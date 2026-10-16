package self

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

const defaultBranch = "trunk"

func Command() *cli.Command {
	return &cli.Command{
		Name:        "self",
		Summary:     "Update shw to the latest version",
		Description: "Pulls the latest trunk and reinstalls the shw binary",
		Usage:       "shw update self",
		Run:         run,
	}
}

func run(_ []string) error {
	repoDir, err := resolveRepoDir()
	if err != nil {
		return err
	}

	branch, err := utils.CaptureCommandInDir(repoDir, "git", "branch", "--show-current")
	if err != nil {
		return fmt.Errorf("failed to read current branch: %w", err)
	}
	if branch = strings.TrimSpace(branch); branch != defaultBranch {
		return fmt.Errorf("shw-cli checkout at %s is on branch %q; switch to %s first", repoDir, branch, defaultBranch)
	}

	fmt.Fprintln(os.Stderr, "Pulling latest trunk...")
	if err := utils.RunCommandInDir(repoDir, "git", "pull", "--ff-only"); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	fmt.Fprintln(os.Stderr, "\nReinstalling shw...")
	installScript := filepath.Join(repoDir, "scripts", "install.sh")
	return utils.RunCommand("bash", installScript)
}

func resolveRepoDir() (string, error) {
	codeRoot := os.Getenv("CODE_ROOT")
	if codeRoot == "" {
		return "", fmt.Errorf("CODE_ROOT environment variable is not set")
	}

	repoDir := filepath.Join(codeRoot, "shw-cli")
	if err := utils.EnsureExistingDir(repoDir); err != nil {
		return "", fmt.Errorf("shw-cli checkout not found at %s: %w", repoDir, err)
	}

	return repoDir, nil
}
