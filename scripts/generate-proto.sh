#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export PATH="${PATH}:$(go env GOPATH)/bin"

if ! command -v protoc >/dev/null; then
  echo "protoc is required (e.g. brew install protobuf)" >&2
  exit 1
fi
if ! command -v protoc-gen-go >/dev/null; then
  echo "protoc-gen-go is required: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" >&2
  exit 1
fi
if ! command -v protoc-gen-swift >/dev/null; then
  echo "protoc-gen-swift is required (e.g. brew install swift-protobuf)" >&2
  exit 1
fi

mkdir -p api/proto/engine macos/Provenance/Generated

protoc \
  --proto_path=api/proto \
  --go_out=api/proto/engine \
  --go_opt=paths=source_relative \
  --swift_out=macos/Provenance/Generated \
  --swift_opt=Visibility=Public \
  api/proto/engine.proto
