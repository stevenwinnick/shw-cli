package self

import (
	"fmt"
	"os"
	"path/filepath"

	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

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
	trunkDir, err := resolvetrunkDir()
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Pulling latest trunk...")
	if err := utils.RunCommandInDir(trunkDir, "git", "pull"); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	fmt.Fprintln(os.Stderr, "\nReinstalling shw...")
	installScript := filepath.Join(trunkDir, "scripts", "install.sh")
	return utils.RunCommand("bash", installScript)
}

func resolvetrunkDir() (string, error) {
	codeRoot := os.Getenv("CODE_ROOT")
	if codeRoot == "" {
		return "", fmt.Errorf("CODE_ROOT environment variable is not set")
	}

	trunkDir := filepath.Join(codeRoot, "shw-cli", "trunk", "shw-cli")
	if err := utils.EnsureExistingDir(trunkDir); err != nil {
		return "", fmt.Errorf("shw-cli trunk not found at %s: %w", trunkDir, err)
	}

	return trunkDir, nil
}
