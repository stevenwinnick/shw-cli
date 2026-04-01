package self

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestResolveTrunkDir(t *testing.T) {
	t.Run("returns error when CODE_ROOT is not set", func(t *testing.T) {
		t.Setenv("CODE_ROOT", "")
		_, err := resolvetrunkDir()
		if err == nil {
			t.Fatal("expected error when CODE_ROOT is empty")
		}
	})

	t.Run("returns error when trunk dir does not exist", func(t *testing.T) {
		t.Setenv("CODE_ROOT", t.TempDir())
		_, err := resolvetrunkDir()
		if err == nil {
			t.Fatal("expected error when trunk dir does not exist")
		}
	})

	t.Run("returns trunk dir when it exists", func(t *testing.T) {
		codeRoot := t.TempDir()
		trunkDir := filepath.Join(codeRoot, "shw-cli", "trunk", "shw-cli")
		if err := os.MkdirAll(trunkDir, 0o755); err != nil {
			t.Fatalf("failed to create trunk dir: %v", err)
		}

		t.Setenv("CODE_ROOT", codeRoot)
		got, err := resolvetrunkDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != trunkDir {
			t.Fatalf("got %q, want %q", got, trunkDir)
		}
	})
}

func TestRunPullsAndInstalls(t *testing.T) {
	testutil.SkipIfWindows(t)

	// Set up a fake git repo as the trunk
	codeRoot := t.TempDir()
	trunkDir := filepath.Join(codeRoot, "shw-cli", "trunk", "shw-cli")
	scriptsDir := filepath.Join(trunkDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	// Initialize a git repo so git pull has something to work with
	runGit(t, trunkDir, "init", "--initial-branch", "trunk")
	runGit(t, trunkDir, "config", "user.name", "Test User")
	runGit(t, trunkDir, "config", "user.email", "test@example.com")

	// Create a marker file that install.sh will touch
	markerFile := filepath.Join(t.TempDir(), "install-ran")
	testutil.WriteExecutable(t, scriptsDir, "install.sh", "#!/bin/bash\ntouch "+markerFile+"\n")

	runGit(t, trunkDir, "add", ".")
	runGit(t, trunkDir, "commit", "-m", "initial")

	// Set up a bare remote so git pull works
	remoteDir := filepath.Join(t.TempDir(), "remote.git")
	runGit(t, "", "clone", "--bare", trunkDir, remoteDir)
	runGit(t, trunkDir, "remote", "add", "origin", remoteDir)
	runGit(t, trunkDir, "fetch", "origin")
	runGit(t, trunkDir, "branch", "--set-upstream-to=origin/trunk", "trunk")

	t.Setenv("CODE_ROOT", codeRoot)

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if _, err := os.Stat(markerFile); err != nil {
		t.Fatalf("install.sh did not run: marker file not found: %v", err)
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
