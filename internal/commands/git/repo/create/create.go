package create

import (
	"shw-cli/internal/cli"
	githubfromlocalcmd "shw-cli/internal/commands/git/repo/create/githubfromlocal"
	localcmd "shw-cli/internal/commands/git/repo/create/local"
	localandgithubcmd "shw-cli/internal/commands/git/repo/create/localandgithub"
)

func Command() *cli.Command {
	create := &cli.Command{
		Name:        "create",
		Summary:     "Create repositories",
		Description: "Create local repos, GitHub repos, or both",
		Usage:       "shw git repo create <command>",
	}

	create.AddChild(localcmd.Command())
	create.AddChild(localandgithubcmd.Command())
	create.AddChild(githubfromlocalcmd.Command())
	return create
}
