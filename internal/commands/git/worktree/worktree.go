package worktree

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

var noUpdateDefaultBranchFlag = cli.Flag{
	Long:        "no-update-default-branch",
	Description: "Skip fetching the default branch before creating the worktree",
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
		"shw git worktree create [flags] <branch-name>",
		runCreate,
		repoDirFlag,
		noUpdateDefaultBranchFlag,
	))
	worktree.AddChild(newLeafCommand(
		"list",
		"List worktrees",
		"List worktrees for a repo",
		"shw git worktree list",
		runList,
		repoDirFlag,
	))
	worktree.AddChild(newLeafCommand(
		"remove",
		"Remove a worktree and local branch",
		"Remove a worktree for a branch and delete the local branch",
		"shw git worktree remove [flags] <branch-name>",
		runRemove,
		repoDirFlag,
	))
	worktree.AddChild(newLeafCommand(
		"prune-stale",
		"Prune stale worktrees",
		"Prune stale worktree metadata and remove worktrees for branches that no longer have remote-tracking refs",
		"shw git worktree prune-stale [flags]",
		runPruneStale,
		repoDirFlag,
	))
	worktree.AddChild(newLeafCommand(
		"path",
		"Print a worktree path",
		"Resolve a branch's worktree in Steven's preferred layout and print its path so you can switch to it with commands like `cd $(shw git worktree path <branch-name>)`",
		"shw git worktree path [flags] <branch-name>",
		runPath,
		repoDirFlag,
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
	noUpdateDefaultBranch, args := cli.ConsumeBoolFlag(args, noUpdateDefaultBranchFlag)
	repoDir, args, err := consumeRepoDir(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: shw git worktree create [flags] <branch-name>")
	}
	return CreateWithOptions(repoDir, args[0], !noUpdateDefaultBranch)
}

func runList(args []string) error {
	repoDir, args, err := consumeRepoDir(args)
	if err != nil {
		return err
	}
	if len(args) != 0 {
		return fmt.Errorf("usage: shw git worktree list")
	}
	return List(repoDir)
}

func runRemove(args []string) error {
	repoDir, args, err := consumeRepoDir(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: shw git worktree remove [flags] <branch-name>")
	}
	return Remove(repoDir, args[0])
}

func runPruneStale(args []string) error {
	repoDir, args, err := consumeRepoDir(args)
	if err != nil {
		return err
	}
	if len(args) != 0 {
		return fmt.Errorf("usage: shw git worktree prune-stale [flags]")
	}
	return CleanAll(repoDir)
}

func runPath(args []string) error {
	repoDir, args, err := consumeRepoDir(args)
	if err != nil {
		return err
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: shw git worktree path [flags] <branch-name>")
	}
	return Path(repoDir, args[0])
}

func consumeRepoDir(args []string) (string, []string, error) {
	repoDir, filtered, err := cli.ConsumeStringFlag(args, repoDirFlag)
	if err != nil {
		return "", nil, err
	}
	if repoDir == "" {
		repoDir = "."
	}
	return repoDir, filtered, nil
}
