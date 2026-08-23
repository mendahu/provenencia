# Provenance

Local-first genealogy software. Native macOS client over a shared Go core. Domain and stack notes live in [`docs/`](docs/). Implementation milestones: [`docs/deployment-plan/`](docs/deployment-plan/).

**Requires macOS 14 (Sonoma) or later.**

## Run the Mac app

1. Install Xcode (this repo was last built with Xcode 26; any current Xcode that can target macOS 14 is fine).
2. Open [`macos/Provenance.xcodeproj`](macos/Provenance.xcodeproj).
3. Select the **Provenance** scheme and **My Mac**.
4. Run (⌘R). You should get an empty window. There is no onboarding or database yet (Spike 1 PR1).

From the command line:

```sh
xcodebuild -project macos/Provenance.xcodeproj -scheme Provenance -destination 'platform=macOS' build
```

## Go core

The Go module is `github.com/mendahu/provenance`. It requires **Go 1.27** or later (`core/` is a stub until later PRs add protobuf FFI and SQLite).

```sh
go test ./...
```

The Mac app does not link the Go core yet.
