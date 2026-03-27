package worktree

import (
	"fmt"

	"shw-cli/internal/cli"
)

var quietFlag = cli.Flag{
	Long:        "quiet",
	Short:       "q",
	Description: "Print only the resulting path when supported",
}

func Command() *cli.Command {
	worktree := &cli.Command{
		Name:        "worktree",
		Summary:     "Manage git worktrees in Steven's preferred layout",
		Description: "Create, list, remove, clean, and resolve git worktrees in Steven's preferred layout",
		Usage:       "shw git worktree <command>",
	}

	worktree.AddChild(newLeafCommand(
		"create",
		"Create a worktree for a branch",
		"Create a worktree for a branch in Steven's preferred layout",
		"shw git worktree create [flags] <repo-dir> <branch-name>",
		runCreate,
		quietFlag,
	))
	worktree.AddChild(newLeafCommand(
		"list",
		"List worktrees and statuses",
		"List worktrees for a repo and show the status of each worktree",
		"shw git worktree list <repo-dir>",
		runList,
	))
	worktree.AddChild(newLeafCommand(
		"remove",
		"Remove a worktree and local branch",
		"Remove a worktree for a branch and delete the local branch",
		"shw git worktree remove <repo-dir> <branch-name>",
		runRemove,
	))
	worktree.AddChild(newLeafCommand(
		"clean-all",
		"Prune stale worktrees",
		"Prune stale worktree metadata and remove worktrees for branches that no longer have remote-tracking refs",
		"shw git worktree clean-all <repo-dir>",
		runCleanAll,
	))
	worktree.AddChild(newLeafCommand(
		"path",
		"Print a worktree path",
		"Resolve a named worktree in Steven's preferred layout and print its path for use with cd or as a workdir",
		"shw git worktree path <repo-dir> <name>",
		runPath,
	))

	return worktree
}

func newLeafCommand(name string, summary string, description string, usage string, run func(args []string) error, flags ...cli.Flag) *cli.Command {
	return &cli.Command{
		Name:        name,
		Summary:     summary,
		Description: description,
		Usage:       usage,
		Flags:       flags,
		Run:         run,
	}
}

func runCreate(args []string) error {
	quiet, args := cli.ConsumeBoolFlag(args, quietFlag)
	if len(args) != 2 {
		return fmt.Errorf("usage: shw git worktree create [flags] <repo-dir> <branch-name>")
	}
	return Create(args[0], args[1], quiet)
}

func runList(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw git worktree list <repo-dir>")
	}
	return List(args[0])
}

func runRemove(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: shw git worktree remove <repo-dir> <branch-name>")
	}
	return Remove(args[0], args[1])
}

func runCleanAll(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: shw git worktree clean-all <repo-dir>")
	}
	return CleanAll(args[0])
}

func runPath(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: shw git worktree path <repo-dir> <name>")
	}
	return Path(args[0], args[1])
}
