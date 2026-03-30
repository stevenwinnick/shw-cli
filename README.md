# shw-cli

Lightweight Golang CLI for Steven's personal shell helpers

## Installation

To install or update the binary, run the installer script:

```bash
./scripts/install.sh
```

## Help

Use `-h` or `--help` at any level to inspect available commands and usage

```bash
shw -h
```

## Worktrees

Manage worktrees in the preferred repo layout with `shw git worktree`

```bash
shw git worktree create steven/add-worktree-commands
shw git worktree create --repo-dir /path/to/repo steven/add-worktree-commands
shw git worktree list
shw git worktree remove steven/add-worktree-commands
shw git worktree prune-stale
cd "$(shw git worktree path steven/add-worktree-commands)"
cd "$(shw git worktree path --repo-dir /path/to/repo steven/add-worktree-commands)"
```

Commands use the current directory as the repo by default. Pass `--repo-dir <repo-dir>` to target a different repo

Use `path` to resolve a worktree by branch name

## Development

### Running Locally

To run the CLI locally, run `go run ./cmd/shw/main.go` from the root of the repository.

### Testing

Before pushing changes, run the full test suite:

```bash
go test ./...
```
