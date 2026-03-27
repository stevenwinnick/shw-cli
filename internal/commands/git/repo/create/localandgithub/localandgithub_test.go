package localandgithub

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunCallsGitThenGitHubInTargetDir(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	gitDirLog := filepath.Join(t.TempDir(), "git-dir.log")
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR_FILE", gitDirLog)
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)

	if err := run([]string{"my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantDir := realPath(t, filepath.Join(rootDir, "repo", "trunk", "repo"))
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, gitDirLog)))); got != wantDir {
		t.Fatalf("git ran in %q, want %q", got, wantDir)
	}
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, ghDirLog)))); got != wantDir {
		t.Fatalf("gh ran in %q, want %q", got, wantDir)
	}

	assertArgs(t, testutil.ReadLines(t, gitArgsLog), []string{"init"})
	assertArgs(t, testutil.ReadLines(t, ghArgsLog), []string{"repo", "create", "my-repo", "--private"})
}

func TestRunAllowsSkippingWorktreeSetup(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	gitDirLog := filepath.Join(t.TempDir(), "git-dir.log")
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR_FILE", gitDirLog)
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)

	if err := run([]string{"--no-worktree-setup", "my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantDir := realPath(t, filepath.Join(rootDir, "repo"))
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, gitDirLog)))); got != wantDir {
		t.Fatalf("git ran in %q, want %q", got, wantDir)
	}
	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, ghDirLog)))); got != wantDir {
		t.Fatalf("gh ran in %q, want %q", got, wantDir)
	}

	assertArgs(t, testutil.ReadLines(t, gitArgsLog), []string{"init"})
	assertArgs(t, testutil.ReadLines(t, ghArgsLog), []string{"repo", "create", "my-repo", "--private"})
}

func TestRunStopsIfGitInitFails(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	binDir := t.TempDir()
	gitDirLog := filepath.Join(t.TempDir(), "git-dir.log")
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_DIR_FILE", gitDirLog)
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GIT_FAIL_ON", "init")

	err := run([]string{"my-repo"})
	if err == nil {
		t.Fatal("expected git init error")
	}
	if !strings.Contains(err.Error(), "command failed: git init") {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(ghDirLog); !os.IsNotExist(statErr) {
		t.Fatalf("expected gh not to run, stat error: %v", statErr)
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

func fakeGhScript() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return `#!/bin/sh
printf '%s\n' "$PWD" >"$FAKE_GH_DIR_FILE"
printf '%s\n' "$@" >"$FAKE_GH_ARGS_FILE"
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
