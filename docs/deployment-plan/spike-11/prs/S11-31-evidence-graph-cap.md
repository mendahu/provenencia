# S11-31 — Bound the loaded Evidence graphs

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 7 — Performance
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine cache gains a cap; nothing on screen changes)
**Board:** —
**Depends on:** S11-19 (landed [#513](https://github.com/mendahu/provenencia/pull/513)), S11-21 (landed [#536](https://github.com/mendahu/provenencia/pull/536))
**Blocks:** —
**Branch:** `s11-31-evidence-graph-cap`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** — (no catalog, FFI, matching-rule or UI change; `graphcache` has no skill of its own)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) "Loads read the catalog's in-memory graph"; S11-19 plan D-21 and §14 (Evidence graph memos); S11-21 plan §6 "Next adds" (this PR adds a subtest to `TestLargeSourceBudgets`); working rule "High tolerance for churn"; dogfood bar 3 ("Fast, and it explains itself")
**Planned:** 2026-10-25 against `main` @ `933cc3c` (first drafted 2026-10-24 against #536's head `74f64d0`; refreshed after S11-21 landed — no Go file changed between the two)
**Review:** [`S11-31-evidence-graph-cap.review.md`](S11-31-evidence-graph-cap.review.md) — Fits with changes · Ready with changes, 1 round, all findings answered

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

The catalog's in-memory graph keeps every Evidence graph it has loaded, with the memos features build on it, until a commit names that graph or the catalog closes. A researcher who opens the evidence sheet across many rows, or browses many Sources in one session, grows that map without limit. After this PR the graph holds at most `maxEvidenceGraphs` (64) loaded Evidence graphs and drops the least recently used one, with its memos, when a load goes past the cap. A dropped graph loads again on its next read and gives the same rows.

**Why now.** S11-19 made the claim evidence query read every member's Evidence graph and memoize member values on it (S11-19 §14, D-21), and S11-21 made warm reads issue no query, so the loaded set now only grows. S11-19's ship log left eviction to this PR.

**What stays.** What a commit drops, `Drop`, what `dropDerived` drops, the memo API's signatures, the loaders, and every caller of `EvidenceGraph` are unchanged. The fixture's warm Evidence pass still issues no query: its 21 graphs fit under the cap.

## 2. User stories

**US-1.** As a researcher in a long session, when I open evidence and Source pages across many Sources, the app's memory for loaded Evidence graphs stays bounded, so a large catalog does not slow the app over the session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Cap 2; graphs A, B, C are read in that order | B and C are loaded; A and A's memos are not |
| AC-1.2 | Cap 2; read A, read B, read A again through `EvidenceGraph`, read C | A and C are loaded; B is not |
| AC-1.3 | Cap 2; read A, read B, a hit on A's memo through `EvidenceGraphMemo`, read C | A and C are loaded; B is not |
| AC-1.4 | A was evicted; A is read again | the returned `EvidenceGraph` deep-equals A's first read; `EvidenceGraphMemo(A, k)` reports no memo until a feature sets it again |
| AC-1.5 | Cap 2; read A, B; `Drop`; read C, D | C and D are loaded |
| AC-1.6 | Cap 2; read A, B, C (A evicted); `Verify` | `Verify` returns nil and A is still not loaded afterwards |

**US-2.** As a researcher, when I open the evidence sheet on every row of a page whose records hold members from many Sources, the sheet shows the same evidence whether or not graphs were evicted along the way.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | S11-21's large fixture; Evidence on every handle row of Source A (21 graphs) with the cap set to 8 | at most 8 Evidence graphs are loaded after each request, and every response deep-equals the response with the default cap |
| AC-2.2 | S11-21's large fixture; the default cap | `TestLargeSourceBudgets` "warm queries" still sees 0 graph queries and 0 SELECTs for a warm Evidence pass |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` holds at most `maxEvidenceGraphs` loaded Evidence graphs and evicts the least recently used one with its memos (D-1, D-3, D-4) | STEP-1, AC-1.1 – AC-1.5 |
| A read through `EvidenceGraph` or a memo hit through `EvidenceGraphMemo` counts as use (D-2) | STEP-1, AC-1.2, AC-1.3 |
| One named constant, sized from S11-21's fixture (D-4) | STEP-1, AC-2.2 |
| `Verify` checks only graphs that are still loaded | STEP-1, AC-1.6 |
| A test-only cap override and loaded count (D-5) | STEP-1, STEP-2 |
| An evicted graph reloads with the same rows and its memo is rebuilt | STEP-1, STEP-2, AC-1.4, AC-2.1 |
| The many-Sources page check: Evidence on every row stays under the cap and gives the same evidence | STEP-2, AC-2.1 |
| Docs: domain doc, completed.md, map, this plan | STEP-3 |

| Out | Home |
| --- | --- |
| Evicting nodes (`g.nodes`), kind indexes, displays or details | not planned — the map's scope entry excludes nodes; nodes are bounded by the catalog |
| Evicting or bounding the match index | not planned — the map's scope entry excludes it; it is one value per catalog |
| A cap sized in bytes, or a cap the researcher sets | not planned — no size measure exists, and a count fits the observed working set (D-4) |
| Pinning graphs for the length of one request | not planned — one request loads at most 2 graphs (§4), far under the cap |
| Changing which graphs a commit drops, or what `dropDerived` drops | not planned |
| Evidence graph card image memory in the app | [`ideas/evidence-graph-card-memory.md`](../../../ideas/evidence-graph-card-memory.md) — a separate budget: Swift card images, not engine rows |

## 4. Current state

Verified at `main` @ `933cc3c`, after S11-21 landed.

| Area | Where | Today |
| --- | --- | --- |
| Loaded graphs | `core/database/graphcache/graph.go` `Graph.evidenceGraphs` (`map[string]EvidenceGraph`), `Graph.memos` (`map[string]map[string]any`) | Two maps keyed by graph id, kept in step by convention ("A graph and its memos go together"). Nothing evicts: `clear` (from `New` and `Drop`) empties both, `dropEvidenceGraphs` removes named ids from both, `dropDerived` replaces `memos` with an empty map and resets `matchIndex` / `matchIndexOK` |
| Reads | `core/database/graphcache/evidence_graph.go` `Graph.EvidenceGraph` | Returns the map entry on a hit; on a miss calls the registered `evidenceGraphLoad`, sets `GraphID`, stores the result, returns it. No use order |
| Commit drop | `evidence_graph.go` `Graph.dropEvidenceGraphs` | Deletes each named id from `evidenceGraphs` and `memos`. Called from `core/database/graphcache/commit.go` `OnCommit` with `set.EvidenceGraphs` |
| Memos | `evidence_graph.go` `Graph.EvidenceGraphMemo`, `Graph.SetEvidenceGraphMemo` | Read returns `memos[graph][key]`. Set stores only when the graph is in `evidenceGraphs`. The doc comment lists what clears a memo; eviction is not on it |
| Verify | `core/database/graphcache/verify.go` `Graph.Verify` | Ranges over `g.evidenceGraphs` and checks each stored graph with `evidenceGraphFn`. It already checks only loaded graphs and loads nothing |
| Test hooks | `graph.go` `SetQueryCounterForTest` (nil-guarded), `LoadedEvidenceGraphForTest` (not nil-guarded), `ResetMatchIndexForTest` | The pattern for a test-only setter and reader on `Graph` |
| Graph readers | `core/database/promotealign/layer.go` (memo first, `EvidenceGraph` on a miss), `core/database/promotealign/evidence.go` `memberValues`, `bridges`, `titles`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/pins.go`, `api/ffi/handlers/subjects.go`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/observations.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/event_titles.go` (each `c.Graph().EvidenceGraph(graphID)`) | Each takes the returned `EvidenceGraph` value and uses it within the call |
| Memo-only readers | `core/database/promotealign/canon.go` `loadCandidatesMemo` (candidates memo), `core/database/promotealign/export_test.go` `MemoSnapshotForTest` | Reach a graph only through `EvidenceGraphMemo` / `SetEvidenceGraphMemo`. With `layer.go` this is why a memo hit is a use (D-2) |
| Working set | measured on S11-21's `newLargeFixture` at `933cc3c` with a throwaway test that counted `LoadedEvidenceGraphForTest` over every `evidence_graphs` id | 23 Evidence graphs in the catalog. A cold `Propose` on Source A loads 1. Evidence on all 62 handle rows of Source A loads 21. One cold Evidence request (after `Drop`) loads at most 2. A cold `Propose` on the one-place Source loads 1, and Evidence on all 33 of its handle rows still loads only that 1: its city handle's other events are bulk rows with no members |
| Budget test | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`; `newLargeFixture`; `evidenceRequests` and `pass`, closures inside `TestLargeSourceBudgets` | Subtests share one fixture and run in order. `evidenceRequests` builds one request per handle row of Source A (`aGraph`). "warm queries" requires a warm Evidence pass to issue 0 graph queries and 0 SELECTs, so every graph of the pass must stay loaded at the default cap |
| Memo tests | `core/database/graphcache/evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource`; fixture `newWalkFixture`, `walkFixture.newSource(typeID, title)` in `core/database/graphcache/evidence_graph_walk_test.go` | Cover memo drop by commit, `Drop`, vocabulary and `SetAliasIndex`. `walkFixture` does not keep the Source type id; `TestWalkEvidenceGraphStaysOnItsGraph` (line 255) gets it again with `sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})` |
| Verify tests | `core/database/graphcache/graph_test.go` (calls `g.Verify()` at lines 51 and 55) | The graphcache test package has the truth recompute registered |
| Domain doc | `docs/promote-graph-alignment.md` bullet "Loads read the catalog's in-memory graph" | Says the graph keeps Evidence graphs between calls; says nothing of a bound |
| Map | `docs/deployment-plan/spike-11/deployment-plan.md` S11-31 row and scope entry | Reworded by this plan's commit to Evidence graphs (review M-2) |

**Precedent.** [#513](https://github.com/mendahu/provenencia/pull/513) (S11-19 D-21) added `memos` beside `evidenceGraphs` and made "a graph and its memos go together"; this PR makes that pairing one entry. Copy `SetQueryCounterForTest` (nil guard included) for the test hooks. There is no earlier eviction in Go; the Swift session cache's `.evict` (`docs/macos-client-patterns.md`) is a different layer and is not reused.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **One entry per loaded graph.** `loadedEvidenceGraph` *(new, `evidence_graph.go`)* is `struct{ graph EvidenceGraph; memos map[string]any; used uint64 }`. `Graph.evidenceGraphs` becomes `map[string]*loadedEvidenceGraph`; the `Graph.memos` field is deleted. `Graph` gains `evidenceTick uint64` and `evidenceCap int` *(new)*. Eviction scans the entries for the smallest `used` | The pairing S11-19 asked for becomes the type: forgetting a graph is one `delete`, and a memo cannot outlive its graph. At a cap of 64 a linear scan on each load is negligible beside the SQL load it follows. Stamps strictly increase, so there are no ties (review M-1) | A third parallel map `evidenceUsed` — three maps kept in step by convention; `container/list` LRU — pointer bookkeeping for an O(1) the cap does not need |
| D-2 | **What counts as use.** `touchEvidenceGraph(e *loadedEvidenceGraph)` *(new)* sets `e.used = ++g.evidenceTick`. It runs in `EvidenceGraph` on a hit and after a load, and in `EvidenceGraphMemo` when it returns a memo. `SetEvidenceGraphMemo` does not touch (it always follows a read of the same graph). `EvidenceGraphMemo`'s doc comment says a hit counts as use | `layer.go` and a warm `Propose` reach the row's graph only through its memo; without the touch the graph Promote is working on would look idle and be evicted by a run of member loads | Touching only in `EvidenceGraph` — evicts the open page's own graph during an evidence pass |
| D-3 | **Where graphs leave.** `dropEvidenceGraphs` deletes each named entry. `evictEvidenceGraphs()` *(new)* deletes the smallest-`used` entry while `len(g.evidenceGraphs) > g.evidenceCap`. `clear` sets `evidenceGraphs` to an empty map; `evidenceTick` and `evidenceCap` are not reset. `dropDerived` sets `e.memos = nil` on every entry (the cap bounds that loop) and resets the match index as today. `SetEvidenceGraphMemo` writes into the entry, creating `e.memos` when nil, and does nothing when there is no entry | Every path that removes a graph removes its memos and its stamp with it, by construction | Keeping a `forgetEvidenceGraph` helper — with one map it is a bare `delete` |
| D-4 | **The cap is `maxEvidenceGraphs = 64`** *(new, unexported const, `evidence_graph.go`)*, commented: "maxEvidenceGraphs is how many Evidence graphs stay loaded. A full evidence pass over S11-21's 100-Subject Source reads 21; 64 keeps about three such pages warm." `New` sets `g.evidenceCap = maxEvidenceGraphs`. `EvidenceGraph` on a miss: load, set `GraphID`, store the entry, touch it, then `evictEvidenceGraphs()` | 64 keeps "warm queries" at 0 queries with room to spare; one graph is one Source page's Subjects and Observations, so 64 is a few megabytes. The just-loaded entry has the largest stamp and is never the one evicted | 32 — under two full pages; 256 — no measured need, and four times the memory; a byte budget — no size measure exists |
| D-5 | **Test hooks**, in `graph.go` beside `LoadedEvidenceGraphForTest`, nil-guarded like `SetQueryCounterForTest`. `SetEvidenceGraphCapForTest(n int)` *(new)*: `n > 0` sets the cap and calls `evictEvidenceGraphs()`; `n <= 0` restores `maxEvidenceGraphs`. `LoadedEvidenceGraphCountForTest() int` *(new)* returns `len(g.evidenceGraphs)`. `LoadedEvidenceGraphForTest` keeps its signature and reads the new map | The graphcache tests need a cap of 2 and the budget test a cap of 8, and both count loaded graphs; `LoadedEvidenceGraphForTest` needs ids the perf test does not hold | Exporting the constant and a production setter — no caller outside tests needs either |
| D-6 | **The budget subtest runs last** in `TestLargeSourceBudgets`, as `t.Run("evidence under a small cap", …)` *(new)* after "births at one place warm". It calls `reqs := evidenceRequests()`, then `f.c.Graph().Drop()`, then records `want[i] = f.evidence(reqs[i])` for every request at the default cap. Then `f.c.Graph().SetEvidenceGraphCapForTest(8)` with `defer f.c.Graph().SetEvidenceGraphCapForTest(0)`, `f.c.Graph().Drop()`, and runs every request in order, failing the subtest when `LoadedEvidenceGraphCountForTest() > 8` after a request or when `!reflect.DeepEqual(got, want[i])`. It runs on Source A, not the one-place Source | Last, so a small cap cannot change the timings or the 0-query check of earlier subtests. 8 is under the 21 the pass touches, so eviction runs, and over the 2 one request needs. Source A is the page whose records span many Sources; a full evidence pass on the one-place Source loads 1 graph (§4), so it would not evict (review M-3) | A new test with its own `newLargeFixture` — the fixture costs seconds; giving `evidenceRequests` a graph parameter and running the one-place Source too — it exercises no eviction |
| D-7 | **`Verify` reads the new map** (`e.graph` for each entry) and otherwise does not change. AC-1.6 pins that it passes after an eviction and does not load the evicted graph | The scope line is already true; a test keeps it true | Adding code to `Verify` — nothing to add |
| D-8 | **T-1's shape.** `TestEvidenceGraphCap` *(new)* has one table for the order cases (AC-1.1 – AC-1.3, AC-1.5) and three plain subtests (T-1d, T-1f, T-1g). The fixture is `newWalkFixture(t)` plus three more Sources made with `f.newSource(typeID, …)`, where `typeID` comes from `sourcetypes.Upsert` with the "book" input, as near line 255 of `evidence_graph_walk_test.go`. `walkFixture` is not changed. Each case starts with `g.Drop()` and `g.SetEvidenceGraphCapForTest(2)` (deferred reset to 0). Commit revisions in T-1g use the literal 10 | Op kinds fit a small table; the cases that capture a value or check a result mid-sequence read better as subtests. Reusing the existing `Upsert` call keeps the shared fixture as it is | Adding `typeID` to `walkFixture` — touches a shared helper for one test |

```go
// T-1 table shape (in evidence_graph_cap_test.go).
type capStep struct {
	op  string // "read" (EvidenceGraph), "memo" (SetEvidenceGraphMemo(id, "k", "v")), "hit" (EvidenceGraphMemo, must be ok), "drop" (g.Drop())
	src int    // index into the four graph ids A, B, C, D
}
cases := []struct {
	name      string
	steps     []capStep
	loaded    []int // must be loaded after the steps
	notLoaded []int
}{ /* T-1a, T-1b, T-1c, T-1e from §8 */ }
```

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `loadedEvidenceGraph` *(new)* | type | `Graph.evidenceGraphs` value `EvidenceGraph` plus `Graph.memos` | Replaces both: the graph, its memos and its stamp in one entry (D-1) |
| `maxEvidenceGraphs` *(new)* | const | `alignment.TopK` and other `matchregistry` knobs | A cache bound is not a matching knob; `matchregistry` is for comparison layers (working rule "One match registry") and `graphcache` does not import it |
| `Graph.evidenceTick`, `Graph.evidenceCap` *(new)* | fields | `Graph.count` (test counter) | Different job; no existing field holds use order or a bound |
| `touchEvidenceGraph`, `evictEvidenceGraphs` *(new)* | func | `dropEvidenceGraphs` | `dropEvidenceGraphs` removes named ids for a commit; eviction picks the id itself. Searched `core/`, `api/` for `evict`, `lru`: no Go eviction exists |
| `SetEvidenceGraphCapForTest` *(new)* | test hook | `SetQueryCounterForTest`, `ResetMatchIndexForTest` | Same pattern, new knob |
| `LoadedEvidenceGraphCountForTest` *(new)* | test hook | `LoadedEvidenceGraphForTest` | That takes an id; the budget subtest needs a count |
| `core/database/graphcache/evidence_graph_cap_test.go` *(new)* `TestEvidenceGraphCap` | test | `evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource` | Same fixture (`newWalkFixture`); a separate file because it is a separate behavior and needs four Sources |

**Builds on.** S11-19's pairing of a graph and its memos (#513); S11-21's large fixture and `TestLargeSourceBudgets`; the `…ForTest` hooks on `Graph`.
**Deletes.** The `Graph.memos` field and its separate upkeep in `clear`, `dropDerived` and `dropEvidenceGraphs`.
**Next adds.** None named. Sizing the cap from memory, if dogfood asks for it, puts a byte count on `loadedEvidenceGraph`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` (`graph.go`) | `evidenceGraphs map[string]*loadedEvidenceGraph`; `memos` field deleted and its comment folded into the `evidenceGraphs` comment; `evidenceTick uint64`, `evidenceCap int` *(new)*. `New` sets `evidenceCap`. `clear` resets `evidenceGraphs` only. `dropDerived` clears each entry's `memos`. `LoadedEvidenceGraphForTest` reads the new map; `SetEvidenceGraphCapForTest`, `LoadedEvidenceGraphCountForTest` *(new)* |
| `core/database/graphcache` (`evidence_graph.go`) | `loadedEvidenceGraph`, `maxEvidenceGraphs`, `touchEvidenceGraph`, `evictEvidenceGraphs` *(new)*, each with a one-line doc comment. `EvidenceGraph`, `dropEvidenceGraphs`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo` move to the entry (D-2, D-3, D-4). `EvidenceGraphMemo`'s doc comment adds: "A hit counts as a use of the graph. The graph and its memos are also dropped when it is the least recently used of more than maxEvidenceGraphs loaded graphs." |
| `core/database/graphcache` (`verify.go`) | `Verify` passes `e.graph` to `evidenceGraphFn` (D-7) |

```go
type loadedEvidenceGraph struct {
	graph EvidenceGraph
	memos map[string]any
	used  uint64
}

const maxEvidenceGraphs = 64

func (g *Graph) touchEvidenceGraph(e *loadedEvidenceGraph)
func (g *Graph) evictEvidenceGraphs()
func (g *Graph) SetEvidenceGraphCapForTest(n int)
func (g *Graph) LoadedEvidenceGraphCountForTest() int
```

`Over` graphs go through `New`, so they get the default cap. No new error.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Cap the loaded Evidence graphs**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/evidence_graph.go`, `core/database/graphcache/verify.go`, `core/database/graphcache/evidence_graph_cap_test.go` *(new)*. Skill: —.
Do: D-1 – D-5, D-7 as in §6 Go core. Write `TestEvidenceGraphCap` per D-8 and §8 T-1a – T-1g, table-driven per `.cursor/rules/go-tests.mdc` for the order cases. Keep `TestEvidenceGraphMemoDroppedWithSource` passing unchanged.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/` passes; `TestEvidenceGraphCap` covers AC-1.1 – AC-1.6.

**STEP-2 — Many-Sources page under a small cap**
Files: `core/database/promotealign/perf_test.go`. Skill: —.
Do: D-6: add the "evidence under a small cap" subtest (T-2) at the end of `TestLargeSourceBudgets`. It uses the `evidenceRequests` closure, `f.evidence` (which fails through the fixture's `t`; that is acceptable), and `reflect` (already imported).
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign/` passes, including "warm queries" (AC-2.2) and the new subtest (AC-2.1); then `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-3 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status, Landed note).
Do:
- `docs/promote-graph-alignment.md`: append to the "Loads read the catalog's in-memory graph" bullet: "It keeps at most 64 Evidence graphs; loading one more drops the least recently used graph with its memos, and a dropped graph loads again on its next read."
- `completed.md`: a new section at the top, above S11-21: `## S11-31 — Bound the loaded Evidence graphs`, then `**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N).` and two or three sentences of what shipped.
- `deployment-plan.md`: the S11-31 row's Status → `Landed #N`.
- This plan: Status → `Landed in [#N](…) (YYYY-MM-DD)` and a **Landed** note under the header.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-31-evidence-graph-cap.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1a | Go | `core/database/graphcache/evidence_graph_cap_test.go` `TestEvidenceGraphCap` *(new)* table row "oldest is evicted" | cap 2; read A, memo A, read B, read C → loaded B, C; not loaded A (so its memo is gone) | AC-1.1, D-3, D-4 |
| T-1b | Go | same, table row "a read is a use" | cap 2; read A, B, A, C → loaded A, C; not loaded B | AC-1.2, D-2 |
| T-1c | Go | same, table row "a memo hit is a use" | cap 2; read A, memo A, read B, hit A, read C → loaded A, C; not loaded B | AC-1.3, D-2 |
| T-1e | Go | same, table row "a drop clears the loaded set" | cap 2; read A, B, drop, read C, D → loaded C, D; not loaded A, B | AC-1.5, D-3 |
| T-1d | Go | same, subtest "an evicted graph reloads the same" | read A (keep the value), memo A, read B, C (A evicted), read A → deep-equals the kept value; `EvidenceGraphMemo(A, "k")` not ok | AC-1.4 |
| T-1f | Go | same, subtest "verify skips an evicted graph" | cap 2; read A, B, C; `g.Verify()` → nil; `LoadedEvidenceGraphForTest(A)` false | AC-1.6, D-7 |
| T-1g | Go | same, subtest "a commit drop removes the entry" | cap 2; read A, B; `g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 10}, Effects: effects.Set{EvidenceGraphs: [][]byte{A}}})`; read C, D → loaded C, D; not loaded A, B (B evicted by D, C kept) | D-3 |
| T-2 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "evidence under a small cap" *(new)* | large fixture; cap 8; Evidence on every handle row of Source A → after each request ≤ 8 loaded; each response deep-equals the default-cap response | AC-2.1, D-6 |
| T-3 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "warm queries" (existing, unchanged) | default cap → 0 graph queries, 0 SELECTs on a warm Evidence pass | AC-2.2, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestEvidenceGraphMemoDroppedWithSource`, `core/database/promotealign` `source_store_test.go` and `memo_test.go`, and `TestLargeSourceBudgets` "warm queries".

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped (STEP-3)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/promote-graph-alignment.md (STEP-3)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Build and run the app on a project with a Source of at least 20 promoted people whose records hold members from other Sources. Open Promote on that Source and open the evidence sheet on each handle row in turn. Each sheet shows the same lines, pins and Sources it showed before this PR, and no error appears.
2. Close the sheets and open five other Sources' pages, then return to the first Source's Promote page and open one evidence sheet again. It shows the same evidence.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A feature reads a graph, loads more than the cap of other graphs, then sets a memo on the first, and loses the memo | `SetEvidenceGraphMemo` keeps nothing for a graph that is not loaded, so the result is a recompute, not a wrong answer; one request loads at most 2 graphs (§4) |
| The cap thrashes on a real catalog with pages larger than the fixture | 64 is three fixture pages; the Landed note records any dogfood finding, and the constant is one line |
| `Drop` in a test resets the cap | `clear` does not touch `evidenceCap` (D-3); T-2 resets it with `defer` |
| A test that inspects memos changes eviction order | Documented on `EvidenceGraphMemo` (D-2); no test depends on order beyond T-1 |

**Stop and report if:**

- `TestLargeSourceBudgets` on `main` no longer ends with "births at one place warm", or `newLargeFixture`, the `evidenceRequests` closure or the "warm queries" subtest differ from §4.
- `graph.go` or `evidence_graph.go` on `main` reads or writes `evidenceGraphs` or `memos` anywhere §4 does not list (a reader the entry change would miss).
- A full Evidence pass over Source A in `TestLargeSourceBudgets` loads more than 64 graphs at the default cap, or "warm queries" fails with the cap in place.
- T-2 finds a response under cap 8 that differs from the default-cap response (a caller depends on a graph staying loaded; report the caller, do not raise the cap).
- `Verify` in the graphcache test package fails with "truth recompute is not registered".

## 12. Open questions

None.

## 13. PR

**Title:** S11-31 — Bound the loaded Evidence graphs
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-31-evidence-graph-cap.md`.
