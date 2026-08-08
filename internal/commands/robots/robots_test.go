package robots

import (
	"strings"
	"testing"
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

func TestSudoArgs(t *testing.T) {
	t.Run("opens a login shell when no command is given", func(t *testing.T) {
		got := strings.Join(sudoArgs(nil), " ")
		want := "-u " + robotsUser + " -i"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("separates the command with a double dash", func(t *testing.T) {
		got := strings.Join(sudoArgs([]string{"git", "status", "-s"}), " ")
		want := "-u " + robotsUser + " -i -- git status -s"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

func TestRunStartRejectsExtraArguments(t *testing.T) {
	if err := runStart([]string{"unexpected"}); err == nil {
		t.Fatal("expected error when extra arguments are passed")
	}
}

func TestRunRunRequiresACommand(t *testing.T) {
	for _, args := range [][]string{nil, {"--"}} {
		if err := runRun(args); err == nil {
			t.Fatalf("expected error for args %v", args)
		}
	}
}
