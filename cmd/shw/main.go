package main

import (
	"os"

	"shw-cli/internal/cli"
	gitcmd "shw-cli/internal/commands/git"
	updatecmd "shw-cli/internal/commands/update"
)

func main() {
	root := cli.RootCommand(gitcmd.Command(), updatecmd.Command())
	os.Exit(cli.Run(root, os.Args[1:], os.Stdout, os.Stderr))
}
