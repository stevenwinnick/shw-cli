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

func TestConsumeStringFlagConsumesSeparateValue(t *testing.T) {
	flag := Flag{Long: "repo-dir", Short: "r", ValueName: "<repo-dir>"}

	value, filtered, err := ConsumeStringFlag([]string{"--repo-dir", "/tmp/repo", "branch"}, flag)
	if err != nil {
		t.Fatalf("ConsumeStringFlag returned error: %v", err)
	}
	if value != "/tmp/repo" {
		t.Fatalf("value got %q, want %q", value, "/tmp/repo")
	}
	if len(filtered) != 1 || filtered[0] != "branch" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}

func TestConsumeStringFlagConsumesEqualsValue(t *testing.T) {
	flag := Flag{Long: "repo-dir", Short: "r", ValueName: "<repo-dir>"}

	value, filtered, err := ConsumeStringFlag([]string{"-r=/tmp/repo", "branch"}, flag)
	if err != nil {
		t.Fatalf("ConsumeStringFlag returned error: %v", err)
	}
	if value != "/tmp/repo" {
		t.Fatalf("value got %q, want %q", value, "/tmp/repo")
	}
	if len(filtered) != 1 || filtered[0] != "branch" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}

func TestConsumeStringFlagStopsAtDoubleDash(t *testing.T) {
	flag := Flag{Long: "repo-dir", ValueName: "<repo-dir>"}

	value, filtered, err := ConsumeStringFlag([]string{"--", "--repo-dir", "/tmp/repo"}, flag)
	if err != nil {
		t.Fatalf("ConsumeStringFlag returned error: %v", err)
	}
	if value != "" {
		t.Fatalf("value got %q, want empty", value)
	}
	if len(filtered) != 3 || filtered[0] != "--" || filtered[1] != "--repo-dir" || filtered[2] != "/tmp/repo" {
		t.Fatalf("filtered args mismatch: %v", filtered)
	}
}

func TestConsumeStringFlagRejectsMissingValue(t *testing.T) {
	flag := Flag{Long: "repo-dir", ValueName: "<repo-dir>"}

	_, _, err := ConsumeStringFlag([]string{"--repo-dir"}, flag)
	if err == nil {
		t.Fatal("expected missing value error")
	}
	if err.Error() != "missing value for --repo-dir <repo-dir>; expected <repo-dir>" {
		t.Fatalf("unexpected error: %v", err)
	}
}
