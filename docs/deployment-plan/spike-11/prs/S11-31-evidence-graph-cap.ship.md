# Ship log — S11-31 Bound the loaded Evidence graphs

**Plan:** [`S11-31-evidence-graph-cap.md`](S11-31-evidence-graph-cap.md) · **Branch:** `s11-31-evidence-graph-cap` · **PR:** #541
**State:** stopped — PRF-1 to user (contradicts the plan's §3 Out line)

## Round 0 — implement
Head `4d9c197d`. Gate PASS (orchestrator run against `22e3878e`). Report: no deviations; manual check (§10) not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (diff touches `core/database/graphcache`, a cache under `core/database/**`). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | A request that reads more than the cap of graphs evicts and reloads its own graphs on every open | To user → user 2026-10-24: Fix now → F-1, supersedes §3 Out line | verified: `promotealign/evidence.go` `memberValues` (l.334), `bridges` (l.424), `titles` (l.479) each read one graph per distinct member Source with no bound; `EvidenceGraph` evicts on every miss. "One request loads at most 2" (§3 Out, §4, §11) is a fixture measurement, not a code bound. Contradicts the §3 Out line "Pinning graphs for the length of one request: not planned" |
| CON-1 | CON | minor | Plan has no Landed note | Fix now → F-2 | DoD asks for one; grep shows only the Status line |
| MNT-1 | MNT | minor | Domain doc repeats the cap's value (64) | Fix now → F-3 | `docs/promote-graph-alignment.md:165`; minor in a file this PR changes |

Nits (logged): `CON-2 · graphcache/graph.go:342 · LoadedEvidenceGraphForTest not nil-guarded like its siblings` · `MNT-2 · evidence_graph.go:129 · local min shadows the builtin; rename oldestUsed` · `MNT-3 · graph.go:361 · LoadedEvidenceGraphCountForTest doc lacks "Tests only."` · `MNT-4 · evidence_graph.go:113 · comment names PR id S11-21; describe the fixture instead`

No SEC findings.

### Fix plan 1

The user decided PRF-1 on 2026-10-24: fix it in this PR with a request-local map. This supersedes the plan's §3 Out line "Pinning graphs for the length of one request: not planned" and the "at most 2 graphs" premise in §4 and §11. Record that in the plan's Landed note (F-2).

**F-1 — An Evidence request reads each Evidence graph at most once** (from PRF-1)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/perf_test.go`.
Failing test first: a new `TestLargeSourceBudgets` subtest `"evidence under a cap of one"`, placed after `"evidence under a small cap"` and following its shape. For each request in `evidenceRequests()`:
- `f.c.Graph().Drop()`, then count the queries for `f.evidence(req)` at the default cap with the existing `SetQueryCounterForTest` (see how `"warm queries"` counts), and keep the response.
- `Drop()` again, call `SetEvidenceGraphCapForTest(1)` (reset it with `defer ...(0)`), count the queries for the same request, and keep the response.
- Assert that the query counts are equal (each graph loads once per request, whatever the cap) and that the responses are `reflect.DeepEqual`.
It must fail at `4d9c197d`: a cold request reads 2 graphs, and `titles`/`bridges` re-read a graph that `memberValues` loaded after it was evicted. If it does not fail, report **Not reproduced** with the output and change nothing in `evidence.go`.
Do:
- Add `graphs map[string]graphcache.EvidenceGraph // by graph id, the graphs this request has read` to `evidenceReader` and initialize it in `newEvidenceReader`.
- Add one method, `func (r *evidenceReader) graph(id []byte) (graphcache.EvidenceGraph, error)`, that returns the local copy or calls `r.g.EvidenceGraph(id)` and stores the result.
- Route the three reads through it: `memberValues` (line ~334), `bridges` (line ~424) and `titles` (line ~479).
- Leave `memberMemo`/`SetEvidenceGraphMemo` as they are. A memo on a graph evicted later in the same request is lost by design. The next open recomputes it, which is correct.
- Leave the row graph read at line 94 (`rowEG`) alone unless it runs on the same reader; if it does, route it too.
Check: `grep -n "r.g.EvidenceGraph(" core/database/promotealign/evidence.go` prints only the line inside `graph`. The new subtest passes. `scripts/check-pr.sh` PASS.

## Round 1 — fix (attempt 1)
Head `4d9c197d`, no commits. F-1 was **not reproduced** on the large fixture. At a cap of one, all 62 requests gave the same query counts (325/330/332) and equal responses, because the fixture's cold requests never re-read a graph after a later load evicts it. F-2 and F-3 were held because they describe F-1. The defect is still real in code (PRF-1 verified by reading the code). The large fixture just does not exercise it, so the user's Fix now decision stands and F-1 moves its red test to a fixture that reads 3 graphs. See F-1 rev. 2.

**F-1 rev. 2 — An Evidence request reads each Evidence graph at most once** (from PRF-1; replaces F-1)
Files: `core/database/promotealign/evidence.go`, `core/database/promotealign/evidence_test.go`.
Failing test first: `TestEvidenceReadsEachGraphOnce` in `evidence_test.go`, placed after `TestEvidenceNamesEverySource`. Build the same catalog as that test: the row Subject on "Obituary" (srcB), the promoted person's members on "Register" (`f.source`) and "Census 1851", and the birth neighbor. That makes one request read 3 Evidence graphs: the row graph, then two member graphs in `memberValues`, and `titles` reads both member graphs again. Then:
- `f.c.Graph().Drop()`. Count the queries for `f.evidence(req)` at the default cap with `SetQueryCounterForTest` (copy how `perf_test.go` "warm queries" uses it), and keep the response.
- `Drop()` again, call `f.c.Graph().SetEvidenceGraphCapForTest(1)` with `defer ...(0)`, count again for the same request, and keep the response.
- Assert equal query counts and `reflect.DeepEqual` responses.
At `4d9c197d` the cap-1 run must count more queries, because `titles` reloads a member graph that `memberValues` evicted. If it does not, report Not reproduced with both counts and change nothing.
Do: as in F-1. Add the `graphs` map to `evidenceReader`, initialized in `newEvidenceReader`, and one `graph(id)` method. Route `memberValues`, `bridges` and `titles` through it. If the line-94 `rowEG` read happens before the reader exists, seed `r.graphs` with it once the reader is built, so the row graph also counts as read. Do not add the perf_test subtest.
Check: `grep -n "r.g.EvidenceGraph(" core/database/promotealign/evidence.go` prints only the line inside `graph`. `TestEvidenceReadsEachGraphOnce` passes. `scripts/check-pr.sh` PASS.

**F-2 — Landed note in the plan** (from CON-1)
Files: `docs/deployment-plan/spike-11/prs/S11-31-evidence-graph-cap.md`.
Do: add a **Landed** note under the header. It says that review round 1 (PRF-1) found one Evidence request reads one graph per distinct member Source, with no bound at 2. By the user's decision on 2026-10-24, `evidenceReader` now keeps the graphs it reads for the length of the request. This supersedes the §3 Out line on pinning and the "at most 2" premise in §4/§11. The note names `TestEvidenceReadsEachGraphOnce` and says the large fixture could not show the reload, because its cold requests never re-read an evicted graph.
Check: plan-docs passes.

**F-3 — Name the cap in the domain doc** (from MNT-1)
Files: `docs/promote-graph-alignment.md` (line ~165).
Do: replace "at most 64 Evidence graphs" with "at most `maxEvidenceGraphs` Evidence graphs (`graphcache`)". Add one clause saying that one Evidence request reads each graph once (F-1). Leave `completed.md` alone.
Check: `grep -n "64" docs/promote-graph-alignment.md` shows no cap value.
