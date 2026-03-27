package local

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

var (
	promptRelativeRepoTargetPaths = utils.PromptRelativeRepoTargetPaths
	ensureRepoTargetPaths         = utils.EnsureRepoTargetPaths
	runCommandInDir               = utils.RunCommandInDir
)

var noWorktreesFlag = cli.Flag{
	Long:        "no-worktrees",
	Description: "Create the repo directly in the prompted directory",
}

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo",
		Description: "Creates a local git repo in your worktree layout by default",
		Usage:       "shw git repo create local [flags] [git-init-args...]",
		Flags:       []cli.Flag{noWorktreesFlag},
		Run:         run,
	}
}

func run(args []string) error {
	noWorktrees, gitInitArgs := cli.ConsumeBoolFlag(args, noWorktreesFlag)

	targetPaths, err := promptRelativeRepoTargetPaths("Relative repo path: ", gitInitArgs, !noWorktrees)
	if err != nil {
		return err
	}
	if err := ensureRepoTargetPaths(targetPaths); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, gitInitArgs...)
	return runCommandInDir(targetPaths.WorkingDir, "git", gitArgs...)
}
