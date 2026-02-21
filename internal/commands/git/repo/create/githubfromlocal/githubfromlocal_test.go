package githubfromlocal

import (
	"errors"
	"testing"
)

func TestRunCallsGitHubCreateInSelectedDir(t *testing.T) {
	origPrompt := promptRelativeDir
	origEnsure := ensureExistingDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureExistingDir = origEnsure
		runCommandInDir = origRun
	}()

	const target = "/tmp/existing-repo"
	var gotDir string
	var gotName string
	var gotArgs []string

	promptRelativeDir = func(_ string) (string, error) { return target, nil }
	ensureExistingDir = func(dir string) error {
		if dir != target {
			t.Fatalf("ensureExistingDir got %q, want %q", dir, target)
		}
		return nil
	}
	runCommandInDir = func(dir string, name string, args ...string) error {
		gotDir = dir
		gotName = name
		gotArgs = append([]string{}, args...)
		return nil
	}

	if err := run([]string{"my-repo", "--public"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if gotDir != target {
		t.Fatalf("run dir got %q, want %q", gotDir, target)
	}
	if gotName != "gh" {
		t.Fatalf("run command got %q, want gh", gotName)
	}
	want := []string{"repo", "create", "my-repo", "--public"}
	if len(gotArgs) != len(want) {
		t.Fatalf("run args len mismatch: got %v want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("run args mismatch at %d: got %q want %q", i, gotArgs[i], want[i])
		}
	}
}

func TestRunStopsIfDirectoryIsMissing(t *testing.T) {
	origPrompt := promptRelativeDir
	origEnsure := ensureExistingDir
	origRun := runCommandInDir
	defer func() {
		promptRelativeDir = origPrompt
		ensureExistingDir = origEnsure
		runCommandInDir = origRun
	}()

	expectedErr := errors.New("missing directory")
	promptRelativeDir = func(_ string) (string, error) { return "./repo", nil }
	ensureExistingDir = func(_ string) error { return expectedErr }
	runCommandInDir = func(_ string, _ string, _ ...string) error {
		t.Fatal("runCommandInDir should not be called when ensureExistingDir fails")
		return nil
	}

	err := run(nil)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected ensureExistingDir error, got %v", err)
	}
}
