# S12-04 — Graph lifecycle

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 2 — Graphs in the catalog
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: one Go package, the delete registries and tests; no FFI, no Swift)
**Board:** —
**Depends on:** S12-03 (landed [#498](https://github.com/mendahu/provenencia/pull/498))
**Blocks:** S12-05
**Branch:** `s12-04-graph-lifecycle`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-03 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`
**Context:** [rules](../source-structure.md#rules), [`catalog-deletes.md`](../../../catalog-deletes.md) § Policy item 10 and § Facet release; working rules: "`core/database/evidencegraphs` owns graphs and links from S12-03", "Every Source has at least one graph", "Name and description are editable; graphs list alphabetically by name", "Refusals are of two kinds", "Artifact ↔ graph links", "One graph list", "One bridge, one function, deleted in S12-08"; the S12-03 [Landed note](S12-03-evidence-graphs-catalog.md) and [ship log](S12-03-evidence-graphs-catalog.ship.md) (MNT-8)
**Planned:** 2026-10-19 against `main` @ `3bd0862`
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

The engine can manage Evidence graphs: list them, create a graph linked to chosen Artifacts, rename it or change its description, and delete it under the rules ([design](../source-structure.md#rules)). One read, `evidencegraphs.List`, returns full graph rows with their linked Artifact ids, sorted by name, for one Source or the whole project; it is the only read of graph rows that S12-05's RPC and every later screen use. Deletes go through `deleteimpact`: a graph with Subjects is refused by the inbound probe S12-03 registered, and a Source's last graph is refused by a domain invariant.

**Why now.** S12-05 wraps these functions in RPCs and the graph list key; slice 4 builds the Graphs section and graph form on them. S12-03 left `KindEvidenceGraph` registered but not deletable, and four test files with their own copy of a `defaultGraphID` helper (S12-03 review MNT-8).

**What stays.** No RPC, no screen and no schema change; nothing changes for researchers. A second graph can now be made in Go, but no RPC exposes it until S12-05 and no screen until S12-12. `evidencegraphs.SoleGraphTx` keeps refusing a Source with two graphs (an internal error), so `subjects.Create` on such a Source fails until S12-08 lets a Subject name its graph; no production path can reach that state before S12-12.

## 2. User stories

**US-1.** As the engine, I read a Source's graphs, or every graph in the project, in one call.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source S has graphs named `smith`, `Default` and `Jones` (made with `Create`); `List(c, S.id)` | Three `Graph` rows in the order `Default`, `Jones`, `smith` (case-insensitive by name), each with `ID`, `Ref`, `SourceID`, `Name`, `Description` |
| AC-1.2 | `Jones` is linked to Artifacts `ART-BBBBB` and `ART-AAAAA` (created in that order) | `Jones.ArtifactIDs` lists `ART-AAAAA`'s id, then `ART-BBBBB`'s (by Artifact ref, case-insensitive); a graph with no link has `ArtifactIDs` nil |
| AC-1.3 | A second Source T with one graph `Archive`; `List(c, nil)` | All four graphs, ordered `Archive`, `Default`, `Jones`, `smith` regardless of Source |
| AC-1.4 | `List(c, <unknown 16-byte id>)` | Empty slice, no error |
| AC-1.5 | `List(c, <15-byte id>)`; `List` on a closed catalog | `ErrInvalid`; `database.ErrClosed` |
| AC-1.6 | Inside one write, `Create` a graph and then `ListTx(tx.Tx, S.id)` | The new graph is in the result |

**US-2.** As the engine, I create a graph linked to at least one Artifact of its Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `Create` on S with `Name: "  Jones household  "`, `Description: ""`, `ArtifactIDs: [A2, A1]` | One `evidence_graphs` row: new UUIDv7 id, valid `GPH-` ref, name `Jones household`, no description. Links (graph, A2) and (graph, A1). Changes `[evidence_graph create, evidence_graph_artifact create (A2), evidence_graph_artifact create (A1)]`. The returned `Graph.ArtifactIDs` is `[A2, A1]` |
| AC-2.2 | `Create` with a name that is empty or only spaces; with no `ArtifactIDs`; with the same Artifact id twice; with a 15-byte Artifact id | `ErrInvalid`; no graph row |
| AC-2.3 | `Create` with an Artifact of another Source, or an unknown Artifact id | `ErrInvalid`; after `writes.Run` returns, no new graph row and no new link |
| AC-2.4 | `Create` on an unknown Source id, a 15-byte Source id, or with an empty user id | `ErrInvalid` |
| AC-2.5 | `Create` with the name `Default` on a Source that already has a `Default` graph | Succeeds: names are not unique (D-6) |

**US-3.** As the engine, I rename a graph and change its description.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `Update` the `Default` graph to `Name: "Smith household"`, same description | One `evidence_graph` update change whose `Fields` holds only `name` (`Default` → `Smith household`); `List` shows the new name |
| AC-3.2 | `Update` with `Description: " Pages 4–6 "` on a graph with none | One update change with `description` old `nil`, new `Pages 4–6`; then `Description: ""` → `description` old `Pages 4–6`, new `nil` |
| AC-3.3 | `Update` with the stored name and description (after trimming) | No changes, no error |
| AC-3.4 | `Update` with an empty name, an unknown graph id, a 15-byte id, or an empty user id | `ErrInvalid`; the row is unchanged |

**US-4.** As the engine, I delete an empty graph that is not its Source's last, and refuse the rest.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | S has `Default` and `Jones`; `Jones` is linked to A1 and has no Subjects; `Delete(Jones)` | `Jones` and its link rows are gone; A1 and `Default` stay. Changes `[evidence_graph_artifact delete (Jones, A1), evidence_graph delete]`; the graph delete change's `Fields` carry `id`, `ref`, `source_id`, `name`, `description` |
| AC-4.2 | S's only graph, with or without Subjects; `Delete` | `ErrLastGraph` (`evidencegraphs.last_graph`, user kind); the graph and its links stay |
| AC-4.3 | S has `Default` holding Subject P and an empty `Jones`; `Delete(Default)` | `ErrInUse` (`evidencegraphs.in_use`, conflict kind); nothing changes. `deleteimpact.Impact(tx, KindEvidenceGraph, Default)` reports gate `inbound` with one group, via `subjects.graph_id`, total 1, listing P |
| AC-4.4 | `deleteimpact.Impact` on the empty `Jones` | `Allowed`, gate `ok`, no groups |
| AC-4.5 | `Delete` of an unknown graph id, a 15-byte id, or with an empty user id; `Impact` on an unknown graph id | `ErrInvalid`; `Impact` gate `not_found` |

**US-5.** As the engine, graph writes are audited under their Source and refresh the graph list.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `Create`, `Update` and `Delete` of a graph on Source A | Each write bumps Source A's scope only; `effects.Resolve` over each write's changes has `ListEvidenceGraphs` in `Set.Lists` |

## 3. Scope

| In | Covered by |
| --- | --- |
| Error codes `evidencegraphs.in_use` (conflict) and `evidencegraphs.last_graph` (user); `ErrInUse`, `ErrLastGraph` | STEP-1, AC-4.2, AC-4.3 |
| `Graph.ArtifactIDs`; `List`, `ListTx` over one unexported `list`; `graphIDsTx` goes, `SoleGraphTx` and `LinkArtifact` read through `list` | STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6 |
| `evidencegraphstest.DefaultGraphID` over `List`; the four `defaultGraphID` copies go (S12-03 review MNT-8) | STEP-2 |
| `Create` with `CreateInput` | STEP-3, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| `Update` with `UpdateInput` | STEP-4, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| `KindEvidenceGraph` deletable: `deletableKinds`, a projector, the `evidence_graph_artifacts.graph_id` facet release | STEP-5, AC-4.3, AC-4.4, AC-4.5 |
| `Delete` with the last-graph invariant and the Subjects probe | STEP-5, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5 |
| Audit scope and list-key test for graph create, update and delete | STEP-6, AC-5.1 |
| Docs this PR makes false | STEP-8 |

| Out | Home |
| --- | --- |
| FFI and store methods for list, create, update, delete, link and unlink; the graph list `CatalogQueryKey`; `L10n.Errors` copy for `evidencegraphs.in_use` and `evidencegraphs.last_graph` (beside the three S12-03 codes) | S12-05 (map scope entry updated with this plan) |
| A graph id on `deleteimpact.Location` (the projector opens the Source page until a graph is a place) | S12-08 |
| Hiding Delete on a Source's only graph: `GetDeleteImpact` does not know the last-graph rule, so the Graphs section reads the graph list to decide | S12-11 |
| Naming a graph's linked Artifacts in its delete confirm (a `Named` facet release) | S12-11, only if brief S12-D2 asks for it |
| Editing a graph's links from its form (through `Link` / `Unlink`) | S12-12 |
| A Subject choosing its graph; deleting `SoleGraphTx` | S12-08 |
| Unique graph names per Source | not planned (D-6) |
| A per-graph `Get` read | not planned (working rule "One graph list": `List` is the read) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Graph package | `core/database/evidencegraphs/evidencegraphs.go` `Graph`, `DefaultName`, `ErrInvalid`, `ErrArtifactInUse`, `ErrLastArtifact`, `CreateDefault`, `insert`, `mintRef`, `SoleGraphTx`, `graphIDsTx`, `SubjectGraphTx`, `IsLinkedTx`, `CitationsUsingArtifactTx`, `checkLink`, `linkFields`, `Link`, `Unlink`, `LinkArtifact`, `EnsureRefs`, `mapConstraint` | `Graph` has no Artifact ids. `insert(tx, sourceID, name, description)` mints the ref through `database.InsertWithRef` and returns the `FullRow` create change; it does not trim. `graphIDsTx` (`SELECT id FROM evidence_graphs WHERE source_id = ? ORDER BY id`) is the only per-Source read; `SoleGraphTx` and `LinkArtifact` call it. `LinkArtifact` refuses a repeated id in its own loop. No `List`, `Create`, `Update`, `Delete` |
| Test fixtures | `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` `SeedRefless`, `Ref`; `core/database/evidencegraphs/export_test.go` `SetMintRef` | `evidencegraphstest` imports `sources` (and through it `deleteimpact`), not `artifacts`, `subjects` or `connect`, so those packages' internal tests can import it |
| Helper copies (MNT-8) | `core/database/artifacts/artifacts_test.go` `defaultGraphID` (3 calls), `core/database/connect/connect_test.go` `defaultGraphID` (1 call), `core/database/subjects/subjects_test.go` `defaultGraphID` (1 call), `core/database/evidencegraphs/evidencegraphs_test.go` `defaultGraphID` (3 calls, over its `graphRows`) | Four copies of one SQL read of a Source's graph id; the first three are internal test packages |
| Kind registry | `core/database/catalogmodel/model.go` `KindEvidenceGraph` (comment "registered for its inbound probe; delete path in S12-04"); `Tables` row `evidence_graphs` (`PK: "id"`, `BucketFacet`) | Registered, not deletable |
| Delete registries | `core/database/deleteimpact/register.go` `deletableKinds`, `inboundEdgeList` (has `fkEdge(KindEvidenceGraph, "subjects.graph_id", …)`); `facets.go` `buildFacetReleases` (link release under `KindArtifact` via `artifact_id`; graph release under `KindSource` via `source_id`); `projectors.go` `projectKind` | No graph release under `KindEvidenceGraph`; no graph projector (the default case returns no `Section`) |
| Honesty tests | `core/database/deleteimpact/pragma_test.go` `TestDeletableKindsResolve`, `TestPragmaHonestyProjectors`, `TestFacetReleaseHonesty` | Making `KindEvidenceGraph` deletable makes the projector test require a `Section` and the facet test require a release for `evidence_graph_artifacts.graph_id` |
| Delete precedent | `core/database/sources/sources.go` `Delete` | Check input and user → load (missing → `ErrInvalid`) → `Impact` → `Refuse(Codes{InUse, NotFound})` → `ReleaseFacets` → `DELETE` → released changes, then its own delete change |
| Update precedent | `core/database/sources/sources.go` `Update` | Trim; empty title → `ErrInvalid`; load (missing → `ErrInvalid`); one `FieldDiff` per changed column, `description` through `database.NullString`; no diff → `nil, nil`; one `ActionUpdate` change |
| Effects | `core/database/effects/registry.go` entries `evidence_graphs` (`SourceScope: field("source_id")`, `Lists: ListEvidenceGraphs`), `evidence_graph_artifacts`; `path.go` `readChangeColumn` | An update change without `source_id` resolves the scope from the live row; a delete change resolves it from its own `source_id` field. No registry change needed |
| Errors | `core/apperr/apperr.go` `CodeEvidenceGraphsInvalid`, `CodeEvidenceGraphsArtifactInUse`, `CodeEvidenceGraphsLastArtifact`; `CodeSourcesInUse` | Every `*.in_use` code is `apperr.KindConflict` |
| Read-in-a-write rule | [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md) § Reads inside a write; `core/database/tx.go` `Querier`; `core/database/artifacts/artifacts.go` `ListBySource` / `ListBySourceTx` | A Catalog read and its `…Tx` twin share one unexported function over `database.Querier`. Artifacts list by `ref COLLATE NOCASE` |
| Null binders | `core/database/null.go` `NullString`, `NullBlob` | `NullBlob(nil)` binds SQL NULL |
| Tests | `core/database/evidencegraphs/evidencegraphs_test.go` `TestCreateDefault`, `TestSoleGraphTx`, `TestLink`, `TestUnlink`, `TestMembership`, `TestEnsureRefs` (helpers `newFixture`, `fixture.createSource`, `graphRows`, `run`, `fixture.newArtifact`, `fixture.link`, `fixture.subjectAndCitation`); `core/database/audit/scopes_test.go` `TestSourceScopes` (case "graph link and unlink"); `core/database/deleteimpact/impact_test.go` `TestImpactSourceAndArtifact` (package `deleteimpact_test`) | Table-driven; each builds a temp catalog with `database.Create` |

**Precedent.** `sources.Delete` and `sources.Update` (`core/database/sources/sources.go`) for the write shapes; `artifacts.ListBySource` / `ListBySourceTx` for the Catalog/Tx read pair; the `KindSource` → `evidence_graphs` `rowFacet` in `facets.go` for the graph delete change's field set; `projectSource` in `projectors.go` for the projector. Copy their shapes; do not copy `sources.Delete`'s per-field `FieldDiff` map for the delete change — use `rowchange.DeletedRow` as `Unlink` and `rowFacet` do.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `Graph` gains `ArtifactIDs [][]byte` (comment: "linked Artifacts, by Artifact ref; filled by List and Create"). `List(c *database.Catalog, sourceID []byte) ([]Graph, error)` and `ListTx(tx *sql.Tx, sourceID []byte) ([]Graph, error)` both call an unexported `list(q database.Querier, sourceID []byte) ([]Graph, error)`. `sourceID` nil means every graph in the project; otherwise it must be 16 bytes (`ErrInvalid`). An unknown Source returns an empty slice. `List` runs `list` inside a read transaction it rolls back (`db.Begin`, as `deleteimpact.Snapshot` does) so its two queries see one state; `ListTx` returns `ErrInvalid` for a nil `tx` | Working rule "One graph list": one read, full rows with linked Artifact ids, sorted by name, optional Source filter. The [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md) Catalog/Tx twin rule. One read transaction because the catalog's single connection can serve a write between two plain queries | A `ListBySource` beside a project-wide `List` — the map review (K-12) chose one function; a per-graph `Get` — the working rule says `List` is the read |
| D-2 | `list` runs two queries and attaches links by graph id: `sqlList = SELECT id, COALESCE(ref, ''), source_id, name, COALESCE(description, '') FROM evidence_graphs WHERE ? IS NULL OR source_id = ? ORDER BY name COLLATE NOCASE, ref, id` and `sqlListLinks = SELECT l.graph_id, l.artifact_id FROM evidence_graph_artifacts l JOIN evidence_graphs g ON g.id = l.graph_id JOIN artifacts a ON a.id = l.artifact_id WHERE ? IS NULL OR g.source_id = ? ORDER BY a.ref COLLATE NOCASE, a.id`, both bound with `database.NullBlob(sourceID)` twice. One unexported `scanGraph` scans a graph row for `list` and `getTx` (D-8) | Working rule "graphs list alphabetically by name"; `ref`, then `id` make ties stable. Artifact ids in the order `artifacts.ListBySource` shows Artifacts. Two flat queries avoid encoding blob ids in `group_concat` | `group_concat(hex(artifact_id))` — a string to split and decode; one query per graph — N+1 |
| D-3 | `graphIDsTx` is deleted. `SoleGraphTx` and `LinkArtifact`'s every-graph branch call `list(tx, sourceID)` and take the ids (in `list`'s order). `SoleGraphTx` keeps its 0 / 1 / many behavior and its doc comment | Map scope entry: `List` absorbs `LinkArtifact`'s every-graph read; one per-Source read in the package | Keeping `graphIDsTx` beside `list` — two reads of the same rows |
| D-4 | `evidencegraphstest.DefaultGraphID(t *testing.T, c *database.Catalog, sourceID []byte) []byte` *(new)* calls `evidencegraphs.List(c, sourceID)`, fails the test unless exactly one graph comes back, and returns its `ID`. The four `defaultGraphID` copies are deleted and every call site uses it (`artifacts_test.go` ×3, `connect_test.go` ×1, `subjects_test.go` ×1, `evidencegraphs_test.go` ×3). `evidencegraphs_test.go` keeps `graphRows` (its tests read `ref` and `name` of graphs on purpose) | S12-03 review MNT-8: the copies move onto `List`. `evidencegraphstest` already exists for shared graph fixtures and imports none of `artifacts`, `subjects`, `connect` (checked with `go list -deps`), so their internal tests can import it | A helper per file over `List` — still four copies; inlining `List` with error checks at eight call sites — more lines than the helper |
| D-5 | `CreateInput{SourceID []byte; Name, Description string; ArtifactIDs [][]byte}` *(new)*. `Create(tx *database.Tx, userID []byte, in CreateInput) (Graph, []rowchange.Change, error)`, in this order: trim `Name` and `Description`; `tx` nil, `SourceID` not 16 bytes, empty `Name` or empty `ArtifactIDs` → `ErrInvalid`; `database.RequireUserID(userID, ErrInvalid)`; a repeated Artifact id → `ErrInvalid`; `SELECT 1 FROM sources WHERE id = ?` with no row → `ErrInvalid`; `insert(tx.Tx, in.SourceID, in.Name, in.Description)`; then `Link(tx, userID, graphID, id)` for each id in the given order, returning its error unchanged. Changes `[graph create, link creates in input order]`; the returned `Graph.ArtifactIDs` is a copy of `in.ArtifactIDs`. The repeated-id loop moves out of `LinkArtifact` into an unexported `requireDistinct(ids [][]byte) error`, which both call | Working rule "Creating a graph needs at least one Artifact" and "`Link` / `Unlink` of (graph, Artifact) pairs that apply the link rules": `Link` already refuses an unknown Artifact, a 15-byte id and another Source's Artifact, and `writes.Run` rolls the graph insert back. One insert (`insert`) and one dedupe | Checking each Artifact's Source in `Create` — a copy of `checkLink`; a separate "no Artifacts" code — the UI never offers Create without one (working rule), so it is bad input |
| D-6 | Graph names are not unique; two graphs of one Source may both be named `Default` or `Smith`. Names are trimmed and must not be empty; there is no length cap | The design and working rules require only a name; refs tell graphs apart and Source titles have no uniqueness rule either. A uniqueness rule is a product call this plan does not add | A `UNIQUE (source_id, name)` constraint — a migration and a refusal no rule asks for |
| D-7 | `UpdateInput{Name, Description string}` *(new)*. `Update(tx *database.Tx, userID, graphID []byte, in UpdateInput) ([]rowchange.Change, error)`: trim both; `tx` nil, `graphID` not 16 bytes or empty `Name` → `ErrInvalid`; `RequireUserID`; `getTx` (no row → `ErrInvalid`); a `FieldDiff` for `name` when it changed and for `description` when it changed (`database.NullString` on both sides, as `sources.Update`); no diff → `nil, nil`; else `UPDATE evidence_graphs SET name = ?, description = ? WHERE id = ?` and one change `{EntityType: "evidence_graph", EntityID: graphID, Action: rowchange.ActionUpdate, Fields: <the diffs>}` | Working rule "Name and description are editable". Same shape as `sources.Update`; `effects` resolves the Source scope of an update change from the live row (`readChangeColumn`) | Taking a whole `Graph` as `sources.Update` takes a `Source` — `ArtifactIDs` would look editable here; links change through `Link` / `Unlink` |
| D-8 | `getTx(tx *sql.Tx, id []byte) (Graph, error)` is unexported: `SELECT id, COALESCE(ref, ''), source_id, name, COALESCE(description, '') FROM evidence_graphs WHERE id = ?`, `sql.ErrNoRows` passed through. Its only callers are `Update` and `Delete` | A write must load the row it changes; `List` stays the only exported read (working rule "One graph list") | An exported `Get` — a second read for S12-05 to wrap |
| D-9 | Error codes *(new)* in `core/apperr/apperr.go` after `CodeEvidenceGraphsLastArtifact`: `CodeEvidenceGraphsInUse = "evidencegraphs.in_use"` and `CodeEvidenceGraphsLastGraph = "evidencegraphs.last_graph"`. In `evidencegraphs`: `ErrInUse = apperr.New(apperr.CodeEvidenceGraphsInUse, apperr.KindConflict)` (comment "returned by Delete while Subjects belong to the graph") and `ErrLastGraph = apperr.New(apperr.CodeEvidenceGraphsLastGraph, apperr.KindUser)` (comment "a Source keeps at least one graph"). No `L10n.Errors` entry in this PR | Working rule "Refusals are of two kinds": Subjects block through `deleteimpact`, like other in-use deletes, so the code is `*.in_use` with the conflict kind every `*.in_use` has; the last graph is a domain invariant with its own user-kind code | One user-kind code for both — the probe refusal would differ from every other in-use delete |
| D-10 | `Delete(tx *database.Tx, userID, graphID []byte) ([]rowchange.Change, error)`, in this order: `tx` nil or `graphID` not 16 bytes → `ErrInvalid`; `RequireUserID`; `getTx` (no row → `ErrInvalid`); `list(tx.Tx, g.SourceID)` has one graph → `ErrLastGraph`; `deleteimpact.Impact(tx.Tx, catalogmodel.KindEvidenceGraph, graphID)` then `deleteimpact.Refuse(report, deleteimpact.Codes{InUse: ErrInUse, NotFound: ErrInvalid})`; `deleteimpact.ReleaseFacets(tx, catalogmodel.KindEvidenceGraph, graphID)`; `DELETE FROM evidence_graphs WHERE id = ?`; return the released changes, then `{EntityType: "evidence_graph", EntityID: graphID, Action: rowchange.ActionDelete, Fields: rowchange.DeletedRow(map[string]any{"id": database.UUIDJSON(id), "ref": database.NullString(g.Ref), "source_id": database.UUIDJSON(g.SourceID), "name": g.Name, "description": database.NullString(g.Description)})}` | [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md) § Domain `Delete`: extra gates before `Impact`. The last-graph refusal comes first because deleting the Subjects would not make that delete possible. The field set matches the graph's create change and the `KindSource` graph release, so a graph's history reads alike | Last graph after `Impact` — a researcher would be told to delete Subjects and then refused again; `Impact` knowing the last-graph rule — the working rule keeps it a domain invariant |
| D-11 | `evidencegraphs` imports `core/database/catalogmodel` and `core/database/deleteimpact` (neither imports `evidencegraphs`; checked with `go list -deps ./core/database/deleteimpact`) | The delete pattern calls `Impact` / `Refuse` / `ReleaseFacets` in the domain package | A composing package for graph delete — nothing else composes |
| D-12 | `deleteimpact`: add `catalogmodel.KindEvidenceGraph: true` to `deletableKinds`. Add `rowFacet(catalogmodel.KindEvidenceGraph, "evidence_graph_artifacts", "graph_id", "evidence_graph_artifact", "graph_id", col{"graph_id", colUUID}, col{"artifact_id", colUUID})` to `buildFacetReleases` directly after the `KindArtifact` link release, with the comment "A graph's links go with it, audited under its id." Not `Named`. Add `case catalogmodel.KindEvidenceGraph: return projectEvidenceGraph(tx, row.ID, row.Ref)` to `projectKind`, and `projectEvidenceGraph` *(new)* shaped like `projectSource`: `SELECT name, source_id FROM evidence_graphs WHERE id = ?`; no row → `(ref, Location{Section: sectionSources, SourceSurface: surfacePage, Ref: ref}, nil)`; else title = trimmed name or `ref`, `Location{Section: sectionSources, SourceID: database.UUIDString(sourceID), SourceSurface: surfacePage, Ref: ref, Title: title}`. Change the `KindEvidenceGraph` comment in `catalogmodel/model.go` to "KindEvidenceGraph is one Evidence graph of a Source." The `KindSource` graph release stays a plain `rowFacet`: a Source is deletable only with no Artifacts, so its graphs have no links | The honesty tests require a release for each audited `CASCADE` whose parent is deletable and a projector with a `Section` for each deletable kind. Same field set as the `KindArtifact` link release. Until S12-08 a graph has no place of its own, so the projector opens its Source page | A `Named` link release — the confirm has no brief yet (Out, S12-11); a custom recursive release for the Source's graphs — unreachable rows |
| D-13 | No FFI, proto, Swift, migration or `VERSION` change | Out lines S12-05 and later | — |
| D-14 | Tests build fixtures with `sources.Create`, `artifacts.Create` and `subjects.Create` inside writes, and new graphs with `evidencegraphs.Create`. A Subject on a graph that is not the Source's only one is made by creating the Subject first, while the Source has one graph, then creating the second graph (`SoleGraphTx` refuses a two-graph Source) | `subjects.Create` takes its graph from `SoleGraphTx` until S12-08 | A direct `INSERT INTO subjects` — S12-03 moved fixtures off direct inserts |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `evidencegraphs.List`, `ListTx`, `list` *(new)* | func | `evidencegraphs.graphIDsTx` (deleted here), `artifacts.ListBySource` / `ListBySourceTx` | `graphIDsTx` returns ids only; `List` replaces it (D-3). Same Catalog/Tx twin shape as `artifacts` |
| `Graph.ArtifactIDs` *(new)* | field | `Graph` (S12-03) | The working rule's "full rows with linked Artifact ids"; the S12-03 plan named this extension |
| `evidencegraphs.Create`, `CreateInput` *(new)* | func / type | `evidencegraphs.CreateDefault`, `insert`, `Link` | `Create` is `insert` plus `Link` per Artifact; no SQL of its own except the Source check |
| `evidencegraphs.Update`, `UpdateInput` *(new)* | func / type | `sources.Update` | Same shape on another table |
| `evidencegraphs.Delete` *(new)* | func | `sources.Delete` | Same pattern, plus the last-graph gate |
| `evidencegraphs.getTx`, `scanGraph` *(new)* | func | `sources.getTx`, `scanSource` | Unexported row load for the writes; `scanGraph` is shared with `list` |
| `evidencegraphs.requireDistinct` *(new)* | func | the repeated-id loop in `LinkArtifact` | Moved out of `LinkArtifact` so `Create` shares it |
| `ErrInUse`, `ErrLastGraph`; `CodeEvidenceGraphsInUse`, `CodeEvidenceGraphsLastGraph` *(new)* | var / const | `sources.ErrInUse`, `evidencegraphs.ErrLastArtifact` | One code per refusal (working rule) |
| `projectEvidenceGraph` *(new)* | func | `projectSource` | One projector per deletable kind; `TestPragmaHonestyProjectors` requires it |
| `rowFacet(KindEvidenceGraph, "evidence_graph_artifacts", "graph_id", …)` *(new)* | registry entry | the `KindArtifact` link release | Same rows, other parent |
| `evidencegraphstest.DefaultGraphID` *(new)* | test helper | the four `defaultGraphID` copies (deleted) | One helper over `List` (MNT-8) |
| `TestList`, `TestCreate`, `TestUpdate`, `TestDelete` *(new)* | test | `TestLink`, `TestUnlink` in the same file | Same fixtures |

**Builds on.** The `deleteimpact` registries (`deletableKinds`, `rowFacet`, `projectKind`) and the `add-catalog-delete` domain `Delete` pattern; `evidencegraphs.insert` and `Link`; the `effects` entries S12-03 added (no change); `evidencegraphstest`.
**Deletes.** `evidencegraphs.graphIDsTx`; the repeated-id loop inside `LinkArtifact` (moved to `requireDistinct`); the four `defaultGraphID` test helpers; the "delete path in S12-04" comment; the sentences in `docs/catalog-deletes.md` that say graphs are not deletable yet.
**Next adds.** S12-05 wraps `List`, `Create`, `Update`, `Delete`, `Link` and `Unlink` in RPCs, maps the five `evidencegraphs.*` codes in `L10n.Errors`, and reads `GetDeleteImpact` for kind `evidence_graph` with no Go change. S12-08 adds a graph id to `deleteimpact.Location` in `projectEvidenceGraph`. S12-11's Delete button reads `List` to hide itself on a Source's only graph.

### Schema

None — no catalog change; the tables, FKs and `catalogmodel` rows are S12-03's.

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | two codes (D-9) |
| `core/database/evidencegraphs` | below |
| `core/database/evidencegraphs/evidencegraphstest` | `DefaultGraphID` *(new)* (D-4) |
| `core/database/catalogmodel` | the `KindEvidenceGraph` comment (D-12) |
| `core/database/deleteimpact` | `deletableKinds`, one `rowFacet`, `projectEvidenceGraph` (D-12) |

`core/database/evidencegraphs/evidencegraphs.go`, additions:

```go
var (
	// ErrInUse is returned by Delete while Subjects belong to the graph.
	ErrInUse = apperr.New(apperr.CodeEvidenceGraphsInUse, apperr.KindConflict)
	// ErrLastGraph: a Source keeps at least one graph.
	ErrLastGraph = apperr.New(apperr.CodeEvidenceGraphsLastGraph, apperr.KindUser)
)

type Graph struct {
	ID          []byte
	Ref         string
	SourceID    []byte
	Name        string
	Description string
	ArtifactIDs [][]byte // linked Artifacts, by Artifact ref; filled by List and Create
}

type CreateInput struct {
	SourceID    []byte
	Name        string
	Description string
	ArtifactIDs [][]byte // at least one, each an Artifact of SourceID
}

type UpdateInput struct {
	Name        string
	Description string
}

func List(c *database.Catalog, sourceID []byte) ([]Graph, error) // nil sourceID = every graph
func ListTx(tx *sql.Tx, sourceID []byte) ([]Graph, error)
func Create(tx *database.Tx, userID []byte, in CreateInput) (Graph, []rowchange.Change, error)
func Update(tx *database.Tx, userID, graphID []byte, in UpdateInput) ([]rowchange.Change, error)
func Delete(tx *database.Tx, userID, graphID []byte) ([]rowchange.Change, error)
```

SQL lives as unexported consts in the same file (`sqlList`, `sqlListLinks`, `sqlGet`, `sqlSourceExists`, `sqlUpdate`, `sqlDelete`); `sqlGraphIDs` goes with `graphIDsTx`.

Errors: `evidencegraphs.invalid` → bad input, unknown Source / graph / Artifact, an Artifact of another Source, a repeated id, an empty name, no Artifacts on create. `evidencegraphs.in_use` → `Delete` of a graph with Subjects. `evidencegraphs.last_graph` → `Delete` of a Source's only graph. None carries params.

### FFI

None — S12-05.

### macOS

None — no Swift change.

### Strings

None — `evidencegraphs.in_use` and `evidencegraphs.last_graph` get `L10n.Errors` entries in S12-05, with the three S12-03 codes.

## 7. Steps

In every step, a test that fails **only** because it called a deleted `defaultGraphID` copy is moved onto `evidencegraphstest.DefaultGraphID` in that step. Any other failure in a test this plan does not name is a Stop condition.

**STEP-1 — Error codes**
Files: `core/apperr/apperr.go`, `core/database/evidencegraphs/evidencegraphs.go`.
Do: D-9 (codes and the two sentinels; nothing returns them yet).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/apperr/ ./core/database/evidencegraphs/` passes. Commit.

**STEP-2 — One graph list, and the test helper on it**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go`, `core/database/artifacts/artifacts_test.go`, `core/database/connect/connect_test.go`, `core/database/subjects/subjects_test.go`. Skill: `add-catalog-query`.
Do: D-1, D-2, D-3 (`Graph.ArtifactIDs`, `List`, `ListTx`, `list`, `scanGraph`; delete `graphIDsTx`). D-4 (`DefaultGraphID`; delete the four copies, move their call sites). Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes; `grep -rn "func defaultGraphID\|graphIDsTx" core` prints nothing. Commit: "List Evidence graphs with their linked Artifacts in one read."

**STEP-3 — Create a graph**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`.
Do: D-5, D-6 (`CreateInput`, `Create`, `requireDistinct`; `LinkArtifact` calls `requireDistinct`). Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/ ./core/database/artifacts/` passes (`TestLink`'s repeated-id case still refuses). Commit: "Create an Evidence graph linked to chosen Artifacts."

**STEP-4 — Rename a graph and edit its description**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`.
Do: D-7, D-8 (`UpdateInput`, `Update`, `getTx`). Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/` passes. Commit: "Rename an Evidence graph and edit its description."

**STEP-5 — Delete a graph under its rules**
Files: `core/database/catalogmodel/model.go`, `core/database/deleteimpact/register.go`, `core/database/deleteimpact/facets.go`, `core/database/deleteimpact/projectors.go`, `core/database/deleteimpact/impact_test.go`, `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`. Skill: `add-catalog-delete`.
Do: D-12 first, then D-10, D-11. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/ ./core/database/catalogmodel/ ./core/database/evidencegraphs/ ./core/database/sources/ ./core/database/artifacts/` passes, including `TestDeletableKindsResolve`, `TestPragmaHonestyProjectors`, `TestFacetReleaseHonesty`. Commit: "Delete an empty Evidence graph, and refuse a graph with Subjects or a Source's last graph."

**STEP-6 — Graph writes are scoped to their Source**
Files: `core/database/audit/scopes_test.go`. Skill: `add-audit-scope`.
Do: Add T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/` passes. Commit.

**STEP-7 — Whole suite**
Files: none, unless the run finds a call site the earlier steps missed (fixed under the rule above).
Do: run the whole Go suite, `go vet` and `scripts/check-pr.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` and `go vet -tags fts5 ./api/... ./core/...` pass; `scripts/check-pr.sh` prints `check-pr: PASS` on macOS.

**STEP-8 — Docs**
Files:
- `docs/source-layer-data-model.md` § 7.1 "Rules the schema and `core/database/evidencegraphs` hold today": add that `List` is the one read of graphs (full rows with linked Artifact ids, by name, one Source or the project); `Create` needs a name and at least one Artifact of the Source, linked through `Link`; name and description are editable and names need not be unique; `Delete` refuses a Source's last graph (`evidencegraphs.last_graph`) and a graph with Subjects (`evidencegraphs.in_use`, through `deleteimpact`), and releases the graph's links audited.
- `docs/catalog-deletes.md` § Policy item 10: graph delete releases its links (`evidence_graph_artifacts.graph_id`) through `ReleaseFacets`; `KindEvidenceGraph` is deletable; the last-graph rule is a domain gate in `evidencegraphs.Delete` that `GetDeleteImpact` does not report. Remove "the `graph_id` release waits for a graph delete path" and "`KindEvidenceGraph` is not deletable yet".
- `docs/deployment-plan/spike-12/deployment-plan.md`: map row status (the S12-05 scope edit was made with this plan).
- `docs/deployment-plan/spike-12/completed.md`; this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-04-graph-lifecycle.md` reports no errors; links resolve. Commit.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits; every commit passes its step's check.

## 8. Tests

All Go, table-driven (`.cursor/rules/go-tests.mdc`). New `evidencegraphs` tests live in `core/database/evidencegraphs/evidencegraphs_test.go` (package `evidencegraphs_test`) and use its fixtures (`newFixture`, `fixture.createSource`, `run`, `fixture.newArtifact`, `fixture.link`) per D-14. A Subject is created with `subjects.Create` the way `fixture.subjectAndCitation` does.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `evidencegraphs_test.go` `TestList` *(new)* | Fixture built with SQL through `c.DB()`, as `TestSoleGraphTx`'s "two graphs" case does (`Create` does not exist yet in STEP-2): S from `fixture.createSource` (its `Default`), plus `INSERT INTO evidence_graphs (id, source_id, name)` rows `Jones` and `smith` with `uuid.NewV7()` ids; Artifacts `B` then `A` from `fixture.newArtifact` (refs minted in that order, so compare by ref, not label), linked to `Jones` with `fixture.link`; T from `fixture.createSource` with its `Default` renamed `Archive` by SQL → the orders and `ArtifactIDs` in AC-1.1, AC-1.2, AC-1.3; unknown id → empty; 15-byte id → `ErrInvalid`; closed catalog → `database.ErrClosed`; inside one `run`, a graph inserted by SQL on `tx.Tx` in that write is in `ListTx(tx.Tx, S.id)` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, D-1, D-2 |
| T-2 | Go | `evidencegraphs_test.go` `TestCreate` *(new)* | AC-2.1 happy case: the row, the `GPH-` ref (`ref.Valid`), changes in order, `Graph.ArtifactIDs`; each AC-2.2, AC-2.3, AC-2.4 input → `ErrInvalid` and `List(c, S.id)` unchanged after the write returns; AC-2.5 → succeeds | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, D-5, D-6 |
| T-3 | Go | `evidencegraphs_test.go` `TestUpdate` *(new)* | AC-3.1, AC-3.2, AC-3.3, AC-3.4: the `Fields` keys and old/new values of each change; unchanged → `nil` changes; each bad input → `ErrInvalid` and `List` shows the old row | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-7 |
| T-4 | Go | `evidencegraphs_test.go` `TestDelete` *(new)* | AC-4.1: changes `[evidence_graph_artifact delete, evidence_graph delete]`, the delete's `Fields` keys, rows gone, `Default` and A1 kept. AC-4.2: only graph without Subjects → `ErrLastGraph`; only graph with a Subject → `ErrLastGraph` (not `ErrInUse`). AC-4.3: Subject on `Default`, then `Create` `Jones`, `Delete(Default)` → `ErrInUse`, both graphs kept. AC-4.5 bad inputs → `ErrInvalid` | AC-4.1, AC-4.2, AC-4.3, AC-4.5, D-10 |
| T-5 | Go | `core/database/deleteimpact/impact_test.go` `TestImpactEvidenceGraph` *(new)* | Source with a Subject on `Default`, then a second graph `Jones` (`evidencegraphs.Create`): `Impact(KindEvidenceGraph, Default)` → gate `inbound`, one group via `subjects.graph_id` with `Total` 1 and the Subject listed with a non-empty `Location.Section`; `Impact(…, Jones)` → `Allowed`, no groups; unknown id → gate `not_found` | AC-4.3, AC-4.4, AC-4.5, D-12 |
| T-6 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` *(new case "graph create, rename and delete")* | On Source A: `evidencegraphs.Create` (linked to A's Artifact), `Update` its name, `Delete` it — each bumps A only; `newScopeSeed` unchanged. Follow the "graph link and unlink" case's shape | AC-5.1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFacetReleaseHonesty` and `TestPragmaHonestyProjectors` (STEP-5), `TestLink` (STEP-3's `requireDistinct` move), and the three internal test packages whose helper moves (STEP-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name for each (AC-5.1's list key is evidenced by T-6 together with the unchanged effects entries; the S12-03 test TestEvidenceGraphsResolve still passes)
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] grep -rn "func defaultGraphID\|graphIDsTx\|delete path in S12-04" core prints nothing
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: source-layer-data-model.md § 7.1, catalog-deletes.md § Policy item 10
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing a researcher sees changes, and no RPC reaches these functions until S12-05. The reviewer runs, on the branch:

1. `CGO_ENABLED=1 go test -tags fts5 -run 'TestList|TestCreate|TestUpdate|TestDelete' -v ./core/database/evidencegraphs/` → every case passes, and the names match §8.
2. `CGO_ENABLED=1 go test -tags fts5 -run 'TestImpactEvidenceGraph|TestFacetReleaseHonesty|TestPragmaHonestyProjectors|TestDeletableKindsResolve' -v ./core/database/deleteimpact/` → pass.
3. Open an existing dogfood project in the app built from this branch: the Sources list, a Source page and its canvas look as on `main` (no schema or read path a screen uses has changed).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test made a second graph and then a Subject, which `SoleGraphTx` refuses | D-14 fixes the fixture order; the refusal is the intended bridge behavior until S12-08 |
| `TestFacetReleaseHonesty` or the projector test fails once the kind is deletable | STEP-5 adds the release and projector before the `Delete` function; D-12 gives both exactly |
| `List` order differs between Go and the UI later | D-2 fixes the SQL order; S12-05 does not re-sort |
| The UI offers Delete on a Source's only graph because `GetDeleteImpact` says allowed | Out line S12-11; `Delete` still refuses with `last_graph` |

**Stop and report if:**

- `core/database/evidencegraphs/evidencegraphs.go` has changed on `main` since `3bd0862` so a §4 row is no longer true (a Spike 11 PR or S12-06 landed first and touched it).
- An import cycle appears when `evidencegraphs` imports `deleteimpact` or `catalogmodel` (D-11 would be wrong), or when an internal test package imports `evidencegraphstest` (D-4 would be wrong).
- `TestDeletableKindsResolve`, `TestPragmaHonestyProjectors` or `TestFacetReleaseHonesty` cannot pass with exactly the registrations in D-12.
- `deleteimpact.ReleaseFacets(tx, KindEvidenceGraph, id)` fails with `release_incomplete`.
- `writes.Run` rejects an `evidence_graph` update change (`audit.Record` or `effects.Resolve` error).
- A test this plan does not name fails for a reason other than the moved `defaultGraphID` helper.

## 12. Open questions

None.

## 13. PR

**Title:** S12-04 — Graph lifecycle
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-04-graph-lifecycle.md`.
