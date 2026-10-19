# Review — S12-16 Add Source reports its effects

**Verdict:** Fits with changes — Ready with changes. All findings fixed or logged.
**Rounds:** 1 (no blocker, no reshape). **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. M-1 changes how the deferred Copy Source count fix will look (one `refreshAll()` line in `SourceCopyModel`); the plan's Out row says so.

## Writer's own calls
- Count after Add Source: `await catalogCounts?.refreshAll()` after `apply` (D-7, from M-1), not a wait on the list reload.
- `FakeStore.insertSource` shared by `createSource` and `copySource` (D-5, from M-2).
- The sidebar count lag after Copy Source stays out of this PR (Out row); raised to the user.
- Tests: extend the existing Go case and Swift create test; one new Swift test proves the reload with a row planted behind the session (D-9). No fake call counter.
- No version bump, as in S12-01 and S12-02.

## Round 1

### M-1 · major · The sidebar count after a Source is created gets a third update path that only works when the list is on screen
**Plan:** D-7 / D-8 waited on `session.readyValue(sourcesListKey)` and then called `syncCatalogCounts()`; a hidden list publishes the old count.
**Codebase:** `CatalogCounts.refreshAll` (doc: for writes with no loaded list to publish from); `PromoteModel` calls `apply` then `await catalogCounts?.refreshAll()`. `SourceCopyModel` and S12-17 could not reuse the wait.
**Better shape:** `apply(effects, .createdSource)` then `await catalogCounts?.refreshAll()`; drop the wait and D-8's caveat.
**Answer:** Fixed — D-7, D-8, §3 In, §4 Counts row (new), §6 code, Codebase fit, Deletes, Risks, STEP-4 check, AC-2.3; the visibility Stop condition is removed; the Out row names the same line for the Copy fix. Verified `FakeStore.workspaceNavCounts` counts `sourcesByProject`, so T-2 and T-3 hold.

### M-2 · minor · `FakeStore.createSource` and `copySource` become near-copies of the same insert
**Plan:** D-5 repeated `copySource`'s append / `bumpSource` / lookup tail.
**Answer:** Fixed — D-5 adds private `FakeStore.insertSource`, used by both; Codebase fit row added; Deletes updated.

### M-3 · nit · How the `readyValue` result is discarded
**Answer:** Moot — M-1 removed the `readyValue` call.

### X-1 · major · Tools for STEP-1 and the Swift checks are missing here
**Answer:** Fixed — header **Runs on:** line and D-10: macOS with `protoc`, `protoc-gen-go`, `protoc-gen-swift` and Xcode; STEP-1 regenerates before STEP-2 because T-1 reads `created.Effects`.

### X-2 · minor · STEP-3 check proves less than it claims
**Answer:** Declined — STEP-3 only has to leave the tree building; T-2 and T-3 in STEP-4 test the fake.

### X-3 · minor · FakeStore "stored row" expression is not spelled out
**Answer:** Fixed — D-5 says the helper runs inside the caller's `withState` block and why `?? source` never fires.

### X-4 · minor · Go test insertion point for T-1 is unspecified
**Answer:** Declined — any place in the case's `after` after `created` is unmarshalled works; not a decision.
