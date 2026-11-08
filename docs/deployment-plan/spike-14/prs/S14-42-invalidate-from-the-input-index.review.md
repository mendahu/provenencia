# Review — S14-42 Invalidate from the input index

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None.

## Writer's own calls
These are calls no source dictated:
- **D-1:** the lookup lives in `searchindex.ReprojectChanged`, called from `autoreconciler.RecomputeReported`. The design said only that `writes.Run` triggers it.
- **D-4:** a `dependent_is_association` column, added to the unreleased `000062.sql`. It is how a deleted association's end rows are told apart from a summary row's rows.
- **D-5:** the ends of a changed link are reported to the app but not reprojected.
- **D-8:** `HandlesObservingSubject` is deleted. The evidence is the resource bucket on `observations.value_subject_id`.
- **D-1 (round 2):** `RecomputeTx` is deleted, and its three test callers move to `RecomputeReported`.
- **D-13 (round 2):** one Result-returning helper, on `evidencelayertest.Records`.
- **D-12:** both new budgets are set at 500 ms, the same as filing one location at the city.
- **Version:** patch, `0.4.1`.
- **Scope:** Subject delete stays in this PR. The step count is 9, so the map's split-at-about-10 trigger did not fire.

## Round 1

### X-1 · blocker · The AC-3.1 / T-11 `rg` command cannot fail
`\|` inside single quotes is a literal pipe to ripgrep, so the check passes on the unchanged tree.
**Answer:** Fixed. The command uses plain `|` and moved out of the table, which forced `\|`, into a fenced block in STEP-7's Check. AC-3.1 and T-11 now point at it.

### X-2 · blocker · STEP-6 and STEP-4's stop condition name checks that need later steps
**Answer:** Fixed. Steps are reordered:
- STEP-4 adds T-3 – T-7 along with the recompute change, and its Check runs `TestLargeSourceBudgets`.
- STEP-6's T-6 already exists by then.
- The old STEP-8 merged into STEP-4 and STEP-5. There are now 9 steps.

### X-3 · major · The scenario tests need `Result.Effects.SummaryDeps`, and the plan does not say how
**Answer:** Fixed. New D-13: a fixture helper `run` over `writes.Run`, the action strings from the FFI handlers, and the note that `Records.Locate` returns the location Subject.

### X-4 · major · Moving the `Dependents` tests leaves the real choices to me
**Answer:** Fixed. D-11 now gives every moved test its write and its expected set, and says `TestJoinReprojectsDependentsOnlyWhenTheRootMoves` stays unchanged. That test does not call `Dependents`, and it holds the "no move, no reprojection" rule under D-2. The oracle is now named in full: S14-40's `TestInputsMatchIsolatedComposition`.

### X-5 · major · T-2's setup has no recipe
**Answer:** Fixed. T-2 now runs `promote.Unpromote(James)` through `writes.Run`, as S14-40's `TestReprojectWritesInputs` already does, instead of raw SQL deletes.

### X-6 · minor · Unplanned choices in the index and kit code
**Answer:** Fixed.
- **Flag:** D-4 carries it in an `inputRow` struct, which replaces the parallel slices. `inputRowsPerInsert` goes to 160, and STEP-2 says how both test printers show the flag.
- **Graph:** one graph per write (M-1).
- **Kit:** D-7 says an uncommitted `pending` is never read.
- **Working rule:** the map's rule text changes in STEP-9.
- **effects README:** the placement is named.
- **structural-debt:** only row 19 changes.

### X-7 · minor · Size
**Answer:** Fixed by the reorder to 9 steps. STEP-3's Do is numbered into three parts.

### M-1 · major · One graph per write; "gone" from that graph
**Answer:** Fixed.
- **D-2:** `gone` is a nil or `Merged` node on `g`.
- **D-5:** `ReprojectHandles` splits into the entry point and `reprojectOn(q, g, ids)`, and `ReprojectChanged` reprojects on its own `g`.
- **STEP-3:** extends `TestReprojectHandlesBuildsOneGraph`.

### M-2 · minor · List the third index read
**Answer:** Fixed. D-10 names `oldEnds`, and the Codebase fit table has a row for it.

### M-3 · minor · Moved tests should check the write's `SummaryDeps`
**Answer:** Fixed. The moved tests in D-11 now assert on `writes.Run`'s `Result.Effects.SummaryDeps`. The three index reads are unexported, and T-1 reaches them through `export_test.go`.

### M-4 · minor · `Rebuild` gets nothing from change detection
**Answer:** Fixed. D-1: `Rebuild` runs `recomputeBatch` per batch and calls `ReprojectHandles(all)` directly.

### M-5 · minor · Kit: hold the observed ids, not each recompute's result
**Answer:** Declined. D-7 keeps `observe`'s documented contract: it rewrites the entity's kept rows at each Observation (`core/database/graphcache/graphcachetest/graphcachetest.go:155`). Moving that to `Commit` would change when the kit's users see values, across graphcache, canonlinks, persondetail and kinship tests. That change is outside this PR, and nobody has checked what it would break. D-7's Rejected cell records it.

### M-6 · minor · The design doc should name where the logic lives
**Answer:** Fixed. STEP-9 rewrites the opening and closing of `summary-inputs.md` § Invalidation.

## Round 2

The maintainability reviewer verdict was "Fits with changes", with no new blockers. The executability reviewer verdict was "Ready with changes", with no blockers. Every round-1 answer was checked against the code and holds; the M-5 decline also holds.

### M (round 2) · notes below blocker
The reviewer raised these as notes below blocker:
- `RecomputeTx` is left with no production caller.
- The §6 Go core table is stale.
- `summary-inputs.md:91`'s first sentence would still be false.
- The recompute batch loop is copied in two places (nit).

**Answer:** Fixed.
- D-1 deletes `RecomputeTx` and adds `recompute`, which `Rebuild` and `RecomputeReported` share. It is in Deletes, in T-11's `rg`, and in STEP-4.
- The Go core table now matches D-4, D-5 and D-10.
- STEP-9 replaces line 91 whole.

### X-1 (round 2) · major · `filedLinkDeps` has no recipe for the filing write's Result
**Answer:** Fixed. D-13 adds `Records.Reported`, set by each `Records` write helper. D-11 deletes `filedLinkDeps`, and each test reads `f.Reported` right after it files the link.

### X-2 (round 2) · major · No helper to cite through `writes.Run` in `conclusionsummaries` tests
**Answer:** Fixed. D-13 has `Records.Cite` make its write through `writes.Run` and set `Reported`, and `Records.Run` covers any other write. Both fixtures embed `Records`.

### X-3 (round 2) · major · `fromJames` positives dropped without saying so
**Answer:** Fixed. D-11 adds a name cited on James, and the write's set must hold the birth, the death, the marriage and the census.

### X-4 (round 2) · major · Key-lookup batching shape not decided
**Answer:** Fixed. D-10 gives each read's statement: row values `IN (VALUES …)` at 250 keys per statement, `input_id IN` through `sqlbatch.Query`, and `oldEnds` by `dependent_id`.

### X-5 (round 2) · minor · Go core table contradicts D-10
**Answer:** Fixed, together with M above. D-10 names the `export_test.go` exports: `DependentsOfForTest`, `DependentsOfHandlesForTest` and `OldEndsForTest`.

### X-6 (round 2) · minor · Test-side choices in STEP-2
**Answer:** Fixed.
- **`searchindex` test:** `inputRow` gains `association bool`, and `hasRow` ignores it.
- **`autoreconciler` test:** `inputRows` appends `|0` or `|1`.

### X-7 (round 2) · minor · T-8 needs the person's inputs written
**Answer:** Fixed. T-8 calls `Commit(P, E, E2)` before it lists persons. That write reprojects P, because a row is its own input, so P's inputs hold E.
