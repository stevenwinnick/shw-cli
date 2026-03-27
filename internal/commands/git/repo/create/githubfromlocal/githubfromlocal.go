package githubfromlocal

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "githubfromlocal",
		Summary:     "Create a GitHub repo from a local directory",
		Description: "Prompts for a relative directory path and runs `gh repo create` there",
		Usage:       "shw git repo create githubfromlocal [gh-repo-create-args...]",
		Run:         run,
	}
}

func run(args []string) error {
	targetDir, err := utils.PromptRelativeDir("Relative directory path: ")
	if err != nil {
		return err
	}
	if err := utils.EnsureExistingDir(targetDir); err != nil {
		return err
	}

	ghArgs := append([]string{"repo", "create"}, args...)
	return utils.RunCommandInDir(targetDir, "gh", ghArgs...)
}
