package worktree

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCreatesWorktreeInPreferredLayout(t *testing.T) {
	fixture := setupRepoFixture(t)

	var output bytes.Buffer
	if err := create(fixture.mainDir, "steven/add-worktree-commands", true, &output, io.Discard); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	worktreePath := fixture.worktreePath("steven/add-worktree-commands")
	resolvedWorktreePath := realPath(t, worktreePath)
	if info, err := os.Stat(worktreePath); err != nil || !info.IsDir() {
		t.Fatalf("expected worktree directory %q to exist: %v", worktreePath, err)
	}

	listing := runGit(t, fixture.mainDir, "worktree", "list", "--porcelain")
	if !strings.Contains(listing, "worktree "+resolvedWorktreePath) {
		t.Fatalf("worktree list missing path %q:\n%s", resolvedWorktreePath, listing)
	}
	if !strings.Contains(listing, "branch refs/heads/steven/add-worktree-commands") {
		t.Fatalf("worktree list missing branch entry:\n%s", listing)
	}
	if !strings.Contains(output.String(), "Worktree created: "+resolvedWorktreePath) {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestCommandsAcceptRepoContainerDir(t *testing.T) {
	fixture := setupRepoFixture(t)

	var createOutput bytes.Buffer
	if err := create(fixture.containerDir, "steven/container-path", true, &createOutput, io.Discard); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	worktreePath := realPath(t, fixture.worktreePath("steven/container-path"))
	if !strings.Contains(createOutput.String(), "Worktree created: "+worktreePath) {
		t.Fatalf("create stdout got %q, want message containing %q", createOutput.String(), "Worktree created: "+worktreePath)
	}

	var pathOutput bytes.Buffer
	if err := switchTo(fixture.containerDir, "steven/container-path", &pathOutput); err != nil {
		t.Fatalf("switchTo returned error: %v", err)
	}
	if pathOutput.String() != worktreePath+"\n" {
		t.Fatalf("path stdout got %q, want %q", pathOutput.String(), worktreePath+"\n")
	}

	var listOutput bytes.Buffer
	if err := list(fixture.containerDir, &listOutput, io.Discard); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if !strings.Contains(listOutput.String(), worktreePath) {
		t.Fatalf("list output missing worktree path:\n%s", listOutput.String())
	}

	var removeOutput bytes.Buffer
	if err := remove(fixture.containerDir, "steven/container-path", &removeOutput, io.Discard); err != nil {
		t.Fatalf("remove returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(worktreePath)); !os.IsNotExist(err) {
		t.Fatalf("expected worktree parent directory to be removed, got err=%v", err)
	}
}

func TestListPrintsOnlyWorktreeListByDefault(t *testing.T) {
	fixture := setupRepoFixture(t)

	featurePath := fixture.createWorktree(t, "steven/feature")

	if err := os.WriteFile(filepath.Join(fixture.mainDir, "README.md"), []byte("updated\n"), 0o644); err != nil {
		t.Fatalf("failed to update main README: %v", err)
	}
	if err := os.WriteFile(filepath.Join(featurePath, "feature.txt"), []byte("feature\n"), 0o644); err != nil {
		t.Fatalf("failed to create feature file: %v", err)
	}

	var output bytes.Buffer
	if err := list(fixture.mainDir, &output, io.Discard); err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	text := output.String()
	if !strings.Contains(text, realPath(t, fixture.mainDir)) {
		t.Fatalf("missing main worktree path in output:\n%s", text)
	}
	if !strings.Contains(text, realPath(t, featurePath)) {
		t.Fatalf("missing feature worktree path in output:\n%s", text)
	}
	if strings.Contains(text, "README.md") || strings.Contains(text, "feature.txt") {
		t.Fatalf("default list output should not include status details:\n%s", text)
	}
}

func TestRemoveDeletesWorktreeDirectoryAndLocalBranch(t *testing.T) {
	fixture := setupRepoFixture(t)
	worktreePath := fixture.createWorktree(t, "steven/feature")

	var output bytes.Buffer
	if err := remove(fixture.mainDir, "steven/feature", &output, io.Discard); err != nil {
		t.Fatalf("remove returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(worktreePath)); !os.IsNotExist(err) {
		t.Fatalf("expected worktree parent directory to be removed, got err=%v", err)
	}
	if got := strings.TrimSpace(runGitAllowFailure(t, fixture.mainDir, "branch", "--list", "steven/feature")); got != "" {
		t.Fatalf("expected local branch to be deleted, got %q", got)
	}
	if strings.Contains(runGit(t, fixture.mainDir, "worktree", "list", "--porcelain"), "steven/feature") {
		t.Fatalf("expected worktree to be removed from git worktree list")
	}
	if !strings.Contains(output.String(), "Worktree removed successfully") {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestCleanAllRemovesOnlyBranchesMissingRemoteTrackingRefs(t *testing.T) {
	fixture := setupRepoFixture(t)
	stalePath := fixture.createWorktree(t, "steven/stale")
	keptPath := fixture.createWorktree(t, "steven/kept")

	runGit(t, keptPath, "push", "-u", "origin", "HEAD")

	var output bytes.Buffer
	if err := cleanAll(fixture.mainDir, &output, io.Discard); err != nil {
		t.Fatalf("cleanAll returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Dir(stalePath)); !os.IsNotExist(err) {
		t.Fatalf("expected stale worktree parent directory to be removed, got err=%v", err)
	}
	if _, err := os.Stat(keptPath); err != nil {
		t.Fatalf("expected kept worktree to remain, got err=%v", err)
	}
	if got := strings.TrimSpace(runGitAllowFailure(t, fixture.mainDir, "branch", "--list", "steven/stale")); got != "" {
		t.Fatalf("expected stale branch to be deleted, got %q", got)
	}
	if got := strings.TrimSpace(runGit(t, fixture.mainDir, "branch", "--list", "steven/kept")); got == "" {
		t.Fatal("expected kept branch to remain")
	}
	if !strings.Contains(output.String(), "branch 'steven/stale' not on remote") {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestSwitchPrintsResolvedPath(t *testing.T) {
	fixture := setupRepoFixture(t)
	worktreePath := fixture.createWorktree(t, "steven/feature")

	var output bytes.Buffer
	if err := switchTo(fixture.mainDir, "steven/feature", &output); err != nil {
		t.Fatalf("switchTo returned error: %v", err)
	}

	want := realPath(t, worktreePath) + "\n"
	if output.String() != want {
		t.Fatalf("stdout got %q, want %q", output.String(), want)
	}
}

func TestSwitchListsAvailableNamesWhenTargetIsMissing(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.createWorktree(t, "steven/feature")

	err := switchTo(fixture.mainDir, "missing", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected missing worktree error")
	}
	if !strings.Contains(err.Error(), "steven/feature") {
		t.Fatalf("error got %q, want available branch name", err.Error())
	}
}

func TestCreateUpdatesDefaultBranchBeforeBranchingByDefault(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.advanceDefaultBranchOnRemote(t)

	var output bytes.Buffer
	if err := create(fixture.mainDir, "steven/from-updated-default", true, &output, io.Discard); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	worktreePath := fixture.worktreePath("steven/from-updated-default")
	if got := readFile(t, filepath.Join(worktreePath, "remote.txt")); got != "from remote\n" {
		t.Fatalf("remote update missing from created worktree, got %q", got)
	}
}

func TestCreateCanSkipDefaultBranchUpdate(t *testing.T) {
	fixture := setupRepoFixture(t)
	fixture.advanceDefaultBranchOnRemote(t)

	var output bytes.Buffer
	if err := create(fixture.mainDir, "steven/from-stale-default", false, &output, io.Discard); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	worktreePath := fixture.worktreePath("steven/from-stale-default")
	if _, err := os.Stat(filepath.Join(worktreePath, "remote.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected remote update to be absent without fetch, got err=%v", err)
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

	var output bytes.Buffer
	if err := create(f.mainDir, branch, true, &output, io.Discard); err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	return f.worktreePath(branch)
}

func (f repoFixture) advanceDefaultBranchOnRemote(t *testing.T) {
	t.Helper()

	cloneDir := filepath.Join(t.TempDir(), "clone")
	runGit(t, "", "clone", f.remoteDir, cloneDir)
	runGit(t, cloneDir, "config", "user.name", "Remote User")
	runGit(t, cloneDir, "config", "user.email", "remote@example.com")
	if err := os.WriteFile(filepath.Join(cloneDir, "remote.txt"), []byte("from remote\n"), 0o644); err != nil {
		t.Fatalf("failed to write remote file: %v", err)
	}
	runGit(t, cloneDir, "add", "remote.txt")
	runGit(t, cloneDir, "commit", "-m", "remote update")
	runGit(t, cloneDir, "push", "origin", f.defaultBranch)
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

func runGitAllowFailure(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output)
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

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %q: %v", path, err)
	}

	return string(data)
}
