# Review — S14-11 Reload only what moved, only when it is on screen

**Verdict:** Fits with changes — Not ready → Fixed (no open blocker after round 2; refresh pass against `fdd37f19`: Fits with changes — Not ready → Fixed, no open blocker after round 4)
**Rounds:** 2, then 2 more for the refresh (3 and 4). **Reviewers:** maintainability (Opus), executability (Sonnet), fresh each round

## To user
None. Every finding was answerable from the code; no product call came up.

## Writer's own calls
- (Refresh, round 3–4.) `.merged` builds a `.run` entry; no new `CatalogQueryRefresh.Action` case. `init(follows:run:)` becomes `fileprivate`, so `.merged` is the only way to declare a partial reload.
- (Refresh.) `apply` decides synchronously whether a `.run` key merges or falls back, and takes and releases the handle's refresh count around each run, so `isFetching` is true before `apply` returns.
- (Refresh.) A stale key queried while a load is in flight restarts that load (D-5), since `isSettled` now reads the stale flag.
- (Refresh.) `patchCached` reports `.awaitingFirstValue`, and `apply` falls back for those keys: every patch on a key whose first load is in flight now restarts it (visible) or marks it stale (hidden) instead of dropping the write. This reaches the Persons, Events, Places and Sources lists and the Source page, not only the Sources-row create.
- (Refresh.) A visible graph list or progress map left in `.error` with no value is retried by the next graph write, as today.
- The read is named `evidence_graphs` (the list it replaces goes in the same PR), keyed by Sources.
- One `source_id` per `ListEvidenceGraphs` call, not a repeated field.
- A hidden `.markStale` key skips the partial merge and goes stale; graph progress (`.revalidate`) keeps merging while hidden, as today.
- A failed partial re-read reloads the whole key (graph list or progress map) instead of keeping old rows; for progress this changes today's silent catch.
- No generation check before a merge (D-7): the read started after the write, and a compare would drop one Source's merge whenever another's landed first.
- The merged graph list appends a Source's re-read rows at the end; no reader uses the flat order.
- `EvidenceGraphStoreTests.writesReturnScriptedEffects` scripts `[.vocabulary]` instead of the removed list.
- The `visibleSources` memo stays local; a shared `Memo` type is logged in `structural-debt.md`.

## Round 1

### M-1 · blocker · `refreshCached` is a second partial-refresh mechanism beside the inline progress refresh
**Answer:** Fixed. D-6 adds the factory `CatalogQueryRefresh.merged` over `refreshCached`, and moves both the graph list and progress onto it. The progress inline closure goes (Deletes). The hidden policy uses option (b): merge unless the kind's `whenHidden` is not `.revalidate`, so the Out row stands. New tests: AC-1.7, T-7b. STEP-4 now moves progress.

### M-2 · blocker · Two owners of "stale"; the rejection premise was unchecked
**Answer:** Fixed. Verified: `readyValue` returns `queryHandle(key)?.value`, `ensureQuery` creates the handle first, `applyHiddenPolicy` guards on the handle, and `evict` removes both. D-5 now makes the handle the only owner. `invalidatedKeys` is deleted, and `QueryHandleResetting` is repurposed as `CachedQuery` (`isStale`, `markStale`, `clearStale`). D-5 lists every site; a new Stop condition guards the premise.

### M-3 · blocker · Partial re-read invisible to `readyValue`
**Answer:** Fixed. D-6 records each refresh in `refreshes[key]`. D-7 has `readyValue` await those tasks after `inFlight`. Reach names the composer's `readyValue` at `CitationComposerModel.swift:1252`, and T-7(a) asserts that `readyValue` waits for the merge. Registering the refresh in `inFlight` itself was rejected, because `ensureQuery` returns early while `inFlight[key]` is set.

### M-4 · blocker · Hand-rolled optional id parse beside `optionalID`
**Answer:** Fixed. D-3 uses `optionalID` and moves it to `catalog.go` beside `parseID`. STEP-2 includes the move.

### M-5 · major · `insertedSourceRow` near-copies `updatedSourceRow`
**Answer:** Fixed. Verified that `updatedSourceShell` returns early for an unwarmed page. D-10 now returns `.createdSource` from `patchedSource` and renames `updatedSourceRow()` to `sourceRow()`, which replaces or appends.

### M-6 · major · STEP-9 and sweep miss guidance
**Answer:** Fixed.
- STEP-9 now covers `add-workspace-place` lines 156 and 173, and the Mark stale row.
- D-8 corrects the `S12-32` comment.
- The DoD sweep adds `isSettled`, `S12-32 removes`, `updatedSourceRow`, `invalidatedKeys` and `QueryHandleResetting`.

### M-7 · minor · Mutation-to-read shims grow inside a generic switch
**Answer:** Fixed. D-8 adds `CatalogMutation.readIDs(for:)`, and both shims live there until S14-16.

### M-8 · minor · Fourth token memo
**Answer:** Logged. Codebase fit lists all four kin. STEP-9 adds a line to `structural-debt.md`, and a new Out row points there. Extracting now would touch three files outside this PR.

### M-9 · minor · D-5 extends dead code
**Answer:** Fixed. `QueryHandleResetting` is repurposed as `CachedQuery`. `resetForInvalidation` leaves the protocol unchanged and gets no stale handling.

### X-1 · blocker · Removing `CatalogList.evidenceGraphs` breaks unnamed test files
**Answer:** Fixed. Verified with `rg`. §4 adds a "Swift tests on the graph list" row naming every site. STEP-5 says what each site becomes, lists the files and runs them in its check.

### X-2 · blocker · Sweep regex misses `.evidenceGraphs` in `lists:`
**Answer:** Fixed. The STEP-5 check and the DoD sweep use `lists: \[[^]]*\.evidenceGraphs`.

### X-3 · blocker · Cannot call `markStale()` on a type-erased handle
**Answer:** Fixed by M-2. `handles` becomes `[CatalogQueryKey: any CachedQuery]`. A key with no handle has nothing to mark, and D-5 says why that is safe.

### X-4 · major · D-5 mixed insert and remove sites; `isSettled` readers undercounted
**Answer:** Fixed. D-5 is now written out site by site. §4 adds an `isSettled` row naming every test. STEP-4's check runs all of those suites, and the Risks row points to it.

### X-5 · major · `refreshCached` cases left open
**Answer:** Fixed.
- No value but first load in flight: D-6 step (1) restarts the load (T-7c).
- Merge with no change: the merge returns `Void` and always writes through `replaceValue`.
- Invalidated, evicted or re-queried during the await: D-7 explains why there is no generation check.
- Several ids: they are read sequentially in one call with one merge, and the `refreshes` count keeps `isFetching` true across overlapping calls.

### X-6 · major · T-7 needs a FakeStore hold
**Answer:** Fixed. D-4 adds `listEvidenceGraphsHold` and records the `sourceID` before the error throw. It also adds `listSourcesCalls` and `listSubjectsGraphIDs`, which T-8 and T-15 need.

### X-7 · major · Manual check does not cover AC-3.3
**Answer:** Fixed. T-17 `hiddenSourcePageReloadsOnlyWhenRevisited` was added, and so was manual step 5.

### X-8 · major · Check lines not runnable as written
**Answer:** Fixed.
- STEP-2 runs `go test … -skip TestSwiftVocabularyMatchesGo`; STEP-5 runs that test.
- STEP-3 notes that it compiles the whole target.
- The STEP-5 `rg` now checks for the list usage itself.
- STEP-8 adds `SourcesListNavigationTests` and `SourcePageModelTests`. `EvidenceGraphModelTests` is not added: it does not read `SourcesModel`.

### X-9 · minor · Smaller choices
**Answer:** Fixed.
- T-1 names `TestEvidenceGraphsResolve` and a `readIDs` test helper.
- STEP-6 spells out the hidden-list test's new seed and assertions, and the `CatalogSource` used at line 968.
- D-8 keeps the `invalidation:` closure.
- D-12 explains why a failure with a kept value cannot change the rows.
- D-12 says `visibleSources` reads the revisions itself.
- STEP-6 names `copy`.
- `TestEvidenceGraphsList`'s generic `Lists` checks move to `TestListsValidate` (T-4b).

## Round 2

### X-1 / M-10 · blocker · Round-1 fixes left the rejected shapes in §6 Go core, the macOS table, the Sketch, Edges, Next adds and STEP-9
**Answer:** Fixed.
- §6 Go core uses `optionalID`.
- The macOS row and the Sketch say `sourceRow()` and `readIDs(for:)`.
- The Sketch carries D-6's real signatures, plus `merged`, `beginRefresh` and the stale methods.
- Edges and Next adds name `sourceRow()` and `readIDs(for:)`.
- STEP-9 rule 4 names `sourceRow()`.
- The sweep adds `insertedSourceRow|deletedSourceId`.

### X-2 · blocker · DoD sweeps cannot print nothing, because the plan and audits sit under `docs/`
**Answer:** Fixed. The sweep is scoped to `docs/*.md` and `docs/ideas/`, and says `docs/deployment-plan/` keeps history and is not swept. It notes that `check-guidance-names.py` does not scan `docs/`.

### X-3 · blocker · `.createdSource` test edits under-specified
**Answer:** Fixed. STEP-6:
- renames and rewrites the test at 636–650 as `applyCreatedSourceInsertsRow`, with its expected cached value;
- gives line 865's new expected array;
- names the constructor for 1169, 1271 and 1284.

STEP-5 deletes `deletedSourceInvalidatesEvidenceGraphsList`, which T-13 replaces.

### M-11 · blocker · D-6 hand-wrote "invalidate and reload" beside `applyHiddenPolicy`
**Answer:** Fixed. D-6 has one exit for "cannot merge", `applyHiddenPolicy(to:)`. It is used for no value with a load in flight, a hidden non-`.revalidate` key, a stale handle, and a failed read. T-7 gains case (d), a hidden key with a first load in flight, which expects no fetch.

### X-4 · major · `sourceRow()` would append on any `patchedSource`
**Answer:** Fixed. D-10 appends only for `.createdSource`; an update or type change only replaces. T-8 asserts both.

### X-5 · major · `updatedSourceShell()`'s `setQueryValue` clears a hidden page's stale flag
**Answer:** Fixed. D-10 has `updatedSourceShell()` write through `patchCached`. A visible page is never stale, so nothing changes there.

### X-6 · major · `beginLoading` clears `isStale`, but the partial path called it
**Answer:** Fixed. D-6 adds `QueryHandle.beginRefresh()`, which sets `isFetching` only. A stale handle takes the `applyHiddenPolicy` exit instead of merging, and the second `beginRefresh()` runs only if the key is not stale.

### X-7 · minor · `readIDs` already exists in the test
**Answer:** Fixed. STEP-1 hoists the existing closure above line 141 and adds no second helper.

### X-8 · minor · Open choices
**Answer:** Fixed.
- The Sketch names `@MainActor` / `@Sendable` on `read` and `merge`.
- The fallback is `applyHiddenPolicy`, so whether it runs off screen follows the kind's policy.
- T-7(b) says the fallback also throws and the handle keeps its rows.
- STEP-5 names the tests at 417, 525, 1726, 1834 and 1932, which switch to `settleWrites()` with their own expectations kept.
- The ghost walk stays guarded by the Stop condition and T-1.

### X-9 · minor · Checks
**Answer:** Fixed. STEP-2 names the protoc tools to install first. The `xcodebuild` steps are unchanged; they run on a macOS machine, as `check-pr.sh` does.

### Below blocker (round 2 maintainability)
- `readIDs(for:)` now derives the Promote case from `evidenceGraphId` (D-8).
- `resetForInvalidation()` is kept as is for its one test caller: removing it is outside this PR's files and scope. Logged, not changed.

## Round 3 — refresh against `fdd37f19`

The plan was refreshed after S13-01 (#655) turned the descriptor's `refresh` slot into `refreshes` (`.revalidate` / `.run`); the implementer stopped on the base check (ship log round 0). Fresh reviewers: maintainability (Opus), executability (Sonnet). Round 3 reviewers re-read the whole plan, not only the refreshed rows, and two of their blockers reshape round 1's answer to M-3 (the `refreshes[key]` map).

### M-1 · blocker · D-6/D-7 add a second task registry and a second "is refreshing" record
`writeRefreshes` already holds every `.run` task; `refreshes[key]` and the hand re-asserted `beginRefresh()` duplicated it.
**Answer:** Fixed. D-7: `trackWrite(for:)` stores the key with the task in `writeRefreshes`, and `readyValue` awaits that key's entries. D-6 (e): `QueryHandle.isFetching` is computed from a load flag and a refresh count (`beginRefresh` / `endRefresh`, ended in a `defer`). The `refreshes` map is gone (Rejected column). T-7 (e) covers overlap.

### M-2 · blocker · The "can this key merge?" gate sits inside `refreshCached`
**Answer:** Fixed. D-6 (a)–(c): `effectivePolicy(for:)` is the one spelling `applyHiddenPolicy` also switches on; `apply` sorts each `.run` key synchronously before its dedupe pass, and a key that cannot merge joins that pass. `refreshCached` keeps only the read-failure exit. T-12 now asserts no run starts.

### M-3 · major · `listSubjectsGraphIDs` beside the unread `listSubjectsCalls`; §4 row wrong
**Answer:** Fixed. Verified `FakeStore.swift:126, 2151`, no reader. §4 row corrected; D-4 replaces it; Codebase fit, Deletes and the DoD sweep name it.

### M-4 · major · S14-27 not in Next adds; its map entry goes false
**Answer:** Fixed. Next adds names S14-27; STEP-9 corrects its scope entry to the Properties workspace alone.

### M-5 · minor · `beginLoading(stale:)` means "keep the value" beside the new `isStale`
**Answer:** Fixed. D-5 renames the parameter `keepingValue:`; the erased protocol is `AnyQueryHandle`, named for what it erases.

### M-6 · minor · The merge re-derives `patchCached`'s guard
**Answer:** Fixed. D-6 (d) merges through `patchCached`.

### M-7 · minor · Comments the change makes false
**Answer:** Fixed. Verified `CatalogQueryRegistry.swift:283, 461`. Deletes lists all four; STEP-4, -5 and -7 rewrite them.

### M-8 · minor · Dead `CatalogQueryStalePolicy`
**Answer:** Fixed as a log line. STEP-9 adds it to `structural-debt.md`. Removing it touches every arm and two tests for no behavior and is not one of the copies this PR touches.

### M-9 · minor · `CatalogMutation.impliedEffects` instead of `readIDs(for:)`
**Answer:** Declined. Both shims exist only until S14-16 gives Promote and DeleteSource effects (map row S14-16), which deletes `readIDs` and the `mutation` parameter of `ids(in:mutation:)` together. A `CatalogWriteEffects` union built now is a concept that PR deletes; `readIDs` is the smaller thing to delete.

### X-1 · blocker · STEP-4 adds T-7 and T-12, which need STEP-5 code
**Answer:** Fixed. T-7 and T-12 move to STEP-5; STEP-4 proves the mechanism through the progress arm (T-7b).

### X-2 · blocker · SourcePageModelTests 417 and 525 have no visible graph list
**Answer:** Fixed. Verified: no `apply(location:)` there. STEP-5 keeps their `waitForQuery(session.query(…))`; only 1726, 1834 and 1932 switch to `settleWrites()`.

### X-3 · blocker · `refreshCached` cannot record itself; Sketch closures
**Answer:** Fixed by M-1: no self-recording. `trackWrite` records the task; `refreshCached` awaits its closures inline, so non-escaping closures compile. Evict leaves the task to finish; `patchCached` then does nothing. `readyValue` loops until no load or refresh is pending (D-7).

### X-4 · major · `assertGraphEffects` has callers in `artifacts_test.go` and no Source id
**Answer:** Fixed. Verified `artifacts_test.go:232, 237` (`f.sourceID`). STEP-2 adds the file and the `sourceID` parameter; T-5 follows.

### X-5 · major · `lists_test.go` locals and `TestWriteEffectsProtoLists`
**Answer:** Fixed. STEP-1 names the replacement rows; STEP-2 uses `{ListVocabulary, ListPropertyUsage}` → `["vocabulary", "property_usage"]`.

### X-6 · major · Unlisted choices in FakeStore and the new tests
**Answer:** Fixed. D-4 places the hold between two `withState` blocks; T-7 (a) uses a flag set by a task, (c) releases the hold twice; T-15 names `WorkspaceLocation(section: .sources, sourceId:, graphId:, sourceSurface: .graph)`; T-9 says FakeStore; T-16 bumps revisions with `setQueryValue`.

### X-7 · major · A Source created while the Sources list's first load is in flight is lost
**Answer:** Fixed. D-10: `sourceRow()` falls back to `applyHiddenPolicy` when `awaitsFirstValue(key)` (shared with D-6). T-8 covers it; Edges names it.

### X-8 · minor · D-6/D-7 gaps
**Answer:** Fixed. The `defer` count ends on every exit, so `isFetching` cannot stick; `patchCached` skips an evicted key; `resetForInvalidation` leaves `isStale` (D-5); no `Sendable` constraint on `Value`/`Part` (D-6 f); Reach gains the stale hidden list reading `.unknown`.

### X-9 · minor · Proto comment and Mark stale row wording
**Answer:** Fixed in §4 and STEP-2, STEP-9.

### X-10 · minor · `.createdSource` sweep hits
**Answer:** Fixed. The DoD lists the `invalidationKind` pair.

### X-11 · nit · `completed.md` path
Logged; STEP-9 names `docs/deployment-plan/spike-14/completed.md`.

## Round 4 — refresh, second round

Fresh reviewers with the log. Raise only new blockers. Round 3's answers held (both reviewers checked M-1, M-2, M-3, M-6, M-7, M-9, X-2, X-4).

### M-10 · blocker · `AnyQueryHandle` cannot say whether a key has a value
`apply` walks keys without `Value`; D-6 (b)/(c) needed "has a value".
**Answer:** Fixed. D-5's protocol gains `hasValue` (and `beginRefresh` / `endRefresh`, X-13); `awaitsFirstValue` reads it.

### M-11 · blocker · D-10 put a second "fall back" decision inside one patch
**Answer:** Fixed. D-10: `patchCached` returns `CatalogPatchResult`; `applyPatches` returns the `.awaitingFirstValue` keys and `apply` adds them to the one dedupe pass. No patch calls a session policy. Edges lists the neighbors (every patch on a first-loading key); T-18 (b). The below-blocker note (make `init(follows:run:)` private) is also taken (D-6 c).

### X-12 · blocker · A key marked stale while its load is in flight stays unsettled
Verified: `ensureQuery` returns early on `inFlight[key] != nil` before it reads the flag, and `applySuccess` does not clear it.
**Answer:** Fixed. D-5: a stale key with a load in flight cancels it and loads again. Reach row 1 says so; T-18 (a).

### X-13 · major · `beginRefresh()` ran in the task, after `apply` returned
**Answer:** Fixed. D-6 (c): `apply` takes the count synchronously and the `trackWrite` closure ends it on the same handle object. It is one begin and one end per run, not the re-assert round 3 removed. `refreshCached` returns when the handle is already gone. T-6 asserts `isFetching` before `apply` returns.

### X-14 · major · `.merged` captures non-`Sendable` generics
**Answer:** Fixed. `Value: Sendable, Part: Sendable` (D-6 f, Sketch); both uses are `Sendable` (`GenealogyStore.swift:94, 1507`).

Below blocker, also fixed: T-6, T-12 and T-13 seed through `setQueryValue`, so the recorder starts empty; T-8's in-flight case makes the list visible; STEP-6 names `.createdSource(source)` at 872; D-6 (c) states that a visible key in `.error` with no value is retried (it falls back for any key without a value).

Round 4's fixes are new text that no reviewer has read. The cap is reached; the ship-pr review of the code is the next check.
