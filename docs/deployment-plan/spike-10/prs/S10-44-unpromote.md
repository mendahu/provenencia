# S10-44 — Unpromote a filed card

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 10 — Take a filing back
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-44--unpromote)
**Brief:** [S10-D6](../design/S10-D6-evidence-graph.md) (EG-8…EG-10), board agreed 2026-10-14
**Board:** Frames 10–13 of the S10-D6 pass in the Claude Design export *Evidence graph canvas.dc.html*. §6 macOS describes each frame this PR ships. The implementer does not need the export.
**Depends on:** S10-D6, S10-20 (landed [#396](https://github.com/mendahu/provenencia/pull/396))
**Blocks:** —
**Branch:** `s10-44-unpromote`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-audit.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/catalog-session.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** working rule "Unpromote deletes the filing. It does not reject it.", [S10-D6](../design/S10-D6-evidence-graph.md) §2 rows "The footer is one height" through "Card delete is a different control" and EG-8…EG-10, [`conclusion-layer-data-model.md` §5 and §5.2](../../../conclusion-layer-data-model.md), dogfood bar item 9 in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `4ea7483`
**Review:** [`S10-44-unpromote.review.md`](S10-44-unpromote.review.md) — Fits with changes, Ready with changes (applied); 2 rounds

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

A researcher who filed a person, event, or place by mistake can take that filing back from the Evidence graph. A promoted primary card gets an Unpromote icon button in its footer. A confirm names the relationships, participations, locations, and place relationships on this Source that were filed through it. Confirm deletes this Subject's Identity Claim and those bridge claims in one audited write. The cards return to Promote. This is dogfood bar item 9 and the slice 10 Check.

**Why now.** S10-20 has landed on the same board (S10-D6), and the footer it left is the one this PR fills. The working rule and EG-8…EG-10 are agreed.

**What stays.** The Evidence cards, their Observations, and the edges stay. Every handle stays, including a handle left with no members and a bridge's association handle. Other members' claims, and their pins, stay as they are; the review of weak claims stays [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md). The handle row still opens the Person, Event, or Place page. Card delete is still the trash can and the impact confirm. Promote and Promote all are unchanged.

## 2. User stories

**US-1.** As a researcher, when a card is filed, I see Unpromote beside the handle row, so I can take the filing back without leaving the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A person card with an accepted claim on `PER-4M2QX`, name "James Robins" | The footer is 36pt. It paints, left to right: the ref, the name, the chevron, a 1pt hairline, then the Unpromote icon button (`arrow.uturn.down`) in a 36pt square at the trailing edge. The handle hit and the Unpromote hit do not intersect, and together with the 1pt hairline they cover the footer width. |
| AC-1.2 | A card whose handle name is "James Alexander Robins of Wellington North Township" | The name truncates before the chevron. The Unpromote square keeps its 36pt width. |
| AC-1.3 | An unpromoted person card, and a relationship card that is filed | The person card's Promote footer shows the `arrow.turn.right.up` icon before "Promote". Neither card shows Unpromote. Neither has an `unpromote` action target. Both footers are unchanged. |
| AC-1.4 | The pointer hovers the Unpromote square, then presses and releases it | Only the 22pt button fill changes (ink at 0.10, then 0.18). The handle row does not highlight. The canvas tooltip reads "Unpromote". Releasing opens the confirm and does not navigate. |
| AC-1.5 | VoiceOver on a promoted person card | The card's custom actions include "Unpromote", after the open-handle action. Running it opens the same confirm. An unpromoted card's actions do not include it. |

**US-2.** As a researcher, when I press Unpromote, the confirm tells me exactly which filings go, so I am not surprised by what else comes off.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The household fixture (D-17): James, John, Mary, and the 1881 census event are filed; son-of (James–John), participation (James in the event), and spouse (John–Mary) are filed | The confirm title is "Unpromote James Robins?". The message is "This filing is removed. The cards stay on the graph, and the handle stays." The detail has the header "Also unfiled", then the son-of and participation card sentences, in bridge-id order, then "The other end of each stays filed." The spouse relationship is not listed. Confirm reads "Unpromote", cancel reads "Keep filing". There is no trash icon. |
| AC-2.2 | James is filed and no filed bridge on this Source has James as an end | Title and message are as in AC-2.1. The detail is empty. |
| AC-2.3 | A relationship touching James exists on this Source but is not filed | It is not listed. |
| AC-2.4 | The confirm is open and the researcher presses "Keep filing" or Escape | The sheet closes. No store call is made. Every footer is as before. |

**US-3.** As a researcher, when I confirm, this filing and the filings that depended on it are removed in one step, and nothing else is.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The household fixture, and Unpromote James is confirmed | One revision, action `unpromote_subject`, deletes James's claim, the son-of claim, and the participation claim, with each claim's pins. James's card, the son-of card, and the participation card show Promote. John's card and the event card still show their handles. The spouse relationship stays filed. `PER-4M2QX` still exists with no members. Each association handle still exists. Every card and edge stays. |
| AC-3.2 | A birth-record Subject is on `PER-4M2QX`, and James joined it with a confirmed pair, so both claims pin both Observations | Only James's claim and its own pins are deleted. The birth-record claim stays accepted with both pins, including the pin on James's Observation. |
| AC-3.3 | `UnfileBridgesTx` is called with a Source id other than the bridge's own (the product cannot build a cross-Source bridge: `connect` refuses ends on another Source) | It returns no changes, and the son-of bridge stays filed. |
| AC-3.4 | The engine is asked to unpromote a Subject that has no accepted claim (a second click after the first landed) | No revision is written and no error is returned. |
| AC-3.5 | The engine is asked to unpromote a relationship Subject, or an id that is not a Subject | `promote.unsupported_type` for the relationship. `promote.invalid` for the unknown id. Nothing is written. |
| AC-3.6 | The write fails (the store throws) | The confirm stays up. Its detail shows a danger callout with `L10n.Errors.message(for:)` of that error, under any bridge list. Confirm is enabled again. "Keep filing" closes it with nothing changed. |
| AC-3.7 | After AC-3.2's write | The `canonical_entity` scope of `PER-4M2QX` moves to that revision. The Source's revision does not move. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Engine: `promote.Unpromote` deletes the accepted claim of one primary Subject, and `identityclaims.UnfileBridgesTx` the accepted claims of the bridges on its Source that have it as an end, with pins, audited (EG-10). | STEP-1, STEP-2, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.7 |
| FFI method `METHOD_UNPROMOTE_SUBJECT` and its handler, returning write effects. | STEP-3, AC-3.1, AC-3.5 |
| Swift store: `GenealogyStore.unpromoteSubject`, `GoStore`, `FakeStore`. | STEP-4, AC-3.1, AC-3.6 |
| EG-8. Promoted footer: handle row, hairline, Unpromote square; hits, hover, canvas tooltip. Not on bridge or unpromoted cards. | STEP-6, AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| EG-8. VoiceOver action. | STEP-7, AC-1.5 |
| The Promote and Unpromote icons (D-15): `.promote` on the Promote footer and on Promote all, `.unpromote` on the Unpromote button. | STEP-5, STEP-6, AC-1.1, AC-1.3 |
| EG-9. `PVConfirm` with an item, tone irreversible, title, message, bridge detail, labels, cancel. | STEP-7, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| EG-10. Confirm writes, the graph refreshes to Promote, errors stay in the confirm. | STEP-7, AC-3.1, AC-3.6 |
| Docs: §5.2 of the data model, `docs/catalog-deletes.md`, the `deleteimpact` README, map, completed, brief archive. | STEP-9 |

| Out | Home |
| --- | --- |
| Rejecting a claim ("not this entity"), the weak-claim queue, re-pin, accept-as-is. | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Deleting a handle, including one left with no members. | not planned — working rule "Unpromote deletes the filing" |
| Deleting or moving Evidence cards or their Observations. | not planned — EG-10 |
| Unpromote from the Person, Event, or Place page. | not planned — S10-D6 §2.1 |
| Unpromote on a bridge card. A bridge's claim comes off only with an end. | not planned — S10-D6 EG-8 |
| An undo, or a Toast after confirm. | not planned — S10-D6 §2.1 |
| Unpromote for a selection of cards. | not planned |
| Adding a `CatalogMutation` case, or changing `conclusionTriggers`. | not planned — D-9 |
| Changing the full-width hit of the Promote footer on unpromoted cards. | not planned |
| Changing `PVIconButton` to read `pvHostInteraction`. | not planned — D-12 |
| Changing `arrowUpRight` at its other callers (`PVDeleteImpact`, `ReconciledValueRow`), or the Promote place's own icons. | not planned — D-15 |
| Help text on any canvas target other than Unpromote. | not planned — D-13 makes the slot |
| Teaching `FakeStore.promoteSubject` to file bridges. | not planned — D-17 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Claims | `core/database/migrations/000032.sql` `identity_claims` | `status IN ('provisional','accepted','rejected')`. At most one accepted claim per Subject (`identity_claims_one_accepted_per_subject`). Pins in `identity_claim_evidence`, ON DELETE CASCADE as a backstop. Nothing in `core/` writes `provisional` or `rejected`. |
| Claim package | `core/database/identityclaims/identityclaims.go` `InsertTx`, `AcceptedEntityForSubjectTx` | Inserts and reads claims. `AcceptedEntityForSubjectTx(tx, subjectID)` returns the accepted `Claim` (with `ID`), or `sql.ErrNoRows`. No delete. |
| Claim delete | `core/database/deleteimpact/facets.go` `releaseSubjectClaims` | The only claim delete. Loops over every claim of a Subject: `ReleaseFacets(tx, catalogmodel.KindIdentityClaim, id)` for the pins, `DELETE FROM identity_claims`, then a `rowchange.DeletedRow` change carrying `id`, `subject_id`, `entity_id`, `subject_type_id`, `status`, `confidence_grade_id`, `argument`. Used only by Subject delete. The per-claim body is inline. `deleteimpact` imports no database package. |
| Bridges | `core/database/identityclaims/bridgefile.go` `bridgesTouching`, `FileBridgesTx`, `fileBridge`, `resolveEndpoint` | `bridgesTouching` selects bridge Subjects (`connectrules.BridgeTypeKeys`) where `s.id = ?` OR an Observation on it has `value_subject_id = ?`, on any Source. `fileBridge` files a bridge once each end property resolves to one accepted handle; it reads every end Observation, with no polarity filter. |
| Promote | `core/database/promote/promote.go` `Save`, `PrimaryKind`, `ErrInvalid`, `ErrUnsupportedType` | Files one primary Subject in a `promote_subject` revision and calls `identityclaims.AppendFiling` to file touching bridges. |
| Scopes | `core/database/effects/registry.go` `identity_claims`, `identity_claim_evidence` | A claim change resolves `HandleScope: field("entity_id")` and `Source` (graph invalidation) from `subject_id`. A delete resolves through the ghost row, so the `DeletedRow` must carry `subject_id`, `entity_id`, `status`. No `source` audit scope for conclusion work. |
| Write runner | `core/writes/run.go` `Run` | One transaction, audit, `autoreconciler.RecomputeReported` on `set.Handles`, search reprojection, commit, notify. An empty change list rolls back and returns no error. |
| Go fixtures | `core/database/promote/bridges_test.go` `newBridgeWorld`, `bridgeWorld.node`, `relationship`, `participation`, `promote`, `join`, `handles`, `members`; `core/database/audit/scopes_test.go` `newScopeSeed`, `scopeSeed.promote`, `handleRevision`, `sourceRevisions`, `latest` | `bridgeWorld` (package `promote_test`) builds and files bridges on one Source. `scopeSeed` has two Sources and no bridges. |
| FFI | `api/proto/engine.proto` `Method`, `DeleteSubjectResponse`, `WriteEffects` | Last method `METHOD_GET_NAME_FORMAT = 99`. `METHOD_PROMOTE_SUBJECT = 85`. |
| FFI dispatch | `api/ffi/dispatch.go` `MethodPromoteSubject`, `MethodDeleteSubject` | Constants and one `case` each. |
| Handlers | `api/ffi/handlers/promote.go` `PromoteSubject`; `api/ffi/handlers/subjects.go` `DeleteSubject` | `DeleteSubject` is the template: `withProjectCatalog`, `writes.Run`, `Effects: writeEffectsProto(wrote)`. |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `deleteSubject`, `promoteSubject`; `macos/App/Platform/GoStore.swift` `deleteSubject`; `macos/App/Platform/CoreInvoke.swift` `CoreMethod.deleteSubject`; `macos/App/Platform/FakeStore.swift` `deleteSubject`, `membershipBySubject`, `claimBySubject`, `observationsBySource`, `setSubjectPositionError` | No unpromote call. `FakeStore.promoteSubject` writes only `membershipBySubject` and `claimBySubject`; it never files bridges. Failure injection is a `var …Error: Error?` per call. |
| Card paint | `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift` `subject`, `paintFooter` | The live card is an AppKit `NSView`. `subject` records the footer region (`promoteActionID` or `openHandleActionID`) as full card width × 36, in two duplicated blocks (cited and uncited). The promoted branch of `paintFooter` paints the ref at `shellPaddingX`, width 72, the name at `shellPaddingX + 78`, width `rect.width - 120`, the chevron at `rect.maxX - 27`, and fills the whole footer on hover. `Painter.fillRound(_:radius:_:)` exists. |
| Canvas pointer | `macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift` `hitTest`; `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift` `updateHoveredAction`, `hoveredCardAction`; `macos/App/Features/GraphCanvas/GraphCanvasHitTarget.swift` `GraphCanvasActionTarget` | The document view returns itself from `hitTest`, so card views never see the mouse. The controller tracks the hovered action. `GraphCanvasActionTarget` has `id` and `frame` only. No canvas target has help text. |
| SwiftUI twin | `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` `membershipRow`, `cardFooter`, `promoteFooter`, `footerHeight`, `openHandleActionID` | Paint-only twin, used by `ghost` and by layout tests. `membershipRow` is a full-width row tagged `openHandleActionID`. |
| Hits | `EvidenceSubjectCard.actionTargets`, `measuredActionTargets`, `footerFrame`; `macos/App/Features/EvidenceGraph/EvidenceCardLayout.swift` `band`, `centeredHit` | The footer target is the full-width `band`, or `footerFrame` without a layout. `centeredHit(id, origin:, minSize: 0)` returns the painted rect in document space. |
| Dispatch | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `performCardAction`, `openHandle`, `beginDelete`, `confirmPendingImpact`, `CreateDraft` | `promoteActionID` navigates; `openHandleActionID` opens the page; delete starts `DeleteImpactFlow`. Line 13, "Snapshot of the subject a v1 Promote confirm names (`.pvConfirm(item:)`).", sits on `let sourceID` and describes nothing that exists. |
| VoiceOver | `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift` private `accessibilityActions(for:)` and the SwiftUI `.accessibilityAction` chain | Footer action is promote or open handle, then delete. Neither list is reachable from a test. |
| Confirm | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` `pvConfirm(item:copy:tone:isRunning:…)`, `PVConfirmTone.irreversible` | Cancel clears the item. The caller dismisses after an async confirm; an error goes in `detail`. Precedent: `macos/App/Features/Names/NameValueEditorForm.swift` `pendingSplit`, `splitConfirmDetail`. |
| Error copy | `L10n.Errors.message(for:)`; `macos/App/DesignSystem/Recipes/DeleteImpact/DeleteImpactFlow.swift` | The app's one mapping from a thrown error to copy. |
| Bridge sentence | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` `sentence(for:in:orders:)` | The card sentence. `beginDelete` already uses it to name a bridge. |
| Snapshot | `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `SourceGraphPlacedSubject.membership`, `SourceGraphPlacedBridge.endpointAID`, `endpointBID`, `membership`, `citedEndpoints(kind:observations:rules:)`, `SourceGraphRows` | Bridge ends come from the first Observation per end property of the connect rule. `SourceGraphRows` has `memberships`. |
| Will-file forecast | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.BridgeFact`, built in `macos/App/Features/Promote/PromoteModel.swift` | Promote forecasts the bridges it will file from the same snapshot fields (`endA`, `endB`, `alreadyFiled`). |
| Refresh | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for:project:)`, `conclusionTriggers`; `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)` | Effects invalidate the Source graph per `effects.sources` and the lists and pages per `effects.touchedHandles`; `apply` refreshes graph progress per `effects.sources`. `conclusionTriggers` is only for writes that report no effects. |
| Graph view | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `.pvDeleteImpact` | The confirm modifiers hang off the view body. |
| Symbols | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol`; `macos/ProvenenciaTests/PVSymbolTests.swift` `symbolResolves` | `arrowUpRight = "arrow.up.right"` ("jump out of a blocker row") is drawn before "Promote" in `EvidenceAppKitCard.paintFooter`, in the twin's `promoteFooter`, and on `EvidenceGraphView.promoteAllButton`; it is also used by `PVDeleteImpact` and `ReconciledValueRow`. No Promote or Unpromote case. Deployment target macOS 14.0. |
| Copy | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph` "Promote + membership footer"; `macos/App/Resources/Localizable.xcstrings` | `evidenceGraph.subject.promote` and siblings. `LocalizedStringResource` conforms to `PVCopy`. |
| Tests | `core/database/promote/promote_test.go` `TestSave`; `core/database/promote/bridges_test.go` `TestBridgeFiling`; `core/database/audit/scopes_test.go` `TestHandleScopes`; `api/ffi/handlers/promote_test.go` `TestPromoteSubject`; `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift` `placed(cited:promoted:kind:)`, `membershipSwapsPromoteForOpenHandle`, `footerTargetIsFullWidthAtTheBottom`; `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `measuredLayoutMatchesThePaint`, `cases`, `fixture`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `graphRows`, `promotableModel`, `waitUntil`; `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift`; `macos/ProvenenciaTests/PVSymbolTests.swift` `symbolResolves`; `macos/ProvenenciaTests/L10nResolutionTests.swift` | `measuredLayoutMatchesThePaint` asserts the footer region is 36pt high and flush with the bottom, and that bands do not overlap. `footerTargetIsFullWidthAtTheBottom` covers unpromoted cards only. `graphRows` takes no memberships. |

**Precedent.** Subject delete (`DeleteSubject` handler, `subjects.Delete`, `releaseSubjectClaims`) is the shape of an audited, multi-row conclusion delete in one `writes.Run`. Copy its handler and its claim-release body; do not copy its Impact check or its manual handle recompute (`writes.Run` recomputes from the returned changes, as `promote.Save` relies on). `promote.Save` + `identityclaims.AppendFiling` is the shape `Unpromote` + `UnfileBridgesTx` mirrors. On the Mac, copy `confirmPendingImpact` for the store call plus `session.apply(effects, …)`, and `NameValueEditorForm` for an irreversible `pvConfirm` with a listed detail. The Promote footer's paint-only kit button in `promoteFooter` is the precedent for the twin.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The writer is `promote.Unpromote(tx *database.Tx, userID, subjectID []byte) ([]rowchange.Change, error)` *(new)* in `core/database/promote/unpromote.go` *(new)*. Action `unpromote_subject`. | `promote` owns filing a primary and already has `PrimaryKind`, `ErrInvalid`, `ErrUnsupportedType`. Unpromote is the inverse of `Save`. | A new package. Putting it in `identityclaims`, which does not know primaries. |
| D-2 | The claim delete is `deleteimpact.ReleaseClaim(tx *sql.Tx, claimID []byte) (Released, error)` *(new)*, extracted from the loop body of `releaseSubjectClaims`, which then selects the Subject's claim ids and calls it for each. It reads the row by id, releases pins through `ReleaseFacets(tx, catalogmodel.KindIdentityClaim, claimID)`, deletes the row, and returns the pin changes followed by the claim's `DeletedRow` change. | One definition of "remove a claim, audited, pins first". `deleteimpact` imports no database package, so `identityclaims` and `promote` can import it. The package's exports carry no `Tx` suffix (`ReleaseFacets`). | A second copy of the `DeletedRow` field list. A raw `DELETE` that relies on the CASCADE — the data model §5 forbids it. |
| D-3 | The bridge side is `identityclaims.UnfileBridgesTx(tx *sql.Tx, sourceID, subjectID []byte) ([]rowchange.Change, error)` *(new)* in `bridgefile.go`, the inverse of `FileBridgesTx`. It selects bridge Subjects whose type is in `connectrules.BridgeTypeKeys()` (origin provenencia), with `s.source_id = sourceID`, an accepted claim, and `touchesEnd`; ordered by bridge Subject id; then calls `deleteimpact.ReleaseClaim` on each claim and returns the concatenated changes. `touchesEnd` *(new, unexported const)* is only the EXISTS arm, with one `?`: `EXISTS (SELECT 1 FROM observations o WHERE o.subject_id = s.id AND o.value_subject_id = ?)`. `bridgesTouching` keeps its own `s.id = ? OR` and uses `touchesEnd` for the second arm; its argument list is unchanged. No polarity filter. | File and unfile live in one file. The end rule is the one filing read (`bridgesTouching`, `resolveEndpoint`), so what Unpromote takes off is what filing put on. | An exported query plus a loop in `promote` — one more exported type for one caller. Exporting `bridgesTouching` — no Source filter, and it returns unfiled bridges. A polarity filter — it would leave filed a bridge whose filing used this Subject. |
| D-4 | `Unpromote` order: `database.RequireUserID(userID, ErrInvalid)`; `len(subjectID) != 16` → `ErrInvalid`; `subjects.GetTx` (`sql.ErrNoRows` → `ErrInvalid`); `subjecttypes.GetByIDTx` + `PrimaryKind` (false → `ErrUnsupportedType`); `identityclaims.AcceptedEntityForSubjectTx` — `sql.ErrNoRows` → return `nil, nil` (AC-3.4). Otherwise `deleteimpact.ReleaseClaim` on that claim's `ID`, then append `identityclaims.UnfileBridgesTx(tx.Tx, subject.SourceID, subject.ID)`. | Mirrors `Save`'s validation. Idempotent for a stale second click, because `writes.Run` rolls back an empty change list with no error. | A new `not_member` error code with L10n — a red callout for a state the refresh already fixes. |
| D-5 | Only accepted claims are deleted, on this Subject and on the bridges. Other claims on the handle, their pins (including pins on this Subject's Observations), and every handle row stay. No `autoreconciler` or `conclusionheaders` call in `Unpromote`. | Working rule and §5.2. `writes.Run` recomputes `set.Handles` (the entity and, for an accepted claim, its observers) and header deps. | Recomputing by hand like `subjects.Delete`. |
| D-6 | FFI: `METHOD_UNPROMOTE_SUBJECT = 100`, `UnpromoteSubjectRequest { string project_dir = 1; string user_id = 2; string subject_id = 3; }`, `UnpromoteSubjectResponse { WriteEffects effects = 1; }`. Handler `UnpromoteSubject` *(new)* in `api/ffi/handlers/promote.go`, after `PromoteSubject`. The engine reads the Source from the Subject. | Mirrors `DeleteSubject`. One source of truth for the Source. | Returning the unfiled ids — the client refreshes. |
| D-7 | The confirm lists bridges from the client snapshot: `snapshot.bridges` where `endpointAID == id || endpointBID == id` and `membership != nil`, sorted by bridge id, each as `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)`. Built once when the confirm opens. | The inverse of Promote's `PromoteFlow.BridgeFact` forecast, from the same snapshot fields. The engine recomputes the authoritative set in the write (D-3); both sort by bridge Subject id. | A preview RPC — a second method and a round trip for what the snapshot already holds. |
| D-8 | `EvidenceGraphModel.PendingUnpromote: Identifiable` *(new, nested)*: `id` (subject id), `name` (`placed.displayName(orders: session.nameFormat)`), `bridges: [String]`, `var error: String?`. Model state *(new)*: `var pendingUnpromote: PendingUnpromote?` and `var isUnpromoting = false`. `beginUnpromote(subjectID:)` *(new)* builds a fresh item for a promoted primary and does nothing otherwise. `confirmUnpromote() async` *(new)*: guard an item and not running; set running; call the store; on success `session.apply(effects)` and set `pendingUnpromote = nil`; on throw set `pendingUnpromote?.error = L10n.Errors.message(for: error)`; clear running in both. Delete line 13's stale comment; `pendingUnpromote` gets the doc comment "Subject whose filing the Unpromote confirm names (`.pvConfirm(item:)`)."; `sourceID` gets no comment. | Nested pending items are the model convention (`CreateDraft`, `PromoteModel.PendingLeave`). A fresh item has no error, so no clearing rule. `L10n.Errors.message(for:)` is the app's error-to-copy path. | A Bool binding. Reusing `DeleteImpactFlow` — the brief says not DeleteImpact. `localizedDescription`. |
| D-9 | Refresh is `session.apply(effects)` with no mutation. | The response's effects carry the Source (graph and progress) and the touched handles (lists and pages). `conclusionTriggers` is for writes that report no effects. | `.promotedSubject(sourceId:)` — it reloads every Conclusion page. A new `CatalogMutation` case. |
| D-10 | One geometry function: `EvidenceSubjectCard.promotedFooterParts(footer: CGRect) -> (handle: CGRect, hairline: CGRect, slot: CGRect)` *(new)*, with `unpromoteSlotWidth: CGFloat = 36` *(new)*. For footer `f`: `slot = (f.maxX - 36, f.minY, 36, f.height)`; `hairline = (slot.minX - 1, f.minY + 8, 1, f.height - 16)`; `handle = (f.minX, f.minY, hairline.minX - f.minX, f.height)`. In `EvidenceAppKitCard.subject`, the two duplicated footer blocks become one private helper `recordFooter(placed: SourceGraphPlacedSubject, y: inout CGFloat, width: CGFloat, regions: inout [String: CGRect]) -> CGRect` *(new)*: it builds `f = (0, y, width, footerHeight)`, advances `y` by `footerHeight`, returns `f` as `footerRect`, and records `promoteActionID` → `f` for an unpromoted card, and `openHandleActionID` → `handle`, `unpromoteActionID` *(new, `"unpromote"`)* → `slot` for a promoted one. `paintFooter`'s promoted branch calls `promotedFooterParts(footer: rect)`: hover fill over `handle` only; ref at x `shellPaddingX`, width 72 (unchanged); name at x `shellPaddingX + 78`, width `handle.width - 120`; chevron at `handle.maxX - 27`; hairline filled `style.line`; button `painter.fillRound(22×22 centered in slot, radius: PVRadius.sm, …)` with `style.ink.opacity(0.10)` on hover and `0.18` on press of `unpromoteActionID`, nothing at rest; `painter.symbol(PVSymbol.unpromote.rawValue, in: 13×13 centered in slot, style.ink)`. | EG-8: own hit, no overlap, 36pt footer, name truncates and the button does not. One function feeds paint, regions, and hits, so they cannot drift. A 36pt square is a full-height hit centered on what is drawn (working rule "Hit boxes are the painted controls"). | Constants repeated in `paintFooter`, `actionTargets`, and the tooltip. A hosted SwiftUI `PVIconButton` on the live card — the card is painted in AppKit. A 22pt hit — smaller than the 28pt header hits. |
| D-11 | Hits for a promoted card: `measuredActionTargets` appends `layout.centeredHit(openHandleActionID, origin:, minSize: 0)` then `layout.centeredHit(unpromoteActionID, origin:, minSize: 0)`. The layout-free branch of `actionTargets` applies `promotedFooterParts(footer: footerFrame(cardFrame: frame))` and appends `handle` and `slot`. Unpromoted cards keep the full-width band. Bridge cards are not touched. | `band` is full width by definition and would overlap. `centeredHit(…, minSize: 0)` is the painted rect. | A new `EvidenceCardLayout` helper. |
| D-12 | The twin's `membershipRow` becomes an `HStack(spacing: 0)` of the existing handle row content (`frame(maxWidth: .infinity)`, tagged `openHandleActionID`, background and hover as today), a `Rectangle().fill(style.line).frame(width: 1, height: 20)`, and a slot `PVIconButton(.unpromote, label: L10n.EvidenceGraph.unpromote, size: .sm, action: {})` with `.allowsHitTesting(false)`, framed `unpromoteSlotWidth × footerHeight`, tagged `unpromoteActionID`, `.accessibilityHidden(true)`. The `HStack` takes `.background(style.band)` and the existing 36pt frame. No hover on the twin's button. | Layout tests measure the twin; its regions must equal `promotedFooterParts`. `PVIconButton` does not read `pvHostInteraction`. | Teaching `PVIconButton` host interaction — Out. |
| D-13 | Tooltip: `GraphCanvasActionTarget` gains `var help: String? = nil` *(new)*. `GraphCanvasPointerController` exposes `var hoveredHelp: String?` *(new, computed)*: the `help` of the target action named by `hoveredCardAction` in `hitTargets`, nil when `hoveredCardAction` is nil, so every path that clears hover clears it. `GraphCanvasDocumentView` sets `toolTip = pointer.hoveredHelp` after each pointer move and on exit, only when the value changed. The `unpromote` target is built with `help: L10n.string(L10n.EvidenceGraph.unpromote)` in both branches of D-11. `GraphCanvasActionTarget.mapFrame(_ transform: (CGRect) -> CGRect) -> GraphCanvasActionTarget` *(new)* returns a copy with only `frame` transformed; `EvidenceGraphView.publishHitTargets()` uses it at both sites that today rebuild `GraphCanvasActionTarget(id: action.id, frame: inDocument(action.frame))`, so `help` reaches `pointer.hitTargets`. | The document view owns the mouse (`hitTest` returns itself), so a tool-tip rect on a card view would never fire. A help field on the generic target gives every later canvas control a slot. | `addToolTip` on `EvidenceAppKitCard` — the card never sees the mouse. |
| D-14 | VoiceOver: the ordered `named` list inside `EvidenceCanvasCardContent.accessibilityActions(for:)` is lifted, unchanged, into `func namedActions(for card: EvidenceCanvasCard) -> [(name: String, run: () -> Void)]` *(new, internal)* on `EvidenceCanvasCardContent`. For a subject card whose `placed.membership != nil` it adds `(L10n.string(L10n.EvidenceGraph.unpromote), { model.beginUnpromote(subjectID: placed.id) })` immediately after the footer action. `accessibilityActions(for:)` maps `namedActions(for:)` to `NSAccessibilityCustomAction` as today. The SwiftUI `.accessibilityAction` chain (the paint-only twin) gets the same Unpromote action by hand after its footer action, guarded by `placed.membership != nil`; the chain is otherwise unchanged. | One list feeds the live card's VoiceOver actions and the test. The card element already names the subject, so the action is just "Unpromote" (EG-8). | A separate names-only helper — a third copy of the order that drives nothing. Replacing the SwiftUI chain with a `ForEach` over the list — wider than this PR. "Unpromote James Robins" — a second format key. |
| D-15 | Two `PVSymbol` cases *(new)*, named for the action, as opposites (product call, answered 2026-10-16): `promote = "arrow.turn.right.up"`, doc comment "Promote: file a Subject on a handle.", and `unpromote = "arrow.uturn.down"`, doc comment "Unpromote: take a filing back (S10-44)." The Unpromote button draws `.unpromote`. The Promote footer (`paintFooter`'s unpromoted branch and the twin's `promoteFooter`) and `promoteAllButton` switch from `.arrowUpRight` to `.promote`. `arrowUpRight` stays for its other callers. Both new cases join `PVSymbolTests.symbolResolves`. | The user chose the pair so the two actions read as opposites; one action, one icon, wherever Promote appears on the graph. | A Lucide-style unlink glyph (the brief's first draft) — replaced by the pair. `arrow.trianglehead.pull` / `arrow.trianglehead.branch` — macOS 15 only; the target is 14.0. Reusing `arrowUpRight` for Promote — it means "jump out" elsewhere. Changing `gitBranch = "arrow.triangle.branch"` — a different symbol with its own use. |
| D-16 | Copy, in `L10n.EvidenceGraph` under the promote footer mark: `unpromote` "Unpromote" (help, VoiceOver action, confirm button); `unpromoteKeep` "Keep filing"; `unpromoteConfirmTitle(name:)` "Unpromote %1$@?"; `unpromoteConfirmMessage` "This filing is removed. The cards stay on the graph, and the handle stays."; `unpromoteConfirmBridgesHeader` "Also unfiled"; `unpromoteConfirmBridgesNote` "The other end of each stays filed." | EG-9: the title names the subject; the message says the filing is removed and what stays, so it is said even when the detail is empty; the detail lists bridges and says the other end stays filed. | "Cards stay, handle stays" in the detail — EG-9 needs the detail empty without bridges. |
| D-17 | Swift fixtures. `FakeStore.unpromoteSubject` *(new)*: records `"unpromoteSubject id=…"`, calls `markCatalogSessionHeld`, throws `unpromoteSubjectError` *(new)* when set; otherwise removes `membershipBySubject[subjectID]` and `claimBySubject[subjectID]` and the same for every Subject on that Subject's Source that has a membership and an Observation in `observationsBySource[sourceID]` with `valueSubjectID == subjectID` (the engine's rule); returns `CatalogWriteEffects` with `sources = [sourceID]` and `touchedHandles` = the removed memberships' entity ids. In `EvidenceGraphModelTests`, `graphRows` gains `memberships: [CatalogSubjectMembership] = []`, and `householdModel()` *(new, private)* seeds James, John, Mary, an event, son-of (`relationship`, `person` = James, `related_to` = John), participation (`person` = James, `event` = event), and spouse (`relationship`, John–Mary), with positions, the end Observations in `store.observationsBySource`, memberships set directly in `store.membershipBySubject` for every Subject, and the same rows in the graph query value. Connect rules are the fake's default `CatalogConnectRule.productMatrix`. Bridge ids are `"b1-sonof"`, `"b2-part"`, `"b3-spouse"`, so son-of sorts first. A variant argument `fileSonOf: Bool = true` leaves the son-of bridge without a membership for AC-2.3. | A fake that files bridges is Out; setting memberships directly is how `SubjectStoreTests` drives the fake. The fake's end rule matches the engine's, and the fixture's Observations also give the snapshot its `endpointAID`/`endpointBID`. | Teaching `FakeStore.promoteSubject` to file bridges. |
| D-18 | Go fixtures. T-1 uses `newBridgeWorld` with its `node`, `relationship`, `participation`, `promote`, `handles`, and `members` helpers. T-1 lives in package `promote_test` and reuses `newBridgeWorld`. T-1 (d) calls `identityclaims.UnfileBridgesTx` inside `writes.Call` with a freshly minted 16-byte id as `sourceID`, because `connect.CreateCitedBridge` refuses a bridge whose ends are on another Source (`core/database/connect/connect.go`), so the Source filter is a guard, not a reachable product state. T-2 copies the first subtest of `TestSavePins` (`newPinFixture`, `person1`, `promote.Save` with `Pairs`, `pinned`): the joining Subject is James. T-3 is a `TestHandleScopes` case on `scopeSeed`: `s.promote(t, s.person)`, then `writes.Call` with `promote.Unpromote`; assert `s.handleRevision(t, h) == s.latest(t)` and `s.sourceRevisions(t)` unchanged. | Reuse the bridge world that already files bridges; the scope seed has the revision readers. Association scopes come from the same registry entry, so the person handle is the case. | A new fixture package. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `promote.Unpromote` *(new)* | func | `promote.Save`, `subjects.Delete` | `Save` creates; `Delete` removes the Subject. This is the inverse of `Save`, beside it. |
| `deleteimpact.ReleaseClaim` *(new)* | func | `releaseSubjectClaims` | Extracted from it; it calls this. No copy. |
| `identityclaims.UnfileBridgesTx`, `touchesEnd` *(new)* | func / SQL fragment | `FileBridgesTx`, `bridgesTouching` | The inverse verb in the file that owns filing; `touchesEnd` is shared with `bridgesTouching`. |
| `METHOD_UNPROMOTE_SUBJECT`, `UnpromoteSubjectRequest`, `UnpromoteSubjectResponse` *(new)* | FFI | `METHOD_DELETE_SUBJECT` | A different write; same message shape. |
| `handlers.UnpromoteSubject` *(new)* | func | `handlers.DeleteSubject` | Same shape, in `promote.go`. |
| `GenealogyStore.unpromoteSubject`, `GoStore.unpromoteSubject`, `FakeStore.unpromoteSubject`, `FakeStore.unpromoteSubjectError`, `CoreMethod.unpromoteSubject` *(new)* | func / var | `deleteSubject` on each | One store method per engine method. |
| `EvidenceGraphModel.PendingUnpromote`, `pendingUnpromote`, `isUnpromoting`, `beginUnpromote`, `confirmUnpromote` *(new)* | type / state / func | `DeleteImpactFlow`, `NameValueEditorForm.pendingSplit`, `EvidenceGraphModel.CreateDraft` | DeleteImpact is ruled out by the brief; the item snapshot follows `pendingSplit`; nesting follows `CreateDraft`. |
| `EvidenceSubjectCard.unpromoteActionID`, `unpromoteSlotWidth`, `promotedFooterParts(footer:)`, `EvidenceCanvasCardContent.namedActions(for:)` *(new)* | constants / func | `openHandleActionID`, `footerFrame(cardFrame:)`, the `named` list in `accessibilityActions(for:)` | New region on the same footer; one geometry source; the action list lifted, not copied. |
| `EvidenceAppKitCard.recordFooter` *(new, private)* | func | the two duplicated footer blocks in `subject` | Replaces both. |
| `GraphCanvasActionTarget.help`, `GraphCanvasActionTarget.mapFrame(_:)`, `GraphCanvasPointerController.hoveredHelp` *(new)* | field / func | the inline rebuild in `publishHitTargets` | The canvas has no help slot today; `mapFrame` keeps every field through the document-space mapping. |
| `PVSymbol.promote`, `PVSymbol.unpromote` *(new)* | key | `PVSymbol.arrowUpRight`, `PVSymbol.gitBranch` | Named for the action, so Promote and Unpromote stay opposites; `arrowUpRight` keeps its "jump out" meaning. |
| `L10n.EvidenceGraph.unpromote…` *(new)* | keys | `L10n.EvidenceGraph.promote…` | New copy in the same section. |
| `householdModel()`, `graphRows(memberships:)` *(new)* | test helper | `promotableModel()`, `graphRows` | The household needs bridges and memberships. |

**Builds on.** `writes.Run` and the effects registry for audit, invalidation, and handle recompute (`add-audit-scope`); `deleteimpact.ReleaseFacets` for pins (`add-catalog-delete`); the `DeleteSubject` handler template (`add-ffi-handler`); `pvConfirm(item:)` (`add-ui-component`); the region-tagged card layout from S10-01; `PromoteFlow.BridgeFact` for the client forecast.
**Deletes.** The stale comment on `EvidenceGraphModel.sourceID`. The inline per-claim body in `releaseSubjectClaims`. The duplicated footer-region blocks in `EvidenceAppKitCard.subject`. The magic `rect.width - 120` / `rect.maxX - 27` offsets now read from the handle rect.
**Next adds.** The weak-claim review (`docs/ideas/identity-claim-review.md`) calls `deleteimpact.ReleaseClaim` for a delete and adds a reject beside it. Any canvas control gets a tooltip by setting `help`. Unpromote from a handle page, if ever planned, calls the same store method.

### Schema

None — no catalog change. Claims and pins are deleted with the existing tables. Audit uses the existing `identity_claim` and `identity_claim_evidence` entity types.

### Go core

| Package | Change |
| --- | --- |
| `core/database/deleteimpact` | `func ReleaseClaim(tx *sql.Tx, claimID []byte) (Released, error)` *(new)* in `facets.go`. `releaseSubjectClaims` selects ids, then calls it for each. Subject delete behavior is unchanged. |
| `core/database/identityclaims` | `func UnfileBridgesTx(tx *sql.Tx, sourceID, subjectID []byte) ([]rowchange.Change, error)` *(new)* and `const touchesEnd` *(new)* in `bridgefile.go`; `bridgesTouching` uses `touchesEnd`. |
| `core/database/promote` | `func Unpromote(tx *database.Tx, userID, subjectID []byte) ([]rowchange.Change, error)` *(new)* in `unpromote.go` *(new)*, per D-4. |

```sql
-- sketch: UnfileBridgesTx
SELECT s.id, ic.id FROM subjects s
JOIN subject_types st ON st.id = s.subject_type_id AND st.origin = 'provenencia'
JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
WHERE s.source_id = ? AND st.key IN (…) AND <touchesEnd>
ORDER BY s.id
```

Errors: `promote.invalid` → bad ids, no user, unknown Subject. `promote.unsupported_type` → not a primary. No new codes.

### FFI

`api/proto/engine.proto`: `METHOD_UNPROMOTE_SUBJECT = 100` *(new)*, `UnpromoteSubjectRequest`, `UnpromoteSubjectResponse` per D-6. Constant `MethodUnpromoteSubject` *(new)* and a `case` in `api/ffi/dispatch.go`. Handler `UnpromoteSubject` in `api/ffi/handlers/promote.go`: `parseUserID`, `parseID`, `withProjectCatalog`, `writes.Run(c, writes.Op{Action: "unpromote_subject", UserID: userID}, …)`, `Effects: writeEffectsProto(wrote)`. Regenerate with `./scripts/generate-proto.sh` and commit `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Promoted footer (live) | Snowflake | `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift` `subject`, `paintFooter` | D-10. |
| Promoted footer (twin) | Snowflake | `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` `membershipRow` | D-12. |
| Unpromote | `PVIconButton` look | painted; twin uses `PVIconButton` | Ghost, small, `PVSymbol.unpromote`. No trash icon. |
| Hits and tooltip | — | `EvidenceSubjectCard.actionTargets`, `measuredActionTargets`; `GraphCanvasActionTarget`, `GraphCanvasPointerController`, `GraphCanvasDocumentView` | D-11, D-13. |
| Dispatch | — | `EvidenceGraphModel.performCardAction` | `unpromoteActionID` → `beginUnpromote(subjectID:)`, returns nil. |
| VoiceOver | — | `EvidenceCanvasCard.accessibilityActions(for:)`, SwiftUI chain | D-14. |
| Confirm | `PVConfirm` | `EvidenceGraphView` body, after `.pvDeleteImpact` | `.pvConfirm(item: $model.pendingUnpromote, copy: { PVConfirmCopy(title: L10n.string(L10n.EvidenceGraph.unpromoteConfirmTitle(name: $0.name)), message: L10n.string(L10n.EvidenceGraph.unpromoteConfirmMessage), confirm: L10n.EvidenceGraph.unpromote, cancel: L10n.EvidenceGraph.unpromoteKeep) }, tone: .irreversible, isRunning: model.isUnpromoting, accessibilityIdentifierPrefix: "evidenceGraph.unpromote", onConfirm: { Task { await model.confirmUnpromote() } }, detail: unpromoteDetail)`. |
| Confirm detail | composed | private `unpromoteDetail(_:)` in `EvidenceGraphView` | Nothing when `bridges` is empty and `error` is nil. Otherwise a `VStack(alignment: .leading, spacing: PVSpacing.space3)`: when bridges exist, the header in `splitConfirmDetail`'s micro-caps style, one `Text(verbatim:)` per sentence, then `unpromoteConfirmBridgesNote` in `PVColor.textSecondary`; then, when `error` is set, `PVCallout(tone: .danger, message: error)`. |

Frames, in words. Frame 10: a filed person's footer reads, left to right, `PER-4M2QX`, James Robins, the chevron, a hairline, and the Unpromote button (`arrow.uturn.down`), on the kind band; the unfiled card beside it has the same footer height with the `arrow.turn.right.up` icon and Promote; a relationship card has no Unpromote. Frame 11: the sheet "Unpromote James Robins?", the message, "ALSO UNFILED", one sentence "James Robins, son of John Robins", the note, buttons "Keep filing" and "Unpromote" (primary, no icon). Frame 12: the same sheet with no detail. Frame 13: James's card and the relationship card show Promote; John's card still shows his handle; all cards and edges are where they were.

States: running — confirm disabled by `isRunning`. Error — callout in the detail. Stale — AC-3.4. Accessibility: D-14; the Unpromote region is not a separate accessibility element.

### Strings

L10n keys *(new)*, all `extractionState: manual`, en translated:

- `L10n.EvidenceGraph.unpromote` → `evidenceGraph.subject.unpromote` = "Unpromote"
- `L10n.EvidenceGraph.unpromoteKeep` → `evidenceGraph.subject.unpromoteKeep` = "Keep filing"
- `L10n.EvidenceGraph.unpromoteConfirmTitle(name:)` → `evidenceGraph.subject.unpromoteConfirmTitle` = "Unpromote %1$@?"
- `L10n.EvidenceGraph.unpromoteConfirmMessage` → `evidenceGraph.subject.unpromoteConfirmMessage` = "This filing is removed. The cards stay on the graph, and the handle stays."
- `L10n.EvidenceGraph.unpromoteConfirmBridgesHeader` → `evidenceGraph.subject.unpromoteConfirmBridgesHeader` = "Also unfiled"
- `L10n.EvidenceGraph.unpromoteConfirmBridgesNote` → `evidenceGraph.subject.unpromoteConfirmBridgesNote` = "The other end of each stays filed."

Removed: none.

## 7. Steps

STEP-1 to STEP-3 run anywhere with Go and cgo; STEP-3 also needs `protoc`, `protoc-gen-go`, and `protoc-gen-swift`. STEP-4 to STEP-8 need macOS with Xcode. If a step's tool is missing, stop and say so; do not skip the check.

**STEP-1 — Shared claim release and the bridge unfile**
Files: `core/database/deleteimpact/facets.go`, `core/database/identityclaims/bridgefile.go`. Skill: `add-catalog-delete`.
Do: D-2 and D-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/... ./core/database/identityclaims/... ./core/database/subjects/... ./core/database/promote/...` passes, including `TestSubjectDeleteReleasesOwnClaimPins`, `TestSubjectDelete`, and `TestBridgeFiling`.

**STEP-2 — `promote.Unpromote`**
Files: `core/database/promote/unpromote.go` *(new)*, `core/database/promote/unpromote_test.go` *(new)*, `core/database/audit/scopes_test.go`. Skill: `add-audit-scope`.
Do: D-1, D-4, D-5, D-18. Write T-1, T-2, T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/audit/...` passes.

**STEP-3 — FFI method**
Files: `api/proto/engine.proto`, `api/ffi/dispatch.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/promote_test.go`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-6. `./scripts/generate-proto.sh`. Write T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `rg -n "UnpromoteSubject" api/proto/engine/engine.pb.go macos/App/Platform/Generated/engine.pb.swift` finds both generated types.

**STEP-4 — Swift store and fake**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`. Skill: `add-swift-test`.
Do: `func unpromoteSubject(projectDir: String, userID: String, subjectID: String) async throws -> CatalogWriteEffects` on the protocol, after `promoteSubject`. `GoStore` copies `deleteSubject` with the new request and `CoreMethod.unpromoteSubject`. `FakeStore` per D-17. `ThrowingStore` in `OnboardingModelTests.swift` gets an `unpromoteSubject` that throws, the way its `deleteSubject` does. No protocol default.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` succeeds.

**STEP-5 — Symbol and copy**
Files: `macos/App/DesignSystem/Components/Icon/PVIcon.swift`, `macos/ProvenenciaTests/PVSymbolTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-15, D-16.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/PVSymbolTests -only-testing:ProvenenciaTests/L10nResolutionTests` passes, including T-12.

**STEP-6 — Footer geometry, paint, hits, tooltip**
Files: `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift`, `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` (`publishHitTargets`), `macos/App/Features/GraphCanvas/GraphCanvasHitTarget.swift`, `macos/App/Features/GraphCanvas/GraphCanvasPointerController.swift`, `macos/App/Features/GraphCanvas/GraphCanvasDocumentView.swift`, `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift`, `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`, `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift`. Skill: `add-ui-component`.
Do: D-10, D-11, D-12, D-13; the Promote footer and `promoteAllButton` icon swap from D-15. The `unpromote` target is dispatched in STEP-7; until then `performCardAction` falls through to its `default` and returns nil. Write T-5, T-6, T-13.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceSubjectCardFooterTests -only-testing:ProvenenciaTests/EvidenceCardLayoutTests -only-testing:ProvenenciaTests/GraphCanvasPointerHitTestingTests` passes.

**STEP-7 — Model, confirm, VoiceOver**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-7, D-8, D-9, D-14, D-17 (test helpers); the `performCardAction` case; the `pvConfirm` and detail in §6 macOS. Write T-7, T-8, T-9, T-10, T-11.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-8 — Whole gate**
Files: none new.
Do: Run the gate. Run the manual check in §10.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Docs**
Files: `docs/conclusion-layer-data-model.md` §5.2, `docs/catalog-deletes.md` item 9, `core/database/deleteimpact/README.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (the S10-44 entry's Plan line and status), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/S10-D6-evidence-graph.md` → `docs/deployment-plan/spike-10/design/archive/`, `docs/deployment-plan/spike-10/design/README.md`.
Do: In §5.2 add: "Unpromote on the Evidence graph deletes a primary Subject's accepted claim and, in the same revision, the accepted claims of the bridges on that Source that have it as an end. It does not reject, and it deletes no handle." Do not describe the review queue. In `docs/catalog-deletes.md` item 9 and the `deleteimpact` README add: "`ReleaseClaim` is the one claim removal (pins first, audited). Subject delete and Unpromote both call it." The rest per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-44-unpromote.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Take a filing back from a promoted card on the Evidence graph."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promote/unpromote_test.go` `TestUnpromote` *(new)* | Table over a `bridgeWorld` household (James, John, Mary, event; son-of, participation, spouse; all promoted so the bridges file): (a) unpromote James → James, son-of, participation claims gone; John's, Mary's, the event's, and spouse claims remain; James's handle and both associations exist (`handles`, `members`); one revision `unpromote_subject`. (b) no filed bridges → only James's claim gone. (c) a relationship touching James whose other end is unpromoted → untouched. (d) `UnfileBridgesTx` with another Source id and James's Subject id → no changes; son-of still filed. (e) no accepted claim → no changes, no error, no new revision. (f) relationship Subject → `ErrUnsupportedType`. (g) unknown id, nil user → `ErrInvalid`. | AC-2.3, AC-3.1, AC-3.3, AC-3.4, AC-3.5, D-3, D-4 |
| T-2 | Go | `core/database/promote/unpromote_test.go` `TestUnpromoteKeepsOtherMembersPins` *(new)* | Per D-18: birth promoted, James (census) joined with the pair → unpromote James → James's claim and its pins gone; `f.pinned(first.Claim.ID)` still equals `idsOf(birthObs[0], censusObs[0])`; the birth claim is still accepted. | AC-3.2, D-5 |
| T-3 | Go | `core/database/audit/scopes_test.go` `TestHandleScopes` case "unpromote moves the handle, not the source" *(new)* | Per D-18 → handle revision is latest; `sourceRevisions` unchanged. | AC-3.7 |
| T-4 | Go | `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject` *(new)*, `runRPC` table | Filed Subject → response `effects` non-empty and `ListSubjectMemberships` no longer lists it; bridge Subject → `promote.unsupported_type`; bad id → invalid error. | AC-3.1, AC-3.5, D-6 |
| T-5 | Swift | `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift` `promotedFooterSplitsHandleAndUnpromote` *(new)* | Promoted person, `layout: nil` → `openHandle` and `unpromote` targets do not intersect, both 36pt high at the bottom; `unpromote` 36pt wide at the trailing edge; handle width + 1 + 36 = card width; `unpromote` help is "Unpromote". Unpromoted person and a bridge → no `unpromote` target. | AC-1.1, AC-1.3, D-10, D-11 |
| T-6 | Swift | `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `measuredLayoutMatchesThePaint` (extended) | `placed(...)` gains `handleName: String? = nil`, which sets `membership.name = CatalogNameValue(form: handleName)`. Add case `"longHandleName+promoted"`: promoted, handle name "James Alexander Robins of Wellington North Township". AC-1.2's truncation is not measured: T-6 proves the slot keeps 36pt and the handle region ends before it; the name stays one line through `lineLimit(1)` on the twin and `lines: 1` in the paint. For it and `cited3+promoted`: `unpromote` region is 36×36, flush right and bottom; `openHandle` region equals `promotedFooterParts(footer:).handle`; the footer-height assertion reads the `openHandle` region's height and minY only; the band-overlap loop includes `unpromote` and passes. | AC-1.1, AC-1.2, D-10, D-12 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteVoiceOverActionFollowsFooterAction` *(new)* | `EvidenceCanvasCardContent(model:navigation: WorkspaceNavigation(), layouts: EvidenceCardLayoutStore())` and a card `EvidenceCanvasCard(placed: .subject(placed, nil), text: EvidenceCardText(subject: placed), state:)` with every state flag false, `canCite: true`, no hover: for James with a membership, `namedActions(for:).map(\.name)` has "Unpromote" right after the open-handle name, and running it sets `model.pendingUnpromote`; without a membership, no "Unpromote". | AC-1.5, D-14 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges` *(new)* | `householdModel()`, `performCardAction(subjectID: james, actionID: unpromoteActionID)` → returns nil; no store call recorded; `pendingUnpromote.name == "James Robins"`; `bridges` equal the son-of and participation sentences in bridge-id order; spouse absent. With `fileSonOf: false` → son-of absent. | AC-1.4, AC-2.1, AC-2.3, D-7 |
| T-9 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteWithNoBridgesHasEmptyDetail` *(new)* | `promotableModel()` with James's membership set → `bridges` empty. Unpromoted card → `beginUnpromote` leaves `pendingUnpromote` nil. | AC-2.2, AC-1.3 |
| T-10 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `confirmUnpromoteRefreshesToPromote` *(new)* | `householdModel()`, begin, `confirmUnpromote()` → `"unpromoteSubject id=…"` recorded; `pendingUnpromote` nil; `waitUntil` the graph handle's memberships no longer list James, son-of, or participation, and still list John, Mary, the event, and spouse. Setting `pendingUnpromote = nil` without confirming records no call. | AC-2.4, AC-3.1, D-8, D-9, D-17 |
| T-11 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteFailureStaysInConfirm` *(new)* | `store.unpromoteSubjectError = CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: [])` → `pendingUnpromote` still set, `pendingUnpromote?.error == L10n.Errors.message(for:)` of that same error, `isUnpromoting` false; `beginUnpromote` again → `error` nil. | AC-3.6, D-8 |
| T-12 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` (extended) | Each §6 key resolves to its default value; `unpromoteConfirmTitle(name: "James Robins")` → "Unpromote James Robins?". | AC-2.1, D-16 |
| T-13 | Swift | `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift` `hoverExposesTargetHelp` *(new)* | A card target with an action `help: "Unpromote"` and one without; `mouseMoved` over the first → `hoveredHelp == "Unpromote"`; over the second → nil; `mapFrame` keeps `help`. | AC-1.4, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSubjectDelete` and `TestSubjectDeleteReleasesOwnClaimPins` (STEP-1 extraction), `TestBridgeFiling` (`touchesEnd`), `measuredLayoutMatchesThePaint` and `membershipSwapsPromoteForOpenHandle` (promoted footer target is no longer full width; `footerTargetIsFullWidthAtTheBottom` covers unpromoted cards and stays as it is), `TestResolversCoverEveryEntityType`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: conclusion-layer-data-model.md §5.2, catalog-deletes.md item 9, deleteimpact README (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has James Robins, John Robins, and Mary Smith cards, a relationship "James Robins, son of John Robins", a relationship between John and Mary, and a participation of James in an event. Before promoting, each card's footer shows the turn-up arrow (`arrow.turn.right.up`) and Promote, and the Promote all button shows the same arrow. Promote all, so every card shows a handle.
2. On James's card, hover the Unpromote square (the U-turn-down arrow) at the right of the footer. See the tooltip "Unpromote" and the small fill; the handle row does not highlight. Click the handle row: James's Person page opens. Go Back.
3. Click Unpromote. See "Unpromote James Robins?", the message, "ALSO UNFILED" with the son-of relationship and the participation, and "The other end of each stays filed." Press "Keep filing": every footer is unchanged.
4. Click Unpromote again and confirm. James's card, the son-of card, and the participation card show Promote. John, Mary, the event, and the John–Mary relationship still show handles. Every card and edge is in place. Open the People list: James's former handle `PER-…` is still listed.
5. Unpromote Mary (her only bridge, John–Mary, is filed): the confirm lists it. Promote the event card on a Source with no filed bridges and unpromote it: the detail is empty.
6. With VoiceOver on, focus a filed card and open its actions: "Unpromote" is listed after the open-handle action and opens the same confirm.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Extracting `ReleaseClaim` changes Subject delete. | STEP-1 runs the Subject delete tests before anything else. |
| The client bridge list (D-7) and the engine set (D-3) disagree. | Both read accepted bridge claims whose end Observation names the Subject; the fake uses the engine's rule. The snapshot draws an end from the first Observation per end property, so a bridge with two Observations on one end property can differ; the write is authoritative and the graph refreshes. |
| Tooltip does not show on the canvas. | D-13 sets `toolTip` on the view that owns the mouse; T-13 pins `hoveredHelp`; manual check step 2. |

**Stop and report if:**

- `releaseSubjectClaims`, `bridgesTouching`, `paintFooter`, or `GraphCanvasDocumentView.hitTest` no longer does what §4 says.
- `identityclaims` or `promote` importing `deleteimpact` makes an import cycle (`go build ./core/...` fails with "import cycle not allowed").
- `newBridgeWorld` cannot file the household's bridges (T-1 (a) setup has an unfiled son-of after promoting both ends).
- `NSImage(systemSymbolName:)` is nil for `arrow.turn.right.up` or `arrow.uturn.down` on the CI macOS image, or the SF Symbols app lists either as available only after macOS 14.0 (the deployment target).
- `measuredLayoutMatchesThePaint` cannot pass without changing the unpromoted footer's full-width band (Out).
- Setting `toolTip` on `GraphCanvasDocumentView` shows no tooltip in the manual check.
- After step 4 of §10 the Person page of James's former handle still shows James as a member, or the People list does not refresh (`session.apply(effects)` missed a key).

## 12. Open questions

None. Q-1 (the icon) was answered 2026-10-16: D-15.

## 13. PR

**Title:** S10-44 — Unpromote a filed card
**Body:** A promoted person, event, or place card on the Evidence graph can take its filing back. What shipped against US-1 (footer button), US-2 (confirm naming the bridges), US-3 (one audited delete of this claim and the bridge claims on this Source); AC evidence by test name; deviations from the plan; the manual check in §10. `Plan: docs/deployment-plan/spike-10/prs/S10-44-unpromote.md`.
