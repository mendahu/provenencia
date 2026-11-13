# S16-13 — Per-graph records in the cache registry, and the Sources section reads them

**Status:** Planned
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Sources section shows the same groups, claims and rows; only where the engine reads them from moves)
**Board:** —
**Depends on:** S16-12 (landed [#800](https://github.com/mendahu/provenencia/pull/800), Source records), S16-15 (landed [#805](https://github.com/mendahu/provenencia/pull/805), the graph index's shells); S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), the registry this PR extends)
**Blocks:** S16-30 (a provenance record composes the Subject record, and adds each Observation's Citation words to it), S16-32 (re-ranks inside the candidates entry this PR registers), S16-24 (the claims read reports its inputs), S15-05, S15-22, S15-24
**Branch:** `s16-13-records-tier`
**Version bump:** none — no change a researcher can see; the wire is unchanged ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc` (a derived cache is a registered entry with Apply, `DerivedFrom`, lifetime, Check and Cap; no field on `Graph`), `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/match-rules.mdc` (the candidates memo is matching-layer state; no knob changes)
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); audit P-4 ([`person-page-audit.md`](../person-page-audit.md)); working rules: "(Batch 2) The point of this batch is simpler code" (as clarified 2026-11-13), "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Per-graph records and memos are registry entries keyed by graph" (K-18, reworded 2026-11-13), "(Batch 2) A Source's words live once", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) Orders are defined and results do not move", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth", "(Batch 2) Batch 1's caps stay where S16-01 puts them"; goal bars 11 and 12
**Planned:** 2026-11-13 against `main` @ `aeb2773`
**Review:** [`S16-13-records-tier.review.md`](S16-13-records-tier.review.md) — round 1: Fits with changes / Not ready; round 2: Fits with changes / Not ready; every finding Fixed, nothing To user; Planned

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

The Person page's claims read (the Sources section's groups and claims) stops loading whole Evidence graphs. A person in 80 graphs today loads all 80 and, because only 64 stay loaded, reloads them on every visit (audit P-4, goal bar 12). After this PR each graph's Subject records (each Subject, its accepted claim, its Observations, and the graph's Subject labels) are a cache entry of their own, and the claims read is lookups: the node's members, the graph index's shell, the Source record, the graph's records. The cache layer gets the plumbing that makes this the norm: every keyed registry entry evicts its least recently used key at its cap; the loaded-graph store, the matching layer and Promote's candidates become typed registry entries keyed by graph; and `graphcache` keeps one eviction rule, one drop path and one Verify path for everything per graph.

**Why now.** S16-12 (Source records) and S16-15 (graph shells) landed, so the claims read can compose from cached records. The user's clarification of "simpler code" (2026-11-13): more cache code is welcome when it makes the reads simple; the construction belongs in the cache layer. The registry gets least-recently-used eviction once, here, instead of a hand-written second store (K-18 reworded, user 2026-11-13). The user chose one PR over a preparatory split (2026-11-13).

**What stays.** The wire; every response's content and order; S16-01's records-per-graph cap (`boundedMembers`); the record links' lines and bounds (record links still build each graph's matching layer from the whole loaded graph, so a person past 64 graphs still reloads graphs there, as `TestRecordLinksBudgets`' past-cap case already measures); Promote's suggestions; 64 loaded graphs; the candidates memo's generation check (S16-32 replaces it). Each Observation's Citation label, page, Artifact id and uncertainty flag join the records with S16-30, their first reader (user, 2026-11-13).

## 2. User stories

**US-1.** As a researcher, when I open the Sources section of a person cited in 80 Evidence graphs, its groups and claims read without loading any whole graph, and read again from memory, so that the section stays fast past 64 graphs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `households(f, 80)` (80 Sources, one graph each, the son promoted onto one entity); `g.Drop()`, then `persondetail.Claims(son)` | 80 groups of one claim each; `g.LoadedEvidenceGraphCountForTest()` is 0 |
| AC-1.2 | The same, read a second time | Zero queries (`SetQueryCounterForTest`); the same 80 groups |
| AC-1.3 | The fixtures of `TestClaimsGroupsBySource`, `TestClaimsOrderSources`, `TestClaimsMissingAndEmpty`, `TestClaimsAfterSourceEdit` | The same groups, claims, Subjects, Observations and order as at the planning base (field paths follow D-9) |
| AC-1.4 | `TestClaimsRecordsPerGraph`, `TestCraftedClaims` (20,000 records in one graph) | Unchanged results; the crafted read within `capBudget` (1 s) |
| AC-1.5 | `households(f, 80)`: cold read (`catalogsessiontest.Cold`), then warm read | Cold within 1 s, warm within 50 ms; the planning-base figures on the same machine are in the ship log |
| AC-1.6 | `ListPersonClaims` over the handler test's Person, one of whose Observations names a second Subject (a `Relationship`) | `value_subject_label` is that Subject's label; the rest of the response as before |
| AC-1.7 | `persondetail.RecordLinks` at `households(f, 80)` | Within `budgetLinks80` (D-15); the planning-base figure in the ship log |

**US-2.** As the engine, I keep each Evidence graph's Subject records exactly, so that a reader composes from them without loading the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A graph's records read through `graphfill.RecordsOf` while the graph is loaded, and again after `graphfill.Records.Drop(g, [][]byte{id})` and `g.Drop()` | The two `*GraphRecords` are `reflect.DeepEqual`; the first fill issues no query |
| AC-2.2 | Records held for graph A; loaded-graph cap set to 1 and graph B loaded (A evicted) | `graphfill.Records.Peek(g, A)` still holds A's records |
| AC-2.3 | Records held for graphs A and B; a write adds an Observation on A (`Set.EvidenceGraphs` = [A]) | A's records are gone, B's are held; the next `RecordsOf(A)` has the new Observation; with `PROVENENCIA_GRAPH_VERIFY=1` the write passes Verify |
| AC-2.4 | Records held; `g.Drop()` | `KeyCounts()["graphfill.records"]` is 0 |
| AC-2.5 | A graph with a Subject that has an accepted claim, a Subject without one, and an Observation whose value is the second Subject | `Subject(id)` returns the claim in `Accepted` for the first and a zero `Accepted` for the second; `Labels()[string(id)]` is each Subject's label; `Entity(id)` is the claimed entity for the first and nil for the second |
| AC-2.6 | An id with no graph row, read through `RecordsOf` | Empty records (no Subjects), stored |

**US-3.** As the engine, every keyed cache entry keeps its most recently used keys at its cap, so that a reader past the cap still gets its value stored.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Cap 2: `Set` A, `Set` B, `Get` A (hit), `Set` C | A and C held, B gone; `KeyCounts` 2 |
| AC-3.2 | Cap 2, A and B held: `Peek` A, then `Set` C | A held (a Peek is a use), B gone |
| AC-3.3 | Cap 2, A held: `GetMany` [A, B, C] filling B and C | Returns A, B and C; holds B and C; A gone (A is touched first, then B and C are stored in the order asked) |
| AC-3.4 | `Check` over a full entry, and `KeyCounts`, then `Set` of a new key | The key evicted is the one that was least recently used before the Check |
| AC-3.5 | The loaded-graph store as an entry: the cases of `TestEvidenceGraphCap` (oldest evicted; a read is a use; a drop clears) | Unchanged results |

**US-4.** As the engine, the matching layer and Promote's candidates are typed registry entries keyed by graph, so that no reader casts a memo and no commit loop clears them by hand.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A layer and candidates held for graph A; a vocabulary commit (`ListPropertyVocabulary`) | Both entries hold no key |
| AC-4.2 | A layer and candidates held for graph A; `nameformats.Invalidate(c)` | Both entries hold no key |
| AC-4.3 | Layers and candidates held for A and B; a write that names A | A's are gone, B's are held |
| AC-4.4 | A layer held for A; the loaded graph A evicted; `evidencelayer.Load(g, A)` | Returns the held layer and loads no Evidence graph |
| AC-4.5 | `evidencelayer.Load` for an unknown graph id, twice | `ErrNoGraph` both times; `Layers` holds no key |
| AC-4.6 | `rg -n 'EvidenceGraphMemo\|SetEvidenceGraphMemo\|memos\|evictEvidenceGraphs\|evidenceTick\|SetEvidenceGraphLoader\|SetEvidenceGraphCheck\|func (g \*Graph) EvidenceGraph' core/database/graphcache -g '!*_test.go'` | No hit; `OnCommit`'s vocabulary branch has no loop |
| AC-4.7 | Promote's suites (`./core/database/promotealign/...`, `./core/database/promote/...`, the matching goldens) | Green, unchanged |
| AC-4.8 | `PROVENENCIA_GRAPH_VERIFY=1` after a proposal | Verify runs the layers and candidates Checks and passes |
| AC-4.9 | With `PROVENENCIA_GRAPH_VERIFY=1`, a commit over held records, layers, candidates and loaded graphs at a full cap | Every entry holds the same keys in the same use order as with Verify off |

## 3. Scope

| In | Covered by |
| --- | --- |
| Keyed registry entries evict the least recently used key at their Cap; Get, GetMany, Peek and Set count as a use; a test-only cap override | STEP-1, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| `graphcache.DropNamedGraphs` *(new)*: the one Apply of an entry keyed by Evidence graph; `graphcache.SubjectLabels` *(new)* | STEP-1, AC-2.3, AC-4.3 |
| The loaded-graph store becomes `graphfill.Graphs` *(new)*, a keyed entry owned by `graphfill`; `graphcache`'s store, its loader and check hooks, LRU, drop, Verify loop and typed field go; the ten callers read `graphfill.EvidenceGraph` | STEP-3, AC-3.5, AC-4.6 |
| The matching layer becomes `evidencelayer.Layers` *(new)*; `Load` reads its own name pattern | STEP-2, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, AC-4.8, AC-4.9 |
| Promote's candidates become `promotealign` `candidates` *(new)*, keeping the generation check | STEP-2, AC-4.1, AC-4.2, AC-4.3, AC-4.7, AC-4.8, AC-4.9 |
| The memo API and `OnCommit`'s vocabulary loop go | STEP-2, AC-4.6 |
| `graphfill` splits its Subject, Observation and member queries into one step both fills call | STEP-4, AC-2.1 |
| `graphfill.Records` *(new)*: each graph's `GraphRecords`, filled from the loaded graph when it is held, else from the split step; checked | STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6 |
| The claims read composes from the node's members, the graph index, Source records and the records entry; `memberGraphs` goes | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| The record links read takes its graphs and claimed entities from the same composition | STEP-6, AC-1.7 |
| `ListPersonClaims` takes each graph's Subject label map from the records | STEP-6, AC-1.6 |
| Budgets and sizes; `evidencegraphstest.ScaleSeed` *(new)* shared by the two graph benchmarks | STEP-7, AC-1.5, AC-1.7 |
| Docs: `catalog-caches.md`, the design doc, the map, the effects README | STEP-8 |

| Out | Home |
| --- | --- |
| Each Observation's Citation label, page, Artifact id and uncertainty flag in the records | S16-30 |
| The provenance sheet and the detail read composing from the records | S16-30 |
| Candidates re-ranking only what moved; the generation check going | S16-32 |
| Record links building layers without the whole loaded graph; a layer or loaded-graph cap above 64 | not planned ([`structural-debt.md`](../../../ideas/structural-debt.md) batch graph read) |
| A batch fill of several graphs' records in one query set | [`structural-debt.md`](../../../ideas/structural-debt.md) (S13-16 PRF-2) |
| The claims read reporting its inputs | S16-24 |
| `BenchmarkSourceRecords` sharing `ScaleSeed` (review M-9) | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Loaded-graph store | `core/database/graphcache/evidence_graph.go` `loadedEvidenceGraph`, `EvidenceGraph`, `touchEvidenceGraph`, `evictEvidenceGraphs`, `dropEvidenceGraphs`, `maxEvidenceGraphs` (64), `SetEvidenceGraphLoader`, `SetEvidenceGraphCheck` | A typed map with its own LRU (a tick scan), its own drop and loader/check hooks; each graph holds `memos map[string]any` |
| Store callers | `g.EvidenceGraph`: `core/database/evidencelayer/evidencelayer.go` `Load`; `core/database/promotealign/promotealign.go` (three: `Propose`'s input check, `loadInputs`, and the read near line 165); `core/database/persondetail/claims.go` `memberGraphs`; `api/ffi/handlers/subjects.go`, `citations.go`, `observations.go`, `promote.go`, `event_titles.go` (one each) | Ten callers; the loader is reached only because `graphfill` registered itself, so tests blank-import `graphfill` |
| Memo API | same file `EvidenceGraphMemo`, `SetEvidenceGraphMemo` | Untyped memos keyed by string, alive while the graph is loaded; callers cast |
| Vocabulary loop | `core/database/graphcache/commit.go` `OnCommit`, the `ListPropertyVocabulary` branch; `g.dropEvidenceGraphs(set.EvidenceGraphs)` | Sets every loaded graph's `memos = nil` (S16-10 D-15 gave it to this PR); drops named graphs |
| Store fields | `core/database/graphcache/graph.go` `Graph.evidenceGraphs`, `evidenceTick`, `evidenceCap`, `clear`, `LoadedEvidenceGraphForTest`, `SetEvidenceGraphCapForTest`, `LoadedEvidenceGraphCountForTest` | The store and its test hooks |
| Store Verify | `core/database/graphcache/verify.go` `Verify` | A loop of `evidenceGraphFn` over loaded graphs beside `verifyEntries` |
| Registry, keyed | `core/database/graphcache/cache.go` `KeyedEntry`, `Keyed[T].data`, `store`, `Get`, `GetMany`, `Peek`, `Set`, `Drop`, `Clear`, `RegisterKeyed`, `KeyCounts`, `verifyEntries` | Values in `map[string]T`; `store` refuses a new key once the map holds `Cap` keys; `GetMany` stores every key the fill returns, in map order |
| Keyed users | `core/database/canonlinks/reader.go` `pivots`, `forAnswers`, `siblingAnswers` (Cap `maxKeys` 65,536, comment "a key not stored is derived again on read"); `Reader.pivot` uses `Peek` then `Get`; `forAnswers.GetMany`, `siblingAnswers.GetMany` | The only keyed entries |
| Matching layer memo | `core/database/evidencelayer/evidencelayer.go` `Load(g, graphID, namePattern)`, `memoKey`, `build` | Loads the whole graph, looks up memo `"evidencelayer\x00"+namePattern`, builds and stores on a miss; `ErrNoGraph` for a graph that is not `Found`. Callers pass the pattern from `valuerules.Load` (`persondetail/recordlinks.go:105`, `promotealign/promotealign.go:135`, `evidencelayertest.go:310`; tests) |
| Name pattern | `core/database/valuerules/valuerules.go` `Load` → `nameformats.LoadFor(g).Key`; `core/database/nameformats/nameformats.go` `Default` (catalog-lifetime whole entry), `Invalidate` | The pattern every caller passes; `Invalidate` drops `Default` and every entry derived from it |
| Candidates memo | `core/database/promotealign/canon.go` `loadCandidatesMemo`, `candidatesMemoKey`, `candidatesMemo{gen, bySubject}`; `index.go` `loadShortlist`, `matchIndexOf`, `shortlistInputs`; `promotealign.go` `loadConfig` | Memo `"promotealign.candidates\x00"+namePattern`, used only when `gen` matches the shortlist index's generation; never changed in place |
| Graph loader | `core/database/graphfill/load.go` `loadEvidenceGraph`, `loadGraphHead`, `loadGraphSubjects`, `loadGraphObservations`, `loadGraphMembers`, `loadTitleEdges`, `checkEvidenceGraph` | Subjects by label then ref, Observations by ref, members (accepted claims) by Subject ref |
| Claims read | `core/database/persondetail/claims.go` `Claims`, `claims`, `boundedMembers`, `memberGraph`, `memberGraphs` | Loads each member's whole graph, finds member and Subject by linear scan, scans all Observations per claim; a graph whose Source has no record is an error; groups and sorts with `conclusionsummaries.SortByTitle` |
| Claim type | same file `Claim{ClaimID, Graph graphcache.EvidenceGraph, Subject, Observations}` | Carries the whole loaded graph |
| Family comment | `core/database/persondetail/family.go` `dropUnread` doc ("as memberGraphs skips") | Names the function this PR deletes |
| Record links | `core/database/persondetail/recordlinks.go` `RecordLinks`, `recordLinks`, `SourcedGraph{Graph graphcache.EvidenceGraph, Source}`, `linksBetween` | Takes graphs from `claims`; `linksBetween` maps Subject → entity from `graph.Members`; `evidencelayer.Load(g, id, namePattern)` |
| Handlers | `api/ffi/handlers/person_detail.go` `ListPersonClaims`, `recordLinksProto`; `api/ffi/handlers/source_read.go` `listedContext.rows`, `rowsOf`, `listedProtosFromGraph` (callers in `citations.go`, `observations.go`) | `rowsOf(eg, obs)` builds a label map keyed `string(s.ID)` from `eg.Subjects` on every call, once per claim |
| Graph shells | `core/database/evidencegraphs/index.go` `IndexOf`, `GraphIndex.Graph`; `evidencegraphs.go` `Graph{ID, Ref, SourceID, Name, Description, ArtifactIDs}` | Every graph's row (S16-15) |
| Source records | `core/database/sources/records.go` `RecordsOf`, `RecordIndex.Source` | Every Source's words (S16-12) |
| Node members | `core/database/graphcache/graph.go` `Member{SubjectID, SourceID, GraphID, Accepted}`, `Node.AcceptedMembers` | No claim id |
| Effects | `core/database/effects/registry.go` `Set.EvidenceGraphs`; `core/database/effects/README.md` line 9 ("graphcache drops each, and nothing else follows it") | The exact signal for a graph's rows |
| Cache doc test | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry`, `TestCacheDocSeesEveryOwner`, blank imports | Every owner must be imported |
| Tests | `core/database/graphcache/cache_test.go` `TestEntryCap`; `core/database/graphcache/evidence_graph_memo_test.go`; `core/database/graphcache/evidence_graph_cap_test.go` `TestEvidenceGraphCap` (memo steps); `core/database/graphfill/load_test.go` (memo lines in `TestRetitleKeepsEvidenceGraph`, `TestGraphWriteDropsOnlyThatGraph`); `core/database/evidencelayer/evidencelayer_test.go` (the memo check after the `ErrNoGraph` loop); `core/database/promotealign/memo_test.go` `TestMemoFollowsCommit`, `TestLoadSharesPromotesMemo`, `TestCandidatesMemoFollowsCommits`; `core/database/persondetail/claims_test.go`, `caps_test.go`, `recordlinks_test.go`, `recordlinks_perf_test.go` (`households`, `TestRecordLinksBudgets`, `budgetLinksPast`); `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims`, the `SourcedGraph` literal | Cover today's behavior; the memo tests read the API this PR removes |
| Benchmarks | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes`; `core/database/evidencegraphs/index_bench_test.go` `BenchmarkGraphIndex` (inline closures `graph`, `subject`, `obs` over one transaction, then a crafted heavy graph and 10,000 Wide graphs) | Sizes per entry; the Scale seed lives inside one benchmark |

**Precedent.** S16-12 ([#800](https://github.com/mendahu/provenencia/pull/800)) and S16-15 ([#805](https://github.com/mendahu/provenencia/pull/805)): an owner package registers an entry from `init`, exports a typed read (`RecordsOf`, `IndexOf`), its Check rebuilds from SQL and `reflect.DeepEqual`s, and the claims read composes from it. Copy the owner-registers-and-exports shape and the Check. `canonlinks`' `pivots` is the precedent keyed entry. `shortlistInputs(g, true)` is the precedent for a Check that rebuilds with fresh inputs.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A keyed entry evicts its least recently used key when a new key arrives at `Cap`. Its data is `keyedValues[T]`: a map from key to a `container/list` element and one list in use order. `Get`/`GetMany` hits, `Peek`, `Set` and stores move a key to the front; `Check`, `KeyCounts` and `Drop` do not. `GetMany` touches held ids in `ids` order, then stores the fill's values in `missing` order; a key the fill returns that was not asked for is neither stored nor returned. `(k Keyed[T]) SetCapForTest(g, n)` *(new)* overrides the Cap for one graph: the override lives in `Graph.capsForTest map[int]int` *(new, by entry id)*, survives `Drop` and `Clear`, applies to a container created later, evicts at once, and goes only with `SetCapForTest(g, 0)`. A Check reads no keyed entry through `Get`, `GetMany`, `Peek` or `Set`: it rebuilds from fresh inputs (review M-11) | One eviction rule serves every per-key cache; O(1) per use; a defined order makes eviction testable; Verify stays read-only | Refusing past the cap (today); a tick scan — O(Cap) per store, 65,536 for `pivots`; an override inside `keyedValues` — `Drop` would clear it |
| D-2 | `graphfill` owns the loaded-graph store as a keyed entry: `graphfill.Graphs` (`Keyed[graphcache.EvidenceGraph]`, `Name: "graphfill.graphs"`, `Cap: MaxEvidenceGraphs` (exported, 64), `Apply: graphcache.DropNamedGraphs`, `Check: checkEvidenceGraph`), registered in `graphfill`'s `init`. `graphfill.EvidenceGraph(g, id) (graphcache.EvidenceGraph, error)` *(new)* is its `Get` with `loadEvidenceGraph` as fill (setting `GraphID`, same invalid-id error); `graphfill.PeekEvidenceGraph(g, id)` *(new)* its `Peek`; `graphfill.LoadEvidenceGraph(g, id)` *(new, exported `loadEvidenceGraph`)* reads without caching, for Checks. The ten callers switch to `graphfill.EvidenceGraph`. From `graphcache` go: `(*Graph).EvidenceGraph`, `EvidenceGraphLoader`, `EvidenceGraphCheck`, `SetEvidenceGraphLoader`, `SetEvidenceGraphCheck`, `evidenceGraphLoad`, `evidenceGraphFn`, `loadedEvidenceGraph`, `touchEvidenceGraph`, `evictEvidenceGraphs`, `dropEvidenceGraphs`, `maxEvidenceGraphs`, `Graph.evidenceGraphs`, `evidenceTick`, `evidenceCap`, the Verify loop and the three `*EvidenceGraph*ForTest` hooks (tests use `graphfill.Graphs.Peek`, `KeyCounts()["graphfill.graphs"]` and `SetCapForTest`). `EvidenceGraph`, its row types and `WalkEvidenceGraph` stay in `graphcache`. Test files that blank-import `graphfill` keep the import | One way to register a per-graph entry; the loaded graphs, their records, loader and Check in one package; no hidden init dependency; `catalog-queries.mdc` forbids a `Graph` field for a derived cache (review M-1, M-10) | The store as a typed field (two LRUs, two drops, two Verify loops); the store as an entry in `graphcache` fed by setter hooks — a second way to register |
| D-3 | Apply of every entry keyed by Evidence graph is `graphcache.DropNamedGraphs[T]` *(new)*: `k.Drop(g, set.EvidenceGraphs)`. The loaded graphs now drop in `applyEntries`, after the commit's reload, instead of first; nothing between reads a loaded graph, and a `Structure` commit still empties every entry through `Drop` | `Set.EvidenceGraphs` is the exact signal; four entries share it | Each owner writing the same lines |
| D-4 | `GraphRecords` (new, in `graphfill`) holds `GraphID`, the graph's `SubjectRecord`s in the loader's Subject order, an index by Subject id, and `labels` (`graphcache.SubjectLabels(subjects)`, keyed `string(id)`). `SubjectRecord{Subject graphcache.EvidenceSubject; Accepted graphcache.EvidenceMember; Observations []graphcache.EvidenceObservation}`: the accepted claim (zero when none; the catalog allows one per Subject) and the Subject's Observations in ref order, nil when it has none. Methods: `Subject(id) (SubjectRecord, bool)`, `Labels() map[string]string` (the shared map), `Entity(subjectID) []byte`. Everything returned is shared and read-only | What the claims read, the handler and record links ask of a graph, built once; reuses graphcache's row types | A new Subject or Observation type; Citation words now — no reader until S16-30 (user, 2026-11-13) |
| D-5 | `graphcache.SubjectLabels(subjects []EvidenceSubject) map[string]string` *(new)* is the one Subject label map; `newGraphRecords` and `listedContext.rows` call it (review M-5) | One label rule | Two loops building the same map |
| D-6 | One constructor `newGraphRecords(graphID, subjects, observations, members)`. `graphfill.RecordsOf(g, graphIDs) (map[string]*GraphRecords, error)` fills through `Records.GetMany`: each missing graph from `PeekEvidenceGraph(g, id)` when held (a held graph that is not `Found` gives empty records), else from `loadGraphRows(q, id)` *(new)*, the step split out of `loadEvidenceGraph`, which calls it too. A nil graph is `database.ErrClosed` | One loader of Subject, Observation and member rows; a graph just viewed costs no SQL; a Peek's use is harmless (review M-8) | Always SQL; filling records inside `EvidenceGraph` — fills them for every Promote load |
| D-7 | `graphfill.Records` (`Keyed[*GraphRecords]`, `Name: "graphfill.records"`, `Cap: maxGraphRecords` = 4,096 with `// decided: user, 2026-11-13`, `Apply: DropNamedGraphs`, `Check`: rebuild from `loadGraphRows` and `reflect.DeepEqual`). An id with no graph row stores empty records; `Set.EvidenceGraphs` names a deleted graph, so they leave with it | Registry carries cap, Verify, drop and sizing (K-18 reworded); 4,096 is 40% of the batch's 10,000 graphs | A second map in the store (K-18's first wording) |
| D-8 | `evidencelayer.Load(g, graphID)` drops its `namePattern` parameter. `Layers` (`Keyed[Layer]`, `Name: "evidencelayer.layers"`, `DerivedFrom: Vocabulary{}, nameformats.Default`, `Cap: graphfill.MaxEvidenceGraphs`, `Apply: graphcache.DropNamedGraphs`). `Load` is `Layers.Get(g, graphID, fill)`; the fill reads `graphfill.EvidenceGraph(g, graphID)`, returns `ErrNoGraph` when it is not `Found` (nothing stored), reads the pattern with `nameformats.LoadFor(g).Key` and calls `Build`. `build` is exported as `Build(g, eg, namePattern)`. Check: for each held id, `graphfill.LoadEvidenceGraph` (fresh, stores nothing) and `Build`, then `reflect.DeepEqual`; a graph no longer `Found` is a mismatch. Callers drop their pattern argument | The pattern always comes from `nameformats.Default`, so declaring it as a source replaces the pattern in the key and the reader's hand check (review M-3); a held layer needs no graph load; the Check fills nothing (M-11) | Pattern in the value with a mismatch branch; `DerivedFrom` the alias index — the layer does not read it (S16-10 D-15); a Check through `Load` — fills the store and `Layers` during Verify |
| D-9 | `promotealign` `candidates` (`Keyed[candidatesMemo]`, `Name: "promotealign.candidates"`, `DerivedFrom: Vocabulary{}, nameformats.Default`, `Cap: graphfill.MaxEvidenceGraphs`, `Apply: graphcache.DropNamedGraphs`). `candidatesMemo{gen, bySubject}` stays; `loadCandidatesMemo` reads with `Peek` (a stale generation is a miss, and its touch is intended) and stores with `Set`. The ranking body becomes `rankCandidates(g, prof, ix, layer graphalign.Layer, subjects []graphalign.Subject, cfg) (map[string][][]byte, error)` *(new, unexported)*: today's `seedHandles` plus `graphalign.Candidates` loop, each Subject's list starting `[][]byte{}` as today; `loadCandidatesMemo` calls it for the missing Subjects. Check, per held graph: `m := matchIndexOf(g)`; skip when `m.shortlist` is nil or `stored.gen != m.gen`; else `eg := graphfill.LoadEvidenceGraph(g, id)` (not `Found` → mismatch), `ev := evidencelayer.Build(g, eg, pattern)` with `cfg, pattern, prof := shortlistInputs(g, true)`, `metas := valuerules.Metas(g)`, `layer := graphalign.Layer{Subjects: ev.Subjects, Bridges: ev.Links, Metas: metas}` (as `loadInputs` builds it); `subjects` are `ev.Subjects` whose ids are keys of `stored.bySubject` (a stored id missing from `ev` → mismatch); `fresh := rankCandidates(g, prof, m.shortlist, layer, subjects, cfg)`; `reflect.DeepEqual(fresh, stored.bySubject)` | Same signals as today's memo plus the name format; a Check for the entry S16-32 will patch (review M-2), filling nothing (M-11); one ranking path | `DerivedFrom` `MatchIndex` — `MatchIndex.Set` runs while profiles fill, which would empty candidates mid-proposal; no Check — against `catalog-queries.mdc` |
| D-10 | The memo API (`EvidenceGraphMemo`, `SetEvidenceGraphMemo`), `evidencelayer.memoKey`, `promotealign.candidatesMemoKey` and `OnCommit`'s vocabulary loop go | Nothing stores memos after D-8 and D-9 | Keeping the API for later |
| D-11 | The claims read: `boundedMembers(n)`; `evidencegraphs.IndexOf`, `sources.RecordsOf`; the shells `idx.Graph(m.GraphID)` (a member with no shell is skipped); `graphfill.RecordsOf` over the graph ids that have a shell; per member, `records.Subject(m.SubjectID)` (skipped when missing or when `Accepted.EntityID` is not `n.ID`, as today); a shell whose Source has no record is an error, as today; groups and orders as today. `Claim` becomes `{Graph evidencegraphs.Graph; graphfill.SubjectRecord; Records *graphfill.GraphRecords}` (embedded, so `cl.Subject`, `cl.Observations` read as before and the claim id is `cl.Accepted.ClaimID`) | The read is lookups; one Subject-record shape that S16-30 extends (review M-4); no empty records stored for graphs that left | Keeping `Graph graphcache.EvidenceGraph`; caching the per-person answer — no stored answer without a measurement |
| D-12 | `SourcedGraph` becomes `{Graph evidencegraphs.Graph; Source sources.Record}`; `recordLinks` collects each distinct graph's `*GraphRecords` from the claims in claim order; `linksBetween(c, page, records []*graphfill.GraphRecords, cfg)` builds each layer with `evidencelayer.Load(g, r.GraphID)` and each member's entity with `r.Entity` | Entities from the same composition; layers still compare the whole graph (Out) | Passing loaded graphs |
| D-13 | `listedContext.rowsOf(labels map[string]string, obs)`; `rows(eg, keep)` calls `graphcache.SubjectLabels(eg.Subjects)` once per call; `ListPersonClaims` passes `cl.Records.Labels()` | The map is built once per graph at fill | A per-claim map, as today |
| D-14 | `evidencegraphstest.ScaleSeed` *(new)*: `NewScaleSeed(tb testing.TB, c *database.Catalog) *ScaleSeed` looks up the person type (`graphcachetest.SubjectTypeID`), the `remark` property and upserts the `book` source type, then opens a transaction; it does not seed. Fields `PersonTypeID`, `PropID`, `TypeID []byte`. Methods (no return; a failed insert is `tb.Fatal`): `Exec(q string, args ...any)`, `Source(i int)` (a Source and its one Artifact), `Graph(i, source int)`, `Subject(i, graph, source int)`, `Obs(i, subject, graph int)`, `Base()` (Sources 1 to `ScaleSources+2`, the last two left for crafted rows; graphs, Subjects and Observations as today), `Commit()`. Constants `ScaleSources` 3,000, `ScaleGraphs` 10,000, `ScaleSubjects` 20,000, `ScaleObservations` 50,000; ids from `databasetest.BenchID` kinds 1–6 as today. `BenchmarkGraphIndex` calls `Base`, then writes its crafted graphs through a second `NewScaleSeed` after measuring, as it does today; its figures stay within noise. `evidencegraphstest` gains imports of `graphcachetest` and `databasetest`, neither of which imports `evidencegraphs` | One Scale seed for both graph benchmarks; the crafted half stays where it is measured | A copy of the insert loop in `graphfill` |
| D-15 | Budgets: `persondetail_test` `TestClaimsBudgets` *(new, `claims_budget_test.go`)* with `budgetClaimsCold = time.Second`, `budgetClaimsWarm = 50 * time.Millisecond` at `households(f, 80)`; `TestRecordLinksBudgets` gains a `households(f, 80)` case under `budgetLinks80` *(new)*, set in STEP-1 to the planning-base figure × 1.25 rounded up to 100 ms (no lower than `budgetLinksPast`) with a `// measured: S16-13 planning base` comment; `graphfill_test` `BenchmarkGraphRecords` *(new, `records_bench_test.go`)* fails past `maxRecordsBuild = 2 * time.Second` or `maxRecordsHeapMB = 35` | Rule "Measure growth" at the map's 80-graph Scale (review M-6); the record links case's base is unknown until measured (X-15) | Unbudgeted; `budgetLinksPast` at 80 — sized for 65 |

**Rules this plan bends.**

- "(Batch 2) Per-graph records live in the Evidence-graph store" (K-18): the records, the loaded graphs and both memos are registry entries keyed by graph; there is no tier inside a hand-written store. User, 2026-11-13 (registry shape, one PR). Updated at planning in the map (rule reworded), the design doc and the S16-32 scope entry; D-2 follows the same answer for the store itself.
- "(Batch 2) The point of this batch is simpler code": not bent; clarified by the user, 2026-11-13. Recorded in the rule at planning.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `keyedValues[T]`, LRU in `Keyed[T]`, `SetCapForTest` | mechanism | `evictEvidenceGraphs`, `SetEvidenceGraphCapForTest` | Extends the registry; replaces both |
| `graphfill.Graphs`, `EvidenceGraph`, `PeekEvidenceGraph`, `LoadEvidenceGraph`, `MaxEvidenceGraphs`; `Graph.capsForTest` | entry, func, const, field | the typed store and its setter hooks in `graphcache` | The store moves to the package that loads it, registered like `Records`; the test override replaces the store's own hook |
| `graphcache.DropNamedGraphs[T]` | func | `canonlinks.clearOnHandles` | Same shape, other signal |
| `graphcache.SubjectLabels` | func | `rowsOf`'s loop | Moved, shared |
| `graphfill.GraphRecords`, `SubjectRecord`, `newGraphRecords`, `Records`, `RecordsOf`, `maxGraphRecords`, `loadGraphRows` | type, func, entry, const | `sources.Records`/`RecordsOf`, `loadEvidenceGraph` | Same owner shape; keyed; the loader split, not copied |
| `evidencelayer.Layers` | entry | `memoKey` + memo API | Replaces it, typed |
| `promotealign` `candidates`, `rankCandidates` | entry, func | `candidatesMemoKey` + memo API; the body of `loadCandidatesMemo` | Replaces the memo; the body is extracted so the Check reuses it |
| `evidencegraphstest.ScaleSeed` | helper | `BenchmarkGraphIndex`'s closures | Moved, not copied |
| Tests (§8) | tests | `TestEntryCap`, `TestRecordLinksBudgets`, `BenchmarkGraphIndex` | New behavior |

**Builds on.** The cache registry (S16-10), the owner-registers shape of S16-12 and S16-15, `graphfill`'s loader registration, `shortlistInputs`, `catalogsessiontest.Cold`, `databasetest.HeapDelta`, `evidencelayertest.Records`, the `households` fixture.
**Deletes.** From `graphcache`: `(*Graph).EvidenceGraph`, `EvidenceGraphLoader`, `EvidenceGraphCheck`, `SetEvidenceGraphLoader`, `SetEvidenceGraphCheck`, `evidenceGraphLoad`, `evidenceGraphFn`, `loadedEvidenceGraph`, `touchEvidenceGraph`, `evictEvidenceGraphs`, `dropEvidenceGraphs`, `maxEvidenceGraphs`, `Graph.evidenceGraphs`, `evidenceTick`, `evidenceCap`, the Verify loop over loaded graphs, `LoadedEvidenceGraphForTest`, `SetEvidenceGraphCapForTest`, `LoadedEvidenceGraphCountForTest`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo`, `OnCommit`'s memo loop and `dropEvidenceGraphs` call; `graphfill`'s `init` calls to the setters; `evidencelayer.memoKey` and `Load`'s `namePattern` parameter; `promotealign.candidatesMemoKey`; `persondetail.memberGraph`, `memberGraphs`, `Claim.ClaimID`; `rowsOf`'s label loop; `TestEntryCap` (replaced), `core/database/graphcache/evidence_graph_memo_test.go`, `core/database/graphcache/evidence_graph_cap_test.go` (its cases move to `core/database/graphfill/graphs_test.go` *(new)*), the memo lines of `core/database/graphfill/load_test.go`; the "refuses past the cap" text (`store`'s comment, `canonlinks.maxKeys`'s comment); `catalog-caches.md`'s typed-field row for `evidenceGraphs` and its "S16-13 owns it" text.
**Next adds.** S16-30 adds Citation words to `SubjectRecord` and composes a provenance record from `Records`; S16-32 changes `candidates`' fill and Apply, keeping its Check honest; S16-24 reports the graphs a claims read used; S16-18's place entries get LRU without code.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A keyed entry evicts instead of refusing | every `Keyed[T]` reader | A reader always gets its value; an evicted key refills (T-1) |
| A held layer or candidates with no loaded graph | `evidencelayer.Load`, `loadCandidatesMemo` | Same value as a fresh build; same drop signals (T-5, T-6, Checks) |
| Records of a graph that left | the claims read | No shell, so the member is skipped and no records are asked for (code; the index is patched on the same commit) |
| A member whose record is missing, or whose claim names another entity | the claims read | Skipped (T-8) |
| A shell whose Source has no record | the claims read | Error, as today (unreachable: a Source with graphs cannot be deleted) |

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; 3,000 Sources, 10,000 graphs, 50,000 Observations; a person in 50 and 80 graphs):

- **Claims read, warm:** per member, three map lookups; no query (AC-1.2). **Cold:** one `loadGraphRows` per graph not held (three queries plus the date and name lookups), 80 graphs → about 400 queries once; budget 1 s (D-15).
- **Record links at 80 graphs:** unchanged shape (layers and loaded graphs capped at 64 thrash); measured under `budgetLinks80` (AC-1.7).
- **Records entry:** at most 4,096 graphs; all 10,000 measured by `BenchmarkGraphRecords` (D-15).
- **Apply per commit:** `DropNamedGraphs` on four entries, O(graphs named), usually one; a Source of 10,000 Observations in one graph names one graph.
- **LRU overhead:** one list element per key; `pivots` at 14,162 keys on the large fixture; `BenchmarkCacheSizes` kinship before and after (Stop past +20%).
- **Layers and candidates:** 64 graphs each, today's memo bound.

### Edges

- **Assumptions.** One accepted claim per Subject (`identity_claims_one_accepted_per_subject`); the loader's Subject and Observation order is the response order (AC-1.3); registry keys are 16 bytes and graph ids are; the name pattern every caller used is `nameformats.LoadFor(g).Key` (§4 lists the callers; T-5 holds `Load` to it).
- **Neighbors.** A key at exactly Cap stores and evicts one (T-1); Peek counts, Check does not (AC-3.2, AC-3.4); a graph loaded vs not fills the same records (AC-2.1); a name-format Invalidate empties layers and candidates (AC-4.2).
- **Removed inputs.** A graph deleted: its loaded graph, records, layer and candidates leave with `Set.EvidenceGraphs` (T-3); its shell leaves the index. A Subject unclaimed: skipped by the read, kept in the record with a zero claim (AC-2.5). A Source renamed: the claims read takes the new title from Source records (AC-1.3).
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** A graph's Subjects and Observations are unbounded, as in the loaded store today; the records count graphs (4,096), and S16-01's per-graph cap bounds the claims the read returns (AC-1.4).

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache/cache.go` | D-1, D-3; `KeyedEntry` doc: "one value per 16-byte id (a handle's, or a graph's for entries whose Apply is `DropNamedGraphs`)"; `Cap`: "the most keys held; past it the least recently used leaves" |
| `core/database/graphcache/evidence_graph.go` | D-2: the store, hooks and memo API go; `SubjectLabels` (D-5); the row types and `WalkEvidenceGraph` stay |
| `core/database/graphcache/commit.go`, `graph.go`, `verify.go`, `cache.go` | D-1 (`capsForTest`), D-2, D-10: the vocabulary loop, `dropEvidenceGraphs` call, store fields and Verify loop go |
| `core/database/graphfill/load.go` | `loadGraphRows`; `loadEvidenceGraph` calls it after `loadGraphHead` |
| `core/database/graphfill/records.go` *(new)* | D-4, D-6, D-7; package doc gains "and each graph's Subject records" |
| `core/database/evidencelayer/evidencelayer.go` | D-8; `evidencelayertest.go` `Layer` drops its pattern argument |
| `core/database/promotealign/canon.go`, `promotealign.go` | D-9; the `evidencelayer.Load` call drops its pattern |
| `core/database/persondetail/claims.go`, `recordlinks.go`, `family.go` (the `dropUnread` comment: "as the claims read skips") | D-11, D-12 |

Errors: none new.

```go
// graphfill/records.go (sketch)
type SubjectRecord struct {
	Subject      graphcache.EvidenceSubject
	Accepted     graphcache.EvidenceMember          // zero: no accepted claim
	Observations []graphcache.EvidenceObservation  // ref order; nil for none
}
type GraphRecords struct {
	GraphID  []byte
	subjects []SubjectRecord
	byID     map[string]int
	labels   map[string]string
}
var Records graphcache.Keyed[*GraphRecords]
func RecordsOf(g *graphcache.Graph, graphIDs [][]byte) (map[string]*GraphRecords, error)

// persondetail/claims.go
type Claim struct {
	Graph evidencegraphs.Graph
	graphfill.SubjectRecord
	Records *graphfill.GraphRecords
}
```

### FFI

None — no proto change. `api/ffi/handlers/person_detail.go`: `ListPersonClaims` maps `cl.Accepted.ClaimID`, `cl.Graph.ID`, `.Ref`, `.Name` and passes `cl.Records.Labels()` to `rowsOf`; `recordLinksProto` maps `g.Graph.ID`, `.Ref`, `.Name`. `api/ffi/handlers/source_read.go`: D-13.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Measure the base; keyed entries evict the least recently used key**
Files: `core/database/graphcache/cache.go`, `core/database/graphcache/graph.go` (`capsForTest`), `core/database/graphcache/cache_test.go`, `core/database/graphcache/evidence_graph.go` (`SubjectLabels` only), `core/database/canonlinks/reader.go` (the `maxKeys` comment only).
Do: first, before any code change, measure the planning base on this machine: `git worktree add ../s16-13-base aeb2773` (outside the repo tree); write `claims_budget_test.go` (T-10's claims test) and the 80-graph `TestRecordLinksBudgets` case into its `core/database/persondetail/` (they call only `persondetail.Claims`, `RecordLinks`, `households` and `catalogsessiontest.Cold`, which exist there); run them and `BenchmarkCacheSizes`; record the figures for the ship log; set `budgetLinks80` per D-15; `git worktree remove ../s16-13-base`. Then D-1, D-3, D-5; replace `TestEntryCap` with T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/canonlinks/...` passes; the base figures are written down.

**STEP-2 — The two memos become entries**
Files: `core/database/graphfill/load.go` (export `LoadEvidenceGraph`; add `MaxEvidenceGraphs`), `core/database/evidencelayer/evidencelayer.go`, `evidencelayer_test.go`, `core/database/evidencelayer/evidencelayertest/evidencelayertest.go`, `core/database/promotealign/canon.go`, `promotealign.go`, `memo_test.go`, `core/database/persondetail/recordlinks.go` (the `Load` call only), `core/database/graphcache/evidence_graph.go`, `commit.go`, `core/database/graphcache/evidence_graph_memo_test.go` (deleted), `core/database/graphcache/evidence_graph_cap_test.go` (memo cases go), `core/database/graphfill/load_test.go`, `core/catalogsession/cache_doc_test.go` (blank-import `evidencelayer`), `docs/catalog-caches.md` (rows `evidencelayer.layers`, `promotealign.candidates`).
Do: D-8, D-9, D-10. The loaded-graph store stays in `graphcache` for this step, without memos (`loadedEvidenceGraph{graph, used}`); `graphfill.MaxEvidenceGraphs` and `graphcache.maxEvidenceGraphs` both read 64 until STEP-3. Memo assertions become entry assertions (`Layers.Peek`, `KeyCounts`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/evidencelayer/... ./core/database/promotealign/... ./core/database/promote/... ./core/database/graphfill/... ./core/database/persondetail/... ./core/catalogsession/...` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes; `rg -n 'EvidenceGraphMemo|SetEvidenceGraphMemo|memos' core/database/graphcache -g '!*_test.go'` finds nothing.

**STEP-3 — The loaded-graph store moves into `graphfill` as a keyed entry**
Files: `core/database/graphcache/evidence_graph.go`, `graph.go`, `commit.go`, `verify.go`, `core/database/graphcache/evidence_graph_cap_test.go` (moved to `core/database/graphfill/graphs_test.go` *(new)*), `core/database/graphfill/load.go`, the ten callers in §4 (`core/database/evidencelayer/evidencelayer.go`, `core/database/promotealign/promotealign.go`, `core/database/persondetail/claims.go`, `api/ffi/handlers/subjects.go`, `citations.go`, `observations.go`, `promote.go`, `event_titles.go`), tests that called the removed `*ForTest` hooks or `g.EvidenceGraph`, `docs/catalog-caches.md` (row `graphfill.graphs`; the typed-field row goes).
Do: D-2. `graphcache.maxEvidenceGraphs` goes; every Cap reads `graphfill.MaxEvidenceGraphs`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; T-1b passes; AC-4.6's full `rg` finds nothing.

**STEP-4 — Per-graph records**
Files: `core/database/graphfill/load.go`, `core/database/graphfill/records.go` *(new)*, `core/database/graphfill/records_test.go` *(new, package `graphfill_test`)*, `core/catalogsession/cache_doc_test.go` (blank-import `graphfill`), `docs/catalog-caches.md` (row `graphfill.records`). Skill: `add-catalog-query`.
Do: D-4, D-6, D-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/... ./core/catalogsession/...` passes, T-2 – T-4 included.

**STEP-5 — The claims read composes from the records**
Files: `core/database/persondetail/claims.go`, `family.go` (comment), `claims_test.go`, `caps_test.go`.
Do: D-11. `TestClaimsAfterSourceEdit`'s "the loaded graph stays loaded" check still holds: record links load the graph through `Layers`' fill. Test edits are field paths only (`Graph.GraphName` → `Graph.Name`, `GraphID` → `ID`, `GraphRef` → `Ref`, `ClaimID` → `Accepted.ClaimID`); expectations stay.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes; T-7, T-8 pass.

**STEP-6 — Record links and the handlers**
Files: `core/database/persondetail/recordlinks.go`, `recordlinks_test.go`, `recordlinks_perf_test.go`, `api/ffi/handlers/person_detail.go`, `source_read.go`, `person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-12, D-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/...` passes; T-9 passes.

**STEP-7 — Budgets and sizes**
Files: `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` (`ScaleSeed`), `core/database/evidencegraphs/index_bench_test.go`, `core/database/graphfill/records_bench_test.go` *(new)*, `core/database/persondetail/claims_budget_test.go` *(new)*, `core/database/persondetail/recordlinks_perf_test.go`, `core/database/promotealign/bench_test.go` (`BenchmarkCacheSizes` gains `graphRecords`: `RecordsOf` over every graph in `IndexOf(g).All()`).
Do: D-14, D-15, with STEP-1's base figures. In the ship log compare: kinship as is; "evidence" before (store with memos) against after (`graphfill.graphs` + `evidencelayer.layers` + `promotealign.candidates`, summed); `graphRecords` after only.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestClaimsBudgets|TestRecordLinksBudgets' ./core/database/persondetail/` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkGraphRecords|BenchmarkGraphIndex' ./core/database/graphfill/ ./core/database/evidencegraphs/` passes D-15; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkCacheSizes ./core/database/promotealign/` runs.

**STEP-8 — Docs**
Files: `docs/catalog-caches.md` (the four rows with sizes from STEP-7; the paragraph after the table gains "Past its Cap a keyed entry drops its least recently used key."; the known-exceptions paragraph loses the vocabulary-loop sentence; the typed-fields intro loses "(S16-13 and later PRs revisit the Evidence-graph store)"), `core/database/effects/README.md` (line 9: "graphcache drops each, and nothing else follows it" → "every registry entry keyed by graph drops each through `graphcache.DropNamedGraphs`: the loaded graphs, their records, layers and candidates"), `docs/deployment-plan/spike-16/cache-architecture.md` (§ Entity records S16-13 bullet: the measured size and time replace "est."; the loaded store is an entry), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-guidance-names.py` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Fixtures: `evidencelayertest.New(t)` / `Records` (`NewSource`, `Person`, `Relationship`, `Graph`, `Run`), `promotetest.Save` for an accepted claim, `households` for the 80-graph person.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/cache_test.go` `TestKeyedEvictsLeastRecentlyUsed` *(replaces `TestEntryCap`)* | Table on a cap-2 entry: Get-hit then Set past cap; Peek then Set; GetMany past cap; Check and KeyCounts then Set; `SetCapForTest(g, 1)` evicts at once and survives `g.Drop()` → the held keys of AC-3.1 – AC-3.4 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-1 |
| T-1b | Go | `core/database/graphfill/graphs_test.go` *(new)* `TestEvidenceGraphCap` *(moved from `core/database/graphcache/evidence_graph_cap_test.go`; memo cases go)* | Oldest evicted; a read is a use; a drop clears, with the cap override kept → unchanged | AC-3.5, D-2 |
| T-2 | Go | `core/database/graphfill/records_test.go` `TestGraphRecordsFill` *(new)* | Table: graph loaded / not → equal records, loaded fill issues no query; claimed and unclaimed Subjects with a `Relationship` → `Subject`, `Labels`, `Entity` per AC-2.5; unknown id → empty records, stored | AC-2.1, AC-2.5, AC-2.6, D-6 |
| T-3 | Go | same file `TestGraphRecordsFollowCommits` *(new)* | Table: Observation added on A (through `Run`) with `PROVENENCIA_GRAPH_VERIFY=1` → A dropped, B held, refill has it; graph A deleted → A dropped; `Drop` → none held | AC-2.3, AC-2.4, D-3, D-7 |
| T-4 | Go | same file `TestGraphRecordsOutliveLoadedGraph` *(new)* | Store cap 1, load B after A → A's records held | AC-2.2 |
| T-5 | Go | `core/database/evidencelayer/evidencelayer_test.go` `TestLayersEntry` *(new; the old memo check moves here)* | Table: vocabulary commit → none held; `nameformats.Invalidate` → none held; write naming A → A gone, B held; A held and graph evicted → `Load` loads no graph; unknown id twice → `ErrNoGraph`, none held | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, D-8 |
| T-6 | Go | `core/database/promotealign/memo_test.go` `TestMemoFollowsCommit`, `TestCandidatesMemoFollowsCommits`, `TestLoadSharesPromotesMemo` *(rewritten to the entries; each gains a vocabulary-commit and a name-format case)* | Propose stores a layer and candidates; vocabulary commit, `Invalidate` and a write naming the graph each empty them; Verify on after a proposal passes | AC-4.1, AC-4.2, AC-4.3, AC-4.8, D-9 |
| T-7 | Go | `core/database/persondetail/claims_test.go` `TestClaimsWarmReadIssuesNoQuery` *(gains an 80-graph case)* | `households(f, 80)`, Drop, Claims → 80 groups, `LoadedEvidenceGraphCountForTest` 0; second read → 0 queries | AC-1.1, AC-1.2 |
| T-8 | Go | same file `TestClaimsSkipsWhatLeft` *(new)* | Table, after a warm read of a person in two graphs A and B: an empty graph's records (`src := f.source("Empty")`, `graphfill.RecordsOf(g, [][]byte{evidencegraphstest.SoleGraph(t, f.C, src.ID).ID})`) set as A's with `graphfill.Records.Set(g, A, r)` → A's claim skipped, B's kept; `Records.Set(g, A, B's records)` (A's Subject absent) → skipped. The missing-Source error keeps today's code (a Source with graphs cannot be deleted) | D-11 |
| T-9 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims` *(the Person gains a `Relationship` to a second Subject)* | `value_subject_label` is that Subject's label; the rest as before | AC-1.6, D-13 |
| T-10 | Go | `core/database/persondetail/claims_budget_test.go` `TestClaimsBudgets` *(new)*; `recordlinks_perf_test.go` `TestRecordLinksBudgets` *(gains 80)* | `households(f, 80)`: claims cold ≤ 1 s, warm ≤ 50 ms; record links ≤ `budgetLinks80` | AC-1.5, AC-1.7, D-15 |
| T-11 | Go bench | `core/database/graphfill/records_bench_test.go` `BenchmarkGraphRecords` *(new)* | `ScaleSeed.Base`, Drop, `RecordsOf` every graph → `build-ms`, `heap-MB`; fails past 2 s or 35 MB | D-14, D-15 |
| T-12 | Go | existing `TestClaims*`, `TestCraftedClaims`, `TestRecordLinks*`, `TestGraphWriteDropsOnlyThatGraph`, `TestRetitleKeepsEvidenceGraph`, `TestCacheDocListsEveryEntry` | Unchanged expectations (field paths and memo → entry assertions only) | AC-1.3, AC-1.4, AC-4.7 |
| T-13 | Sweep | STEP-2's and STEP-3's `rg` (AC-4.6) | No hit | AC-4.6, D-10 |
| T-14 | Go | `core/database/graphfill/records_test.go` `TestVerifyLeavesEntriesAlone` *(new)* | Records, layers, candidates and loaded graphs held at a full `SetCapForTest` cap; a write through `Run` with Verify on, and the same with Verify off → the same held keys in the same order (compared by `Peek`-free reads of `KeyCounts` and a sequence of `Set`s that reveals the next eviction) | AC-4.9, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs (its Xcode half runs on macOS; on Linux run its Go and doc checks and say so). Most likely to break: `./core/database/promotealign/...`, `./core/database/graphcache/...`, `./core/catalogsession/...`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/` comments and `L10n.swift` with no hit left (`EvidenceGraphMemo`, `SetEvidenceGraphMemo`, `memberGraphs`, `candidatesMemoKey`, `maxEvidenceGraphs`, `evictEvidenceGraphs`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (done at planning; confirm the map text still matches)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none: none
- [ ] Ship log: TestClaimsBudgets and the 80-graph TestRecordLinksBudgets at the planning base and after; BenchmarkGraphRecords; BenchmarkCacheSizes per STEP-7
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated): not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md (four rows, the Cap sentence, the store row gone), core/database/effects/README.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person cited in more than 64 Evidence graphs. Open the person's Sources section. See every Source group and claim as before.
2. Leave and reopen the person. The groups draw at once.
3. Rename one of those Sources from its Source page. Back on the person, the group shows the new title.
4. Add an Observation to one cited Subject in its Evidence graph. Back on the person, the claim shows it.
5. Run one Promote proposal on a Source before and after a vocabulary edit (add a term): suggestions are the same as before the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| LRU bookkeeping slows the kinship reads | O(1) per use; `BenchmarkCacheSizes` and `TestCraftedFamily` before and after |
| Moving the store into `graphfill` changes Promote's loading pattern | Same cap, same use rule (reads and Peeks count); T-1b; AC-4.7 |
| A memo outliving its loaded graph changes Promote's output | Same drop signals plus the name format; the generation check stays; Checks under Verify; AC-4.7 |
| Records filled from a loaded graph differ from SQL | One constructor; T-2 and the Check under Verify |

**Stop and report if:**

- A row of §4 is no longer true at your base (in particular `EvidenceGraphMemo` gained a caller, `Keyed` gained a method, `evidencelayer.Load` gained a caller passing a pattern not from `valuerules.Load`, or `persondetail.Claims` gained a caller outside the handlers).
- AC-4.7: any Promote suite or matching golden changes output.
- The candidates Check (D-9) cannot rebuild the stored candidates without a second ranking path beside `rankCandidates`, or the layers Check fails on a graph no write touched.
- `TestClaimsBudgets` warm is above 50 ms or cold above 1 s; the 80-graph record links read is above `budgetLinks80` or slower than the planning base by more than 25%.
- `BenchmarkGraphRecords` passes 2 s or 35 MB, or `BenchmarkCacheSizes`' kinship figure grows by more than 20%.
- `g.Verify()` reports any of the four entries mismatched after a write that went through `writes.Run`: name the write and its `Set.EvidenceGraphs`.
- An import cycle: `persondetail`, `evidencelayer`, `promotealign` or `api/ffi/handlers` → `graphfill`; `evidencelayer` → `nameformats`; `evidencegraphstest` → `graphcachetest`; or `catalogsession`'s test → `evidencelayer`/`graphfill`.
- A test outside §8's list needs more than a field path or a memo → entry assertion to pass.

## 12. Open questions

None. Answered 2026-11-13: the Citation fields join with S16-30; the records, the loaded graphs and both memos are registry entries keyed by graph, with least-recently-used eviction added to the registry, in one PR; "simpler code" means simple reads, with construction in the cache layer.

## 13. PR

**Title:** S16-13 — Per-graph records in the cache registry, and the Sources section reads them
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; TestClaimsBudgets, TestRecordLinksBudgets, BenchmarkGraphRecords and BenchmarkCacheSizes figures; `Plan: docs/deployment-plan/spike-16/prs/S16-13-records-tier.md`.
