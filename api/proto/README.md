# FFI protobuf

The Mac (and later Windows) client talks to the Go core over a **C ABI** whose payloads are Protocol Buffers, not JSON.

## Schema

[`engine.proto`](engine.proto) is the source of truth. Generated Go lives in [`engine/`](engine/); generated Swift in [`macos/App/Platform/Generated/`](../../macos/App/Platform/Generated/). Commit both after regenerating so `go test` and Xcode do not require `protoc` on every machine.

## Plugins

| Tool | Role | Typical install |
| --- | --- | --- |
| `protoc` | Compiler | `brew install protobuf` |
| `protoc-gen-go` | Go types | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| `protoc-gen-swift` | Swift types | `brew install swift-protobuf` |

Put `$(go env GOPATH)/bin` on `PATH` so `protoc` finds `protoc-gen-go`.

Regenerate:

```sh
./scripts/generate-proto.sh
```

## C ABI

A single pair of exports (see `api/libprovenance`):

- `provenance_call(method, in, in_len, out, out_len)` — `method` is `provenance.engine.v1.Method`; on success `in`/`out` are protobuf bytes for that RPC. On failure (status `1`) `out` is UTF-8 `err.Error()`, not protobuf.
- `provenance_free` — caller frees `out`.

Do not add per-field C getters.

## Rebuild the Mac library

SQLite (`mattn/go-sqlite3`) is compiled into this dylib. Identity and session RPCs take directory paths in protobuf (caller supplies Application Support / Documents; Go does not hardcode them). There are no Source tables.

Building the Mac app (⌘R or `xcodebuild`) runs [`scripts/build-macos-core.sh`](../../scripts/build-macos-core.sh). You can run that script from a terminal too. It writes a universal (`arm64` + `x86_64`) `macos/Core/libprovenance.dylib` with install name `@rpath/libprovenance.dylib` and `MACOSX_DEPLOYMENT_TARGET=14.0`. **Do not commit the dylib.**
