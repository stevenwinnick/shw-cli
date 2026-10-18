package localandgithub

import (
	"fmt"
	"os"
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "localandgithub",
		Summary:     "Create a local repo, then create a GitHub repo",
		Description: "Creates a local repo, then runs `gh repo create` there",
		Usage:       "shw git repo create localandgithub [gh-repo-create-args...]",
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

	if err := utils.RunCommandInDir(repoDir, "git", "init"); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, "\n***Important: Answer \"Push an existing repository to GitHub\" with a path of \".\"***\n\n")

	ghCommandArgs := append([]string{"repo", "create"}, args...)
	return utils.RunCommandInDir(repoDir, "gh", ghCommandArgs...)
}
