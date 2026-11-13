# S16-13 — Per-graph records in the cache registry, and the Sources section reads them

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Sources section shows the same groups, claims and rows; only where the engine reads them from moves)
**Board:** —
**Depends on:** S16-12 (landed [#800](https://github.com/mendahu/provenencia/pull/800), Source records), S16-15 (landed [#805](https://github.com/mendahu/provenencia/pull/805), the graph index's shells); S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), the registry this PR extends)
**Blocks:** S16-30 (a provenance record composes the Subject record, and adds each Observation's Citation words to it), S16-32 (re-ranks inside the candidates entry this PR registers), S16-24 (the claims read reports its inputs), S15-05, S15-22, S15-24
**Branch:** `s16-13-records-tier`
**Version bump:** none — no change a researcher can see; the wire is unchanged ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/match-rules.mdc` (the candidates memo is matching-layer state; no knob changes)
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); audit P-4 ([`person-page-audit.md`](../person-page-audit.md)); working rules: "(Batch 2) The point of this batch is simpler code" (as clarified 2026-11-13), "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Per-graph records and memos are registry entries keyed by graph" (K-18, reworded 2026-11-13), "(Batch 2) A Source's words live once", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) Orders are defined and results do not move", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth", "(Batch 2) Batch 1's caps stay where S16-01 puts them"; goal bars 11 and 12
**Planned:** 2026-11-13 against `main` @ `aeb2773`
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

The Person page's Sources section stops loading whole Evidence graphs. A person in 80 graphs today loads all 80 and, because only 64 stay loaded, reloads them on every visit (audit P-4, goal bar 12). After this PR each graph's Subject records (each Subject, its accepted claim, its Observations, and the graph's Subject labels) are held in a cache entry of their own, and the claims read is a few lookups: the node's members, the graph index's shell, the Source record, the graph's records. The cache layer gets the plumbing that makes this possible: every keyed registry entry evicts its least recently used key at its cap, and the two memos that today hang off a loaded graph (the matching layer, Promote's candidates) become typed registry entries too, so the Evidence-graph store holds only loaded graphs.

**Why now.** S16-12 (Source records) and S16-15 (graph shells) landed, so the claims read can compose from cached records. The user's clarification of "simpler code" (2026-11-13): more cache code is welcome when it makes the reads simple; the construction belongs in the cache layer, not in the readers. The registry gets least-recently-used eviction once, here, instead of a hand-written second store (K-18 reworded, user 2026-11-13). The user chose one PR over a preparatory split (2026-11-13).

**What stays.** The wire, every response's content and order, S16-01's records-per-graph cap (`boundedMembers`), the record links' lines and bounds (record links still build each graph's matching layer from the whole loaded graph), Promote's suggestions, the 64-graph limit of the loaded-graph store, and the candidates memo's generation check (S16-32 replaces it). Each Observation's Citation label, page, Artifact id and uncertainty flag join the records with S16-30, their first reader (user, 2026-11-13).

## 2. User stories

**US-1.** As a researcher, when I open the Sources section of a person cited in 80 Evidence graphs, it opens without loading any whole graph, and opens again from memory, so that the section is fast and stays fast past 64 graphs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `households(f, 80)` (80 Sources, one graph each, the son promoted onto one entity); `g.Drop()`, then `persondetail.Claims(son)` | 80 groups of one claim each; `g.LoadedEvidenceGraphCountForTest()` is 0 |
| AC-1.2 | The same, read a second time | Zero queries (`SetQueryCounterForTest`); the same 80 groups |
| AC-1.3 | The fixtures of `TestClaimsGroupsBySource`, `TestClaimsOrderSources`, `TestClaimsMissingAndEmpty`, `TestClaimsAfterSourceEdit` | The same groups, claims, Subjects, Observations and order as at the planning base (field names follow D-9) |
| AC-1.4 | `TestClaimsRecordsPerGraph`, `TestCraftedClaims` (20,000 records in one graph) | Unchanged results; the crafted read within `capBudget` (1 s) |
| AC-1.5 | `households(f, 80)`: cold read (`catalogsessiontest.Cold`), then warm read | Cold within 1 s, warm within 50 ms; the planning-base figures on the same machine are in the ship log |
| AC-1.6 | `ListPersonClaims` over the handler test's Person with a claim on two graphs, one Observation's value a second Subject | The response equals the planning base's, `value_subject_label` included |

**US-2.** As the engine, I keep each Evidence graph's Subject records exactly, so that a reader composes from them without loading the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A graph read through `graphfill.RecordsOf` while it is loaded, and the same graph read through it after `Records.Drop` and the loaded graph evicted | The two `*GraphRecords` are `reflect.DeepEqual`; the first issues no query |
| AC-2.2 | Records held for graph A; loaded-graph cap set to 1 and graph B loaded (A evicted) | `Records.Peek(g, A)` still holds A's records |
| AC-2.3 | Records held for graphs A and B; a write adds an Observation on A (`Set.EvidenceGraphs` = [A]) | A's records are gone, B's are held; the next `RecordsOf(A)` has the new Observation; `PROVENENCIA_GRAPH_VERIFY=1` passes |
| AC-2.4 | Records held; `g.Drop()` | `KeyCounts()["graphfill.records"]` is 0 |
| AC-2.5 | A Subject with an accepted claim, a Subject without one, and an Observation whose value is another Subject | `Subject(id)` returns the claim for the first and a zero `Claim` for the second; `Labels()` maps both ids to their labels; `Entity(id)` is the claimed entity, nil for the second |

**US-3.** As the engine, every keyed cache entry keeps its most recently used keys at its cap, so that a reader past the cap still gets its value stored.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Cap 2: `Set` A, `Set` B, `Get` A (hit), `Set` C | A and C held, B gone; `KeyCounts` 2 |
| AC-3.2 | Cap 2, A and B held: `Peek` A, then `Set` C | A held (a Peek is a use), B gone |
| AC-3.3 | Cap 2, A held: `GetMany` [A, B, C] filling B and C | Returns A, B and C; holds B and C (the two most recent); A gone |
| AC-3.4 | `Check` over a full entry, and `KeyCounts` | Neither changes which key is evicted next |

**US-4.** As the engine, the matching layer and Promote's candidates are typed registry entries keyed by graph, so that no reader casts a memo and no commit loop clears them by hand.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A layer and candidates held for graph A; a vocabulary commit (`ListPropertyVocabulary`) | Both entries hold no key |
| AC-4.2 | Layers and candidates held for A and B; a write that names A | A's are gone, B's are held |
| AC-4.3 | A layer held for A; the loaded graph A evicted; `evidencelayer.Load(g, A, pattern)` | Returns the held layer and loads no Evidence graph (`LoadedEvidenceGraphCountForTest` unchanged) |
| AC-4.4 | `evidencelayer.Load` for an unknown graph id, twice | `ErrNoGraph` both times; `Layers` holds no key |
| AC-4.5 | A layer held under name pattern P; `Load` under pattern Q | Builds and stores the Q layer in A's place |
| AC-4.6 | `rg -n 'EvidenceGraphMemo\|SetEvidenceGraphMemo\|memos' core/database/graphcache/*.go` (non-test) | Nothing; `OnCommit`'s vocabulary branch has no loop |
| AC-4.7 | Promote's suites (`./core/database/promotealign/...`, `./core/database/promote/...`, the matching goldens) | Green, unchanged |

## 3. Scope

| In | Covered by |
| --- | --- |
| Keyed registry entries evict the least recently used key at their Cap; Get, GetMany, Peek and Set count as a use | STEP-1, AC-3.1 – AC-3.4 |
| `graphcache.DropNamedGraphs` *(new)*: the one Apply of an entry keyed by Evidence graph | STEP-1, AC-2.3, AC-4.2 |
| The matching layer becomes `evidencelayer.Layers` *(new)*, a keyed entry by graph id | STEP-2, AC-4.1 – AC-4.5 |
| Promote's candidates become `promotealign` `candidates` *(new)*, a keyed entry by graph id, keeping the generation check | STEP-2, AC-4.1, AC-4.2, AC-4.7 |
| The loaded-graph store loses its memos: `loadedEvidenceGraph.memos`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo` and the vocabulary loop in `OnCommit` go | STEP-2, AC-4.6 |
| `graphfill` splits its Subject, Observation and member queries into one step both fills call | STEP-3, AC-2.1 |
| `graphfill.Records` *(new)*: each graph's `GraphRecords`, filled from the loaded graph when it is held, else from the split step; Check is graphfill's | STEP-3, AC-2.1 – AC-2.5 |
| The claims read composes from the node's members, the graph index, Source records and the records entry; `memberGraphs` goes | STEP-4, AC-1.1 – AC-1.5 |
| The record links read takes its graphs and claimed entities from the same composition | STEP-5, US-1 (record links tests) |
| `ListPersonClaims` builds each graph's Subject label map once, from the records | STEP-5, AC-1.6 |
| Budgets and sizes: claims at 80 graphs; records at 10,000 graphs and 50,000 Observations; `BenchmarkCacheSizes` gains the three entries | STEP-6, AC-1.5 |
| Docs: `catalog-caches.md`, the design doc, the map, the effects README sentence on `Set.EvidenceGraphs` | STEP-7 |

| Out | Home |
| --- | --- |
| Each Observation's Citation label, page, Artifact id and uncertainty flag in the records | S16-30 |
| The provenance sheet and the detail read composing from the records | S16-30 |
| Candidates re-ranking only what moved; the generation check going | S16-32 |
| Record links building layers without the whole loaded graph; the layer cap above 64 | not planned ([`structural-debt.md`](../../../ideas/structural-debt.md) batch graph read) |
| A batch fill of several graphs' records in one query set | [`structural-debt.md`](../../../ideas/structural-debt.md) (S13-16 PRF-2) |
| The claims read reporting its inputs | S16-24 |
| Raising `maxEvidenceGraphs` | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Loaded-graph store | `core/database/graphcache/evidence_graph.go` `loadedEvidenceGraph`, `EvidenceGraph`, `evictEvidenceGraphs`, `dropEvidenceGraphs`, `maxEvidenceGraphs` (64) | Holds whole graphs, LRU by a tick scan; each holds `memos map[string]any` |
| Memo API | same file `EvidenceGraphMemo`, `SetEvidenceGraphMemo` | Untyped memos keyed by string, alive while the graph is loaded; callers cast |
| Vocabulary loop | `core/database/graphcache/commit.go` `OnCommit`, the `ListPropertyVocabulary` branch | Sets every loaded graph's `memos = nil`: the store's one hand-written drop (S16-10 D-15 gave it to this PR) |
| Store fields | `core/database/graphcache/graph.go` `Graph.evidenceGraphs`, `clear`, `LoadedEvidenceGraphForTest`, `SetEvidenceGraphCapForTest`, `LoadedEvidenceGraphCountForTest` | The store and its test hooks; no read of a held graph that does not count as a use |
| Store Verify | `core/database/graphcache/verify.go` `Verify` | Runs `evidenceGraphFn` over every loaded graph |
| Registry, keyed | `core/database/graphcache/cache.go` `KeyedEntry`, `Keyed[T].data`, `store`, `Get`, `GetMany`, `Peek`, `Set`, `Drop`, `Clear`, `RegisterKeyed` | Values in `map[string]T` under `g.entries[id]`; `store` refuses a new key once the map holds `Cap` keys |
| Keyed users | `core/database/canonlinks/reader.go` `pivots`, `forAnswers`, `siblingAnswers` (Cap `maxKeys` 65,536); `Reader.pivot` uses `Peek` then `Get` | The only keyed entries; real data stays under the cap |
| Matching layer memo | `core/database/evidencelayer/evidencelayer.go` `Load`, `memoKey` | Loads the whole graph, then looks up memo `"evidencelayer\x00"+namePattern`, builds and stores on a miss; `ErrNoGraph` for a graph that is not `Found` |
| Candidates memo | `core/database/promotealign/canon.go` `loadCandidatesMemo`, `candidatesMemoKey`, `candidatesMemo{gen, bySubject}` | Memo `"promotealign.candidates\x00"+namePattern`, used only when `gen` matches the shortlist index's generation; never changed in place |
| Graph loader | `core/database/graphfill/load.go` `loadEvidenceGraph`, `loadGraphHead`, `loadGraphSubjects`, `loadGraphObservations`, `loadGraphMembers`, `loadTitleEdges`, `checkEvidenceGraph` | One loader; Subjects by label then ref, Observations by ref, members (accepted claims) by Subject ref |
| Claims read | `core/database/persondetail/claims.go` `Claims`, `claims`, `boundedMembers`, `memberGraph`, `memberGraphs` | Loads each member's whole graph (`c.Graph().EvidenceGraph`), finds the member and Subject by linear scan, and scans all of the graph's Observations per claim; groups by `graph.SourceID`; sorts groups by Source title and claims by graph name (`conclusionsummaries.SortByTitle`) |
| Claim type | same file `Claim{ClaimID, Graph graphcache.EvidenceGraph, Subject, Observations}` | Carries the whole loaded graph |
| Record links | `core/database/persondetail/recordlinks.go` `RecordLinks`, `recordLinks`, `SourcedGraph{Graph graphcache.EvidenceGraph, Source}`, `linksBetween` | Takes graphs from `claims`; `linksBetween` maps Subject → entity from `graph.Members` and builds layers with `evidencelayer.Load` |
| Handlers | `api/ffi/handlers/person_detail.go` `ListPersonClaims`, `recordLinksProto`; `api/ffi/handlers/source_read.go` `listedContext.rows`, `rowsOf`, `listedProtosFromGraph` | `rowsOf(eg, obs)` rebuilds a Subject label map from `eg.Subjects` on every call, once per claim |
| Graph shells | `core/database/evidencegraphs/index.go` `IndexOf`, `GraphIndex.Graph`; `evidencegraphs.go` `Graph{ID, Ref, SourceID, Name, Description, ArtifactIDs}` | Every graph's row, held and patched (S16-15) |
| Source records | `core/database/sources/records.go` `RecordsOf`, `RecordIndex.Source` | Every Source's words (S16-12) |
| Node members | `core/database/graphcache/graph.go` `Member{SubjectID, SourceID, GraphID, Accepted}`, `Node.AcceptedMembers` | No claim id; the claim comes from the graph |
| Effects | `core/database/effects/registry.go` `Set.EvidenceGraphs`; entries `subjects`, `observations`, `citations`, value columns, `identity_claims`, `evidence_graphs` name graphs | The exact signal for anything a graph's rows hold |
| Cache doc test | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry`, `TestCacheDocSeesEveryOwner` and its blank imports | Imports every owner; a new owner package must join |
| Tests | `core/database/graphcache/cache_test.go` `TestEntryCap`; `evidence_graph_memo_test.go` `TestEvidenceGraphMemoDroppedWithSource`; `evidence_graph_cap_test.go` (memo steps); `core/database/graphfill/load_test.go` `TestRetitleKeepsEvidenceGraph`, `TestGraphWriteDropsOnlyThatGraph` (memo lines); `core/database/evidencelayer/evidencelayer_test.go` (line 93 memo check); `core/database/promotealign/memo_test.go` `TestLoadSharesPromotesMemo`; `core/database/persondetail/claims_test.go`, `caps_test.go` `TestCraftedClaims`, `recordlinks_test.go`, `recordlinks_perf_test.go` (`households`, `TestRecordLinksBudgets`); `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims` and the `SourcedGraph` literal near line 440 | Cover today's behavior; the memo tests read the API this PR removes |
| Benchmarks | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes`; `core/database/evidencegraphs/index_bench_test.go` `BenchmarkGraphIndex` (inline seed of 3,000 Sources, 10,000 graphs, 20,000 Subjects, 50,000 Observations) | Sizes per entry; the Scale seed lives inside one benchmark |

**Precedent.** S16-12 ([#800](https://github.com/mendahu/provenencia/pull/800)) and S16-15 ([#805](https://github.com/mendahu/provenencia/pull/805)): an owner package registers an entry from `init`, exports a typed read (`RecordsOf`, `IndexOf`), its Check rebuilds from SQL and `reflect.DeepEqual`s, and the claims read composes from it. Copy the owner-registers-and-exports shape and the Check. Do not copy the whole-entry fill: this entry is keyed, filled through `GetMany`. `canonlinks`' `pivots` is the precedent keyed entry (Apply, Check, Cap).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A keyed entry evicts its least recently used key when a new key arrives at `Cap`. Each keyed entry's data is a map from key to a `container/list` element plus one list in use order. `Get` and `GetMany` hits, `Peek` and `Set`/fill stores move a key to the front; `Check`, `KeyCounts` and `Drop` do not | One eviction rule in the registry serves the records, both memos and the kinship pivots; O(1) per use; `Peek` is a reader's lookup in `Reader.pivot` | Refusing past the cap (today) — a reader past it recomputes on every read; a tick scan like `evictEvidenceGraphs` — O(Cap) per store, 65,536 for `pivots` |
| D-2 | The records are a keyed entry, `graphfill.Records graphcache.Keyed[*GraphRecords]`, keyed by graph id, `Name: "graphfill.records"`, `Cap: maxGraphRecords` = 4,096 with `// decided: user, 2026-11-13` | The registry carries cap, Verify, the commit drop, `Drop` and sizing (K-18 reworded, user 2026-11-13); 4,096 graphs is 40% of the batch's 10,000 graphs and bounds the entry near 14 MB at the design's 25–35 MB for all of them | A second map in the loaded-graph store (K-18's first wording) — a second hand-written cap, eviction, Verify loop and drop |
| D-3 | Apply of every entry keyed by Evidence graph is `graphcache.DropNamedGraphs[T]` *(new)*: `k.Drop(g, set.EvidenceGraphs)` | `Set.EvidenceGraphs` is the exact signal the loaded-graph store already drops on; one function for three entries | Each owner writing the same two lines |
| D-4 | `GraphRecords` (new, in `graphfill`) holds `GraphID`, the graph's `SubjectRecord`s in the loader's Subject order, an index by Subject id, and the Subject id → label map. `SubjectRecord{Subject graphcache.EvidenceSubject; Claim graphcache.EvidenceMember; Observations []graphcache.EvidenceObservation}`: the Subject's accepted claim (zero when none; the catalog allows one per Subject) and its Observations in ref order. Methods: `Subject(id) (SubjectRecord, bool)`, `Labels() map[string]string`, `Entity(subjectID) []byte` | It is what the claims read, the handler and record links ask of a graph, built once; it reuses graphcache's row types | A new Subject or Observation type — duplicates `EvidenceSubject`/`EvidenceObservation`; Citation words now — no reader until S16-30 (user, 2026-11-13) |
| D-5 | One constructor `newGraphRecords(graphID, subjects, observations, members)` builds records. The fill (`graphfill.RecordsOf(g, graphIDs [][]byte) (map[string]*GraphRecords, error)`, through `Records.GetMany`) takes each missing graph from `g.HeldEvidenceGraph(id)` *(new)* when the store holds it, else from `loadGraphRows(q, id)` *(new)*, the step split out of `loadEvidenceGraph`, which calls it too | One loader of Subject, Observation and member rows; a graph the user just looked at costs no SQL | Always SQL — re-queries a graph already in memory; deriving records inside `EvidenceGraph` — fills records for every Promote load that never reads them |
| D-6 | `HeldEvidenceGraph(graphID) (EvidenceGraph, bool)` *(new)* on `graphcache.Graph` returns a loaded graph without loading and without counting a use | The records fill must not reorder the store's LRU nor load | Using `EvidenceGraph` — loads on a miss, which is what this PR removes |
| D-7 | A graph id with no row fills empty records (no Subjects) and stores them; `Set.EvidenceGraphs` names a deleted graph (the `evidence_graphs` entry's `self()`), so they leave with it | Same as the loaded store's `Found: false` graph; no second state | An error — a graph deleted mid-read would fail the page |
| D-8 | `graphfill`'s Check for the entry rebuilds each held graph's records from `loadGraphRows` and `reflect.DeepEqual`s them | The same check shape as `checkEvidenceGraph`; AC-2.1 holds both fill paths to it | No Check — the registry's Verify is the batch's guard |
| D-9 | The claims read: `boundedMembers(n)`; `evidencegraphs.IndexOf`, `sources.RecordsOf`, `graphfill.RecordsOf` over the members' graph ids; per member, the shell `idx.Graph(m.GraphID)`, the Source record of `shell.SourceID`, and `records.Subject(m.SubjectID)`; a member is skipped when the shell is missing, the Subject is missing, or its claim's `EntityID` is not `n.ID` (today's skip); groups and orders as today. `Claim` becomes `{ClaimID []byte; Graph evidencegraphs.Graph; Subject graphcache.EvidenceSubject; Observations []graphcache.EvidenceObservation; Records *graphfill.GraphRecords}` | The read is lookups; the shell carries ref, name and Source; `Records` is shared, read-only, for labels and record links; S16-01's cap stays the first line | Keeping `Graph graphcache.EvidenceGraph` — needs the whole graph; caching the per-person answer — a stored composition with no measurement (rule "No new persisted projections, and no stored answer without a measurement") |
| D-10 | `SourcedGraph` becomes `{Graph evidencegraphs.Graph; Source sources.Record}`; `recordLinks` collects each distinct graph's `*GraphRecords` from the claims, and `linksBetween(c, page, records []*graphfill.GraphRecords, cfg)` builds each layer with `evidencelayer.Load(g, r.GraphID, namePattern)` and each member's entity with `r.Entity` | The claimed entities come from the same composition; the layer still compares the whole graph (Out) | Passing loaded graphs — the reason for this PR |
| D-11 | `listedContext.rowsOf(labels map[string]string, obs)`; `rows(eg, keep)` builds `labels` from `eg.Subjects` once; `ListPersonClaims` passes `cl.Records.Labels()` | The label map is built once per graph, at fill (scope: "the handlers build each graph's label map once") | A per-claim map, as today |
| D-12 | `evidencelayer.Layers graphcache.Keyed[patternLayer]` *(new)*, `Name: "evidencelayer.layers"`, `DerivedFrom: Vocabulary{}`, `Apply: DropNamedGraphs`, `Cap: maxLayers` = 64 (`// decided: user, 2026-11-13`, the loaded-graph store's size), no Check. `patternLayer{pattern string; layer Layer}`. `Load` asks `Layers.Get` first; a held value with another pattern is a miss whose fill replaces it; the fill loads the graph and returns `ErrNoGraph` (stores nothing) for a graph that is not `Found` | The layer reads the vocabulary and `canonlinks.LoadTerms`, as the loop dropped it; 64 keeps today's memory; a held layer needs no graph load (AC-4.3) | `DerivedFrom` `nameformats.Default` — today the pattern is in the key and the alias index and name format do not drop memos (S16-10 D-15); keyed by graph and pattern — the registry keys are 16-byte ids |
| D-13 | `promotealign` `candidates graphcache.Keyed[candidatesMemo]` *(new)*, `Name: "promotealign.candidates"`, `DerivedFrom: Vocabulary{}`, `Apply: DropNamedGraphs`, `Cap: 64`, no Check. `candidatesMemo` gains `pattern string`; a held value with another pattern or generation is unknown, as today. `loadCandidatesMemo` reads with `Peek` and stores with `Set` | Same signals as today's memo; S16-32 changes the re-rank inside this entry instead of adding a slot | `DerivedFrom` `MatchIndex` — `MatchIndex.Set` runs as profiles fill, which would empty candidates during a proposal |
| D-14 | The loaded-graph store keeps only graphs: `loadedEvidenceGraph{graph, used}`, `EvidenceGraphMemo`, `SetEvidenceGraphMemo` and `OnCommit`'s vocabulary loop go | Nothing else stores memos after D-12 and D-13 | Keeping the API for later — a second way to cache per graph |
| D-15 | `evidencegraphstest.SeedScale(tb, c)` *(new)* holds `BenchmarkGraphIndex`'s seed (3,000 Sources, 10,000 graphs, 20,000 Subjects, 50,000 Observations); `BenchmarkGraphIndex` and `BenchmarkGraphRecords` *(new)* call it | One Scale seed for the graph entries | A copy of the insert loop in `graphfill` |
| D-16 | Budgets: `TestClaimsBudgets` *(new)*: cold 1 s, warm 50 ms at 80 graphs; `BenchmarkGraphRecords` stops the build when filling every graph's records past 2 s or 35 MB | Rule "Measure growth"; 35 MB is the design's estimate ceiling | Unbudgeted |

**Rules this plan bends.**

- "(Batch 2) Per-graph records live in the Evidence-graph store" (K-18): the records tier and both memos are registry entries keyed by graph, not a tier inside the store; `dropEvidenceGraphs` drops loaded graphs only, and each entry drops on the same `Set.EvidenceGraphs`. User, 2026-11-13 ("Yes, one PR" to the registry shape, after asking that cache code carry the complexity). Updated at planning in the map (rule reworded), the design doc and the S16-32 scope entry.
- "(Batch 2) The point of this batch is simpler code": not bent; clarified by the user, 2026-11-13 (simple reads, with construction in the cache layer). Recorded in the rule at planning.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| LRU order in `Keyed[T]` | mechanism | `evictEvidenceGraphs` (tick scan), `store`'s refusal | Extends the registry in place; the scan is O(Cap) |
| `graphcache.DropNamedGraphs[T]` | func | `canonlinks.clearOnHandles` | Same shape, other signal; three entries share it |
| `(*Graph).HeldEvidenceGraph` | func | `LoadedEvidenceGraphForTest` | Test-only and returns no graph |
| `graphfill.GraphRecords`, `SubjectRecord`, `newGraphRecords` | type, func | `graphcache.EvidenceGraph` | The graph holds flat slices; records index them by Subject and carry the label map, without title edges |
| `graphfill.Records`, `RecordsOf`, `maxGraphRecords` | entry, func, const | `sources.Records`/`RecordsOf`, `evidencegraphs.Index`/`IndexOf` | Same owner shape; keyed, not whole |
| `graphfill.loadGraphRows` | func | `loadEvidenceGraph` | Split out of it; both call it |
| `graphfill` records Check | func | `checkEvidenceGraph` | Same shape over records |
| `evidencelayer.Layers`, `patternLayer`, `maxLayers` | entry, type, const | `memoKey` + `EvidenceGraphMemo` | Replaces it, typed |
| `promotealign` `candidates`, `candidatesMemo.pattern` | entry, field | `candidatesMemoKey` + `EvidenceGraphMemo` | Replaces it, typed |
| `evidencegraphstest.SeedScale` | helper | `BenchmarkGraphIndex`'s inline seed | Moved, not copied |
| `TestClaimsBudgets`, `BenchmarkGraphRecords`, `TestGraphRecords*`, `TestKeyedEvictsLeastRecentlyUsed`, `TestLayersEntry` | tests | `TestRecordLinksBudgets`, `BenchmarkGraphIndex`, `TestEntryCap` | New behavior |

**Builds on.** The cache registry (S16-10), the owner-registers shape of S16-12 and S16-15, `graphfill`'s loader registration, `catalogsessiontest.Cold`, `databasetest.HeapDelta`, the `households` fixture.
**Deletes.** `loadedEvidenceGraph.memos`; `Graph.EvidenceGraphMemo`, `SetEvidenceGraphMemo`; `OnCommit`'s memo loop and its comment; `evidencelayer.memoKey`; `promotealign.candidatesMemoKey`; `persondetail.memberGraph`, `memberGraphs`; `Claim.Graph`'s `graphcache.EvidenceGraph` type; `SourcedGraph.Graph`'s `graphcache.EvidenceGraph` type; `rowsOf`'s per-call label map; `TestEvidenceGraphMemoDroppedWithSource` and the memo steps of `core/database/graphcache/evidence_graph_cap_test.go` and `core/database/graphfill/load_test.go` (their cases move to the entries' tests); the "refuses past the cap" sentences (`store`'s comment, `canonlinks.maxKeys`'s comment, `TestEntryCap`); `catalog-caches.md`'s `memos` text and "S16-13 owns it".
**Next adds.** S16-30 adds Citation words to `SubjectRecord` and composes a provenance record from `Records`; S16-32 changes `candidates`' fill and Apply and drops the generation check; S16-24 reports the graphs a claims read used from the members it composed; S16-18's place entries get LRU from D-1 without code.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A keyed entry evicts instead of refusing | every `Keyed[T]` reader (`canonlinks`, records, layers, candidates) | A reader always gets its value (fill or held); an evicted key refills; T-1 |
| A held layer with no loaded graph | `evidencelayer.Load` callers (Promote, record links) | Same value as a fresh build; dropped by the same signals (T-5) |
| A candidates memo outliving its loaded graph | `loadCandidatesMemo` | Pattern and generation checked as today (T-6, AC-4.7) |
| Records of a deleted graph | the claims read | The shell is missing from the index, so the member is skipped (T-8) |
| A member whose Subject's claim names another entity (a merge mid-read) | the claims read | Skipped, as `memberGraphs` skips today (T-8) |

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; 3,000 Sources, 10,000 graphs, 50,000 Observations; a person in 50 and 80 graphs):

- **Claims read, warm:** per member, three map lookups; 80 graphs → 80 lookups each, no query (AC-1.2). **Cold:** one `loadGraphRows` per graph not held (three queries plus the date and name lookups), 80 graphs → about 400 queries once; no title-edge walks; budget 1 s (D-16).
- **Records entry:** 4,096 graphs at most; filling all 10,000 graphs measured by `BenchmarkGraphRecords` (stop past 2 s or 35 MB); the entry holds the most recent 4,096.
- **Apply per commit:** `DropNamedGraphs` on three entries, O(graphs named), usually one; a Source of 10,000 Observations in one graph names one graph.
- **LRU overhead:** one list element per key; `pivots` at 14,162 keys on the large fixture; `BenchmarkCacheSizes` kinship figure before and after in the ship log (Stop past +20%).
- **Layers and candidates:** 64 graphs each, the same bound as today's memos.

### Edges

- **Assumptions.** One accepted claim per Subject (`identity_claims_one_accepted_per_subject`), so `SubjectRecord.Claim` is one member; the loader's Subject and Observation order is the response order, held by AC-1.3. The registry's keys are 16 bytes; graph ids are.
- **Neighbors.** A key at exactly Cap stores and evicts one (T-1); Peek counts, Check does not (AC-3.2, AC-3.4); a layer under another pattern replaces (AC-4.5); a graph loaded vs not fills the same records (AC-2.1).
- **Removed inputs.** A graph deleted: its records, layer and candidates leave with `Set.EvidenceGraphs` (T-3); its shell leaves the index, so the claims read skips it. A Subject unclaimed: the read skips it; the record keeps it with a zero claim (AC-2.5). A Source renamed: the claims read takes the new title from Source records; records are untouched (AC-1.3).
- **Real data.** None — no table, seed or configuration drives a rule here.
- **Untrusted size.** A graph's Subjects and Observations are unbounded, as in the loaded store today; the records count graphs (4,096), and S16-01's per-graph cap bounds the claims the read returns (AC-1.4). One crafted graph with 20,000 Subjects fills once and stays under `capBudget`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache/cache.go` | `keyedValues[T]` *(new, unexported)*: `map[string]*list.Element` and `*list.List` of `{key string; v T}`; `data`, `store`, `Get`, `GetMany`, `Peek`, `Set`, `Drop`, `Clear`, `KeyCounts`, the keyed `Check` loop use it (D-1). `func DropNamedGraphs[T any](g *Graph, k Keyed[T], _ [][]byte, set effects.Set) error` *(new)* (D-3). `KeyedEntry` doc: "one value per 16-byte id (a handle's, or a graph's for entries that Apply `DropNamedGraphs`)"; `Cap`: "the most keys held; past it the least recently used leaves" |
| `core/database/graphcache/evidence_graph.go` | `loadedEvidenceGraph{graph, used}`; `EvidenceGraphMemo`, `SetEvidenceGraphMemo` go; `func (g *Graph) HeldEvidenceGraph(graphID []byte) (EvidenceGraph, bool)` *(new)* (D-6); `dropEvidenceGraphs` comment drops "with its memos" |
| `core/database/graphcache/commit.go` | The vocabulary branch loses the memo loop (D-14) |
| `core/database/graphcache/graph.go` | `evidenceGraphs` field comment: loaded graphs only |
| `core/database/graphfill/load.go` | `loadGraphRows(q, graphID) (subjects, observations, members, error)` *(new)*; `loadEvidenceGraph` calls it after `loadGraphHead` |
| `core/database/graphfill/records.go` *(new)* | `GraphRecords`, `SubjectRecord`, `newGraphRecords`, `Records`, `RecordsOf`, `maxGraphRecords`, the Check, registered in `init` (D-2, D-4, D-5, D-7, D-8). Package doc gains "and each graph's Subject records" |
| `core/database/evidencelayer/evidencelayer.go` | `Layers`, `patternLayer`, `maxLayers`; `Load` through `Layers.Get` (D-12); `memoKey` goes |
| `core/database/promotealign/canon.go` | `candidates` entry; `candidatesMemo.pattern`; `loadCandidatesMemo` through `Peek`/`Set` (D-13); `candidatesMemoKey` goes; `candidatesMemo` doc: dropped by the vocabulary and a commit naming the graph |
| `core/database/persondetail/claims.go` | `Claim` per D-9; `claims` composes per D-9; `memberGraph`, `memberGraphs` go; `boundedMembers` doc: the cap bounds the records the read returns, not a graph scan |
| `core/database/persondetail/recordlinks.go` | `SourcedGraph`, `recordLinks`, `linksBetween` per D-10 |

Errors: none new. `RecordsOf` on a nil graph returns `database.ErrClosed`; `IndexOf` and `RecordsOf` of `sources` keep theirs.

```go
// graphfill/records.go (sketch)
type SubjectRecord struct {
	Subject      graphcache.EvidenceSubject
	Claim        graphcache.EvidenceMember // zero: no accepted claim
	Observations []graphcache.EvidenceObservation // ref order
}
type GraphRecords struct {
	GraphID  []byte
	subjects []SubjectRecord
	byID     map[string]int
	labels   map[string]string
}
var Records graphcache.Keyed[*GraphRecords]
func RecordsOf(g *graphcache.Graph, graphIDs [][]byte) (map[string]*GraphRecords, error)
```

### FFI

None — no proto change. `api/ffi/handlers/person_detail.go`: `ListPersonClaims` maps `cl.Graph.ID`, `.Ref`, `.Name` and passes `cl.Records.Labels()` to `rowsOf`; `recordLinksProto` maps `g.Graph.ID`, `.Ref`, `.Name`. `api/ffi/handlers/source_read.go`: `rowsOf` and `rows` per D-11.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Keyed entries evict the least recently used key**
Files: `core/database/graphcache/cache.go`, `core/database/graphcache/cache_test.go`, `core/database/canonlinks/reader.go` (the `maxKeys` comment only).
Do: D-1 and D-3. Rewrite `TestEntryCap` as `TestKeyedEvictsLeastRecentlyUsed` (T-1).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/canonlinks/...` passes.

**STEP-2 — The two Evidence-graph memos become entries; the store holds only graphs**
Files: `core/database/evidencelayer/evidencelayer.go`, `evidencelayer_test.go`, `core/database/promotealign/canon.go`, `memo_test.go`, `core/database/graphcache/evidence_graph.go`, `commit.go`, `graph.go`, `evidence_graph_memo_test.go` (deleted), `evidence_graph_cap_test.go`, `core/database/graphfill/load_test.go`, `core/catalogsession/cache_doc_test.go` (blank-import `evidencelayer`), `docs/catalog-caches.md` (rows `evidencelayer.layers`, `promotealign.candidates`; the `evidenceGraphs` typed-field row loses `memos` and the loop).
Do: D-12, D-13, D-14. Memo assertions in the listed tests become entry assertions (`Layers.Peek`, `KeyCounts`); the cap test keeps its graph steps and drops its memo steps.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/evidencelayer/... ./core/database/promotealign/... ./core/database/promote/... ./core/database/graphfill/... ./core/catalogsession/...` passes; AC-4.6's `rg` finds nothing.

**STEP-3 — Per-graph records**
Files: `core/database/graphfill/load.go`, `core/database/graphfill/records.go` *(new)*, `core/database/graphfill/records_test.go` *(new)*, `core/database/graphcache/evidence_graph.go` (`HeldEvidenceGraph`), `core/catalogsession/cache_doc_test.go` (blank-import `graphfill`), `docs/catalog-caches.md` (row `graphfill.records`). Skill: `add-catalog-query`.
Do: D-2, D-4 – D-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/... ./core/catalogsession/...` passes, T-2 – T-4 included.

**STEP-4 — The claims read composes from the records**
Files: `core/database/persondetail/claims.go`, `claims_test.go`, `caps_test.go`.
Do: D-9. Test edits are field renames only (`Graph.GraphName` → `Graph.Name`, `GraphID` → `ID`, `GraphRef` → `Ref`); expectations stay.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/...` passes; T-7, T-8 pass.

**STEP-5 — Record links and the handlers**
Files: `core/database/persondetail/recordlinks.go`, `recordlinks_test.go`, `recordlinks_perf_test.go`, `api/ffi/handlers/person_detail.go`, `source_read.go`, `person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-10, D-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/...` passes; `TestRecordLinksBudgets` within its budgets; T-9 passes.

**STEP-6 — Budgets and sizes**
Files: `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` (`SeedScale`), `core/database/evidencegraphs/index_bench_test.go`, `core/database/graphfill/records_bench_test.go` *(new)*, `core/database/persondetail/claims_budget_test.go` *(new)*, `core/database/promotealign/bench_test.go` (`BenchmarkCacheSizes` gains `graphRecords`: `RecordsOf` over every graph in `IndexOf(g).All()`; the `evidence` measure now reports layers and candidates with it).
Do: D-15, D-16. Measure the planning base on the same machine: in a worktree at `aeb2773`, copy `claims_budget_test.go` in (it calls only `persondetail.Claims`, `households` and `catalogsessiontest.Cold`, which exist there) and run it and `BenchmarkCacheSizes`; record those figures and this branch's in the ship log.
Check: `CGO_ENABLED=1 go test -tags fts5 -run TestClaimsBudgets ./core/database/persondetail/` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkGraphRecords|BenchmarkGraphIndex' ./core/database/graphfill/ ./core/database/evidencegraphs/` reports within D-16; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkCacheSizes ./core/database/promotealign/` runs.

**STEP-7 — Docs**
Files: `docs/catalog-caches.md` (the three new rows with sizes from STEP-6; the Cap paragraph: "past its Cap a keyed entry drops its least recently used key"; the known-exceptions paragraph loses the vocabulary loop sentence; the typed-fields intro loses "(S16-13 and later PRs revisit the Evidence-graph store)"), `core/database/effects/README.md` (where `Set.EvidenceGraphs` is described: "graphcache drops each loaded graph, and the entries keyed by graph (`graphfill.records`, `evidencelayer.layers`, `promotealign.candidates`) drop theirs through `graphcache.DropNamedGraphs`"), `docs/deployment-plan/spike-16/cache-architecture.md` (§ Entity records S16-13 bullet: the measured size and time replace "est."), `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row status), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-guidance-names.py` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/cache_test.go` `TestKeyedEvictsLeastRecentlyUsed` *(replaces `TestEntryCap`)* | Table: Get-hit then Set past cap; Peek then Set; GetMany past cap; Check and KeyCounts then Set → the held keys of AC-3.1 – AC-3.4 | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-1 |
| T-2 | Go | `core/database/graphfill/records_test.go` `TestGraphRecordsFill` *(new)* | Table: graph loaded / not loaded → equal records, loaded issues no query; a Subject with and without a claim and a value-Subject Observation → `Subject`, `Labels`, `Entity` per AC-2.5; unknown graph id → empty records | AC-2.1, AC-2.5, D-5, D-7 |
| T-3 | Go | same file `TestGraphRecordsFollowCommits` *(new)* | Table: Observation added on A → A dropped, B held, refill has it, Verify on; graph A deleted → A dropped; `Drop` → none held | AC-2.3, AC-2.4, D-3 |
| T-4 | Go | same file `TestGraphRecordsOutliveLoadedGraph` *(new)* | Cap 1 on the store, load B after A → A's records held | AC-2.2 |
| T-5 | Go | `core/database/evidencelayer/evidencelayer_test.go` `TestLayersEntry` *(new; the line-93 memo check moves here)* | Table: vocabulary commit → none held; write naming A → A gone, B held; A held and graph evicted → `Load` loads no graph; unknown id twice → `ErrNoGraph`, none held; other pattern → replaced | AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, D-12 |
| T-6 | Go | `core/database/promotealign/memo_test.go` `TestLoadSharesPromotesMemo` *(reads `Layers`)*, `TestCandidatesEntry` *(new)* | Propose stores a layer and candidates; vocabulary commit empties both; write naming the graph drops both | AC-4.1, AC-4.2, D-13 |
| T-7 | Go | `core/database/persondetail/claims_test.go` `TestClaimsLoadNoGraph` *(new)* | `households(f, 80)`, Drop, Claims → 80 groups, loaded count 0; second read → 0 queries | AC-1.1, AC-1.2 |
| T-8 | Go | same file `TestClaimsSkipsWhatLeft` *(new)* | Table: graph deleted after the node was read (index lacks it) → skipped; a member whose record's claim names another entity → skipped | D-9, Reach |
| T-9 | Go | `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims` *(gains an Observation whose value is the second Subject)* | Response `value_subject_label` is that Subject's label; rest as before | AC-1.6, D-11 |
| T-10 | Go | `core/database/persondetail/claims_budget_test.go` `TestClaimsBudgets` *(new)* | `households(f, 80)`: cold ≤ 1 s, warm ≤ 50 ms | AC-1.5, D-16 |
| T-11 | Go bench | `core/database/graphfill/records_bench_test.go` `BenchmarkGraphRecords` *(new)* | `SeedScale`, Drop, `RecordsOf` every graph → `build-ms`, `heap-MB`; stop past 2 s or 35 MB | D-15, D-16, Scale |
| T-12 | Go | existing `TestClaims*`, `TestCraftedClaims`, `TestRecordLinks*`, `TestRecordLinksBudgets`, `TestGraphWriteDropsOnlyThatGraph`, `TestRetitleKeepsEvidenceGraph`, `TestCacheDocListsEveryEntry` | Unchanged expectations (field renames and memo → entry assertions only) | AC-1.3, AC-1.4, AC-4.7 |
| T-13 | Sweep | STEP-2's check: `rg -n 'EvidenceGraphMemo\|SetEvidenceGraphMemo\|memos' core/database/graphcache/*.go -g '!*_test.go'` | No hit | AC-4.6, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `./core/database/promotealign/...` (memo move), `./core/database/graphcache/...` (LRU), `./core/catalogsession/...` (cache doc rows and owner imports).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`EvidenceGraphMemo`, `SetEvidenceGraphMemo`, `memberGraphs`, `candidatesMemoKey`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (done at planning; confirm the map text still matches)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none: none
- [ ] Ship log: TestClaimsBudgets at the planning base and after; BenchmarkGraphRecords; BenchmarkCacheSizes kinship, evidence and graphRecords before and after
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated): not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md (three rows, the Cap rule, the store row), core/database/effects/README.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person cited in more than 64 Evidence graphs (or run the app on a catalog built with `households(f, 80)`'s shape). Open the person's Sources section. See every Source group and claim as before.
2. Leave and reopen the person. The section draws at once.
3. Rename one of those Sources from its Source page. Back on the person, the group shows the new title.
4. Add an Observation to one cited Subject in its Evidence graph. Back on the person, the claim shows it.
5. Run one Promote proposal on a Source before and after a vocabulary edit (add a term): suggestions are the same as before the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| LRU bookkeeping slows the kinship reads | O(1) per use; `BenchmarkCacheSizes` and `TestCraftedFamily` before and after |
| A memo outliving its loaded graph changes Promote's output | Same drop signals as today; the generation and pattern checks stay; AC-4.7 |
| Records filled from a loaded graph differ from SQL | One constructor; T-2 and the Check under Verify |
| The cold 80-graph read is slower than today (records fill per graph) | Budget 1 s; planning-base figure recorded |

**Stop and report if:**

- A row of §4 is no longer true at your base (in particular `EvidenceGraphMemo` gained a caller, `Keyed` gained a method, or `persondetail.Claims` gained a caller outside the handlers).
- AC-4.7: any Promote suite or matching golden changes output.
- `TestClaimsBudgets` warm is above 50 ms or cold above 1 s, or the planning-base cold figure is faster than the new cold by more than 25%.
- `BenchmarkGraphRecords` passes 2 s or 35 MB, or `BenchmarkCacheSizes`' kinship figure grows by more than 20%.
- `g.Verify()` reports `graphfill.records` mismatched after a write that went through `writes.Run`: name the write and its `Set.EvidenceGraphs`.
- `graphfill` cannot be imported by `persondetail` without a cycle, or `evidencelayer` by `catalogsession`'s test.
- A test outside §8's list needs more than a field rename or a memo → entry assertion to pass.

## 12. Open questions

None. Answered 2026-11-13: the Citation fields join with S16-30; the records tier and both memos are registry entries keyed by graph, with least-recently-used eviction added to the registry, in one PR; "simpler code" means simple reads, with construction in the cache layer.

## 13. PR

**Title:** S16-13 — Per-graph records in the cache registry, and the Sources section reads them
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; TestClaimsBudgets, BenchmarkGraphRecords and BenchmarkCacheSizes figures; `Plan: docs/deployment-plan/spike-16/prs/S16-13-records-tier.md`.
