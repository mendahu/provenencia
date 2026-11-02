# Ship log — S14-19 A merge leaves no stale link in the graph cache

**Plan:** [`S14-19-merge-leaves-no-stale-link.md`](S14-19-merge-leaves-no-stale-link.md) · **Branch:** `s14-19-merge-leaves-no-stale-link` · **PR:** #681
**State:** ready

## Round 0 — implement
Head `4aad6f05`. Gate PASS (orchestrator run against `51dccd48`). PR #681 (draft). Report deviations: version bump in its own commit; map row reads `Landed #681` to match other rows; base moved since `e2efbd0` only in `graphcachetest.go` and `promotealign_test.go` (no Current state row changed); `TestCommitOfAMergedHandle` committed with the fix (confirmed red first).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes; the pbxproj change is the version bump). SEC+PRF dispatched because the diff is in `core/database/**` graph-cache code and `mergedNeighbors` loops over links loaded from the catalog.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | minor | `mergedNeighbors` runs the whole `loadLinks`, including a `disambiguation` query whose result it discards | Fix now → F-1 | verified: `loadLinks` (`load.go:200`) ends in `finishLinks`, which calls `g.disambiguation` (`load.go:326`); `mergedNeighbors` reads only `Link.Neighbor`. Minor in code this PR adds; plan §6 Scale says one link query |
| MNT-1 | MNT | nit → minor | `mergedNeighbors` doc comment ("linked to each merged handle before it was merged") is wrong for a handle merged in an earlier commit (D-3) | Fix now → F-1 | stale comment in a function this PR adds; F-1 rewrites that function anyway |

Nits (logged): `MNT-2 · core/database/graphcache/verify.go:11 · wrap the VerifyMode doc comment at ~80 columns` · `CON-1 · docs/deployment-plan/spike-14/completed.md · double space after "#681)."`

SEC: no findings. CON: no findings. MNT: no findings above nit.

### Fix plan 1
**F-1 — The merged-handle neighbor scan runs the link query only** (from PRF-1, MNT-1)
Files: `core/database/graphcache/load.go`, `core/database/graphcache/commit.go`, `core/database/graphcache/graph_test.go`. Skill: none.
Failing test first: in `graph_test.go`, a query-count test for a merge commit where neither the merged Person nor its Event is loaded (the "neither loaded" setup of `TestCommitOfAMergedHandle`), using `g.SetQueryCounterForTest` as `TestCommitReloadsASharedEndpointOnce` does. Set the counter after setup, before `canon.Commit(p, q)`. Pin the count the fixed code makes (no reload happens, so the neighbor scan should be one link query per chunk and no disambiguation query); it must fail at `4aad6f05` by exactly the one disambiguation query. Say in a comment what the count is made of, like the existing test. Alternatively add a `/neither loaded` count check inside `TestCommitOfAMergedHandle` if that reads better; one test, not both.
Do: split the scan half of `loadLinks` (`load.go:200`) into `linkRaws(ids [][]byte) ([]linkRaw, error)` — the `linkLoadSQL` query, the `connectrules.LookupBridge` / `bridgeEndpoints` filter and the clones — and make `loadLinks` call `linkRaws` then `finishLinks` (collect `assocIDs` from the raws). `mergedNeighbors` calls `linkRaws` inside its `sqlbatch.Each` and reads `r.neighbor`. No behavior change to `loadLinks`. Rewrite the `mergedNeighbors` doc comment so it holds for a handle merged in this commit or an earlier one, e.g. "mergedNeighbors finds the unmerged nodes each merged handle's own links name. reload is the ones already loaded, whose links to the handle a fresh load would drop; touched is all of them, which a stored row may still name." Do not touch the double loading of the merged handle's own links in the "both loaded" case (PRF-1's third bullet): that is the existing reload path.
Check: the new count test fails before the change and passes after; `TestCommitOfAMergedHandle` and `TestCommitReloadsASharedEndpointOnce` pass; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/promotealign/... ./core/database/canonlinks/...` ok; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `4a68fdf4`. Gate PASS (orchestrator run). F-1 done: `loadLinks` split into `linkRaws` + `finishLinks`; `mergedNeighbors` reads `linkRaws`; new `TestCommitOfAMergedHandleNeitherLoadedQueries` pins 2 queries (measured 3 at `4aad6f05`, the extra disambiguation query). Implementer deviation: the count test is its own test, as the fix plan allowed.
- PRF-1: **Resolved** (orchestrator read `4aad6f05..4a68fdf4`; the scan runs only `linkLoadSQL`, `loadLinks` behavior unchanged).
- MNT-1: **Resolved** (comment rewritten as given).

## Round 2 — review
Not run: round 1 had no Fix-now blocker or major.

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: — · CON miss: none
