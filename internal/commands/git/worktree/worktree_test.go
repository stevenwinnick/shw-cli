package worktree

import (
	"strings"
	"testing"
)

func TestCommandIncludesExpectedSubcommands(t *testing.T) {
	cmd := Command()

	got := map[string]bool{}
	for _, child := range cmd.Children {
		got[child.Name] = true
	}

	for _, name := range []string{"create", "list", "remove", "clean-all", "switch"} {
		if !got[name] {
			t.Fatalf("expected subcommand %q to be present", name)
		}
	}
}

func TestLeafHandlersRejectWrongArgCounts(t *testing.T) {
	tests := []struct {
		name    string
		run     func([]string) error
		args    []string
		wantMsg string
	}{
		{name: "create", run: runCreate, args: []string{"repo"}, wantMsg: "usage: shw git worktree create"},
		{name: "list", run: runList, args: nil, wantMsg: "usage: shw git worktree list"},
		{name: "remove", run: runRemove, args: []string{"repo"}, wantMsg: "usage: shw git worktree remove"},
		{name: "clean-all", run: runCleanAll, args: []string{"repo", "extra"}, wantMsg: "usage: shw git worktree clean-all"},
		{name: "switch", run: runSwitch, args: []string{"repo"}, wantMsg: "usage: shw git worktree switch"},
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
