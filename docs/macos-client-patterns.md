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
| `DesignSystem/` | Layer roots `Components/` / `Recipes/` / optional `Snowflakes/` — **one folder per control** (e.g. `Components/Button/PVButton.swift`); plus `Tokens/`. See [`DesignSystem/README.md`](../macos/App/DesignSystem/README.md) and [`design-system-layers.md`](design-system-layers.md). No `Core`/`Forms`/`Research` nesting; no loose `.swift` at a layer root. | `Components/Button/`, `Recipes/Marks/` |
| `Features/<Name>/` | One product flow (view + model). Typical home for snowflake helpers (`private`). | `Onboarding`, `Workspace`, `EvidenceGraph`, `CitationComposer` |
| `Features/Catalog/` | Catalog-wide markers and session counts — not a screen. Prefer a product concept over a `Shared/` dump. | `CatalogCounts`, `OriginMarkers`, `MetadataFieldBadges` |
| `Features/CatalogVocabulary/` | Reusable vocabulary-browser shell (list + detail chrome) for Fields, Types, and later vocab destinations. | `VocabularyListPane`, `VocabularyChrome` |
| `Features/Onboarding/` | File vs new, then contributor or names; home stub; relaunch uses active project. | `OnboardingView` |
| `Features/Workspace/` | Sidebar, toolbar, navigation history, session cache, place registry. | `WorkspaceView`, `WorkspaceSession` |
| `Features/Sources/` | Sources list and source page. | `SourcePageView`, `SourcesListView` |
| `Features/Metadata/` | Source metadata-field vocabulary (the Metadata page). | `MetadataView` |
| `Features/SourceTypes/` | Source-type vocabulary. | `SourceTypesView` |
| `Features/Properties/` | Properties and their Subject-type bindings (the Properties page). | `PropertiesView` |
| `Features/EvidenceGraph/` | Evidence graph for one source: cards, connect, composer handoff. | `EvidenceGraphView` |
| `Features/GraphCanvas/` | Product-agnostic pan/zoom canvas used by the evidence graph. | `GraphCanvasScrollView` |
| `Features/CitationComposer/` | Citation place: artifact, locator, observations. | `CitationComposerView` |
| `Features/OCR/` | Reusable on-device text recognition (Vision). Not a screen. | `OCREngine`, `OCRImage` |
| `Features/ArtifactViewer/` | Image and PDF viewer plus region geometry. | `ArtifactViewer` |
| `Features/Names/` | Shared name-value editor. | `NameValueEditorForm` |
| `Features/Dates/` | Shared date-value editor. | `DateValueEditorForm` |
| `Resources/` | Assets, entitlements, bridging header, String Catalogs | `Localizable.xcstrings`, `InfoPlist.xcstrings` |

The Go dylib is written to [`macos/Core/`](../macos/Core/) as a **build artifact** (gitignored). Do not commit `libprovenencia.dylib`.

Open **`macos/Provenencia.xcodeproj`**, not `macos/App`. The latter is the source tree; it is not an Xcode project.

### UI composition layers

UI building blocks are **components** (content-agnostic `PV*`), **recipes** (Provenencia-specific, reused), or **snowflakes** (one screen). Under each DesignSystem layer root, use **one folder per control** (`Components/Button/…`) — not UI-category nesting and not loose files at the layer root. Public Components/Recipes types use the **`PV` prefix** so they do not collide with SwiftUI/AppKit (`PVButton`, not `Button`). Stop at the highest layer you need. Full model: [`design-system-layers.md`](design-system-layers.md). Agent workflows: `add-ui-component`, `evaluate-ui-component`.

---


## 1. SwiftUI describes UI; Go is the engine

Treat SwiftUI like React: it renders. It does not own SQLite, identity, or genealogy.

```text
SwiftUI View  →  small Swift model / store  →  FFI  →  Go core
```

Views may format strings and enable/disable buttons. They must not open `provenencia.sqlite` or mint UUIDs. Launch and the open picker only `stat` that the catalog file exists (`InstallPaths`); Go still owns a real open.

### Catalog session (workspace)

Go holds one exclusive catalog session per open project and serializes FFI ops ([`docs/ideas/archive/catalog-access-serialization.md`](ideas/archive/catalog-access-serialization.md), `.cursor/skills/use-catalog-session/SKILL.md`).

- **Enter:** mount `WorkspaceView` and start catalog traffic (`CatalogCounts.refreshAll`, feature loads). First store call opens the session; do **not** gate content on badge refresh.
- **Leave:** `WorkspaceView.onDisappear` calls `GenealogyStore.closeCatalogSession(projectDir:)`. Sign-out / project switch also close in Go.
- **FakeStore:** tracks `heldCatalogProjectDir` / `lastClosedCatalogProjectDir` for unit tests — no real locking.

### Workspace session & place registry (Spike 4)

Navigation history (`WorkspaceNavigation`) answers **where**. The **workspace session** answers **what catalog data is already loaded** for the open project.

```text
go(to: location)  →  navigation stack (sync)
                  →  session.apply(location)  →  PlaceRegistry  →  warm CatalogQueryKeys
WorkspaceDestinationHost  →  presentation id  →  destination view
Destination view  →  reads QueryHandle(s)  →  patches / invalidates on mutate
```

| Piece | Role |
| --- | --- |
| `WorkspaceSession` | Project-scoped cache: get-or-load, in-flight dedupe, patch vs invalidate+revalidate |
| `CatalogQueryRegistry` | Declarative loaders + mutation → cache key map |
| `PlaceRegistry` | `WorkspaceLocation` → place id, presentation, required query keys |
| `WorkspaceDestinationHost` | Routes by presentation (list vs detail are separate views) |
| `QueryHandle` | Observable load state per key (`.ready`, `.loading`, stale `isFetching`) |

**Cache contract (required):**

1. The registry is the only loader. Features do not rebuild a `CatalogQueryKey` payload from `GenealogyStore`.
2. `session.apply(mutation)` patches if the mutation says so, then invalidates and **revalidates every already-warmed key**. Features never `session.query` after `apply`.
3. `readyValue` waits out in-flight work, then returns `nil` if the key is still stale.
4. `setQueryValue` is an overlay (optimistic drag, list-row patch the write returned). It is not a substitute for `apply`.
5. Views observe handles. Models keep form/interaction state only. Evidence graph cards are a **pure join** of `sourceGraph` (`SourceGraphRows`) and `propertiesWorkspace` types — that join is not a third cache key.
6. If a place reads a key, that key is in `PlaceRegistry.queryKeys`.
7. `getCitation` is the exception: a one-shot into the composer fields draft. Do not cache the open citation.

**Sidebar sections (S9-D1):** the sidebar is built from `WorkspaceSidebarSections` — titled `PVSidebarNav` groups, **Source** and **Conclude** at the top, **Configure** bottom-aligned above the footer. Research rows carry counts; Configure rows carry none. Every destination is a top-level row; a new destination goes into its section's list there, not into the view.

**Conclusion keys (Spike 9):**

- Every Conclusion key (Persons list today; Events, Places, and details later) invalidates on one set, `CatalogQueryRegistry.conclusionTriggers`. Go's resolved-values cache recomputes on the same writes, so busting them all together is cheap and never misses a dependency between handles.
- **Planned exception to rule 2 (S9-15):** Conclusion *detail* keys that aren't on screen are evicted rather than revalidated, so a trigger doesn't reload every Person page the researcher has visited. The visible page revalidates as usual. List keys follow rule 2.
- **Go returns structures; Swift makes text.** Payloads carry NameValues, DateValues, term ids, and title parts — never display strings. Formatting lives in the app (`NameValueDisplay`, `PersonHeaderDisplay`, later the event-title templates).

**View rules (required):**

- Warm loads in `.task { model.warm…Queries() }` or rely on navigation `apply(location:)` — not in `body`.
- Read display data via `session.queryHandle(_)`; observe with `@Bindable var handle: QueryHandle<…>` in a child view.
- Never call `session.query()` from computed properties that `body` reads every frame.

**Mutations:** patch list/detail synchronously when the save response is enough (`updatedSource`); `apply` invalidates+reloads on create/delete/ambiguous busts. Name mutations for the **read model** they change (`.mutatedSourceGraph`, `.savedCitation`), not the UI verb. Stale-while-revalidate is for **navigation reads**, not the only path after a write.

**One cache owns each list.** A payload that duplicates another key's data has to
be invalidated whenever that data changes anywhere, which is a bust with no
natural scope. `GetSourceWorkspace` used to fold in the source-type, credibility-grade,
and metadata-field vocabulary to save catalog opens; once `catalogsession` held the
catalog open, that saved nothing and only meant a field added on the Metadata
page left every open Source page stale. The page now reads those three lists from
their own keys (warmed together by `PlaceRegistry`), so adding vocabulary
invalidates one list and no pages.

Reach for a bust only where a payload is genuinely *derived* rather than duplicated:

- `workspace.metadata` is a Go-computed join (the type's suggestions ∪ stored values,
  filtered by per-source dismissal, in layout order) and each row embeds its field.
  Relabelling a field or changing a type's suggestions restates it, and the mutation
  can't name which sources are affected — hence
  `CatalogQueryInvalidation.allCached(.sourceWorkspace)`. Keep that join in Go; don't
  rebuild it in Swift to chase a narrower key.
- A `CatalogTypeSuggestion` embeds a whole field row, so a relabel restates suggestion
  lists for types the edit never named.
- Derived counts move when another table is written: `usedBy` / `suggestedFieldCount`
  gate the Delete buttons, so writing a metadata value stales the field list and
  adding a source stales the type list.

A `setQueryValue` patch is an optimization layered on top of invalidation, never a
substitute for it: it silently no-ops when the key was never cached, and it only
fixes the one row the write returned.

**Leave guard and row commits.** Places with unsaved drafts implement `WorkspaceLeaveGuard` and set `WorkspaceNavigation.leaveGuard` while they are on screen. `go(to:)`, `goBack()`, `goForward()`, and `go(toIndex:)` ask the guard; `fallbackToSectionRoot()` and `attachProject` do not. The citation composer holds when citation fields are dirty, an Observation row is edited or a non-empty draft, or a connection is touched — then **Discard changes** resumes the held move and **Keep editing** cancels it. Observation, citation-field, and connection writes are separate store calls; nothing deletes an Observation because it was missing from a payload.

Agent workflow: [`.cursor/skills/add-workspace-place/SKILL.md`](../.cursor/skills/add-workspace-place/SKILL.md). Design: [`ideas/archive/page-navigation-performance.md`](ideas/archive/page-navigation-performance.md).

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

Agent workflow: [`.cursor/skills/add-localized-string/SKILL.md`](../.cursor/skills/add-localized-string/SKILL.md). Enforcement when editing Mac UI: [`.cursor/rules/macos-l10n.mdc`](../.cursor/rules/macos-l10n.mdc).

- **Catalogs:** [`Localizable.xcstrings`](../macos/App/Resources/Localizable.xcstrings) for UI strings; [`InfoPlist.xcstrings`](../macos/App/Resources/InfoPlist.xcstrings) for Info.plist TCC/copy keys.
- **Keys:** semantic dotted names (`onboarding.chooseFile.welcomeTitle`), with English in `defaultValue` and in the catalog.
- **Call sites:** `Text(L10n.Onboarding.welcomeTitle)` is fine. When copy is needed as a `String` — `Button` / `TextField` / `Picker` titles on our macOS 14 deployment target (CI Xcode 16), AppKit, models — use `L10n.string(L10n.…)`. Never `String(localized:)`: resolving a `LocalizedStringResource` that way re-reads and re-parses the whole strings table on every call (~1.5 ms, against ~1 µs for `L10n.string`), which turns into dropped frames when it runs in a view body.
- **Copy with arguments:** a `%@` / `%lld` catalog format (positional `%1$@` when there is more than one argument) filled by an `L10n` format function that returns `String` via `L10n.format(_:_:)`. Never interpolate into a `LocalizedStringResource` (`defaultValue: "\(count) items"`) — `L10n.string` cannot read its arguments. `L10n.format` formats in a locale whose language matches the resolved localization, so plural variations pick the right rules; don't call `String(format:)` on copy yourself. Design-system components take labels as `PVCopy`, which accepts either a static resource or a formatted `String`.
- **Counts:** a count with a noun ("3 fields", "in use on 1 source") is one catalog key with `plural` variations (`one` / `other` in English; other languages add `few`, `many`, …), filled by `L10n.format`. Never pick between singular and plural keys in Swift (`count == 1 ? one : other`) — that only works for English. When the count sits inside a longer string with other arguments, use a catalog substitution (`%1$@ · %#@citations@`). A zero case that reads differently ("no sources yet") can stay its own key. `L10nPluralTests` resolves every plural key in the catalog for one and several, and checks a second language's rules through the same path.
- **Non-translated glyphs / brand / numbers:** use `Text(verbatim:)` (e.g. `/`, `·`, `—`, `›`, `Provenencia`, `"\(index + 1)"`). Bare `Text("—")` or `Text("\(n)")` is treated as a localizable key; with `SWIFT_EMIT_LOC_STRINGS = YES`, Xcode rewrites `Localizable.xcstrings` on build and re-adds junk entries (`%lld`, `%@`, `""`) after you delete them. Same for interpolated accessibility: `Text(verbatim: "\(count)")`, not `Text("\(count)")`. Unlabeled `TextField("", …)` / `Picker("", …)` also extract as `""` — use `TextField(text:prompt:) { EmptyView() }` or `Picker(selection:) { … } label: { EmptyView() }` instead.
- **CI guard:** `scripts/check-localizable-xcstrings.py` (run in the macOS workflow) fails on non-dotted keys, `extracted_with_value`, `isCommentAutoGenerated`, `L10n` keys missing from the catalog, multi-argument values without numbered arguments, interpolated `LocalizedStringResource`s, and any `String(localized:)` call.
- **Add a string:** add a `LocalizedStringResource` under `L10n`, use it at the call site, and add the matching key/`en` value in the String Catalog (or build once so extraction can pick it up). Prune unused `L10n` members and catalog keys together.
- **Add a language later:** Xcode → String Catalog → add locale → translate. No call-site changes.
- **Go/FFI errors:** status `1` returns a protobuf `Error` (`code`, `kind`, `params`) — never English prose. Map codes with `L10n.Errors` (`error.<domain>.<name>` catalog keys). Unknown codes use `error.internal.unknown`. Brand name `Provenencia` stays untranslated.

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
