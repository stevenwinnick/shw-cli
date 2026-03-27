package localandgithub

import (
	"fmt"
	"os"
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
		Name:        "localandgithub",
		Summary:     "Create a local repo, optionally using your worktree layout, then create a GitHub repo",
		Description: "Prompts for a relative repo path, creates the repo in your worktree layout by default, runs `git init`, then runs `gh repo create` there",
		Usage:       "shw git repo create localandgithub [--no-worktrees] [gh-repo-create-args...]",
		Notes: []string{
			"Pass --no-worktrees to create the repo directly in the prompted directory",
		},
		Run: run,
	}
}

func run(args []string) error {
	useWorktrees, ghArgs := parseNoWorktreesFlag(args)

	targetPaths, err := promptRelativeRepoTargetPaths("Relative repo path: ", nil, useWorktrees)
	if err != nil {
		return err
	}
	if err := ensureRepoTargetPaths(targetPaths); err != nil {
		return err
	}

	if err := runCommandInDir(targetPaths.WorkingDir, "git", "init"); err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, "\n***Important: Answer \"Push an existing repository to GitHub\" with a path of \".\"***\n\n")

	ghCommandArgs := append([]string{"repo", "create"}, ghArgs...)
	return runCommandInDir(targetPaths.WorkingDir, "gh", ghCommandArgs...)
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
