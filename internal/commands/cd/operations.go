package cd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"shw-cli/internal/utils"
)

func Repo(repoName string, stdout io.Writer) error {
	codeRoot := strings.TrimSpace(os.Getenv("CODE_ROOT"))
	if codeRoot == "" {
		return fmt.Errorf("CODE_ROOT is not set")
	}

	containerDir := filepath.Join(codeRoot, repoName)
	info, err := os.Stat(containerDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("repo %q not found at %s", repoName, containerDir)
	}

	_, err = fmt.Fprintln(stdout, containerDir)
	return err
}

func Default(dir string, stdout io.Writer) error {
	repoRoot, repoName, err := resolveRepoRoot(dir)
	if err != nil {
		return err
	}

	defaultBranch, err := detectDefaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}

	return Branch(dir, defaultBranch, stdout)
}

func Branch(dir string, branchName string, stdout io.Writer) error {
	repoRoot, repoName, err := resolveRepoRoot(dir)
	if err != nil {
		return err
	}

	entries, err := listWorktreeEntries(repoRoot)
	if err != nil {
		return err
	}

	availableBranches := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.branch == "" {
			continue
		}
		availableBranches = append(availableBranches, e.branch)
		if e.branch == branchName {
			if info, err := os.Stat(e.path); err == nil && info.IsDir() {
				_, err = fmt.Fprintln(stdout, e.path)
				return err
			}
			return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, e.path)
		}
	}

	defaultBranch, _ := detectDefaultBranch(repoRoot, repoName)
	if branchName == defaultBranch {
		mainPath := mainWorktreePath(entries, repoName)
		if mainPath != "" {
			if info, err := os.Stat(mainPath); err == nil && info.IsDir() {
				_, err = fmt.Fprintln(stdout, mainPath)
				return err
			}
			return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, mainPath)
		}
	}

	if !containsString(availableBranches, defaultBranch) && defaultBranch != "" {
		availableBranches = append(availableBranches, defaultBranch)
	}
	if len(availableBranches) == 0 {
		return fmt.Errorf("branch %q not found; available: (none)", branchName)
	}
	return fmt.Errorf("branch %q not found; available: %s", branchName, strings.Join(availableBranches, ", "))
}

// resolveRepoRoot finds a git repo root from the given directory, handling
// both directories inside a repo and repo container directories.
func resolveRepoRoot(dir string) (string, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve directory %q: %w", dir, err)
	}

	if repoRoot, err := gitTopLevel(absDir); err == nil {
		return repoRoot, filepath.Base(repoRoot), nil
	}

	repoRoot, err := repoRootFromContainer(absDir)
	if err != nil {
		return "", "", fmt.Errorf("not inside a git repo or recognized repo container: %s", absDir)
	}
	return repoRoot, filepath.Base(repoRoot), nil
}

func gitTopLevel(dir string) (string, error) {
	out, err := utils.CaptureCommandInDir(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "", fmt.Errorf("resolved git repo root is empty")
	}
	return out, nil
}

// repoRootFromContainer resolves a repo root when given the container directory
// (the parent that holds default-branch and worktrees dirs).
func repoRootFromContainer(containerDir string) (string, error) {
	repoName := filepath.Base(containerDir)
	dirEntries, err := os.ReadDir(containerDir)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, de := range dirEntries {
		if !de.IsDir() || de.Name() == "worktrees" {
			continue
		}
		candidate := filepath.Join(containerDir, de.Name(), repoName)
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		if root, err := gitTopLevel(candidate); err == nil {
			candidates = append(candidates, root)
		}
	}

	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("not a recognized repo container")
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple repo roots found under %q", containerDir)
	}
}

func detectDefaultBranch(repoRoot string, repoName string) (string, error) {
	entries, err := listWorktreeEntries(repoRoot)
	if err != nil {
		return "", err
	}

	mainPath := mainWorktreePath(entries, repoName)
	if mainPath != "" {
		return filepath.Base(filepath.Dir(mainPath)), nil
	}

	remoteHead, err := utils.CaptureCommandInDir(repoRoot, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(remoteHead), "origin/"), nil
	}

	return "", fmt.Errorf("failed to detect default branch for %s", repoRoot)
}

type worktreeEntry struct {
	path   string
	branch string
}

func listWorktreeEntries(repoRoot string) ([]worktreeEntry, error) {
	output, err := utils.CaptureCommandInDir(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeEntries(output), nil
}

func parseWorktreeEntries(output string) []worktreeEntry {
	var entries []worktreeEntry
	var current *worktreeEntry

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
			current = &worktreeEntry{path: strings.TrimPrefix(line, "worktree ")}
			continue
		}
		if current != nil && strings.HasPrefix(line, "branch refs/heads/") {
			current.branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries
}

func mainWorktreePath(entries []worktreeEntry, repoName string) string {
	for _, e := range entries {
		parentDir := filepath.Dir(e.path)
		if filepath.Base(filepath.Dir(parentDir)) == repoName {
			return e.path
		}
	}
	if len(entries) == 0 {
		return ""
	}
	return entries[0].path
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
