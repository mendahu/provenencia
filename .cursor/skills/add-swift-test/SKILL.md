---
name: add-swift-test
description: >-
  Adds Provenencia macOS unit tests under macos/ProvenenciaTests using Swift Testing.
  Use when adding or changing OnboardingModel, InstallPaths, FakeStore, ProvenenciaTests,
  xcodebuild test, Swift Testing @Test/#expect, or wiring a .swift file into the
  ProvenenciaTests Xcode target. Not for Go tests, XCUITest, or GoStore/dylib tests.
---

# Add a Swift unit test

Product logic under test lives in `macos/App/`. Tests live in **`macos/ProvenenciaTests/`** (sibling of `App/`, not co-located). They compile in the **ProvenenciaTests** bundle, host the Provenencia app, and `@testable import Provenencia`.

## What to test

- Models (`OnboardingModel`), path helpers (`InstallPaths`), store fakes.
- Inject `FakeStore` and `OnboardingFolders` with dirs under `FileManager.default.temporaryDirectory`.
- **No** `GoStore`, `provenencia_call`, dylib, or generated `engine.pb.swift`.
- **No** XCUITest / view snapshots unless the user asks.

Domain, SQLite, and identity stay in `go test`.

## File

Copy [`macos/ProvenenciaTests/OnboardingModelTests.swift`](macos/ProvenenciaTests/OnboardingModelTests.swift) or [`InstallPathsTests.swift`](macos/ProvenenciaTests/InstallPathsTests.swift).

```swift
import Foundation
import Testing
@testable import Provenencia

@Suite
struct FooTests {
    @Test func doesTheThing() throws {
        #expect(true)
    }
}
```

- Swift Testing only: `import Testing`, `@Test`, `#expect`. Not XCTest.
- `@MainActor` on the suite (or tests) when touching `OnboardingModel`.
- `@Suite` is optional for a single struct; use it when grouping.
- Prefer one `@Test` per behavior (load matrix rows, not one mega-test) unless the setup is identical and a loop stays readable.
- Sort unordered `FileManager` listings before comparing.

## Xcode project

Add the file to **ProvenenciaTests** only (never the Provenencia app target).

In [`macos/Provenencia.xcodeproj/project.pbxproj`](macos/Provenencia.xcodeproj/project.pbxproj), copy an existing test file’s four objects and give them unused `AA00000000000000000000xx` IDs (do not reuse `B0`–`BE`):

1. `PBXFileReference` in group `AA00000000000000000000B5` (`path = ProvenenciaTests`)
2. `PBXBuildFile` in Sources phase `AA00000000000000000000B7`
3. Do not add a new native target or scheme entry; those already exist.

Bundle id stays `app.provenencia.tests`. Host: Provenencia.app.

## Run

```sh
xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'
```

⌘U in Xcode (Provenencia scheme). CI: `.github/workflows/macos-test.yml` (non-draft PRs to `main`).

## Do not

- Mix test sources into `macos/App/`
- Call `GoStore` / the dylib. Fixtures that pretend to be a project should include `provenencia.sqlite` (file exists; do not open the DB).
- Bump `VERSION` for tests-only
- Skip signing flags in local `xcodebuild`; CI sets `CODE_SIGNING_ALLOWED=NO`
