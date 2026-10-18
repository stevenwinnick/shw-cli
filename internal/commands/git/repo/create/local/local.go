package local

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo",
		Description: "Creates a local git repo",
		Usage:       "shw git repo create local [git-init-args...]",
		Run:         run,
	}
}

func run(args []string) error {
	repoDir, err := utils.PromptRelativeDir("Relative repo path: ")
	if err != nil {
		return err
	}
	if err := utils.EnsureDir(repoDir); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, args...)
	return utils.RunCommandInDir(repoDir, "git", gitArgs...)
}
