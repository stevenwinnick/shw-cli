package repo

import (
	"shw-cli/internal/cli"
	createcmd "shw-cli/internal/commands/git/repo/create"
)

func Command() *cli.Command {
	repo := &cli.Command{
		Name:        "repo",
		Summary:     "Repository lifecycle helpers",
		Description: "Commands for creating local and GitHub repositories",
		Usage:       "shw git repo <command>",
	}

	repo.AddChild(createcmd.Command())
	return repo
}
