package local

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunCreatesRepoInPromptedDir(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	if err := run([]string{"--initial-branch", "feature"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	repoDir := filepath.Join(rootDir, "repo")
	if !isDir(t, filepath.Join(repoDir, ".git")) {
		t.Fatalf("expected git directory %q to exist", filepath.Join(repoDir, ".git"))
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
