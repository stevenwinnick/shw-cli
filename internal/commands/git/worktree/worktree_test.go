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

	for _, name := range []string{"create", "list", "remove", "prune-stale"} {
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
	if len(create.Flags) != 2 {
		t.Fatalf("unexpected create flags: %+v", create.Flags)
	}
	if create.Flags[0].Long != "repo-dir" || create.Flags[0].Short != "r" {
		t.Fatalf("missing repo-dir flag: %+v", create.Flags)
	}
	if create.Flags[1].Long != "no-update-default-branch" {
		t.Fatalf("missing no-update-default-branch flag: %+v", create.Flags)
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

func TestRunCreatePassesOptionFlags(t *testing.T) {
	originalCreateWorktree := createWorktree
	defer func() {
		createWorktree = originalCreateWorktree
	}()

	called := false
	createWorktree = func(repoDir string, branchName string, updateDefaultBranch bool) error {
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
		return nil
	}

	if err := runCreate([]string{"--repo-dir", "/tmp/repo", "--no-update-default-branch", "steven/feature"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}
	if !called {
		t.Fatal("expected createWorktree to be called")
	}
}
