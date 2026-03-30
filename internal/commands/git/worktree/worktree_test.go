package worktree

import (
	"strings"
	"testing"

	"shw-cli/internal/cli"
)

func TestCommandIncludesExpectedSubcommands(t *testing.T) {
	cmd := Command()

	got := map[string]bool{}
	for _, child := range cmd.Children {
		got[child.Name] = true
	}

	for _, name := range []string{"create", "list", "remove", "prune-stale", "path"} {
		if !got[name] {
			t.Fatalf("expected subcommand %q to be present", name)
		}
	}
}

func TestCreateCommandIncludesExpectedFlags(t *testing.T) {
	cmd := Command()

	var create *cli.Command
	for _, child := range cmd.Children {
		if child.Name == "create" {
			create = child
			break
		}
	}
	if create == nil {
		t.Fatal("expected create command")
	}
	if len(create.Flags) != 3 {
		t.Fatalf("unexpected create flags: %+v", create.Flags)
	}
	if create.Flags[0].Long != "repo-dir" || create.Flags[0].Short != "r" {
		t.Fatalf("missing repo-dir flag: %+v", create.Flags)
	}
	if create.Flags[1].Long != "no-update-default-branch" {
		t.Fatalf("missing no-update-default-branch flag: %+v", create.Flags)
	}
	if create.Flags[2].Long != "no-copy-navigation-command" {
		t.Fatalf("missing no-copy-navigation-command flag: %+v", create.Flags)
	}
}

func TestLeafHandlersRejectWrongArgCounts(t *testing.T) {
	tests := []struct {
		name    string
		run     func([]string) error
		args    []string
		wantMsg string
	}{
		{name: "create", run: runCreate, args: nil, wantMsg: "usage: shw git worktree create"},
		{name: "list", run: runList, args: []string{"extra"}, wantMsg: "usage: shw git worktree list"},
		{name: "remove", run: runRemove, args: nil, wantMsg: "usage: shw git worktree remove"},
		{name: "prune-stale", run: runPruneStale, args: []string{"extra"}, wantMsg: "usage: shw git worktree prune-stale"},
		{name: "path", run: runPath, args: nil, wantMsg: "usage: shw git worktree path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run(tt.args)
			if err == nil {
				t.Fatal("expected usage error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Fatalf("error got %q, want substring %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestPathCommandHelpUsesBranchNamePlaceholder(t *testing.T) {
	cmd := Command()

	var pathCmd *cli.Command
	for _, child := range cmd.Children {
		if child.Name == "path" {
			pathCmd = child
			break
		}
	}
	if pathCmd == nil {
		t.Fatal("expected path command")
	}
	if !strings.Contains(pathCmd.Usage, "<branch-name>") {
		t.Fatalf("usage got %q, want branch-name placeholder", pathCmd.Usage)
	}
	if !strings.Contains(pathCmd.Description, "`cd $(shw git worktree path <branch-name>)`") {
		t.Fatalf("description got %q, want cd example", pathCmd.Description)
	}
}

func TestRunCreatePassesOptionFlags(t *testing.T) {
	originalCreateWorktree := createWorktree
	defer func() {
		createWorktree = originalCreateWorktree
	}()

	called := false
	createWorktree = func(repoDir string, branchName string, updateDefaultBranch bool, copyNavigationCommand bool) error {
		called = true
		if repoDir != "/tmp/repo" {
			t.Fatalf("repoDir got %q, want %q", repoDir, "/tmp/repo")
		}
		if branchName != "steven/feature" {
			t.Fatalf("branchName got %q, want %q", branchName, "steven/feature")
		}
		if updateDefaultBranch {
			t.Fatal("expected no-update-default-branch to disable default-branch updates")
		}
		if copyNavigationCommand {
			t.Fatal("expected no-copy-navigation-command to disable clipboard copy")
		}
		return nil
	}

	if err := runCreate([]string{"--repo-dir", "/tmp/repo", "--no-update-default-branch", "--no-copy-navigation-command", "steven/feature"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}
	if !called {
		t.Fatal("expected createWorktree to be called")
	}
}
