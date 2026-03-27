package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func PromptRelativeDir(prompt string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return ResolveRelativeDir(strings.TrimSpace(input))
}

func ResolveRelativeDir(relativePath string) (string, error) {
	cleaned := strings.TrimSpace(relativePath)
	if cleaned == "" {
		return "", fmt.Errorf("directory path is required")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("directory path must be relative")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Join(cwd, filepath.Clean(cleaned)), nil
}

func EnsureDir(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to access directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}

	return nil
}

func EnsureExistingDir(dir string) error {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	if err != nil {
		return fmt.Errorf("failed to access directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}

	return nil
}
