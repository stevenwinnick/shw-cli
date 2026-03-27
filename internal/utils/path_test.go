package utils

import (
	"os"
	"path/filepath"
	"testing"

	"shw-cli/internal/testutil"
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

func TestResolveRelativeRepoTargetPathsUsesDetectedDefaultBranch(t *testing.T) {
	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	paths, err := ResolveRelativeRepoTargetPaths("tmp/repo", []string{"--initial-branch", "trunk"}, true)
	if err != nil {
		t.Fatalf("ResolveRelativeRepoTargetPaths returned error: %v", err)
	}

	wantRoot := filepath.Join(workingDir, "tmp", "repo")
	if realPath(t, paths.RepoRoot) != wantRoot {
		t.Fatalf("RepoRoot got %q, want %q", paths.RepoRoot, wantRoot)
	}
	if paths.DefaultBranch != "trunk" {
		t.Fatalf("DefaultBranch got %q, want trunk", paths.DefaultBranch)
	}
	wantMain := filepath.Join(wantRoot, "trunk", "repo")
	if paths.WorkingDir != wantMain {
		t.Fatalf("WorkingDir got %q, want %q", paths.WorkingDir, wantMain)
	}
	wantWorktrees := filepath.Join(wantRoot, "worktrees")
	if paths.AdditionalWorkDir != wantWorktrees {
		t.Fatalf("AdditionalWorkDir got %q, want %q", paths.AdditionalWorkDir, wantWorktrees)
	}
	if !paths.UsesWorktrees {
		t.Fatal("expected UsesWorktrees to be true")
	}
}

func TestResolveRelativeRepoTargetPathsWithoutWorktreesUsesRepoRoot(t *testing.T) {
	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	paths, err := ResolveRelativeRepoTargetPaths("tmp/repo", []string{"--bare"}, false)
	if err != nil {
		t.Fatalf("ResolveRelativeRepoTargetPaths returned error: %v", err)
	}

	wantRoot := filepath.Join(workingDir, "tmp", "repo")
	if realPath(t, paths.RepoRoot) != wantRoot {
		t.Fatalf("RepoRoot got %q, want %q", paths.RepoRoot, wantRoot)
	}
	if paths.WorkingDir != wantRoot {
		t.Fatalf("WorkingDir got %q, want %q", paths.WorkingDir, wantRoot)
	}
	if paths.DefaultBranch != "" {
		t.Fatalf("DefaultBranch got %q, want empty", paths.DefaultBranch)
	}
	if paths.AdditionalWorkDir != "" {
		t.Fatalf("AdditionalWorkDir got %q, want empty", paths.AdditionalWorkDir)
	}
	if paths.UsesWorktrees {
		t.Fatal("expected UsesWorktrees to be false")
	}
}

func TestEnsureRepoTargetPathsCreatesWorkingAndAdditionalDirs(t *testing.T) {
	parent := t.TempDir()
	paths := RepoTargetPaths{
		WorkingDir:        filepath.Join(parent, "repo", "trunk", "repo"),
		AdditionalWorkDir: filepath.Join(parent, "repo", "worktrees"),
	}

	if err := EnsureRepoTargetPaths(paths); err != nil {
		t.Fatalf("EnsureRepoTargetPaths returned error: %v", err)
	}

	for _, dir := range []string{paths.WorkingDir, paths.AdditionalWorkDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %q to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %q to be a directory", dir)
		}
	}
}

func TestEnsureRepoTargetPathsWithoutAdditionalDirCreatesOnlyWorkingDir(t *testing.T) {
	parent := t.TempDir()
	paths := RepoTargetPaths{
		WorkingDir: filepath.Join(parent, "repo"),
	}

	if err := EnsureRepoTargetPaths(paths); err != nil {
		t.Fatalf("EnsureRepoTargetPaths returned error: %v", err)
	}

	info, err := os.Stat(paths.WorkingDir)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", paths.WorkingDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", paths.WorkingDir)
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

func realPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to resolve path %q: %v", path, err)
	}

	return absolute
}
