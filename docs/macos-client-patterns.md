# macOS client patterns

## Status

Working notes for the Swift/SwiftUI Mac client. Domain rules and persistence stay in Go ([`application-stack.md`](application-stack.md)). This document is how the Mac UI should be structured so it stays testable and replaceable (a future Windows client should not need these types).

Audience: someone comfortable with JavaScript and Go, new to Apple platforms.

## Folder layout (`macos/App/`)

| Folder | Scope | Examples |
| --- | --- | --- |
| `App/` | Process entry (`ProvenenciaMain` picks the app or the test host), scenes, menus. Stay thin. | `ProvenenciaMain.swift`, `ProvenenciaApp.swift` |
| `Platform/` | Reused for the whole Mac client: store protocol, FFI, fakes, `L10n`, value formatting. Not a screen. | `GenealogyStore`, `GoStore`, `CoreInvoke`, `L10n` |
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
| `Features/EvidenceGraph/` | One Evidence graph, opened by graph id: cards, connect, composer handoff. | `EvidenceGraphView` |
| `Features/GraphCanvas/` | Product-agnostic pan/zoom canvas used by the evidence graph; the Artifact region overlay borrows its `GraphCanvasCoordinates` and drag threshold; the item contract is in its [README](../macos/App/Features/GraphCanvas/README.md#items-appkit-owns-where-items-are). | `GraphCanvasScrollView` |
| `Features/CitationComposer/` | Citation place: artifact, locator, observations. | `CitationComposerView` |
| `Features/OCR/` | Reusable on-device text recognition (Vision). Not a screen. | `OCREngine`, `OCRImage` |
| `Features/ArtifactViewer/` | Image and PDF viewer plus region geometry. | `ArtifactViewer` |
| `Features/Names/` | Shared name-value editor. | `NameValueEditorForm` |
| `Features/Dates/` | Shared date-value editor. | `DateValueEditorForm` |
| `Resources/` | Assets, entitlements, bridging header, String Catalogs | `Localizable.xcstrings`, `InfoPlist.xcstrings` |

The Go dylib is written to [`macos/Core/`](../macos/Core/) as a **build artifact** (gitignored). Do not commit `libprovenencia.dylib`.

Open **`macos/Provenencia.xcodeproj`**, not `macos/App`. The latter is the source tree; it is not an Xcode project.

`App/App`, `App/Platform`, `App/Features`, `App/DesignSystem` and `ProvenenciaTests` are synchronized folders: a file on disk there is in the build. `App/Resources` is a classic group. A new top-level folder under `macos/App` is not built until it is added as a synchronized folder; `scripts/check-guidance-names.py` fails until it is.

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

Go holds one exclusive catalog session per open project and serializes FFI ops ([`.agents/skills/use-catalog-session/SKILL.md`](../.agents/skills/use-catalog-session/SKILL.md)).

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
| `QueryHandle` | Observable load state per key (`.ready`, `.loading`, stale `isFetching`; `isStale` = a write moved the value and a reload is owed; `isSettled` = ready, not refetching and not stale; `read` folds value / error / loading and is for a reader that shows only a value or a failure; `reading` is for a reader that shows a refresh failure beside its value (loading, failed, or loaded with an optional `refreshFailure`); `failureRevision` bumps on each failure so a notice can speak again; `isReading` = a first load or a refetch is running) |

**Cache contract (required):**

1. The registry is the only loader. Features do not rebuild a `CatalogQueryKey` payload from `GenealogyStore`.
2. `session.apply(mutation)` patches if the mutation says so, then invalidates and **revalidates every already-warmed key**. Features never `session.query` after `apply`.
   A write that returns `WriteEffects` is applied with `apply(effects)`; one that also has a mutation naming keys the effects cannot (a Source page's own workspace) is applied with `apply(effects, mutation)`, which patches the moved rows, unions the invalidations, and never invalidates a key twice. Every write under a Source returns its effects, so the Sources list and the Conclusion lists are patched or invalidated from `sourceScopes` (activity: patches the Sources row, and the detail keys whose records cite the Source follow it through `.source`) and `handles`, never from a blanket mutation trigger; `evidenceGraphs` names the Evidence graphs whose rows changed and reloads each, so a note does not; `deleted` names the ids the write deleted by entity (its `canonical_entity` entry, read as `deletedHandles`, holds the handles: an Unpromote of a person's last member, a Subject delete of a handle's last member), and `apply` evicts the hidden keys whose kind declares that entity (`deletedWith`: the per-entity Conclusion keys declare `.handle`); a page that shows a deleted id applies the effects with `apply(_:_:leavingPage:)`, which hides every key, applies (evicting the leaving page's keys), then leaves, as the Source page's delete and the Person page's Unpromote pass `navigation.fallbackToSectionRoot` through `UnpromoteFlow.confirm(session:userID:leavingPage:announce:)`, which leaves only when `deletedHandles` holds its entity; `reads` names the keyed partial reads a write moved (`citations`, `evidence_graph_progress` per graph, `evidence_graphs` per Source, `property_terms` per Property) with only the ids they moved for, and each kind's `refreshes` follow its read for those ids alone (the graph list and graph progress re-read just those ids and merge them into the warmed value): the Evidence graph, list, detail and person-claims keys follow `evidenceGraphs` and `touchedHandles` through their descriptor `refreshes`, revalidated once alongside the registry's invalidations in `apply` (see `add-workspace-place` Refresh). The catalog-wide `lists` and `structure` reach their keys the same way, through `.list(…)` and `.structure` entries on each descriptor; the registry invalidates only what a mutation names. A vocabulary write applies its effects, whose `source_vocabulary` (Source types, metadata fields) or `property_vocabulary` (Properties, terms, subject types, bindings) each label-holding key follows; a term write also reports a `property_terms` read for its Property, which only that Property's term list follows. The mutations rule 2 keeps are the Source-row mutations (`.createdSource`, `.updatedSource`, `.changedSourceType`, `.deletedSource`), `.mutatedSourceWorkspace`, `.updatedMetadataField` / `.deletedMetadataField` (the Source pages a field relabel or delete restates, which a list cannot tell from a field create), and the type–field join (`.assignedTypeSuggestion`, `.removedTypeSuggestion`: unaudited, no effects). The Source row's own writes and Promote report their effects too and apply them beside the mutation that carries the row or names the Source page: `createSource` and Copy Source apply `.createdSource(source)`, whose row the Sources list inserts (a new row the effects cannot describe, but the response carries); `updateSource` and `setSourceCover` apply `.updatedSource(source)` (`.changedSourceType(source)` for a type change); `deleteSource` applies `.deletedSource(id:)`; a Promote Done applies its effects alone. The sidebar recounts when a write's `WriteEffects.lists` has `counts` (a Source or a handle was created, deleted or merged).
   A registry `whenHidden` policy says what an invalidated key that isn't on screen does ("on screen" is the keys of the place from the last `apply(location:)`, `visibleKeys`): `.revalidate` (the default) reloads it now; `.markStale` keeps the old value and reloads on the next `query`. The People, Events, Places, and Sources lists, a visited graph canvas (`evidenceGraph`), a visited Source page (`sourceWorkspace`) and the Conclusion detail and Person section keys are `.markStale`. One save made away from a list reloads no list until the researcher looks at one.
3. `readyValue` waits out in-flight work, then returns `nil` if the key is still stale.
4. `setQueryValue` and `patchCached` are overlays (optimistic drag, a list-row patch). A patch is never a substitute for invalidation **unless the write reported everything the patch needs**: a write that reports its activity (`WriteEffects.revision`, `revision_at`) moves a list's last-updated cell through a `CatalogQueryPatch.lastUpdated` on that list's descriptor (`follows` names the ids), with no read. Invalidation stays for changes the client cannot know (a value, title, or membership changed). A create that returns its row patches list membership the same way (`sourceRow()` appends the row a `.createdSource(source)` carries). Patches go through `WorkspaceSession.patchCached`, so an in-flight load is cancelled and restarted and an older answer cannot overwrite the patch.
5. Views observe handles. Models keep form/interaction state only. Evidence graph cards are a **pure join** of `evidenceGraph` (`EvidenceGraphRows`) and `propertiesWorkspace` types — that join is not a third cache key.
6. If a place reads a key, that key is in `PlaceRegistry.queryKeys`.
7. `getCitation` is the exception: a one-shot into the composer fields draft. Do not cache the open citation.

**Sidebar sections (S9-D1):** the sidebar is built from `WorkspaceSidebarSections` — titled `PVSidebarNav` groups, **Source** and **Conclude** at the top, **Configure** bottom-aligned above the footer. Research rows carry counts; Configure rows carry none. Every destination is a top-level row; a new destination goes into its section's list there, not into the view.

**Conclusion keys (Spike 9):**

- Every Conclusion key (the Persons, Events, and Places lists, `conclusionDetail` and the Person page's reads) follows `touchedHandles` through its descriptor `refreshes`, so a write reloads only the handles it reports, and `property_vocabulary` for a label change. No mutation reaches a Conclusion key.
- **Per-entity keys while hidden.** `conclusionDetail` and the Person section keys (claims, family, events, record links) are `whenHidden: .markStale`: a trigger flags the pages that aren't on screen and reloads none of them, so Back shows the last value with "Refreshing" while it reloads. The visible page revalidates as usual. A write that deleted a handle evicts that handle's hidden keys (`deletedWith: .handle`); a merged handle's keys go stale and read gone on return. At most `WorkspaceSession.maxHiddenEntities` (50) hidden entities keep keys, the least recently shown dropped first, every key of an entity together (`apply(location:)`). The detail key follows its own handle, property labels and the Sources its records cite (`.source`). The three lists are `.markStale`: the visible one reloads, the others reload when shown.
- **One detail key for every kind.** `conclusionDetail(project:, entityId:)` reads `GetConclusionDetail`: each bound Property's state, every auto-reconciled value with its reason, every Observation's outcome with its evidence, and the handle's **summary** for its kind (a Person's birth and death, an Event's title and places, a Place's names) in the same call. For a Person it also carries the records behind its life rows, keyed by the owning Event or Place (`derivedRows`). A page's header and default sections use this key alone, so they have one load, one error, and one consistent read; don't add a per-entity summary key beside it. A page section may read its own key, declared in `ConclusionPageSection.reads` (the Person's Family does). `ReconciledValueDisplay` words its states and outcomes.
- **One sectioned detail page.** `ConclusionDetailPage` draws the shared header, then fills everything below it from its `pageSections` slot. The convenience initializer supplies `ConclusionDetailDefaultPageSections` (Details rows, the gallery slot, Place relationship rows), so Event and Place pass only the `userID`: a Conclusion kind passes `userID`, and the page owns the lead model (`ConclusionLeadModel`, which chooses a gallery lead) and its toast; a kind that wants its own sections uses the full initializer rather than a page type of its own. `ConclusionDetailBody.sections` and `ConclusionDetailSection` are the Place hierarchy rows (Part of, Contains, Succession) and are drawn inside the default page section; they are not the page's sections. A kind with sections is offered them by `ConclusionPageSection.offered(by:)` (the Person page: Vitals, Family, Timeline, Gallery, Sources); the stored one lives on `WorkspaceLocation.pageSection` (nil is the first), so switching sections pushes history and a reopened project returns to it, and `WorkspaceLocation.openPageSection` is the one shown (stored, else first). The page, not the kind, reads it from navigation, draws the switcher (`PVSegmentedControl`, whose focus handle lands on the selected segment; a switch chosen on it says "{Section} shown." and keeps the keyboard on the switcher when it was in the sections column or on the switcher, `ConclusionPageSectionSwitch`) and pins the header, and passes it to the fill in `ConclusionDetailPageSectionsContext.openPageSection`; the kind's fill is one `switch` over it. A section declares its reads once in `ConclusionPageSection.reads(project:entityId:)`, a list of `ConclusionSectionRead`s tagged own, companion or supporting; the place warms each (`Role.warmsWithPlace`), the switcher's "Refreshing" follows the own and companion ones (`Role.showsRefreshing`), and the section reads the same keys. That section sits on `ConclusionPageSectionHost`, which gets the handle of its `read:` role (own by default), starts the read and draws the spinner, error Callout and empty state around the section's content (Family, the Timeline and Sources; Sources also reads the Source types list as a supporting read, and its record links as a companion read, drawn in its own region by a second host: `read: .companion` and `placement: .column(loadingHeight:)` give it a column's states). A section computes its identifier root once, `ConclusionPageSection.accessibilityIdentifierRoot(pageIdentifier:)`, and passes it down as `accessibilityIdentifierRoot:` to the host and every view that builds an identifier; `ConclusionPageSection.fillsWidth` opts a section into the full-width mode (Timeline, Gallery and Sources). Every long list in a section folds through `ConclusionFold` (a declared first limit, then pages of 50) and holds its revealed pages in one `ConclusionFoldState`; the section's "Show N more" (`ConclusionShowMoreRow`) words the next page and, after a key press, claims the keyboard for the first item the page adds through `FocusClaim`.
- **Detail pages show what was recorded (S9-16).** The detail read returns only the Properties a handle's records speak to. A page names the few rows it always shows, empty or not (a Person: name and the four life rows); every other field appears once recorded.
- **Detail pages share their rows (S9-16).** `ReconciledValueRow` (label · value · state badge · Sources · disclosures) and its Why (`ReconciliationReasoningView`, `ReconciliationRecordRows`) are snowflakes in `Features/Conclusions/`, shared by the Person, Event and Place pages rather than promoted to recipes. A Why record's Source opens its Citation in the composer (`WorkspaceLocation.composer(for:)`), so Back returns to the page. Disclosures are kit `PVDisclosureButton`s. The Person page's Vitals use `.sheet` mode instead (S13-11): the row ends with a provenance `PVIconButton` and `ConclusionDetailRowsSection` presents one `PVProvenanceSheet` (kit recipe `Recipes/ProvenanceSheet/`, S13-36) for all rows on `.pvSheet(item:)`, holding the row's id (`ProvenanceSheetItem`) rather than a snapshot so the sheet follows its row across a refresh and closes when the row goes. A row holds what its records are built from, and `provenanceContent()` builds the `PVProvenanceSheetContent` when a sheet or a Why opens. The sheet's table is `PVProvenanceRecordTable`; `ReconciliationRecordRows` is the Why's own table and draws the recipe's parts. A record opens three places through one callback `(Record, PVProvenanceLink)`, mapped by `WorkspaceLocation.provenance(_:_:)`; a link closes the sheet first (`\.pvSheetClose`), so Back lands on the page.
- **Go returns structures; Swift makes text.** Payloads carry NameValues, DateValues, term ids, and title parts — never display strings. Formatting lives in the app (`NameValueDisplay`, `PersonTitleDisplay`, `ReconciledValueDisplay`, `EventTitleDisplay`). Choosing *which* naming-matrix rule titles an Event is a rule, so Go does it (`core/eventtitle`) and sends `EventTitle { rule, parts }`; Swift only fills that rule's template.

**View rules (required):**

- Warm loads in `.task { model.warm…Queries() }` or rely on navigation `apply(location:)` — not in `body`.
- Read display data via `session.queryHandle(_)`; observe with `@Bindable var handle: QueryHandle<…>` in a child view.
- Never call `session.query()` from computed properties that `body` reads every frame.

**Mutations:** patch list/detail synchronously when the save response is enough (`updatedSource`); `apply` invalidates+reloads on create/delete/ambiguous busts. Name mutations for the **read model** they change (`.deletedSource`), not the UI verb. A write that returns `WriteEffects` needs no mutation for what its effects name: the Evidence graph follows `evidenceGraphs`. Stale-while-revalidate is for **navigation reads**, not the only path after a write.

**One cache owns each list.** A payload that duplicates another key's data has to
be invalidated whenever that data changes anywhere, which is a bust with no
natural scope. `GetSourceWorkspace` used to fold in the source-type, credibility-grade,
and metadata-field vocabulary to save catalog opens; once `catalogsession` held the
catalog open, that saved nothing and only meant a field added on the Metadata
page left every open Source page stale. The page now reads those three lists from
their own keys (warmed together by `PlaceRegistry`), so adding vocabulary
invalidates one list and no pages. A Source-vocabulary write reloads the Source
vocabulary lists and no page; a field relabel or delete marks the Source pages
stale, because their derived rows embed the field.

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

**Leave guard and row commits.** Places with unsaved drafts implement `WorkspaceLeaveGuard` and set `WorkspaceNavigation.leaveGuard` while they are on screen. `go(to:)`, `goBack()`, `goForward()`, and `go(toIndex:)` ask the guard; `fallbackToSectionRoot()` and `attachProject` do not. The citation composer holds when citation fields are dirty, an observation draft is open, or the edit dialog is dirty — then **Discard changes** resumes the held move and **Keep editing** cancels it. Observation, citation-field, and connection writes are separate store calls; nothing deletes an Observation because it was missing from a payload.

Agent workflow: [`.agents/skills/add-workspace-place/SKILL.md`](../.agents/skills/add-workspace-place/SKILL.md).

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

Agent workflow: [`.agents/skills/add-localized-string/SKILL.md`](../.agents/skills/add-localized-string/SKILL.md). Enforcement when editing Mac UI: [`.cursor/rules/macos-l10n.mdc`](../.cursor/rules/macos-l10n.mdc).

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

## 9. Window widths: Regular and Wide

The window opens at `PVSpacing.widthWorkspaceDefault` × `heightWorkspaceDefault` (1366 × 900pt, clamped to the screen), which is laptop-shaped. People also stretch it across a large external monitor. A view designed only for the default size then sits in its 1180pt `widthContentMax` column, with a wall of empty space beside it (centered on the Source, Person, Event and Place pages and lists). Every view is designed for both widths.

### The breakpoint

**One breakpoint: 1600pt of pane width.**

| Width class | Pane width | Typical setup |
| --- | --- | --- |
| Regular | below 1600pt | The default window; any MacBook, even full screen; half of a big monitor |
| Wide | 1600pt and up | Full screen or most of a 1920pt-or-wider monitor; a 16" MacBook full screen with the sidebar collapsed |

**Pane width** is what the destination itself gets: the window minus the sidebar (264pt, 78pt collapsed). A page with a side pane (an inspector, a detail pane) plans that pane in its own Wide frame; it does not re-measure. It is not the screen or the window. Read it with `@Environment(\.pvWidthClass)`; `WorkspaceDestinationHost` fills the detail area and measures it once. Collapsing the sidebar can move a view across the line, and that is intended.

Where the numbers come from (points, sidebar open):

| Setup | Window | Pane | Class |
| --- | --- | --- | --- |
| Default window | 1366 | ~1102 | Regular |
| 13"–15" MacBook Air full screen | 1470–1710 | ~1206–1446 | Regular |
| 16" MacBook Pro full screen | 1728 | ~1464 | Regular |
| 4K monitor at "looks like 1920 × 1080" | 1920 | ~1656 | Wide |
| 27" 5K / 1440p monitor | 2560 | ~2296 | Wide |
| 32" 6K monitor | 3008 | ~2744 | Wide |

1600 is also the smallest pane that holds a full `widthContentMax` column (1180), a gap (32), an inspector-width column (`widthInspector`, 340) and two page gutters (2 × 32) without squeezing either column much. Below that, a second column steals from the first.

The breakpoint is a token (`PVSpacing.widthWideBreakpoint = 1600`) and the class is `PVWidthClass` (`Components/WidthClass/PVWidthClass.swift`): `WorkspaceDestinationHost` measures the pane once with `pvMeasuresWidthClass()`, and a page reads `@Environment(\.pvWidthClass)` instead of measuring itself.

`PVPageColumns` (`Components/PageColumns/PVPageColumns.swift`) is the record-page scaffold: a header slot (scrolling, or pinned), a primary column and an optional secondary column, sized by `PVPageColumnsMetrics.columnWidths`, stacked at Regular, side by side at Wide, and centered. `ConclusionDetailPage` is built on it, and the Source page (`SourcePageView`: the pinned header over `primaryFill` with `fillSplitSideMax`, and `SourcePageOverviewLayout` placing Metadata in a side column at Wide).

### Rules for Wide

The same six rules are in the [`add-design-brief`](../.agents/skills/add-design-brief/SKILL.md) Paste block, word for word, because Claude Design does not follow links. Change both together.

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band; with `fillSplitSideMax` the primary and the band are the main-and-side group the page lays out inside it, as the Source page does); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Fill the width only with something useful.** A list, table, grid or canvas takes the extra width when Wide has more to show in it: more columns, more cards, the detail beside the list. When it has nothing more to show, it keeps a maximum width (1600pt of content for a list) and is centered, with the page gutters outside it. Empty space stretched between a row's columns is never the Wide design. Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Rule 3's list maximum is a token, `PVSpacing.widthListWideMax` (1600pt of content, gutters outside; S15-04 adds it, and Promote's list reads it): a list whose Wide arrangement adds no column, or only a narrow one, centers at that width with hairlines at its sides through `pvListBand(_:)` (`Components/PageColumns/PVPageColumns.swift`), as the Sources list does (S15-04, user 2026-11-10: a list stretched across a 2296pt pane left a wide empty band between its title and its graphs).

A full-width section's empty, loading and error states keep the page band: apply `pvKeepsPageBand()` (`Components/PageColumns/PVPageColumns.swift`) to them, so they stay centered at `widthContentMax` instead of stretching across a Wide pane. A page that passes `fillSplitSideMax` has a wider band (the main-and-side group); its states sit in the primary, which is that band.

`FocusKeeper` (`Platform/FocusKeeper.swift`, applied with `focusKeeper(...)`, which takes a pure `map` from a stop to the stop that can take the keyboard and a `scroll` that only a claim calls) is how a view keeps rule 5's focus: it hands the keyboard back to the stop that survived a width change, to a fallback when a card was folded away, and to a tile a lazy grid dropped.

### Patterns to reach for

| View shape | Regular | Wide |
| --- | --- | --- |
| List → detail page (Sources, Conclusions, vocabulary) | List; select navigates to the detail | List column (about 420–520pt) and the selected detail side by side. Until a list's redesign brings the detail beside it, the list adds what columns it has (the Sources list's Published column) and centers at `widthListWideMax` |
| Record / detail page (Source page, Conclusion detail) | One column; supporting sections stacked below | Primary column (≤ `widthContentMax`) plus a secondary column (340–480pt) of supporting content: metadata, linked records, activity, artifacts |
| List with a pinned side column at both widths (Person's Sources) | The rows with the side column beside them while the pane holds the rows' minimum, else stacked under them | The same, the side column up to 720pt; `ConclusionPinnedSideColumn` holds it in view while the rows scroll |
| Grid (artifacts, galleries) | Columns that fit at the card minimum | More columns at the same card size |
| Table | Essential columns | Columns that were hidden at Regular |
| Form / editor | One column at the form measure | Field groups in two columns, or the form beside a live preview or reference panel |
| Canvas (evidence graph, viewer) | Canvas; inspector on demand | Canvas with the inspector docked open |
| List with a side pane (Properties, Metadata, Source types) | List and a fixed side pane | The list takes the extra width (more columns), or the side pane widens; drawn in the page's brief |

A view that fits none of these says in its brief what its Wide frame does and why. Briefs carry these rules in their Paste block, because Claude Design does not follow links ([`add-design-brief`](../.agents/skills/add-design-brief/SKILL.md)).
