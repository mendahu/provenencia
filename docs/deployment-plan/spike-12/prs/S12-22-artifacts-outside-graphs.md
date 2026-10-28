# S12-22 — Artifacts can stand outside every graph

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the engine half of A-3 and A-8; the views keep what they show today, and S12-38 builds the chrome on brief S12-D7)
**Board:** —
**Depends on:** —
**Blocks:** S12-38 (the Swift half reads `GraphLinks`, the pair-shaped `UnlinkImpact` and the new reason); S12-25 lands after this PR (map: "S12-22 before S12-25, so its new writes follow the narrowed effects")
**Branch:** `s12-22-artifacts-outside-graphs`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-21 did; no schema change, so the catalog format stays as it is ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-3, A-8; [rules](../source-structure.md#rules); [`catalog-deletes.md`](../../../catalog-deletes.md); working rules "An Artifact is the parent; it may stand in no graph", "An Artifact with Citations keeps at least one graph", "One unlink check, one link writer", "An Artifact card's Save is one write", "Refusals are of two kinds", "FakeStore is a seeded fake, not a second engine", "High tolerance for churn"; [map review](../deployment-plan.review.md) K-3, K-6, K-7, K-12; the S12-05 plan (D-13, the `viaFormats` precedent) and the S12-13 Landed note (the card's save loop)
**Planned:** 2026-10-28 against `main` @ `c409b9d`
**Review:** [`S12-22-artifacts-outside-graphs.review.md`](S12-22-artifacts-outside-graphs.review.md) — round 1 done; M-6 answered by the user (2026-10-28); round 2 pending

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

The engine lets an Artifact stand in no Evidence graph and still keeps every Citation reachable. A new or edited Artifact says which graphs it is in with one wire message, `GraphLinks`, where absent means the default and present and empty means none. An Artifact's update saves its fields and its links in one write. One check decides whether a (graph, Artifact) link may go, so the refusal and the preflight cannot disagree, and it adds the rule that an Artifact with Citations keeps at least one graph. A graph delete that would leave such an Artifact in no graph is refused and lists the Citations.

**Why now.** Audit A-3 (High): today a Citation with no Observation can be stranded in no graph, through the card or a graph delete, and its Artifact then can't be deleted. Audit A-8: the card's link save is several transactions run from Swift. The user's fix for both (2026-10-27) makes the Artifact the parent; S12-38 builds the views on this engine half.

**What stays.** Every view looks and behaves as today: the Add artifact dialog still hides its boxes on a one-graph Source and still needs a box checked, the card still saves its links with the per-graph link and unlink RPCs, and the graph form still locks only Artifacts its Citations use. The link and unlink RPCs stay until S12-38 removes their last caller. Graph create, the `last_graph` gate, and Artifact delete keep their rules.

## 2. User stories

**US-1.** As the engine, when an Artifact is created or updated, I link it to exactly the graphs the request names, or to none, so that a page can stand outside every graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source with graphs Default and Jones; `CreateArtifact` with no `graph_links` | The Artifact links Default and Jones |
| AC-1.2 | Same Source; `CreateArtifact` with `graph_links { graph_ids: [Jones] }` | The Artifact links only Jones |
| AC-1.3 | Same Source; `CreateArtifact` with `graph_links {}` (present, empty) | The Artifact exists and links no graph; no link row is written, so the response's `effects.lists` has no `evidence_graphs` |
| AC-1.4 | `CreateArtifact` with `graph_links { graph_ids: [a graph of another Source] }` | `evidencegraphs.invalid`; no Artifact is created |
| AC-1.5 | Artifact P linked to Default; `UpdateArtifact` with a new label and `graph_links { graph_ids: [Jones] }` | One revision holds the label change, the Jones link and the Default unlink; P links only Jones |
| AC-1.6 | Artifact P linked to Default; `UpdateArtifact` with a new label and no `graph_links` | The label changes; P's links are unchanged |
| AC-1.7 | Artifact P with no Citations, linked to Default and Jones; `UpdateArtifact` with the same label and `graph_links {}` | P links no graph; one revision holds both unlinks |
| AC-1.8 | `UpdateArtifact` whose links are refused (AC-2.2) and whose label also changed | The error is the refusal's code; the label, every link and every unlink roll back |

**US-2.** As the engine, I refuse every unlink that would strand a Citation, and I report each refusal before it is tried, so that the preflight and the write agree.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Artifact X with one Citation C that has no Observation, linked to Default and Jones; unlink X from Default | Allowed: X stays in Jones |
| AC-2.2 | Then unlink X from Jones, its last graph | `evidencegraphs.citations_need_graph` *(new)*; the link stays |
| AC-2.3 | Artifact Y with no Citation, linked only to Jones, and Jones also links X; unlink Y from Jones | Allowed: Y is in no graph |
| AC-2.4 | `UnlinkImpact` over the five pairs of T-5's fixture: one per reason, one allowed, one not linked | One block per pair whose unlink is refused, in input order, each with its graph id, Artifact id and reason; the allowed pair and the unlinked pair have none |
| AC-2.5 | For every pair of the T-5 fixture | `Unlink` returns exactly the reason's error when `UnlinkImpact` reports a block, and succeeds when it reports none |
| AC-2.6 | A `citations_need_graph` block | Its Citations group has via `evidencegraphs.citations_need_graph`, kind `citation`, and lists C |
| AC-2.7 | A `last_artifact` block | Its Citations group is empty |

**US-3.** As the engine, I refuse a graph delete that would leave an Artifact with Citations in no graph, and the delete preview lists those Citations.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Source with P1 (Citation C1, no Observation) and P2 linked to Default; Jones made with P2; `GetDeleteImpact(evidence_graph, Default)` | Not allowed, gate `inbound`, one group via `evidencegraphs.citations_need_graph`, kind `citation`, listing C1 |
| AC-3.2 | Same; `DeleteEvidenceGraph(Default)` | `evidencegraphs.in_use`; Default and its links stay |
| AC-3.3 | Same Source but P1 has no Citation | The delete succeeds; P1 is in no graph |
| AC-3.4 | X with a Citation linked to Default and Jones; delete Jones (no Subjects) | Allowed: X stays in Default |

**US-4.** As a researcher, when a refusal from this PR reaches the app, I read what it means.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The engine returns `evidencegraphs.citations_need_graph` | "Citations on this artifact need it to stay in an evidence graph." |
| AC-4.2 | The engine returns `evidencegraphs.in_use` | "Subjects or citations still need this evidence graph." |
| AC-4.3 | The DeleteImpact recipe heads a `evidencegraphs.citations_need_graph` group of 1 / of 3 | "1 citation would be left in no evidence graph" / "3 citations would be left in no evidence graph" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `GraphLinks` on `CreateArtifactRequest` and `UpdateArtifactRequest`; `CreateArtifactRequest.graph_ids` (7) reserved | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8 |
| One Go type, `evidencegraphs.GraphLinks`, that never tells nil from empty | STEP-1, STEP-4, D-1 |
| The Artifact update applies its links through `applyLinks`, lifted from `evidencegraphs.Update`, in one write | STEP-1, STEP-4, AC-1.5, AC-1.6, AC-1.7, AC-1.8 |
| One unlink check behind `Unlink` and `UnlinkImpact`; `UnlinkImpact` over (graph, Artifact) pairs; `EvidenceGraphUnlinkBlock` gains `graph_id` and `reason` | STEP-2, STEP-5, AC-2.4, AC-2.5, AC-2.6, AC-2.7 |
| `Unlink` refuses the last graph of an Artifact with Citations with `evidencegraphs.citations_need_graph` | STEP-2, AC-2.1, AC-2.2, AC-2.3 |
| Graph `Delete` gains an inbound probe for the Citations it would strand, with a `viaFormats` heading; `evidencegraphs.in_use` reworded | STEP-3, STEP-7, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-4.2, AC-4.3 |
| Store methods (`GenealogyStore`, `GoStore`, `FakeStore`) for the new request shapes; FakeStore returns scripted links and refusals | STEP-6, D-12, D-14 |
| The card's link locks come from one `UnlinkImpact` call; `isOnlyArtifact` and the per-graph lock loop go (review M-6, user 2026-10-28) | STEP-6, D-13, T-16 |
| `L10n.Errors` copy for the new code; the reworded `in_use` | STEP-7, AC-4.1, AC-4.2 |
| `source-layer-data-model.md`, `interpretation-layer-data-model.md`, `catalog-deletes.md`, `source-structure.md` rules | STEP-8 |

| Out | Home |
| --- | --- |
| Any visible change: graph boxes on a one-graph Source, none checked allowed, the "Not in an evidence graph" card, the card's one-write Save, the `citations_need_graph` lock in the card and the graph form, a new graph starting with the Artifacts in no graph | S12-38 |
| Deleting `LinkEvidenceGraphArtifact` / `UnlinkEvidenceGraphArtifact` (handlers, proto methods, store methods, `CoreInvoke`, tests), `graphListReloaded()`, the partial-save strings | S12-38 (the card is their last caller) |
| Naming, in a graph delete's confirm, the Artifacts it leaves in no graph | [`docs/ideas/artifact-outside-graph-marker.md`](../../../ideas/artifact-outside-graph-marker.md) |
| `UpdateEvidenceGraphRequest.artifact_ids` moving onto `GraphLinks` (a graph always needs an Artifact, so empty is never "none" there) | not planned |
| Narrowing the Artifact and Citation effects | S12-25 |
| Repairing Citations already stranded in an existing project by A-3's paths | not planned (no user outside the product owner; such a Citation still shows in its Artifact's delete preview) |
| Removing the engine rules FakeStore carries today (`createArtifact`'s linking) | S12-35 |
| Sorting accented graph names | audit L-1, [not scheduled](../app-health-audit.md#not-scheduled) |

## 4. Current state

Verified at `c409b9d`.

| Area | Where | Today |
| --- | --- | --- |
| Graph link writer | `core/database/evidencegraphs/evidencegraphs.go` `Link`, `Unlink`, `checkLink` | `Unlink` refuses `ErrArtifactInUse` (from `CitationsUsingArtifactTx`) then `ErrLastArtifact` (`sqlLinkCount == 1`). Nothing stops the last graph of an Artifact with Citations (A-3) |
| Graph edit | same file `Update` | Computes the stored links from `list(tx, g.SourceID)`, then a links loop and an unlinks loop over `Link` / `Unlink`. Empty `UpdateInput.ArtifactIDs` = links unchanged |
| New Artifact's links | same file `LinkArtifact(tx, userID, sourceID, artifactID, graphIDs [][]byte)` | Empty `graphIDs` = every graph of the Source; there is no way to say none |
| Unlink preflight | same file `UnlinkImpact(c, graphID, artifactIDs)`, `UnlinkBlock{ArtifactID, Citations}`, `UnlinkVia = "evidencegraphs.artifact_in_use"` | One graph, many Artifacts; reports only `artifact_in_use`, so it and `Unlink` can disagree on the last link |
| Sentinels | same file `ErrInvalid`, `ErrArtifactInUse`, `ErrLastArtifact`, `ErrInUse` (KindConflict), `ErrLastGraph`; `core/apperr/apperr.go` `CodeEvidenceGraphs*` (lines 71–75) | No code for "the Artifact's Citations need a graph" |
| Graph delete | same file `Delete` → `deleteimpact.Impact(tx, KindEvidenceGraph, id)` → `deleteimpact.Refuse(report, Codes{InUse: ErrInUse, LastGraph: ErrLastGraph, NotFound: ErrInvalid})` | Refused for Subjects (`subjects.graph_id` probe) and the last graph only (A-3 path 2) |
| Inbound probes | `core/database/deleteimpact/register.go` `inboundEdgeList`, `joinEdge` in `probes.go`; `extraGateList` | One `KindEvidenceGraph` probe, `subjects.graph_id`. `joinEdge` is the precedent for a probe that is not one FK column (`source_metadata.field_id` lists Sources) |
| Artifact write | `core/database/artifacts/artifacts.go` `CreateInput.GraphIDs`, `Create` (calls `evidencegraphs.LinkArtifact`), `Update(tx, userID, a Artifact)` | `Update` writes label, description and first-attach `file_id` only and returns no changes for an unchanged row; links are not part of it (A-8) |
| Other `artifacts.Create` callers | `core/database/sourcecopy/sourcecopy.go` (`in := a.In`), `core/database/artifacts/artifactstest/artifactstest.go`, `core/database/observations/observationstest/observationstest.go` | None sets `GraphIDs`: all rely on "every graph" |
| Fixture that unlinks | `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` (line 96, `Unlink` of the Default graph) | Runs before Citations exist in its callers |
| Handlers | `api/ffi/handlers/artifacts.go` `CreateArtifact` (`parseIDs(req.GetGraphIds())`), `UpdateArtifact`, `IngestArtifactFile` (calls `artifacts.Update` with the new `file_id`); `api/ffi/handlers/evidence_graphs.go` `GetEvidenceGraphUnlinkImpact`, `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact` | `UpdateArtifact` sends label and description only |
| Proto | `api/proto/engine.proto` `CreateArtifactRequest` (`graph_ids = 7`, "empty = every graph"), `UpdateArtifactRequest` (fields 1–5), `GetEvidenceGraphUnlinkImpactRequest` (`graph_id = 2`, `artifact_ids = 3`), `EvidenceGraphUnlinkBlock` (`artifact_id = 1`, `citations = 2`), `DeleteImpactGate` enum (the enum precedent) | — |
| Effects | `core/database/effects/registry.go` `"evidence_graph_artifacts"` | A link row change already reports `ListEvidenceGraphs` and its `CitationSource`; an Artifact update that writes link rows reports them with no registry change |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `createArtifact(… graphIDs: [String])`, `updateArtifact(… label:, description:)`, `getEvidenceGraphUnlinkImpact(projectDir:, graphID:, artifactIDs:)`, `CatalogEvidenceGraphUnlinkBlock{artifactID, citations}`; `macos/App/Platform/GoStore.swift` same three methods and `mapDeleteImpactGate` (unknown enum handling) | — |
| FakeStore | `macos/App/Platform/FakeStore.swift` `createArtifact` (empty `graphIDs` = every graph of the Source), `updateArtifact` (`updateArtifactCalls`, `updateArtifactError`), `getEvidenceGraphUnlinkImpact` (`unlinkImpactBlocksByGraph[graphID]`, `getEvidenceGraphUnlinkImpactGraphIDs`) | `createArtifact` carries an engine rule (every graph); S12-35 removes such rules |
| Test stub | `macos/ProvenenciaTests/OnboardingModelTests.swift` (its own `GenealogyStore` conformance, lines 432–495) | Must follow every protocol change |
| Callers | `macos/App/Features/Sources/SourceArtifactsSection.swift` `loadLinkLocks` (one `getEvidenceGraphUnlinkImpact` per graph), `saveFields` (`updateArtifact` + per-graph link / unlink), `saveDraft` path (`createArtifact(graphIDs: draftGraphIDs)`, retry `updateArtifact`); `draftGraphIDs` returns `[]` when the boxes are hidden; `macos/App/Features/Sources/SourceGraphFormModel.swift` `loadKeptLinks` (one call over `graph.artifactIDs`) | — |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Errors.evidenceGraphsInUse` ("This evidence graph still has subjects."), the `evidencegraphs.*` cases of `L10n.Errors.message`, `L10n.DeleteImpact.viaFormats` (`"evidencegraphs.artifact_in_use"` entry is the precedent); `macos/App/Resources/Localizable.xcstrings` | — |
| Tests | `core/database/evidencegraphs/evidencegraphs_test.go` `TestUnlink`, `TestUnlinkImpact`, `TestUpdate`, `TestDelete`; `core/database/artifacts/artifacts_test.go` `TestArtifacts`; `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactGraphIDs`; `api/ffi/handlers/evidence_graphs_test.go` `graphFixture.createArtifactResp(…, graphIDs []string)`, `TestGetEvidenceGraphUnlinkImpact`; `macos/ProvenenciaTests/EvidenceGraphFixtures.swift` `unlinkBlock`; `EvidenceGraphStoreTests.createArtifactLinksEveryGraphWhenNoneChosen`; `PVDeleteImpactTests` (the via heading table, line ~153); `L10nResolutionTests` (line 89, the `in_use` copy) | — |

**Precedent.** S12-05 ([#511](https://github.com/mendahu/provenencia/pull/511)) added `CreateArtifactRequest.graph_ids`, the unlink preflight and its `viaFormats` heading (`"evidencegraphs.artifact_in_use"` → `deleteImpact.via.evidenceGraphsArtifactInUse`, a plural variation): copy that shape for the new via and the new `L10n.Errors` case. S12-04 ([#505](https://github.com/mendahu/provenencia/pull/505)) made graph edit one write (`evidencegraphs.Update`): lift its loop, do not copy it. `source_metadata.field_id` in `inboundEdgeList` is the `joinEdge` precedent for the delete probe. Do not copy `TestCreateArtifactGraphIDs`'s nil-means-every slice: the new test passes a message or none.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `evidencegraphs.GraphLinks` *(new)*: `type GraphLinks struct { Given bool; IDs [][]byte }`. The zero value means "the default": every graph on create, unchanged on update. `Given` with no `IDs` means none. No code reads `len(IDs) == 0` as a choice | Working rule "An Artifact is the parent" (map review K-6). `evidencegraphs` owns links and already owns `LinkArtifact`; the zero value keeps `sourcecopy`, `artifactstest` and `observationstest` unchanged | Nil vs empty slices — K-6; a type in `artifacts` — the link rules live in `evidencegraphs` |
| D-2 | Proto: `message GraphLinks { repeated string graph_ids = 1; }` *(new)*. `CreateArtifactRequest`: `reserved 7; reserved "graph_ids";` and `GraphLinks graph_links = 8;`. `UpdateArtifactRequest`: `GraphLinks graph_links = 6;`. A message field has presence, so absent and empty differ. Handlers read it through `parseGraphLinks(*engine.GraphLinks) (evidencegraphs.GraphLinks, error)` *(new)* in `api/ffi/handlers/catalog.go`, beside `parseIDs` (review M-7): nil → zero value; non-nil → `Given: true` and `parseIDs` of its ids | Working rule; map scope entry ("`graph_ids` (7) is reserved") | Keeping `graph_ids` beside a bool — two fields for one choice |
| D-3 | `artifacts.CreateInput.GraphIDs` becomes `Graphs evidencegraphs.GraphLinks`. `evidencegraphs.LinkArtifact(tx, userID, sourceID, artifactID, links GraphLinks)`: not `Given` → every graph of the Source (today's rule); `Given` → exactly `IDs`, none allowed; repeated ids → `ErrInvalid`; another Source's graph → `ErrInvalid` (from `Link`'s `checkLink`) | "A re-keyed symbol is renamed in the PR that re-keys it"; one meaning per field | Keeping the name `GraphIDs` with a new type — the name would still read as a slice |
| D-4 | `artifacts.Update(tx, userID, a Artifact, graphs evidencegraphs.GraphLinks)`. It writes the fields as today (an unchanged row writes no `artifacts` change), then, when `graphs.Given`, calls `evidencegraphs.SetArtifactGraphs(tx, userID, a.ID, graphs.IDs)` *(new)* and appends its changes. `IngestArtifactFile` passes `evidencegraphs.GraphLinks{}` | Working rule "An Artifact card's Save is one write": the update owns its links as `Create` does; one `writes.Run`, one revision | The handler calling `SetArtifactGraphs` in the same `Run` — splits one Artifact write across two layers; an `UpdateInput` struct — every other field of `Artifact` would need copying |
| D-5 | `type Pair struct { GraphID, ArtifactID []byte }` *(new)* and one writer, `applyLinks(tx *database.Tx, userID []byte, stored, want []Pair) ([]rowchange.Change, error)` *(new, unexported)*: refuses a repeated pair in `want` (`ErrInvalid`), links every pair of `want` not in `stored` through `Link` (in `want` order), then unlinks every pair of `stored` not in `want` through `Unlink` (in `stored` order), and returns the first error unchanged. Every link loop goes through it: graph `Create` (`stored` nil, `want` = its Artifacts), graph `Update` (its stored links, from `list`), `LinkArtifact` (`stored` nil, `want` = the given graphs, or every graph of the Source when not `Given`) and `SetArtifactGraphs` *(new)* (the Artifact's stored links from `sqlArtifactGraphs` *(new)*: `SELECT graph_id FROM evidence_graph_artifacts WHERE artifact_id = ? ORDER BY graph_id`). `requireDistinct` stays only where an id list is not a pair list (graph `Create` / `Update` inputs before they become pairs go through `applyLinks`'s check, so drop their own calls). An unknown Artifact in `SetArtifactGraphs` is `ErrInvalid` | Working rule "One unlink check, one link writer" (K-3): one diff, one distinct check, one loop. Links before unlinks lets a swap move an Artifact with Citations from its only graph to another | A link list and an unlink list per caller — each caller keeps its own diff (review M-4) |
| D-6 | One check, `checkUnlink(tx *sql.Tx, p Pair) (*UnlinkBlock, []deleteimpact.Listed, error)` *(new, unexported)*: nil when the pair is not linked or may be unlinked. Otherwise the first of, in order: (1) the graph's Citations use the Artifact (`CitationsUsingArtifactTx`) → `ErrArtifactInUse` with those rows; (2) it is the graph's only link → `ErrLastArtifact`, no rows (its block's `Citations` is the zero `deleteimpact.Group`; on the wire `citations` is unset); (3) the Artifact's Citations that this graph alone holds, read with `deleteimpact.SQLCitationsOnlyInGraph + " AND c.artifact_id = ? ORDER BY c.ref COLLATE NOCASE"` (graph id, then Artifact id), are not empty → `ErrCitationsNeedGraph` with those rows. `Unlink` runs `checkLink`, then `checkUnlink`, returns `block.Reason` when there is one, else deletes. One reason per pair | The working rule's point: `Unlink` refuses from it and `UnlinkImpact` reports it. (1) and (2) keep today's order, so existing refusals keep their codes. (3) uses the same SQL as the graph delete probe (D-10), so the unlink refusal, its preflight and the delete preview cannot drift (review M-1) | Reporting every reason of a pair — `Unlink` can return only one code; a count of the Artifact's links plus a Citations query — a second spelling of D-10's rule |
| D-7 | `UnlinkBlock` becomes `{ GraphID, ArtifactID []byte; Reason error; Citations deleteimpact.Group }`. `Reason` is the sentinel itself (`ErrArtifactInUse`, `ErrLastArtifact`, `ErrCitationsNeedGraph`). `checkUnlink` returns the rows; `UnlinkImpact` turns them into a `Group` through `deleteimpact.ListedGroup` with via `UnlinkVia` for (1) and `deleteimpact.ViaCitationsNeedGraph` for (3); (2) has the zero `Group` | The write path does not pay for projecting titles and locations; the sentinel is the one source of truth for the reason | A separate reason enum in Go — a second name for each code |
| D-8 | `UnlinkImpact(c *database.Catalog, pairs []Pair) ([]UnlinkBlock, error)`. A malformed id or a repeated pair → `ErrInvalid`; an unknown graph → `ErrInvalid` (as today); a pair not linked → no block. Blocks come in input order. Pairs may span graphs and Artifacts | Working rule: "the graph form asks for its graph and each Artifact, the card for each graph and its Artifact" | Two entry points (by graph, by Artifact) — two shapes for one check |
| D-9 | New code `apperr.CodeEvidenceGraphsCitationsNeedGraph = "evidencegraphs.citations_need_graph"` *(new)*, `evidencegraphs.ErrCitationsNeedGraph`, `apperr.KindUser` (like `last_artifact`) | Working rule "An Artifact with Citations keeps at least one graph": "its own user-kind code". The name says why the link stays | `evidencegraphs.artifact_last_graph` — reads like the existing `last_graph` (a Source's last graph) |
| D-10 | In `core/database/deleteimpact/register.go`: `const ViaCitationsNeedGraph = "evidencegraphs.citations_need_graph"` *(new)* and `const SQLCitationsOnlyInGraph` *(new, exported)*, the one spelling of "the Citations an Evidence graph alone holds": `citations c WHERE c.artifact_id IN (SELECT l.artifact_id FROM evidence_graph_artifacts l WHERE l.graph_id = ? AND NOT EXISTS (SELECT 1 FROM evidence_graph_artifacts o WHERE o.artifact_id = l.artifact_id AND o.graph_id <> l.graph_id))`, binding one graph id. The graph delete probe is `joinEdge(KindEvidenceGraph, ViaCitationsNeedGraph, KindCitation, "SELECT COUNT(*) FROM " + SQLCitationsOnlyInGraph, "SELECT c.id, c.ref FROM " + SQLCitationsOnlyInGraph + " ORDER BY c.ref COLLATE NOCASE LIMIT ?")` in `inboundEdgeList`, after `subjects.graph_id`; D-6 (3) selects `c.id, c.ref FROM` it too. The consts live in `deleteimpact` because `evidencegraphs` imports `deleteimpact`, not the reverse. `Delete`'s `Codes` are unchanged: the inbound gate maps to `ErrInUse` | Working rule (map review K-7): an inbound probe that lists the Citations, no new extra gate. One SQL, one via and one heading serve the delete confirm and the unlink block, as `evidencegraphs.SQLObservationsInGraph` serves `Unlink` and `citations`. A domain via is named for its refusal code (`core/database/deleteimpact/README.md`) | A `stranded_citations` extra gate — K-7; an FK-named via such as `evidence_graph_artifacts.graph_id` — that column is a released facet; "SoleGraph" in the name — `evidencegraphstest.SoleGraph` means a Source's only graph |
| D-11 | `ErrInUse` keeps `KindConflict` and its code; only its copy changes (Strings) | Working rule: "`evidencegraphs.in_use` is reworded to cover Subjects and these Citations" | A second inbound code — the rule says one |
| D-12 | Swift store: `createArtifact(projectDir:userID:sourceID:fileID:label:description:graphIDs: [String]?)` — `nil` sends no `graphLinks`; `updateArtifact(projectDir:userID:artifactID:label:description:graphIDs: [String]?)` — `nil` sends none; `getEvidenceGraphUnlinkImpact(projectDir:pairs: [CatalogEvidenceGraphArtifactPair])` with `struct CatalogEvidenceGraphArtifactPair: Sendable, Equatable { var graphID: String; var artifactID: String }` *(new)*. `CatalogEvidenceGraphUnlinkBlock` gains `graphID: String` and `reasonCode: String`. Proto `EvidenceGraphUnlinkBlock` gains `string graph_id = 3; string reason_code = 4;`, which the handler sets from `apperr.From(b.Reason).Code()` (`evidencegraphs.artifact_in_use`, `evidencegraphs.last_artifact`, `evidencegraphs.citations_need_graph`) | Swift's optional array says nil vs empty without a second type. The reason is the refusal's own code, as `Error.code` is: Swift reads its copy through `L10n.Errors.message(code:)`, so the lock and the refused write say the same thing (review M-2) | A proto enum, a Swift enum and two mappers — a second vocabulary for each code |
| D-13 | The views keep what they show today. `saveDraft` passes `draftGraphIDs.isEmpty ? nil : draftGraphIDs` (empty today means the boxes are hidden; `draftNeedsGraph` blocks an empty visible choice). Every `updateArtifact` call passes `graphIDs: nil`. `loadLinkLocks` sends **one** `getEvidenceGraphUnlinkImpact` with a pair (graph, Artifact) for every graph of `edit.graphIDs` that links the Artifact (none → ready with no locks, no call), and maps each block by `reasonCode`: `evidencegraphs.artifact_in_use` → `PVDeleteImpactCopy.heading(block.citations)` (today's copy), `evidencegraphs.last_artifact` → `L10n.string(L10n.Sources.artifactGraphOnlyArtifact)` (today's copy), anything else → no lock, with a one-line comment that S12-38 shows it. `CatalogEvidenceGraph.isOnlyArtifact` and the per-graph loop are deleted. `loadKeptLinks` (graph form) sends one pair per linked Artifact and keeps only `artifact_in_use` blocks, as today | Scope: no visible change; the card's locks come from the engine's one check, so the Swift copy of the last-link rule goes now (review M-6, user 2026-10-28) | Keeping the loop and `isOnlyArtifact` until S12-38 — rewrites lines only to delete them |
| D-14 | FakeStore adds no rule. `createArtifact` appends its `graphIDs` to `createArtifactGraphIDs: [[String]?]` *(new)* and keeps today's linking expression with the optional unwrapped (`graphIDs == nil` → every graph of its Source, else the graphs it names); `updateArtifact` appends to `updateArtifactGraphIDs: [[String]?]` *(new)* and moves no link. `unlinkImpactBlocksByGraph` becomes `unlinkImpactBlocks: [CatalogEvidenceGraphUnlinkBlock]` and `getEvidenceGraphUnlinkImpactGraphIDs` becomes `getEvidenceGraphUnlinkImpactPairs: [CatalogEvidenceGraphArtifactPair]`; `getEvidenceGraphUnlinkImpact` records the pairs and returns, in pair order, each scripted block whose (`graphID`, `artifactID`) is a requested pair | Working rule "FakeStore is a seeded fake": calls are recorded and blocks scripted, each fact in one place (review M-3, M-5). Today's `createArtifact` linking is an existing rule S12-35 removes; this PR only changes its argument's type | Testing a none-means-none branch in the fake — the rule is tested in Go (T-1, T-9) |
| D-15 | `EvidenceGraphFixtures.swift` `unlinkBlock(_ artifactID: String, graphID: String, citations: Int, reasonCode: String = "evidencegraphs.artifact_in_use")`, with no default graph; the group's via is the reason code for `artifact_in_use` and `citations_need_graph` (`evidencegraphs.citations_need_graph`), and `CatalogDeleteImpactGroup(via: "", kind: "", total: 0, listed: [])` for `last_artifact`, what `GoStore` maps from an unset `citations`. Each existing call passes the graph its test scripted it under (`SourcePageModelTests` `g-s`, `g-j`; `SourceGraphFormModelTests` `g-smith`, `g-pair`; `EvidenceGraphStoreTests` `g1`) | One source for each scripted fact | A default graph id — it disagrees with the scripted key (review M-5) |
| D-16 | Handler test helpers `graphFixture.createArtifactResp(t, sourceID, label string, links *engine.GraphLinks)` and its wrapper `graphFixture.createArtifact` take `*engine.GraphLinks`; `nil` sends none (today's callers pass `nil`) | The helpers mirror the wire | — |

**Rules this plan bends.** None. D-14 changes an existing FakeStore rule only so that it stops contradicting the wire (empty is no longer every graph) and adds none; "FakeStore is a seeded fake" binds new methods and S12-35 removes the old rules.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `evidencegraphs.GraphLinks` | type | `[][]byte` with "empty = every graph" in `LinkArtifact`, `UpdateInput.ArtifactIDs` | It replaces the `LinkArtifact` slice; the graph edit's slice keeps its meaning (Out) |
| `evidencegraphs.Pair` | type | `UnlinkBlock` (Artifact id only) | Inputs need both ids; blocks embed the same pair |
| `applyLinks` | func | `Create`'s, `Update`'s and `LinkArtifact`'s loops and `Update`'s diff | It is those, lifted into one |
| `SetArtifactGraphs` | func | `LinkArtifact` (create only, add only) | Update needs the diff and unlinks; `LinkArtifact` stays the create path |
| `checkUnlink` | func | `Unlink`'s body, `UnlinkImpact`'s loop | It is those two, merged |
| `sqlArtifactGraphs` | SQL | `sqlListLinksBySource` (by Source) | One Artifact's links inside a write |
| `ErrCitationsNeedGraph`, `CodeEvidenceGraphsCitationsNeedGraph` | error code | `ErrLastArtifact` | A different rule and copy |
| `deleteimpact.ViaCitationsNeedGraph`, `deleteimpact.SQLCitationsOnlyInGraph`, one `joinEdge` entry | probe and shared SQL | `joinEdge(KindMetadataField, …)`; `evidencegraphs.SQLObservationsInGraph` (one exported spelling of a rule) | One more entry in the existing list; the SQL follows the `SQLObservationsInGraph` precedent |
| `parseGraphLinks` | handler helper | `parseIDs` | Wraps it with presence |
| `GraphLinks`, `EvidenceGraphArtifactPair` | proto | `CreateEvidenceGraphRequest.artifact_ids` | New wire concepts; the reason rides as a code string like `Error.code` |
| `CatalogEvidenceGraphArtifactPair` | Swift type | — | Mirrors the proto pair |
| `FakeStore.createArtifactGraphIDs`, `updateArtifactGraphIDs`, `unlinkImpactBlocks`, `getEvidenceGraphUnlinkImpactPairs` | test hooks | `updateArtifactCalls`, `unlinkImpactBlocksByGraph`, `getEvidenceGraphUnlinkImpactGraphIDs` | Record the new arguments; the last two replace the graph-keyed hooks |
| `L10n.Errors.evidenceGraphsCitationsNeedGraph`, `deleteImpact.via.evidenceGraphsCitationsNeedGraph` | copy | `evidenceGraphsLastArtifact`, `deleteImpact.via.evidenceGraphsArtifactInUse` | New code and via |

Searched: `rg "GraphIDs|graph_ids|UnlinkImpact|UnlinkBlock|LinkArtifact|joinEdge|viaFormats"` across `core/`, `api/` and `macos/`.

**Builds on.** `evidencegraphs` as the one link writer (working rule "`core/database/evidencegraphs` owns graphs and links"); `deleteimpact`'s inbound list and `ListedGroup` ([`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md)); the effects registry's `evidence_graph_artifacts` entry (no change); the S12-05 `viaFormats` and `L10n.Errors` shapes.
**Deletes.** `CreateArtifactRequest.graph_ids` (reserved), `GetEvidenceGraphUnlinkImpactRequest.graph_id` and `.artifact_ids` (reserved), `artifacts.CreateInput.GraphIDs`, `Update`'s inline links-then-unlinks loops, `Unlink`'s inline checks and `UnlinkImpact`'s per-Artifact loop (both now `checkUnlink`), `TestCreateArtifactGraphIDs` (replaced by T-9). `CatalogEvidenceGraph.isOnlyArtifact` (`GenealogyStore.swift:109`, its only caller is `loadLinkLocks`) and `loadLinkLocks`'s per-graph loop. Nothing else loses its last caller: the link and unlink RPCs are still called by the card's Save (S12-38 deletes them). `L10n.Sources.artifactGraphOnlyArtifact` stays: the `last_artifact` lock still reads it.
**Next adds.** S12-38 sends `graphIDs: []` from the Add dialog and the card, sends the card's links in its one `updateArtifact`, reads each lock's copy from `L10n.Errors.message(code: block.reasonCode)` or the Citations heading, and drops D-13's filters. A card swap (uncheck the only graph, check another) is allowed by the write because links come first; S12-38 decides how a `citations_need_graph` lock reads when another box is checked. A later reason is one case in `checkUnlink` and its `L10n.Errors` copy.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Artifact in no graph, made through the wire (no view sends it yet) | Citation composer (`EvidenceGraphView` / composer Artifact list) | Offers only a graph's linked Artifacts (S12-10), so an unlinked Artifact is not offered; its Citations can't exist (US-2, US-3) |
| Same | Source page card (`SourceArtifactsSection`) | Shows the card with every box unchecked when two or more graphs; on one graph, no boxes. The "Not in an evidence graph" line is S12-38 |
| Same | Graph form (`SourceGraphFormModel`) | Lists every Artifact of the Source with its box state from the graph; unaffected |
| A graph delete preview with a `evidencegraphs.citations_need_graph` group | DeleteImpact recipe (`PVDeleteImpact`) | Known via after STEP-7 (else the generic heading); each Citation opens through `projectCitation`'s location as today |
| A Citation already stranded by A-3 in an existing project | `projectCitation` (`GraphID ""`) | Unchanged: still reachable through its Artifact's delete preview (Out) |

### Scale

At the map's Scale (a census with 60 graphs and 300 Citations on one page):

- `checkUnlink` per pair: `IsLinkedTx` (PK), `CitationsUsingArtifactTx` (`citations_artifact_id_idx` then `SQLObservationsInGraph`), `sqlLinkCount` (PK prefix), and for (3) `SQLCitationsOnlyInGraph` narrowed to one Artifact (`citations_artifact_id_idx`, then one `NOT EXISTS` on `evidence_graph_artifacts_artifact_id_idx`). A 60-pair `UnlinkImpact` is 60 × these, one transaction; the card's locks are one call from this PR on (60 sequential calls today).
- The graph delete probe walks the graph's links with one indexed `NOT EXISTS` each, then the Citations of the Artifacts it finds; capped by `joinEdge`'s list limit.
- An Artifact update with links writes one `evidence_graph_artifacts` change per added or dropped graph. Its effects add the graph list key and the Source's citation reads, as the link RPCs do today; the Artifact row's own effects (every graph of the Source) are unchanged until S12-25.

### Schema

None — no catalog change. The probe reads existing tables through existing indexes.

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | `CodeEvidenceGraphsCitationsNeedGraph = "evidencegraphs.citations_need_graph"` *(new)* |
| `core/database/deleteimpact` | `const ViaCitationsNeedGraph` and `const SQLCitationsOnlyInGraph` *(new)*; one `joinEdge` entry in `inboundEdgeList` (D-10) |
| `core/database/evidencegraphs` | `ErrCitationsNeedGraph` *(new)*; `type GraphLinks` *(new)*; `type Pair` *(new)*; `applyLinks` *(new)*; `SetArtifactGraphs(tx *database.Tx, userID, artifactID []byte, graphIDs [][]byte) ([]rowchange.Change, error)` *(new)*; `checkUnlink` *(new)*; `Unlink` calls `checkUnlink`; `Create`, `Update` and `LinkArtifact` call `applyLinks`; `LinkArtifact(…, links GraphLinks)`; `UnlinkBlock` gains `GraphID` and `Reason`; `UnlinkImpact(c, pairs []Pair)` |
| `core/database/artifacts` | `CreateInput.Graphs evidencegraphs.GraphLinks` (replaces `GraphIDs`); `Update(tx, userID, a Artifact, graphs evidencegraphs.GraphLinks)` |

```go
// GraphLinks says which graphs an Artifact is in. The zero value is the
// default: every graph of its Source on create, unchanged on update. Given
// with no IDs is none.
type GraphLinks struct {
	Given bool
	IDs   [][]byte
}

// Pair is one (graph, Artifact) link.
type Pair struct{ GraphID, ArtifactID []byte }

// UnlinkBlock is one pair whose unlink is refused, with the refusal and, for
// a Citation reason, those Citations shaped for the DeleteImpact recipe.
type UnlinkBlock struct {
	GraphID, ArtifactID []byte
	Reason              error // ErrArtifactInUse, ErrLastArtifact or ErrCitationsNeedGraph
	Citations           deleteimpact.Group
}
```

Errors: `evidencegraphs.citations_need_graph` → `Unlink` (directly, through `Update`, or through `artifacts.Update`) of an Artifact's only link while it has a Citation. `evidencegraphs.in_use` → graph `Delete` with Subjects or with an Artifact whose only link is this graph and which has a Citation.

### FFI

`api/proto/engine.proto`:

```proto
message GraphLinks {
  repeated string graph_ids = 1; // present and empty = none
}

message CreateArtifactRequest {
  // … fields 1–6 unchanged
  reserved 7;
  reserved "graph_ids";
  GraphLinks graph_links = 8; // absent = every graph of the Source
}

message UpdateArtifactRequest {
  // … fields 1–5 unchanged
  GraphLinks graph_links = 6; // absent = links unchanged
}

message EvidenceGraphArtifactPair {
  string graph_id = 1;
  string artifact_id = 2;
}

message GetEvidenceGraphUnlinkImpactRequest {
  string project_dir = 1;
  reserved 2, 3;
  reserved "graph_id", "artifact_ids";
  repeated EvidenceGraphArtifactPair pairs = 4;
}

message EvidenceGraphUnlinkBlock {
  string artifact_id = 1;
  DeleteImpactGroup citations = 2; // empty for LAST_ARTIFACT
  string graph_id = 3;
  string reason_code = 4; // the refusal's apperr code
}
```

Update the `UpdateArtifactRequest` comment to say it also carries the links. Handlers: `api/ffi/handlers/artifacts.go` `CreateArtifact` and `UpdateArtifact` read `parseGraphLinks(req.GetGraphLinks())` and pass it on; `IngestArtifactFile` passes `evidencegraphs.GraphLinks{}`. `api/ffi/handlers/evidence_graphs.go` `GetEvidenceGraphUnlinkImpact` parses pairs (`parseID` each) and sets `reason_code` from `apperr.From(b.Reason).Code()`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

None visible. Store and caller edits, per D-12 to D-15; the card's locks read as today from one call.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store protocol | `GenealogyStore` | `macos/App/Platform/GenealogyStore.swift` | D-12 signatures; the two new types beside `CatalogEvidenceGraphUnlinkBlock` |
| Go store | `GoStore` | `macos/App/Platform/GoStore.swift` | `req.graphLinks` set only for non-nil; pairs; `reasonCode` copied as is (D-12) |
| Fake | `FakeStore` | `macos/App/Platform/FakeStore.swift` | D-14 |
| Callers | `SourceArtifactsSection` (`loadLinkLocks` rewritten, one call), `SourceGraphFormModel` | `macos/App/Features/Sources/` | D-13 |

### Strings

The PR changes what `evidencegraphs.in_use` covers, so its copy changes.

- `L10n.Errors.evidenceGraphsCitationsNeedGraph` *(new)*, key `error.evidencegraphs.citations_need_graph` = "Citations on this artifact need it to stay in an evidence graph." (comment "FFI error evidencegraphs.citations_need_graph"), with its `case` in `L10n.Errors.message`.
- `L10n.Errors.evidenceGraphsInUse`, key `error.evidencegraphs.in_use`: "This evidence graph still has subjects." → "Subjects or citations still need this evidence graph."
- `L10n.DeleteImpact.viaFormats` `"evidencegraphs.citations_need_graph"` *(new)* → key `deleteImpact.via.evidenceGraphsCitationsNeedGraph`, plural variation: one "%lld citation would be left in no evidence graph", other "%lld citations would be left in no evidence graph" (comment "Heading over the Citations an unlink or a graph delete would leave in no graph; argument is how many").
- Old-scope copy checked and kept: `deleteImpact.via.subjectsGraphId` ("This graph has %lld subjects") still heads the Subjects group; the `last_graph` gate title and body speak of the last graph only.

## 7. Steps

**STEP-1 — One link writer and `GraphLinks`**
Files: `core/apperr/apperr.go`, `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`.
Do: add the code (D-9) and `ErrCitationsNeedGraph`; add `GraphLinks`, `Pair`, `applyLinks`, `sqlArtifactGraphs` and `SetArtifactGraphs` (D-1, D-5); make `Create` and `Update` call `applyLinks`; change `LinkArtifact` to take `GraphLinks` (D-3). Update `artifacts.Create`'s call (`in.GraphIDs` → `in.Graphs`, STEP-4 renames the field) so the tree builds. Write T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/... ./core/database/artifacts/...` passes; `TestCreate` and `TestUpdate` pass unchanged.

**STEP-2 — One unlink check and the pair-shaped preflight**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`. Skill: `add-catalog-delete` (preflight groups).
Do: add `deleteimpact.ViaCitationsNeedGraph` and `deleteimpact.SQLCitationsOnlyInGraph` (D-10; STEP-3 adds the probe that also uses them), then `checkUnlink` (D-6); route `Unlink` through it; change `UnlinkBlock` and `UnlinkImpact` (D-7, D-8). Rewrite `TestUnlinkImpact` and extend `TestUnlink` (T-3 – T-5).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/...` passes.

**STEP-3 — The graph delete probe**
Files: `core/database/deleteimpact/register.go`, `core/database/evidencegraphs/evidencegraphs_test.go`. Skill: `add-catalog-delete`.
Do: add the `joinEdge` entry over `SQLCitationsOnlyInGraph` (D-10). Extend `TestDelete` (T-6, T-7).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/... ./core/database/evidencegraphs/...` passes, `TestResourceAndOwnedHonesty` included.

**STEP-4 — The Artifact writes carry their links**
Files: `core/database/artifacts/artifacts.go`, `core/database/artifacts/artifacts_test.go`, `core/database/audit/scopes_test.go` (`runArtifactUpdate` passes `evidencegraphs.GraphLinks{}`), `api/ffi/handlers/artifacts.go` (only the three call sites, with the zero value, so the tree builds; STEP-5 wires the wire).
Do: rename `CreateInput.GraphIDs` to `Graphs evidencegraphs.GraphLinks`; add the `graphs` parameter to `Update` (D-4). In `TestArtifacts`, the row "create links to the Default graph when GraphIDs is empty or names it" becomes "create links to the Default graph when Graphs is the zero value or names it" (`{}` and `{Given: true, IDs: {graph}}`), and the refusal row passes `{Given: true, IDs: ids}`. Write T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-5 — Wire and handlers**
Files: `api/proto/engine.proto`, generated code, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/evidence_graphs.go`, `api/ffi/handlers/artifacts_test.go`, `api/ffi/handlers/evidence_graphs_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: the proto changes in §6 FFI; `./scripts/generate-proto.sh`; `parseGraphLinks` (in `catalog.go`); the three handlers; D-16's helper. Replace `TestCreateArtifactGraphIDs` with T-9; add T-10; update `TestGetEvidenceGraphUnlinkImpact` to T-11.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./...` passes.

**STEP-6 — Swift store methods and callers**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourceGraphFormModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphFixtures.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`, and each test file that calls the changed methods (`rg "createArtifact\(|updateArtifact\(|getEvidenceGraphUnlinkImpact\(" macos/ProvenenciaTests`). Skill: `add-swift-test`.
Do: D-12, D-13, D-14, D-15. In `EvidenceGraphStoreTests`, every `createArtifact(… graphIDs: [])` that means "every graph" passes `graphIDs: nil` (`createArtifactLinksOnlyItsProjectsGraphs`, `createArtifactLinksEveryGraphWhenNoneChosen`, `createSourceMakesADefaultGraph`); `unlinkImpactFiltersToAskedArtifacts` becomes T-13. Every `unlinkImpactBlocksByGraph[g] = blocks` becomes `unlinkImpactBlocks = blocks` with each `unlinkBlock` naming `g` (D-15); `cardEditLocksInUseAndOnlyArtifactGraphs` becomes T-16. Add T-12. In `SourceArtifactsSection.swift`, rewrite `loadLinkLocks` per D-13 and delete `isOnlyArtifact` from `GenealogyStore.swift`.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes with no Swift warning; `rg "graphIDs: \[\]|unlinkImpactBlocksByGraph|getEvidenceGraphUnlinkImpactGraphIDs" macos` has no hit.

**STEP-7 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PVDeleteImpactTests.swift`, `macos/ProvenenciaTests/L10nResolutionTests.swift`. Skill: `add-localized-string`.
Do: §6 Strings. Add the new via to the heading table test and the new code to the error table; change the `in_use` row.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; T-14 and T-15 pass.

**STEP-8 — Docs**
Files: `docs/source-layer-data-model.md` (line 171 "as every new Artifact is linked to every graph of its Source by default" → "as a new Artifact is by default"; lines 421–422: `GraphLinks`, none allowed, the new refusal, the update's links, `UnlinkImpact` over pairs), `docs/interpretation-layer-data-model.md` (a numbered invariant after 24: an Artifact with Citations stays linked to at least one Evidence graph; a Citation with no Observation is reached through any graph that links its Artifact), `docs/catalog-deletes.md` (rule 10: the `evidencegraphs.citations_need_graph` probe; line 89: one block per refused pair with its graph and reason), `docs/deployment-plan/spike-12/source-structure.md` (Rules: nothing changes, confirm), `completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-guidance-names.py` passes.

Where checks run: STEP-5's `generate-proto.sh` and STEP-6, STEP-7 and the DoD's `scripts/check-pr.sh` need `protoc`, the Go and Swift protoc plugins and `xcodebuild`: run them on the Mac. The generated `engine.pb.go` and `engine.pb.swift` are never edited by hand.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Go tests are table-driven (`.cursor/rules/go-tests.mdc`). Fixtures: the `newFixture` / `evidencegraphstest` helpers the package tests use today, and `graphFixture` in the handlers.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestLinkArtifact` *(new)* | Source with Default and Jones: zero `GraphLinks` → both; `{Given, [Jones]}` → Jones; `{Given}` → none; `{Given, [Jones, Jones]}` → `ErrInvalid`; another Source's graph → `ErrInvalid` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-3 |
| T-2 | Go | same file `TestSetArtifactGraphs` *(new)* | P in Default → `[Jones]`: one create and one delete link change, Jones linked first (change order); `[Jones, Jones]` → `ErrInvalid`; P with no Citations in Default and Jones → `[]`: two deletes; X with a Citation only in Default → `[]`: `ErrCitationsNeedGraph`; X only in Default → `[Jones]` (swap): allowed; unknown Artifact → `ErrInvalid` | AC-1.5, AC-1.7, D-5 |
| T-3 | Go | same file `TestUnlink` (extended) | Today's rows stay; plus X with a Citation and no Observation in Default and Jones: unlink Default ok; unlink Jones → `ErrCitationsNeedGraph`, link stays; Y with no Citation, Jones's other link: unlink → ok | AC-2.1, AC-2.2, AC-2.3 |
| T-4 | Go | same file `TestUpdate` (extended) | A graph edit dropping X's only link while X has a Citation → `ErrCitationsNeedGraph`, name change rolled back | D-5, D-6 |
| T-5 | Go | same file `TestUnlinkImpact` (rewritten) | Fixture: graphs Default, Jones and Solo. Default links P, U, W, X, Y; Jones links U, X; Solo links only P. U has a Citation with an Observation on a Subject of Default; W has a Citation with no Observation and no other graph; X has a Citation with no Observation; P and Y have no Citation. Pairs (Default,U), (Solo,P), (Default,W), (Jones,X), (Jones,Y) → three blocks, in that order: (Default,U) `ErrArtifactInUse` with U's Citation under `UnlinkVia`; (Solo,P) `ErrLastArtifact` with an empty group; (Default,W) `ErrCitationsNeedGraph` with W's Citation under `deleteimpact.ViaCitationsNeedGraph`. (Jones,X) is allowed and (Jones,Y) is not linked: no block. Then, for each of the five pairs, `Unlink` in a write that is rolled back returns the block's `Reason` or succeeds. A repeated pair → `ErrInvalid`; an unknown graph → `ErrInvalid` | AC-2.4, AC-2.5, AC-2.6, AC-2.7, D-6 – D-8 |
| T-6 | Go | same file `TestDelete` (extended) | A-3 probe 2: P1 (Citation C1) and P2 in Default; Jones with P2; `deleteimpact.Impact(KindEvidenceGraph, Default)` → gate `inbound`, a group via `ViaCitationsNeedGraph` listing C1; `Delete(Default)` → `ErrInUse`, Default and its links stay | AC-3.1, AC-3.2 |
| T-7 | Go | same file `TestDelete` (extended) | P1 without Citations → `Delete(Default)` succeeds, P1 links nothing; X with a Citation in Default and Jones → `Delete(Jones)` succeeds | AC-3.3, AC-3.4 |
| T-8 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifactUpdateGraphs` *(new)* | In one `writes.Run`: label change + `{Given, [Jones]}` → one revision with one `artifact` update, one link create, one link delete; label unchanged + `{Given, []}` on a Citation-free Artifact → only link deletes; label change + zero links → links unchanged; label change + a refused unlink → error, label unchanged after rollback | AC-1.5, AC-1.6, AC-1.7, AC-1.8, D-4 |
| T-9 | Go | `api/ffi/handlers/artifacts_test.go` `TestCreateArtifactGraphLinks` *(replaces `TestCreateArtifactGraphIDs`)* | Absent → Default and Jones, effects have `evidence_graphs`; `{graph_ids: [Jones]}` → Jones; `{}` → none and no `evidence_graphs` in `lists`; another Source's graph → `evidencegraphs.invalid` and no Artifact | AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| T-10 | Go | same file `TestUpdateArtifactGraphLinks` *(new)* | Label + `{graph_ids: [Jones]}` → one revision (`effects.revision` = previous + 1), links Jones only; no `graph_links` → links unchanged; a refused unlink → `evidencegraphs.citations_need_graph`, the label is unchanged in `GetSourceWorkspace` | AC-1.5, AC-1.6, AC-1.8 |
| T-11 | Go | `api/ffi/handlers/evidence_graphs_test.go` `TestGetEvidenceGraphUnlinkImpact` (rewritten) | Pairs → blocks with `graph_id`, `artifact_id`, `reason_code` and the group (via, kind, listed ids) for each of the three reasons; `last_artifact`'s `citations` unset; bad proto → error; malformed id → error | AC-2.4, AC-2.6, AC-2.7 |
| T-12 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `addArtifactSendsDefaultGraphLinksWhenBoxesHidden` *(new)* | One-graph Source (boxes hidden): Add artifact → `store.createArtifactGraphIDs == [nil]`; two-graph Source with only Jones checked → `[["g-j"]]`; a card label save → `store.updateArtifactGraphIDs == [nil]` | D-13 |
| T-16 | Swift | `macos/ProvenenciaTests/SourcePageModelTests.swift` `cardEditLocksInUseAndOnlyArtifactGraphs` (rewritten) | P3 linked to g-j and g-s, blocks scripted (g-s,P3) `artifact_in_use` with 2 Citations → one call whose pairs are (g-j,P3), (g-s,P3); `locked["g-s"]` = "2 citations in this graph use this artifact", `locked["g-j"] == nil`. P1, g-b's only Artifact, scripted (g-b,P1) `last_artifact` → `locked["g-b"]` = `artifactGraphOnlyArtifact`. A scripted `citations_need_graph` block → no lock. An Artifact linked to no graph → no call | D-13 |
| T-13 | Swift | `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `unlinkImpactReturnsScriptedBlocksPerPair` (replaces `unlinkImpactFiltersToAskedArtifacts`) | Scripted blocks (g1,a1) and (g2,a1); pairs (g2,a1), (g1,a1), (g2,a9) → the two blocks in pair order; `getEvidenceGraphUnlinkImpactPairs` equals the three pairs | D-14 |
| T-14 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` (heading table) | `("evidencegraphs.citations_need_graph", "citation", 1, "1 citation would be left in no evidence graph")`, and 3 → "3 citations would be left in no evidence graph" | AC-4.3 |
| T-15 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` (error table) | `"evidencegraphs.citations_need_graph"` → AC-4.1's text; `"evidencegraphs.in_use"` → AC-4.2's text | AC-4.1, AC-4.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (the card's locks and saves, `cardSaveRefusedKeepsEditAndRereadsLocks` included), `SourceGraphFormModelTests` (`unlinkImpactBlocksByGraph`), `deleteimpact` honesty tests, `sourcecopy` tests (zero-value links).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left — `CreateInput.GraphIDs`, `CreateArtifactRequest.graph_ids`, `GetEvidenceGraphUnlinkImpactRequest.artifact_ids`, `unlinkImpactBlocksByGraph`, `getEvidenceGraphUnlinkImpactGraphIDs`, `isOnlyArtifact`, "every graph of its Source by default"; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: source-layer-data-model.md, interpretation-layer-data-model.md, catalog-deletes.md (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The PR changes no view, so the check is that nothing a researcher sees moved, plus the one refusal a view can reach today.

1. Open a project with a Source that has graphs Default and Jones and two pages, Page 2 linked to both. Add artifact "Page 3" with only Jones checked. See it on the Source page; open Jones's composer and see Page 3 offered; Default's composer does not offer it.
2. On "Page 3", add a Citation in Jones's composer and save it with no Observation. On its card, uncheck Jones and Save. See the card's save error include "Citations on this artifact need it to stay in an evidence graph."; Page 3 stays in Jones.
3. On a Source where Page 1 has a Citation and is linked only to Default, and Jones links Page 2: delete Default from the Graphs section. See the confirm list Page 1's Citation under "1 citation would be left in no evidence graph" with Delete unavailable.
4. Edit a card's label only. See it save as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test builds an Artifact with Citations in one graph, then unlinks it, and now hits `citations_need_graph` (`evidencegraphstest.Split` links first, so it is safe) | Link the second graph before unlinking; do not weaken the rule |
| The card's existing per-graph unlink order (links before unlinks across separate RPCs) now meets the new refusal on a swap done in one Save: the unlink of the old only graph runs after the link, so it is allowed | Covered by D-5's order in the write; the card already links before it unlinks |
| An existing project holds a Citation already stranded by A-3 | Out (not planned); it stays reachable through its Artifact's delete preview |
| The `NOT EXISTS` probe scans on a graph with many links | It uses `evidence_graph_artifacts_artifact_id_idx`; `TestListQueryPlan` is the pattern if a plan test is wanted (not required) |

**Stop and report if:**

- `evidencegraphs.Unlink`, `Update`, `UnlinkImpact` or `LinkArtifact` no longer match §4 at your base.
- `deleteimpact` cannot hold `ViaCitationsNeedGraph` without importing `evidencegraphs` or `artifacts` (an import cycle).
- A `deleteimpact` honesty test (`TestResourceAndOwnedHonesty`, `TestPragmaHonestyProjectors`) refuses an inbound entry that is not one FK column.
- An existing Go or Swift test fails for a reason other than the signature changes, the reworded `in_use` copy, or a fixture order the rule now refuses (first risk row).
- Keeping a view's behavior (D-13) needs more than the call-site changes listed there.
- `swift-protobuf` does not expose presence (`hasGraphLinks`) for the `GraphLinks` message field.

## 12. Open questions

None. (Review M-6, retiring the card's Swift last-link rule here, answered yes by the user 2026-10-28: D-13.)

## 13. PR

**Title:** S12-22 — Artifacts can stand outside every graph
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-22-artifacts-outside-graphs.md`.
