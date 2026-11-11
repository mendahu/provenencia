# Ship log — S16-10 One cache registry

**Plan:** [`S16-10-one-cache-registry.md`](S16-10-one-cache-registry.md) · **Branch:** `s16-10-one-cache-registry` · **PR:** #789
**State:** ready

## Round 0 — implement
First pass stopped: the implementer's `git commit` was denied by the permission system; the user answered "try again" and the implementer committed and pushed.
Head `a351801f`. Gate PASS (orchestrator, against `01fa954b`).
Report deviations: (1) `graphcache.Vocabulary` is an existing struct, so it implements `Source` instead of a `var Vocabulary Source` (D-1 naming); (2) `TestListPersonsQueryCountIsConstant` gained a `nameformats.Warm` setup line; (3) `TestLoadFor` and the namealiases nil-graph case now assert `ErrClosed` per D-10; (4) `Reader.For` stores nothing for a non-node id; (5) `Reader.Relationships` reads siblings from `GetMany`'s map; (6) Whole Checks take `(g, stored)` per D-8; (7) one commit for registry + owners (compile together); (8) §9 docs lines pending the PR number.
Measured: Go caches ≈118 MB at 16,000 persons against the 40 MB budget (nodes 58.2, evidence 44.8, match index 9.8, kinship 4.4 MB); the plan says report, not fix.

## Round 1 — review
Reviewers: CON, MNT, SEC+PRF (`core/catalogsession`, `api/ffi/**`, `core/database/namealiases`, graph and cache code). Skipped: UX — the only `api/ffi` change drops a `c.DB()` argument in `nameformat.go`; no `macos/App/**`, strings, `core/apperr` or new error codes.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major (filed blocker) | Doc test sees only owners `catalogsession` imports by hand | Fix now → F-1 | verified: `cache_doc_test.go` blank-imports four owners; nothing fails when a new `graphcache.Register*` package is not imported. Re-rated major: nothing is wrong at merge, but AC-1.9/D-19's guarantee does not hold. Fix keeps D-19's location (moving the test to `api/ffi` would contradict it) |
| MNT-2 | MNT | major | `Reader.For` stores then drops instead of using `GetMany`'s "leave it out" | Fix now → F-2 | verified: `reader.go:386-401` `absent` flag + `forAnswers.Drop`; `Keyed.Get` (`cache.go:230`) always stores |
| SEC-1 | SEC | minor | `pivot` no longer checks the stored result's kind | Fix now → F-3 | verified: `reader.go:363` `pivots.Get` never reads `res.kind`; base compared `res.kind == kind`. Regression in code this PR changes |
| CON (from MNT outside-lens) | CON | minor | `namealiases`/`nameformats` `Warm`/`Invalidate` return nil on a closed catalog | Fix now → F-4 | verified: `Catalog.Close` (`catalog.go:213`) clears `db`, leaves the graph; base returned `ErrClosed` via `c.DB()`. D-10's nil-graph rule stays; add the closed check |
| MNT-3 | MNT | minor | Profiler setup written three times in `promotealign` | Fix now → F-5 | same-pass cleanup; two copies added by this PR |
| MNT-4, CON-1 | MNT, CON | minor | Garbled `Graph` struct comment | Fix now → F-6 | `graph.go:34-35` |
| MNT-5 | MNT | minor | Registry rule omits the benchmark and doc-test steps | Fix now → F-6 | `.cursor/rules/catalog-queries.mdc:18` |

Nits (logged): `MNT-6 · graphcache/commit.go:13-14, graph.go Drop, evidence_graph.go · rewrap long comments` · `MNT-7 · graphcache/cache.go:158-160 · info.build any + dead guard` · `MNT-8 · persondetail/family.go:29 · second Kinship rules copy` · `CON-2 · graphcache/commit.go:13-14 · rewrap` · `CON-3 · spike-16/deployment-plan.md · "Reads" → "DerivedFrom"; capital T in S16-20 row` · `CON-4 · graphcache/cache_test.go:208 · AC-1.10 on Over(db), not a tx`

### Fix plan 1

**F-1 — The doc test fails when a cache owner is not linked** (from MNT-1)
Files: `core/catalogsession/cache_doc_test.go`.
Failing test first: add `TestCacheDocSeesEveryOwner` (same file). It walks `core/` and `api/` for non-`_test.go` Go files that call `graphcache.RegisterKeyed(` or `graphcache.RegisterWhole(`, maps each to its package import path (`github.com/mendahu/provenencia/` + dir), parses `cache_doc_test.go`'s own imports with `go/parser`, and fails naming each owner package not imported. Prove it red by temporarily removing the `promotealign` blank import, then restore.
Do: keep the test in `catalogsession` (D-19). Update the comment: the companion test proves this file imports every owner.
Check: both tests pass; scripts/check-pr.sh PASS.

**F-2 — One rule for "no value" in a keyed entry** (from MNT-2)
Files: `core/database/canonlinks/reader.go`.
Failing test first: none new; the existing `For` tests on a non-node id and `PROVENENCIA_GRAPH_VERIFY=1` canonlinks tests cover it.
Do: write `Reader.For` on `forAnswers.GetMany(r.g, [][]byte{id}, …)` with a fill that leaves the id out of its map when `around` finds no node. Delete the `absent` flag and the `forAnswers.Drop` call. Do not change `Keyed.Get`.
Check: `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/...` passes; scripts/check-pr.sh PASS.

**F-3 — A kinship pivot stored under one kind is not served for another** (from SEC-1)
Files: `core/database/canonlinks/reader.go`, `reader_test.go`.
Failing test first: `TestPivotKindMismatch` — a catalog with a relationship association whose second end is an event handle; call `For` on that person first, then on a participant of the event; assert the participant's implied links equal those from a fresh reader that only reads the participant.
Do: in `pivot`, `Peek` first; if a value is held and its `kind` matches, return it; if held with another kind, derive and return without storing; else `Get` as today.
Check: test red then green; scripts/check-pr.sh PASS.

**F-4 — `Warm` and `Invalidate` refuse a closed catalog** (from MNT outside-lens → CON)
Files: `core/database/namealiases/namealiases.go`, `core/database/nameformats/nameformats.go`, their tests.
Failing test first: `TestWarmClosedCatalog` in each package — open, `Close`, `Warm` and `Invalidate` return `database.ErrClosed`.
Do: before the nil-graph check, call `c.DB()` and return its error (as base did). Keep D-10's nil-graph `ErrClosed`.
Check: tests red then green; scripts/check-pr.sh PASS.

**F-5 — One helper for the shortlist's inputs** (from MNT-3)
Files: `core/database/promotealign/index.go`, `export_test.go`.
Do: add one unexported helper returning the config and profiler (`loadConfig(g)`, `valuerules.Metas(g)`, `newProfiler`/`newProfilerWith`); `Warm`, `checkMatchIndex` and `EmptyShortlistForTest` call it.
Check: `grep -n "valuerules.Metas" core/database/promotealign/*.go` prints only the helper's line; scripts/check-pr.sh PASS.

**F-6 — Comment and guide fixes** (from MNT-4, CON-1, MNT-5)
Files: `core/database/graphcache/graph.go`, `.cursor/rules/catalog-queries.mdc`.
Do: end the `displays` sentence at `graph.go:34`; make "The typed fields below are not registry entries (docs/catalog-caches.md)." its own sentence. In the rule line 18, add that an owner adds its fill to `BenchmarkCacheSizes` and a blank import to `core/catalogsession/cache_doc_test.go`.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `3b91dc2d`. Gate PASS (orchestrator). F-1 – F-6 done. Deviations: F-3's test fixture differs from the plan's (the plan's passed before the fix; the repo's test has a person who is both an event participant and the parent between, and failed red first); F-5 leaves `promotealign.go:139`'s `valuerules.Metas` in `loadInputs`, which is the layer's own read, not a profiler setup copy.
Minors checked by the orchestrator: SEC-1 **Resolved** (`pivot` Peeks and refuses a kind mismatch without storing); CON/F-4 **Resolved** (`c.DB()` checked first in both `Warm` and `Invalidate`); MNT-3 **Resolved** (`shortlistInputs`); MNT-4/CON-1 **Resolved**; MNT-5 **Resolved**.

## Round 2 — review
Reviewers: MNT (had Fix-now majors MNT-1, MNT-2). CON had none; SEC+PRF had only a minor.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | — | Doc test owner coverage | Resolved | `TestCacheDocSeesEveryOwner` finds all four registering packages; red with an import removed |
| MNT-2 | MNT | — | `Reader.For` store-then-drop | Resolved | `For` on `GetMany`; no `absent`/`Drop` |

No new findings. No Fix plan 2.

## Result
CI: first run of Go tests failed `TestLargeSourceBudgets/warm` (376 ms vs 375 ms) and `/compose_every_summary` (519 ms vs 500 ms) on the runner. Locally base `01fa954b` vs head, three runs each: warm 124–134 vs 128–138 ms, compose 187–196 vs 190–194 ms; no regression. Re-ran the failed job: pass.

Gate PASS · CI green · open: none · follow-ups: — · to user: Go caches ≈118 MB at 16,000 persons vs the 40 MB budget (reported per plan §10; the caches are today's)
