# S14-10 — A Citation belongs to one Evidence graph

**Status:** Landed in #645 (2026-10-30)
**Spike / slice:** Spike 14 · Slice 4 — Evidence graphs finished (batch 4; moved from Spike 12, where it was S12-52, on 2026-10-29; S12-25 landed in #621)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the composer passes an id it already holds; the graph delete confirm gains one line through the existing DeleteImpact recipe)
**Board:** —
**Depends on:** S12-25 ([#621](https://github.com/mendahu/provenencia/pull/621), its `effects.Reads` and `prefer` / `citationArtifactGraphs` paths)
**Blocks:** —
**Branch:** `s14-10-citation-in-one-graph`
**Version bump:** none — a catalog format step and engine rules; nothing a researcher sees as a product version ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/catalog-audit.mdc`
**Context:** S12-25 ship log CON-1 (the gap this closes); working rules "Graphs are fully isolated", "A Citation belongs to one Evidence graph" (user, 2026-10-29), "High tolerance for churn"; [`catalog-deletes.md`](../../../catalog-deletes.md) policy 10; [`source-layer-data-model.md`](../../../source-layer-data-model.md) (graph links); [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) rule 25; precedent S12-18 (a required `GraphID` and its test-literal sweep)
**Planned:** 2026-10-29 against `s12-25-writes-name-graphs-they-move` @ `9775a92e` (S12-25's head; `main` @ `4145ccdd` plus #621)
**Review:** [`S14-10-citation-in-one-graph.review.md`](S14-10-citation-in-one-graph.review.md) — Fits with changes / Ready with changes → Fixed, 2 rounds
**Landed:** The migration is `000059.sql` (S13-25 took `000058.sql`). T-1 and T-2 are subtests of `TestCitations` ("Citations belong to their graph", "create needs a linked graph"). `WorkspaceSessionTests.citationReadsRefreshOnlyTheirGraph` relied on the fake's per-graph Artifact filter that D-10 drops, so `FakeStore` records `citationCountsGraphIDs` and the test asserts which graph was asked for. The map's Graph rules and Audit fixes bullets D-11 names were no longer in the Spike 14 map; the new working rule was already there, and the superseded rule survives only in `app-health-audit.md`, untouched. S12-25's Landed note is not in the tree, so its CON-1 gap was not edited.

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

A Citation stores the Evidence graph it was made in (`citations.graph_id`), and that graph never changes. A graph's file counts and the composer's citation stepper list that graph's Citations and no others, so a Citation made in the Jones composer no longer shows in Default before it has an Observation, and no write can move a Citation in or out of another graph. A graph delete takes its Citations with it.

**Why now.** S12-25 review CON-1: a Citation with no Observations shows in every graph its Artifact is linked to, so its first Observation (or the delete of its last) changes other graphs' citation counts and stepper, and S12-25's per-graph `reads` do not name those graphs. The user chose to remove the fallback instead of widening the reads (2026-10-29): "having citation be associated with the graph is smart because it makes all the code simpler, and citations are cheap. If two graphs need the same citation, they can just make it again."

**What stays.** A Citation's Artifact, locator, fields and notes; the composer offering only the graph's linked Artifacts; Copy Source copying no Citations; Promote, the auto-reconciler and search (none read Citation graph membership); the Swift `CatalogCitation` type (nothing reads a Citation's graph in the app).

## 2. User stories

**US-1.** As a researcher, when I make a Citation in one graph's composer, it shows only in that graph, before and after I record from it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Artifact A linked to Default and Jones; a Citation created in Jones with no Observations | `ListByArtifact(Jones, A)` lists it; `ListByArtifact(Default, A)` does not; `CountByEvidenceGraph(Default)` has no entry for A; `CountByEvidenceGraph(Jones)[A]` = 1 |
| AC-1.2 | Create a Citation on A with a graph A is not linked to, or with no graph id | `citations.invalid`; nothing written |
| AC-1.3 | A Jones Citation; an Observation on a Default Subject | `observations.other_graph` (create and update) |
| AC-1.4 | The AC-1.1 Citation gains its first Observation, on a Jones Subject | `Set.Reads` `citations` = Jones; the Default list and counts are the same before and after |
| AC-1.5 | `CreateCitedBridge` with no `citation_id`, endpoints in Jones | The new Citation's `graph_id` is Jones |

**US-2.** As a researcher, deleting a graph deletes its Citations, and unlinking an Artifact is refused while the graph has Citations on it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Jones has no Subjects and two Citations (one with a note) on A | `GetDeleteImpact(Jones)` is allowed and its `cascades` has one group, via `citations.graph_id`, kind `citation`, listing both refs; `Delete` removes both Citations and the note, each with an audited delete change |
| AC-2.2 | Jones has one Citation with no Observations on A; unlink (Jones, A) | `evidencegraphs.artifact_in_use`; `UnlinkImpact` lists that Citation |
| AC-2.3 | The graph delete confirm for AC-2.1 | It shows "The graph’s citations are deleted with it: CIT-A and CIT-B." (refs joined by `cascadeLines`' list format) |
| AC-2.4 | Any unlink or graph delete | `evidencegraphs.citations_need_graph` never appears; the code, its probe and its strings are gone |

**US-3.** As a contributor, an existing catalog opens with every Citation in one graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Before the step: Citation X with Observations on Jones Subjects; Citation Y with none on an Artifact linked to Default and Jones; Citation Z with none on an Artifact linked to no graph | After: X → Jones; Y → Default (first linked graph by name); Z → the Source's first graph by name, and Z's Artifact is linked to that graph. Manual check step 1 on the dogfood project that has two such Y Citations |

## 3. Scope

| In | Covered by |
| --- | --- |
| Migration: `citations.graph_id` NOT NULL, backfill, indexes (D-1) | STEP-1, AC-3.1 |
| `catalogmodel` FK; `deleteimpact` `ReleaseCitation`, graph facet release, probe removal (D-1, D-6) | STEP-1, STEP-2, AC-2.1 |
| `citations`: required `GraphID`, link check, reads by `graph_id`; `Delete` through `ReleaseCitation` (D-2, D-6) | STEP-2, AC-1.1, AC-1.2 |
| `observations.requireSameGraph` on the Citation's graph; `connect` bridge graph (D-3, D-4) | STEP-3, AC-1.3, AC-1.5 |
| `evidencegraphs` unlink rule; `citations_need_graph` removed end to end (D-5) | STEP-3, STEP-6, AC-2.2, AC-2.4 |
| Effects: Citation `citations` read by `graph_id`; `prefer` and `citationArtifactGraphs` deleted; link rows lose their `citations` read (D-7) | STEP-4, AC-1.4 |
| Go test-literal sweep (D-8) | STEP-5 |
| FFI `CreateCitationWithObservationsRequest.graph_id` (D-9) | STEP-6 |
| Swift: store signature, composer passes its graph, unlink reason case and strings removed, cascade line (D-10) | STEP-7, AC-2.3 |
| Docs (D-11) | STEP-8 |

| Out | Home |
| --- | --- |
| `graph_id` on the `Citation` proto message or Swift `CatalogCitation` | not planned (no reader) |
| Moving a Citation to another graph | not planned (working rule: a Citation's graph never changes) |
| Plural forms for the new cascade line | S12-29 (plural forms and a catalog check) |
| Copy Source copying Citations | not planned |
| The graph list `reads` entry | S12-26 |

## 4. Current state

Verified at `9775a92e`.

| Area | Where | Today |
| --- | --- | --- |
| Schema | `core/database/migrations/000030.sql` (rebuild), `000031.sql` (`citations_artifact_id_idx`), `000050.sql` (`description` → `label`); last step `000057.sql` | `citations (id, ref, artifact_id, locator_json, transcription, label, transcription_uncertain, transcription_note)`; no graph column |
| Model | `core/database/catalogmodel/model.go` `Tables` / `FKs` | `citations` resource; `citations.artifact_id` resource NO ACTION; `citation_notes.citation_id` audited CASCADE |
| Membership SQL | `core/database/citations/citations.go` `sqlUsedByGraph`, `sqlInGraph`, `sqlListByArtifact`, `sqlCountByEvidenceGraph`, `ListByArtifact`, `CountByEvidenceGraph` | In a graph = an Observation on one of its Subjects, **or no Observation at all**; counts join `evidence_graph_artifacts` |
| Create | same file `CreateInput`, `normalizeCreateInput`, `InsertWithObservationsTx`, `sqlInsert`, `sqlGet`, `scanOne` | No graph; checks only that the Artifact exists; create change has no `graph_id` |
| Delete | same file `Delete` | `Impact` → `Refuse` → `ReleaseFacets(KindCitation)` → `DELETE` → its own delete change (no `graph_id`) |
| Graph SQL | `core/database/evidencegraphs/evidencegraphs.go` `SQLObservationsInGraph`, `sqlCitationsUsing`, `sqlCitationsNeedGraph`, `CitationsUsingArtifactTx`, `checkUnlink`, `UnlinkBlock`, `ErrCitationsNeedGraph`, `ErrInUse` doc | Unlink refused when the graph's Citations use the Artifact (through Observations), when it is the last link, or when the Artifact's Citations would be left in no graph |
| Projector | `core/database/deleteimpact/projectors.go` `projectCitation`, `sqlFirstArtifactGraph` | A listed Citation's graph = its first Observation's Subject's graph, else the Artifact's first linked graph by name |
| Delete registry | `core/database/deleteimpact/register.go` `ViaCitationsNeedGraph`, `SQLCitationsOnlyInGraph`, the `joinEdge(KindEvidenceGraph, ViaCitationsNeedGraph, …)` probe; `facets.go` `buildFacetReleases`, `releaseClaim` / `ReleaseClaim` (the composing precedent) | Graph delete refused while it is the only graph of an Artifact with Citations |
| Isolation | `core/database/observations/observations.go` `requireSameGraph`, `sqlCitationArtifact`; `update.go` caller | Subject's graph must be **linked to the Citation's Artifact** |
| Bridge | `core/database/connect/connect.go` `CreateCitedBridge` | New Citation from `in.Citation` with no graph; bridge Subject takes `from.GraphID` |
| Effects | `core/database/effects/registry.go` `citationObservationGraphs`, `citationArtifactGraphs`, `"citations"` (`Reads: prefer(citationObservationGraphs, citationArtifactGraphs)`), `"evidence_graph_artifacts"` (`Reads: field("graph_id")`); `path.go` `prefer` | S12-25 shape; `prefer`'s only caller is the Citation row |
| Effects tests | `reads_resolve_test.go` `TestCitationReadsPreferArtifactGraphs`; `reads_test.go` `TestReadsDeclare`; `evidencegraphs_test.go` `TestLinkWritesMoveCitationReads` | Pin the fallback and the link-row read |
| Error code | `core/apperr/apperr.go` `CodeEvidenceGraphsCitationsNeedGraph` | Used by `evidencegraphs`, `deleteimpact` |
| FFI | `api/proto/engine.proto` `CreateCitationWithObservationsRequest` (fields 1–10); `api/ffi/handlers/citations.go` `CreateCitationWithObservations` | No graph id |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `createCitationWithObservations`; `GoStore.swift`; `FakeStore.swift` (its own method, its bridge call, `citationCountsByEvidenceGraph` filtering by `graph.artifactIDs`); `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | No graph id |
| Composer | `macos/App/Features/CitationComposer/CitationComposerContext.swift` `createCitationIfNeeded` | Holds `graphID`; does not pass it |
| Unlink reason | `macos/App/Platform/ConclusionVocabulary.swift` `EvidenceGraphUnlinkReason.citationsNeedGraph` | Decoded from the deleted code; used only by tests (`SourceGraphFormModelTests`, `SourcePageModelTests`, `EvidenceGraphFixtures.swift`) and pinned by `api/ffi/swift_vocabulary_test.go` (unlink reasons row) |
| Strings | `L10n.swift` `L10n.Errors.evidenceGraphsCitationsNeedGraph` (`error.evidencegraphs.citations_need_graph`), `L10n.DeleteImpact.viaFormats["evidencegraphs.citations_need_graph"]` (`deleteImpact.via.evidenceGraphsCitationsNeedGraph`); `Localizable.xcstrings` | Live only for the deleted code |
| Cascade copy | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `cascadeLines` | Unknown vias fall back to `alsoAffects` ("This also changes …"), which reads wrong for a delete |
| Fixtures | `core/database/evidencegraphs/evidencegraphstest` `SoleGraph`, `Split`, `Create`; `core/database/graphcache/graphcachetest/graphcachetest.go` (raw `INSERT INTO citations`); about 100 `citations.CreateInput{` literals in `core`, 32 `CreateCitationWithObservationsRequest{` in `api` | No graph id |
| Old-behavior tests | `core/database/artifacts/artifacts_test.go` (~1223) and `api/ffi/handlers/artifacts_test.go` (~329) expect `citations_need_graph`; `api/ffi/handlers/citations_test.go` (~369) expects a bare Citation's read to name every linked graph; `core/database/evidencegraphs/evidencegraphs_test.go` (~614) says a bare Citation belongs to no graph; `core/database/deleteimpact/impact_test.go` (~1111) projects a bare Citation to the first linked graph; `macos/ProvenenciaTests/L10nResolutionTests.swift` (~87–90) and `PVDeleteImpactTests.swift` (~155–156) pin the deleted strings | Pin the old rules |
| Docs | `docs/catalog-deletes.md` policy 10; `docs/interpretation-layer-data-model.md` rule 25; `docs/source-layer-data-model.md` lines 420–425; `core/database/deleteimpact/README.md`; `.agents/skills/add-catalog-delete/SKILL.md` § Domain `Delete`; map **Audit fixes (slices 5–7)** rule "An Artifact with Citations keeps at least one graph"; `core/database/effects/README.md`; map **Graph rules** | Teach the no-Observation fallback and `citations_need_graph` |

**Precedent.** S12-18 (plan (Spike 12, in git history)) made `subjects.CreateInput.GraphID` required and moved every test literal onto `evidencegraphstest.SoleGraph` / `Split`: copy that sweep. `000030.sql` is the table-rebuild shape (`-- migrate: foreign_keys off`, `__new`, copy, drop, rename) and `000031.sql` recreates the dropped index. `releaseClaim` / `ReleaseClaim` in `core/database/deleteimpact/facets.go` is the one-row removal that a parent release and the row's own `Delete` share: copy it for `releaseCitation` / `ReleaseCitation`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Migration `000058.sql` *(new)* rebuilds `citations` with `graph_id BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE` after `artifact_id`, under `-- migrate: foreign_keys off`. Backfill per row, by scalar subquery in this order: (1) the graph of the Subject of the Citation's earliest Observation (`ORDER BY o.id`); (2) the first graph linked to its Artifact, `ORDER BY g.name COLLATE NOCASE, g.ref, g.id` (the graph list's order, `evidencegraphs.sqlListOrder`); (3) the first graph of the Artifact's Source, same order. After the rebuild, `INSERT OR IGNORE INTO evidence_graph_artifacts (graph_id, artifact_id)` every (graph, Artifact) pair a Citation now holds that is not linked, so every stored Citation's graph is linked to its Artifact (unaudited, as S12-03's backfill links were). No split. Recreate `citations_artifact_id_idx`; add `citations_graph_artifact_idx ON citations(graph_id, artifact_id)` *(new)*. `catalogmodel`: `{From: "citations", Column: "graph_id", To: "evidence_graphs", OnDelete: "CASCADE", Bucket: BucketFacet, Audited: true}` | User: a Citation is part of its graph and goes with it. The user's three dogfood catalogs have **no** Citation with Observations in two graphs (query run 2026-10-29), so (1) never meets a split; one has two Citations of shape (2). (3) keeps NOT NULL from failing on an Artifact with no link | Splitting multi-graph Citations into copies (no data needs it); refusing to open (no data hits it); a NULL-able column filled at open (`add-catalog-migration` prefers SQL backfill) |
| D-2 | `citations.CreateInput.GraphID []byte` *(new)*, required: `normalizeCreateInput` returns `ErrInvalid` when it is not 16 bytes, beside the `ArtifactID` check. `InsertWithObservationsTx` checks `evidencegraphs.IsLinkedTx(tx, GraphID, ArtifactID)`; not linked → `ErrInvalid`. Insert writes `graph_id`; the create change's `Fields` include `"graph_id"`. `Citation.GraphID` *(new)* read by `sqlGet` / `scanOne`. `sqlListByArtifact` filters `c.artifact_id = ? AND c.graph_id = ?` and its count column becomes `(SELECT COUNT(*) FROM observations o WHERE o.citation_id = c.id)`; `sqlCountByEvidenceGraph` becomes `SELECT artifact_id, COUNT(*) FROM citations WHERE graph_id = ? GROUP BY artifact_id`. Delete `sqlUsedByGraph`, `sqlInGraph`; reword the `sqlPersonNamesByArtifact` comment (it names `SQLObservationsInGraph`). Bind order of each changed query follows its new text | One membership rule, stored. The link join is redundant: D-5 keeps every Citation's graph linked to its Artifact | A user-kind code for "not linked" (the composer offers only linked Artifacts; it is bad input); defaulting `GraphID` from the Artifact's sole graph (a hidden second path) |
| D-3 | `observations.requireSameGraph` compares the Subject's graph with the Citation's `graph_id` (`sqlCitationGraph = SELECT graph_id FROM citations WHERE id = ?` *(new)*) and returns `ErrOtherGraph` when they differ; the Artifact-link check goes. Delete `sqlCitationArtifact` (its one caller is `requireSameGraph`). The Subject-value rules below it are unchanged | User: refuse an Observation about a person in another graph (2026-10-29). Same check, same code | A separate check beside `requireSameGraph` |
| D-4 | `connect.CreateCitedBridge` sets `in.Citation.GraphID = from.GraphID` before `NormalizeCreateInput` when it creates the Citation. A reused `CitationID` is covered by D-3 through `InsertManyTx`. `CreateCitedBridgeRequest` is unchanged | The bridge Subject already takes `from.GraphID` | A `graph_id` on the bridge request (the engine knows it) |
| D-5 | `evidencegraphs.sqlCitationsUsing` becomes `SELECT c.id, c.ref FROM citations c WHERE c.artifact_id = ? AND c.graph_id = ? ORDER BY c.ref COLLATE NOCASE`: any Citation of the graph on the Artifact, with or without Observations, refuses the unlink with `ErrArtifactInUse`. `checkUnlink` loses its third reason. Delete `ErrCitationsNeedGraph`, `sqlCitationsNeedGraph`, `SQLObservationsInGraph`, `apperr.CodeEvidenceGraphsCitationsNeedGraph`, `deleteimpact.ViaCitationsNeedGraph`, `deleteimpact.SQLCitationsOnlyInGraph` and the graph probe built on them. `UnlinkBlock.Reason`, `ErrInUse` and `CitationsUsingArtifactTx` docs drop the third reason | Every Citation's graph stays linked to its Artifact, so an Artifact with Citations is always linked to every graph they are in: "keeps at least one graph" holds by construction | Keeping `citations_need_graph` as a second, now unreachable, refusal |
| D-6 | `deleteimpact.releaseCitation(tx *sql.Tx, id)` / `ReleaseCitation(tx *database.Tx, id)` *(new, in `facets.go`)*: the one Citation removal. It reads every column of the row, releases its facets (`releaseFacets(tx, KindCitation, id)`), deletes the row, and appends the delete change `rowchange.DeletedRow` over every column (`graph_id` included; a NULL column as `nil`), the `releaseClaim` shape. A facet release *(new)* `{Parent: KindEvidenceGraph, Via: "citations.graph_id", Named: true, Child: KindCitation, Count: countSQLFn("SELECT COUNT(*) FROM citations WHERE graph_id = ?"), List: listSQLFn("SELECT id, ref FROM citations WHERE graph_id = ? ORDER BY ref COLLATE NOCASE LIMIT ?"), Remaining: "SELECT COUNT(*) FROM citations WHERE graph_id = ?", Release: releaseEach(…)}`. Add `releaseEach(selectIDs string, one func(*sql.Tx, []byte) (released, error))` *(new, `facets.go`)*: select ids by the parent id, close the rows, release each through `one`, merge. `releaseSubjectClaims` moves onto it and the graph release is `releaseEach("SELECT id FROM citations WHERE graph_id = ? ORDER BY id", releaseCitation)`. `citations.Delete` keeps `Impact` / `Refuse` and then calls `ReleaseCitation` instead of its own release, `DELETE` and change. `projectors.go` `projectCitation` reads `c.graph_id` in its `SELECT … FROM citations c` for the location's graph and keeps the Observation query only for the Subject ref and title; delete `sqlFirstArtifactGraph` | User: a graph delete takes its Citations (2026-10-29). One removal shape for both paths, as `ReleaseClaim` is for claims. A graph with Subjects is still refused (`subjects.graph_id`), so a released Citation has no Observation (D-3) | A second collect-then-release loop beside `releaseSubjectClaims`; `rowFacet` (it does not release `citation_notes` first); deleting through `citations.Delete` from `deleteimpact` (import cycle) |
| D-7 | Effects: `"citations"` `Reads: []ReadPath{{ReadCitations, field("graph_id")}}`; `EvidenceGraph` stays `citationObservationGraphs`. Delete `citationArtifactGraphs`, `prefer` (`path.go`) and its test cases. `"evidence_graph_artifacts"` drops `Reads` (a link write moves no Citation). `"observations"` is unchanged: its Citation's graph is its Subject's graph | The graph is stored, so the read is the row's own field. CON-1 closes because no write changes a Citation's membership | Keeping `prefer` for a fallback that no longer exists |
| D-8 | Every `citations.CreateInput{` literal and raw `INSERT INTO citations` in Go tests and test helpers gets a graph id: the graph the test's Subjects or Artifact links already use, through `evidencegraphstest.SoleGraph(t, c, sourceID).ID` or the graph from `evidencegraphstest.Split` / `Create`. Shared helpers change first (STEP-2): `observationstest.SaveCitation(c, userID, graphID, art, inputs...)` gains `graphID []byte` after `userID` and keeps `art artifacts.Artifact` (12 callers); `graphcachetest`'s raw insert writes `graph_id`. A test whose Artifact is linked to no graph links it to the graph it uses through `evidencegraphstest.Create` / the fixture, unless the test is about the unlinked case, which then expects `citations.invalid`. No helper default | S12-18 precedent; explicit graphs keep the tests honest | A test-only default that picks the sole graph |
| D-9 | Proto: `string graph_id = 11;` on `CreateCitationWithObservationsRequest`; the handler parses it with `parseID` (bad or empty → the same error path as `artifact_id`) and passes `GraphID`. Regenerate with `./scripts/generate-proto.sh`. The 32 handler-test requests set `GraphId` | The composer holds the graph | Deriving it from the Artifact on the engine side |
| D-10 | Swift: `createCitationWithObservations(projectDir:userID:graphID:artifactID:…)` on `GenealogyStore`, `GoStore`, `FakeStore` (`graphID` *(new)* after `userID`); `CitationComposerContext.createCitationIfNeeded` passes `graphID`; FakeStore's bridge call passes the bridge's graph. FakeStore records `"createCitationWithObservations graph=\(graphID) observations=\(drafts.count)"`; the two exact-match asserts in `CitationComposerModelTests` (today lines 621 and 650) expect the new string. FakeStore `citationCountsByEvidenceGraph` drops its `graph.artifactIDs` filter and counts every seeded Citation per Artifact, as `listCitationsByArtifact` returns every seeded Citation (no membership rule in the fake, `swift-tests.mdc`). Delete `EvidenceGraphUnlinkReason.citationsNeedGraph` and its two wire lines; tests that used it use `.artifactInUse`. Delete `L10n.Errors.evidenceGraphsCitationsNeedGraph`, its `case` in the error switch, the `viaFormats` entry, and both keys in `Localizable.xcstrings`. Add `L10n.DeleteImpact.deletesCitations(refs:)` *(new)* = "The graph’s citations are deleted with it: %@." and a `case "citations.graph_id":` in `PVDeleteImpact.cascadeLines` | Closed vocabulary; the fallback line "This also changes" misstates a delete | Keeping the dead case for old engines (one engine ships with the app) |
| D-11 | Docs in this PR: `catalog-deletes.md` policy 10 (graph delete releases its Citations, audited; no `citations_need_graph`; add `citations.graph_id` to the Audited column); `interpretation-layer-data-model.md` rule 25 ("A Citation belongs to the one Evidence graph it was made in; its graph stays linked to its Artifact"); `source-layer-data-model.md` lines 420–425 (isolation, unlink, graph reads, graph delete); `core/database/deleteimpact/README.md` and the `evidencegraphs.Delete` doc comment (graph delete releases its Citations); `core/database/effects/README.md` (Citation read by `graph_id`; no fallback); the map's **Graph rules** bullets on `citations` reads, unlinking and the composer, the new working rule under **Graph data model**, and the **Audit fixes (slices 5–7)** rule "An Artifact with Citations keeps at least one graph" marked "Superseded 2026-10-29 by 'A Citation belongs to one Evidence graph' (S14-10)". `.agents/skills/add-catalog-delete/SKILL.md`: one bullet — a row with its own `Delete` that is also an audited facet of another parent gets one `releaseX` / `ReleaseX` in `facets.go`, called by the parent's release and by its `Delete` (`ReleaseClaim`, `ReleaseCitation`) — and the § Domain `Delete` example shows `ReleaseCitation` | Map rule "Each PR updates the docs and skills it makes false" | — |

**Rules this plan bends.** "Artifact ↔ graph links" (map **Graph rules**): "From S12-22 also when it is the last graph of an Artifact that has Citations" → subsumed by D-5; and "The Citation composer offers … on each only the graph's Citations: those with an Observation on one of its Subjects, plus those with none yet" → the graph's own Citations. User answer: 2026-10-29 ("having citation be associated with the graph is smart"). Also the **Audit fixes (slices 5–7)** rule "An Artifact with Citations keeps at least one graph" (a Citation with no Observation reached through any linked graph; the graph-delete probe; `in_use` covering these Citations): D-5, D-6 and **Strings** replace each sentence, under the same user answer. STEP-8 updates both bullets and marks that rule superseded.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `000058.sql` *(new)* | migration | `000030.sql` rebuild | Extends the same shape |
| `citations.graph_id`, `citations_graph_artifact_idx` *(new)* | column / index | `subjects.graph_id` | Same idea for Citations; no `source_id` on `citations`, so no composite FK |
| `CreateInput.GraphID`, `Citation.GraphID` *(new)* | field | `subjects.CreateInput.GraphID` | — |
| `sqlCitationGraph` *(new)* | SQL | `sqlCitationArtifact` | Replaces it in `requireSameGraph` |
| `releaseCitation` / `ReleaseCitation` *(new)* | func | `releaseClaim` / `ReleaseClaim` | Same shape for a different row |
| `releaseEach` *(new)* | func | `releaseSubjectClaims` | Its loop becomes the shared helper; the graph release is one call |
| `CreateCitationWithObservationsRequest.graph_id` *(new)* | proto field | `CreateSubjectRequest.graph_id` | — |
| `L10n.DeleteImpact.deletesCitations(refs:)` *(new)* | string | `leavesHandle`, `alsoAffects` | `alsoAffects` says "changes" |

**Builds on.** `deleteimpact` facet releases (one call per delete, releases compose); `evidencegraphs.IsLinkedTx`; S12-25 `effects.Reads`; `evidencegraphstest` fixtures.
**Deletes.** `sqlUsedByGraph`, `sqlInGraph`, `SQLObservationsInGraph`, `sqlCitationsNeedGraph`, `sqlCitationArtifact`, `ErrCitationsNeedGraph`, `CodeEvidenceGraphsCitationsNeedGraph`, `ViaCitationsNeedGraph`, `SQLCitationsOnlyInGraph` and its probe, `sqlFirstArtifactGraph`, `releaseSubjectClaims`' own loop (onto `releaseEach`), `prefer`, `citationArtifactGraphs`, the link row's `citations` read, `EvidenceGraphUnlinkReason.citationsNeedGraph`, two L10n keys; the CON-1 known gap in S12-25's Landed note becomes history.
**Next adds.** None named.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A graph delete that removes Citations | Composer and graph surfaces of other graphs | Those graphs' keys are untouched (different `graph_id`); the deleted graph's own keys go with the graph (S12-21 deleted-graph fallback) |
| An Artifact unlink now refused by an empty Citation | Source page Artifact card and graph form (S12-38 preflight) | Same `artifact_in_use` lock and listed Citations as today; no new UI state |

### Scale

At the map's Scale (60 graphs, 300 Citations on one Source): `ListByArtifact` and `CountByEvidenceGraph` read one graph's rows through `citations_graph_artifact_idx` instead of an `EXISTS` per Citation; a Citation write names one `citations` read id; a graph delete releases that graph's Citations one by one (each a primary-key delete plus its notes), bounded by one graph's Citations.

### Schema

```sql
-- migrate: foreign_keys off
CREATE TABLE citations__new (
	id BLOB PRIMARY KEY, ref TEXT UNIQUE NOT NULL,
	artifact_id BLOB NOT NULL REFERENCES artifacts(id),
	graph_id BLOB NOT NULL REFERENCES evidence_graphs(id) ON DELETE CASCADE,
	locator_json TEXT NOT NULL, transcription TEXT, label TEXT,
	transcription_uncertain INTEGER NOT NULL DEFAULT 0, transcription_note TEXT,
	CHECK (transcription_uncertain IN (0, 1))
) STRICT;
-- INSERT INTO citations__new SELECT every column of c, with graph_id =
--   COALESCE(<D-1 step 1>, <D-1 step 2>, <D-1 step 3>) FROM citations c;
DROP TABLE citations; ALTER TABLE citations__new RENAME TO citations;
CREATE INDEX citations_artifact_id_idx ON citations(artifact_id);
CREATE INDEX citations_graph_artifact_idx ON citations(graph_id, artifact_id);
```

Copy the column list and types from the live schema (`sqlite3 … .schema citations` on a fresh catalog), not from this sketch. Delete class: audited facet of its graph (D-6). Audit: create and delete changes carry `graph_id`; scope stays `source` through `artifact_id`.

### Go core

| Package | Change |
| --- | --- |
| `core/database/citations` | D-2, D-6 |
| `core/database/observations` | D-3 |
| `core/database/connect` | D-4 |
| `core/database/evidencegraphs` | D-5 |
| `core/database/deleteimpact` | D-5, D-6 |
| `core/database/effects` | D-7 |
| `core/apperr` | delete `CodeEvidenceGraphsCitationsNeedGraph` |

Errors: `citations.invalid` → missing or unlinked graph (AC-1.2); `observations.other_graph` → Subject in another graph (AC-1.3); `evidencegraphs.artifact_in_use` → unlink with any Citation of the graph on the Artifact (AC-2.2).

### FFI

D-9. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Create call | — | `CitationComposerContext.createCitationIfNeeded` | passes `graphID` |
| Cascade line | `PVDeleteImpact` recipe | `PVDeleteImpact.cascadeLines` | `citations.graph_id` case |
| Unlink reason | — | `ConclusionVocabulary.swift` | case removed |

States: unchanged. Accessibility: unchanged.

### Strings

New: `deleteImpact.confirm.deletesCitations` = "The graph’s citations are deleted with it: %@." (argument: the Citation refs, joined). Removed: `error.evidencegraphs.citations_need_graph`, `deleteImpact.via.evidenceGraphsCitationsNeedGraph`. Old-scope copy checked and kept: `error.evidencegraphs.in_use` ("Subjects or citations still need this evidence graph.") stays true only for Subjects — reword to "Subjects still need this evidence graph." `error.evidencegraphs.artifact_in_use`, `sources.graphForm.keptLinkedMessage` and `sources.page.artifact.graphs.none` stay.

## 7. Steps

**STEP-1 — Schema**
Files: `core/database/migrations/000058.sql` *(new)*, `core/database/catalogmodel/model.go`. Skill: `add-catalog-migration`, `add-catalog-model`.
Do: D-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/catalogmodel/...` passes or fails only on the missing facet release (STEP-2 adds it).

**STEP-2 — Citations and their removal**
Files: `core/database/citations/citations.go`, `core/database/deleteimpact/facets.go`, `register.go`, their tests. Skill: `add-catalog-delete`, `add-catalog-query`.
Files also: `core/database/deleteimpact/projectors.go`, `core/database/observations/observationstest/observationstest.go`, `core/database/graphcache/graphcachetest/graphcachetest.go`.
Do: D-2, D-6 and D-8's shared helpers (the D-5 deletions wait for STEP-3, since `evidencegraphs` still builds on them). Sweep these packages' own tests: `citations`, `deleteimpact` (the `impact_test.go` bare-Citation row expects its own `graph_id`, set up so it differs from the first linked graph), `catalogmodel`.
Check: `CGO_ENABLED=1 go build ./...`; `CGO_ENABLED=1 go test -tags fts5 ./core/database/citations/... ./core/database/deleteimpact/... ./core/database/catalogmodel/...` pass.

**STEP-3 — Isolation, bridge, unlink**
Files: `core/database/observations/observations.go`, `update.go`, `core/database/connect/connect.go`, `core/database/evidencegraphs/evidencegraphs.go`, `core/apperr/apperr.go`, their tests.
Files also: `core/database/artifacts/artifacts_test.go`.
Files also: `core/database/deleteimpact/register.go`.
Do: D-3, D-4, and all of D-5 (`evidencegraphs`, `deleteimpact` and `apperr` deletions together). Sweep these packages' own tests. `artifacts_test.go`'s last-graph unlink case now expects `evidencegraphs.ErrArtifactInUse` (its Citations are in that graph); the `evidencegraphs_test.go` bare-Citation comment and case say the Citation is in its own graph.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/... ./core/database/connect/... ./core/database/evidencegraphs/... ./core/database/artifacts/... ./core/database/deleteimpact/...` pass; `CGO_ENABLED=1 go build ./...` succeeds.

**STEP-4 — Effects**
Files: `core/database/effects/registry.go`, `path.go`, `reads_resolve_test.go`, `reads_test.go`, `path_test.go` (if it covers `prefer`), `core/database/evidencegraphs/evidencegraphs_test.go` (`TestLinkWritesMoveCitationReads` → `TestLinkWritesMoveNoCitationReads`). Skill: `add-audit-scope`.
Do: D-7. Rename `TestCitationReadsPreferArtifactGraphs` to `TestCitationReadsNameItsGraph`, remove `evidence_graph_artifacts` from `TestReadsDeclare`'s `want[ReadCitations]`, and add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/database/evidencegraphs/...`.

**STEP-5 — Go test sweep**
Files: every Go test and test helper the build names. Skill: go-tests.
Do: D-8 for every package STEP-2 to STEP-4 did not sweep. A missing `GraphID` compiles and fails at run time with `citations.invalid`, so the test run is the check.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes.

**STEP-6 — FFI**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/citations.go`, handler tests. Skill: `add-ffi-handler`.
Do: D-9 and the 32 request literals. Rows that pinned `citations_need_graph` now pin `evidencegraphs.artifact_in_use` (`evidence_graphs_test.go`, `artifacts_test.go`) or the graph cascade; `citations_test.go`'s zero-Observation case expects `reads` `citations` = the request's graph and empty `evidence_graphs`; `swift_vocabulary_test.go` drops `CodeEvidenceGraphsCitationsNeedGraph` from the unlink-reasons row.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./api/...`.

**STEP-7 — Swift**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CitationComposerContext.swift`, `ConclusionVocabulary.swift`, `L10n.swift`, `Localizable.xcstrings`, `PVDeleteImpact.swift`, tests named in §8, `OnboardingModelTests.swift` (`ThrowingStore`), `L10nResolutionTests.swift` (drop the `citations_need_graph` row; the `in_use` row expects the new copy), `PVDeleteImpactTests.swift` (drop the `citations_need_graph` rows), `SourceGraphFormModelTests.swift`, `SourcePageModelTests.swift`, `EvidenceGraphFixtures.swift` (`.citationsNeedGraph` → `.artifactInUse`), `CitationComposerModelTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-10 and the `in_use` reword in **Strings**.
Check: `scripts/check-pr.sh` PASS.

**STEP-8 — Docs**
Files: D-11's docs and skill, this plan (status), `deployment-plan.md` (map row), `completed.md`, the S12-25 plan's Landed note (CON-1 gap → "closed by S14-10").
Do: D-11 and the DoD.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-10-citation-in-one-graph.md`; the DoD sweep.

Commits: one per behavior ("Store the Evidence graph a Citation belongs to.", "Delete a graph's Citations with it.", "Refuse an Observation outside its Citation's graph.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/citations/citations_test.go` `TestCitationsBelongToTheirGraph` *(new)* | AC-1.1 setup → lists and counts per graph | AC-1.1, D-2 |
| T-2 | Go | same file `TestCreateNeedsALinkedGraph` *(new)* | no graph id; graph not linked to A → `citations.invalid` | AC-1.2 |
| T-3 | Go | `core/database/observations/observations_test.go` (update the other-graph table) | Jones Citation + Default Subject, create and update → `observations.other_graph` | AC-1.3, D-3 |
| T-4 | Go | `core/database/connect/connect_test.go` (add a case) | new bridge Citation's `GraphID` = endpoints' graph | AC-1.5, D-4 |
| T-5 | Go | `core/database/effects/reads_resolve_test.go` `TestFirstObservationMovesNoOtherGraph` *(new)* | Citation in Jones, Artifact linked to Default and Jones; add first Observation → `Reads[citations]` = {Jones} | AC-1.4, D-7 |
| T-6 | Go | `core/database/deleteimpact/impact_test.go` (graph cases) and `core/database/evidencegraphs/evidencegraphs_test.go` delete test | AC-2.1 → allowed; one cascade group via `citations.graph_id`; Delete removes Citations and notes; audit changes carry `graph_id` | AC-2.1, D-6 |
| T-7 | Go | `evidencegraphs_test.go` `TestUnlink`, `TestUnlinkImpact` (update) | empty Jones Citation on A, A also linked to Default (sorts first) → `artifact_in_use` listing it, its `Location.GraphID` = Jones; no `citations_need_graph` row | AC-2.2, AC-2.4, D-5 |
| T-8 | Go | `api/ffi/handlers/citations_test.go` `TestCreateCitationWithObservations` (update) | request with `GraphId` → Citation in that graph; missing → error | D-9 |
| T-9 | Swift | `macos/ProvenenciaTests/PVDeleteImpactTests.swift` (or the file that tests `cascadeLines`) | via `citations.graph_id`, refs CIT-A, CIT-B → "The graph’s citations are deleted with it: CIT-A and CIT-B." | AC-2.3, D-10 |
| T-10 | Swift | `CitationComposerModelTests` (the create tests at today's 621 / 650) | `recordedCalls` contains `"createCitationWithObservations graph=<composer graph> observations=0"` | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/deleteimpact`, `core/database/evidencegraphs`, `core/database/connect`, `api/ffi/handlers`, `SourceGraphFormModelTests`, `SourcePageModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `citations_need_graph`, `CitationsNeedGraph`, `citationsNeedGraph`, `SQLCitationsOnlyInGraph`, `SQLObservationsInGraph`, `sqlInGraph`, `citationArtifactGraphs`, `sqlFirstArtifactGraph`, `prefer(`, and the phrases "no Observation yet" / "belongs to no graph" rg'd across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core`, `api`, `macos/App`, `macos/ProvenenciaTests` and `L10n.swift` with no hit left (the superseded map rule keeps its text under its marker); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: catalog-deletes.md, interpretation-layer-data-model.md, source-layer-data-model.md, core/database/effects/README.md, core/database/deleteimpact/README.md; add-catalog-delete skill
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-3.1) Quit the app. Back up the dogfood project whose query reported two empty Citations on a page linked to two graphs. Open it with this build: it opens, and each of those two Citations shows only in the first of its page's graphs by name (the file dropdown's count and the composer stepper).
2. On a Source with graphs A and B both linked to one page, make a Citation in A's composer and save it with no Observation. B's file dropdown count for that page does not change; A's goes up by one.
3. Record an Observation from that Citation about a person in A. B's count still does not change.
4. Try to uncheck graph A on that page's card: the box locks, listing the Citation.
5. Delete a graph that has a Citation but no Subjects: the confirm says "The graph’s citations are deleted with it: CIT-…"; after Delete, the Citation is gone.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A catalog has a Citation with Observations in two graphs | D-1 puts it in the earliest Observation's graph; none exist in the dogfood data. Later Observation writes on it are refused by D-3 until the researcher remakes it, and deleting its graph would fail on `observations.citation_id` with a raw FK error (accepted: no data has the shape) |
| The facet FK on a resource table fails a `catalogmodel` / `deleteimpact` honesty test | Stop (below) |

**Stop and report if:**

- `TestPragmaHonesty` or a `deleteimpact` honesty test refuses an audited `BucketFacet` FK on the `citations` resource table, or requires a `Kind` change for `citations`.
- `releaseFacets` cannot run `releaseCitation` from a facet release without an import cycle.
- A Go caller outside tests creates a Citation where no graph is known (other than `connect`, D-4).
- The migration fails NOT NULL on the dogfood project in manual check step 1.

## 12. Open questions

None.

## 13. PR

**Title:** S14-10 — A Citation belongs to one Evidence graph
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-10-citation-in-one-graph.md`.
