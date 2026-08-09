package robots

import (
	"os"
	"path/filepath"
	"testing"

	"shw-cli/internal/testutil"
)

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "robots" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}

	usages := map[string]string{}
	for _, child := range cmd.Children {
		if child.Run == nil {
			t.Fatalf("expected Run handler for %q", child.Name)
		}
		usages[child.Name] = child.Usage
	}

	want := map[string]string{
		"start": "shw robots start",
		"run":   "shw robots run -- <command>",
	}
	for name, usage := range want {
		if usages[name] != usage {
			t.Fatalf("usage for %q got %q, want %q", name, usages[name], usage)
		}
	}
	if len(usages) != len(want) {
		t.Fatalf("unexpected children: %v", usages)
	}
}

func TestRunStartOpensALoginShellAsTheRobotsUser(t *testing.T) {
	argsLog := installFakeSudo(t)

	if err := runStart(nil); err != nil {
		t.Fatalf("runStart returned error: %v", err)
	}

	assertSudoArgs(t, argsLog, "-u "+robotsUser+" -i")
}

func TestRunStartRejectsExtraArguments(t *testing.T) {
	if err := runStart([]string{"unexpected"}); err == nil {
		t.Fatal("expected error when extra arguments are passed")
	}
}

func TestRunRunPassesTheCommandToSudo(t *testing.T) {
	t.Run("with a leading double dash", func(t *testing.T) {
		argsLog := installFakeSudo(t)

		if err := runRun([]string{"--", "git", "status", "-s"}); err != nil {
			t.Fatalf("runRun returned error: %v", err)
		}

		assertSudoArgs(t, argsLog, "-u "+robotsUser+" -i -- git status -s")
	})

	t.Run("without a leading double dash", func(t *testing.T) {
		argsLog := installFakeSudo(t)

		if err := runRun([]string{"git", "status", "-s"}); err != nil {
			t.Fatalf("runRun returned error: %v", err)
		}

		assertSudoArgs(t, argsLog, "-u "+robotsUser+" -i -- git status -s")
	})
}

func TestRunRunRequiresACommand(t *testing.T) {
	for _, args := range [][]string{nil, {"--"}} {
		if err := runRun(args); err == nil {
			t.Fatalf("expected error for args %v", args)
		}
	}
}

// installFakeSudo puts a fake sudo on PATH that logs its arguments, and returns the log path
func installFakeSudo(t *testing.T) string {
	t.Helper()
	testutil.SkipIfWindows(t)

	binDir := t.TempDir()
	argsLog := filepath.Join(t.TempDir(), "sudo-args.log")
	testutil.WriteExecutable(t, binDir, "sudo", "#!/bin/bash\nprintf '%s\\n' \"$*\" >>\"$FAKE_SUDO_ARGS_FILE\"\n")

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_SUDO_ARGS_FILE", argsLog)
	return argsLog
}

func assertSudoArgs(t *testing.T, argsLog string, want string) {
	t.Helper()

	lines := testutil.ReadLines(t, argsLog)
	if len(lines) != 1 {
		t.Fatalf("expected one sudo invocation, got %v", lines)
	}
	if lines[0] != want {
		t.Fatalf("sudo args got %q, want %q", lines[0], want)
	}
}
