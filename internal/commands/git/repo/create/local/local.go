package local

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var (
	promptRelativeNewRepoPaths = utils.PromptRelativeNewRepoPaths
	ensureNewRepoLayout        = utils.EnsureNewRepoLayout
	runCommandInDir            = utils.RunCommandInDir
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo in your worktree layout",
		Description: "Prompts for a relative repo path, creates the default-branch working copy, and runs `git init` there",
		Usage:       "shw git repo create local [git-init-args...]",
		Run:         run,
	}
}

func run(args []string) error {
	targetPaths, err := promptRelativeNewRepoPaths("Relative repo path: ", args)
	if err != nil {
		return err
	}
	if err := ensureNewRepoLayout(targetPaths); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, args...)
	return runCommandInDir(targetPaths.MainWorktreeDir, "git", gitArgs...)
}
