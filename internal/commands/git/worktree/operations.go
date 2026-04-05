package worktree

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"shw-cli/internal/utils"
)

func Create(repoDir string, branchName string) error {
	return CreateWithOptions(repoDir, branchName, true)
}

func CreateWithOptions(repoDir string, branchName string, updateDefaultBranch bool) error {
	return create(repoDir, branchName, updateDefaultBranch, os.Stdout, os.Stderr)
}

func create(repoDir string, branchName string, updateDefaultBranch bool, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree create [flags] <branch-name>")
	}

	repoRoot, repoName, err := utils.RepoDetails(repoDir)
	if err != nil {
		return err
	}

	containerDir, err := utils.RepoContainerDir(repoRoot, repoName)
	if err != nil {
		return err
	}

	defaultBranch, err := utils.RepoDefaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}
	if updateDefaultBranch {
		if err := updateBranchBaseRef(repoRoot, defaultBranch, stderr); err != nil {
			return err
		}
	}

	baseRef, err := baseRef(repoRoot, defaultBranch)
	if err != nil {
		return err
	}

	worktreeDir := strings.ReplaceAll(branchName, "/", "--")
	worktreePath := filepath.Join(containerDir, "worktrees", worktreeDir, repoName)
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory %q: %w", filepath.Dir(worktreePath), err)
	}

	addArgs := []string{"worktree", "add", "-b", branchName, worktreePath, baseRef}
	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", addArgs...); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "Worktree created: %s\nTo switch to it, run: shwcd %s\n", worktreePath, branchName)
	return err
}

func List(repoDir string) error {
	return list(repoDir, os.Stdout, os.Stderr)
}

func list(repoDir string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree list")
	}

	repoRoot, _, err := utils.RepoDetails(repoDir)
	if err != nil {
		return err
	}

	return utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "list")
}

func Remove(repoDir string, branchName string) error {
	return remove(repoDir, branchName, os.Stdout, os.Stderr)
}

func remove(repoDir string, branchName string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree remove [flags] <branch-name>")
	}

	repoRoot, repoName, err := utils.RepoDetails(repoDir)
	if err != nil {
		return err
	}

	entries, err := utils.WorktreeListEntries(repoRoot)
	if err != nil {
		return err
	}

	var targetEntry *utils.WorktreeEntry
	for _, entry := range entries {
		if entry.Branch == branchName {
			entryCopy := entry
			targetEntry = &entryCopy
			break
		}
	}

	if targetEntry == nil {
		listOutput, listErr := utils.CaptureCommandInDir(repoRoot, "git", "worktree", "list")
		if listErr != nil {
			return fmt.Errorf("no worktree found for branch %q", branchName)
		}
		return fmt.Errorf("no worktree found for branch %q\n\nAvailable worktrees:\n%s", branchName, strings.TrimSpace(listOutput))
	}

	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "remove", "--force", targetEntry.Path); err != nil {
		return err
	}

	parentDir := filepath.Dir(targetEntry.Path)
	if utils.WorktreeIsPreferredLayoutPath(targetEntry.Path, repoName) {
		if _, err := os.Stat(parentDir); err == nil {
			if _, err := fmt.Fprintf(stdout, "Removing directory: %s\n", parentDir); err != nil {
				return err
			}
			if err := os.RemoveAll(parentDir); err != nil {
				return fmt.Errorf("failed to remove directory %q: %w", parentDir, err)
			}
		}
	}

	if refExists(repoRoot, "refs/heads/"+branchName) {
		_ = utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "branch", "-D", branchName)
	}

	_, err = fmt.Fprintln(stdout, "Worktree removed successfully")
	return err
}

func CleanAll(repoDir string) error {
	return cleanAll(repoDir, os.Stdout, os.Stderr)
}

func cleanAll(repoDir string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree prune-stale [flags]")
	}

	repoRoot, repoName, err := utils.RepoDetails(repoDir)
	if err != nil {
		return err
	}

	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "prune"); err != nil {
		return err
	}

	entries, err := utils.WorktreeListEntries(repoRoot)
	if err != nil {
		return err
	}

	mainWorktree := utils.WorktreeMainPath(entries, repoName)
	for _, entry := range entries {
		if entry.Path == mainWorktree || entry.Branch == "" {
			continue
		}

		if refExists(repoRoot, "refs/remotes/origin/"+entry.Branch) {
			continue
		}

		if _, err := fmt.Fprintf(stdout, "Removing worktree %s: branch '%s' not on remote\n", entry.Path, entry.Branch); err != nil {
			return err
		}
		if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "remove", "--force", entry.Path); err != nil {
			return err
		}
		if utils.WorktreeIsPreferredLayoutPath(entry.Path, repoName) {
			if err := os.RemoveAll(filepath.Dir(entry.Path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to remove directory %q: %w", filepath.Dir(entry.Path), err)
			}
		}
		_ = utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "branch", "-D", entry.Branch)
	}

	_, err = fmt.Fprintln(stdout, "Worktree cleanup complete")
	return err
}

func baseRef(repoRoot string, defaultBranch string) (string, error) {
	remoteRef := "refs/remotes/origin/" + defaultBranch
	if refExists(repoRoot, remoteRef) {
		return "origin/" + defaultBranch, nil
	}

	localRef := "refs/heads/" + defaultBranch
	if refExists(repoRoot, localRef) {
		return defaultBranch, nil
	}

	return "", fmt.Errorf("could not find a base ref for default branch %q", defaultBranch)
}

func updateBranchBaseRef(repoRoot string, defaultBranch string, stderr io.Writer) error {
	if !refExists(repoRoot, "refs/remotes/origin/"+defaultBranch) {
		return nil
	}

	return utils.RunCommandInDirWithWriters(repoRoot, io.Discard, stderr, "git", "fetch", "origin", defaultBranch)
}

func refExists(repoRoot string, ref string) bool {
	_, err := utils.CaptureCommandInDir(repoRoot, "git", "show-ref", "--verify", ref)
	return err == nil
}
