#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TARGET="$INSTALL_DIR/shw"

TMP_BIN="$(mktemp "$REPO_ROOT/shw.XXXXXX")"
trap 'rm -f "$TMP_BIN"' EXIT

cd "$REPO_ROOT"
go build -o "$TMP_BIN" ./cmd/shw

if [[ -w "$INSTALL_DIR" ]]; then
  install -m 0755 "$TMP_BIN" "$TARGET"
else
  sudo install -m 0755 "$TMP_BIN" "$TARGET"
fi

echo "Installed shw to $TARGET"
