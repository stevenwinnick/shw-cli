# shw-cli

Lightweight Golang CLI for Steven's personal shell helpers

## Installation

To install or update the binary, run the installer script:

```bash
./scripts/install.sh
```

## Shell Setup

`shw pd` prints directory paths for repos and worktrees. To turn it into a `cd` command, add this shell function:

```bash
shwcd() { cd "$(shw pd "$@")" }
```

Then use it to navigate:

```bash
shwcd repo <repo-name>     # Go to a repo's container directory
shwcd default              # Go to the default branch worktree
shwcd <branch-name>        # Go to a branch's worktree
```

## Help

Use `-h` or `--help` at any level to inspect available commands and usage

```bash
shw -h
```

## Development

### Running Locally

To run the CLI locally, run `go run ./cmd/shw/main.go` from the root of the repository.

### Testing

Before pushing changes, run the full test suite:

```bash
go test ./...
```
