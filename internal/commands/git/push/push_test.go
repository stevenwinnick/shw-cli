package push

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"shw-cli/internal/testutil"
)

func TestRunCallsGitPushWithUpstreamToHead(t *testing.T) {
	testutil.SkipIfWindows(t)

	binDir := t.TempDir()
	gitArgsLog := filepath.Join(t.TempDir(), "git-args.log")
	testutil.WriteExecutable(t, binDir, "git", fakeGitScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GIT_ARGS_FILE", gitArgsLog)

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	want := []string{"push", "-u", "origin", "HEAD"}
	got := testutil.ReadLines(t, gitArgsLog)
	if len(got) != len(want) {
		t.Fatalf("run args len mismatch: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("run args mismatch at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "push" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if cmd.Usage != "shw git push" {
		t.Fatalf("unexpected usage: %q", cmd.Usage)
	}
	if cmd.Run == nil {
		t.Fatal("expected Run handler")
	}
}

func fakeGitScript() string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return `#!/bin/sh
printf '%s\n' "$@" >"$FAKE_GIT_ARGS_FILE"
`
}
