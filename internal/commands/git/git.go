package git

import (
	"shw-cli/internal/cli"
	pushcmd "shw-cli/internal/commands/git/push"
	repocmd "shw-cli/internal/commands/git/repo"
)

func Command() *cli.Command {
	git := &cli.Command{
		Name:        "git",
		Summary:     "Git helper commands",
		Description: "Commands for local and GitHub repository setup",
		Usage:       "shw git <command>",
	}

	git.AddChild(repocmd.Command())
	git.AddChild(pushcmd.Command())
	return git
}
