# S12-07 — Graph-keyed reads

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 3 — Graph id end to end
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Go reads change their key; no wire, Swift or copy change)
**Board:** —
**Depends on:** S12-03 (landed [#498](https://github.com/mendahu/provenencia/pull/498)), S11-14 (landed [#493](https://github.com/mendahu/provenencia/pull/493)), S11-15 (landed [#497](https://github.com/mendahu/provenencia/pull/497)); tests use S12-04's `evidencegraphs.Create` through `evidencegraphstest.Split` (landed [#505](https://github.com/mendahu/provenencia/pull/505))
**Blocks:** S12-08
**Branch:** `s12-07-graph-keyed-reads`, branched from the commit that adds this plan (so STEP-9 edits it in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-06 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (handler bodies only; no proto change), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [data model](../source-structure.md#data-model), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) § "The layer", [`core/database/effects/README.md`](../../../../core/database/effects/README.md); working rules: "A re-keyed symbol is renamed in the PR that re-keys it", "Effects are graph-keyed, in one step", "`subjects.source_id` stays", "One bridge, one function, deleted in S12-08", "Agreement counts per Source", "High tolerance for churn", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-22 against `main` @ `e823d6c`
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

The engine's in-memory Evidence graph cache, its loader, Subject positions, event titles and graph progress read **one Evidence graph** instead of one Source, and are named for it. Two graphs of one census load and count apart, so S12-08 can hand a graph id from the wire straight to these reads. The FFI wire and Swift do not change: each handler that takes a Source id finds the Source's one graph through the S12-03 bridge (`SoleGraphTx`) and reads that graph.

**Why now.** S12-03 put `subjects.graph_id` in the catalog and S12-06 made graphs isolated, but every read still filters on `subjects.source_id`. S12-08 re-keys the RPCs, effects and Swift to graph ids in one step; it needs graph-keyed Go reads under it. S11-14 and S11-15 have landed, so `promotealign.loadLayer` and the `graphcache` shortlist index are in their final Spike 11 shape.

**What stays.** Every RPC request and response, `effects.Set.Sources` (keyed by Source until S12-08), the auto-reconciler's per-Source agreement counts, credibility, `subjects.source_id`, Promote's public signatures (`Propose`, `Evidence`, `ResolvePins` take a Source id until S12-09). The app looks and behaves exactly as before.

## 2. User stories

**US-1.** As the engine, when a Source has two Evidence graphs, I load, walk and memoize each graph on its own, so a canvas or Promote layer for one graph never holds the other graph's Subjects.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source split by `evidencegraphstest.Split` into Default (Subject Ada) and Jones (Subject Bea); `g.EvidenceGraph(def.ID)` and `g.EvidenceGraph(jones.ID)` | Default's `Subjects` is [Ada], Jones's is [Bea]; both carry the Source's id and title and `Found` true; each has its own `GraphID` |
| AC-1.2 | Both graphs loaded with a memo on each; a write that creates a Subject on Jones commits | Both graphs and both memos are dropped (the commit names the Source, D-4); a graph of another Source stays loaded |
| AC-1.3 | Both graphs loaded; the Source's title is edited | Both graphs are dropped and reload with the new title |
| AC-1.4 | Jones loaded and empty; Jones is deleted with `evidencegraphs.Delete` | Jones is no longer loaded; `g.EvidenceGraph(jones.ID)` reads `Found` false (D-5) |
| AC-1.5 | A handle with one accepted member Subject on Jones (a two-graph Source) | `graphcache.Node(...).Members[0]` carries `SourceID` = the Source and `GraphID` = Jones |

**US-2.** As the engine, I count progress, list Subject positions and title events per graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Default holds 2 persons (1 filed) and 1 Observation; Jones holds 1 place and no Observation; `graphprogress.List` | Two rows: Default {SubjectCount 2, PrimaryCount 2, FiledCount 1, ObservationCount 1}, Jones {SubjectCount 1, PrimaryCount 1, FiledCount 0, ObservationCount 0}; each row has its `GraphID` and the Source's `SourceID` |
| AC-2.2 | Same; `graphprogress.Get(c, jones.ID)` | Jones's counts only; an unknown graph id reads zeros |
| AC-2.3 | Ada placed at (1, 2) on Default, Bea at (3, 4) on Jones; `subjectpositions.ListByEvidenceGraph(c, jones.ID)` | [Bea at (3, 4)] |
| AC-2.4 | An Event on Default and an Event on Jones; `grapheventtitles.FromStored` over the loaded Jones graph | One title, for Jones's Event |

**US-3.** As a researcher, the app works exactly as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Every existing handler test for `ListSubjects`, `ListObservationsBySource`, `GetCitation`, `ListSubjectMemberships`, `ListSourceEventTitles`, `ListSubjectPositions`, `ListSourceGraphProgress`, `GetSourceGraphProgress` and the Promote RPCs | Each passes unchanged; the wire still carries Source ids |
| AC-3.2 | `ListSubjects` and `GetSourceGraphProgress` with a well-formed id no Source has | An empty list and zero counts, no error, as today (D-7) |
| AC-3.3 | Claim evidence for a handle whose member Subject is on the Jones graph of another, two-graph Source | The member is listed with the values Jones records and that Source's title |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` caches per Evidence graph: `EvidenceGraph` type and method, graph memos, `WalkEvidenceGraph`, loader and check registration, drop of every graph of a changed Source (D-1, D-2, D-4) | STEP-2, AC-1.1, AC-1.2, AC-1.3 |
| `graphcache.Member.GraphID`, loaded with the claim's Subject and checked by `Verify` (D-3) | STEP-2, AC-1.5 |
| `sourcefill` → `graphfill`, loading by `subjects.graph_id` (D-6) | STEP-2, AC-1.1 |
| `evidencegraphs.SoleGraph` *(new)*, the catalog-level call of `SoleGraphTx` (D-7) | STEP-1 |
| Handlers keep their wire and read the Source's sole graph (D-7) | STEP-3, AC-3.1, AC-3.2 |
| `promotealign` reads the layer, fixed anchors and pins from the Source's sole graph, and claim evidence members from each member's own graph (D-8) | STEP-4, AC-3.3 |
| `evidence_graphs` registry entry names its Source in `Effect.Source` (D-5) | STEP-5, AC-1.4 |
| `subjectpositions.ListBySource` → `ListByEvidenceGraph` | STEP-6, AC-2.3 |
| `sourceeventtitles` → `grapheventtitles` | STEP-6, AC-2.4 |
| `sourcegraphprogress` → `graphprogress`, grouped by `subjects.graph_id` (D-9) | STEP-7, AC-2.1, AC-2.2 |
| Two-graph tests for each read | STEP-8 |
| Docs and skills the renames make false | STEP-9 |

| Out | Home |
| --- | --- |
| Any `engine.proto` field, RPC name or Swift change; handlers taking a graph id | S12-08 |
| `effects.Set.Sources`, `Effect.Source` and `WriteEffects.sources` carrying graph ids; dropping one graph instead of every graph of its Source | S12-08 |
| Deleting `SoleGraphTx` and `SoleGraph` | S12-08 (Promote's use: S12-09) |
| `promotealign.Propose`, `Evidence`, `ResolvePins`, `promote.SaveBatch` taking a graph id; Promote's key and location | S12-09 |
| Renaming the element types `SourceSubject`, `SourceObservation`, `SourceMember`, `SourceEdge` (D-2) | not planned |
| Renaming the handler functions and RPCs (`ListSourceEventTitles`, `ListSourceGraphProgress`, `GetSourceGraphProgress`, `ListObservationsBySource`) | S12-08 |
| Evicting loaded graphs beyond a cap | S11-31 (re-worded to Evidence graphs in STEP-9) |
| Per-graph progress in the Sources list | S12-14 |
| Agreement counted per graph | not planned (working rule "Agreement counts per Source") |

## 4. Current state

Verified at `e823d6c`.

| Area | Where | Today |
| --- | --- | --- |
| Cache entry | `core/database/graphcache/source.go` `SourceGraph`, `SourceSubject`, `SourceObservation`, `SourceMember`, `SourceEdge` | One loaded Source: `SourceID`, `Title`, `Found` (the Source row exists), Subjects, Observations, accepted Members, title edges |
| Cache read | `core/database/graphcache/source.go` `(*Graph).Source(id)` | Loads through the registered `SourceLoader` on first read and stores under `g.sources[sourceID]`, a not-found Source included |
| Memos | `core/database/graphcache/source.go` `SourceMemo`, `SetSourceMemo`; `graph.go` `Graph.memos` | Per Source id, then a feature key; kept only while the Source is loaded |
| Drop | `core/database/graphcache/source.go` `dropSources`; `commit.go` `OnCommit` (`g.dropSources(set.Sources)`) | Deletes `g.sources[id]` and `g.memos[id]` for each Source id in `Set.Sources` |
| Walk | `core/database/graphcache/source.go` `WalkSource(sg, h, from)` | Walks the given loaded graph in memory; reaches only Subjects on it |
| Registration | `core/database/graphcache/source.go` `SourceLoader`, `SourceCheck`, `SetSourceLoader`, `SetSourceCheck` | `sourcefill` registers both in `init` |
| Verify | `core/database/graphcache/verify.go` `Verify` (`for key, stored := range g.sources`), `sameMembers` | Re-reads each loaded Source; compares member `SubjectID`, `SourceID`, `Accepted` |
| Members | `core/database/graphcache/load.go` `loadMembers`; `graph.go` `Member` | `SELECT c.entity_id, c.subject_id, s.source_id, c.status …`; `Member{SubjectID, SourceID, Accepted}` |
| Test hooks | `core/database/graphcache/graph.go` `LoadedSourceForTest`, `CorruptMemberSourceForTest` | By Source id |
| Loader | `core/database/sourcefill/load.go` `loadSource`, `loadSourceTitle`, `loadSourceSubjects`, `loadSourceObservations`, `loadSourceMembers`, `loadTitleEdges`, `checkSource` | Every query filters `WHERE s.source_id = ?`; title from `sources` |
| Handlers | `api/ffi/handlers/subjects.go` `ListSubjects`; `observations.go` `ListObservationsBySource`; `citations.go` `GetCitation`; `promote.go` `ListSubjectMemberships`; `event_titles.go` `ListSourceEventTitles` | Each calls `c.Graph().Source(sourceID)`; an unknown Source reads as an empty graph, so the response is empty |
| Handler helpers | `api/ffi/handlers/source_read.go` `listedProtosFromSource`, `membershipsFromSource` | Take a `graphcache.SourceGraph` |
| Positions | `core/database/subjectpositions/subjectpositions.go` `ListBySource`, `sqlListBySource`; `api/ffi/handlers/subjects.go` `ListSubjectPositions` | `WHERE s.source_id = ?` |
| Event titles | `core/database/sourceeventtitles/sourceeventtitles.go` `FromStored(sg graphcache.SourceGraph, …)` | Reads one loaded Source; no query |
| Progress | `core/database/sourcegraphprogress/sourcegraphprogress.go` `Progress`, `List`, `Get`, four SQL constants; `api/ffi/handlers/sourcegraphprogress.go` `ListSourceGraphProgress`, `GetSourceGraphProgress`, `protoProgress` | Grouped by `s.source_id`; `Get` of an unknown Source is zeros |
| Promote layer | `core/database/promotealign/promotealign.go` `loadInputs`, `mergeFixed`; `layer.go` `loadLayerMemo`, `loadLayer`, `loadLayerBridges` | `c.Graph().Source(sourceID)`; layer memo under the Source id; `WalkSource` |
| Claim evidence | `core/database/promotealign/evidence.go` `evidence` (`rowSG`), `evidenceReader.sourceOf`, `accepted`, `memberMemo`, `memberValues`, `bridges`, `titles` | Reads each member's **Source** graph, grouped by `Member.SourceID`; member memo under the Source id; titles from `SourceGraph.Title` |
| Pins | `core/database/promotealign/pins.go` `ResolvePins` | `c.Graph().Source(sourceID)` for the filed check |
| Bridge | `core/database/evidencegraphs/evidencegraphs.go` `SoleGraphTx(tx *sql.Tx, sourceID)`; `List(c, sourceID)` | `SoleGraphTx` returns `ErrInvalid` for a Source with no graph and a plain error for more than one; `List` shows the read-transaction pattern over a catalog |
| Effects | `core/database/effects/registry.go` `"evidence_graphs"` entry | `SourceScope` and `Lists` only; no `Source`, so a graph write drops no cached graph |
| Index | `core/database/migrations/000055.sql` `subjects_graph_id_idx` | The loader's new filter has an index; no migration |
| Tests | `core/database/graphcache/source_memo_test.go` `TestSourceMemoDroppedWithSource`; `source_walk_test.go` `TestWalkSourceEdges`, `TestWalkSourceStaysOnItsSource`, `TestMemberSourceID`; `core/database/sourcefill/load_test.go` `TestRetitleDropsSourceGraph`, `TestOnlyATitleChangeDropsSourceGraph`; `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions`; `core/database/sourceeventtitles/sourceeventtitles_test.go` `TestForSource`; `core/database/sourcegraphprogress/sourcegraphprogress_test.go` `TestSourceGraphProgress`, `TestPrimaryAndFiledCounts`; `core/database/promotealign/source_store_test.go`, `memo_test.go`, `export_test.go` `MemoSnapshotForTest`; `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve`; `api/ffi/handlers/sourcegraphprogress_test.go` | All keyed by Source id |
| Test callers of `g.Source` outside these packages | `core/database/identityclaims/identityclaims_test.go` (line 371), `core/database/subjects/subjects_test.go` (line 509) | Read a Source graph by Source id |
| Fixture | `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` `Split`, `SoleGraph`, `Create` | Makes a second graph and moves Artifacts onto it |

**Precedent.** `58e2f5da` ("Keep a Source's title and existence on its cached graph, drop the graph on a retitle") is the shape of D-1 and D-5: what the cache entry stores decides which registry `Effect.Source` path drops it, and `TestOnlyATitleChangeDropsSourceGraph` tests the drop through a real write. Copy that test's write-then-`LoadedSourceForTest` shape for AC-1.2 to AC-1.4. `4afd43a1` (S11-19) moved the walk in memory over one loaded graph; keep its walk body unchanged here, only its name and parameter type move.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `graphcache` stores loaded graphs in `g.evidenceGraphs map[string]EvidenceGraph` keyed by graph id, and memos in `g.memos` keyed by graph id. `EvidenceGraph` *(renamed from `SourceGraph`)* has `GraphID`, `SourceID`, `SourceTitle` *(renamed from `Title`)*, `Found` (the graph row exists), and the same `Subjects`, `Observations`, `Members`, `EventPeople`, `EventPlaces`. `(*Graph).EvidenceGraph(graphID)` sets `GraphID`; the loader sets `SourceID` and `SourceTitle` from the graph's Source | Working rule "A re-keyed symbol is renamed"; the rule names `graphcache` `EvidenceGraph(id)`. A graph's Source is part of its row, so the loader reads it in the same query as the title | Keeping `Source(id)` and adding `EvidenceGraph(id)` beside it — the rule "replaces Source-keyed fields instead of adding graph ones"; keeping `Title` — it is the Source's title, not the graph's name |
| D-2 | Renamed with the key: `Source` → `EvidenceGraph`, `SourceGraph` → `EvidenceGraph`, `SourceMemo` / `SetSourceMemo` → `EvidenceGraphMemo` / `SetEvidenceGraphMemo`, `WalkSource` → `WalkEvidenceGraph`, `SourceLoader` / `SetSourceLoader` → `EvidenceGraphLoader` / `SetEvidenceGraphLoader`, `SourceCheck` / `SetSourceCheck` → `EvidenceGraphCheck` / `SetEvidenceGraphCheck`, `LoadedSourceForTest` → `LoadedEvidenceGraphForTest`, `dropSources` → `dropGraphsOfSources`. Files: `core/database/graphcache/source.go` → `evidence_graph.go` *(new)*, `source_memo_test.go` → `evidence_graph_memo_test.go` *(new)*, `source_walk_test.go` → `evidence_graph_walk_test.go` *(new)* (`git mv`). The element types `SourceSubject`, `SourceObservation`, `SourceMember`, `SourceEdge` keep their names | These are the symbols whose key changes. The element types are rows of the Source's document (a `SourceSubject` still carries `SourceID`), not keyed reads; renaming them touches about 20 more sites in `promotealign` and the handlers for no change of meaning | Renaming the element types too — churn with no re-key; a type named `EvidenceGraphs` — clashes with `evidencegraphs` the package |
| D-3 | `graphcache.Member` gains `GraphID` *(new)*, read in `loadMembers` as `s.graph_id` beside `s.source_id`; `sameMembers` compares it. `CorruptMemberSourceForTest` stays | Claim evidence must find a member's graph without a query per member (D-8); a Subject never changes graph (working rule "Graphs are fully isolated"), so the stored id never goes stale | A `SubjectGraphTx` query per member — a query per row on every evidence sheet, against S11-19's "a warm run issues no graph query" |
| D-4 | `OnCommit` calls `g.dropGraphsOfSources(set.Sources)`: it removes every loaded graph whose `SourceID` is in the set, with its memos, by scanning `g.evidenceGraphs`. No query | Working rule "Effects are graph-keyed, in one step": `Set.Sources` stays keyed by Source until S12-08, and `graphcache` drops every graph of a changed Source. The loaded map is small (S11-31 will cap it) | A Source → graphs index on `Graph` — a second structure to keep in step for one PR's life; S12-08 replaces the scan with a direct delete by graph id |
| D-5 | The `"evidence_graphs"` registry entry gains `Source: field("source_id")`. `"evidence_graph_artifacts"` stays without one | `Found` now means the graph row exists, so creating or deleting a graph row must drop that Source's loaded graphs, or a deleted graph reads as an empty, found graph. A rename drops them too, which costs one reload. A cached graph holds no links, so a link write needs no drop. S12-03 D-5 left `Source` off because "No read loads these tables until S12-05 / S12-07"; this PR is that read | Not caching a not-found graph and leaving the stale delete to S12-08 — defers a defect in code this PR adds; dropping every graph on `ListEvidenceGraphs` — every Artifact add would drop every loaded Promote layer |
| D-6 | `core/database/sourcefill` becomes `core/database/graphfill` (`git mv`; package comment "loads one Evidence graph"). `loadEvidenceGraph(g, graphID)` reads: `SELECT g.source_id, s.title FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id = ?` (no row → `Found` false, everything else empty, no further query); Subjects, Observations and Members filter `WHERE s.graph_id = ?` with their current `ORDER BY`; `loadTitleEdges` calls `WalkEvidenceGraph`. `checkEvidenceGraph` reloads by graph id. Functions are renamed `loadGraphTitle` → reads `(sourceID, title, found)`, `loadGraphSubjects`, `loadGraphObservations`, `loadGraphMembers` | Working rule "renamed to match"; `subjects_graph_id_idx` serves the filter | `evidencegraphfill` — longer with no gain; the rule's own example is `graphprogress` |
| D-7 | `evidencegraphs.SoleGraph(c *database.Catalog, sourceID []byte) ([]byte, error)` *(new)* opens a read transaction and returns `SoleGraphTx`, the way `List` wraps `list`. It carries the same "S12-08 deletes it" comment. The handlers get one helper, `sourceEvidenceGraph(c, sourceID) (graphcache.EvidenceGraph, error)` *(new)* in `api/ffi/handlers/source_read.go`: `SoleGraph`, then `c.Graph().EvidenceGraph(id)`; when `SoleGraph` returns `evidencegraphs.ErrInvalid` (a Source with no graph, so no Source) it returns `graphcache.EvidenceGraph{}` and no error, so an unknown Source reads empty as today. `ListSubjectPositions` and `GetSourceGraphProgress` call `SoleGraph` and return an empty list / zero counts carrying the request's Source id on `ErrInvalid` | Working rule "One bridge, one function": resolution stays in `SoleGraphTx`; `SoleGraph` only opens the transaction, as `List` does for `list`. Six handlers would otherwise repeat Begin / Rollback / error mapping. The unknown-Source behaviour is what the app sees today (AC-3.2) | A `graphcache` method taking a Source id — a Source-keyed read beside the graph one; returning an error for an unknown Source — a behaviour change no PR asked for |
| D-8 | `promotealign`: `inputs` gains `graphID`, set in `loadInputs` from `evidencegraphs.SoleGraph` (its `ErrInvalid` becomes `promote.ErrInvalid`); `loadLayerMemo`, `loadLayer`, `mergeFixed`, `evidence`'s row graph and `ResolvePins`'s filed check take `in.graphID`. `evidenceReader` gains `graphOf map[string][]byte` (Subject → graph) and `graphOfSource map[string][]byte` (Source → the first member graph seen), both filled in `accepted` from `Member.GraphID`; `memberValues` groups by graph and memoizes under the graph id; `bridges` reads each member's graph once; `titles` reads `SourceTitle` from `graphOfSource[source]`. `claimevidence.Member.Source` and `Bridge.Source` stay Source ids | Scope "`promotealign.loadLayer` reads its graph through `SoleGraphTx` until S12-09". A member of another Source may sit on any of its graphs (AC-3.3), so evidence must read the member's own graph, not the Source's sole one, which fails for a two-graph Source. Titles stay per Source (the evidence names Sources) | Resolving members through `SoleGraph` — fails on a two-graph Source; adding graph ids to `claimevidence` — wire-facing, S12-09 |
| D-9 | `core/database/sourcegraphprogress` becomes `core/database/graphprogress` (`git mv`). `Progress` gains `GraphID` and keeps `SourceID`. `List` groups by `s.graph_id` and reads the Source through `JOIN evidence_graphs eg ON eg.id = s.graph_id` (one row per graph with any counted Subject; graphs with none are omitted, as Sources are today). `Get(c, graphID)` filters `WHERE s.graph_id = ?` and reads `SourceID` with one `SELECT source_id FROM evidence_graphs WHERE id = ?` (no row: zeros, `SourceID` empty). `ErrInvalid` keeps `CodeSourcesInvalid`. `ListSourceGraphProgress` writes one wire row per graph with `source_id` = `row.SourceID` — one per Source while each Source has one graph | Working rule names `graphprogress` and "Per-graph progress is its own key, keyed by graph id"; S12-14 reads it per graph after S12-08 moves the wire | A new error code — no new failure exists; summing graphs per Source in the handler — an adapter the churn rule forbids, and unreachable before a second graph has UI |
| D-10 | `subjectpositions.ListBySource` → `ListByEvidenceGraph(c, graphID)`, `sqlListBySource` → `sqlListByGraph` with `WHERE s.graph_id = ?`. `sourceeventtitles` → `grapheventtitles` (`git mv`), `FromStored(eg graphcache.EvidenceGraph, …)`; its doc comment says "one Evidence graph". | Working rule "renamed to match" | Leaving the names — the rule binds this PR |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache.EvidenceGraph` type and `(*Graph).EvidenceGraph` *(renamed)* | type / func | `graphcache.SourceGraph`, `(*Graph).Source` | It **is** that code, re-keyed and renamed (D-1, D-2) |
| `graphcache.Member.GraphID` *(new)* | field | `Member.SourceID` | Extends the same row and query (D-3) |
| `(*Graph).dropGraphsOfSources` *(renamed)* | func | `dropSources` | Same job, now a scan by `SourceID` (D-4) |
| `evidencegraphs.SoleGraph` *(new)* | func | `evidencegraphs.SoleGraphTx`, `List` / `list` | Wraps `SoleGraphTx` in a read transaction the way `List` wraps `list`; resolution stays in one function (D-7) |
| `handlers.sourceEvidenceGraph` *(new)* | func | `listedProtosFromSource` in the same file | The one place the handlers map a Source id to its loaded graph until S12-08 (D-7) |
| `core/database/graphfill` *(renamed)* | package | `core/database/sourcefill` | Renamed, filter moves to `graph_id` (D-6) |
| `core/database/grapheventtitles` *(renamed)* | package | `core/database/sourceeventtitles` | Renamed (D-10) |
| `core/database/graphprogress` *(renamed)*, `Progress.GraphID` *(new)* | package / field | `core/database/sourcegraphprogress` | Renamed and regrouped (D-9) |
| `subjectpositions.ListByEvidenceGraph` *(renamed)* | func | `ListBySource` | Renamed, filter moves (D-10) |
| `evidenceReader.graphOf`, `graphOfSource` *(new)* | field | `evidenceReader.sourceOf` | Same fill point (`accepted`) and same lifetime (D-8) |
| Two-graph test cases | test | `evidencegraphstest.Split` | Reused as the fixture for every new case; no new helper |

**Builds on.** `effects` registry paths (`field("source_id")`, as `subjects` has); the `List` / `list` read-transaction pattern; the S11-19 in-memory walk and memo slots; `evidencegraphstest.Split`.
**Deletes.** `graphcache` `Source`, `SourceGraph`, `SourceMemo`, `SetSourceMemo`, `WalkSource`, `SourceLoader`, `SetSourceLoader`, `SourceCheck`, `SetSourceCheck`, `LoadedSourceForTest`, `dropSources`, `g.sources`; packages `sourcefill`, `sourceeventtitles`, `sourcegraphprogress`; `subjectpositions.ListBySource`; every `WHERE s.source_id = ?` in those reads. The `TestEvidenceGraphsResolve` assertion that graph rows set no `Sources`.
**Next adds.** S12-08 changes `sourceEvidenceGraph`'s callers to `c.Graph().EvidenceGraph(graphID)` and deletes the helper and `SoleGraph`; `dropGraphsOfSources` becomes a delete by graph id when `Set` carries graph ids. S12-09 replaces `loadInputs`'s `SoleGraph` with the request's graph id. S12-14 reads `graphprogress` rows per graph. S11-31 caps `g.evidenceGraphs`.

### Schema

None — `subjects.graph_id` and `subjects_graph_id_idx` exist (migration 55).

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | D-1 – D-4. `func (g *Graph) EvidenceGraph(graphID []byte) (EvidenceGraph, error)`: same validation and error strings with "evidence graph" for "source"; `func (g *Graph) WalkEvidenceGraph(eg EvidenceGraph, h hops.Hop, from [][]byte) ([]SourceEdge, error)`: body unchanged. `Verify` iterates `g.evidenceGraphs` with `checkFn(g, graphID, stored)`. `clear` resets `evidenceGraphs` |
| `core/database/graphfill` | D-6 |
| `core/database/evidencegraphs` | `func SoleGraph(c *database.Catalog, sourceID []byte) ([]byte, error)` *(new)*, D-7 |
| `core/database/effects` | `"evidence_graphs"`: `Source: field("source_id")` (D-5). `README.md` sentence updated (STEP-9) |
| `core/database/promotealign` | D-8. `export_test.go` `MemoSnapshotForTest(c, graphID, memberGraphID []byte)` |
| `core/database/subjectpositions` | D-10 |
| `core/database/grapheventtitles` | D-10 |
| `core/database/graphprogress` | D-9. SQL sketch: |

```sql
-- sqlCountSubjectsByType
SELECT s.graph_id, eg.source_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
FROM subjects s
JOIN evidence_graphs eg ON eg.id = s.graph_id
JOIN subject_types t ON t.id = s.subject_type_id
LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
GROUP BY s.graph_id, eg.source_id, t.key, t.origin
-- sqlCountObservationsByGraph
SELECT s.graph_id, eg.source_id, COUNT(*)
FROM observations o JOIN subjects s ON s.id = o.subject_id
JOIN evidence_graphs eg ON eg.id = s.graph_id
GROUP BY s.graph_id, eg.source_id
```

Errors: none new. `SoleGraph` returns `SoleGraphTx`'s errors unchanged.

### FFI

None — no `engine.proto` change. Handler bodies only: `ListSubjects`, `ListObservationsBySource`, `GetCitation`, `ListSubjectMemberships`, `ListSourceEventTitles` call `sourceEvidenceGraph`; `listedProtosFromSource` and `membershipsFromSource` take `graphcache.EvidenceGraph`; `ListSubjectPositions` calls `SoleGraph` then `subjectpositions.ListByEvidenceGraph`; `ListSourceGraphProgress` / `GetSourceGraphProgress` / `protoProgress` use `graphprogress` (D-7, D-9). No generated code changes, so `./scripts/generate-proto.sh` is not run.

### macOS

None — no Swift change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — `SoleGraph` over a catalog**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`. Skill: `add-catalog-query`.
Do: add `SoleGraph` (D-7) directly under `SoleGraphTx`, with the doc comment "The catalog read of SoleGraphTx. S12-08 deletes both." Add T-1's cases.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/` passes.

**STEP-2 — Key the graph cache and its loader by graph**
Files: `core/database/graphcache/{source.go → evidence_graph.go, graph.go, commit.go, verify.go, load.go, source_memo_test.go → evidence_graph_memo_test.go, source_walk_test.go → evidence_graph_walk_test.go}`, `core/database/sourcefill/ → core/database/graphfill/` (`load.go`, `load_test.go`), and every caller so the tree builds: `api/ffi/handlers/{source_read.go, subjects.go, observations.go, citations.go, promote.go, event_titles.go}`, `core/database/promotealign/{promotealign.go, layer.go, evidence.go, pins.go, export_test.go, source_store_test.go, memo_test.go, promotealign_test.go}`, `core/database/sourceeventtitles/sourceeventtitles.go` (parameter type only; the package moves in STEP-6), `core/database/identityclaims/identityclaims_test.go`, `core/database/subjects/subjects_test.go`. Skill: `use-catalog-session`.
Do: D-1, D-2, D-3, D-4, D-6 in `graphcache` and `graphfill`; the handler helper and its callers (D-7); `promotealign` per D-8. Change every blank and named import of `database/sourcefill` to `database/graphfill`. Tests that read a Source graph get the graph id from `evidencegraphstest.SoleGraph(t, c, sourceID).ID`. Rename `TestSourceMemoDroppedWithSource` → `TestEvidenceGraphMemoDroppedWithSource`, `TestWalkSourceEdges` → `TestWalkEvidenceGraphEdges`, `TestWalkSourceStaysOnItsSource` → `TestWalkEvidenceGraphStaysOnItsGraph`, `TestMemberSourceID` → `TestMemberSourceAndGraph` (asserts `GraphID` too, T-4), `TestRetitleDropsSourceGraph` → `TestRetitleDropsEvidenceGraph`, `TestOnlyATitleChangeDropsSourceGraph` → `TestOnlyATitleChangeDropsEvidenceGraph`, `TestSourceGraphSecondReadIssuesNoQueries` → `TestEvidenceGraphSecondReadIssuesNoQueries`.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` is clean; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `rg -n '\.Source\(|SourceMemo|WalkSource|SourceGraph\b|LoadedSourceForTest|sourcefill' core api` prints nothing (`Source(` on other types, such as `deleteimpact`, is not a match only when the receiver is not a `*graphcache.Graph`; check each remaining hit).

**STEP-3 — Behaviour stays for an unknown Source**
Files: `api/ffi/handlers/subjects_test.go`, `api/ffi/handlers/sourcegraphprogress_test.go`.
Do: add T-9 and T-10 (AC-3.2). They pass on STEP-2's code; they pin D-7's mapping.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/ -run 'TestListSubjects|TestGetSourceGraphProgress'` passes.

**STEP-4 — Claim evidence reads each member's graph**
Files: `core/database/promotealign/evidence_test.go`. Skill: —.
Do: add T-8 (AC-3.3). STEP-2 made the change; this step proves the two-graph case.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/` passes.

**STEP-5 — A graph row drops its Source's graphs**
Files: `core/database/effects/registry.go`, `core/database/effects/evidence_graphs_test.go`, `core/database/graphfill/load_test.go`. The `graphfill` and renamed `graphcache` test paths are *(new)* (moved by this PR).
Do: D-5. In `TestEvidenceGraphsResolve`, the first `set` (graph + link create) now has `Sources` = [the new Source]; the rename and delete sets have `Sources` = [the Source]; add a link-only resolve (`evidencegraphs.Unlink` changes) whose `Sources` is empty (T-7). In `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity`, the `evidence_graphs` row becomes `{"evidence_graphs", true, true, false}` (T-7). Add T-6 (AC-1.4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/database/graphfill/` passes.

**STEP-6 — Positions and event titles by graph**
Files: `core/database/subjectpositions/subjectpositions.go`, `subjectpositions_test.go`; `core/database/sourceeventtitles/ → core/database/grapheventtitles/` (`git mv`), its test; `api/ffi/handlers/subjects.go` `ListSubjectPositions`; `api/ffi/handlers/event_titles.go`.
Do: D-10. `TestForSource` → `TestForEvidenceGraph`. Add T-11 and T-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectpositions/ ./core/database/grapheventtitles/ ./api/ffi/handlers/` passes.

**STEP-7 — Progress by graph**
Files: `core/database/sourcegraphprogress/ → core/database/graphprogress/` (`git mv` both files), `api/ffi/handlers/sourcegraphprogress.go`. Skill: `add-catalog-query`.
Do: D-9. `TestSourceGraphProgress` → `TestGraphProgress`; its expectations gain `GraphID`. Add T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphprogress/ ./api/ffi/handlers/` passes; `rg -n 'sourcegraphprogress|sourceeventtitles' core api` prints nothing.

**STEP-8 — Two graphs load apart**
Files: `core/database/graphfill/load_test.go`, `core/database/graphcache/evidence_graph_memo_test.go`. The `graphfill` and renamed `graphcache` test paths are *(new)* (moved by this PR).
Do: add T-2, T-3 and T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/ ./core/database/graphcache/` passes; `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/ ./core/database/promotealign/` passes (Verify re-reads every loaded graph after each commit).

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row; S12-07 scope **Check**), this plan (status, Landed note), `core/database/effects/README.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-11/deployment-plan.md` (S11-31 scope entry), `.agents/skills/review-pr/lenses/conformance.md`, `.agents/skills/review-pr/lenses/maintainability.md`, `.agents/skills/review-pr/lenses/performance.md`.
Do: per the Definition of done. `core/database/effects/README.md`: graph rows (`evidence_graphs`) now set `Effect.Source`, so a graph create, edit or delete drops the Source's loaded Evidence graphs; link rows still do not; `Effect.Source` "is the Source whose graphs changed (graphcache drops every loaded Evidence graph of it; graph progress counts it)". `promote-graph-alignment.md` § "The layer" item 1: `g.Source` → `g.EvidenceGraph` on "the Source's one Evidence graph" (S12-09 rewrites the paragraph). S11-31: "Source graphs" → "Evidence graphs", `SourceMemo` → `EvidenceGraphMemo`, `g.sources` → `g.evidenceGraphs`. The three lens files: `sourcegraphprogress` → `graphprogress`.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-07-graph-keyed-reads.md` reports no errors; `rg -n 'sourcegraphprogress|sourcefill|sourceeventtitles|SourceMemo|WalkSource' docs .agents core api --glob '!docs/deployment-plan/spike-1[12]/prs/**' --glob '!docs/deployment-plan/archive/**'` prints only this spike's map history lines.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestSoleGraphTx` *(cases added)* | Each existing row also runs through `SoleGraph(f.c, …)` with the same result; a new row: a Source with a second graph from `evidencegraphs.Create` → an error from both | D-7 |
| T-2 | Go | `core/database/graphfill/load_test.go` `TestTwoEvidenceGraphsLoadApart` *(new)* | `Split` Default (Ada) / Jones (Bea) → each graph's `Subjects`, `GraphID`, `SourceID`, `SourceTitle`, `Found` per AC-1.1; Observations on Bea appear only on Jones | AC-1.1 |
| T-3 | Go | `core/database/graphfill/load_test.go` `TestSourceWriteDropsEveryGraph` *(new)* | Both graphs and a graph of a second Source loaded, a memo on each; `subjects.Create` on Jones through `writes.Run` → Default and Jones not loaded, memos gone; the other Source's graph and memo kept | AC-1.2, D-4 |
| T-4 | Go | `core/database/graphcache/evidence_graph_walk_test.go` *(new)* `TestMemberSourceAndGraph` *(renamed)* | James promoted from the fixture Source → `Members[0].SourceID` = Source, `GraphID` = its graph; a second case with James on a `Split` graph → that graph | AC-1.5, D-3 |
| T-5 | Go | `core/database/graphfill/load_test.go` *(new)* `TestRetitleDropsEvidenceGraph` *(renamed)* | Gains a `Split` case: retitle → both graphs dropped; next read has the new `SourceTitle` | AC-1.3 |
| T-6 | Go | `core/database/graphfill/load_test.go` `TestDeletedGraphIsNotFound` *(new)* | Jones (no Subjects) loaded; `evidencegraphs.Delete` through `writes.Run` → `LoadedEvidenceGraphForTest(jones)` false; `EvidenceGraph(jones)` → `Found` false | AC-1.4, D-5 |
| T-7 | Go | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve`, `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` *(changed)* | Graph create, rename, delete → `Sources` = [Source]; link-only change → `Sources` empty; the registry row for `evidence_graphs` has `Source` set | D-5 |
| T-8 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsMemberGraph` *(new)* | Row Source R; handle H with an accepted member Subject on the Jones graph of Source X (X `Split` in two) → `Evidence` lists the member with Jones's values and `Sources[X]` = X's title; no error | AC-3.3, D-8 |
| T-9 | Go | `api/ffi/handlers/subjects_test.go` `TestListSubjects` *(case added)* | A fresh UUIDv7 no Source has → empty `Subjects`, no error | AC-3.2, D-7 |
| T-10 | Go | `api/ffi/handlers/sourcegraphprogress_test.go` `TestGetSourceGraphProgress` *(case added)* | Unknown Source id → zero counts, `SourceId` = the request's id | AC-3.2, D-7, D-9 |
| T-11 | Go | `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions` *(case added)* | Ada on Default at (1, 2), Bea on Jones at (3, 4) → `ListByEvidenceGraph(jones)` = [Bea (3, 4)], Default = [Ada (1, 2)] | AC-2.3 |
| T-12 | Go | `core/database/grapheventtitles/grapheventtitles_test.go` *(new)* `TestForEvidenceGraph` *(case added)* | One Event on each graph → the loaded Jones graph titles only Jones's Event | AC-2.4 |
| T-14 | Go | `api/ffi/handlers/` — every existing test of the RPCs in AC-3.1 | Unchanged expectations → pass | AC-3.1 |
| T-13 | Go | `core/database/graphprogress/graphprogress_test.go` *(new)* `TestGraphProgress` *(case added)* | The AC-2.1 catalog → two rows as listed; `Get(jones)` → Jones's counts; `Get(unknown)` → zeros | AC-2.1, AC-2.2, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (memo and evidence tests), `api/ffi/handlers` promote and citation tests, `core/database/graphcache` Verify under `PROVENENCIA_GRAPH_VERIFY=1`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — none: no brief
- [ ] Domain doc updated if a rule in docs/*.md changed: core/database/effects/README.md (graph rows set Effect.Source), docs/promote-graph-alignment.md (layer read name), S11-31 scope entry, the three review-pr lens files
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-3.1) Open an existing project with a Source that has Subjects, Citations and promoted Subjects. Open its Evidence graph: the same bubbles sit in the same places, and event cards show the same titles.
2. Open the Sources list: that Source's progress counts are the same as on `main`.
3. Open Promote on that Source and open a row's evidence sheet: the same rows, members and Source titles as on `main`.
4. Add a Subject on the canvas: it appears at once, and the Sources list count moves by one.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A caller of `g.Source` is missed and still keys by Source | The method is deleted, so the compiler lists every caller (STEP-2 check) |
| `Verify` diverges because a stored graph and a fresh read differ in `GraphID` / `SourceID` placement | STEP-8 runs the graph-touching packages with `PROVENENCIA_GRAPH_VERIFY=1` |
| Claim evidence opens one query per member | D-3 reads `GraphID` with the members; `TestEvidenceWarmIssuesNoQuery` must stay green |
| A graph write now reports its Source in `WriteEffects.sources`, so Swift reloads that Source's graph | No Swift view writes graphs before S12-11; the reload is the correct refresh once it does |

**Stop and report if:**

- `evidencegraphs.SoleGraphTx` no longer returns `ErrInvalid` for a Source with no graph, or `List` no longer opens its own read transaction (D-7 is built on both).
- `TestEvidenceWarmIssuesNoQuery`, `TestProposeWarmIssuesNoQuery` or `TestEvidenceGraphSecondReadIssuesNoQueries` fails after STEP-2 and passes on `main` (a graph read issues a query a memo used to save).
- Any handler test in AC-3.1 needs its expected output changed to pass.
- A test in `core/database/effects` other than the two STEP-5 names (`TestEvidenceGraphsResolve`, `TestSourceEvidenceIsNotActivity`) fails on `Source` for `"evidence_graphs"`.
- A query in D-6 or D-9 needs a column `subjects_graph_id_idx` does not cover to stay under the existing `TestPragmaHonesty` / query-plan checks.
- Something outside §4 calls `subjectpositions.ListBySource`, `sourcegraphprogress` or `sourceeventtitles` (the rename would then reach code this plan did not read).

## 12. Open questions

None.

## 13. PR

**Title:** S12-07 — Graph-keyed reads
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-07-graph-keyed-reads.md`.
