package commit

import (
	"os/exec"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "commit" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if len(cmd.Children) != 1 {
		t.Fatalf("unexpected child count: %d", len(cmd.Children))
	}

	empty := cmd.Children[0]
	if empty.Name != "emptyasshw" {
		t.Fatalf("unexpected child name: %q", empty.Name)
	}
	if empty.Usage != "shw git commit emptyasshw [<message>]" {
		t.Fatalf("unexpected usage: %q", empty.Usage)
	}
	if empty.Run == nil {
		t.Fatal("expected Run handler")
	}
}

func TestRunEmptyAsShwCreatesEmptyCommitWithShwIdentity(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "--initial-branch", "trunk")
	runGit(t, repoDir, "config", "user.name", "Robots User")
	runGit(t, repoDir, "config", "user.email", "robots@example.com")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	testutil.SetWorkingDir(t, repoDir)

	if err := runEmptyAsShw(nil); err != nil {
		t.Fatalf("runEmptyAsShw returned error: %v", err)
	}

	got := strings.TrimSpace(string(runGit(t, repoDir, "log", "-1", "--format=%an|%ae|%cn|%ce|%s")))
	want := strings.Join([]string{shwName, shwEmail, shwName, shwEmail, defaultEmptyMsg}, "|")
	if got != want {
		t.Fatalf("commit metadata got %q, want %q", got, want)
	}

	if diff := strings.TrimSpace(string(runGit(t, repoDir, "diff", "--name-only", "HEAD~1", "HEAD"))); diff != "" {
		t.Fatalf("expected an empty commit, got changes: %q", diff)
	}
}

func TestRunEmptyAsShwUsesProvidedMessage(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "--initial-branch", "trunk")
	runGit(t, repoDir, "config", "user.name", "Robots User")
	runGit(t, repoDir, "config", "user.email", "robots@example.com")
	runGit(t, repoDir, "commit", "--allow-empty", "-m", "initial")
	testutil.SetWorkingDir(t, repoDir)

	if err := runEmptyAsShw([]string{"Custom message"}); err != nil {
		t.Fatalf("runEmptyAsShw returned error: %v", err)
	}

	if got := strings.TrimSpace(string(runGit(t, repoDir, "log", "-1", "--format=%s"))); got != "Custom message" {
		t.Fatalf("commit subject got %q, want %q", got, "Custom message")
	}
}

func TestRunEmptyAsShwRejectsExtraArguments(t *testing.T) {
	if err := runEmptyAsShw([]string{"first", "second"}); err == nil {
		t.Fatal("expected error when extra arguments are passed")
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
