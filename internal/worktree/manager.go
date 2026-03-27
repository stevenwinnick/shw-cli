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

type entry struct {
	Path   string
	Branch string
}

var (
	runCommandInDir               = utils.RunCommandInDir
	captureCommandInDir           = utils.CaptureCommandInDir
	stdout              io.Writer = os.Stdout
	stat                          = os.Stat
	mkdirAll                      = os.MkdirAll
	removeAll                     = os.RemoveAll
	readDir                       = os.ReadDir
	getenv                        = os.Getenv
)

func Create(repoDir string, branchName string) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree create <repo-dir> <branch-name>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	containerDir, err := repoContainerDir(repoRoot, repoName)
	if err != nil {
		return err
	}

	defaultBranch, err := defaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}

	baseRef, err := baseRef(repoRoot, defaultBranch)
	if err != nil {
		return err
	}

	worktreeDir := strings.ReplaceAll(branchName, "/", "--")
	worktreePath := filepath.Join(containerDir, "worktrees", worktreeDir, repoName)
	if err := mkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory %q: %w", filepath.Dir(worktreePath), err)
	}

	if err := runCommandInDir(repoRoot, "git", "worktree", "add", "-b", branchName, worktreePath, baseRef); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "Worktree created: %s\n", worktreePath)
	return err
}

func List(repoDir string) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree list <repo-dir>")
	}

	repoRoot, _, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	if err := runCommandInDir(repoRoot, "git", "worktree", "list"); err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(stdout, "%s:\n", entry.Path); err != nil {
			return err
		}

		if _, err := stat(entry.Path); err != nil {
			if _, err := fmt.Fprintln(stdout, "(not accessible)"); err != nil {
				return err
			}
			continue
		}

		status, err := captureCommandInDir(entry.Path, "git", "status", "--short")
		if err != nil {
			if _, err := fmt.Fprintln(stdout, "(not accessible)"); err != nil {
				return err
			}
			continue
		}

		if status == "" {
			continue
		}

		if _, err := fmt.Fprint(stdout, status); err != nil {
			return err
		}
		if !strings.HasSuffix(status, "\n") {
			if _, err := fmt.Fprintln(stdout); err != nil {
				return err
			}
		}
	}

	return nil
}

func Remove(repoDir string, branchName string) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree remove <repo-dir> <branch-name>")
	}

	repoRoot, _, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}

	worktreePath := ""
	for _, entry := range entries {
		if entry.Branch == branchName {
			worktreePath = entry.Path
			break
		}
	}

	if worktreePath == "" {
		listOutput, listErr := captureCommandInDir(repoRoot, "git", "worktree", "list")
		if listErr != nil {
			return fmt.Errorf("no worktree found for branch %q", branchName)
		}
		return fmt.Errorf("no worktree found for branch %q\n\nAvailable worktrees:\n%s", branchName, strings.TrimSpace(listOutput))
	}

	if _, err := fmt.Fprintf(stdout, "Removing worktree: %s (branch: %s)\n", worktreePath, branchName); err != nil {
		return err
	}
	if err := runCommandInDir(repoRoot, "git", "worktree", "remove", "--force", worktreePath); err != nil {
		return err
	}

	parentDir := filepath.Dir(worktreePath)
	if _, err := stat(parentDir); err == nil {
		if _, err := fmt.Fprintf(stdout, "Removing directory: %s\n", parentDir); err != nil {
			return err
		}
		if err := removeAll(parentDir); err != nil {
			return fmt.Errorf("failed to remove directory %q: %w", parentDir, err)
		}
	}

	if refExists(repoRoot, "refs/heads/"+branchName) {
		if _, err := fmt.Fprintf(stdout, "Deleting local branch: %s\n", branchName); err != nil {
			return err
		}
		_ = runCommandInDir(repoRoot, "git", "branch", "-D", branchName)
	}

	_, err = fmt.Fprintln(stdout, "Worktree removed successfully")
	return err
}

func CleanAll(repoDir string) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree clean-all <repo-dir>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	if err := runCommandInDir(repoRoot, "git", "worktree", "prune"); err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}

	mainWorktree := mainWorktreePath(entries, repoName)
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
		if err := runCommandInDir(repoRoot, "git", "worktree", "remove", "--force", entry.Path); err != nil {
			return err
		}
		if err := removeAll(filepath.Dir(entry.Path)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to remove directory %q: %w", filepath.Dir(entry.Path), err)
		}
		_ = runCommandInDir(repoRoot, "git", "branch", "-D", entry.Branch)
	}

	_, err = fmt.Fprintln(stdout, "Worktree cleanup complete")
	return err
}

func Switch(repoDir string, name string) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("usage: shw git worktree switch <repo-dir> <name>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	containerDir, err := repoContainerDir(repoRoot, repoName)
	if err != nil {
		return err
	}

	worktreePath := filepath.Join(containerDir, "worktrees", name, repoName)
	info, err := stat(worktreePath)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("worktree path exists but is not a directory: %s", worktreePath)
		}
		_, err = fmt.Fprintln(stdout, worktreePath)
		return err
	}

	names, listErr := availableWorktreeNames(filepath.Join(containerDir, "worktrees"))
	if listErr != nil {
		return listErr
	}
	if len(names) == 0 {
		return fmt.Errorf("worktree %q not found; available: (none)", name)
	}

	return fmt.Errorf("worktree %q not found; available: %s", name, strings.Join(names, ", "))
}

func repoDetails(repoDir string) (string, string, error) {
	absoluteRepoDir, err := filepath.Abs(repoDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve repo dir %q: %w", repoDir, err)
	}

	repoRoot, err := captureCommandInDir(absoluteRepoDir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve git repo from %q: %w", repoDir, err)
	}

	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return "", "", fmt.Errorf("failed to resolve git repo from %q", repoDir)
	}

	return repoRoot, filepath.Base(repoRoot), nil
}

func repoContainerDir(repoRoot string, repoName string) (string, error) {
	parentDir := filepath.Dir(repoRoot)
	grandparentDir := filepath.Dir(parentDir)
	if filepath.Base(grandparentDir) == repoName {
		return grandparentDir, nil
	}

	if filepath.Base(grandparentDir) == "worktrees" {
		containerDir := filepath.Dir(grandparentDir)
		if filepath.Base(containerDir) == repoName {
			return containerDir, nil
		}
	}

	codeRoot := strings.TrimSpace(getenv("CODE_ROOT"))
	if codeRoot != "" {
		return filepath.Join(codeRoot, repoName), nil
	}

	return "", fmt.Errorf("repo %q is not in the preferred worktree layout and CODE_ROOT is not set", repoRoot)
}

func defaultBranch(repoRoot string, repoName string) (string, error) {
	entries, err := listEntries(repoRoot)
	if err != nil {
		return "", err
	}

	mainWorktree := mainWorktreePath(entries, repoName)
	if mainWorktree != "" {
		return filepath.Base(filepath.Dir(mainWorktree)), nil
	}

	remoteHead, err := captureCommandInDir(repoRoot, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(remoteHead), "origin/"), nil
	}

	return "", fmt.Errorf("failed to detect default branch for %s", repoRoot)
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

func refExists(repoRoot string, ref string) bool {
	_, err := captureCommandInDir(repoRoot, "git", "show-ref", "--verify", ref)
	return err == nil
}

func listEntries(repoRoot string) ([]entry, error) {
	output, err := captureCommandInDir(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseEntries(output), nil
}

func parseEntries(output string) []entry {
	var entries []entry
	var current *entry

	for _, line := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &entry{Path: strings.TrimPrefix(line, "worktree ")}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}

func mainWorktreePath(entries []entry, repoName string) string {
	for _, entry := range entries {
		if isMainWorktreePath(entry.Path, repoName) {
			return entry.Path
		}
	}
	if len(entries) == 0 {
		return ""
	}
	return entries[0].Path
}

func isMainWorktreePath(path string, repoName string) bool {
	parentDir := filepath.Dir(path)
	return filepath.Base(filepath.Dir(parentDir)) == repoName
}

func availableWorktreeNames(worktreesDir string) ([]string, error) {
	dirEntries, err := readDir(worktreesDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees in %q: %w", worktreesDir, err)
	}

	names := make([]string, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			names = append(names, dirEntry.Name())
		}
	}

	return names, nil
}
