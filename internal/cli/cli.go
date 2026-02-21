package cli

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Command struct {
	Name        string
	Summary     string
	Description string
	Usage       string
	Notes       []string
	Children    []*Command
	Run         func(args []string) error
	parent      *Command
}

func RootCommand(children ...*Command) *Command {
	root := &Command{
		Name:        "shw",
		Summary:     "Small helper CLI for Steven's personal workflows",
		Description: "Small helper CLI for Steven's personal workflows",
		Usage:       "shw <command>",
	}

	for _, child := range children {
		root.AddChild(child)
	}

	return root
}

func Run(root *Command, args []string, stdout io.Writer, stderr io.Writer) int {
	if err := execute(root, args, stdout); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func execute(cmd *Command, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		if cmd.Run != nil && len(cmd.Children) == 0 {
			return cmd.Run(nil)
		}
		printHelp(cmd, stdout)
		return nil
	}

	if isHelpArg(args[0]) {
		printHelp(cmd, stdout)
		return nil
	}

	for _, child := range cmd.Children {
		if child.Name == args[0] {
			return execute(child, args[1:], stdout)
		}
	}

	if cmd.Run != nil {
		return cmd.Run(args)
	}

	printHelp(cmd, stdout)
	return fmt.Errorf("unknown subcommand %q for %s", args[0], cmd.Path())
}

func (c *Command) AddChild(child *Command) {
	child.parent = c
	c.Children = append(c.Children, child)
}

func (c *Command) Path() string {
	parts := []string{c.Name}
	for p := c.parent; p != nil; p = p.parent {
		parts = append([]string{p.Name}, parts...)
	}
	return strings.Join(parts, " ")
}

func printHelp(cmd *Command, stdout io.Writer) {
	fmt.Fprintf(stdout, "%s\n\n", cmd.Description)
	fmt.Fprintln(stdout, "Usage:")
	fmt.Fprintf(stdout, "  %s\n", cmd.Usage)

	if len(cmd.Children) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Commands:")
		w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
		for _, child := range cmd.Children {
			fmt.Fprintf(w, "  %s\t%s\n", child.Name, child.Summary)
		}
		_ = w.Flush()
	}

	if len(cmd.Notes) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Notes:")
		for _, line := range cmd.Notes {
			fmt.Fprintf(stdout, "  %s\n", line)
		}
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Flags:")
	fmt.Fprintln(stdout, "  -h, --help  Show help for this command")
}

func isHelpArg(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}
