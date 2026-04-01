package cd

import (
	"fmt"
	"os"

	"shw-cli/internal/cli"
)

func Command() *cli.Command {
	cd := &cli.Command{
		Name:        "cd",
		Summary:     "Print a directory path for shell navigation",
		Description: "Resolve and print directory paths for repos and worktrees, designed for use with a shell wrapper like: shwcd() { cd \"$(shw cd \"$@\")\" }",
		Usage:       "shw cd <repo <name> | default | branch-name>",
		Run:         runBranch,
	}

	cd.AddChild(&cli.Command{
		Name:        "repo",
		Summary:     "Print a repo's container directory",
		Description: "Print the container directory for a repo, looking it up in $CODE_ROOT",
		Usage:       "shw cd repo <repo-name>",
		Run:         runRepo,
	})

	return cd
}

func runRepo(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw cd repo <repo-name>")
	}
	return Repo(args[0], os.Stdout)
}

func runBranch(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw cd <repo <name> | default | branch-name>")
	}
	name := args[0]
	if name == "default" {
		return Default(".", os.Stdout)
	}
	return Branch(".", name, os.Stdout)
}
