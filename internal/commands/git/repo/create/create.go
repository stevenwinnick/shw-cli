package create

import (
	"shw-cli/internal/cli"
	githubfromlocalcmd "shw-cli/internal/commands/git/repo/create/githubfromlocal"
)

func Command() *cli.Command {
	create := &cli.Command{
		Name:        "create",
		Summary:     "Create repositories",
		Description: "Create GitHub repos",
		Usage:       "shw git repo create <command>",
	}

	create.AddChild(githubfromlocalcmd.Command())
	return create
}
