package push

import "testing"

func TestRunCallsGitPushWithUpstreamToHead(t *testing.T) {
	origRun := runCommand
	defer func() {
		runCommand = origRun
	}()

	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) error {
		gotName = name
		gotArgs = append([]string{}, args...)
		return nil
	}

	if err := run(nil); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if gotName != "git" {
		t.Fatalf("run command got %q, want git", gotName)
	}

	want := []string{"push", "-u", "origin", "HEAD"}
	if len(gotArgs) != len(want) {
		t.Fatalf("run args len mismatch: got %v want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("run args mismatch at %d: got %q want %q", i, gotArgs[i], want[i])
		}
	}
}

func TestCommandShape(t *testing.T) {
	cmd := Command()

	if cmd.Name != "push" {
		t.Fatalf("unexpected name: %q", cmd.Name)
	}
	if cmd.Usage != "shw git push" {
		t.Fatalf("unexpected usage: %q", cmd.Usage)
	}
	if cmd.Run == nil {
		t.Fatal("expected Run handler")
	}
}
