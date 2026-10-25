# S12-10 — Composer offers linked Artifacts

**Status:** Landed in [#540](https://github.com/mendahu/provenencia/pull/540) (2026-10-24)
**Spike / slice:** Spike 12 · Slice 3 — Graph id end to end
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the composer's Artifact picker, Citation stepper and empty callout keep their look; only what they list and two strings change)
**Board:** —
**Depends on:** S12-05 (landed [#511](https://github.com/mendahu/provenencia/pull/511)), S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528))
**Blocks:** —
**Branch:** `s12-10-composer-linked-artifacts`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-09 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [data model § rules](../source-structure.md#rules), [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 7.1 (the graph rules list), the S12-08 [plan](S12-08-graph-id-end-to-end.md) D-1 (how a Source-keyed RPC is re-keyed) and D-8 (keys); working rules: "The Citation composer offers only the graph's linked Artifacts", "One graph list", "Graphs are fully isolated", "A re-keyed symbol is renamed in the PR that re-keys it", "High tolerance for churn", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-24 against `main` @ `7a1398c`
**Review:** [`S12-10-composer-linked-artifacts.review.md`](S12-10-composer-linked-artifacts.review.md) — Fits with changes / Ready with changes, 2 rounds; every finding answered

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

In a census with a Jones graph and a Smith graph, the Citation composer opened on the Jones graph offers only the Artifacts linked to Jones, and on any of those Artifacts it steps only through Jones's Citations, with Jones's counts and names. Today it offers every Artifact of the Source and every Citation on the chosen Artifact, so the researcher can pick a page the engine will refuse (`observations.other_graph`) and steps through another graph's work.

**Why now.** Slice 3 makes every surface work on one Evidence graph; the composer is the last one that still lists by Source. Working rule "The Citation composer offers only the graph's linked Artifacts"; the Citation half was added at planning (user, 2026-10-24).

**What stays.** The composer's layout, picker, stepper and default-Artifact rule (`defaultPendingArtifactID`). Every Source has one graph today, linked to all its Artifacts, so on an existing project the composer lists exactly what it lists now. The canvas's own cite gate and Files list still read the Source's Artifacts (S12-13); the canvas's Files menu counts become the graph's counts, because they share the re-keyed count read (D-5).

## 2. User stories

**US-1.** As a researcher, when I open the Citation composer on one Evidence graph, I see only the Artifacts linked to that graph, so I never cite a page the graph cannot use.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source with Artifacts `Page 1` (art-1), `Page 2` (art-2); graph Jones links art-2 only; open the composer on Jones | The Artifact picker lists `Page 2` only; it is selected |
| AC-1.2 | Same Source; graph Default links art-1 and art-2 | The picker lists `Page 1`, `Page 2`, in the Source page's Artifact order |
| AC-1.3 | Source has art-1; graph Smith links no Artifact (its only Artifact was deleted) | The composer shows the empty callout: title "This evidence graph has no artifacts", message "Every citation needs an artifact linked to this evidence graph. Add one from the Source page, then return here.", button "Go to Source page" |
| AC-1.4 | Source has no Artifact; Default links none | The same callout as AC-1.3 |
| AC-1.5 | An Add property entry (`CitationComposerEntry.addProperty`) opens the composer on Jones with `artifactID` = art-1 (not linked to Jones; a stale Back entry) | The composer opens on `Page 2` with a new Citation: `prepare` drops the unlisted Artifact and the one-Artifact branch selects `Page 2` |
| AC-1.6 | The project's graph list fails to load | The composer shows its load-failed state with "catalogUnavailable" copy, as when the graph rows fail today |

**US-2.** As a researcher, when a page is linked to two graphs, the composer on one graph steps only through that graph's Citations, so the other graph's work does not mix into mine.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | art-2 linked to Jones and Smith. CIT-1 on art-2 has an Observation on a Jones Subject; CIT-2 on art-2 has an Observation on a Smith Subject; CIT-3 on art-2 has no Observation | On Jones the stepper lists CIT-1, CIT-3; on Smith it lists CIT-2, CIT-3 |
| AC-2.2 | Same; CIT-1's Observations: 2 on Jones Subjects (Ada, Ben), 1 on a Smith Subject (Clara) | On Jones, CIT-1's row reads 2 observations and names Ada, Ben; on Smith it reads 1 and names Clara |
| AC-2.3 | Same data as AC-2.1, Artifact picker on Jones | `Page 2`'s menu count reads 2 (CIT-1, CIT-3) |
| AC-2.4 | One-graph Source with CIT-1 (Observations) and CIT-3 (none) on art-1 | The stepper lists CIT-1, CIT-3 and counts 2: unchanged from today |
| AC-2.5 | On Jones, save the first Observation of a new Citation on art-2 | The stepper refreshes and lists the new Citation on Jones; on Smith it is no longer listed |

## 3. Scope

| In | Covered by |
| --- | --- |
| `citations.ListByArtifact` takes a graph; `CountBySource` becomes `CountByEvidenceGraph`; one membership filter (D-1, D-2, D-3) | STEP-1, AC-2.1 – AC-2.4 |
| Link rows report `CitationSource` (D-6) | STEP-1, T-12 |
| Proto, handlers, dispatch: `ListCitationsByArtifactRequest.graph_id`; `CitationCountsBySource` → `CitationCountsByEvidenceGraph` (D-4) | STEP-2, AC-2.1, AC-2.3 |
| Swift store methods, query keys, registry fetch and invalidations re-keyed by graph (D-5, D-6) | STEP-3, STEP-4, AC-2.5 |
| `CatalogEvidenceGraph.linkedArtifacts(_:)` *(new)*; the composer's `artifacts` reads the graph list (D-7, D-8) | STEP-5, AC-1.1 – AC-1.6 |
| Composer place keys and warm include the graph list (D-8) | STEP-5, AC-1.6 |
| Empty callout copy names the graph (D-9) | STEP-6, AC-1.3, AC-1.4 |
| `source-layer-data-model.md` § 7.1 rule for a graph's Citations | STEP-8 |

| Out | Home |
| --- | --- |
| The canvas's cite gate (`EvidenceGraphModel.canCite`) and Files list (`fileRows`) reading the graph's linked Artifacts | S12-13 (its scope entry already says so) |
| Editing links in the UI | S12-12, S12-13 |
| Deleting a Citation whose Observations span two graphs from one graph's composer | not planned — `citations.Delete` already refuses a Citation with Observations (`deleteimpact`), so no cross-graph delete is reachable |
| Replacing `WriteEffects.citation_sources` with graph ids (D-6) | not planned |
| A per-graph label for the Default graph | S12-11 |

## 4. Current state

Verified at `7a1398c`.

| Area | Where | Today |
| --- | --- | --- |
| Citation list | `core/database/citations/citations.go` `ListByArtifact(c, artifactID)`, `sqlListByArtifact`, `attachPersonNames`, `sqlPersonNamesByArtifact` | Every Citation on the Artifact, `observation_count` over all its Observations, person names over all its Observations. Wrong: not scoped to a graph |
| Citation counts | same file, `CountBySource(c, sourceID)`, `sqlCountBySource` | Counts every Citation per Artifact of a Source. Wrong: keyed by Source |
| Link-row effect | `core/database/effects/registry.go` `registry["evidence_graph_artifacts"]`; `registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` row `{"evidence_graph_artifacts", false, true, false}` | Sets `SourceScope` and `Lists` only, no `CitationSource`. Wrong once D-4's counts read links |
| Graph's Citations, unlink check | `core/database/evidencegraphs/evidencegraphs.go` `CitationsUsingArtifactTx`, `sqlCitationsUsing` | Citations on an Artifact with ≥1 Observation on a Subject of the graph; ID and Ref only. The unlink refusal; stays as is (D-3) |
| Proto | `api/proto/engine.proto` `METHOD_CITATION_COUNTS_BY_SOURCE = 73`, `METHOD_LIST_CITATIONS_BY_ARTIFACT = 74`, `CitationCountsBySourceRequest { project_dir = 1; source_id = 2 }`, `CitationCountsBySourceResponse`, `ListCitationsByArtifactRequest { project_dir = 1; artifact_id = 2 }` | Source-keyed count, unscoped list |
| Handlers | `api/ffi/handlers/citations.go` `CitationCountsBySource`, `ListCitationsByArtifact`; `api/ffi/dispatch.go` `MethodCitationCountsBySource`, `MethodListCitationsByArtifact` | Call the two reads above |
| Go tests | `core/database/citations/citations_test.go` `TestCitations` cases "list by artifact" and the `CountBySource` cases (lines 426 – 552); `api/ffi/handlers/citations_test.go` `TestCitationCountsBySource`, `TestListCitationsByArtifact`, and the list call at line 245 | Source-keyed / unscoped |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `citationCountsBySource(projectDir:sourceID:)`, `listCitationsByArtifact(projectDir:artifactID:)`; `GoStore.swift` (lines 1760, 1956); `FakeStore.swift` (lines 2815, 3325); `CoreInvoke.swift` `CoreMethod.citationCountsBySource`; `macos/ProvenenciaTests/OnboardingModelTests.swift` stub (lines 785, 788) | Source-keyed / unscoped |
| Graph model | `GenealogyStore.swift` `struct CatalogEvidenceGraph` (`artifactIDs`); `extension CatalogArtifact` (line 108) holds `displayTitle` | No "linked Artifacts" helper |
| Query keys | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.citationCounts(project:sourceId:)`, `.citationsByArtifact(project:artifactId:)` | Source / Artifact keyed |
| Registry | `CatalogQueryRegistry.swift` fetch cases (lines 233, 274); `invalidations(for effects:)` (line 311: one `.key(.citationCounts(sourceId))` per `effects.citationSources`, `.allCached(.citationsByArtifact)`); `invalidation(for kind:by mutation:)` `.citationCounts` (line 418: `.deletedSubject(sourceId)` → that Source's key) | Source-keyed |
| Mutation | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `case deletedSubject(sourceId: String)` (line 36) | Its payload is read only by the registry's `.citationCounts` case (line 420) |
| Places | `PlaceRegistry.swift` `.sourceCitationComposer` (line 42) and `.sourceGraph` specs | Composer: `.citationCounts(sourceId)`, `.citationsByArtifact(artifactId)`, no graph list. Canvas: `.citationCounts(sourceId)` |
| Composer | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `artifacts` (line 139) = `workspace?.artifacts`; `hasNoArtifacts`; `citationCountsKey`, `citationsKey(artifactID:)`; `warmCatalogHandles` | Lists every Artifact of the Source |
| Composer context | `CitationComposerContext.swift` `applyWrite`, `reloadListedCitations` | Query `.citationsByArtifact(artifactId)` |
| Empty gate | `CitationComposerView.swift` `noArtifactGate`; `L10n.CitationComposer.noArtifactsTitle` ("This source has no artifacts"), `.noArtifactsCallout`; `macos/App/Resources/Localizable.xcstrings` keys `citationComposer.noArtifactsTitle`, `citationComposer.noArtifactsCallout` | Source wording |
| Canvas | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `citationCountsKey` (line 185), `let graphID` | Counts by Source |
| Swift tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` (`makeStore`, `seedArtifact`, `ghostArtifactResetsToANewCitation`); `CatalogQueryRegistryTests.swift` (lines 142, 209 – 223, 634 – 635, 886 – 907); `PlaceRegistryTests.swift` (lines 82, 167, 186, 350); `WorkspaceSessionTests.swift` (lines 737, 772, 797) | Use the Source / Artifact keys |

**Precedent.** S12-08 (#528) re-keyed Source-keyed RPCs to graphs: its D-1 kept method numbers, renamed messages named for a Source, and reserved the old field (`reserved 2; reserved "source_id"; string graph_id = 3;`). Copy that shape for D-4. S12-08 D-8 shows the key rename (`.sourceGraph(sourceId:)` → `.evidenceGraph(graphId:)`). `extension CatalogArtifact { displayTitle }` (S12-02) is the precedent for a small shared read on a catalog model type.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A graph's Citations on an Artifact** are the Citations on that Artifact with at least one Observation on a Subject of the graph, plus the Citations with no Observation at all. Two SQL fragments *(new)* in `citations.go`: `sqlUsedByGraph` = `EXISTS (SELECT 1 FROM observations o JOIN subjects s ON s.id = o.subject_id WHERE o.citation_id = c.id AND s.graph_id = ?)`, commented as the same rule as `evidencegraphs.sqlCitationsUsing` (the Unlink check; change both together), and `sqlInGraph` = `(` + `sqlUsedByGraph` + ` OR NOT EXISTS (SELECT 1 FROM observations o WHERE o.citation_id = c.id))`, used by both reads | A new Citation is created with no Observation (`CitationComposerContext.createCitationIfNeeded`), so it must stay reachable on the graph that made it; it belongs to no graph yet. The predicate reads no link row; the counts (D-4) do, so link rows report `CitationSource` (D-6) | Hiding Observation-less Citations — the researcher loses the Citation they just made; adding `citations.graph_id` — a migration and a second source of truth beside the Observations |
| D-2 | `citations.ListByArtifact(c *database.Catalog, graphID, artifactID []byte)` — `ErrInvalid` unless both are 16 bytes. Rows filtered by D-1. `observation_count` counts only Observations on Subjects of the graph. `attachPersonNames` takes the graph and keeps only Observations whose `subject_id` is in the graph (both union branches add `JOIN subjects so ON so.id = o.subject_id AND so.graph_id = ?`) | AC-2.2: a Citation can hold Observations in two graphs (each Observation is checked alone by `requireSameGraph`); the composer already loads only the graph's Observations (`fetchOpenCitation` filters the graph rows), so the stepper's count matches what opens | Counting all Observations — the stepper says 3 and the composer opens 2 |
| D-3 | `CitationsUsingArtifactTx` and its SQL stay as they are | The unlink refusal is about Citations the graph uses; an Observation-less Citation does not block an unlink, and must not | Sharing `sqlInGraph` with it — would block an unlink on an empty Citation |
| D-4 | Proto, numbers kept: `METHOD_CITATION_COUNTS_BY_SOURCE = 73` → `METHOD_CITATION_COUNTS_BY_EVIDENCE_GRAPH = 73`; `CitationCountsBySourceRequest/Response` → `CitationCountsByEvidenceGraphRequest/Response`, request `reserved 2; reserved "source_id"; string graph_id = 3;`. `ListCitationsByArtifactRequest` gains `string graph_id = 3;` (required; `parseID`). `citations.CountBySource` → `CountByEvidenceGraph(c, graphID)`: counts per Artifact linked to the graph (`JOIN evidence_graph_artifacts l ON l.artifact_id = c.artifact_id AND l.graph_id = ?`) filtered by D-1, so Go and Swift (D-7) mean the same thing by "the graph's Artifacts". Until S12-13 the canvas Files menu shows no count for an Artifact not linked to its graph (only reachable on a multi-graph Source). Handler `CitationCountsBySource` → `CitationCountsByEvidenceGraph`; `dispatch.go` constant → `MethodCitationCountsByEvidenceGraph` | Working rules "replaces Source-keyed fields" and "A re-keyed symbol is renamed"; S12-08 D-1 shape. `ListCitationsByArtifact` still lists by Artifact, so it keeps its name | Keeping `source_id` beside `graph_id` — the churn rule; reusing field 2 — an old Source id would decode as a graph id |
| D-5 | Swift: `GenealogyStore.citationCountsByEvidenceGraph(projectDir:graphID:)` *(renamed)*, `listCitationsByArtifact(projectDir:graphID:artifactID:)`; `CoreMethod.citationCountsByEvidenceGraph`; keys `.citationCounts(project:graphId:)`, `.citationsByArtifact(project:graphId:artifactId:)`. The canvas's `citationCountsKey` and the `.sourceGraph` place read `.citationCounts(graphId: graphID)` | One count read, keyed by what it counts. The canvas key moves because the old key is deleted | A second count key for the canvas — two reads of one thing |
| D-6 | Invalidation: `WriteEffects.citation_sources` stays and is the one path to the citation reads. Link rows join it: `registry["evidence_graph_artifacts"]` in `core/database/effects/registry.go` gains `CitationSource: up("graph_id", "source_id")`, because D-4's counts read links (review M-8); a link, unlink or graph edit then reaches the citation reads with no Swift change. `invalidations(for effects:)` appends `.allCached(.citationCounts)` and `.allCached(.citationsByArtifact)` once when `effects.citationSources` is non-empty (the per-Source loop goes). The `.citationCounts` and `.citationsByArtifact` specs drop `.deletedSubject` from `invalidateOn`; the `.citationCounts` case body in `CatalogQueryKey.Kind.invalidation(project:mutation:)` (line 418; the switch has no `default`) becomes `return .allCached(.citationCounts)`, as `.citationsByArtifact` reads at line 436. `.deletedSubject(sourceId:)` → `.deletedSubject` (no payload) in `CatalogMutation.swift` line 36 (its doc comment becomes "A Subject (including a bridge) was erased. Releases connection facets so Observation `usedBy` must reload; the citation reads follow the write's effects."): its last reader was that case; callers `CitationComposerModel.swift` line 583 and `EvidenceGraphModel.swift` line 1004 change; it still drives `.propertiesWorkspace`. The `.citationCounts` spec gains `whenHidden: .markStale`, as `.citationsByArtifact` has. Subject erase still reloads both reads: its Observation erase (`deleteimpact` `eraseObservation`) reports `CitationSource` through the `observations` effect | The effect names Sources and the keys name graphs; the session holds at most the open graphs' counts, so dropping all is cheap, and `.citationsByArtifact` is already invalidated this way. The one Go effects change is the link-row line (review M-8) | Re-keying the effect to graph ids or driving the keys from `effects.evidenceGraphs` — touches `effects` registry, wire and FakeStore for no visible gain (subject moves and label edits also set `evidenceGraphs`); keeping `.deletedSubject`'s payload — a second path to the same invalidation, with a dead value (review M-1) |
| D-7 | `extension CatalogEvidenceGraph { func linkedArtifacts(_ artifacts: [CatalogArtifact]) -> [CatalogArtifact] }` *(new)* in `GenealogyStore.swift`, after the struct: returns `artifacts` filtered to `artifactIDs`, in the order of `artifacts`. The session read beside `graphsBySource(in:)` in `SourcesListNavigation.swift`: `@MainActor static func linkedArtifacts(graphID: String, artifacts: [CatalogArtifact], in session: WorkspaceSession) -> [CatalogArtifact]?` *(new)* — `nil` while the graph list has no value; `[]` when the list has no entry with that id; else that entry's `linkedArtifacts(artifacts)` | One rule and one lookup for "the graph's Artifacts" that S12-13 calls as is for the canvas (review M-4); order stays the Source page's | Filtering inline in the composer — S12-13 would copy it; ordering by `artifactIDs` — that order is the link order, not page order |
| D-8 | Composer: `artifacts` = `SourcesListNavigation.linkedArtifacts(graphID: graphID, artifacts: workspace?.artifacts ?? [], in: session) ?? []`. `evidenceGraphsKey` *(new)* = `.evidenceGraphsList(project:)`, queried after `graphKey` in `warmCatalogHandles` and added to its `guard`; `warmCatalogHandles` queries it and throws `CatalogWarmError.handleMissing` when it has no ready value. The `.sourceCitationComposer` place appends `.evidenceGraphsList(project:)` after `.citationCounts` (before the optional `.citationsByArtifact`) | Working rule "One graph list": the composer filter reads it; `GetSourceWorkspace` does not gain graphs. A missing list is a load failure like the graph rows | Falling back to every Artifact while the list loads — would offer pages the engine refuses |
| D-9 | Copy, keys unchanged: `citationComposer.noArtifactsTitle` = "This evidence graph has no artifacts" (comment "Callout title when the Evidence graph links no Artifacts"); `citationComposer.noArtifactsCallout` = "Every citation needs an artifact linked to this evidence graph. Add one from the Source page, then return here." (comment "Left-pane callout when the Evidence graph links no Artifacts"). One gate for both cases | User, 2026-10-24: graph wording always. "Add one" is true today: a new Artifact links to every graph (S12-05) | Two copies, Source and graph — user chose one |
| D-11 | `CitationComposerContext.citationsKey(artifactID: String) -> CatalogQueryKey` *(new)* builds `.citationsByArtifact(project:graphId: graphID, artifactId:)`; `applyWrite`, `reloadListedCitations` and the model's `citationsKey(artifactID:)` (which becomes `context.citationsKey(artifactID:)`) call it | One builder for the re-keyed key instead of three (review M-5) | Re-keying three inline sites |
| D-10 | FakeStore mirrors D-1 / D-2 through one private helper `citationIsInGraph(_ citationID: String, graphID: String) -> Bool` *(new)*: true when `observationsByGraph[graphID]` holds one of its Observations, or no entry of `observationsByGraph` holds any. `listCitationsByArtifact(projectDir:graphID:artifactID:)` filters with it; `observationCount` counts only `observationsByGraph[graphID]`; `personNames(for:graphID:projectDir:)` reads only `observationsByGraph[graphID]` and `subjectsByGraph[graphID]`. `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact` and `updateEvidenceGraph` return `citationSources: [graph.sourceID]` beside their list effect, as the engine now does. `citationCountsByEvidenceGraph(projectDir:graphID:)` finds the graph in `evidenceGraphsByProject[projectDir]` (no graph → `[:]`) and counts Citations on its `artifactIDs` that pass the same helper (D-4) | Swift tests exercise the same rule the engine holds; list and counts cannot drift | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `sqlUsedByGraph`, `sqlInGraph` *(new)* | SQL fragment | `evidencegraphs.sqlCitationsUsing` | Same "used by the graph" rule plus Citations with no Observation; the comment names its twin (D-1); `citations` owns the Citation list reads |
| `citations.CountByEvidenceGraph` *(renamed)* | func | `citations.CountBySource` | Replaces it |
| `CitationCountsByEvidenceGraph*` *(renamed)* | FFI | `CitationCountsBySource*` | Replaces it (D-4) |
| `ListCitationsByArtifactRequest.graph_id` *(new)* | FFI field | — | Extends the existing request |
| `CatalogEvidenceGraph.linkedArtifacts(_:)`, `SourcesListNavigation.linkedArtifacts(graphID:artifacts:in:)` *(new)* | func | `SourcesListNavigation.graphsBySource(in:)`, `entryGraph` | Those group graphs; none maps a graph to its Artifacts or finds a graph by id. Searched `artifactIDs.contains`, `artifactIDs`, `id == graphID` in `App/` |
| `CitationComposerContext.citationsKey(artifactID:)` *(new)* | key builder | `CitationComposerModel.citationsKey(artifactID:)` (private) | Moves it to the context so its two inline sites share it (D-11) |
| `CitationComposerModel.evidenceGraphsKey` *(new)* | key accessor | `graphKey`, `workspaceKey` | Same pattern, one more key |

**Builds on.** S12-05's one graph list key and its list invalidation; S12-08's re-key shape; the catalog query registry.
**Deletes.** `citations.CountBySource`, `sqlCountBySource`, the `CitationCountsBySource` handler, method name and messages, `citationCountsBySource` in every store, the Source-keyed `.citationCounts` and `.citationsByArtifact` cases, the per-Source loop in `invalidations(for effects:)`, the `.citationCounts` case's `.deletedSubject` branch in `invalidation(project:mutation:)`, `.deletedSubject`'s `sourceId` payload.
**Next adds.** S12-13 points `EvidenceGraphModel.canCite` and `fileRows` at `linkedArtifacts`.

### Schema

None — no catalog change. D-1 is a derived read.

### Go core

| Package | Change |
| --- | --- |
| `core/database/citations` | `ListByArtifact(c, graphID, artifactID)` (D-2); `CountByEvidenceGraph(c, graphID) (map[string]int32, error)` (D-4); `attachPersonNames(db, graphID, artifactID, list)`; `sqlInGraph` (D-1) |

Errors: `citations.invalid` (`ErrInvalid`) for a bad id length; an unknown graph returns an empty list or map.

Sketch of the list SQL (bind order: graph for the count, artifact, graph for the filter):

```sql
SELECT c.id, c.ref, c.artifact_id, c.locator_json, ...,
  (SELECT COUNT(*) FROM observations o JOIN subjects s ON s.id = o.subject_id
     WHERE o.citation_id = c.id AND s.graph_id = ?)
FROM citations c WHERE c.artifact_id = ? AND <sqlInGraph>
ORDER BY c.ref COLLATE NOCASE
```

### FFI

`api/proto/engine.proto` per D-4. Handlers in `api/ffi/handlers/citations.go`; `api/ffi/dispatch.go` case renamed. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Artifact picker | `PVSelect` (unchanged) | `CitationComposerEvidenceColumn.swift` `artifactSelect` | Reads `model.artifacts`; no view change |
| Empty gate | `PVCallout` (unchanged) | `CitationComposerView.swift` `noArtifactGate` | Copy only (D-9) |

States: empty → the gate (AC-1.3, AC-1.4); graph list failed → load failed (AC-1.6). Accessibility: unchanged; the picker's position string counts the filtered list.

### Strings

Changed values: `L10n.CitationComposer.noArtifactsTitle`, `L10n.CitationComposer.noArtifactsCallout` and their `Localizable.xcstrings` entries (D-9). None added, none removed.

## 7. Steps

**STEP-1 — Citation reads per graph (Go)**
Files: `core/database/citations/citations.go`, `core/database/citations/citations_test.go`, `core/database/effects/registry.go`, `core/database/effects/registry_completeness_test.go`. Skill: `add-catalog-query`.
Do: D-1, D-2, D-4's `CountByEvidenceGraph`; D-6's link-row `CitationSource` line, with the `TestSourceEvidenceIsNotActivity` row flipped to `{"evidence_graph_artifacts", false, true, true}` and T-12; delete `CountBySource` and `sqlCountBySource`. `sqlInGraph` is a `const` concatenated into `sqlListByArtifact`, `sqlCountByEvidenceGraph` *(renamed)* and both branches of `sqlPersonNamesByArtifact`; bind order is the order of `?` in the final string (list: graph, artifact, graph; names: per branch graph, artifact, graph; count: graph, graph). Each existing case that calls `ListByArtifact` (lines 426, 461) gets the graph from `evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)`; the two `CountBySource` cases become `CountByEvidenceGraph` cases on that graph. Add T-1 – T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/citations/... ./core/database/effects/... ./core/database/evidencegraphs/...` passes. `api/` does not build yet; that is STEP-2.

**STEP-2 — Proto and handlers**
Files: `api/proto/engine.proto`, generated Go and Swift, `api/ffi/handlers/citations.go`, `api/ffi/handlers/citations_test.go`, `api/ffi/dispatch.go`. Skill: `add-ffi-handler`.
Do: D-4; `./scripts/generate-proto.sh`. `TestCitationCountsBySource` → `TestCitationCountsByEvidenceGraph`; add the `graph_id` cases to `TestListCitationsByArtifact` (T-4) and pass `GraphId` at line 245 (`TestCreateCitationRefusesSymlinkedCrop`): read it with `ListEvidenceGraphs` for the project, the one graph of the fixture's Source.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` and `go vet -tags fts5 ./api/... ./core/...` pass.

**STEP-3 — Swift store and keys**
Files: `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`, `CatalogQueryKey.swift` (both case patterns at lines 110 and 115 gain the new arity), `CatalogQueryRegistry.swift` (fetch cases and D-6), `CatalogMutation.swift` (D-6), `WorkspaceSession.swift` (only if a case pattern needs the new arity), `PlaceRegistry.swift`, `EvidenceGraphModel.swift` (`citationCountsKey`), `CitationComposerModel.swift` (`citationCountsKey`, `citationsKey`), `CitationComposerContext.swift` (`applyWrite`, `reloadListedCitations`), `macos/ProvenenciaTests/OnboardingModelTests.swift`. Skill: `use-catalog-session`.
Do: D-5, D-6, D-10, D-11. Every caller passes its `graphID`, including `cachedCitation(id:)`'s loop in `CitationComposerModel.swift` (lines 1107 – 1123).
Check: the app builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'`), and `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerModelTests` still passes (the composer filter is STEP-5; this proves the re-keyed FakeStore reads keep today's one-graph behaviour).

**STEP-4 — Key and invalidation tests**
Files: `CatalogQueryRegistryTests.swift`, `PlaceRegistryTests.swift`, `WorkspaceSessionTests.swift`.
Do: update the §4 test lines to the new keys with graph id `"g1"`; `seedStore` in `CatalogQueryRegistryTests` adds `CatalogEvidenceGraph(id: "g1", ref: "GPH-1", sourceID: "s1", name: "Default", description: "", artifactIDs: ["art-0"])` to `evidenceGraphsByProject[projectDir]`, and `citationCountsLoadBySource` is renamed `citationCountsLoadByEvidenceGraph`; change the `.deletedSubject(sourceId: …)` test calls (lines 634, 674, 695, 904, 928) to `.deletedSubject` and their expectations to D-6; add T-5. No app code changes in this step.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/PlaceRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-5 — Composer lists the graph's Artifacts**
Files: `GenealogyStore.swift` (D-7), `CitationComposerModel.swift`, `PlaceRegistry.swift`, `CitationComposerModelTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-7, D-8 (`SourcesListNavigation.swift` too). In `CitationComposerModelTests.makeStore`, the fixture graph links `["art-0", "art-1"]` (every Artifact any test seeds, so `steppingWithDirtyCitationFieldsAsks` keeps selecting art-1); T-6 – T-9 set `store.evidenceGraphsByProject[projectDir]` themselves. `CitationComposerDepictionTests` (line 55) and `CitationAutoTranscribeTests` (line 381) keep their art-0-only graphs: they select no other Artifact. Add T-6 – T-9.
Check: `xcodebuild test … -only-testing:ProvenenciaTests/CitationComposerModelTests -only-testing:ProvenenciaTests/CitationComposerDepictionTests -only-testing:ProvenenciaTests/PlaceRegistryTests` passes.

**STEP-6 — Empty callout copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-7 — Composer Citations per graph**
Files: `CitationComposerModelTests.swift`.
Do: add to `CitationComposerModelTests`: `makeModel` gains `graphID: String? = nil` (nil → `self.graphID`), passed to every entry case; `seedSecondGraph(_ store: FakeStore)` *(new)* appends graph `g-2` (ref `GPH-2`, name `Smith`, Source `src-1`, links `["art-0"]`) to `evidenceGraphsByProject[projectDir]` and person `sub-person-2` (ref `CPR-2`, label `Clara`) to `subjectsByGraph["g-2"]`. T-10, T-11 against the FakeStore rule (D-10); no model change beyond STEP-3.
Check: `xcodebuild test … -only-testing:ProvenenciaTests/CitationComposerModelTests` passes.

**STEP-8 — Docs**
Files: `docs/source-layer-data-model.md` § 7.1 rules list, `.agents/skills/add-workspace-place/SKILL.md` (lines 172 – 173), `.agents/skills/add-catalog-delete/SKILL.md` (lines 57, 111), `core/database/effects/registry.go` (comments at lines 57 – 59 and 89 – 91), `core/database/effects/README.md` (the link-row sentence), `docs/deployment-plan/spike-12/deployment-plan.md` (S12-10 row status only; the scope entries are already current), `docs/deployment-plan/spike-12/completed.md`, this plan.
Do: add after the `Unlink` bullet (line 420): "A graph's Citations on an Artifact are those the graph uses (the `Unlink` rule above) plus those with no Observation yet; `citations.ListByArtifact` and `CountByEvidenceGraph` read them, and the Citation composer lists only the graph's linked Artifacts." Skills: graph and composer deletes apply `session.apply(effects, .deletedSubject)`; the citation reads are keyed by graph and dropped project-wide when `citationSources` is non-empty; Subject erase reloads Observation `usedBy` through `.deletedSubject` and the citation reads through the effects. `registry.go` comments: `CitationSource` is the Source whose change drops the citation reads of every graph (the app keys them by graph). README: link rows also set `Effect.CitationSource`, because a graph's citation counts read its links. Then the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-10-composer-linked-artifacts.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/citations/citations_test.go` `TestCitations` "list by artifact is per graph" *(new)* | From `mustSeed`: Default = `evidencegraphstest.SoleGraph(t, c, s.artifact.SourceID)`; `evidencegraphstest.Create` a second graph Smith linked to `s.artifact`; a Smith person via `subjectstest.On(t, c, userID, smith.ID, "person", "Clara")`; the Default Subject is `s.person`; CIT-A with an Observation on the Default Subject, CIT-B on the Smith Subject, CIT-C with none → Default lists A, C; Smith lists B, C | AC-2.1, D-1 |
| T-2 | Go | same, "list counts and names the graph's Observations" *(new)* | CIT-A made by one `CreateWithObservations` call with three inputs (allowed: `requireSameGraph` checks each Observation alone): `s.person` (Alice) and a second Default person Ben (`subjectstest.On`) on the Default graph, Smith person Clara → Default count 2, names [Alice, Ben]; Smith count 1, names [Clara] | AC-2.2, D-2 |
| T-3 | Go | same, "count by evidence graph" *(new; replaces the CountBySource cases)* | T-1 data → Default {art: 2}, Smith {art: 2}; a second Artifact of the Source, created with `artifactstest.Create` **before** Smith (so only Default links it; Smith is created linking `s.artifact` only), with a Citation → absent from Smith's map; one-graph Source → today's counts; 1-byte id → `ErrInvalid` | AC-2.3, AC-2.4, D-4 |
| T-4 | Go | `api/ffi/handlers/citations_test.go` `TestCitationCountsByEvidenceGraph` *(renamed)*, `TestListCitationsByArtifact` | `citationFixture` has one graph; its id comes from `ListEvidenceGraphs` in each test. Missing or bad `graph_id` → error; valid → the cases' existing expectations hold (the fixture's Citations are all on that graph) | D-4 |
| T-5 | Swift | `CatalogQueryRegistryTests` effects test (lines 886 – 907) | effects with `citationSources: ["s1"]` (as a Subject erase reports) → `.allCached(.citationCounts)`, `.allCached(.citationsByArtifact)`; `invalidations(by: .deletedSubject)` holds neither citation kind; `hiddenPolicy` of `.citationCounts(project:graphId:)` is `.markStale` | AC-2.5, D-6 |
| T-6 | Swift | `CitationComposerModelTests` `composerListsOnlyTheGraphsLinkedArtifacts` *(new)* | art-0, art-1 on the Source; graph links art-1 → `artifacts.map(\.id) == ["art-1"]`, `selectedArtifactID == "art-1"` | AC-1.1, D-7, D-8 |
| T-7 | Swift | same, `composerKeepsPageOrderForLinkedArtifacts` *(new)* | links [art-1, art-0] → `["art-0", "art-1"]` | AC-1.2, D-7 |
| T-8 | Swift | same, `graphWithNoLinksShowsTheNoArtifactGate` *(new)* | Source has art-0; graph links none → `phase == .compose`, `hasNoArtifacts` | AC-1.3, AC-1.4 |
| T-9 | Swift | same, `unlinkedEntryArtifactOpensANewCitation` *(new)*; `graphListFailureFailsTheLoad` *(new)* | `seedArtifact` plus art-1 on the Source; graph `g-1` links art-1 only; `makeModel(store:)` (the `.addProperty` entry with `artifactID: "art-0"`) → `selectedArtifactID == "art-1"`, `activeCitationID == nil`; `listEvidenceGraphsError` set → `phase == .loadFailed` | AC-1.5, AC-1.6, D-8 |
| T-10 | Swift | same, `stepperListsOnlyThisGraphsCitations` *(new)* | `seedArtifact`, `seedSecondGraph`; `citationsByID`: `cit-a` (CIT-A), `cit-b` (CIT-B), `cit-c` (CIT-C) on art-0; `observationsByGraph["g-1"]` holds one Observation of cit-a on `subjectID`, `observationsByGraph["g-2"]` one of cit-b on `sub-person-2` → on `g-1` `listedCitations` refs = [CIT-A, CIT-C], `listedCount(for: "art-0") == 2`; on `g-2` = [CIT-B, CIT-C] | AC-2.1, D-10 |
| T-11 | Swift | same, `firstObservationMovesANewCitationIntoTheGraph` *(new)* | `seedArtifact`, `seedSecondGraph`; on `g-1`, save the first Observation (as `firstRowSaveOnNewCreatesCitationAndExposesRef` does) → the new Citation is in `listedCitations`; `makeModel(store:graphID: "g-2", subjectID: "sub-person-2")` after `prepare()` does not list it | AC-2.5 |
| T-12 | Go | `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` row, plus a case in `core/database/evidencegraphs/evidencegraphs_test.go` *(new)*: `Link` then `Unlink` through `writes.Run` → each write's `res.Effects.CitationSources` (the `writes.Result`, read as `likenesses_test.go` line 424 reads `res.Effects`) holds the graph's Source | D-6, review M-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (fixture graph links only art-0, line 109), `CitationComposerDepictionTests`, `PlaceRegistryTests`, `WorkspaceSessionTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no brief; no view code changes
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — n/a
- [ ] Domain doc updated: source-layer-data-model.md § 7, the graph's Citations rule (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Until S12-12 a second graph cannot be made in the app; make one in a scratch project with the engine test helpers, or check the one-graph case and rely on T-1 – T-11 for two graphs.

1. Open a project with a one-graph Source with two Artifacts and a Citation on each. Open the composer. The picker lists both; the stepper lists each Artifact's Citations as before (AC-1.2, AC-2.4).
2. On a Source with no Artifact, open the composer from a Subject. The callout reads "This evidence graph has no artifacts" (AC-1.4).
3. With a two-graph scratch catalog (Jones linked to Page 2 only): open the composer on Jones. The picker lists Page 2 only; Smith's Citations on Page 2 are not in the stepper (AC-1.1, AC-2.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A test or caller passes a Source id where a graph id now goes | Every renamed signature takes `graphID`; T-3 and T-4 cover bad ids |
| The canvas's Files menu counts change on a multi-graph Source | Intended (D-5); the Files list itself is S12-13 |

**Stop and report if:**

- A Subject erase's `WriteEffects` has empty `citation_sources` while the Subject had Observations (D-6 then leaves the citation reads stale).

- `citations.ListByArtifact` or `CountBySource` has a caller not listed in §4.
- `CitationComposerModel.artifacts` is read outside the composer and that caller needs every Source Artifact.
- The composer place cannot list `.evidenceGraphsList` without the `PlaceRegistryTests` expectation for another place changing.

## 12. Open questions

None.

## 13. PR

**Title:** S12-10 — Composer offers linked Artifacts
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-10-composer-linked-artifacts.md`.
