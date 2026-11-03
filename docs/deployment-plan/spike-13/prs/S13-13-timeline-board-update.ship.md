# Ship log — S13-13 Timeline section: the board's 2026-11-03 update

**Plan:** [`S13-13-timeline-board-update.md`](S13-13-timeline-board-update.md) · **Branch:** `s13-13-timeline-section` · **PR:** #691
**State:** ready

Base: branch @ `905d2083` (PR #691, ready for review, CI green on `d87045a1`), on `main` @ `92f87dba`. The user changed the board on 2026-11-03 and asked for the update on the same branch; plan Planned after two review rounds (user answers: age in Go, `.compact` reshaped, 29 February → 1 March). The first ship log, [`S13-13-timeline-section.ship.md`](S13-13-timeline-section.ship.md), covers the section as first built.

## Round 0 — implement
Head `ef3d0259`. Gate PASS (orchestrator run against `92f87dba`: all nine checks). Report deviations: (1) `PersonTimelinePlacement.lifeline` is optional, nil until every item has a height; (2) the flag font is `PVTypeScale.micro` (11); (3) `Row.age`, `Row.ageHelp` and one `ageText(_:)` rule; (4) `PersonTimelineAgeFlag(isDashed:)` from "undated row"; (5) the curve ends at the flag's leading edge (x = 96) with no end dot; (6) docs as STEP-7; (7) nothing outside the plan. Not run in the app: click-to-scroll over `PersonTimelineStack`, the frames, tethers on scroll, Wide resize.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, strings. SEC+PRF: `core/datevalue` and `core/valuecompare` (matching's leap-day windows via `ShiftYears`, `core/subjectprofile`), `api/ffi/**`, `core/database/**`, feature models and list layout.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Unfolding a fold rescales the line; rows above move | Fix now → F-3 | verified: `PersonTimelinePlacement` `base` reads the sum of all heights; a participant unfold changes it; nothing anchors the scroll |
| UX-2 | UX | major | A narrow Regular pane crushes the title beside the 300pt people column | Fix now → F-4 | verified arithmetic: 492pt fixed per row; a 700pt pane leaves about 52pt of text |
| MNT-1 (+ MNT-OL) | MNT | major | The 12pt line gap in three places; tether end x twice; the text cap omits it | Fix now → F-1 | `PersonTimelineList.lineGap` private; bare `12` in `PersonTimelineLifeline.tethers` and `canvasWidth` |
| MNT-2 (+ CON-2) | MNT | major | Two stacking rules for unmeasured items; an unfold drops the line for a frame | Fix now → F-2 | `PersonTimelinePlacement.stack` read only by tests; `PersonTimelineStack` stacks again; view keys on `lifeline == nil` |
| CON-1 | CON | major | Manual checks and frames unrun; stop condition 2 rests on them | To user (manual check) | ship-pr: the user does the manual check; flagged in the report |
| MNT-3 | MNT | minor | `eventDating` twice per row; the birth placed per row | Fix now → F-5 | `order.go` `EventSpan`, `EventAge` |
| MNT-4 / UX-6 | MNT, UX | minor | The age-range translator comment says "by a year" | Fix now → F-6 | ranges reach 26–32 |
| CON-OL (→ UX) | CON | minor | VoiceOver reads "age 41–42" with a dash | Fix now → F-6 | `timelineAgeSpoken` |
| CON-3 | CON | minor | `add-date-value` skill does not list `CompletedYears` or the 1 March rule | Fix now → F-7 | `.agents/skills/add-date-value/SKILL.md:22` |
| CON-4 | CON | minor | `TestShiftYears` has no 29 February row | Fix now → F-7 | `core/valuecompare/datewindow_test.go` |
| UX-3 | UX | minor | A truncated role on the compact card has no tooltip | Fix now → F-8 | `PVPersonCardContent.helpText` has no role |
| UX-4 | UX | minor | A tall row loses its curve while it fills the viewport | Fix now → F-8 | `tethers` keeps only anchors in view |
| UX-5 | UX | minor | Wide → Regular drops focus held on cards 6–12 | Fix now → F-3 | rule 5: crossing keeps focus |

SEC+PRF: clean (no findings; 2,000,000-pair probe of `AddYears`/`CompletedYears`; matching windows move a day at 29 February edges, nothing stored).

Nits (logged): `MNT-5 · PVCard padding: init could forward to insets:` · `MNT-6 · ageHelp re-calls ageText` · `MNT-7 · README says PVTileCardSize owns every measure; recipe owns type sizes` · `MNT-8 · kept! twice in the scale search` · `MNT-9 · the 8pt open-bar inset is a bare literal` · `CON-5 · S13-35 Ships text still says 52pt`

### Fix plan 1

**F-1 — One geometry constant for the line gap and the flag's x** (from MNT-1, MNT-OL)
Files: `macos/App/Features/Conclusions/PersonTimelineLifeline.swift`, `PersonTimelineSection.swift`, `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift`.
Do: `PersonTimelineLifeline.listGap: CGFloat = 12` and `static var flagX: CGFloat { columnWidth + listGap + PersonTimelinePlacement.flagInset }`; `tethers`, `PersonTimelineTethers.canvasWidth` and the list's leading spacer read them; the private `lineGap` goes; `PersonTimelineEventRow`'s text cap subtracts `listGap` too; the test asserts `flagX`, not 96.
Check: `rg -n "lineGap|\+ 12 \+|12 \+ PersonTimelinePlacement" macos/App/Features/Conclusions` prints nothing; tests pass.

**F-2 — One rule for items not yet measured; the line stays drawn** (from MNT-2, CON-2)
Files: `PersonTimelinePlacement.swift`, `PersonTimelineSection.swift`, `macos/ProvenenciaTests/PersonTimelinePlacementTests.swift`.
Do: an item with no measured height takes `PersonTimelinePlacement.estimatedHeight(heights:)`: the mean of the measured heights, or 120 when none is measured, so `PersonTimelinePlacement` always places every item and `lifeline` is set whenever the content has a range. `stack(_:heights:)` goes; `PersonTimelineStack` always receives complete tops (its own measured-fallback stacking goes); the view passes `placement.tops` with no `lifeline == nil` check.
Failing test first: `PersonTimelinePlacementTests/rowsSitByTheirPoints` — with one height missing, every item still has a top and the line exists (the 120pt mean).
Check: the test passes; `rg -n "func stack" macos/App/Features/Conclusions` prints nothing.

**F-3 — Unfolding and resizing keep the researcher's place** (from UX-1, UX-5)
Files: `PersonTimelinePlacement.swift`, `PersonTimelineSection.swift`, `PersonTimelinePlacementTests.swift`.
Do: `PersonTimelinePlacement.init` gains `pointsPerYear: CGFloat? = nil`; when given, the scale search is skipped. `PersonTimelineList` keeps the chosen scale in `@State` and re-chooses it only when the items' ids or the viewport height change, so a participant unfold (heights only) moves only rows below it. When an unfold changes the ids (the dated or undated fold), after the new placement lands it sets `scrollTarget` to `firstRevealed` and bumps `scrollRequest`, as `select` does, so the revealed row is in view where focus lands. On a width-class change, if `focused` names a card no longer shown, focus moves to that row's title.
Failing test first: `PersonTimelinePlacementTests/aFixedScaleMovesOnlyRowsBelow` *(new)* — `johann` placed, then row 4 made 600pt taller with the same `pointsPerYear`: tops of rows 1 – 4 unchanged, rows below never higher.
Check: the test passes; scripts/check-pr.sh PASS.

**F-4 — The people column goes under the text on a narrow list** (from UX-2)
Files: `PersonTimelineSection.swift`, `PersonTimelinePlacement.swift` (the threshold).
Do: `PersonTimelinePlacement.stacksPeople(listWidth:) -> Bool` *(new)*: true when the text column would be under 280pt (list width − gutter − picture − people column − 3 × 16 < 280). The row then puts the people column under the title and date line, indented past the gutter and picture (the shipped S13-13 place), as a `VStack(spacing: 4)` at `compact.minimumWidth`. The list width is the section's measured width.
Failing test first: `PersonTimelinePlacementTests/narrowListStacksPeople` *(new)* — list width 946 → false; 608 → true.
Check: the test passes; the "Full timeline" preview at a 700pt pane shows the title readable.

**F-5 — One dating per row; the birth placed once** (from MNT-3)
Files: `core/database/conclusionsummaries/order.go`, `compose.go`, `conclusionsummaries_test.go`, `core/database/persondetail/timeline.go`.
Do: `conclusionsummaries.PlaceBirth(birth *datevalue.Value) (PlacedBirth, bool)` *(new; `PlacedBirth` wraps the private `placedDate`)* and `EventSpanAndAge(h EventSummary, birth PlacedBirth, hasBirth bool) (span TimelineSpan, dated bool, age AgeRange, aged bool)` *(new)*, one `eventDating` call; `EventAge` becomes a thin call through them (its tests unchanged); `PersonEvents` places the birth once before the loop and calls `EventSpanAndAge` per row.
Check: `TestEventAge`, `TestEventSpan`, `TestPersonEvents` pass.

**F-6 — Age copy** (from MNT-4, UX-6, CON-OL)
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `PersonTimelineContent.swift`, `PersonTimelineContentTests.swift`.
Do: the range key's comment: "Age at an event when the dates leave it uncertain; arguments are the lowest and highest possible age in years". A spoken range reads `timelineAgeSpokenRange(lo:hi:)` *(new)* "age %lld to %lld"; a single age keeps "age %@".
Check: event 3's spoken label ends "age 17 to 18"; `check-localizable-xcstrings.py` passes.

**F-7 — Date skill and the shared year move's test** (from CON-3, CON-4)
Files: `.agents/skills/add-date-value/SKILL.md`, `core/valuecompare/datewindow_test.go`.
Do: the skill's `calendar.go` list adds `CompletedYears` (the inverse of `AddYears`) and says `AddYears` moves 29 February to 1 March in a common year. `TestShiftYears` gains a row: a 29 Feb 1904 end shifted by 1 → 1 Mar 1905.
Check: `go test -tags fts5 ./core/valuecompare/...` passes.

**F-8 — Card role tooltip; tethers for tall rows** (from UX-3, UX-4)
Files: `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift`, `macos/ProvenenciaTests/PVPersonCardTests.swift`, `PersonTimelineLifeline.swift`, `PersonTimelineLifelineTests.swift`, `PersonTimelineSection.swift`.
Do: `helpText` ends with the role when `role` is set ("Anna Weber\nb. 3 May 1790 · Bremen\nd. 1851\nMother"). `tethers(anchors:spans:visible:)`: `spans: [String: ClosedRange<CGFloat>]` (each anchored item's top to bottom, from placement); a tether is drawn while its item's range meets `visible`, its end y clamped to `visible`.
Check: a `PVPersonCardTests` case for the role line; a `PersonTimelineLifelineTests` case: an item 2000pt tall whose anchor is above `visible` but whose range meets it → one tether, end y at `visible.minY`.

## Round 1 — fix
Head `28bd8383`. Gate PASS (orchestrator). F-1 – F-8 done. Deviations: F-2, F-3, F-4 and F-8's tests were written with the code, not run red first; one combined commit `d4c5a272`; `rowPicture` and `rowGap` moved into `PersonTimelinePlacement`.
Read by the orchestrator: CON-3 **Resolved** (`add-date-value` lists `CompletedYears` and the 1 March rule); CON-4 **Resolved** (`TestShiftYears` "29 February moves to 1 March in a common year"). CI: run 41650651620 on `5ab59d3d` (ship log only) failed `promotealign` `TestLargeSourceBudgets/warm`, a timing budget; the same code passed CI on `ef3d0259` → runner timing, not this PR.

## Round 2 — review
Reviewers: MNT, UX (each had a Fix-now major). Scope `ef3d0259...28bd8383`.
- MNT-1, MNT-2 **Resolved**; UX-1 **Resolved** (`chosenScale` kept across height-only changes; `aFixedScaleMovesOnlyRowsBelow`).
- **MNT-10 · blocker** (= UX-2 Not resolved): the row's width budget is derived twice (`PersonTimelinePlacement.stacksPeople` and the view's `textWidth`/`peopleWidth`) and they disagree: `listWidth` is measured on the outer `HStack` (includes `columnWidth + listGap`, 92pt) and `stacksPeople` counts 3 row gaps where the row has 4, so titles shrink to about 172pt before the people stack; `narrowListStacksPeople` cannot tell; completed.md says 946. Verified: `PersonTimelineSection.swift` measures `listWidth` on the `HStack` whose first child is the 92pt spacer. → F-9.
- UX-5 (minor, round 1) **Not resolved** in one case: `keepFocus()` guards on `focused`, which SwiftUI sets nil when the focused card leaves the tree. → F-10 (rides along).
- Minor note: `conclusionsummaries.EventSpan` and `EventAge` lost their last production callers in F-5 → F-10 (rides along).

### Fix plan 2

**F-9 — One owner for the row's widths; the stacking threshold counts every column** (from MNT-10, UX-2)
Files: `macos/App/Features/Conclusions/PersonTimelinePlacement.swift`, `PersonTimelineSection.swift`, `macos/ProvenenciaTests/PersonTimelinePlacementTests.swift`, `docs/deployment-plan/spike-13/completed.md`.
Failing test first: `PersonTimelinePlacementTests/narrowListStacksPeople` — at `.regular`, a list (the items' width, after the line column) of 863 → true and 865 → false; at `.wide` the same rule with the 930pt column.
Do: `PersonTimelinePlacement` owns the row's horizontal measures beside its vertical ones: `static func peopleWidth(_ widthClass: PVWidthClass) -> CGFloat` (one compact column at Regular; three plus two `cardColumnGap` at Wide; `cardColumnGap` moves here), `static func textWidth(itemsWidth: CGFloat, widthClass:) -> CGFloat` = itemsWidth − `gutterWidth` − `rowPicture` − `peopleWidth` − 4 × `rowGap` (gutter|picture, picture|text, text|spacer, spacer|people), and `static func textCap` = `PVSpacing.widthContentMax − columnWidth − listGap − gutterWidth − rowPicture − 2 × rowGap`; `stacksPeople(itemsWidth:widthClass:)` is `textWidth < minimumTextWidth`. The view's `textWidth`/`peopleWidth` read these; `itemsWidth` is measured on `PersonTimelineStack` (the items), not the outer `HStack`. completed.md states the real threshold (an items width under 864pt at Regular).
Check: the test passes; `rg -n "cardColumnGap|3 \* (rowGap|Self.gap)" macos/App/Features/Conclusions` shows the measures only in `PersonTimelinePlacement.swift`; scripts/check-pr.sh PASS.

**F-10 — Riders** (from UX-5, MNT note)
Files: `PersonTimelineSection.swift`, `core/database/conclusionsummaries/order.go`, `conclusionsummaries_test.go`.
Do: `keepFocus()` also runs when `focused` turns nil while `focusedOwner` is set and that card is no longer shown, moving focus to the owner row's title. Delete `EventSpan` and `EventAge` (no production caller after F-5); `TestEventSpan` and `TestEventAge` call `EventSpanAndAge` (with `PlaceBirth`) and keep their cases.
Check: `rg -n "func EventSpan\(|func EventAge\(" core` prints nothing; Go tests pass; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `600b1c1d`. Gate PASS (orchestrator). F-9: `narrowListStacksPeople` red first; `PersonTimelinePlacement` owns `peopleWidth`, `textWidth(itemsWidth:widthClass:)`, `textCap`, `cardColumnGap`, `stacksPeople`; `itemsWidth` measured on `PersonTimelineStack`; the view reads them. The fix plan's 863/865 were outer widths; on the items the formula gives 772 at Regular (1402 at Wide), which the test and completed.md use. MNT-10 / UX-2 **Resolved** (read by the orchestrator). F-10: the focus rider (`lastFocused`, no unit test) and `EventSpan`/`EventAge` deleted, their tests on `EventSpanAndAge` (`81558a65`).

## Result
Gate PASS · CI green (runs 41654254260, 41654254343 on `8d6eab9c`) · open: none · follow-ups: — · to user: CON-1 (the manual check, incl. click-to-scroll over `PersonTimelineStack`, plan stop condition 2) · CON miss: none
