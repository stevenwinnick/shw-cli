package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(name string, args ...string) error {
	return RunCommandInDir("", name, args...)
}

func RunCommandInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)
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
