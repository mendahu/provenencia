# Ship log — S14-11 Reload only what moved, only when it is on screen

**Plan:** [`S14-11-reload-only-what-moved.md`](S14-11-reload-only-what-moved.md) · **Branch:** `s14-11-reload-only-what-moved` · **PR:** #657
**State:** fixing 1

## Round 0 — implement
Base `main` @ `fdd37f19`.
Stopped on the base check (implement-pr round 0 step 2), before any code change; no branch, no commit, no PR. `main` moved 42 commits past the Planned `85a5ce2`; S13-01 (#655) changed:
- `CatalogQueryDescriptor.refresh: CatalogQueryRefresh?` → `refreshes: [CatalogQueryRefresh]` (verified: `CatalogQueryRegistry.swift:42`); `CatalogQueryRefresh` is now `{ follows, action: .revalidate | .run }`.
- `applyHiddenPolicy(follows:id:)` → `applyHiddenPolicy(follows: CatalogQueryFollows, ids:)`, plus `followsValue:` and `applyHiddenPolicyOnAny` variants.
- `CatalogQueryFollows` gained `.touchedHandles` and `.evidenceGraphs`.
- `WorkspaceSession.apply` no longer calls `apply(invalidations:)`; it dedupes `invalidated + revalidatedKeys(for:)` and runs `.run` refreshes in their own loop. The registry no longer emits per-graph `.evidenceGraph` keys; that arm now revalidates via `refreshes: [.applyHiddenPolicy(follows: .evidenceGraphs)]`.
- Smaller: `.personClaims` key kind; `"sources"` effect now reports `EvidenceGraph` on title/source_type change; `listPersonClaims` on the stores.

Affected: §4 rows (patch/refresh slots, Follows, Session, progress refresh, graph list arm), D-5–D-8, D-10, D-11, STEPs 4–7, T-12, T-15. D-1–D-4, D-9 look unaffected. Next: plan refresh with `plan-pr` against `fdd37f19`, then resume this log.

Plan refreshed with `plan-pr` against `fdd37f19` (2026-11-01); `review-plan` rounds 3–4 in the review log, no open blocker. Resumed: round 0 again.

Base `main` @ `fdd37f19`. Head `1e23f521`. Gate PASS (orchestrator run; xcode-test 66s). First commit `22fb7d07` carries the plan refresh. Report deviations (in the plan's Landed note): `TestListsValidate` in `registry_completeness_test.go` (unexported `validate`); T-4, T-7, T-8 split into more tests; tests that relied on a hidden graph reloading now apply the canvas location in setup only (`EvidenceGraphModelTests.makeModel`, `CitationComposerModelTests.makeModel`, `applyCitationMutationReloadsGraphWithoutASecondQuery`); two unplanned assertion changes (`hiddenPolicies` expects `.markStale` for `sourceWorkspace`; `createRefreshesListThroughSession` → `createInsertsTheRowThroughSession`); `CatalogPatchResult` landed in the STEP-4 commit; a comment on `CatalogMutation.evidenceGraphId`. This log was committed in `e52153b1` by the implementer (meant to stay untracked; harmless, it belongs on the branch).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. UX: the diff touches `macos/App/**` (Sources models, session cache). SEC+PRF: `api/ffi/**` (new `source_id` parse) and session cache / `Features/**` models.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | Each graph or link write rebuilds the whole `EvidenceGraphIndex` on the main actor | Fix now → F-1 | verified: `WorkspaceSession+EvidenceGraphs.swift` `currentEvidenceGraphIndex` rebuilds all three maps on any revision move, and the merge bumps the revision. A-9 (`app-health-audit.md:112`) names "rebuilds `EvidenceGraphIndex` on the main actor" as part of the problem this PR fixes; measured 2.45 ms per write at 3,060 graphs (the map's Scale), 8.97 ms at 12,240. The plan's Scale bullet accepted it because the old reload paid it too, which is not a reason to keep it. Code this PR adds |
| MNT-1 | MNT | major | `.merged` is not the only way to build a `.run` entry | Fix now → F-2 | verified: `CatalogLastUpdated.swift` `init(follows:action:)` and `Action.run` are internal; only the four factories call the designated init. D-6 (c) promised one way |
| MNT-2 | MNT | major | `VisibleRowsToken` is a fourth hand-written token memo | Fix now → F-3 (same-pass cleanup) | this PR adds a copy; copies: `CitationComposerContext.MemoToken`/`refreshMemo`, `DepictionPreviewMemo`, `SourcesModel.VisibleRowsToken`. Supersedes the plan's Out line "A shared `Memo` type" and the STEP-9 `structural-debt.md` line; no behavior a researcher sees changes |
| MNT-3 | MNT | minor | `refreshCached` is callable outside `apply`'s gate | Fix now → F-4 | verified: only caller is `.merged`; file this PR adds to |
| MNT-4 | MNT | minor | New Go test hand-marshals and repeats a check | Fix now → F-5 | verified: `rpc` helper in `harness_test.go:103`; the only `proto.Marshal(&engine…Request` in handler tests |
| UX-1 | UX | minor | A revisited canvas swaps in fresh cards with no reloading sign | Fix now (doc) → F-6; indicator → Follow-up | verified: `isRereadingGraph` is drawn only in `readErrorCallout`. Plan Reach row 1 and Manual check 3 claim it shows; correct them. An indicator is chrome and needs a brief |
| CON-1 | CON | minor | Rows memo could hit after a handle is evicted and recreated | Rejected | none of the four keys in the token evicts: `.sourcesList` and `.evidenceGraphsList` are `.markStale`, `.sourceTypesList` and `.evidenceGraphProgress` `.revalidate`; `evict` is reached only through `applyHiddenPolicy`'s `.evict` arm (`WorkspaceSession.swift:376`) |

Nits (logged): `CON-2 · WorkspaceSession.swift refreshCached · dropped patchCached result, add a comment` · `MNT-5 · lists_test.go:51 · rename "graph then vocabulary"` · `MNT-6 · CatalogLastUpdated.swift:98 · isCreate as a let` · `MNT-7 · WorkspaceSession.swift:224 · shorten apply's doc, point to the skill`

### Fix plan 1

**F-1 — Splice one Source into the graph index instead of rebuilding it** (from PRF-1)
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `CatalogLastUpdated.swift` (`.merged`), `CatalogQueryRegistry.swift` (`.evidenceGraphsList` arm), `WorkspaceSession.swift` (`refreshCached`), `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Failing test first: `WorkspaceSessionTests.graphListMergeSplicesTheIndex` *(new)* — warm the list (S: 2 graphs, T: 1), read `evidenceGraph(id:)` so the index is built, apply `reads: [evidence_graphs: [S]]` with S now 3 graphs, `settleWrites()`; then `evidenceGraphIndexBuildCount` *(new test hook, counts full `EvidenceGraphIndex(revision:rows:)` builds)* is unchanged, `evidenceGraphs(sourceID: S)` returns the 3 graphs in the re-read's order, T's graphs and `linkedArtifactIDs` for T are unchanged, and `evidenceGraph(id:)` finds a new S graph and no longer finds a removed one.
Do:
- `EvidenceGraphIndex.replacing(sources: [String: [CatalogEvidenceGraph]], revision: Int) -> EvidenceGraphIndex` *(new)*: for each Source given, removes that Source's old ids from `byID`, sets `bySource[id]` (dropping the key when empty) and `linkedArtifactIDsBySource[id]`, and adds the new ids to `byID`. No other Source is touched.
- `.merged` gains an optional trailing `afterMerge: (@MainActor @Sendable (WorkspaceSession, CatalogQueryKey, _ parts: [String: Part], _ previousRevision: Int) -> Void)? = nil`; `refreshCached` calls it after `patchCached` answers `.applied`, passing the handle revision read just before the patch. Progress passes none.
- The `.evidenceGraphsList` arm passes `afterMerge: { session, _, parts, previous in session.spliceEvidenceGraphIndex(parts, previousRevision: previous) }`. `spliceEvidenceGraphIndex` *(new, in `WorkspaceSession+EvidenceGraphs.swift`)*: if the stored `evidenceGraphIndex?.revision == previousRevision` and the handle's revision is now `previousRevision + 1`, store `index.replacing(sources: parts, revision: handle.revision)`; otherwise do nothing (the next read builds in full, as today).
- `currentEvidenceGraphIndex` keeps the full build for every other revision move (full loads).
Check: the test passes; `graphListMergesOneSourcesReRead` and the T-7 tests still pass; scripts/check-pr.sh PASS.

**F-2 — Make `.merged` the only way to declare a partial reload** (from MNT-1)
Files: `CatalogLastUpdated.swift`.
Do: make `CatalogQueryRefresh.init(follows:action:)` `fileprivate` (the factories in the same file stay the only builders), and keep the doc comment "`merged` is the only way to declare one" true. No test change expected.
Check: `rg -n "CatalogQueryRefresh\(" macos` shows only `CatalogLastUpdated.swift`; scripts/check-pr.sh PASS.

**F-3 — One `Memo` type for the token memos** (from MNT-2; same-pass cleanup)
Files: `macos/App/Platform/Memo.swift` *(new, add to the app target in `project.pbxproj`)*, `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Features/CitationComposer/CitationComposerContext.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`, `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-14/prs/S14-11-reload-only-what-moved.md` (Landed note). Skill: `add-swift-test`.
Failing test first: `macos/ProvenenciaTests/MemoTests.swift` *(new)* `buildsOncePerKey` — `var memo = Memo<Int, String>()`; `value(for: 1) { "a" }` twice → build closure runs once; `value(for: 2)` runs it again; `value(for:)` (lookup) on a missing key → nil; `store("b", for: 3)` then lookup 3 → "b".
Do:
- `struct Memo<Key: Equatable, Value>` *(new)*: `private var entry: (key: Key, value: Value)?`; `func value(for key: Key) -> Value?` (the stored value when the key matches); `mutating func store(_ value: Value, for key: Key)`; `mutating func value(for key: Key, build: () -> Value) -> Value` (lookup, else build and store). Doc: "The last value built, kept while its key is unchanged. A key holds every input the build reads."
- `SourcesModel`: `VisibleRowsToken` stays as the key; `visibleRowsToken` and `visibleRows` become one `@ObservationIgnored private var visibleRowsMemo = Memo<VisibleRowsToken, [CatalogSource]>()`; `visibleSourcesBuildCount` is incremented inside the build closure.
- `CitationComposerContext`: `MemoToken` becomes the key with its `terms` computed up front each read (`termRevisions(in: fieldsHandle?.value ?? .empty)`, the same reads `termRevisionsMatch` does today); `memoToken`, `memoSnapshot` and `memoVocabulary` become one `Memo<MemoToken, (snapshot: EvidenceGraphSnapshot, vocabulary: CitationComposerVocabulary)>`; `refreshMemo` reads it with `value(for:build:)`. Behavior unchanged; its existing tests stay green.
- `DepictionPreviewMemo` is deleted; `depictionPreviewMemo` becomes `Memo<DepictionCutKey, DepictionCut?>` with `struct DepictionCut { let cgImage: CGImage; let image: NSImage }` *(new, replaces the tuple)*; `depictionPreviewState` uses `value(for:)` (`.cutting` when nil, `.failed` when the stored cut is nil); `depictionImage(for:)` keeps its "store only while the draft is still on screen" guard and calls `store(_:for:)`.
- Delete the `Memo` line from `docs/ideas/structural-debt.md`. The Landed note says this supersedes the plan's Out line and STEP-9 log line.
- `EvidenceGraphIndex` is not moved: after F-1 it splices on a merge, which `Memo` does not model.
Check: `rg -nE "private struct \w*Token: Equatable" macos/App` prints only `SourcesModel.VisibleRowsToken`, `CitationComposerContext.MemoToken` (both now `Memo` keys) and `EvidenceGraphView.PreviewToken` (a change detector, not a memo); `rg -n "DepictionPreviewMemo|memoSnapshot|visibleRows\b" macos` prints nothing; scripts/check-pr.sh PASS.

**F-4 — Mark `refreshCached` as `.merged`'s entry point** (from MNT-3)
Files: `WorkspaceSession.swift`.
Do: its doc says it is called only by `CatalogQueryRefresh.merged`, inside `apply`'s gate and count, and is not a feature API. If `.merged` can move into a `WorkspaceSession` extension in `WorkspaceSession.swift` without moving the other factories, make `refreshCached` `fileprivate` instead; otherwise the doc line.
Check: scripts/check-pr.sh PASS.

**F-5 — Use the RPC harness in the new handler test** (from MNT-4)
Files: `api/ffi/handlers/evidence_graphs_test.go`.
Do: the "bad source id" case calls `rpc(t, ListEvidenceGraphs, &engine.ListEvidenceGraphsRequest{ProjectDir: f.dir, SourceId: "x"}, &engine.ListEvidenceGraphsResponse{})` and checks `errors.Is(err, errInvalidID)`; drop the `proto` import. Keep one per-Source assertion (in `TestListEvidenceGraphsSourceID`), and drop the duplicate in `TestListEvidenceGraphs`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run ListEvidenceGraphs` passes.

**F-6 — Say what a revisited canvas shows** (from UX-1)
Files: `docs/deployment-plan/spike-14/prs/S14-11-reload-only-what-moved.md` (§6 Reach row 1, §10 Manual check 3, Landed note).
Do: Reach row 1 and Manual check 3 say the canvas shows its cached cards with no reloading sign and swaps in the current ones when the reload lands (`isRereadingGraph` is drawn only in the read-error callout). No code change.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-11-reload-only-what-moved.md` reports no errors.

Follow-ups: UX-1 (a reloading sign on a revisited canvas, driven by `isRereadingGraph`; needs a brief; suggested home `docs/ideas/`).
