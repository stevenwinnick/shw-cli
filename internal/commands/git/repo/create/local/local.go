package local

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var (
	promptRelativeDir = utils.PromptRelativeDir
	ensureDir         = utils.EnsureDir
	runCommandInDir   = utils.RunCommandInDir
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo in a relative directory",
		Description: "Prompts for a relative directory path and runs `git init` there",
		Usage:       "shw git repo create local [git-init-args...]",
		Run:         run,
	}
}

func run(args []string) error {
	targetDir, err := promptRelativeDir("Relative directory path: ")
	if err != nil {
		return err
	}
	if err := ensureDir(targetDir); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, args...)
	return runCommandInDir(targetDir, "git", gitArgs...)
}
