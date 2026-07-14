package pr

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"shw-cli/internal/commands/git/worktree"
	"shw-cli/internal/testutil"
)

func TestMergeSquashMergesAndRemovesWorktree(t *testing.T) {
	f := setupFixture(t)
	ghArgsLog := installFakeGh(t, 0, "MERGED")
	worktreePath := f.createWorktree(t, "steven/feature")

	// Merge from within the worktree, letting the branch default to the one
	// checked out there.
	if err := merge(worktreePath, "", io.Discard, io.Discard); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}

	assertWorktreeRemoved(t, f, worktreePath, "steven/feature")

	args := testutil.ReadLines(t, ghArgsLog)
	if len(args) != 1 {
		t.Fatalf("expected gh to be called once, got %d call(s): %v", len(args), args)
	}
	if args[0] != "pr merge steven/feature --squash --delete-branch" {
		t.Fatalf("unexpected gh merge invocation: %q", args[0])
	}
}

func TestMergeAcceptsExplicitBranchArgument(t *testing.T) {
	f := setupFixture(t)
	installFakeGh(t, 0, "MERGED")
	worktreePath := f.createWorktree(t, "steven/feature")

	// Run from the main worktree and name the branch explicitly.
	if err := merge(f.mainDir, "steven/feature", io.Discard, io.Discard); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}

	assertWorktreeRemoved(t, f, worktreePath, "steven/feature")
}

func TestMergeToleratesLocalBranchDeletionFailure(t *testing.T) {
	f := setupFixture(t)
	// gh fails (as it does when it cannot delete the checked-out local branch),
	// but the PR is reported as merged, so cleanup should still proceed.
	ghArgsLog := installFakeGh(t, 1, "MERGED")
	worktreePath := f.createWorktree(t, "steven/feature")

	if err := merge(worktreePath, "", io.Discard, io.Discard); err != nil {
		t.Fatalf("merge returned error: %v", err)
	}

	assertWorktreeRemoved(t, f, worktreePath, "steven/feature")

	args := testutil.ReadLines(t, ghArgsLog)
	if len(args) != 2 {
		t.Fatalf("expected gh merge then gh view, got: %v", args)
	}
	if !strings.HasPrefix(args[1], "pr view steven/feature") {
		t.Fatalf("expected a gh pr view call to check merge state, got: %q", args[1])
	}
}

func TestMergeReturnsErrorWhenMergeFails(t *testing.T) {
	f := setupFixture(t)
	// gh fails and the PR is not merged, so this is a genuine merge failure.
	installFakeGh(t, 1, "OPEN")
	worktreePath := f.createWorktree(t, "steven/feature")

	if err := merge(worktreePath, "", io.Discard, io.Discard); err == nil {
		t.Fatal("expected merge to return an error when the PR did not merge")
	}

	// The worktree and branch must be left intact so no work is lost.
	if _, err := os.Stat(worktreePath); err != nil {
		t.Fatalf("expected worktree %q to still exist after a failed merge: %v", worktreePath, err)
	}
	if !branchRefExists(t, f.mainDir, "steven/feature") {
		t.Fatal("expected local branch to still exist after a failed merge")
	}
}

// Fixture and helpers.

type fixture struct {
	repoName     string
	containerDir string
	mainDir      string
}

func setupFixture(t *testing.T) fixture {
	t.Helper()
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	repoName := "demo"
	containerDir := filepath.Join(rootDir, repoName)
	mainDir := filepath.Join(containerDir, "trunk", repoName)
	remoteDir := filepath.Join(rootDir, "remote", repoName+".git")

	mustMkdirAll(t, mainDir)
	mustMkdirAll(t, filepath.Dir(remoteDir))

	runGit(t, "", "init", "--bare", "--initial-branch", "trunk", remoteDir)
	runGit(t, "", "init", "--initial-branch", "trunk", mainDir)
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "initial")
	runGit(t, mainDir, "remote", "add", "origin", remoteDir)
	runGit(t, mainDir, "push", "-u", "origin", "trunk")

	return fixture{repoName: repoName, containerDir: containerDir, mainDir: mainDir}
}

func (f fixture) worktreePath(branch string) string {
	return filepath.Join(f.containerDir, "worktrees", strings.ReplaceAll(branch, "/", "--"), f.repoName)
}

func (f fixture) createWorktree(t *testing.T, branch string) string {
	t.Helper()

	if err := worktree.Create(f.mainDir, branch); err != nil {
		t.Fatalf("failed to create worktree for %q: %v", branch, err)
	}
	return f.worktreePath(branch)
}

// installFakeGh puts a fake gh on PATH whose `pr merge` exits with mergeExit and
// whose `pr view` reports viewState. It returns the path to the args log.
func installFakeGh(t *testing.T, mergeExit int, viewState string) string {
	t.Helper()

	binDir := t.TempDir()
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	script := "#!/bin/sh\n" +
		// Record each invocation's arguments, one line per call.
		"printf '%s\\n' \"$*\" >>\"$FAKE_GH_ARGS_FILE\"\n" +
		"case \"$2\" in\n" +
		"  merge) exit ${FAKE_GH_MERGE_EXIT} ;;\n" +
		"  view) printf '%s\\n' \"${FAKE_GH_VIEW_STATE}\" ;;\n" +
		"esac\n" +
		"exit 0\n"
	testutil.WriteExecutable(t, binDir, "gh", script)

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)
	t.Setenv("FAKE_GH_MERGE_EXIT", fmt.Sprintf("%d", mergeExit))
	t.Setenv("FAKE_GH_VIEW_STATE", viewState)
	return ghArgsLog
}

func assertWorktreeRemoved(t *testing.T, f fixture, worktreePath string, branch string) {
	t.Helper()

	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Fatalf("expected worktree %q to be removed, stat err=%v", worktreePath, err)
	}
	if branchRefExists(t, f.mainDir, branch) {
		t.Fatalf("expected local branch %q to be deleted", branch)
	}
}

func branchRefExists(t *testing.T, dir string, branch string) bool {
	t.Helper()

	cmd := exec.Command("git", "show-ref", "--verify", "refs/heads/"+branch)
	cmd.Dir = dir
	return cmd.Run() == nil
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create dir %q: %v", dir, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return string(output)
}
