package pd

import (
	"fmt"
	"os"

	"shw-cli/internal/cli"
)

func Command() *cli.Command {
	cmd := &cli.Command{
		Name:        "pd",
		Summary:     "Print a directory path for shell navigation",
		Description: "Resolve and print directory paths for repos and worktrees, designed for use with a shell wrapper like: shwcd() { cd \"$(shw pd \"$@\")\" }",
		Usage:       "shw pd <repo <name> | default | branch-name>",
		Notes: []string{
			"default      Print the default branch worktree (requires being inside a repo)",
			"<branch>     Print a branch's worktree (requires being inside a repo)",
		},
		Run: runBranch,
	}

	cmd.AddChild(&cli.Command{
		Name:        "repo",
		Summary:     "Print a repo's container directory",
		Description: "Print the container directory for a repo, looking it up in $CODE_ROOT",
		Usage:       "shw pd repo <repo-name>",
		Run:         runRepo,
	})

	return cmd
}

func runRepo(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw pd repo <repo-name>")
	}
	return Repo(args[0], os.Stdout)
}

func runBranch(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw pd <repo <name> | default | branch-name>")
	}
	name := args[0]
	if name == "default" {
		return Default(".", os.Stdout)
	}
	return Branch(".", name, os.Stdout)
}
