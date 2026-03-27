package local

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunUsesWorktreeSetupAndGitInit(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	if err := run([]string{"--initial-branch", "trunk", "--bare"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	repoRoot := filepath.Join(rootDir, "repo")
	bareRepoDir := filepath.Join(repoRoot, "trunk", "repo")
	if !isDir(t, bareRepoDir) {
		t.Fatalf("expected bare repo directory %q to exist", bareRepoDir)
	}
	if !isDir(t, filepath.Join(repoRoot, "worktrees")) {
		t.Fatalf("expected worktrees directory %q to exist", filepath.Join(repoRoot, "worktrees"))
	}

	if got := strings.TrimSpace(string(runGit(t, "", "--git-dir", bareRepoDir, "rev-parse", "--is-bare-repository"))); got != "true" {
		t.Fatalf("expected bare repository, got %q", got)
	}
	head := strings.TrimSpace(string(mustReadFile(t, filepath.Join(bareRepoDir, "HEAD"))))
	if head != "ref: refs/heads/trunk" {
		t.Fatalf("HEAD got %q, want %q", head, "ref: refs/heads/trunk")
	}
}

func TestRunAllowsSkippingWorktreeSetup(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	if err := run([]string{"--no-worktree-setup", "--initial-branch", "feature"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	repoDir := filepath.Join(rootDir, "repo")
	if !isDir(t, filepath.Join(repoDir, ".git")) {
		t.Fatalf("expected git directory %q to exist", filepath.Join(repoDir, ".git"))
	}
	if _, err := os.Stat(filepath.Join(repoDir, "worktrees")); !os.IsNotExist(err) {
		t.Fatalf("expected worktrees directory to be absent, stat error: %v", err)
	}

	currentBranch := strings.TrimSpace(string(runGit(t, repoDir, "branch", "--show-current")))
	if currentBranch != "feature" {
		t.Fatalf("current branch got %q, want %q", currentBranch, "feature")
	}
}

func TestRunStopsOnPromptError(t *testing.T) {
	testutil.SetStdin(t, "")

	err := run(nil)
	if err == nil {
		t.Fatal("expected prompt error")
	}
	if !strings.Contains(err.Error(), "directory path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %q: %v", path, err)
	}

	return content
}

func realPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to resolve path %q: %v", path, err)
	}

	return absolute
}

func runGit(t *testing.T, dir string, args ...string) []byte {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}

	return output
}

func isDir(t *testing.T, path string) bool {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
