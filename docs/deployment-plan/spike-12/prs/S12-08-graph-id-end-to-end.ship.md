# Ship log — S12-08 Graph id end to end

**Plan:** [`S12-08-graph-id-end-to-end.md`](S12-08-graph-id-end-to-end.md) · **Branch:** `s12-08-graph-id-end-to-end` · **PR:** #528
**State:** ready

## Round 0 — implement
Base `e33db0f0`. Head `b8ae1245`. Gate PASS (orchestrator run, Go test cache cleared). Report deviations:
- Base had moved 131 commits since the Planned commit `3c2ff0da` (S11-20 to S11-26 Promote work); no Current-state row was false.
- Handler tests get graph ids through the engine (`soleGraphID`, `graphLinking` over `ListEvidenceGraphs`), not `evidencegraphstest`: they hold a project dir, not a catalog.
- `FakeStore` is lenient where the engine refuses: `createSubject` on an unknown graph does not throw `subjects.invalid`; `sourceOfGraph` falls back to a Subject's `sourceID`; `bumpSource` reports graphs even when the Source is in no list; `createSource` reports no graph. `FakeStore` now records `getCitation <id> graphID=<id>`.
- `subjects.Create` no longer checks `SourceID` itself; `InsertTx` holds the one rule.
- `mutatedSourceGraph` renamed then deleted, as D-8 says.
- `TestSwiftVocabularyMatchesGo` path fixed to the renamed file (not named in the plan).
- Manual check (§10) not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `api/ffi/**`, `core/database/**`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1, MNT-1 | PRF, MNT | major | Sources list scans the whole graph list per row (status filter O(N×G) on the main actor) | Fix now → F-1 | verified: `SourcesModel.visibleSources` calls `entryGraph(for:)` per Source; `SourcesListNavigation.entryGraph` is `graphs.first { … }`; PRF measured 57 ms/pass at 2,000×2,000 |
| MNT-3 | MNT | minor | Two models each read the graph list and call `graphLocation` | Fix now → F-1 | same-pass cleanup: `SourcePageModel.swift:67-71` and `SourcesModel.swift:160-166` copy the read; F-1 adds the one helper |
| UX-1 | UX | major | Sources list counts, status filter and error ignore the graph list's loading/error state | Fix now → F-2 | verified: `graphCountsLoading`, `isAwaitingFilterData`, `statusCountsError` read only `progressHandle`; `graphs` defaults to `[]`; Source page tooltip says "needs an Artifact" whenever `location == nil`. D-12's "disabled while loading" stands; the wrong counts and swallowed error are not in D-12 |
| MNT-2 | MNT | major | FakeStore has `graphOfSubject` plus four inline copies | Fix now → F-3 | same-pass cleanup: `grep -n 'subjectsByGraph.first(where' FakeStore.swift` → 1607, 2470, 2632, 3431, 3539 (helper body) |
| UX-2 | UX | minor | Disabled Open graph zone does not say why | Fix now → F-4 | minor in a file this PR changes. Not via the zone condition (D-12 keeps the zone on `hasArtifact`); via help and hint |
| MNT-4 | MNT | minor | `WorkspaceLocation.evidenceGraph(…)` lacks `subjectId`; FakeStore mutates it after | Fix now → F-5 | verified: `FakeStore.swift:2904` is the only `.subjectId =` in `macos/App` |
| MNT-5 | MNT | minor | "First graph by name" has three tie-breaks | Fix now → F-6 | verified: `projectors.go` `ORDER BY g.name COLLATE NOCASE, g.id` vs `evidencegraphs.sqlListOrder` `name COLLATE NOCASE, ref, id` |
| MNT-6 | MNT | minor | FakeStore graph literal written twice | Fix now → F-7 | same-pass cleanup: `GPH-FAKE1` at 753 and 3503 |
| CON-1, CON-2 | CON | minor | FakeStore lenient where the engine refuses; `createSource` reports no graph | Fix now → F-8 | D-14 says "FakeStore mirrors the engine … report the same graphs the engine reports"; the leniency defeats the distinct-ids check the plan relies on |
| MNT-7 | MNT | nit→minor | Stale `valueSource` / `ownedValue` comments in `effects/registry.go` | Fix now → F-9 | stale comment in a file this PR changes |
| PRF-2 | PRF | minor | `.deletedSource` reloads every cached graph, hidden ones too | Follow-up | D-8 chose `.deletedSource → .allCached(.evidenceGraph)`; quiet path, cost bounded by graphs visited. Idea: resolve the Source's graph ids from `.evidenceGraphsList`, or `whenHidden: .markStale` on `.evidenceGraph` |

Nits (logged): `CON-3 · CatalogQueryRegistry.swift · comment that .deletedSource → .allCached is wider by D-8` · `CON-4 · deleteimpact/projectors.go:180-200 · Citation with no Observation and no graph-linked Artifact falls back to the Source page (unreachable while Default links every Artifact)` · `CON-5 · §10 manual check not run` · `UX-3 · SourcesSplitRow.swift:118 · { onOpenGraph?() } can't run with nil` · `MNT-8 · deleteimpact/kinds.go:42-45 · GraphID doc comment splits field alignment`

### Fix plan 1

**F-1 — Group the graph list by Source once; one helper for both models** (from PRF-1, MNT-1, MNT-3)
Files: `macos/App/Features/Sources/SourcesListNavigation.swift`, `SourcesModel.swift`, `SourcePageModel.swift`, `SourcesListView.swift`, tests in `macos/ProvenenciaTests/SourcesListNavigationTests.swift`. Skill: `add-swift-test`.
Failing test first: `SourcesListNavigationTests.graphsBySourceKeepsListOrder` — two Sources, graphs [A·Default, B·Default, A·Jones] (list order by name) → map `A: [Default, Jones]`, `B: [Default]`; `entryGraph` reads `.first`.
Do: add `SourcesListNavigation.graphsBySource(_ graphs: [CatalogEvidenceGraph]) -> [String: [CatalogEvidenceGraph]]` (one pass, keeps list order). `entryGraph(sourceID:in:)` and `graphLocation(for:graphs:)` take the grouped map (dictionary read, `.first`). Add one session helper on `SourcesListNavigation` (e.g. `graphsBySource(in session: WorkspaceSession)`) that reads `.evidenceGraphsList`, and have both `SourcesModel` and `SourcePageModel.evidenceGraphLocation` call it; delete `SourcesModel.graphs` and the private `evidenceGraphsKey(for:)` wrapper if it has no other caller. In `SourcesModel.visibleSources`, build the map once per pass before the filter. Memoize the map in `SourcesModel` against the handle's value (as `searchKeyCache` is reused) so the row builders do a dictionary read.
Check: the test passes; `rg -n 'graphs\.first \{' macos/App/Features/Sources` prints nothing; scripts/check-pr.sh PASS.

**F-2 — The Sources list's loading and error state follow the graph list too** (from UX-1)
Files: `macos/App/Features/Sources/SourcesModel.swift`, `SourcesListView.swift`, `SourcePageIdentityHeader.swift`; tests in `macos/ProvenenciaTests/SourcesModelTests.swift` (or the existing Sources list test file). Skill: `add-swift-test`.
Failing test first: `SourcesModelTests.graphListFailureShowsTheStatusError` — FakeStore fails the `.evidenceGraphsList` read; progress loads; status filter Not started set → `statusCountsError` is non-nil and `isAwaitingFilterData` false; and `graphListLoadingKeepsCountsLoading` — graph list pending, progress ready → `graphCountsLoading` true, `isAwaitingFilterData` true with a status filter.
Do: `graphCountsLoading` is true while either the progress handle or the graph-list handle has no value and has not reached `.ready`/`.error`. `statusCountsError` returns the graph-list handle's error when it failed (else the progress one). `SourcesListView` passes that combined flag as `countsLoading:` to `SourcesSplitRowProgress` (not the progress handle alone). `SourcePageIdentityHeader.evidenceGraphJump`: use `needsArtifactTooltip` only when `!source.hasArtifact`; otherwise keep `jumpToEvidenceGraph` as the help. No new strings. Use the FakeStore's existing error-hook pattern for the progress read to fail the graph-list read (add a hook if none exists for it).
Check: both tests pass; scripts/check-pr.sh PASS.

**F-3 — FakeStore calls `graphOfSubject` everywhere** (from MNT-2)
Files: `macos/App/Platform/FakeStore.swift`.
Do: replace the inline scans at the four sites (today 1607, 2470, 2632, 3431) with `graphOfSubject(_:)`, keeping each site's nil handling (`guard`, `?? ""`).
Check: `grep -n 'subjectsByGraph.first(where' macos/App/Platform/FakeStore.swift` prints only the `graphOfSubject` body; scripts/check-pr.sh PASS.

**F-4 — A disabled Open graph zone says why** (from UX-2)
Files: `macos/App/Features/Sources/SourcesSplitRow.swift`.
Do: keep D-12's zone condition (`source.hasArtifact`). When `onOpenGraph == nil` and the counts are not loading (F-2's flag), add `.help(L10n.string(L10n.Sources.needsArtifactTooltip))` and the matching `.accessibilityHint`, as the no-Artifact branch does. Existing strings only. Fold UX-3: pass the optional and call it with a single guard.
Check: scripts/check-pr.sh PASS (xcode-test, localizable).

**F-5 — `WorkspaceLocation.evidenceGraph(…)` takes `subjectId`** (from MNT-4)
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Platform/FakeStore.swift`.
Do: add `subjectId: String? = nil` to the `evidenceGraph` factory, passed through the initializer (so `nilIfEmpty` applies). The FakeStore delete-impact fixture passes it instead of assigning `location.subjectId` afterwards.
Check: `rg -n '\.subjectId = ' macos/App` prints nothing; scripts/check-pr.sh PASS.

**F-6 — One tie-break for "first graph by name"** (from MNT-5)
Files: `core/database/deleteimpact/projectors.go`, `macos/App/Platform/FakeStore.swift`.
Do: `sqlFirstArtifactGraph` orders by `g.name COLLATE NOCASE, g.ref, g.id`, with a comment naming `evidencegraphs.sqlListOrder`. FakeStore `firstGraphID(ofArtifact:)` sorts by case-insensitive name, then ref, then id.
Check: existing deleteimpact tests pass; scripts/check-pr.sh PASS.

**F-7 — One FakeStore graph append** (from MNT-6)
Files: `macos/App/Platform/FakeStore.swift`.
Do: add a private `appendGraph(projectDir:sourceID:name:description:artifactIDs:) -> CatalogEvidenceGraph` and call it from `createEvidenceGraph` and `insertSource`.
Check: `grep -n 'GPH-FAKE1' macos/App/Platform/FakeStore.swift` prints one line; scripts/check-pr.sh PASS.

**F-8 — FakeStore refuses and reports as the engine does** (from CON-1, CON-2; D-14)
Files: `macos/App/Platform/FakeStore.swift`, the Swift tests that seed Subjects by graph id without a graph. Skill: `add-swift-test`.
Failing test first: `EvidenceGraphStoreTests.createSubjectOnAnUnknownGraphThrows` — `createSubject(graphID: "src-1")` where `src-1` is a Source id and no graph has that id → throws. And extend `EvidenceGraphStoreTests.createSourceMakesADefaultGraph` to expect the created Source's effects to name the Default graph.
Do: `createSubject` throws (the fake's `subjects.invalid` equivalent, as other fake refusals do) when `sourceOfGraph(graphID)` is nil. Remove `sourceOfGraph`'s fallback to a Subject's `sourceID` (graph's Source comes from `evidenceGraphsByProject` only). `createSource` reports `[graph.id]` (delete the `reportsGraph:` parameter if it then has one value). Tests that relied on the leniency seed a `CatalogEvidenceGraph` literal in `evidenceGraphsByProject` with a graph id distinct from the Source id, as D-14 asks; tests counting effects after `createSource` adjust by the Default graph only.
Check: both tests pass; scripts/check-pr.sh PASS.

**F-9 — Effects registry comments name the graph path** (from MNT-7)
Files: `core/database/effects/registry.go`.
Do: update the `valueSource` comment ("CitationSource and SourceScope walk it") and the `ownedValue` comment to match the graph-keyed effects.
Check: scripts/check-pr.sh PASS.

Update the plan's Landed note with F-8 (FakeStore no longer lenient) replacing the leniency deviation.

## Round 1 — fix
Head `43b0b098`. Gate PASS (orchestrator run). F-1 to F-9 done. Report deviations:
- F-2's two tests were not seen failing first (code written before the tests ran); they pass.
- F-8 broke `PromoteModelTests.aStaleDoneReproposes` and `CatalogQueryRegistryTests.conclusionDetailLoadsAPersonsNameValuesAndOutcomes`, which relied on the leniency; both now seed a `CatalogEvidenceGraph` literal (D-14).
- FakeStore's `listEvidenceGraphs` shares the new `graphListOrder` (name, ref, id).
- Added `FakeStore.listEvidenceGraphsError` for F-2's test.
- Plan Landed note: leniency line replaced.

Minors whose reviewer does not return: CON-1, CON-2 — **Resolved** (read `FakeStore.swift` diff: `createSubject` throws `subjects.invalid` when `sourceOfGraph` is nil; no Subject fallback; `insertSource` always reports the Default graph, `reportsGraph:` removed).

## Round 2 — review
Reviewers: MNT, UX, SEC+PRF (each had a Fix-now major). Not sent: CON (minors only). Scope `b8ae1245...43b0b098`.

| ID | Lens | Round-1 finding | Status | Evidence |
| --- | --- | --- | --- | --- |
| PRF-1, MNT-1 | PRF, MNT | Per-row graph scan | Resolved | both reviewers: `graphsBySource` one pass, cached on `QueryHandle.revision`; `.evidenceGraphsList` is `.markStale`, never evicted |
| MNT-2 | MNT | FakeStore inline scans | Resolved | only `graphOfSubject` body matches the grep |
| UX-1 | UX | List ignores graph-list loading/error | Resolved | combined `graphCountsLoading`; `statusCountsError` returns the graph-list error; Source page tooltip on `hasArtifact` |
| MNT-3 | MNT | Two reads of the graph list | Resolved, partly | `SourcePageModel` uses `graphsBySource(in:)`; `SourcesModel` reads the handle itself because its cache needs `revision` |
| MNT-4 to MNT-7 | MNT | Minors | Resolved | MNT round-2 checks |
| UX-2 | UX | Disabled Open graph says why | **Not resolved** | F-4's hint is the "add an artifact" string; it shows for a Source that has an Artifact when the graph list failed with no status filter, or when the entry graph links no Artifact. The Source page shows "Jump to evidence graph" in the same state (MNT outside-lens note). Minor: no Fix plan 2 (round 2 fixes blockers only) → user |

New blockers: none (MNT, UX, SEC+PRF all Clean).
Outside-lens notes (logged): `UX → MNT/PRF · SourcesModel.swift:162-170 · cache keyed on revision only; safe while .evidenceGraphsList is never evicted` · `MNT → UX · SourcePageIdentityHeader.swift:201 · disabled jump with an Artifact gives no reason`

## Result
Gate PASS · CI green · open: none · follow-ups: PRF-2 · to user: UX-2 (answered: fixed, U-1) · CON miss: none

## User fix — UX-2
The user asked (2026-10-23): "Fix UX-2 on PR #528: show the disabled Open graph hint only when the graph list has a value, and give the Source page's disabled jump the same reason."

### Fix plan U
**U-1 — One rule for why Open graph / Jump to evidence graph is blocked** (from UX-2, MNT outside-lens note on `SourcePageIdentityHeader.swift:201`)
Files: `macos/App/Features/Sources/SourcesListNavigation.swift`, `SourcesModel.swift`, `SourcePageModel.swift`, `SourcesSplitRow.swift`, `SourcesListView.swift`, `SourcePageIdentityHeader.swift`; tests in `macos/ProvenenciaTests/SourcesListNavigationTests.swift`. Skill: `add-swift-test`.
Failing test first: `SourcesListNavigationTests.graphBlockedReasonNeedsTheGraphList` — (a) no Artifact → the needs-Artifact reason; (b) has an Artifact, graph list has no value (loading or failed), no location → nil; (c) has an Artifact, graph list has a value, no location → the needs-Artifact reason; (d) a location → nil.
Do: add one function on `SourcesListNavigation`, e.g. `graphBlockedReason(hasArtifact: Bool, graphListLoaded: Bool, location: WorkspaceLocation?) -> String?`, returning `L10n.string(L10n.Sources.needsArtifactTooltip)` per the cases above. "Graph list has a value" is `.evidenceGraphsList`'s handle `value != nil`; expose it once (e.g. a `graphListLoaded` on `SourcesModel` reading the handle it already holds, and the same on `SourcePageModel` through the session helper's handle) rather than re-deriving it at each view. `SourcesSplitRow`: the disabled zone's `.help` / `.accessibilityHint` come from that function (passed in from `SourcesListView` as the reason, or as the inputs), replacing `graphBlockedReason`'s `countsLoading` test; no hint when it returns nil. `SourcePageIdentityHeader.evidenceGraphJump`: `.help(reason ?? L10n.string(L10n.Sources.jumpToEvidenceGraph))` and the same `.accessibilityHint` when the reason is non-nil. Existing strings only; no new copy.
Check: the test passes; `rg -n 'needsArtifactTooltip' macos/App/Features/Sources` shows the reason coming from the one function at the row's graph zone and the page's jump (the row's separate no-Artifact branch may keep its own use); scripts/check-pr.sh PASS.

## User fix — implement
Head `9a12e073`. Gate PASS (orchestrator run). U-1 done: `SourcesListNavigation.graphBlockedReason(hasArtifact:graphListLoaded:location:)` feeds the row's disabled zone and the Source page's jump (`.help` and `.accessibilityHint`); no hint while the graph list has no value. Test `SourcesListNavigationTests.graphBlockedReasonNeedsTheGraphList` (cases a–d); it could only fail to compile before the function existed. Orchestrator read the diff: matches the user's ask. UX-2 **Resolved**.
