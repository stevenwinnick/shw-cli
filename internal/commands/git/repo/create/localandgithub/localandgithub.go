package localandgithub

import (
	"fmt"
	"os"
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var noWorktreesFlag = cli.Flag{
	Long:        "no-worktree-setup",
	Description: "Skip the default worktree setup and create the repo directly in the prompted directory",
}

func Command() *cli.Command {
	return &cli.Command{
		Name:        "localandgithub",
		Summary:     "Create a local repo, then create a GitHub repo",
		Description: "Creates a local repo, then runs `gh repo create` there",
		Usage:       "shw git repo create localandgithub [flags] [gh-repo-create-args...]",
		Flags:       []cli.Flag{noWorktreesFlag},
		Run:         run,
	}
}

func run(args []string) error {
	noWorktrees, ghArgs := cli.ConsumeBoolFlag(args, noWorktreesFlag)

	targetPaths, err := utils.PromptRelativeRepoTargetPaths("Relative repo path: ", nil, !noWorktrees)
	if err != nil {
		return err
	}
	if err := utils.EnsureRepoTargetPaths(targetPaths); err != nil {
		return err
	}

	if err := utils.RunCommandInDir(targetPaths.WorkingDir, "git", "init"); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, "\n***Important: Answer \"Push an existing repository to GitHub\" with a path of \".\"***\n\n")

	ghCommandArgs := append([]string{"repo", "create"}, ghArgs...)
	return utils.RunCommandInDir(targetPaths.WorkingDir, "gh", ghCommandArgs...)
}
