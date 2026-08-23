---
name: add-ffi-handler
description: >-
  Adds a Provenance FFI RPC handler under api/ffi/handlers and tests it with runRPC.
  Use when adding or changing engine.proto Method, ffi.Call cases, handler functions,
  GetInstallIdentity/CompleteOnboarding-style RPCs, or handler table tests.
---

# Add an FFI handler

[`api/ffi/dispatch.go`](api/ffi/dispatch.go) is **only** a method switch. Unmarshal, domain calls, and protobuf replies live in **`api/ffi/handlers/`**. `handlers` must not import `ffi` (cycle).

## Layout

1. Add the enum value and messages in [`api/proto/engine.proto`](api/proto/engine.proto). Run `./scripts/generate-proto.sh` and commit generated Go + Swift.
2. Add `Method…` next to the other constants in `dispatch.go` and one `case` that calls `handlers.<Name>`.
3. Implement `func <Name>(in []byte) ([]byte, error)` in `api/ffi/handlers/<name>.go` (copy ping/identity/onboarding).
4. Test in `api/ffi/handlers/<name>_test.go` with **`runRPC`** from `harness_test.go`. Do **not** put a large `run func(*testing.T)` in the table.

`dispatch_test.go` stays router-only (unspecified / unknown method).

After C ABI or handler behavior changes that ship in the dylib: `./scripts/build-macos-core.sh`.

## Handler tests (`runRPC`)

```go
func TestFoo(t *testing.T) {
	runRPC(t, Foo, []rpcTest{
		{name: "ok", req: &engine.FooRequest{…}, want: &engine.FooResponse{…}},
		{name: "bad proto", raw: []byte{0xff, 0xff, 0xff, 0xff}, wantErr: true},
	})
}
```

| Field | Use |
| --- | --- |
| `req` | Static protobuf request |
| `reqFn` | Request that needs `t.TempDir()` / fixtures (`saveIdentity`) |
| `raw` | Bytes that are not a valid message |
| `calls` | Repeat the same input (default 1). `wantErr`/`want`/`after` apply to the **last** call |
| `wantErr` | Last call must fail |
| `want` | Populated proto3 fields must match; omit generated ids |
| `exact` | `proto.Equal` including zeros (empty ping echo) |
| `after` | Disk / follow-up RPC only; use harness helpers (`assertOnboardingWroteProject`, `assertIdentityMatchesComplete`, `assertIdentityNotFound`) |

Do not unmarshal in the table row. Add a `t.Helper()` in `harness_test.go` if a side effect will be reused.

Run `CGO_ENABLED=1 go test ./api/ffi/...`.

## Do not

- Grow `dispatch.go` with unmarshal or `core/` imports
- Add handler tests under `api/ffi/` itself (except router tests)
- Hardcode Application Support / Documents in Go
