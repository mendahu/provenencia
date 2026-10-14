#!/usr/bin/env bash
# Builds macos/Core/libprovenencia.dylib for the Xcode app target.
# Xcode runs this phase on every build (alwaysOutOfDate); the freshness check
# below decides whether to rebuild, so there is no input list to maintain.
# It looks at every file under api/ and core/, which covers go:embed assets
# such as core/database/migrations/*.sql.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p macos/Core
out="macos/Core/libprovenencia.dylib"

if [[ -f "$out" ]] \
  && [[ ! go.mod -nt "$out" ]] \
  && [[ ! go.sum -nt "$out" ]] \
  && [[ ! scripts/build-macos-core.sh -nt "$out" ]] \
  && [[ -z "$(find api core -type f -newer "$out" -print -quit)" ]]; then
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

export CGO_ENABLED=1
# The app's deployment target. Xcode exports it to this build phase; the
# fallback matches the project for builds run by hand.
export MACOSX_DEPLOYMENT_TARGET="${MACOSX_DEPLOYMENT_TARGET:-14.0}"
# fts5: catalog omnibar projection (go-sqlite3 amalgamation opt-in).
GOARCH=arm64 go build -tags fts5 -buildmode=c-shared -o "$tmp/libprovenencia_arm64.dylib" ./api/libprovenencia
GOARCH=amd64 go build -tags fts5 -buildmode=c-shared -o "$tmp/libprovenencia_amd64.dylib" ./api/libprovenencia
lipo -create -output "$out" \
  "$tmp/libprovenencia_arm64.dylib" \
  "$tmp/libprovenencia_amd64.dylib"
install_name_tool -id @rpath/libprovenencia.dylib "$out"
