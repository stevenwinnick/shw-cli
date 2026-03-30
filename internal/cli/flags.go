package cli

import "strings"

type Flag struct {
	Long        string
	Short       string
	ValueName   string
	Description string
}

func (f Flag) Label() string {
	label := ""
	switch {
	case f.Short != "" && f.Long != "":
		label = "-" + f.Short + ", --" + f.Long
	case f.Long != "":
		label = "--" + f.Long
	case f.Short != "":
		label = "-" + f.Short
	}

	if label == "" {
		return ""
	}
	if f.ValueName == "" {
		return label
	}

	return label + " " + f.ValueName
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

func ConsumeStringFlag(args []string, flag Flag) (string, []string, error) {
	filtered := make([]string, 0, len(args))
	parsingFlags := true

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if parsingFlags && arg == "--" {
			parsingFlags = false
			filtered = append(filtered, arg)
			continue
		}
		if !parsingFlags {
			filtered = append(filtered, arg)
			continue
		}

		if value, matched := consumeStringFlagValue(arg, flag); matched {
			if value != "" {
				return value, append(filtered, args[i+1:]...), nil
			}
			if i+1 >= len(args) || args[i+1] == "--" {
				return "", nil, missingFlagValueError(flag)
			}
			return args[i+1], append(filtered, args[i+2:]...), nil
		}

		filtered = append(filtered, arg)
	}

	return "", filtered, nil
}

func (f Flag) matches(arg string) bool {
	return (f.Long != "" && arg == "--"+f.Long) || (f.Short != "" && arg == "-"+f.Short)
}

func consumeStringFlagValue(arg string, flag Flag) (string, bool) {
	if flag.Long != "" {
		prefix := "--" + flag.Long + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix), true
		}
	}
	if flag.Short != "" {
		prefix := "-" + flag.Short + "="
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix), true
		}
	}

	return "", flag.matches(arg)
}

func missingFlagValueError(flag Flag) error {
	valueName := flag.ValueName
	if valueName == "" {
		valueName = "<value>"
	}

	return &flagValueError{message: "missing value for " + flag.Label() + "; expected " + valueName}
}

type flagValueError struct {
	message string
}

func (e *flagValueError) Error() string {
	return e.message
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
