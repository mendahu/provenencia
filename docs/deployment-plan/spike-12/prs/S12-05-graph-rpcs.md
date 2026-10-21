# S12-05 — Graph RPCs and graph list

**Status:** Landed in [#511](https://github.com/mendahu/provenencia/pull/511) (2026-10-21)
**Landed:** Two differences from the plan.
- **`TestProvenenciaCall`'s unknown method.** `api/libprovenencia/call_test.go` used method 103 as its unknown method; 103 is now `ListEvidenceGraphs`, so the case uses 9999.
- **Handler tests (D-16).** They call the handlers through a small `rpc` / `mustRPC` helper in `evidence_graphs_test.go` instead of `runRPC` tables, because each case builds several dependent calls; `TestListEvidenceGraphs` keeps the bad-proto case.
**Spike / slice:** Spike 12 · Slice 2 — Graphs in the catalog
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: RPCs, store methods, one query key and error and delete-confirm copy; no view changes)
**Board:** —
**Depends on:** S12-04 (landed [#505](https://github.com/mendahu/provenencia/pull/505))
**Blocks:** S12-08, S12-10, S12-11
**Branch:** `s12-05-graph-rpcs`, branched from the commit that adds this plan (so STEP-9 edits it in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-04 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [rules](../source-structure.md#rules), [`catalog-deletes.md`](../../../catalog-deletes.md) § FFI; working rules: "Refusals are of two kinds", "Artifact ↔ graph links", "One graph list", "The graph list has its own effect", "A re-keyed symbol is renamed in the PR that re-keys it" (naming), "High tolerance for churn"; the S12-04 [Landed note](S12-04-graph-lifecycle.md) and [ship log](S12-04-graph-lifecycle.ship.md) (UX-1); the S12-03 [ship log](S12-03-evidence-graphs-catalog.ship.md) (UX-1, UX-2)
**Planned:** 2026-10-21 against `main` @ `d214e66`
**Review:** [`S12-05-graph-rpcs.review.md`](S12-05-graph-rpcs.review.md) — Fits / Ready with changes, 2 rounds

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

The app can reach every Evidence graph operation the engine has: list every graph in the project, create, edit and delete a graph, link and unlink an Artifact, and create an Artifact linked to chosen graphs. The graph list is one cached read that refreshes itself after any write that touches a graph or a link. Every graph refusal reaches the researcher as its own sentence, a graph delete preview reads correctly, and a graph edit that would unlink an Artifact its Citations use can be told, before it saves, which Artifact blocks and which Citations.

**Why now.** S12-04 finished the engine half (`evidencegraphs` `List`, `Create`, `Update`, `Delete`; the `last_graph` gate). Slice 3 (S12-08, S12-10) and slice 4 (S12-11 – S12-13) build their views on these RPCs and on one graph list key ([map](../deployment-plan.md#s12-05--graph-rpcs-and-graph-list)). Five error codes have no copy yet (S12-03 review UX-2, S12-04 D-9), the `last_graph` gate maps to `UNSPECIFIED` on the wire, and a graph delete with Subjects falls back to "… reference this" (S12-03 review UX-1).

**What stays.** No view changes: nothing on screen reads the graph list or calls the new methods yet. `GetSourceWorkspace` is untouched. Add Artifact keeps linking a new Artifact to every graph of its Source (it sends no `graph_ids`). The engine's link rules, refusals and effects are S12-03 / S12-04's and do not change.

## 2. User stories

**US-1.** As the app, I list every Evidence graph of the project in one call, sorted by name, each with its linked Artifact ids, and I cache it under one key that any write touching a graph or link invalidates.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Project with Source A (graphs `Smith`, `Default`) and Source B (graph `Default`); `ListEvidenceGraphs` | Three rows in `evidencegraphs.List` order: `Default`, `Default`, `Smith` (ties by ref), each with `source_id`, `ref` (`GPH-…`), `name`, `description` and `artifact_ids` |
| AC-1.2 | A write whose `CatalogWriteEffects.lists` holds `.evidenceGraphs` | `CatalogQueryRegistry.invalidations(for:project:)` contains `.key(.evidenceGraphsList(project:))` exactly once, and no `.propertiesWorkspace` key |
| AC-1.3 | `CreateEvidenceGraph`, `UpdateEvidenceGraph`, `DeleteEvidenceGraph`, `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact` and a `CreateArtifact` that links | Each response's `effects.lists` contains `evidence_graphs` and `effects.revision` > 0 |

**US-2.** As the app, I create, edit and delete graphs and link and unlink Artifacts through the engine's rules, and a refusal comes back as its code.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CreateEvidenceGraph` on Source A, name ` Jones `, description `Household 12`, `artifact_ids` = [A's page 1] | Response `graph` has a `GPH-` ref, name `Jones`, description `Household 12`, `artifact_ids` = [page 1]; `ListEvidenceGraphs` shows it |
| AC-2.2 | `CreateEvidenceGraph` with no `artifact_ids`, or with Source B's Artifact | Error code `evidencegraphs.invalid`; no graph added |
| AC-2.3 | `UpdateEvidenceGraph` on `Jones`: name `Jones family`, `artifact_ids` = [page 2] | Graph renamed, linked to page 2 only; the response's `effects.revision` is one revision above the create's |
| AC-2.4 | Default graph linked to pages 1 and 2; a Citation on page 1 has an Observation on a Default Subject; `UpdateEvidenceGraph` on Default with name `Household` and `artifact_ids` = [page 2] | Error code `evidencegraphs.artifact_in_use`; name and links unchanged |
| AC-2.5 | `DeleteEvidenceGraph` on a Source's only graph | Error code `evidencegraphs.last_graph` |
| AC-2.6 | `DeleteEvidenceGraph` on a non-last graph with a Subject | Error code `evidencegraphs.in_use` |
| AC-2.7 | `DeleteEvidenceGraph` on an empty, non-last graph | Graph gone from `ListEvidenceGraphs` |
| AC-2.8 | `LinkEvidenceGraphArtifact` (Jones, page 2) then `UnlinkEvidenceGraphArtifact` (Jones, page 2) on a graph linked to pages 1 and 2 | Link added, then removed; each response has effects |
| AC-2.9 | `UnlinkEvidenceGraphArtifact` on a graph's only link | Error code `evidencegraphs.last_artifact` |

**US-3.** As the app, I create an Artifact linked to the graphs I choose.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Source A with graphs `Default` and `Jones`; `CreateArtifact` with `graph_ids` = [Jones] | The new Artifact is linked to `Jones` only |
| AC-3.2 | Same Source; `CreateArtifact` with no `graph_ids` | Linked to `Default` and `Jones` (today's behavior) |
| AC-3.3 | `CreateArtifact` with `graph_ids` = [Source B's graph] | Error code `evidencegraphs.invalid`; no Artifact created |

**US-4.** As the app, before a graph edit saves, I learn which of the Artifacts it removes are used by that graph's Citations, with those Citations shaped for the DeleteImpact recipe.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Default graph linked to pages 1 and 2; Citation `CIT-…` on page 1 has an Observation on a Default Subject; `GetEvidenceGraphUnlinkImpact(Default, [page 2, page 1])` | One block: `artifact_id` page 1, `citations` group `via` = `evidencegraphs.artifact_in_use`, `kind` = `citation`, `total` = 1, one listed row with the Citation's id, ref, title and a `citationComposer` location |
| AC-4.2 | Same graph; `[page 2]` (no Citations) | No blocks |
| AC-4.3 | Same Source after `CreateEvidenceGraph` `Jones` linked to pages 1 and 2; `GetEvidenceGraphUnlinkImpact(Jones, [page 1])` | No blocks: the Citation's Observations are on a Default Subject, not a Jones one |
| AC-4.4 | `deleteimpact.ListedGroup` with 25 Citation rows | `total` = 25, 20 listed (the recipe's cap), in input order |
| AC-4.5 | Unknown graph id, or a repeated Artifact id | Error code `evidencegraphs.invalid` |

**US-5.** As a researcher, every graph refusal and graph delete preview reads as a sentence about my data.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `L10n.Errors.message(code:)` for each of `evidencegraphs.invalid`, `evidencegraphs.artifact_in_use`, `evidencegraphs.last_artifact`, `evidencegraphs.in_use`, `evidencegraphs.last_graph` | The §6 Strings copy; none equals `L10n.Errors.message(code: "internal.unknown")` |
| AC-5.2 | `GetDeleteImpact` (kind `evidence_graph`) on a Source's only graph | `allowed` false, `gate` = `DELETE_IMPACT_GATE_LAST_GRAPH`; Swift maps it to `.lastGraph` and `PVDeleteImpactCopy.extraGate` returns the §6 title and body; no destructive button |
| AC-5.3 | `PVDeleteImpactCopy.heading(via: "subjects.graph_id", kind: "subject", total:)` with 1 and 3 | `This graph has 1 subject`, `This graph has 3 subjects` |
| AC-5.4 | `PVDeleteImpactCopy.heading(via: "evidencegraphs.artifact_in_use", kind: "citation", total:)` with 1 and 2 | `1 citation in this graph uses this artifact`, `2 citations in this graph use this artifact` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `deleteimpact.ListedGroup` and `evidencegraphs.UnlinkImpact` (the `artifact_in_use` preflight read over a set of removed Artifacts; S12-04 review UX-1) | STEP-1, AC-4.1 – AC-4.5 |
| Proto: seven methods, their messages, `EvidenceGraph`, `CreateArtifactRequest.graph_ids`, `DELETE_IMPACT_GATE_LAST_GRAPH` | STEP-2 |
| `parseIDs` in handlers; `CopySource`, `ReorderSourceMetadata`, `ApplyPromoteGraphAlignment` move onto it | STEP-3 |
| `CreateArtifact` passes `graph_ids`; `gateToProto` maps `last_graph` | STEP-3, AC-3.1 – AC-3.3, AC-5.2 |
| Handlers `ListEvidenceGraphs`, `CreateEvidenceGraph`, `UpdateEvidenceGraph`, `DeleteEvidenceGraph`, `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact`, `GetEvidenceGraphUnlinkImpact` and their dispatch cases | STEP-4, AC-1.1, AC-1.3, AC-2.1 – AC-2.9 |
| Swift models, `GenealogyStore` / `GoStore` methods, `CoreMethod` constants, `createArtifact(graphIDs:)`, `CatalogDeleteImpactGate.lastGraph` | STEP-5 |
| `FakeStore` graph state and methods | STEP-6, AC-1.3 |
| `CatalogQueryKey.evidenceGraphsList`, its registry spec and loader, the `CatalogList.evidenceGraphs` mapping line, `warmQuery` arm | STEP-7, AC-1.2 |
| `L10n.Errors` copy for the five codes; `last_graph` gate title and body; `viaFormats` headings for `subjects.graph_id` (S12-03 review UX-1) and `evidencegraphs.artifact_in_use` | STEP-8, AC-5.1 – AC-5.4 |
| Docs: map row, `completed.md`, this plan, `catalog-deletes.md` § FFI preflight line | STEP-9 |

| Out | Home |
| --- | --- |
| Any view or model that shows or edits graphs; the Graphs section; reading the graph list on screen | S12-10, S12-11, S12-12, S12-14 |
| Graph checkboxes in Add Artifact and the Artifact card (Add Artifact sends no `graph_ids` here) | S12-13 |
| Warming the graph list from a place (`PlaceRegistry` `queryKeys`) | S12-11 (its first reader) |
| The delete-confirm kind noun for `evidence_graph` (`L10n.DeleteImpact` noun list) and a graph `WorkspaceLocation` in `projectEvidenceGraph` | S12-11, S12-08 |
| A per-graph list RPC, a `Get` for one graph, or graphs in `GetSourceWorkspace` | not planned (working rule "One graph list") |
| A `last_artifact` preflight (the form never sends an empty link set: `UpdateInput.ArtifactIDs` empty means unchanged) | S12-12 disables Save with no Artifact checked |
| `FakeStore.createSource` / `copySource` minting a Default graph | not planned: Swift tests seed `evidenceGraphsByProject` directly (D-14) |
| Re-keying effects, query keys or locations by graph id | S12-08 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Graph reads and writes | `core/database/evidencegraphs/evidencegraphs.go` `List`, `Create`, `Update`, `Delete`, `Link`, `Unlink`, `LinkArtifact` | Engine-complete since S12-04; no FFI caller. `List(c, nil)` returns every graph by name; `Create` returns `Graph` with `Ref` and `ArtifactIDs`; `Update` with empty `ArtifactIDs` leaves links unchanged |
| Citations blocking an unlink | `evidencegraphs.CitationsUsingArtifactTx` | Ids of the Citations on an Artifact with an Observation on a Subject of the graph, ordered by `ref COLLATE NOCASE`; the ref is selected and dropped. Only caller: `Unlink` |
| Error codes | `core/apperr/apperr.go` `CodeEvidenceGraphsInvalid`, `CodeEvidenceGraphsArtifactInUse`, `CodeEvidenceGraphsLastArtifact`, `CodeEvidenceGraphsInUse`, `CodeEvidenceGraphsLastGraph` | Defined; no Swift copy, so the app shows the unknown-error default |
| Artifact create | `core/database/artifacts` `CreateInput.GraphIDs`; `api/ffi/handlers/artifacts.go` `CreateArtifact` | Engine takes `GraphIDs` (empty = every graph); the handler never sets it |
| Delete gates | `core/database/deleteimpact/kinds.go` `GateLastGraph`; `api/ffi/handlers/delete_impact.go` `gateToProto` | `last_graph` exists in Go; `gateToProto` has no case, so it goes on the wire as `DELETE_IMPACT_GATE_UNSPECIFIED` |
| Report projection | `core/database/deleteimpact/projectors.go` `projectListed`, `projectKind`, `projectCitation`; `core/database/deleteimpact/impact.go` `collectGroups`; `listedCap` = 20 in `kinds.go` | Unexported: only `Impact` can shape rows into a `Group` with titles and locations |
| Subjects of a graph edge | `core/database/deleteimpact/register.go` `fkEdge(catalogmodel.KindEvidenceGraph, "subjects.graph_id", …)` | Registered by S12-03; the Swift heading falls back to "%lld subjects reference this" |
| Id lists in handlers | `api/ffi/handlers/sources.go` `CopySource` (artifact ids), `ReorderSourceMetadata` (field ids); `api/ffi/handlers/promote_batch.go` `ApplyPromoteGraphAlignment` (skip bridge ids) | Three copies of the same `for … parseID … append` loop |
| Dispatch | `api/ffi/dispatch.go` `MethodCopySource` (102) and its `case` | The highest method number is 102 |
| Wire effects | `api/ffi/handlers/write_effects.go` `writeEffectsProto`; `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogList.evidenceGraphs` | The `evidence_graphs` list key already reaches Swift (S12-03) |
| List-key mapping | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `static func invalidations(for list: CatalogList, project:)` | `.evidenceGraphs` returns `[]` with the comment "S12-05 adds the graph list's key here." |
| Query keys | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `CatalogQueryKey`, `Kind`; `CatalogQueryRegistry` `specs`, `load`, `invalidation(kind:mutation:)` arms; `WorkspaceSession.swift` `warmQuery` | No graph key. `sourceGraphProgress(project:)` is the project-wide precedent |
| Store | `macos/App/Platform/GenealogyStore.swift` protocol; `GoStore.swift`; `FakeStore.swift`; `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore`; `macos/App/Platform/CoreInvoke.swift` `CoreMethod` | No graph methods. `createArtifact` has no graph parameter; its one app caller is `macos/App/Features/Sources/SourceArtifactsSection.swift` `create()` |
| Delete recipe | `GenealogyStore.swift` `CatalogDeleteImpactGate`; `GoStore.swift` `mapDeleteImpactGate`; `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactCopy.extraGate`, `PVDeleteImpactControls.isExtraGate`; `macos/App/Platform/L10n.swift` `L10n.DeleteImpact.viaFormats`, `gateEdgeLockedTitle` | No `lastGraph` case anywhere in Swift |
| Error copy | `L10n.swift` `L10n.Errors` (`citationsInUse`, `observationsEdgeLocked`) and `message(code:params:)`; `macos/App/Resources/Localizable.xcstrings` | No `evidencegraphs.*` entry |
| Tests | `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactAndIngest`; `api/ffi/handlers/sources_test.go` `TestCopySource`; `api/ffi/handlers/delete_impact_test.go`; `core/database/evidencegraphs/evidencegraphs_test.go`; `core/database/deleteimpact/impact_test.go`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `listInvalidations` (case "evidence graphs alone"); `PVDeleteImpactTests.swift` `extraGatesHaveNoDestructiveButtonAndNoInventedList`, `knownViaHeadingsMatchEngineEdges`; `L10nResolutionTests.swift` `depictionSaveErrorsHaveCopy` | Cover today's behavior; the shapes below extend them |

**Precedent.** S12-01 ([#478](https://github.com/mendahu/provenencia/pull/478)) added `CopySource` end to end: proto method and messages, `handlers.CopySource` through `writes.Run` returning `writeEffectsProto`, a `dispatch.go` case, `GoStore.copySource`, `FakeStore.copySource`, `CoreMethod.copySource`. Copy that shape for every write here. S12-16 ([#492](https://github.com/mendahu/provenencia/pull/492)) is the precedent for a response carrying `WriteEffects` that the app applies. `sourceGraphProgress(project:)` is the precedent for a project-wide cached list key. Do not copy `CopySource`'s inline id loop (D-6).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Seven methods, appended after `METHOD_COPY_SOURCE = 102`: `METHOD_LIST_EVIDENCE_GRAPHS = 103`, `METHOD_CREATE_EVIDENCE_GRAPH = 104`, `METHOD_UPDATE_EVIDENCE_GRAPH = 105`, `METHOD_DELETE_EVIDENCE_GRAPH = 106`, `METHOD_LINK_EVIDENCE_GRAPH_ARTIFACT = 107`, `METHOD_UNLINK_EVIDENCE_GRAPH_ARTIFACT = 108`, `METHOD_GET_EVIDENCE_GRAPH_UNLINK_IMPACT = 109`. All handlers live in one file, `api/ffi/handlers/evidence_graphs.go` *(new)*, one per method | Working rule: the concept is `EvidenceGraph` in code. One file per domain package, as `artifacts.go` | Bare `Graph` names — taken by `graphcache.Graph` and the Source graph RPCs; separate link and unlink only through `Update` — S12-13's Artifact card edits one Artifact's links across graphs (map scope entry S12-13) |
| D-2 | `ListEvidenceGraphs` takes only `project_dir` and returns `repeated EvidenceGraph graphs` from `evidencegraphs.List(c, nil)`, in its order. `message EvidenceGraph { string id = 1; string ref = 2; string source_id = 3; string name = 4; string description = 5; repeated string artifact_ids = 6; }`; ids are `database.UUIDString`. One unexported `evidenceGraphProto(g evidencegraphs.Graph) *engine.EvidenceGraph` builds it for List and Create | Working rule "One graph list": one project-wide RPC; readers filter by Source in Swift | A `source_id` filter on the RPC — a second cached shape per Source; `List` keeps its filter for Go callers |
| D-3 | Writes go through `writes.Run` with actions `create_evidence_graph`, `update_evidence_graph`, `delete_evidence_graph`, `link_evidence_graph_artifact`, `unlink_evidence_graph_artifact`. `CreateEvidenceGraphResponse { EvidenceGraph graph = 1; WriteEffects effects = 2; }`; every other write response is `{ WriteEffects effects = 1; }`. Requests carry `project_dir = 1`, `user_id = 2`, then: Create `source_id = 3, name = 4, description = 5, repeated string artifact_ids = 6`; Update `graph_id = 3, name = 4, description = 5, repeated string artifact_ids = 6` (comment "empty = links unchanged; otherwise the full set"); Delete `graph_id = 3`; Link and Unlink `graph_id = 3, artifact_id = 4` | `add-ffi-handler` § Catalog RPCs: a write returns its effects. Create returns the graph because the S12-12 form opens it; Update and Delete need nothing but the list refresh the effects cause. No re-read after Update: `List` is the only read (S12-04 D-8) | Returning the graph from Update — needs a `Get` the working rule forbids or a full `List` per edit |
| D-4 | `GetEvidenceGraphUnlinkImpactRequest { project_dir = 1; graph_id = 2; repeated string artifact_ids = 3; }` → `GetEvidenceGraphUnlinkImpactResponse { repeated EvidenceGraphUnlinkBlock blocked = 1; }`, `EvidenceGraphUnlinkBlock { string artifact_id = 1; DeleteImpactGroup citations = 2; }`. Blocks follow the request's Artifact order; an Artifact with no blocking Citation, or not linked to the graph, has no block. The handler runs inside `withProjectCatalog`, maps each block's group with `groupToProto(g deleteimpact.Group) *engine.DeleteImpactGroup` *(new)*, the loop body of `groupsToProto` moved into its own function (`groupsToProto` calls it), and is not a write | S12-04 review UX-1 (map scope entry): the preflight takes the removed set so S12-12 names each blocking Artifact. `DeleteImpactGroup` is the recipe's shape, so S12-12 renders the Citations with `PVDeleteImpact` | One (graph, Artifact) pair per call — N calls per save; attaching the list to the `artifact_in_use` error — `catalog-deletes.md` § FFI: refusals are plain codes |
| D-5 | `deleteimpact.ListedGroup(tx *sql.Tx, via string, kind catalogmodel.Kind, rows []Listed) (Group, error)` *(new)*, in `impact.go` below `collectGroups`. Callers pass rows with `ID` and `Ref` set, from the query that found them. `tx` nil → `ErrInvalid`. `Total` = `len(rows)`; the first `listedCap` rows are replaced by `projectListed(tx, kind, probeRow{ID: r.ID, Ref: r.Ref})`, in input order. Empty `rows` → `Group{Via: via, Kind: kind}`. Doc comment: "ListedGroup shapes rows a domain read found into one report group, projected and capped as Impact lists inbound rows." `CitationsUsingArtifactTx` returns `([]deleteimpact.Listed, error)` with `ID` and `Ref` from the ref it already selects (only those two fields set); `Unlink` keeps its `len(using) > 0` check; its doc comment says "returns the Citations on the Artifact that have an Observation on a Subject of the graph, with their refs, ordered by ref" | `deleteimpact` owns projection (S12-03 review M-1). The ref comes from the one query that found the row, as the probes' `listSQLFn` scans `(id, ref)` together; `evidencegraphs` already imports `deleteimpact`. Review M-1 | A per-row `SELECT ref` built from `tableByKind` — a second way to read refs plus a bucket-has-ref rule; exporting `projectListed` / `probeRow`; a `CitationRef` type — a near-copy of `Listed` |
| D-6 | `parseIDs(raws []string) ([][]byte, error)` *(new)* in `api/ffi/handlers/catalog.go` below `parseID`: `parseID` each, first error returned unchanged; an empty input returns `nil, nil`. `CopySource`, `ReorderSourceMetadata` and `ApplyPromoteGraphAlignment` replace their loops with it, and every new handler that takes ids uses it | Three copies today, four more coming. No caller checks nil against empty (`rg "IDs == nil" core` is empty) | A fourth inline loop per handler |
| D-7 | `CreateArtifactRequest` gains `repeated string graph_ids = 7; // empty = every graph of the Source`. `CreateArtifact` parses it with `parseIDs` before `withProjectCatalog` and passes it as `artifacts.CreateInput.GraphIDs` | Working rule "Artifact ↔ graph links": `graph_ids` carries the choice; empty and absent are one case in proto3 (S12-03 D-11) | A separate `LinkArtifact` call after create — two writes |
| D-8 | `evidencegraphs.UnlinkImpact(c *database.Catalog, graphID []byte, artifactIDs [][]byte) ([]UnlinkBlock, error)` *(new)* and `type UnlinkBlock struct { ArtifactID []byte; Citations deleteimpact.Group }` *(new)*, with `const UnlinkVia = "evidencegraphs.artifact_in_use"` *(new)*, commented "UnlinkVia names the preflight group of Citations that keep an Artifact linked. A domain-preflight via is named for its refusal code, not an FK edge." In this order: `graphID` not 16 bytes or an Artifact id not 16 bytes → `ErrInvalid`; `requireDistinct(artifactIDs)`; open a read transaction as `List` does; `getTx(tx, graphID)` (`sql.ErrNoRows` → `ErrInvalid`); for each Artifact id in order: `IsLinkedTx` false → skip; `CitationsUsingArtifactTx` empty → skip; else `deleteimpact.ListedGroup(tx, UnlinkVia, catalogmodel.KindCitation, rows)` and append the block. Empty `artifactIDs` → `nil, nil` after the graph check | The domain read stays in `evidencegraphs` (working rule "`core/database/evidencegraphs` owns graphs and links") and reuses the exact query `Unlink` refuses on, so preview and write agree. One read transaction, as `List` and `deleteimpact.Snapshot` | The handler calling `CitationsUsingArtifactTx` itself — domain logic in `handlers` (`add-ffi-handler` § Do not) |
| D-9 | `gateToProto` gains `case deleteimpact.GateLastGraph: return engine.DeleteImpactGate_DELETE_IMPACT_GATE_LAST_GRAPH`; proto `DELETE_IMPACT_GATE_LAST_GRAPH = 7`. Swift: `CatalogDeleteImpactGate.lastGraph = "last_graph"`, a `.lastGraph` arm in `GoStore.mapDeleteImpactGate`, `PVDeleteImpactCopy.extraGate` returns `(L10n.DeleteImpact.gateLastGraphTitle, L10n.string(L10n.DeleteImpact.gateLastGraphBody))`, `CatalogDeleteImpactGate` becomes `CaseIterable`, and `PVDeleteImpactControls.isExtraGate` (only a test calls it) is deleted: `extraGatesHaveNoDestructiveButtonAndNoInventedList` loops over `CatalogDeleteImpactGate.allCases` where `PVDeleteImpactCopy.extraGate(for:ref:) != nil` instead of a literal list | Map scope entry; the recipe shows every extra gate the same way, so S12-11's Delete needs no rule of its own | A Swift-side "last graph" check — the gate exists so `GetDeleteImpact` and `Delete` agree (S12-04 D-15) |
| D-10 | Swift store methods, in `GenealogyStore` after `ingestArtifactFile`: `listEvidenceGraphs(projectDir:) -> [CatalogEvidenceGraph]`; `createEvidenceGraph(projectDir:userID:sourceID:name:description:artifactIDs:) -> (CatalogEvidenceGraph, CatalogWriteEffects)`; `updateEvidenceGraph(projectDir:userID:graphID:name:description:artifactIDs:) -> CatalogWriteEffects`; `deleteEvidenceGraph(projectDir:userID:graphID:) -> CatalogWriteEffects`; `linkEvidenceGraphArtifact(projectDir:userID:graphID:artifactID:) -> CatalogWriteEffects`; `unlinkEvidenceGraphArtifact(projectDir:userID:graphID:artifactID:) -> CatalogWriteEffects`; `getEvidenceGraphUnlinkImpact(projectDir:graphID:artifactIDs:) -> [CatalogEvidenceGraphUnlinkBlock]`; all `async throws`. `createArtifact` gains `graphIDs: [String]` after `description`; `SourceArtifactsSection.create()` passes `graphIDs: []` | Mirrors the RPCs one to one, as every store method does. Models in `GenealogyStore.swift` beside `CatalogArtifact`: `struct CatalogEvidenceGraph: Sendable, Equatable, Identifiable { var id, ref, sourceID, name, description: String; var artifactIDs: [String] }` and `struct CatalogEvidenceGraphUnlinkBlock: Sendable, Equatable { var artifactID: String; var citations: CatalogDeleteImpactGroup }`. GoStore maps with `mapEvidenceGraph` *(new)* and the existing `mapDeleteImpactGroup` | A protocol-extension default for `graphIDs` — hides the parameter S12-13 must pass; a new App file — `add-swift-test`'s project wiring for no gain |
| D-11 | `CatalogQueryKey.evidenceGraphsList(project: ProjectKey)` *(new)*, `Kind.evidenceGraphsList`; value type `[CatalogEvidenceGraph]`. Registry `Spec(kind: .evidenceGraphsList, stalePolicy: .sessionFresh, invalidateOn: [])` with the comment "Graph and link writes report the `evidence_graphs` list in `WriteEffects`; a Source delete releases its graphs the same way." Default `whenHidden` (`.revalidate`). Loader: `store.listEvidenceGraphs(projectDir:)`. `invalidation(kind:mutation:)` arm: `.key(.evidenceGraphsList(project: project))`. `static func invalidations(for list:)`: `case .evidenceGraphs: return [.key(.evidenceGraphsList(project: project))]` (the comment goes). `WorkspaceSession.warmQuery` arm `let _: QueryHandle<[CatalogEvidenceGraph]> = query(key)` | Working rules "One graph list" and "The graph list has its own effect". Name follows `sourcesList` / `personsList`; S12-08 adds `evidenceGraph(graphId:)` for the per-graph canvas, so the list key needs the `List` suffix. Revalidate: the Sources list accordion (S12-14) and Source page (S12-11) both keep it on screen | `invalidateOn: [.deletedSource]` — the effects already carry it; `.markStale` — the readers are list screens that must show a new graph at once |
| D-12 | Error copy and keys (§6 Strings), mapped in `L10n.Errors.message(code:params:)` beside `citations.in_use`. No params on any code | Working rule "Refusals are of two kinds": plain codes. Copy follows the voice of `citationsInUse` and `observationsEdgeLocked`; "evidence graph" lower-case in sentences, as `gateEdgeLockedBody` | One shared "can't change this graph" message — the researcher could not tell the five apart |
| D-13 | `viaFormats` gains `"subjects.graph_id"` → key `deleteImpact.via.subjectsGraphId`, "This graph has %lld subjects" (one: "This graph has %lld subject"), and `"evidencegraphs.artifact_in_use"` → key `deleteImpact.via.evidenceGraphsArtifactInUse`, "%lld citations in this graph use this artifact" (one: "%lld citation in this graph uses this artifact"). Both are plural variations in `Localizable.xcstrings`, as `deleteImpact.via.subjectsSourceId` | Map scope entry (S12-03 review UX-1); the second is the heading S12-12 shows over the preflight group | Reusing `citations.artifact_id` ("%lld citations are drawn from this artifact") for the preflight — wrong for Citations that are blocked only in this graph |
| D-14 | `FakeStore` gains `var evidenceGraphsByProject: [String: [CatalogEvidenceGraph]] = [:]` and `var unlinkImpactBlocks: [CatalogEvidenceGraphUnlinkBlock] = []`. Methods: list returns the project's graphs sorted by `name` (`localizedCaseInsensitiveCompare`), then `ref`; create appends a graph with id `UUID().uuidString.lowercased()` and ref `GPH-FAKE1`, throws `StoreBoom.boom` for an empty trimmed name or no Artifact; update replaces name and description and, when `artifactIDs` is non-empty, the links; delete removes the graph; link and unlink add or remove the id; each throws `StoreBoom.boom` for an unknown graph id. `getEvidenceGraphUnlinkImpact` returns `unlinkImpactBlocks` filtered to the asked Artifact ids, in their order. Every write returns `CatalogWriteEffects(lists: [.evidenceGraphs])`. `createArtifact` links the new Artifact to the given graphs, or to every graph of its Source when `graphIDs` is empty, and always adds `.evidenceGraphs` to its effects' `lists`, as Go does (every Source has a graph). `deleteArtifact` removes the id from every graph's `artifactIDs`; `deleteSource` drops that Source's graphs. Each method calls `markCatalogSessionHeld(projectDir)` as `deleteArtifact` does | A fake that keeps the list shape and the effects the session reacts to; refusal rules stay in Go and are tested there | Re-implementing the link rules in the fake — a second copy of the engine rules |
| D-15 | `ThrowingStore` (`OnboardingModelTests.swift`) gains the seven methods, each `throw` like its neighbors, and the `graphIDs` parameter | It conforms to `GenealogyStore`; the compiler requires it | — |
| D-16 | Go handler tests live in `api/ffi/handlers/evidence_graphs_test.go` *(new)* and use `runRPC` and handlers only. Fixtures start from `citationFixture` (`citations_test.go`: a Source with its Default graph, one Artifact, a place Subject and the `toponym` property), add a second Artifact with `CreateArtifact`, and add a Citation on the first Artifact with one `toponym` Observation on the Subject with `CreateCitationWithObservations`, as `citations_test.go` does. A second graph comes from `CreateEvidenceGraph`. Source B (AC-1.1, AC-2.2, AC-3.3) is made in the same project: `ListSourceTypes` for a type id, then `CreateSource` and `CreateArtifact`. Arrange in the test body with direct handler calls, as `citations_test.go` does; `runRPC` runs only the call under test. Deleting the Default graph once a second exists is the "graph with Subjects" case (the Subject was made while the Source had one graph, as S12-04 D-14). Swift graph store tests go in `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` *(new)*, wired into the test target per `add-swift-test` | Handler tests exercise the wire, as `TestCreateArtifactAndIngest`; `citationFixture` already builds the Subject and Citation the in-use cases need | A new fixture helper beside `citationFixture` — a near-copy; `observationstest` helpers inside `withProjectCatalog` — reaches past the handlers under test |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `deleteimpact.ListedGroup` *(new)* | func | `collectGroups`, `projectListed` | `collectGroups` lists through registered probes; the preflight's rows come from a domain read. `ListedGroup` wraps the same `projectListed` and `listedCap` and takes the rows' refs from the caller (D-5) |
| `CitationsUsingArtifactTx` returns `[]deleteimpact.Listed` (changed) | func | its own `c.ref` column, selected and dropped today | Keeps the ref the query already reads (D-5) |
| `evidencegraphs.UnlinkImpact`, `UnlinkBlock`, `UnlinkVia` *(new)* | func / type / const | `CitationsUsingArtifactTx`, `Unlink`, `deleteimpact.Snapshot` | A read over many Artifacts in one transaction, built on the query `Unlink` refuses with (D-8) |
| `groupToProto` *(new)* | func | `groupsToProto`'s loop body | One group's conversion, shared by the report and the preflight (D-4) |
| `parseIDs` *(new)* | func | `parseID`; the three inline loops (replaced) | One helper for every repeated id field (D-6) |
| `api/ffi/handlers/evidence_graphs.go` *(new)*: seven handlers, `evidenceGraphProto` | file / FFI | `artifacts.go`, `sources.go` `CopySource` | One handler per method, as for every domain package (D-1) |
| Proto `EvidenceGraph`, seven request / response pairs, `EvidenceGraphUnlinkBlock`, `DELETE_IMPACT_GATE_LAST_GRAPH`, `CreateArtifactRequest.graph_ids` *(new)* | FFI | `Artifact`, `DeleteImpactGroup`, `CopySourceRequest.artifact_ids` | Reuses `WriteEffects` and `DeleteImpactGroup`; no new effect or report shape |
| `CatalogEvidenceGraph`, `CatalogEvidenceGraphUnlinkBlock` *(new)* | type | `CatalogArtifact`, `CatalogDeleteImpactGroup` | Swift mirror of the proto rows (D-10) |
| Seven `GenealogyStore` methods, `GoStore.mapEvidenceGraph`, seven `CoreMethod` constants *(new)* | func / const | `copySource`, `mapArtifact` | One store method per RPC |
| `CatalogQueryKey.evidenceGraphsList` *(new)* | key | `sourceGraphProgress(project:)` | The one graph list key (D-11) |
| `CatalogDeleteImpactGate.lastGraph`, `L10n.DeleteImpact.gateLastGraphTitle` / `Body` *(new)* | case / copy | `edgeLocked`, `gateEdgeLockedTitle` / `Body` | One more extra gate in the recipe (D-9) |
| `L10n.Errors.evidenceGraphsInvalid`, `evidenceGraphsArtifactInUse`, `evidenceGraphsLastArtifact`, `evidenceGraphsInUse`, `evidenceGraphsLastGraph` *(new)* | copy | `citationsInUse`, `observationsEdgeLocked` | One entry per code (D-12) |
| Two `viaFormats` entries *(new)* | copy | `subjects.source_id` | (D-13) |
| `FakeStore.evidenceGraphsByProject`, `unlinkImpactBlocks` *(new)* | test double | `artifactsBySource` | (D-14) |
| `TestListedGroup`, `TestUnlinkImpact`, `TestListEvidenceGraphs`, `TestCreateEvidenceGraph`, `TestUpdateEvidenceGraph`, `TestDeleteEvidenceGraph`, `TestLinkUnlinkEvidenceGraphArtifact`, `TestGetEvidenceGraphUnlinkImpact`, `TestParseIDs` *(new)*; `EvidenceGraphStoreTests` *(new)* | test | `TestCopySource`, `SubjectStoreTests` | Same harnesses |

**Builds on.** `evidencegraphs` (S12-03 / S12-04) for every rule; `writes.Run` and `writeEffectsProto` for effects; the `deleteimpact` report shape and the `PVDeleteImpact` recipe; the `CatalogList` → key mapping S12-03 left a slot for; `add-ffi-handler`, `add-workspace-place`, `add-localized-string`.
**Deletes.** `PVDeleteImpactControls.isExtraGate` and the test's literal gate list (D-9); the per-group loop body in `groupsToProto` (moved to `groupToProto`); the inline id-parse loops in `CopySource`, `ReorderSourceMetadata` and `ApplyPromoteGraphAlignment`; the "S12-05 adds the graph list's key here." comment and its `return []`.
**Next adds.** S12-08 changes `CreateArtifact`'s and the Subject RPCs' Source ids to graph ids beside these RPCs. S12-10 filters `evidenceGraphsList` by graph to offer linked Artifacts. S12-11 reads `evidenceGraphsList` for one Source, deletes through `GetDeleteImpact` (kind `evidence_graph`) and `deleteEvidenceGraph`, and adds the kind noun. S12-12 calls `createEvidenceGraph` (using the returned graph only for its id; display reads the list key) / `updateEvidenceGraph` and, before an edit that removes Artifacts, `getEvidenceGraphUnlinkImpact`, rendering each block's group with `PVDeleteImpact`. S12-13 passes `graphIDs` from Add Artifact and calls link / unlink from the card. Any later preflight over a domain read reuses `deleteimpact.ListedGroup`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/deleteimpact` | `func ListedGroup(tx *sql.Tx, via string, kind catalogmodel.Kind, rows []Listed) (Group, error)` *(new)* in `impact.go` (D-5) |
| `core/database/evidencegraphs` | `CitationsUsingArtifactTx` returns `[]deleteimpact.Listed` (D-5); `const UnlinkVia = "evidencegraphs.artifact_in_use"` *(new)*; `type UnlinkBlock struct { ArtifactID []byte; Citations deleteimpact.Group }` *(new)*; `func UnlinkImpact(c *database.Catalog, graphID []byte, artifactIDs [][]byte) ([]UnlinkBlock, error)` *(new)* (D-8) |

Errors: `deleteimpact.ErrInvalid` → `ListedGroup` with a nil `tx`; `evidencegraphs.ErrInvalid` → `UnlinkImpact` with a bad or unknown id or a repeated Artifact id.

### FFI

`api/proto/engine.proto`:

```proto
// enum Method, after METHOD_COPY_SOURCE = 102:
METHOD_LIST_EVIDENCE_GRAPHS = 103;
METHOD_CREATE_EVIDENCE_GRAPH = 104;
METHOD_UPDATE_EVIDENCE_GRAPH = 105;
METHOD_DELETE_EVIDENCE_GRAPH = 106;
METHOD_LINK_EVIDENCE_GRAPH_ARTIFACT = 107;
METHOD_UNLINK_EVIDENCE_GRAPH_ARTIFACT = 108;
METHOD_GET_EVIDENCE_GRAPH_UNLINK_IMPACT = 109;

// CreateArtifactRequest:
repeated string graph_ids = 7; // empty = every graph of the Source

// enum DeleteImpactGate:
DELETE_IMPACT_GATE_LAST_GRAPH = 7;

message EvidenceGraph {
  string id = 1;
  string ref = 2;
  string source_id = 3;
  string name = 4;
  string description = 5;
  repeated string artifact_ids = 6; // by Artifact ref from List; in request order from Create
}

message ListEvidenceGraphsRequest { string project_dir = 1; }
message ListEvidenceGraphsResponse { repeated EvidenceGraph graphs = 1; }

message CreateEvidenceGraphRequest {
  string project_dir = 1;
  string user_id = 2;
  string source_id = 3;
  string name = 4;
  string description = 5;
  repeated string artifact_ids = 6; // at least one
}
message CreateEvidenceGraphResponse { EvidenceGraph graph = 1; WriteEffects effects = 2; }

message UpdateEvidenceGraphRequest {
  string project_dir = 1;
  string user_id = 2;
  string graph_id = 3;
  string name = 4;
  string description = 5;
  repeated string artifact_ids = 6; // empty = links unchanged; otherwise the full set
}
message UpdateEvidenceGraphResponse { WriteEffects effects = 1; }

message DeleteEvidenceGraphRequest { string project_dir = 1; string user_id = 2; string graph_id = 3; }
message DeleteEvidenceGraphResponse { WriteEffects effects = 1; }

message LinkEvidenceGraphArtifactRequest { string project_dir = 1; string user_id = 2; string graph_id = 3; string artifact_id = 4; }
message LinkEvidenceGraphArtifactResponse { WriteEffects effects = 1; }
message UnlinkEvidenceGraphArtifactRequest { string project_dir = 1; string user_id = 2; string graph_id = 3; string artifact_id = 4; }
message UnlinkEvidenceGraphArtifactResponse { WriteEffects effects = 1; }

// Which of the Artifacts a graph edit removes are used by the graph's Citations.
message GetEvidenceGraphUnlinkImpactRequest { string project_dir = 1; string graph_id = 2; repeated string artifact_ids = 3; }
message EvidenceGraphUnlinkBlock { string artifact_id = 1; DeleteImpactGroup citations = 2; }
message GetEvidenceGraphUnlinkImpactResponse { repeated EvidenceGraphUnlinkBlock blocked = 1; }
```

Lay the messages out one field per line, as the rest of `engine.proto`; the sketch is compressed. Put them after the `CopySource` messages. Regenerate with `./scripts/generate-proto.sh` (Go `api/proto/engine/engine.pb.go` and Swift `macos/App/Platform/Generated/engine.pb.swift`, both committed).

Handlers (`api/ffi/handlers/evidence_graphs.go` *(new)*), each: unmarshal (`unmarshalErr("<action>", err)`), `parseUserID` for writes, `parseID` / `parseIDs` before the catalog, then `withProjectCatalog`. Writes call the matching `evidencegraphs` function inside `writes.Run` (D-3) and return `writeEffectsProto(wrote)`. `ListEvidenceGraphs` and `GetEvidenceGraphUnlinkImpact` are reads (`evidencegraphs.List(c, nil)`, `evidencegraphs.UnlinkImpact`). Dispatch: seven `Method…` constants and `case`s in `api/ffi/dispatch.go` after `MethodCopySource`.

### macOS

No view. Store and cache only:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Graph row, unlink block | `CatalogEvidenceGraph`, `CatalogEvidenceGraphUnlinkBlock` | `macos/App/Platform/GenealogyStore.swift` | D-10 |
| Store methods | `GenealogyStore`, `GoStore`, `FakeStore`, `ThrowingStore` | `macos/App/Platform/`, `macos/ProvenenciaTests/OnboardingModelTests.swift` | D-10, D-14, D-15 |
| Method ids | `CoreMethod.listEvidenceGraphs`, `createEvidenceGraph`, `updateEvidenceGraph`, `deleteEvidenceGraph`, `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact`, `getEvidenceGraphUnlinkImpact` | `macos/App/Platform/CoreInvoke.swift` | After `copySource` |
| Graph list key | `CatalogQueryKey.evidenceGraphsList` | `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift` | D-11 |
| Last-graph gate | `CatalogDeleteImpactGate.lastGraph` | `GenealogyStore.swift`, `GoStore.swift`, `PVDeleteImpact.swift` | D-9 |

States: none on screen. Accessibility: none (no view).

### Strings

L10n keys *(new)*, all in `macos/App/Platform/L10n.swift` and `macos/App/Resources/Localizable.xcstrings`:

- `L10n.Errors.evidenceGraphsInvalid` — `error.evidencegraphs.invalid` = "That evidence graph isn’t valid."
- `L10n.Errors.evidenceGraphsArtifactInUse` — `error.evidencegraphs.artifact_in_use` = "Citations in this evidence graph use that artifact, so it stays linked."
- `L10n.Errors.evidenceGraphsLastArtifact` — `error.evidencegraphs.last_artifact` = "An evidence graph needs at least one artifact."
- `L10n.Errors.evidenceGraphsInUse` — `error.evidencegraphs.in_use` = "This evidence graph still has subjects."
- `L10n.Errors.evidenceGraphsLastGraph` — `error.evidencegraphs.last_graph` = "A source keeps at least one evidence graph."
- `L10n.DeleteImpact.gateLastGraphTitle` — `deleteImpact.gate.lastGraph.title` = "This is the source’s only evidence graph"
- `L10n.DeleteImpact.gateLastGraphBody` — `deleteImpact.gate.lastGraph.body` = "A source keeps at least one evidence graph. Add another graph before you delete this one."
- `viaFormats["subjects.graph_id"]` — `deleteImpact.via.subjectsGraphId` = one "This graph has %lld subject", other "This graph has %lld subjects"
- `viaFormats["evidencegraphs.artifact_in_use"]` — `deleteImpact.via.evidenceGraphsArtifactInUse` = one "%lld citation in this graph uses this artifact", other "%lld citations in this graph use this artifact"

Comments: "FFI error <code>" for each error; "Extra-gate title for last_graph", "Extra-gate body for last_graph"; "Inbound heading for subjects.graph_id; argument is how many"; "Heading over the Citations that keep an Artifact linked to a graph; argument is how many". Removed: none.

## 7. Steps

**STEP-1 — Preflight read in the engine**
Files: `core/database/deleteimpact/impact.go`, `core/database/deleteimpact/impact_test.go`, `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`. Skill: `add-catalog-delete`.
Do: add `ListedGroup` and change `CitationsUsingArtifactTx`'s return (D-5); add `UnlinkImpact`, `UnlinkBlock`, `UnlinkVia` (D-8); write T-1 and T-2; `TestMembership` (the one test that calls `CitationsUsingArtifactTx`) is updated for the new return type only and also checks the returned ref.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/... ./core/database/evidencegraphs/...` passes.

**STEP-2 — Proto**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`. Skill: `add-ffi-handler`.
Do: add the methods, messages, `graph_ids` and `DELETE_IMPACT_GATE_LAST_GRAPH` (§6 FFI); run `./scripts/generate-proto.sh`.
Check: `go build ./...` passes; `git diff --stat` shows both generated files changed.

**STEP-3 — Shared id parsing, Artifact graph ids, last-graph gate on the wire**
Files: `api/ffi/handlers/catalog.go`, `api/ffi/handlers/sources.go`, `api/ffi/handlers/promote_batch.go`, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/delete_impact.go`, `api/ffi/handlers/catalog_id_test.go`, `api/ffi/handlers/artifacts_test.go`, `api/ffi/handlers/delete_impact_test.go`. Skill: `add-ffi-handler`.
Do: add `parseIDs` and move the three loops onto it (D-6); pass `graph_ids` in `CreateArtifact` (D-7); add the `GateLastGraph` case (D-9); write T-3, T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, including `TestCopySource` and the promote batch tests unchanged.

**STEP-4 — Graph handlers**
Files: `api/ffi/handlers/evidence_graphs.go` *(new)*, `api/ffi/handlers/evidence_graphs_test.go` *(new)*, `api/ffi/dispatch.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: the seven handlers (D-1 – D-4), the dispatch constants and cases; write T-6 – T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-5 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-swift-test` (for `ThrowingStore`).
Do: models and methods (D-10), `CoreMethod` constants, `createArtifact(graphIDs:)` and its call site, `CatalogDeleteImpactGate.lastGraph` with its `GoStore` and recipe arms (D-9; the copy keys land in STEP-8, so add them to `L10n.DeleteImpact` here with the §6 English defaults), `ThrowingStore` (D-15). Delete `isExtraGate` and rewrite `extraGatesHaveNoDestructiveButtonAndNoInventedList` over `allCases` here (T-15's first half), so the test target keeps compiling. `FakeStore.createArtifact` gains the `graphIDs` parameter and ignores it until STEP-6; the seven new `FakeStore` methods are stubs that `throw StoreBoom.boom` until STEP-6. `check-localizable-xcstrings.py` may fail on the two gate keys until STEP-8 adds their catalog entries.
Check: `xcodebuild build` (the `scripts/check-pr.sh` Xcode command without `test`) succeeds on macOS.

**STEP-6 — FakeStore graphs**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-14; wire the new test file per `add-swift-test`; write T-12.
Check: `xcodebuild test` with `-only-testing:ProvenenciaTests/EvidenceGraphStoreTests` passes.

**STEP-7 — Graph list key**
Files: `macos/App/Features/Workspace/Session/CatalogQueryKey.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place` (§3 only).
Do: D-11. In `CatalogQueryKey.swift`: the case, the `Kind` case, the `kind` arm and the `project` arm. In `CatalogQueryRegistry.swift`: the `Spec` directly after the `sourceGraphProgress` spec, the loader arm, the `invalidation(kind:mutation:)` arm and the list mapping line. Write T-13.
Check: `xcodebuild test -only-testing:ProvenenciaTests/CatalogQueryRegistryTests` passes.

**STEP-8 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/L10nResolutionTests.swift`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`. Skill: `add-localized-string`.
Do: §6 Strings (D-12, D-13); write T-14, `lastGraphGateCopy` (T-15's second half) and T-16.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -only-testing:ProvenenciaTests/L10nResolutionTests -only-testing:ProvenenciaTests/PVDeleteImpactTests` passes.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status, Landed note), `docs/catalog-deletes.md` (§ FFI, after the "report proto is only on `GetDeleteImpact`" bullet, this bullet: "An Evidence graph unlink is refused with a plain code too; its preflight is `GetEvidenceGraphUnlinkImpact`, which returns one `DeleteImpactGroup` of Citations per blocking Artifact.").
Also: one line in `.agents/skills/add-catalog-delete/SKILL.md` and one in `core/database/deleteimpact/README.md`: "A domain invariant's refusal (a link, not a row delete) gets a preflight that returns a `deleteimpact.Group` through `ListedGroup`, from the same query the write refuses on; its via is named for the refusal code." One line in `.agents/skills/add-ffi-handler/SKILL.md` § Layout: "Parse a repeated id field with `parseIDs`."
When: the last commit, after the draft PR is open, so its number goes in the plan's Status, the map row (`Landed #N`) and `completed.md` (a "**Landed YYYY-MM-DD** in [#N]" line linking the PR), as S12-04 did.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-05-graph-rpcs.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/deleteimpact/impact_test.go` `TestListedGroup` *(new)*; Citations built in a loop the way `TestImpactCitationAndSubject` builds one | three Citation rows → `Total` 3, three listed with their refs, titles and `citationComposer` locations; 25 rows → `Total` 25, the first 20 listed in input order; empty → `Total` 0, no rows; nil `tx` → `ErrInvalid` | AC-4.4, D-5 |
| T-2 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestUnlinkImpact` *(new)* | AC-4.1 fixture (built with the file's `newFixture`, `fixture.newArtifact` and `fixture.subjectAndCitation`, as `TestUnlink` builds its in-use case) → one block for page 1; `[page 2]` → none; a second graph linked to page 1 → none for that graph (fixture: `newFixture`, `fixture.newArtifact`, `fixture.subjectAndCitation`, then `fixture.createGraph`); unlinked Artifact → none; unknown graph, 15-byte id, repeated id → `ErrInvalid`; blocks follow input order | AC-4.1, AC-4.2, AC-4.3, AC-4.5, D-8 |
| T-3 | Go | `api/ffi/handlers/catalog_id_test.go` `TestParseIDs` *(new)* | two v7 ids → two 16-byte ids in order; empty → nil, nil; one bad → `errInvalidID` | D-6 |
| T-4 | Go | `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactGraphIDs` *(new)* | Source with a second graph (made with `CreateEvidenceGraph`): `graph_ids` [second] → linked to second only (`ListEvidenceGraphs`); none → both; another Source's graph → `evidencegraphs.ErrInvalid`, no Artifact | AC-3.1, AC-3.2, AC-3.3, D-7 |
| T-5 | Go | `api/ffi/handlers/delete_impact_test.go` `TestGetDeleteImpactLastGraph` *(new)* | `sourceFixture` + `CreateSource`, then `ListEvidenceGraphs` for its Default graph; that graph, kind `evidence_graph` → `allowed` false, gate `DELETE_IMPACT_GATE_LAST_GRAPH` | AC-5.2, D-9 |
| T-6 | Go | `api/ffi/handlers/evidence_graphs_test.go` `TestListEvidenceGraphs` *(new)* | AC-1.1 fixture → three rows in that order with refs, `source_id`, `artifact_ids`; bad proto → error | AC-1.1, D-2 |
| T-7 | Go | same file `TestCreateEvidenceGraph` *(new)* | AC-2.1 → graph fields and `effects.lists` ∋ `evidence_graphs`, `revision` > 0; no Artifacts or other Source's Artifact → `evidencegraphs.ErrInvalid`; bad user id → error | AC-2.1, AC-2.2, AC-1.3, D-3 |
| T-8 | Go | same file `TestUpdateEvidenceGraph` *(new)* | AC-2.3 on a graph made with `CreateEvidenceGraph` → name and links per `ListEvidenceGraphs`, effects; AC-2.4 → `evidencegraphs.ErrArtifactInUse`, list unchanged | AC-2.3, AC-2.4, AC-1.3 |
| T-9 | Go | same file `TestDeleteEvidenceGraph` *(new)* | AC-2.5 → `ErrLastGraph`; AC-2.6 → `ErrInUse`; AC-2.7 → gone, effects | AC-2.5, AC-2.6, AC-2.7, AC-1.3 |
| T-10 | Go | same file `TestLinkUnlinkEvidenceGraphArtifact` *(new)* | AC-2.8 → link then unlink, each with effects; AC-2.9 → `ErrLastArtifact` | AC-2.8, AC-2.9, AC-1.3 |
| T-11 | Go | same file `TestGetEvidenceGraphUnlinkImpact` *(new)* | AC-4.1 fixture → one `blocked` entry, `citations.via` `evidencegraphs.artifact_in_use`, `kind` `citation`, `total` 1, one listed with ref; malformed graph id → `errInvalidID` | AC-4.1, D-4 |
| T-12 | Swift | `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `createThenListSortsByName`, `writesReportTheGraphList`, `createArtifactLinksEveryGraphWhenNoneChosen`, `deletesDropLinksAndGraphs`, `unlinkImpactFiltersToAskedArtifacts` *(new)* | FakeStore: create `Smith` then `Jones` → list `Jones`, `Smith`; every write's `lists` == [.evidenceGraphs]; `createArtifact(graphIDs: [])` on a two-graph Source → linked to both, effects list `.evidenceGraphs`; `deleteArtifact` → its id gone from every graph; `deleteSource` → its graphs gone; seeded blocks for pages 1 and 2, asked [2] → block 2 only | AC-1.3, D-14 |
| T-13 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `listInvalidations`: `ListCase` gains `let graphListInvalidations: Int`, 1 for "evidence graphs alone" and 0 for the other five cases, and the test also expects that count of `.key(.evidenceGraphsList(project: project))`; `evidenceGraphsListLoadsFromStore` *(new)* | effects `lists: [.evidenceGraphs]` → the graph list key once, no Properties key; the loader returns FakeStore's seeded graphs | AC-1.2, D-11 |
| T-14 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` `evidenceGraphErrorsHaveCopy` *(new)* | each of the five codes → the §6 sentence, not the unknown default | AC-5.1, D-12 |
| T-15 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` `extraGatesHaveNoDestructiveButtonAndNoInventedList` (loops `allCases` with an extra gate; asserts `.notFound`, `.edgeLocked`, `.originLocked`, `.infra`, `.lastGraph` are among them), `lastGraphGateCopy` *(new)* | `.lastGraph` → extra gate, no destructive action, §6 title and body | AC-5.2, D-9 |
| T-16 | Swift | same file `knownViaHeadingsMatchEngineEdges` (two new rows) | `subjects.graph_id` 1 → "This graph has 1 subject", 3 → "This graph has 3 subjects"; `evidencegraphs.artifact_in_use` 1 and 2 → the §6 forms | AC-5.3, AC-5.4, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestCopySource`, `TestReorderSourceMetadata`-family and promote batch handler tests (STEP-3 moves their id parsing); `CatalogQueryRegistryTests` (exhaustive key switches); `SourcePageModelTests` (`createArtifact` signature).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — n/a, not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: catalog-deletes.md § FFI names the unlink preflight
- [ ] PR description: what shipped, AC evidence, deviations, manual check
- [ ] evidencegraphs.invalid has its own L10n.Errors copy, not the unknown default (T-14)
- [ ] evidencegraphs.artifact_in_use has its own L10n.Errors copy (T-14)
- [ ] evidencegraphs.last_artifact has its own L10n.Errors copy (T-14)
- [ ] evidencegraphs.in_use has its own L10n.Errors copy (T-14)
- [ ] evidencegraphs.last_graph has its own L10n.Errors copy (T-14)
- [ ] subjects.graph_id has its viaFormats heading, not the "reference this" fallback (T-16)
```

## 10. Manual check

No screen reads the new methods, so the check is that nothing visible changed and Add Artifact still links:

1. Open a project with a Source that has Artifacts and Subjects. The Source page, its Evidence graph and the composer look and behave as on `main`.
2. Add an Artifact to that Source from the Artifacts section. It saves as before. Its Evidence graph composer offers the new Artifact (it is linked to the Default graph because `graph_ids` is empty).
3. Delete an Artifact with no Citations. The delete confirm reads as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `parseIDs` returning nil for an empty list changes `CopySource`, `ReorderSourceMetadata` or `ApplyPromoteGraphAlignment` | No callee distinguishes nil from empty today; their existing handler tests run in STEP-3 |
| A new `CatalogQueryKey` case misses a switch | Swift switches over `CatalogQueryKey` and `Kind` are exhaustive; the compiler lists each |
| `CitationsUsingArtifactTx`'s new return type breaks a caller | Its callers are `Unlink` and the new `UnlinkImpact` (`rg CitationsUsingArtifactTx core`); the compiler lists any other |
| `protoc` / `protoc-gen-go` / `protoc-gen-swift` or `xcodebuild` missing on the host | STEP-2 needs the proto tools and Steps 5 – 8 need macOS. Never hand-edit `engine.pb.go` or `engine.pb.swift`; report each Xcode check as not run rather than passed, as S12-03 X-8 recorded |

**Stop and report if:**

- `evidencegraphs.List`, `Create`, `Update`, `Delete`, `Link`, `Unlink`, `LinkArtifact` or `CitationsUsingArtifactTx` no longer has the signature §4 names.
- `CitationsUsingArtifactTx` has a caller besides `Unlink` that needs ids only and cannot take `[]deleteimpact.Listed`.
- `go list -deps ./core/database/deleteimpact` lists `evidencegraphs` (an import cycle; today `evidencegraphs` imports `deleteimpact`, not the reverse).
- Any existing handler test changes behavior after the `parseIDs` move.
- A method number 103 – 109 is already used in `engine.proto`.
- `./scripts/generate-proto.sh` cannot run (no `protoc` or plugins): stop after STEP-1 and report; the generated files are never edited by hand.
- `CatalogList.evidenceGraphs` is missing or the S12-03 comment in `CatalogQueryRegistry.invalidations(for list:)` is gone.
- Showing a new code's copy needs a change in a view (Out).

## 12. Open questions

None.

## 13. PR

**Title:** S12-05 — Graph RPCs and graph list
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-05-graph-rpcs.md`.
