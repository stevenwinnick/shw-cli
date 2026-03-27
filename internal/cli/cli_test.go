package cli

import (
	"bytes"
	"strings"
	"testing"
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
