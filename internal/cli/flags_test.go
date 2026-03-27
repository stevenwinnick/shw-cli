package cli

import "testing"

func TestConsumeBoolFlagRemovesMatchingFlag(t *testing.T) {
	flag := Flag{Long: "no-worktree-setup"}

	consumed, filtered := ConsumeBoolFlag([]string{"--no-worktree-setup", "--bare"}, flag)
	if !consumed {
		t.Fatal("expected flag to be consumed")
	}
	if len(filtered) != 1 || filtered[0] != "--bare" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}

func TestConsumeBoolFlagStopsAtDoubleDash(t *testing.T) {
	flag := Flag{Long: "no-worktree-setup"}

	consumed, filtered := ConsumeBoolFlag([]string{"--", "--no-worktree-setup"}, flag)
	if consumed {
		t.Fatal("expected flag after -- to be left alone")
	}
	if len(filtered) != 2 || filtered[0] != "--" || filtered[1] != "--no-worktree-setup" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}
