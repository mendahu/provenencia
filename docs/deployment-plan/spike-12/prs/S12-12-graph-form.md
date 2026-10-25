# S12-12 — Graph form

**Status:** Blocked — on board S12-D3
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D3](../design/S12-D3-graph-form.md), board not yet agreed
**Board:** not yet drawn. When it is agreed, the planner checks every frame against §6 and fills §6 Frames in words for an implementer who cannot open the link, as S12-11 did
**Depends on:** S12-02 (landed [#485](https://github.com/mendahu/provenencia/pull/485)), S12-11 (landed [#549](https://github.com/mendahu/provenencia/pull/549)), S12-D3
**Blocks:** S12-13
**Branch:** `s12-12-graph-form`, branched from `main` after the plan branch merges (so STEP-9 edits this file in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-11 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md) (the checklist's promotion to a recipe), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`
**Context:** [rules](../source-structure.md#rules), [where graphs show](../source-structure.md#where-graphs-show); working rules: "Artifact ↔ graph links", "One title and type form, one Artifact checklist", "Name and description are editable; graphs list alphabetically by name", "Refusals are of two kinds", "One graph list", "The graph list has its own effect", "High tolerance for churn"; the S12-04 plan D-7 (`Update` is one write), the S12-05 plan D-4 / D-8 / D-13 (the unlink preflight and its heading), the S12-11 [Landed note](S12-11-source-page-graphs.md) and [ship log](S12-11-source-page-graphs.ship.md) (UX-3)
**Planned:** 2026-10-25 against `main` @ `55611a1`
**Review:** [`S12-12-graph-form.review.md`](S12-12-graph-form.review.md) — round 2 pending

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

The researcher creates and edits Evidence graphs on the Source page. **Add evidence graph** in the Evidence graphs section header and **Edit** on each row open one form: name, description, and the Source's Artifacts as a checklist, all checked for a new graph and the linked ones for an edit. An Artifact the graph's Citations use is shown checked and locked, with the Citations listed, so an edit cannot hit the unlink refusal. This makes the dogfood bar's census case reachable in the app: "A census with three graphs", each with its own name, description and pages ([dogfood bar](../deployment-plan.md#goal-dogfood-bar) outcomes 2 and 8).

**Why now.** S12-04 and S12-05 put create, one-write update and the unlink preflight on the wire; S12-11 drew the section this form hangs off. S12-13 (Artifact graph checkboxes) needs a second graph to exist before its checkboxes ever show.

**What stays.** The engine, FFI and store protocol are unchanged. The section, its rows, Open and Delete are as S12-11 built them. The Copy Source form looks and behaves the same; its checklist moves to a shared recipe. On a one-graph Source nothing changes but the two new buttons.

## 2. User stories

**US-1.** As a researcher on a census Source, I add an Evidence graph for one household and link only the pages it is on, so that its composer offers only those pages.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with Artifacts P1 – P6; Add evidence graph | The form opens titled "Add evidence graph", name empty, description empty, all six checked, heading "6 of 6" |
| AC-1.2 | Name "Jones household", description "Household 14, page 3", only P3 checked; Add | `createEvidenceGraph` is called once with that name, description and `[P3]`; the dialog closes; the section lists the new graph in name order; the toast reads "Evidence graph added" / "{ref} is ready to open." |
| AC-1.3 | Name blank (or spaces); Add | No store call; the name field shows "Give the evidence graph a name."; the form stays open |
| AC-1.4 | Every Artifact unchecked | Add is disabled; "Check at least one artifact." shows under the checklist |
| AC-1.5 | A Source with no Artifacts | Add evidence graph is disabled; tooltip and VoiceOver hint "Add an artifact before you add an evidence graph" |
| AC-1.6 | `createEvidenceGraph` throws | The form stays open with the danger callout "Couldn’t save the evidence graph" and the error's message; name, description and checks are unchanged |

**US-2.** As a researcher, I rename a graph and change its pages in one save, and I can see which pages it must keep and why.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Edit on "Smith household" (`GPH-SMITH`) linked to P3 and P4 | The form opens titled "Edit evidence graph", subtitle "GPH-SMITH", name and description filled, P3 and P4 checked, the others unchecked |
| AC-2.2 | Before the preflight returns | The checklist area shows the loading state; Save is disabled; name and description are editable |
| AC-2.3 | The preflight returns one block for P3 with 2 Citations | P3 is checked and disabled with the hint "2 citations in this graph use this artifact"; the Kept linked group lists both Citations under P3's label; unchecking All Artifacts leaves P3 checked and All reads mixed |
| AC-2.4 | Rename to "Smith family", uncheck P4, check P5; Save | `updateEvidenceGraph` is called once with "Smith family", the description, and `[P3, P5]` in Source page order; the dialog closes; the row reads "Smith family"; no toast |
| AC-2.5 | The preflight throws | A danger callout "Couldn’t read the artifacts" with **Read again** replaces the checklist; Save is disabled; Read again repeats the preflight |
| AC-2.6 | `updateEvidenceGraph` throws (`evidencegraphs.artifact_in_use` in a race) | The form stays open with the save-error callout carrying that code's `L10n.Errors` copy; every choice kept |
| AC-2.7 | Edit Default on a Source with no Artifacts; rename to "Whole census"; Save | The checklist area shows "This Source has no Artifacts."; Save is enabled; `updateEvidenceGraph` gets `artifactIDs: []` (links unchanged) |
| AC-2.8 | A Citation row in the Kept linked group is clicked | The form closes without saving and the Citation's location opens |

**US-3.** As a researcher deleting the Source's only graph, the notice tells me the control to use.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Delete on the only graph | The notice body reads "A source keeps at least one evidence graph. Use Add evidence graph to make another, then delete this one. A new evidence graph needs an artifact." |

**US-4.** As a researcher copying a Source, the Copy form's Artifact checklist works as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The Copy Source form | Same heading, All Artifacts, rows, count and accessibility ids (`sources.copy.artifacts…`) as before; `SourceCopyModelTests` pass unchanged |

## 3. Scope

| In | Covered by |
| --- | --- |
| The Artifact checklist promoted to `PVArtifactChecklist` with locked rows (brief inventory) | STEP-1, AC-4.1, AC-2.3 |
| The DeleteImpact group's rows as a recipe view shared by the notice and the form (GF-8) | STEP-2, AC-2.3, AC-2.8 |
| One load and read-error area for the Copy and graph forms (D-14) | STEP-2, AC-2.5, AC-4.1 |
| FakeStore call counters and error hooks for create, update and the preflight | STEP-3 |
| Graph form model: create and edit modes, validation, preflight, one save (GF-4 – GF-11, GF-13) | STEP-4, AC-1.1 – AC-1.6, AC-2.1 – AC-2.7 |
| The form dialog view (GF-4 – GF-11, GF-14, GF-15) | STEP-5, AC-2.8, manual check |
| Add evidence graph in the section header, Edit on each row (GF-1 – GF-3) | STEP-6, AC-1.5 |
| The `last_graph` notice body (GF-12; S12-11 review UX-3) | STEP-7, AC-3.1 |

| Out | Home |
| --- | --- |
| Graph checkboxes in the Add Artifact dialog and on Artifact cards; the canvas's cite gate and Files list reading linked Artifacts; an id index behind the session's graph accessors | S12-13 (board S12-D4) |
| Per-graph progress; the Sources list per-graph Open | S12-14 (board S12-D5) |
| Graphs in the omnibar | S12-15 |
| A toast after an edit | not planned (brief GF-13) |
| Opening the new graph after Add | not planned — user 2026-10-25: the researcher stays on the Source page |
| Two `last_graph` bodies chosen by whether the Source has Artifacts | not planned — user 2026-10-25: one body for both cases |
| Showing or editing a graph's ref in the form | not planned (brief §2.1) |
| Moving Subjects between graphs, splitting or merging graphs | not planned — graphs are fully isolated ([README](../README.md#not-this-spike)) |
| A `pvFormDialog(item:)` presenter so model-backed forms stop copying `sourceCopyDialog`'s binding (review M-5) | S12-13 (its scope entry names it, written by this plan's branch) |
| Asking "Discard changes?" before a Kept linked row navigates | not planned — the row closes the form as the delete notice's rows close the notice (D-10) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Checklist | `macos/App/Features/Sources/SourceArtifactChecklist.swift` `SourceArtifactChecklist` | A snowflake: `artifacts`, `@Binding excluded: Set<String>`, `projectDir`, `sourceTypeIconKey`, `isDisabled`. State is the unchecked set, so an Artifact that appears later starts checked. Header text `L10n.Sources.copySourceArtifactsHeading` and count `copySourceArtifactCount(_:_:)` ("%1$lld of %2$lld"); `allRow` sets `excluded = on ? [] : all ids`; rows `PVCheckbox(verbatim:isChecked:isDisabled:)` (no description) or box-only with `accessibilitySpokenLabel`; empty line `copySourceNoArtifacts`; ids hard-coded `sources.copy.artifacts`, `.all`, `.empty`, `.<artifact id>`. Doc comment: "A snowflake until the Graph form reuses it (promote then)". One caller: `SourceCopyDialog.swift` `SourceCopyForm.parts` |
| Copy strings | `macos/App/Platform/L10n.swift` `L10n.Sources.copySourceArtifactsHeading` (`sources.copy.artifactsHeading` "Artifacts"), `copySourceAllArtifacts` (`sources.copy.allArtifacts` "All Artifacts"), `copySourceNoArtifacts` (`sources.copy.noArtifacts` "This Source has no Artifacts."), `copySourceArtifactCount` (`sources.copy.artifactCount`) | Read only by the checklist |
| Recipes | `macos/App/DesignSystem/Recipes/` (`DeleteImpact`, `LastUpdated`, `Marks`, `OmnibarHitRow`); `macos/App/DesignSystem/README.md` recipe table (lines 180 – 182); `docs/design-system-layers.md` tree (line 53) | `LastUpdated` has its own `L10n.LastUpdated` namespace |
| Delete notice group | `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` `PVDeleteImpactNotice.groupView(_:)` (private) | Heading `PVDeleteImpactCopy.heading(group)` (+ raw `via` when unknown), then the listed rows as buttons calling `PVDeleteImpactControls.activate(row, onNavigate:dismiss:)`, then "and N more" (`overflow`), in a raised rounded box. Private to the notice |
| Preflight heading | `PVDeleteImpactCopy.heading(_ group:)` with via `evidencegraphs.artifact_in_use` | "%lld citations in this graph use this artifact" (one: "… uses …"), from S12-05 D-13 |
| Graph store API | `macos/App/Platform/GenealogyStore.swift` `createEvidenceGraph(projectDir:userID:sourceID:name:description:artifactIDs:) -> (CatalogEvidenceGraph, CatalogWriteEffects)`, `updateEvidenceGraph(…graphID:name:description:artifactIDs:) -> CatalogWriteEffects` ("Empty `artifactIDs` leaves the links unchanged"), `getEvidenceGraphUnlinkImpact(projectDir:graphID:artifactIDs:) -> [CatalogEvidenceGraphUnlinkBlock]` (`artifactID`, `citations: CatalogDeleteImpactGroup`) | No Swift caller of create, update or the preflight yet |
| Engine rules | `core/database/evidencegraphs/evidencegraphs.go` `Create`, `Update` | Both trim name and description; blank name or (Create) no Artifact → `evidencegraphs.invalid`; names need not be unique; `Update` links before it unlinks, refuses with `artifact_in_use` / `last_artifact`, and rolls back the rename on any refusal |
| FakeStore | `macos/App/Platform/FakeStore.swift` `createEvidenceGraph`, `updateEvidenceGraph`, `getEvidenceGraphUnlinkImpact` (returns `unlinkImpactBlocks` filtered by the given ids), `graphEffects(sourceID:)`, `deleteEvidenceGraphCalls`, `listEvidenceGraphsError` | Create throws `StoreBoom.boom` on a blank name or no Artifact. No call counters or error hooks for create, update or the preflight |
| Section model | `macos/App/Features/Sources/SourceGraphsSection.swift` `SourceGraphsSection` | `rows`, `loadError`, `location(for:)`, `blockedReason(for:)` over a `SourcePageContext`. No form state |
| Section view | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphsView`, private `SourcePageGraphRow` | `PVSectionHeader(title:meta:)` with no `actions`; row buttons `HStack(spacing: PVSpacing.space4)` of Open then Delete, `.fixedSize()` |
| Page context | `macos/App/Features/Sources/SourcePageContext.swift` | `sourceID`, `projectDir`, `userID`, `store`, `session`, `workspace` (its `artifacts` in page order), `source`, `toast`, `clearPageError()` |
| Page dialogs | `macos/App/Features/Sources/SourcePageView.swift` | `.pvFormDialog(isPresented:…)` for Add artifact, Add file, Add metadata; `.sourceCopyDialog($model.copying)`; `.vocabularyToastOverlay` |
| Precedent form | `SourceCopyModel.swift`, `SourceCopyDialog.swift` | A model presented while non-nil through a `View` extension modifier over `pvFormDialog` (520pt, `isRunning`, `isDirty`, `confirmDisabled`); load state `.loading` / `.failed(String)` / `.ready` with skeleton bars and a Read again callout; save error callout at the top |
| Add artifact form | `SourcePageArtifactsView.swift` `addForm`; `SourceArtifactsSection.create()` | Label `PVField(…, error:, marker: .required)` + `PVInput`, the error cleared `.onChange`; the error set only on confirm |
| Edit button | `SourcePageArtifactsView.swift` Artifact card | `PVIconButton(.penLine, label:, size: .sm)` |
| Section header add | `SourcePageArtifactsView.swift` | `PVButton(L10n.Sources.addArtifact, variant: .primary, size: .sm, icon: .plus)` in `PVSectionHeader` `actions` |
| Last-graph copy | `L10n.DeleteImpact.gateLastGraphBody` (`deleteImpact.gate.lastGraph.body`) | "A source keeps at least one evidence graph. Add another graph before you delete this one." |
| Xcode target | `macos/Provenencia.xcodeproj/project.pbxproj` | Hand-maintained `AA…` ids. `SourceArtifactChecklist.swift`: lines 53, 480, 934 (Sources group), 2260. The `DeleteImpact` recipe group is `D5C000000000000000000022` (line 1814), child of `Recipes` `D5C000000000000000000020` (line 1869) |
| Tests | `macos/ProvenenciaTests/SourceCopyModelTests.swift`, `SourcePageModelTests.swift` (`graphRowsAreTheSourcesGraphsByName`, `deleteEmptyGraphCallsStoreAndToasts`), `PVDeleteImpactTests.swift` `lastGraphGateCopy` (asserts today's body) | `lastGraphGateCopy` changes in STEP-7 |

**Precedent.** S12-02 ([#485](https://github.com/mendahu/provenencia/pull/485)) built the Copy Source form: copy its model / dialog split (`SourceCopyModel` + `View.sourceCopyDialog`), its load states and its save-error callout. Copy Add artifact's name validation (error on confirm, cleared on edit). Do not copy `SourceArtifactsSection.create()`'s hand-written workspace patch: a graph write changes no workspace field, and its effects (`lists: [.evidenceGraphs]`) refresh the graph list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `SourceArtifactChecklist` moves to `macos/App/DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift` *(new)* as `PVArtifactChecklist`, in a new `ArtifactChecklist` group (id `D5C000000000000000000023`, the next after the highest group id `D5C000000000000000000022`) under `Recipes` in `project.pbxproj`; the old file is deleted. Its inputs: today's five plus `locked: [String: String] = [:]` (Artifact id → hint; `PVArtifactChecklistState` takes its keys) and `accessibilityIdentifierPrefix: String`. The Copy form passes `"sources.copy.artifacts"`, so its ids do not change. | Its second caller is here; `add-ui-component` step 3 (Provenencia-specific, two call sites) and the brief inventory. | A second checklist in the graph form; keeping it a Sources snowflake with two callers. |
| D-2 | The checklist's strings move to `L10n.ArtifactChecklist` *(new)*, an enum placed directly before `enum LastUpdated` in `L10n.swift`: `heading` (`artifactChecklist.heading` "Artifacts"), `all` (`artifactChecklist.all` "All Artifacts"), `empty` (`artifactChecklist.empty` "This Source has no Artifacts."), `count(_:_:)` (`artifactChecklist.count` "%1$lld of %2$lld"). The four `sources.copy.*` keys and their `L10n.Sources` members are deleted. | A recipe does not read a feature's copy namespace; `L10n.LastUpdated` is the recipe precedent. Values unchanged, so AC-4.1 holds. | Keeping `copySource*` names in a shared recipe. |
| D-3 | Locked rows: the pure helper `PVArtifactChecklistState` *(new)*, in the recipe file, owns the arithmetic, with `static func liveExcluded(ids:excluded:locked:)`, `checkedCount`, `isAllChecked`, `isAllIndeterminate`, and `toggledAll(on:ids:locked:) -> Set<String>` (off → every id not in `locked`; on → `[]`). Signatures, all over `ids: [String]` (the list's Artifact ids in order), `excluded: Set<String>`, `locked: Set<String>`: `liveExcluded(ids:excluded:locked:) -> Set<String>` = `excluded ∩ ids − locked`; `checkedCount(ids:excluded:locked:) -> Int` = `ids.count − liveExcluded.count`; `isAllChecked(ids:excluded:locked:) -> Bool` = `liveExcluded.isEmpty`; `isAllIndeterminate(ids:excluded:locked:) -> Bool` = `!liveExcluded.isEmpty && checkedCount > 0`; `checkedIDs(ids:excluded:locked:) -> [String]` = `ids` in order minus `liveExcluded`, and `checkedCount` is its count. `SourceCopyModel.artifactIDsToCopy` and `SourceGraphFormModel.checkedArtifactIDs` (D-6) both return `checkedIDs` (the Copy form passes `locked: []`), so the heading count and the ids a save sends are one function. A locked row always uses the labelled form, `PVCheckbox(verbatim: artifact.displayTitle, description: hint, isChecked: .constant(true), isLocked: true)` — `displayTitle` falls back to the ref, so an unlabelled locked Artifact shows its ref as the label and the hint visibly. `PVCheckbox`'s `verbatim` init gains `isLocked: Bool = false` *(new)*: forwarded to `PVCheckboxMark(isLocked:)` (the kit's lock glyph on a paper fill, "a binding the researcher cannot clear", as `PropertiesView` draws it), the button disabled with no opacity change, and the accessibility value `L10n.DesignSystem.checkboxLocked` *(new)* (`designSystem.checkbox.locked` "locked") in `CheckboxAccessibility`, beside "partly checked". Unlocked rows keep today's two forms and take `isDisabled`. A locked id is never inserted into `excluded`. | The All row must not uncheck a locked Artifact (GF-7); a pure helper is testable without a view. One checked-ids rule for every caller (review M-2). The kit already has one look for a link that can't be cleared (review M-3); S12-13's card checkboxes get it as a parameter. | Locking by disabling the whole checklist; storing locked ids in `excluded`; a dimmed check beside the kit's lock mark. |
| D-4 | The notice's group view becomes `PVDeleteImpactGroupList` *(new)*, an internal struct in `PVDeleteImpact.swift`: `title: String`, `group: CatalogDeleteImpactGroup`, `onNavigate: (WorkspaceLocation) -> Void`, `dismiss: () -> Void`. It draws the title line (with the unknown-`via` mono tag as today), the rows and the overflow line in today's box, and a row calls `PVDeleteImpactControls.activate(row, onNavigate:, dismiss:)` itself. `PVDeleteImpactNotice` passes `PVDeleteImpactCopy.heading(group)`, its `onNavigate` and `onDismiss`. | One Citation list for the notice and the form (brief inventory "Extend"); the form titles each list with its Artifact. The dismiss-then-navigate rule stays in `activate`, for both callers (review M-1). | Making `groupView` public as is — its heading is the via line, which repeats the locked row's hint. |
| D-5 | `SourceGraphFormModel` *(new)*, `@MainActor @Observable final class` in `macos/App/Features/Sources/SourceGraphFormModel.swift` *(new)*, built from a `SourcePageContext` and `enum Mode { case create, edit(CatalogEvidenceGraph) }`. State: `draft: SourceGraphFormDraft` *(new)* (`name`, `description`, `excludedArtifactIDs: Set<String>`; `Equatable`), `prefill`, `nameError: String?`, `saveError: String?`, `isSaving`, `loadState: SourceFormLoadState` (D-14), `blocks: [CatalogEvidenceGraphUnlinkBlock]`. Prefill: create → empty name and description, `excluded = []`; edit → the graph's name and description, `excluded` = page Artifact ids not in `graph.artifactIDs`. | Same shape as `SourceCopyModel`; `excluded` semantics keep "a new Artifact starts linked" (brief §2) true while the form is open. One load-state enum for both forms (D-14). | A `checked` set (an Artifact added mid-edit would start unchecked); two models for create and edit. |
| D-6 | Reads on the model: `artifacts` = `context.workspace?.artifacts ?? []`; `projectDir` = `context.projectDir`; `sourceTypeIconKey` = `context.types.first { $0.id == context.source?.sourceTypeID }?.iconKey` (as `SourceCopyModel.sourceTypeIconKey`); `locked: [String: String]` = each block's `artifactID` → `PVDeleteImpactCopy.heading(block.citations)`; `checkedArtifactIDs` = `PVArtifactChecklistState.checkedIDs(ids: artifacts ids, excluded: draft.excludedArtifactIDs, locked: Set(locked.keys))` (D-3); `needsArtifact` = `!artifacts.isEmpty && checkedArtifactIDs.isEmpty`; `canSave` = `loadState == .ready && !isSaving && !needsArtifact`; `isDirty` = `draft != prefill`. | The hint is the S12-05 heading for that via, so the row and the engine word it alike. | A second plural string for the hint. |
| D-7 | Preflight: `init` sets `loadState` to `.loading` in edit mode when `graph.artifactIDs` is non-empty, `.ready` otherwise (and in create mode); the section then starts `loadKeptLinks()` *(new)*, which calls `getEvidenceGraphUnlinkImpact(graphID:, artifactIDs: graph.artifactIDs)` once and sets `blocks` and `.ready`, or `.failed(L10n.Errors.message(for:))`. `loadKeptLinks()` returns at once for a graph with no links. **Read again** calls `loadKeptLinks()`. | User 2026-10-25: lock in-use Artifacts up front. The preflight takes every linked id (S12-05 D-4), so one call covers the whole form. | Running the preflight on Save; one call per unchecked box. |
| D-8 | `save() async -> Bool`: guard `!isSaving`; trim name; blank → `nameError = L10n.string(L10n.Sources.graphFormNameRequired)`, return false; guard `canSave`. Create: `createEvidenceGraph(sourceID: context.sourceID, name:, description: trimmed, artifactIDs: checkedArtifactIDs)`, `context.session.apply(effects)`, `context.toast = VocabularyToast(title: L10n.string(L10n.Sources.toastGraphAddedTitle), body: L10n.Sources.toastGraphAddedBody(ref: graph.ref), tone: .success)`. Edit: `updateEvidenceGraph(graphID:, name:, description:, artifactIDs: checkedArtifactIDs)` (empty when the Source has no Artifacts, which means unchanged), `context.session.apply(effects)`, no toast. Success returns true; a throw sets `saveError = L10n.Errors.message(for: error)` and returns false with `draft` untouched. | One write per save (S12-04 D-7); the effects refresh the graph list with no hand-written cache write (working rule "The graph list has its own effect"). User 2026-10-25: stay on the page after Add. | Separate link / unlink calls; a write-through into the graph list. |
| D-9 | `SourceGraphsSection` gains `var form: SourceGraphFormModel?`, `canAdd: Bool` (`!(context.workspace?.artifacts.isEmpty ?? true)`), `addBlockedReason: String?` (`L10n.string(L10n.Sources.addEvidenceGraphNeedsArtifact)` when `!canAdd`), `openAdd()` (no-op when `!canAdd`), `ownGraph(id:) -> CatalogEvidenceGraph?` *(new)* (`context.session.evidenceGraph(id:)` when its `sourceID == context.sourceID`; `SourcePageModel.askDeleteGraph(id:)` calls it in place of its own check, review M-7), `openEdit(id:)` (no-op unless `context.workspace` is non-nil and `ownGraph(id:)` finds the graph; then sets `form` and starts `Task { await form?.loadKeptLinks() }`). | S12-11 D-1 named this class as the home of the form state. | Form state on `SourcePageModel`. |
| D-10 | The dialog: `View.sourceGraphFormDialog(_ model: Binding<SourceGraphFormModel?>)` *(new)* in `macos/App/Features/Sources/SourceGraphFormDialog.swift` *(new)*, as `sourceCopyDialog`: `pvFormDialog` at width 520, `isRunning: isSaving`, `isDirty`, `confirmDisabled: !canSave`, prefix `sources.graphForm`; on confirm `if await current.save() { model = nil }`. Body (private `SourceGraphForm`): save-error callout; Name field; Description field; `PVDivider()`; the checklist area through `SourceFormLoadArea` (D-14; ready → `PVArtifactChecklist(… locked: model.locked, accessibilityIdentifierPrefix: "sources.graphForm.artifacts")`); the `needsArtifact` line; the Kept linked group when `blocks` is non-empty. Each Kept linked list gets `onNavigate: { navigation.go(to: $0) }` and `dismiss: { model = nil }` (D-4). | One modifier line on the page, as Copy. Rows close and navigate as the delete notice's rows do. | A second `pvFormDialog` per mode; a "Discard changes?" before navigating. |
| D-11 | Section header: `actions` holds `PVButton(L10n.Sources.addEvidenceGraph, variant: .primary, size: .sm, icon: .plus)`, `.disabled(!canAdd)`, `.help` / `.accessibilityHint` from `addBlockedReason`, action `model.graphs.openAdd()`, id `sources.page.addEvidenceGraph`. The header shows the button while the graph list loads or fails too. Row: `PVIconButton(.penLine, label: L10n.Sources.editEvidenceGraphLabel(name:), size: .sm)` between Open and Delete, action `model.graphs.openEdit(id: graph.id)`, id `sources.page.graph.edit.<ref>`, `.disabled(model.source == nil)`. | Brief GF-1 – GF-3; the Artifacts header and card. | Hiding Add while the Source has no Artifact — the working rule says unavailable, and the tooltip explains. |
| D-12 | `deleteImpact.gate.lastGraph.body` becomes "A source keeps at least one evidence graph. Use Add evidence graph to make another, then delete this one. A new evidence graph needs an artifact." No recipe change. | User 2026-10-25 (one body); S12-11 review UX-3. The body names the control's label exactly. | Two bodies keyed on whether the Source has Artifacts. |
| D-13 | Name field: `PVField(label: L10n.Sources.graphFormName, error: nameError, marker: .required)` + `PVInput(text:, prompt: L10n.Sources.graphFormNamePlaceholder, isInvalid:, activateOnAppear: true)` (focused on open, as `SourceTitleTypeFields`' title), the error cleared `.onChange` of the name. Description: `PVField(label: L10n.Sources.graphFormDescription, hint: L10n.Sources.graphFormDescriptionHint)` + `PVTextArea(text:, lineLimit: 3...3)`. Description is trimmed on save. | Add artifact's validation pattern; brief GF-5. | Validating as the researcher types. |
| D-14 | `SourceCopyLoadState` is renamed `SourceFormLoadState` and moves, with the Copy form's private loading and failed bodies, into `macos/App/Features/Sources/SourceFormLoadArea.swift` *(new)*: `struct SourceFormLoadArea<Ready: View>: View` with `state: SourceFormLoadState`, `loadingText: LocalizedStringResource`, `errorTitle: LocalizedStringResource`, `errorMessage: (String) -> String`, `onReadAgain: () -> Void`, `accessibilityIdentifierPrefix: String`, `@ViewBuilder ready: () -> Ready`. It draws today's `SourceCopyForm.loading` (three bars, italic caption, one a11y label; id `<prefix>.loading`) and the compact danger callout with **Read again** (ids `<prefix>.loadError`, `<prefix>.readAgain`). `L10n.Sources.copySourceReadAgain` becomes `L10n.Sources.formReadAgain` (`sources.form.readAgain` "Read again"). `SourceCopyForm.parts` calls it with `copySourceLoading`, `copySourceLoadErrorTitle`, `copySourceLoadErrorMessage`, prefix `sources.copy` (ids unchanged) and its part checkboxes as `ready`; the graph form with `graphFormLoading`, `graphFormLoadErrorTitle`, `{ $0 }`, prefix `sources.graphForm` and the checklist. | One loading and read-error body for Source-page forms; no Copy-named enum in the graph form (review M-4, M-6). | A copy of `SourceCopyForm.loading` in the graph form; reusing `SourceCopyLoadState` under its Copy name. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `PVArtifactChecklist` *(new)*, `Recipes/ArtifactChecklist/` *(new)* | component (recipe) | `SourceArtifactChecklist` | It **is** that view, promoted and extended with locked rows (D-1, D-3); the old file is deleted. |
| `PVArtifactChecklistState` *(new)* | helper | `SourceArtifactChecklist.liveExcluded` and `allRow` binding | The same arithmetic, moved out of the view and taught about locked ids. |
| `L10n.ArtifactChecklist` *(new)* | key namespace | `L10n.Sources.copySource*` (four keys), `L10n.LastUpdated` | Moves the four keys; recipe namespace precedent (D-2). |
| `PVDeleteImpactGroupList` *(new)* | component (recipe view) | `PVDeleteImpactNotice.groupView` | Extracted from it; the notice calls it (D-4). |
| `SourceGraphFormModel`, `SourceGraphFormDraft` *(new)* | type | `SourceCopyModel`, `SourceCopyChoices` | Different write and fields; same shape. |
| `SourceFormLoadArea` *(new)*, `SourceFormLoadState` (renamed) | component (snowflake) / type | `SourceCopyForm.loading` and its `.failed` branch, `SourceCopyLoadState` | Moved out of the Copy form so both forms draw one area (D-14). |
| `PVCheckbox` `isLocked` *(new)*, `L10n.DesignSystem.checkboxLocked` *(new)* | component param / key | `PVCheckboxMark.isLocked`, `L10n.Properties.bindingLocked` | The mark already draws the lock; the labelled checkbox forwards it (D-3). A kit key, not the Properties one. |
| `PVArtifactChecklistState.checkedIDs` *(new)* | helper | `SourceCopyModel.artifactIDsToCopy` | Becomes the one checked-ids rule; the Copy model calls it (D-3). |
| `SourceGraphsSection.ownGraph(id:)` *(new)* | func | `SourcePageModel.askDeleteGraph`'s inline check | Shared by delete and edit (D-9). |
| `View.sourceGraphFormDialog` *(new)*, `SourceGraphFormDialog.swift` *(new)* | component (snowflake) | `View.sourceCopyDialog` | Same presenter pattern, another form. |
| `SourceGraphsSection.form`, `canAdd`, `addBlockedReason`, `openAdd()`, `openEdit(id:)` *(new)* | func | `SourceArtifactsSection.openAdd()` / `isAdding` | Extends the section S12-11 made for this. |
| `L10n.Sources.addEvidenceGraph`, `addEvidenceGraphNeedsArtifact`, `editEvidenceGraphLabel(name:)`, `graphForm*`, `toastGraphAdded*` *(new)* | key | `L10n.Sources.addArtifact`, `toastGraphDeleted*` | New copy in the existing namespace. |
| FakeStore `createEvidenceGraphCalls`, `updateEvidenceGraphCalls`, `getEvidenceGraphUnlinkImpactCalls`, `createEvidenceGraphError`, `updateEvidenceGraphError`, `getEvidenceGraphUnlinkImpactError` *(new)* | helper | `deleteEvidenceGraphCalls`, `listEvidenceGraphsError` | Same counter and error-hook pattern. |

**Builds on.** `pvFormDialog`; the Copy form's presenter, load states and callouts; the DeleteImpact recipe's list and its activate rule; the S12-05 preflight and its heading; the session's one graph list and its effect; `add-ui-component`'s promotion rule.
**Deletes.** `Features/Sources/SourceArtifactChecklist.swift` and its four `project.pbxproj` entries; `L10n.Sources.copySourceArtifactsHeading`, `copySourceAllArtifacts`, `copySourceNoArtifacts`, `copySourceArtifactCount` and their `Localizable.xcstrings` entries; the private `PVDeleteImpactNotice.groupView`; `SourceCopyForm.loading` and its `.failed` branch body (moved, D-14); `L10n.Sources.copySourceReadAgain` (renamed); `askDeleteGraph`'s inline graph check.
**Next adds.** S12-13 adds graph checkboxes on the Artifacts section, reusing the locked-row hint (D-6) for its unlink refusal. Any later form that picks Artifacts uses `PVArtifactChecklist`. Any later in-form refusal list over a `CatalogDeleteImpactGroup` uses `PVDeleteImpactGroupList`.

### Schema

None — no catalog change.

### Go core

None — `evidencegraphs.Create`, `Update` and `UnlinkImpact` do what the form needs.

### FFI

None — `CreateEvidenceGraph`, `UpdateEvidenceGraph` and `GetEvidenceGraphUnlinkImpact` exist since S12-05.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Form dialog | `pvFormDialog` (520pt) | `Features/Sources/SourceGraphFormDialog.swift` *(new)* | D-10. Create: `L10n.Sources.graphFormAddTitle`, `graphFormAddSubtitle`, confirm `graphFormAddConfirm`. Edit: `graphFormEditTitle`, `verbatimSubtitle: graph.ref`, confirm `graphFormSaveConfirm`. Cancel `L10n.Sources.cancelAction` |
| Name, Description | `PVField` + `PVInput` / `PVTextArea` | same | D-13; ids `sources.graphForm.name`, `sources.graphForm.description` |
| Artifact checklist | `PVArtifactChecklist` | `DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift` *(new)* | D-1, D-3; `isDisabled: isSaving` |
| None-checked line | `Text(L10n.Sources.graphFormNoneChecked)` caption, `PVColor.textMuted` | `SourceGraphFormDialog.swift` | When `needsArtifact`; id `sources.graphForm.noneChecked` |
| Kept linked | `PVCallout(tone: .neutral, title: L10n.Sources.graphFormKeptLinkedTitle, message: L10n.Sources.graphFormKeptLinkedMessage, compact: true)`, then one `PVDeleteImpactGroupList(title: artifact.displayTitle, group: block.citations, onNavigate:, dismiss:)` per block in page order | same | D-4, D-10; id `sources.graphForm.keptLinked` |
| Loading and read error | `SourceFormLoadArea` | `Features/Sources/SourceFormLoadArea.swift` *(new)* | D-14; ids `sources.graphForm.loading`, `.loadError`, `.readAgain` |
| Save error | `PVCallout(tone: .danger, title: graphFormSaveErrorTitle, message: saveError)` | same | id `sources.graphForm.error` |
| Add evidence graph | `PVButton(.primary, .sm, .plus)` in `PVSectionHeader` `actions` | `SourcePageGraphsView.swift` | D-11 |
| Edit | `PVIconButton(.penLine, .sm)` | `SourcePageGraphsView.swift` `SourcePageGraphRow` | D-11 |
| Presenter | `.sourceGraphFormDialog($model.graphs.form)` | `SourcePageView.swift` | After `.sourceCopyDialog($model.copying)` |

**Frames.** Filled from the agreed board (Handoff of S12-D3). Until then the brief's §5 frames and §4 requirements are the target.

States: create; edit loading, ready, read error; locked rows with Kept linked; none checked; blank name; save error; Source with no Artifacts (Add disabled; edit of Default allowed); 40 Artifacts (the checklist scrolls, as Copy). Accessibility: Add's hint names why it is disabled; Edit's label names the graph; a locked row reads label, checked, "locked" and its hint; Kept linked rows keep the recipe's "{ref}, {title}" label.

### Strings

L10n keys *(new)*, to be checked against the agreed board:

- `L10n.ArtifactChecklist.heading` `artifactChecklist.heading` = "Artifacts"; `.all` `artifactChecklist.all` = "All Artifacts"; `.empty` `artifactChecklist.empty` = "This Source has no Artifacts."; `.count(_:_:)` `artifactChecklist.count` = "%1$lld of %2$lld"
- `L10n.Sources.addEvidenceGraph` `sources.page.addEvidenceGraph` = "Add evidence graph"
- `L10n.Sources.addEvidenceGraphNeedsArtifact` `sources.graph.addNeedsArtifact` = "Add an artifact before you add an evidence graph"
- `L10n.Sources.editEvidenceGraphLabel(name:)` `sources.page.graph.editLabel` = "Edit %@"
- `L10n.Sources.graphFormAddTitle` `sources.graphForm.addTitle` = "Add evidence graph"
- `L10n.Sources.graphFormAddSubtitle` `sources.graphForm.addSubtitle` = "A part of this source you work and promote on its own."
- `L10n.Sources.graphFormAddConfirm` `sources.graphForm.addConfirm` = "Add"
- `L10n.Sources.graphFormEditTitle` `sources.graphForm.editTitle` = "Edit evidence graph"
- `L10n.Sources.graphFormSaveConfirm` `sources.graphForm.saveConfirm` = "Save"
- `L10n.Sources.graphFormName` `sources.graphForm.name` = "Name"; `graphFormNamePlaceholder` `sources.graphForm.namePlaceholder` = "Jones household"; `graphFormNameRequired` `sources.graphForm.nameRequired` = "Give the evidence graph a name."
- `L10n.Sources.graphFormDescription` `sources.graphForm.description` = "Description"; `graphFormDescriptionHint` `sources.graphForm.descriptionHint` = "Which part of the source this graph covers"
- `L10n.Sources.graphFormNoneChecked` `sources.graphForm.noneChecked` = "Check at least one artifact."
- `L10n.Sources.graphFormKeptLinkedTitle` `sources.graphForm.keptLinkedTitle` = "Kept linked"; `graphFormKeptLinkedMessage` `sources.graphForm.keptLinkedMessage` = "These artifacts stay linked while citations in this graph use them."
- `L10n.Sources.graphFormLoading` `sources.graphForm.loading` = "Reading the artifacts…"
- `L10n.Sources.graphFormLoadErrorTitle` `sources.graphForm.loadErrorTitle` = "Couldn’t read the artifacts"
- `L10n.DesignSystem.checkboxLocked` `designSystem.checkbox.locked` = "locked"
- `L10n.Sources.graphFormSaveErrorTitle` `sources.graphForm.saveErrorTitle` = "Couldn’t save the evidence graph"
- `L10n.Sources.toastGraphAddedTitle` `sources.page.toastGraphAddedTitle` = "Evidence graph added"; `toastGraphAddedBody(ref:)` `sources.page.toastGraphAddedBody` = "%@ is ready to open."

Changed: `deleteImpact.gate.lastGraph.body` (D-12). Renamed: `sources.copy.readAgain` → `sources.form.readAgain` (`L10n.Sources.formReadAgain`, D-14). Removed: `sources.copy.artifactsHeading`, `sources.copy.allArtifacts`, `sources.copy.noArtifacts`, `sources.copy.artifactCount`.

## 7. Steps

`XCT` below stands for `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`, the command `scripts/check-pr.sh` runs, with an `-only-testing:` class appended. A new Swift file goes in its target by copying an existing file's four `project.pbxproj` entries (a `PBXBuildFile`, a `PBXFileReference`, a group child, a Sources-phase entry) with new ids counting up from the highest `AA…` id in the file. Test files go in the `ProvenenciaTests` target the same way (`add-swift-test`).

**STEP-1 — Promote the Artifact checklist to a recipe**
Files: `macos/App/DesignSystem/Recipes/ArtifactChecklist/PVArtifactChecklist.swift` *(new)*, `macos/App/Features/Sources/SourceArtifactChecklist.swift` (deleted), `macos/App/DesignSystem/Components/Checkbox/PVCheckbox.swift` (`isLocked`, with a preview case), `SourceCopyDialog.swift`, `SourceCopyModel.swift` (`artifactIDsToCopy` onto `checkedIDs`), `L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `project.pbxproj` (new `ArtifactChecklist` group under `Recipes`, as `DeleteImpact`; the old entries removed), `macos/App/DesignSystem/README.md` (recipe table row), `docs/design-system-layers.md` (tree), `macos/ProvenenciaTests/PVArtifactChecklistTests.swift` *(new)*. Skills: `add-ui-component`, `add-localized-string`, `add-swift-test`.
Do: D-1, D-2, D-3. The view keeps today's layout; add a `#Preview` with static sample Artifacts, one locked. Add T-1.
Check: `XCT -only-testing:ProvenenciaTests/PVArtifactChecklistTests -only-testing:ProvenenciaTests/SourceCopyModelTests` passes; `rg -n 'SourceArtifactChecklist|copySourceArtifactsHeading|copySourceAllArtifacts|copySourceNoArtifacts|copySourceArtifactCount' macos` prints nothing; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — One DeleteImpact group list and one form load area**
Files: `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/App/Features/Sources/SourceFormLoadArea.swift` *(new)*, `SourceCopyModel.swift`, `SourceCopyDialog.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `SourceCopyModelTests.swift` (rename only). Skills: `add-ui-component`, `add-localized-string`.
Do: D-4, D-14. The notice and the Copy form draw exactly as before.
Check: `XCT -only-testing:ProvenenciaTests/PVDeleteImpactTests -only-testing:ProvenenciaTests/SourceCopyModelTests` passes; `rg -n 'func groupView|SourceCopyLoadState|copySourceReadAgain' macos` prints nothing.

**STEP-3 — FakeStore hooks**
Files: `macos/App/Platform/FakeStore.swift`.
Do: the six members in §6 Codebase fit. Each counter increments at the top of its method; each error hook throws inside `withState` before any state change when set (`getEvidenceGraphUnlinkImpact`'s body becomes `try withState`).
Check: `XCT -only-testing:ProvenenciaTests/SourcePageModelTests` passes (no hook is set, so behavior is unchanged) and the build succeeds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`).

**STEP-4 — Graph form model**
Files: `macos/App/Features/Sources/SourceGraphFormModel.swift` *(new)*, `SourceGraphsSection.swift`, `SourcePageModel.swift` (`askDeleteGraph`), `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift` *(new)*. Skills: `add-localized-string`, `add-swift-test`.
Do: D-5 – D-9 (with `ownGraph(id:)` and `askDeleteGraph` onto it). Add T-2 – T-9.
Check: `XCT -only-testing:ProvenenciaTests/SourceGraphFormModelTests` passes; the xcstrings check passes.

**STEP-5 — Graph form dialog**
Files: `macos/App/Features/Sources/SourceGraphFormDialog.swift` *(new)*, `SourcePageView.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-10, D-13; §6 macOS; strings as §6.
Check: the build succeeds. The dialog is first reachable in STEP-6, whose run-the-app check opens it.

**STEP-6 — Add and Edit on the section**
Files: `SourcePageGraphsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-11. Add T-10.
Check: `XCT -only-testing:ProvenenciaTests/SourcePageModelTests` passes; run the app (README "Run the Mac app") on any project and see Add evidence graph in the section header and Edit on the Default row; Add opens the form with every Artifact checked, and Edit opens it filled from Default.

**STEP-7 — Last-graph notice body**
Files: `L10n.swift` (`gateLastGraphBody` default value), `Localizable.xcstrings`, `PVDeleteImpactTests.swift`.
Do: D-12; update `lastGraphGateCopy`'s expected body (T-11).
Check: `XCT -only-testing:ProvenenciaTests/PVDeleteImpactTests` passes.

**STEP-8 — Full check**
Files: none new.
Do: run the whole gate.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md`, this plan. (The brief is archived at board agreement.)
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-12-graph-form.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Fixtures for T-2 – T-10: a FakeStore with Source `src-1` (Artifacts `a1` – `a6` in that page order) and its graph list, a session with the Source page location applied (so the graph list is a visible key, as S12-11's T-5), and `SourcePageModel.warmFromSession()` then a settle of `session.evidenceGraphsHandle!`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PVArtifactChecklistTests.swift` `stateHonorsLockedRows` *(new)*, parameterized over the five cases in this row | ids [a1, a2, a3], locked {a2}: `toggledAll(on: false)` → {a1, a3}, then `checkedCount` 1, `isAllChecked` false, `isAllIndeterminate` true; `toggledAll(on: true)` → []; no locked and every id excluded → `isAllIndeterminate` false; a stale excluded id not in ids does not change `checkedCount`; `checkedIDs` over ids [a1, a2, a3], excluded {a1, a2}, locked {a2} → [a2, a3] in that order | AC-2.3, AC-4.1, D-3 |
| T-2 | Swift | `SourceGraphFormModelTests.swift` `createPrefillsEveryArtifactChecked` *(new)* | `openAdd()` → `form` mode create, `draft.name == ""`, `checkedArtifactIDs == [a1…a6]`, `loadState == .ready`, `isDirty` false | AC-1.1, D-5 |
| T-3 | Swift | `createSavesOneGraphAndToasts` *(new)* | name "Jones household", description " Household 14, page 3 ", exclude all but a3; `save()` → true, `createEvidenceGraphCalls == 1`; after settle the list holds a graph on `src-1` named "Jones household", description "Household 14, page 3", `artifactIDs == [a3]`; toast title "Evidence graph added" | AC-1.2, D-8 |
| T-4 | Swift | `blankNameAndNoArtifactBlockSave` *(new)*, parameterized | name "  " → `save()` false, `nameError == "Give the evidence graph a name."`, no create call; every Artifact excluded → `needsArtifact`, `canSave` false | AC-1.3, AC-1.4, D-6 |
| T-5 | Swift | `saveFailureKeepsChoices` *(new)*, parameterized over create and edit | `createEvidenceGraphError` / `updateEvidenceGraphError` set → `save()` false, `saveError` non-nil, `draft` equal to before | AC-1.6, AC-2.6 |
| T-6 | Swift | `editPrefillsAndLocksInUseArtifacts` *(new)* | graph `g-smith` linked to [a3, a4]; `unlinkImpactBlocks` = one block for a3 with a group of total 2; `openEdit(id: "g-smith")`, await `loadKeptLinks` → `getEvidenceGraphUnlinkImpactCalls == 1`, excluded = {a1, a2, a5, a6}, `locked == [a3: "2 citations in this graph use this artifact"]`, `blocks.count == 1`, `loadState == .ready` | AC-2.1, AC-2.3, D-6, D-7 |
| T-7 | Swift | `editSavesNameAndFullLinkSet` *(new)* | from T-6: name "Smith family", exclude a4, include a5; `save()` → true, `updateEvidenceGraphCalls == 1`; after settle `g-smith` has name "Smith family" and `artifactIDs == [a3, a5]`; no toast | AC-2.4, D-8 |
| T-8 | Swift | `editPreflightFailureThenReadAgain` *(new)* | `getEvidenceGraphUnlinkImpactError` set → `.failed`, `canSave` false; clear it, `loadKeptLinks()` → `.ready` | AC-2.2, AC-2.5, D-7 |
| T-9 | Swift | `editDefaultWithNoArtifactsKeepsLinks` *(new)* | Source with no Artifacts and its Default graph with no links; `openEdit`, no preflight call, `canSave` true; name "Whole census"; `save()` → true and the list's Default reads "Whole census" with `artifactIDs == []` | AC-2.7, D-7, D-8 |
| T-10 | Swift | `SourcePageModelTests.swift` `addGraphNeedsAnArtifact` *(new)* | Source with no Artifacts → `graphs.canAdd` false, `addBlockedReason == "Add an artifact before you add an evidence graph"`, `openAdd()` leaves `form` nil; with one Artifact → `canAdd` true, `openAdd()` sets `form` | AC-1.5, D-9 |
| T-11 | Swift | `PVDeleteImpactTests.swift` `lastGraphGateCopy` (edit) | The body equals D-12's copy | AC-3.1, D-12 |

AC-2.8 (a Kept linked row closes the form and navigates) is a manual check: the closure lives in the view.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourceCopyModelTests` (STEP-1, STEP-2), `PVDeleteImpactTests` (STEP-2, STEP-7), `check-localizable-xcstrings.py` (moved and removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (the four sources.copy.* checklist keys removed)
- [ ] UI matches the brief inventory: the PR description lists each component used per inventory row, with screenshots of the census Add frame and the Edit frame with kept links
- [ ] Version bump: none (VERSION stays 0.2.0)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief: moved to design/archive/ and design/README.md updated by the planner at board agreement, before this PR starts (check it is there; the implementer does not move it)
- [ ] Domain doc: none — no rule in docs/*.md changes; DesignSystem/README.md and design-system-layers.md list the new recipe
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a real project with a census Source holding at least three page scans. In **Evidence graphs**, click **Add evidence graph**: every page is checked, "3 of 3" (AC-1.1).
2. Uncheck all pages: Add is disabled and "Check at least one artifact." shows (AC-1.4). Check one page, clear the name, click Add: the name error shows (AC-1.3).
3. Name it "Jones household", add a description, keep one page, click **Add**. The dialog closes, the row appears in name order, the toast reads "Evidence graph added" / "GPH-… is ready to open." (AC-1.2). Open it: the composer offers only that page.
4. In that graph, cite the page on a Subject. Back on the Source page, **Edit** "Jones household": the page is checked and locked, its hint reads "1 citation in this graph uses this artifact", and Kept linked lists the Citation (AC-2.3). Click the Citation: the form closes and the Citation opens (AC-2.8).
5. Edit again: rename to "Jones family", check a second page, **Save**: the row shows the new name and moves to its place (AC-2.4).
6. Delete the second graph so one is left, then **Delete** the last one: the notice body names Add evidence graph (AC-3.1).
7. Add a new Source with no Artifacts: Add evidence graph is disabled with its tooltip (AC-1.5); **Edit** Default, rename it, Save (AC-2.7).
8. **Copy…** the census Source: the Artifact checklist looks and works as before (AC-4.1).
9. A locked row shows the kit's lock mark; VoiceOver reads its label, checked, "locked" and the hint; Edit reads "Edit {name}".

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving the checklist file loses its target membership | STEP-1 checks the build and runs `SourceCopyModelTests`. |
| The graph list does not refresh after a create (not a visible key in tests) | T-3 applies the Source page location first, as S12-11's T-5. |
| The preflight's id order differs from the page order | D-6 builds every list from the page's Artifacts, not the block order. |

**Stop and report if:**

- The agreed board draws the locked row without the kit's lock mark (D-3 then needs the user).
- `createEvidenceGraph` or `updateEvidenceGraph`'s effects do not refresh the graph list without a hand-written cache write (the row stays stale after T-3's or T-7's settle).
- `getEvidenceGraphUnlinkImpact` returns a block for an Artifact the graph does not link (§4's engine rule is not true).
- The agreed board draws a control or state the inventory does not list.

## 12. Open questions

None. The user answered the three product calls on 2026-10-25 (D-7, D-8, D-12). The plan stays `Blocked — on board S12-D3` until the board is agreed.

## 13. PR

**Title:** S12-12 — Graph form
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-12-graph-form.md`.
