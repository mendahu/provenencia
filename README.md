# Provenance

Local-first genealogy software. Native macOS client over a shared Go core. Domain and stack notes live in [`docs/`](docs/). Mac UI patterns: [`docs/macos-client-patterns.md`](docs/macos-client-patterns.md). Product versioning: [`docs/versioning.md`](docs/versioning.md). Implementation milestones: [`docs/deployment-plan/`](docs/deployment-plan/). FFI protobuf: [`api/proto/README.md`](api/proto/README.md).

**Requires macOS 14 (Sonoma) or later** and **Go 1.27** or later (the Xcode project builds `libprovenance.dylib` on compile).

## Run the Mac app

1. Install Xcode (this repo was last built with Xcode 26; any current Xcode that can target macOS 14 is fine) and Go 1.27+.
2. Open **[`macos/Provenance.xcodeproj`](macos/Provenance.xcodeproj)** (the file that ends in `.xcodeproj`). Do not open the `macos/App` folder; that is source, not the Xcode project. From Terminal: `open macos/Provenance.xcodeproj`
3. Select the **Provenance** scheme and **My Mac**.
4. Run (⌘R). The **Build Go core** phase runs [`scripts/build-macos-core.sh`](scripts/build-macos-core.sh) (skipped when the dylib is already newer than Go sources). First launch asks whether you already have a project. Creating one writes `identity.json`, `active-project.json`, and `{Name}.provenance` under Documents. Opening an existing folder lists contributors so you can keep the same ID (new Mac) or add a new one. A later launch skips the picker if that active project is still on disk.

From the command line:

```sh
xcodebuild -project macos/Provenance.xcodeproj -scheme Provenance -destination 'platform=macOS' build
xcodebuild test -project macos/Provenance.xcodeproj -scheme Provenance -destination 'platform=macOS'
```

The first `xcodebuild` resolves the SwiftProtobuf package from GitHub. `macos/Core/libprovenance.dylib` is a local build artifact; do not commit it. You can still run `./scripts/build-macos-core.sh` from a terminal.

## Go core

The Go module is `github.com/mendahu/provenance`. It requires **Go 1.27** or later.

```sh
go test ./api/... ./core/...
```

PRs to `main` that are **not drafts** run this in GitHub Actions (`.github/workflows/go-test.yml`). The same PRs run Swift tests on a macOS runner (`.github/workflows/macos-test.yml`).

Regenerate protobuf (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`):

```sh
./scripts/generate-proto.sh
```
