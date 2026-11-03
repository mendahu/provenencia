# S13-13 — Timeline section: the board's 2026-11-03 update

**Status:** Planned
**Spike / slice:** Spike 13 · Slice 4 — Timeline
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D4](../design/archive/S13-D4-timeline-section.md), board updated 2026-11-03 (turn 12, decisions TL-5b, TL-5c, TL-6x), agreed by the user the same day
**Board:** Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Timeline.dc.html*, turn 12 as updated 2026-11-03. The implementer may not be able to open it: **Design → macOS** and the Decisions give every measure read from it
**Depends on:** S13-13 as built on this branch ([plan](S13-13-timeline-section.md), PR [#691](https://github.com/mendahu/provenencia/pull/691), not merged)
**Blocks:** —
**Branch:** `s13-13-timeline-section` (the same branch and PR as S13-13)
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [person-detail.md § 3. Timeline](../person-detail.md#3-timeline); the S13-13 plan, its Landed note and its [ship log](S13-13-timeline-section.ship.md); the S13-33 plan ([`S13-33-timeline-positions.md`](S13-33-timeline-positions.md)); working rules "The timeline is a lifeline", "Timeline lists fold", "The person card lives in the design system; display is decoupled from content"
**Planned:** 2026-11-03 against `s13-13-timeline-section` @ `604883bf` (on `main` @ `92f87dba`)
**Review:** [`S13-13-timeline-board-update.review.md`](S13-13-timeline-board-update.review.md) — round 1 Fits with changes / Not ready; round 2 Fits with changes / Not ready; every finding Fixed or answered by the user; none open

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

The Timeline follows the board's update of 2026-11-03 (turn 12, TL-5b, TL-5c, TL-6x). Each row sits beside its own point on the lifeline, so stretches with no records open as gaps between rows instead of only on the line. The curve from the line ends in a small flag in the row's gutter that holds the person's age at the event, worked out in Go. The people on an event move into a narrow column at the row's right, as 38pt cards with the role inside the card.

**Why now.** The user changed the board on PR #691 before merge and asked for the update on the same branch (user, 2026-11-03).

**What stays.** The read's order, the markers and the folds of 50 events (S13-33, S13-13); the event's title, picture, date and place; the switcher, header and every other section; click-to-scroll, `whenRequested`, the two drawing layers; Family's regular cards.

## 2. User stories

**US-1.** As a researcher, I see where in the life each event falls, by where its row sits, not only by its point on the line (TL-5b).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann (fixture `johann`) at Regular | Each dated row's flag centre sits on its point unless the row above pushes it down; rows never overlap; the space between the baptism (1817) and the 1835 census opens as a gap (frame 12a; D-3) |
| AC-1.2 | Johann | The median distance between a row's flag and its point is at most 24pt, or the scale is twice its base (D-3, T-4) |
| AC-1.3 | Birth 12 Mar 1817 and baptism 16 Mar 1817 | The baptism sits directly under the birth row, pushed, not on its own point (D-3) |
| AC-1.4 | A short list on a long life | The list is at least the viewport height minus 320pt tall at the scale's floor; the list may run past the end of the line (D-3) |
| AC-1.5 | Only undated events (`ida`) | Rows stack from the top: "Birth unknown", Undated, its two rows, "Death unknown" (D-3) |

**US-2.** As a researcher, I read the person's age at each event at a glance (TL-5c).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Birth 12 Mar 1817; the birth itself | Flag "0" |
| AC-2.2 | Birth 12 Mar 1817; marriage 4 Jun 1844 | Flag "27" |
| AC-2.3 | Birth 12 Mar 1817; census "1835" | Flag "17–18" |
| AC-2.4 | An event "bef. 1880", an undated event, or a person with no dated birth | Flag "?", tooltip "Age unknown" |
| AC-2.5 | An undated row | A dashed flag "?" and no curve |
| AC-2.6 | Hover a flag "27" | Tooltip "Age 27"; VoiceOver reads the row's label ending "…, as Subject, age 27" |
| AC-2.7 | Born 29 Feb 1820; an event 28 Feb 1821 | Age 0; an event 1 Mar 1821 → 1 (D-1) |

**US-3.** As a researcher, the people on an event sit beside it, small, with their role on the card (TL-6x).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Event 1 at Regular | A 300pt column at the row's right: Heinrich Weber "1788 – 1849" FATHER, Anna Weber "1790 – 1851" MOTHER, each 38pt tall with a 28pt likeness; no caption above (frame 12a; D-5, D-6) |
| AC-3.2 | The 41-person burial at Regular | Five cards, then "Show 36 more" (frame 12g; D-7) |
| AC-3.3 | The same at Wide | A 930pt column of three, twelve cards, then "Show 29 more" (frame 12k; D-7) |
| AC-3.4 | VoiceOver on Anna's card | Unchanged: "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851"; hover shows the three-line tooltip |
| AC-3.5 | The person's own roles | Badges lead the date line: SUBJECT · 12 Mar 1817 · Bremen (frame 12a; D-8) |

## 3. Scope

| In | Covered by |
| --- | --- |
| The age at each event on the timeline read: rule, wire, Swift value (D-1, D-2) | STEP-1, STEP-2, AC-2.1 – AC-2.4, AC-2.7 |
| The compact person card reshaped to the board's 38pt card with its role inside (D-5) | STEP-3, AC-3.1, AC-3.4 |
| Row placement against the line, the scale chosen from the rows (D-3, D-4) | STEP-4, AC-1.1 – AC-1.5 |
| The age flag, the row's new columns, the role Badges on the date line, the participant folds by width (D-6 – D-10) | STEP-5, STEP-6, AC-2.5, AC-2.6, AC-3.1 – AC-3.3, AC-3.5 |

| Out | Home |
| --- | --- |
| The board's bracket shapes for "bef." and "aft." marks and its 20 × 2pt birth and death caps | not planned: not among the three changes the user named (user, 2026-11-03); the shipped marks stay |
| The board's 25-year tick step | not planned: same reason; the shipped steps and the 200-tick cap stay |
| An estimated age for an unknown birth | not planned (user, 2026-11-02: no date is invented) |
| Family's "Show N more" focus gap | the follow-up already logged in the S13-13 ship log |

## 4. Current state

Planning base: branch `s13-13-timeline-section` @ `604883bf` (PR #691), on `main` @ `92f87dba`.

| Area | Where | Today |
| --- | --- | --- |
| Timeline read (Go) | `core/database/persondetail/timeline.go` `PersonEvent { Event; Roles; Participants; Span }`, `PersonEvents` (reads the person's own summary as `me`, sets `Span` from `conclusionsummaries.EventSpan(h)`) | No age |
| Where a date sits (Go) | `core/database/conclusionsummaries/order.go` `EventSpan(h EventSummary) (TimelineSpan, bool)`; `compose.go` `datePlace`, `eventDating` (the one rule for an event's chosen date), `TimelineSpan` | The owner of "where an event's date sits" |
| Date ends | `core/valuecompare/datewindow.go` `DateEnd { Year, Month, Day; Lo, Hi; From, To, Mid; Open }`, `DateEnds(d)`; `core/datevalue/calendar.go` `DaysFromCivil`, private `civilFromDays`, `YearLine` | `Lo`/`Hi` are day numbers |
| Wire | `api/proto/engine.proto` `PersonEvent` (fields 1 – 3, 5; 4 reserved), `TimelineSpan`; `api/ffi/handlers/person_detail.go` builds `engine.PersonEvent` and sets `Span`; `scripts/generate-proto.sh` | — |
| Swift value | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonEvent { event; roles; participants; span }`; `macos/App/Platform/GoStore.swift` maps `e.hasSpan` | — |
| Card | `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift` `PVTileCardSize` `.compact`: height 52, thumbnail 40, padding 6, `minimumWidth` 232, card `HStack(spacing: 10)`; `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` `lines` at `.compact`: the years in mono caption; `PVPersonCardContent.role`; tests `macos/ProvenenciaTests/PVTileCardTests.swift`, `PVPersonCardTests.swift` | The Timeline is `.compact`'s only caller |
| Row and line | `macos/App/Features/Conclusions/PersonTimelineSection.swift` `PersonTimelineList` (a `VStack` of items, `itemFrames` measured per item, `PersonTimelineLifeline(content:itemFrames:)` rebuilt on change), private `PersonTimelineEventRow` (picture, title, date line, Badges at the trailing edge, a `LazyVGrid` of compact cards with captions below), `PersonTimelineLine` (static `Canvas`, click buttons), `PersonTimelineTethers` (scroll layer) | Rows stack; the line stretches to the list's height |
| Geometry | `macos/App/Features/Conclusions/PersonTimelineLifeline.swift` `PersonTimelineLifeline` (`length` = first item top to last item bottom; scale = (length − stubs) ÷ years; `columnWidth` 80; `quarterTickYearHeight` 136; `tethers(itemFrames:visible:)` ending at x = 80, row `minY + 12`) | — |
| Content | `macos/App/Features/Conclusions/PersonTimelineContent.swift` `eventFoldLimit` 50, `participantFoldLimit` 15, `Row`, `spokenLabel`, `firstRevealed(by:)`, `PersonCardLink` participants | — |
| Title | `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift` `Style.display(compact:)`: h3 (18) when compact, else display3 (34); `PVTypeScale.h2` 22 | The board titles an event row at h2 |

**Precedent.** S13-33 ([#687](https://github.com/mendahu/provenencia/pull/687)) added `span` to the same read: a Go field on `PersonEvent` from a `conclusionsummaries` helper, a proto field, the FFI line in `person_detail.go`, the `GoStore` mapping, table tests in `conclusionsummaries_test.go` and `timeline_test.go`. Copy that path for `age`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The age rule is Go's (user, 2026-11-03). **One leap-day rule** (user, 2026-11-03: a 29 February birthday is reached on 1 March in a common year): `datevalue.AddYears` moves 29 February to 1 March in a common year (today 28 February; its doc comment and `TestAddYears`'s "leap day to common year" row change), and `datevalue.CompletedYears(born, on int) int` *(new)* in `core/datevalue/calendar.go` is its inverse on day numbers (`DaysFromCivil`): the largest n with `AddYears(born's date, n)` on or before `on`, negative when `on` is before `born`. Matching's lifespan windows (`valuecompare.ShiftYears`, `subjectprofile`) move a day for a 29 February end. **One date check** (round 1 M-1): `datePlace` returns `(placedDate, bool)` with `placedDate { key int; span TimelineSpan; first, last valuecompare.DateEnd }` *(new, private)*, the ends it already computes; `dateKeyOf` and `eventDating` read its fields, and `eventDated` carries `first` and `last`; `conclusionsummaries.EventAge(h EventSummary, birth *datevalue.Value) (AgeRange, bool)` *(new)* in `order.go` beside `EventSpan` reads `eventDating(h.Date, h.StartDate)` and `datePlace(birth)` and adds only the age's own rule: false when either is not placed or any of the four ends is open (BEF, AFT, a one-sided range); else `Hi = CompletedYears(birthFirst.Lo, eventLast.Hi)`, false when `Hi < 0`, and `Lo = max(0, CompletedYears(birthLast.Hi, eventFirst.Lo))`. `type AgeRange struct { Lo, Hi int }` *(new)* in `compose.go` beside `TimelineSpan` | One owner of date arithmetic and of whether a date can be placed; one leap-day rule in `datevalue` | A Swift rule over the spans; fractional years from the spans (a birthday reads "26–27"); a second `DateEnds` call and window check in `EventAge`; two leap-day rules |
| D-2 | `persondetail.PersonEvent.Age *conclusionsummaries.AgeRange` *(new)*, set in `PersonEvents` from `me.Birth.Date` (the header's birth): the `summaries.Person(entityID)` lookup moves before the loop that sets `Span`, and that loop sets `Age` too; with no `me` or a nil `me.Birth.Date`, `Age` is nil on every row; `message AgeRange { int32 lo = 1; int32 hi = 2; }` *(new)* and `PersonEvent.age = 6` *(new)* in `engine.proto`; `person_detail.go` sets it; `CatalogAgeRange { lo: Int; hi: Int }` *(new)* and `CatalogPersonEvent.age: CatalogAgeRange?` *(new)* in `GenealogyStore.swift`, mapped in `GoStore.swift` from `e.hasAge`; `api/ffi/handlers/person_detail_test.go` asserts `GetAge()` beside `GetSpan()` | S13-33's path | A per-section Swift read of the birth |
| D-3 | Placement (TL-5b) is pure, and owns every vertical measure of an item (round 1 M-5): `PersonTimelinePlacement` *(new)* in `macos/App/Features/Conclusions/PersonTimelinePlacement.swift` *(new)*. `static func topPadding(_ item: Item, after previous: Item?) -> CGFloat`: 4 for the first item, 22 for the Undated heading, 14 for the item after it, else 20; `static func hasDivider(_ item: Item, after previous: Item?) -> Bool`: false for the first item, the Undated heading and the item after it, else true; `static let bottomPadding: CGFloat = 20`, `flagInset: CGFloat = 4`, `flagWidth: CGFloat = 52`, `flagHeight: CGFloat = 20`, `gutterWidth: CGFloat = flagInset + flagWidth` (56). An event row's (and a recorded marker's) anchor offset is `(divider ? 1 : 0) + topPadding + flagInset + flagHeight / 2`; a placeholder's is half its height; the Undated heading and a "Show N more" row have none. `init(content: PersonTimelineContent, heights: [String: CGFloat], viewportHeight: CGFloat)`, heights being each item's whole height (divider and paddings included), gives `pointsPerYear`, `lifeline: PersonTimelineLifeline`, `tops: [String: CGFloat]`, `anchors: [String: CGFloat]` (each anchored item's top plus its offset) and `height` (the larger of the last item's bottom and the line's `length`). An anchored item wants `top = mark.pointY − offset`; each top is `max(previous bottom, want)`, or the previous bottom when it has no anchor. **Scale:** over `PersonTimelineLifeline.range(content)` (D-4), `years = to − from`; `floor = max(300, viewportHeight − 320) ÷ years`, `viewportHeight` being `pvPageViewportSize.height` read by `PersonTimelineList` (0 before its first measure, so the floor is then 300 ÷ years); `base = max(floor, min(sum of heights ÷ years, 40))`; try `base × f` for f in 1, 1.25, 1.5, 1.75, 2: keep the first whose median `top − want` over anchored items is at most 24pt, else the one with the smallest median (a later try replaces the kept one only when its median is more than 1pt smaller). With no range, items stack. Until every item has a height, items stack with no gaps and no line is drawn | The board's rule, testable without SwiftUI; one owner of the paddings the view, the anchors and the tethers read | A SwiftUI `Layout` that also chooses the scale; the paddings in the view and again in the placement |
| D-4 | `PersonTimelineLifeline.init(content:pointsPerYear:)` *(changed from `init(content:itemFrames:)`)*. `static func range(_ content: PersonTimelineContent) -> (from: Double, to: Double)?` *(new)*: the earliest `from` to the latest `to` of `allSpans`, `to` at least `from + 0.25` (the board's minimum, replacing the shipped one year), nil with no span; placement and the line both read it. `origin` goes: the line is drawn in the section's coordinates from y = 18. Every marker with no span of its own, placeholder or recorded, has a 40pt stub, inserted where the shipped **Mapping** puts it (a birth stub at the `from` of the row its marker precedes, a death stub at the `to` of its `besideEventId` row). `y(t) = 18 + (t − from) × pointsPerYear + 40 × (birth stubs with t ≥ their insertion year) + 40 × (death stubs with t > their insertion year)`, as the shipped `map` closure counts them; a birth stub runs from `y(at) − 40` to `y(at)` with its cap at the top, a death stub from `y(at)` to `y(at) + 40` with its cap at the bottom; `length = 18 + (to − from) × pointsPerYear + 40 × (all stubs) + 18`. With no range: the birth cap at 18, the death cap at 58, one dashed stub from 18 to 58, `length` 76. `Mark.pointY` (the existing property; round 1 M-4) becomes: a dot or cap at its y; a closed bar its middle; an open-start bar its known end minus 8; an open-end bar its start plus 8. `quarterTickYearHeight` becomes 56. `tethers(anchors: [String: CGFloat], visible: CGRect) -> [Tether]` *(changed)*: for each anchor inside `visible`'s vertical range, a curve from its mark's `pointY` to (`columnWidth` + 12 + `PersonTimelinePlacement.flagInset`, the anchor). A recorded birth or death with no span has a mark (its cap on its own stub), so it is anchored at the cap, gets a curve and a solid flag; its age reads "?" (round 2 X-21). An undated row in the Undated group has no mark, so no anchor and no curve, and a dashed flag. `dotInset`, `Tether.isDashed` and the private `targets(_:pointY:)` and `Target` go. Both `Canvas` layers are `placement.height` tall | The line no longer follows the list's height; one answer to "where is a row's point" | Keeping item frames as the input; a second `pointY(id)` beside `Mark.pointY` |
| D-5 | `PVTileCardSize.compact` is the board's card (user, 2026-11-03): `height` 38, `thumbnail` 28, `titleLineLimit` 1, `linesLineLimit` 1, `minimumWidth` 300. `padding` becomes `EdgeInsets` *(changed from `CGFloat`)*: `.regular` 10 on each side (as today), `.compact` top 4, leading 4, bottom 4, trailing 8; `contentSpacing` *(new)*: `.regular` 10, `.compact` 8, read by `PVTileCard`'s `HStack`. `PVCard` gains `init(…, insets: EdgeInsets)` *(new)* beside its `padding: CGFloat?` init; both store one `EdgeInsets?` (round 1 M-6, X-4). `PVPersonCard` at `.compact`: the title in `PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium)`; the `lines` slot is one line, a `HStack(spacing: 12)` of the years (`PVFont.mono(size: PVTypeScale.micro)`, `textMuted`) and, when `content.role` is set, a `Spacer(minLength: 0)` and the role in `.pvMicroCaps()` `textMuted` | One compact size; its only caller moves with it; every size measure in `PVTileCardSize` | A third size beside an unused 52pt one; padding around the shell in the feature |
| D-6 | The event row (`PersonTimelineEventRow`) is a `HStack(alignment: .top, spacing: 16)` of the flag gutter (`PersonTimelinePlacement.gutterWidth` wide; the flag `flagInset` from its leading and top edges), the picture (`PersonTimelineEventRow.picture`, 72), the text column (the title, then the date line; `.frame(maxWidth: PVSpacing.widthContentMax − PersonTimelineLifeline.columnWidth − gutterWidth − picture − 2 × 16, alignment: .leading)`, so Wide rule 2's measure holds) and the people column: `PVTileCardSize.compact.minimumWidth` wide at Regular, a `VStack(spacing: 4)` of cards; at Wide `3 × minimumWidth + 2 × 15` (930) wide, a `LazyVGrid` of three `.flexible()` columns (spacing 15, row spacing 4); then `ConclusionShowMoreRow` when folded. The list sits 12pt right of the line column. The item wrapper (`PersonTimelineItemView`) owns the divider and the paddings, read from `PersonTimelinePlacement` (`hasDivider`, `topPadding`, `bottomPadding`): `VStack(spacing: 0) { divider when set; content with that top and bottom padding }`; the shipped `.padding(.vertical, PVSpacing.space5)` and trailing `PVDivider` go. The captions above cards and the adaptive grid go (the role is on the card, D-5); the section still fills the pane at Wide | The board's TL-6x and row | Keeping the grid under the title; an unbounded text column |
| D-7 | The participant fold limit follows the width class (the board's five and twelve): `PersonTimelineContent.init(timeline:expanded:locale:orders:widthClass: PVWidthClass = .regular)`, `static func participantFoldLimit(_ widthClass: PVWidthClass) -> Int` *(new)*: 5 at `.regular`, 12 at `.wide`; the static `participantFoldLimit` (15) goes; the private `row(...)` takes the limit; `firstRevealed(by: .participants)` follows it. `PersonTimelineSection` reads `@Environment(\.pvWidthClass)` and passes it in the host's `make` closure, which re-runs when it changes; the Wide previews pass `.wide` (round 1 X-8, M-10) | One width input; callers cannot pass an arbitrary number | A fixed 15; a bare `Int` parameter |
| D-8 | The person's role Badges lead the date line: `HStack(spacing: 8) { Badges; EventSecondaryLine(...) }`, still hidden from VoiceOver | Frame 12a | Badges at the trailing edge |
| D-9 | The flag (TL-5c) is private `PersonTimelineAgeFlag` *(new)* in `PersonTimelineSection.swift`: `PersonTimelinePlacement.flagWidth` × `flagHeight`, `PVFont.mono(size: 11)`, `PVColor.textSecondary`, a 1pt `PVColor.borderDefault` border (dashed `[3, 3]` on an undated row), corner radius 3, `PVColor.surfaceCard` fill; text `Row.ageText` *(new)*: "27", `L10n.Conclusions.timelineAgeRange(lo:hi:)` "17–18", or "?" (`Text(verbatim:)`); `.help` `L10n.Conclusions.timelineAgeHelp(_:)` "Age %@" or `timelineAgeUnknown` "Age unknown"; `.accessibilityHidden(true)`; the row's `spokenLabel` ends with `L10n.Conclusions.timelineAgeSpoken(_:)` "age %@" when the age is known. Placeholders have no flag | The board's flag; the age spoken in the row's one label | The age only in a tooltip |
| D-10 | The event title is `PVTitleText(row.title, style: .heading, lineLimit: 2)`. `PVTitleText.Style.heading` *(new)*: the display face at `PVTypeScale.h2` (22), a ref mono at `PVTypeScale.h3` (18). `displayText` takes a private `DisplayMetrics { titleSize; refSize }` chosen per style (`.display(compact: true)` → h3, h4; `.display(compact: false)` → display3, h1; `.heading` → h2, h3), so the name / label / ref switch stays one (round 1 M-7); medium weight, display tracking, `textDisplay`, `.isHeader` as today. The header is unchanged | The board titles the row at h2 (`var(--text-h2)`); one title component, one switch | A third copy of the display switch; a font override outside `PVTitleText` |

**Rules this plan bends.** (1) Working rule "The timeline is a lifeline" (`docs/deployment-plan/spike-13/deployment-plan.md`) says "Every event sits at its own date, wherever that falls" and "Positions are data: Go places each dated event and the markers (S13-33) with the date rule matching already uses (`valuecompare.DateEnds`), and Swift only scales them"; `person-detail.md` R-24a says "Every event sits at its own date (user, 2026-11-02)" → D-3 places each row as near its point as the rows above allow → user, 2026-11-03 (the board, TL-5b) → STEP-7 rewords both: "Every event's point sits at its own date; its row sits as near that point as the rows above allow, so silent stretches also open as gaps between rows", and "Positions are data: Go places each dated event and the markers (S13-33) with the date rule matching already uses (`valuecompare.DateEnds`); Swift scales them and places each row by its point". (2) Working rule "Timeline lists fold" ("each event's participants fold at 15") and the S13-13 map row ("participants 15") → D-7's 5 at Regular and 12 at Wide → user, 2026-11-03 (the board, TL-6x) → STEP-7 rewords both. (3) `datevalue.AddYears`'s 29 February rule → D-1 → user, 2026-11-03 → STEP-1 changes the code and its doc comment. Wide rule 2 (`docs/macos-client-patterns.md` § 9) is kept: D-6 caps the text column.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `datevalue.CompletedYears`, `conclusionsummaries.EventAge`, `AgeRange` | func, type | `YearLine`, `EventSpan`, `TimelineSpan` | Beside them; the same owner |
| `PersonEvent.Age`, proto `AgeRange`, `CatalogAgeRange` | field, message, type | `Span`, `TimelineSpan`, `CatalogTimelineSpan` | S13-33's path |
| `PersonTimelinePlacement`; `PersonTimelineStack` *(private)* | pure layout; `Layout` | `PersonTimelineLifeline`; the `VStack` it replaces | Placement needs heights and a scale search; the lifeline keeps drawing geometry; the stack only places |
| `PVTitleText.Style.heading` | style | `.display(compact:)` | One more size in the one title component |
| `PersonTimelineAgeFlag` | snowflake | — | One caller |

**Builds on.** `eventDating`, `valuecompare.DateEnds`, `PVTileCardSize`, `PVPersonCard`, `ConclusionShowMoreRow`, `PVDivider`, `pvWidthClass`, the shipped two-layer drawing and `whenRequested`.
**Deletes.** `PersonTimelineLifeline.Tether.isDashed`, `targets(_:pointY:)` and `Target`; the dashed-dot branch in `PersonTimelineTethers.draw`; `PersonTimelineItemView.isLast`; the 52pt compact card's measures; the participant captions above cards; the adaptive `GridItem(.adaptive(minimum: PVTileCardSize.compact.minimumWidth))` grid under the title; `PersonTimelineContent.participantFoldLimit` (15); `PersonTimelineLifeline.init(content:itemFrames:)`, `tethers(itemFrames:visible:)`, `origin` and `dotInset`; `itemFrames` and its pruning in `PersonTimelineList.rebuild()`; `PersonTimelineItemView`'s `.padding(.vertical, PVSpacing.space5)` and trailing `PVDivider`; D-15's `widthContentMax − columnWidth` frame on the row's top line.
**Next adds.** None named.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A row with no dated birth for the person | `PersonTimelineAgeFlag` | "?" (AC-2.4) |
| Resize across Regular and Wide | participant folds | The limit follows the width class; `expanded` is kept |

### Scale

Go: `EventAge` is one `DateEnds` per event plus one for the birth, constant per row. Swift: placement tries at most five scales, each linear in the items shown (at most about 60 folded, 120 unfolded at the brief's largest), on a height or fold change, never per scroll step; the scroll layer still computes only tethers.

### Edges

- **Assumptions.** Rows are in the read's order and only move down: no row is placed above the one before it. The view reads its paddings from `PersonTimelinePlacement`, so the anchor and the drawn flag agree.
- **Neighbors.** An event on the birthday → one number; the day before → one less. A 29 Feb birthday → reached 1 Mar in a common year, and matching's windows use the same rule. An event dated before the birth (a mis-dated record) → `Hi < 0` → "?". A birth "abt. 1817" → a range on every row.
- **Removed inputs.** A birth unpromoted or merged → the next read has no birth → every flag "?".
- **Real data.** None — no table drives it.
- **Untrusted size.** Unchanged from S13-13: years bounded by `MinYear…MaxYear`; rows bounded by the folds.

### Schema

None — no catalog change.

### Go core

D-1, D-2. `CompletedYears` beside `DaysFromCivil` (uses `civilFromDays`). `EventAge` beside `EventSpan`.

### FFI

`PersonEvent.age = 6`, `AgeRange`; regenerate with `scripts/generate-proto.sh`; `person_detail.go` sets it next to `Span`. No new method.

### macOS

Frames in words (turn 12, update 2026-11-03):

- **12a Regular.** The 80pt line column, a 12pt gap, then the list. Row 1 at the top: a flag "0" in the gutter at the line's birth point, the picture, "Birth of Johann Weber" at 22pt, SUBJECT · 12 Mar 1817 · Bremen, and at the right the 300pt column of Heinrich Weber FATHER and Anna Weber MOTHER. The baptism directly below; then a gap to the 1835 census ("17–18"); rows from 1844 on close together; Undated with "?" dashed flags before the death row ("63") and the burial ("63").
- **12d Wide.** The list fills the pane; each people column 930pt, three cards a row.
- **12g / 12h / 12k Large event.** Five cards and "Show 36 more" at Regular; all 41 unfolded; twelve and "Show 29 more" at Wide.
- **12f Row sheet.** Every row state with the flag and the people column.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Placement | pure | `PersonTimelinePlacement.swift` *(new)* | D-3 |
| Age flag | snowflake | `PersonTimelineSection.swift` | D-9 |
| People column | `PVPersonCard(size: .compact)` | `PersonTimelineSection.swift` | D-5, D-6, D-7 |
| Title | `PVTitleText(style: .heading)` | `PVTitleText.swift` | D-10 |

### Strings

L10n keys *(new)* under `L10n.Conclusions`: `timelineAgeRange(lo:hi:)` = "%lld–%lld"; `timelineAgeHelp(_ age: String)` = "Age %@"; `timelineAgeUnknown` = "Age unknown"; `timelineAgeSpoken(_ age: String)` = "age %@".

## 7. Steps

**STEP-1 — The age rule in Go**
Files: `core/datevalue/calendar.go`, `core/datevalue/calendar_test.go`, `core/database/conclusionsummaries/order.go`, `compose.go`, `conclusionsummaries_test.go`, `core/subjectprofile/lifespan_test.go` (the row "AC-2.10 a leap day moves to the last day of February" becomes "AC-2.10 a leap day moves to 1 March", ends 1894-3-1 and 2024-3-1), `core/valuecompare/datewindow_test.go` (any 29 February expectation). Skill: `add-date-value`.
Do: D-1.
Check: T-1, T-2 pass; `go test -tags fts5 ./core/...` passes (matching's tests with the new 29 February rule).

**STEP-2 — The age on the read and the wire**
Files: `core/database/persondetail/timeline.go`, `timeline_test.go`, `api/proto/engine.proto`, generated code (`scripts/generate-proto.sh`), `api/ffi/handlers/person_detail.go`, `person_detail_test.go`, `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift`. Skill: `add-ffi-handler`.
Do: D-2. The fixture's `event(...)` helper gains `age: CatalogAgeRange? = nil`, with these literals (birth 12 Mar 1817 for `johann` and `unknownEnds`, 3 Apr 1845 for `ernst`): johann 1 (0,0), 2 (0,0), 3 (17,18), 4 (27,27), 5 (27,27), 6 (28,29), 7 (32,33), 8 (41,42), 9 (43,44), 10 (63,63), 11 (63,63), 12 and 13 none; `unknownEnds` the same rows with none on every row (no dated birth); ernst e1 (0,0), e2 (0,0), e3 (2,2); `largeEvent`'s burial (63,63); `lise` and `ida` none.
Check: T-3 passes; the handler test passes; the app builds.

**STEP-3 — The compact card**
Files: `macos/App/DesignSystem/Components/TileCard/PVTileCard.swift`, `macos/App/DesignSystem/Components/Card/PVCard.swift`, `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift`, `macos/ProvenenciaTests/PVTileCardTests.swift`, `PVPersonCardTests.swift`, `macos/App/DesignSystem/README.md` (Card, TileCard and PersonCard rows). Skill: `add-ui-component`.
Do: D-5.
Check: T-6 passes; `PVPersonCardTests` pass.

**STEP-4 — Placement and the line**
Files: `macos/App/Features/Conclusions/PersonTimelinePlacement.swift` *(new)*, `PersonTimelineLifeline.swift`, `PersonTimelineContent.swift` (D-7), `macos/ProvenenciaTests/PersonTimelinePlacementTests.swift` *(new)*, `PersonTimelineLifelineTests.swift`, `PersonTimelineContentTests.swift`. Skill: `add-swift-test`.
Do: D-3, D-4, D-7.
Check: T-4, T-5, T-7 pass.

**STEP-5 — The row**
Files: `PersonTimelineSection.swift`, `PersonTimelineContent.swift` (`ageText`, spoken age), `macos/App/DesignSystem/Components/TitleText/PVTitleText.swift`, `macos/App/DesignSystem/README.md` (TitleText row), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-6, D-8, D-9, D-10: the row's columns, the flag, the people column, the title; `PersonTimelineList` positions items with `PersonTimelineStack` *(new, private)*, a `Layout` that takes `tops: [CGFloat]` (in item order, from `PersonTimelinePlacement.tops`) and `height`, proposes each subview the full width, places it at its top, and reports `height`; with no tops it stacks. Heights are measured per item with `onGeometryChange` of `size.height` (a height does not depend on a top). The line and tethers come from `placement.lifeline`.
Check: T-8 passes; the app builds; the previews "Full timeline", "Large event" and "Row sheet" read as frames 12a, 12g and 12f.

**STEP-6 — Clicks and the scroll layer**
Files: `PersonTimelineSection.swift`.
Do: `PersonTimelineTethers` reads `placement.anchors` instead of item frames; click-to-scroll and `whenRequested` unchanged; `firstRevealed` focus unchanged.
Check: scripts/check-pr.sh PASS.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-13/prs/S13-13-timeline-section.md` (Landed note: this update, D-15's measure dropped), this plan (status, Landed note), `docs/deployment-plan/spike-13/deployment-plan.md` (the working rules "The timeline is a lifeline" and "Timeline lists fold", the S13-13 map row and the S13-13 scope entry's "folds at 50 events and 15 participants" reworded as **Rules this plan bends** says), `docs/deployment-plan/spike-13/design/archive/S13-D4-timeline-section.md` (Board line: the 2026-11-03 update, TL-5b, TL-5c, TL-6x), `docs/deployment-plan/spike-13/person-detail.md` § 3 (R-24a as **Rules this plan bends** says; R-27: people beside the event; the age at each event), `docs/deployment-plan/spike-13/completed.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` on both plans reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/datevalue/calendar_test.go` `TestCompletedYears` *(new)*, `TestAddYears` (row changed) | 12 Mar 1817 → 12 Mar 1817: 0; → 4 Jun 1844: 27; → 11 Mar 1844: 26; 29 Feb 1820 → 28 Feb 1821: 0, → 1 Mar 1821: 1; on before born: −1; for each case `AddYears(born, n)` is on or before `on` and `AddYears(born, n + 1)` after it; `AddYears(1904, 2, 29, 110)` → 2014-3-1 | AC-2.7, D-1 |
| T-2 | Go | `core/database/conclusionsummaries/conclusionsummaries_test.go` `TestEventAge` *(new)* | birth 12 Mar 1817: event same day → {0,0}; 4 Jun 1844 → {27,27}; "1835" → {17,18}; "1844 – 1849" → {26,32}; "bef. 1880" → false; birth "abt. 1817" and event 4 Jun 1844 → {26,27}; nil birth → false; event before birth → false; year outside the window → false | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-1 |
| T-3 | Go | `core/database/persondetail/timeline_test.go` `TestPersonEvents` subtest `age` *(new)*; `api/ffi/handlers/person_detail_test.go` (assertion added) | a person born 12 Mar 1817 (a day) with events on 4 Jun 1844 → {27,27} and one undated → nil; a person with no birth → nil on every row; in `person_detail_test.go` the first fixture's `Events[0]` (1820 against a birth in 1820) → `GetAge()` {0,0}, and in the "carries spans and markers" test `Events[0]` → {0,0}, `Events[1]` (undated) → nil | D-2 |
| T-4 | Swift | `macos/ProvenenciaTests/PersonTimelinePlacementTests.swift` *(new)* `rowsSitByTheirPoints` | `johann`, each item 120pt, viewport 900 → no two items overlap; the birth row's top is 0 (its point 18 minus offset 4 + 4 + 10); the baptism's top is 120 (pushed under the birth); the census row's top is its `pointY` minus 35 (divider 1 + 20 + 4 + 10); median distance ≤ 24 or the factor is 2; `anchors` equal tops plus offsets; with any height missing → stacked tops, no line | AC-1.1, AC-1.2, AC-1.3, D-3 |
| T-5 | Swift | same file `scaleFloorAndRunPast` | `ernst` (2.87 years), viewport 900 → points per year ≥ 580 ÷ that range; viewport 0 → ≥ 300 ÷ that range; `ida` → stacked, no range; `johann` with each item 400pt → `height` is the last item's bottom, more than the line's `length` | AC-1.4, AC-1.5, D-3 |
| T-6 | Swift | `macos/ProvenenciaTests/PVTileCardTests.swift` (cases changed) | `.compact` → height 38, thumbnail 28, minimumWidth 300, titleLineLimit 1 | AC-3.1, D-5 |
| T-7 | Swift | `macos/ProvenenciaTests/PersonTimelineContentTests.swift` (cases changed) | `largeEvent` with limit 5 → 5 shown, 36 hidden, `firstRevealed` the 6th; limit 12 → 12, 29; `ageText` "0" (event 1), "27" (event 4), "17–18" (event 3), "?" (event 12); event 4's spoken label ends "age 27", event 12's has no age; the existing assertion on event 2's `spokenLabel` gains ", age 0" | AC-2.1 – AC-2.6, AC-3.2, AC-3.3, D-7, D-9 |
| T-8 | Swift | `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift` (cases changed) | `init(content:pointsPerYear:)`: Johann (no stub) at 43pt a year → length = 18 + (range years) × 43 + 18; `unknownEnds` at 43pt a year → length = 18 + 40 + 40 + (to − from) × 43 + 18, the birth cap at 18, the first dated point at 58; a range under 0.25 years drawn as 0.25; quarter ticks from 56pt a year; `pointY` of "bef. 1900" is its end minus 8; `tethers(anchors:visible:)` end at x = 96 at each anchor; none for undated rows or anchors outside `visible` | D-4 |

Most likely to break: `PersonTimelineLifelineTests` (signature change), `PVTileCardTests` (padding type), `PersonTimelineContentTests` (fold limit, spoken age), matching tests with 29 February ends (`core/valuecompare`, `core/subjectprofile`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `participantFoldLimit =`, `init(content:itemFrames:)`, `tethers(itemFrames:`, `itemFrames`, `dotInset`, `lifeline.origin`, "fold at 15", "participants 15", "15 participants", "Swift only scales them" and "Every event sits at its own date" `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `docs/deployment-plan/spike-13/*.md` and `macos/` with no hit left outside plan and review history
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches frames 12a, 12d, 12f, 12g, 12h, 12k of the 2026-11-03 board
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #691 (date); a Landed note lists every difference from the plan
- [ ] completed.md: the S13-13 paragraph names the age flag, the placement and the people column
- [ ] PR description: the update's ACs, evidence, deviations, manual check
```

## 10. Manual check

1. (AC-1.1 – AC-1.3, AC-2.1 – AC-2.3, AC-3.1, AC-3.5) Open Johann on Timeline at Regular: frame 12a — flags "0", "0", "17–18", "27", gaps where the record is silent, people at the right with roles on the cards.
2. (AC-2.4, AC-2.5, AC-2.6) Undated rows show dashed "?" flags; hover a flag; VoiceOver reads the age at the end of the row.
3. (AC-3.2, AC-3.3) The "Large event" preview at Regular and at Wide: five, then twelve, and the counts.
4. (AC-1.4, AC-1.5) "Short life" and "Only undated" previews.
5. Click a point: the row scrolls to the top, again on a second click; scroll: curves follow the rows in view.
6. (AC-3.4) Family's regular cards look as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A row placed by the `Layout` does not scroll into view by id | `PersonTimelineStack` sets real frames, which `scrollTo` reads; manual check 5 |
| Heights change after placement (an image loads, a fold opens) | Heights feed `PersonTimelinePlacement`; a change re-places, never per scroll step |

**Stop and report if:**

- `valuecompare.DateEnds` gives no day bounds (`Lo`, `Hi`) for a year-precision date, so D-1 cannot take its ends from it.
- `scrollTo` does not land on a row placed by `PersonTimelineStack` (manual check 5).
- The board shows a measure this plan does not give.

## 12. Open questions

None. The user answered on 2026-11-03: the age is Go's; `.compact` is reshaped.

## 13. PR

**Title:** S13-13 — Timeline section (PR #691, same branch)
**Body:** the update's ACs, evidence and deviations added to #691's description; `Plan: docs/deployment-plan/spike-13/prs/S13-13-timeline-board-update.md`.
