# S10-44 — Unpromote a filed card

**Status:** Draft
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
**Planned:** 2026-10-16 against `main` @ `536d278`
**Review:** — (not reviewed yet)

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
| AC-1.1 | A person card with an accepted claim on `PER-4M2QX`, name "James Robins" | The footer is 36pt. It paints, left to right: the ref, the name, the chevron, a 1pt hairline, then the unlink icon button in a 36pt square at the trailing edge. The handle hit and the Unpromote hit do not intersect, and together they cover the footer except the 1pt hairline. |
| AC-1.2 | A card with a long handle name ("James Alexander Robins of Wellington North Township") | The name truncates before the chevron. The Unpromote square keeps its 36pt width. |
| AC-1.3 | An unpromoted person card, and a relationship card that is filed | Neither shows Unpromote. Neither has an `unpromote` action target. The relationship card's footer is unchanged. |
| AC-1.4 | The pointer hovers the Unpromote square, then presses it | Only the 22pt button fill changes (ink at 0.10, then 0.18). The handle row does not highlight. Releasing opens the confirm. It does not navigate. The tooltip over the square reads "Unpromote". |
| AC-1.5 | VoiceOver on a promoted person card | The card's custom actions include "Unpromote", after the open-handle action. Running it opens the same confirm. An unpromoted card's actions do not include it. |

**US-2.** As a researcher, when I press Unpromote, the confirm tells me exactly which filings go, so I am not surprised by what else comes off.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | James is filed. A relationship "James Robins, son of John Robins" is filed. A participation of James in the 1881 census event is filed. A relationship between John and Mary is filed | The confirm title is "Unpromote James Robins?". The message is "This filing is removed. The cards stay on the graph, and the handle stays." The detail has the header "Also unfiled", then the two bridge card sentences that touch James, in the order of D-7, then "The other end of each stays filed." The John–Mary relationship is not listed. Confirm reads "Unpromote", cancel reads "Keep filing". There is no trash icon. |
| AC-2.2 | James is filed and no filed bridge on this Source has James as an end | Title and message are as in AC-2.1. The detail is empty. |
| AC-2.3 | A relationship touching James exists on this Source but is not filed (John is unpromoted) | That relationship is not listed. |
| AC-2.4 | The confirm is open and the researcher presses "Keep filing" or Escape | The sheet closes. No store call is made. Every footer is as before. |

**US-3.** As a researcher, when I confirm, this filing and the filings that depended on it are removed in one step, and nothing else is.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The catalog of AC-2.1, and Unpromote James is confirmed | One revision, action `unpromote_subject`, deletes James's claim, the son-of relationship's claim, and the participation's claim, with each claim's pins. James's card, the son-of card, and the participation card show Promote. John's card and the 1881 census event card still show their handles. The John–Mary relationship stays filed. `PER-4M2QX` still exists with no members. Each association handle still exists. Every card and edge stays. |
| AC-3.2 | James and a second Subject on another Source are both on `PER-4M2QX`, and their claims pin each other's Observations | Only James's claim and its own pins are deleted. The other Subject's claim keeps its status and its pins, including pins on James's Observations. |
| AC-3.3 | A relationship on a *different* Source has James's Subject as an end | It cannot: a bridge's ends are Subjects on that bridge's Source. The query filters by this Source. T-1 has a case for a bridge on another Source that names the Subject, and it stays filed. |
| AC-3.4 | The engine is asked to unpromote a Subject that has no accepted claim (a second click after the first landed) | No revision is written and no error is returned. The graph refreshes and shows Promote. |
| AC-3.5 | The engine is asked to unpromote a relationship Subject, or an id that is not a Subject | `promote.unsupported_type` for the relationship. `promote.invalid` for the unknown id. Nothing is written. |
| AC-3.6 | The write fails (the store throws) | The confirm stays up. Its detail shows a danger callout with the error message under any bridge list. Confirm is enabled again. "Keep filing" closes it with nothing changed. |
| AC-3.7 | After AC-3.1, the revision's audit scopes | The `canonical_entity` scope of `PER-4M2QX` and of each association handle moves. No `source` scope is recorded. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Engine: `promote.Unpromote` deletes the accepted claim of one primary Subject and the accepted claims of the bridges on its Source that have it as an end, with pins, audited (EG-10). | STEP-1, STEP-2, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.7 |
| FFI method `METHOD_UNPROMOTE_SUBJECT` and its handler, returning write effects. | STEP-3, AC-3.1, AC-3.5 |
| Swift store: `GenealogyStore.unpromoteSubject`, `GoStore`, `FakeStore`. | STEP-4, AC-3.1, AC-3.6 |
| EG-8. Promoted footer: handle row, hairline, Unpromote square; hits, hover, tooltip, VoiceOver action. Not on bridge or unpromoted cards. | STEP-5, STEP-6, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| EG-9. `PVConfirm` with an item, tone irreversible, title, message, bridge detail, labels, cancel. | STEP-7, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| EG-10. Confirm writes, the graph refreshes to Promote, errors stay in the confirm. | STEP-7, AC-3.1, AC-3.4, AC-3.6 |
| Docs: `conclusion-layer-data-model.md` §5.2 notes Unpromote; map, completed, brief archive. | STEP-9 |

| Out | Home |
| --- | --- |
| Rejecting a claim ("not this entity"), the weak-claim queue, re-pin, accept-as-is. | [`docs/ideas/identity-claim-review.md`](../../../ideas/identity-claim-review.md) |
| Deleting a handle, including one left with no members. | not planned — working rule "Unpromote deletes the filing" |
| Deleting or moving Evidence cards or their Observations. | not planned — EG-10 |
| Unpromote from the Person, Event, or Place page. | not planned — S10-D6 §2.1 |
| Unpromote on a bridge card. A bridge's claim comes off only with an end. | not planned — S10-D6 EG-8 |
| An undo, or a Toast after confirm. | not planned — S10-D6 §2.1 |
| Unpromote for a selection of cards. | not planned |
| Renaming `CatalogMutation.promotedSubject` to a name that covers both directions. | not planned — D-9 |
| Changing the full-width hit of the Promote footer on unpromoted cards. | not planned |
| Changing `PVIconButton` to read `pvHostInteraction`. | not planned — D-10 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Claims | `core/database/migrations/000032.sql` `identity_claims` | `status IN ('provisional','accepted','rejected')`. At most one accepted claim per Subject (`identity_claims_one_accepted_per_subject`). Pins in `identity_claim_evidence`, ON DELETE CASCADE as a backstop. Nothing in `core/` writes `provisional` or `rejected`. |
| Claim package | `core/database/identityclaims/identityclaims.go` `InsertTx`, `AcceptedEntityForSubjectTx` | Inserts and reads claims. No delete. |
| Claim delete | `core/database/deleteimpact/facets.go` `releaseSubjectClaims` | The only claim delete. Loops over every claim of a Subject: `ReleaseFacets(tx, catalogmodel.KindIdentityClaim, id)` for the pins, `DELETE FROM identity_claims`, then a `rowchange.DeletedRow` change carrying `id`, `subject_id`, `entity_id`, `subject_type_id`, `status`, `confidence_grade_id`, `argument`. Used only by Subject delete. The per-claim body is inline. |
| Bridges | `core/database/identityclaims/bridgefile.go` `bridgesTouching`, `fileBridge`, `resolveEndpoint` | `bridgesTouching` selects bridge Subjects (`connectrules.BridgeTypeKeys`) that are the Subject or that have an Observation whose `value_subject_id` is the Subject, on any Source. `fileBridge` files a bridge once each end property resolves to one accepted handle; it reads every end Observation, with no polarity filter. |
| Promote | `core/database/promote/promote.go` `Save`, `PrimaryKind`, `ErrInvalid`, `ErrUnsupportedType` | Files one primary Subject in a `promote_subject` revision and calls `identityclaims.AppendFiling` to file touching bridges. |
| Scopes | `core/database/effects/registry.go` `identity_claims`, `identity_claim_evidence` | A claim change resolves `HandleScope: field("entity_id")` and `Source` (graph cache) from `subject_id`. A delete resolves through the ghost row, so the `DeletedRow` must carry `subject_id`, `entity_id`, `status`. No `source` audit scope for conclusion work. |
| Write runner | `core/writes/run.go` `Run` | One transaction, audit, `autoreconciler.RecomputeReported` on `set.Handles`, search reprojection, commit, notify. An empty change list rolls back and returns no error. |
| FFI | `api/proto/engine.proto` `Method`, `DeleteSubjectResponse`, `WriteEffects` | Last method `METHOD_GET_NAME_FORMAT = 99`. `METHOD_PROMOTE_SUBJECT = 85`. |
| FFI dispatch | `api/ffi/dispatch.go` `MethodPromoteSubject`, `MethodDeleteSubject` | Constants and one `case` each. |
| Handlers | `api/ffi/handlers/promote.go` `PromoteSubject`; `api/ffi/handlers/subjects.go` `DeleteSubject` | `DeleteSubject` is the template: `withProjectCatalog`, `writes.Run`, `Effects: writeEffectsProto(wrote)`. |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `deleteSubject`, `promoteSubject`; `macos/App/Platform/GoStore.swift` `deleteSubject`; `macos/App/Platform/CoreInvoke.swift` `CoreMethod.deleteSubject`; `macos/App/Platform/FakeStore.swift` `deleteSubject` | No unpromote call. |
| Card paint | `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift` `measure` / `subject`, `paintFooter` | The live card is an AppKit `NSView`. The footer region (`promoteActionID` or `openHandleActionID`) is full card width × 36. The promoted branch paints the ref at x 13 width 72, the name at x 91 width `rect.width - 120`, the chevron at `rect.maxX - 27`, and fills the whole footer on hover. The card has no tooltips. |
| SwiftUI twin | `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` `membershipRow`, `cardFooter`, `footerHeight`, `openHandleActionID` | Paint-only twin, used by `ghost` and by layout tests. `membershipRow` is a full-width row tagged `openHandleActionID`. |
| Hits | `EvidenceSubjectCard.actionTargets`, `measuredActionTargets`, `footerFrame`; `macos/App/Features/EvidenceGraph/EvidenceCardLayout.swift` `band`, `centeredHit` | The footer target is the full-width `band`, or `footerFrame` without a layout. `centeredHit(id, origin:, minSize: 0)` returns the painted rect in document space. |
| Dispatch | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `performCardAction`, `openHandle`, `beginDelete`, `confirmPendingImpact` | `promoteActionID` navigates; `openHandleActionID` opens the page; delete starts `DeleteImpactFlow`. The comment above `let sourceID` names "a v1 Promote confirm" that no longer exists. |
| VoiceOver | `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift` `accessibilityActions(for:)` and the SwiftUI `.accessibilityAction` chain | Footer action is promote or open handle, then delete. |
| Confirm | `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift` `pvConfirm(item:copy:tone:isRunning:…)`, `PVConfirmTone.irreversible` | Cancel clears the item. The caller dismisses after an async confirm; an error goes in `detail`. Precedent: `macos/App/Features/Names/NameValueEditorForm.swift` `pendingSplit`, `splitConfirmDetail`. |
| Bridge sentence | `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` `sentence(for:in:orders:)` | The card sentence. `beginDelete` already uses it to name a bridge. |
| Snapshot | `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `SourceGraphPlacedSubject.membership`, `SourceGraphPlacedBridge.endpointAID`, `endpointBID`, `membership` | The graph's view of who is filed and which bridge touches which card. |
| Refresh | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `promotedSubject(sourceId:)`; `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` | `.promotedSubject` invalidates the Source graph, memberships, graph progress, and every Conclusion list and page. |
| Graph view | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `.pvDeleteImpact` | The confirm modifiers hang off the view body. |
| Symbols | `macos/App/DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` | `link = "link"`. No unlink case. |
| Copy | `macos/App/Platform/L10n.swift` `L10n.EvidenceGraph` "Promote + membership footer"; `macos/App/Resources/Localizable.xcstrings` | `evidenceGraph.subject.promote` and siblings. |
| Tests | `core/database/promote/promote_test.go` `TestSave`; `core/database/promote/bridges_test.go` `TestBridgeFiling`; `core/database/audit/scopes_test.go` `TestHandleScopes`; `api/ffi/handlers/promote_test.go` `TestPromoteSubject`; `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift`; `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `measuredLayoutMatchesThePaint`; `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `promotedCardShowsMembershipAndNoLongerOffersPromote`; `macos/ProvenenciaTests/PVSymbolTests.swift` `symbolResolves` | `measuredLayoutMatchesThePaint` asserts footer bands do not overlap and that the footer target is 36pt high at the card bottom. |

**Precedent.** Subject delete (`DeleteSubject` handler, `subjects.Delete`, `releaseSubjectClaims`) is the shape of an audited, multi-row conclusion delete in one `writes.Run`. Copy its handler and its claim-release body; do not copy its Impact check or its manual handle recompute (`writes.Run` recomputes from the returned changes, as `promote.Save` relies on). On the Mac, copy `confirmPendingImpact` for the store call plus `session.apply(effects, …)`, and `NameValueEditorForm` for an irreversible `pvConfirm` with a listed detail. The Promote footer's paint-only kit button in `promoteFooter` is the precedent for the twin.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The writer is `promote.Unpromote(tx *database.Tx, userID, subjectID []byte) ([]rowchange.Change, error)` *(new)* in `core/database/promote/unpromote.go` *(new)*. Action `unpromote_subject`. | `promote` owns filing a primary and already validates `PrimaryKind` and reuses `ErrInvalid` / `ErrUnsupportedType`. Unpromote is its inverse. | A new package — one more concept for one function. `identityclaims` — it does not know primaries. |
| D-2 | The claim delete is `deleteimpact.ReleaseClaimTx(tx *sql.Tx, claimID []byte) (Released, error)` *(new)*, extracted from the loop body of `releaseSubjectClaims`, which then calls it per claim. It reads the row by id, releases pins through `ReleaseFacets(…, KindIdentityClaim, …)`, deletes, and returns the `DeletedRow` change. | One definition of "remove a claim, audited, pins first". `deleteimpact` imports no database package, so `promote` can import it; `identityclaims` cannot be imported by `deleteimpact` without a cycle risk, and a copy would duplicate the `DeletedRow` field list. | A second copy in `identityclaims` or `promote`. A raw `DELETE` relying on the CASCADE — the domain doc forbids it. |
| D-3 | The bridge set is `identityclaims.FiledSourceBridgesTouchingTx(tx *sql.Tx, sourceID, subjectID []byte) ([]FiledBridge, error)` *(new)*, `FiledBridge{SubjectID, ClaimID []byte}` *(new)*. A bridge qualifies when its Subject type is in `connectrules.BridgeTypeKeys()` (origin provenencia), `s.source_id = sourceID`, it has an accepted claim, and some Observation on it has `value_subject_id = subjectID`. No polarity filter. Ordered by bridge Subject id. The end predicate is one SQL fragment `touchesEnd` *(new)* shared with `bridgesTouching`. | It is the rule filing used (`bridgesTouching`, `resolveEndpoint` read every end Observation), so what Unpromote takes off is what filing put on. The Source filter is the working rule. | Exporting `bridgesTouching` and filtering in Go — it has no Source filter and returns unfiled bridges. A polarity filter — it would leave a bridge filed whose filing used this Subject. |
| D-4 | Order inside `Unpromote`: validate the user id and Subject id (`ErrInvalid`), load the Subject (`ErrInvalid` when missing), `PrimaryKind` (`ErrUnsupportedType`), read its accepted claim id. No accepted claim → return `nil, nil` (AC-3.4). Otherwise release this Subject's claim, then each bridge claim from D-3, and return the changes in that order. | Idempotent for a stale second click; `writes.Run` rolls back an empty change list with no error. | `identityclaims.ErrNotMember` *(would be new)* — a new error code, L10n entry, and a red callout for a state the graph already fixes on refresh. |
| D-5 | Only accepted claims are deleted, on this Subject and on the bridges. Other claims on the handle, their pins (including pins on this Subject's Observations), and every handle row stay. No `autoreconciler` or `conclusionheaders` call in `Unpromote`. | Working rule and §5.2. `writes.Run` recomputes `set.Handles` (the entity and, for an accepted claim, its observers) and header deps. | Recomputing by hand like `subjects.Delete` — duplicate work. |
| D-6 | FFI: `METHOD_UNPROMOTE_SUBJECT = 100`, `UnpromoteSubjectRequest { project_dir = 1; user_id = 2; subject_id = 3; }`, `UnpromoteSubjectResponse { WriteEffects effects = 1; }`. Handler `UnpromoteSubject` *(new)* in `api/ffi/handlers/promote.go`. The source id is not in the request; the engine reads it from the Subject. | Mirrors `DeleteSubject`. One source of truth for the Source. | Returning the unfiled ids — the client refreshes the graph and does not need them. |
| D-7 | The confirm lists bridges from the client snapshot: `snapshot.bridges` where `endpointAID == id || endpointBID == id` and `membership != nil`, sorted by bridge id, each as `EvidenceBridgeEdgeSummary.sentence(for:in:orders:)`. Built once when the confirm opens, into `PendingUnpromote` *(new)*. | The graph already shows those edges and footers; the engine recomputes the authoritative set in the write (D-3). Both sort by bridge Subject id. A mismatch only happens on a stale graph, and the write then follows the catalog. | A preview RPC — a second method and a round trip for what the snapshot already holds. |
| D-8 | `PendingUnpromote: Identifiable` *(new)*, in `EvidenceGraphModel.swift`: `id` (subject id), `name` (`placed.displayName(orders:)`), `bridges: [String]`. Model state: `pendingUnpromote: PendingUnpromote?`, `isUnpromoting: Bool`, `unpromoteError: String?` *(all new)*. `beginUnpromote(subjectID:)` *(new)* builds it for a promoted primary and does nothing otherwise. `confirmUnpromote() async` *(new)* sets running, calls the store, then `session.apply(effects, .promotedSubject(sourceId: sourceID))` and clears the item; on throw it keeps the item and sets `unpromoteError` to the error's `localizedDescription`. Opening the confirm clears `unpromoteError`. | Item snapshot is the `pvConfirm` contract. Errors in the detail are the `pvConfirm` contract. The stale v1 comment above `let sourceID` is replaced by the doc comment of `pendingUnpromote`. | A Bool binding. Reusing `DeleteImpactFlow` — the brief says not DeleteImpact. |
| D-9 | Refresh uses the existing `CatalogMutation.promotedSubject(sourceId:)`. | It invalidates exactly what changes: the graph, memberships, progress, and Conclusion lists and pages. A new case means editing every switch for the same invalidation. | `.unpromotedSubject` *(would be new)*, or a rename — Out. |
| D-10 | Footer geometry, in card coordinates, footer rect `f` (height 36): Unpromote slot `CGRect(x: f.maxX - 36, y: f.minY, width: 36, height: 36)` tagged `unpromoteActionID = "unpromote"` *(new)*. Hairline 1pt wide at `x = f.maxX - 37`, from `f.minY + 8` to `f.maxY - 8`, colour `style.line`. Handle region `CGRect(x: 0, y: f.minY, width: f.width - 37, height: 36)`, still tagged `openHandleActionID`. Name width `f.width - 157`. Chevron at `f.maxX - 37 - 27`. The handle hover fill covers only the handle region. The button paints a 22×22 rounded rect (`PVRadius.sm`) centered in the slot, filled `style.ink` at 0.10 on hover and 0.18 on press, clear at rest, and the unlink symbol at 13pt in `style.ink`. These constants live on `EvidenceSubjectCard` as `unpromoteSlotWidth = 36` *(new)*. | EG-8: own hit, no overlap, 36pt footer, name truncates and the button does not. 36pt square is a full-height hit centered on what is drawn (working rule "Hit boxes are the painted controls"). Ink fills match the handle row's ghost look on the kind band. | A hosted SwiftUI `PVIconButton` on the live card — the card is painted in AppKit (`EvidenceAppKitCard`). A 22pt hit — smaller than the 28pt header icon hits. |
| D-11 | Hit targets read regions, not bands, for the promoted footer: `measuredActionTargets` uses `layout.centeredHit(openHandleActionID, origin:, minSize: 0)` and the same for `unpromoteActionID`. The layout-free fallback in `actionTargets` splits `footerFrame` the same way with `unpromoteFrame(cardFrame:)` *(new)* and `handleFrame(cardFrame:)` *(new)*. The unpromoted Promote footer keeps its full-width band. | `band` is full card width by definition and would overlap. `centeredHit` with `minSize: 0` is the painted rect. | A new `EvidenceCardLayout.rect(_:origin:)` helper — `centeredHit(…, minSize: 0)` already is that. |
| D-12 | The SwiftUI twin's `membershipRow` becomes an `HStack(spacing: 0)` of the existing handle row (frame width `.infinity`, tagged `openHandleActionID`), a 1pt `Rectangle` filled `style.line` at height 20, and a 36×36 slot holding `PVIconButton(.unlink, label: L10n.EvidenceGraph.unpromote, size: .sm, action: {})` with `.allowsHitTesting(false)`, tagged `unpromoteActionID`, `.accessibilityHidden(true)`. The twin shows no hover on the button. | Layout tests measure the twin; it must match the AppKit paint. `PVIconButton` does not read `pvHostInteraction`. | Teaching `PVIconButton` host interaction — Out. |
| D-13 | Tooltip: `EvidenceAppKitCard` adds one tool-tip rect for the `unpromoteActionID` region (offset by `GraphCanvasItemLayer.paintOutset`) with `addToolTip(_:owner:userData:)`, owner `self`, and conforms to `NSViewToolTipOwner`, returning `L10n.string(L10n.EvidenceGraph.unpromote)`. | EG-8 asks for help "Unpromote"; the live card is an `NSView`, so `.help` does not reach it. | `toolTip` on the whole card — it would show over every control. |
| D-14 | VoiceOver: a custom action named `L10n.string(L10n.EvidenceGraph.unpromote)` ("Unpromote") right after the footer action, only when `placed.membership != nil`, in both `accessibilityActions(for:)` and the SwiftUI chain. It calls `model.beginUnpromote(subjectID:)`. | EG-8: the accessible label is Unpromote. The card element already names the subject. | "Unpromote James Robins" — a second format key for a name the card label already reads. |
| D-15 | `PVSymbol.unlink` *(new)*, doc comment "Lucide `unlink` — Unpromote on a promoted card (S10-44).", raw value from Q-1. It is added to `PVSymbolTests.symbolResolves`. | The brief names the unlink icon; `PVSymbol` is the only symbol registry. | `PVSymbol.link` — it reads as "start a bridge", the opposite action. |
| D-16 | Copy, in `L10n.EvidenceGraph` under the promote footer mark: `unpromote` "Unpromote" (button help, VoiceOver action, confirm button); `unpromoteKeep` "Keep filing"; `unpromoteConfirmTitle(name:)` "Unpromote %1$@?"; `unpromoteConfirmMessage` "This filing is removed. The cards stay on the graph, and the handle stays."; `unpromoteConfirmBridgesHeader` "Also unfiled"; `unpromoteConfirmBridgesNote` "The other end of each stays filed." Keys in §6 Strings. | EG-9: title names the subject, message says the filing is removed and what stays, detail lists bridges and says the other end stays filed, empty detail with no bridges. The message carries "cards stay, handle stays" so it is said even when the detail is empty. | Putting "cards stay, handle stays" in the detail — EG-9 needs the detail empty without bridges. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `promote.Unpromote` *(new)* | func | `promote.Save`, `subjects.Delete` | `Save` creates; `Delete` removes the Subject. This is the inverse of `Save` and lives beside it. |
| `deleteimpact.ReleaseClaimTx` *(new)* | func | `releaseSubjectClaims` | Extracted from it; `releaseSubjectClaims` calls it. No copy. |
| `identityclaims.FiledSourceBridgesTouchingTx`, `FiledBridge`, `touchesEnd` *(new)* | func / type / SQL fragment | `bridgesTouching`, `FileSourceBridgesTx` | Neither filters by Source and accepted claim together. `touchesEnd` is shared with `bridgesTouching`. |
| `METHOD_UNPROMOTE_SUBJECT`, `UnpromoteSubjectRequest`, `UnpromoteSubjectResponse` *(new)* | FFI | `METHOD_DELETE_SUBJECT` | A different write; same message shape. |
| `handlers.UnpromoteSubject` *(new)* | func | `handlers.DeleteSubject` | Same shape, in `promote.go` beside `PromoteSubject`. |
| `GenealogyStore.unpromoteSubject`, `GoStore.unpromoteSubject`, `FakeStore.unpromoteSubject`, `CoreMethod.unpromoteSubject` *(new)* | func | `deleteSubject` on each | One store method per engine method. |
| `PendingUnpromote`, `beginUnpromote`, `confirmUnpromote`, `pendingUnpromote`, `isUnpromoting`, `unpromoteError` *(new)* | type / func / state | `DeleteImpactFlow`, `NameValueEditorForm.pendingSplit` | DeleteImpact is ruled out by the brief; the item snapshot follows `pendingSplit`. |
| `EvidenceSubjectCard.unpromoteActionID`, `unpromoteSlotWidth`, `unpromoteFrame(cardFrame:)`, `handleFrame(cardFrame:)` *(new)* | constants / func | `openHandleActionID`, `footerFrame(cardFrame:)` | New region on the same footer. |
| `PVSymbol.unlink` *(new)* | key | `PVSymbol.link` | The brief draws unlink, not link. |
| `L10n.EvidenceGraph.unpromote…` *(new)* | keys | `L10n.EvidenceGraph.promote…` | New copy in the same section. |
| Tool-tip owner on `EvidenceAppKitCard` *(new)* | conformance | none on the canvas | First tooltip on a painted card. |

**Builds on.** `writes.Run` and the effects registry for audit and handle recompute (`add-audit-scope`); `deleteimpact.ReleaseFacets` for pins (`add-catalog-delete`); the `DeleteSubject` handler template (`add-ffi-handler`); `pvConfirm(item:)` (`add-ui-component`); the region-tagged card layout from S10-01.
**Deletes.** The stale comment "Snapshot of the subject a v1 Promote confirm names" in `EvidenceGraphModel.swift`. The inline per-claim body in `releaseSubjectClaims`, which moves into `ReleaseClaimTx`.
**Next adds.** The weak-claim review (`docs/ideas/identity-claim-review.md`) can call `ReleaseClaimTx` for a delete and add a reject beside it. Unpromote from a handle page, if ever planned, calls the same store method.

### Schema

None — no catalog change. Claims and pins are deleted with the existing tables. Audit uses the existing `identity_claim` and `identity_claim_evidence` entity types.

### Go core

| Package | Change |
| --- | --- |
| `core/database/deleteimpact` | `func ReleaseClaimTx(tx *sql.Tx, claimID []byte) (Released, error)` *(new)* in `facets.go`. `releaseSubjectClaims` selects ids, then calls it for each. Behavior of Subject delete is unchanged. |
| `core/database/identityclaims` | `type FiledBridge struct { SubjectID, ClaimID []byte }` *(new)*; `func FiledSourceBridgesTouchingTx(tx *sql.Tx, sourceID, subjectID []byte) ([]FiledBridge, error)` *(new)* in `bridgefile.go`; `touchesEnd` *(new)* used by it and by `bridgesTouching`. |
| `core/database/promote` | `func Unpromote(tx *database.Tx, userID, subjectID []byte) ([]rowchange.Change, error)` *(new)* in `unpromote.go` *(new)*, steps per D-4. |

```go
// sketch: bridge query
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
| Promoted footer (live) | Snowflake | `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift` `paintFooter`, `subject` | D-10 geometry. Regions per D-10. Tooltip per D-13. |
| Promoted footer (twin) | Snowflake | `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift` `membershipRow` | D-12. |
| Unpromote | `PVIconButton` look | painted; twin uses `PVIconButton` | Ghost, small, `PVSymbol.unlink`. No trash icon. |
| Hits | — | `EvidenceSubjectCard.actionTargets`, `measuredActionTargets` | D-11. Bridge cards and unpromoted cards get no `unpromote` target. |
| Dispatch | — | `EvidenceGraphModel.performCardAction` | `unpromoteActionID` → `beginUnpromote(subjectID:)`, returns nil. |
| Confirm | `PVConfirm` | `EvidenceGraphView` body, after `.pvDeleteImpact` | `.pvConfirm(item: $model.pendingUnpromote, copy:, tone: .irreversible, isRunning: model.isUnpromoting, accessibilityIdentifierPrefix: "evidenceGraph.unpromote", onConfirm: { Task { await model.confirmUnpromote() } }, detail:)`. Copy: title `unpromoteConfirmTitle(name:)`, message `unpromoteConfirmMessage`, confirm `unpromote`, cancel `unpromoteKeep`. |
| Confirm detail | composed | private `unpromoteDetail(_:)` in `EvidenceGraphView` | Empty when `bridges` is empty and there is no error. Else a `VStack(alignment: .leading, spacing: PVSpacing.space3)`: the header in the `splitConfirmDetail` micro-caps style, one `Text(verbatim:)` line per sentence, then `unpromoteConfirmBridgesNote` in `PVColor.textSecondary`. Under it, when `unpromoteError` is set, `PVCallout(tone: .danger, message:)`. |

Frames, in words. Frame 10: a filed person's footer reads, left to right, `PER-4M2QX`, James Robins, the chevron, a hairline, and the unlink button, on the kind band; the unfiled card beside it has the same footer height with Promote; a relationship card has no Unpromote. Frame 11: the sheet "Unpromote James Robins?", the message, "ALSO UNFILED", one sentence "James Robins, son of John Robins", the note, buttons "Keep filing" and "Unpromote" (primary, no icon). Frame 12: the same sheet with no detail. Frame 13: James's card and the relationship card show Promote; John's card still shows his handle; all cards and edges are where they were.

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

**STEP-1 — Extract the claim release and add the bridge query**
Files: `core/database/deleteimpact/facets.go`, `core/database/identityclaims/bridgefile.go`. Skill: `add-catalog-delete`.
Do: D-2 and D-3. `releaseSubjectClaims` keeps its behavior.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/... ./core/database/identityclaims/... ./core/database/subjects/...` passes, including `TestSubjectDeleteReleasesOwnClaimPins` and `TestSubjectDelete`.

**STEP-2 — `promote.Unpromote`**
Files: `core/database/promote/unpromote.go` *(new)*, `core/database/promote/unpromote_test.go` *(new)*, `core/database/audit/scopes_test.go`. Skill: `add-audit-scope`.
Do: D-1, D-4, D-5. Write T-1, T-2, T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/audit/...` passes.

**STEP-3 — FFI method**
Files: `api/proto/engine.proto`, `api/ffi/dispatch.go`, `api/ffi/handlers/promote.go`, `api/ffi/handlers/promote_test.go`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-6. `./scripts/generate-proto.sh`. Write T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `git diff --stat` shows both generated files.

**STEP-4 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/CoreInvoke.swift`, `macos/App/Platform/FakeStore.swift`. Skill: `add-swift-test`.
Do: `func unpromoteSubject(projectDir: String, userID: String, subjectID: String) async throws -> CatalogWriteEffects` on the protocol, after `promoteSubject`. `GoStore` copies `deleteSubject`. `FakeStore` copies `deleteSubject`'s recording (`"unpromoteSubject id=…"`), calls `markCatalogSessionHeld`, removes the Subject's membership and the memberships of the bridges on its Source whose ends include it, and throws `unpromoteSubjectError` *(new)* when it is set, the way `setSubjectPositionError` is used.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'` succeeds.

**STEP-5 — Symbol and copy**
Files: `macos/App/DesignSystem/Components/Icon/PVIcon.swift`, `macos/ProvenenciaTests/PVSymbolTests.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: `PVSymbol.unlink` per Q-1's answer; add it to `symbolResolves` arguments. §6 Strings.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `symbolResolves` passes for `.unlink`.

**STEP-6 — Footer paint, hits, tooltip, VoiceOver**
Files: `macos/App/Features/EvidenceGraph/EvidenceAppKitCard.swift`, `macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift`, `macos/App/Features/EvidenceGraph/EvidenceCanvasCard.swift`, `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift`, `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift`. Skill: `add-ui-component`.
Do: D-10, D-11, D-12, D-13, D-14. Write T-5, T-6, T-7.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceSubjectCardFooterTests -only-testing:ProvenenciaTests/EvidenceCardLayoutTests` passes.

**STEP-7 — Confirm and write**
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift`, `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: D-7, D-8, D-9; the `performCardAction` case; the `pvConfirm` and detail in §6 macOS. Write T-8, T-9, T-10, T-11.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-8 — Whole gate**
Files: none new.
Do: Run the gate. Run the manual check in §10.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Docs**
Files: `docs/conclusion-layer-data-model.md` §5.2, `completed.md`, `deployment-plan.md` (S10-44 entry: Plan line and status), this plan (status, Landed note), `design/S10-D6-evidence-graph.md` → `design/archive/`, `design/README.md`.
Do: In §5.2 add one paragraph: "Unpromote on the Evidence graph deletes a primary Subject's accepted claim and, in the same revision, the accepted claims of the bridges on that Source that have it as an end. It does not reject, and it deletes no handle." Do not describe the review queue. The rest per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-44-unpromote.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Take a filing back from a promoted card on the Evidence graph."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promote/unpromote_test.go` `TestUnpromote` *(new)* | Table: (a) James filed, son-of and participation filed, John–Mary filed → James, son-of, participation claims gone; John's, the event's, John–Mary claims remain; `PER-4M2QX` and associations exist; one revision `unpromote_subject`. (b) no filed bridges → only James's claim gone. (c) relationship touching James unfiled → untouched. (d) a bridge on another Source whose Observation names James's Subject → stays filed. (e) no accepted claim → no changes, no error. (f) relationship Subject → `ErrUnsupportedType`. (g) unknown id, nil user → `ErrInvalid`. | AC-2.3, AC-3.1, AC-3.3, AC-3.4, AC-3.5, D-3, D-4 |
| T-2 | Go | `core/database/promote/unpromote_test.go` `TestUnpromoteKeepsOtherMembersPins` *(new)* | James and a Subject on another Source joined with pinned pairs → James's claim and its pins gone; the other claim accepted with its pins, including those on James's Observations. | AC-3.2, D-5 |
| T-3 | Go | `core/database/audit/scopes_test.go` `TestHandleScopes` (new case "unpromote moves the handle and each association") | After (a) of T-1 → `canonical_entity` scopes of `PER-4M2QX` and both associations moved; no `source` scope row for that revision. | AC-3.7 |
| T-4 | Go | `api/ffi/handlers/promote_test.go` `TestUnpromoteSubject` *(new)*, `runRPC` table | Filed Subject → response `effects` non-empty and `ListSubjectMemberships` no longer lists it; bridge Subject → `promote.unsupported_type`; bad id → invalid error. | AC-3.1, AC-3.5, D-6 |
| T-5 | Swift | `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift` `promotedFooterSplitsHandleAndUnpromote` *(new)* | Promoted person, `layout: nil` → `openHandle` and `unpromote` targets do not intersect, both 36pt high at the bottom, `unpromote` 36pt wide at the trailing edge, the two plus 1pt equal the card width. Unpromoted person and a bridge → no `unpromote` target. | AC-1.1, AC-1.3, D-11 |
| T-6 | Swift | `macos/ProvenenciaTests/EvidenceCardLayoutTests.swift` `measuredLayoutMatchesThePaint` (extended) | For `cited3+promoted` and a new `longName+promoted` case → `unpromote` region exists, 36×36, flush right and bottom; the band-overlap loop includes it and passes. | AC-1.1, AC-1.2, D-10, D-12 |
| T-7 | Swift | `macos/ProvenenciaTests/EvidenceSubjectCardFooterTests.swift` `footerAccessibilityNames` (extended) | Promoted card's custom actions contain "Unpromote" right after the open-handle name; unpromoted card's do not. | AC-1.5, D-14 |
| T-8 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteActionOpensConfirmNamingBridges` *(new)* | FakeStore graph of AC-2.1, `performCardAction(subjectID: james, actionID: unpromoteActionID)` → returns nil; `pendingUnpromote.name == "James Robins"`; `bridges` equal the two sentences in bridge-id order; John–Mary absent; unfiled relationship absent. | AC-1.4, AC-2.1, AC-2.3, D-7 |
| T-9 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteWithNoBridgesHasEmptyDetail` *(new)* | No filed bridges → `bridges` empty. Unpromoted card → `beginUnpromote` leaves `pendingUnpromote` nil. | AC-2.2, AC-1.3 |
| T-10 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `confirmUnpromoteRefreshesToPromote` *(new)* | `confirmUnpromote()` → FakeStore recorded `"unpromoteSubject id=…"`; `pendingUnpromote` nil; `waitUntil` the graph handle shows James and the son-of bridge with nil membership and John still filed. Setting `pendingUnpromote = nil` without confirming records no call. | AC-2.4, AC-3.1, D-8, D-9 |
| T-11 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `unpromoteFailureStaysInConfirm` *(new)* | FakeStore throws → `pendingUnpromote` still set, `unpromoteError` non-nil, `isUnpromoting` false; reopening clears the error. | AC-3.6 |
| T-12 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` (extended) | Each §6 key resolves to its default value; `unpromoteConfirmTitle(name: "James Robins")` → "Unpromote James Robins?". | AC-2.1, D-16 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestSubjectDelete` and `TestSubjectDeleteReleasesOwnClaimPins` (STEP-1 extraction), `measuredLayoutMatchesThePaint` and `membershipSwapsPromoteForOpenHandle` (the promoted footer target is no longer full width; `footerTargetIsFullWidthAtTheBottom` covers only unpromoted cards and stays as it is), `TestResolversCoverEveryEntityType`.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: conclusion-layer-data-model.md §5.2 paragraph from STEP-9
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has James Robins, John Robins, and Mary Smith cards, a relationship "James Robins, son of John Robins", a relationship between John and Mary, and a participation of James in an event. Promote all, so every card shows a handle.
2. On James's card, hover the unlink square at the right of the footer. See the tooltip "Unpromote" and the small fill; the handle row does not highlight. Click the handle row: James's Person page opens. Go Back.
3. Click Unpromote. See "Unpromote James Robins?", the message, "ALSO UNFILED" with the son-of relationship and the participation, and "The other end of each stays filed." Press "Keep filing": every footer is unchanged.
4. Click Unpromote again and confirm. James's card, the son-of card, and the participation card show Promote. John, Mary, the event, and the John–Mary relationship still show handles. Every card and edge is in place. Open the People list: James's former handle `PER-…` is still listed.
5. Unpromote Mary (her only bridge, John–Mary, is filed): the confirm lists it. Re-promote James, then unpromote the event card from step 1 with no filed bridges left: the detail is empty.
6. With VoiceOver on, focus a filed card and open its actions: "Unpromote" is listed after the open-handle action and opens the same confirm.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Extracting `ReleaseClaimTx` changes Subject delete. | STEP-1 runs the Subject delete tests before anything else. |
| The client bridge list (D-7) and the engine set (D-3) disagree. | Both read accepted bridge claims whose end Observation names the Subject; T-1 and T-8 use the same fixture. A disagreement only follows a stale graph, and the write is authoritative. |
| The tooltip rect drifts from the painted slot. | It is the same `unpromoteActionID` region the paint and hits read. |

**Stop and report if:**

- `releaseSubjectClaims`, `bridgesTouching`, or `paintFooter` no longer does what §4 says.
- `promote` importing `deleteimpact` makes an import cycle (`go build ./core/...` fails with "import cycle not allowed").
- A filed bridge on this Source has this Subject as an end only through an Observation the client snapshot does not draw as an edge (`endpointAID` / `endpointBID`), so T-8 and T-1 cannot share a fixture.
- `NSImage(systemSymbolName: PVSymbol.unlink.rawValue)` is nil on the CI macOS image.
- `measuredLayoutMatchesThePaint` cannot pass without changing the unpromoted footer's full-width band (Out).
- Deleting the bridge claims leaves `autoreconciler` reporting a handle `writes.Run` did not recompute (a stale Person page name after step 4 of §10).

## 12. Open questions

- **Q-1. Which SF Symbol is the unlink icon?** The brief says "unlink icon" (Lucide `unlink`). `PVSymbol` maps to SF Symbols, and this plan was written where SF Symbols cannot be checked. Recommendation: the planner confirms on a Mac whether `link.slash` (or another slashed-link symbol) resolves with `NSImage(systemSymbolName:)`, and D-15 names it. If none exists, a custom glyph would be a new kit asset, which is a design call for the user.

## 13. PR

**Title:** S10-44 — Unpromote a filed card
**Body:** A promoted person, event, or place card on the Evidence graph can take its filing back. What shipped against US-1 (footer button), US-2 (confirm naming the bridges), US-3 (one audited delete of this claim and the bridge claims on this Source); AC evidence by test name; deviations from the plan; the manual check in §10. `Plan: docs/deployment-plan/spike-10/prs/S10-44-unpromote.md`.
