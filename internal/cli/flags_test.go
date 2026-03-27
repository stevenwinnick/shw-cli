package cli

import "testing"

func TestConsumeBoolFlagRemovesMatchingFlag(t *testing.T) {
	flag := Flag{Long: "no-worktrees"}

	consumed, filtered := ConsumeBoolFlag([]string{"--no-worktrees", "--bare"}, flag)
	if !consumed {
		t.Fatal("expected flag to be consumed")
	}
	if len(filtered) != 1 || filtered[0] != "--bare" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}

func TestConsumeBoolFlagStopsAtDoubleDash(t *testing.T) {
	flag := Flag{Long: "no-worktrees"}

	consumed, filtered := ConsumeBoolFlag([]string{"--", "--no-worktrees"}, flag)
	if consumed {
		t.Fatal("expected flag after -- to be left alone")
	}
	if len(filtered) != 2 || filtered[0] != "--" || filtered[1] != "--no-worktrees" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}
