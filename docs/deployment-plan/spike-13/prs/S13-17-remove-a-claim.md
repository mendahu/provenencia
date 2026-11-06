# S13-17 — Remove a claim

**Status:** Blocked — on S13-37, S13-16
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/S13-D6-sources-section.md) (revision 2), board agreed 2026-11-05
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Sources.dc.html*, turn "S13-D6 rev 2 · Sources as Source rows, Remove and the last-record warning", frames 14n – 14u, 14y (this PR). **Design → macOS** describes them in words
**Depends on:** S13-08 (landed [#707](https://github.com/mendahu/provenencia/pull/707)), S13-37 (the engine reports what Unpromote removes), S13-16, S13-D6 (board agreed 2026-11-05)
**Blocks:** —
**Branch:** `s13-17-remove-a-claim`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § Identity claim controls](../person-detail.md#identity-claim-controls) (R-38, R-39); working rules "One Unpromote, deletes through the registry", "Sections are places"; brief [S13-D6](../design/S13-D6-sources-section.md) RC-1 – RC-7; the S13-08 plan (D-14: `getUnpromoteImpact`; D-16: the per-kind deletes copy) and the S13-16 plan
**Planned:** 2026-11-05 against `main` @ `124d492` (board agreed; refreshed when S13-37 and S13-16 land, then `Planned`)
**Review:** [`S13-17-remove-a-claim.review.md`](S13-17-remove-a-claim.review.md) — round 1 Reshape / Not ready, reshaped with the user's answers; round 2 (board agreed) Ready with changes, every finding Fixed; refreshed once S13-37 and S13-16 land

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

A researcher can take a record out of a person from the Person page: Unpromote on a record's line under its Source in the Sources section unpromotes that Subject, exactly as Unpromote on the Evidence graph does (R-38), through the same confirm and the same flow (user, 2026-11-05). Removing the last claim deletes the person; the confirm says so first, and on proceeding the app lands on the Persons list (R-39).

**Why now.** S13-08 made the engine delete a handle whose last member leaves and gave Unpromote its preflight; S13-37 makes the engine report the connections Unpromote takes back and the handles a write deletes, so neither page re-derives them; S13-16 draws the Sources rows and a line per record, which this PR ends with the control. This PR closes slice 6 (bar 7, removal half).

**What stays.** The engine (S13-37 is its change). The Evidence graph's Unpromote confirm keeps its title, buttons, list, tone and placement; its open, write and error handling move into one shared flow both pages hold, as both hold a `DeleteImpactFlow` for deletes. Its single-handle deletes message becomes the per-kind sentence the user approved for S13-08 (2026-11-04, S13-08 D-16), which the landed code defines but does not show (D-5).

## 2. User stories

**US-1.** As a researcher, when a Subject was promoted into the wrong person, I unpromote it from the Person page, so that I do not have to find its Evidence graph first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber (PER-3HK9W) with seven claims; Unpromote on the Household 388 line (Subject "J. Weber", CPR-2M8XF; the preflight names one relationship and one participation) | The confirm opens: title "Unpromote J. Weber?", message "This promotion is removed. The cards stay on the graph, and the handle stays.", "Also unpromoted" with the two connection sentences worded from the claim's graph and "The other end of each stays promoted.", buttons "Unpromote" (irreversible) and "Keep promotion" (RC-1, RC-2) |
| AC-1.2 | Confirm | `unpromoteSubject` runs once for CPR-2M8XF; the confirm closes; the line leaves after the reload (an emptied Source row leaves too); the page stays on Sources, scroll unchanged; focus is on the next line's Unpromote (Volume 3) (RC-6) |
| AC-1.3 | Keep promotion | No write; the confirm closes |
| AC-1.4 | The write fails (`promote.stale`) | The confirm stays open with the error under the list; "Unpromote" is enabled to retry (RC-5) |
| AC-1.5 | While an Unpromote opens or writes | Every Unpromote control in the section is disabled (RC-1) |
| AC-1.6 | The claim's graph rows cannot be read | The confirm opens with the preflight's message and, in place of the list, "The connections that come off with it could not be listed. The other end of each stays promoted." (D-4) |

**US-2.** As a researcher, when I remove a person's last record, I am told the person is deleted first, and land on the Persons list after.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Maria Weber (PER-8JD2L) with one claim; Unpromote | The preflight's deletes group lists PER-8JD2L only; danger tone; "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." (RC-3; D-5) |
| AC-2.2 | The preflight lists PER-8JD2L and two association handles | The message is the DeleteImpact line naming all three, danger tone (RC-3; D-5) |
| AC-2.3 | The preflight fails | Danger tone, "The check for whether this person is deleted could not run. Confirming may still delete this person. The cards stay on the graph.", the error under it; Unpromote enabled (RC-4) |
| AC-2.4 | Confirm on AC-2.1 | The write's effects list PER-8JD2L's id in `deletedHandles`; the app shows the Persons list (`navigation.fallbackToSectionRoot()`) without Maria; Back does not return to her page (RC-7; D-6, D-7) |
| AC-2.5 | Confirm on AC-2.3 | Lands on the Persons list exactly when the write's `deletedHandles` holds the person; otherwise stays on Sources (D-6) |

**US-3.** As a researcher on the Evidence graph, Unpromote behaves as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Unpromote on an Evidence graph card whose handle keeps a member | The same confirm and list as today (the list now from the preflight's bridge ids); the graph stays |
| AC-3.2 | Unpromote of a person's last member on the graph | The per-kind sentence of AC-2.1; danger tone; the graph stays (working rule "One Unpromote") |

## 3. Scope

| In | Covered by |
| --- | --- |
| One Unpromote flow and confirm for both pages, lifted from the Evidence graph: `UnpromoteFlow`, `UnpromoteConfirmItem`, its message, the `unpromoteConfirm(flow:…)` modifier (RC-2 – RC-5) | STEP-1, STEP-2, AC-3.1, AC-3.2 |
| The list from the preflight's bridge ids; the Swift copy of the engine's bridge rule goes (S13-37) | STEP-2, AC-1.1, AC-3.1 |
| The single-handle deletes sentence approved for S13-08 (RC-3) | STEP-2, AC-2.1, AC-3.2 |
| One session read of a graph's snapshot, shared with Promote | STEP-3, AC-1.1, AC-1.6 |
| Unpromote at the end of each record's line, disabled while busy; focus after a kept Unpromote; landing on the Persons list (RC-1, RC-6, RC-7) | STEP-5, AC-1.2, AC-1.5, AC-2.4 |
| `userID` reaches page sections | STEP-4 |
| Strings | each step that adds a key; STEP-6 checks the catalog |
| Docs, brief archive | STEP-7 |

| Out | Home |
| --- | --- |
| Engine fields for the bridges and the deleted handles | S13-37 |
| A Toast after Unpromote | not planned (brief RC-6), unless the board adds one |
| Leaving a Conclusion page whose entity another write deleted (Event, Place, another window) | their redesign (R-1); S13-37's `deletedHandles` gives it the fact |
| Bulk Unpromote, Reject, Provisional | not planned (brief § 2.1) |
| How the records connect | S13-28 |

## 4. Current state

Rows marked *(S13-37)* or *(S13-16)* describe what those PRs add; the implementer checks them against their Landed notes.

| Area | Where | Today |
| --- | --- | --- |
| Unpromote state | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `PendingUnpromote{id, name, kind, bridges, deletedHandles, impactCheckFailed, error}` and `confirmTone`; `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen()` | The graph model owns the item, the busy flag and the stale-open guard |
| Unpromote begin | `EvidenceGraphModel.beginUnpromote(subjectID:)` | From the graph's snapshot: bridges with the Subject as an end and a membership, sorted by id, worded by `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)` (a Swift copy of `identityclaims.UnfileBridges`); then `store.getUnpromoteImpact` (`total > 0` → `deletedHandles`; a throw → `impactCheckFailed`, `error`) |
| Unpromote write | `EvidenceGraphModel.confirmUnpromote()` | `store.unpromoteSubject`, `session.apply(effects)`, closes; a failure stays as `error` |
| Unpromote confirm | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `.pvConfirm(item: $model.pendingUnpromote, copy: …)` and `unpromoteDetail(_:)` | The message is built in the `copy:` closure: `PVDeleteImpactCopy.cascadeLines` first line, else the preflight-failed copy, else `unpromoteConfirmMessage` |
| The approved deletes copy | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` | No caller in `macos/App`; only `macos/ProvenenciaTests/PVDeleteImpactTests.swift` reads it |
| Flow precedent | `macos/App/DesignSystem/Recipes/DeleteImpact/DeleteImpactFlow.swift` `DeleteImpactFlow` (`request`, `isRunning`, `error`, `confirm(perform:) -> Bool`); held by `EvidenceGraphModel.deleteImpact` and `SourcePageModel.deleteImpact`, bound with `.pvDeleteImpact(flow:…)` | The shape `UnpromoteFlow` copies |
| Snapshot read | `macos/App/Features/Promote/PromoteModel.swift` `graphSnapshot()` | `readyValue` on the graph, properties and rules keys, then `EvidenceGraphSnapshot.build(rows:types:rules:)`; nil when the rows are nil |
| Preflight | `GetUnpromoteImpactResponse{deletes}`; `GenealogyStore.getUnpromoteImpact(projectDir:subjectID:) -> CatalogDeleteImpactGroup?` | The listed rows carry each deleted handle's id; `deleteimpact.ListedGroup` caps `listed` (`listedCap`), the subject's own handle first. *(S13-37)* adds the bridge Subject ids the write will unpromote |
| Effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects` | No deleted-handles field. *(S13-37)* adds `deletedHandles: [String]` |
| Not-found error | `core/database/conclusiondetails/conclusiondetails.go` `ErrNotFound` (`apperr.KindUser`) | Not used by this plan (D-6 reads the effects) |
| Leaving a page | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `fallbackToSectionRoot()`; `macos/App/Features/Sources/SourcePageView.swift` (`if await model.confirmResourceDelete() { navigation.fallbackToSectionRoot() }`) | Replaces the current entry with the list |
| Page sections | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPageSectionsContext{projectDir, refreshing, pageIdentifier, lead, pageSection}`; `ConclusionLeadModel(session:userID:entityId:)` built by the page from its `userID`, writing through `session.store` | Sections get no `userID` |
| Sources section *(S13-16)* | `PersonSourcesSection.swift` in `Features/Conclusions/` | `PersonSourcesList` draws each row's `PersonSourcesContent.Line{id (claimID), subjectID, graphTitle, graphRef, subjectName, subjectRef, location}` as one caption line ending with the Subject ref; folds past six lines (S13-16 D-3, D-7, D-10) |
| Tests | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges`, `unpromoteWithNoBridgesHasEmptyDetail`, `confirmUnpromoteRefreshesToPromote`, `unpromoteFailureStaysInConfirm`, `unpromoteLastRecordNamesTheDeletedHandle` (asserts `deletedHandles?.listed.first?.ref`), `unpromotePreflightFailureStaysInTheConfirm` | Cover the graph's flow |

**Precedent.** `DeleteImpactFlow`: one observable flow object a model holds and a view modifier binds, shared by the Evidence graph and the Source page; copy its shape for `UnpromoteFlow`. The Source page's resource delete: a successful write that removes the page's entity, then `fallbackToSectionRoot()`; copy that.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One flow: `@MainActor @Observable final class UnpromoteFlow` *(new)* in `macos/App/Features/EvidenceGraph/UnpromoteFlow.swift` *(new)* owns `pending: UnpromoteConfirmItem?`, `isRunning`, `isOpening`, the open generation, `open(subjectID:name:kind:handleID:snapshot: EvidenceGraphSnapshot?, orders:session:) async` (it calls the preflight through `session.store`), `confirm(session:userID:) async -> UnpromoteResult?` (write, `session.apply(effects)`, close; a failure stays as `pending.error` and returns nil) and `cancel()`. `UnpromoteResult` *(new)* holds the `CatalogWriteEffects`. `EvidenceGraphModel` holds `let unpromote = UnpromoteFlow()` (as it holds `deleteImpact`); its `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen` and `confirmUnpromote` go; `beginUnpromote(subjectID:)` keeps its guards and calls `unpromote.open` | One place for the guard, the generation and the error rule; the `DeleteImpactFlow` precedent (round 1 M-1) | A second state machine in a Person-page model |
| D-2 | `UnpromoteConfirmItem` *(new, same file)* = today's `PendingUnpromote` fields plus `handleID: String?` (the Subject's own handle) and `bridgesUnlisted: Bool`; computed `deletesOwnHandle` (`deletedHandles` non-nil, `total == 1`, its one listed row's id == `handleID`), `confirmTone` (today's rule), and `message: String` (D-5). The modifier `View.unpromoteConfirm(flow: UnpromoteFlow, accessibilityIdentifierPrefix: String, onConfirm: @escaping () -> Void)` *(new, same file)* holds today's `pvConfirm` call and `unpromoteDetail` view, reading `item.message`. `EvidenceGraphView` calls it with `"evidenceGraph.unpromote"` | The copy is testable on the item (round 1 X-3); one rule for "deletes its own handle" (round 1 M-5) | Message logic in a view closure |
| D-3 | The list comes from the engine: `open` takes the preflight's bridge Subject ids *(S13-37)* and words each with `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)` from the snapshot, in the preflight's order; an id not in the snapshot is skipped. The Swift filter in `beginUnpromote` goes | The engine's rule in one layer (round 1 M-3; user, 2026-11-05: preparatory PR S13-37) | Keeping the Swift filter |
| D-4 | One snapshot read: `WorkspaceSession.evidenceGraphSnapshot(graphID:) async -> EvidenceGraphSnapshot?` *(new)* in `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`: `session.query` then `readyValue` on `.evidenceGraph(project: projectKey, graphId:)`, `.propertiesWorkspace(project:)` and `.connectRules(project:)`; nil when the rows are nil; missing types or rules → `[]` (Promote's rule). `PromoteModel.graphSnapshot()` moves onto it. When it is nil, or the Subject is not in it, `open` sets `bridgesUnlisted` and the title is the name the caller passes (the page passes `Line.subjectName`) | One place decides a snapshot's keys (round 1 M-4); a missing list never hides the confirm | A third copy of the three reads |
| D-5 | `UnpromoteConfirmItem.message`: `deletesOwnHandle` → `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` with the listed row's ref; another non-empty deletes group → `PVDeleteImpactCopy.cascadeLines(...)`'s first line (today's); `impactCheckFailed` → the preflight-failed copy; else `unpromoteConfirmMessage`. `bridgesUnlisted` shows `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted` *(new)* in the detail in place of the list and its note | The copy the user approved on 2026-11-04 (S13-08 D-16), defined and never called; the many-handle case keeps today's wording, as the brief says | New Person-page copy; leaving the key dead |
| D-6 | The person is gone when the write's `effects.deletedHandles` *(S13-37)* contains the page's entity id. No preflight or detail read decides it | One fact, one source: what the write did (round 1 M-2; user, 2026-11-05) | Inferring from the preflight; a detail read keyed to an error code |
| D-7 | On "gone" the section calls `navigation.fallbackToSectionRoot()`, which replaces the Person page's entry with the Persons list. It calls it right after `confirm` returns, before the view reads the refreshed detail, so the not-found Callout is not shown: `session.apply` marks `.conclusionDetail` stale and its reload is async, and the navigation unmounts the page first | The Source page precedent; Back does not return to a deleted page (RC-7; round 1 X-6) | `go(to: .sectionRoot(.persons))` |
| D-8 | `ConclusionDetailPageSectionsContext` gains `userID: String` *(new field)*, set by `ConclusionDetailPage` from its `userID`. `PersonDetailSections` passes `context.userID` to `PersonSourcesSection(entityId:userID:)`. The section holds `@State private var unpromote = UnpromoteFlow()` and resets it (`cancel()`) in `.onChange(of: entityId)`; it reads `@Environment(WorkspaceNavigation.self)` | Every section can write without threading props (round 1 M-7); writes use `session.store` | Adding `userID` to one section's initializer from the view |
| D-9 | At the end of each record's line, 6pt after its Subject ref: `PVButton(L10n.EvidenceGraph.unpromote, variant: .ghost, size: .sm, icon: .unpromote)` (the glyph and the word, board 14a: an icon alone at the far right read as detached, user 2026-11-05), its VoiceOver label `L10n.Conclusions.sourcesUnpromoteAccessibility(name:)` ("Unpromote J. Weber"), id `persons.detail.sources.unpromote.<claimID>`, part of the line (the lines' Unpromote buttons do not align in a column), disabled while `unpromote.isOpening || unpromote.isRunning`. Its action, from S13-16's `Line`: `Task { let snapshot = await session.evidenceGraphSnapshot(graphID: line.graphID); await unpromote.open(subjectID: line.subjectID, name: line.subjectName, kind: .person, handleID: entityId, snapshot: snapshot, orders: session.nameFormat, session: session) }`. `line.subjectName` is the graph card's name rule (S13-16 D-2), so the line, the button's label and the confirm's title agree with or without a snapshot. The section's modifier: `unpromoteConfirm(flow: unpromote, accessibilityIdentifierPrefix: "persons.detail.sources.unpromote", onConfirm: …)`; `onConfirm` awaits `confirm`, then D-6 / D-7, else D-10 | The action is called Unpromote everywhere in code (round 1 M-6) | "Remove" names in code |
| D-10 | `PersonSourcesList` gains `unpromote: (PersonSourcesContent.Line) -> Void`, `unpromoteDisabled: Bool` and `focus: FocusState<PersonSourcesFocus?>.Binding` *(new parameters)*; `PersonSourcesFocus` *(new enum, in the section file)* is `.unpromote(claimID: String)` or `.sourceTitle(sourceID: String)`; the S13-16 title link and each Unpromote button take `.focused(focus, equals:)`. After a kept Unpromote, focus goes to the next line's Unpromote in reading order (Source rows in order, their visible lines in order, flattened), else the previous one's, else the first Source title. The section keeps the content as it was before the write, and applies the target when the reloaded content no longer holds the removed claim. No Toast (board 14t) | Brief RC-6, board 14t; focus never lands on a removed control (round 1 X-8) | Moving focus at confirm time (the line is still there) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `UnpromoteFlow`, `UnpromoteResult` *(new)* | type | `DeleteImpactFlow`; `EvidenceGraphModel`'s Unpromote state | Replaces the model's state; same shape as the delete flow |
| `UnpromoteConfirmItem` *(new)* | type | `EvidenceGraphModel.PendingUnpromote` | It is that type, moved, with `handleID`, `bridgesUnlisted`, `message` |
| `View.unpromoteConfirm(flow:…)` *(new)* | modifier | `.pvDeleteImpact(flow:…)`; `EvidenceGraphView`'s inline `.pvConfirm` | Moved |
| `WorkspaceSession.evidenceGraphSnapshot(graphID:)` *(new)* | func | `PromoteModel.graphSnapshot()` | Its body, moved to the session |
| `ConclusionDetailPageSectionsContext.userID` *(new)* | field | `lead` (built from `userID`) | Sections that write need it |
| `unpromoteConfirmBridgesUnlisted`, `sourcesUnpromoteAccessibility(name:)` *(new keys)* | keys | `unpromoteConfirmPreflightFailed`, `L10n.EvidenceGraph.unpromote` | New sentences |
| `UnpromoteFlowTests` *(new)* | test file | `EvidenceGraphModelTests`' Unpromote cases | The flow is tested once; the callers test their calls |

**Builds on.** S13-37's preflight bridge ids and `deletedHandles`; S13-08's `getUnpromoteImpact` and `unpromoteSubject`; `DeleteImpactFlow`'s shape; `EvidenceGraphSnapshot.build`; `fallbackToSectionRoot`; `PVConfirm`; S13-16's record line (`PersonSourcesContent.Line`, `PersonSourcesList`).
**Deletes.** `EvidenceGraphModel.PendingUnpromote`, `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen`, `confirmUnpromote`; the Swift bridge filter in `beginUnpromote`; `EvidenceGraphView`'s inline Unpromote `.pvConfirm` and `unpromoteDetail`; `PromoteModel.graphSnapshot()` (moved). `unpromoteConfirmMessageDeletes` gains its first caller.
**Next adds.** An Event or Place page with sections holds an `UnpromoteFlow`; S13-28 reads graphs through `evidenceGraphSnapshot`; any page that leaves when its entity is deleted reads `deletedHandles`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Unpromote on a Person page → the shared confirm | `UnpromoteFlow.swift` *(new)* | The item is built fresh per open; a stale open is dropped by the generation (T-4) |
| A claim whose graph rows cannot be read, or whose Subject is gone from it | `UnpromoteFlow.open` | Opens with `bridgesUnlisted` (AC-1.6); the engine re-checks in the write |
| The person deleted by this page's write | `PersonDetailView` | Persons list before the detail reload paints (D-7) |
| Another person's page while a confirm is open | `PersonSourcesSection` | `cancel()` on `entityId` change (D-8) |

### Scale

One Unpromote: one `evidenceGraphSnapshot` (warm when the researcher came from that graph; else one `.evidenceGraph` load, the cost of opening it; the two lists are warm, S13-16 warms them), one preflight, one write. The effects refresh the per-person keys as Unpromote on the graph does today. Nothing grows with the person's claim count.

### Edges

- **Assumptions.** The preflight's bridge ids are exactly what the write unfiles (S13-37's contract, tested there). `listed` is capped but lists the Subject's own handle first; `deletesOwnHandle` needs only `total == 1`, so the cap never hides it. The gone decision reads the effects, not `listed` (D-6).
- **Neighbors.** Two claims: one removed keeps the person; the last deletes. A single deleted row that is an association handle (the person keeps a member): `deletesOwnHandle` is false (id ≠ `handleID`), so the `cascadeLines` line shows (T-2).
- **Removed inputs.** The claim already removed elsewhere: the preflight or the write fails with the engine's error in the confirm (AC-1.4); the reload drops the line. A bridge id not in the cached snapshot: skipped (T-3).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The bridge list is the Subject's filed bridges on one graph (today's bound).

### Schema

None — no catalog change (S13-37 has none either).

### Go core

None — S13-37.

### FFI

None — S13-37 adds the fields.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Unpromote control | `PVButton` (ghost, `.sm`, `PVSymbol.unpromote` and the word) at the end of the line | `PersonSourcesSection.swift` (S13-16's file) | D-9, RC-1 |
| Flow and confirm | `UnpromoteFlow`, `unpromoteConfirm` over `PVConfirm` | `macos/App/Features/EvidenceGraph/UnpromoteFlow.swift` *(new)* | D-1 – D-5, RC-2 – RC-5 |
| Deletes line | `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes`, `PVDeleteImpactCopy` | as today | D-5, RC-3 |
| Landing | `WorkspaceNavigation.fallbackToSectionRoot()` | — | D-7, RC-7 |

States: opening (controls disabled), writing (`isRunning`), failed (error under the list), preflight failed, list unavailable. Accessibility: the control reads "Unpromote J. Weber"; the confirm reads as today; focus per D-10. The board's frames are described here in words on agreement.

### Strings

L10n keys *(new)*: `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted` = "The connections that come off with it could not be listed. The other end of each stays promoted."; `L10n.Conclusions.sourcesUnpromoteAccessibility(name:)` = "Unpromote %@". Existing copy now also shown from the Person page, each still true there (the cards are on the claim's graph): `evidenceGraph.subject.unpromoteConfirmMessage` ("This promotion is removed. The cards stay on the graph, and the handle stays."), `unpromoteConfirmPreflightFailed.*`, `unpromoteConfirmMessageDeletes.*`. Removed: none.

## 7. Steps

**STEP-1 — One Unpromote flow**
Files: `macos/App/Features/EvidenceGraph/UnpromoteFlow.swift` *(new)*, `EvidenceGraphModel.swift`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift` *(new)*, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-1, and D-2's item without the message rule; the graph model moves onto the flow.
Check: T-1, T-4, T-5, T-6 pass.

**STEP-2 — One confirm, its message and its list**
Files: `UnpromoteFlow.swift`, `EvidenceGraphView.swift`, `EvidenceGraphModel.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift` *(new)* (created in STEP-1). Skill: `add-localized-string`.
Do: D-2's modifier and message, D-3, D-5 (with `unpromoteConfirmBridgesUnlisted`).
Check: T-2, T-3 pass; `rg 'endpointAID == subjectID' macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` finds nothing.

**STEP-3 — One snapshot read**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `UnpromoteFlow.swift`.
Do: D-4.
Check: T-7 passes; `PromoteModelTests` pass unchanged.

**STEP-4 — Sections get the user id**
Files: `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `PersonDetailView.swift`.
Do: D-8's context field.
Check: the build has no warning; `PersonDetailTests` pass.

**STEP-5 — Unpromote on the record lines**
Files: `PersonSourcesSection.swift` (S13-16's file in `Features/Conclusions/`), `PersonDetailView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-6 – D-10, with `sourcesUnpromoteAccessibility(name:)`; update `evidenceGraph.subject.unpromote`'s comment in `L10n.swift` and `Localizable.xcstrings` to say it is also the Person page's Unpromote button on a record line.
Check: T-8 passes; manual check items 1 – 4.

**STEP-6 — Strings check**
Files: `macos/App/Resources/Localizable.xcstrings`.
Do: confirm both new keys have entries and comments.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/design/S13-D6-sources-section.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` existing Unpromote cases (moved onto `model.unpromote`) | Unchanged behavior; `unpromoteLastRecordNamesTheDeletedHandle` also asserts `item.message` is the per-kind sentence | AC-3.1, AC-3.2, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `messageFollowsThePreflight` *(new)* | `handleID` "h-maria", group of one listed "h-maria" (PER-8JD2L) → "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph."; group of three → the `cascadeLines` first line; group of one listed "h-rel" → the `cascadeLines` line; `impactCheckFailed` → the preflight-failed copy; none → `unpromoteConfirmMessage`; tone `.danger` for the first four | AC-2.1, AC-2.2, AC-2.3, AC-3.2, D-2, D-5 |
| T-3 | Swift | same · `listComesFromThePreflight` *(new)* | Preflight bridge ids [b2, b1] with both in the snapshot → two sentences in that order; [b2, gone] → one; snapshot nil → `bridgesUnlisted` true and no sentences | AC-1.1, AC-1.6, D-3, D-4 |
| T-4 | Swift | same · `staleOpenIsDropped` *(new)* | `open` for A, then for B before A's preflight returns → `pending` is B's | D-1 |
| T-5 | Swift | same · `confirmReportsTheEffects` *(new)* | Write succeeds with `deletedHandles` ["h-maria"] → result's effects hold it, `pending` nil, `isRunning` false; `isRunning` true while the scripted write is suspended | AC-1.2, AC-1.5, AC-2.4, D-1, D-6 |
| T-6 | Swift | same · `confirmFailureStaysOpen` *(new)* | Write throws → result nil, `pending.error` set, retry allowed | AC-1.4 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `evidenceGraphSnapshotReadsTheThreeKeys` *(new)* | Seeded graph rows, properties and rules → a snapshot with the Subject; rows load fails → nil | AC-1.6, D-4 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `unpromoteOutcome` *(new)* | `PersonSourcesSection.outcome(effects:entityId:removed claimID:in content: PersonSourcesContent) -> Outcome` *(new, static)*: `deletedHandles` holds the entity → `.leave`; else `.stay(focus: PersonSourcesFocus?)`: the next visible line's `.unpromote` in reading order (across Source rows), else the previous one's, else `.sourceTitle` of the first row, else nil. Cases: a middle line → the next; the page's last line → the previous; the only line of the only Source (the person kept another member elsewhere and that row is gone) → nil | AC-2.4, AC-2.5, D-6, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceGraphModelTests`, `PromoteModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected (R-38 and R-39 already say this)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Dogfood project, a Person with three claims, one of whose Subjects has a filed relationship. Sources → Unpromote on that record's line: the confirm names the relationship under "Also unpromoted"; Keep promotion: nothing changes (AC-1.1, AC-1.3). While it opens, the other Unpromote controls are disabled (AC-1.5).
2. Unpromote again and confirm: the line leaves, the page stays on Sources, focus is on the next Unpromote; open that graph: the Subject is unpromoted and its card stays on the graph (AC-1.2).
3. A Person with one claim: Unpromote shows "PER-… has no other records, so this person is deleted. …" in the danger tone; confirm: the Persons list, without the person, no not-found Callout seen; Back does not show the deleted page (AC-2.1, AC-2.4).
4. On an Evidence graph, Unpromote a person's last member: the same sentence; the graph stays (AC-3.2). Dark mode for the confirm.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Lifting the flow changes the graph's behavior | T-1 runs every graph Unpromote test on the flow |
| The not-found Callout flashes before leaving | D-7 leaves before the detail reload paints; manual item 3 |

**Stop and report if:**

- S13-37 has not landed, or lands without the preflight's bridge ids or `CatalogWriteEffects.deletedHandles` (D-3, D-6).
- S13-16 has not landed, or its `Line` view has no end to append the control to (D-9).
- The board's frames 14n – 14u and this plan disagree.
- `EvidenceGraphSnapshot.build` needs more than the three cached reads (D-4).
- The approved per-kind sentence is no longer what the user wants on the Evidence graph (D-5 changes that confirm's copy).

## 12. Open questions

None. (The board answered RC-1, RC-3 and RC-6 on 2026-11-05: Unpromote as a ghost Button with the glyph and the word at the end of the line; no name beside the ref; no Toast.)

## 13. PR

**Title:** S13-17 — Remove a claim
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md`.
