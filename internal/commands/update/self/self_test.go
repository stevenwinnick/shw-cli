package self

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "self" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if cmd.Usage != "shw update self" {
		t.Fatalf("unexpected usage: %q", cmd.Usage)
	}
	if cmd.Run == nil {
		t.Fatal("expected Run handler")
	}
}

func TestResolveRepoDir(t *testing.T) {
	t.Run("returns error when CODE_ROOT is not set", func(t *testing.T) {
		t.Setenv("CODE_ROOT", "")
		_, err := resolveRepoDir()
		if err == nil {
			t.Fatal("expected error when CODE_ROOT is empty")
		}
	})

	t.Run("returns error when repo dir does not exist", func(t *testing.T) {
		t.Setenv("CODE_ROOT", t.TempDir())
		_, err := resolveRepoDir()
		if err == nil {
			t.Fatal("expected error when repo dir does not exist")
		}
	})

	t.Run("returns repo dir when it exists", func(t *testing.T) {
		codeRoot := t.TempDir()
		repoDir := filepath.Join(codeRoot, "shw-cli")
		if err := os.MkdirAll(repoDir, 0o755); err != nil {
			t.Fatalf("failed to create repo dir: %v", err)
		}

		t.Setenv("CODE_ROOT", codeRoot)
		got, err := resolveRepoDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != repoDir {
			t.Fatalf("got %q, want %q", got, repoDir)
		}
	})
}

func TestRunPullsAndInstalls(t *testing.T) {
	testutil.SkipIfWindows(t)

	// Set up a fake shw-cli checkout on trunk
	codeRoot := t.TempDir()
	repoDir := filepath.Join(codeRoot, "shw-cli")
	scriptsDir := filepath.Join(repoDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	// Initialize a git repo so git pull has something to work with
	runGit(t, repoDir, "init", "--initial-branch", "trunk")
	runGit(t, repoDir, "config", "user.name", "Test User")
	runGit(t, repoDir, "config", "user.email", "test@example.com")

	// Create a marker file that install.sh will touch
	markerFile := filepath.Join(t.TempDir(), "install-ran")
	testutil.WriteExecutable(t, scriptsDir, "install.sh", "#!/bin/bash\ntouch "+markerFile+"\n")

	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "initial")

	// Set up a bare remote so git pull works
	remoteDir := filepath.Join(t.TempDir(), "remote.git")
	runGit(t, "", "clone", "--bare", repoDir, remoteDir)
	runGit(t, repoDir, "remote", "add", "origin", remoteDir)
	runGit(t, repoDir, "fetch", "origin")
	runGit(t, repoDir, "branch", "--set-upstream-to=origin/trunk", "trunk")

	t.Setenv("CODE_ROOT", codeRoot)

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if _, err := os.Stat(markerFile); err != nil {
		t.Fatalf("install.sh did not run: marker file not found: %v", err)
	}
}

func TestRunRejectsNonTrunkBranch(t *testing.T) {
	codeRoot := t.TempDir()
	repoDir := filepath.Join(codeRoot, "shw-cli")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}
	runGit(t, repoDir, "init", "--initial-branch", "steven/feature")

	t.Setenv("CODE_ROOT", codeRoot)

	err := run(nil)
	if err == nil || !strings.Contains(err.Error(), "switch to trunk first") {
		t.Fatalf("expected non-trunk branch error, got %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
