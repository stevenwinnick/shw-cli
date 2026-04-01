package aiagentconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "ai-agent-config",
		Summary:     "Update AI agent configuration",
		Description: "Pulls the latest trunk and applies the configuration to all agents",
		Usage:       "shw update ai-agent-config",
		Run:         run,
	}
}

func run(_ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	script := filepath.Join(homeDir, ".ai-agent-config", "scripts", "apply-updated-trunk-config.sh")
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("ai-agent-config not found at %s: %w", script, err)
	}

	return utils.RunCommand("bash", script)
}
