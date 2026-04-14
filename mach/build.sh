#!/bin/bash
set -e

# Build script for mach

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Get version info
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "Building mach v${VERSION} (${COMMIT})..."

# Build flags
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"
LDFLAGS="${LDFLAGS} -s -w"  # Strip debug info for smaller binary

# Create build directory
mkdir -p build

# Build for current platform
echo "Building for current platform..."
go build -ldflags "${LDFLAGS}" -o build/mach .

# Also build for common platforms
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64"

for platform in $PLATFORMS; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    output="build/mach-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building for ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "${LDFLAGS}" -o "$output" . 2>/dev/null || echo "  (skipped - may require cross-compilation setup)"
done

echo ""
echo "Build complete. Binaries in ./build/"
echo "Run: ./build/mach"
