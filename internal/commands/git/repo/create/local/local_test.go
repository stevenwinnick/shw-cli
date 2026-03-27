package local

import (
	"errors"
	"fmt"
	"shw-cli/internal/utils"
	"testing"
)

func TestRunUsesPromptEnsureAndGitInit(t *testing.T) {
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
	var gotPrompt string
	var gotEnsure utils.RepoTargetPaths
	var gotDir string
	var gotName string
	var gotArgs []string

	promptRelativeRepoTargetPaths = func(prompt string, gitInitArgs []string, useWorktrees bool) (utils.RepoTargetPaths, error) {
		gotPrompt = prompt
		if !useWorktrees {
			t.Fatal("expected worktree layout to remain enabled")
		}
		if len(gitInitArgs) != 1 || gitInitArgs[0] != "--bare" {
			t.Fatalf("git init args mismatch: %v", gitInitArgs)
		}
		return targetPaths, nil
	}
	ensureRepoTargetPaths = func(paths utils.RepoTargetPaths) error {
		gotEnsure = paths
		return nil
	}
	runCommandInDir = func(dir string, name string, args ...string) error {
		gotDir = dir
		gotName = name
		gotArgs = append([]string{}, args...)
		return nil
	}

	if err := run([]string{"--bare"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if gotPrompt == "" {
		t.Fatal("expected prompt to be shown")
	}
	if gotEnsure != targetPaths {
		t.Fatalf("ensureRepoTargetPaths got %#v, want %#v", gotEnsure, targetPaths)
	}
	if gotDir != target {
		t.Fatalf("run dir got %q, want %q", gotDir, target)
	}
	if gotName != "git" {
		t.Fatalf("run command got %q, want git", gotName)
	}
	wantArgs := []string{"init", "--bare"}
	if len(gotArgs) != len(wantArgs) {
		t.Fatalf("run args len %d, want %d: %v", len(gotArgs), len(wantArgs), gotArgs)
	}
	for i := range wantArgs {
		if gotArgs[i] != wantArgs[i] {
			t.Fatalf("run args mismatch at %d: got %q want %q", i, gotArgs[i], wantArgs[i])
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

	promptRelativeRepoTargetPaths = func(_ string, gitInitArgs []string, useWorktrees bool) (utils.RepoTargetPaths, error) {
		if useWorktrees {
			t.Fatal("expected --no-worktrees to disable the worktree layout")
		}
		if len(gitInitArgs) != 1 || gitInitArgs[0] != "--bare" {
			t.Fatalf("git init args mismatch: %v", gitInitArgs)
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
		if dir != target {
			t.Fatalf("run dir got %q, want %q", dir, target)
		}
		if name != "git" {
			t.Fatalf("run command got %q, want git", name)
		}
		wantArgs := []string{"init", "--bare"}
		if len(args) != len(wantArgs) {
			t.Fatalf("run args len %d, want %d: %v", len(args), len(wantArgs), args)
		}
		for i := range wantArgs {
			if args[i] != wantArgs[i] {
				t.Fatalf("run args mismatch at %d: got %q want %q", i, args[i], wantArgs[i])
			}
		}
		return nil
	}

	if err := run([]string{"--no-worktrees", "--bare"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestRunStopsOnPromptError(t *testing.T) {
	origPrompt := promptRelativeRepoTargetPaths
	origEnsure := ensureRepoTargetPaths
	origRun := runCommandInDir
	defer func() {
		promptRelativeRepoTargetPaths = origPrompt
		ensureRepoTargetPaths = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("prompt failed")
	promptRelativeRepoTargetPaths = func(_ string, _ []string, _ bool) (utils.RepoTargetPaths, error) {
		return utils.RepoTargetPaths{}, expectedErr
	}
	ensureRepoTargetPaths = func(_ utils.RepoTargetPaths) error {
		return fmt.Errorf("should not be called")
	}
	runCommandInDir = func(_ string, _ string, _ ...string) error {
		return fmt.Errorf("should not be called")
	}

	err := run(nil)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected prompt error, got %v", err)
	}
}
