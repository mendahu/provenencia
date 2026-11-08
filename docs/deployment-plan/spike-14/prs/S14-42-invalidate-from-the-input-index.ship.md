# Ship log — S14-42 Invalidate from the input index

**Plan:** [`S14-42-invalidate-from-the-input-index.md`](S14-42-invalidate-from-the-input-index.md) · **Branch:** `s14-42-invalidate-from-the-input-index` (stacked on `s14-40-build-the-input-index`, #746) · **PR:** #748
**State:** review 1

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
