package local

import (
	"errors"
	"fmt"
	"shw-cli/internal/utils"
	"testing"
)

func TestRunUsesPromptEnsureAndGitInit(t *testing.T) {
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
	var gotPrompt string
	var gotEnsure utils.NewRepoPaths
	var gotDir string
	var gotName string
	var gotArgs []string

	promptRelativeNewRepoPaths = func(prompt string, gitInitArgs []string) (utils.NewRepoPaths, error) {
		gotPrompt = prompt
		if len(gitInitArgs) != 1 || gitInitArgs[0] != "--bare" {
			t.Fatalf("git init args mismatch: %v", gitInitArgs)
		}
		return targetPaths, nil
	}
	ensureNewRepoLayout = func(paths utils.NewRepoPaths) error {
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
		t.Fatalf("ensureNewRepoLayout got %#v, want %#v", gotEnsure, targetPaths)
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

func TestRunStopsOnPromptError(t *testing.T) {
	origPrompt := promptRelativeNewRepoPaths
	origEnsure := ensureNewRepoLayout
	origRun := runCommandInDir
	defer func() {
		promptRelativeNewRepoPaths = origPrompt
		ensureNewRepoLayout = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("prompt failed")
	promptRelativeNewRepoPaths = func(_ string, _ []string) (utils.NewRepoPaths, error) {
		return utils.NewRepoPaths{}, expectedErr
	}
	ensureNewRepoLayout = func(_ utils.NewRepoPaths) error {
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
