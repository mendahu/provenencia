# S11-31 — Bound the Source cache

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 7 — Performance
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: an engine cache gains a cap; nothing on screen changes)
**Board:** —
**Depends on:** S11-19 (landed [#513](https://github.com/mendahu/provenencia/pull/513)), S11-21 (open as [#536](https://github.com/mendahu/provenencia/pull/536); this plan is written against its head, user 2026-10-24)
**Blocks:** —
**Branch:** `s11-31-source-cache-cap`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** — (no catalog, FFI, matching-rule or UI change; `graphcache` has no skill of its own)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) "Loads read the catalog's in-memory graph"; S11-19 plan D-21 and §14 (Evidence graph memos); S11-21 plan §6 "Next adds" (this PR adds a subtest to `TestLargeSourceBudgets`); working rule "High tolerance for churn"; dogfood bar 3 ("Fast, and it explains itself")
**Planned:** 2026-10-24 against S11-21's branch `s11-21-large-source-performance` @ `74f64d0` ([#536](https://github.com/mendahu/provenencia/pull/536) head; `main` @ `6152d4b`)
**Review:** — (not reviewed yet)

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

**What stays.** Commit drops (`dropEvidenceGraphs`), `Drop`, `dropDerived`, the memo API, the loaders, and every caller of `EvidenceGraph` are unchanged. The fixture's warm Evidence pass still issues no query: its 21 graphs fit under the cap.

## 2. User stories

**US-1.** As a researcher in a long session, when I open evidence and Source pages across many Sources, the app's memory for loaded Evidence graphs stays bounded, so a large catalog does not slow the app over the session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Cap 2; graphs A, B, C are read in that order | B and C are loaded; A and A's memos are not |
| AC-1.2 | Cap 2; read A, read B, read A again through `EvidenceGraph`, read C | A and C are loaded; B is not |
| AC-1.3 | Cap 2; read A, read B, a hit on A's memo through `EvidenceGraphMemo`, read C | A and C are loaded; B is not |
| AC-1.4 | A was evicted; A is read again | the returned `EvidenceGraph` deep-equals A's first read; `EvidenceGraphMemo(A, k)` reports no memo until a feature sets it again |
| AC-1.5 | Cap 2; read A, B; `Drop`; read C, D | C and D are loaded (no stale use order from before the drop evicts either) |
| AC-1.6 | Cap 2; read A, B, C (A evicted); `Verify` | `Verify` returns nil and A is still not loaded afterwards |

**US-2.** As a researcher, when I open the evidence sheet on every row of a large page, the sheet shows the same evidence whether or not graphs were evicted along the way.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | S11-21's large fixture; Evidence on every handle row of Source A (21 graphs) with the cap set to 8 | at most 8 Evidence graphs are loaded after each request, and every response deep-equals the response with the default cap |
| AC-2.2 | S11-21's large fixture; the default cap | `TestLargeSourceBudgets` "warm queries" still sees 0 graph queries and 0 SELECTs for a warm Evidence pass |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` holds at most `maxEvidenceGraphs` loaded Evidence graphs and evicts the least recently used one with its memos (D-1, D-2, D-3) | STEP-1, AC-1.1 – AC-1.5 |
| A read through `EvidenceGraph` or a memo hit through `EvidenceGraphMemo` counts as use (D-2) | STEP-1, AC-1.2, AC-1.3 |
| One named constant, sized from S11-21's fixture (D-4) | STEP-1, AC-2.2 |
| `Verify` checks only graphs that are still loaded | STEP-1, AC-1.6 |
| A test-only cap override and loaded count (D-5) | STEP-1, STEP-2 |
| An evicted graph reloads with the same rows and its memo is rebuilt | STEP-1, STEP-2, AC-1.4, AC-2.1 |
| The hub-page check: Evidence on every row stays under the cap and gives the same evidence | STEP-2, AC-2.1 |
| Docs: domain doc, completed.md, map, this plan | STEP-3 |

| Out | Home |
| --- | --- |
| Evicting nodes (`g.nodes`), kind indexes, displays or details | not planned — the map's scope entry excludes it; nodes are bounded by the catalog |
| Evicting or bounding the match index | not planned — the map's scope entry excludes it; it is one value per catalog |
| A cap sized in bytes, or a cap the researcher sets | not planned — no size measure exists, and a count fits the observed working set (D-4) |
| Pinning graphs for the length of one request | not planned — one request loads at most 2 graphs (§4), far under the cap |
| Changing `dropEvidenceGraphs`, `dropDerived`, or what a commit drops | not planned |

## 4. Current state

Verified at S11-21's head `74f64d0`.

| Area | Where | Today |
| --- | --- | --- |
| Loaded graphs | `core/database/graphcache/graph.go` `Graph.evidenceGraphs` (`map[string]EvidenceGraph`), `Graph.memos` (`map[string]map[string]any`) | Keyed by graph id. Nothing evicts: `clear` (from `New` and `Drop`) empties both, `dropEvidenceGraphs` removes named ids, `dropDerived` empties `memos` only |
| Reads | `core/database/graphcache/evidence_graph.go` `Graph.EvidenceGraph` | Returns the map entry on a hit; on a miss calls the registered `evidenceGraphLoad`, stores the result, returns it. No use order |
| Commit drop | `evidence_graph.go` `Graph.dropEvidenceGraphs` | Deletes each named id from `evidenceGraphs` and `memos`. Called from `core/database/graphcache/commit.go` `OnCommit` with `set.EvidenceGraphs` |
| Memos | `evidence_graph.go` `Graph.EvidenceGraphMemo`, `Graph.SetEvidenceGraphMemo` | Read returns `memos[graph][key]`. Set stores only when the graph is in `evidenceGraphs`. The doc comment lists what clears a memo; eviction is not on it |
| Verify | `core/database/graphcache/verify.go` `Graph.Verify` | Ranges over `g.evidenceGraphs` and checks each stored graph with `evidenceGraphFn`. It already checks only loaded graphs and loads nothing; this PR adds a test, not code |
| Test hooks | `graph.go` `SetQueryCounterForTest`, `LoadedEvidenceGraphForTest`, `ResetMatchIndexForTest` | The pattern for a test-only setter and reader on `Graph` |
| Callers | `core/database/promotealign/layer.go` (memo first, then `EvidenceGraph` on a miss), `core/database/promotealign/canon.go` `loadCanon` (candidates memo), `core/database/promotealign/evidence.go` `memberValues`, `bridges`, `titles`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/pins.go`, `api/ffi/handlers/subjects.go`, `citations.go`, `observations.go`, `promote.go`, `event_titles.go` | Each takes the returned `EvidenceGraph` value and uses it within the call. `layer.go` and a warm `Propose` reach the row's graph only through `EvidenceGraphMemo`, so a memo hit is a use (D-2) |
| Working set | measured on S11-21's `newLargeFixture` at `74f64d0` with a throwaway test | 23 Evidence graphs in the catalog. A cold `Propose` on Source A loads 1. Evidence on all 62 handle rows of Source A loads 21. One cold Evidence request loads at most 2 |
| Budget test | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `newLargeFixture`, `evidenceRequests`, `pass` | Subtests share one fixture and run in order. "warm queries" requires a warm Evidence pass to issue 0 graph queries and 0 SELECTs, so every graph of the pass must stay loaded at the default cap |
| Memo tests | `core/database/graphcache/evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource`; fixture `newWalkFixture`, `walkFixture.newSource` in `core/database/graphcache/evidence_graph_walk_test.go` | Cover memo drop by commit, `Drop`, vocabulary and `SetAliasIndex`. `newSource` makes a Source; `evidencegraphstest.SoleGraph` gives its graph id |
| Verify tests | `core/database/graphcache/graph_test.go` (calls `g.Verify()` at lines 51 and 55) | The graphcache test package already has the truth recompute registered |
| Domain doc | `docs/promote-graph-alignment.md` bullet "Loads read the catalog's in-memory graph" | Says the graph keeps Evidence graphs between calls; says nothing of a bound |

**Precedent.** [#513](https://github.com/mendahu/provenencia/pull/513) (S11-19 D-21) added `memos` beside `evidenceGraphs` and made "a graph and its memos go together"; this PR keeps that pairing and adds one more map in step with them. Copy `SetQueryCounterForTest` for the test-only setter. There is no earlier eviction in Go; the Swift session cache's `.evict` (`docs/macos-client-patterns.md`) is a different layer and is not reused.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Use order is a stamp per loaded graph.** `Graph` gains `evidenceUsed map[string]uint64` *(new)*, `evidenceTick uint64` *(new)* and `evidenceCap int` *(new)*. A use sets `evidenceUsed[id] = ++evidenceTick`. Eviction scans `evidenceUsed` for the smallest stamp | At a cap of 64 a linear scan on each load is negligible beside the SQL load it follows, and a map in step with `evidenceGraphs` is the shape `memos` already has. Stamps strictly increase, so there are no ties and the order is deterministic | `container/list` LRU — pointer bookkeeping for an O(1) the cap does not need; storing the stamp inside the `evidenceGraphs` value — changes the map's value type and every reader (`Verify`, `LoadedEvidenceGraphForTest`) for no gain |
| D-2 | **What counts as use.** `EvidenceGraph` on a hit and after a load, and `EvidenceGraphMemo` when it returns a memo. `SetEvidenceGraphMemo` does not stamp (it is always called right after a read of the same graph) | `layer.go` and a warm `Propose` reach the row's graph only through its memo; without the memo stamp the graph Promote is working on would look idle and be evicted by a run of member loads | Stamping only in `EvidenceGraph` — evicts the open page's own graph during an evidence pass |
| D-3 | **One place forgets a graph.** `forgetEvidenceGraph(key string)` *(new, `evidence_graph.go`)* deletes the key from `evidenceGraphs`, `memos` and `evidenceUsed`. `dropEvidenceGraphs` calls it in place of its two `delete`s; eviction calls it. `clear` resets `evidenceUsed` to an empty map with the other two; `evidenceTick` is not reset. `dropDerived` stays as it is (it drops memos, not graphs) | The three maps cannot drift: every path that removes a graph goes through one function or through `clear` | Deleting the stamp separately in each path — a fourth path added later forgets it |
| D-4 | **The cap is `maxEvidenceGraphs = 64`** *(new, unexported const, `evidence_graph.go`)*. `New` sets `g.evidenceCap = maxEvidenceGraphs`. After `EvidenceGraph` stores a newly loaded graph and stamps it, `evictEvidenceGraphs()` *(new)* forgets the smallest-stamped graph while `len(g.evidenceGraphs) > g.evidenceCap` | S11-21's fixture loads 21 graphs for a full Evidence pass over a 100-Subject Source; 64 keeps about three such pages warm, so "warm queries" stays at 0 queries, and one graph is one Source page's Subjects and Observations, so 64 is a few megabytes. The just-loaded graph has the largest stamp and is never the one evicted | 32 — under two full pages; 256 — no measured need, and four times the memory; no cap with a byte budget — no size measure exists |
| D-5 | **Test hooks.** `SetEvidenceGraphCapForTest(n int)` *(new, `graph.go`)*: `n > 0` sets the cap and evicts down to it at once; `n <= 0` restores `maxEvidenceGraphs`. `LoadedEvidenceGraphCountForTest() int` *(new, `graph.go`)* returns `len(g.evidenceGraphs)` | The graphcache tests need a cap of 2 and the budget test a cap of 8, and both need to count loaded graphs; `LoadedEvidenceGraphForTest` needs ids the perf test does not hold | Exporting the constant and a production setter — no caller outside tests needs either |
| D-6 | **The budget subtest runs last** in `TestLargeSourceBudgets`, as `t.Run("evidence under a small cap", …)` *(new)* after "births at one place warm". It builds `reqs` with `evidenceRequests()`, records each response at the default cap, then sets the cap to 8 with `defer f.c.Graph().SetEvidenceGraphCapForTest(0)`, calls `f.c.Graph().Drop()`, and runs every request in order, checking after each that `LoadedEvidenceGraphCountForTest() <= 8` and that the response deep-equals the recorded one (`reflect.DeepEqual`) | Last, so a small cap cannot change the timings or the 0-query check of earlier subtests. 8 is under the 21 the pass touches, so eviction runs, and over the 2 one request needs | A new test with its own `newLargeFixture` — the fixture costs seconds; asserting timing under the small cap — not a budget this PR owns |
| D-7 | **`Verify` is unchanged.** It already ranges over `g.evidenceGraphs` only; AC-1.6 pins that it passes after an eviction and does not load the evicted graph | The scope line is already true; a test keeps it true | Adding code to `Verify` — nothing to add |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `maxEvidenceGraphs` *(new)* | const | `alignment.TopK` and other `matchregistry` knobs | A cache bound is not a matching knob; `matchregistry` is for comparison layers (working rule "One match registry") and `graphcache` does not import it |
| `Graph.evidenceUsed`, `evidenceTick`, `evidenceCap` *(new)* | fields | `Graph.memos` | `memos` holds feature values, not use order; a parallel map is the same shape (D-1) |
| `forgetEvidenceGraph` *(new)* | func | `dropEvidenceGraphs` | Extended: `dropEvidenceGraphs` now calls it per id; eviction needs the single-id form |
| `evictEvidenceGraphs` *(new)* | func | — | Searched `graphcache` and `core/` for `evict`, `lru`, `cap`: no existing eviction in Go |
| `SetEvidenceGraphCapForTest` *(new)* | test hook | `SetQueryCounterForTest`, `ResetMatchIndexForTest` | Same pattern, new knob |
| `LoadedEvidenceGraphCountForTest` *(new)* | test hook | `LoadedEvidenceGraphForTest` | That takes an id; the budget subtest needs a count |
| `core/database/graphcache/evidence_graph_cap_test.go` *(new)* `TestEvidenceGraphCap` | test | `evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource` | Same fixture (`newWalkFixture`); a separate file because it is a separate behavior and needs three Sources |

**Builds on.** S11-19's pairing of a graph and its memos (#513); S11-21's large fixture and `TestLargeSourceBudgets`; the `…ForTest` hooks on `Graph`.
**Deletes.** The two `delete` lines in `dropEvidenceGraphs`, replaced by `forgetEvidenceGraph`. Nothing else becomes dead.
**Next adds.** A cap on another per-key cache in `graphcache` (displays, details) reuses the stamp shape. A later need to size the cap from memory changes one constant.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` (`graph.go`) | `Graph` gains `evidenceUsed map[string]uint64`, `evidenceTick uint64`, `evidenceCap int` *(new)*, commented next to `memos`. `New` sets `evidenceCap = maxEvidenceGraphs`. `clear` sets `evidenceUsed = map[string]uint64{}`. `SetEvidenceGraphCapForTest`, `LoadedEvidenceGraphCountForTest` *(new)* beside `LoadedEvidenceGraphForTest` |
| `core/database/graphcache` (`evidence_graph.go`) | `const maxEvidenceGraphs = 64` *(new)* with a comment naming S11-21's 21-graph pass. `useEvidenceGraph(key string)` *(new)*: stamps. `forgetEvidenceGraph(key string)` *(new)*. `evictEvidenceGraphs()` *(new)*. `EvidenceGraph`: stamp on a hit; after storing a load, stamp, then evict. `EvidenceGraphMemo`: stamp when it returns `ok`. `dropEvidenceGraphs`: calls `forgetEvidenceGraph`. `EvidenceGraphMemo`'s doc comment adds "or the graph is evicted as least recently used" to what clears a memo |

```go
// maxEvidenceGraphs is how many Evidence graphs stay loaded. …
const maxEvidenceGraphs = 64

func (g *Graph) useEvidenceGraph(key string)
func (g *Graph) forgetEvidenceGraph(key string)
func (g *Graph) evictEvidenceGraphs()
func (g *Graph) SetEvidenceGraphCapForTest(n int)
func (g *Graph) LoadedEvidenceGraphCountForTest() int
```

`Over` graphs go through `New`, so they get the default cap; nothing else changes for them. No new error.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Cap the loaded Evidence graphs**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/evidence_graph.go`, `core/database/graphcache/evidence_graph_cap_test.go` *(new)*. Skill: —.
Do: D-1 – D-5 and D-7 as in §6 Go core. Write `TestEvidenceGraphCap` (T-1) table-driven per `.cursor/rules/go-tests.mdc`: each case drops the graph, sets the cap with `g.SetEvidenceGraphCapForTest(2)` (deferred reset to 0), and runs a list of operations (`read` id, `memo hit` id, `set memo` id, `drop`, `verify`) over three Sources A, B, C (and D for AC-1.5) made with `walkFixture.newSource`, then checks `LoadedEvidenceGraphForTest` for each id. Keep `TestEvidenceGraphMemoDroppedWithSource` passing unchanged.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/` passes; `TestEvidenceGraphCap` covers AC-1.1 – AC-1.6.

**STEP-2 — Hub-page check under a small cap**
Files: `core/database/promotealign/perf_test.go`. Skill: —.
Do: D-6: add the "evidence under a small cap" subtest (T-2) at the end of `TestLargeSourceBudgets`. Use `evidenceRequests`, `f.evidence`, `reflect` (already imported).
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign/` passes, including "warm queries" (AC-2.2) and the new subtest (AC-2.1); then `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-3 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/completed.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (map row), this plan (status, Landed note).
Do: in `docs/promote-graph-alignment.md`, the "Loads read the catalog's in-memory graph" bullet adds one sentence: the graph keeps at most 64 Evidence graphs and drops the least recently used one with its memos; a dropped graph loads again on its next read. Then per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-31-source-cache-cap.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1a | Go | `core/database/graphcache/evidence_graph_cap_test.go` `TestEvidenceGraphCap` *(new)* "oldest is evicted" | cap 2; read A, B, C → B, C loaded; A not; `EvidenceGraphMemo(A, "k")` set before eviction is gone | AC-1.1, D-3, D-4 |
| T-1b | Go | same, "a read is a use" | cap 2; read A, B, A, C → A, C loaded; B not | AC-1.2, D-2 |
| T-1c | Go | same, "a memo hit is a use" | cap 2; read A, set memo A, read B, memo hit A, read C → A, C loaded; B not | AC-1.3, D-2 |
| T-1d | Go | same, "an evicted graph reloads the same" | read A (keep the value), read B, C (A evicted), read A → deep-equals the kept value; `EvidenceGraphMemo(A, "k")` not ok | AC-1.4 |
| T-1e | Go | same, "a drop clears the use order" | cap 2; read A, B; `Drop`; read C, D → C, D loaded | AC-1.5, D-3 |
| T-1f | Go | same, "verify skips an evicted graph" | cap 2; read A, B, C; `Verify()` → nil; A still not loaded | AC-1.6, D-7 |
| T-1g | Go | same, "a commit drop forgets the stamp" | cap 2; read A, B; a commit naming A (`OnCommit` with `effects.Set{EvidenceGraphs: [A]}`, as `TestEvidenceGraphMemoDroppedWithSource` "a graph drop" does); read C, D → C, D loaded; B evicted by D, not C | D-3 |
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
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
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
| A feature reads a graph, loads more than the cap of other graphs, then sets a memo on the first, and loses the memo | `SetEvidenceGraphMemo` already keeps nothing for a graph that is not loaded, so the result is a recompute, not a wrong answer; one request loads at most 2 graphs (§4) |
| The cap thrashes on a real catalog with pages larger than the fixture | 64 is three fixture pages; the Landed note records any dogfood finding, and the constant is one line |
| `Drop` in a test resets the cap | `clear` does not touch `evidenceCap` (D-3, D-5); T-2 resets it with `defer` |

**Stop and report if:**

- [#536](https://github.com/mendahu/provenencia/pull/536) has not merged into `main` when you start, or merged with `newLargeFixture`, `evidenceRequests`, `pass` or the "warm queries" subtest different from §4.
- `graph.go` or `evidence_graph.go` on `main` has a path that removes a graph from `evidenceGraphs` other than `clear` and `dropEvidenceGraphs`.
- A full Evidence pass over Source A in `TestLargeSourceBudgets` loads more than 64 graphs at the default cap, or "warm queries" fails with the cap in place.
- T-2 finds a response under cap 8 that differs from the default-cap response (a caller depends on a graph staying loaded; report the caller, do not raise the cap).
- `Verify` in the graphcache test package fails with "truth recompute is not registered".

## 12. Open questions

None.

## 13. PR

**Title:** S11-31 — Bound the Source cache
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-31-source-cache-cap.md`.
