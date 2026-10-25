# Ship log — S12-11 Source page Graphs section

**Plan:** [`S12-11-source-page-graphs.md`](S12-11-source-page-graphs.md) · **Branch:** `s12-11-source-page-graphs` · **PR:** #549
**Board:** [Claude Design](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share) (S12-D2, agreed 2026-10-25; described in the plan's §6 Frames)
**State:** ready

## Round 0 — implement
Head `0a38d473`. Gate PASS (orchestrator run: diff-check, gofmt, go-vet, go-test, localizable, plan-docs, xcode-test). Report deviations: `PVCard` default padding is 0, so the row and loading card pad with `space6` (D-17 assumed a padded default); `rows`/`loadError` are computed from the session; T-5 applies the Source page location first and waits with `waitUntil` (list is `whenHidden: .markStale`); `PlaceRegistryTests.resolvesEvidenceGraphDistinctFromPage` also gained `.evidenceGraphsList` (follows from D-10); test helper `entryGraphLocation` replaces removed `evidenceGraphLocation`; unused `WorkspaceNavigation` environment read removed from `SourcePageIdentityHeader`. Not done: frame screenshots and §10 manual check (app not run).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — UX matches `macos/App/**` and `Localizable.xcstrings`; SEC+PRF matches performance paths (`macos/App/Features/**` models, graph and list code).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Third/fourth private `CatalogEvidenceGraph` test factory; copied query-wait helper | Fix now → F-1, F-2 | verified: `rg -n 'func graph\(' macos/ProvenenciaTests` → 4 factories; same-pass cleanup. Body search found 4 more wait copies: `SourcePageModelTests.settle`, `SourceTypesModelTests.waitForQuery` (private), `WorkspaceSessionTests.waitForFetchComplete`, `CatalogQueryRegistryTests.waitForFetchComplete` |
| MNT-2 | MNT | major | Source page locations still hand-built after D-16's `.sourcePage` | Fix now → F-3 | verified: `rg -n 'sourceSurface: \.page' macos` → FakeStore:3073, :3112 and 5 test sites; same-pass cleanup |
| MNT-3 | MNT | major | `askDeleteGraph` finds the graph inline, not via `evidenceGraph(id:)` (D-2 owner) | Fix now → F-4 | verified: `SourcePageModel.swift:187` |
| PRF-1 | PRF | minor | Section reads `rows` twice per render; each read scans all graphs | Fix now → F-5 (single read); index stays S12-13 per plan Out line | verified: `SourcePageGraphsView.swift:9,13` |
| MNT-4, UX-5 | MNT, UX | minor | `WorkspaceLocation.sourceTitle` doc comment names the removed crumb | Fix now → F-6 | verified: `WorkspaceLocation.swift:66-67` |
| MNT-5 | MNT | minor | Third copy of the Source page section top padding | Fix now → F-7 | verified: `rg 'space11 - PVSpacing.space9' macos/App` → 3 lines; same-pass cleanup |
| UX-1 | UX | minor | Loading card hidden from VoiceOver | Fix now → F-8 | verified: `SourcePageGraphsView.swift` `loadingRow` `.accessibilityHidden(true)`; `PVList.swift:388-389` precedent |
| UX-2 | UX | minor | Truncated graph name/description has no way to read in full | Fix now → F-8 | file added by this PR; rubric "Truncation needs a way to read the whole value" |
| UX-3 | UX | minor | Last-graph notice says "Add another graph" before Add graph exists | Later PR (S12-12) | written into the S12-12 map scope entry, `4f328395` |

Nits (logged): `CON-1 · WorkspaceToolbar.swift:239 · Source crumb location has no ref (D-9 says ref: nil)` · `MNT-6 · SourcesListNavigation.swift:23-24 · rewrap doc comment` · `MNT-7 · SourcePageModelTests.swift:1178 · redundant first?.name expect` · `UX-4 · SourcePageGraphsView.swift:231 · enabled Open tooltip could name the graph`

### Fix plan 1

Supersedes nothing in the plan's Decisions; F-1, F-2, F-3 and F-7 are same-pass cleanups that reach test files and `FakeStore` beyond the plan's file list. No behavior change a researcher can see except F-8.

**F-1 — One `CatalogEvidenceGraph` test fixture** (from MNT-1)
Files: new `macos/ProvenenciaTests/EvidenceGraphFixtures.swift` (beside `RegionFixtures.swift`; add it to the ProvenenciaTests target in `project.pbxproj`), `EvidenceGraphStoreTests.swift`, `SourcesListNavigationTests.swift`, `SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: add one internal factory, e.g. `func evidenceGraph(_ id: String, ref: String? = nil, sourceID: String = "s1", name: String = "Default", description: String = "", artifactIDs: [String] = []) -> CatalogEvidenceGraph` (pick defaults that keep every caller's current values; `ref` defaults to a value derived from `id` if callers rely on one). Delete the four private `graph(...)` factories (`EvidenceGraphStoreTests.swift:19`, `SourcesListNavigationTests.swift:18` and `:114`, `SourcePageModelTests.swift:1154`) and move their callers onto it. Inline `CatalogEvidenceGraph(` literals elsewhere may stay.
Check: `rg -n 'func graph\(' macos/ProvenenciaTests` prints nothing; scripts/check-pr.sh PASS.

**F-2 — One query-wait helper** (from MNT-1)
Files: `macos/ProvenenciaTests/QueryHandleWaiting.swift`, `SourcesListNavigationTests.swift`, `SourcePageModelTests.swift`, `SourceTypesModelTests.swift`, `WorkspaceSessionTests.swift`, `CatalogQueryRegistryTests.swift`.
Do: give the shared `waitForQuery` an optional `timeoutNanoseconds: UInt64 = 2_000_000_000` parameter (body otherwise unchanged). Delete `SourcesListNavigationTests.settle` (`:103`), `SourcePageModelTests.settle` (`:804`), the private `SourceTypesModelTests.waitForQuery` (`:59`), and the private `waitForFetchComplete` in `WorkspaceSessionTests` and `CatalogQueryRegistryTests`; point every caller at `waitForQuery` (pass the timeout where a caller passed one). The SourceTypes copy did not wait for `!isFetching`; the shared one does, which only waits longer.
Check: `rg -n 'func (settle|waitForFetchComplete|waitForQuery)' macos/ProvenenciaTests` prints only `QueryHandleWaiting.swift`; scripts/check-pr.sh PASS.

**F-3 — Every Source page location through `.sourcePage`** (from MNT-2)
Files: `macos/App/Platform/FakeStore.swift`, `SourcePageModelTests.swift`, `PlaceRegistryTests.swift`, `WorkspaceNavigationTests.swift`.
Do: replace `WorkspaceLocation(section: .sources, sourceId:, sourceSurface: .page, …)` with `WorkspaceLocation.sourcePage(sourceId:ref:title:)` at `FakeStore.swift:3073` and `:3112`, `SourcePageModelTests.swift:1205`, `PlaceRegistryTests.swift:64` and `:232`, `WorkspaceNavigationTests.swift:293` and `:469`. Leave `FakeStore.swift:2884` (Artifact anchor with `artifactId`).
Check: `rg -n 'sourceSurface: \.page' macos` prints only `WorkspaceLocation.swift:177` and `FakeStore.swift:2884`; scripts/check-pr.sh PASS.

**F-4 — `askDeleteGraph` reads through the session accessor** (from MNT-3)
Files: `macos/App/Features/Sources/SourcePageModel.swift`.
Do: `guard let graph = context.session.evidenceGraph(id: id), graph.sourceID == <this page's source id> else { return }` (use whatever the model already holds for the Source id).
Check: T-5 and T-6 still pass; `rg -n 'rows\?\.first' macos/App/Features/Sources` prints nothing; scripts/check-pr.sh PASS.

**F-5 — Read the graph rows once per render** (from PRF-1)
Files: `macos/App/Features/Sources/SourcePageGraphsView.swift`.
Do: in `body`, `let rows = model.graphs.rows` once and use it for both the meta count and the rows branch.
Check: scripts/check-pr.sh PASS.

**F-6 — Reword the `sourceTitle` doc comment** (from MNT-4, UX-5)
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift:66-67`.
Do: "Denormalized Source title, the Source crumb's label while the Sources list has no value (composer, Promote); ignored for navigation identity."
Check: `rg -n 'Evidence graph for' macos` prints nothing.

**F-7 — One Source page section top-padding modifier** (from MNT-5)
Files: `macos/App/Features/Sources/SourcePageChrome.swift`, `SourcePageArtifactsView.swift`, `SourcePageNotesView.swift`, `SourcePageGraphsView.swift`.
Do: add one modifier in `SourcePageChrome.swift` (match its existing modifier style) that applies `.padding(.top, PVSpacing.space11 - PVSpacing.space9)`, and use it in all three sections.
Check: `rg -n 'space11 - PVSpacing.space9' macos/App` prints only the `SourcePageChrome.swift` line; scripts/check-pr.sh PASS.

**F-8 — Loading announces; truncated text has a tooltip** (from UX-1, UX-2)
Files: `macos/App/Features/Sources/SourcePageGraphsView.swift`.
Do: on `loadingRow`, replace `.accessibilityHidden(true)` with `.accessibilityElement(children: .ignore)` and `.accessibilityLabel(Text(L10n.DesignSystem.listLoading))` (as `PVList.swift:388-389`). Add `.help(graph.name)` to the name `Text` and `.help(description)` to the description `Text`. No new strings.
Check: scripts/check-pr.sh PASS; hover check goes in the PR's manual check.

## Round 1 — fix
Head `e09ae66a` (fix commits `2758b90a`..`05dbbd52`, then the orchestrator's S12-12 scope commit cherry-picked as `e09ae66a`; the original `4f328395` had landed on a detached HEAD). Gate PASS (orchestrator run). F-1 to F-8 done. Report deviations: commits grouped by file overlap; fixture defaults `ref` to `GPH-<ID>` and `sourceID` to `src-1`, with old-default callers passing values explicitly; F-8 hover not checked in the app.

Minors read by the orchestrator (no round-2 reviewer): PRF-1 **Resolved** (`SourcePageGraphsView.body` reads `rows` once); UX-1 **Resolved** (`loadingRow` combines and labels with `L10n.DesignSystem.listLoading`); UX-2 **Resolved** (`.help` on name and description; hover goes to manual check); MNT-4/UX-5, MNT-5 go with MNT's round-2 review.

## Round 2 — review
Reviewers: MNT (had majors MNT-1, MNT-2, MNT-3). Not re-sent: CON (clean), UX and SEC+PRF (minors only).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) | Resolved | one fixture `EvidenceGraphFixtures.swift`; one `waitForQuery`; greps clean |
| MNT-2 | MNT | major | (round 1) | Resolved | `rg 'sourceSurface: \.page' macos` → only `WorkspaceLocation.swift:177`, `FakeStore.swift:2884` |
| MNT-3 | MNT | major | (round 1) | Resolved | `SourcePageModel.swift:187` via `session.evidenceGraph(id:)` |
| MNT-4, MNT-5 | MNT | minor | (round 1) | Resolved | doc comment reworded; one `sourcePageSectionTopPadding()` |

No new findings. Nits: none. No Fix plan 2.

## Finish
Merged `origin/main` (`c7c78526`, S11-37 plan docs only) as `d6adeee1`; gate PASS.

## After ready — user report
The user reported a Swift 6 warning: `WorkspaceToolbar.swift:198` `BreadcrumbNames.none` is not concurrency-safe. `check-pr.sh`'s incremental `xcode-test` does not surface warnings, so the gate missed it.

**F-9 — `BreadcrumbNames.none` as a computed static** · head `cdf20523` · Gate PASS (orchestrator run). The implementer also marked `WorkspaceToolbarBreadcrumbTests` `@MainActor` to clear the "sending 'self.names'" warnings from the suite's new `names` property (and older main-actor call warnings in the same suite). A clean `build-for-testing` leaves 6 warnings in changed files, all in `WorkspaceSessionTests.swift` (55, 62, 69, 169, 847, 877), from `c07f6c02c` and `d4e666b8c` on `main`; left as is.

## Result
Gate PASS · CI green · open: none · follow-ups: none (UX-3 written into S12-12 scope) · to user: — · CON miss: none
