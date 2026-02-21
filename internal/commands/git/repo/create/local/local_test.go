package local

import (
	"errors"
	"fmt"
	"testing"
)

func TestRunUsesPromptEnsureAndGitInit(t *testing.T) {
	origPrompt := promptRelativeDir
	origEnsure := ensureDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureDir = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/repo"
	var gotPrompt string
	var gotEnsure string
	var gotDir string
	var gotName string
	var gotArgs []string

	promptRelativeDir = func(prompt string) (string, error) {
		gotPrompt = prompt
		return target, nil
	}
	ensureDir = func(dir string) error {
		gotEnsure = dir
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
	if gotEnsure != target {
		t.Fatalf("ensureDir got %q, want %q", gotEnsure, target)
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
	origPrompt := promptRelativeDir
	origEnsure := ensureDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureDir = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("prompt failed")
	promptRelativeDir = func(_ string) (string, error) { return "", expectedErr }
	ensureDir = func(_ string) error {
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
