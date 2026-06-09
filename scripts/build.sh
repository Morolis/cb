#!/usr/bin/env bash
# Legacy build script. Prefer: make cross-compile
# For releases: goreleaser release --snapshot --clean
set -euo pipefail
cd "$(dirname "$0")/.."
exec make cross-compile
