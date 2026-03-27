package localandgithub

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunCallsGitThenGitHubInTargetDir(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")
	t.Setenv("GIT_CONFIG_GLOBAL", writeGlobalGitConfig(t, "[init]\n\tdefaultBranch = trunk\n"))

	binDir := t.TempDir()
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)

	if err := run([]string{"my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	repoRoot := filepath.Join(rootDir, "repo")
	repoDir := filepath.Join(repoRoot, "trunk", "repo")
	if !isDir(t, filepath.Join(repoDir, ".git")) {
		t.Fatalf("expected git directory %q to exist", filepath.Join(repoDir, ".git"))
	}
	if !isDir(t, filepath.Join(repoRoot, "worktrees")) {
		t.Fatalf("expected worktrees directory %q to exist", filepath.Join(repoRoot, "worktrees"))
	}

	currentBranch := strings.TrimSpace(string(runGit(t, repoDir, "branch", "--show-current")))
	if currentBranch != "trunk" {
		t.Fatalf("current branch got %q, want %q", currentBranch, "trunk")
	}

	wantDir := realPath(t, repoDir)
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, ghDirLog)))); got != wantDir {
		t.Fatalf("gh ran in %q, want %q", got, wantDir)
	}

	assertArgs(t, testutil.ReadLines(t, ghArgsLog), []string{"repo", "create", "my-repo", "--private"})
}

func TestRunAllowsSkippingWorktreeSetup(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)

	if err := run([]string{"--no-worktree-setup", "my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	repoDir := filepath.Join(rootDir, "repo")
	if !isDir(t, filepath.Join(repoDir, ".git")) {
		t.Fatalf("expected git directory %q to exist", filepath.Join(repoDir, ".git"))
	}
	if _, err := os.Stat(filepath.Join(rootDir, "repo", "worktrees")); !os.IsNotExist(err) {
		t.Fatalf("expected worktrees directory to be absent, stat error: %v", err)
	}

	wantDir := realPath(t, repoDir)
	currentBranch := strings.TrimSpace(string(runGit(t, repoDir, "branch", "--show-current")))
	if currentBranch == "" {
		t.Fatal("expected a current branch in the created repository")
	}
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, ghDirLog)))); got != wantDir {
		t.Fatalf("gh ran in %q, want %q", got, wantDir)
	}

	assertArgs(t, testutil.ReadLines(t, ghArgsLog), []string{"repo", "create", "my-repo", "--private"})
}

func TestRunStopsOnPromptErrorBeforeGhRuns(t *testing.T) {
	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "")

	binDir := t.TempDir()
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	testutil.SkipIfWindows(t)
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)

	err := run(nil)
	if err == nil {
		t.Fatal("expected prompt error")
	}
	if !strings.Contains(err.Error(), "directory path is required") {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(ghDirLog); !os.IsNotExist(statErr) {
		t.Fatalf("expected gh not to run, stat error: %v", statErr)
	}
}

func fakeGhScript() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return `#!/bin/sh
printf '%s\n' "$PWD" >"$FAKE_GH_DIR_FILE"
printf '%s\n' "$@" >"$FAKE_GH_ARGS_FILE"
`
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

func assertArgs(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("arg length mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
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

func writeGlobalGitConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	return path
}
