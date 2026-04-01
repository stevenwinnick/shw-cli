package update

import (
	"shw-cli/internal/cli"
	aiagentconfigcmd "shw-cli/internal/commands/update/aiagentconfig"
	selfcmd "shw-cli/internal/commands/update/self"
)

func Command() *cli.Command {
	update := &cli.Command{
		Name:        "update",
		Summary:     "Update shw and its dependencies",
		Description: "Commands for updating the CLI and related tools",
		Usage:       "shw update <command>",
	}

	update.AddChild(selfcmd.Command())
	update.AddChild(aiagentconfigcmd.Command())
	return update
}
