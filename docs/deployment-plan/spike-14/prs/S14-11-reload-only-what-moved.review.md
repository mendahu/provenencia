# Review — S14-11 Reload only what moved, only when it is on screen

**Verdict:** Fits with changes — Not ready → Fixed (round 2 pending)
**Rounds:** 1 so far. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None. Every finding was answerable from the code; no product call came up.

## Writer's own calls
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
