package commit

import (
	"fmt"

	"shw-cli/internal/cli"
	"shw-cli/internal/utils"
)

const (
	shwName         = "Steven Winnick"
	shwEmail        = "112898228+stevenwinnick@users.noreply.github.com"
	defaultEmptyMsg = "Empty commit"
)

func Command() *cli.Command {
	commit := &cli.Command{
		Name:        "commit",
		Summary:     "Git commit helpers",
		Description: "Commands for creating commits",
		Usage:       "shw git commit <command>",
	}

	commit.AddChild(&cli.Command{
		Name:        "emptyasshw",
		Summary:     "Create an empty commit authored by Steven Winnick",
		Description: fmt.Sprintf("Runs `git commit --allow-empty` with the author and committer set to %s <%s>", shwName, shwEmail),
		Usage:       "shw git commit emptyasshw [<message>]",
		Run:         runEmptyAsShw,
	})

	return commit
}

func runEmptyAsShw(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: shw git commit emptyasshw [<message>]")
	}

	message := defaultEmptyMsg
	if len(args) == 1 {
		message = args[0]
	}

	return utils.RunCommand(
		"git",
		"-c", "user.name="+shwName,
		"-c", "user.email="+shwEmail,
		"commit",
		"--allow-empty",
		"-m", message,
	)
}
