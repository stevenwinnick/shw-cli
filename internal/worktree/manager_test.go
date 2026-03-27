package worktree

import (
	"bytes"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateUsesPreferredLayout(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	var mkdirPath string
	mkdirAll = func(path string, _ fs.FileMode) error {
		mkdirPath = path
		return nil
	}

	var commandDir string
	var commandName string
	var commandArgs []string
	runCommandInDir = func(dir string, name string, args ...string) error {
		commandDir = dir
		commandName = name
		commandArgs = append([]string{}, args...)
		return nil
	}

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		switch key(append([]string{name}, args...)...) {
		case key("git", "rev-parse", "--show-toplevel"):
			return "/code/demo/trunk/demo\n", nil
		case key("git", "worktree", "list", "--porcelain"):
			return "worktree /code/demo/trunk/demo\nbranch refs/heads/trunk\n", nil
		case key("git", "show-ref", "--verify", "refs/remotes/origin/trunk"):
			return "sha refs/remotes/origin/trunk\n", nil
		default:
			return "", errors.New("unexpected command")
		}
	}

	var output bytes.Buffer
	stdout = &output

	if err := Create("/code/demo/trunk/demo", "steven/add-worktree-commands"); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	wantPath := "/code/demo/worktrees/steven--add-worktree-commands/demo"
	if mkdirPath != filepath.Dir(wantPath) {
		t.Fatalf("mkdir path = %q, want %q", mkdirPath, filepath.Dir(wantPath))
	}
	if commandDir != "/code/demo/trunk/demo" || commandName != "git" {
		t.Fatalf("unexpected command invocation: dir=%q name=%q", commandDir, commandName)
	}

	wantArgs := []string{"worktree", "add", "-b", "steven/add-worktree-commands", wantPath, "origin/trunk"}
	assertArgsEqual(t, commandArgs, wantArgs)
	if !strings.Contains(output.String(), "Worktree created: "+wantPath) {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestListPrintsStatusesAndMarksMissingPaths(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	var ranList bool
	runCommandInDir = func(dir string, name string, args ...string) error {
		if dir == "/code/demo/trunk/demo" && name == "git" && key(args...) == key("worktree", "list") {
			ranList = true
			return nil
		}
		return errors.New("unexpected command")
	}

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		switch {
		case dir == "/code/demo/trunk/demo" && key(append([]string{name}, args...)...) == key("git", "rev-parse", "--show-toplevel"):
			return "/code/demo/trunk/demo\n", nil
		case dir == "/code/demo/trunk/demo" && key(append([]string{name}, args...)...) == key("git", "worktree", "list", "--porcelain"):
			return "worktree /code/demo/trunk/demo\nbranch refs/heads/trunk\n\nworktree /code/demo/worktrees/steven--feature/demo\nbranch refs/heads/steven/feature\n", nil
		case dir == "/code/demo/trunk/demo" && key(append([]string{name}, args...)...) == key("git", "status", "--short"):
			return " M README.md\n", nil
		case dir == "/code/demo/worktrees/steven--feature/demo" && key(append([]string{name}, args...)...) == key("git", "status", "--short"):
			return "", nil
		default:
			return "", errors.New("unexpected command")
		}
	}

	stat = func(path string) (fs.FileInfo, error) {
		if path == "/code/demo/trunk/demo" {
			return fakeFileInfo{name: "demo", dir: true}, nil
		}
		return nil, errors.New("missing")
	}

	var output bytes.Buffer
	stdout = &output

	if err := List("/code/demo/trunk/demo"); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if !ranList {
		t.Fatal("expected git worktree list to run")
	}

	text := output.String()
	if !strings.Contains(text, "/code/demo/trunk/demo:\n M README.md\n") {
		t.Fatalf("missing main worktree status in output: %q", text)
	}
	if !strings.Contains(text, "/code/demo/worktrees/steven--feature/demo:\n(not accessible)\n") {
		t.Fatalf("missing inaccessible worktree marker in output: %q", text)
	}
}

func TestRemoveDeletesWorktreeDirectoryAndBranch(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	var commands [][]string
	runCommandInDir = func(dir string, name string, args ...string) error {
		commands = append(commands, append([]string{dir, name}, args...))
		return nil
	}

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		switch key(append([]string{name}, args...)...) {
		case key("git", "rev-parse", "--show-toplevel"):
			return "/code/demo/trunk/demo\n", nil
		case key("git", "worktree", "list", "--porcelain"):
			return "worktree /code/demo/trunk/demo\nbranch refs/heads/trunk\n\nworktree /code/demo/worktrees/steven--feature/demo\nbranch refs/heads/steven/feature\n", nil
		case key("git", "show-ref", "--verify", "refs/heads/steven/feature"):
			return "sha refs/heads/steven/feature\n", nil
		default:
			return "", errors.New("unexpected command")
		}
	}

	var removed string
	removeAll = func(path string) error {
		removed = path
		return nil
	}
	stat = func(path string) (fs.FileInfo, error) {
		return fakeFileInfo{name: filepath.Base(path), dir: true}, nil
	}

	var output bytes.Buffer
	stdout = &output

	if err := Remove("/code/demo/trunk/demo", "steven/feature"); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}

	if removed != "/code/demo/worktrees/steven--feature" {
		t.Fatalf("removed path = %q, want parent worktree dir", removed)
	}

	if len(commands) != 2 {
		t.Fatalf("unexpected command count: %v", commands)
	}
	assertArgsEqual(t, commands[0], []string{"/code/demo/trunk/demo", "git", "worktree", "remove", "--force", "/code/demo/worktrees/steven--feature/demo"})
	assertArgsEqual(t, commands[1], []string{"/code/demo/trunk/demo", "git", "branch", "-D", "steven/feature"})
	if !strings.Contains(output.String(), "Worktree removed successfully") {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestCleanAllRemovesBranchesMissingRemoteTrackingRefs(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	var commands [][]string
	runCommandInDir = func(dir string, name string, args ...string) error {
		commands = append(commands, append([]string{dir, name}, args...))
		return nil
	}

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		switch key(append([]string{name}, args...)...) {
		case key("git", "rev-parse", "--show-toplevel"):
			return "/code/demo/trunk/demo\n", nil
		case key("git", "worktree", "list", "--porcelain"):
			return "worktree /code/demo/trunk/demo\nbranch refs/heads/trunk\n\nworktree /code/demo/worktrees/steven--stale/demo\nbranch refs/heads/steven/stale\n\nworktree /code/demo/worktrees/steven--kept/demo\nbranch refs/heads/steven/kept\n", nil
		case key("git", "show-ref", "--verify", "refs/remotes/origin/steven/stale"):
			return "", errors.New("missing")
		case key("git", "show-ref", "--verify", "refs/remotes/origin/steven/kept"):
			return "sha refs/remotes/origin/steven/kept\n", nil
		default:
			return "", errors.New("unexpected command")
		}
	}

	var removed []string
	removeAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	var output bytes.Buffer
	stdout = &output

	if err := CleanAll("/code/demo/trunk/demo"); err != nil {
		t.Fatalf("CleanAll returned error: %v", err)
	}

	if len(commands) != 3 {
		t.Fatalf("unexpected command count: %v", commands)
	}
	assertArgsEqual(t, commands[0], []string{"/code/demo/trunk/demo", "git", "worktree", "prune"})
	assertArgsEqual(t, commands[1], []string{"/code/demo/trunk/demo", "git", "worktree", "remove", "--force", "/code/demo/worktrees/steven--stale/demo"})
	assertArgsEqual(t, commands[2], []string{"/code/demo/trunk/demo", "git", "branch", "-D", "steven/stale"})
	if len(removed) != 1 || removed[0] != "/code/demo/worktrees/steven--stale" {
		t.Fatalf("unexpected removed directories: %v", removed)
	}
	if !strings.Contains(output.String(), "Worktree cleanup complete") {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestSwitchPrintsResolvedPath(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		if key(append([]string{name}, args...)...) == key("git", "rev-parse", "--show-toplevel") {
			return "/code/demo/trunk/demo\n", nil
		}
		return "", errors.New("unexpected command")
	}
	stat = func(path string) (fs.FileInfo, error) {
		return fakeFileInfo{name: filepath.Base(path), dir: true}, nil
	}

	var output bytes.Buffer
	stdout = &output

	if err := Switch("/code/demo/trunk/demo", "steven--feature"); err != nil {
		t.Fatalf("Switch returned error: %v", err)
	}

	if output.String() != "/code/demo/worktrees/steven--feature/demo\n" {
		t.Fatalf("unexpected stdout: %q", output.String())
	}
}

func TestSwitchListsAvailableNamesWhenMissing(t *testing.T) {
	restore := stubEnvironment(t)
	defer restore()

	captureCommandInDir = func(dir string, name string, args ...string) (string, error) {
		if key(append([]string{name}, args...)...) == key("git", "rev-parse", "--show-toplevel") {
			return "/code/demo/trunk/demo\n", nil
		}
		return "", errors.New("unexpected command")
	}
	stat = func(path string) (fs.FileInfo, error) {
		return nil, errors.New("missing")
	}
	readDir = func(path string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{
			fakeDirEntry{name: "steven--feature", dir: true},
			fakeDirEntry{name: "steven--bugfix", dir: true},
		}, nil
	}

	err := Switch("/code/demo/trunk/demo", "missing")
	if err == nil {
		t.Fatal("expected error for missing worktree")
	}
	if !strings.Contains(err.Error(), "steven--feature, steven--bugfix") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func stubEnvironment(t *testing.T) func() {
	t.Helper()

	origRun := runCommandInDir
	origCapture := captureCommandInDir
	origStdout := stdout
	origStat := stat
	origMkdirAll := mkdirAll
	origRemoveAll := removeAll
	origReadDir := readDir
	origGetenv := getenv

	runCommandInDir = func(string, string, ...string) error { return nil }
	captureCommandInDir = func(string, string, ...string) (string, error) { return "", nil }
	stdout = &bytes.Buffer{}
	stat = func(path string) (fs.FileInfo, error) { return fakeFileInfo{name: filepath.Base(path), dir: true}, nil }
	mkdirAll = func(string, fs.FileMode) error { return nil }
	removeAll = func(string) error { return nil }
	readDir = func(string) ([]fs.DirEntry, error) { return nil, nil }
	getenv = func(string) string { return "" }

	return func() {
		runCommandInDir = origRun
		captureCommandInDir = origCapture
		stdout = origStdout
		stat = origStat
		mkdirAll = origMkdirAll
		removeAll = origRemoveAll
		readDir = origReadDir
		getenv = origGetenv
	}
}

func assertArgsEqual(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("arg len mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func key(parts ...string) string {
	return strings.Join(parts, "\x00")
}

type fakeFileInfo struct {
	name string
	dir  bool
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() fs.FileMode  { return 0o755 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }

type fakeDirEntry struct {
	name string
	dir  bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.dir }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return fakeFileInfo{name: f.name, dir: f.dir}, nil }
