#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

LD_FLAGS=$(printf '%s' \
  "-X github.com/pidisk/pidisk/internal/version.Version=${VERSION} " \
  "-X github.com/pidisk/pidisk/internal/version.Commit=${COMMIT} " \
  "-X github.com/pidisk/pidisk/internal/version.BuildTime=${BUILD_TIME}")

wails build -clean -platform "linux/amd64" -ldflags "${LD_FLAGS}"
