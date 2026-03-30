package worktree

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"shw-cli/internal/utils"
)

type entry struct {
	Path   string
	Branch string
}

func Create(repoDir string, branchName string) error {
	return CreateWithOptions(repoDir, branchName, true)
}

func CreateWithOptions(repoDir string, branchName string, updateDefaultBranch bool) error {
	return create(repoDir, branchName, updateDefaultBranch, os.Stdout, os.Stderr)
}

func create(repoDir string, branchName string, updateDefaultBranch bool, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree create [flags] <branch-name>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	containerDir, err := repoContainerDir(repoRoot, repoName)
	if err != nil {
		return err
	}

	defaultBranch, err := defaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}
	if updateDefaultBranch {
		if err := updateBranchBaseRef(repoRoot, defaultBranch, stderr); err != nil {
			return err
		}
	}

	baseRef, err := baseRef(repoRoot, defaultBranch)
	if err != nil {
		return err
	}

	worktreeDir := strings.ReplaceAll(branchName, "/", "--")
	worktreePath := filepath.Join(containerDir, "worktrees", worktreeDir, repoName)
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return fmt.Errorf("failed to create worktree parent directory %q: %w", filepath.Dir(worktreePath), err)
	}

	addArgs := []string{"worktree", "add", "-b", branchName, worktreePath, baseRef}
	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", addArgs...); err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "Worktree created: %s\n", worktreePath)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "To switch to it, run: `%s`\n", navigationCommand(repoDir, branchName))
	return err
}

func List(repoDir string) error {
	return list(repoDir, os.Stdout, os.Stderr)
}

func list(repoDir string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree list")
	}

	repoRoot, _, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	return utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "list")
}

func Remove(repoDir string, branchName string) error {
	return remove(repoDir, branchName, os.Stdout, os.Stderr)
}

func remove(repoDir string, branchName string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree remove [flags] <branch-name>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}

	var targetEntry *entry
	for _, entry := range entries {
		if entry.Branch == branchName {
			entryCopy := entry
			targetEntry = &entryCopy
			break
		}
	}

	if targetEntry == nil {
		listOutput, listErr := utils.CaptureCommandInDir(repoRoot, "git", "worktree", "list")
		if listErr != nil {
			return fmt.Errorf("no worktree found for branch %q", branchName)
		}
		return fmt.Errorf("no worktree found for branch %q\n\nAvailable worktrees:\n%s", branchName, strings.TrimSpace(listOutput))
	}

	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "remove", "--force", targetEntry.Path); err != nil {
		return err
	}

	parentDir := filepath.Dir(targetEntry.Path)
	if isPreferredLayoutWorktreePath(targetEntry.Path, repoName) {
		if _, err := os.Stat(parentDir); err == nil {
			if _, err := fmt.Fprintf(stdout, "Removing directory: %s\n", parentDir); err != nil {
				return err
			}
			if err := os.RemoveAll(parentDir); err != nil {
				return fmt.Errorf("failed to remove directory %q: %w", parentDir, err)
			}
		}
	}

	if refExists(repoRoot, "refs/heads/"+branchName) {
		_ = utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "branch", "-D", branchName)
	}

	_, err = fmt.Fprintln(stdout, "Worktree removed successfully")
	return err
}

func CleanAll(repoDir string) error {
	return cleanAll(repoDir, os.Stdout, os.Stderr)
}

func cleanAll(repoDir string, stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(repoDir) == "" {
		return fmt.Errorf("usage: shw git worktree prune-stale [flags]")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "prune"); err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}

	mainWorktree := mainWorktreePath(entries, repoName)
	for _, entry := range entries {
		if entry.Path == mainWorktree || entry.Branch == "" {
			continue
		}

		if refExists(repoRoot, "refs/remotes/origin/"+entry.Branch) {
			continue
		}

		if _, err := fmt.Fprintf(stdout, "Removing worktree %s: branch '%s' not on remote\n", entry.Path, entry.Branch); err != nil {
			return err
		}
		if err := utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "worktree", "remove", "--force", entry.Path); err != nil {
			return err
		}
		if isPreferredLayoutWorktreePath(entry.Path, repoName) {
			if err := os.RemoveAll(filepath.Dir(entry.Path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to remove directory %q: %w", filepath.Dir(entry.Path), err)
			}
		}
		_ = utils.RunCommandInDirWithWriters(repoRoot, stdout, stderr, "git", "branch", "-D", entry.Branch)
	}

	_, err = fmt.Fprintln(stdout, "Worktree cleanup complete")
	return err
}

func Path(repoDir string, name string) error {
	return switchTo(repoDir, name, os.Stdout)
}

func switchTo(repoDir string, branchName string, stdout io.Writer) error {
	if strings.TrimSpace(repoDir) == "" || strings.TrimSpace(branchName) == "" {
		return fmt.Errorf("usage: shw git worktree path [flags] <branch-name>")
	}

	repoRoot, repoName, err := repoDetails(repoDir)
	if err != nil {
		return err
	}

	entries, err := listEntries(repoRoot)
	if err != nil {
		return err
	}
	defaultBranch, err := defaultBranch(repoRoot, repoName)
	if err != nil {
		return err
	}

	availableBranches := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Branch == "" {
			continue
		}
		availableBranches = append(availableBranches, entry.Branch)
		if entry.Branch != branchName {
			continue
		}

		info, err := os.Stat(entry.Path)
		if err != nil {
			return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, entry.Path)
		}
		if !info.IsDir() {
			return fmt.Errorf("worktree path exists but is not a directory: %s", entry.Path)
		}

		_, err = fmt.Fprintln(stdout, entry.Path)
		return err
	}

	if branchName == defaultBranch {
		mainWorktree := mainWorktreePath(entries, repoName)
		if mainWorktree != "" {
			info, err := os.Stat(mainWorktree)
			if err != nil {
				return fmt.Errorf("worktree for branch %q is not accessible at %s", branchName, mainWorktree)
			}
			if !info.IsDir() {
				return fmt.Errorf("worktree path exists but is not a directory: %s", mainWorktree)
			}

			_, err = fmt.Fprintln(stdout, mainWorktree)
			return err
		}
	}

	if !containsString(availableBranches, defaultBranch) {
		availableBranches = append(availableBranches, defaultBranch)
	}
	if len(availableBranches) == 0 {
		return fmt.Errorf("worktree branch %q not found; available: (none)", branchName)
	}

	return fmt.Errorf("worktree branch %q not found; available: %s", branchName, strings.Join(availableBranches, ", "))
}

func repoDetails(repoDir string) (string, string, error) {
	absoluteRepoDir, err := filepath.Abs(repoDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve repo dir %q: %w", repoDir, err)
	}

	if repoRoot, err := resolveGitTopLevel(absoluteRepoDir); err == nil {
		return repoRoot, filepath.Base(repoRoot), nil
	}

	repoRoot, err := resolveRepoRootFromContainer(absoluteRepoDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve git repo from %q: %w", repoDir, err)
	}

	return repoRoot, filepath.Base(repoRoot), nil
}

func resolveGitTopLevel(dir string) (string, error) {
	repoRoot, err := utils.CaptureCommandInDir(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return "", fmt.Errorf("resolved git repo root is empty")
	}

	return repoRoot, nil
}

func resolveRepoRootFromContainer(containerDir string) (string, error) {
	repoName := filepath.Base(containerDir)
	dirEntries, err := os.ReadDir(containerDir)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() || dirEntry.Name() == "worktrees" {
			continue
		}

		candidate := filepath.Join(containerDir, dirEntry.Name(), repoName)
		if _, err := os.Stat(candidate); err != nil {
			continue
		}

		repoRoot, err := resolveGitTopLevel(candidate)
		if err != nil {
			continue
		}

		candidates = append(candidates, repoRoot)
	}

	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("not a git repo or recognized repo container")
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple repo roots found under %q", containerDir)
	}
}

func repoContainerDir(repoRoot string, repoName string) (string, error) {
	parentDir := filepath.Dir(repoRoot)
	grandparentDir := filepath.Dir(parentDir)
	if filepath.Base(grandparentDir) == repoName {
		return grandparentDir, nil
	}

	if filepath.Base(grandparentDir) == "worktrees" {
		containerDir := filepath.Dir(grandparentDir)
		if filepath.Base(containerDir) == repoName {
			return containerDir, nil
		}
	}

	codeRoot := strings.TrimSpace(os.Getenv("CODE_ROOT"))
	if codeRoot != "" {
		return filepath.Join(codeRoot, repoName), nil
	}

	return "", fmt.Errorf("repo %q is not in the preferred worktree layout and CODE_ROOT is not set", repoRoot)
}

func defaultBranch(repoRoot string, repoName string) (string, error) {
	entries, err := listEntries(repoRoot)
	if err != nil {
		return "", err
	}

	mainWorktree := mainWorktreePath(entries, repoName)
	if mainWorktree != "" {
		return filepath.Base(filepath.Dir(mainWorktree)), nil
	}

	remoteHead, err := utils.CaptureCommandInDir(repoRoot, "git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err == nil {
		return strings.TrimPrefix(strings.TrimSpace(remoteHead), "origin/"), nil
	}

	return "", fmt.Errorf("failed to detect default branch for %s", repoRoot)
}

func baseRef(repoRoot string, defaultBranch string) (string, error) {
	remoteRef := "refs/remotes/origin/" + defaultBranch
	if refExists(repoRoot, remoteRef) {
		return "origin/" + defaultBranch, nil
	}

	localRef := "refs/heads/" + defaultBranch
	if refExists(repoRoot, localRef) {
		return defaultBranch, nil
	}

	return "", fmt.Errorf("could not find a base ref for default branch %q", defaultBranch)
}

func updateBranchBaseRef(repoRoot string, defaultBranch string, stderr io.Writer) error {
	if !refExists(repoRoot, "refs/remotes/origin/"+defaultBranch) {
		return nil
	}

	return utils.RunCommandInDirWithWriters(repoRoot, io.Discard, stderr, "git", "fetch", "origin", defaultBranch)
}

func refExists(repoRoot string, ref string) bool {
	_, err := utils.CaptureCommandInDir(repoRoot, "git", "show-ref", "--verify", ref)
	return err == nil
}

func listEntries(repoRoot string) ([]entry, error) {
	output, err := utils.CaptureCommandInDir(repoRoot, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	return parseEntries(output), nil
}

func parseEntries(output string) []entry {
	var entries []entry
	var current *entry

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
			current = &entry{Path: strings.TrimPrefix(line, "worktree ")}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasPrefix(line, "branch refs/heads/") {
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}

func mainWorktreePath(entries []entry, repoName string) string {
	for _, entry := range entries {
		if isMainWorktreePath(entry.Path, repoName) {
			return entry.Path
		}
	}
	if len(entries) == 0 {
		return ""
	}
	return entries[0].Path
}

func isMainWorktreePath(path string, repoName string) bool {
	parentDir := filepath.Dir(path)
	return filepath.Base(filepath.Dir(parentDir)) == repoName
}

func isPreferredLayoutWorktreePath(path string, repoName string) bool {
	parentDir := filepath.Dir(path)
	if filepath.Base(path) != repoName {
		return false
	}
	if filepath.Base(filepath.Dir(parentDir)) != "worktrees" {
		return false
	}

	return filepath.Base(filepath.Dir(filepath.Dir(parentDir))) == repoName
}

func navigationCommand(repoDir string, branchName string) string {
	pathArgs := []string{"shw", "git", "worktree", "path"}
	if repoDir != "." {
		pathArgs = append(pathArgs, "--repo-dir", shellQuote(repoDir))
	}
	pathArgs = append(pathArgs, shellQuote(branchName))

	return fmt.Sprintf("cd $(%s)", strings.Join(pathArgs, " "))
}

func shellQuote(value string) string {
	if value == "" {
		return `""`
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return !(r == '-' || r == '_' || r == '.' || r == '/' || r == ':' || r == '@' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) == -1 {
		return value
	}

	return strconv.Quote(value)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
