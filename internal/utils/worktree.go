package utils

import (
	"path/filepath"
	"strings"
)

// WorktreeEntry represents a git worktree entry.
type WorktreeEntry struct {
	Path   string
	Branch string
}

// WorktreeListEntries returns all worktree entries for a repo.
func WorktreeListEntries(repoRoot string) ([]WorktreeEntry, error) {
	output, err := CaptureCommandInDir(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return WorktreeParseEntries(output), nil
}

// WorktreeParseEntries parses git worktree list --porcelain output into entries.
func WorktreeParseEntries(output string) []WorktreeEntry {
	var entries []WorktreeEntry
	var current *WorktreeEntry

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
			current = &WorktreeEntry{Path: strings.TrimPrefix(line, "worktree ")}
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

// WorktreeMainPath returns the path of the main worktree (on the default
// branch) in the preferred layout, or the first entry as a fallback.
func WorktreeMainPath(entries []WorktreeEntry, repoName string) string {
	for _, e := range entries {
		if WorktreeIsMainPath(e.Path, repoName) {
			return e.Path
		}
	}
	if len(entries) == 0 {
		return ""
	}
	return entries[0].Path
}

// WorktreeIsMainPath checks whether a worktree path is the main worktree
// in the preferred layout (i.e., <repo>/<default-branch>/<repo>).
func WorktreeIsMainPath(path string, repoName string) bool {
	return filepath.Base(filepath.Dir(filepath.Dir(path))) == repoName
}

// WorktreeIsPreferredLayoutPath checks whether a worktree path follows the
// preferred layout for non-default branches (i.e., <repo>/worktrees/<branch>/<repo>).
func WorktreeIsPreferredLayoutPath(path string, repoName string) bool {
	if filepath.Base(path) != repoName {
		return false
	}
	parentDir := filepath.Dir(path)
	if filepath.Base(filepath.Dir(parentDir)) != "worktrees" {
		return false
	}
	return filepath.Base(filepath.Dir(filepath.Dir(parentDir))) == repoName
}
