package pr

import (
	"fmt"
	"io"
	"os"
	"strings"

	"shw-cli/internal/commands/git/worktree"
	"shw-cli/internal/utils"
)

// Merge squash-merges the pull request for a branch, deletes the remote branch,
// and removes the local worktree and branch. When branch is empty it defaults to
// the branch checked out in repoDir.
func Merge(repoDir string, branch string) error {
	return merge(repoDir, branch, os.Stdout, os.Stderr)
}

func merge(repoDir string, branch string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git pr merge [flags] [<branch>]")
	}

	repoRoot, repoName, err := utils.RepoDetails(repoDir)
	if err != nil {
		return err
	}

	if strings.TrimSpace(branch) == "" {
		branch, err = currentBranch(repoRoot)
		if err != nil {
			return err
		}
	}

	// Run gh and the local cleanup from the main worktree rather than the
	// worktree we are about to remove. Run from the branch's own worktree and
	// gh would try to switch it off the branch before deleting it, which fails
	// because the default branch is already checked out in the main worktree.
	entries, err := utils.WorktreeListEntries(repoRoot)
	if err != nil {
		return err
	}
	mainRoot := utils.WorktreeMainPath(entries, repoName)
	if mainRoot == "" {
		mainRoot = repoRoot
	}

	// Squash-merge and delete the remote branch. gh also attempts to delete the
	// local branch, but that fails in this worktree layout because the branch is
	// checked out; we remove the worktree and branch ourselves below, so tolerate
	// a gh failure as long as the PR actually merged.
	mergeErr := utils.RunCommandInDirWithWriters(mainRoot, stdout, stderr, "gh", "pr", "merge", branch, "--squash", "--delete-branch")
	if mergeErr != nil {
		if !branchMerged(mainRoot, branch) {
			return mergeErr
		}
		fmt.Fprintln(stderr, "gh could not delete the local branch (it is checked out in a worktree); removing it directly.")
	}

	// Remove the worktree and delete the local branch.
	return worktree.Remove(mainRoot, branch)
}

func currentBranch(dir string) (string, error) {
	out, err := utils.CaptureCommandInDir(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	branch := strings.TrimSpace(out)
	if branch == "" || branch == "HEAD" {
		return "", fmt.Errorf("could not determine the current branch in %q; pass a branch name explicitly", dir)
	}
	return branch, nil
}

// branchMerged reports whether the pull request for a branch is already merged,
// used to decide whether a gh error was only the (expected) local-branch deletion
// failure rather than a failed merge.
func branchMerged(dir string, branch string) bool {
	out, err := utils.CaptureCommandInDir(dir, "gh", "pr", "view", branch, "--json", "state", "--jq", ".state")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "MERGED"
}
