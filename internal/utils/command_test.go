package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunCommandInDirUsesProvidedDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell")
	}

	targetDir := t.TempDir()
	marker := filepath.Join(targetDir, "marker.txt")

	err := RunCommandInDir(targetDir, "sh", "-c", "touch marker.txt")
	if err != nil {
		t.Fatalf("RunCommandInDir returned error: %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("expected marker file to be created in target dir: %v", err)
	}
}

func TestRunCommandReturnsWrappedError(t *testing.T) {
	err := RunCommand("definitely-not-a-real-command-binary")
	if err == nil {
		t.Fatal("expected command error")
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Fatalf("expected wrapped error, got: %v", err)
	}
}
