package pr

import (
	"os"
	"path/filepath"
	"testing"

	"shw-cli/internal/testutil"
)

func TestMergeSquashesAndDeletesBranch(t *testing.T) {
	ghArgsLog := installFakeGh(t)

	if err := runMerge(nil); err != nil {
		t.Fatalf("runMerge returned error: %v", err)
	}

	assertGhArgs(t, ghArgsLog, "pr merge --squash --delete-branch")
}

func TestMergePassesBranchAndExtraArgsThrough(t *testing.T) {
	ghArgsLog := installFakeGh(t)

	if err := runMerge([]string{"steven/feature", "--admin"}); err != nil {
		t.Fatalf("runMerge returned error: %v", err)
	}

	assertGhArgs(t, ghArgsLog, "pr merge steven/feature --admin --squash --delete-branch")
}

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "pr" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if len(cmd.Children) != 1 || cmd.Children[0].Name != "merge" {
		t.Fatalf("expected a single merge subcommand, got %v", cmd.Children)
	}
	if cmd.Children[0].Run == nil {
		t.Fatal("expected merge Run handler")
	}
}

func installFakeGh(t *testing.T) string {
	t.Helper()
	testutil.SkipIfWindows(t)

	binDir := t.TempDir()
	ghArgsLog := filepath.Join(t.TempDir(), "gh-args.log")
	testutil.WriteExecutable(t, binDir, "gh", "#!/bin/sh\nprintf '%s\\n' \"$*\" >>\"$FAKE_GH_ARGS_FILE\"\n")

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH_ARGS_FILE", ghArgsLog)
	return ghArgsLog
}

func assertGhArgs(t *testing.T, ghArgsLog string, want string) {
	t.Helper()

	args := testutil.ReadLines(t, ghArgsLog)
	if len(args) != 1 {
		t.Fatalf("expected gh to be called once, got %d call(s): %v", len(args), args)
	}
	if args[0] != want {
		t.Fatalf("gh args got %q, want %q", args[0], want)
	}
}
