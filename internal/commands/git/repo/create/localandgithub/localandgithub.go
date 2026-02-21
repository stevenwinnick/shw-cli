package localandgithub

import (
	"fmt"
	"os"
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
		Name:        "localandgithub",
		Summary:     "Create local repo in a relative directory, then create a GitHub repo",
		Description: "Prompts for a relative directory path, runs `git init`, then runs `gh repo create` there",
		Usage:       "shw git repo create localandgithub [gh-repo-create-args...]",
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

	if err := runCommandInDir(targetDir, "git", "init"); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, "\n***Important: Answer \"Push an existing repository to GitHub\" with a path of \".\"***\n\n")

	ghArgs := append([]string{"repo", "create"}, args...)
	return runCommandInDir(targetDir, "gh", ghArgs...)
}
