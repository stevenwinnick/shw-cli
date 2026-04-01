package repo

import (
	"path/filepath"
	"strings"

	"shw-cli/internal/utils"
)

// Entry represents a git worktree entry.
type Entry struct {
	Path   string
	Branch string
}

// ListEntries returns all worktree entries for a repo.
func ListEntries(repoRoot string) ([]Entry, error) {
	output, err := utils.CaptureCommandInDir(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return ParseEntries(output), nil
}

// ParseEntries parses git worktree list --porcelain output into entries.
func ParseEntries(output string) []Entry {
	var entries []Entry
	var current *Entry

	for _, line := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &Entry{Path: strings.TrimPrefix(line, "worktree ")}
			continue
		}
		if current != nil && strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries
}

// MainWorktreePath returns the path of the main worktree (on the default
// branch) in the preferred layout, or the first entry as a fallback.
func MainWorktreePath(entries []Entry, repoName string) string {
	for _, e := range entries {
		if IsMainWorktreePath(e.Path, repoName) {
			return e.Path
		}
	}
	if len(entries) == 0 {
		return ""
	}
	return entries[0].Path
}

// IsMainWorktreePath checks whether a worktree path is the main worktree
// in the preferred layout (i.e., <repo>/<default-branch>/<repo>).
func IsMainWorktreePath(path string, repoName string) bool {
	return filepath.Base(filepath.Dir(filepath.Dir(path))) == repoName
}

// IsPreferredLayoutWorktreePath checks whether a worktree path follows the
// preferred layout for non-default branches (i.e., <repo>/worktrees/<branch>/<repo>).
func IsPreferredLayoutWorktreePath(path string, repoName string) bool {
	if filepath.Base(path) != repoName {
		return false
	}
	parentDir := filepath.Dir(path)
	if filepath.Base(filepath.Dir(parentDir)) != "worktrees" {
		return false
	}
	return filepath.Base(filepath.Dir(filepath.Dir(parentDir))) == repoName
}
