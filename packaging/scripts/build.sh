#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-0.1.0}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
DIST="$ROOT_DIR/dist"

rm -rf "$DIST"
mkdir -p "$DIST"

COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE"

build() {
    local goos=$1 goarch=$2 suffix=$3
    local ext=""
    [ "$goos" = "windows" ] && ext=".exe"

    echo "Building $goos/$goarch..."
    GOOS=$goos GOARCH=$goarch go build -trimpath -ldflags "$LDFLAGS" \
        -o "$DIST/cb-${goos}-${goarch}${ext}" .

    if [ "$goos" = "windows" ]; then
        (cd "$DIST" && zip -q "cb-windows-${goarch}.zip" "cb-windows-${goarch}${ext}")
    else
        tar -czf "$DIST/cb-${goos}-${goarch}.tar.gz" -C "$DIST" "cb-${goos}-${goarch}"
    fi
}

cd "$ROOT_DIR"

build linux   amd64
build linux   arm64
build darwin  amd64
build darwin  arm64
build windows amd64

# .deb package (requires nfpm)
if command -v nfpm &>/dev/null; then
    VERSION=$VERSION nfpm package -f "$SCRIPT_DIR/../linux/nfpm.yaml" -p deb -t "$DIST/"
    echo "Built .deb package"
else
    echo "Skipping .deb (nfpm not installed)"
fi

# Inno Setup installer
# 1. Check local directory
# 2. Check system PATH
INNO_DIR="$SCRIPT_DIR/../windows/inno"
ISCC=""
if [ -f "$INNO_DIR/ISCC.exe" ]; then
    ISCC="$INNO_DIR/ISCC.exe"
elif command -v iscc &>/dev/null; then
    ISCC="iscc"
fi

if [ -n "$ISCC" ]; then
    # Inno Setup is Windows-only, use Wine if available
    if command -v wine &>/dev/null; then
        wine "$ISCC" //DAppVersion="$VERSION" "$SCRIPT_DIR/../windows/cb.iss" >/dev/null
        echo "Built Windows installer"
    elif command -v iscc &>/dev/null; then
        iscc //DAppVersion="$VERSION" "$SCRIPT_DIR/../windows/cb.iss" >/dev/null
        echo "Built Windows installer"
    else
        echo "Skipping Windows installer (Wine not available to run ISCC)"
    fi
else
    echo "Skipping Windows installer (ISCC not found in packaging/windows/inno/ or PATH)"
fi

# Checksums
(cd "$DIST" && sha256sum *.tar.gz *.zip *.deb 2>/dev/null > "checksums-${VERSION}.txt" || true)

echo ""
echo "Build complete. Artifacts in $DIST:"
ls -lh "$DIST"/*.{tar.gz,zip,deb,exe} 2>/dev/null || true
