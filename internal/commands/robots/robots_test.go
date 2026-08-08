package robots

import "testing"

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "robots" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if len(cmd.Children) != 1 {
		t.Fatalf("unexpected child count: %d", len(cmd.Children))
	}

	start := cmd.Children[0]
	if start.Name != "start" {
		t.Fatalf("unexpected child name: %q", start.Name)
	}
	if start.Usage != "shw robots start" {
		t.Fatalf("unexpected usage: %q", start.Usage)
	}
	if start.Run == nil {
		t.Fatal("expected Run handler")
	}
}

func TestRunStartRejectsExtraArguments(t *testing.T) {
	if err := runStart([]string{"unexpected"}); err == nil {
		t.Fatal("expected error when extra arguments are passed")
	}
}
