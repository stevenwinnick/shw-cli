package localandgithub

import (
	"errors"
	"shw-cli/internal/utils"
	"testing"
)

type invocation struct {
	dir  string
	name string
	args []string
}

func TestRunCallsGitThenGitHubInTargetDir(t *testing.T) {
	origPrompt := promptRelativeNewRepoPaths
	origEnsure := ensureNewRepoLayout
	origRun := runCommandInDir
	defer func() {
		promptRelativeNewRepoPaths = origPrompt
		ensureNewRepoLayout = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/repo"
	targetPaths := utils.NewRepoPaths{
		MainWorktreeDir: target,
		WorktreesDir:    "/tmp/worktrees",
	}
	var calls []invocation

	promptRelativeNewRepoPaths = func(_ string, gitInitArgs []string) (utils.NewRepoPaths, error) {
		if gitInitArgs != nil {
			t.Fatalf("expected nil git init args, got %v", gitInitArgs)
		}
		return targetPaths, nil
	}
	ensureNewRepoLayout = func(paths utils.NewRepoPaths) error {
		if paths != targetPaths {
			t.Fatalf("ensureNewRepoLayout got %#v, want %#v", paths, targetPaths)
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
	origPrompt := promptRelativeNewRepoPaths
	origEnsure := ensureNewRepoLayout
	origRun := runCommandInDir
	defer func() {
		promptRelativeNewRepoPaths = origPrompt
		ensureNewRepoLayout = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("git init failed")
	const target = "/tmp/repo"
	targetPaths := utils.NewRepoPaths{
		MainWorktreeDir: target,
		WorktreesDir:    "/tmp/worktrees",
	}
	var calls int

	promptRelativeNewRepoPaths = func(_ string, _ []string) (utils.NewRepoPaths, error) {
		return targetPaths, nil
	}
	ensureNewRepoLayout = func(_ utils.NewRepoPaths) error { return nil }
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
