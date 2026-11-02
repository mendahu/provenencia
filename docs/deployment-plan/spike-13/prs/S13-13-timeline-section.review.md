# Review — S13-13 Timeline section (and brief S13-D4)

**Verdict:** round 1 Fits with changes — Ready with changes; round 2 Fits with changes — Not ready, every finding Fixed or answered by the user; none open. After round 2 the card work moved to S13-34 and S13-35 (below); the refreshed plan has not been re-reviewed. Status `Blocked — on S13-33, S13-35`.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet).

## To user

- **M-1: the person tile and the layering rule.** Answered 2026-11-02. The user wants the tile in the design system, but only as a content-agnostic shell: "The components that should be promoted into the design system should be content-agnostic. They should just be the little shells that have slots for content to go in, and then the callers can inject the content they need … The design system can have the basic components and things like the gender colors. At the repository level, we can also have shared helpers, like just basic things that resolve values, but the content should live at the call sites." A preparatory PR came first: S13-32 "The person tile becomes a kit shell". **Superseded the same day:** the board (turn 12) drew compact participant cards, which the user kept ("I decided to make them smaller … if we don't need to do that anymore, you can cancel those PRs"). S13-32 is descoped; the tile stays in `Features/Conclusions` with two sizes (D-4); the working rule is now "One person tile, two sizes, in the feature". Its shared pieces moved back into this plan: the term-name overload (D-5), the Return modifier (D-10).
- **Board answers (2026-11-02).** Compact cards kept; participants ordered by role, then name, in Go (S13-33 D-10); full width at Wide (D-17, which bends Wide rule 2: STEP-9 rewords it in § 9 and the add-design-brief Paste block together).
- **Round 2 M-14: one term-name rule.** Answered 2026-11-02: unify in this PR. Vitals rows, place titles and the gender badge read terms through `PropertyTermDisplay` (D-18, STEP-2); a relabelled built-in term shows the product name in Vitals and on the badge (AC-7.1).
- **Round 2 M-11: the compact card's years.** Answered 2026-11-02: reuse the Persons list's life span at year precision ("1820 –", not the board's "b. 1820") (D-4, D-20).

## Writer's own calls

- D-6: the timeline speaks an event without its ref (`EventsList.spokenLine`); the ref adds noise on a person's own timeline.
- D-11: participants sit in an adaptive `LazyVGrid` at `PersonTile.Size.compact.minimumWidth` (the Gallery's precedent), not in Family's column formula.
- D-7: the board's 72pt event picture over the Events list's 44pt.
- D-6: a fold covers the dated rows before the death marker; the death marker, the undated group and the rows after death always show (the board drew no folded death).
- D-15: a repeat click on a point re-scrolls by resetting the target to nil first.
- D-12: Family moves onto a shared section host in this PR, not in a preparatory one: the host is about 70 lines, and Family is its first caller.
- D-14: the accessibility-id scheme.
- (Superseded) M-2, M-4 and M-8 were placed in S13-32's scope; with S13-32 descoped they are answered in this plan again (round 2 M-20).

## Scope change after round 1 (user, 2026-11-02)

The user asked for a lifeline: a line to scale over the life, each event pointing to its place, gaps shown as empty stretches. Answers: an unknown birth or death is a placeholder with no invented date, placed by which event types may come before birth or after death (matching's lifespan window pads 110 years, so it is not reused); every event sits at its own date; undated events sit between the birth and death markers (R-24 changed, R-24a added); gaps are not labelled. Positions are Go's (new map row S13-33, Timeline positions, using `valuecompare.DateEnds`); the plan gains US-5, D-15, D-16, T-8, T-9. Round 2 reviews this text.

## Scope change after round 2 (user, 2026-11-02)

The user asked to hoist the person card into the design system, regular and compact, because a family tree in another feature will reuse it, and to decouple display from content first. An evaluation of `PersonTile` (2026-11-02) found that the card and the display pieces it is built from sit in `Features/`; the map re-cut the work as two preparatory PRs (working rule "The person card lives in the design system; display is decoupled from content"). This plan now builds on them and no longer carries:

- **To S13-34** (display pieces move down a layer): D-5's `PropertyTermDisplay.name(_:propertyKey:)` overload and D-18's one term rule for Vitals and the badge (with the former US-7, its AC-7.1 and AC-7.2, STEP-2, T-1, T-12); D-20's year precision on `DateRowDisplay`. Round 2 M-11, M-14 and M-22's resolver half are answered there.
- **To S13-35** (the person card joins the design system): D-4's compact size, `years`, `LifeLine.text` and `helpText` (STEP-4, T-2); D-10's `pvActivatesOnReturn`; D-19's focus-ring style, now `PVFocusRingButtonStyle` in the kit (round 2 M-18). Round 2 M-21's minimum width is `PVTileCardSize.compact.minimumWidth`.
- **Kept here, renumbered:** role badges (D-4), items (D-5), picture (D-6), folds (D-7), fold row (D-8), card grid (D-9, now `PVPersonCard(size: .compact)`), host (D-10), fixtures (D-11), ids (D-12), lifeline (D-13), placeholders (D-14), full width (D-15), and the event row (D-16, now `PVTileCard(size: .row)` with the title button in its title slot). Tests T-3 – T-11 became T-1 – T-9; Steps went from STEP-0 – STEP-9 to STEP-0 – STEP-7.
- **Rules this plan bends** no longer lists the tile rule or the term-name change: the prep PRs carry them.

A round 3 (or a fresh round on the three plans together) should check this refresh against S13-34 and S13-35.

## Round 1

### M-1 · blocker · `PVPersonTile` in `Recipes/` would import the Conclusions feature
The tile reads `ConclusionListRow.title`, `ConclusionTitleSource`, `ConclusionDatePlaceLine`, `PersonTitleDisplay`, `PersonLifeDisplay`, `PersonGenderTreatment` and `L10n.Conclusions.family*`; `docs/design-system-layers.md` says lower layers do not import higher ones. This conflicts with the map rule "One person tile, built once".
**Answer:** To user → answered (see To user). D-4 rewritten; S13-32 added to the map.

### M-2 · blocker · `PersonTimelineRole.label` is a second term-label rule
`PropertyTermDisplay.name(key:origin:propertyKey:catalogLabel:)` already falls back to the key; T-5 contradicted it.
**Answer:** Fixed. D-5 adds `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey:)` in this PR (re-answered in round 2: S13-32 is descoped); `PersonTimelineRole` is gone. The older rule in `ReconciledValueDisplay.string(for:)` is unified in this PR (round 2 M-14, D-18), not logged.

### M-3 · blocker · `Row.spokenLabel` is a second spoken label for an event
**Answer:** Fixed; reshaped in round 2 (M-19): D-6 composes `EventsList.spokenLine` with the roles; T-7 pins it.

### M-4 · major · The title button has no owner; Return handling would be a third hand copy
**Answer:** Fixed; re-answered in round 2 (M-18): the Return modifier is added in this PR (D-10) and the title's style is `ConclusionFocusRingButtonStyle` (D-19).

### M-5 · major · D-11 copied Family's column formula
**Answer:** Fixed; re-answered in round 2 (M-21): the minimum width is `PersonTile.Size.compact.minimumWidth` (D-4, D-11). Executability X-2 is closed by the same fix.

### M-6 · major · The section host is copied from Family, with S13-16 as the third copy
**Answer:** Fixed. D-12 `ConclusionPageSectionHost`; Family moves onto it in STEP-2.

### M-7 · major · Two copies of the preview and test data
**Answer:** Fixed. D-13 `PersonTimelinePreviewData`, shared by the previews and the tests, with people from `PersonFamilyPreviewData.summary`. Closes X-5.

### M-8 · major · The shared tile keeps Family-named keys
**Answer:** Declined (re-answered in round 2): a compact card is never the focus, so the Timeline never reads `familyThisPerson` / `familyThisPersonSpoken`; the keys keep their one reader (Family's focus tile). The Out row says so.

### M-9 · minor · `Participant` nearly copies `PersonFamilyContent.Tile`
**Answer:** Declined (re-answered in round 2): `Participant` adds a `caption` that Family's `Tile` never has, so a shared struct would carry an optional field one caller never sets; both hold the same `PersonTileContent`. Codebase fit says so.

### M-10 · minor · Brief: the Event title status, and TL-6's role on the tile
**Answer:** Fixed. The title row is Ship. TL-6 allows only a caption above the tile.

### X-1 · blocker · The review log did not exist
**Answer:** Fixed: this file.

### X-2 · blocker · D-11's width rule could not be applied
**Answer:** Fixed with M-5.

### X-3 · major · STEP-4 listed L10n files with nothing to add
**Answer:** Fixed. STEP-4 adds no key; the keys are in STEP-1 and STEP-5.

### X-4 · major · Spoken label and role joins only half decided
**Answer:** Fixed. D-6 fixes the composition; TL-12's wording ("as Subject") joins the frame-dependent list as § 12 question 9; T-7 tests the reused part.

### X-5 · major · The fixture was not specified
**Answer:** Fixed with M-7.

### X-6 · major · Manual check data for 41 participants and 120 events
**Answer:** Fixed. AC-3.1 and AC-3.2 cite T-6 and the "Large event" and "Long life" previews; manual check 6 uses the previews.

### X-7 · major · Accessibility ids collide
**Answer:** Fixed. D-14.

### X-8 · minor · The crumb form differs from the brief
**Answer:** Fixed. The brief's TL-8 now reads *Persons › Weber, Anna › Timeline*, the sorted title, as the code gives.

### X-9 · minor · The recipe move crosses a layer boundary
**Answer:** Fixed with M-1.

### X-10 · minor · STEP-1's file list
**Answer:** Fixed: the tile move left this plan (S13-32).

### X-11 · minor · `PVCard`, `PVBadge`, `PVSectionHeader` not checked
**Answer:** Logged. They depend on the board; STEP-0 checks them when the frames are fixed.

### X-12 · minor · `sectionQueryKeys` asserts the other sections are nil
**Answer:** Fixed. STEP-3 says `.timeline` leaves the nil list.

## Round 2

Reviewed the plan rewritten after the board (turn 12), S13-32's descoping and the S13-33 plan. Maintainability: Fits with changes. Executability: Not ready (two blockers). All answered below; S13-33's final names (Planned) re-checked against § 4, D-6 and D-15: `dated` is computed from `span`, a recorded marker is exactly the header's birth or death, a non-canonical header birth is a placeholder; no `HasEventDate` mention remains.

### M-11 · blocker · `PersonLifeDisplay.years` and the `lifeYears*` keys are a second life-span rule
**Answer:** To user → answered 2026-11-02: reuse the Persons list's span. D-20 adds `DateRowDisplay.Precision.year` (qualifiers and ranges kept); D-4's `years` is `dateSpan` at year precision; the three keys are dropped; T-2 gains "abt." and one-sided cases.

### M-12 · blocker · The row order is worked out twice
**Answer:** Fixed. D-6 builds one `items` sequence (markers, rows, undated header, more rows) with the folds applied; the view and `PersonTimelineLifeline` read it; the folded-death, undated-death and undated-birth cases are decided; T-5 and T-6 assert the sequences.

### M-13 · blocker · A second scroll mechanism and an unowned viewport
**Answer:** Fixed. D-15: clicks set `scrollTarget` for `pvScrollsRowIntoView(_, anchor: .top)`; `PVPageColumns` publishes `pvPageVisibleRect` from the state it already measures; frames are measured in named spaces. T-9 covers tethers against `visible`.

### M-14 · blocker · Two term-name rules
**Answer:** To user → answered 2026-11-02: unify in this PR. D-18, STEP-2, T-12, AC-7.1, AC-7.2; the structural-debt Out row is removed; the Objective names the visible change.

### M-15 · major · The brief and map still describe S13-32
**Answer:** Fixed in the planning commit: the brief's Depends, Ships, TL-6, TL-W1, TL-W2, Wide table, frame 13, inventory (tile, title, secondary column row removed) and Out rows, and the map's S13-D4 track row, S13-12 Context and S13-13 Ships; the DoD sweep adds "One person tile, built once" and `S13-32`.

### M-16 · major · D-17 bends Wide rule 2; one `group` cannot serve columns and band
**Answer:** Fixed. Rules this plan bends names rule 2 with the user's answer; STEP-9 rewords it in § 9 and the add-design-brief Paste block together; `columnWidths` returns a `band`; T-10 asserts the band stays 1180.

### M-17 · major · The host takes key and glyph instead of the section
**Answer:** Fixed. D-12: `ConclusionPageSectionHost(section:entityId:…)` derives the key, glyph and id root; Family's hand-built key goes.

### M-18 · major · The title's button style is not chosen
**Answer:** Fixed. D-19: `ConclusionFocusRingButtonStyle`, moved from the tile, for the tile and the title.

### M-19 · major · `includesRef:` is a mode Bool and joins speech with the visual separator
**Answer:** Fixed. D-6: `EventsList.spokenLine` with `spokenJoin`; `accessibilityLabel` unchanged.

### M-20 · minor · Round-1 answers rested on S13-32
**Answer:** Fixed: To user records S13-32's descoping and the user's answers; M-1, M-2, M-4, M-5, M-8 and M-9 are re-answered above.

### M-21 · minor · The card's minimum width belongs on the card
**Answer:** Fixed. `PersonTile.Size.minimumWidth` (D-4), read by D-11 and by `PersonFamilyLayout` in place of its literal 120.

### M-22 · minor · "Subject" needs the seeded constant and the origin check
**Answer:** Fixed. D-5 `isSubject`; T-7 has a researcher's term keyed `subject`.

### X-13 · blocker · The viewport source for tethers
**Answer:** Fixed with M-13 (D-15 Coordinates).

### X-14 · major · A new `ScrollViewReader`
**Answer:** Fixed with M-13: one mechanism, `.top` anchor, nil-then-id reset, unfold-then-scroll ordering, cap clicks scroll to their marker item.

### X-15 · blocker · The row order is not decided by the model
**Answer:** Fixed with M-12 (D-6 Order and Folds).

### X-16 · major · STEP-4's check needed a later step
**Answer:** Fixed. `fillsWidth` turns true for `.timeline` in STEP-5 with T-3; STEP-1 adds it all false.

### X-17 · major · T-8's tick step contradicted D-15
**Answer:** Fixed. T-8: 1-year ticks at about 43pt a year; `lise` heights given (4000pt → 1-year, 1000pt → 5-year); `ernst` 1200pt.

### X-18 · major · A recorded marker's spoken label decided twice
**Answer:** Fixed. D-6: a recorded marker reads its row's label; "Born …" is dropped; T-7.

### X-19 · major · The tooltip text was not specified
**Answer:** Fixed. D-4: `LifeLine.text` and `PersonTileContent.helpText`, with T-2's example.

### X-20 · major · The title's style, font and line limit
**Answer:** Fixed. D-19: `PVFont.display(size: PVTypeScale.h3)`, two lines, only the title is the hit area; the date at `PVTypeScale.body` through `EventSecondaryLine(dateSize:)`.

### X-21 · major · `primaryFill` and the three initializers
**Answer:** Fixed. D-17: all three initializers, the band layout on `band`, ignored with a secondary or an unbounded proposal; T-10.

### X-22 · major · Empty roles and `isSubject`
**Answer:** Fixed. D-5, D-6: no Badges and no "as" clause; nil caption and spoken role; `isSubject` defined; AC-1.10, T-7.

### X-23 · minor · Fixture details
**Answer:** Fixed. D-13 lists each fixture's facts (the rest is the writer's, tests check only named properties); the brief's sample now carries Johann's death 18 Sep 1880 and burial 22 Sep 1880; T-5's event list is unambiguous.

### X-24 · minor · Smaller items
**Answer:** Fixed: the compact card's lift (D-4); the fold row's style (D-9); the host's signature (D-12); the caption's join and nil (D-5); no new years keys (M-11); the Events list tests are `PersonsListTests` (§ 4, § 8); STEP-7's check includes the repeat-click scroll; the fold reset is the section's `.onChange`, as Family (§ 8 note).

