# Review — S13-13 Timeline section: the board's 2026-11-03 update

**Verdict:** round 1 Fits with changes — Not ready; round 2 Fits with changes — Not ready, every finding Fixed; none open (cap reached).
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **Where the age is worked out.** Answered 2026-11-03: in Go, on the timeline read (D-1, D-2).
- **The compact card.** Answered 2026-11-03: reshape `.compact` into the board's 38pt card (D-5).
- **M-2: the 29 February birthday.** Answered 2026-11-03: reached on 1 March in a common year. `AddYears` changes with it, so the engine keeps one rule; matching's windows move a day for a 29 February end (D-1, T-1).

## Writer's own calls

- The board's 22pt row title (`var(--text-h2)`) replaces the shipped 18pt (D-10): it was not one of the three changes the user named, but the board is the source.
- The board's bracket shapes for "bef."/"aft.", its bar-shaped birth and death caps and its 25-year tick step stay as shipped (§ 3 Out).
- The text column keeps a measure, so Wide rule 2 holds and is not bent (D-6).
- The age is spoken at the end of the row's label ("age 27"), and the flag is hidden from VoiceOver (D-9).
- Lise's generated events carry no age (flags "?") so the fixture keeps no date arithmetic.

## Round 1

- **M-1 · blocker · `EventAge` checks a date's ends a second way.** Fixed: `datePlace` returns its `first`/`last` ends, `eventDated` carries them, `EventAge` reads `eventDating` and `datePlace(birth)` (D-1).
- **M-2 · blocker · A second 29 February rule.** To user, answered: 1 March, and `AddYears` changes to match (D-1, T-1).
- **M-3 · blocker · Rules bent but not listed; one misquoted.** Fixed: **Rules this plan bends** lists "The timeline is a lifeline" (its real sentences), R-24a, "Timeline lists fold" and the map row, and `AddYears`; STEP-7 rewords each; Wide rule 2 is kept by capping the text column (D-6). The Sweep searches text that exists.
- **M-4 · major · A second `pointY(id)`.** Fixed: `Mark.pointY` changes (D-4).
- **M-5 · major · Paddings and the flag offset in three places.** Fixed: `PersonTimelinePlacement` owns `topPadding`, `hasDivider`, `bottomPadding`, `flagInset`, `flagHeight` and `anchors`; the view and `tethers(anchors:visible:)` read them (D-3, D-4, D-6).
- **M-6 · major · Card measures outside `PVTileCardSize`; `PVCard` has one padding.** Fixed: `padding` becomes `EdgeInsets`, `contentSpacing` is new, `PVCard` gains `insets:`; the people column reads `compact.minimumWidth` (D-5, D-6, STEP-3).
- **M-7 · major · `.heading` would copy the display switch.** Fixed: a private `DisplayMetrics` per style, one switch (D-10).
- **M-8 · major · Deletes and Sweep miss dead code.** Fixed: `dotInset`, `origin`, `rebuild()`'s pruning, the adaptive grid and the item wrapper's padding and divider are in **Deletes**; the Sweep names them.
- **M-9 · minor · Next adds names an unrecorded future.** Fixed: "None named."
- **M-10 · nit.** Taken: `widthClass: PVWidthClass` with one `participantFoldLimit(_:)` switch (D-7).
- **X-1 · blocker · T-2's "1844 – 1849" value.** Fixed: {26,32}; "abt. 1817" with 4 Jun 1844 → {26,27} (T-2).
- **X-2 · blocker · The flag offset stated three ways.** Fixed with M-5: one anchor offset; T-4 gives 0, 120 and point − 35.
- **X-3 · blocker · Divider and padding owner.** Fixed with M-5: the item wrapper owns them, from placement (D-3, D-6).
- **X-4 · blocker · `PVCard` takes one padding.** Fixed with M-6.
- **X-5 · blocker · The line range and the placement scale.** Fixed: `PersonTimelineLifeline.range(_:)`, minimum 0.25 years, read by both; the no-range line is 18 to 58, length 76 (D-4).
- **X-6 · major · `origin`, `pointY`, `length` and the drawing.** Fixed: `origin` goes, the line starts at 18, both `Canvas` layers are `placement.height` tall, `Mark.pointY` changes in place (D-4).
- **X-7 · major · The misquoted rule.** Fixed with M-3.
- **X-8 · major · The fold limit's callers and width source.** Fixed with M-10: a defaulted `widthClass`, the section reads `pvWidthClass` in the host's `make`, previews pass `.wide` (D-7).
- **X-9 · major · `viewportHeight`.** Fixed: `PersonTimelineList` reads `pvPageViewportSize.height`; 0 gives a floor of 300 ÷ years (D-3).
- **X-10 · major · Fixture ages.** Fixed: the helper gains `age:`; every value listed in STEP-2.
- **X-11 · major · Handler test and Go order.** Fixed: the `me` lookup moves before the span loop; `person_detail_test.go` asserts `GetAge()` (D-2, T-3).
- **X-12 · minor.** Fixed: the card title font and the one-line lines slot (D-5); T-8 "no stub"; the 12pt gap and 16pt spacing (D-6, § 6); the TitleText README row in STEP-5; the stack is a placing `Layout` (STEP-5); Wide previews pass `.wide` (D-7).
- **X-13 · minor.** Fixed: T-4 and T-5 inputs and values; T-7 values per event; T-3 subtest `age`.

## Round 2

Fresh reviewers with the log; only new blockers asked for.

- **Maintainability: Fits with changes, no new blockers.** Its below-the-bar items, all Fixed: the S13-13 scope entry's "15 participants" is reworded in STEP-7 and swept; `Tether.isDashed`, `targets`/`Target`, the dashed-dot draw branch and `PersonTimelineItemView.isLast` are in **Deletes**; the flag's measures are named constants in placement (`flagWidth`, `gutterWidth`) and D-6's cap and D-9 read them; `datePlace` returns a private `placedDate`; STEP-1 names the matching tests.
- **X-20 · blocker · `y(t)` double-counts the birth stub; the death stub is ambiguous.** Fixed: `y(t)` is the shipped `map` (≥ for a birth stub, > for a death stub), every marker with no span has a stub, `length = 18 + (to − from) × pointsPerYear + 40 × stubs + 18`; T-8 gives `unknownEnds`'s length and points (D-4).
- **X-21 · blocker · A recorded marker with no span.** Fixed: anchored at its cap, a curve, a solid flag reading "?"; only Undated-group rows have no anchor, no curve and a dashed flag (D-4).
- **X-22 · major · The leap-day change breaks `lifespan_test.go`.** Fixed: STEP-1 names the file, the renamed row and the new ends.
- **X-23 · minor · T-3's values.** Fixed: the cases and values are in T-3.

**Writer's own calls (round 2):** a recorded but undated birth or death is tethered to its cap like a dated row.
