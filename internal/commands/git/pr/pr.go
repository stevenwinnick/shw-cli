package pr

import (
	"fmt"

	"shw-cli/internal/cli"
)

var repoDirFlag = cli.Flag{
	Long:        "repo-dir",
	Short:       "r",
	ValueName:   "<repo-dir>",
	Description: "Use a repo other than the current directory",
}

func Command() *cli.Command {
	pr := &cli.Command{
		Name:        "pr",
		Summary:     "GitHub pull request helpers",
		Description: "Commands for working with GitHub pull requests",
		Usage:       "shw git pr <command>",
	}

	pr.AddChild(&cli.Command{
		Name:        "merge",
		Summary:     "Squash-merge a PR and clean up its worktree",
		Description: "Squash-merge the pull request for a branch, delete the remote branch, and remove the local worktree and branch. Defaults to the branch checked out in the current directory.",
		Usage:       "shw git pr merge [flags] [<branch>]",
		Flags:       []cli.Flag{repoDirFlag},
		Run:         runMerge,
	})

	return pr
}

func runMerge(args []string) error {
	repoDir, args, err := cli.ConsumeStringFlag(args, repoDirFlag)
	if err != nil {
		return err
	}
	if repoDir == "" {
		repoDir = "."
	}
	if len(args) > 1 {
		return fmt.Errorf("usage: shw git pr merge [flags] [<branch>]")
	}

	branch := ""
	if len(args) == 1 {
		branch = args[0]
	}

	return Merge(repoDir, branch)
}
