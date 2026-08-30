# macOS client patterns

## Status

Working notes for the Swift/SwiftUI Mac client. Domain rules and persistence stay in Go ([`application-stack.md`](application-stack.md)). This document is how the Mac UI should be structured so it stays testable and replaceable (a future Windows client should not need these types).

Audience: someone comfortable with JavaScript and Go, new to Apple platforms.

## Folder layout (`macos/App/`)

| Folder | Scope | Examples |
| --- | --- | --- |
| `App/` | Process entry: scenes, menus. Stay thin. | `ProvenenciaApp.swift` |
| `Platform/` | Reused for the whole Mac client: store protocol, FFI, fakes, `L10n`. Not a screen. | `GenealogyStore`, `GoStore`, `CoreInvoke`, `L10n` |
| `Platform/Generated/` | `protoc` output. Do not edit by hand. | `engine.pb.swift` |
| `Features/<Name>/` | One product flow (view + model). | `Onboarding` |
| `Features/Onboarding/` | File vs new, then contributor or names; home stub; relaunch uses active project. | `OnboardingView` |
| `Resources/` | Assets, entitlements, bridging header, String Catalogs | `Localizable.xcstrings`, `InfoPlist.xcstrings` |

The Go dylib is written to [`macos/Core/`](../macos/Core/) as a **build artifact** (gitignored). Do not commit `libprovenencia.dylib`.

Open **`macos/Provenencia.xcodeproj`**, not `macos/App`. The latter is the source tree; it is not an Xcode project.

---


## 1. SwiftUI describes UI; Go is the engine

Treat SwiftUI like React: it renders. It does not own SQLite, identity, or genealogy.

```text
SwiftUI View  →  small Swift model / store  →  FFI  →  Go core
```

Views may format strings and enable/disable buttons. They must not open `provenencia.sqlite` or mint UUIDs. Launch and the open picker only `stat` that the catalog file exists (`InstallPaths`); Go still owns a real open.

---

## 2. Views are value types; `body` stays cheap

SwiftUI `View`s are **structs** (copied values), not long-lived class components.

- Put logic in a **model** (`struct` or `@Observable` class), not in `body`.
- `body` should be a pure description of UI from current state, like a React function component without side effects.
- Side effects (create project, I/O) belong in actions the model calls, or in `.task` / button handlers that call the store — not in `body` itself.

That is the same split as “don’t put fetch and SQL in the JSX.”

---

## 3. Protocol + fake for the Go core

Do not call the dylib from every view. Introduce a small Swift protocol for application operations, for example `GenealogyStore` with `createProject`, `currentUser`, and so on.

- **Production:** type that talks protobuf/FFI.
- **Previews and tests:** `FakeStore` in memory.

UI tests and ViewModel tests then run without SQLite or a dylib. Go still tests the real catalog.

---

## 4. AppKit is an escape hatch, not the default

Prefer SwiftUI. Use AppKit (`NSColor`, `NSOpenPanel`, menus, document packages) when Mac-native behavior is not available or is awkward in SwiftUI. Keep the escape in a small wrapper, not scattered `NSViewRepresentable`s in every screen.

`Color(nsColor:)` in the blank window is that pattern at the smallest scale.

---

## 5. Accessibility identifiers if you UI-test

XCUITest finds controls by **accessibility** (label, identifier), not by pixel position.

When a control matters to a UI test, set `.accessibilityIdentifier("onboarding.continue")` (stable, dotted names). Do not query by English title alone if the title will be localized.

Previews (`#Preview`) are Storybook-like. They are not CI tests.

---

## 6. Localization

User-facing Mac UI copy goes through typed `L10n` (`LocalizedStringResource` in [`macos/App/Platform/L10n.swift`](../macos/App/Platform/L10n.swift)). Views and models must not hard-code English (or raw catalog keys).

- **Catalogs:** [`Localizable.xcstrings`](../macos/App/Resources/Localizable.xcstrings) for UI strings; [`InfoPlist.xcstrings`](../macos/App/Resources/InfoPlist.xcstrings) for Info.plist TCC/copy keys.
- **Keys:** semantic dotted names (`onboarding.chooseFile.welcomeTitle`), with English in `defaultValue` and in the catalog.
- **Call sites:** `Text(L10n.Onboarding.welcomeTitle)`, `Button(L10n.Onboarding.continueAction)`, or `String(localized: L10n.Onboarding.missingProject)` for `String` / AppKit. Interpolated copy uses `%@` in the catalog and `String(format:locale:_:)` inside `L10n` helpers.
- **Add a string:** add a `LocalizedStringResource` under `L10n`, use it at the call site, and add the matching key/`en` value in the String Catalog (or build once so extraction can pick it up). Prune unused `L10n` members and catalog keys together.
- **Add a language later:** Xcode → String Catalog → add locale → translate. No call-site changes.
- **Go/FFI errors:** still English `err.Error()` text until mapped to `L10n` (prefer stable codes later). Brand name `Provenencia` stays untranslated.

Keep using `.accessibilityIdentifier` for UI tests, not localized titles (§5).

---

## 7. Testing ladder

| Layer | Tool | What belongs there |
| --- | --- | --- |
| Domain, SQLite, identity, project folders | `go test` | Almost all product behavior |
| Path trim, `OnboardingModel`, `InstallPaths` | `ProvenenciaTests` (Swift Testing) | FakeStore + temp folders; no dylib. `xcodebuild test -scheme Provenencia` |
| First-run happy path | One XCUITest later | Optional; needs identifiers |
| Pixel snapshots | Third-party, sparingly | Easy to churn across macOS versions |

Host logic so it can be tested **without** the live Go store. The unit-test bundle still hosts in Provenencia.app. Add a UI-test target when the happy path is worth locking.

CI for Go runs on Linux. Swift tests run on a macOS GitHub runner (`.github/workflows/macos-test.yml`).

---

## 8. Bundle and windows

- **Bundle id** (`app.provenencia`) is the OS identity of the app (defaults, later UTIs), not the product SemVer.
- **`App` / `Scene` / `View`:** process entry and windows vs content inside a window. Keep `ProvenenciaApp` thin (scenes, menus). Screens live in their own files.
- **`WindowGroup`** is a generic window. A real `.provenencia` document flow may later use a document scene; do not force that until we open files.

Product version remains [`versioning.md`](versioning.md) (`VERSION` / About box), not the bundle id.
