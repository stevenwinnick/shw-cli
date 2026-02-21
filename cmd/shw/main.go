package main

import (
	"os"

	"shw-cli/internal/cli"
	gitcmd "shw-cli/internal/commands/git"
)

func main() {
	root := cli.RootCommand(gitcmd.Command())
	os.Exit(cli.Run(root, os.Args[1:], os.Stdout, os.Stderr))
}
