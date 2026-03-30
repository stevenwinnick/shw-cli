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

## Development

### Running Locally

To run the CLI locally, run `go run ./cmd/shw/main.go` from the root of the repository.

### Testing

Before pushing changes, run the full test suite:

```bash
go test ./...
```
