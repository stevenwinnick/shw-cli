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

func Command() *cli.Command {
	return &cli.Command{
		Name:        "local",
		Summary:     "Create a local git repo, optionally using your worktree layout",
		Description: "Prompts for a relative repo path, creates the repo in your worktree layout by default, and runs `git init` there",
		Usage:       "shw git repo create local [--no-worktrees] [git-init-args...]",
		Notes: []string{
			"Pass --no-worktrees to create the repo directly in the prompted directory",
		},
		Run: run,
	}
}

func run(args []string) error {
	useWorktrees, gitInitArgs := parseNoWorktreesFlag(args)

	targetPaths, err := promptRelativeRepoTargetPaths("Relative repo path: ", gitInitArgs, useWorktrees)
	if err != nil {
		return err
	}
	if err := ensureRepoTargetPaths(targetPaths); err != nil {
		return err
	}

	gitArgs := append([]string{"init"}, gitInitArgs...)
	return runCommandInDir(targetPaths.WorkingDir, "git", gitArgs...)
}

func parseNoWorktreesFlag(args []string) (bool, []string) {
	useWorktrees := true
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--no-worktrees" {
			useWorktrees = false
			continue
		}
		filtered = append(filtered, arg)
	}

	return useWorktrees, filtered
}
