package local

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var noWorktreesFlag = cli.Flag{
	Long:        "no-worktree-setup",
	Description: "Skip the default worktree setup and create the repo directly in the prompted directory",
}

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo",
		Description: "Creates a local git repo",
		Usage:       "shw git repo create local [flags] [git-init-args...]",
		Flags:       []cli.Flag{noWorktreesFlag},
		Run:         run,
	}
}

func run(args []string) error {
	noWorktrees, gitInitArgs := cli.ConsumeBoolFlag(args, noWorktreesFlag)

	targetPaths, err := utils.PromptRelativeRepoTargetPaths("Relative repo path: ", gitInitArgs, !noWorktrees)
	if err != nil {
		return err
	}
	if err := utils.EnsureRepoTargetPaths(targetPaths); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, gitInitArgs...)
	return utils.RunCommandInDir(targetPaths.WorkingDir, "git", gitArgs...)
}
