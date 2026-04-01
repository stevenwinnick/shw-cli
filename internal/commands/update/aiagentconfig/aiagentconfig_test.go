package aiagentconfig

import (
	"os"
	"path/filepath"
	"testing"

	"shw-cli/internal/testutil"
)

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "ai-agent-config" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if cmd.Usage != "shw update ai-agent-config" {
		t.Fatalf("unexpected usage: %q", cmd.Usage)
	}
	if cmd.Run == nil {
		t.Fatal("expected Run handler")
	}
}

func TestRunCallsApplyScript(t *testing.T) {
	testutil.SkipIfWindows(t)

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	scriptsDir := filepath.Join(homeDir, ".ai-agent-config", "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	markerFile := filepath.Join(t.TempDir(), "apply-ran")
	testutil.WriteExecutable(t, scriptsDir, "apply-updated-trunk-config.sh", "#!/bin/bash\ntouch "+markerFile+"\n")

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if _, err := os.Stat(markerFile); err != nil {
		t.Fatalf("apply script did not run: marker file not found: %v", err)
	}
}

func TestRunReturnsErrorWhenScriptMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	if err := run(nil); err == nil {
		t.Fatal("expected error when script does not exist")
	}
}
