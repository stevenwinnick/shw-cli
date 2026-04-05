package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepoDetails resolves the repo root and name from any directory inside a repo
// or from a repo container directory in the preferred worktree layout.
func RepoDetails(dir string) (root string, name string, err error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve directory %q: %w", dir, err)
	}

	if repoRoot, err := resolveGitTopLevel(absDir); err == nil {
		return repoRoot, filepath.Base(repoRoot), nil
	}

	repoRoot, err := resolveRepoRootFromContainer(absDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve git repo from %q: %w", dir, err)
	}
	return repoRoot, filepath.Base(repoRoot), nil
}

// RepoContainerDir finds the container directory for a repo in the preferred
// worktree layout. The container holds the default-branch directory and
// the worktrees directory.
func RepoContainerDir(repoRoot string, repoName string) (string, error) {
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

	codeRoot := strings.TrimSpace(os.Getenv("CODE_ROOT"))
	if codeRoot != "" {
		return filepath.Join(codeRoot, repoName), nil
	}

	return "", fmt.Errorf("repo %q is not in the preferred worktree layout and CODE_ROOT is not set", repoRoot)
}

// RepoDefaultBranch detects the default branch for a repo by checking the
// preferred layout structure first, then falling back to the remote HEAD.
func RepoDefaultBranch(repoRoot string, repoName string) (string, error) {
	entries, err := WorktreeListEntries(repoRoot)
	if err != nil {
		return "", err
	}

	mainPath := WorktreeMainPath(entries, repoName)
	if mainPath != "" {
		return filepath.Base(filepath.Dir(mainPath)), nil
	}

	remoteHead, err := CaptureCommandInDir(repoRoot, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(remoteHead), "origin/"), nil
	}

	return "", fmt.Errorf("failed to detect default branch for %s", repoRoot)
}

func resolveGitTopLevel(dir string) (string, error) {
	repoRoot, err := CaptureCommandInDir(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return "", fmt.Errorf("resolved git repo root is empty")
	}
	return repoRoot, nil
}

func resolveRepoRootFromContainer(containerDir string) (string, error) {
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
		if root, err := resolveGitTopLevel(candidate); err == nil {
			candidates = append(candidates, root)
		}
	}

	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("not a git repo or recognized repo container")
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple repo roots found under %q", containerDir)
	}
}
