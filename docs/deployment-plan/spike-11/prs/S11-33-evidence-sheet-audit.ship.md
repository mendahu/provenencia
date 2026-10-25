# Ship log — S11-33 Evidence sheet audit follow-up

**Plan:** [`S11-33-evidence-sheet-audit.md`](S11-33-evidence-sheet-audit.md) · **Branch:** `s11-20-claim-evidence-sheet` (shared with S11-20, by the user's choice) · **PR:** #525
**State:** ready

## Round 0 — implement
Head `01922f4c` (STEP-1 `d9aff3e8` … STEP-12 `01922f4c`, warnings `7a11277e`). Gate PASS (orchestrator re-ran). Clean build shows no warning in changed files; also fixed 12 pre-existing warnings in `PVPanelFormDialogTests.swift`.
Deviations reported:
- D-13: strong emphasis does not show in a semibold display face, so the evidence heading's names no longer stand out from the semibold heading. This regresses S11-20 round 17 (UX-3). → for triage.
- `StyledFormat.attributed(literal:)` keeps "joins" muted.
- `PVSheet.Policy`.
- `EvidenceBoard.build(evidence:neighbor:labels:worded:)`.
- Shared views landed in STEP-7.
- T-4 is a mapping test.
- Extra dead code removed (`BridgeFact.kind`, three `EvidenceBridgeEdgeSummary` statics).
- Two plan greps cannot print nothing (existing `WorkspaceSidebar` comments; `LikenessTileView`).
- Manual check not run, including Escape with focus in the argument box.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. All four routes match: `macos/App/**`, strings, `api/ffi/**`, `core/database/promotealign`, `core/segmentcompare`, Features models. Scope `a860a37c...01922f4c`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | A click in a nested sheet counts as outside its parent: the Name form's split confirm shows the discard question or closes the form mid-click | Fix now → F-1 | verified by the reviewer's AppKit probe: `beginSheet` sets `sheetParent`, not `parent`; `PVOutsideClick.isInsideSheet` walks `parent` only. CON's "nested is inside" relied on the same `parent` walk. §11 stop condition 5 |
| CON-1, UX-2 | CON, UX | major | Evidence heading names no longer stand out by weight (regresses S11-20 UX-3/CON-4) | Fix now → F-2 | verified: `PVSectionHeader` title fixed semibold; `.stronglyEmphasized` cannot exceed it. Writer's call: restore the round-17 look with a documented kit rule (an `AttributedString` title draws regular; its emphasized runs carry the weight). That keeps D-13's "no `titleWeight`, no per-call font". Reported to the user |
| MNT-1 | MNT | blocker | Place-relationship shares are a second wording table (registry switch + four `share*Place` functions) | Fix now → F-3 | verified: `ConnectionSentenceRegistry.swift:173-183`; `wording` has label only for place terms |
| MNT-2 | MNT | major | `PVSheet` forwarding layers: `dismissRoute` ×2, `Policy`, three `*Policy` helpers | Fix now → F-4 | verified `PVConfirm.swift:214-215` builds a policy twice to unpack it |
| MNT-3 | MNT | major | `add-match-rule` still teaches Swift `SegmentCause` | Fix now → F-5 | verified `.agents/skills/add-match-rule/SKILL.md:75` |
| MNT-4 | MNT | major | Two `Member` types; each record worded twice | Fix now → F-6 | verified |
| MNT-5, PRF-1 | MNT, PRF | minor | `sentence(bridges:)` builds an index per call; Promote never passes one | Fix now → F-7 | D-11 said "built once" |
| MNT-6 | MNT | minor | Handler re-looks-up the bridge type for the term's property | Fix now → F-8 | core knowledge belongs in `promotealign` |
| MNT-7 | MNT | minor | Three writers of `flow.nameOrders`; view `onChange` hop | Fix now → F-9 | |
| MNT-8 | MNT | minor | Value-column switch written twice | Fix now → F-10 | |
| UX-4 | UX | minor | Hidden Escape button may be a Tab stop | Fix now → F-11 | `.focusable(false)` |
| CON-3 | CON | minor | T-6 tests a string stand-in, not `EvidenceLine ==` | Fix now → F-12 | |
| CON-2 | CON | major | Escape routing through the hidden button is unverified; no unit test can reach it | To user (manual check) | UI interaction; the four steps go in the report. Fallback if it fails: `.onExitCommand` |
| UX-3 | UX | minor | `LazyVStack` may stop Tab and VoiceOver at the last built card | To user (manual check) | low confidence; if it stops early, `VStack` (lines stay `.equatable()`) |

Nits (ride along in F-13):
- `MNT-9 · PromoteEvidenceDisplay.swift:4 · doc names removed bridge helpers`
- `MNT-10 · DesignSystem/README.md:204 · PVFormDialog presents through .pvSheet`
- `MNT-11 · PromoteReadout.swift:3-5 · doc lists the Updated caption`
- `MNT-12 · PVSectionHeader.swift:70 · preview indent`
- `UX-5 · L10n.swift:1576 · spokenRelationshipCount comment, parameter named link`
- `UX-6 · L10n.swift · "a event" in comments`
- `UX-7 · focus after Try again (logged, not fixed)`
- `CON-4 · plan "…" warnings are intended`

### Fix plan 1

**F-1 — A nested sheet is inside its parent** (from UX-1)
Files: `macos/App/DesignSystem/Components/Panel/PVOutsideClick.swift`, `macos/ProvenenciaTests/PVPanelFormDialogTests.swift`.
Failing test first: `aSheetOnASheetIsInside`. Two real windows with `beginSheet` (as the reviewer's probe did). A click in the child sheet's window is inside the parent sheet.
Do: `isInsideSheet` walks `window.sheetParent ?? window.parent` up the chain.
Check: test passes; check-pr PASS.

**F-2 — The evidence heading's names stand out again** (from CON-1, UX-2)
Files: `macos/App/DesignSystem/Components/SectionHeader/PVSectionHeader.swift`, `macos/App/DesignSystem/README.md`, `macos/ProvenenciaTests/TextEmphasisRenderingTests.swift` (or the existing rendering test file).
Failing test first: render `PVSectionHeader` with an `AttributedString` title that has one `.stronglyEmphasized` run; the emphasized run renders heavier than the rest (the same pixel comparison the existing tests use).
Do: a `PVSectionHeader` given an `AttributedString` title draws its base at regular weight, so emphasized runs carry the weight. A `String`/`LocalizedStringResource` title stays semibold. One documented rule; no `titleWeight` option. Update D-13's Landed-note line.
Check: test passes; check-pr PASS.

**F-3 — Place terms live in the one wording table** (from MNT-1)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Catalog/Connect/ConnectionSentenceRegistry.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/ConnectionSentenceRegistryTests.swift`.
Failing test first: the Go "every seeded term has wording" subtest also requires a share for every role, event type, relationship type and place relationship term. It fails on `part_of` / `succeeded_by` before the fix.
Do:
- `Wording` gains `shareOtherEnd` / `shareOtherEndSimilar`, the phrase used when the term names the row, not the neighbor (`namesNeighbor == false`).
- `part_of`: `share` "a place within it", `shareOtherEnd` "a parent place". `succeeded_by`: `share` "a predecessor place", `shareOtherEnd` "a successor place". Each has a similar form.
- The place closure reads them through the seeded lookup.
- Delete the switch and the four `share*Place` functions.
- Name every similar key `….similar` (place, location, connection and untyped-event keys included).
Check: T-5 still passes (same phrases); check-pr PASS.

**F-4 — One dismiss rule, no forwarders** (from MNT-2)
Files: `macos/App/DesignSystem/Components/Panel/PVSheet.swift`, `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`, `macos/App/DesignSystem/Components/Confirm/PVConfirm.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift`, `macos/ProvenenciaTests/PVPanelFormDialogTests.swift`.
Do: `PVSheetChrome.dismiss` calls `PVSheetDismiss.action` directly. Delete `PVSheet.dismissRoute` (both), `PVSheet.Policy`, `PVFormDialogControls.sheetPolicy`, `PVConfirmControls.sheetPolicy` and `PVDeleteImpactControls.noticePolicy`. Presenters pass the Bools (the notice passes none). Fold T-2 into T-1.
Check: `rg "dismissRoute|sheetPolicy|noticePolicy|PVSheet.Policy" macos` prints nothing; check-pr PASS.

**F-5 — `add-match-rule` matches the code** (from MNT-3)
Files: `.agents/skills/add-match-rule/SKILL.md`.
Do: drop "or a segment cause" and `SegmentCause` from line 75's paragraph. Say segment causes stay in Go; the sheet does not read them.
Check: `rg SegmentCause .agents` prints nothing.

**F-6 — One member shape** (from MNT-4)
Files: `macos/App/Features/Promote/{PromoteEvidenceBoard,PromoteEvidenceDisplay}.swift`, `macos/ProvenenciaTests/PromoteEvidenceDisplayTests.swift`.
Do: `spokenMember` and `spokenLabel` take `source:` and `value:` strings. `memberRow` and `lineRow` compute them once. Delete `PromoteEvidenceDisplay.Member`.
Check: check-pr PASS.

**F-7 — Bridge index built once** (from MNT-5, PRF-1)
Files: `macos/App/Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/EvidenceBridgeEdgeSummaryTests.swift`.
Do: `PromoteModel.bridgeFacts` skips the endpoint lookup: its bridges come from the snapshot itself. `sentence(for:in:…)` without `bridges:` falls back to `snapshot.bridges.first(where:)` only when an endpoint is nil, with no dictionary built per call. Callers that hold an index pass it.
Check: check-pr PASS.

**F-8 — The term's property comes from core** (from MNT-6)
Files: `core/claimevidence/claimevidence.go`, `core/database/promotealign/evidence.go`, `api/ffi/handlers/promote_evidence.go` and tests.
Do: `readFromNeighbor` records the term's property (`NeighborTermProperty`) on `Relationship`. The handler maps fields only; drop its `connectrules.LookupBridge` call.
Check: `go test -tags fts5 ./api/... ./core/...` PASS.

**F-9 — One writer of the name order** (from MNT-7)
Files: `macos/App/Features/Promote/{PromoteModel,PromoteView}.swift`.
Do: the board records the `NameFormatOrders` it was built with. `setSheetSubject` and `board(of:)` compare them with `session.nameFormat` and rebuild on a mismatch. Delete `nameFormatChanged()` and the view's `onChange`; `flow.nameOrders` is set in one place.
Check: an existing or new `PromoteModelTests` case: change `session.nameFormat`, open the sheet → the board's worded names use the new order.

**F-10 — One value view** (from MNT-8)
Files: `macos/App/Features/Promote/PromoteEvidenceCards.swift`.
Do: one private `EvidenceValueView(value:weight:projectDir:)` used by `EvidenceLine` and `EvidenceMemberRow`.

**F-11 — The hidden Escape button is not a Tab stop** (from UX-4)
Files: `macos/App/DesignSystem/Components/Panel/PVSheet.swift`.
Do: `.focusable(false)` on the hidden button.

**F-12 — Pin the line equality** (from CON-3)
Files: `macos/App/Features/Promote/PromoteEvidenceCards.swift` (`EvidenceLine` internal), `macos/ProvenenciaTests/PromoteEvidenceBoardTests.swift`.
Do: test `EvidenceLine ==` directly: same `line`, `checked` and `isOpen` → equal; a different `checked` or `isOpen` → unequal. Replace the string stand-in.

**F-13 — Nits** (MNT-9 – MNT-12, UX-5, UX-6)
Fix each at the line given above. `spokenRelationship*`'s parameter is renamed `outcome`.

Check for the round: scripts/check-pr.sh PASS; a fresh `-derivedDataPath` build shows no warning in a changed file.

## Round 1 — fix
Head `b41fca17`. Gate PASS (orchestrator re-ran). F-1, F-2 and F-3 failed first; F-8 and F-12 failed to compile first.

Deviations, all accepted:
- F-9: no recorded order on the board. Every open re-reads and words the evidence in the session's current order. A sheet already open when the format changes picks the new order up on its next open.
- F-2: the base weight is chosen by `title is AttributedString`.
- F-7: `bridges: [:]` skips the lookup.
- F-3: place keys moved under `placeRelationship.*`.

Also: `@MainActor` on `PromoteEvidenceBoardTests` for two new warnings.

PRF-1: **Resolved** (orchestrator read F-7).

## Round 2 — review
Reviewers: MNT, UX, CON (Sonnet), each of which had a Fix-now blocker or major. Scope `01922f4c...b41fca17`.
- **MNT:** MNT-1 – MNT-4 **Resolved**. Nit: `MNT-13 · EvidenceBridgeEdgeSummary.swift:30,42-54 · no app caller passes a real index; drop bridges(of:) and the parameter, or name the skip`.
- **UX:** UX-1 and UX-2 **Resolved**. Noted: plan D-9 still says an open sheet rewords on a name-order change; the Landed note records the actual behavior.
- **CON:** CON-1 and CON-3 **Resolved**. No regressions from F-3, F-4 or F-9.
- No new blockers.

## Finish
`main` unchanged. Gate PASS on `b41fca17`.

## Result
Gate PASS · CI see PR · open: none · to user: manual Escape check (CON-2) and long-sheet Tab/VoiceOver check (UX-3) · follow-ups: MNT-13 nit · CON miss: none
