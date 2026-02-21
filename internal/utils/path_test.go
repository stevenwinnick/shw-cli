package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRelativeDir(t *testing.T) {
	resolved, err := ResolveRelativeDir("tmp/repo")
	if err != nil {
		t.Fatalf("ResolveRelativeDir returned error: %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("expected absolute path, got %q", resolved)
	}
}

func TestResolveRelativeDirRejectsAbsolutePath(t *testing.T) {
	abs := filepath.Clean(string(filepath.Separator) + "tmp/repo")
	if _, err := ResolveRelativeDir(abs); err == nil {
		t.Fatal("expected error for absolute path")
	}
}

func TestResolveRelativeDirRejectsEmpty(t *testing.T) {
	if _, err := ResolveRelativeDir("   "); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestEnsureDirCreatesDirectory(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "new", "repo")

	if err := EnsureDir(target); err != nil {
		t.Fatalf("EnsureDir returned error: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got non-directory at %q", target)
	}
}

func TestEnsureDirRejectsExistingFile(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "file.txt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if err := EnsureDir(target); err == nil {
		t.Fatal("expected error when path is an existing file")
	}
}

func TestEnsureExistingDirRequiresExistingDirectory(t *testing.T) {
	parent := t.TempDir()
	missing := filepath.Join(parent, "missing")
	if err := EnsureExistingDir(missing); err == nil {
		t.Fatal("expected error for missing directory")
	}

	existing := filepath.Join(parent, "existing")
	if err := os.Mkdir(existing, 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := EnsureExistingDir(existing); err != nil {
		t.Fatalf("EnsureExistingDir returned error: %v", err)
	}
}
