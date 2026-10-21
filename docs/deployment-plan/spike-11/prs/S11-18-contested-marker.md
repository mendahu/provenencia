# S11-18 — Contested marker on Promote rows

**Status:** Blocked — on board S11-D1
**Spike / slice:** Spike 11 · Slice 4 — Segments and graph alignment
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S11-D1](../design/S11-D1-promote-rows.md) §4.1 (CM-1 – CM-6), board not agreed yet
**Board:** the S11-D1 Claude Design board, read through the Claude Design tools; its link is recorded here when the user agrees the board. If you cannot open it, §6 macOS describes the frames in words; build from §6
**Depends on:** S11-17 (landed [#504](https://github.com/mendahu/provenencia/pull/504)), S11-D1 (board)
**Blocks:** —
**Branch:** `s11-18-contested-marker`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [bands](../comparison-layers.md#bands) ("A conflict shows a contested marker; it never caps the band"); [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) §3.1; [S11-17 plan](S11-17-contested-ffi.md) D-4, D-7 (the fields this PR reads) and its Out lines; working rules "Bands come from evidence", "One wire vocabulary"
**Planned:** 2026-10-20 against `main` @ `9cf8e33c`
**Review:** [`S11-18-contested-marker.review.md`](S11-18-contested-marker.review.md) — round 1: Fits with changes / Not ready (blocked on board) → fixed; board still open

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

A Promote row whose comparisons include a conflict says so: a **Contested** badge sits beside its band, at any band. A strong match with a contradicting sex at birth no longer looks clean. When re-proposing moves a row because of a neighbor, the row's "Updated · …" caption names the neighbor and says what it did. This makes dogfood bars 3 ("updated because…") and 4 ("a conflicting row shows contested") true on the page; S11-16 computes both and S11-17 carries them to `PromoteFlow.Row`.

**Why now.** S11-17 landed `Row.contested` and `Row.move` with an interim, unreachable `reasonText` arm that waits for this PR's copy. Slice 4's Check ("a conflicting row shows contested") closes with this PR.

**What stays.** The row's columns and widths, the band badges and tones, the target menu, the Updated and Decided badges, the conflict-with-decided and duplicate callouts, the evidence sheet and its subtitle, the connections panel (S11-27 replaces it), grouping, Done and Cancel. No Go, proto or FFI change.

## 2. User stories

**US-1.** As a researcher, when a proposed record conflicts with my Source on something (a sex at birth, a spouse), I see **Contested** beside its band, so a strong band does not hide the conflict.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A row proposed on PER-1, band Strong, `contested = true` | The assessment column shows the Strong badge, then the Contested badge (danger tone, warning icon), then the reason text (CM-1) |
| AC-1.2 | The same row with `contested = false` | Strong badge and reason only; no Contested badge (CM-2) |
| AC-1.3 | A contested row at band Weak (target drafted on its first alternative) | Weak badge, then Contested (CM-1) |
| AC-1.4 | A row with `contested = true` whose target is New, Skip or unset, or an anchor row | No assessment column and no Contested badge (CM-2) |
| AC-1.5 | A contested row on PER-1; the researcher picks PER-2 from the menu | The Contested badge goes at once; after the re-proposal it shows again only if the proposal says PER-2 is contested (D-2) |
| AC-1.6 | VoiceOver on the assessment button of a contested Strong row "Mary Smith" whose reason is "via John Smith → PER-7QH3D (son)" | Reads "Evidence for Mary Smith: Strong, contested, via John Smith → PER-7QH3D (son)"; the clean row reads "Evidence for Mary Smith: Strong, via John Smith → PER-7QH3D (son)" (CM-4) |

**US-2.** As a researcher, when deciding one row moves another, the moved row's Updated caption names the neighbor that moved it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Row "Bea" is Updated with `move = .neighborPlaced(neighborID: "a")`; row "a" is "Ada" on handle PER-e1 | Caption "Updated · after Ada → PER-e1 was matched" (CM-5) |
| AC-2.2 | Same, `move = .neighborMoved(neighborID: "a")` | "Updated · after Ada moved to PER-e1" (CM-5) |
| AC-2.3 | Same, `move = .neighborReleased(neighborID: "a")` | "Updated · after Ada was unmatched" (CM-5) |
| AC-2.4 | A move whose neighbor ID is not a row on the page | "Updated · after a connected row changed" (D-5) |
| AC-2.5 | Row Updated with `move = nil`, reason via Ada | Caption as today: "Updated · via Ada → PER-e1" (CM-6) |
| AC-2.6 | Row not Updated with a move | No caption (CM-6; today's `if row.updated` guard) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Contested badge in the assessment column, only with a selection readout (CM-1, CM-2, CM-3) | D-1, D-3, STEP-1, STEP-4, AC-1.1 – AC-1.4 |
| Changing the target clears contested and move until the next proposal | D-2, STEP-1, AC-1.5 |
| VoiceOver label with "contested" (CM-4) | D-4, STEP-2, STEP-3, STEP-4, AC-1.6 |
| The three move sentences and their fallback, naming the neighbor as `.via` does (CM-5) | D-5, D-6, STEP-2, STEP-3, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| The Updated caption reads the move when there is one (CM-5, CM-6) | D-7, STEP-1, STEP-4, AC-2.5, AC-2.6 |
| `promote-graph-alignment.md` §3.1 says both | STEP-5 |

| Out | Home |
| --- | --- |
| Connection subrows, the "+N connections" pointer, removing the panel (S11-D1 CS-1 – CS-9) | S11-27 |
| Contested or the move in the evidence sheet's subtitle, and which comparison conflicts | S11-20 (behind S11-D2) |
| A contested mark per alternative in the target menu | Not planned — the proposal carries `contested` for the row's own pair only (S11-D1 §2.1) |
| A move shown on the first proposal, or on a row that is not Updated | Not planned — user decision 2026-10-20 (S11-D1 "Where a move shows") |
| `FakeStore` producing contested rows or moves | Not planned — tests set the fields on rows, as S11-17's tests do |
| Any Go, proto or handler change | Not planned — S11-17 sends everything this PR reads |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row fields | `macos/App/Features/Promote/PromoteFlow.swift` `PromoteFlow.Row.contested`, `.move` | Set by `makeRow` from the proposal (S11-17 D-7): fresh rows and decided-keep rows take `proposal.contested` and `Reason(move:)`; anchor rows keep the defaults. Nothing reads them |
| Readout | `PromoteFlow.swift` `PromoteFlow.SelectionReadout` (`assessment`, `reason`) and `Row.selectionReadout` | Non-nil only when `target` is `.handle` and that handle is in `menu`; Skip, New, unset give nil. Carries no contested |
| Draft target | `PromoteFlow.swift` `draftTarget` | Band none → `.unset`; proposal handle (and first in the menu) → that handle; otherwise `taken` → `.unset`; Weak → `.handle(menu.first)` (the first alternative, the pair Go's contested describes for a Skip row); otherwise New or Skip. So the readout always shows the pair `contested` describes |
| Retarget | `PromoteFlow.swift` `PromoteFlow.setTarget(subjectID:token:)` | On a target change clears `comparisons`, `pins` and the drafted argument ("The lines and pins compared the old target"), marks decided, clears `updated`. Leaves `contested` and `move` describing the old pair until the re-proposal `PromoteModel.setTarget` starts |
| Move reasons | `PromoteFlow.swift` `PromoteFlow.Reason` `.neighborPlaced/.neighborMoved/.neighborReleased(neighborID:)` | From `init?(move:)`; an empty neighbor ID or unknown code gives `.empty` |
| Wording | `macos/App/Features/Promote/PromoteModel.swift` `PromoteModel.reasonText(_:on:)` | One switch over `Reason`. `.via` looks the neighbor up in `flow.rows`, names it with its handle ref (or its Source ref), adds the bridge phrase; falls back to `L10n.Promote.reasonVia` when the neighbor is not a row. The three move cases hit the interim arm `return L10n.string(L10n.Promote.reasonEmpty)` with the comment "Worded in S11-18; no row reason is a move reason." |
| Caption reason | `PromoteModel.swift` `PromoteModel.reasonText(for:)` | `reasonText(row.reason, on: row)`. Its only app caller is the Updated caption |
| Row view | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Assessment column: when `row.selectionReadout` is set, a plain `Button` (opens the sheet) holding `PVBadge(band)`, the reason `Text` (`lineLimit(1)`), a chevron; `.accessibilityLabel(L10n.Promote.evidenceFor(name:assessment:reason:))`, identifier `promote.row.<ref>.evidence`. Caption: `if row.updated { Text(verbatim: L10n.Promote.updatedDetail(model.reasonText(for: row))) }` |
| Sheet subtitle | `PromoteView.swift` `PromoteEvidenceSheet.sheetSubtitle` | `L10n.Promote.sheetSubtitle(assessment:reason:)` from the readout. Stays as it is (Out) |
| Badge | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` `PVBadge.init(_:tone:icon:subtle:)`, `PVBadgeTone.danger` | Text badge with optional icon; body ends in `.fixedSize()`, so a badge never truncates. `PVSymbol.warning` is `exclamationmark.triangle.fill` (`DesignSystem/Components/Icon/PVIcon.swift`) |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Promote` (`updatedDetail`, `evidenceFor`, `via`, `viaNeighbor`, `reasonVia`, `reasonEmpty`); `macos/App/Resources/Localizable.xcstrings` | No contested or move strings |
| Tests | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aContestedRowAndItsMoveReachTheFlow`, `aDecidedRowTakesTheLatestContestedFlag`, `anAnchorRowIsNeverContested`; `macos/ProvenenciaTests/PromoteModelTests.swift` `reasonsAreWordedWithLabelsAndRefs` (calls `reasonText(for:)`) | Parsing and flow only; no wording or readout tests for contested or moves |

**Precedent.** The Updated and Decided marks in `PromoteAlignmentRow` (S9-44) are `PVBadge`s with a tone and icon; copy that shape. `setTarget` clearing `comparisons` on a retarget is the precedent for D-2. `.via`'s neighbor lookup in `reasonText` is the precedent for the move sentences: D-6 moves it into one helper both use, as S11-17 review M-1 planned ("the move cases reuse `.via`'s neighbor naming").

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **The readout carries contested.** `PromoteFlow.SelectionReadout` gains `var contested = false`, declared after `reason` so the existing `SelectionReadout(assessment:reason:)` literals in `PromoteFlowTests` keep compiling (doc "The proposal says this pair conflicts somewhere; the page marks it beside the band"). `Row.selectionReadout` sets it from `row.contested`. The view reads `readout.contested`; nothing else reads `row.contested` | The marker shows exactly when the band shows (user, 2026-10-20: "only with a handle"); the readout is the one place that decides that, and `draftTarget` makes the readout's handle the pair Go's contested describes (§4). One computed value is testable in `PromoteFlowTests` without a view | A `row.contested && row.selectionReadout != nil` check in the view — the rule would live in a view and go untested; showing it on Skip rows — user chose not to |
| D-2 | **A retarget clears contested and move.** In `PromoteFlow.setTarget`, inside `if next != rows[index].target`, beside the `comparisons` and `pins` resets, set `rows[index].contested = false` and `rows[index].move = nil`. Change the comment to: "The lines, pins and contested flag compared the old target; the next proposal drafts them for this one." | `contested` describes the old pair once the target changes; the re-proposal refreshes it (decided-keep path, S11-17 D-7), as it re-drafts the lines | Leaving the old flag until the re-proposal — the badge would sit on a record it does not describe |
| D-3 | **Marker.** In `PromoteAlignmentRow`'s assessment `Button` label, right after the band `PVBadge`: `if readout.contested { PVBadge(L10n.Promote.contested, tone: .danger, icon: .warning).accessibilityIdentifier("promote.row.\(row.ref).contested") }`. No other layout change: `PVBadge` is `fixedSize`, and the reason `Text` is already `lineLimit(1)`, so the reason truncates first (CM-3) | Brief inventory: "Contested mark — Badge, not a new control and not a Chip"; danger keeps it apart from the Weak band's warning tone | A glyph-only badge (`init(icon:label:)`) — CM-1 draws text; a Callout under the row — the brief puts it on the assessment line |
| D-4 | **VoiceOver: contested changes the spoken band, not the sentence.** `L10n.Promote.assessment(_ a: PromoteFlow.Assessment) -> LocalizedStringResource` *(new)*, beside `L10n.Promote.outcome(_:)`, switches to the existing `assessmentStrong` / `assessmentMedium` / `assessmentWeak` / `assessmentNone`; it replaces `private enum PromoteAssessmentCopy` in `PromoteView.swift`, whose four call sites move onto it and which is deleted. `PromoteModel.assessmentText(_ readout: PromoteFlow.SelectionReadout) -> String` *(new)*, beside `reasonText`, returns `L10n.Promote.assessmentContested(L10n.string(L10n.Promote.assessment(readout.assessment)))` *(new key, "%@, contested")* when `readout.contested`, else `L10n.string(L10n.Promote.assessment(readout.assessment))`. The row button's label stays `L10n.Promote.evidenceFor(name:assessment:reason:)`, now with `assessment: model.assessmentText(readout)`. The sheet subtitle keeps the plain band (Out). The Contested badge gets no accessibility modifier: an explicit `.accessibilityLabel` on the button replaces its children's, as the band badge relies on today | One place decides how a contested band is spoken; S11-20 switches `sheetSubtitle` to `assessmentText` with no new key (maintainability review M-2). The band wording moves out of a file-private view enum the model and tests cannot reach (M-1); `outcome(_:)` is the precedent | A second full sentence `evidenceForContested` — S11-20 would double `sheetSubtitle` the same way; appending ", contested" in the view — word order is the translator's, and untestable |
| D-5 | **Move sentences.** `reasonText(_:on:)` replaces the interim arm and its comment with three arms. With the neighbor found (D-6): `.neighborPlaced` → `L10n.Promote.neighborPlaced(neighbor:ref:)` *(new)* "after %1$@ → %2$@ was matched"; `.neighborMoved` → `L10n.Promote.neighborMoved(neighbor:ref:)` *(new)* "after %1$@ moved to %2$@"; `.neighborReleased` → `L10n.Promote.neighborReleased(neighbor:)` *(new)* "after %@ was unmatched". Neighbor not a row on the page: all three → `L10n.Promote.reasonMoveUnknown` *(new)* "after a connected row changed". No bridge phrase | Brief CM-5 copy. The sentences fill `updatedDetail`'s `%@`, giving "Updated · after Ada moved to PER-e1". A move neighbor is a layer Subject, so it is a row; the fallback mirrors `.via`'s `reasonVia` for the case it is not | Reusing the via phrase with the role — the brief's sentences have none; a separate move switch — S11-17 D-4 put moves in `Reason` so there is one switch |
| D-6 | **A row's shown ref lives in the flow.** `PromoteFlow.Target.handleRef: String?` *(new)*, beside `handleID` (`if case .handle(_, let ref, _) = self { return ref }`). `PromoteFlow.Row.shownRef: String` *(new)*, doc "The ref the page names this row by: its record's when it has one, else its Source ref", `target.handleRef ?? ref`. In `reasonText`, `.via` and D-5's arms find the neighbor with `flow.rows.first { $0.subjectID == neighborID }` and name it by `.name` and `.shownRef`; `.via` keeps its phrase lookup and its `reasonVia` fallback. The other `if case .handle(_, let ref, _)` sites (`noteDuplicates`, `targetColumn`) are not touched | The "which ref names a row" rule is written once, in the flow, where S11-27 builds subrow sentences that name ends the same way (review M-3) | A private `neighborNaming` tuple in `PromoteModel` — S11-27 would copy it into the flow; copying the lookup into each move arm |
| D-7 | **Caption reason in the flow.** `PromoteFlow.Row.captionReason: Reason` *(new)*, beside `selectionReadout`, doc "What the Updated caption words: the move when settling moved this row, else its reason", `move ?? reason`. `PromoteModel.reasonText(for:)` is deleted; the caption calls `model.reasonText(row.captionReason, on: row)`; `if row.updated` stays the guard | User, 2026-10-20: the move shows in the Updated line, replacing the reason, and only there (CM-5, CM-6). Both "which reason shows" rules sit in `PromoteFlow`, testable there, and the model keeps one wording function (review M-4) | Renaming the forwarder to `updatedReasonText(for:)` — a second wording entry point; a second caption line — user chose the Updated line |
| D-8 | **Copy comes from the agreed board.** §6 Strings is the brief's copy. If the agreed board words CM-1, CM-4 or CM-5 differently, the plan refresh (before `Planned`) copies the board's words into §6 and AC-1.6 / AC-2.x | The brief is not the board; the board is agreed before this plan leaves Blocked | Implementing against the brief's copy when the board differs |
| D-9 | **Tests are Swift Testing on flow and model**, in the existing suites; no view or UI test. Rows get `contested` and `moveReason` set directly on `CatalogPromoteGraphAlignmentRow`, as S11-17's tests do | Every rule in this PR is in `PromoteFlow`, `PromoteModel` or `L10n` (D-1, D-2, D-4, D-5, D-6, D-7); the view only places a badge | A `FakeStore` knob for contested — S11-17 Out |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SelectionReadout.contested` *(new)* | field | `SelectionReadout` (`assessment`, `reason`) | Extends it (D-1) |
| `L10n.Promote.assessment(_:)` *(new)* | func | `private enum PromoteAssessmentCopy` (`PromoteView.swift`), `L10n.Promote.outcome(_:)` | Replaces the private enum with the `outcome(_:)` shape so the model can reach it (D-4) |
| `PromoteModel.assessmentText(_:)` *(new)* | func | the `assessment:` argument built inline in `PromoteAlignmentRow` | One spoken band for the row now and the sheet in S11-20 (D-4) |
| `Target.handleRef`, `Row.shownRef` *(new)* | computed props | `Target.handleID`; the inline `if case .handle(_, let ref, _)` in `reasonText` | Extends `Target` beside `handleID`; one rule for the ref a row is named by (D-6) |
| `Row.captionReason` *(new)* | computed prop | `Row.selectionReadout`; `PromoteModel.reasonText(for:)` | Beside the readout; replaces the forwarder (D-7) |
| `L10n.Promote.contested`, `assessmentContested`, `neighborPlaced`, `neighborMoved`, `neighborReleased`, `reasonMoveUnknown` *(new)* | L10n keys | `L10n.Promote.updated`, `evidenceFor`, `viaNeighbor`, `reasonVia` | Same namespace, same shapes; searched `rg -n "contested\|moved\|unmatched" macos/App/Platform/L10n.swift` — no existing copy |

**Builds on.** `PVBadge` (kit), `PromoteFlow.Reason`'s move cases and `PromoteReasonCode` (S11-17), `reasonText`'s one switch, `add-localized-string`.
**Deletes.** The interim arm and its comment `// Worded in S11-18; no row reason is a move reason.` in `PromoteModel.reasonText(_:on:)`; `PromoteModel.reasonText(for:)`; `private enum PromoteAssessmentCopy` in `PromoteView.swift`.
**Next adds.** S11-27 adds subrows and the "+N connections" Button to `PromoteAlignmentRow` (S11-D1 CS-1 – CS-9) and names ends by `Row.shownRef`. S11-20 speaks the sheet's band through `assessmentText` and names neighbor groups by `shownRef` (noted on its scope). A new move reason is one `Reason` case the compiler asks `reasonText` to word, plus one string. Logged, not done: `noteDuplicates` and `targetColumn` could read `Target.handleRef`.

### Schema

None — no catalog change.

### Go core

None — S11-16 computes contested and moves; S11-17 sends them.

### FFI

None — `PromoteGraphAlignmentRow.contested`, `move_reason`, `move_neighbor_subject_id` exist since S11-17.

### macOS

From the S11-D1 inventory (§6 of the brief), the rows this PR touches:

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Promote row | snowflake, Extend | `macos/App/Features/Promote/PromoteView.swift` `PromoteAlignmentRow` | Contested badge in the assessment button (D-3); label's `assessment:` from `model.assessmentText` (D-4); caption from `row.captionReason` (D-7) |
| Contested mark | `PVBadge(_:tone:icon:)`, `.danger`, `.warning` | `macos/App/DesignSystem/Components/Badge/PVBadge.swift` | Instance; no new kit piece |
| Band, Updated, Decided marks | `PVBadge` | same | Unchanged |
| Updated caption | snowflake, Extend | `PromoteAlignmentRow` | Same type, indent and guard |

Frames in words (S11-D1 §5 frames 1 and 2): a row's assessment column reads, left to right, the band badge, the Contested badge (red tint, triangle icon, caps "CONTESTED"), the italic reason, a chevron. A clean row has no second badge. Mary's Updated row has the blue Updated badge in the badge column and, under the row at the note indent, the italic caption "Updated · after John Smith moved to PER-7QH3D".

States: no empty, loading or error state is added; while a re-proposal is in flight a retargeted row shows no Contested badge (D-2). Large: the badge is fixed-size; the reason truncates. Accessibility: D-4; the badge's identifier `promote.row.<ref>.contested`.

### Strings

L10n keys *(new)* in `L10n.Promote`, each with an entry in `Localizable.xcstrings` (`extractionState` `manual`, `en` translated, as `promote.row.updatedDetail`):

| Swift | Key | English | Comment |
| --- | --- | --- | --- |
| `contested` (`LocalizedStringResource`) | `promote.row.contested` | Contested | Badge beside a Promote row's band when a comparison behind its pair conflicts |
| `assessmentContested(_:)` | `promote.row.assessmentContested` | %@, contested | Spoken band of a contested Promote row, inside the evidence button label; argument is the band name ("Strong") |
| `neighborPlaced(neighbor:ref:)` | `promote.reason.neighborPlaced` | after %1$@ → %2$@ was matched | Why a row moved on the last proposal: a neighbor was matched; 1 = neighbor name, 2 = its ref. Follows "Updated · " |
| `neighborMoved(neighbor:ref:)` | `promote.reason.neighborMoved` | after %1$@ moved to %2$@ | Why a row moved: a neighbor was matched to another record; 1 = neighbor name, 2 = its new ref. Follows "Updated · " |
| `neighborReleased(neighbor:)` | `promote.reason.neighborReleased` | after %@ was unmatched | Why a row moved: a neighbor lost its match; argument is the neighbor name. Follows "Updated · " |
| `reasonMoveUnknown` (`LocalizedStringResource`) | `promote.reason.moveUnknown` | after a connected row changed | Why a row moved, when the neighbor is not on the page. Follows "Updated · " |

Removed: none. `reasonEmpty` stays (it words `.empty`).

## 7. Steps

The implementer evidences every step with XT and `scripts/check-pr.sh`. §10's manual checks need a person at the app; the implementer lists them as "for the user" and does not tick them.

Xcode check, used below as **XT**: `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/<Suite> CODE_SIGNING_ALLOWED=NO` (from `scripts/check-pr.sh`).

**STEP-1 — Readout, retarget, shown ref and caption reason**
Files: `macos/App/Features/Promote/PromoteFlow.swift`, `macos/ProvenenciaTests/PromoteFlowTests.swift`. Skill: `add-swift-test`.
Do: D-1, D-2, and the flow halves of D-6 (`Target.handleRef`, `Row.shownRef`) and D-7 (`Row.captionReason`). Add T-1, T-2, T-3, T-5.
Check: XT with `PromoteFlowTests` passes.

**STEP-2 — Strings**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: the six keys in §6 Strings, beside `updatedDetail` and `evidenceFor` in `L10n.Promote`, and `L10n.Promote.assessment(_:)` (D-4) beside `outcome(_:)`. In `Localizable.xcstrings`, insert each entry where the file's existing key order puts it.
Check: the app builds (XT with `PromoteModelTests` passes); `rg -c '"promote\.(row\.contested|row\.assessmentContested|reason\.neighbor(Placed|Moved|Released)|reason\.moveUnknown)"' macos/App/Resources/Localizable.xcstrings` prints 6; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — Wording**
Files: `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.
Do: D-4's `assessmentText`, D-5, D-6's use in `reasonText`, and delete `reasonText(for:)` (D-7). In `reasonsAreWordedWithLabelsAndRefs`, `model.reasonText(for: x)` becomes `model.reasonText(x.captionReason, on: x)`. Add T-4, T-6.
Check: XT with `PromoteModelTests` passes; `rg -n "Worded in S11-18" macos/App` prints nothing.

**STEP-4 — Row view**
Files: `macos/App/Features/Promote/PromoteView.swift`. Skill: `add-ui-component`.
Do: D-3 in `PromoteAlignmentRow`; in its `.accessibilityLabel`, pass `assessment: model.assessmentText(readout)` to `evidenceFor`; the caption calls `model.reasonText(row.captionReason, on: row)`. Move the four `PromoteAssessmentCopy.title(x)` call sites (section titles, the band badge, the label's old argument, `sheetSubtitle`) onto `L10n.Promote.assessment(x)` and delete `PromoteAssessmentCopy` (D-4). Nothing else in the file changes.
Check: XT with `PromoteModelTests` and `PromoteFlowTests` passes (the app builds); `rg -n "reasonText\(for:|PromoteAssessmentCopy" macos` prints nothing.

**STEP-5 — Docs**
Files: `docs/promote-graph-alignment.md` §3.1, `docs/deployment-plan/spike-11/completed.md` (a `## S11-18 — Contested marker on Promote rows` section above S11-17's, in its shape: **Landed** date and PR link, then one paragraph), `deployment-plan.md` (map row), this plan (status).
Do: in §3.1, replace the bullet "**Rows whose suggestion changed** get a brief "updated" mark with the reason ("because Gracie → PER-Z")." with "**Rows whose suggestion changed** get a brief "updated" mark with the reason ("via Gracie → PER-Z"), or, when settling moved the row because of a neighbor, what that neighbor did ("after Gracie moved to PER-Z")." and add after it "- **A contested row** (a comparison behind its pair conflicts) shows a Contested mark beside its band, at any band; the band is unchanged." The rest per the Definition of done. Do **not** move the S11-D1 brief: it also gates S11-27, which archives it.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-18-contested-marker.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `aContestedRowShowsContestedInItsReadout` *(new)* | Build with the file's `row(id:target:handleID:assessment:pinned:)` and set fields after the call. `a = row(id: "a")`, `a.contested = true`; `b = row(id: "b", handleID: "e3")`, `b.handleRef = "PER-3"` (clean); `c = row(id: "c", target: "skip", handleID: "", assessment: "weak")`, `c.contested = true`, `c.alternatives[0].assessment = "weak"` (its alternative is the helper's `e2`). `load` with facts a, b, c. Expect a's `selectionReadout?.contested == true`, b's `== false`, c's readout `== PromoteFlow.SelectionReadout(assessment: .weak, reason: .empty, contested: true)` (the alternative's reason defaults to `.empty`), with `c.target.handleID == "e2"` | AC-1.1, AC-1.2, AC-1.3, D-1 |
| T-2 | Swift | same file · `aContestedRowWithoutAHandleShowsNoReadout` *(new)* | Each row has `contested = true`, set after `row(...)`: `row(id: "n", target: "new", handleID: "")` (drafts New); `row(id: "t", target: "skip", handleID: "", assessment: "medium")` with `reason = .taken` (drafts unset); `row(id: "z", target: "skip", handleID: "", assessment: "none")` (drafts unset); `row(id: "x")` loaded with `fact("x", anchor: true)`. Expect `selectionReadout == nil` for each | AC-1.4, D-1 |
| T-3 | Swift | same file · `retargetingClearsContestedAndMoveUntilTheNextProposal` *(new)* | Load a (handle e1, `contested = true`, `moveReason = .neighborMoved`, neighbor b) and b; `mapRest()`; `setTarget(subjectID: "a", token: "handle:e2")` (the helper's alternative, so the target changes) → a's `contested == false`, `move == nil`. Then `merge` a proposal with `row(id: "a", handleID: "e2")`, `contested = true` → a's `target.handleID == "e2"` and `contested == true` | AC-1.5, D-2 |
| T-4 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `movesAreWordedNamingTheNeighbor` *(new)* | Subjects Ada (a, handle e1) and Bea (b, handle e2); after `load()`: `reasonText(.neighborPlaced(neighborID: "a"), on: bea) == L10n.Promote.neighborPlaced(neighbor: "Ada", ref: "PER-e1")`; `.neighborMoved` → `neighborMoved(neighbor: "Ada", ref: "PER-e1")`; `.neighborReleased` → `neighborReleased(neighbor: "Ada")`; `.neighborMoved(neighborID: "zz")` → `L10n.string(L10n.Promote.reasonMoveUnknown)` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5, D-6 |
| T-5 | Swift | `macos/ProvenenciaTests/PromoteFlowTests.swift` `theCaptionReasonIsTheMoveWhenThereIsOne` *(new)* | `a = row(id: "a")` with `moveReason = .neighborMoved`, `moveNeighborSubjectID = "b"`; `b = row(id: "b", handleID: "e3")`. After `load`: a's `captionReason == .neighborMoved(neighborID: "b")`; b's `captionReason == b.reason`. Shown ref: a's `shownRef == "PER-1"` (its handle ref); after `setTarget(subjectID: "a", token: "skip")`, a's `shownRef == "CPR-a"` | AC-2.5, D-6, D-7 |
| T-6 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `theContestedBandIsSpoken` *(new)* | Ada on e1 strong with `contested = true`: `assessmentText(ada.selectionReadout!) == L10n.Promote.assessmentContested(L10n.string(L10n.Promote.assessment(.strong)))`; with `contested = false`: `== L10n.string(L10n.Promote.assessment(.strong))` | AC-1.6, D-4 |

AC-2.6 is today's `if row.updated` guard, unchanged (STEP-4) and checked in §10 step 3. AC-1.1's order and tones are checked in §10 step 1.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PromoteModelTests` (the deleted forwarder) and `PromoteFlowTests` (the readout's new field in any `SelectionReadout(…)` literal; `rg -n "SelectionReadout(" macos` finds them).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump: none (header)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief stays in design/ (it gates S11-27 too); design/README.md unchanged
- [ ] Domain doc updated: promote-graph-alignment.md §3.1 (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a filed person "Mary Smith" (sex at birth Male). Add a Source with a person "Mary Smith" whose sex at birth is recorded as Female. Open Promote on that Subject. See: the row on Mary Smith's record with its band, then a red **Contested** badge, then the reason. The other rows show no Contested badge.
2. On that row, pick New. See: the assessment column and the badge go. Pick Mary Smith's record again. See: after the re-proposal, the Contested badge is back.
3. On any page, decide a row so another suggested row changes target. See: the changed row shows Updated and a caption "Updated · …". With the default registry a settling move is rare, so the caption usually reads the via or agrees reason; T-5 is the evidence for the move sentence, named in the PR body.
4. With VoiceOver on, focus the contested row's assessment. Hear "Evidence for Mary Smith: Strong, contested, …" (or its band).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The agreed board words or places the marker differently from the brief | D-8: the plan refresh copies the board before `Planned` |
| A row shows Contested on a record the proposal did not compare | D-1 ties it to the readout; T-1 covers the Weak draft-on-alternative case |

**Stop and report if:**

- `draftTarget` drafts a handle other than the proposal's handle or its first alternative for a row Go marks contested (the readout would describe a pair `contested` does not).
- `Row.contested` or `Row.move` is not set as §4 says (`makeRow`'s fresh and decided-keep paths).
- The agreed board puts the marker outside the assessment button, or shows a move outside the Updated caption (D-3, D-7 would not hold).
- Any step needs a Go, proto or `FakeStore` change.
- XT or `scripts/check-pr.sh` cannot run in your environment: report it; do not mark a step's check done without its output.

## 12. Open questions

1. **Is the S11-D1 board agreed?** Recommended: draw it in Claude Design from [`S11-D1-promote-rows.md`](../design/S11-D1-promote-rows.md), then refresh this plan (D-8), record the board link, and run `review-plan`.

## 13. PR

**Title:** S11-18 — Contested marker on Promote rows
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-18-contested-marker.md`.
