package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// CommandExitError reports a command that ran with its output streamed to the terminal and exited
// non-zero. The command has already written its own error output, so callers can propagate ExitCode
// without reporting the error again.
type CommandExitError struct {
	ExitCode int
	err      error
}

func (e *CommandExitError) Error() string {
	return e.err.Error()
}

func (e *CommandExitError) Unwrap() error {
	return e.err
}

func RunCommand(name string, args ...string) error {
	return RunCommandInDir("", name, args...)
}

func RunCommandInDir(dir string, name string, args ...string) error {
	return RunCommandInDirWithWriters(dir, os.Stdout, os.Stderr, name, args...)
}

func RunCommandInDirWithWriters(dir string, stdout io.Writer, stderr io.Writer, name string, args ...string) error {
	PrintCommand(stderr, name, args...)

	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		failure := fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() > 0 {
			return &CommandExitError{ExitCode: exitErr.ExitCode(), err: failure}
		}

		return failure
	}

	return nil
}

func CaptureCommand(name string, args ...string) (string, error) {
	return CaptureCommandInDir("", name, args...)
}

func CaptureCommandInDir(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"command failed: %s %s: %w%s%s",
			name,
			strings.Join(args, " "),
			err,
			formatCapturedOutput("stdout", stdout.String()),
			formatCapturedOutput("stderr", stderr.String()),
		)
	}

	return stdout.String(), nil
}

func formatCapturedOutput(label string, output string) string {
	if strings.TrimSpace(output) == "" {
		return ""
	}

	return fmt.Sprintf("; %s: %s", label, strings.TrimSpace(output))
}

func PrintCommand(w io.Writer, name string, args ...string) {
	if w == nil {
		return
	}

	_, _ = fmt.Fprintf(w, "Executing command: %s\n", formatCommand(name, args...))
}

func formatCommand(name string, args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteCommandArg(name))
	for _, arg := range args {
		parts = append(parts, quoteCommandArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteCommandArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if strings.IndexFunc(arg, func(r rune) bool {
		return !(r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r == '@' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) == -1 {
		return arg
	}

	return strconv.Quote(arg)
}
