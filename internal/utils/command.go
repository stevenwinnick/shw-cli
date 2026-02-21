package utils

import (
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
