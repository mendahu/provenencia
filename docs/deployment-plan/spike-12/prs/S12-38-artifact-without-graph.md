# S12-38 — Add an Artifact without a graph; the card saves links in one write

**Status:** Blocked — on board S12-D7
**Spike / slice:** Spike 12 · Slice 5 — Audit fixes: High
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D7](../design/S12-D7-artifact-in-no-graph.md), board not agreed yet
**Board:** the S12-D7 frame row on the Source page board ([Claude Design](https://claude.ai/design/p/de1e1ccc-aa35-455f-9b9c-3ae269593dc7?via=share)); until it is agreed, §6 macOS describes each frame in words from the brief
**Depends on:** S12-22 (landed [#600](https://github.com/mendahu/provenencia/pull/600)), S12-D7
**Blocks:** S12-30 (map: "S12-38 before S12-30, which then opts the card's and the graph form's rewritten error callouts into announcing")
**Branch:** `s12-38-artifact-without-graph`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-22 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (its effects list names the deleted RPCs)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [app-health audit](../app-health-audit.md) A-3, A-8; [rules](../source-structure.md#rules); working rules "An Artifact is the parent; it may stand in no graph", "An Artifact in no graph says so on its card", "An Artifact with Citations keeps at least one graph", "One unlink check, one link writer", "An Artifact card's Save is one write", "A new graph starts with the Artifacts in no graph checked", "FakeStore is a seeded fake, not a second engine"; [map review](../deployment-plan.review.md) K-3, K-11, K-12; the S12-22 plan (D-12 – D-14, **Next adds**) and ship log (UX-2); boards [S12-D3](../design/archive/S12-D3-graph-form.md) and [S12-D4](../design/archive/S12-D4-artifact-graph-links.md) for the chrome this reuses
**Planned:** 2026-10-28 against `main` @ `b1bf4c5`
**Review:** [`S12-38-artifact-without-graph.review.md`](S12-38-artifact-without-graph.review.md) — round 1 answered; round 2 runs once the board is agreed and the Strings are final

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

The researcher can add an Artifact that belongs to no Evidence graph yet, see such an Artifact marked on its card, and link it from there. The graph boxes show on every Source, one graph included. A card's Save writes the label, the description and every link in one write, so a refusal leaves nothing half applied and never asks to "save again to finish". Every lock on a box is the engine's refusal as its preflight reports it; Swift decides no lock. A new graph starts with the Artifacts that are in no graph.

**Why now.** Audit A-3 and A-8 (High), and the user's fix (2026-10-27): S12-22 built the engine half; this PR is the Swift half and deletes what the one write makes dead.

**What stays.** The card grid, the resting card of an Artifact that is in a graph, the Add file dialog, Artifact delete, cover and Use as thumbnail; the graph form's layout and its `needsArtifact` rule; the engine's rules (no Go rule changes).

## 2. User stories

**US-1.** As a researcher, when I add an Artifact, I choose its graphs on every Source, and may choose none, so that an Artifact can wait for its graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source with one graph "Default"; open Add artifact | The Evidence graphs group shows one checked box, Default |
| AC-1.2 | Same; label "Page 7", Add | `createArtifact` is sent `graphIDs: ["g-1"]` |
| AC-1.3 | Census (Brown, Jones, Smith); every box unchecked | Add is enabled; the line "This artifact won’t be in an evidence graph. Citations can’t use it until it is linked." shows; Add sends `graphIDs: []` |
| AC-1.4 | The session has no graph for this Source (list loading, failed, or not refreshed) | No group; Add sends `graphIDs: nil` (the engine links every graph) |

**US-2.** As a researcher, I see which Artifacts are in no graph and can link one from its card.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | ART-P6 is linked by no graph; the Source's graphs are in the session | ART-P6's resting card shows the "Not in an evidence graph" badge and **Link to a graph** |
| AC-2.2 | ART-P3 is linked by Jones | No badge, no button |
| AC-2.3 | The session has no graph for this Source | No card shows the badge |
| AC-2.4 | Click **Link to a graph** on ART-P6 | ART-P6 opens in edit with every box unchecked; check Jones and Save → `updateArtifact` with `graphIDs: ["g-j"]` |

**US-3.** As a researcher, a card Save writes everything or nothing, and its locks say why a graph keeps the Artifact.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | ART-P3 in Jones and Smith; uncheck Jones, check Brown, change the label; Save | One `updateArtifact` with the label and `graphIDs: ["g-b", "g-s"]`; no other write; the card leaves edit |
| AC-3.2 | Change only the label; Save | `updateArtifact` with `graphIDs: nil` |
| AC-3.3 | One-graph Source, Default links ART-P3 and ART-P4; card edit on ART-P3 | The group shows; the preflight runs for (Default, ART-P3); unchecking Default and saving sends `graphIDs: []` |
| AC-3.4 | Preflight blocks (Smith, ART-P3) `artifact_in_use` with 2 Citations | Smith locked, hint "2 citations in this graph use this artifact" |
| AC-3.5 | Preflight blocks (Brown, ART-P9) `last_artifact` | Brown locked, hint "An evidence graph needs at least one artifact." |
| AC-3.6 | ART-P9 only in Brown; preflight blocks (Brown, ART-P9) `citations_need_graph` with 1 Citation | Brown locked, hint "1 citation would be left in no evidence graph" |
| AC-3.7 | Same; check Jones | Brown stays locked; Save → `graphIDs: ["g-b", "g-j"]`. Edit again with the preflight now reporting no block: uncheck Brown, Save → `graphIDs: ["g-j"]` (two Saves move it; user, 2026-10-28) |
| AC-3.8 | Save refused `evidencegraphs.artifact_in_use`; the re-read preflight blocks (Jones, ART-P3), which the researcher unchecked | Card stays in edit; callout "Couldn’t save the artifact" / "Jones household keeps this artifact. Your changes are kept. {error}"; Jones checked and locked; label buffer kept |
| AC-3.9 | ART-P9 only in Brown, no block at edit open; uncheck Brown; Save refused `evidencegraphs.citations_need_graph`; the re-read blocks (Brown, ART-P9) `citations_need_graph` | Brown ends checked and locked; the callout names Brown household; no text says "save again" (S12-22 UX-2) |
| AC-3.10 | Save refused, and no re-read block locks a graph the researcher unchecked, or the re-read fails | "Your changes are kept. {error}" |
| AC-3.11 | A graph created during edit links ART-P3; uncheck Jones; Save | `graphIDs` keeps the new graph: `["g-s", <new id>]` |

**US-4.** As a researcher, the graph form locks every Artifact the engine says it must keep, and a new graph starts with the Artifacts in no graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Edit Smith (a3, a4); preflight blocks (Smith, a3) `citations_need_graph` with 1 Citation and (Smith, a4) `last_artifact` | a3 locked with "1 citation would be left in no evidence graph" and listed under Kept linked; a4 locked with "An evidence graph needs at least one artifact." and not listed under Kept linked |
| AC-4.2 | Graph form save refused `evidencegraphs.citations_need_graph` | The kept links are read again; the refusal stays in view |
| AC-4.3 | Six Artifacts; a1 – a4 linked by a graph, a5 and a6 by none; Add evidence graph | a5 and a6 checked, a1 – a4 unchecked; the form is not dirty |
| AC-4.4 | Every Artifact is in a graph, or the session has no graph for this Source | Every Artifact checked (today's default) |

**US-5.** As a contributor, the per-graph link RPCs are gone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `rg "LinkEvidenceGraphArtifact\|linkEvidenceGraphArtifact" api macos --glob '!*.pb.*'` | No hit (the `Unlink…` forms included by the suffix match) |
| AC-5.2 | `api/proto/engine.proto` | `107, 108` and both method names reserved; the four messages deleted |

## 3. Scope

| In | Covered by |
| --- | --- |
| One session read of the Artifacts in no graph (D-6) | STEP-1, AC-2.3, AC-4.4 |
| The group on every Source the session has graphs for, in the Add dialog and the card (ND-1, ND-6) | STEP-2, STEP-3, AC-1.1, AC-3.3 |
| No box checked allowed, with its line (ND-3, ND-6) | STEP-2, STEP-3, AC-1.3 |
| Every preflight block locks its box, worded by one hint (ND-7, ND-8) | STEP-3, AC-3.4 – AC-3.7 |
| The card Save as one `updateArtifact` with its links | STEP-4, AC-3.1, AC-3.2, AC-3.11 |
| The refused-save callout and re-read (ND-9) | STEP-4, AC-3.8 – AC-3.10 |
| The resting card's badge and **Link to a graph** (ND-4, ND-5, ND-10, ND-11) | STEP-5, AC-2.1 – AC-2.4 |
| Graph form: every block locks; Kept linked line (ND-12); refusal re-read; new graph's default checks | STEP-6, AC-4.1 – AC-4.4 |
| One FakeStore hold type for the held writes | STEP-4 |
| Delete the link and unlink RPCs on both sides, `graphListReloaded()`, the partial-save, needs-one and only-artifact strings | STEP-2, STEP-3, STEP-4, STEP-7, AC-5.1, AC-5.2 |

| Out | Home |
| --- | --- |
| Lifting a lock in Swift when another box is checked (a one-Save move of a cited Artifact) | not planned — two Saves (user, 2026-10-28: "No logic in Swift") |
| Naming, in a graph delete's confirm, the Artifacts it leaves in no graph | [`docs/ideas/artifact-outside-graph-marker.md`](../../../ideas/artifact-outside-graph-marker.md) |
| Patching the cached graph list from a link write instead of reloading it (review M-6) | S12-26 (reload only what moved) |
| Moving the test-side `Gate` in `WorkspaceSessionTests.swift` onto `FakeHold` | not planned (logged, review M-5) |
| A count or filter of Artifacts in no graph (section header, Sources list) | not planned |
| Announcing the card's and graph form's error callouts | S12-30 |
| Removing FakeStore's `createArtifact` linking rule | S12-35 |
| Narrowing the Artifact update's effects | S12-25 |
| Any engine rule change | not planned (S12-22 owns the rules) |

## 4. Current state

Verified at `b1bf4c5`.

| Area | Where | Today |
| --- | --- | --- |
| Session graph reads | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `EvidenceGraphIndex` (`byID`, `bySource`), `evidenceGraphs(sourceID:)` (nil with no list value, `[]` when the list has no graph for the Source), `evidenceGraphsRead(sourceID:)` (treats `[]` as loading) | No read of which Artifacts no graph links |
| Group visibility | `macos/App/Features/Sources/SourceArtifactsSection.swift` `visibleGraphs` | Non-nil only with `graphs.count >= 2` |
| Add's links | same file `draftGraphIDs` (`[String]`, `[]` when hidden), `draftNeedsGraph`, `canSubmitDraft`, `create()` (`draftGraphIDs.isEmpty ? nil : draftGraphIDs`) | Every box unchecked blocks Add |
| Card edit state | same file `LinkEdit` (`locked: [String: String]`), `beginEdit` (only when `visibleGraphs`), `checkedGraphIDs`, `linkedGraphIDs`, `liveGraphIDs`, `linksDirty` | — |
| Card locks | same file `loadLinkLocks` | One `getEvidenceGraphUnlinkImpact` over pairs; a switch on `block.reason`: `.lastArtifact` → `L10n.Sources.artifactGraphOnlyArtifact`; `.citationsNeedGraph` → no lock (S12-22 D-13) |
| Card save | same file `saveFields`, `saveErrorBody`, `graphListReloaded` | `updateArtifact(graphIDs: nil)`, then one `linkEvidenceGraphArtifact` per added graph and one `unlinkEvidenceGraphArtifact` per dropped graph; partial-save copy; waits on the graph list reload |
| Card views | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `addForm` (needs-one line, id `sources.page.addArtifact.graphs.needsGraph`), `graphLinks`, `editFields`, `restingText`, `descriptionLine` (3-line `minHeight`) | — |
| Graph rows | `macos/App/Features/Sources/SourceArtifactGraphLinks.swift` `SourceArtifactGraphLinks` over `PVChecklistFrame` / `PVChecklistCheckbox` (`macos/App/DesignSystem/Recipes/ArtifactChecklist/PVChecklistRows.swift`) | Takes `locked: [String: String]` |
| Graph form | `macos/App/Features/Sources/SourceGraphFormModel.swift` `init` (create prefill: nothing excluded), `blocks`, `locked`, `keptLinked`, `loadKeptLinks` (keeps `.artifactInUse` only), `save` (re-reads on the literal `"evidencegraphs.artifact_in_use"`); `macos/App/Features/Sources/SourceGraphFormDialog.swift` `keptLinked` | — |
| Store | `macos/App/Platform/GenealogyStore.swift` `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact` (protocol); `macos/App/Platform/GoStore.swift` (lines 517–552); `macos/App/Platform/CoreInvoke.swift` (lines 66–67); `macos/ProvenenciaTests/OnboardingModelTests.swift` (stub, lines 479–490) | `updateArtifact(… graphIDs: [String]?)` and `createArtifact(… graphIDs: [String]?)` already carry links (S12-22 D-12) |
| FakeStore | `macos/App/Platform/FakeStore.swift` link hold (`holdsLinkEvidenceGraphArtifact`, `heldLinkEvidenceGraphArtifacts`, `heldLinkEvidenceGraphArtifactCount`, `releaseLinkEvidenceGraphArtifact`, lines 119–123, 832–862); `linkEvidenceGraphArtifactError`, `unlinkEvidenceGraphArtifactError`; `listEvidenceGraphsCalls` (line 124); `setSubjectPosition` hold (`holdsSetSubjectPosition`, `heldSetSubjectPositions`, `heldSetSubjectPositionCount`, `releaseSetSubjectPosition`, lines 152–156, 2232–2270); `updateArtifact` (records `updateArtifactCalls`, `updateArtifactGraphIDs`, moves no link) | Two hand-rolled holds |
| Go | `api/ffi/handlers/evidence_graphs.go` `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact`; `api/ffi/dispatch.go` (lines 108–109, 303–306); `api/proto/engine.proto` (method enum lines 119–120, messages lines 506–526) | `evidencegraphs.Link` / `Unlink` stay: `applyLinks` calls them |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources.artifactGraphsNeedsOne`, `artifactGraphOnlyArtifact`, `artifactCardSaveErrorKept`, `artifactCardSaveErrorGraphKeeps`, `artifactCardSaveErrorMessage`, `graphFormKeptLinkedMessage`; `macos/App/Resources/Localizable.xcstrings` | — |
| Guidance | `.agents/skills/add-workspace-place/SKILL.md` line 172 | Names `linkEvidenceGraphArtifact` and `unlinkEvidenceGraphArtifact` |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` (section "Artifact ↔ graph links (S12-13)", lines 1455–1823); `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`; `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `writesReportTheGraphList`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` (the `setSubjectPosition` hold); `api/ffi/handlers/evidence_graphs_test.go` `TestLinkUnlinkEvidenceGraphArtifact` | — |

**Precedent.** `e7063636` deleted three RPCs on both sides in one commit (`GetPersonHeader` and siblings): proto methods reserved by number and name, messages deleted, handler and test files, `dispatch.go`, `CoreInvoke.swift`, `GenealogyStore`, `GoStore`, `FakeStore`, the `OnboardingModelTests` stub. Copy that list. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built the group and the card's edit; extend its views, do not add a second group. `SourceGraphFormModel` stores raw `blocks` and derives its locks from them: the card follows that shape (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | "The session has graphs for this Source" means `context.session.evidenceGraphs(sourceID:)` is non-nil **and** non-empty, as `evidenceGraphsRead` treats it (every Source has a graph, so `[]` is a list not yet refreshed). `visibleGraphs` returns those graphs, else nil. Every read in this PR (D-2, D-6, D-8) uses this one condition. Tests make "no graphs" with `store.listEvidenceGraphsError = CoreInvokeError.failed(status: 3)` set before the page warms | Review X-2: one meaning for "loaded" | `evidenceGraphsHandle?.value != nil` — every card would show the badge while a refreshed list lacks the Source |
| D-2 | `draftGraphIDs` becomes `[String]?`: nil when `visibleGraphs` is nil, else the checked ids (possibly `[]`). `create()` sends it as is. `draftNeedsGraph` is deleted; `draftInNoGraph` *(new)* = `draft.createdID == nil && draftGraphIDs == []` drives the ND-3 line | Working rule "An Artifact is the parent": boxes on every Source, none allowed; nil keeps the engine default when nothing is shown | Keeping `[]` as "hidden" — the S12-22 D-13 hack this PR retires |
| D-3 | **Swift decides no lock** (user, 2026-10-28: "No logic in Swift"). `LinkEdit.locked` is replaced by `LinkEdit.blocks: [CatalogEvidenceGraphUnlinkBlock]` *(the form's shape)*. Every block whose `artifactID` is this card's locks its `graphID`, whatever its reason. One hint for both models: `extension CatalogEvidenceGraphUnlinkBlock { var lockHint: String }` *(new)* in `SourceArtifactGraphLinks.swift` = `citations.total > 0 ? PVDeleteImpactCopy.heading(citations) : L10n.Errors.message(code: reason.wire)`. `SourceArtifactsSection.lockedGraphs: [String: String]` *(new, computed)* maps graph id → `lockHint` (first block per graph); `checkedGraphIDs`, the view and D-5 read it. No `switch` on `reason` in either model | The engine's preflight is the one unlink check (working rule); a later reason costs no Swift change (S12-22 **Next adds**, review M-3). A cited Artifact moves graphs in two Saves: link the new graph, then, with the preflight read again, unlink the old (AC-3.7) | A Swift lift of the `citations_need_graph` lock when another box is checked — a client copy of the engine rule (review M-1, user 2026-10-28); a reason switch per model — review M-3 |
| D-4 | `saveFields` makes one call: `updateArtifact(… label:, description:, graphIDs:)` with `graphIDs` = `savedGraphIDs` *(new)* when `linksDirty`, else nil. `savedGraphIDs` = `checkedGraphIDs` followed by the ids of `context.session.evidenceGraphs(sourceID:)` that link the Artifact and are not in `linkEdit.graphIDs` (made after edit opened), in list order. Success: `replace(updated, effects:)` (its `applyWrite` reloads the graph list from the effects), then leave edit. Nothing is sent when neither text nor links changed | Working rule "An Artifact card's Save is one write"; the full-set write would otherwise unlink a graph the card never showed | Sending links always — a label save would rewrite links it did not show |
| D-5 | A refusal: when links were sent and `linkEdit` is still this card's, `await loadLinkLocks()` first; then, only if the card is still the one in edit, set `cardSaveError`. Body: the first id of `linkEdit.graphIDs` that is in `excluded` and in `lockedGraphs` after the re-read, read through `context.session.evidenceGraph(id:)` → `L10n.Sources.artifactCardSaveErrorGraphKeeps(graph: name, error:)`; otherwise (no such graph, or the re-read failed and left `loadState` `.failed`) `artifactCardSaveErrorKept(error)` | Nothing is half applied, so the body never speaks of a partial save (UX-2); the engine's lock names the graph | Waiting on the graph list reload — no link changed, so nothing reloads |
| D-6 | One session read: `EvidenceGraphIndex` gains `linkedArtifactIDsBySource: [String: Set<String>]` *(new)*, built once per revision with `bySource`. `WorkspaceSession.artifactIDsInNoGraph(sourceID: String, artifactIDs: [String]) -> Set<String>?` *(new, in `WorkspaceSession+EvidenceGraphs.swift`)*: nil unless D-1's condition holds; else the ids of `artifactIDs` not in that Source's set. `SourceArtifactsSection.inNoGraph: Set<String>` *(new)* = that read over `items` (`?? []`); the resting row shows for `inNoGraph.contains(art.id)`; **Link to a graph** calls `beginEdit(art.id)` | One definition of "in no graph" in the file that owns graph link reads; a set lookup per card (review M-2) | A per-card walk of the list — review M-2; a field on `CatalogArtifact` — a second copy of the links |
| D-7 | Graph form: `loadKeptLinks` keeps every block (no filter); `locked` maps each block's `artifactID` → `lockHint`; `keptLinked` lists only blocks with `citations.total > 0` (the Kept linked group lists Citations). So an Artifact the graph must keep as its last one is locked too, and a swap inside the form takes two Saves, as on the card. `save` re-reads kept links when the refusal's code reads as a known unlink reason: `(error as? CoreInvokeError)?.code.map(EvidenceGraphUnlinkReason.init(wire:))` is non-nil and not `.unrecognized`; a nil code does not re-read | D-3's rule on the form's view; the wire vocabulary is spelled once (`.cursor/rules/macos-client.mdc`); review M-3 | Filtering out `.lastArtifact` — a Swift reason rule; a list of code literals |
| D-8 | New graph prefill: in `SourceGraphFormModel.init` `.create`, `context.session.artifactIDsInNoGraph(sourceID:artifactIDs:)` over the page's Artifacts; when it is non-nil and non-empty, `excludedArtifactIDs` = the page's Artifacts not in it; otherwise empty | Working rule "A new graph starts with the Artifacts in no graph checked"; the prefill is the draft's baseline, so the form opens clean | Computing it in `SourceGraphsSection.openAdd` — the model owns its prefill |
| D-9 | Delete `LinkEvidenceGraphArtifact` and `UnlinkEvidenceGraphArtifact`: proto `reserved 107, 108;` and `reserved "METHOD_LINK_EVIDENCE_GRAPH_ARTIFACT", "METHOD_UNLINK_EVIDENCE_GRAPH_ARTIFACT";` beside the method enum's other reserved lines; the four messages deleted; handlers, `dispatch.go` cases and constants, `TestLinkUnlinkEvidenceGraphArtifact`, `CoreInvoke` constants, the protocol methods, `GoStore`, the two `FakeStore` methods and their `linkEvidenceGraphArtifactError` / `unlinkEvidenceGraphArtifactError`, the stub, and the two calls in `writesReportTheGraphList`. `evidencegraphs.Link` / `Unlink` stay | K-12; precedent `e7063636` | Keeping them for a later caller — none is planned |
| D-10 | One hold type: `final class FakeHold: @unchecked Sendable` *(new, in `FakeStore.swift`)* with its own `NSLock`: `var isOn: Bool`, `var count: Int` (waiting calls), `func wait() async` (returns at once when `isOn` is false, else suspends on a continuation), `func release()` (resumes the oldest). FakeStore: `let updateArtifactHold = FakeHold()` and `let setSubjectPositionHold = FakeHold()`; the link hold and `holdsSetSubjectPosition`, `heldSetSubjectPositions`, `heldSetSubjectPositionCount`, `releaseSetSubjectPosition` go, and `EvidenceGraphModelTests` uses `setSubjectPositionHold.isOn`, `.count`, `.release()`. `updateArtifact` order: record `updateArtifactCalls` and `updateArtifactGraphIDs` inside `withState`; `await updateArtifactHold.wait()` outside it; then throw `updateArtifactError` or write. `setSubjectPosition` keeps its order (record, read its error, wait, throw) | The lens (Q5): a third hand-rolled hold is extracted when one is added (review M-5); the refusal arrives after the pause, as `cardSaveErrorStaysOnTheCardThatSaved` needs (review X-8) | A third copy for `updateArtifact` |
| D-11 | Copy changes per §6 Strings. `artifactGraphOnlyArtifact` (STEP-3), `artifactGraphsNeedsOne` (STEP-2) and `artifactCardSaveErrorMessage` (STEP-4) are deleted from `L10n.swift` and the catalog | What the one write makes dead (K-12) | — |

**Rules this plan bends.** None. The user's answer to review M-1 (2026-10-28) keeps the lock the engine's: a cited Artifact moves in two Saves; STEP-8 records it under the map's working rule "An Artifact with Citations keeps at least one graph".

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `EvidenceGraphIndex.linkedArtifactIDsBySource`, `WorkspaceSession.artifactIDsInNoGraph(sourceID:artifactIDs:)` | session read | `linkedArtifacts(graphID:artifacts:)`, `evidenceGraphs(sourceID:)` | Extends the same index and file; one more read beside them |
| `CatalogEvidenceGraphUnlinkBlock.lockHint` | computed copy | `SourceGraphFormModel.locked` (heading only) | Shared by the card and the form |
| `LinkEdit.blocks`, `lockedGraphs`, `savedGraphIDs`, `draftInNoGraph`, `inNoGraph`, `cardInNoGraph(id:)` | reads on the section model | `LinkEdit.locked`, `checkedGraphIDs`, `draftNeedsGraph` | `locked` and `draftNeedsGraph` are replaced |
| `FakeHold` | test hook type | FakeStore's link and `setSubjectPosition` holds | Replaces both |
| `noGraphLine` | private view | the needs-one `Text` in `addForm` | One caption for ND-3 and ND-6 (review M-7) |
| `L10n.Sources.artifactGraphsNone`, `artifactNotInGraph`, `artifactLinkToGraph`, `artifactLinkToGraphLabel(ref:)` | copy | `artifactGraphsNeedsOne` (deleted) | New meaning |

Searched: `rg "visibleGraphs|draftNeedsGraph|graphListReloaded|holdsLink|holdsSetSubject|artifactGraph|linkedArtifacts|EvidenceGraphIndex" macos`.

**Builds on.** S12-13's group and card edit; S12-22's `updateArtifact(graphIDs:)`, pair preflight and `EvidenceGraphUnlinkReason`; `PVDeleteImpactCopy.heading`; `WorkspaceSession+EvidenceGraphs`'s index; the `e7063636` RPC removal.
**Deletes.** `LinkEvidenceGraphArtifact` / `UnlinkEvidenceGraphArtifact` everywhere (D-9); `graphListReloaded()`; the save loop's `attemptedLink`, `wroteAnything`, `refusedGraphID`, `linkEffects`; `draftNeedsGraph`; `LinkEdit.locked` and `loadLinkLocks`'s reason switch; `SourceGraphFormModel.loadKeptLinks`'s reason filter and `save`'s code literal; FakeStore's link hold, the `setSubjectPosition` hold fields (moved onto `FakeHold`), `linkEvidenceGraphArtifactError`, `unlinkEvidenceGraphArtifactError`, `listEvidenceGraphsCalls` (its only reader, `cardSaveReloadsTheGraphListOnce`, goes); `L10n.Sources.artifactGraphOnlyArtifact`, `artifactGraphsNeedsOne`, `artifactCardSaveErrorMessage` and their catalog entries; `SourcePageModelTests` `linkCalls`, `cardSaveReloadsTheGraphListOnce`. Comments made false: `SourceArtifactsSection.swift` lines 19, 56–57, 85–87, 167–169, 230–232, 314–318, 560–561; `SourcePageArtifactsView.swift` 288–289; `SourceGraphFormModel.swift` 33, 75–76, 123–124 — each rewritten in the Step that changes its code.
**Next adds.** S12-30 opts the card's and graph form's callouts into announcing. A later unlink reason is one engine case and its `L10n.Errors` copy; no Swift change (D-3). A Sources-list count of Artifacts in no graph would read `artifactIDsInNoGraph`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Artifact in no graph, made from the app | Composer, canvas (`CitationComposerModel`, `EvidenceGraphModel`) | Offer only linked Artifacts (S12-10); the no-artifact callouts already name the graph (S12-D4 LK-11, LK-12) |
| Same | Graph form edit | Lists every Artifact unchecked for this graph; unaffected |
| **Link to a graph** → card edit | `SourceArtifactsSection.beginEdit` | Same path as the pencil; closes another card's edit first |
| A one-graph Source's card in edit | `loadLinkLocks` | One pair; `last_artifact` when it is Default's only Artifact |
| A graph created or deleted during card edit | `savedGraphIDs`, `liveGraphIDs` | Created: kept in the write (D-4). Deleted: dropped (today) |
| A graph whose only Artifact is locked in the graph form (new: `last_artifact` now locks there) | `SourceGraphFormModel.locked` | The box shows the lock and hint; `needsArtifact` never fires for it |

### Scale

At the map's Scale (a census with 60 graphs and 300 Citations on one Source): the card's edit is one preflight over at most 60 pairs (as since S12-22) and one write. `linkedArtifactIDsBySource` is built once per list revision (every link of the project, once); each resting card is one set lookup. The update's effects are the engine's; S12-25 narrows them and S12-26 reloads one Source's graphs.

### Schema

None — no catalog change.

### Go core

None — no `core/` change. `evidencegraphs.Link` and `Unlink` stay as `applyLinks`'s steps.

### FFI

`api/proto/engine.proto`: methods 107 and 108 reserved by number and name; `LinkEvidenceGraphArtifactRequest`, `LinkEvidenceGraphArtifactResponse`, `UnlinkEvidenceGraphArtifactRequest`, `UnlinkEvidenceGraphArtifactResponse` deleted. Handlers and dispatch entries deleted (D-9). Regenerate with `./scripts/generate-proto.sh`.

### macOS

From brief S12-D7's inventory (binding).

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Add dialog group | `PVField` + `SourceArtifactGraphLinks` | `SourcePageArtifactsView.addForm` | Shows when `visibleGraphs != nil`; `noGraphLine` when `draftInNoGraph`, id `sources.page.addArtifact.graphs.none` (replaces `…graphs.needsGraph`) |
| Card edit group | same | `SourcePageArtifactsView.graphLinks` | `locked:` is `lockedGraphs`; `noGraphLine` under the group, below `SourceFormLoadArea`, when `cardInNoGraph(id:)` = `linkEdit?.artifactID == id`, `loadState == .ready`, `checkedGraphIDs.isEmpty`; id `sources.page.artifact.<id>.graphs.none` |
| No-graph line | `noGraphLine(_ id: String)` *(new, private)* | `SourcePageArtifactsView.swift` | Today's needs-one style: `PVFont.body(size: PVTypeScale.caption, italic: true)`, `PVColor.textMuted` |
| Resting row | `PVFlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space3)` holding `PVBadge(L10n.Sources.artifactNotInGraph, tone: .neutral, icon: .link)` and `PVButton(L10n.Sources.artifactLinkToGraph, variant: .ghost, size: .sm)` | `SourcePageArtifactsView.restingText` | Last child of the `VStack`, under `descriptionLine`, only for `inNoGraph`. Both children `.fixedSize()`, so the Button wraps under the Badge. Cards are not levelled: a card with the row is taller, as a card in edit is (the grid aligns tops). Button `.accessibilityLabel(L10n.Sources.artifactLinkToGraphLabel(ref: art.ref))`; ids `sources.page.artifact.<id>.noGraph` (badge), `sources.page.artifact.<id>.linkToGraph` |
| Refused save | `PVCallout` danger compact | `editFields` (unchanged position) | Body from D-5 |
| Graph form kept group | `keptLinked` | `SourceGraphFormDialog.swift` | ND-12 copy |

Frames in words (until the board is agreed): **Add, one graph** — the dialog with one checked "Default" row and its ref. **Add, nothing checked** — every row unchecked, Add enabled, the ND-3 caption under the group. **Resting cards** — a card in no graph shows, under its description, a neutral badge "Not in an evidence graph" and a ghost "Link to a graph" button; other cards unchanged. **Cited Artifact** — Brown locked with "1 citation would be left in no evidence graph"; checking Jones leaves Brown locked. **Save refused** — the danger callout above Label, naming the graph; that row checked and locked.

States: no graphs for the Source → no group, no badge; read error → S12-D4's compact callout with Read again (the ND-6 line hides, since `loadState` is not ready); many graphs → the group scrolls past six rows (S12-D4). Accessibility per ND-10, ND-11.

### Strings

Final copy is the agreed board's; these are the brief's and are replaced by the board's at agreement.

- *(new)* `L10n.Sources.artifactGraphsNone`, key `sources.page.artifact.graphs.none` = "This artifact won’t be in an evidence graph. Citations can’t use it until it is linked." Comment: "Muted line under the evidence graph group in the Add artifact dialog and on an artifact card in edit when no box is checked".
- *(new)* `L10n.Sources.artifactNotInGraph`, key `sources.page.artifact.notInGraph` = "Not in an evidence graph". Comment: "Badge on a resting artifact card when no evidence graph links the artifact".
- *(new)* `L10n.Sources.artifactLinkToGraph`, key `sources.page.artifact.linkToGraph` = "Link to a graph". Comment: "Button on a resting artifact card in no evidence graph; opens the card's edit with the graph boxes".
- *(new)* `L10n.Sources.artifactLinkToGraphLabel(ref:)` (a `static func` returning `String` through `L10n.format`, like `editArtifactLabel(ref:)`), key `sources.page.artifact.linkToGraphLabel` = "Link %@ to an evidence graph". Comment: "VoiceOver label of the Link to a graph button; argument is the artifact ref".
- Changed `sources.page.artifact.saveErrorGraphKeeps`: "%1$@ keeps this artifact. %2$@" → "%1$@ keeps this artifact. Your changes are kept. %2$@". Comment: "Callout message on an artifact card in edit when an evidence graph refused to drop the artifact and nothing was saved; %1$@ is the graph name, %2$@ is the error text".
- Changed `sources.graphForm.keptLinkedMessage`: → "These artifacts stay linked: citations need them in this evidence graph."
- Removed: `sources.page.artifact.graphs.needsOne`, `sources.page.artifact.graphs.onlyArtifact`, `sources.page.artifact.saveErrorMessage`.
- Old-scope copy checked and kept: `sources.page.artifact.graphs.hint` ("Which evidence graphs can cite this artifact") reads true with none checked; `sources.page.artifact.saveErrorKept` unchanged; `error.evidencegraphs.last_artifact` now also reads as a lock hint and fits.

## 7. Steps

**STEP-1 — One session read of the Artifacts in no graph**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-swift-test`.
Do: D-6's index field and read. Write T-0.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes.

**STEP-2 — The Add dialog on every Source, none allowed**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-1, D-2; `noGraphLine` in `addForm`; `artifactGraphsNone` in; `artifactGraphsNeedsOne` out; rewrite the comments on `excludedGraphIDs`, `visibleGraphs`, `draftGraphIDs` and the `create()` `[]` comment. Write T-1, T-2. Existing tests: `addArtifactLinksOnlyTheCheckedGraphs` stays (non-nil ids now); `addArtifactExcludedGraphIsDirty` stays; `addArtifactRetryIgnoresGraphChoice` rewritten to assert `store.createArtifactGraphIDs.count == 1` after the retry (no `linkCalls`).
Check: `xcodebuild test` passes; `rg "draftNeedsGraph|artifactGraphsNeedsOne" macos` has no hit.

**STEP-3 — The card's group and locks on every Source**
Files: `SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourceArtifactGraphLinks.swift`, `SourcePageArtifactsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`.
Do: `beginEdit` builds `LinkEdit` when `visibleGraphs` is non-nil; D-3 (`blocks`, `lockHint`, `lockedGraphs`); the view passes `lockedGraphs`; `noGraphLine` under the card's group; `artifactGraphOnlyArtifact` out; rewrite the comments on `linkEdit`, `beginEdit`, `loadLinkLocks`, `graphLinks`. Write T-3, T-4, T-5. Existing tests: `cardEditOnOneGraphHasNoLinks` → T-3; `cardEditLocksInUseAndOnlyArtifactGraphs` → T-4; `cardLinkReadFailsAndReadAgain` stays; `cardCancelDropsLinkChanges` stays with its `linkCalls(store).isEmpty` line replaced by `store.updateArtifactCalls == 0`; `cardFieldsOnlySaveErrorSaysChangesKept` gains `await model.artifacts.linkLoadTask?.value` after `beginEdit` (a one-graph card now reads its locks before it can save).
Check: `xcodebuild test` passes; `rg "case .citationsNeedGraph|case .lastArtifact|case .artifactInUse" macos/App/Features` has no hit.

**STEP-4 — The card's Save is one write**
Files: `SourceArtifactsSection.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`.
Do: D-4, D-5, D-10 (`FakeHold`; the link hold, the `setSubjectPosition` hold fields and `listEvidenceGraphsCalls` go; the link methods and their error vars stay until STEP-7); reword `artifactCardSaveErrorGraphKeeps`; delete `artifactCardSaveErrorMessage`, `graphListReloaded`; rewrite the `saveFields` and `saveErrorBody` comments. Write T-6 – T-9b. Existing tests: `cardSaveErrorStaysOnTheCardThatSaved` uses `store.updateArtifactHold` (`isOn = true`, wait for `count == 1`, `release()`) with `updateArtifactError` set; `cardSaveSkipsAGraphDeletedDuringEdit` expects `store.updateArtifactGraphIDs == [nil]`; delete `cardSaveLinksThenUnlinks` (T-6), `cardSaveReloadsTheGraphListOnce`, `cardSaveRefusedKeepsEditAndRereadsLocks` (T-8), and the `linkCalls` helper.
Check: `xcodebuild test` passes; `rg "linkEvidenceGraphArtifact|unlinkEvidenceGraphArtifact|linkCalls|listEvidenceGraphsCalls|holdsSetSubjectPosition" macos/App/Features macos/ProvenenciaTests` has no hit.

**STEP-5 — The resting card marks an Artifact in no graph**
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-6's `inNoGraph`; the row per §6 macOS and the brief inventory. Write T-10.
Check: `xcodebuild test` passes; manual check step 2 shows the row.

**STEP-6 — The graph form**
Files: `macos/App/Features/Sources/SourceGraphFormModel.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`.
Do: D-7, D-8; the `keptLinkedMessage` copy; rewrite the comments on `blocks`, `locked`, `loadKeptLinks`. Write T-11, T-12, T-13. `createPrefillsEveryArtifactChecked` stays green unchanged (no graph links any Artifact there); `editPrefillsAndLocksInUseArtifacts` and `editRefusedInUseRereadsKeptLinks` stay.
Check: `xcodebuild test` passes; `rg '"evidencegraphs\.' macos/App/Features` has no hit.

**STEP-7 — Delete the link and unlink RPCs**
Files: `api/proto/engine.proto`, generated code, `api/ffi/handlers/evidence_graphs.go`, `api/ffi/handlers/evidence_graphs_test.go`, `api/ffi/dispatch.go`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`. Skill: `add-ffi-handler`.
Do: D-9; `./scripts/generate-proto.sh`. In `writesReportTheGraphList`, drop `linked` and `unlinked` from the loop.
Check: AC-5.1's `rg` has no hit; `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `xcodebuild test` passes.

**STEP-8 — Docs and guidance**
Files: `.agents/skills/add-workspace-place/SKILL.md` (line 172: "Artifact ↔ graph link writes (`updateArtifact` with `graphIDs`, `updateEvidenceGraph`): `session.apply(effects)`" in place of the two deleted names), `docs/deployment-plan/spike-12/deployment-plan.md` (map row status; under working rule "An Artifact with Citations keeps at least one graph" one sentence: "On the card and in the graph form the lock is the engine's preflight and does not lift; moving such an Artifact to another graph takes two saves (user, 2026-10-28)."), `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/design/README.md` (S12-D7 to Archived), the brief to `design/archive/` with every relative link rewritten for the archive folder, one level deeper (match the links of the archived S12-D4 brief), this plan's **Brief:** link pointed at the archived file, this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no broken link; `python3 scripts/check-guidance-names.py` passes.

Where checks run: `generate-proto.sh` and `xcodebuild` need the Mac.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Swift Testing (`add-swift-test`). Fixtures: `makeCensus` (Brown `g-b` P9; Jones `g-j` P3 P4; Smith `g-s` P3 P6) and `unlinkBlock(_:graphID:citations:reason:)`. "No graphs" is `store.listEvidenceGraphsError = CoreInvokeError.failed(status: 3)` set before the page warms (D-1).

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-0 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `artifactIDsInNoGraphReadsTheIndex` *(new)* | Graphs of `s1` link a1, a2: `artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1","a2","a3"]) == ["a3"]`; a list with no graph for `s1` → nil; no list value → nil | AC-2.3, AC-4.4, D-1, D-6 |
| T-1 | Swift | `SourcePageModelTests` `addArtifactOnOneGraphShowsTheGroup` *(replaces `addArtifactOnOneGraphSendsEveryGraph` and `addArtifactSendsDefaultGraphLinksWhenBoxesHidden`)* | One graph `g-1`: `visibleGraphs` ids `["g-1"]`; create → `createArtifactGraphIDs == [["g-1"]]`. No graphs: `visibleGraphs == nil`, `draftGraphIDs == nil`, create → `[nil]` | AC-1.1, AC-1.2, AC-1.4, D-2 |
| T-2 | Swift | same `addArtifactWithNoGraphChecked` *(replaces `addArtifactNeedsAGraph`)* | Census, every box excluded: `canSubmitDraft`, `draftInNoGraph`; create → `createArtifactGraphIDs == [[]]`, dialog closed | AC-1.3, D-2 |
| T-3 | Swift | same `cardEditOnOneGraphShowsTheGroup` *(replaces `cardEditOnOneGraphHasNoLinks`)* | One graph `g-1` linking P3 and P4: `linkEdit` non-nil, one preflight call with pair (g-1, P3); exclude `g-1`, Save → `updateArtifactGraphIDs == [[]]` | AC-3.3 |
| T-4 | Swift | same `cardEditLocksEveryBlock` *(rewrites `cardEditLocksInUseAndOnlyArtifactGraphs`)* | (g-s,P3) in use, 2 → `lockedGraphs["g-s"]` = "2 citations in this graph use this artifact"; (g-b,P9) `.lastArtifact` → `L10n.Errors.message(code: "evidencegraphs.last_artifact")`; (g-b,P9) `.citationsNeedGraph`, 1 → "1 citation would be left in no evidence graph"; (g-b,P9) `.unrecognized("x.y")` → locked with `L10n.Errors.message(code: "x.y")`; an Artifact no graph links → no call, no locks | AC-3.4, AC-3.5, AC-3.6, D-3 |
| T-5 | Swift | same `cardCitationsLockHoldsUntilTheLinkIsSaved` *(new)* | Graphs Brown (P9, P4), Jones (P3, P4), Smith (P6); (g-b,P9) `.citationsNeedGraph`, 1. Un-exclude g-j → `lockedGraphs["g-b"]` still set, `checkedGraphIDs == ["g-b","g-j"]`; Save → `updateArtifactGraphIDs == [["g-b","g-j"]]`. Seed Jones linking P9 in the store, apply, clear the scripted blocks, edit again, exclude g-b, Save → last `updateArtifactGraphIDs` is `["g-j"]` | AC-3.7, D-3 |
| T-6 | Swift | same `cardSaveIsOneWrite` *(replaces `cardSaveLinksThenUnlinks`)* | P3: exclude g-j, un-exclude g-b, label "Page 3, front" → `updateArtifactCalls == 1`, `updateArtifactGraphIDs == [["g-b","g-s"]]`, label on the page, edit closed | AC-3.1, D-4 |
| T-7 | Swift | same `cardSaveSendsNoLinksWhenOnlyTextChanged` *(rewrites `cardSaveFieldsOnlyHasNoLinkCalls`)* | Label only → `updateArtifactGraphIDs == [nil]` | AC-3.2, D-4 |
| T-8 | Swift | same `cardSaveRefusedNamesTheGraphAndLocksIt` *(replaces `cardSaveRefusedKeepsEditAndRereadsLocks`)* | Exclude g-j; `updateArtifactError` `artifact_in_use`; blocks then (g-j,P3) in use, 1 → card in edit, `cardSaveError == artifactCardSaveErrorGraphKeeps(graph: "Jones household", error:)`, `checkedGraphIDs` contains g-j, `excluded` still has g-j. Second case: graphs as T-5 with no block at open; exclude g-b; refusal `citations_need_graph`; blocks then (g-b,P9) `.citationsNeedGraph` → g-b in `lockedGraphs` and `checkedGraphIDs`; the error names Brown household. Third: refusal with no block → `artifactCardSaveErrorKept` | AC-3.8, AC-3.9, AC-3.10, D-5 |
| T-9 | Swift | same `cardSaveKeepsAGraphMadeDuringEdit` *(new)* | Edit P3, await `linkLoadTask`; `store.createEvidenceGraph` linking P3, effects applied, list awaited; exclude g-j; Save → `updateArtifactGraphIDs == [["g-s", newID]]` | AC-3.11, D-4 |
| T-9b | Swift | same `cardSaveReloadFailureOffersReadAgain` (rewritten) | Edit P3, await `linkLoadTask`; then exclude g-j, set `updateArtifactError` and `getEvidenceGraphUnlinkImpactError`; Save → `linkEdit.loadState` is `.failed`, `cardSaveError == artifactCardSaveErrorKept(...)` | AC-3.10, D-5 |
| T-10 | Swift | same `restingCardMarksAnArtifactInNoGraph` *(new)* | Graphs omit P6: `inNoGraph == ["P6"]`; P3 not in it; no graphs → `inNoGraph` empty; `beginEdit("P6")` → `excluded` is every graph id | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-11 | Swift | `SourceGraphFormModelTests` `editLocksEveryBlock` *(new)* | `openSmithEdit` with (g-smith,a3) `.citationsNeedGraph` 1 and (g-smith,a4) `.lastArtifact` → `locked` = [a3: "1 citation would be left in no evidence graph", a4: "An evidence graph needs at least one artifact."], `keptLinked` is a3 only | AC-4.1, D-7 |
| T-12 | Swift | same `createPrefillsTheArtifactsInNoGraph` *(new)* | Graph linking a1 – a4 → `checkedArtifactIDs == ["a5","a6"]`, `!isDirty`; graph linking a1 – a6 → every Artifact checked | AC-4.3, AC-4.4, D-8 |
| T-13 | Swift | same `editRefusedCitationsNeedGraphRereadsKeptLinks` *(new)* | As `editRefusedInUseRereadsKeptLinks` with code `evidencegraphs.citations_need_graph` and a `.citationsNeedGraph` block → a2 locked; a refusal with code nil (`CoreInvokeError.failed`) → no re-read (`getEvidenceGraphUnlinkImpactCalls` unchanged) | AC-4.2, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (the S12-13 section), `SourceGraphFormModelTests`, `EvidenceGraphStoreTests`, `EvidenceGraphModelTests` (the moved hold), `TestSwiftVocabularyMatchesGo`, `L10nResolutionTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each deleted name `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left — `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact`, `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact`, `graphListReloaded`, `artifactGraphOnlyArtifact`, `artifactGraphsNeedsOne`, `artifactCardSaveErrorMessage`, `draftNeedsGraph`, `listEvidenceGraphsCalls`, `holdsSetSubjectPosition`; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none; the M-1 answer is recorded in STEP-8)
- [ ] UI matches the brief inventory; no new kit primitive (the brief marks none New)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the working-rule sentence; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ with its links rewritten, and design/README.md updated
- [ ] Domain doc updated: none — `source-layer-data-model.md` already describes `GraphLinks` and none allowed (S12-22)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice Check: on a two-graph Source, a card that unlinks one graph and links another saves in one revision; a refused save names the graph and leaves nothing half applied; a `citations_need_graph` refusal ends with the box checked and locked, and no "save again to finish" text; no app code calls the deleted RPCs.

Seed: a project with Source A (one graph, Default, linking ART-1) and Source B, a census with graphs Brown, Jones and Smith; on B, ART-3 linked to Jones and Smith, ART-9 linked to Brown and Jones, ART-4 and ART-6 linked to Jones.

1. On Source A, Add artifact "Page 7": the group shows Default, checked. Uncheck it: the ND-3 line shows, Add stays enabled. Add.
2. See Page 7's card show "Not in an evidence graph" and **Link to a graph**. Click it: edit opens with Default unchecked. Check Default and Save. The badge goes.
3. On Source B, open ART-3; uncheck Jones, check Brown, change the label; Save. Open the History panel: one revision holds the label, the Brown link and the Jones unlink.
4. Open ART-9; uncheck Jones; Save. ART-9 is in Brown only. In Brown's composer, make a Citation on ART-9 and save it with no Observation. Open ART-9's card: Brown is locked with "1 citation would be left in no evidence graph". Check Smith: Brown stays locked. Save. Open the card again: Brown is no longer locked; uncheck it and Save. ART-9 is in Smith only.
5. Open ART-4's card (Jones only, no Citation, no lock); uncheck Jones and leave the card in edit. In a second window, in Jones's composer, make a Citation on ART-4 with no Observation. Back on the card, Save: the callout reads "Jones household keeps this artifact. Your changes are kept. …" with nothing about saving again; Jones is checked and locked.
6. Edit Smith in the graph form: ART-9 is locked under Kept linked with its Citation.
7. On Source B, Add artifact "Page 8" with every box unchecked. Add evidence graph: only Page 8 starts checked.
8. AC-5.1, AC-5.2: run `rg "LinkEvidenceGraphArtifact|linkEvidenceGraphArtifact" api macos --glob '!*.pb.*'` (no hit) and read the method enum's reserved lines in `api/proto/engine.proto`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `applyWrite` after an update with links does not reload the graph list, so the badge stays | The engine reports `evidence_graphs` for link rows (S12-22 §6 Scale); manual step 2 checks it |
| The graph form now locks a graph's only Artifact (`last_artifact`), so swapping a graph's only Artifact takes two saves | Intended (D-7, user 2026-10-28); brief ND-12 draws it |

**Stop and report if:**

- `loadLinkLocks`, `saveFields`, `SourceGraphFormModel.loadKeptLinks` or `EvidenceGraphIndex` no longer match §4 at your base.
- `updateArtifact` with links returns effects without the `evidence_graphs` list, so the session list does not reload after a card Save.
- The agreed board's frames contradict a Decision here (for example a lock that lifts in Swift, or Add disabled with no box checked).
- An existing Go or Swift test outside §7's lists fails for a reason other than the deleted RPCs, the moved hold or the changed copy.
- Deleting a proto method breaks `TestSwiftVocabularyMatchesGo` or another pinning test in a way the reserved lines do not fix.

## 12. Open questions

1. **Board S12-D7.** The brief's copy and the resting row's layout are recommendations until the board is agreed. Recommended: agree the brief as written; this plan's Strings follow the board.

## 13. PR

**Title:** S12-38 — Add an Artifact without a graph; the card saves links in one write
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-38-artifact-without-graph.md`.
