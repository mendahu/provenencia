# Ship log — S11-49 Kinship links in the graph cache

**Plan:** [`S11-49-kinship-graph-cache.md`](S11-49-kinship-graph-cache.md) · **Branch:** `s11-49-kinship-graph-cache` · **PR:** #569
**State:** ready

## Round 0 — implement
Head `2537e8a3`. Gate PASS (orchestrator run, against `cf186b21`); `PROVENENCIA_GRAPH_VERIFY=1` over `canonlinks` and `graphcache` passes. Report deviations:
1. D-12: `graphcachetest.Canon` backs every value with an Observation on an accepted member Subject and runs `autoreconciler.RecomputeTx`, because raw `auto_reconciler_values` inserts fail graph Verify (AC-2.6). `Term` finds a term by key at any origin.
2. D-4: an event result keeps only links with an `Event`; a chain inside a malformed event belongs to the person's result.
3. STEP-7 and STEP-8 share one commit.
Implementer note: `Deps` reads the cached node, so an entity deleted alone (no association in the commit) names no neighbors; untested.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`macos/App/Platform/*.swift`, `Localizable.xcstrings`), SEC+PRF (`core/matchregistry`, `core/database/**`, graph cache).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | blocker | `canonlinks` repeats kinship's parent-step rule (`reader.go:117`, `isParentStep`) | Fix now → F-2 | verified: `kinship.go:80`/`:84` (`Keys` + `Edge.reads`) vs `reader.go:117`, `:331-335` |
| SEC-1 | SEC/PRF | major | `kinship.chains` is quadratic in duplicate parent steps; D-4 makes each filed link its own step | Fix now → F-1 | verified: `kinship.go:219-231` loops every step × `byG[up.m]` before `made[t]`; reviewer's scratch run 8,000 duplicates → 2.1 s, doubling ×4 |
| MNT-2 | MNT | major | `EventTypeProperty` is a third key+origin → id scan | Fix now → F-3 | verified: `canonlinks.go:45-52`, `evidence_graph.go:236-243` closure |
| CON-1 | CON | minor | An entity deleted alone drops no neighbor results | Fix now → F-4 | `commit.go:48-58`: `removeNode` runs before `dropMemos`, so `deps` sees no node; in code this PR adds |
| MNT-3 | MNT | minor | `strict` flag on `deriver` | Fix now → F-5 | file this PR adds |
| MNT-4 | MNT | minor | `For` and `chainLinks` repeat the event-neighbor/birth loop | Fix now → F-6 | same-pass; `reader.go:167-195`, `:265-284` |
| MNT-5 | MNT | minor | Test copies the bound with a wrong reason | Fix now → F-7 | `reader_test.go:21-22` |
| MNT-6 | MNT | minor | No rule names `RegisterNodeMemo`; `displays` move unrecorded | Fix now → F-8 (rule line) · Follow-up (the `displays` move) | the move is an Out line; PR body follow-up |
| PRF-1 | PRF | minor | Verify rebuilds the deriver per stored result | Rejected | verify mode only (`PROVENENCIA_GRAPH_VERIFY=1`); release builds never run it; per-pass state would need a new hook on the shared `MemoCheck` |
| UX-1 | UX | minor | Pre-existing projects keep a locked, unlocalized `spouse` role that implies nothing | Rejected (as code) | user decision 2026-10-26: no migration or backfill; noted in the PR body's deviations |
| CON (outside MNT) | CON | minor | `EdgeOf` reads only the provenencia `event_type`; the old `keptTerm` matched any origin | Rejected | plan D-7 names origin provenencia; noted in the Landed note |

Nits (logged): `MNT-7 · promotealign/layer.go:148 · neighborTypeProperty beside profiler.eventTypeProp` · `MNT-8 · canonlinks/reader_test.go:67-77 · literals where vocabkey constants exist`

### Fix plan 1

**F-1 — Linear grandparent chains over duplicate parent steps** (SEC-1)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`.
Failing test first: `TestDeriveManyAssociations` gains a case: n = 8,000 single-association edges g→m and 8,000 m→c, `Rules{MaxPivotEnd: 10}` → one grandparent link, under 1 s.
Do: in `chains`, iterate distinct (g, m) pairs in first-seen order (an order slice beside `byPair`) and, per pair, the distinct children of m in first-seen order (an order beside `byG`); build each triple's `From` and associations from `byPair` as today. Output order unchanged for existing tests.
Check: the new case passes; `go test ./core/kinship/` passes unchanged otherwise.

**F-2 — One parent-step rule, in `core/kinship`** (MNT-1)
Files: `core/kinship/kinship.go`, `core/kinship/kinship_test.go`, `core/database/canonlinks/reader.go`.
Do: export `func ParentStep(e Edge, terms segmentcompare.Terms) (parent, child []byte, ok bool)` built on `Edge.reads` with the parent keys and `appendStep`'s self-link rule (false when parent equals child). `Derive` uses it to collect steps. In `canonlinks`, delete `deriver.parentFirst`, `parentSecond` and `isParentStep`; `For` and `chainLinks` call `kinship.ParentStep` and take the far end from its result. Add a `TestParentStep` table (direct, reversed ends, a `child`-term edge read through `Keys`, a self-link, a spouse edge).
Check: `go test ./core/kinship/` and `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonlinks/` pass; `grep -n 'parentFirst\|isParentStep' core/database/canonlinks` prints nothing.

**F-3 — One property-id lookup by key and origin** (MNT-2)
Files: `core/database/graphcache/vocab.go`, `core/database/graphcache/evidence_graph.go`, `core/database/canonlinks/canonlinks.go`, `core/database/canonlinks/reader.go`, `core/database/promotealign/profiles.go` (and any other `EventTypeProperty` caller).
Do: add `func (v Vocabulary) PropertyID(key, origin string) []byte` beside `PropertiesByID`; `evidence_graph.go`'s `propID` closure calls it; delete `canonlinks.EventTypeProperty`; callers use `v.PropertyID(vocabkey.PropertyEventType, vocabkey.OriginProvenencia)`.
Check: `grep -rn 'EventTypeProperty\|propID := func' core` prints nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes.

**F-4 — A deleted entity drops its neighbors' memos** (CON-1)
Files: `core/database/graphcache/commit.go`, `core/database/canonlinks/reader_test.go`.
Failing test first: `TestForFollowsCommits` subtest "an event deleted alone": AC-1.1, `Remove(M1)` (the event only, not its participations) and `Commit(M1)` → `For(Ann)` is empty.
Do: in `OnCommit`, before `g.removeNode(id)` for an entity whose row is gone, append the cached node's `Link.Neighbor`s to `roots`.
Check: the subtest passes, also under `PROVENENCIA_GRAPH_VERIFY=1`.

**F-5 — Inject the node read instead of a `strict` flag** (MNT-3)
Files: `core/database/canonlinks/reader.go`.
Do: `deriver` gets `node func([]byte) (*graphcache.Node, error)`; `NewReader` sets `g.Node`; `check` sets a closure over `g.Cached` that errors on a missing node; drop `strict`.
Check: canonlinks tests pass, also under verify mode.

**F-6 — One event-neighbor loop** (MNT-4)
Files: `core/database/canonlinks/reader.go`.
Do: one helper, `eventLinksAt(n *graphcache.Node, ...) ([]kinship.Implied, error)`, returns the event-result links with the node's id as an end, each event once; `For` keeps them all, `chainLinks` filters to `PatternBirthParents`.
Check: canonlinks tests pass.

**F-7 — Tests read the registry's bound** (MNT-5)
Files: `core/database/canonlinks/reader_test.go`.
Do: replace the copied `shipped` rules and its comment with `matchregistry.Default().Kinship`.
Check: canonlinks tests pass.

**F-8 — The rule names the per-node memo** (MNT-6)
Files: `.cursor/rules/catalog-queries.mdc`.
Do: one line: a per-node result derived from the graph cache registers a `graphcache.NodeMemo` (`RegisterNodeMemo`, with its `Deps` and `Check`); never a new map field on `Graph`.
Check: `scripts/check-pr.sh` PASS.

Also: the plan's Landed note gains lines for F-1 – F-4 and the `event_type` origin (D-7).

## Round 1 — fix
Head `2b52062c`, then gate fix `b0f7e00d`. Gate PASS (orchestrator run); verify mode over `canonlinks`, `graphcache`, `kinship` passes. F-1 – F-8 done. Deviations: F-1's test is `TestDeriveManyDuplicateParentSteps` (failed at 2.1 s first); F-4's planned test shape could not run (deleting an event alone fails its foreign key), so the subtest removes a birth with its participations and commits only the birth — it failed before the fix; it commits a deliberately inconsistent handle set, so it skips under verify mode (gate fix, comment and Landed note); `chainLinks` feeds relationship edges before birth edges; only `For` calls `ParentStep`.
CON-1 (minor, F-4): **Resolved** — read `fc0fb244` and `b0f7e00d`.

## Round 2 — review
Reviewers: MNT (MNT-1 blocker, MNT-2 major), SEC+PRF (SEC-1 major). Scope `2537e8a3...b0f7e00d`.
MNT: MNT-1 and MNT-2 **Resolved** (`ParentStep` in `core/kinship`; `Vocabulary.PropertyID`; old copies gone). No new blocker. Outside its lens: `ParentStep` recomputes keys per edge (constant factor) — logged as a follow-up.
SEC+PRF: SEC-1 **Resolved** (`chains` loops distinct triples; 8,000 duplicates in 0.03 s). No new blocker.
`origin/main` had not moved; gate PASS on `b0f7e00d`.

## Result
Gate PASS · CI pending · open: none · follow-ups: move `displays` onto memo slots; `ParentStep` per-edge key cost · to user: — · CON miss: none
