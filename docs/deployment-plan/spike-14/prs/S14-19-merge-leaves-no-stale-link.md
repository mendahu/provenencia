# S14-19 — A merge leaves no stale link in the graph cache

**Status:** Landed in [#681](https://github.com/mendahu/provenencia/pull/681) (2026-11-01)
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the in-memory graph's commit path and two Go tests only; every view draws as before)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-19-merge-leaves-no-stale-link`
**Version bump:** patch — a cache correctness fix that changes no feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** none. No table, query package, FFI method, catalog session or vocabulary changes; `use-catalog-session`, named in the map's Context, does not apply (the fix is inside `graphcache.OnCommit`, which runs under whatever session the writer holds)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit L-16](../app-health-audit.md) (its `a_handle_merged` and verify-mode half); [`docs/conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §2.3 and §9 Canonical merge; [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) (loads read `graphcache`); the S13-03 ship log follow-up (in git with #666, "fails `TestMemoFollowsCommit/a_handle_merged` (`links: count 1, catalog 0`)"); working rules: "Audit items keep their IDs", "Every audited write returns its effects, and a test holds them to it"
**Landed:** [#681](https://github.com/mendahu/provenencia/pull/681), 2026-11-01. Differences from the plan: none, except that the version bump (0.2.7 to 0.2.8) is its own commit.
**Planned:** 2026-11-02 against `main` @ `e2efbd0`
**Review:** [`S14-19-merge-leaves-no-stale-link.review.md`](S14-19-merge-leaves-no-stale-link.review.md) — Fits — Ready, 2 rounds

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The catalog's in-memory graph (`core/database/graphcache`) drops a link to a handle once that handle is merged, the same way a fresh load does. Today, when a commit names a handle that has just been merged, the graph reloads that handle but not the loaded nodes that link to it, so a neighbor (the birth Event of a merged Person) keeps a link the catalog no longer has. `PROVENENCIA_GRAPH_VERIFY=1` catches it in `promotealign` `TestMemoFollowsCommit/a_handle_merged` (audit L-16; the S13-03 ship log reproduced it on `main`). After this PR the verify-mode `promotealign` suite passes.

**Why now.** Bar outcome 19 ("a merge leaves no stale link in the graph cache"), and the bug half of S12-43, which moved here when Spike 12 closed. No product action merges handles yet; the first merge writer would inherit the stale link.

**What stays.** What a fresh load reads (`linkLoadSQL` still drops links to merged neighbors); how a delete or an association commit is applied; every Propose result, which equals a fresh read before and after; every non-verify test.

## 2. User stories

**US-1.** As the engine, when a commit names a handle that is now merged, every loaded node's links match what a fresh load reads, so Promote, kinship and family reads never walk through a merged handle.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Person P and a birth Event E joined by a `participation`; E and P are loaded. P is set `merged_into_id` = Q, and the commit names P and Q | E has no link to P; P is loaded with `Merged` true and keeps its link to E; `g.Verify()` returns nil |
| AC-1.2 | The same, but only E is loaded (P never read) | E has no link to P; P is not loaded by the commit; `g.Verify()` returns nil |
| AC-1.3 | The same, but neither E nor P is loaded | The commit loads neither E nor P; `g.Verify()` returns nil |
| AC-1.4 | `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./core/database/promotealign/...` | Passes, `TestMemoFollowsCommit/a_handle_merged` included (the map's Check) |
| AC-1.5 | `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` with and without `PROVENENCIA_GRAPH_VERIFY=1` | Every package that passed at the planning base still passes |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.OnCommit` reloads the loaded neighbors of every committed handle that is merged, and adds them to the roots whose displays and memos drop (D-1, D-2) | STEP-3, AC-1.1 – AC-1.3 |
| `graphcache.VerifyMode()` replaces the hand-typed verify-mode checks (D-7) | STEP-4 |
| A table-driven `graphcache` test for the merge commit (D-5) | STEP-2, AC-1.1 – AC-1.3 |
| `TestLargeSourceBudgets` logs its timings without failing them when `PROVENENCIA_GRAPH_VERIFY=1` (D-4), so the map's Check holds on any machine | STEP-4, AC-1.4 |
| `OnCommit`'s doc comment says what a merged handle does | STEP-3 |
| The map's S14-19 Context line names `graphcache.OnCommit` instead of "the merge path in `canonicalentities`", which does not exist (§4) | STEP-6 |

| Out | Home |
| --- | --- |
| A product merge writer (re-pointing claims, the Merge action) | not planned (no map row; [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) §9 describes it) |
| Un-merging a handle (clearing `merged_into_id`) | not planned: nothing clears it, and the data model has no un-merge |
| The delete branch of `OnCommit` (a removed node's neighbors go to roots but are not reloaded; writers commit the associations they delete, as `canonlinks` "an event deleted alone" documents) | not planned |
| S12-43's performance work: the shortlist index kept on commit, profile memos, budgets at Scale | [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md) |
| `TestLargeSourceBudgets` flaking over its budgets on CI runners without verify mode | not planned in this PR (a known runner-speed flake in five ship logs; the budgets are [`docs/ideas/promote-ranking.md`](../../../ideas/promote-ranking.md)'s) |
| Running verify mode in CI or in `scripts/check-pr.sh` | not planned (L-16 asks only that the suite be clean) |
| L-16's other items (`TestCallRouter`'s method 100 row; `c.DB()` inside `writes.Run`) | not planned in this map; audit L-16 keeps them |

## 4. Current state

Verified by reading the code at `e2efbd0`. The planning session could not run Go tests (its sandbox refused running the repo's code), so the failure itself is taken from the S13-03 ship log, which reproduced it on `main` at `067be776`: `links: count 1, catalog 0`. STEP-1 reproduces it before any change.

| Area | Where | Today |
| --- | --- | --- |
| Commit path | `core/database/graphcache/commit.go` `(*Graph).OnCommit` | For each committed id, reads the entity row (`entityRows`). A row that exists and is not a bridge is an entity: it is reloaded when loaded or when its kind index is loaded (`entityNeedsReload`). Nothing looks at `row.merged`. The merged handle's neighbors are neither reloaded nor added to `roots`. A missing row (deleted) adds the cached node's link neighbors to `roots` and removes the node |
| Entity row | `core/database/graphcache/load.go` `entityRow.merged`, `(*Graph).entityRows` | Reads `e.merged_into_id IS NOT NULL` into `merged`, for merged and unmerged rows alike |
| Link load | `load.go` `linkLoadSQL`, `(*Graph).loadLinks` | Selects links whose endpoint `a.value_entity_id` is in the ids; the association and the **neighbor** must be unmerged (`assoc.merged_into_id IS NULL`, `n.merged_into_id IS NULL`). Self is not filtered: a merged node loads its own links |
| Holders | `load.go` `install`, `attach`, `detach`, `removeNode`; `Graph.holders` | `holders[association]` lists the loaded nodes that hold a link through it; `install` replaces a node and re-attaches |
| Kind index | `load.go` `syncKind` | A reinstalled node with `Merged` true leaves its loaded kind list |
| Verify | `core/database/graphcache/verify.go` `(*Graph).Verify`, `verifyIDs`, `sameLinks` | Compares each loaded node's `Links` with `loadLinks` of its id. Runs after every `OnCommit` when `PROVENENCIA_GRAPH_VERIFY=1` (`commit.go` `verifyIfSet`) |
| Memos | `core/database/graphcache/memo.go` `dropMemos`, `NodeMemo.Deps` | Drops each slot's results for `roots` and for what the slot's `Deps` names for them |
| Merge in tests | `core/database/graphcache/graphcachetest/graphcachetest.go` `(*Canon).Merge`, `(*Canon).Commit` | `Merge(id, into)` sets `merged_into_id`; `Commit(ids...)` commits `effects.Set{Handles: ids}` |
| Merge in product | `core/database/canonicalentities/canonicalentities.go` | Reads `merged_into_id` only (`MergedIntoID`, counts skip merged). No Go code outside tests writes `merged_into_id`: `rg 'merged_into_id\s*=' --type go --glob '!*_test.go'` hits only `graphcachetest.go` (a test helper package); eight `_test.go` files merge with raw SQL |
| Failing test | `core/database/promotealign/memo_test.go` `TestMemoFollowsCommit` "a handle merged" | Warms Propose, merges `w.personHandle` into another promoted Person with SQL, commits `Handles: [personHandle]`, and compares warm with fresh. In verify mode the commit's `Verify` fails: the birth handle (loaded by Propose) still links to the merged Person |
| Budget test | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, its `over` helper, `budget*` constants | Fails a subtest when a timing is over budget. Audit L-16 lists it as failing in verify mode; five Spike 13 – 14 ship logs record it missing its budgets on CI runners without verify mode. Its timed sections (`since(...)`) contain no commit |
| Graph tests | `core/database/graphcache/graph_test.go` `TestCommitOfAnAssociationLoadsEndpoints`, `TestCommitReloadsASharedEndpointOnce`, helper `loaded` | Cover association commits; no test commits a merged handle |
| Verify-mode skip | `core/database/canonlinks/reader_test.go` "an event deleted alone" | Skips under `PROVENENCIA_GRAPH_VERIFY=1` with a one-line reason |

**Precedent.** The association branch of `OnCommit` (`detachAssociation`: old holders that are loaded go to `reload`, every holder and endpoint goes to `roots`) is the shape to copy: neighbors found, loaded ones reloaded, all of them roots. `TestCommitOfAnAssociationLoadsEndpoints` is the test shape (a `graphcachetest.Canon`, a commit, `loaded` and `n.Links` asserted). The verify-mode skip in `core/database/canonlinks/reader_test.go` is the precedent for D-4's environment check.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The fix lives in `graphcache.OnCommit`. For every committed id whose entity row has `merged` true and is not a bridge, the graph finds that handle's catalog links with the existing `loadLinks`, and each link's `Neighbor` that is loaded goes to `reload`. Every neighbor found, loaded or not, goes to `roots` | `linkLoadSQL`'s `n.merged_into_id IS NULL` is the graph's own rule, so the graph owns its invalidation; any future merge writer only names the handles it changed. `loadLinks` is the indexed query the graph already runs (self is not filtered, so it returns a merged handle's links) | (a) Make the merge writer also commit the merged handle's associations — rejected: a merge changes no association row, and every writer would have to know the cache's filter. (b) Use the cached merged node's `Links` only — rejected: misses AC-1.2, where the neighbor is loaded and the merged handle is not. (c) Scan every loaded node's links — rejected: O(graph) per commit at the map's Scale |
| D-2 | The neighbor lookup runs before the single `g.loadIDs(reload, true)`, so one reload covers the merged handle and its neighbors. It runs once per commit over all merged ids, through `sqlbatch.Each` so a commit naming many merged handles stays under SQLite's variable limit | One reload, one link query batch; a shared neighbor reloads once (`loadIDs` uniques) | A second reload after the first — rejected: two passes for no gain |
| D-3 | The lookup runs for every committed merged handle, not only for one whose merge is new to the cache | The graph cannot tell "newly merged" for a handle it never loaded (AC-1.2). For a handle already merged, `loadLinks` still finds the neighbors, which by then hold no link to it, so the extra work is one query and at most a reload of loaded neighbors | Compare with the cached `Node.Merged` — rejected: wrong when the handle is not loaded |
| D-4 | `TestLargeSourceBudgets`'s `over` helper logs and returns without failing when `graphcache.VerifyMode()` *(new, D-7)* is true. Every other assertion in the test (the fixture's row counts, "memo matches a fresh run", "warm queries") still runs in verify mode | Verify mode checks the cache, not speed: it rebuilds every loaded node after each commit, and the budgets were set without it. The map's Check runs the whole package in verify mode, which a timing test cannot pass on every machine | Skip the whole test in verify mode — rejected: loses its correctness subtests under verify. Leave it — rejected: the map's Check then depends on the machine |
| D-5 | The new test is `graphcache` `TestCommitOfAMergedHandle` *(new)* in `graph_test.go`, one table over AC-1.1 – AC-1.3 (`mergedLoaded`, `eventLoaded` booleans). It builds with `graphcachetest.NewCanon`, `Person`, `Birth`, `Merge`, `Commit`, and calls `g.Verify()` directly, so it holds the rule whether or not the environment sets verify mode | The rule is the graph's; `promotealign`'s test reaches it only through Propose and only in verify mode | Rely on `TestMemoFollowsCommit` alone — rejected: it fails only in verify mode, which CI does not run |
| D-7 | `graphcache` exports `func VerifyMode() bool` *(new)*, in `verify.go`: `os.Getenv("PROVENENCIA_GRAPH_VERIFY") == "1"`. `verifyIfSet` in `commit.go`, the `canonlinks` "an event deleted alone" skip, and D-4's `over` call it; no other code reads the variable | The env name and its `== "1"` rule are hand-typed in two places today (`commit.go` `verifyIfSet`, `core/database/canonlinks/reader_test.go`); D-4 would add a third, in a third package | A third literal copy — rejected: a rename or a new accepted value would miss one (review M-1) |
| D-6 | Product version bump: patch | `docs/versioning.md`: a bugfix | none — the code changes, so docs-only does not apply |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `(*Graph).mergedNeighbors(ids [][]byte) (reload, touched [][]byte, err error)` *(new)*, in `commit.go` | func | `(*Graph).detachAssociation` (returns `reload, touched` for an association); `(*Graph).loadLinks` | Same return shape as `detachAssociation` on purpose; it calls `loadLinks` instead of adding a query. `detachAssociation` itself removes the association node and loads fresh endpoints, which a merge does not need. Searched: `rg 'merged' core/database/graphcache`, `rg 'value_entity_id' core/database/graphcache core/database/canonlinks core/database/canonicalgraph` |
| `graphcache.VerifyMode() bool` *(new)*, in `verify.go` | func | `commit.go` `verifyIfSet`'s `os.Getenv` check; the literal in `core/database/canonlinks/reader_test.go` | It is that check, extracted so the three readers share it (D-7). Searched: `rg PROVENENCIA_GRAPH_VERIFY --type go` |
| `TestCommitOfAMergedHandle` *(new)* | test | `TestCommitOfAnAssociationLoadsEndpoints` | A different commit shape; same helpers (`graphcachetest.Canon`, `loaded`) |

**Builds on.** `OnCommit`'s `reload` / `roots` pattern; `loadLinks` and `sqlbatch.Each`; `graphcachetest.Canon.Merge`.
**Deletes.** The two hand-typed `os.Getenv("PROVENENCIA_GRAPH_VERIFY")` checks (`commit.go` `verifyIfSet`, `core/database/canonlinks/reader_test.go`) become `graphcache.VerifyMode()` calls, and `reader_test.go`'s `"os"` import goes with its only use; `commit.go` drops `"os"` too (nothing else in it uses it). The `canonlinks` verify skip itself stays (it commits a deliberately inconsistent set, unrelated to merges).
**Next adds.** A product merge writer on `writes.Run` (not planned) needs nothing extra: `effects.Resolve` already names the merged handle (`core/database/effects/registry.go` `"canonical_entities"` `Handles: self()`) and the survivor (re-pointed `"identity_claims"` `field("entity_id")`), and the graph drops the stale links itself. [`docs/ideas/promote-ranking.md#name-index-on-commit`](../../../ideas/promote-ranking.md#name-index-on-commit)'s "a merge: remove both, add the survivor" can read the `merged` list `OnCommit` now builds. S14-20's `checkFixed` release of a pair whose record was merged reads the graph this PR keeps true.

### Reach

None — no new path into existing code. The state "a loaded node links to a merged handle" stops being reachable; the state "a committed handle is merged" was already reachable through `graphcachetest.Canon.Merge`.

### Scale

At the map's Scale (37,000 Subjects; 6,000 events on one place): a commit naming a merged handle adds one batched link query over the merged ids and reloads the merged handle's loaded neighbors. The worst named case, a merged place with 6,000 events, reloads at most the 6,000 events that are loaded, in `sqlbatch` chunks: the same cost an association commit on that place's events already pays. A commit with no merged handle runs no new query. No product writer merges today, so no budget changes; no budget is set or leaned on.

### Edges

- **Assumptions.** A bridge has exactly two endpoints (`bridgeEndpoints`), so a link M → N through association A exists in the catalog exactly when N → M does while both are unmerged: `loadLinks(M)` names every node that held a link to M. Holds for every registered bridge (`connectrules.Bridges`); `TestCommitOfAMergedHandle` covers `participation`.
- **Neighbors.** A merged handle whose neighbor is also merged: `loadLinks` drops that neighbor (`n.merged_into_id IS NULL`), and it held no link to M already. A neighbor named in the same commit: reloaded once. A committed handle that is not merged: unchanged path, no new query (AC-1.5 holds the rest of the suite).
- **Removed inputs.** A merged handle's links leave its neighbors; memos and displays that read those neighbors drop because the neighbors join `roots`; `dropMatchIndex` and `dropHandleMemos` already run for any handle commit. The merged node itself stays loaded with its own links and `Merged` true, as a fresh load gives.
- **Real data.** None — no table, seed or configuration drives the rule.
- **Untrusted size.** The merged ids come from the writer's effects set, not from catalog data; `loadLinks` runs in `sqlbatch` chunks (D-2).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `commit.go` `OnCommit`: in the `ids` loop, a found, non-bridge row with `row.merged` adds `row.id` to a local `merged` list (it still goes to `entities` as today). Before `g.loadIDs(reload, true)`, `mergedNeighbors(merged)` *(new)* returns `reload` (link neighbors that are in `g.nodes`) and `touched` (every link neighbor); append them to `reload` and `roots`. The doc comment gains two sentences, side by side: "A deleted handle's neighbors become roots; its writer commits the associations it removed. A merged handle's loaded neighbors are reloaded by the graph, so none keeps a link a fresh load would drop." |
| `core/database/graphcache` | `mergedNeighbors(ids)`: returns nil, nil, nil for no ids; otherwise `sqlbatch.Each(ids, ...)` over `g.loadLinks(chunk)`, collecting each `Link.Neighbor` |
| `core/database/graphcache` | `verify.go`: `// VerifyMode reports whether PROVENENCIA_GRAPH_VERIFY=1, which rebuilds and checks the cache after every commit.` `func VerifyMode() bool` *(new)*, with `"os"` added to `verify.go`'s imports. `commit.go` `verifyIfSet` returns nil when `!VerifyMode()`; drop `commit.go`'s `"os"` import |
| `core/database/canonlinks` | `reader_test.go` "an event deleted alone": `if graphcache.VerifyMode()` in place of the `os.Getenv` check; drop the `"os"` import (its only use) |
| `core/database/promotealign` | `perf_test.go` `TestLargeSourceBudgets` `over`: after `t.Logf`, return when `graphcache.VerifyMode()`, with a comment: "Verify mode rebuilds the cache on every commit; budgets are measured without it." |

Errors: `mergedNeighbors` returns `loadLinks`'s error, and `OnCommit` returns it as it does for `entityRows`.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Reproduce**
Files: none.
Do: on a clean checkout of the base, run the verify-mode suite and the whole Go suite in verify mode, and record every failure.
Check: `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m -v ./core/database/promotealign/...` fails `TestMemoFollowsCommit/a_handle_merged` with a `links: count 1, catalog 0` error. For `TestLargeSourceBudgets`, record each failing subtest and its message, and sort them into budget lines (the `t.Errorf` in `over`) and anything else (`Fatalf`, `Errorf` in "warm queries" or "memo matches a fresh run", a timeout). Anything else is a Stop condition. Every verify-mode command in this plan uses `-timeout 30m`. Then `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./api/... ./core/...`: record the list of failing packages (the AC-1.5 baseline).

**STEP-2 — The merge-commit test**
Files: `core/database/graphcache/graph_test.go`.
Do: add `TestCommitOfAMergedHandle` (D-5), table-driven per `.cursor/rules/go-tests.mdc`: `tests := []struct{ name string; mergedLoaded, eventLoaded bool }` with rows "both loaded" (true, true; AC-1.1), "only the event loaded" (false, true; AC-1.2), "neither loaded" (false, false; AC-1.3). Each row, in `t.Run`: `c := graphcachetest.OpenCatalog(t)`, `canon := graphcachetest.NewCanon(t, c)`, P and Q from `canon.Person()`, `E, joins := canon.Birth(P, nil, nil)` (a birth Event with P as subject), `canon.Commit(joins...)`, `g := c.Graph()`, `g.Drop()`. Then `g.Node(E)` when `eventLoaded` and `g.Node(P)` when `mergedLoaded` (`Node` loads only the id it is given, `graph.go` `(*Graph).Node`). Then `canon.Merge(P, Q)` and `canon.Commit(P, Q)`. Assert: `loaded(g, E) == eventLoaded` and `loaded(g, P) == mergedLoaded`; when E is loaded, `n, ok := g.Cached(E)`, fail if `!ok`, and no entry of `n.Links` has `Neighbor` equal to P (`bytes.Equal`); when P is loaded, `m, ok := g.Cached(P)`, fail if `!ok`, `m.Merged` is true and `m.Links` still hold one link with `Neighbor` E (a merged node keeps its own links, as a fresh load gives); `g.Verify()` returns nil. Do not touch Q's neighbors: Q has none, and D-1 reloads only a merged handle's.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestCommitOfAMergedHandle ./core/database/graphcache/` fails "both loaded" and "only the event loaded" (E still links to P), and passes "neither loaded".

**STEP-3 — Reload a merged handle's neighbors**
Files: `core/database/graphcache/commit.go`.
Do: per §6 Go core: the `merged` list in `OnCommit`, `mergedNeighbors` *(new)*, appends to `reload` and `roots` before `loadIDs`, and the doc comment sentence.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes, `TestCommitOfAMergedHandle` included; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m -run TestMemoFollowsCommit ./core/database/promotealign/` passes.

**STEP-4 — Budgets log in verify mode**
Files: `core/database/graphcache/verify.go`, `core/database/graphcache/commit.go`, `core/database/canonlinks/reader_test.go`, `core/database/promotealign/perf_test.go`.
Do: D-7 first (`VerifyMode`, then `verifyIfSet` and the `canonlinks` skip onto it, per §6 Go core), then D-4 in `over`, importing `github.com/mendahu/provenencia/core/database/graphcache` in `perf_test.go` if it is not imported yet.
Check: `rg PROVENENCIA_GRAPH_VERIFY --type go` hits only `verify.go` (code and comment) and comments elsewhere; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./core/database/promotealign/... ./core/database/canonlinks/...` passes (AC-1.4); `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign/` still fails a subtest when a budget constant is set to `time.Nanosecond` locally (revert after checking; do not commit it).

**STEP-5 — Whole suite**
Files: none.
Do: run the suite both ways and compare with STEP-1's baseline.
Check: `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./api/... ./core/...` fails no package beyond STEP-1's list minus `promotealign`; `scripts/check-pr.sh` prints `check-pr: PASS` (AC-1.5).

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row status; S14-19 scope entry: the whole **Context.** line becomes "**Context.** Bar outcome 19; `core/database/graphcache` `OnCommit` (memos and `Deps`)."; the **Ships** sentence "The merge's effects drop that link, and the verify-mode test passes." becomes "`graphcache.OnCommit` reloads the loaded neighbors of a committed merged handle, so none keeps the link, and the verify-mode test passes."), `docs/conclusion-layer-data-model.md` §9 (after the sentence ending "re-pointed onto the survivor by application rules (§2.3)." add: "The merge write declares no extra effects for the in-memory graph: `effects.Resolve` names the merged handle (`canonical_entities`) and the survivor (re-pointed `identity_claims.entity_id`), and the graph drops the merged handle's links from its loaded neighbors itself."), this plan (status, Landed note), `docs/deployment-plan/spike-14/app-health-audit.md` (L-16: replace "`PROVENENCIA_GRAPH_VERIFY=1` is not suite-clean (`promotealign` `TestMemoFollowsCommit/a_handle_merged`, `TestLargeSourceBudgets`);" with "`PROVENENCIA_GRAPH_VERIFY=1` is not suite-clean (`promotealign` `TestMemoFollowsCommit/a_handle_merged`, `TestLargeSourceBudgets`; **fixed in S14-19**);" and leave the rest of the bullet as it is), `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj`.
Do: per the Definition of done. Version: the next patch after what `VERSION` holds on your base (0.2.7 → 0.2.8 at the planning base), in all three places. `completed.md`: a new `## S14-19 — A merge leaves no stale link in the graph cache` section at the top (newest first), one paragraph starting `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` Map row: Plan column links this file; Status `Landed in #N (YYYY-MM-DD)` like the other landed rows.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-19-merge-leaves-no-stale-link.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/graph_test.go` `TestCommitOfAMergedHandle` *(new)* "both loaded" | P, E loaded; merge P into Q; commit P, Q → E has no link to P; P loaded with `Merged` and its link to E; `Verify` nil | AC-1.1, D-1, D-5 |
| T-2 | Go | same, "only the event loaded" | E loaded, P not; merge; commit → E has no link to P; P not loaded; `Verify` nil | AC-1.2, D-1, D-3 |
| T-3 | Go | same, "neither loaded" | nothing loaded; merge; commit → E and P not loaded; `Verify` nil | AC-1.3, D-2 |
| T-4 | Go | `core/database/promotealign/memo_test.go` `TestMemoFollowsCommit` "a handle merged" (unchanged) under `PROVENENCIA_GRAPH_VERIFY=1` | warm Propose, merge, commit → no verify error; warm equals fresh | AC-1.4 |
| T-5 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` under `PROVENENCIA_GRAPH_VERIFY=1` | timings logged, not failed; other subtests assert as before | AC-1.4, D-4 |
| T-6 | Go | the whole Go suite with and without verify mode (STEP-5) | no new failing package | AC-1.5 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `graphcache` `TestCommitReloadsASharedEndpointOnce` (query count; an unmerged commit must add no query), `promotealign` `TestProposeWarmIssuesNoQuery`, `TestMemoFollowsCommit`, `TestCandidatesMemoFollowsCommits`, `persondetail` `TestPersonFamily` AC-1.9 and AC-1.12 (they merge with `Canon.Merge`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./core/database/promotealign/... passes (the map's Check)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it) — nothing is renamed or deleted
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump: patch (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, the S14-19 Context line and Ships sentence (STEP-6), and `rg "merge's effects drop" docs` finds nothing outside this plan and its review log; completed.md: one short paragraph of what shipped
- [ ] app-health-audit.md L-16: the two verify-mode items marked fixed in S14-19
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/conclusion-layer-data-model.md §9 gains the merge-effects sentence (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: no product action merges handles. The reviewer runs:

1. `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 -timeout 30m ./core/database/promotealign/...` → `ok`.
2. `CGO_ENABLED=1 go test -tags fts5 -run TestCommitOfAMergedHandle -v ./core/database/graphcache/` → three subtests pass.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The extra neighbor lookup adds a query to every handle commit | D-1 runs it only when a committed row is merged; `TestCommitReloadsASharedEndpointOnce` holds the unmerged query count |
| Verify mode finds a second inconsistency once the link is fixed (values or members of the merged handle) | Stop condition below; the plan does not widen to it |

**Stop and report if:**

- STEP-1's verify-mode `TestMemoFollowsCommit/a_handle_merged` passes, or fails with an error other than a `links:` mismatch on a neighbor of the merged handle.
- After STEP-3, `TestMemoFollowsCommit/a_handle_merged` still fails in verify mode, with any error (values, members, identity, a memo check).
- `TestLargeSourceBudgets` fails in verify mode on anything other than a budget line written by `over` (a `Fatal` in "fixture", "memo matches a fresh run" or "warm queries", or the 30-minute test timeout).
- `TestCommitReloadsASharedEndpointOnce` needs its query ceiling raised.
- A package that passed in STEP-1 fails in STEP-5, in either mode.
- `rg 'merged_into_id\s*=' --type go --glob '!*_test.go'` finds a hit outside `core/database/graphcache/graphcachetest/graphcachetest.go` (a product merge path landed after the base).

## 12. Open questions

None.

## 13. PR

**Title:** S14-19 — A merge leaves no stale link in the graph cache
**Body:** objective in one sentence; what shipped against US-1; AC evidence (the verify-mode `promotealign` run, `TestCommitOfAMergedHandle`); deviations from the plan; the manual check commands; `Plan: docs/deployment-plan/spike-14/prs/S14-19-merge-leaves-no-stale-link.md`.
