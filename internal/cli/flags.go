package cli

import "strings"

type Flag struct {
	Long        string
	Short       string
	Description string
}

func (f Flag) Label() string {
	switch {
	case f.Short != "" && f.Long != "":
		return "-" + f.Short + ", --" + f.Long
	case f.Long != "":
		return "--" + f.Long
	case f.Short != "":
		return "-" + f.Short
	default:
		return ""
	}
}

func ConsumeBoolFlag(args []string, flag Flag) (bool, []string) {
	consumed := false
	filtered := make([]string, 0, len(args))
	parsingFlags := true

	for _, arg := range args {
		if parsingFlags && arg == "--" {
			parsingFlags = false
			filtered = append(filtered, arg)
			continue
		}
		if parsingFlags && flag.matches(arg) {
			consumed = true
			continue
		}

		filtered = append(filtered, arg)
	}

	return consumed, filtered
}

func (f Flag) matches(arg string) bool {
	return (f.Long != "" && arg == "--"+f.Long) || (f.Short != "" && arg == "-"+f.Short)
}

func helpFlag() Flag {
	return Flag{
		Long:        "help",
		Short:       "h",
		Description: "Show help for this command",
	}
}

func allFlags(cmd *Command) []Flag {
	flags := make([]Flag, 0, len(cmd.Flags)+1)
	flags = append(flags, cmd.Flags...)
	flags = append(flags, helpFlag())
	return flags
}

func formatFlagLabel(flag Flag) string {
	return strings.TrimSpace(flag.Label())
}
