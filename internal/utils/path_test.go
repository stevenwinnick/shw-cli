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

func TestResolveRelativeNewRepoPathsUsesDetectedDefaultBranch(t *testing.T) {
	origDetect := detectGitDefaultBranch
	defer func() {
		detectGitDefaultBranch = origDetect
	}()

	detectGitDefaultBranch = func(args []string) (string, error) {
		if len(args) != 1 || args[0] != "--bare" {
			t.Fatalf("got git init args %v", args)
		}
		return "trunk", nil
	}

	paths, err := ResolveRelativeNewRepoPaths("tmp/repo", []string{"--bare"})
	if err != nil {
		t.Fatalf("ResolveRelativeNewRepoPaths returned error: %v", err)
	}

	wantRoot := filepath.Join(mustGetwd(t), "tmp", "repo")
	if paths.RepoRoot != wantRoot {
		t.Fatalf("RepoRoot got %q, want %q", paths.RepoRoot, wantRoot)
	}
	if paths.DefaultBranch != "trunk" {
		t.Fatalf("DefaultBranch got %q, want trunk", paths.DefaultBranch)
	}
	wantMain := filepath.Join(wantRoot, "trunk", "repo")
	if paths.MainWorktreeDir != wantMain {
		t.Fatalf("MainWorktreeDir got %q, want %q", paths.MainWorktreeDir, wantMain)
	}
	wantWorktrees := filepath.Join(wantRoot, "worktrees")
	if paths.WorktreesDir != wantWorktrees {
		t.Fatalf("WorktreesDir got %q, want %q", paths.WorktreesDir, wantWorktrees)
	}
}

func TestEnsureNewRepoLayoutCreatesMainAndWorktreesDirs(t *testing.T) {
	parent := t.TempDir()
	paths := NewRepoPaths{
		MainWorktreeDir: filepath.Join(parent, "repo", "trunk", "repo"),
		WorktreesDir:    filepath.Join(parent, "repo", "worktrees"),
	}

	if err := EnsureNewRepoLayout(paths); err != nil {
		t.Fatalf("EnsureNewRepoLayout returned error: %v", err)
	}

	for _, dir := range []string{paths.MainWorktreeDir, paths.WorktreesDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %q to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", dir)
		}
	}
}

func TestDetectGitDefaultBranchPrefersExplicitInitialBranch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "long flag", args: []string{"--initial-branch", "main"}, want: "main"},
		{name: "short flag", args: []string{"-b", "feature"}, want: "feature"},
		{name: "equals form", args: []string{"--initial-branch=trunk"}, want: "trunk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectGitDefaultBranch(tt.args)
			if err != nil {
				t.Fatalf("DetectGitDefaultBranch returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("DetectGitDefaultBranch got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectGitDefaultBranchRejectsMissingExplicitBranchValue(t *testing.T) {
	for _, args := range [][]string{{"-b"}, {"--initial-branch"}, {"--initial-branch="}} {
		if _, err := DetectGitDefaultBranch(args); err == nil {
			t.Fatalf("expected error for args %v", args)
		}
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

func mustGetwd(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	return cwd
}
