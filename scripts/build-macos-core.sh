#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p macos/Core
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

export CGO_ENABLED=1
export MACOSX_DEPLOYMENT_TARGET=14.0
GOARCH=arm64 go build -buildmode=c-shared -o "$tmp/libprovenance_arm64.dylib" ./api/libprovenance
GOARCH=amd64 go build -buildmode=c-shared -o "$tmp/libprovenance_amd64.dylib" ./api/libprovenance
lipo -create -output macos/Core/libprovenance.dylib \
  "$tmp/libprovenance_arm64.dylib" \
  "$tmp/libprovenance_amd64.dylib"
install_name_tool -id @rpath/libprovenance.dylib macos/Core/libprovenance.dylib
