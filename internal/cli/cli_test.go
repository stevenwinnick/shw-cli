package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"shw-cli/internal/utils"
)

func TestRootHelp(t *testing.T) {
	root := RootCommand(&Command{Name: "git", Summary: "Git commands"})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"-h"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	text := stdout.String()
	if !strings.Contains(text, "shw <command>") {
		t.Fatalf("help missing usage: %s", text)
	}
	if !strings.Contains(text, "git") {
		t.Fatalf("help missing command: %s", text)
	}
}

func TestNestedHelpIncludesLeaf(t *testing.T) {
	leaf := &Command{Name: "local", Summary: "local repo", Description: "leaf", Usage: "shw git repo create local"}
	create := &Command{Name: "create", Summary: "create", Description: "create", Usage: "shw git repo create"}
	create.AddChild(leaf)

	repo := &Command{Name: "repo", Summary: "repo", Description: "repo", Usage: "shw git repo"}
	repo.AddChild(create)

	git := &Command{Name: "git", Summary: "git", Description: "git", Usage: "shw git"}
	git.AddChild(repo)

	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	root.AddChild(git)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"git", "repo", "create", "-h"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	text := stdout.String()
	if !strings.Contains(text, "local") {
		t.Fatalf("help missing nested leaf command: %s", text)
	}
}

func TestUnknownSubcommandReturnsError(t *testing.T) {
	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"unknown"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown subcommand") {
		t.Fatalf("stderr missing unknown subcommand error: %s", stderr.String())
	}
}

func TestLeafReceivesArgs(t *testing.T) {
	var got []string
	leaf := &Command{
		Name:        "local",
		Description: "leaf",
		Usage:       "shw local",
		Run: func(args []string) error {
			got = append(got, args...)
			return nil
		},
	}
	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	root.AddChild(leaf)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"local", "--bare"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if len(got) != 1 || got[0] != "--bare" {
		t.Fatalf("leaf did not receive passthrough args: %v", got)
	}
}

func TestRunPropagatesCommandExitCode(t *testing.T) {
	leaf := &Command{
		Name:        "fail",
		Description: "leaf",
		Usage:       "shw fail",
		Run: func(_ []string) error {
			return utils.RunCommandInDirWithWriters("", io.Discard, io.Discard, "sh", "-c", "exit 3")
		},
	}
	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	root.AddChild(leaf)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"fail"}, stdout, stderr)
	if code != 3 {
		t.Fatalf("expected exit code 3, got %d", code)
	}
	if stderr.String() != "" {
		t.Fatalf("expected no wrapper error for a failed command, got: %s", stderr.String())
	}
}

func TestRunReportsErrorsFromCapturedCommands(t *testing.T) {
	leaf := &Command{
		Name:        "fail",
		Description: "leaf",
		Usage:       "shw fail",
		Run: func(_ []string) error {
			_, err := utils.CaptureCommand("sh", "-c", "echo boom >&2; exit 3")
			return err
		},
	}
	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	root.AddChild(leaf)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"fail"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Fatalf("expected captured output in error, got: %s", stderr.String())
	}
}

func TestHelpIncludesLocalFlags(t *testing.T) {
	leaf := &Command{
		Name:        "local",
		Description: "leaf",
		Usage:       "shw local [flags] [args...]",
		Flags: []Flag{
			{Long: "no-worktree-setup", Description: "Skip the default worktree setup and create the repo directly in the prompted directory"},
		},
	}
	root := &Command{Name: "shw", Description: "root", Usage: "shw <command>"}
	root.AddChild(leaf)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(root, []string{"local", "-h"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	text := stdout.String()
	if !strings.Contains(text, "shw local [flags] [args...]") {
		t.Fatalf("help missing usage: %s", text)
	}
	if !strings.Contains(text, "--no-worktree-setup") {
		t.Fatalf("help missing local flag: %s", text)
	}
	if !strings.Contains(text, "-h, --help") {
		t.Fatalf("help missing help flag: %s", text)
	}
}
