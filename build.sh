#!/bin/bash
set -e

# Get the current git tag and commit hash
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse HEAD)
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Inject them into the binary
LDFLAGS="-X 'expe3000/internal/version.Version=$VERSION' \
         -X 'expe3000/internal/version.GitCommit=$COMMIT' \
         -X 'expe3000/internal/version.BuildTime=$BUILD_TIME'"

echo "Building expe3000 and expe3000-gui with version $VERSION ($COMMIT)..."

# Build CLI version
go build -ldflags "$LDFLAGS" -o expe3000 ./cmd/expe3000

# Build GUI version
go build -ldflags "$LDFLAGS" -o expe3000-gui ./cmd/expe3000-gui

echo "Done."
