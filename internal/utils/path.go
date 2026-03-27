package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NewRepoPaths struct {
	RepoRoot        string
	DefaultBranch   string
	MainWorktreeDir string
	WorktreesDir    string
}

var (
	getWorkingDir          = os.Getwd
	detectGitDefaultBranch = DetectGitDefaultBranch
)

func PromptRelativeDir(prompt string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return ResolveRelativeDir(strings.TrimSpace(input))
}

func PromptRelativeNewRepoPaths(prompt string, gitInitArgs []string) (NewRepoPaths, error) {
	fmt.Fprint(os.Stdout, prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return NewRepoPaths{}, fmt.Errorf("failed to read input: %w", err)
	}

	return ResolveRelativeNewRepoPaths(strings.TrimSpace(input), gitInitArgs)
}

func ResolveRelativeDir(relativePath string) (string, error) {
	cleaned := strings.TrimSpace(relativePath)
	if cleaned == "" {
		return "", fmt.Errorf("directory path is required")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("directory path must be relative")
	}

	cwd, err := getWorkingDir()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(cwd, filepath.Clean(cleaned)), nil
}

func ResolveRelativeNewRepoPaths(relativePath string, gitInitArgs []string) (NewRepoPaths, error) {
	repoRoot, err := ResolveRelativeDir(relativePath)
	if err != nil {
		return NewRepoPaths{}, err
	}

	defaultBranch, err := detectGitDefaultBranch(gitInitArgs)
	if err != nil {
		return NewRepoPaths{}, err
	}

	repoName := filepath.Base(repoRoot)

	return NewRepoPaths{
		RepoRoot:        repoRoot,
		DefaultBranch:   defaultBranch,
		MainWorktreeDir: filepath.Join(repoRoot, defaultBranch, repoName),
		WorktreesDir:    filepath.Join(repoRoot, "worktrees"),
	}, nil
}

func EnsureNewRepoLayout(paths NewRepoPaths) error {
	if err := EnsureDir(paths.MainWorktreeDir); err != nil {
		return err
	}
	if err := EnsureDir(paths.WorktreesDir); err != nil {
		return err
	}

	return nil
}

func DetectGitDefaultBranch(gitInitArgs []string) (string, error) {
	if explicitBranch, ok, err := parseGitInitInitialBranch(gitInitArgs); err != nil {
		return "", err
	} else if ok {
		return explicitBranch, nil
	}

	cmd := exec.Command("git", "config", "--get", "init.defaultBranch")
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch, nil
		}
	}

	tempDir, err := os.MkdirTemp("", "shw-cli-git-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir for default branch detection: %w", err)
	}
	defer os.RemoveAll(tempDir)

	initCmd := exec.Command("git", "init", "--quiet")
	initCmd.Dir = tempDir
	if err := initCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to detect git default branch: %w", err)
	}

	headCmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	headCmd.Dir = tempDir
	headOutput, err := headCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read git default branch: %w", err)
	}

	branch := strings.TrimSpace(string(headOutput))
	if branch == "" {
		return "", fmt.Errorf("git default branch is empty")
	}

	return branch, nil
}

func parseGitInitInitialBranch(args []string) (string, bool, error) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-b" || args[i] == "--initial-branch":
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
				return "", false, fmt.Errorf("git init option %q requires a branch name", args[i])
			}
			return args[i+1], true, nil
		case strings.HasPrefix(args[i], "--initial-branch="):
			branch := strings.TrimSpace(strings.TrimPrefix(args[i], "--initial-branch="))
			if branch == "" {
				return "", false, fmt.Errorf("git init option %q requires a branch name", "--initial-branch")
			}
			return branch, true, nil
		}
	}

	return "", false, nil
}

func EnsureDir(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to access directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}

	return nil
}

func EnsureExistingDir(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return fmt.Errorf("failed to access directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}

	return nil
}
