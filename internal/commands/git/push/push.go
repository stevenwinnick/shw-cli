package push

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var runCommand = utils.RunCommand

func Command() *cli.Command {
	return &cli.Command{
		Name:        "push",
		Summary:     "Push HEAD to origin and set upstream",
		Description: "Runs `git push -u origin HEAD`",
		Usage:       "shw git push",
		Run:         run,
	}
}

func run(_ []string) error {
	return runCommand("git", "push", "-u", "origin", "HEAD")
}
