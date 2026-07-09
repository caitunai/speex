#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
mkdir -p "$ROOT_DIR/dist"
cp "$ROOT_DIR/src/index.js" "$ROOT_DIR/dist/index.js"
cp "$ROOT_DIR/src/index.d.ts" "$ROOT_DIR/dist/index.d.ts"
