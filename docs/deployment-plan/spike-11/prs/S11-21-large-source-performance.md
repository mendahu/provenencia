# S11-21 — Large-Source performance check

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 7 — Performance
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: engine timing, caching and one test; nothing on screen changes)
**Board:** —
**Depends on:** S11-16 (landed [#501](https://github.com/mendahu/provenencia/pull/501)), S11-19 (landed [#513](https://github.com/mendahu/provenencia/pull/513))
**Blocks:** S11-31 (sizes its cap from this PR's fixture)
**Branch:** `s11-21-large-source-performance`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (graphalign and segmentcompare change; no new knob), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the open-time warm in `catalogsession`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/catalog-session.mdc`
**Context:** [Layer 6 procedure](../comparison-layers.md#layer-6-procedure) steps 2–6; [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) "Thresholds and bands"; [`docs/matching-model.md`](../../../matching-model.md) knob inventory (`alignment.TopK`, `alignment.TieCap`); ship logs S11-13 PRF-1, S11-15 PRF-1 and PRF-3, S11-16 PRF-1 and PRF-2, S11-19 round 3 CON-1; working rules "Registry numbers: carry over, then estimate", "One match registry", "High tolerance for churn"; dogfood bar 3 ("Fast, and it explains itself")
**Planned:** 2026-10-24 against `main` @ `c93b3fb`
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

A Source with 100 Subjects proposes without a noticeable wait, re-proposes after a row decision at once, and a page of births at a big city no longer freezes Promote. A budgeted Go test on a realistic catalog holds Promote to those times in every `scripts/check-pr.sh` run, so a later change that makes Promote slow fails a test instead of a dogfood session.

**Why now.** Each layer PR since S11-13 logged a performance follow-up to this PR (map scope entry). Measured while planning, on the fixture this plan specifies (4-core Linux, the class of CI's `ubuntu-latest`), `main` @ `c93b3fb`:

| Run | `main` | Prototype of this plan |
| --- | --- | --- |
| 100-Subject Source, cold (graph dropped) | 1.24 s | 1.3 s |
| 100-Subject Source, warm repropose | 450 ms | 105 ms |
| 100-Subject Source, first proposal after a handle commit | 650 ms | 680 ms |
| Hub page (30 births at a 6,000-event city), cold | 18.6 s | 2.9 s |
| Hub page, warm | 16 s | 1.6 s |
| Evidence on all 58 handle rows, warm | 20 ms, 0 graph queries | same |
| Warm Evidence / Propose, raw SELECTs | 7 / 8 | 2 / 3 (`SoleGraph`'s 2, plus `MAX(revision)`) |

The hub page spends its time scoring every one of the city's 6,000 events for each of the 30 births, several times (S11-16 PRF-1). The warm repropose spends two thirds of its time in `graphalign.Candidates` re-running subject compare on about 61,000 shortlisted pairs (S11-15 PRF-3: a common given name keys 10% of the catalog), which no decision changes. The user set the budgets and the remedies (2026-10-24): §5 D-1 and D-2.

**What stays.** Every proposal: the same rows, bands, scores and reasons as `main` on every existing test and golden (the prototype passed `core/graphalign`, `core/database/promotealign`, `api/...` unchanged). The shortlist index and profiles still drop on every handle commit. `Propose` still reads `MAX(revision)`. No registry number changes.

## 2. User stories

**US-1.** As a researcher, when I open Promote on a Source with 100 Subjects, it proposes within 2.5 s cold, re-proposes within 250 ms after I decide a row, and within 1.5 s after a Done or any other handle commit, so that Promote keeps up with me.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The large fixture (§6 Go core, D-4); `Propose` on Source A after `Graph().Drop()` | Takes at most 2.5 s |
| AC-1.2 | Source A proposed once; `Propose` again with no commit between (fastest of 3) | Takes at most 250 ms |
| AC-1.3 | A person on another Source promoted (a handle commit); then `Propose` on Source A (fastest of 2) | Takes at most 1.5 s |
| AC-1.4 | Source A warm; the same `Propose` after `Graph().Drop()` | Returns a `Proposal` equal (`reflect.DeepEqual`) to the warm one |

**US-2.** As a researcher, when I open Promote on a page whose births are all at one big city, it proposes within 5 s cold and 3 s warm, so that a hub page is slow but usable instead of frozen.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The large fixture's hub page (30 births located at the hub place, which has 6,000 filed events); `Propose` after `Graph().Drop()` | Takes at most 5 s |
| AC-2.2 | Hub page proposed once; `Propose` again (fastest of 3) | Takes at most 3 s |
| AC-2.3 | A placed pair whose handle has 30 corresponding canon neighbors of the layer neighbor's kind, distinct similarities, `TopK` 5 | The layer neighbor gets exactly 5 nominated candidates: the 5 most similar |
| AC-2.4 | Same, all 30 tied on similarity, `TieCap` 20 | 20 nominated candidates |
| AC-2.5 | Same, one corresponding canon neighbor whose similarity is under `MinSimilarity` | It is still nominated |
| AC-2.6 | Same, and one of the 30 is already a candidate of the neighbor from `Canon.Candidates` | That one stays a candidate and the bound still adds 5 others |

**US-3.** As a researcher, when I open the evidence sheet on rows of a large Source a second time, it opens from memory, so that paging through rows costs nothing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Source A's proposal; `Evidence` for every row with a handle target, then again (fastest of 3 passes) | The second pass over all rows takes at most 100 ms |
| AC-3.2 | Same warm pass under `Graph().SetQueryCounterForTest` | 0 graph queries |
| AC-3.3 | Same warm pass under `databasetest.CountSelects` (D-10) | As many SELECTs as `evidencegraphs.SoleGraph` alone issues, times the number of rows (S12-09 deletes `SoleGraph`; the count then is 0) |
| AC-3.4 | A warm `Propose` on Source A under `databasetest.CountSelects` | `SoleGraph`'s SELECTs plus 1 (`MAX(revision)`) |
| AC-3.5 | A catalog opened through `catalogsession.Do` | `valuerules.Load` issues 0 SELECTs |

**US-4.** As a contributor, the memo this PR adds never serves a stale answer, so that a fast proposal is a correct one.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A Source proposed once, then nothing committed | The warm proposal equals the proposal after `Graph().Drop()` |
| AC-4.2 | A Source proposed once, then another "John Smith" with the same birth filed and promoted on a new Source (a handle commit) | The next proposal equals the proposal after `Graph().Drop()`, and lists the new handle among the row's alternatives |
| AC-4.3 | A Source proposed once, then a name observation added to one of its persons (a commit naming the Source's Evidence graph) | The next proposal equals the proposal after `Graph().Drop()` |

## 3. Scope

| In | Covered by |
| --- | --- |
| A realistic large catalog for tests: `bulkCanon` takes a catalog shape; a large shape with Zipf names, dated births and deaths (some BEF / AFT / undated), one hub city of 6,000 events and 50 towns | STEP-1, D-4 |
| A budgeted `Test…` (not a `Benchmark…`) with the D-1 budgets, logging each measured time | STEP-1, STEP-4, STEP-5, STEP-6, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| A test helper that counts the raw SELECTs a catalog prepares | STEP-1, D-10 |
| `segmentcompare.CompareNeighbor`: segment compare with the neighbor result handed in, computed only when the placement needs it | STEP-2, D-5 |
| A per-`Align` memo of subject compare per (layer Subject, canon handle) | STEP-3, D-6 |
| Bounded nominations: a placed pair offers each layer neighbor at most `TopK` plus ties, capped at `TieCap`, ranked as `Candidates` ranks, with no similarity gate | STEP-4, D-7, AC-2.3, AC-2.4, AC-2.5, AC-2.6 |
| A cross-proposal memo of each Subject's candidates, held on the Source's Evidence graph memo and valid for one shortlist index | STEP-5, D-8, AC-1.2, AC-1.4, AC-4.1, AC-4.2, AC-4.3 |
| The catalog's default name format warmed onto the graph at project open, like the alias tables | STEP-6, D-9, AC-3.3, AC-3.4, AC-3.5 |
| Docs: nominations are bounded; the knob inventory; the map's scope entry says which remedies shipped and which were not needed | STEP-7 |

| Out | Home |
| --- | --- |
| Updating the shortlist index in place after a handle commit (`Index.Remove` plus `Add`) | not planned — measured inside budget: 680 ms against 1.5 s (AC-1.3) |
| Scoring shortlisted handles on key values first and profiling only the top few | not planned — the cold proposal is inside budget: 1.3 s against 2.5 s (AC-1.1) |
| A memo of neighbor compares that outlives one `Align` (hub warm 1.6 s → 0.6 s, about 20 MB per hub Source) | not planned (user, 2026-10-24) |
| An unplaced neighbor's segment reading only that neighbor's candidates | not planned — measured: hub warm only 1.9 → 1.7 s, and it changed results (strong rows 51 → 43) |
| The canon walk reaching the whole connected component on every proposal (12,051 handles on the fixture) | not planned — inside the warm budget at this fixture; a larger fixture that fails the budget reopens it |
| A commit effect for name-format writes | not planned — the only writer is `nameformats.Install` at onboarding; D-9 follows the alias tables instead |
| Removing `SoleGraph`'s SELECTs from warm Evidence | S12-09 (deletes `SoleGraph`); AC-3.3's expected count goes to 0 there |
| Bounding the loaded Source graphs and their memos | S11-31 |
| Making `segmentcompare.Result.Neighbor` smaller | not planned — no reader outside `segmentcompare`, and no budget needs it |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Proposal entry | `core/database/promotealign/promotealign.go` `propose` | Reads `MAX(revision)`, `loadInputs`, `mergeFixed`, `newProfiler`, `loadCanon(c.Graph(), prof, layer, anchors, cfg, in.terms)`, then `graphalign.Align(layer, canon, anchors, cfg)` |
| Inputs | `promotealign.go` `loadInputs` | `evidencegraphs.SoleGraph` (2 SELECTs), `valuerules.Load(c.Graph(), q)` (5 SELECTs on a catalog with the default name format installed), `loadTerms`, `loadLayerMemo`. Shared by `Propose` and `Evidence` |
| Candidates | `core/database/promotealign/canon.go` `loadCanon` | For the unfixed weighted Subjects: `loadShortlist`, `seedHandles` (one node load, one profile batch), then `graphalign.Candidates(s.Kind, s.Values, shortlisted[i], layer.Metas, cfg)` per Subject, on every proposal. On the fixture: 61,306 shortlisted pairs for Source A, 61% of a warm proposal's CPU |
| Canon walk | `canon.go` `loadCanon` BFS | Walks every link from the seeds until no new handle: 12,051 handles on the fixture. Out of scope (§3) |
| Layer memo | `core/database/promotealign/layer.go` `loadLayerMemo`, `layerMemoKey` | Precedent for an Evidence graph memo: `g.EvidenceGraphMemo(graphID, layerMemoKey+namePattern)`, stored with `g.SetEvidenceGraphMemo` |
| Match index | `core/database/promotealign/index.go` `matchIndex`, `loadShortlist`, `matchIndexOf` | The shortlist index and the profiles, in the graph's match index slot. Dropped by a handle commit, a vocabulary commit, `SetAliasTables` and `Drop` (`core/database/graphcache/graph.go` `dropDerived`, `commit.go` `OnCommit`) |
| Evidence graph memos | `core/database/graphcache/evidence_graph.go` `EvidenceGraphMemo`, `SetEvidenceGraphMemo` | Held while the Evidence graph is loaded; dropped by a commit naming the graph, `Drop`, a vocabulary commit, `SetAliasTables` (all through `dropDerived` or `dropEvidenceGraphs`) |
| Candidate ranking | `core/graphalign/candidates.go` `Candidates` | Compares each handle (`subjectcompare.Compare`), drops similarity under `MinSimilarity`, sorts by similarity, evidence, `Ref`, `ID`, keeps `TopK`, extends ties at the cutoff, caps at `TieCap` |
| Placing | `core/graphalign/align.go` `(*state).place` | For each layer link, every canon neighbor of the placed handle through a corresponding link, of the layer neighbor's kind, becomes that neighbor's candidate (`st.candidate`), scored at once; no bound (S11-16 D-7). On the hub page: 6,000 per birth |
| Scoring | `align.go` `score`, `evaluate`, `segments`, `compareSegments`, `compare` | `score` calls `evaluate` (subject compare of the pair) and `segments`; `segments` calls `compareSegments` → `compare` → `segmentcompare.Compare` for each corresponding canon link, which runs subject compare on the neighbors. Nothing is memoized: the fixed hub place's own score compares its 30 births with 6,000 events at every re-score |
| Duplicates and alternatives | `core/graphalign/rows.go` (line 155 `st.evaluate`, line 200 `st.compare`) | Compare layer Subject with layer Subject; unchanged |
| Segment compare | `core/segmentcompare/segmentcompare.go` `Compare` | Computes `neighbor := subjectcompare.Compare(…)` unless the keys do not correspond or `placed == Apart`. Callers: `graphalign` `compare`, `core/claimevidence/build.go` |
| Value rules | `core/database/valuerules/valuerules.go` `Load` | `nameformats.Load(q)` (1 SELECT, plus 4 through `nameformats.LoadID` when a default format is stored), then `namealiases.Rules(g, q, …)`, which reads the graph's warm alias tables when present |
| Alias warm precedent | `core/database/namealiases/namealiases.go` `Warm`, `Rules`; `core/database/graphcache/graph.go` `AliasTables`, `SetAliasTables`; `core/catalogsession/session.go` `openResearcher` | Loaded once at project open onto the graph; `Drop` keeps them; `SetAliasTables` calls `dropDerived`; readers fall back to the catalog when not warm |
| Name formats | `core/database/nameformats/nameformats.go` `Load`, `Loaded`; `core/database/nameformats/registry.go` `Install` | `Install` (called only from `core/onboarding/ready.go`) is the only writer; the tables are `None` effects in `core/database/effects/registry.go` |
| Benchmarks | `core/database/promotealign/bench_test.go` `BenchmarkProposeColdHub`, `BenchmarkProposeWarmHub`, `BenchmarkProposeAfterCommit`, `bulkCanon`, `benchName`, `benchID` | 20k persons named `GivenN SurnameN`, 30k undated births round-robin over 5 hubs; reported, never run by `scripts/check-pr.sh` |
| Fixture | `core/database/promotealign/promotealign_test.go` `newFixture`, `person`, `participation`, `locate`, `place`, `promote`, `newSource`, `datedOn`; `evidence_test.go` `filedBirth`, `(*fixture).evidence` | The builders this PR's fixture composes |
| Query counting | `core/database/graphcache/graph.go` `SetQueryCounterForTest`; `memo_test.go` `TestEvidenceWarmIssuesNoQuery`, `TestProposeWarmIssuesNoQuery` | Counts only the graph's own queries; nothing counts raw catalog SELECTs |
| One connection | `core/database/catalog.go` (`db.SetMaxOpenConns(1)`) | Every catalog query runs on one SQLite connection, which D-10 relies on |
| Graph tests | `core/database/graphcache/graph_test.go` `TestDropLeavesAliasTables`, `TestMatchIndexDrops` | The precedent tests D-9's slot mirrors |
| Session tests | `core/catalogsession/alias_cache_test.go` `TestDoKeepsAliasTablesForTheOpenProject` | Opens a project through `onboarding.Complete` and `catalogsession.Do` |
| Graphalign tests | `core/graphalign/align_test.go` `alignWith`, `partLink`, `locLink`, `edge`, `bridge`, `vals`, `termProp`, `dateProp`, `eventMetas`; `invariants_test.go` | Every golden and invariant this PR must keep green |
| Docs | `docs/promote-graph-alignment.md` "Thresholds and bands" bullet; `docs/deployment-plan/spike-11/comparison-layers.md` "Built in S11-16" paragraph; `docs/matching-model.md` rows `alignment.TopK`, `alignment.TieCap` | Say nominations have no bound and `TopK` / `TieCap` size only the shortlisted candidates |

S11-27 is in progress on `s11-27-connection-subrows` and changes only Swift and docs (`deployment-plan.md`, `completed.md`); this PR's Go files do not overlap it. Rebase the docs step on whatever has landed.

**Precedent.** S11-19 D-21 / D-22 ([#513](https://github.com/mendahu/provenencia/pull/513)): the layer and member memos on the Evidence graph memo, and profiles shared through the match index. Copy the memo shape of `loadLayerMemo` (key constant with the name pattern after it, a read-only value, rebuild and store on a miss) for D-8. `namealiases.Warm` and `TestDropLeavesAliasTables` for D-9. `core/locator/validate_test.go` `TestValidateDocCapBeforeCrossing` is the repo's one wall-clock budget in a `Test…`: `time.Now()` / `time.Since`, `t.Fatalf` over the budget.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Budgets (user, 2026-10-24).** Source A cold ≤ 2.5 s; warm ≤ 250 ms; first proposal after a handle commit ≤ 1.5 s. Hub page cold ≤ 5 s; warm ≤ 3 s. Evidence on every handle row of Source A, warm, ≤ 100 ms for the pass. Warm repeats take the fastest of 3 runs, after-commit the fastest of 2, cold runs one run. Each subtest logs its measured time with `t.Logf` | 1.7–3× headroom over the prototype on a CI-class machine, so the test does not flake and still catches a 2× regression | Tighter budgets (warm 150 ms, hub 1 s) — need the cross-run neighbor memo the user declined; a `Benchmark…` — `scripts/check-pr.sh` never runs it |
| D-2 | **Remedies (user, 2026-10-24).** Ship D-5 – D-9. Do not ship the Out lines of §3 | Measured: these meet D-1; the others were inside budget, declined, or changed results | — |
| D-3 | **One test, one fixture, subtests in order.** `TestLargeSourceBudgets` *(new)* builds the large fixture once and runs subtests in this order: `cold`, `warm`, `memo matches a fresh run`, `after a handle commit`, `evidence warm`, `warm queries`, `hub page cold`, `hub page warm`. Budgets are named constants at the top of `perf_test.go` | Fixture build is about 2 s; one build keeps `go test` of the package under 25 s more | One test per budget — rebuilds the fixture eight times |
| D-4 | **The large shape.** `bulkCanon(db, hub, shape)` takes `catalogShape` *(new)*: `Persons`, `Hubs`, `HubEvents`, `Towns`, `TownEvents`, `Name func(r *rand.Rand, i int) string`, `Dates bool`. Event `i` of the hub events is the birth of person `i` when `i < Persons`, else the death of person `i % Persons`; hub events go round-robin over the hubs; town events are deaths of a random person at a random town; each town has a `toponym` value. With `Dates`, a person's birth year is 1780 plus `[0, 120)`, a death is 20 plus `[0, 60)` years later, and each event date is BEF the year with chance 1/20, AFT with 1/20, absent with 1/20, else a full day (month `[1, 12]`, day `[1, 28]`). `benchShape` = 20,000 persons, 5 hubs, 30,000 hub events, `benchName`, no dates (the benchmarks keep their catalog except that events past 20,000 become deaths). `largeShape` = 4,000 persons, 1 hub, 6,000 hub events, 50 towns, 2,000 town events, Zipf names (given names: 200 words, exponent 0.9; surnames: 3,000 words, exponent 0.7; words are `Giv` / `Sur` plus a base-26 letter suffix of the rank), dates on. All randomness is `rand.New(rand.NewPCG(1, 2))` from `math/rand/v2` | The scope asks to reuse `bulkCanon`; one builder with a shape, not two near-copies. Zipf gives a top given name about 13% and a top surname about 3% (S11-15 PRF-3's common-word case); BEF / AFT reach the `wide` key | A second builder beside `bulkCanon` — a near-copy; real name lists — no gain for a timing test |
| D-5 | **Segment compare takes the neighbor lazily.** `segmentcompare.CompareNeighbor(a, b Segment, placed Placement, own float64, neighbor func() subjectcompare.Result, rules Rules) Result` *(new)* holds `Compare`'s body; it calls `neighbor()` only after the correspondence and `Apart` checks. `Compare` keeps its signature and calls `CompareNeighbor` with a closure over `subjectcompare.Compare(a.NeighborKind, a.Neighbor, b.Neighbor, metas, subject, value)` | `graphalign` hands in its memoized result; `claimevidence` and the tests keep calling `Compare` | Changing `Compare`'s signature — touches `claimevidence` for nothing; passing a computed result — computes it for `Apart` and non-corresponding links |
| D-6 | **Pair memo within one `Align`.** `state` gains `pairs map[[2]string]subjectcompare.Result` *(new)*, keyed by (layer Subject id, handle id), made in `newState`. `(*state).pair(s *Subject, h *Handle) subjectcompare.Result` *(new)* returns the stored result or `st.evaluate(s.Kind, s.Values, h.Values)`, stored. `score` calls `st.pair(s, h)` instead of `st.evaluate`. `compareSegments` calls `segmentcompare.CompareNeighbor(…, func() subjectcompare.Result { return st.pair(ns, nh) }, st.cfg.Segment)` and no longer calls `st.compare`. `rows.go`'s `evaluate` and `compare` calls stay (they compare two layer Subjects) | Values do not change during one `Align`; the same pair is compared by its own score, by every neighbor's segments, and by D-7's ranking | Memoizing inside `subjectcompare` — it is pure and keyless |
| D-7 | **Bounded nominations.** In `place`, for each layer link, the corresponding canon neighbors of the placed handle of the neighbor's kind split into those already in `st.cands[nk]` (kept, `via` set as today) and fresh ones. Fresh ones are ranked by `st.pair(ns, h)` and cut by the same rule as `Candidates` with no `MinSimilarity` gate: similarity high first, then evidence, `Ref`, `ID`; keep `TopK`; extend ties at the cutoff; cap at `TieCap`. Only the kept ones become candidates with `via`. The sort-and-cut moves out of `Candidates` into `keepTop(found []scored, r Rules) []Handle` *(new, `candidates.go`)*, with `scored` at package level; `Candidates` gates on `MinSimilarity` then calls it | The scope's remedy for S11-16 PRF-1, and one ranking rule for shortlisted and nominated candidates. The prototype: hub warm 16 s → 1.9 s with D-6, every existing test green | A bound with its own knob — a second size rule beside `TopK` / `TieCap`; gating nominations on `MinSimilarity` — the scope keeps "a handle reached through a link is still offered at any score" |
| D-8 | **Candidates memo across proposals.** `candidatesMemoKey = "promotealign.candidates\x00"` *(new, `canon.go`)*, plus the name pattern, on the Source's Evidence graph memo. Value `candidatesMemo{index *shortlist.Index; bySubject map[string][][]byte}` *(new)*: each Subject's `Candidates` handle ids, in order, and the shortlist index they were ranked over. `loadCanon` gains a `graphID []byte` parameter (after `g`); after `loadShortlist` it reads the memo, treats it as empty when `memo.index != ix`, runs `seedHandles` and `Candidates` only for the `todo` Subjects with no entry, then stores a **new** map holding the old entries and the new ones (a stored memo is read-only). Every candidate id still goes through `add` as today | Candidates depend on the Subject's values (dropped with the Evidence graph memo), the handles' profiles and the index (dropped together on every handle commit), the vocabulary and aliases (`dropDerived` drops both) and the name pattern (in the key). Comparing the index pointer ties the memo to the one index it was ranked against. Prototype: warm 450 → 105 ms, identical proposals | Keying on the match index alone — misses a Source edit; caching the whole `Canon` — fixed rows change it on every decision |
| D-9 | **Warm the name format at open.** `graphcache.Graph` gains `nameFormat any`, `nameFormatOK bool`, `NameFormat() (any, bool)` and `SetNameFormat(v any)` *(new, `graph.go`)*, beside the alias tables: `SetNameFormat` calls `dropDerived`; `clear` (so `Drop`) leaves the slot. `valuerules.Warm(c *database.Catalog) error` *(new)* stores `nameformats.Load(db)`'s `Loaded` with `SetNameFormat`. `valuerules.Load` uses the warm value when `g.NameFormat()` holds a `nameformats.Loaded`, else reads `nameformats.Load(q)` as today. `catalogsession.openResearcher` calls `valuerules.Warm(c)` right after `namealiases.Warm(c)` | Same lifetime and owner shape as the alias tables (the S11-19 round 3 CON-1 fix the scope names): loaded once per open, read by every compare. `Install` runs at onboarding before a researcher open, so no write drops it | A commit effect on `project_settings` and the `name_format_*` tables — no write path goes through `writes.Run`, and `project_settings` holds unrelated keys; moving the alias slot into a generic map — churn in `namealiases` for one more value |
| D-10 | **Counting raw SELECTs.** `core/database/databasetest` *(new package)*, file `selects.go` *(new, `//go:build cgo`)*: `CountSelects(t testing.TB, c *database.Catalog) *Selects` takes the catalog's one connection (`db.Conn`, `Raw`) and registers a `sqlite3` authorizer that counts `sqlite3.SQLITE_SELECT` and returns `sqlite3.SQLITE_OK`; `(*Selects).N() int` and `(*Selects).Reset()`. The doc comment says it relies on `SetMaxOpenConns(1)` in `core/database/catalog.go`. AC-3.3 and AC-3.4 compare against `SoleGraph`'s own count measured under the same counter in the same test, not a literal | Counts what SQLite prepares, whoever issues it; measuring `SoleGraph` in place keeps the assertion true when its SQL changes | A counting driver — needs a second driver name and an open option on `database.Create`; counting through `graphcache` — misses the raw queries this AC is about |
| D-11 | **The fixture's catalog is set up like a researcher open.** The large fixture runs `nameformats.Install`, `namealiases.Install`, then `namealiases.Warm` and (from STEP-6) `valuerules.Warm` | A warm run then reads what the app reads | Leaving `newFixture` as is — measures an alias table read per call that the app never makes |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `catalogShape`, `benchShape`, `largeShape` *(new, `bench_test.go`)* | test type, values | `bulkCanon`, `benchPersons`, `benchEvents`, `benchHubs` | Extends `bulkCanon`; the three constants become `benchShape`'s fields and go |
| `core/database/promotealign/perf_test.go` *(new)* `TestLargeSourceBudgets`, `newLargeFixture`, `largeFixture`, `qualifiedEvent`, `fastest` | test file, helpers | `bench_test.go` benchmarks; `promotealign_test.go` `datedOn` | A `Test…` runs in `check-pr.sh`; `datedOn` takes no qualifier, `qualifiedEvent(src, art, eventType string, d *datevalue.Value)` takes a `datevaluetest` value or nil |
| `core/database/databasetest/selects.go` *(new)* `CountSelects`, `Selects` | test helper package | `graphcache` `SetQueryCounterForTest` | That counts only the graph's queries |
| `segmentcompare.CompareNeighbor` *(new)* | func | `segmentcompare.Compare` | `Compare` becomes a one-line wrapper over it |
| `(*state).pair`, `state.pairs` *(new)* | method, field | `(*state).evaluate` | Wraps it with the memo; `evaluate` stays for `rows.go` |
| `keepTop`, package-level `scored` *(new / moved)* | func, type | `Candidates`' local `scored` and its tail | Extracted, shared by `Candidates` and `place` |
| `core/graphalign/export_test.go` *(new)* `NominatedForTest` | test seam | `promotealign/export_test.go` | `graphalign` has no seam; AC-2.3, AC-2.4, AC-2.5, AC-2.6 need a Subject's candidate set |
| `candidatesMemoKey`, `candidatesMemo` *(new, `canon.go`)* | const, type | `layerMemoKey`, `layerMemo` (`layer.go`); `membersMemoKey` | Same mechanism, one more key |
| `Graph.NameFormat`, `Graph.SetNameFormat` *(new)* | methods | `Graph.AliasTables`, `SetAliasTables` | Same contract, a value owned by `valuerules` |
| `valuerules.Warm` *(new)* | func | `namealiases.Warm` | Same shape; `valuerules` owns "the catalog's rules" |

**Builds on.** The Evidence graph memo (S11-19 D-21), the match index lifetime (S11-15 D-9), the open-time warm (`namealiases.Warm`), `Candidates`' ranking rule, the `add-match-rule` invariant that a layer stays pure (no number other than 0 and 1 enters `graphalign` or `segmentcompare`).
**Deletes.** `benchPersons`, `benchEvents`, `benchHubs` (become `benchShape`); `Candidates`' local `scored` type and sort-and-cut tail (moved to `keepTop`); the "no depth limit; the size knobs are `TopK` and `TieCap`" wording that says nominations are unbounded.
**Next adds.** S11-31 sizes its Source cap from `largeFixture` and adds a subtest to `TestLargeSourceBudgets`. S12-09 keys `candidatesMemo` per graph for free (it already is) and changes AC-3.3's expectation to 0 SELECTs. A future budget is one subtest and one constant.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/segmentcompare` | `func CompareNeighbor(a, b Segment, placed Placement, own float64, neighbor func() subjectcompare.Result, rules Rules) Result` *(new)*; `Compare` calls it (D-5) |
| `core/graphalign` | `state.pairs`, `(*state).pair` *(new)*; `score` and `compareSegments` use it (D-6). `place` bounds fresh nominations with `keepTop` (D-7). `candidates.go`: package-level `scored{h Handle; evidence, similarity float64}`, `func keepTop(found []scored, r Rules) []Handle` *(new)*; `Candidates` keeps its signature and doc, plus one sentence: "A placed pair's nominations are cut by the same rule, with no similarity gate (`keepTop`)." |
| `core/database/promotealign` | `loadCanon(g *graphcache.Graph, graphID []byte, prof *profiler, layer graphalign.Layer, fixed []graphalign.Fixed, cfg graphalign.Config, terms segmentcompare.Terms)`; `propose` passes `in.graphID`. `candidatesMemoKey`, `candidatesMemo` (D-8) |
| `core/database/graphcache` | `NameFormat`, `SetNameFormat` and their fields (D-9) |
| `core/database/valuerules` | `Warm` *(new)*; `Load` reads the warm value first (D-9) |
| `core/catalogsession` | `openResearcher` calls `valuerules.Warm(c)` after `namealiases.Warm(c)`, closing the catalog on error the same way |
| `core/database/databasetest` *(new)* | `CountSelects`, `Selects` (D-10) |

Sketch of the place bound (D-7):

```go
var fresh []scored
for _, ce := range st.corresponding(link, hid) {
	g := st.handles[string(ce.neighbor)]
	if g == nil || g.Kind != ns.Kind { continue }
	if c := st.cands[nk][string(g.ID)]; c != nil { /* keep: today's fixed-neighbor via rule */ continue }
	r := st.pair(ns, g)
	fresh = append(fresh, scored{h: *g, evidence: r.Evidence, similarity: r.Similarity})
}
for _, h := range keepTop(fresh, st.cfg.Alignment) { /* st.candidate, then set c.via and c.viaKey as today */ }
```

`keepTop` sorts in place by similarity, evidence, `Ref`, `ID` and returns the kept handles; for the `via` assignment the kept `Handle` is looked up again as `st.handles[string(h.ID)]`.

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — The large fixture and the budgets `main` already meets**
Files: `core/database/promotealign/bench_test.go`, `core/database/promotealign/perf_test.go` *(new)*, `core/database/databasetest/selects.go` *(new)*. Rule: `.cursor/rules/go-tests.mdc`.
Do: D-4 (`catalogShape`, `bulkCanon(db, hub, shape)`, the three benchmarks pass `benchShape`), D-10, D-11 (without `valuerules.Warm`). `newLargeFixture(t)` builds, in this order: `newFixture`; the installs and `namealiases.Warm`; Source "Hub page" with place "Hub" promoted (its handle is the hub); 20 persons, each on its own Source "Census N" (N from 1 to 20) through `filedBirth` with a `largeShape` name and birth year 1850 + N; `bulkCanon(db, hub, largeShape)`; Source A on `f.source`: 40 persons, persons 0–19 named as the 20 filed persons with a birth in their year and a death 50 years later, persons 20–39 with fresh `largeShape` names and a birth in 1800 plus `[0, 100)`, each event dated by D-4's rule through `qualifiedEvent`, each joined by a `subject` participation (100 Subjects); the hub page: 30 persons with a dated birth each, the birth located at the "Hub" place. `TestLargeSourceBudgets` subtests `cold`, `memo matches a fresh run` (AC-1.4; it holds on `main` too), `after a handle commit` (each run first promotes a fresh person "Added N Person" on Source "Other", outside the timing), `evidence warm` (AC-3.1, AC-3.2) and `warm queries` with the graph-query half only (AC-3.2 for `Propose` too). No hub or warm budget yet.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestLargeSourceBudgets|^$' -bench Propose -benchtime 1x -v ./core/database/promotealign/` passes and logs each time; `go vet -tags fts5 ./core/database/...` is clean.

**STEP-2 — Segment compare takes the neighbor lazily**
Files: `core/segmentcompare/segmentcompare.go`, `core/segmentcompare/segmentcompare_test.go`. Skill: `add-match-rule`.
Do: D-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/segmentcompare/ ./core/matchregistry/ ./core/claimevidence/` passes, `TestCompareNeighborMatchesCompare` included.

**STEP-3 — One subject compare per pair within an alignment**
Files: `core/graphalign/align.go`. Skill: `add-match-rule`.
Do: D-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/ ./core/database/promotealign/ ./core/matchregistry/` passes with no test edited.

**STEP-4 — Bound a placed pair's nominations**
Files: `core/graphalign/candidates.go`, `core/graphalign/align.go`, `core/graphalign/export_test.go` *(new)*, `core/graphalign/align_test.go`, `core/database/promotealign/perf_test.go`. Skill: `add-match-rule`.
Do: D-7. `NominatedForTest(layer Layer, canon Canon, fixed []Fixed, cfg Config, subjectID []byte) [][]byte` runs `newState`, `anchor`, `seed`, `walk`, `settle` and returns the handle ids in `st.cands[string(subjectID)]`, sorted by `bytes.Compare`. Add T-5 and the `hub page cold` / `hub page warm` subtests (AC-2.1, AC-2.2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign/ ./core/database/promotealign/ ./core/matchregistry/ ./api/...` passes, every existing golden unchanged; the hub subtests log under 5 s and 3 s.

**STEP-5 — Remember each Subject's candidates until the index changes**
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/memo_test.go`, `core/database/promotealign/perf_test.go` *(new)* (made in STEP-1).
Do: D-8. Add T-6 and the `warm` subtest (AC-1.2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ ./api/...` passes; `TestProposeWarmIssuesNoQuery` and `TestProposeWithOnlyFixedSubjectsBuildsNoIndex` unchanged and green.

**STEP-6 — Warm the name format at project open**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/graph_test.go`, `core/database/valuerules/valuerules.go`, `core/database/valuerules/valuerules_test.go`, `core/catalogsession/session.go`, `core/catalogsession/alias_cache_test.go`, `core/database/promotealign/perf_test.go` *(new)* (made in STEP-1). Skill: `use-catalog-session`.
Do: D-9; the large fixture calls `valuerules.Warm` after `namealiases.Warm`; `warm queries` gains the SELECT half (AC-3.3, AC-3.4). Add T-7, T-8, T-9.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-7 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/comparison-layers.md`, `docs/matching-model.md`, `completed.md`, `deployment-plan.md` (map row and the S11-21 scope entry), this plan (status).
Do: per the Definition of done. In `promote-graph-alignment.md`, the "Thresholds and bands" bullet says: the walk has no depth limit; `TopK` and `TieCap` bound both the shortlisted candidates and the handles one placed pair nominates for each neighbor (S11-21). In `comparison-layers.md`, the "Built in S11-16" paragraph's `place` clause gains "at most `TopK` plus ties, capped at `TieCap`, ranked as `Candidates` ranks, at any similarity (S11-21)". In `matching-model.md`, the `alignment.TopK` and `alignment.TieCap` rows' last column adds "and per layer neighbor of a placed pair". The S11-21 scope entry gains one line: "**Shipped.** D-1 budgets; nominations bounded, pair memo per `Align`, candidates memo, name format warmed at open. Not needed at the budgets: in-place index update, key-value profiling, name-format commit effect."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` *(new)* subtests `cold`, `warm`, `after a handle commit` | Source A of the large fixture → each run within D-1's budget, time logged | AC-1.1, AC-1.2, AC-1.3, D-1, D-3, D-4 |
| T-2 | Go | same, `memo matches a fresh run` | Warm `Propose` on Source A, then `Graph().Drop()` and `Propose` → `reflect.DeepEqual` | AC-1.4, D-8 |
| T-3 | Go | same, `hub page cold`, `hub page warm` | The hub page → within 5 s cold, 3 s warm | AC-2.1, AC-2.2, D-6, D-7 |
| T-4 | Go | same, `evidence warm`, `warm queries` | `Evidence` on every handle row of Source A's proposal, the request's `Neighbors` being every handle row → second pass ≤ 100 ms, 0 graph queries, SELECTs = rows × `SoleGraph`'s; a warm `Propose` → `SoleGraph`'s SELECTs + 1 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-10, D-11 |
| T-5 | Go | `core/graphalign/align_test.go` `TestPlacedPairNominatesTopK` *(new)*, table-driven | A place Subject fixed to handle `h-hub`; three birth Subjects each linked to it by `locLink()`; `h-hub` has 30 canon births by `locLink()`. Cases: births in years 1800–1829 and a layer birth of 1815, `TopK` 5 → `NominatedForTest` for that birth returns 5 ids, the 1815 handle among them; all 30 undated births → 20 ids (`TieCap`); 1 canon event of type `death` and the layer birth (similarity under `MinSimilarity`) → nominated; the 1800 handle listed in `Canon.Candidates` for the birth → it and 5 others | AC-2.3, AC-2.4, AC-2.5, AC-2.6, D-7 |
| T-6 | Go | `core/database/promotealign/memo_test.go` `TestCandidatesMemoFollowsCommits` *(new)*, table-driven on `newWarmFixture` | Subtests: no commit; a handle commit (another "John Smith" born 1817 on a new Source, filed and promoted); a Source edit (a second name observation cited on `bPerson`). Each: propose, change, propose → equals `Propose` after `Graph().Drop()`; the handle-commit case also lists the new handle in the row's alternatives or target | AC-4.1, AC-4.2, AC-4.3, D-8 |
| T-7 | Go | `core/segmentcompare/segmentcompare_test.go` `TestCompareNeighborMatchesCompare` *(new)* | Every case of the existing `Compare` table → `CompareNeighbor` with a counting closure returns the same `Result`; for `Apart` and non-corresponding keys the closure is never called | D-5 |
| T-8 | Go | `core/database/graphcache/graph_test.go` `TestDropLeavesNameFormat` *(new)*; `TestMatchIndexDrops` gains a `SetNameFormat` case | `SetNameFormat("western")`, `Drop` → `NameFormat()` still `"western", true`; `SetNameFormat` after `SetMatchIndex` → match index gone | D-9 |
| T-9 | Go | `core/database/valuerules/valuerules_test.go` `TestLoad` gains subtest `a warmed catalog reads no name format table`; `core/catalogsession/alias_cache_test.go` `TestDoWarmsNameFormat` *(new)* | Installed catalog, `namealiases.Warm`, `valuerules.Warm`, then `Load` under `databasetest.CountSelects` → 0 SELECTs and the same rules as an unwarmed `Load`; a project opened through `catalogsession.Do` → `c.Graph().NameFormat()` holds a `nameformats.Loaded` | AC-3.5, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` goldens and invariants (`align_test.go`, `invariants_test.go`), `core/database/promotealign` (`promotealign_test.go`, `memo_test.go`, `evidence_test.go`), `core/matchregistry` `guard_test.go` (no numeric literal other than 0 and 1 in `graphalign` or `segmentcompare`), `api/ffi/handlers` Promote tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each (T-1 – T-9)
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] The PR description lists the measured time of every TestLargeSourceBudgets subtest (go test -v output), on the machine that ran it
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, S11-21 scope entry Shipped line; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: promote-graph-alignment.md, comparison-layers.md, matching-model.md (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project, add a Source, and cite 40 people with a birth each (common names, years 1800–1900). Open Promote. See the proposal appear within about a second. Set one row to New. See the page re-propose with no visible wait.
2. Open a Source whose births are all located at one place that already has many filed events (`go test -run TestLargeSourceBudgets -v` is the measured version; in the app, any dogfood Source located at a city with hundreds of filed events). Open Promote. See it propose within a few seconds instead of hanging.
3. On the first Source, open the evidence sheet on a row, close it, and open it again. See it open at once with the same lines.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A budget flakes on a slow CI runner | D-1's headroom is 1.7× at its tightest (hub cold); warm runs take the fastest of 3 |
| The bound drops the handle that would have won | It keeps the most similar `TopK` plus ties, the same rule shortlisted candidates already follow; every golden stays green |
| The candidates memo serves a stale list | It lives on the Source's Evidence graph memo and checks the index pointer (D-8); T-2 and T-6 compare with a dropped graph |
| The pair memo grows large | It dies with one `Align`; the prototype's heap without a cross-run memo was 52 MB on the hub fixture |

**Stop and report if:**

- Any existing test in `core/graphalign`, `core/segmentcompare`, `core/claimevidence`, `core/database/promotealign` or `api/ffi/handlers` changes result after STEP-3 or STEP-4 (the prototype changed none; a change means the bound or the memo differs from D-6 / D-7).
- A `TestLargeSourceBudgets` subtest exceeds its D-1 budget after the step that should meet it, on the machine that runs `check-pr.sh`; report the measured times and the profile (`go test -cpuprofile`).
- The authorizer in D-10 does not see the queries (`CountSelects` reads 0 for a cold `Propose`).
- `catalog.go` no longer sets `SetMaxOpenConns(1)`.
- `matchIndex` or `loadShortlist` no longer builds a new `*shortlist.Index` after every handle commit (D-8's pointer check relies on it).
- `nameformats.Install` is called from anywhere other than `core/onboarding/ready.go` and tests (D-9 relies on no other writer).

## 12. Open questions

None.

## 13. PR

**Title:** S11-21 — Large-Source performance check
**Body:** objective in one sentence; what shipped against each US; AC evidence (the `TestLargeSourceBudgets -v` times); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-21-large-source-performance.md`.
