package localandgithub

import (
	"errors"
	"testing"
)

type invocation struct {
	dir  string
	name string
	args []string
}

func TestRunCallsGitThenGitHubInTargetDir(t *testing.T) {
	origPrompt := promptRelativeDir
	origEnsure := ensureDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureDir = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/repo"
	var calls []invocation

	promptRelativeDir = func(_ string) (string, error) { return target, nil }
	ensureDir = func(dir string) error {
		if dir != target {
			t.Fatalf("ensureDir got %q, want %q", dir, target)
		}
		return nil
	}
	runCommandInDir = func(dir string, name string, args ...string) error {
		calls = append(calls, invocation{dir: dir, name: name, args: append([]string{}, args...)})
		return nil
	}

	if err := run([]string{"my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(calls))
	}

	if calls[0].dir != target || calls[0].name != "git" {
		t.Fatalf("first call mismatch: %#v", calls[0])
	}
	if len(calls[0].args) != 1 || calls[0].args[0] != "init" {
		t.Fatalf("first call args mismatch: %v", calls[0].args)
	}

	if calls[1].dir != target || calls[1].name != "gh" {
		t.Fatalf("second call mismatch: %#v", calls[1])
	}
	wantSecond := []string{"repo", "create", "my-repo", "--private"}
	if len(calls[1].args) != len(wantSecond) {
		t.Fatalf("second call args len mismatch: got %v want %v", calls[1].args, wantSecond)
	}
	for i := range wantSecond {
		if calls[1].args[i] != wantSecond[i] {
			t.Fatalf("second call arg mismatch at %d: got %q want %q", i, calls[1].args[i], wantSecond[i])
		}
	}
}

func TestRunStopsIfGitInitFails(t *testing.T) {
	origPrompt := promptRelativeDir
	origEnsure := ensureDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureDir = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("git init failed")
	const target = "/tmp/repo"
	var calls int

	promptRelativeDir = func(_ string) (string, error) { return target, nil }
	ensureDir = func(_ string) error { return nil }
	runCommandInDir = func(_ string, name string, _ ...string) error {
		calls++
		if name == "git" {
			return expectedErr
		}
		return nil
	}

	err := run([]string{"my-repo"})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected git init error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected only git command to run, got %d calls", calls)
	}
}
