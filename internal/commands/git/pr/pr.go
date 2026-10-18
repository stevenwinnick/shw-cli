package pr

import (
	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

func Command() *cli.Command {
	pr := &cli.Command{
		Name:        "pr",
		Summary:     "GitHub pull request helpers",
		Description: "Commands for working with GitHub pull requests",
		Usage:       "shw git pr <command>",
	}

	pr.AddChild(&cli.Command{
		Name:        "merge",
		Summary:     "Squash-merge a PR and delete its branch",
		Description: "Runs `gh pr merge [<branch>] --squash --delete-branch`. Defaults to the PR for the current branch; any extra args are passed through to gh.",
		Usage:       "shw git pr merge [<branch>] [gh-pr-merge-args...]",
		Run:         runMerge,
	})

	return pr
}

func runMerge(args []string) error {
	ghArgs := append([]string{"pr", "merge"}, args...)
	ghArgs = append(ghArgs, "--squash", "--delete-branch")
	return utils.RunCommand("gh", ghArgs...)
}
