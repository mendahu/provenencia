# S13-17 — Remove a claim

**Status:** Blocked — on board S13-D6
**Spike / slice:** Spike 13 · Slice 6 — Sources
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D6](../design/S13-D6-sources-section.md), board not drawn yet
**Board:** — (not drawn yet; on agreement this line names the Claude Design turn and frames, and **Design → macOS** describes every frame in words)
**Depends on:** S13-08 (landed [#707](https://github.com/mendahu/provenencia/pull/707)), S13-16, S13-D6 (board)
**Blocks:** —
**Branch:** `s13-17-remove-a-claim`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md) (leaving a page whose entity is gone), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [`../person-detail.md` § Identity claim controls](../person-detail.md#identity-claim-controls) (R-38, R-39); working rules "One Unpromote, deletes through the registry", "Sections are places"; brief [S13-D6](../design/S13-D6-sources-section.md) RC-1 – RC-7; the S13-08 plan (D-14: `getUnpromoteImpact`; D-16: the per-kind deletes copy) and the S13-16 plan
**Planned:** 2026-11-05 against `main` @ `124d492` (draft: refreshed when S13-16 lands and the board is agreed)
**Review:** [`S13-17-remove-a-claim.review.md`](S13-17-remove-a-claim.review.md) — round 1 on the draft

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

A researcher can take a record out of a person from the Person page: Remove on a claim card in the Sources section unpromotes that Subject, exactly as Unpromote on the Evidence graph does (R-38), through the same confirm (user, 2026-11-05). Removing the last claim deletes the person; the confirm says so first, and on proceeding the app lands on the Persons list (R-39).

**Why now.** S13-08 made the engine delete a handle whose last member leaves and gave Unpromote its preflight (`GetUnpromoteImpact`); S13-16 draws the claim cards with a footer slot for Remove. This PR closes slice 6 (bar 7, removal half).

**What stays.** The engine: `UnpromoteSubject` and `GetUnpromoteImpact` are called as they are; no Go or proto change. The Evidence graph's Unpromote confirm keeps its title, buttons, list, tone and placement; it moves into one shared confirm both pages use, and its single-handle deletes message becomes the per-kind sentence the user approved for S13-08 (2026-11-04, S13-08 D-16), which the landed code defined but does not show (D-6).

## 2. User stories

**US-1.** As a researcher, when a Subject was promoted into the wrong person, I remove it from the Person page, so that I do not have to find its Evidence graph first.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann Weber (PER-3HK9W) with seven claims; click Remove on the Household 388 card (Subject "J. Weber", CPR-2M8XF, with one relationship and one participation filed) | The confirm opens (`UnpromoteConfirmItem` with `handleID` PER-3HK9W's id): title "Unpromote J. Weber?", message "This promotion is removed. The cards stay on the graph, and the handle stays.", "Also unpromoted" with the two connection sentences and "The other end of each stays promoted.", buttons "Unpromote" (irreversible tone) and "Keep promotion" (RC-1, RC-2) |
| AC-1.2 | Confirm | `unpromoteSubject` runs once for CPR-2M8XF; the confirm closes; the card leaves the section after the reload; the Person page stays on Sources (RC-6) |
| AC-1.3 | Keep promotion | No write; the confirm closes; nothing changes |
| AC-1.4 | The write fails (`promote.stale`) | The confirm stays open with the error under the list; "Unpromote" is enabled to retry (RC-5) |
| AC-1.5 | While a Remove is writing | Every Remove control in the section is disabled (RC-1) |

**US-2.** As a researcher, when I remove a person's last record, I am told the person is deleted before it happens, and land on the Persons list after.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Maria Weber (PER-8JD2L) with one claim; click Remove | The preflight lists PER-8JD2L only; the confirm is in the danger tone with "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." (RC-3; D-6) |
| AC-2.2 | The preflight lists PER-8JD2L and two association handles | The message is the DeleteImpact line naming all three (`PVDeleteImpactCopy.cascadeLines`), danger tone (RC-3; D-6) |
| AC-2.3 | The preflight fails | Danger tone, "The check for whether this person is deleted could not run. Confirming may still delete this person. The cards stay on the graph.", the error under it; Unpromote enabled (RC-4) |
| AC-2.4 | Confirm on AC-2.1 | The write runs; the app shows the Persons list (`navigation.fallbackToSectionRoot()`), which no longer lists Maria; Back does not return to her page (RC-7; D-8) |
| AC-2.5 | Confirm on AC-2.3 when the write did delete the person | Same as AC-2.4 (D-7) |
| AC-2.6 | Confirm on AC-2.3 when the person kept another record | The page stays on Sources (D-7) |

**US-3.** As a researcher on the Evidence graph, Unpromote behaves as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Unpromote on a graph card whose handle keeps a member | The same confirm, copy and list as today; the graph stays where it is |
| AC-3.2 | Unpromote of a person's last member on the graph | The per-kind sentence of AC-2.1 (the S13-08 copy), danger tone; the graph stays where it is (working rule "One Unpromote") |

## 3. Scope

| In | Covered by |
| --- | --- |
| One Unpromote confirm for both pages: its item, its builder (connections + preflight) and its modifier, lifted from the Evidence graph (RC-2 – RC-5) | STEP-1, STEP-2, AC-3.1, AC-3.2 |
| The single-handle deletes message the user approved for S13-08 (RC-3) | STEP-1, AC-2.1, AC-3.2 |
| `PersonClaimRemoval`: begin from a claim (reads the claim's graph from the session), confirm, the person-gone check (RC-1, RC-3, RC-7) | STEP-3, AC-1.1 – AC-1.5, AC-2.1 – AC-2.6 |
| Remove on each claim card, disabled while writing; focus after a kept Remove (RC-1, RC-6) | STEP-4, AC-1.2, AC-1.5 |
| Landing on the Persons list (RC-7) | STEP-4, AC-2.4 |
| Strings | STEP-5 |
| Docs, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Any Go, proto or store change | not planned: S13-08's RPCs carry it |
| A Toast after Remove | not planned (brief RC-6), unless the board adds one |
| A general "deleted entity → list" rule for every Conclusion page (Event, Place, a page left open while another window unpromotes) | their redesign (R-1); this PR navigates only after its own write (D-8) |
| Bulk Remove, Reject, Provisional | not planned (brief § 2.1) |
| How the records connect | S13-28 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Unpromote item | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `PendingUnpromote{id, name, kind, bridges, deletedHandles, impactCheckFailed, error}`, `confirmTone` | Nested in the graph model; `confirmTone` is `.danger` when a handle is deleted, the check failed or there is an error |
| Unpromote begin | `EvidenceGraphModel.beginUnpromote(subjectID:)` | From the graph's snapshot: the bridges that have the Subject as an end and are promoted, sorted by id, worded by `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)`; then `store.getUnpromoteImpact`; a generation guard drops a stale open |
| Unpromote write | `EvidenceGraphModel.confirmUnpromote()` | `store.unpromoteSubject`, `session.apply(effects)`, closes; a failure stays as `error` |
| Unpromote confirm | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `.pvConfirm(item: $model.pendingUnpromote, …)` and `unpromoteDetail(_:)` | Message: `PVDeleteImpactCopy.cascadeLines` of the deleted group (first line), else the preflight-failed copy, else `unpromoteConfirmMessage`; the detail lists the bridges with `unpromoteConfirmBridgesHeader` / `Note` and the error Callout |
| The approved deletes copy | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref:)` | Defined by S13-08 (D-16) with person, event and place keys; `rg unpromoteConfirmMessageDeletes macos/App` finds no caller: the confirm shows the generic `cascadeLines` sentence instead |
| Store | `macos/App/Platform/GenealogyStore.swift` `unpromoteSubject(projectDir:userID:subjectID:) -> CatalogWriteEffects`, `getUnpromoteImpact(projectDir:subjectID:) -> CatalogDeleteImpactGroup?`; `CatalogDeleteImpactGroup{via, kind, total, listed: [CatalogDeleteImpactListed{id, ref, title}]}` | The preflight's listed rows carry each deleted handle's id |
| FakeStore | `macos/App/Platform/FakeStore.swift` `unpromoteImpactBySubject`, `getUnpromoteImpactError`, `getUnpromoteImpactCalls`, `unpromoteSubject` | Scripted |
| Effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects` | No "deleted" flag: a deleted handle shows only as a touched handle |
| Reading a session value | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `readyValue(_:)` | Waits out an in-flight load and returns the cached value, nil when never warmed, stale or failed; `CitationComposerModel` reads `.evidenceGraph`, `.propertiesWorkspace` with it |
| Snapshot | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphSnapshot.build(rows:types:rules:)` | Builds the graph's placed subjects and bridges from the three cached reads (`.evidenceGraph`, `.propertiesWorkspace` types, `.connectRules`) |
| Leaving a page | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `fallbackToSectionRoot()` | Replaces the current entry with the section's list; `SourcePageView`'s resource delete uses it after a successful delete |
| Error kind | `macos/App/Platform/CoreInvoke.swift` `CoreErrorKind.notFound`, `CoreInvokeError.coded(status:code:kind:params:)` | A missing handle's detail read throws `conclusiondetails.not_found` with kind `notFound` |
| Sources section | `PersonSourcesSection.swift` (in `Features/Conclusions/`, added by S13-16) | Draws each claim card with a footer graph control and no Remove (S13-16 D-10) |
| Tests | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges`, `unpromoteWithNoBridgesHasEmptyDetail`, `confirmUnpromoteRefreshesToPromote`, `unpromoteFailureStaysInConfirm`, `unpromoteLastRecordNamesTheDeletedHandle`, `unpromotePreflightFailureStaysInTheConfirm` | Cover the graph's flow; they keep passing on the shared item |

**Precedent.** The Source page's resource delete (`macos/App/Features/Sources/SourcePageView.swift`, `.pvDeleteImpact(... onConfirm: { if await model.confirmResourceDelete() { navigation.fallbackToSectionRoot() } })`): a write that removes the page's own entity reports success, and the view leaves with `fallbackToSectionRoot`. Copy that. `CitationComposerModel` reads the graph's three cached inputs with `session.readyValue`; copy that for the claim's graph.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One confirm. `PendingUnpromote` moves out of `EvidenceGraphModel` as `UnpromoteConfirmItem` *(new)* in `macos/App/Features/EvidenceGraph/UnpromoteConfirm.swift` *(new)*, same fields and `confirmTone`. `EvidenceGraphModel.pendingUnpromote` becomes `UnpromoteConfirmItem?` | User, 2026-11-05: the same confirm; R-38 "the same behavior" | A Person-page copy of the item and copy |
| D-2 | One builder: `UnpromoteConfirmItem.open(subject: EvidenceGraphPlacedSubject, in snapshot: EvidenceGraphSnapshot, orders: NameFormatOrders, store: any GenealogyStore, projectDir: String) async -> UnpromoteConfirmItem` *(new)*: the bridge sentences exactly as `beginUnpromote` builds them today, then the preflight (a group with `total > 0` → `deletedHandles`; a throw → `impactCheckFailed` and `error`). `beginUnpromote` keeps its guards and generation check and calls it | The graph and the page must list the same connections; one place | A second bridge query |
| D-3 | One modifier: `View.unpromoteConfirm(item: Binding<UnpromoteConfirmItem?>, isRunning: Bool, accessibilityIdentifierPrefix: String, onConfirm: @escaping () -> Void)` *(new)* in the same file, holding today's `pvConfirm` copy closure and `unpromoteDetail` view. `EvidenceGraphView` calls it with `"evidenceGraph.unpromote"`; the Sources section with `"persons.detail.sources.unpromote"` | The copy and the detail live once | Two `pvConfirm` call sites |
| D-4 | The file stays in `Features/EvidenceGraph/`: the confirm unpromotes a Subject of an Evidence graph, and its strings stay in `L10n.EvidenceGraph` (no key renamed) | It is the graph's domain; the Person page is a second caller in another feature. Not a kit recipe: Unpromote is Provenencia meaning with two callers in features | A `DesignSystem/Recipes/` piece (it would import `EvidenceBridgeEdgeSummary`, a feature type: the kit check fails) |
| D-5 | `PersonClaimRemoval` *(new)*, `@MainActor @Observable`, in `macos/App/Features/Conclusions/PersonClaimRemoval.swift` *(new)*: `init(entityId:session:store:userID:)`; `pending: UnpromoteConfirmItem?`; `isWriting`; `begin(_ claim: CatalogPersonClaim) async`: starts and awaits `.evidenceGraph(graphID)`, `.propertiesWorkspace` and `.connectRules` with `session.query` then `session.readyValue`, builds the snapshot with `EvidenceGraphSnapshot.build(rows:types:rules:)`, finds the claim's Subject in it, and calls `UnpromoteConfirmItem.open`; a generation counter drops a stale open (as the graph's). When the graph rows are nil (the read failed) or the Subject is not on the graph, it opens with no bridges and `bridgesUnlisted = true` *(new field on the item)* | Reads the claim's graph from the session cache (user, 2026-11-05); the precedent is `CitationComposerModel` | Calling the store for the graph directly (bypasses the cache and its invalidation) |
| D-6 | The item gains `handleID: String?` *(new)*, the Subject's own handle (the graph: `placed.membership?.entity.id`; the page: the Person's entity id). The message: a deleted group with `total == 1` whose one listed row's id is `handleID` → `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes(kind:ref: listed[0].ref)`; a larger group → `PVDeleteImpactCopy.cascadeLines(...)`'s first line (today's); the check failed → the preflight-failed copy; else `unpromoteConfirmMessage`. `bridgesUnlisted` adds `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted` *(new)* in the detail in place of the list | The single-handle sentence is the copy the user approved on 2026-11-04 (S13-08 D-16), which the landed confirm defines and does not call; the many-handle case keeps today's wording, as the brief says | New Person-page copy; leaving `unpromoteConfirmMessageDeletes` dead |
| D-7 | The person-gone check after a successful write: gone when `pending.deletedHandles?.listed` contains the person's entity id; when the check had failed, `PersonClaimRemoval` calls `store.getConclusionDetail(projectDir:entityID:)` once and treats a `CoreInvokeError.coded` with kind `.notFound` as gone; any other result or error is "kept". `confirm() async -> Outcome` *(new enum `Outcome { case kept, personDeleted, failed }`)* | The effects carry no deleted flag (§ 4); the preflight is the engine's own forecast, and a failed one needs one direct read | Waiting for the `.conclusionDetail` key to fail (indirect, depends on refresh timing); reading the detail after every write (a query per Remove) |
| D-8 | On `.personDeleted` the section calls `navigation.fallbackToSectionRoot()`, which replaces the Person page's history entry with the Persons list | The Source page delete precedent; Back does not return to a deleted page (brief RC-7) | `go(to: .sectionRoot(.persons))` (leaves the dead entry on Back) |
| D-9 | `PersonSourcesSection` owns one `PersonClaimRemoval` in `@State`, built on first appearance with the page's `entityId` and the session's user id (passed in by `PersonDetailSections` from `PersonDetailView.userID`; `PersonSourcesSection(entityId:userID:)`), reset when `entityId` changes. The card footer gains `PVIconButton(.unpromote, label: L10n.EvidenceGraph.unpromote, accessibilityLabel: L10n.Conclusions.sourcesRemoveAccessibility(name:), size: .sm)` at its end with help and label `L10n.EvidenceGraph.unpromote` / `L10n.Conclusions.sourcesRemoveAccessibility(name:)` *(new)*, id `persons.detail.sources.remove.<claimID>`, disabled while `PersonClaimRemoval.canBegin` *(new)* is false (a write, or a `begin` in flight) | The page's flow lives in the page's feature; the Remove control is the graph's glyph (brief SS-4) | A model on the section's content value |
| D-10 | After a kept Remove the section moves focus with `@FocusState` keyed by claim id: the next card's Remove in reading order, else the previous, else none (the section). **Refreshed from the board's RC-6 answer** | Brief RC-6 | Leaving focus on a removed control |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `UnpromoteConfirmItem` *(new)* | type | `EvidenceGraphModel.PendingUnpromote` | It is that type, moved so a second caller can hold it |
| `UnpromoteConfirmItem.open(...)` *(new)* | func | `EvidenceGraphModel.beginUnpromote` | Its body, lifted; the guards stay in the model |
| `View.unpromoteConfirm(...)` *(new)* | modifier | `EvidenceGraphView`'s `.pvConfirm` block and `unpromoteDetail` | Moved, not copied |
| `PersonClaimRemoval` *(new)* | model | `EvidenceGraphModel`'s Unpromote part; `SourcePageModel.confirmResourceDelete()` | The page has no graph model; the outcome tells the view to leave |
| `bridgesUnlisted` *(new field)*, `unpromoteConfirmBridgesUnlisted` *(new key)* | field, key | `impactCheckFailed`, `unpromoteConfirmPreflightFailed` | One more honest state of the same confirm |
| `sourcesRemoveAccessibility(name:)` *(new key)* | key | `L10n.EvidenceGraph.unpromote` | The page names the Subject in VoiceOver ("Unpromote J. Weber") |
| `PersonClaimRemovalTests` *(new)* | test file | `EvidenceGraphModelTests`' Unpromote cases | One file per model |

**Builds on.** S13-08's `getUnpromoteImpact` and `unpromoteSubject`; `EvidenceGraphSnapshot.build`; `session.readyValue`; `fallbackToSectionRoot`; `PVConfirm`; S13-16's card footer.
**Deletes.** `EvidenceGraphModel.PendingUnpromote` (moved); `EvidenceGraphView`'s inline `.pvConfirm` for Unpromote and `unpromoteDetail` (moved). `unpromoteConfirmMessageDeletes` gains its first caller, so nothing in it is dead after this PR.
**Next adds.** An Event or Place page with sections calls the same modifier with its own Remove; a family tree's "remove from person" (if ever) calls `PersonClaimRemoval`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Remove on a Person page → the Unpromote confirm | `UnpromoteConfirm.swift` | New caller; the item is built fresh per open, so `error` starts nil (as today) |
| A claim whose graph read fails, or whose Subject is gone from the graph | `PersonClaimRemoval.begin` | Opens with `bridgesUnlisted` (D-5); the write is the engine's, which re-checks |
| The person deleted while the page shows it | `PersonDetailView` | After this PR's own write: the Persons list (D-8). A delete from elsewhere still shows today's not-found Callout (Out) |
| Another Person page opened on Sources while a Remove is open | `PersonSourcesSection` | `PersonClaimRemoval` resets on `entityId` change (D-9); a stale `begin` is dropped by its generation |

### Scale

One Remove reads the claim's graph from the cache (warm when the researcher came from it; else one `.evidenceGraph` load, the cost of opening that graph), the properties workspace and the connect rules (both usually warm), one preflight and one write. The write's effects refresh `.personClaims` and the other per-person keys as Unpromote on the graph does today (S13-01 `refreshes`). No list or loop grows with the person's claims.

### Edges

- **Assumptions.** The preflight's `listed` names the subject's own handle when it is deleted (S13-08: `UnpromoteImpact` lists the primary first); D-7 checks membership by id, not position. The bridge list is informational: the engine computes the set it removes in the write (today's comment on `beginUnpromote`).
- **Neighbors.** A person with two claims: removing one keeps the person (AC-1.2); removing the last of the remaining one deletes (AC-2.1). A group of one row that is not the Subject's handle (an emptied association handle while the person keeps a member): D-6 compares the row's id with `handleID`, so it gets the `cascadeLines` line, not "this person is deleted" (T-2 pins it).
- **Removed inputs.** The claim already removed elsewhere (another window): the preflight or the write fails with the engine's error, shown in the confirm (AC-1.4); the reload drops the card.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** The bridge list is the Subject's filed bridges on one graph (today's bound on the graph's confirm).

### Schema

None — no catalog change.

### Go core

None — S13-08's RPCs.

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Remove | `PVIconButton` (`PVSymbol.unpromote`, `.sm`) | `PersonSourcesSection.swift` (S13-16's file) | D-9, RC-1 |
| Confirm | `PVConfirm` through `unpromoteConfirm` | `macos/App/Features/EvidenceGraph/UnpromoteConfirm.swift` *(new)* | D-1 – D-3, D-6, RC-2 – RC-5 |
| Deletes line | `PVDeleteImpactCopy`, `L10n.EvidenceGraph.unpromoteConfirmMessageDeletes` | as today | D-6, RC-3 |
| Landing | `WorkspaceNavigation.fallbackToSectionRoot()` | — | D-8, RC-7 |

States: confirm opening (Remove disabled), writing (`isRunning`), failed (error under the list), preflight failed (danger copy), graph unread (`bridgesUnlisted`). Accessibility: Remove reads "Unpromote J. Weber"; the confirm reads as today; after a kept Remove focus moves per D-10. The board's frames are described here in words on agreement.

### Strings

L10n keys *(new)*: `L10n.EvidenceGraph.unpromoteConfirmBridgesUnlisted` = "The connections that come off with it could not be listed. The other end of each stays promoted."; `L10n.Conclusions.sourcesRemoveAccessibility(name:)` = "Unpromote %@". Existing copy that names the action's scope and is now shown on a second page: `evidenceGraph.subject.unpromoteConfirmMessage` ("This promotion is removed. The cards stay on the graph, and the handle stays."), `unpromoteConfirmPreflightFailed.*`, `unpromoteConfirmMessageDeletes.*` — each still reads true from the Person page (the cards are on the claim's graph), so none changes. Removed: none.

## 7. Steps

**STEP-1 — Lift the Unpromote confirm**
Files: `macos/App/Features/EvidenceGraph/UnpromoteConfirm.swift` *(new)*, `EvidenceGraphModel.swift`, `EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`. Skill: `add-ui-component`.
Do: D-1 – D-4, and D-6's message rule.
Check: every `EvidenceGraphModelTests` Unpromote case passes; `unpromoteLastRecordNamesTheDeletedHandle` now expects the per-kind sentence (T-2).

**STEP-2 — The confirm can say the list is missing**
Files: `UnpromoteConfirm.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: `bridgesUnlisted` and its line (D-5, D-6).
Check: T-3 passes.

**STEP-3 — The Person page's removal**
Files: `macos/App/Features/Conclusions/PersonClaimRemoval.swift` *(new)*, `macos/ProvenenciaTests/PersonClaimRemovalTests.swift` *(new)*.
Do: D-5, D-7.
Check: T-4 – T-8 pass.

**STEP-4 — Remove on the claim cards**
Files: `PersonSourcesSection.swift` (S13-16's file in `Features/Conclusions/`), `macos/App/Features/Conclusions/PersonDetailView.swift`. Skill: `add-ui-component`, `add-workspace-location`.
Do: D-8 – D-10.
Check: T-9 passes; manual check items 1 – 4.

**STEP-5 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: `sourcesRemoveAccessibility(name:)`.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-13/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-13/design/S13-D6-sources-section.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` existing Unpromote cases | Unchanged behavior on the shared item | AC-3.1, D-1 – D-3 |
| T-2 | Swift | `EvidenceGraphModelTests.swift` `unpromoteLastRecordNamesTheDeletedHandle` (changed) and `unpromoteMessageForSeveralHandles` *(new)* | A group of one listed PER-… of kind `canonical_entity` for a person subject → "PER-… has no other records, so this person is deleted. The cards stay on the graph."; a group of three → the `cascadeLines` first line; a group of one association handle → the `cascadeLines` line | AC-2.1, AC-2.2, AC-3.2, D-6 |
| T-3 | Swift | `macos/ProvenenciaTests/PersonClaimRemovalTests.swift` `graphUnreadOpensWithoutTheList` *(new)* | `.evidenceGraph` load throws → item with `bridges` empty, `bridgesUnlisted` true; the preflight still read | D-5, D-6 |
| T-4 | Swift | same · `beginNamesTheClaimsConnections` *(new)* | FakeStore graph with J. Weber and two filed bridges → item title "J. Weber", two sentences equal to what `EvidenceGraphModel.beginUnpromote` builds for the same Subject | AC-1.1, D-2, D-5 |
| T-5 | Swift | same · `confirmKeepsThePerson` *(new)* | No deleted group; confirm → `unpromoteSubject` called once with CPR-2M8XF; outcome `.kept`; `pending` nil | AC-1.2, D-7 |
| T-6 | Swift | same · `confirmFailureStaysOpen` *(new)* | `unpromoteSubject` throws → outcome `.failed`; `pending.error` set; `isWriting` false | AC-1.4 |
| T-7 | Swift | same · `lastClaimDeletesThePerson` *(new)* | Deleted group lists PER-8JD2L (the entity id) → outcome `.personDeleted`; no detail read | AC-2.1, AC-2.4, D-7 |
| T-8 | Swift | same · `failedCheckReadsTheDetailOnce` *(new)* | `getUnpromoteImpactError` set; after the write `getConclusionDetail` throws `conclusiondetails.not_found` → `.personDeleted`; it returns a detail → `.kept`; one detail read each | AC-2.3, AC-2.5, AC-2.6, D-7 |
| T-9 | Swift | same · `staleBeginIsDropped` *(new)* | `begin` for claim A, then for claim B before A's preflight returns → `pending` is B's | D-5, D-9 |
| T-10 | Swift | same · `writingDisablesRemove` *(new)* | During `confirm()` (FakeStore write suspended) `isWriting` true and `canBegin` *(new computed)* false; after → true | AC-1.5, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `EvidenceGraphModelTests` (the lifted confirm).

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

1. Dogfood project, a Person with three claims, one of whose Subjects has a filed relationship. Sources → Remove on that card: the confirm names the relationship under "Also unpromoted"; Keep promotion: nothing changes (AC-1.1, AC-1.3).
2. Remove again and Unpromote: the card leaves, the page stays on Sources, focus is on the next Remove; open that Evidence graph: the Subject is unpromoted and its card stays (AC-1.2).
3. A Person with one claim: Remove shows the danger confirm "PER-… has no other records, so this person is deleted. …"; Unpromote: the Persons list, without the person; Back does not show the deleted page (AC-2.1, AC-2.4).
4. On an Evidence graph, Unpromote a person's last member: the same sentence; the graph stays (AC-3.2). Dark mode for the confirm.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Lifting the confirm changes the graph's behavior | T-1 runs every existing graph Unpromote test unchanged except the approved message (T-2) |
| The Persons list shows the deleted person for a moment | The write's effects refresh the list before the view leaves (`session.apply` runs before the outcome returns) |

**Stop and report if:**

- S13-16 has not landed, or its card footer has no slot for Remove (S13-16 D-10): this plan stays Blocked.
- The board is not agreed, or it moves Remove out of the card's footer (brief SS-4, RC-1).
- `GetUnpromoteImpact`'s listed rows do not carry the person's entity id for a last-member Unpromote (D-7).
- `EvidenceGraphSnapshot.build` cannot be called with only the three cached reads (D-5).
- The approved per-kind sentence is no longer what the user wants on the Evidence graph (D-6 changes that confirm's visible copy).

## 12. Open questions

1. **The board's answers** to RC-3 (the person's name beside its ref), RC-6 (Toast, focus) and any change to RC-1's placement. Recommended: the brief's recommendations; D-9, D-10 and D-6 are written for them.
2. **The Evidence graph's deletes sentence** (D-6): the landed confirm shows the generic DeleteImpact line, not the per-kind sentence approved for S13-08 on 2026-11-04. Recommended: show the approved sentence on both pages, as D-6 says.

## 13. PR

**Title:** S13-17 — Remove a claim
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-17-remove-a-claim.md`.
