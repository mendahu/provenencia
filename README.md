# Provenance

Local-first genealogy software. Native macOS client over a shared Go core. Domain and stack notes live in [`docs/`](docs/). Mac UI patterns: [`docs/macos-client-patterns.md`](docs/macos-client-patterns.md). Product versioning: [`docs/versioning.md`](docs/versioning.md). Implementation milestones: [`docs/deployment-plan/`](docs/deployment-plan/). FFI protobuf: [`api/proto/README.md`](api/proto/README.md).

**Requires macOS 14 (Sonoma) or later.**

## Run the Mac app

1. Install Xcode (this repo was last built with Xcode 26; any current Xcode that can target macOS 14 is fine).
2. If you changed Go FFI code, rebuild the embedded library: `./scripts/build-macos-core.sh`
3. Open **[`macos/Provenance.xcodeproj`](macos/Provenance.xcodeproj)** (the file that ends in `.xcodeproj`). Do not open the `macos/App` folder; that is source, not the Xcode project. From Terminal: `open macos/Provenance.xcodeproj`
4. Select the **Provenance** scheme and **My Mac**.
5. Run (⌘R). First launch asks for **researcher name** and a **family / project name**, then writes `~/Library/Application Support/Provenance/identity.json` and `{Name}.provenance` under Documents. A later launch skips the form if that identity file exists. Last-project reopen is spike PR9.

From the command line:

```sh
./scripts/build-macos-core.sh
xcodebuild -project macos/Provenance.xcodeproj -scheme Provenance -destination 'platform=macOS' build
```

The first `xcodebuild` resolves the SwiftProtobuf package from GitHub.

## Go core

The Go module is `github.com/mendahu/provenance`. It requires **Go 1.27** or later.

```sh
go test ./api/... ./core/...
```

PRs to `main` that are **not drafts** run this in GitHub Actions (`.github/workflows/go-test.yml`).

Regenerate protobuf (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`):

```sh
./scripts/generate-proto.sh
```
