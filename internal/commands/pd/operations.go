package pd

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
	repoRoot, repoName, err := utils.RepoDetails(dir)
	if err != nil {
		return err
	}

	defaultBranch, err := utils.RepoDefaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}

	return Branch(dir, defaultBranch, stdout)
}

func Branch(dir string, branchName string, stdout io.Writer) error {
	repoRoot, repoName, err := utils.RepoDetails(dir)
	if err != nil {
		return err
	}

	entries, err := utils.WorktreeListEntries(repoRoot)
	if err != nil {
		return err
	}

	availableBranches := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Branch == "" {
			continue
		}
		availableBranches = append(availableBranches, e.Branch)
		if e.Branch == branchName {
			if info, err := os.Stat(e.Path); err == nil && info.IsDir() {
				_, err = fmt.Fprintln(stdout, e.Path)
				return err
			}
			return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, e.Path)
		}
	}

	defaultBranch, _ := utils.RepoDefaultBranch(repoRoot, repoName)
	if branchName == defaultBranch {
		mainPath := utils.WorktreeMainPath(entries, repoName)
		if mainPath != "" {
			if info, err := os.Stat(mainPath); err == nil && info.IsDir() {
				_, err = fmt.Fprintln(stdout, mainPath)
				return err
			}
			return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, mainPath)
		}
	}

	if defaultBranch != "" && !containsString(availableBranches, defaultBranch) {
		availableBranches = append(availableBranches, defaultBranch)
	}
	if len(availableBranches) == 0 {
		return fmt.Errorf("branch %q not found; available: (none)", branchName)
	}
	return fmt.Errorf("branch %q not found; available: %s", branchName, strings.Join(availableBranches, ", "))
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
