package push

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCallsGitPushWithUpstreamToHead(t *testing.T) {
	rootDir := t.TempDir()
	remoteDir := filepath.Join(rootDir, "remote.git")
	localDir := filepath.Join(rootDir, "local")

	runGit(t, "", "init", "--bare", "--initial-branch", "trunk", remoteDir)
	runGit(t, "", "init", "--initial-branch", "trunk", localDir)
	runGit(t, localDir, "config", "user.name", "Test User")
	runGit(t, localDir, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(localDir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	runGit(t, localDir, "add", "README.md")
	runGit(t, localDir, "commit", "-m", "initial")
	runGit(t, localDir, "remote", "add", "origin", remoteDir)

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(localDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	upstream := strings.TrimSpace(string(runGit(t, localDir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")))
	if upstream != "origin/trunk" {
		t.Fatalf("upstream got %q, want %q", upstream, "origin/trunk")
	}

	if got := strings.TrimSpace(string(runGit(t, "", "--git-dir", remoteDir, "show-ref", "--verify", "--hash", "refs/heads/trunk"))); got == "" {
		t.Fatal("expected remote trunk ref to exist after push")
	}
}

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "push" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if cmd.Usage != "shw git push" {
		t.Fatalf("unexpected usage: %q", cmd.Usage)
	}
	if cmd.Run == nil {
		t.Fatal("expected Run handler")
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
