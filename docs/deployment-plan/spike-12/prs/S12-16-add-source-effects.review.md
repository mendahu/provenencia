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

## Round 2 — refresh after S12-17 (`main` @ `6191fc8`)

S12-17 ([#488](https://github.com/mendahu/provenencia/pull/488)) landed after round 1. It adds `SourcesModel.copying` / `openCopy`, `.sourceCopyDialog` in `SourcesListView`, and `openCopySetsCopyingForThatSource`; none of the code this PR edits moved. Both reviewers re-ran on the refreshed plan (round-2 rules: new blockers only). Verdict: Fits with changes — Ready with changes. No blocker.

### M-4 · major · The recount after a create is a new pattern the skill does not teach
**Plan:** STEP-5 edited only `add-workspace-place` line 172; the skill (line 41, line 156) still teaches publishing from the list. A Source delete has the Copy count lag too.
**Answer:** Fixed — STEP-5 adds the apply-then-`refreshAll()` rule after `add-workspace-place` line 156; the Out row names the delete path beside both Copy builders.

### M-5 · minor · Two doc comments become false
**Answer:** Fixed — STEP-4 rewrites `SourcesModel.syncCatalogCounts()`, `CatalogCounts.publishSources` and `CatalogCounts.refreshAll` comments; Files and Deletes updated.

### M-6 · minor · Header promised a refresh section the log lacked
**Answer:** Fixed — this section; the plan's **Review:** line points at round 2.

### X-5 · minor · Docs cited as "rule 3"; the text is cache contract item 2
**Answer:** Fixed — header Context, §4 Docs row, STEP-5, DoD and the Out row now say cache contract item 2 (line 87).

### X-6 · major · T-2 and T-3 assert before the list reload lands
**Answer:** Fixed — T-2 and T-3 `await waitForQuery` on the Sources list handle (`QueryHandleWaiting.swift`) after `create()`. Round 1's M-1 removed the in-model wait, which opened this gap.

### X-7 · minor · Stop condition 2 would fire on a timing failure
**Answer:** Fixed — the condition now reads "after waiting for the list handle".

### X-8 · nit · D-5 does not say how `createSource` mints the row
**Answer:** Logged — it compiles as written (`updated` defaults to `CatalogActivity()`).
