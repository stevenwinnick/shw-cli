package local

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunUsesWorktreeSetupAndGitInit(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	gitDirLog := filepath.Join(t.TempDir(), "git-dir.log")
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR_FILE", gitDirLog)
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)

	if err := run([]string{"--initial-branch", "trunk", "--bare"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantDir := realPath(t, filepath.Join(rootDir, "repo", "trunk", "repo"))
	gotDir := realPath(t, strings.TrimSpace(string(mustReadFile(t, gitDirLog))))
	if gotDir != wantDir {
		t.Fatalf("git ran in %q, want %q", gotDir, wantDir)
	}

	wantArgs := []string{"init", "--initial-branch", "trunk", "--bare"}
	gotArgs := testutil.ReadLines(t, gitArgsLog)
	assertArgs(t, gotArgs, wantArgs)
}

func TestRunAllowsSkippingWorktreeSetup(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	gitDirLog := filepath.Join(t.TempDir(), "git-dir.log")
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR_FILE", gitDirLog)
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)

	if err := run([]string{"--no-worktree-setup", "--bare"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantDir := realPath(t, filepath.Join(rootDir, "repo"))
	gotDir := realPath(t, strings.TrimSpace(string(mustReadFile(t, gitDirLog))))
	if gotDir != wantDir {
		t.Fatalf("git ran in %q, want %q", gotDir, wantDir)
	}

	wantArgs := []string{"init", "--bare"}
	gotArgs := testutil.ReadLines(t, gitArgsLog)
	assertArgs(t, gotArgs, wantArgs)
}

func TestRunStopsOnPromptError(t *testing.T) {
	testutil.SetStdin(t, "")

	err := run(nil)
	if err == nil {
		t.Fatal("expected prompt error")
	}
	if !strings.Contains(err.Error(), "directory path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func fakeGitScript() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return `#!/bin/sh
if [ "$1" = "config" ] && [ "$2" = "--get" ] && [ "$3" = "init.defaultBranch" ]; then
  printf '%s\n' "${FAKE_GIT_DEFAULT_BRANCH:-trunk}"
  exit 0
fi
printf '%s\n' "$PWD" >"$FAKE_GIT_DIR_FILE"
printf '%s\n' "$@" >"$FAKE_GIT_ARGS_FILE"
if [ -n "${FAKE_GIT_FAIL_ON:-}" ] && [ "$1" = "$FAKE_GIT_FAIL_ON" ]; then
  exit 1
fi
`
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %q: %v", path, err)
	}

	return content
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

func assertArgs(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("arg length mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
