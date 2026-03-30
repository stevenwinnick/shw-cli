package worktree

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var copyTextToClipboard = systemClipboardCopy

func systemClipboardCopy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if strings.TrimSpace(stderr.String()) == "" {
			return fmt.Errorf("clipboard command %q failed: %w", "pbcopy", err)
		}
		return fmt.Errorf("clipboard command %q failed: %w: %s", "pbcopy", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
