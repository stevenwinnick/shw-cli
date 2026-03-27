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
	origPrompt := promptRelativeRepoTargetPaths
	origEnsure := ensureRepoTargetPaths
	origRun := runCommandInDir
	defer func() {
		promptRelativeRepoTargetPaths = origPrompt
		ensureRepoTargetPaths = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/repo"
	targetPaths := utils.RepoTargetPaths{
		WorkingDir:        target,
		AdditionalWorkDir: "/tmp/worktrees",
		UsesWorktrees:     true,
	}
	var calls []invocation

	promptRelativeRepoTargetPaths = func(_ string, gitInitArgs []string, useWorktrees bool) (utils.RepoTargetPaths, error) {
		if !useWorktrees {
			t.Fatal("expected worktree layout to remain enabled")
		}
		if gitInitArgs != nil {
			t.Fatalf("expected nil git init args, got %v", gitInitArgs)
		}
		return targetPaths, nil
	}
	ensureRepoTargetPaths = func(paths utils.RepoTargetPaths) error {
		if paths != targetPaths {
			t.Fatalf("ensureRepoTargetPaths got %#v, want %#v", paths, targetPaths)
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

func TestRunAllowsSkippingWorktreeLayout(t *testing.T) {
	origPrompt := promptRelativeRepoTargetPaths
	origEnsure := ensureRepoTargetPaths
	origRun := runCommandInDir
	defer func() {
		promptRelativeRepoTargetPaths = origPrompt
		ensureRepoTargetPaths = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/repo"
	targetPaths := utils.RepoTargetPaths{
		WorkingDir:    target,
		UsesWorktrees: false,
	}
	var calls []invocation

	promptRelativeRepoTargetPaths = func(_ string, gitInitArgs []string, useWorktrees bool) (utils.RepoTargetPaths, error) {
		if useWorktrees {
			t.Fatal("expected --no-worktree-setup to disable the worktree layout")
		}
		if gitInitArgs != nil {
			t.Fatalf("expected nil git init args, got %v", gitInitArgs)
		}
		return targetPaths, nil
	}
	ensureRepoTargetPaths = func(paths utils.RepoTargetPaths) error {
		if paths != targetPaths {
			t.Fatalf("ensureRepoTargetPaths got %#v, want %#v", paths, targetPaths)
		}
		return nil
	}
	runCommandInDir = func(dir string, name string, args ...string) error {
		calls = append(calls, invocation{dir: dir, name: name, args: append([]string{}, args...)})
		return nil
	}

	if err := run([]string{"--no-worktree-setup", "my-repo", "--private"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(calls))
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
	origPrompt := promptRelativeRepoTargetPaths
	origEnsure := ensureRepoTargetPaths
	origRun := runCommandInDir
	defer func() {
		promptRelativeRepoTargetPaths = origPrompt
		ensureRepoTargetPaths = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("git init failed")
	const target = "/tmp/repo"
	targetPaths := utils.RepoTargetPaths{
		WorkingDir:        target,
		AdditionalWorkDir: "/tmp/worktrees",
		UsesWorktrees:     true,
	}
	var calls int

	promptRelativeRepoTargetPaths = func(_ string, _ []string, _ bool) (utils.RepoTargetPaths, error) {
		return targetPaths, nil
	}
	ensureRepoTargetPaths = func(_ utils.RepoTargetPaths) error { return nil }
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
