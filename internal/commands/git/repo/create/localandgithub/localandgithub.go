package localandgithub

import (
	"fmt"
	"os"
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
		Name:        "localandgithub",
		Summary:     "Create a local repo in your worktree layout, then create a GitHub repo",
		Description: "Prompts for a relative repo path, creates the default-branch working copy, runs `git init`, then runs `gh repo create` there",
		Usage:       "shw git repo create localandgithub [gh-repo-create-args...]",
		Run:         run,
	}
}

func run(args []string) error {
	targetPaths, err := promptRelativeNewRepoPaths("Relative repo path: ", nil)
	if err != nil {
		return err
	}
	if err := ensureNewRepoLayout(targetPaths); err != nil {
		return err
	}

	if err := runCommandInDir(targetPaths.MainWorktreeDir, "git", "init"); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, "\n***Important: Answer \"Push an existing repository to GitHub\" with a path of \".\"***\n\n")

	ghArgs := append([]string{"repo", "create"}, args...)
	return runCommandInDir(targetPaths.MainWorktreeDir, "gh", ghArgs...)
}
