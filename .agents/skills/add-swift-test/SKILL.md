---
name: add-swift-test
description: >-
  Adds Provenencia macOS unit tests under macos/ProvenenciaTests using Swift Testing.
  Use when adding or changing OnboardingModel, InstallPaths, FakeStore, ProvenenciaTests,
  xcodebuild test, Swift Testing @Test/#expect, or adding a test file under macos/ProvenenciaTests. Not for Go tests, XCUITest, or GoStore/dylib tests.
---

# Add a Swift unit test

Product logic under test lives in `macos/App/`. Tests live in **`macos/ProvenenciaTests/`** (sibling of `App/`, not co-located). They compile in the **ProvenenciaTests** bundle, host the Provenencia app, and `@testable import Provenencia`.

## What to test

- Models (`OnboardingModel`), path helpers (`InstallPaths`), store fakes.
- Artifact files on disk (an image via `writeImage(fileType:)` in PNG, JPEG or TIFF, with `dpi:`, `orientation:` or `grayscale:` when the test needs them, a PNG header that only declares a size via `writeDeclaredPNGHeader`, raw bytes via `write(_:)`, a PDF (`sizedPDFDocument(size:)` for one page of a known size, `pdfHost(document:)` for the PDF host view on it), the catalog row that points at them): use `ArtifactFileFixtures` in `macos/ProvenenciaTests/ArtifactFileFixtures.swift`; do not write another copy.
- A `Layout`'s placement: test it through a `static func arrange(width:…fit:)` with made-up sizes, not a hosted window (`SourcePageOverviewLayout.arrange`, `PVPageColumnsLayout.arrange`; the `Layout` passes `subviews.fit` to its `arrange`), or a pure placement value built from measured sizes (`PersonTimelinePlacement`). Host a view only for SwiftUI's own behavior (environment, geometry modifiers, view identity), then call `layoutSubtreeIfNeeded()` on the host and assert at once; do not `waitUntil` on layout. To host a SwiftUI or AppKit view in a window, use `hostInBorderlessWindow` or `hostInBorderlessWindow(nsView:size:)` from `macos/ProvenenciaTests/HostedWindow.swift` (or `testWindow(size:styleMask:)` for a bare window, and `trackTestWindow(_:)` around a panel the builder cannot make) on a suite marked `.hostedWindows`: the window is never shown and the trait closes it when the test ends. Do not hold the window yourself or build an untracked window. Do not synthesize mouse or key events through a window: test the model, the action closure, or the view's own handler instead.
- Inject `FakeStore` and `OnboardingFolders` with dirs under `FileManager.default.temporaryDirectory`.
- **FakeStore:** new methods seed rows and script return values (`nextWriteEffects`, `…Error` hooks, impact maps); do not add graph-list order, Citation membership, effects-resolution, or graph / Copy Source refusal logic.
- When adding a catalog `GenealogyStore` API on `FakeStore`, call `markCatalogSessionHeld` (same as real traffic) so `CatalogSessionStoreTests` / overlap coverage stay honest.
- To pause a fake call mid-write, give it a `FakeHold` (`isOn`, `count`, `release()`); do not hand-roll continuations.
- **No** `GoStore`, `provenencia_call`, dylib, or generated `engine.pb.swift`.
- **No** XCUITest / view snapshots unless the user asks.

Domain, SQLite, and identity stay in `go test`.

## File

Copy [`macos/ProvenenciaTests/OnboardingModelTests.swift`](../../../macos/ProvenenciaTests/OnboardingModelTests.swift) or [`InstallPathsTests.swift`](../../../macos/ProvenenciaTests/InstallPathsTests.swift).

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

`macos/ProvenenciaTests` is a synchronized folder: a `.swift` file saved there is in the **ProvenenciaTests** target. Do not edit `project.pbxproj`, and never add a test file under `macos/App`.

Bundle id stays `app.provenencia.tests`. Host: Provenencia.app, which runs `TestHostApp` under tests: no window, no install identity, no project.

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
