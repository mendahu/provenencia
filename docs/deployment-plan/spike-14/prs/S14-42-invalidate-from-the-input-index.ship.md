# Ship log — S14-42 Invalidate from the input index

**Plan:** [`S14-42-invalidate-from-the-input-index.md`](S14-42-invalidate-from-the-input-index.md) · **Branch:** `s14-42-invalidate-from-the-input-index` (stacked on `s14-40-build-the-input-index`, #746) · **PR:** #748
**State:** review 2

## Round 0 — implement
Base `f55bf6c1` (S14-40's head). Head `573b5669`. The orchestrator re-ran the gate on this head (see below).

Budgets (warm, fastest of 3; before = STEP-1 on S14-40's code):

| Subtest | Before | After |
| --- | --- | --- |
| Unpromote at the city | 41.1 ms | 44.1 ms |
| Delete a location at the city | 1,272.6 ms (over budget) | 41.9 ms |
| Locate at the city | 81.6 ms | 75.0 ms |
| Second part_of | 103.9 ms | 44.5 ms |
| Compose every summary | 195.2 ms | 203.6 ms |

Deviations:
1. **`HandlesObservingSubject`'s test callers.** It had five more callers, in `effects/conclusion_handles_test.go`, that §4 did not list. The query now lives in that test file as `observerHandles`.
2. **Fixture `row` counter.** The `autoreconciler` fixture gained a counter, because the new scenarios need Subject positions for `CreateCitedBridge`.
3. **One commit for STEPs 3–7.** The tree does not compile between those steps.
4. **D-11 citation values.** The moved tests cite "Yorke", "Jim Robins", 1818 and "Upper Canada".
5. **Manual check not run** in the app.

The orchestrator's gate on `573b5669`: PASS. The implementer's uncached `go-test` run covered the same code, so the cached `go-test` here is acceptable.

## Round 1 — review
Base `f55bf6c1`, head `573b5669`.

Reviewers:
- **CON, MNT:** always run.
- **SEC+PRF:** migration `000062`, `core/database/**`, search, the graph cache.

Skipped: UX. There is no `macos/App/**`, no strings and no error code; the only `macos` change is `MARKETING_VERSION`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | `ReprojectChanged` loads written handles one at a time on a cold graph | Fix now → F-1 | Confirmed: the `range written` loop calls `g.Node(id)`. `graph.go` `node` calls `loadIDs([][]byte{id})` per id. `g.Nodes` (the batched loader) is unused here |
| MNT-1 | MNT | blocker | `scanDependents` is a second id-scan loop beside `searchindex.listIDs` | Fix now → F-2 | Same-pass cleanup. `rg -n 'func (listIDs\|scanDependents)' core/database/searchindex` |
| MNT-2 | MNT | major | `holds` / `holdsID` re-implement `database.ContainsBlobID` | Fix now → F-3 | Same-pass cleanup; this PR adds both copies |
| CON-1 | CON | minor | The seeded sequences never file, unpromote or delete a link | Fix now → F-4 | It guards branch (c), the S14-38 fix, which this PR adds. The reviewer's probe (720 steps, all equal to a rebuild) is the shape to port |
| MNT-3 | MNT | minor | Sorting is done twice; `sortedIDs` is new | Fix now → F-5 | Code this PR adds |
| CON-3 | CON | nit | `summary-inputs.md` still describes the walk as current | Fix now → F-6 | A stale doc in a file this PR changes |
| CON-4 | CON | nit | Two `filedlinks_test.go` tests are named after the deleted walk | Fix now → F-6 (rides along) | A file this PR changes |

Nits (logged):
- `MNT-4 · perf_test.go:287-332 · fixture f.unpromote / f.deleteSubject helpers`
- `MNT-5 · searchindex/inputs_test.go:251 · hasKey vs holds naming`
- `MNT-6 · graphcachetest.go:52-58 · Canon field alignment`
- `CON-2 · perf_test.go:307,330 · double log of the figure`
- `CON-5 · TestIndexReads checks the unknown id on oldEnds only`
- `CON-6 · T-2 checks the title, not full rebuild equality`

### Fix plan 1

**F-1 — Load the written handles in one batch** (from PRF-1)
Files: `core/database/searchindex/inputs.go`, `inputs_test.go`.
Failing test first: `TestReprojectChangedLoadsWrittenInBatches`. Give `ReprojectChanged` 1,200 written ids, a few real and the rest unknown, on a graph whose query counter (`Graph.SetQueryCounterForTest`) is set. The count must be under 40, where today's per-id loads cost about 4 × 1,200.
Do:
- Call `g.Nodes(written)` once, before the gone loop.
- Pass `g.AssociationInputs` only the moved ids that are not gone, since a gone id has no new ends.
Check: the test passes; T-1, T-2 and `TestReprojectHandlesBuildsOneGraph` pass.

**F-2 — One id scan in `searchindex`** (from MNT-1)
Files: `core/database/searchindex/searchindex.go`, `inputs.go`.
Do:
- `listIDs(q Querier, query string, args ...any)`.
- `dependentsOf` calls it; `scanDependents` is deleted.
- `dependentsOfHandles` and `oldEnds` drop their `append([]byte(nil), id...)` copies, because `Scan` into `*[]byte` already copies.
Check: `rg -n 'func scanDependents' core` prints nothing. `rg -nE 'rows\.Scan\(&id\)' core/database/searchindex` prints only `listIDs` and the `sqlbatch.Query` callbacks.

**F-3 — Reuse `database.ContainsBlobID`** (from MNT-2)
Files: `core/database/searchindex/inputs_test.go`, `core/database/autoreconciler/autoreconciler_test.go`.
Do: delete `holds` (in `searchindex`) and `holdsID`, and call `database.ContainsBlobID` at their call sites.
Check: `rg -nE 'func (holds|holdsID)\(ids' core` prints nothing.

**F-4 — Seeded sequences file and remove links** (from CON-1)
Files: `core/database/autoreconciler/autoreconciler_test.go`.
Do: add ops to `TestRebuildEqualsUpkeep_SeededSequences`'s switch, or add `TestRebuildEqualsUpkeep_SeededLinks` with the same seeds and `assertUpkeepEqualsRebuild` after each write. The ops:
- File a participation, a location, a part_of and a succeeded_by between random existing Subjects, with the `Records` helpers and `f.row` positions.
- Unpromote a random promoted Subject (`unpromote_subject`).
- Delete a random link Subject (`delete_subject`).
- Use the existing `tolerate` for refused ops.
The reviewer's probe at `/private/tmp/claude-501/-Users-jakerobins--claude-spares-spare-H816z9/02abe4b5-d454-4d02-a97d-57d62bbf3798/scratchpad/probe/probe_test.go` is the reference shape (12 seeds × 60 writes).
Check: the test passes; its runtime stays under the existing seeded test's runtime × 2.

**F-5 — Sort once, at the boundary** (from MNT-3)
Files: `core/database/searchindex/inputs.go`.
Do:
- The three reads return unique ids, unsorted.
- `ReprojectChanged` sorts its return value once, with `slices.SortFunc(out, bytes.Compare)`.
- Delete `sortedIDs` unless another caller remains.
- A test that compares order sorts its own copy.
Check: `rg -n 'func sortedIDs' core` prints nothing, or prints one caller.

**F-6 — Docs and names left from the walk** (from CON-3, CON-4)
Files: `docs/deployment-plan/spike-14/summary-inputs.md` (lines 13–22), `core/database/conclusionsummaries/filedlinks_test.go`.
Do:
- In `summary-inputs.md`, retitle the column "How a write found what to rebuild before S14-42", and put the list after it in the past tense.
- Rename `TestFiledLocationWalksOnlyFromItsEvent` → `TestFiledLocationReportsOnlyItsEvent`.
- Rename `TestFiledPartOfWalksFromThePart` → `TestFiledPartOfReportsThePart`.
Check: `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `c6849e1d`. The orchestrator ran the gate: PASS. F-1 through F-6 are done.

Deviations:
- **F-1.** The gone check uses the existing `Graph.Cached` after `g.Nodes(written)`, because a missing id is never cached and `Node` would reload it. The test went from 2,404 queries to under 40.
- **F-4.** Runs 4 seeds × 60 writes, not 12, to stay under 2× the existing seeded test (1.68 s against 1.95 s).
- **Commits are split by file**, not by fix item.

Minors and nits, checked by the orchestrator:

| ID | Result | Evidence |
| --- | --- | --- |
| CON-1 | Resolved | `TestRebuildEqualsUpkeep_SeededLinks` (`autoreconciler_test.go:2096`) |
| MNT-3 | Resolved | `rg 'func sortedIDs'` prints nothing; `ReprojectChanged` sorts its return value once |
| CON-3, CON-4 | Resolved | `summary-inputs.md` diff; the renamed tests |

## Round 2 — review
Scope `573b5669...c6849e1d`. Reviewers: MNT (MNT-1 blocker, MNT-2 major) and SEC+PRF (PRF-1 major).
