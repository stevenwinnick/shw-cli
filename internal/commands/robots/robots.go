package robots

import (
	"fmt"

	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

const robotsUser = "stevenwinnickrobots"

func Command() *cli.Command {
	robots := &cli.Command{
		Name:        "robots",
		Summary:     "Helpers for running AI agents",
		Description: "Commands for running AI agents as the sandboxed robots user",
		Usage:       "shw robots <command>",
	}

	robots.AddChild(&cli.Command{
		Name:        "start",
		Summary:     "Open a login shell as the robots user",
		Description: fmt.Sprintf("Runs `sudo -u %s -i` so agents run as the sandboxed robots user instead of the current user", robotsUser),
		Usage:       "shw robots start",
		Run:         runStart,
	})

	robots.AddChild(&cli.Command{
		Name:        "run",
		Summary:     "Run a single command as the robots user",
		Description: fmt.Sprintf("Runs `sudo -u %s -i -- <command>`, exiting with the command's own exit code", robotsUser),
		Usage:       "shw robots run -- <command>",
		Run:         runRun,
	})

	return robots
}

func runStart(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: shw robots start")
	}

	return utils.RunCommand("sudo", "-u", robotsUser, "-i")
}

func runRun(args []string) error {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: shw robots run -- <command>")
	}

	// `-i` runs the command through the robots user's login shell so it gets that user's environment
	sudoArgs := append([]string{"-u", robotsUser, "-i", "--"}, args...)
	return utils.RunCommand("sudo", sudoArgs...)
}
