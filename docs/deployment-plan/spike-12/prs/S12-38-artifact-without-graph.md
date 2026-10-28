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
**Review:** — (not reviewed yet; review-plan runs once the board is agreed and the Strings are final)

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

The researcher can add a page that belongs to no Evidence graph yet, see such a page marked on its card, and link it from there. The graph boxes show on every Source, one graph included. A card's Save writes the label, the description and every link in one write, so a refusal leaves nothing half applied and never asks to "save again to finish". A cited page can move from one graph to another in one Save. A new graph starts with the pages that are in no graph.

**Why now.** Audit A-3 and A-8 (High), and the user's fix (2026-10-27): S12-22 built the engine half; this PR is the Swift half and deletes what the one write makes dead.

**What stays.** The card grid, the resting card of an Artifact that is in a graph, the Add file dialog, Artifact delete, cover and Use as thumbnail; the graph form's layout; graph create's `needsArtifact` rule; the engine's rules (no Go rule changes).

## 2. User stories

**US-1.** As a researcher, when I add a page, I choose its graphs on every Source, and may choose none, so that a page can wait for its graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source with one graph "Default"; open Add artifact | The Evidence graphs group shows one checked box, Default |
| AC-1.2 | Same; label "Page 7", Add | `createArtifact` is sent `graphIDs: ["g-1"]` |
| AC-1.3 | Census (Brown, Jones, Smith); every box unchecked | Add is enabled; the line "This artifact won’t be in an evidence graph. Citations can’t use it until it is linked." shows; Add sends `graphIDs: []` |
| AC-1.4 | Graph list not loaded yet | No group; Add sends `graphIDs: nil` (the engine links every graph) |

**US-2.** As a researcher, I see which pages are in no graph and can link one from its card.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Page 6 is linked by no graph; the graph list is loaded | Page 6's resting card shows the "Not in an evidence graph" badge and **Link to a graph** |
| AC-2.2 | Page 3 is linked by Jones | No badge, no button |
| AC-2.3 | The graph list is loading | No card shows the badge |
| AC-2.4 | Click **Link to a graph** on Page 6 | Page 6 opens in edit with every box unchecked; check Jones and Save → `updateArtifact` with `graphIDs: ["g-j"]` |

**US-3.** As a researcher, a card Save writes everything or nothing, and its locks say why a graph keeps the page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Page 3 in Jones and Smith; uncheck Jones, check Brown, change the label; Save | One `updateArtifact` with the label and `graphIDs: ["g-b", "g-s"]`; no other write; the card leaves edit |
| AC-3.2 | Change only the label; Save | `updateArtifact` with `graphIDs: nil` |
| AC-3.3 | Card edit on a one-graph Source | The group shows; the preflight runs for (Default, Page 3); unchecking Default and saving sends `graphIDs: []` |
| AC-3.4 | Preflight blocks (Smith, Page 3) `artifact_in_use` with 2 Citations | Smith locked, hint "2 citations in this graph use this artifact" |
| AC-3.5 | Preflight blocks (Brown, Page 9) `last_artifact` | Brown locked, hint "An evidence graph needs at least one artifact." |
| AC-3.6 | Page 5 only in Jones; preflight blocks (Jones, Page 5) `citations_need_graph` with 1 Citation | Jones locked, hint "1 citation would be left in no evidence graph" |
| AC-3.7 | Same; check Smith | Jones unlocks and stays checked; uncheck Jones; Save → `graphIDs: ["g-s"]`. Uncheck Smith instead → Jones locked and checked again |
| AC-3.8 | Save refused `evidencegraphs.artifact_in_use`; the re-read preflight blocks (Jones, Page 3), which the researcher unchecked | Card stays in edit; callout "Couldn’t save the artifact" / "Jones household keeps this artifact. Your changes are kept. {error}"; Jones checked and locked; label buffer kept |
| AC-3.9 | Page 5 only in Jones, no block at edit open; uncheck Jones; Save refused `evidencegraphs.citations_need_graph`; the re-read blocks (Jones, Page 5) `citations_need_graph` | Jones ends checked and locked; the callout names Jones household; no text says "save again" (S12-22 UX-2) |
| AC-3.10 | Save refused with a code no re-read lock explains | "Your changes are kept. {error}" |
| AC-3.11 | A graph created during edit links Page 3; uncheck Jones; Save | `graphIDs` keeps the new graph: `["g-s", <new id>]` |

**US-4.** As a researcher, the graph form locks a page whose Citations would be left in no graph, and a new graph starts with the pages in no graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Edit Smith (a3, a4); preflight blocks (Smith, a3) `citations_need_graph` with 1 Citation and (Smith, a4) `last_artifact` | a3 locked with "1 citation would be left in no evidence graph" and listed under Kept linked; a4 not locked |
| AC-4.2 | Graph form save refused `evidencegraphs.citations_need_graph` | The kept links are read again; the refusal stays in view |
| AC-4.3 | Six pages; a1 – a4 linked by a graph, a5 and a6 by none; Add evidence graph | a5 and a6 checked, a1 – a4 unchecked; the form is not dirty |
| AC-4.4 | Every page is in a graph, or the graph list has no value | Every page checked (today's default) |

**US-5.** As a contributor, the per-graph link RPCs are gone.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `rg "LinkEvidenceGraphArtifact\|linkEvidenceGraphArtifact" api macos --glob '!*.pb.*'` | No hit (the `Unlink…` forms included by the pattern's suffix match) |
| AC-5.2 | `api/proto/engine.proto` | `107, 108` and both method names reserved; the four messages deleted |

## 3. Scope

| In | Covered by |
| --- | --- |
| The group on every Source with a loaded graph list, in the Add dialog and the card (ND-1, ND-6) | STEP-1, STEP-2, AC-1.1, AC-3.3 |
| No box checked allowed, with its line (ND-3, ND-6) | STEP-1, STEP-2, AC-1.3 |
| The resting card's badge and **Link to a graph** (ND-4, ND-5, ND-10, ND-11) | STEP-4, AC-2.1 – AC-2.4 |
| The card Save as one `updateArtifact` with its links (D-4, D-5) | STEP-3, AC-3.1, AC-3.2, AC-3.11 |
| Lock hints by reason; the swap-aware Citations lock (ND-7, ND-8) | STEP-2, AC-3.4 – AC-3.7 |
| The refused-save callout and re-read (ND-9) | STEP-3, AC-3.8 – AC-3.10 |
| Graph form: the Citations lock and Kept linked line (ND-12); refusal re-read | STEP-5, AC-4.1, AC-4.2 |
| New graph's default checks | STEP-5, AC-4.3, AC-4.4 |
| Delete the link and unlink RPCs on both sides, `graphListReloaded()`, the partial-save and needs-one strings, `artifactGraphOnlyArtifact` | STEP-6, STEP-7, AC-5.1, AC-5.2 |

| Out | Home |
| --- | --- |
| Naming, in a graph delete's confirm, the Artifacts it leaves in no graph | [`docs/ideas/artifact-outside-graph-marker.md`](../../../ideas/artifact-outside-graph-marker.md) |
| A count or filter of Artifacts in no graph (section header, Sources list) | not planned |
| Announcing the card's and graph form's error callouts | S12-30 |
| Removing FakeStore's `createArtifact` linking rule | S12-35 |
| Narrowing the Artifact update's effects | S12-25 |
| Any engine rule change | not planned (S12-22 owns the rules) |

## 4. Current state

Verified at `b1bf4c5`.

| Area | Where | Today |
| --- | --- | --- |
| Group visibility | `macos/App/Features/Sources/SourceArtifactsSection.swift` `visibleGraphs` | Non-nil only with `graphs.count >= 2` |
| Add's links | same file `draftGraphIDs` (`[String]`, `[]` when hidden), `draftNeedsGraph`, `canSubmitDraft`, `create()` (`draftGraphIDs.isEmpty ? nil : draftGraphIDs`) | Every box unchecked blocks Add |
| Card edit state | same file `LinkEdit` (`locked: [String: String]`), `beginEdit` (only when `visibleGraphs`), `checkedGraphIDs`, `linkedGraphIDs`, `liveGraphIDs`, `linksDirty` | — |
| Card locks | same file `loadLinkLocks` | One `getEvidenceGraphUnlinkImpact` over pairs; `.lastArtifact` → `L10n.Sources.artifactGraphOnlyArtifact`; `.citationsNeedGraph` → no lock (S12-22 D-13) |
| Card save | same file `saveFields`, `saveErrorBody`, `graphListReloaded` | `updateArtifact(graphIDs: nil)`, then one `linkEvidenceGraphArtifact` per added graph and one `unlinkEvidenceGraphArtifact` per dropped graph; partial-save copy; waits on the graph list reload |
| Card views | `macos/App/Features/Sources/SourcePageArtifactsView.swift` `addForm` (needs-one line), `graphLinks`, `editFields`, `restingText`, `metaRow` | — |
| Graph rows | `macos/App/Features/Sources/SourceArtifactGraphLinks.swift` `SourceArtifactGraphLinks` | Takes `locked: [String: String]` |
| Graph form | `macos/App/Features/Sources/SourceGraphFormModel.swift` `init` (create prefill: nothing excluded), `loadKeptLinks` (keeps `.artifactInUse` only), `save` (re-reads on `evidencegraphs.artifact_in_use`); `macos/App/Features/Sources/SourceGraphFormDialog.swift` `keptLinked` | — |
| Store | `macos/App/Platform/GenealogyStore.swift` `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact` (protocol); `macos/App/Platform/GoStore.swift` (lines 517–552); `macos/App/Platform/FakeStore.swift` (`linkEvidenceGraphArtifactError`, `holdsLinkEvidenceGraphArtifact`, `heldLinkEvidenceGraphArtifactCount`, `releaseLinkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifactError`, both methods); `macos/App/Platform/CoreInvoke.swift` (lines 66–67); `macos/ProvenenciaTests/OnboardingModelTests.swift` (stub, lines 479–490) | `updateArtifact(… graphIDs: [String]?)` and `createArtifact(… graphIDs: [String]?)` already carry links (S12-22 D-12); FakeStore records them in `updateArtifactGraphIDs`, `createArtifactGraphIDs` |
| Go | `api/ffi/handlers/evidence_graphs.go` `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact`; `api/ffi/dispatch.go` (lines 108–109, 303–306); `api/proto/engine.proto` (method enum lines 119–120, messages lines 506–526) | `evidencegraphs.Link` / `Unlink` stay: `applyLinks` calls them |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources.artifactGraphsNeedsOne`, `artifactGraphOnlyArtifact`, `artifactCardSaveErrorKept`, `artifactCardSaveErrorGraphKeeps`, `artifactCardSaveErrorMessage`, `graphFormKeptLinkedMessage`; `macos/App/Resources/Localizable.xcstrings` | — |
| Guidance | `.agents/skills/add-workspace-place/SKILL.md` line 172 | Names `linkEvidenceGraphArtifact` and `unlinkEvidenceGraphArtifact` |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` (section "Artifact ↔ graph links (S12-13)", lines 1455–1830); `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`; `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `writesReportTheGraphList`; `api/ffi/handlers/evidence_graphs_test.go` `TestLinkUnlinkEvidenceGraphArtifact` | — |

**Precedent.** `e7063636` deleted three RPCs on both sides in one commit (`GetPersonHeader` and siblings): proto methods reserved by number and name, messages deleted, handler and test files, `dispatch.go`, `CoreInvoke.swift`, `GenealogyStore`, `GoStore`, `FakeStore`, the `OnboardingModelTests` stub. Copy that list. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built the group and the card's edit; extend its views, do not add a second group.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `visibleGraphs` returns the Source's graphs whenever the session has a list value (`count >= 1`); nil only while it has none. `draftGraphIDs` becomes `[String]?`: nil when `visibleGraphs` is nil, else the checked ids (possibly `[]`). `create()` sends `draftGraphIDs` as is. `draftNeedsGraph` is deleted; `draftInNoGraph` *(new)* = `draft.createdID == nil && draftGraphIDs == []` drives the ND-3 line | Working rule "An Artifact is the parent": boxes on every Source, none allowed; nil keeps the engine default when nothing is shown | Keeping `[]` as "hidden" — the S12-22 D-13 hack this PR exists to retire |
| D-2 | `beginEdit` builds `LinkEdit` whenever `visibleGraphs` is non-nil (one graph included). `LinkEdit` gains `needsGraph: [String: String]` *(new)*: the graph id → hint of a `.citationsNeedGraph` block. `locked` keeps the other two reasons | The Citations lock behaves differently (ND-8), so it is stored apart from the hard locks | One map with a reason enum — every reader would re-filter |
| D-3 | Lock hints: `.artifactInUse` and `.citationsNeedGraph` → `PVDeleteImpactCopy.heading(block.citations)`; `.lastArtifact` → `L10n.Errors.message(code: block.reason.wire)`; `.unrecognized` → `L10n.Errors.message(code: block.reason.wire)` in `locked` (a refusal the app has no words for still locks). `effectiveLocks` *(new, computed)* = `locked` plus each `needsGraph` entry whose graph is the only checked candidate: no other live graph `h` with `!excluded.contains(h) \|\| locked[h] != nil`. `checkedGraphIDs`, the view and `saveErrorBody` read `effectiveLocks` | S12-22 **Next adds**: "reads each lock's copy from `L10n.Errors.message(code: block.reason.wire)` or the Citations heading"; ND-7, ND-8. The engine allows the swap because `applyLinks` links first (S12-22 D-5), so a Save the card allows the engine accepts | Locking the Citations graph always — blocks moving a cited page; no lock and a refusal — UX-2 |
| D-4 | `saveFields` makes one call: `updateArtifact(… label:, description:, graphIDs:)` with `graphIDs` = `savedGraphIDs` *(new)* when `linksDirty`, else nil. `savedGraphIDs` = `checkedGraphIDs` followed by the session graphs of this Source that link the Artifact and are not in `linkEdit.graphIDs` (made after edit opened). Success: `replace(updated, effects:)` (its `applyWrite` reloads the graph list from the effects), then leave edit. Nothing is sent when neither text nor links changed | Working rule "An Artifact card's Save is one write"; the full-set write would otherwise unlink a graph the card never showed | Sending the links always — a label save would rewrite links it did not show; a separate link write — A-8 |
| D-5 | A refusal: `cardSaveError` is set only if the card is still the one in edit. When links were sent and `linkEdit` is this card's, `await loadLinkLocks()` first. Body: the first graph of `linkEdit.graphIDs` that is in `excluded` and in `effectiveLocks` after the re-read → `L10n.Sources.artifactCardSaveErrorGraphKeeps(graph:error:)`; otherwise `artifactCardSaveErrorKept`. A failed re-read leaves `loadState` `.failed` (Read again) | Nothing is half applied, so the body never speaks of a partial save (UX-2); the lock names the graph | Waiting on the graph list reload — no link changed, so nothing reloads |
| D-6 | Resting badge: `SourceArtifactsSection.isInNoGraph(_ art:)` *(new)*: the session list has a value and no graph of this Source links `art.id`. The row sits in `restingText` under `descriptionLine` (brief inventory "Resting card text", Extend). **Link to a graph** calls `beginEdit(art.id)` | One read of the session list, the same one the card's boxes use | A field on `CatalogArtifact` — a second copy of the links |
| D-7 | Graph form: `loadKeptLinks` keeps `.artifactInUse` and `.citationsNeedGraph` blocks (not `.lastArtifact`: `needsArtifact` already stops an empty graph, and locking the only Artifact would stop a swap). `save` re-reads kept links on `evidencegraphs.artifact_in_use` and `evidencegraphs.citations_need_graph`, read through `EvidenceGraphUnlinkReason(wire:)` against `.artifactInUse` / `.citationsNeedGraph`, not string literals | ND-12; `.cursor/rules/macos-client.mdc` (wire vocabulary spelled once) | — |
| D-8 | New graph prefill: in `SourceGraphFormModel.init` `.create`, when the session list has a value and some Artifacts of the page are linked by none of this Source's graphs, `excludedArtifactIDs` = the Artifacts that some graph links; otherwise empty. Read through `context.session.evidenceGraphs(sourceID:)` | Working rule "A new graph starts with the Artifacts in no graph checked"; the prefill is the draft's baseline, so the form opens clean | Computing it in `SourceGraphsSection.openAdd` — the model owns its prefill |
| D-9 | Delete `LinkEvidenceGraphArtifact` and `UnlinkEvidenceGraphArtifact`: proto `reserved 107, 108;` and `reserved "METHOD_LINK_EVIDENCE_GRAPH_ARTIFACT", "METHOD_UNLINK_EVIDENCE_GRAPH_ARTIFACT";` beside the method enum's other reserved lines; the four messages deleted; handlers, `dispatch.go` cases and constants, `TestLinkUnlinkEvidenceGraphArtifact`, `CoreInvoke` constants, the protocol, `GoStore`, `FakeStore` methods and hooks, the stub, and the two calls in `writesReportTheGraphList`. `evidencegraphs.Link` / `Unlink` stay | K-12; precedent `e7063636` | Keeping them for a later caller — none is planned |
| D-10 | FakeStore's hold moves to the one write: `holdsUpdateArtifact`, `heldUpdateArtifactCount`, `releaseUpdateArtifact()` *(new)* replace the link hold, in the same shape (`withState`, continuations). `updateArtifact` still moves no link (S12-22 D-14) | `cardSaveErrorStaysOnTheCardThatSaved` needs to pause a save; the write it pauses is now `updateArtifact` | A new fake rule that moves links — "FakeStore is a seeded fake" |
| D-11 | Copy changes per §6 Strings. `artifactGraphOnlyArtifact`, `artifactGraphsNeedsOne`, `artifactCardSaveErrorMessage` are deleted from `L10n.swift` and the catalog | What the one write makes dead (K-12) | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `LinkEdit.needsGraph` | field | `LinkEdit.locked` | Different lift rule (D-2) |
| `effectiveLocks`, `savedGraphIDs`, `draftInNoGraph`, `isInNoGraph`, `cardInNoGraph` | computed reads | `checkedGraphIDs`, `draftNeedsGraph` | Small reads on the one section model; `draftNeedsGraph` is replaced |
| FakeStore `holdsUpdateArtifact`, `heldUpdateArtifactCount`, `releaseUpdateArtifact` | test hooks | `holdsLinkEvidenceGraphArtifact` (deleted) | Same shape, on the write that remains |
| `L10n.Sources.artifactGraphsNone`, `artifactNotInGraph`, `artifactLinkToGraph`, `artifactLinkToGraphLabel(ref:)` | copy | `artifactGraphsNeedsOne` (deleted) | New meaning |

Searched: `rg "visibleGraphs|draftNeedsGraph|graphListReloaded|holdsLink|artifactGraph" macos`.

**Builds on.** S12-13's group and card edit; S12-22's `updateArtifact(graphIDs:)`, pair preflight and `EvidenceGraphUnlinkReason`; `PVDeleteImpactCopy.heading`; the `e7063636` RPC removal.
**Deletes.** `LinkEvidenceGraphArtifact` / `UnlinkEvidenceGraphArtifact` everywhere (D-9); `graphListReloaded()`; the save loop's `attemptedLink`, `wroteAnything`, `refusedGraphID`, `linkEffects`; `draftNeedsGraph`; FakeStore's link hold and errors; `L10n.Sources.artifactGraphOnlyArtifact`, `artifactGraphsNeedsOne`, `artifactCardSaveErrorMessage` and their catalog entries; `SourcePageModelTests` `linkCalls`, `cardSaveReloadsTheGraphListOnce`; the S12-22 comment in `loadLinkLocks`.
**Next adds.** S12-30 opts the card's and graph form's callouts into announcing. The idea file's delete confirm can reuse `isInNoGraph`'s read.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An Artifact in no graph, made from the app | Composer, canvas (`CitationComposerModel`, `EvidenceGraphModel`) | Offer only linked Artifacts (S12-10); the no-artifact callouts already name the graph (S12-D4 LK-11, LK-12) |
| Same | Graph form edit | Lists every Artifact unchecked for this graph; unaffected |
| **Link to a graph** → card edit | `SourceArtifactsSection.beginEdit` | Same path as the pencil; closes another card's edit first |
| A one-graph Source's card in edit | `loadLinkLocks` | One pair; `.lastArtifact` when it is Default's only Artifact |
| A graph created or deleted during card edit | `savedGraphIDs`, `liveGraphIDs` | Created: kept in the write (D-4). Deleted: dropped (today) |

### Scale

At the map's Scale (a census with 60 graphs and 300 Citations on one page): the card's edit is one preflight over at most 60 pairs (as since S12-22) and one write; `effectiveLocks` walks at most 60 ids per read. `isInNoGraph` walks the session list once per resting card: 60 graphs × their `artifactIDs`; the card grid is lazy. The update's effects are the engine's (S12-25 narrows them).

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
| Add dialog group | `PVField` + `SourceArtifactGraphLinks` | `SourcePageArtifactsView.addForm` | Shows when `visibleGraphs != nil`; the ND-3 line when `draftInNoGraph` (muted italic caption, as the deleted needs-one line) |
| Card edit group | same | `SourcePageArtifactsView.graphLinks` | `locked:` is `effectiveLocks`; the ND-6 line under the group when `cardInNoGraph(id)` (`linkEdit` ready and `checkedGraphIDs` empty) |
| Resting row | `PVBadge(…, tone: .neutral, icon: .link)` + `PVButton(…, variant: .ghost, size: .sm)` | `SourcePageArtifactsView.restingText` | Under `descriptionLine`, only when `isInNoGraph(art)`; wraps with `PVFlowLayout` (`macos/App/DesignSystem/Components/FlowLayout/PVFlowLayout.swift`) at narrow widths; button `.accessibilityLabel(L10n.Sources.artifactLinkToGraphLabel(ref:))`; ids `sources.page.artifact.<id>.noGraph`, `.linkToGraph` |
| Refused save | `PVCallout` danger compact | `editFields` (unchanged position) | Body from D-5 |
| Graph form kept group | `keptLinked` | `SourceGraphFormDialog.swift` | ND-12 copy |

Frames in words (until the board is agreed): **Add, one graph** — the dialog with one checked "Default" row and its ref. **Add, nothing checked** — every row unchecked, Add enabled, the ND-3 caption under the group. **Resting cards** — a card in no graph shows, under its description, a neutral badge "Not in an evidence graph" and a ghost "Link to a graph" button; other cards unchanged. **Cited page, moving** — Jones locked with "1 citation would be left in no evidence graph"; after Smith is checked, Jones shows its ref and no lock mark. **Save refused** — the danger callout above Label, naming the graph; that row checked and locked.

States: loading list → no group, no badge; read error → S12-D4's compact callout with Read again; a lot of graphs → the group scrolls past six rows (S12-D4). Accessibility per ND-10, ND-11.

### Strings

Final copy is the agreed board's; these are the brief's and are replaced by the board's at agreement.

- *(new)* `L10n.Sources.artifactGraphsNone`, key `sources.page.artifact.graphs.none` = "This artifact won’t be in an evidence graph. Citations can’t use it until it is linked." (comment: "Muted line under the evidence graph group in the Add artifact dialog and on an artifact card in edit when no box is checked").
- *(new)* `L10n.Sources.artifactNotInGraph`, key `sources.page.artifact.notInGraph` = "Not in an evidence graph" (badge on a resting card).
- *(new)* `L10n.Sources.artifactLinkToGraph`, key `sources.page.artifact.linkToGraph` = "Link to a graph" (button).
- *(new)* `L10n.Sources.artifactLinkToGraphLabel(ref:)`, key `sources.page.artifact.linkToGraphLabel` = "Link %@ to an evidence graph" (VoiceOver label; argument is the Artifact ref).
- Changed `sources.page.artifact.saveErrorGraphKeeps`: "%1$@ keeps this artifact. %2$@" → "%1$@ keeps this artifact. Your changes are kept. %2$@" (comment: a graph refused to drop the artifact; nothing was saved).
- Changed `sources.graphForm.keptLinkedMessage`: → "These artifacts stay linked: citations need them in this evidence graph."
- Removed: `sources.page.artifact.graphs.needsOne`, `sources.page.artifact.graphs.onlyArtifact`, `sources.page.artifact.saveErrorMessage`.
- Old-scope copy checked and kept: `sources.page.artifact.graphs.hint` ("Which evidence graphs can cite this artifact") reads true with none checked; `sources.page.artifact.saveErrorKept` unchanged.

## 7. Steps

**STEP-1 — The Add dialog on every Source, none allowed**
Files: `macos/App/Features/Sources/SourceArtifactsSection.swift`, `macos/App/Features/Sources/SourcePageArtifactsView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-1; the ND-3 line with `artifactGraphsNone`; delete `artifactGraphsNeedsOne`. Write T-1, T-2; rewrite `addArtifactRetryIgnoresGraphChoice` to assert `store.createArtifactGraphIDs.count == 1` after the retry.
Check: `xcodebuild test` (as `scripts/check-pr.sh` runs it) passes; `rg "draftNeedsGraph|artifactGraphsNeedsOne" macos` has no hit.

**STEP-2 — The card's group and locks on every Source**
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skill: `add-swift-test`.
Do: D-2, D-3 (`loadLinkLocks`, `effectiveLocks`, `checkedGraphIDs`); the view passes `effectiveLocks`; the ND-6 line. Write T-3, T-4, T-5.
Check: `xcodebuild test` passes.

**STEP-3 — The card's Save is one write**
Files: `SourceArtifactsSection.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`.
Do: D-4, D-5, D-10; reword `artifactCardSaveErrorGraphKeeps`; delete `artifactCardSaveErrorMessage`, `graphListReloaded`. Write T-6 – T-9; rewrite `cardSaveErrorStaysOnTheCardThatSaved` with `holdsUpdateArtifact`, `cardSaveSkipsAGraphDeletedDuringEdit` to expect `store.updateArtifactGraphIDs == [nil]`, `cardSaveReloadFailureOffersReadAgain` as T-9b; delete `cardSaveLinksThenUnlinks` (T-6 replaces it), `cardSaveReloadsTheGraphListOnce`, `cardSaveRefusedKeepsEditAndRereadsLocks` (T-8 replaces it).
Check: `xcodebuild test` passes; `rg "linkEvidenceGraphArtifact|unlinkEvidenceGraphArtifact" macos/App/Features` has no hit.

**STEP-4 — The resting card marks an Artifact in no graph**
Files: `SourceArtifactsSection.swift`, `SourcePageArtifactsView.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcePageModelTests.swift`. Skills: `add-ui-component`, `add-localized-string`.
Do: D-6; the row per §6 macOS and the brief inventory. Write T-10.
Check: `xcodebuild test` passes; the manual check's step 2 shows the row.

**STEP-5 — The graph form**
Files: `macos/App/Features/Sources/SourceGraphFormModel.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/SourceGraphFormModelTests.swift`.
Do: D-7, D-8; the `keptLinkedMessage` copy. Write T-11, T-12, T-13; `createPrefillsEveryArtifactChecked` stays green unchanged.
Check: `xcodebuild test` passes.

**STEP-6 — Delete the link and unlink RPCs**
Files: `api/proto/engine.proto`, generated code, `api/ffi/handlers/evidence_graphs.go`, `api/ffi/handlers/evidence_graphs_test.go`, `api/ffi/dispatch.go`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`. Skill: `add-ffi-handler`.
Do: D-9; `./scripts/generate-proto.sh`. In `writesReportTheGraphList`, drop `linked` and `unlinked` from the loop.
Check: AC-5.1's `rg` has no hit; `./scripts/generate-proto.sh` leaves no diff on a second run; `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `xcodebuild test` passes.

**STEP-7 — Docs and guidance**
Files: `.agents/skills/add-workspace-place/SKILL.md` (line 172: Artifact ↔ graph link writes are `updateArtifact` with links and `updateEvidenceGraph`), `docs/source-layer-data-model.md` (any sentence that says Add artifact needs a graph or that one-graph Sources hide the boxes), `docs/deployment-plan/spike-12/deployment-plan.md` (map row status; working rule "Single-graph Sources hide link checkboxes" stays marked superseded), `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/design/README.md` (S12-D7 to Archived), the brief to `design/archive/`, this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-guidance-names.py` passes.

Where checks run: `generate-proto.sh` and `xcodebuild` need the Mac.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Swift Testing (`add-swift-test`). Fixtures: `makeCensus` (Brown `g-b` P9; Jones `g-j` P3 P4; Smith `g-s` P3 P6) and `unlinkBlock(_:graphID:citations:reason:)`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcePageModelTests` `addArtifactOnOneGraphShowsTheGroup` *(replaces `addArtifactOnOneGraphSendsEveryGraph` and `addArtifactSendsDefaultGraphLinksWhenBoxesHidden`)* | One graph `g-1`: `visibleGraphs` ids `["g-1"]`; create → `createArtifactGraphIDs == [["g-1"]]`. A store with no graph list value: `visibleGraphs == nil`, `draftGraphIDs == nil`, create → `[nil]` | AC-1.1, AC-1.2, AC-1.4, D-1 |
| T-2 | Swift | same `addArtifactWithNoGraphChecked` *(replaces `addArtifactNeedsAGraph`)* | Census, every box excluded: `canSubmitDraft`, `draftInNoGraph`; create → `createArtifactGraphIDs == [[]]`, dialog closed | AC-1.3, D-1 |
| T-3 | Swift | same `cardEditOnOneGraphShowsTheGroup` *(replaces `cardEditOnOneGraphHasNoLinks`)* | One graph `g-1` linking P3: `linkEdit` non-nil, one preflight call with pair (g-1, P3); exclude `g-1`, Save → `updateArtifactGraphIDs == [[]]` | AC-3.3, D-2 |
| T-4 | Swift | same `cardEditLocksByReason` *(rewrites `cardEditLocksInUseAndOnlyArtifactGraphs`)* | (g-s,P3) in use, 2 → `effectiveLocks["g-s"]` = "2 citations in this graph use this artifact"; (g-b,P9) `.lastArtifact` → `L10n.Errors.message(code: "evidencegraphs.last_artifact")`; a P9 that no graph links → no call, no locks | AC-3.4, AC-3.5, D-3 |
| T-5 | Swift | same `cardCitationsLockLiftsWhenAnotherGraphIsChecked` *(new)* | Graphs Brown (P9, P4), Jones (P3, P4), Smith (P6); P9 with (g-b,P9) `.citationsNeedGraph`, 1: `effectiveLocks["g-b"]` = "1 citation would be left in no evidence graph", `checkedGraphIDs == ["g-b"]`; un-exclude `g-j` → no lock on g-b; exclude `g-b` → `checkedGraphIDs == ["g-j"]`; Save → `updateArtifactGraphIDs == [["g-j"]]`. Second run: exclude `g-j` again → g-b locked and checked | AC-3.6, AC-3.7, D-3 |
| T-6 | Swift | same `cardSaveIsOneWrite` *(replaces `cardSaveLinksThenUnlinks`)* | P3: exclude g-j, un-exclude g-b, label "Page 3, front" → `updateArtifactCalls == 1`, `updateArtifactGraphIDs == [["g-b","g-s"]]`, label on the page, edit closed; no recorded call contains "EvidenceGraphArtifact" | AC-3.1, D-4 |
| T-7 | Swift | same `cardSaveSendsNoLinksWhenOnlyTextChanged` *(rewrites `cardSaveFieldsOnlyHasNoLinkCalls`)* | Label only → `updateArtifactGraphIDs == [nil]` | AC-3.2, D-4 |
| T-8 | Swift | same `cardSaveRefusedNamesTheGraphAndLocksIt` *(replaces `cardSaveRefusedKeepsEditAndRereadsLocks`)* | Exclude g-j; `updateArtifactError` `artifact_in_use`; blocks scripted (g-j,P3) in use, 1 → card in edit, `cardSaveError == artifactCardSaveErrorGraphKeeps(graph: "Jones household", error:)`, `checkedGraphIDs` contains g-j, `excluded` still has g-j. Second case: graphs as T-5 but no block at open; exclude g-b; refusal `citations_need_graph`; blocks then (g-b,P9) `.citationsNeedGraph` → g-b in `effectiveLocks` and checked; the error names Brown household. Third: refusal with no block → `artifactCardSaveErrorKept` | AC-3.8, AC-3.9, AC-3.10, D-5 |
| T-9 | Swift | same `cardSaveKeepsAGraphMadeDuringEdit` *(new)* | Edit P3; `createEvidenceGraph` linking P3, effects applied, list awaited; exclude g-j; Save → `updateArtifactGraphIDs == [["g-s", newID]]` | AC-3.11, D-4 |
| T-9b | Swift | same `cardSaveReloadFailureOffersReadAgain` (rewritten) | Exclude g-j; refusal; `getEvidenceGraphUnlinkImpactError` set → `linkEdit.loadState` is `.failed` | D-5 |
| T-10 | Swift | same `restingCardMarksAnArtifactInNoGraph` *(new)* | Graphs omit P6: `isInNoGraph(P6)`, not P3; a store with no list value → false for P6; `beginEdit("P6")` → `excluded` is every graph id | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-11 | Swift | `SourceGraphFormModelTests` `editLocksCitationsNeedGraphArtifacts` *(new)* | `openSmithEdit` with (g-smith,a3) `.citationsNeedGraph` 1 and (g-smith,a4) `.lastArtifact` → `locked` = [a3: "1 citation would be left in no evidence graph"], `keptLinked` is a3 only | AC-4.1, D-7 |
| T-12 | Swift | same `createPrefillsTheArtifactsInNoGraph` *(new)* | Graph linking a1 – a4 → `checkedArtifactIDs == ["a5","a6"]`, `!isDirty`; graph linking a1 – a6 → every Artifact checked | AC-4.3, AC-4.4, D-8 |
| T-13 | Swift | same `editRefusedCitationsNeedGraphRereadsKeptLinks` *(new)* | As `editRefusedInUseRereadsKeptLinks` with code `evidencegraphs.citations_need_graph` and a `.citationsNeedGraph` block → a2 locked | AC-4.2, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcePageModelTests` (the S12-13 section), `SourceGraphFormModelTests`, `EvidenceGraphStoreTests`, `TestSwiftVocabularyMatchesGo`, `L10nResolutionTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each deleted name `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left — `LinkEvidenceGraphArtifact`, `UnlinkEvidenceGraphArtifact`, `linkEvidenceGraphArtifact`, `unlinkEvidenceGraphArtifact`, `graphListReloaded`, `artifactGraphOnlyArtifact`, `artifactGraphsNeedsOne`, `artifactCardSaveErrorMessage`, `draftNeedsGraph`; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive (the brief marks none New)
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain doc updated: source-layer-data-model.md where it describes Add artifact's graph choice
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice Check: on a two-graph Source, a card that unlinks one graph and links another saves in one revision; a refused save names the graph and leaves nothing half applied; a `citations_need_graph` refusal ends with the box checked and locked, and no "save again to finish" text; no app code calls the deleted RPCs.

1. Open a project with a one-graph Source. Add artifact "Page 7": the group shows Default, checked. Uncheck it: the ND-3 line shows, Add stays enabled. Add.
2. See Page 7's card show "Not in an evidence graph" and **Link to a graph**. Click it: edit opens with Default unchecked. Check Default and Save. The badge goes.
3. On a census with Brown, Jones and Smith: open Page 3 (in Jones and Smith); uncheck Jones, check Brown, change the label; Save. Open the History panel: one revision holds the label, the Brown link and the Jones unlink.
4. Give Page 5 (only in Jones) a Citation in Jones's composer, with no Observation. Open its card: Jones is locked with "1 citation would be left in no evidence graph". Check Smith: Jones unlocks. Uncheck Jones; Save. Page 5 is in Smith only.
5. Open Page 5 in a second window's card edit, then add a Citation in Smith's composer from the first; in the card uncheck Smith and Save. The callout names Smith household and says nothing about saving again; Smith is checked and locked.
6. Edit Smith in the graph form: Page 5 is locked under Kept linked with its Citation.
7. Delete every page's links from one graph so Page 8 is in no graph; Add evidence graph: only Page 8 starts checked.
8. AC-5.1, AC-5.2: run `rg "LinkEvidenceGraphArtifact|linkEvidenceGraphArtifact" api macos --glob '!*.pb.*'` (no hit) and read the method enum's reserved lines in `api/proto/engine.proto`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `applyWrite` after an update with links does not reload the graph list, so the badge stays | The engine reports `evidence_graphs` for link rows (S12-22 §6 Scale); STEP-4's manual step 2 checks it |
| A cited Artifact in two graphs with Citations needing one: unchecking both is refused only at Save | D-5's callout and re-read; no Swift copy of the rule |

**Stop and report if:**

- `loadLinkLocks`, `saveFields` or `SourceGraphFormModel.loadKeptLinks` no longer match §4 at your base.
- `updateArtifact` with links returns effects without the `evidence_graphs` list, so the session list does not reload after a card Save.
- The agreed board's frames contradict a Decision here (for example a lock that never lifts, or Add disabled with no box checked).
- An existing Go or Swift test outside §8's list fails for a reason other than the deleted RPCs or the changed copy.
- Deleting a proto method breaks `TestSwiftVocabularyMatchesGo` or another pinning test in a way the reserved lines do not fix.

## 12. Open questions

1. **Board S12-D7.** The brief's copy and the resting row's layout are recommendations until the board is agreed. Recommended: agree the brief as written; this plan's Strings follow the board.

## 13. PR

**Title:** S12-38 — Add an Artifact without a graph; the card saves links in one write
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-38-artifact-without-graph.md`.
