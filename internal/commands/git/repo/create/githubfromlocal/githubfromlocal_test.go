package githubfromlocal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunCallsGitHubCreateInSelectedDir(t *testing.T) {
	testutil.SkipIfWindows(t)

	rootDir := t.TempDir()
	targetDir := filepath.Join(rootDir, "existing-repo")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "existing-repo\n")

	binDir := t.TempDir()
	ghDirLog := filepath.Join(t.TempDir(), "gh-dir.log")
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "gh", fakeGhScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_DIR_FILE", ghDirLog)
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)

	if err := run([]string{"my-repo", "--public"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if got := realPath(t, strings.TrimSpace(string(mustReadFile(t, ghDirLog)))); got != realPath(t, targetDir) {
		t.Fatalf("gh ran in %q, want %q", got, targetDir)
	}
	assertArgs(t, testutil.ReadLines(t, ghArgsLog), []string{"repo", "create", "my-repo", "--public"})
}

func TestRunStopsIfDirectoryIsMissing(t *testing.T) {
	rootDir := t.TempDir()
	testutil.SetWorkingDir(t, rootDir)
	testutil.SetStdin(t, "repo\n")

	err := run(nil)
	if err == nil {
		t.Fatal("expected missing directory error")
	}
	if !strings.Contains(err.Error(), "directory does not exist") {
		t.Fatalf("unexpected error: %v", err)
	}
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
