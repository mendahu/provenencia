# S12-13 — Artifact graph checkboxes

**Status:** Landed in [#562](https://github.com/mendahu/provenencia/pull/562) (2026-10-25)
**Landed:** Differences from the plan: `evidenceGraphIndex` is internal, not private, because its builder lives in `WorkspaceSession+EvidenceGraphs.swift`; `SourceArtifactLinkEdit` is the nested `SourceArtifactsSection.LinkEdit`; the 264pt cap, the bordered row frame and the row checkbox with its locked spoken value are shared recipe views (`PVChecklistFrame`, `PVChecklistCheckbox` in `PVChecklistRows.swift`, key `checklist.lockedValue`), which supersedes D-12's copied frame (review round 1, F-4); `SourceArtifactsSection` gains `linkEditGraphs` (the card's rows from the session) and `FakeStore` gains `updateArtifactCalls` (T-8 and T-9 assert no or one `updateArtifact`); `saveFields` returns without a call while the link read has not finished; when the reload after a refused save has no value, the card's Read again error is the graph list's error; the new file's project ids are `AA121400000000000000000D` / `...0E`; `EvidenceGraphModelTests.makeStore` now links `a`, `b`, `c` to its graph and `selectStaysAvailableWithoutAnArtifact` seeds a graph with no links instead of a Source without Artifacts. Review round 1 changes: the card's save error shows only on the card that saved (F-1); a card save applies its link effects once, so the graph list reloads once (F-2, `CatalogWriteEffects.merging`); after a refused file the Add dialog's graph boxes are disabled and ignored on retry (F-3); a graph deleted during edit is skipped, and a failed reload offers Read again (F-5); membership reads as `CatalogEvidenceGraph.links` / `isOnlyArtifact` (F-6); and, by the user's decision of 2026-10-25, the card stays in edit and its save-error body follows what is left to save, which supersedes the AC-2.7 copy: `artifactCardSaveErrorKept` ("Your changes are kept. %@") when nothing was written, `artifactCardSaveErrorGraphKeeps` ("%1$@ keeps this artifact. %2$@") when a graph refused and nothing else is left, else the original wording (F-7). `FakeStore` gains `holdsLinkEvidenceGraphArtifact`, `listEvidenceGraphsCalls` and `updateArtifactError`; the new file's project ids are `AA1214000000000000000F01` / `...F02`. The manual check and the twelve-graph dialog stop condition were not run in the app by the implementer.
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D4](../design/archive/S12-D4-artifact-graph-links.md), board agreed 2026-10-25
**Board:** [Claude Design](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share), agreed 2026-10-25 (the S12-D4 row of frame buttons on the Source page board). The planner rendered all thirteen S12-D4 frames from the exported board and checked them against §6; §6 Frames describes them in words for an implementer who cannot open the link
**Depends on:** S12-12 (landed [#557](https://github.com/mendahu/provenencia/pull/557)), S12-11 (landed [#549](https://github.com/mendahu/provenencia/pull/549)), S12-10 (landed [#540](https://github.com/mendahu/provenencia/pull/540)), S12-D4 (board agreed 2026-10-25)
**Blocks:** —
**Branch:** `s12-13-artifact-graph-checkboxes`, branched from `main` after the plan branch merges
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-12 did; Swift only, no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (session cache reads)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [rules](../source-structure.md#rules); working rules: "Artifact ↔ graph links", "Single-graph Sources hide link checkboxes", "One graph list", "The graph list has its own effect", "High tolerance for churn"; the S12-05 plan D-10 (store methods), the S12-10 plan D-7 and [ship log](S12-10-composer-linked-artifacts.ship.md) PRF-1 / UX-1, the S12-11 plan D-2, the S12-12 [Landed note](S12-12-graph-form.md) (`pvFormDialog(item:)`, the locked checkbox)
**Planned:** 2026-10-25 against `main` @ `181dd41`
**Review:** [`S12-13-artifact-graph-checkboxes.review.md`](S12-13-artifact-graph-checkboxes.review.md) — Fits with changes / Ready with changes, 2 rounds; every blocker Fixed, M-9 logged

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

On a Source with two or more Evidence graphs, the researcher chooses an Artifact's graphs from the Source page itself: the Add artifact dialog lists the graphs as checkboxes, all checked, and an Artifact card in edit lists them checked for its current links, saved with the label and description. A graph that must keep the Artifact (its Citations use it, or it is the graph's only Artifact) shows the box locked with the reason. The canvas then reads the graph's own links: its cite tools and Files menu follow the graph, not the Source. This makes the bar's census case true: "a new page can be linked to one graph only" ([dogfood bar](../deployment-plan.md#goal-dogfood-bar) outcomes 2 and 8).

**Why now.** S12-12 made a second graph possible, so a new Artifact silently joining every household is now a real annoyance. S12-05 put link, unlink, the unlink preflight and `graph_ids` on the wire; nothing in the app calls link or unlink yet.

**What stays.** Engine, FFI and store protocol. The cards at rest, the grid, Add file, delete and cover. A one-graph Source looks and behaves exactly as now. The graph form and the Copy Source form.

## 2. User stories

**US-1.** As a researcher adding page 7 to a census with three household graphs, I link it only to the Jones household, so that the other households' composers do not offer it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source with one graph; Add artifact | No Evidence graphs group; Add sends `graphIDs: []` (today's call) |
| AC-1.2 | Source with graphs Brown (`g-b`), Jones (`g-j`), Smith (`g-s`); Add artifact | The group lists Brown, Jones, Smith in that order, each with its ref, all checked |
| AC-1.3 | Uncheck Brown and Smith; label "Page 7"; Add | `createArtifact` is called once with `graphIDs: ["g-j"]`; the dialog closes; the existing toast shows |
| AC-1.4 | Every box unchecked | Add is disabled (`canSubmitDraft == false`); "Check at least one evidence graph." shows under the group |
| AC-1.5 | Only Brown unchecked, label empty, no file; click outside | The discard confirm asks (`addSheetIsDirty == true`) |
| AC-1.6 | The graph list has no value yet (loading) on a three-graph Source | No group; Add sends `graphIDs: []` |

**US-2.** As a researcher, I change which graphs an existing page belongs to from its card, and I see which links I can't remove and why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | One-graph Source; Edit on a card | Today's edit, no group, no preflight call |
| AC-2.2 | Three graphs; P3 linked to Jones and Smith; Edit on P3 | The group shows; while the preflight runs, Save is disabled and the loading line shows; label and description are editable |
| AC-2.3 | The preflight returns a block for Smith with 2 Citations | Smith is checked and locked with "2 citations in this graph use this artifact"; Jones checked and clickable; Brown unchecked |
| AC-2.4 | P9 is Brown's only Artifact; Edit on P9 | Brown is checked and locked with "The only artifact in this evidence graph"; no preflight call for Brown |
| AC-2.5 | On P3: check Brown, uncheck Jones; Save | `linkEvidenceGraphArtifact(g-b, P3)` then `unlinkEvidenceGraphArtifact(g-j, P3)`, once each; no `updateArtifact` call (fields unchanged); the card leaves edit |
| AC-2.6 | On P3: change label only; Save | `updateArtifact` once; no link or unlink call |
| AC-2.7 | Link changes and the unlink throws `evidencegraphs.artifact_in_use` | The card stays in edit with a compact danger callout at the top of its fields: "Couldn’t save the artifact" / "Some changes may not be saved. Every field and box keeps its value — save again to finish. {that code's `L10n.Errors` copy}"; `pageError` stays nil; the boxes keep their values; the preflight runs again, so Jones shows locked |
| AC-2.8 | The preflight throws | A danger callout "Couldn’t read the evidence graphs" with Read again replaces the group; Save disabled; Read again repeats the preflight |
| AC-2.9 | Uncheck Jones; Cancel | No store call; reopening edit shows Jones checked |
| AC-2.10 | Only a box changed | Save is enabled (`canSaveFields == true`) |

**US-3.** As a researcher on one household's canvas, the cite tools and Files menu follow that graph's links.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Graph Jones links P3 only; the Source has P1–P3 with files | `fileRows` ids are `["P3"]` |
| AC-3.2 | Graph Brown links no Artifact, the Source has Artifacts | `canCite == false`; the callout reads "No artifacts in this evidence graph" |
| AC-3.3 | The graph list has no value | `canCite == true`, `fileRows` empty (unknown, as today while the Sources list loads) |
| AC-3.4 | The Citation composer on a graph with no links | The callout body reads the LK-12 copy |

**US-4.** As a contributor, graph-list reads cost one dictionary lookup.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The graph list's value changes (revision bumps) | `evidenceGraph(id:)`, `evidenceGraphs(sourceID:)` and `evidenceGraphsBySource` read the new rows; between changes they read a cached index |
| AC-4.2 | `SourcesModel` | `graphsBySourceCache` is gone; `rg graphsBySourceCache macos` prints nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| Graph checkbox group in Add artifact, all on, sent as `graph_ids` (LK-1 – LK-4) | STEP-3, STEP-5, AC-1.* |
| Graph checkbox group in card edit, saved with Save through link / unlink (LK-5, LK-9) | STEP-4, STEP-5, AC-2.* |
| Locked boxes: Citations use it (preflight) and only artifact (LK-6, LK-7) | STEP-4, AC-2.3, AC-2.4 |
| Card loading and read-error states; compact `SourceFormLoadArea` (LK-8) | STEP-4, STEP-5, AC-2.2, AC-2.8 |
| A refused card save shows on the card, not the page (LK-9 as agreed on the board) | STEP-4, STEP-5, AC-2.7 |
| Canvas `canCite` and `fileRows` read the graph's links; canvas callout copy (LK-11) | STEP-6, AC-3.1 – AC-3.3 |
| Composer callout copy (LK-12, S12-10 review UX-1) | STEP-6, AC-3.4 |
| Id-indexed graph list in the session behind its accessors; `SourcesModel.graphsBySourceCache` and `SourcesListNavigation.graphsBySource` folded into it (S12-10 PRF-1) | STEP-1, AC-4.* |
| `CatalogEvidenceGraph.hasArtifact`, read by the canvas and both `SourcesListNavigation` checks (review M-3) | STEP-1, STEP-6, AC-3.2 |
| `PVArtifactChecklistState` renamed `PVChecklistState` and reused by the graph group (review M-2) | STEP-3, AC-1.3 |
| `CoreInvokeError.code`, read by the graph form and the card (review M-6) | STEP-4, AC-2.7 |
| VoiceOver for the group and locked rows (LK-10) | STEP-5, manual check |

| Out | Home |
| --- | --- |
| Links shown on a card at rest | not planned (product owner, 2026-10-25) |
| Citation list under a locked box on the card | not planned (product owner, 2026-10-25); the graph form's Kept linked group lists them |
| An All row or a filter in the graph group | not planned |
| Graph create or edit | S12-12 (landed) |
| Per-graph progress | S12-14 |
| Moving Add artifact onto `pvFormDialog(item:)` | not planned — its draft lives on `SourceArtifactsSection`, not a model object (D-11) |
| Engine, FFI or store protocol changes | not planned; S12-05 shipped them |
| Deleting `SourcesListNavigation.entryGraph` | S12-14 |
| Folding `editingArtifactID`, `labels`, `descriptions`, `fieldErrors`, `linkEdit` and `cardSaveError` into one card-edit value | not planned; logged in the review log (round 2 note) as a follow-up for the next card change |
| The unlink preflight also reporting the last-link block, so Swift stops restating `evidencegraphs.last_artifact` | not planned; logged in the review log (M-9) for when the preflight next changes |

## 4. Current state

Verified at `181dd41`.

| Area | Where | Today |
| --- | --- | --- |
| Add artifact model | `macos/App/Features/Sources/SourceArtifactsSection.swift` `Draft`, `openAdd()`, `canSubmitDraft`, `addSheetIsDirty`, `create()` | `create()` calls `context.store.createArtifact(…, graphIDs: [])` — every graph. `Draft` has no graph choice |
| Card edit model | same file, `beginEdit(_:)`, `fieldsDirty(_:)`, `canSaveFields(_:)`, `cancelFields(id:)`, `saveFields(id:)` | Label and description only; one `updateArtifact` call; a refusal goes to `context.pageError`, the callout at the top of the page (`SourcePageView`, `if let pageError = model.pageError`) |
| Add artifact form | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `addForm` | Label, Description, File (optional) |
| Card edit view | same file, `editFields(_:)` | Label, Description (`PVTextArea`), Save / Cancel |
| Dialog presentation | `macos/App/Features/Sources/SourcePageView.swift` `.pvFormDialog(isPresented:)` bound to `model.artifacts.isAdding` (line ~144, binding at ~315) | Bool-backed |
| Graph-list reads | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `evidenceGraph(id:)`, `evidenceGraphs(sourceID:)`, `linkedArtifacts(graphID:artifacts:)` | Linear `first` / `filter` over the list on every read |
| Session storage | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` (`@ObservationIgnored private var graphCameras`) | No graph index |
| Sources list grouping | `macos/App/Features/Sources/SourcesModel.swift` `graphsBySourceCache`, `graphsBySource` (lines ~173–181, read at ~151, ~157, ~319) | Its own revision-keyed cache of `SourcesListNavigation.graphsBySource(_:)` (one line, `Dictionary(grouping:by: \.sourceID)`; also called by `SourcesListNavigationTests`) |
| Canvas cite gate | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `canCite` | `sourceRow?.hasArtifact` — the Source's Artifacts, read per card in `EvidenceGraphView` hit targets and `cards` |
| Canvas Files | same file, `fileRows` | `EvidenceGraphFiles.fileBearing(workspace.artifacts, …)` — every Artifact of the Source |
| Dead code | same file, `refreshCanCite()` | No callers (`rg refreshCanCite macos` lists only its definition) |
| Canvas callout | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `noArtifactCallout`; `L10n.EvidenceGraph.noArtifactTitle` / `noArtifactMessage` | "No Artifacts on this source" / "Citing needs an Artifact…" |
| Composer callout | `macos/App/Features/CitationComposer/CitationComposerView.swift` `noArtifactGate`; `L10n.CitationComposer.noArtifactsCallout` | "…Add one from the Source page, then return here." |
| Load area | `macos/App/Features/Sources/SourceFormLoadArea.swift` `SourceFormLoadArea`, `SourceFormLoadState` | Three skeleton bars; one size |
| Locked checkbox | `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` `init(verbatim:description:isChecked:isIndeterminate:isDisabled:isLocked:)` | Exists (S12-12) |
| Citation hint | `PVDeleteImpactCopy.heading(_:)` (used by `SourceGraphFormModel.locked`) | "%lld citations in this graph use this artifact" |
| Store | `macos/App/Platform/GenealogyStore.swift` `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact`, `getEvidenceGraphUnlinkImpact`, `createArtifact(graphIDs:)` | All exist |
| Fake store | `macos/App/Platform/FakeStore.swift` `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact`; `getEvidenceGraphUnlinkImpact` with `unlinkImpactBlocks: [CatalogEvidenceGraphUnlinkBlock]`; `recordedCalls: [String]` | Link and unlink are not recorded and have no error hooks. The preflight ignores its `graphID` (`unlinkImpactBlocks.first { $0.artifactID == id }`), so one Artifact gets the same block in every graph. Used by `SourceGraphFormModelTests` (4 places) and `EvidenceGraphStoreTests` (1) |
| "Graph links an Artifact", graph form | `SourceGraphFormModel` `init` and `loadKeptLinks()`, each `graph.artifactIDs.isEmpty` | Two more inline copies |
| Write paths on the page | `macos/App/Features/Sources/SourcePageContext.swift` `applyWrite(_:)` = `session.apply(effects, .mutatedSourceWorkspace(sourceId:))` (reloads `GetSourceWorkspace`); `SourceGraphFormModel.save()` applies graph writes with `context.session.apply(effects)` | Graph-link writes take the second path |
| Checklist arithmetic | `macos/App/DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift` `enum PVArtifactChecklistState` (`liveExcluded`, `checkedIDs`, `checkedCount`, `isAllChecked`, `isAllIndeterminate`, `toggledAll`); callers `PVArtifactChecklist`, `SourceGraphFormModel.checkedArtifactIDs`, `SourceCopyModel`; `PVArtifactChecklist.maxRowsHeight = 264` | Id-generic, named for Artifacts |
| "Graph links an Artifact" | `SourcesListNavigation.graphBlockedReason(for:)` and `graphLocation(for:graph:)`, each `graph.artifactIDs.isEmpty` | Written out twice |
| Error codes | `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError.coded(status:code:kind:params:)`; `SourceGraphFormModel.save()` matches `if case .coded(_, "evidencegraphs.artifact_in_use", _, _)` | No code accessor |
| Canvas prepare | `EvidenceGraphModel.prepare()` awaits the graph, fields, rules, Sources list, workspace and counts, then `if !canCite { disarmCiteTools() }`; it does not wait for the graph list. `EvidenceGraphView` `.onChange(of: model.canCite)` only republishes hit targets | |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` (`createFilelessArtifact`, `cancelArtifactFieldsRestoresDrafts`, `editingArtifactIsOneCardAtATime`), `EvidenceGraphModelTests.swift` (`noArtifactDisablesArmingAndComposerLocation` seeds the Sources list, `resolvedFileFallsBackAndSelectFileClearsTheError`), `SourcesListNavigationTests.swift`, `SourceGraphFormModelTests.swift` | |

**Precedent.** S12-12 ([#557](https://github.com/mendahu/provenencia/pull/557)) `SourceGraphFormModel`: excluded-set draft, `loadKeptLinks()` with `SourceFormLoadState`, the `PVDeleteImpactCopy.heading` hint, and the re-read on `evidencegraphs.artifact_in_use`. Copy that shape onto the card. `SourcesModel.graphsBySourceCache` is the precedent for the revision-keyed, `@ObservationIgnored` index (D-10).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The group shows when `context.session.evidenceGraphs(sourceID:)` has a value with `count >= 2`. Hidden: Add sends `graphIDs: []`; a card save sends no link call | Working rule "Single-graph Sources hide link checkboxes"; `[]` means every graph (S12-03 D-11), which equals all checked | Showing a disabled single box — the rule says hide |
| D-2 | Both surfaces store the **unchecked** set, as the Artifact checklist does. Add: `Draft.excludedGraphIDs: Set<String>` *(new)*. When visible, `create()` sends `PVChecklistState.checkedIDs(ids: visibleGraphs.map(\.id), excluded: draft.excludedGraphIDs, locked: [])`, in list order, never `[]`. `PVArtifactChecklistState` is renamed `PVChecklistState` (same file, same members; its three callers and `PVArtifactChecklistTests` move to the new name) | One representation and one tested helper for every checklist; a graph made while the dialog is open starts checked (review M-2) | A `checked` set with a hand-written filter beside the helper |
| D-3 | `canSubmitDraft` is also false when the group is visible and no id is checked; `draftNeedsGraph` *(new)* drives the muted line, which `addForm` places under the group itself. `addSheetIsDirty` is also true when `PVChecklistState.liveExcluded(ids:excluded:locked: [])` is not empty | LK-4; `[]` would link every graph, the opposite of the choice | Letting Add send `[]` |
| D-4 | The retry path (`draft.createdID` set after a refused file) sends no link calls: the Artifact was linked when it was created | The first `createArtifact` already wrote the links; the dialog keeps the boxes but they are not re-applied | Re-linking on retry |
| D-5 | Card link state is one `linkEdit: SourceArtifactLinkEdit?` *(new struct)* on `SourceArtifactsSection`, set by `beginEdit` when D-1 shows the group and cleared by `cancelFields` / a successful save. It holds ids and choices only, never graph rows: `artifactID`, `graphIDs: [String]` (the `visibleGraphs` ids when edit opened, in the session's name order; a graph made later is not shown until edit reopens), `excluded: Set<String>` (unchecked graph ids, seeded with the graphs that do not link the Artifact), `locked: [String: String]` (graph id → hint), `loadState: SourceFormLoadState`. Rows, names, refs and current links are read from `context.session.evidenceGraph(id:)` (the D-10 index) each time. `linkedGraphIDs` *(computed)* = the `graphIDs` whose session graph links the Artifact; `checkedGraphIDs` = `PVChecklistState.checkedIDs(ids: graphIDs, excluded:, locked: Set(locked.keys))` | One card is in edit at a time; one source for a graph's links, the session the canvas and composer read (review M-10, the Known trap of a model copying `handle.value`) | A snapshot of the graph rows on the card — a second copy that goes stale after a write |
| D-6 | `loadLinkLocks()` *(new)* reads the session's graphs for `graphIDs`. Locks: a graph whose `artifactIDs == [artifactID]` is locked with `L10n.Sources.artifactGraphOnlyArtifact` locally, with no read. Every other graph that links the Artifact gets one `getEvidenceGraphUnlinkImpact(graphID:, artifactIDs: [artifactID])` call, in name order, sequentially; a returned block locks it with `PVDeleteImpactCopy.heading(block.citations)`, which wins over the only-artifact hint. No graph needing a read → `loadState = .ready` at once. `beginEdit` starts it as `linkLoadTask = Task { await loadLinkLocks() }` (`@ObservationIgnored var linkLoadTask: Task<Void, Never>?` *(new)*; tests `await section.linkLoadTask?.value`); Read again calls `loadLinkLocks()` directly. `loadLinkLocks()` captures `linkEdit.artifactID` before its first call and writes `locked` / `loadState` only while `linkEdit?.artifactID` still equals it, so a result for a cancelled or switched card is dropped | The preflight is per graph (S12-05 D-8); a census has a handful of graphs; sequential keeps the fake store's counters deterministic | A new batched RPC — out of scope (Swift only) |
| D-7 | `fieldsDirty(_:)` also returns true when `linkEdit`'s checked set differs from `linkedGraphIDs` (the session's current links); `canSaveFields(_:)` also requires `linkEdit == nil || linkEdit.loadState == .ready` | LK-9, LK-8 | A separate Save for links — product owner chose one Save |
| D-8 | `saveFields(id:)`: guard label not blank (as today); then return without a call when neither the fields nor the links changed (replacing today's fields-only early return, so a links-only change reaches the loop); `context.clearPageError()` stays; `savingID = id` covers every write below. Order: `updateArtifact` when the fields changed, applied through the existing `replace(_:effects:)` (`applyWrite`, the page payload moved); then for each id in `graphIDs`, `linkEvidenceGraphArtifact` where checked and not in `linkedGraphIDs`, then `unlinkEvidenceGraphArtifact` where unchecked and in `linkedGraphIDs`, both sets read once before the loop. Each link and unlink's effects go through `context.session.apply(effects)`, as `SourceGraphFormModel.save()` applies graph writes, not `applyWrite`: a link does not change the page payload (review M-1). After the loop, on success or error, when any link or unlink was attempted, the save awaits `_ = await context.session.readyValue(context.session.evidenceGraphsKey) as [CatalogEvidenceGraph]?` once (`session.apply` only starts the reload; review X-12, M-10), so the session's links are current before anything reads them. Success: the card leaves edit. The first error stops the loop, sets `cardSaveError = L10n.Sources.artifactCardSaveErrorMessage(L10n.Errors.message(for:))` *(new, `String?` on `SourceArtifactsSection`, cleared by `beginEdit`, `cancelFields` and a successful save)*, keeps the card in edit and `excluded` as the researcher left it, and, after the await, runs `loadLinkLocks()` again, so a partial save, an in-use Artifact and a last link all show from the reloaded list; Save again then sends only what is left. If `readyValue` returns nil (the reload failed), `loadState = .failed(…)` with the graph list's error, so Read again is offered and Save waits. `saveFields` no longer writes `context.pageError`, on every Source, one graph or many (the board moved the refusal onto the card). `CoreInvokeError.code` *(new, `var code: String?`, nil unless `.coded`)* replaces the code match in `SourceGraphFormModel.save()` (review M-6) | Links before unlinks so an Artifact never passes through zero graphs needlessly; the engine has no multi-graph link write; the re-read mirrors S12-12's save | One `updateEvidenceGraph` per graph with the full set — writes every other Artifact's link row again and races other edits; a local copy of the links patched per write — a second source (M-10) |
| D-9 | `CatalogEvidenceGraph.hasArtifact` *(new, `!artifactIDs.isEmpty`)* in `GenealogyStore.swift` beside `linkedArtifacts(_:)`; `SourcesListNavigation.graphBlockedReason(for:)`, `graphLocation(for:graph:)` and the two inline checks in `SourceGraphFormModel` (`init`'s `loadState` and `loadKeptLinks()`'s guard) read it (review M-3). `EvidenceGraphModel.canCite`: `guard let graph = session.evidenceGraph(id: graphID) else { return session.evidenceGraphsHandle?.value == nil }` then `graph.hasArtifact`. `fileRows` keeps its `status == .ready` workspace guard and passes `session.linkedArtifacts(graphID:artifacts: workspace.artifacts) ?? []` to `EvidenceGraphFiles.fileBearing`. `prepare()` adds `let _: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)` and `_ = await session.readyValue(session.evidenceGraphsKey) as [CatalogEvidenceGraph]?` before its `canCite` check, so a graph with no links disarms the tools on open. `refreshCanCite()` is deleted | Map scope (moved from S12-10); a graph missing from a loaded list cannot be cited | Keeping `sourceRow.hasArtifact` — wrong on a multi-graph Source |
| D-10 | `WorkspaceSession` gains `@ObservationIgnored private var evidenceGraphIndex: EvidenceGraphIndex?` *(new)*, where `struct EvidenceGraphIndex { let revision: Int; let byID: [String: CatalogEvidenceGraph]; let bySource: [String: [CatalogEvidenceGraph]] }` lives in `WorkspaceSession+EvidenceGraphs.swift`, built from `evidenceGraphsHandle.value` when `revision` differs (`bySource` = `Dictionary(grouping:by: \.sourceID)` in the builder; `SourcesListNavigation.graphsBySource(_:)` is deleted, its tests build the map the same way, and `graphsBySourceKeepsListOrder` moves into T-1, review M-5). `evidenceGraph(id:)`, `evidenceGraphs(sourceID:)` (`bySource[id] ?? []`) and a new `evidenceGraphsBySource: [String: [CatalogEvidenceGraph]]` (empty while no value) read it. `SourcesModel` passes `session.evidenceGraphsBySource` at its three reads; its private `graphsBySource` and `graphsBySourceCache` are deleted | `canCite` is read per card; S12-10 PRF-1 and S12-11 D-2 sent the index here; one cache instead of two | A second cache in the canvas model |
| D-11 | Add artifact stays on `pvFormDialog(isPresented:)`; this PR adds no dialog. The map scope sentence "Its dialogs present through the kit's `pvFormDialog(item:)`" is replaced in the map by this decision | The draft lives on the section; there is no model object to make the `item` | Wrapping the draft in a throwaway model |
| D-12 | `SourceArtifactGraphLinks` *(new view)* in `Features/Sources/SourceArtifactGraphLinks.swift` *(new)* takes `graphs`, `excluded: Binding<Set<String>>`, `locked: [String: String]` and `accessibilityIdentifierPrefix`; rows are `PVCheckbox(verbatim: graph.name, description: locked[id] ?? graph.ref, isChecked:, isLocked: locked[id] != nil)` with a locked row's `accessibilityValue` as `PVArtifactChecklist` sets it; the frame copies `PVArtifactChecklist`'s (clipped `RoundedRectangle(cornerRadius: PVRadius.sm)`, `PVColor.borderSubtle` stroke, `PVDivider` between rows, `ScrollView` capped at the same 264pt, about six rows; the board's 280 is rounded to the kit's one value). Add binds `$draft.excludedGraphIDs`; the card binds `linkEdit.excluded`. The form dialog body does not scroll; the 264pt cap keeps a twelve-graph Add dialog in a normal window | One view for both call sites (brief inventory); one checklist frame size on the page (review M-7) | `PVArtifactChecklist` — Artifacts with thumbnails; a `showsNeedsGraph` Bool |
| D-13 | `SourceFormLoadArea` gains `var compact = false`; compact draws one skeleton bar at 0.48 width, 12pt high, and the caption. The graph form and Copy form keep the default | Brief inventory "Extend"; one loading body | A local loading line on the card |
| D-14 | Accessibility ids: `sources.page.addArtifact.graphs` (group), `.graphs.<graph id>` (row), `.graphs.needsGraph`; card: `sources.page.artifact.<id>.graphs`, `.graphs.<graph id>`, `.graphs.loading`, `.graphs.loadError`, `.graphs.readAgain` (the last three from `SourceFormLoadArea`'s prefix) | Match the existing `sources.page.…` ids | |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourceArtifactGraphLinks` *(new)* | view | `PVArtifactChecklist` (its frame is the near-copy: border, dividers, 264pt scroll) | Rows are graphs (name, ref), no thumbnail, no All row; searched `PVCheckbox(` call sites in `Features/`. A third bordered checklist extracts the frame into the kit |
| `SourceArtifactsSection.visibleGraphs` *(new)* | computed | `SourceGraphsSection.rows` | The one home of the working rule "Single-graph Sources hide link checkboxes": the Source's graphs when there are two or more, else nil |
| `SourceArtifactsSection.cardSaveError`, `linkLoadTask` *(new)* | stored | `SourceGraphFormModel.saveError` | The card's own refusal and its preflight task |
| `CatalogEvidenceGraph.hasArtifact` *(new)* | computed | `CatalogSource.hasArtifact` | Same name at graph level; replaces two inline checks |
| `CoreInvokeError.code` *(new)* | computed | the `if case .coded(_, "…", _, _)` matches | One accessor for both models |
| `PVChecklistState` | rename | `PVArtifactChecklistState` | Renamed, not copied |
| `SourceArtifactLinkEdit` *(new)* | struct | `SourceGraphFormDraft` / `SourceGraphFormModel` | That is a whole form with its own write; the card needs only ids and choices, inside the existing card edit; graph rows stay in the session |
| `SourceArtifactsSection.loadLinkLocks()` *(new)* | func | `SourceGraphFormModel.loadKeptLinks()` | Per-graph calls for one Artifact vs one call for one graph's Artifacts |
| `Draft.excludedGraphIDs`, `draftNeedsGraph` *(new)* | field / computed | `SourceGraphFormDraft.excludedArtifactIDs`, `needsArtifact` | Same shape, different object |
| `EvidenceGraphIndex`, `evidenceGraphIndex`, `evidenceGraphsBySource` *(new)* | struct / stored / accessor | `SourcesModel.graphsBySourceCache` | Moved here and deleted there |
| `SourceFormLoadArea.compact` *(new)* | parameter | `SourceFormLoadArea` | Extended, not copied |
| `FakeStore` link / unlink entries in `recordedCalls` (`"linkEvidenceGraphArtifact graph=<id> artifact=<id>"`, `"unlinkEvidenceGraphArtifact graph=<id> artifact=<id>"`), `linkEvidenceGraphArtifactError`, `unlinkEvidenceGraphArtifactError` *(new)*, `getEvidenceGraphUnlinkImpactGraphIDs: [String]` *(new)*, and `unlinkImpactBlocksByGraph: [String: [CatalogEvidenceGraphUnlinkBlock]]` replacing `unlinkImpactBlocks` | test hooks | `recordedCalls` (ordered write log), the `…Error` / `…Calls` hooks beside them | Writes join the existing ordered log so T-8 can assert link-before-unlink (review M-4); the preflight returns `unlinkImpactBlocksByGraph[graphID]` filtered to the asked ids, and the five existing `unlinkImpactBlocks` uses move to the graph they test (review X-1) |
| L10n keys in §6 Strings *(new)* | keys | `L10n.Sources.graphForm…` | New copy |

**Builds on.** The session's graph-list accessors (S12-11 D-2); `SourceFormLoadState` (S12-12); `PVCheckbox` `isLocked` (S12-12); `PVChecklistState`; `PVDeleteImpactCopy.heading`; the write effects paths (`context.applyWrite` for the Artifact row, `session.apply` for graph links, as `add-workspace-place` §6 says for graph writes).
**Deletes.** `EvidenceGraphModel.refreshCanCite()`; `SourcesModel.graphsBySourceCache` and its private `graphsBySource`; `SourcesListNavigation.graphsBySource(_:)`; `FakeStore.unlinkImpactBlocks`; the name `PVArtifactChecklistState`; the old `L10n.EvidenceGraph.noArtifactTitle` / `noArtifactMessage` / `noArtifactAction` values and `L10n.CitationComposer.noArtifactsCallout` value (replaced in place, same keys, `Localizable.xcstrings` updated); `SourceFormLoadArea`'s doc comment (three skeleton bars, one body for both forms) is rewritten for the compact size and a third caller.
**Next adds.** S12-14 reads `evidenceGraphsBySource` for the accordion, gates per-graph Open on `CatalogEvidenceGraph.hasArtifact`, and deletes `entryGraph`; S12-15 opens a graph by id through `evidenceGraph(id:)`.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None — every RPC exists (S12-05).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Graph group | `PVField` + `SourceArtifactGraphLinks` *(new)* | `Features/Sources/SourceArtifactGraphLinks.swift` *(new)* | D-12 |
| Graph row | `PVCheckbox(verbatim:description:isChecked:isLocked:)` | kit | ref or hint as description |
| Add placement | `SourcePageArtifactsView.addForm` | between Description and File (optional) | LK-3 |
| Card placement | `SourcePageArtifactsView.editFields` | between Description and the buttons | LK-5 |
| Card loading / error | `SourceFormLoadArea(compact: true)` | `SourceFormLoadArea.swift` | LK-8 |
| Save refused | `PVCallout(tone: .danger, title: L10n.Sources.artifactCardSaveErrorTitle, message: cardSaveError, compact: true)` | `SourcePageArtifactsView.editFields`, first item, above Label | LK-9; id `sources.page.artifact.<id>.saveError` |
| Canvas callout | `noArtifactCallout` | `EvidenceGraphView.swift` | copy only |
| Composer callout | `noArtifactGate` | `CitationComposerView.swift` | copy only |

States: graph list loading → group hidden (D-1); preflight loading / failed (LK-8); twelve graphs → scrolls. Accessibility: LK-10 — `PVCheckbox` already speaks locked; the group `PVField` label is the container label.

**Frames.** From the agreed S12-D4 board. Census fixture: Pages 1–6, 8 and 9 (`ART-0401` to `ART-0409`, no `ART-0407`); Brown household (`GPH-BR6KD`, links Page 9 only), Jones household (`GPH-JONES`, Pages 3, 4, 5, 9), Smith household (`GPH-7SM2T`, Pages 3 and 6; two Citations on Page 3).

1. **Add, one graph** — today's Add artifact dialog: Label, Description, File (optional). No group.
2. **Add, census** — label "Page 7"; between Description and File, a `PVField` "Evidence graphs" with the hint "Which evidence graphs can cite this artifact" over a bordered list: Brown household / `GPH-BR6KD`, Jones household / `GPH-JONES`, Smith household / `GPH-7SM2T`, each a checked kit checkbox with the ref as its description line.
3. **Add, Jones only** — Brown and Smith unchecked; Add enabled.
4. **Add, nothing checked** — all three unchecked; the muted italic caption "Check at least one evidence graph." under the field; **Add artifact** disabled.
5. **Edit, one graph** — today's card edit on Page 3: Label, Description, Save artifact / Cancel. No group.
6. **Edit, loading** — Page 3's card in edit: Label and Description filled and editable; under Description one skeleton bar (about half the card width) and the italic caption "Reading the evidence graphs…"; Save artifact disabled.
7. **Edit, ready** — Page 3: Brown unchecked with `GPH-BR6KD`; Jones checked with `GPH-JONES`; Smith checked and locked (lock mark, not clickable) with "2 citations in this graph use this artifact" in place of its ref. Save artifact disabled until something changes.
8. **Edit, only artifact** — Page 9: Brown checked and locked with "The only artifact in this evidence graph"; Jones checked and clickable; Smith unchecked.
9. **Read error** — in the group's place, the compact danger callout "Couldn’t read the evidence graphs" / "Label and description are kept. Read the evidence graphs again to change the links." with the error and a **Read again** secondary button; Save artifact disabled.
10. **Save refused** — Page 3 with a changed description and Jones unchecked: a compact danger callout as the first item of the card's edit fields, above Label, "Couldn’t save the artifact" with the body and the error (copy in §6 Strings); every field and box keeps its value.
11. **Twelve, card** — twelve households on Page 3's card; the list scrolls inside its border (264pt, D-12); Smith is locked.
12. **Twelve, dialog** — the same twelve in Add artifact; the list scrolls and the dialog's buttons stay in view.
13. **Canvas and composer** — the canvas callout (warning) "No artifacts in this evidence graph" / LK-11 body / **Go to Source page**; the composer callout with the LK-12 body, title and button unchanged.

The board draws the locked row with a stand-in because the web kit's checkbox has no lock; the app uses `PVCheckbox` `isLocked`, as S12-12 did.

### Strings

L10n keys *(new)*, in `L10n.Sources`, final copy as agreed on the board. Catalog keys are `sources.page.artifact.graphs.<name>` (`.label`, `.hint`, `.needsOne`, `.onlyArtifact`, `.loading`, `.loadErrorTitle`, `.loadErrorMessage`) and `sources.page.artifact.saveErrorTitle` / `.saveErrorMessage`, each with a translator comment saying where it shows. The two messages are `static func …(_ error: String) -> String` through `L10n.format`, as `graphFormSaveErrorMessage` is (its trailing space before an empty error is accepted there too):

- `artifactGraphsLabel` = "Evidence graphs"
- `artifactGraphsHint` = "Which evidence graphs can cite this artifact"
- `artifactGraphsNeedsOne` = "Check at least one evidence graph."
- `artifactGraphOnlyArtifact` = "The only artifact in this evidence graph"
- `artifactGraphsLoading` = "Reading the evidence graphs…"
- `artifactGraphsLoadErrorTitle` = "Couldn’t read the evidence graphs"
- `artifactGraphsLoadErrorMessage` = "Label and description are kept. Read the evidence graphs again to change the links. %@"
- `artifactCardSaveErrorTitle` = "Couldn’t save the artifact"
- `artifactCardSaveErrorMessage` = "Some changes may not be saved. Every field and box keeps its value — save again to finish. %@" (the board's "Nothing was changed" body reworded, user 2026-10-25: a save of several writes can stop partway)

Changed values (same keys): `L10n.EvidenceGraph.noArtifactTitle` = "No artifacts in this evidence graph"; `L10n.EvidenceGraph.noArtifactMessage` = "Citing needs an artifact linked to this evidence graph. Place tools, Connect, and Add property stay disabled until one is linked."; `L10n.EvidenceGraph.noArtifactAction` = "Go to Source page"; `L10n.CitationComposer.noArtifactsCallout` = "Every citation needs an artifact linked to this evidence graph. Add one on the Source page, or link an existing one from its card, then return here."

## 7. Steps

**STEP-1 — Graph index in the session**
Files: `WorkspaceSession.swift`, `WorkspaceSession+EvidenceGraphs.swift`, `SourcesModel.swift`, `SourcesListNavigation.swift`, `GenealogyStore.swift` (`hasArtifact`), `macos/ProvenenciaTests/SourcesListNavigationTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-10 and the `hasArtifact` half of D-9.
Check: T-1 passes; `rg 'graphsBySourceCache|SourcesListNavigation.graphsBySource' macos` prints nothing; `SourcesModelTests`, `SourcesListNavigationTests` pass.

**STEP-2 — Fake store hooks**
Files: `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`, `EvidenceGraphStoreTests.swift`, `macos/App/Platform/CoreInvoke.swift`, `SourceGraphFormModel.swift`.
Do: the FakeStore row in Codebase fit (each error hook throws before mutating and before recording); `CoreInvokeError.code`, and `SourceGraphFormModel.save()` reads it.
Check: `SourceGraphFormModelTests` and `EvidenceGraphStoreTests` pass on `unlinkImpactBlocksByGraph`.

**STEP-3 — Add artifact model**
Files: `SourceArtifactsSection.swift`, `PVArtifactChecklist.swift` and its callers (`SourceGraphFormModel.swift`, `SourceCopyModel.swift`, `macos/ProvenenciaTests/PVArtifactChecklistTests.swift`, the ArtifactChecklist row in `macos/App/DesignSystem/README.md`) for the rename. Skill: `add-swift-test`.
Do: D-1 – D-4; `visibleGraphs: [CatalogEvidenceGraph]?` (nil when hidden; the session's order) read by Add and edit.
Check: T-2 – T-5 pass.

**STEP-4 — Card edit model**
Files: `SourceArtifactsSection.swift`.
Do: D-5 – D-8; `cancelFields` clears `linkEdit` and `cardSaveError`.
Check: T-6 – T-11 pass; `cancelArtifactFieldsRestoresDrafts`, `editingArtifactIsOneCardAtATime` pass unchanged.

**STEP-5 — Views and strings**
Files: `SourceArtifactGraphLinks.swift` *(new)*, `SourcePageArtifactsView.swift`, `SourceFormLoadArea.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-12 – D-14 and the new keys; the save-refused callout; the `SourceFormLoadArea` doc comment; match §6 Frames.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the app builds; the manual check steps 1–4.

**STEP-6 — Canvas and composer follow the graph**
Files: `EvidenceGraphModel.swift`, `EvidenceGraphView.swift` (copy only through L10n), `L10n.swift`, `Localizable.xcstrings`.
Do: D-9; the changed values in §6 Strings.
Check: T-12 – T-14 pass; `noArtifactDisablesArmingAndComposerLocation` moved onto a graph list seed passes; `rg refreshCanCite macos` prints nothing.

**STEP-7 — Gate**
Do: `scripts/check-pr.sh`.
Check: prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `completed.md`, `deployment-plan.md` (S12-13 row → Landed), this plan (status, Landed note), `design/README.md` (S12-D4 archived line names this PR), `.agents/skills/add-workspace-place/SKILL.md` §6 (name graph link writes, `link…` / `unlink…` / `updateEvidenceGraph`, under the `session.apply(effects)` bullet). The brief is already in `design/archive/` (moved when the board was agreed).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Link a new artifact to chosen evidence graphs."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `WorkspaceSessionTests.swift` `evidenceGraphIndexFollowsTheListRevision` *(new)* | seed list [g1 (s1), g2 (s1), g3 (s2)] → `evidenceGraph(id: "g2")`, `evidenceGraphs(sourceID: "s1")` = [g1, g2]; set a new value without g2 → `evidenceGraph(id: "g2") == nil`; no value → `evidenceGraphs(sourceID:) == nil`, `evidenceGraphsBySource` empty | AC-4.1, AC-4.2, D-10 |
| T-2 | Swift | `SourcePageModelTests.swift` `addArtifactOnOneGraphSendsEveryGraph` *(new)* | one graph → no group (`visibleGraphs == nil`); create → `graphIDs == []` | AC-1.1, AC-1.6 |
| T-3 | Swift | same, `addArtifactLinksOnlyTheCheckedGraphs` *(new)* | three graphs, exclude g-b, g-s → `graphIDs == ["g-j"]` | AC-1.2, AC-1.3, D-2 |
| T-4 | Swift | same, `addArtifactNeedsAGraph` *(new)* | exclude all → `canSubmitDraft == false`, `draftNeedsGraph == true`; create → no store call | AC-1.4, D-3 |
| T-5 | Swift | same, `addArtifactExcludedGraphIsDirty` *(new)* | exclude g-b only → `addSheetIsDirty` | AC-1.5 |
| T-6 | Swift | same, `cardEditOnOneGraphHasNoLinks` *(new)* | one graph; `beginEdit` → `linkEdit == nil`, `getEvidenceGraphUnlinkImpactCalls == 0` | AC-2.1 |
| T-7 | Swift | same, `cardEditLocksInUseAndOnlyArtifactGraphs` *(new)* | P3 in g-j, g-s; `unlinkImpactBlocksByGraph = ["g-s": [block]]`, block built as `SourceGraphFormModelTests.block` does (`via: "evidencegraphs.artifact_in_use"`, `kind: "citation"`, `total: 2`); `await linkLoadTask?.value` → P3: `locked["g-s"]` = "2 citations in this graph use this artifact", `locked["g-j"] == nil`, `getEvidenceGraphUnlinkImpactGraphIDs == ["g-j", "g-s"]`; P9 only in g-b → `locked["g-b"]` = "The only artifact in this evidence graph", no call for g-b | AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-8 | Swift | same, `cardSaveLinksThenUnlinks` *(new)* | check g-b, uncheck g-j → `recordedCalls` ends with `["linkEvidenceGraphArtifact graph=g-b artifact=P3", "unlinkEvidenceGraphArtifact graph=g-j artifact=P3"]` in that order, no `updateArtifact`, `editingArtifactID == nil` | AC-2.5, AC-2.10, D-8 |
| T-9 | Swift | same, `cardSaveFieldsOnlyHasNoLinkCalls` *(new)* | label change → `updateArtifact` once, no link calls | AC-2.6 |
| T-10 | Swift | same, `cardSaveRefusedKeepsEditAndRereadsLocks` *(new)* | check g-b, uncheck g-j; `unlinkEvidenceGraphArtifactError` = coded `evidencegraphs.artifact_in_use` → still editing, `cardSaveError` set, `context.pageError == nil`, `excluded` unchanged, the link to g-b recorded before the refusal; after the save returns (it awaited the graph list), `linkedGraphIDs` contains g-b, impact calls incremented; Save again records only the unlink | AC-2.7, D-8 |
| T-11 | Swift | same, `cardLinkReadFailsAndReadAgain` *(new)*, `cardCancelDropsLinkChanges` *(new)* | impact error → `loadState` `.failed`, `canSaveFields == false`; clear error, `loadLinkLocks()` → `.ready`. Uncheck g-j, cancel → no calls, reopen → g-j checked | AC-2.8, AC-2.9 |
| T-12 | Swift | `EvidenceGraphModelTests.swift` `fileRowsAreTheGraphsLinkedArtifacts` *(new)* | graph links P3 of P1–P3 → `fileRows.map(\.id) == ["P3"]` | AC-3.1, D-9 |
| T-13 | Swift | same, `canCiteNeedsALinkedArtifact` *(new)* | graph with `artifactIDs: []` → false, and after `prepare()` an armed tool is disarmed; no list value → true | AC-3.2, AC-3.3 |
| T-14 | Swift | `L10n` lookups in T-13 and `CitationComposerModelTests.swift` | callout strings equal §6 Strings | AC-3.2, AC-3.4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceGraphModelTests` (canCite seeds the Sources list today), `SourcePageModelTests`, `SourcesModelTests`, `SourceGraphFormModelTests` (load area).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] design/README.md: the S12-D4 archived line names this PR (the brief moved to design/archive/ at board agreement)
- [ ] Domain doc updated if a rule in docs/*.md changed: none — the link rules in source-structure.md are unchanged
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Set up once: a new project; Add Source "1881 census"; add Artifacts "Page 1", "Page 2", "Page 3" (any image files). The Source has one graph, Default, linked to all three.

1. Edit Page 1's card and open Add artifact: no Evidence graphs group in either (AC-1.1, AC-2.1). Cancel both.
2. In Evidence graphs, Add evidence graph "Jones household" with only Page 3 checked, and "Smith household" with only Page 2 checked. Open Add artifact: the group lists Default, Jones household, Smith household, all checked (AC-1.2). Uncheck all three: Add is disabled with "Check at least one evidence graph." (AC-1.4). Check only Jones; label "Page 7"; Add (AC-1.3).
3. Open Smith household's graph: the Files menu lists only Page 2. Open Jones household's: Page 3 and Page 7 (AC-3.1).
4. In Jones's Citation composer, cite Page 3 once. Back on the Source page, Edit Page 3's card: Default checked, Jones checked and locked with "1 citation in this graph uses this artifact", Smith unchecked (AC-2.3).
5. Edit Page 2's card: Smith is checked and locked with "The only artifact in this evidence graph" (AC-2.4).
6. On Page 3's card, check Smith and uncheck Default; Save. Each graph's Edit form shows the new links (AC-2.5).
7. Add evidence graph "Empty test" with only Page 1 checked, then on Page 1's card confirm "Empty test" is locked as its only artifact. Open "Empty test" (its canvas cites), return to the Source page and delete Page 1 (uncited, so the delete is allowed and leaves "Empty test" with no links). Press Back to "Empty test"'s canvas: the callout reads "No artifacts in this evidence graph" and the cite tools are disabled (AC-3.2). The composer's LK-12 body is checked by T-14 (AC-3.4): the composer cannot be opened on a graph with no links.
8. VoiceOver on Page 3's card group: each row reads its name, state, ref or hint, and "locked" when locked (LK-10).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A save that fails halfway leaves some links written | D-8 diffs against the session at save time, so Save again sends only the rest; the error stays on the page |
| `canCite` per card on a big graph list | D-10 index; one dictionary read |
| The graph list refreshes during card edit and adds a graph | `linkEdit.graphIDs` is fixed when edit opens; the new graph is not shown until edit reopens; the save diff covers only `graphIDs` |
| A twelve-graph Add dialog is taller than the window | D-12's 264pt cap; Stop condition below |

**Stop and report if:**

- `createArtifact(graphIDs:)` with explicit ids does not link only those graphs (the S12-03 D-11 rule) in `EvidenceGraphStoreTests` or the Go handler tests.
- `getEvidenceGraphUnlinkImpact` refuses a single Artifact id for a graph.
- `QueryHandle.revision` does not change when the graph list's value changes, so D-10 cannot key on it.
- A test outside §8's "most likely to break" list fails because `canCite` changed meaning.
- The Add artifact dialog with twelve graphs is cut off at the default window size (the dialog body does not scroll).
- The agreed board draws something §5 or the inventory does not cover (a resting-card link line, a Citation list on the card, an All row).

## 12. Open questions

None.

## 13. PR

**Title:** S12-13 — Artifact graph checkboxes
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-13-artifact-graph-checkboxes.md`.
