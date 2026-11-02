# Review — S14-19 A merge leaves no stale link in the graph cache

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None.

## Writer's own calls
- D-1: the fix lives in `graphcache.OnCommit` (the graph reloads a merged handle's loaded neighbors), not in the merge writer's effects. No product merge writer exists; the map's "merge path in `canonicalentities`" does not exist, and the map's Context and Ships lines are corrected in STEP-6.
- D-4: `TestLargeSourceBudgets` logs its timings without failing them in verify mode, so the map's Check (the whole `promotealign` package in verify mode) holds on any machine. Its correctness subtests still run.
- D-5: a new table-driven `graphcache` test that calls `Verify` directly, because CI does not run verify mode.
- D-6: patch bump.
- The planning session could not run Go tests (the sandbox refused running repo code). The failure is taken from the S13-03 ship log, which reproduced it on `main`; STEP-1 reproduces it first, with stop conditions.

## Round 1

### X-1 · blocker · The sixth stop condition fires at STEP-1, and Current state row "Merge in product" is false
`rg 'merged_into_id\s*=' --type go` also matches raw SQL in eight `_test.go` files.
**Answer:** Fixed — §4 "Merge in product" and §11 use `rg 'merged_into_id\s*=' --type go --glob '!*_test.go'`, which hits only `graphcachetest.go`.

### X-2 · major · `perf_test.go` needs a new `os` import the plan does not mention
**Answer:** Fixed, then superseded by M-1 — `over` calls `graphcache.VerifyMode()`; STEP-4 says to import `graphcache` in `perf_test.go`, and §6 says `verify.go` gains `"os"`.

### X-3 · major · STEP-2's setup leaves choices open
**Answer:** Fixed — STEP-2 names the table struct, the `g.Node` calls per row (`Node` loads only the id it is given, `graph.go` `(*Graph).Node`), and every assertion, including that a loaded merged P keeps its link to E (added to AC-1.1 and T-1). It also says not to touch Q's neighbors.

### X-4 · major · STEP-1's `TestLargeSourceBudgets` baseline cannot tell a budget miss from a Fatal
**Answer:** Fixed — every verify-mode command uses `-timeout 30m`; STEP-1 runs with `-v` and sorts failures into budget lines (`over`'s `t.Errorf`) and anything else, which is a Stop condition.

### X-5 · minor · The audit L-16 edit has no "Fixed" paragraph to copy
**Answer:** Fixed — STEP-6 gives the exact replacement text for L-16's verify-mode clause and leaves the rest of the bullet.

### X-6 · minor · Several Step 6 and DoD edits depend on files the plan only names
**Answer:** Fixed — STEP-6 gives the version target (next patch, 0.2.7 → 0.2.8 at the base, in `VERSION`, `core/version.go` and every `MARKETING_VERSION` line), the `completed.md` heading and position, and the map row's Plan and Status values.

### M-1 · major · Third hand-typed copy of the `PROVENENCIA_GRAPH_VERIFY` check
**Answer:** Fixed — D-7: `graphcache.VerifyMode()` *(new)* in `verify.go`; `verifyIfSet`, the `canonlinks` skip and `over` call it. Codebase fit, Deletes, §6 Go core, Scope and STEP-4 updated.

### M-2 · major · The map's Ships line still says "the merge's effects drop that link", which D-1 makes false
**Answer:** Fixed — STEP-6 rewrites the Ships sentence too, and the DoD `rg`s the old wording.

### M-3 · minor · "Next adds" misstates how a merge writer reaches `OnCommit`, and the contract is written down only in `commit.go`
Checked: `core/database/effects/registry.go` `"canonical_entities"` resolves `Handles: self()` and `"identity_claims"` resolves `Handles: union(field("entity_id"), …)`.
**Answer:** Fixed — Next adds now says a writer on `writes.Run` needs nothing extra, and STEP-6 adds the sentence to `docs/conclusion-layer-data-model.md` §9; the DoD's domain-doc line names it.

### M-4 · nit · The `OnCommit` doc comment should put the delete contract next to the new merge contract
**Answer:** Fixed — STEP-3's doc comment states both rules side by side.


## Round 2

Executability (sonnet): every round-1 answer confirmed against the code; no blocker. Ready with changes.

### X-7 · minor · The DoD `rg "merge's effects drop" docs` cannot pass as written
**Answer:** Fixed — the DoD excludes this plan and its review log.

### X-8 · minor · STEP-2 reads `Cached` as a single return value
**Answer:** Fixed — STEP-2 uses `n, ok := g.Cached(E)` and `m, ok := g.Cached(P)`, failing on `!ok` where the row says loaded.

### X-9 · minor · STEP-6's Context-line edit is ambiguous
**Answer:** Fixed — STEP-6 gives the whole replacement Context line.
