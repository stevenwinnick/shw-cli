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
shw git worktree create /path/to/repo steven/add-worktree-commands
shw git worktree create --quiet /path/to/repo steven/add-worktree-commands
shw git worktree list /path/to/repo
shw git worktree remove /path/to/repo steven/add-worktree-commands
shw git worktree clean-all /path/to/repo
cd "$(shw git worktree path /path/to/repo steven--add-worktree-commands)"
```

Use `path` to resolve an existing worktree path. Use `create --quiet` when you want only the created path without extra output

## Development

### Running Locally

To run the CLI locally, run `go run ./cmd/shw/main.go` from the root of the repository.

### Testing

Before pushing changes, run the full test suite:

```bash
go test ./...
```
