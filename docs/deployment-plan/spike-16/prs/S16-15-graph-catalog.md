# S16-15 — An Evidence-graph catalog in the graph cache

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Sources list shows the same counts and graph rows; only where the engine reads them from moves)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), the cache registry this entry joins); S16-12 (landed [#800](https://github.com/mendahu/provenencia/pull/800)) is not a dependency but is the precedent whole entry
**Blocks:** S16-13 (orders a person's graphs from `OfSource` and reads graph shells from `Graph`), S15-24 (a move re-lists both Sources; this entry's Apply handles a graph that changes Source)
**Branch:** `s16-15-graph-catalog`
**Version bump:** none — no change a researcher can see; the project-wide graph list's order changes on the wire and no client reads it ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) Cache entities and results per key; compose answers", "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) A write's one-shot graph fills only what it reads", "(Batch 2) Warm only what every proposal needs; no background re-warm", "(Batch 2) Orders are defined and results do not move", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth"; goal bar 12 (the progress clause)
**Planned:** 2026-11-12 against `main` @ `064e1f3`
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

The Sources list shows each Evidence graph's Subject, filed and Observation counts. Today every list read runs two full-table `GROUP BY`s over `subjects` and `observations` (`graphprogress.List`), and every per-graph refresh runs two more scoped queries (`graphprogress.Get`); the graph list itself is two queries per read. After this PR the graph cache holds one catalog of every graph (its row, its linked Artifacts and its counts) and each Source's graphs in list order, built on first read and patched from the commit's `effects.Set.Reads` (`evidence_graph_progress` recounts the named graphs, `evidence_graphs` re-lists the named Sources). The progress and graph-list reads become map reads, and `graphprogress` folds into `evidencegraphs`, so one package owns graph rows, their counts and the cache over both.

**Why now.** S16-13 orders a person's graphs from this catalog and reads graph shells from it instead of loading whole graphs; S15-24's move re-lists both Sources through it. Goal bar 12: "The Sources list's graph progress needs no full-table count."

**What stays.** Every count the researcher sees and its rules (a Source's reification Subject is not counted; filed means a Promote primary kind with an accepted claim; a graph with all-zero counts is omitted from the progress list and reads zeros alone); each Source's graph order (name case-insensitive, then ref, then id); the wire messages; every write path, which keeps reading graph rows through its transaction (`list(tx, …)`, `getTx`, `SourceOf`); the Swift keys `.evidenceGraphProgress` and `.evidenceGraphsList`, which already merge per graph and per Source on the same two reads.

## 2. User stories

**US-1.** As a researcher, when I open the Sources list and work in a Source, each graph's counts are the same as today and follow every edit, and the list reads them from memory.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `TestGraphProgress`'s fixture: Source "Worked" whose Default graph has 2 counted Subjects, 1 Observation; Source "Empty"; a second graph "Jones" with 1 place Subject; one person Subject with an accepted claim | `ListProgress` returns Default `{SubjectCount 2, FiledCount 1, ObservationCount 1}` and Jones `{1, 0, 0}`, and no row for a graph with all zeros; `GetProgress(unknown id)` returns `{GraphID: unknown}`; `GetProgress([]byte{1})` returns `ErrInvalid` with code `evidencegraphs.invalid` |
| AC-1.2 | The catalog is warm (one `ListProgress`); a query counter is set on `c.Graph()` | `ListProgress`, `GetProgress(def)`, `List(c, nil)` and `List(c, worked.ID)` together issue 0 queries |
| AC-1.3 | Catalog warm; an Observation is added on a Subject of Default through `writes.Run` | `GetProgress(def).ObservationCount` is 2 with no `Drop`, and `g.Verify()` passes |

**US-2.** As the engine, the catalog is exact after every commit and holds a defined order.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Catalog warm; each write of T-2's table commits through `writes.Run` (graph create, rename, link, unlink, Subject create, claim accept, Observation create, Subject delete, graph delete, Source rename, second-Source delete) | After each commit `g.Verify()` passes; an entry registered `DerivedFrom: evidencegraphs.Catalog` is emptied by each commit whose `Set.Reads` names `evidence_graphs` or `evidence_graph_progress`, and kept by the Source rename (which names neither); the values listed in T-2 hold |
| AC-2.2 | Graph G of Source A has its `source_id` set to Source B by SQL, then `g.OnCommit` runs with `Reads: [{evidence_graphs, [A, B]}]`, and again from the start with `[B, A]` | Both orders: `OfSource(A)` lacks G, `OfSource(B)` holds G in list order, `Graph(G).SourceID == B`, `Verify` passes |
| AC-2.3 | `CatalogOf(graphcache.Over(tx))`; `CatalogOf(nil)`; `List`, `ListProgress`, `GetProgress` on a closed catalog | `ErrCatalogOnWrite` and nothing stored; `database.ErrClosed`; `database.ErrClosed` for all three |
| AC-2.4 | Source "Census 1881" with graphs "Default", "Jones", "smith"; Source "Parish register" (created second) with one graph renamed "Archive" by SQL before the first read | `List(c, census)` → Default, Jones, smith (case-insensitive by name); `List(c, nil)` → Default, Jones, smith, Archive (Sources by id ascending, each in list order); Jones' `ArtifactIDs` by Artifact ref |

**US-3.** As a contributor, graph rows and graph counts have one loader each, in the package that owns the table, and one cache in front of both.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `rg -n "graphprogress" core api .agents .cursor docs/*.md` | No hit; `core/database/graphprogress/` does not exist; `sqlCountSubjectsByTypeOne` and `sqlCountObservationsOne` do not exist |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphprogress` folds into `evidencegraphs`: `Progress`, one count loader for all graphs or named graphs, `ListProgress`, `GetProgress`; the package and its per-graph queries go | STEP-1, AC-1.1, AC-3.1 |
| `list` reads through any `database.Querier`, so the cache and the writes share it | STEP-2 |
| `evidencegraphs.Catalog` whole registry entry: Build, Apply on `Set.Reads`, Check, graph lifetime, `CatalogOf`, `GraphCatalog` with `Graph`, `OfSource`, `All`, `Progress`, `AllProgress` | STEP-3, AC-2.1, AC-2.2, AC-2.3 |
| `List`, `ListProgress`, `GetProgress` read the catalog | STEP-4, AC-1.2, AC-1.3, AC-2.4 |
| Size and Apply measured; `docs/catalog-caches.md` row; `BenchmarkCacheSizes` fill; doc-test import | STEP-5 |
| Docs: design line, map scope entry, completed.md | STEP-7 |

| Out | Home |
| --- | --- |
| The claims read ordering a person's graphs from the catalog, and the records tier | S16-13 |
| `graphfill.loadGraphHead` reading a graph's ref, name and Source from the catalog (it also runs on a write's graph, which this entry does not fill) | S16-13 (splits `core/database/graphfill/load.go`) |
| `evidencegraphs.SourceOf`, `SubjectGraphTx`, `IsLinkedTx`, `getTx` and `list(tx, …)` inside writes reading the catalog | not planned: a write reads its own transaction, and both `SourceOf` callers (`subjects.Create`, `promote.SaveBatch`) are writes |
| `searchindex`'s and `deleteimpact`'s graph SQL | not planned: write paths |
| A graph moved between Sources by a real write | S15-24 (AC-2.2 holds the Apply now) |
| The app's `EvidenceGraphIndex` (Swift) | not planned: Swift's per-view key is enough (rule "No Go cache for a view that is one query") |
| A batch graph read | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Progress | `core/database/graphprogress/graphprogress.go` `List`, `Get`, `Progress`, `add`, `scanSubjects`, `scanSubjectsOne`, `scanGrouped`, `ErrInvalid` | `List` runs `sqlCountSubjectsByType` and `sqlCountObservationsByGraph` over every row (`GROUP BY`), skips `key == "source"` rows before making an entry, and returns rows in map order; `Get` runs `sqlCountSubjectsByTypeOne` and `sqlCountObservationsOne` for one graph and returns zeros for an unknown or empty graph. `ErrInvalid` is `apperr.CodeEvidenceGraphsInvalid`, the same code as `evidencegraphs.ErrInvalid` |
| Progress tests | `core/database/graphprogress/graphprogress_test.go` `TestGraphProgress`, `TestFiledCounts` (package `graphprogress`, internal) | Hold the counts, zeros, `ErrInvalid` code and `ErrClosed`. `TestGraphProgress`' "two graphs" case inserts a canonical entity and an accepted claim by SQL, then cites an Observation through a write |
| Graph rows | `core/database/evidencegraphs/evidencegraphs.go` `List`, `list(tx *sql.Tx, sourceID)`, `scanGraph`, `sqlList`, `sqlListBySource`, `sqlListLinks`, `sqlListLinksBySource`, `Graph` | `List` opens a read transaction and calls `list`; `list` runs the graph query (`ORDER BY name COLLATE NOCASE, ref, id`) then the link query (`ORDER BY a.ref COLLATE NOCASE, a.id`) and fills `ArtifactIDs` (nil when a graph has no link). `Update` and `LinkArtifact` call `list(tx.Tx, …)` inside writes |
| Graph list tests | `core/database/evidencegraphs/evidencegraphs_test.go` `TestList`, `TestListQueryPlan`; `export_test.go` `ListSQL` | `TestList` "the whole project" wants Archive, Default, Jones, smith (global name order) |
| Readers | `api/ffi/handlers/evidence_graph_progress.go` `ListEvidenceGraphProgress`, `GetEvidenceGraphProgress`, `protoProgress`; `api/ffi/handlers/evidence_graphs.go` `ListEvidenceGraphs` | The only production callers of `graphprogress.List`, `graphprogress.Get` and `evidencegraphs.List`. `ListEvidenceGraphs`' comment says "by name" |
| Handler tests | `api/ffi/handlers/evidence_graph_progress_test.go` `TestListEvidenceGraphProgress`, `TestGetEvidenceGraphProgress`, `TestEvidenceGraphProgressPerGraph`; `api/ffi/handlers/evidence_graphs_test.go` `TestListEvidenceGraphs`, `TestListEvidenceGraphsSourceID` | Through the wire |
| Wire | `api/proto/engine.proto` `ListEvidenceGraphsRequest` (`source_id`: "empty: every graph of the project"), `EvidenceGraphProgress`, `ListEvidenceGraphProgressResponse`, `GetEvidenceGraphProgressResponse` | Unchanged by this PR |
| Reads | `core/database/effects/registry.go` `ReadEvidenceGraphProgress` (keyed by graph: `subjects`, `observations`, `identity_claims`), `ReadEvidenceGraphs` (keyed by Source: `evidence_graphs`, `evidence_graph_artifacts`), `Set.Reads []ReadIDs`; `core/database/effects/reads_test.go` `TestReadsDeclare` pins the tables | A graph delete (`deleteimpact.releaseGraph`) and an Artifact delete (its link facets) write audited rows with `source_id` / `graph_id`, so both reads name the Source; `subject_types` keys are never updated (no `UPDATE subject_types` in `core`) |
| Swift follows | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `.evidenceGraphProgress` (`.merged(follows: .reads(.evidenceGraphProgress))`), `.evidenceGraphsList` (`.merged(follows: .reads(.evidenceGraphs))`, replaces one Source's rows and appends them) | Swift already trusts these two reads; `EvidenceGraphIndex` groups rows by Source and reads no cross-Source order |
| Registry | `core/database/graphcache/cache.go` `RegisterWhole`, `WholeEntry`, `Whole.Get`, `Whole.Set`, `Whole.Drop`, `LifeGraph` | A `LifeGraph` whole entry is neither built nor stored on an `Over` graph (`Get` → `ok == false`); Apply runs only while held; `Set` stores and empties derived entries; `Drop()` of the graph empties every `LifeGraph` entry |
| Precedent entry | `core/database/sources/records.go` `Records`, `RecordIndex`, `RecordsOf`, `ErrRecordsOnWrite`, `loadRecords` | S16-12's whole entry: Build through `loadRecords(g, nil)`, Apply removes every named Source before inserting any (review CON-4), then `Records.Set`; Check rebuilds and `reflect.DeepEqual`s |
| Precedent tests | `core/database/sources/records_test.go` `TestRecordsApply`, `TestRecordsApplySwap`, `TestRecordsOf`, `derivedFromRecords`; `core/database/sources/records_bench_test.go` `BenchmarkSourceRecords` | Table of writes with `g.Verify()` after each; a crafted bench by direct SQL that times `g.OnCommit` |
| Querier | `core/database/tx.go` `database.Querier` (`Query`, `QueryRow`); `core/database/graphcache/graph.go` `Graph.Query`, `Graph.QueryRow`, `SetQueryCounterForTest` | `*graphcache.Graph` and `*sql.Tx` both satisfy `database.Querier`; queries through `g` are counted |
| Batching | `core/database/sqlbatch/sqlbatch.go` `Query(q, head, tail, args, ids, scan)`, `Scan` | `head` + placeholders + `tail` per batch |
| Open | `core/catalogsession/session.go` `ReconcileOnOpen` | `evidencegraphs.EnsureRefs` (an unaudited ref backfill) runs before any read of the session |
| Doc test | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry`, `TestCacheDocSeesEveryOwner` | Blank-imports each owner package; `evidencegraphs` is not imported yet |
| Size benchmark | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` | One `measure` per cache; `sourceRecords` is the latest |
| Indexes | `core/database/migrations/000055.sql` `subjects_graph_id_idx`; `000029.sql` `observations_subject_id_idx` | A per-graph recount seeks both |

**Precedent.** [#800](https://github.com/mendahu/provenencia/pull/800) (S16-12) registered `sources.Records`: copy its file layout (entry, value type, read function and `ErrRecordsOnWrite`-style error in one file of the table's package), its Build/Apply/Check through one loader taking an id list or nil, its "remove every named key before inserting any" Apply, its `TestRecordsApply` table with `g.Verify()` and a derived probe entry, and `BenchmarkSourceRecords`' direct-SQL fixture and `g.OnCommit` timing. Do not copy its Source-scoped signal: this entry follows `Set.Reads`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The entry lives in `core/database/evidencegraphs`, file `catalog.go` *(new)*. `graphprogress` folds into the same package as `progress.go` *(new)*: `Progress`, `ListProgress(c)`, `GetProgress(c, graphID)`, the loader `loadProgress`; `core/database/graphprogress/` is deleted and `api/ffi/handlers/evidence_graph_progress.go` imports `evidencegraphs`. The tests move to `core/database/evidencegraphs/progress_test.go` *(new)* in package `evidencegraphs_test` (the old file imports `evidencegraphstest` and `sources`, which import `evidencegraphs`, so an internal test would cycle) | The table's package owns its cache (S16-12 D-1); the Build needs the count loader, so `graphprogress` would either be imported by `evidencegraphs` (and could not then read the catalog: cycle) or be left a wrapper over another package's cache | `graphprogress` keeps its package and reads `CatalogOf` — its SQL must still move, leaving a package that only converts; the entry in `graphprogress` — graph rows cached outside their package, and `list` exported |
| D-2 | Lifetime `graphcache.LifeGraph`, entry name `"evidencegraphs.catalog"`, handle `var Catalog graphcache.Whole[*GraphCatalog]`. `CatalogOf(g *graphcache.Graph) (*GraphCatalog, error)` *(new)*: nil graph → `database.ErrClosed`; `ok == false` on a non-nil graph → `ErrCatalogOnWrite` *(new)*, `errors.New("evidencegraphs: the graph catalog is not read on a write's graph")` | Rule "A write's one-shot graph fills only what it reads"; same contract as `sources.RecordsOf`, which S16-13 composes beside | A SQL fallback on `Over` — a second read path |
| D-3 | `GraphCatalog` holds `graphs map[string]Graph` (graph id → the row `list` returns: ID, Ref, SourceID, Name, Description, ArtifactIDs), `progress map[string]Progress` (graph id → counts, only the graphs `loadProgress` returns a row for) and `bySource map[string][][]byte` (Source id → graph ids in list order). Methods: `Graph(id []byte) (Graph, bool)`; `OfSource(sourceID []byte) []Graph` (list order; nil for none); `All() []Graph` (Sources by `bytes.Compare` of their id, each Source's graphs in list order); `Progress(id []byte) Progress` (the held counts, else zeros, `GraphID` set to a copy of `id`); `AllProgress() []Progress` (the held rows with a non-zero count, in `All` order). Every method returns a new slice; `Graph` values share their `ArtifactIDs` with the held index and readers treat them as read-only | `List` returns Description and ArtifactIDs, so the catalog holds the whole row to serve it; three maps keyed by id keep Apply O(named) | A slim shell without Description and links — `List` would keep its SQL beside the cache (a second path); one global ordered slice — Apply would need a Go copy of SQLite's `NOCASE` order to insert into it |
| D-4 | Project-wide order (`All`, `List(c, nil)`, `AllProgress`) is Sources by id ascending, then each Source's list order. `ListEvidenceGraphs`' comment and `TestList`'s "whole project" rows change; the proto comment gains "grouped by Source" | Per-Source order comes from SQL (Build's global query grouped stably by Source; Apply's per-Source query), so no Go collation exists; `.evidenceGraphsList` already appends a re-listed Source at the end and `EvidenceGraphIndex` groups by Source, so no reader depends on cross-Source order | Global name order — needs the Go `NOCASE` comparator rejected in D-3 |
| D-5 | `list` takes `q database.Querier` in place of `tx *sql.Tx`; `Update` and `LinkArtifact` pass `tx.Tx` as today. Build calls `list(g, nil)`; Apply calls `list(g, sourceID)` per named Source. `List`'s own `db.Begin()` goes | One loader for graph rows, shared by writes and the cache | A second loader for the cache |
| D-6 | `loadProgress(q database.Querier, graphIDs [][]byte) (map[string]Progress, error)`: `graphIDs == nil` runs today's two full `GROUP BY`s; otherwise the same two queries with `WHERE s.graph_id IN (` + placeholders + `) GROUP BY s.graph_id, t.key, t.origin` (Subjects) and `) GROUP BY s.graph_id` (Observations) through `sqlbatch.Query`. It keeps today's rules: a `key == "source"` row makes no entry, an Observation row makes one; `add` is unchanged. `sqlCountSubjectsByTypeOne`, `sqlCountObservationsOne`, `scanSubjectsOne` go | One count loader for Build, Apply, Check and the uncached `ListProgress`/`GetProgress` of STEP-1 | Keeping the per-graph queries for Apply — two loaders of one count |
| D-7 | Apply loops over `set.Reads` once: `ReadEvidenceGraphs` gives Source ids, `ReadEvidenceGraphProgress` graph ids; returns at once when both are empty. Then (1) `fresh := list(g, sid)` for every named Source, all loaded before any change; (2) for every named Source, delete its held graph ids from `graphs` and `bySource`; (3) insert each fresh list into `graphs` and `bySource[sid]` (an empty list leaves no `bySource` key); (4) delete the `progress` row of every graph removed in (2) and absent after (3); (5) `loadProgress(g, graphIDs)`, delete each named id's row, store each returned row whose graph is in `graphs`; (6) `Catalog.Set(g, idx)` with the same pointer | O(named) per commit; remove-before-insert keeps a graph that moved between two named Sources (S16-12 review CON-4); `Set` is the registry's one write path, so an entry `DerivedFrom: Catalog` (S16-13) is emptied | A new `effects.Set` helper for one loop — a name with one caller; deltas from rowchanges — a special case per table |
| D-8 | Check rebuilds with `list(g, nil)` and `loadProgress(g, nil)` and compares `graphs`, `progress` and `bySource` with `reflect.DeepEqual`; error `errors.New("evidencegraphs: the stored graph catalog does not match the catalog tables")` | S16-12's Check shape; Verify holds every commit in T-2 | A Check of counts only — misses order |
| D-9 | `List(c, sourceID)`, `ListProgress(c)`, `GetProgress(c, graphID)` validate their ids as today (`ErrInvalid`), call `c.DB()` first (so a closed catalog is `database.ErrClosed`: `Close` leaves `c.Graph()` non-nil), then read `CatalogOf(c.Graph())`. `List` with an empty id returns `All()`, else `OfSource`; an unknown Source returns nil | The held value must not outlive the connection; signatures and errors stay, so callers and handler tests do not change | `CatalogOf` checking the connection — it has only the graph |
| D-10 | The entry is filled on first read only; nothing warms it at open | Rule "Warm only what every proposal needs" | Warming in `WarmOnOpen` — the rule asks for an open-time measurement first |
| D-11 | A test that writes evidence rows by direct SQL after it has read the catalog calls `c.Graph().Drop()` before its next read, with a one-line comment. Expectations of existing tests do not change | The S16-12 and S16-11 idiom (`TestForEntityQueryCountIsConstant` drops before counting); a direct-SQL write reports no effects | A production invalidation for fixture writes |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/evidencegraphs/catalog.go` *(new)*: `Catalog`, `GraphCatalog` with `Graph`, `OfSource`, `All`, `Progress`, `AllProgress`, `CatalogOf`, `ErrCatalogOnWrite`, `newGraphCatalog` | whole registry entry, type, funcs | `core/database/sources/records.go` `Records`, `RecordIndex`, `RecordsOf` | Searched `RegisterWhole`, `evidence_graphs` readers and `graphprogress`: no cache holds graph rows or counts; the shape is the precedent's |
| `core/database/evidencegraphs/progress.go` *(new)*: `Progress`, `ListProgress`, `GetProgress`, `loadProgress` | moved type and funcs | `core/database/graphprogress/graphprogress.go` | Moved, not added beside (D-1); the package is deleted |
| `core/database/evidencegraphs/progress_test.go` *(new)* | moved tests | `core/database/graphprogress/graphprogress_test.go` | Moved (D-1) |
| `core/database/evidencegraphs/catalog_test.go` *(new)*, `catalog_bench_test.go` *(new)* | tests | `core/database/sources/records_test.go`, `records_bench_test.go` | New entry |

**Builds on.** The S16-10 registry (`RegisterWhole`, Apply on `effects.Set`, `Whole.Set`, Check under `PROVENENCIA_GRAPH_VERIFY`); `effects.Set.Reads` and its two keyed reads, which Swift already follows; `sqlbatch.Query`; S16-12's entry shape.

**Deletes.**
- `core/database/graphprogress/` (`graphprogress.go`, `graphprogress_test.go`): `List`, `Get`, `ErrInvalid` (same code as `evidencegraphs.ErrInvalid`), `scanSubjectsOne`, `scanGrouped`'s package copy, `sqlCountSubjectsByTypeOne`, `sqlCountObservationsOne`.
- `List`'s own read transaction (`db.Begin()` / `tx.Rollback()`).
- The two full `GROUP BY`s and the per-graph count queries on every read path (they run in Build, Apply and Check only).

**Next adds.** S16-13 reads `CatalogOf(c.Graph())`: `OfSource` orders a person's graphs and `Graph` gives a graph's ref, name and Source without a loaded graph; it may register an entry `DerivedFrom: evidencegraphs.Catalog` (D-7 empties it). S15-24's move updates `evidence_graphs.source_id`, whose change names old and new Source under `ReadEvidenceGraphs` (`readChangeColumn`); D-7 and AC-2.2 already handle it.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Progress and graph-list RPCs served from a held value | `ListEvidenceGraphProgress`, `GetEvidenceGraphProgress`, `ListEvidenceGraphs` | Same messages; an unknown graph reads zeros and an unknown Source no graphs, as today (AC-1.1, AC-2.4); a closed catalog is `ErrClosed` (D-9) |
| `CatalogOf` on a write's graph or a nil graph | none today; S16-13's readers run on `c.Graph()` | `ErrCatalogOnWrite` / `ErrClosed` (AC-2.3) |
| A graph whose Source changes | Apply | AC-2.2 |

### Scale

At the map's Batch 2 Scale (3,000 Sources, 10,000 Evidence graphs, 50,000 Observations; large fixture at 4,000 and 16,000 persons):

- **Build** (first progress or list read of a session, or after `Drop`): four queries: every graph row, every link, the Subject `GROUP BY` and the Observation `GROUP BY` (today's progress cost, paid once instead of per read). Design estimate 30–80 ms and about 2 MB; the whole rows with Description and links add an estimated 1–2 MB. Measured by `BenchmarkGraphCatalog` (T-9); **Stop** over 150 ms or 8 MB.
- **Apply per commit.** Only while held. An Observation, Subject or claim write names its graph's progress: two scoped `GROUP BY`s over that one graph (`subjects_graph_id_idx`, `observations_subject_id_idx`). Real graphs hold tens to hundreds of Observations; crafted, a graph with 10,000 Observations (bar 2's axis) recounts 10,000 rows on each such write: budget 20 ms (T-9). A graph or link write names its Source: two queries listing that Source's graphs. Real Sources hold one to five graphs; crafted, a Source with 10,000 graphs re-lists 10,000 rows on each graph write: budget 50 ms (T-9). No cap: a capped list or count would be wrong, and both costs are linear in one Source or one graph, not in the catalog. A commit naming many graphs or Sources is batched by `sqlbatch` (progress) or loops per Source (lists); no write today names more than the graphs of one Source.
- **Reads:** `ListProgress` and `List(c, nil)` walk the maps, O(graphs); per-graph and per-Source reads are map lookups. No SQL once warm (AC-1.2).
- **Memory:** `docs/catalog-caches.md` gets the entry's size at 4,000 and 16,000 persons (`BenchmarkCacheSizes`) and at the Scale (T-9); est. under 8 MB against the 250 MB ceiling.

### Edges

- **Assumptions.** The two reads name every graph whose counts and every Source whose graph rows a commit changes: `subjects`, `observations`, `identity_claims` → progress; `evidence_graphs`, `evidence_graph_artifacts` → list (`TestReadsDeclare` pins them; Swift relies on the same set today). Counts depend on `subject_types.key` and `origin`, which no code updates. `EnsureRefs` writes refs unaudited and runs in `ReconcileOnOpen` before any read of the session. A write that bypasses `writes.Run` would leave the catalog stale; Verify's Check (D-8) catches it in T-2 and in any test run with `PROVENENCIA_GRAPH_VERIFY=1`.
- **Neighbors.** A Source rename, Artifact relabel, citation or note edit names neither read: the catalog and any entry derived from it stay (T-2's Source rename row). An Observation edit that changes nothing counted still recounts its graph (one graph, same values). A graph rename re-lists its Source and can move the graph within `OfSource` (T-2).
- **Removed inputs.** A deleted graph (only one with no Subjects can be deleted) leaves `graphs`, `bySource` and `progress` (D-7 step 4). A deleted Subject drops its graph's counts by its own and its Observations' contribution (the recount reads the table). A deleted Source (only one with no Artifacts or Subjects) names its graphs' deletes under its id, so `OfSource` returns nil. A deleted Artifact's link facets name the graph's Source, so `ArtifactIDs` drops it. T-2 holds each.
- **Real data.** Counts are not driven by a table beyond `vocabkey.IsPrimaryType` (person, event, place from `provenencia`), which this PR does not change; the moved `TestFiledCounts` keeps its case per seeded type.
- **Untrusted size.** Build is linear in graphs, links, Subjects and Observations. Apply multiplies by graphs per Source (re-list) and Observations per graph (recount), both budgeted above. Ids per commit are chunked by `sqlbatch`. No loop over graphs runs inside a loop over Sources except grouping Build's one ordered result.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/evidencegraphs` | `progress.go` *(new)* (D-1, D-6): `type Progress struct{ GraphID []byte; SubjectCount, ObservationCount, FiledCount int32 }` with `add`; `ListProgress(c *database.Catalog) ([]Progress, error)`; `GetProgress(c *database.Catalog, graphID []byte) (Progress, error)`; `loadProgress`. `catalog.go` *(new)* (D-2, D-3, D-7, D-8). `evidencegraphs.go`: `list` takes `database.Querier` (D-5); `List` reads the catalog (D-9) |
| `core/database/graphprogress` | Deleted |
| `core/catalogsession` | `cache_doc_test.go` blank-imports `core/database/evidencegraphs` |
| `core/database/promotealign` | `bench_test.go` `BenchmarkCacheSizes` gains `measure("graphCatalog", func() { _, err := evidencegraphs.CatalogOf(g); must(b, err) })` after `sourceRecords` |

Errors: `evidencegraphs.ErrCatalogOnWrite` (internal invariant, not an `apperr`); `ErrInvalid` as today. No new `apperr` code.

### FFI

`api/ffi/handlers/evidence_graph_progress.go`: `evidencegraphs.ListProgress`, `evidencegraphs.GetProgress`, `protoProgress(row evidencegraphs.Progress)`. `api/ffi/handlers/evidence_graphs.go`: `ListEvidenceGraphs`' comment reads "returns the graphs of one Source in list order, or of the whole project grouped by Source". `api/proto/engine.proto` `ListEvidenceGraphsRequest.source_id` comment becomes "empty: every graph of the project, grouped by Source"; regenerate with `./scripts/generate-proto.sh` (comment-only; if `protoc` is missing, leave the generated files and say so in the report). No message changes.

### macOS

None — no Swift change; the two keys already follow these reads.

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — Graph progress moves into `evidencegraphs`**
Files: `core/database/evidencegraphs/progress.go` *(new)*, `core/database/evidencegraphs/progress_test.go` *(new)*, `core/database/graphprogress/` (deleted), `api/ffi/handlers/evidence_graph_progress.go`. Skill: `add-catalog-query`.
Do: Before any change, run `BenchmarkCacheSizes` on `main` and put its figures in your report. Move `Progress`, `add`, the two full-table queries and their scans into `progress.go`; write `loadProgress` (D-6); `ListProgress` and `GetProgress` run `loadProgress` on `c.DB()` for now (`nil` and `[][]byte{graphID}`), with today's id check, zeros and order. Move both tests to `progress_test.go` in package `evidencegraphs_test`, qualifying names (`evidencegraphs.ListProgress`, `evidencegraphs.Progress`, `evidencegraphs.ErrInvalid`); expectations unchanged. Repoint the handler; delete the package.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/... ./api/ffi/handlers/...` passes; `rg -n "graphprogress" core api` finds nothing.

**STEP-2 — One graph-row loader for writes and the cache**
Files: `core/database/evidencegraphs/evidencegraphs.go`.
Do: D-5: `list(q database.Querier, sourceID []byte)`; `List` calls `list(db, sourceID)` with no transaction (still SQL in this step); `Update` and `LinkArtifact` pass `tx.Tx`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/...` passes unchanged, `TestListQueryPlan` included.

**STEP-3 — The catalog entry**
Files: `core/database/evidencegraphs/catalog.go` *(new)*, `core/database/evidencegraphs/catalog_test.go` *(new)*. Skill: `add-catalog-query`.
Do: Register `Catalog` in `init()` (D-2): Build = `list(g, nil)` + `loadProgress(g, nil)` → `newGraphCatalog`, grouping the ordered rows by `SourceID` without re-sorting; Apply per D-7; Check per D-8. `CatalogOf` and the methods per D-3. Write T-1 – T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestCatalog' ./core/database/evidencegraphs/...` passes.

**STEP-4 — Reads go through the catalog**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/progress.go` *(new)*, `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/evidencegraphs/progress_test.go` *(new)*, `api/ffi/handlers/evidence_graphs.go`, `api/proto/engine.proto`. Skill: `add-ffi-handler`.
Do: D-9 for `List`, `ListProgress` (`AllProgress`), `GetProgress` (`Progress`). Update `TestList`'s "whole project" rows (D-4, AC-2.4); add T-7 and T-8. Apply D-11 where a progress test's fixture writes by SQL after a read. Update the two comments and regenerate (FFI section).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/... ./api/ffi/handlers/...` passes.

**STEP-5 — Measure and document the entry**
Files: `core/database/evidencegraphs/catalog_bench_test.go` *(new)*, `core/database/promotealign/bench_test.go`, `core/catalogsession/cache_doc_test.go`, `docs/catalog-caches.md`.
Do: T-9 and T-10. Add the `docs/catalog-caches.md` "Registry entries" row: Cache `evidencegraphs.catalog`; Owner `core/database/evidencegraphs/catalog.go` *(new)* · `Catalog`; Key "one value"; Fill "Build: four queries (every graph row, every link, the Subject and Observation counts), on first read"; Updated by "Apply re-lists the Sources in `effects.Set.Reads` `evidence_graphs` and recounts the graphs in `evidence_graph_progress`, then `Set`"; Lifetime graph; Cap none; Size: both `BenchmarkCacheSizes` figures and T-9's. Blank-import the package in `cache_doc_test.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestCacheDoc' ./core/catalogsession/...` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkGraphCatalog' -benchtime 20x ./core/database/evidencegraphs/` reports build-ms, heap-MB, apply-progress-10k-ms and apply-list-10k-ms inside the Scale budgets; `-bench BenchmarkCacheSizes` in `./core/database/promotealign/` reports `graphCatalog` at both sizes.

**STEP-6 — Whole suite**
Files: any test file the suite names, D-11 only.
Do: Run the whole Go suite. Where a test fails only because it wrote evidence rows by SQL after a catalog read, apply D-11. Any other failure is a Stop condition.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/... ./core/database/sources/... ./api/ffi/handlers/...` passes; `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-16/completed.md`, `docs/deployment-plan/spike-16/deployment-plan.md` (map row status; S16-15 scope entry: "`graphprogress` reads it" becomes "`graphprogress` folds into `evidencegraphs`, whose progress and graph list read it", and "`SourceOf` on read paths" becomes "`SourceOf` stays on its write paths"), `docs/deployment-plan/spike-16/cache-architecture.md` (§ Entity records S16-15 bullet: same two changes; "and its subject" → "its Subject"), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py --links docs/deployment-plan/spike-16/deployment-plan.md docs/deployment-plan/spike-16/cache-architecture.md docs/deployment-plan/spike-16/completed.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/evidencegraphs/catalog_test.go` `TestCatalogBuild` *(new)* | `evidencelayertest.New` plus a second Source; graphs "Default", "jones", "Apple" on the first (one linked to two Artifacts), a person Subject with one Observation → `OfSource` equals `List`'s SQL order before this PR (Apple, Default, jones); `Graph(id)` equals the SQL row with `ArtifactIDs` by Artifact ref; `Progress(default)` `{1, 0, 1}`; `All()` is the lower Source id's graphs first | D-3, D-4 |
| T-2 | Go | `catalog_test.go` `TestCatalogApply` *(new)* | Table of writes through `r.Run` / helpers, each with a probe entry `DerivedFrom: evidencegraphs.Catalog` set before and `g.Verify()` after: graph create "Smith" → in `OfSource`; rename "jones" → "Zed" → last in `OfSource`; link an Artifact → in `ArtifactIDs`; unlink → gone; Subject create in "Smith" → `SubjectCount` 1; canonical entity inserted by SQL then `identityclaims.Create` accepted through `r.Run` → `FiledCount` 1; Observation create → `ObservationCount` +1; Subject delete → counts back; graph delete "Smith" → `Graph` false, `Progress` zeros, not in `OfSource`; Source rename → probe kept; second Source deleted → `OfSource` nil. The probe is emptied on every row but the Source rename | AC-1.3, AC-2.1, D-7, D-8 |
| T-3 | Go | `catalog_test.go` `TestCatalogApplyMove` *(new)* | Table: Reads order `[A, B]` and `[B, A]`; graph G's `source_id` set to B by SQL; `g.OnCommit(effects.Commit{Effects: effects.Set{Reads: …}})` → `OfSource(A)` lacks G, `OfSource(B)` holds G, `Graph(G).SourceID == B`, `Verify` passes | AC-2.2, D-7 |
| T-4 | Go | `catalog_test.go` `TestCatalogOf` *(new)* | `CatalogOf(nil)` → `database.ErrClosed`; inside `writes.Run`, `CatalogOf(graphcache.Over(tx.Tx))` → `ErrCatalogOnWrite` | AC-2.3, D-2 |
| T-5 | Go | `core/database/evidencegraphs/progress_test.go` *(new)* `TestGraphProgress`, `TestFiledCounts` (moved) | Today's cases and expectations, through `ListProgress` / `GetProgress`, including zeros, the reification-only graph, `evidencegraphs.invalid` and `ErrClosed` | AC-1.1, AC-2.3, D-6 |
| T-6 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestList` (changed) | "the whole project" and "empty id" want Default, Jones, smith, Archive; the other rows unchanged; closed catalog → `ErrClosed` | AC-2.4, AC-2.3, D-4, D-9 |
| T-7 | Go | `progress_test.go` `TestWarmReadsIssueNoQuery` *(new)* | After one `ListProgress`, `c.Graph().SetQueryCounterForTest(func(){ n++ })`; `ListProgress`, `GetProgress(def)`, `List(c, nil)`, `List(c, worked)` → `n == 0` | AC-1.2 |
| T-8 | Go | `api/ffi/handlers/evidence_graph_progress_test.go` `TestListEvidenceGraphProgress`, `TestGetEvidenceGraphProgress`, `TestEvidenceGraphProgressPerGraph`; `api/ffi/handlers/evidence_graphs_test.go` `TestListEvidenceGraphs`, `TestListEvidenceGraphsSourceID` (unchanged) | Pass as they are through the wire | AC-1.1, D-9 |
| T-9 | Go | `core/database/evidencegraphs/catalog_bench_test.go` `BenchmarkGraphCatalog` *(new)* | Direct-SQL fixture in the shape of `BenchmarkSourceRecords` and `graphcachetest.NewCanon`: 3,000 Sources each with one Artifact; 10,000 graphs (graph i on Source i mod 3,000, linked to its Artifact) each with one Citation; 20,000 person Subjects; 50,000 Observations (`value_text`, the seeded `remark` Property) → reports `build-ms` (≤ 150) and `heap-MB` (≤ 8). Then a graph with 10,000 Observations and a Source with 10,000 graphs; `g.OnCommit` with the matching read → `apply-progress-10k-ms` (≤ 20), `apply-list-10k-ms` (≤ 50) | Scale, D-7 |
| T-10 | Go | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` (changed); `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry` | `graphCatalog` measured at 4,000 and 16,000 persons; the doc row matches the registry | Scale, D-2 |

AC-3.1 is checked in the manual check and the DoD sweep.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: tests that seed evidence rows by SQL after calling `evidencegraphstest.SoleGraph` or `artifactstest.Split` (D-11), and `TestReadsDeclare` if a read table is touched by mistake.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md row and sizes; cache-architecture.md S16-15 bullet
- [ ] PR description: what shipped, AC evidence, deviations, manual check
- [ ] Sweep for this PR: `rg -n "graphprogress" core api .agents .cursor docs/*.md` finds nothing (AC-3.1); the ship log records BenchmarkCacheSizes before and after, and T-9's four figures
```

## 10. Manual check

1. `rg -n "graphprogress" core api .agents .cursor docs/*.md` prints nothing, and `ls core/database/graphprogress` fails (AC-3.1).
2. Open a project with at least two Sources and a few Evidence graphs with Subjects, one filed. Open Sources: each graph row shows the same Subject, filed and Observation counts as on `main`.
3. Open a graph, add a person Subject and an Observation, go back: that graph's counts move by one Subject and one Observation; no other graph's counts move.
4. Create a graph named "Aardvark" on a Source: it lists first among that Source's graphs. Delete it: it leaves the list.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A write path changes counted rows without naming either read, leaving the catalog stale | Verify's Check after every T-2 write; the suite with `PROVENENCIA_GRAPH_VERIFY=1` (STEP-6) |
| Many tests seed by SQL after a read and need `Drop` | D-11, counted in the report; a Stop past five files |
| The project-wide order change reaches a reader | D-4 names the Swift readers; the Go suite holds the rest |

**Stop and report if:**

- A row of §4 is no longer true at your base (in particular `graphprogress.List` / `Get` gained a caller, or `effects.Reads` changed).
- A test fails after STEP-4 for a reason other than D-11, or more than five test files need D-11's `Drop`.
- `g.Verify()` reports `evidencegraphs.catalog` mismatched after a write that went through `writes.Run`: name the write and its `Set.Reads`.
- T-9 exceeds a budget: Build over 150 ms or 8 MB, `apply-progress-10k-ms` over 20, or `apply-list-10k-ms` over 50.
- A Go reader outside `evidencegraphs` turns out to depend on the project-wide list order (D-4).
- `evidencegraphs` cannot import `graphcache` without a cycle.

## 12. Open questions

None.

## 13. PR

**Title:** S16-15 — An Evidence-graph catalog in the graph cache
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; BenchmarkCacheSizes and T-9 figures; `Plan: docs/deployment-plan/spike-16/prs/S16-15-graph-catalog.md`.
