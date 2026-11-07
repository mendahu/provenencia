# S13-17 — Remove a claim

**Status:** Landed in #729 (2026-11-06)
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/archive/S13-D6-sources-section.md) (revision 2), board agreed 2026-11-05
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Sources.dc.html*, turn "S13-D6 rev 2 · Sources as Source rows, Remove and the last-record warning", frames 14n – 14u, 14y (this PR). **Design → macOS** describes them in words
**Depends on:** S13-08 (landed [#707](https://github.com/mendahu/provenencia/pull/707)), S13-37 (landed [#725](https://github.com/mendahu/provenencia/pull/725)), S13-16 (landed [#722](https://github.com/mendahu/provenencia/pull/722)), S13-D6 (board agreed 2026-11-05)
**Blocks:** —
**Branch:** `s13-17-remove-a-claim`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md) (a page that leaves when its entity is gone)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § Identity claim controls](../person-detail.md#identity-claim-controls) (R-38, R-39); working rules "One Unpromote, deletes through the registry", "Sections are places"; brief [S13-D6](../design/archive/S13-D6-sources-section.md) RC-1 – RC-7; the S13-08 plan (D-14: `getUnpromoteImpact`; D-16: the per-kind deletes copy) and the S13-16 plan
**Planned:** 2026-11-06 against `main` @ `f48846c3` (refreshed after S13-37 and S13-16 landed)
**Landed:** the code follows the plan, except where review round 1's Fix plan 1 supersedes a Decision; eleven notes. (1) `UnpromoteFlowTests` and `evidenceGraphSnapshotReadsTheThreeKeys` seed their graph through `FakeStore` (`subjectsByGraph`, `observationsByGraph`, positions) and let the session load it; `graphRows(…)` stays a private method of `EvidenceGraphModelTests`, without its ignored `types` parameter. (2) `unpromoteLastRecordNamesTheDeletedHandle` lists the Subject's own handle id (`ent-s-james`) so `deletesOwnHandle` holds, which its `message` assertion needs. (3) Supersedes D-10's source of the removed claim: `PersonSourcesList` keeps no `unpromotingClaimID`; `confirmAndTrackFocus` reads the Subject id from `unpromote.pending?.id`, `lineAfterRemoving(subjectID:)` returns the target's claim id, and `applyPendingFocus` tests lines by Subject id (F-3). Keep promotion returns focus to the Unpromote button that had it (`returnFocus`), and Return on a disabled button does nothing. (4) The Source page's `.onChange(of: workspaceHandle.status)` keeps `model.sync(from:)`; only its leave rule moved onto `leavesForSectionRoot`. (5) Supersedes D-1's `beforeApply` and return type, and D-7's order: `UnpromoteFlow.confirm(session:userID:apply:) async -> Bool` (`@discardableResult`) hands the effects to `apply` or applies them itself, and the Person page leaves with `WorkspaceSession.apply(_:_:leavingPage:)`, which hides every key, applies, then leaves, so the Persons list loads once (F-9). The Source page's resource delete leaves the same way (`SourcePageModel.confirmResourceDelete(leaving:)` passes `navigation.fallbackToSectionRoot`), so the deleted Source is not reloaded (F-11). (6) Supersedes D-11's page rule and `ConclusionDetailsError`: `ConclusionDetailPresentation` has a `.gone` case (the page draws its loading page), the page shows a Toast with `conclusionDetailsNotFound` as it leaves (Person, Event and Place), `QueryRead.isConclusionNotFound` is gone, and `EngineErrorCode` (`conclusionNotFound`, `promoteStale`, `sourceMetadataInvalid`; `CoreInvokeError.engineCode`) replaces it with its parity row (F-4, F-6). (7) `PVUnpromoteConfirmContent` and `UnpromoteConfirmItem` carry `connections: PVUnpromoteConnections` (`listed([String])` or `unlisted`) in place of `bridges` and `bridgesUnlisted`; the list scrolls under `PVDeleteImpactLayout.listMaxHeight`, and a write error is always announced (F-1, F-2). (8) A deleted person is announced to VoiceOver (`sourcesPersonDeletedAnnouncement(ref:)`), and each Unpromote button reads graph, record and ref (F-5). (9) The Unpromote copy lives in `L10n.Unpromote` with `unpromote.*` keys (F-10); the S13-23 plan still cites the old `evidenceGraph.subject.unpromote*` keys. (10) Manual check §10 items 1 – 4 were not run in the app. (11) A Person page section whose own read reports the handle gone draws its error Callout with `conclusionDetailsNotFound` instead of a spinner, and `ConclusionDetailLoaded.body` reads `presentation` once (F-12).
**Review:** [`S13-17-remove-a-claim.review.md`](S13-17-remove-a-claim.review.md) — refresh rounds 1 and 2 (2026-11-06): every finding Fixed, no blocker open

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
| AC-1.6 | The claim's graph rows cannot be read, and the preflight names at least one connection | The confirm opens with the preflight's message and, in place of the list, "The connections that come off with it could not be listed. The other end of each stays promoted." (D-3, D-4) |

**US-2.** As a researcher, when I remove a person's last record, I am told the person is deleted first, and land on the Persons list after.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Maria Weber (PER-8JD2L) with one claim; Unpromote | The preflight's deletes group lists PER-8JD2L only; danger tone; "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." (RC-3; D-5) |
| AC-2.2 | The preflight lists PER-8JD2L and two association handles | The message is the DeleteImpact line naming all three, danger tone (RC-3; D-5) |
| AC-2.3 | The preflight fails | Danger tone, "The check for whether this person is deleted could not run. Confirming may still delete this person. The cards stay on the graph.", the error under it, and the could-not-be-listed sentence in place of the list; Unpromote enabled (RC-4; D-3) |
| AC-2.4 | Confirm on AC-2.1 | The write's effects list PER-8JD2L's id in `deletedHandles`; the app shows the Persons list (`navigation.fallbackToSectionRoot()`) without Maria (RC-7; D-6, D-7) |
| AC-2.5 | Confirm on AC-2.3 | Lands on the Persons list exactly when the write's `deletedHandles` holds the person; otherwise stays on Sources (D-6) |
| AC-2.6 | After AC-2.4, the history still holds an earlier entry of Maria's page (Persons list → Maria → Sources is two entries); Back | The app does not stay on her page: the detail read reports `conclusiondetails.not_found` and that entry becomes the Persons list, with no Callout left on screen (RC-7; D-11) |

**US-3.** As a researcher on the Evidence graph, Unpromote behaves as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Unpromote on an Evidence graph card whose handle keeps a member | The same confirm and list as today (the list now from the preflight's bridge ids); the graph stays |
| AC-3.2 | Unpromote of a person's last member on the graph | The per-kind sentence of AC-2.1; danger tone; the graph stays (working rule "One Unpromote") |

## 3. Scope

| In | Covered by |
| --- | --- |
| One Unpromote flow and confirm for both pages, lifted from the Evidence graph: `UnpromoteFlow`, `UnpromoteConfirmItem`, its message, the `PVUnpromoteConfirm` recipe and the `unpromoteConfirm(flow:…)` modifier over it (RC-2 – RC-5) | STEP-2, STEP-3, AC-3.1, AC-3.2 |
| The list from the preflight's bridge ids; the Swift copy of the engine's bridge rule goes (S13-37) | STEP-2, AC-1.1, AC-3.1 |
| The single-handle deletes sentence approved for S13-08 (RC-3) | STEP-3, AC-2.1, AC-3.2 |
| One session read of a graph's snapshot, shared with Promote | STEP-1, AC-1.1, AC-1.6 |
| Unpromote at the end of each record's line, disabled while busy; focus after a kept Unpromote; landing on the Persons list (RC-1, RC-6, RC-7) | STEP-5, AC-1.2, AC-1.5, AC-2.4 |
| A Conclusion page whose handle no longer exists leaves for its list, through one leave modifier the Source page and the graph also use (RC-7) | STEP-6, AC-2.6 |
| One read of a coded error (`CoreInvokeError.code`), which every code-only match moves onto | STEP-6 |
| `userID` reaches the Sources section | STEP-4 |
| Strings | each step that adds a key; STEP-7 checks the catalog |
| Docs, brief archive | STEP-8 |

| Out | Home |
| --- | --- |
| Engine fields for the bridges and the deleted handles | landed in S13-37 |
| A Toast after Unpromote | not planned (brief RC-6), unless the board adds one |
| Leaving a Conclusion page that stays open while another write deletes its entity (Event, Place, another window): it holds a value, so D-11 does not fire | their redesign (R-1); S13-37's `deletedHandles` gives it the fact |
| Removing a deleted entity's entries from the history store | not planned: D-11 rewrites each entry when Back reaches it |
| Bulk Unpromote, Reject, Provisional | not planned (brief § 2.1) |
| How the records connect | S13-28 |

## 4. Current state

The S13-37 and S13-16 rows below describe the landed code.

| Area | Where | Today |
| --- | --- | --- |
| Unpromote state | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `PendingUnpromote{id, name, kind, bridges, deletedHandles, impactCheckFailed, error}` and `confirmTone`; `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen()` | The graph model owns the item, the busy flag and the stale-open guard |
| Unpromote begin | `EvidenceGraphModel.beginUnpromote(subjectID:)` | Guards on `currentSnapshot()`, `primary(in:id:)` and `placed.membership`. Filters `snapshot.bridges` itself (a Swift copy of `identityclaims.UnfileBridges`), sorted by id, worded by `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)`; then `store.getUnpromoteImpact` (`impact.deletes` with `total > 0` → `deletedHandles`; a throw → `impactCheckFailed`, `error`). It ignores `impact.unpromotedBridgeSubjectIDs` |
| Unpromote write | `EvidenceGraphModel.confirmUnpromote()` | `store.unpromoteSubject`, `session.apply(effects)`, closes; a failure stays as `error` |
| Unpromote confirm | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `.pvConfirm(item: $model.pendingUnpromote, copy: …)` and `unpromoteDetail(_:)` | The message is built in the `copy:` closure: `PVDeleteImpactCopy.cascadeLines` first line, else the preflight-failed copy, else `unpromoteConfirmMessage` |
| The approved deletes copy | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` | No caller in `macos/App`; only `macos/ProvenenciaTests/PVDeleteImpactTests.swift` reads it |
| Flow precedent | `macos/App/DesignSystem/Recipes/DeleteImpact/DeleteImpactFlow.swift` `DeleteImpactFlow` (`request`, `isRunning`, `error`, `confirm(perform:) -> Bool`); held by `EvidenceGraphModel.deleteImpact` and `SourcePageModel.deleteImpact`, bound with `.pvDeleteImpact(flow:…)` | The shape `UnpromoteFlow` copies |
| Snapshot read | `macos/App/Features/Promote/PromoteModel.swift` `graphSnapshot()` | `readyValue` on the graph, properties and rules keys, then `EvidenceGraphSnapshot.build(rows:types:rules:)`; nil when the rows are nil |
| Preflight | `GetUnpromoteImpactResponse{deletes, unpromoted_bridge_subject_ids}`; `GenealogyStore.getUnpromoteImpact(projectDir:subjectID:) -> CatalogUnpromoteImpact{deletes: CatalogDeleteImpactGroup?, unpromotedBridgeSubjectIDs: [String]}` | The bridge ids are in bridge Subject id order, and `EvidenceGraphPlacedBridge.id` is that Subject's id. Each `listed` row (`CatalogDeleteImpactListed`) carries the deleted handle's `id` and `ref`; `deleteimpact.ListedGroup` caps `listed` (`listedCap`), the subject's own handle first |
| Effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects` | `deleted: [CatalogWriteEffectsDeleted]` (S13-37 landed it as one list keyed by entity) with computed `deletedHandles: [String]`, the `.handle` entry. `session.apply` ignores `deleted` |
| Not-found error | `core/database/conclusiondetails/conclusiondetails.go` `ErrNotFound` (`apperr.CodeConclusionDetailsNotFound`, wire `conclusiondetails.not_found`, "an unknown or merged handle") | `getConclusionDetail` throws it. D-6 does not read it; D-11 does |
| Leaving a page | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `fallbackToSectionRoot()`; `macos/App/Features/Sources/SourcePageView.swift` (`if await model.confirmResourceDelete() { navigation.fallbackToSectionRoot() }`, and `.onChange(of: workspaceHandle.status)` on `.failed`); `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`.onChange(of: model.graphIsGone, initial: true)`) | Replaces the current entry with the list; earlier entries stay |
| Section history | `ConclusionDetailPage.swift` the switcher's `navigation.go(to: navigation.currentLocation.withPageSection($0))` | A section switch pushes an entry, so a Person page reached through a switch has two or more entries, and `fallbackToSectionRoot` replaces only the last |
| Page on a failed read | `ConclusionDetailPage.swift` the private page body: `presentation` from `ConclusionDetailPresentation(handle.read, …)`; `.failed(message)` draws `PVCallout` | Back to a deleted person stays on a danger Callout. `.conclusionDetail` is `whenHidden: .evict`, so Back reloads and fails. `QueryHandle.read` is `.failed` only with no value held; a page holding a value keeps it |
| View identity | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `.id(place?.viewIdentity)` | Another person's page is another view; `PersonSourcesSection`'s `@State` starts fresh |
| Claims reload | `CatalogQueryRegistry.swift` `.personClaims` refreshes | Follows `touchedHandles` and `.evidenceGraphs`, so the section reloads after Unpromote with no change here |
| FakeStore | `macos/App/Platform/FakeStore.swift` `unpromoteImpactBySubject`, `getUnpromoteImpactError`, `unpromoteSubjectError`, `nextWriteEffects`; `FakeHold` (`updateArtifactHold`) | `getUnpromoteImpact` defaults to no bridge ids, so a test scripts them. `unpromoteSubject` has no hold |
| Page sections | `macos/App/Features/Conclusions/PersonDetailView.swift` `PersonDetailView{session, entityId, userID}`; private `PersonDetailSections(entityId:content:context:)`; `PersonSourcesSection(entityId: entityId)` | The Person view holds `userID` and sections do not get it |
| Coded errors | `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError` | No `code` accessor. Three matches read only the code: `DeleteImpactFlow.isInUse`, `SourceMetadataSection.applySetError`, `PromoteModel`'s `promote.stale` catch |
| Wire parity | `api/ffi/swift_vocabulary_test.go` `wireCase` rows; `macos/App/Platform/ConclusionVocabulary.swift` | `EvidenceGraphUnlinkReason` is an `apperr` code enum with a row; `conclusiondetails.not_found` has none |
| Delayed focus | `macos/App/Platform/FocusClaim.swift` `FocusClaim.claim(_:_:then:)` | Writes twice, because a dismissing sheet can swallow the first; `ConclusionGallerySection` uses it |
| Recipes | `macos/App/DesignSystem/Recipes/DeleteImpact/` `pvDeleteImpact(item:…)` over `.pvConfirm`; `macos/App/DesignSystem/README.md` built table; `scripts/check-guidance-names.py` | A recipe takes neutral props and may not name a `Features/` type; each new kit folder needs a README row |
| Sources section | `macos/App/Features/Conclusions/PersonSourcesContent.swift` `PersonSourcesContent.Line{id (claimID), subjectID, graphID, graphTitle, graphRef, subjectName, subjectRef, accessibilityLabel, location}`; `PersonSourcesSection.swift` `PersonSourcesSection(entityId:)`, `PersonSourcesList`, private `PersonSourcesRowView` and `PersonSourcesLineView` | The section builds the content inside `ConclusionPageSectionHost`'s closures, so only that closure sees it. `PersonSourcesList` owns `@FocusState focusedLine: String?` (the graph links), set on an unfold from `firstRevealedLineID` / `firstRevealedRowLineID`. The line is an `HStack(spacing: 0)` ending with the Subject ref; the name truncates first. The file's `#if DEBUG` preview builds `PersonSourcesList` |
| Tests | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges` (its second half asserts the Swift filter drops an unfiled bridge), `unpromoteWithNoBridgesHasEmptyDetail`, `confirmUnpromoteRefreshesToPromote`, `unpromoteFailureStaysInConfirm`, `unpromoteLastRecordNamesTheDeletedHandle` (asserts `deletedHandles?.listed.first?.ref`), `unpromotePreflightFailureStaysInTheConfirm`, `unpromoteVoiceOverActionFollowsFooterAction`; `macos/ProvenenciaTests/PersonDetailTests.swift` | Cover the graph's flow; `PersonDetailTests` asserts `.failed` presentation for `conclusiondetails.not_found` |

**Precedent.** `DeleteImpactFlow`: one observable flow object a model holds and a view modifier binds, shared by the Evidence graph and the Source page; copy its shape for `UnpromoteFlow`. The Source page's resource delete: a successful write that removes the page's entity, then `fallbackToSectionRoot()`; copy that.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One flow: `@MainActor @Observable final class UnpromoteFlow` *(new)* in `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` *(new)* owns `pending: UnpromoteConfirmItem?`, `isRunning`, `isBusy` (an open is in progress or the write runs; a private open count, so a dropped stale open cannot clear it), the open generation, `open(subjectID:name:kind:handleID:graphID:session:) async` and `confirm(session:userID:beforeApply: @MainActor (CatalogWriteEffects) -> Void = { _ in }) async -> CatalogWriteEffects?` (write through `session.store`, `beforeApply(effects)`, `session.apply(effects)`, close, return the effects; a failure stays as `pending.error` and returns nil) and `cancel()`. `EvidenceGraphModel` holds `let unpromote = UnpromoteFlow()` (as it holds `deleteImpact`); its `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen` and `confirmUnpromote` go; `beginUnpromote(subjectID:)` keeps its guards and calls `unpromote.open` with `handleID: placed.membership?.entity.id` and `graphID: graphID`. Names come from `session.nameFormat` | One place for the guard, the generation and the error rule; the `DeleteImpactFlow` precedent (round 1 M-1). `Features/Catalog` is where cross-feature code sits, so Conclusions does not reach into EvidenceGraph (user, 2026-11-06) | A second state machine in a Person-page model; a `snapshot:` value or closure; a result type around the effects |
| D-2 | Recipe and item, split (user, 2026-11-06). `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` *(new)* holds `PVUnpromoteConfirmContent{id, name, message, tone: PVConfirmTone, bridges: [String], bridgesUnlisted: Bool, error: String?, announcesError: Bool}` *(new)* and `View.pvUnpromoteConfirm(item: Binding<PVUnpromoteConfirmContent?>, isRunning: Bool, accessibilityIdentifierPrefix: String, onConfirm: @escaping () -> Void)` *(new)*: today's `pvConfirm` call and `unpromoteDetail` view, moved. It names no `Features/` type. `UnpromoteConfirmItem` *(new, `UnpromoteFlow.swift`)* = today's `PendingUnpromote` fields plus `handleID: String?` (the Subject's own handle) and `bridgesUnlisted: Bool`; computed `deletesOwnHandle` (`deletedHandles` non-nil, `total == 1`, its one listed row's id == `handleID`), `confirmTone` (today's rule), `message: String` (D-5) and `content: PVUnpromoteConfirmContent` (`announcesError` is `impactCheckFailed`). `View.unpromoteConfirm(flow: UnpromoteFlow, accessibilityIdentifierPrefix: String, onConfirm: @escaping () -> Void)` *(new, same file)* calls the recipe with a binding whose getter is `flow.pending?.content` and whose nil setter is `flow.cancel()`, as `pvDeleteImpact(flow:…)` does. `EvidenceGraphView` calls it with `"evidenceGraph.unpromote"` | `add-ui-component`: a second call site goes to a recipe, and a recipe maps domain to props. The copy is testable on the item (round 1 X-3); one rule for "deletes its own handle" (round 1 M-5) | Message logic in a view closure; the whole flow in the recipe (it needs `WorkspaceSession` and `EvidencePrimaryKind`, which the kit may not name) |
| D-3 | The list comes from the engine. `open` calls the preflight first; when `impact.unpromotedBridgeSubjectIDs` is not empty it reads `session.evidenceGraphSnapshot(graphID:)` and words each id, in the preflight's order, with `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)`, matching `snapshot.bridges` on `id`; an id not in the snapshot is skipped. `bridgesUnlisted` is true when the preflight failed, or when it named bridges and no sentence came out. The Swift filter in `beginUnpromote` goes | The engine's rule in one layer (round 1 M-3; user, 2026-11-05: preparatory PR S13-37). With no preflight there are no ids, and an empty list would read as "no connections" where today's confirm listed them | Keeping the Swift filter; a silent empty list on a failed preflight |
| D-4 | One snapshot read: `WorkspaceSession.evidenceGraphSnapshot(graphID:) async -> EvidenceGraphSnapshot?` *(new)* in `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`: `query` (which warms a cold or stale key), then `readyValue`, on `.evidenceGraph(project: projectKey, graphId:)`, `.propertiesWorkspace(project:)` and `.connectRules(project:)`; nil when the rows are nil; missing types or rules → `[]` (Promote's rule). `PromoteModel.graphSnapshot()` becomes a call to it, so Promote's read now warms a cold key; the Promote place already warms all three, so a real visit is unchanged. Promote's own `session.reread(graphKey)` stays, and `propertiesKey` and `rulesKey` stay for `PromoteView`. The confirm's title is always the name the caller passes (the page passes `Line.subjectName`) | One place decides a snapshot's keys (round 1 M-4); a missing list never hides the confirm | A third copy of the three reads; a `warm:` flag or a second read for Promote |
| D-5 | `UnpromoteConfirmItem.message`: `deletesOwnHandle` → `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` with the listed row's ref; another non-empty deletes group → `PVDeleteImpactCopy.cascadeLines(...)`'s first line (today's); `impactCheckFailed` → the preflight-failed copy; else `unpromoteConfirmMessage`. `bridgesUnlisted` shows `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted` *(new)* in the detail in place of the list and its note, above any error | The copy the user approved on 2026-11-04 (S13-08 D-16), defined and never called; the many-handle case keeps today's wording, as the brief says | New Person-page copy; leaving the key dead |
| D-6 | The person is gone when the write's `effects.deletedHandles` contains the page's entity id. No preflight or detail read decides it | One fact, one source: what the write did (round 1 M-2; user, 2026-11-05) | Inferring from the preflight; a detail read deciding it |
| D-7 | On "gone" the section calls `navigation.fallbackToSectionRoot()` from `confirm`'s `beforeApply`, so the app leaves before `session.apply`. `onLocationCommit` updates the session's visible keys in the same call, so the person's `.conclusionDetail` and `.personClaims` are hidden when `apply` runs: their `.evict` policy drops them and nothing reloads a handle the write reported deleted. The entry is replaced by the Persons list | No failing read and no not-found Callout to race; the Source page's `fallbackToSectionRoot` precedent, where a replaced entry cannot be returned to (RC-7; round 1 X-6) | `go(to: .sectionRoot(.persons))`; leaving after `apply` and racing the reload |
| D-8 | `userID` reaches the section through the Person page only: `PersonDetailView` passes it to `PersonDetailSections(entityId:userID:content:context:)`, which passes it to `PersonSourcesSection(entityId:userID:)`. The section holds `@State private var unpromote = UnpromoteFlow()`; another person's page is another view (`viewIdentity`), so no reset is needed | The Person view already holds `userID` beside `entityId`; no Event or Place section reads it | A field on the shared `ConclusionDetailPageSectionsContext`, with one reader and a route through private `ConclusionDetailLoaded` (this supersedes round 1 M-7); an `.onChange(of: entityId)` that never fires |
| D-9 | At the end of each record's line, `PVSpacing.space3` (6pt) after its Subject ref: `PVButton(L10n.EvidenceGraph.unpromote, variant: .ghost, size: .sm, icon: .unpromote)` (the glyph and the word, board 14a: an icon alone at the far right read as detached, user 2026-11-05), its VoiceOver label `L10n.Conclusions.sourcesUnpromoteAccessibility(name:)` ("Unpromote J. Weber"), id `persons.detail.sources.unpromote.<claimID>`, outside the line's `accessibilityHidden` group, part of the line (the lines' Unpromote buttons do not align in a column), disabled while `unpromote.isBusy`, `.fixedSize()` so the Subject name truncates first. Section side: `begin(line)` runs `Task { await unpromote.open(subjectID: line.subjectID, name: line.subjectName, kind: .person, handleID: entityId, graphID: line.graphID, session: session) }`; `confirm() async -> Bool` runs `await unpromote.confirm(session: session, userID: userID, beforeApply: { if $0.deletedHandles.contains(entityId) { navigation.fallbackToSectionRoot() } }) != nil`. `line.subjectName` is the graph card's name rule (S13-16 D-2), so the line, the button's label and the confirm's title agree with or without a snapshot | The action is called Unpromote everywhere in code (round 1 M-6) | "Remove" names in code |
| D-10 | Signatures: `enum PersonSourcesFocus: Hashable { case graph(claimID: String), unpromote(claimID: String) }` *(new, section file)*; `PersonSourcesList` gains `unpromote: UnpromoteFlow`, `beginUnpromote: (PersonSourcesContent.Line) -> Void` and `confirmUnpromote: () async -> Bool`; `PersonSourcesContent.lineAfterRemoving(claimID: String) -> String?` *(new)* returns the claim id of the next visible line in reading order (Source rows in order, their visible lines in order, flattened), else the previous one, else nil. The list keeps its focus state, typed `@FocusState private var focus: PersonSourcesFocus?` (it replaces `focusedLine`; the unfold paths set `.graph(claimID:)` from `firstRevealedLineID` and `firstRevealedRowLineID`, and `PersonSourcesRowView` and `PersonSourcesLineView` take `FocusState<PersonSourcesFocus?>.Binding`, `unpromoteDisabled: Bool` (`unpromote.isBusy`) and `beginUnpromote`). The Unpromote button takes `.focused(focus, equals: .unpromote(claimID: line.id))` and `.pvActivatesOnReturn(when:)` as the graph link does. The list carries `.unpromoteConfirm(flow: unpromote, accessibilityIdentifierPrefix: "persons.detail.sources.unpromote", onConfirm: …)`; `onConfirm` stores `@State pendingFocus: (claimID: String, target: String?)?` from `lineAfterRemoving`, then `Task { if await !confirmUnpromote() { pendingFocus = nil } }`. `.onChange(of: content)` on the list applies it when no visible line of the new content has that claim id: `FocusClaim.claim($focus, .unpromote(claimID: target))`, then clears it. The `#if DEBUG` preview passes a fresh `UnpromoteFlow()` and no-op closures. No Toast (board 14t) | Brief RC-6, board 14t; focus never lands on a removed control (round 1 X-8). The list already holds `content`, the focus and both unfold paths, and the walk sits with `firstRevealedLineID`. Captured before the write, a reload that beats `confirm`'s return cannot miss it. RC-6's third branch, the first Source title, is unreachable: every row has a line, so with no other line there is no row | Moving focus at confirm time (the line is still there); lifting focus to the section; a bare focus write after a sheet closes |
| D-11 | A Conclusion page leaves when its read reports the handle is gone. `CoreInvokeError.code: String?` *(new, `CoreInvoke.swift`)*, written `switch self { case .coded(_, let code, _, _): code; case .failed: nil }`. `enum ConclusionDetailsError: Sendable, Equatable { case notFound; init?(wire: String) }` *(new, `ConclusionVocabulary.swift`, in the `case "…": self = ` shape)*. `QueryRead.isConclusionNotFound` *(new, `ConclusionDetailPage.swift`)*: `.failed` whose error's `code` maps to `.notFound`. `View.leavesForSectionRoot(when: Bool)` *(new, a `ViewModifier` reading `@Environment(WorkspaceNavigation.self)` and its `View` extension, `WorkspaceNavigation.swift`)*: `.onChange(of: gone, initial: true)`, calling `fallbackToSectionRoot()` when true. `ConclusionDetailLoaded` applies it with `handle.read.isConclusionNotFound`, and its `.failed` branch draws `loadingPage` while that is true, so the Callout never paints. Any other failure keeps the Callout. `EvidenceGraphView` (`model.graphIsGone`) and `SourcePageView` (`workspaceHandle.read.isFailed`, *new* beside `isLoading`) move onto the modifier; the Source page keeps `.onChange(of: workspaceHandle.status) { model.sync(from:) }`, only its `fallbackToSectionRoot()` branch moves, and it gains `initial: true`, which only adds an already failed read on mount. The shared page means Event and Place gain it | RC-7 says Back does not return, and section switches push entries that D-7 cannot reach. It is the third "leave when gone", so one modifier; the parity row fails CI on a renamed code | Removing the entity's entries from `NavigationHistoryStore`; leaving on any failure as the Source page does (a transient read failure would drop the researcher's place); a fourth inline `.coded` match; moving `ErrNotFound` to `apperr.KindNotFound` (an engine change this PR does not need) |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `UnpromoteFlow` *(new)* | type | `DeleteImpactFlow`; `EvidenceGraphModel`'s Unpromote state | Replaces the model's state; same shape as the delete flow |
| `UnpromoteConfirmItem` *(new)* | type | `EvidenceGraphModel.PendingUnpromote` | It is that type, moved, with `handleID`, `bridgesUnlisted`, `message`, `content` |
| `PVUnpromoteConfirm` recipe, `View.unpromoteConfirm(flow:…)` *(new)* | recipe, modifier | `Recipes/DeleteImpact`, `.pvDeleteImpact(flow:…)`; `EvidenceGraphView`'s inline `.pvConfirm` | Moved; the recipe takes neutral props, the modifier maps the flow onto them |
| `WorkspaceSession.evidenceGraphSnapshot(graphID:)` *(new)* | func | `PromoteModel.graphSnapshot()` | Its body, moved to the session |
| `unpromoteConfirmBridgesUnlisted`, `sourcesUnpromoteAccessibility(name:)` *(new keys)* | keys | `unpromoteConfirmPreflightFailed`, `L10n.EvidenceGraph.unpromote` | New sentences |
| `UnpromoteFlowTests` *(new)* | test file | `EvidenceGraphModelTests`' Unpromote cases | The flow is tested once; the callers test their calls |
| `PersonSourcesFocus`, `PersonSourcesContent.lineAfterRemoving(claimID:)` *(new)* | enum, func | `focusedLine: String?`; `firstRevealedLineID` | Two control kinds on a line now take focus; the enum carries both |
| `CoreInvokeError.code`, `ConclusionDetailsError`, `QueryRead.isConclusionNotFound`, `QueryRead.isFailed` *(new)* | accessor, enum, properties | The three inline `.coded` matches; `EvidenceGraphUnlinkReason` | One read of a coded error; the three move onto it (same-pass cleanup); the enum gets the parity row |
| `View.leavesForSectionRoot(when:)` *(new)* | modifier | `SourcePageView`'s `.onChange(of: workspaceHandle.status)`, `EvidenceGraphView`'s `.onChange(of: model.graphIsGone)` | Both move onto it |
| `FakeStore.unpromoteSubjectHold`, `getUnpromoteImpactHold` *(new)* | test seams | `updateArtifactHold` | The same `FakeHold`, on these two calls |

**Builds on.** S13-37's `unpromotedBridgeSubjectIDs` and `deletedHandles`; S13-08's `getUnpromoteImpact` and `unpromoteSubject`; `DeleteImpactFlow`'s shape; `EvidenceGraphSnapshot.build`; `fallbackToSectionRoot`; `FocusClaim`; `PVConfirm`; `pvActivatesOnReturn`; `FakeHold`; S13-16's record line (`PersonSourcesContent.Line`, `PersonSourcesList`).
**Deletes.** `EvidenceGraphModel.PendingUnpromote`, `pendingUnpromote`, `isUnpromoting`, `unpromoteOpenGeneration`, `beginUnpromoteOpen`, `confirmUnpromote`; the Swift bridge filter in `beginUnpromote`; `EvidenceGraphView`'s inline Unpromote `.pvConfirm` and `unpromoteDetail`; the body of `PromoteModel.graphSnapshot()`; `PersonSourcesList.focusedLine`; the inline `.coded` matches in `DeleteImpactFlow.isInUse`, `SourceMetadataSection.applySetError` and `PromoteModel`; the `.onChange` leave rules in `SourcePageView` and `EvidenceGraphView`. `unpromoteConfirmMessageDeletes` gains its first caller.
**Next adds.** An Event or Place page with sections holds an `UnpromoteFlow`; any page that leaves when its entity is deleted reads `deletedHandles` before `apply`, as D-7 does; any other coded error reads `CoreInvokeError.code`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Unpromote on a Person page → the shared confirm | `UnpromoteFlow.swift` *(new)* | The item is built fresh per open; a stale open is dropped by the generation (T-4) |
| A claim whose graph rows cannot be read, or whose Subject is gone from it | `UnpromoteFlow.open` | Opens with `bridgesUnlisted` (AC-1.6); the engine re-checks in the write |
| The person deleted by this page's write | `PersonSourcesSection` | Persons list before `apply` (D-7) |
| An earlier history entry of a deleted person | `ConclusionDetailPage` | The read fails with not-found; D-11 rewrites the entry, and the Callout does not paint (AC-2.6) |
| A Person page unmounting while an open is in flight | `UnpromoteFlow` | The flow is the section's `@State`, so the late result lands on a flow no view shows; a page never sees another person's item (`viewIdentity`) |
| A reload that empties the section while the confirm is open | `PersonSourcesList` | The sheet needs another write to do this, and the sheet is modal. If it happens, the list and its modifier unmount and the sheet closes; `pending` stays on the flow and is replaced on the next open |

### Scale

One Unpromote: one preflight, one write and, when the preflight names bridges, one `evidenceGraphSnapshot` (warm when the researcher came from that graph; else one `.evidenceGraph` load, the cost of opening it; the two lists are warm, S13-16 warms them). The effects refresh the per-person keys as Unpromote on the graph does today. Nothing grows with the person's claim count; one PVButton is added per drawn line, and drawn lines are bounded by S13-16's folds (5 per Source until unfolded, 50 Sources).

### Edges

- **Assumptions.** The preflight's bridge ids are exactly what the write unfiles (S13-37's contract, tested there). `listed` is capped but lists the Subject's own handle first; `deletesOwnHandle` needs only `total == 1`, so the cap never hides it. The gone decision reads the effects, not `listed` (D-6). `onLocationCommit` runs `session.apply(location:)` in the same call (`WorkspaceView`), which D-7 relies on; T-5 proves the keys are evicted, not reloaded.
- **Neighbors.** Two claims: one removed keeps the person; the last deletes. A single deleted row that is an association handle (the person keeps a member): `deletesOwnHandle` is false (id ≠ `handleID`), so the `cascadeLines` line shows (T-2).
- **Removed inputs.** The claim already removed elsewhere: the preflight or the write fails with the engine's error in the confirm (AC-1.4); the reload drops the line. A bridge id not in the cached snapshot: skipped; all of them skipped sets `bridgesUnlisted` (T-3). A merged handle: D-11 lands on the list, as it does for a deleted one.
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
| Unpromote control | `PVButton` (ghost, `.sm`, `PVSymbol.unpromote` and the word) at the end of the line | `PersonSourcesSection.swift` (`PersonSourcesLineView`) | D-9, RC-1 |
| Flow and confirm | `UnpromoteFlow`, `unpromoteConfirm`; `pvUnpromoteConfirm` over `PVConfirm` | `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` *(new)*; `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` *(new)* | D-1 – D-5, RC-2 – RC-5 |
| Deletes line | `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes`, `PVDeleteImpactCopy` | as today | D-5, RC-3 |
| Landing | `WorkspaceNavigation.fallbackToSectionRoot()` | — | D-7, RC-7 |

States: opening (controls disabled), writing (`isRunning`), failed (error under the list), preflight failed, list unavailable. Accessibility: the control reads "Unpromote J. Weber"; the confirm reads as today; focus per D-10. The board's frames are described here in words on agreement.

### Strings

L10n keys *(new)*. `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted`: static `LocalizedStringResource`, key `evidenceGraph.subject.unpromoteConfirmBridgesUnlisted`, value "The connections that come off with it could not be listed. The other end of each stays promoted.", comment "Unpromote confirm, in place of the connection list, when the engine or the graph could not supply it". `L10n.Conclusions.sourcesUnpromoteAccessibility(name:) -> String` through `L10n.format`, as `sourcesOpenGraph`: key `conclusions.sources.unpromoteAccessibility`, value "Unpromote %@", comment "Spoken label of the Unpromote button on a record's line on a Person's Sources section; the argument is the record's name". `evidenceGraph.subject.unpromote`'s comment becomes "Footer icon button's tooltip and VoiceOver action on a promoted card, the confirm button, and the Person page's button on a record's line, that takes its promotion back". Existing copy now also shown from the Person page, each still true there (the cards are on the claim's graph): `evidenceGraph.subject.unpromoteConfirmMessage` ("This promotion is removed. The cards stay on the graph, and the handle stays."), `unpromoteConfirmPreflightFailed.*`, `unpromoteConfirmMessageDeletes.*`. Removed: none.

## 7. Steps

**STEP-1 — One snapshot read**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`.
Do: D-4, with the two `PromoteModelTests` rewrites.
Check: T-7, T-11 pass; the other `PromoteModelTests` pass unchanged.

**STEP-2 — One Unpromote flow**
Files: `macos/App/Features/Catalog/Unpromote/UnpromoteFlow.swift` *(new)*, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `EvidenceGraphView.swift` (binds `model.unpromote` for now), `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/EvidenceGraphRowsFixture.swift` *(new)*, `macos/ProvenenciaTests/UnpromoteFlowTests.swift` *(new)*, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-1, D-3, and D-2's item without the `message` rule or `content`; the graph model moves onto the flow. `FakeStore` gains `unpromoteSubjectHold` and `getUnpromoteImpactHold` (`FakeHold`): each method becomes two `withState` blocks with the hold awaited between them, recording the call in the first, as `updateArtifact` does. `UnpromoteFlowTests` builds its session as `EvidenceGraphModelTests.makeModel` does (`WorkspaceSession(projectKey:store:)` on a `FakeStore`), seeds the graph with `store.subjectsByGraph` and `session.setQueryValue` (`graphRows(…)`, which moves from a private method of `EvidenceGraphModelTests` to an internal function in `macos/ProvenenciaTests/EvidenceGraphRowsFixture.swift` *(new)*), the Subject's membership with `store.seedPromotion(projectDir:subjectID:)`, and the effects with `store.nextWriteEffects = CatalogWriteEffects(deleted: [CatalogWriteEffectsDeleted(entity: .handle, ids: ["h-maria"])])`.
Check: T-1 (without its `item.message` assertion), T-3, T-4, T-5, T-6 pass; `rg 'endpointAID == subjectID' macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` finds nothing.

**STEP-3 — One confirm and its message**
Files: `UnpromoteFlow.swift`, `macos/App/DesignSystem/Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` *(new)*, `macos/App/DesignSystem/README.md` (a row for the new folder), `EvidenceGraphView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/UnpromoteFlowTests.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-2's recipe, modifier, `message` and `content`, D-5 (with `unpromoteConfirmBridgesUnlisted`); `unpromoteDetail` moves into the recipe and shows the sentence when `bridgesUnlisted` is true.
Check: T-2 and T-1's `item.message` assertion pass.

**STEP-4 — The Sources section gets the user id**
Files: `macos/App/Features/Conclusions/PersonDetailView.swift`, `PersonSourcesSection.swift`.
Do: D-8's parameters.
Check: the build has no warning; `PersonDetailTests` pass.

**STEP-5 — Unpromote on the record lines**
Files: `macos/App/Features/Conclusions/PersonSourcesSection.swift`, `PersonSourcesContent.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/PersonSourcesContentTests.swift`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-6, D-7, D-9, D-10, with `sourcesUnpromoteAccessibility(name:)` and `evidenceGraph.subject.unpromote`'s new comment.
Check: T-8 passes; manual check items 1 – 4.

**STEP-6 — A gone handle leaves its page**
Files: `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/ConclusionVocabulary.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/App/Features/Workspace/Session/QueryHandle.swift`, `macos/App/Features/Workspace/WorkspaceNavigation.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/DeleteImpactFlow.swift`, `macos/App/Features/Sources/SourceMetadataSection.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`. Skill: `add-workspace-location`.
Do: D-11, and move the three code-only matches onto `CoreInvokeError.code`.
Check: T-9, T-10 pass; `rg -n "\.coded\(_, let code, _, _\)" macos/App` prints the one line of `CoreInvokeError.code`; manual check item 3.

**STEP-7 — Strings check**
Files: `macos/App/Resources/Localizable.xcstrings`.
Do: confirm both new keys have entries and comments.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `docs/macos-client-patterns.md` (its "until S13-17" sentence: a page that shows a deleted id leaves before `apply`), `macos/App/Features/Workspace/Session/CatalogMutation.swift` (the same on `deleted`'s comment), `docs/deployment-plan/spike-13/design/S13-D6-sources-section.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` existing Unpromote cases (moved onto `model.unpromote`) | Behavior unchanged, with the bridge ids scripted in `store.unpromoteImpactBySubject`; `unpromoteActionOpensConfirmNamingBridges`'s `fileSonOf: false` half goes (the engine's rule, tested in S13-37 T-2) and `unpromotePreflightFailureStaysInTheConfirm` also asserts `bridgesUnlisted`; `unpromoteLastRecordNamesTheDeletedHandle` also asserts `item.message` is the per-kind sentence | AC-3.1, AC-3.2, D-1 |
| T-2 | Swift | `macos/ProvenenciaTests/UnpromoteFlowTests.swift` `messageFollowsThePreflight` *(new)* | `handleID` "h-maria", group of one listed "h-maria" (PER-8JD2L) → "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph."; group of three → the `cascadeLines` first line; group of one listed "h-rel" → the `cascadeLines` line; `impactCheckFailed` → the preflight-failed copy; none → `unpromoteConfirmMessage`; tone `.danger` for the first four | AC-2.1, AC-2.2, AC-2.3, AC-3.2, D-2, D-5 |
| T-3 | Swift | same · `listComesFromThePreflight` *(new)* | Preflight bridge ids [b2, b1] with both in the snapshot → two sentences in that order; [b2, gone] → one, `bridgesUnlisted` false; [gone] → `bridgesUnlisted` true; rows load fails (`listSubjectsError`) with ids → `bridgesUnlisted` true, no sentences; no ids and a failing rows load → `bridgesUnlisted` false; preflight throws → `bridgesUnlisted` true | AC-1.1, AC-1.6, D-3, D-4 |
| T-4 | Swift | same · `staleOpenIsDropped` *(new)* | `open` for A, then for B before A's preflight returns (`getUnpromoteImpactHold`) → `pending` is B's; `isBusy` stays true until B ends | D-1 |
| T-5 | Swift | same · `confirmReportsTheEffects` *(new)* | Write succeeds with `deletedHandles` ["h-maria"] → the returned effects hold it, `pending` nil, `isBusy` false; with none → `deletedHandles` empty (the person stayed); `isBusy` true while `unpromoteSubjectHold` suspends the write and while `getUnpromoteImpactHold` suspends the open; `beforeApply` sees the effects before `apply`: when it calls `session.apply(location: .sectionRoot(.persons))`, her `.conclusionDetail` key is evicted and no load runs | AC-1.2, AC-1.5, AC-2.4, AC-2.5, D-1, D-6, D-7 |
| T-6 | Swift | same · `confirmFailureStaysOpen` *(new)* | Write throws → `confirm` returns nil, `pending.error` set, retry allowed | AC-1.4 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `evidenceGraphSnapshotReadsTheThreeKeys` *(new)* | Seeded graph rows, properties and rules → a snapshot with the Subject; a stale graph key → a snapshot, after `query` reloads it; rows load fails → nil | AC-1.6, D-4 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `lineAfterRemoving` *(new)* | A middle line → the next line; a Source's last line → the first line of the next Source; the page's last line → the previous; the only line of the only Source → nil; a line folded away by `lineFoldLimit` is never returned | AC-1.2, D-10 |
| T-9 | Swift | `macos/ProvenenciaTests/PersonDetailTests.swift` `conclusionNotFoundIsReadFromTheFailure` *(new)* | `QueryRead.failed(CoreInvokeError.coded(… "conclusiondetails.not_found" …))` → true; `.failed` with another code → false; `.failed(CoreInvokeError.failed(status: 1))` → false; `.loading`, `.loaded` → false | AC-2.6, D-11 |
| T-10 | Go | `api/ffi/swift_vocabulary_test.go` new row "conclusion detail errors", enum `ConclusionDetailsError`, `re: wireCase`, want `apperr.CodeConclusionDetailsNotFound` | The Swift enum's wire strings match the Go code | D-11 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `aPageWithoutItsGraphSaysSoInsteadOfLookingEmpty`, `tryAgainReadsAFailedGraphAgain` (rewritten) | `store.listSubjectsError` set, then `load()` → `loadError == L10n.Promote.graphUnavailable`; the second clears the error and tries again. Neither relies on a cold key | D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceGraphModelTests`, `PromoteModelTests`, and any `SourcePageView` test of a failed workspace read.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: `docs/macos-client-patterns.md` (STEP-8); R-38 and R-39 already say the rest
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Dogfood project, a Person with three claims, one of whose Subjects has a filed relationship. Sources → Unpromote on that record's line: the confirm names the relationship under "Also unpromoted"; Keep promotion: nothing changes (AC-1.1, AC-1.3). While it opens, the other Unpromote controls are disabled (AC-1.5).
2. Unpromote again and confirm: the line leaves, the page stays on Sources, focus is on the next Unpromote; open that graph: the Subject is unpromoted and its card stays on the graph (AC-1.2).
3. A Person with one claim: Unpromote shows "PER-… has no other records, so this person is deleted. …" in the danger tone; confirm: the Persons list, without the person, no not-found Callout seen (AC-2.1, AC-2.4). Open another one-claim Person from the Persons list, switch to Sources, Unpromote and confirm, then Back: the app lands on the Persons list, not her page (AC-2.6).
4. On an Evidence graph, Unpromote a person's last member: the same sentence; the graph stays (AC-3.2). Dark mode for the confirm.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Lifting the flow changes the graph's behavior | T-1 runs every graph Unpromote test on the flow |
| The not-found Callout flashes before leaving | D-7 leaves before `apply`; D-11 draws the loading page; manual item 3 |
| A Conclusion link to a merged handle now lands on the list, not on a Callout saying it was merged | Accepted in D-11: a merged handle has no page either way, and the page's Callout offers no way forward |

**Stop and report if:**

- The board's frames 14n – 14u and this plan disagree.
- `EvidenceGraphSnapshot.build` needs more than the three cached reads (D-4) (D-4).
- An existing test or flow depends on a Conclusion page holding its Callout for `conclusiondetails.not_found` on a location other than a deleted handle (D-11).
- The approved per-kind sentence is no longer what the user wants on the Evidence graph (D-5 changes that confirm's copy).

## 12. Open questions

None. (The board answered RC-1, RC-3 and RC-6 on 2026-11-05: Unpromote as a ghost Button with the glyph and the word at the end of the line; no name beside the ref; no Toast.)

## 13. PR

**Title:** S13-17 — Remove a claim
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md`.
