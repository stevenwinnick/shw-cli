package cd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepoPrintsContainerDir(t *testing.T) {
	fixture := setupRepoFixture(t)

	var output bytes.Buffer
	if err := Repo(fixture.repoName, &output); err != nil {
		t.Fatalf("Repo returned error: %v", err)
	}

	want := fixture.containerDir + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestRepoErrorsWhenCodeRootNotSet(t *testing.T) {
	t.Setenv("CODE_ROOT", "")

	var output bytes.Buffer
	err := Repo("some-repo", &output)
	if err == nil {
		t.Fatal("expected error when CODE_ROOT is not set")
	}
	if !strings.Contains(err.Error(), "CODE_ROOT is not set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRepoErrorsWhenRepoNotFound(t *testing.T) {
	t.Setenv("CODE_ROOT", t.TempDir())

	var output bytes.Buffer
	err := Repo("nonexistent", &output)
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultPrintsDefaultBranchWorktree(t *testing.T) {
	fixture := setupRepoFixture(t)

	var output bytes.Buffer
	if err := Default(fixture.mainDir, &output); err != nil {
		t.Fatalf("Default returned error: %v", err)
	}

	want := realPath(t, fixture.mainDir) + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestDefaultWorksFromContainerDir(t *testing.T) {
	fixture := setupRepoFixture(t)

	var output bytes.Buffer
	if err := Default(fixture.containerDir, &output); err != nil {
		t.Fatalf("Default returned error: %v", err)
	}

	want := realPath(t, fixture.mainDir) + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestBranchPrintsWorktreePath(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.createWorktree(t, "steven/feature")

	var output bytes.Buffer
	if err := Branch(fixture.mainDir, "steven/feature", &output); err != nil {
		t.Fatalf("Branch returned error: %v", err)
	}

	want := realPath(t, fixture.worktreePath("steven/feature")) + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestBranchWorksFromContainerDir(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.createWorktree(t, "steven/feature")

	var output bytes.Buffer
	if err := Branch(fixture.containerDir, "steven/feature", &output); err != nil {
		t.Fatalf("Branch returned error: %v", err)
	}

	want := realPath(t, fixture.worktreePath("steven/feature")) + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

func TestBranchErrorsWhenNotFound(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.createWorktree(t, "steven/feature")

	err := Branch(fixture.mainDir, "missing", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for missing branch")
	}
	if !strings.Contains(err.Error(), "steven/feature") {
		t.Fatalf("expected available branches in error, got: %v", err)
	}
}

func TestBranchWorksFromSubdirectory(t *testing.T) {
	fixture := setupRepoFixture(t)
	wtPath := fixture.createWorktree(t, "steven/feature")

	subDir := filepath.Join(fixture.mainDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	var output bytes.Buffer
	if err := Branch(subDir, "steven/feature", &output); err != nil {
		t.Fatalf("Branch returned error: %v", err)
	}

	want := realPath(t, wtPath) + "\n"
	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

type repoFixture struct {
	repoName      string
	containerDir  string
	mainDir       string
	remoteDir     string
	defaultBranch string
}

func setupRepoFixture(t *testing.T) repoFixture {
	t.Helper()

	rootDir := t.TempDir()
	repoName := "demo"
	containerDir := filepath.Join(rootDir, repoName)
	mainDir := filepath.Join(containerDir, "trunk", repoName)
	remoteDir := filepath.Join(rootDir, "remote", repoName+".git")

	if err := os.MkdirAll(mainDir, 0o755); err != nil {
		t.Fatalf("failed to create main dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(remoteDir), 0o755); err != nil {
		t.Fatalf("failed to create remote parent dir: %v", err)
	}

	runGit(t, "", "init", "--bare", "--initial-branch", "trunk", remoteDir)
	runGit(t, "", "init", "--initial-branch", "trunk", mainDir)
	runGit(t, mainDir, "config", "user.name", "Test User")
	runGit(t, mainDir, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(mainDir, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("failed to write README: %v", err)
	}
	runGit(t, mainDir, "add", "README.md")
	runGit(t, mainDir, "commit", "-m", "initial")
	runGit(t, mainDir, "remote", "add", "origin", remoteDir)
	runGit(t, mainDir, "push", "-u", "origin", "trunk")

	t.Setenv("CODE_ROOT", rootDir)

	return repoFixture{
		repoName:      repoName,
		containerDir:  containerDir,
		mainDir:       mainDir,
		remoteDir:     remoteDir,
		defaultBranch: "trunk",
	}
}

func (f repoFixture) worktreePath(branch string) string {
	return filepath.Join(f.containerDir, "worktrees", strings.ReplaceAll(branch, "/", "--"), f.repoName)
}

func (f repoFixture) createWorktree(t *testing.T, branch string) string {
	t.Helper()

	wtPath := f.worktreePath(branch)
	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		t.Fatalf("failed to create worktree parent dir: %v", err)
	}
	runGit(t, f.mainDir, "worktree", "add", "-b", branch, wtPath, f.defaultBranch)
	return wtPath
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return string(output)
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
