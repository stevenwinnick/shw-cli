package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func SkipIfWindows(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("test uses POSIX shell scripts")
	}
}

func SetWorkingDir(t *testing.T, dir string) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

func SetStdin(t *testing.T, input string) {
	t.Helper()

	file := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(file, []byte(input), 0o644); err != nil {
		t.Fatalf("failed to write stdin fixture: %v", err)
	}

	handle, err := os.Open(file)
	if err != nil {
		t.Fatalf("failed to open stdin fixture: %v", err)
	}

	original := os.Stdin
	os.Stdin = handle

	t.Cleanup(func() {
		os.Stdin = original
		if err := handle.Close(); err != nil {
			t.Fatalf("failed to close stdin fixture: %v", err)
		}
	})
}

func WriteExecutable(t *testing.T, dir string, name string, script string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write executable %q: %v", name, err)
	}

	return path
}

func ReadLines(t *testing.T, path string) []string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %q: %v", path, err)
	}

	text := strings.TrimSpace(string(content))
	if text == "" {
		return nil
	}

	return strings.Split(text, "\n")
}
