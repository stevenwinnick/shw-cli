package worktree

import "testing"

func TestCommandIncludesExpectedSubcommands(t *testing.T) {
	cmd := Command()

	got := map[string]bool{}
	for _, child := range cmd.Children {
		got[child.Name] = true
	}

	for _, name := range []string{"create", "list", "remove", "clean-all", "switch"} {
		if !got[name] {
			t.Fatalf("expected subcommand %q to be present", name)
		}
	}
}

func TestRunCreatePassesArgs(t *testing.T) {
	orig := createWorktree
	defer func() {
		createWorktree = orig
	}()

	var gotRepo string
	var gotBranch string
	createWorktree = func(repoDir string, branchName string) error {
		gotRepo = repoDir
		gotBranch = branchName
		return nil
	}

	if err := runCreate([]string{"repo", "steven/feature"}); err != nil {
		t.Fatalf("runCreate returned error: %v", err)
	}
	if gotRepo != "repo" || gotBranch != "steven/feature" {
		t.Fatalf("unexpected args: repo=%q branch=%q", gotRepo, gotBranch)
	}
}

func TestRunCreateRejectsWrongArgCount(t *testing.T) {
	if err := runCreate([]string{"repo"}); err == nil {
		t.Fatal("expected usage error")
	}
}

func TestRunListPassesArgs(t *testing.T) {
	orig := listWorktrees
	defer func() {
		listWorktrees = orig
	}()

	var gotRepo string
	listWorktrees = func(repoDir string) error {
		gotRepo = repoDir
		return nil
	}

	if err := runList([]string{"repo"}); err != nil {
		t.Fatalf("runList returned error: %v", err)
	}
	if gotRepo != "repo" {
		t.Fatalf("unexpected repo arg: %q", gotRepo)
	}
}

func TestRunRemovePassesArgs(t *testing.T) {
	orig := removeWorktree
	defer func() {
		removeWorktree = orig
	}()

	var gotRepo string
	var gotBranch string
	removeWorktree = func(repoDir string, branchName string) error {
		gotRepo = repoDir
		gotBranch = branchName
		return nil
	}

	if err := runRemove([]string{"repo", "steven/feature"}); err != nil {
		t.Fatalf("runRemove returned error: %v", err)
	}
	if gotRepo != "repo" || gotBranch != "steven/feature" {
		t.Fatalf("unexpected args: repo=%q branch=%q", gotRepo, gotBranch)
	}
}

func TestRunCleanAllPassesArgs(t *testing.T) {
	orig := cleanAllWorktree
	defer func() {
		cleanAllWorktree = orig
	}()

	var gotRepo string
	cleanAllWorktree = func(repoDir string) error {
		gotRepo = repoDir
		return nil
	}

	if err := runCleanAll([]string{"repo"}); err != nil {
		t.Fatalf("runCleanAll returned error: %v", err)
	}
	if gotRepo != "repo" {
		t.Fatalf("unexpected repo arg: %q", gotRepo)
	}
}

func TestRunSwitchPassesArgs(t *testing.T) {
	orig := switchWorktree
	defer func() {
		switchWorktree = orig
	}()

	var gotRepo string
	var gotName string
	switchWorktree = func(repoDir string, name string) error {
		gotRepo = repoDir
		gotName = name
		return nil
	}

	if err := runSwitch([]string{"repo", "steven--feature"}); err != nil {
		t.Fatalf("runSwitch returned error: %v", err)
	}
	if gotRepo != "repo" || gotName != "steven--feature" {
		t.Fatalf("unexpected args: repo=%q name=%q", gotRepo, gotName)
	}
}
