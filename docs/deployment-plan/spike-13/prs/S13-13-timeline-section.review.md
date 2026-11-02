# Review — S13-13 Timeline section (and brief S13-D4)

**Verdict:** Fits with changes — Ready with changes (round 1, on the draft blocked on S13-32 and the board)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). Round 2 runs after S13-32 lands and the board is agreed (plan STEP-0).

## To user

- **M-1: the person tile and the layering rule.** Answered 2026-11-02. The user wants the tile in the design system, but only as a content-agnostic shell: "The components that should be promoted into the design system should be content-agnostic. They should just be the little shells that have slots for content to go in, and then the callers can inject the content they need … The design system can have the basic components and things like the gender colors. At the repository level, we can also have shared helpers, like just basic things that resolve values, but the content should live at the call sites." A preparatory PR comes first. **Applied:** map row and scope entry S13-32 "The person tile becomes a kit shell"; the working rule "One person tile, built once" is rewritten; the plan depends on S13-32 (D-4, STEP-0).

## Writer's own calls

- D-6: `EventsList.accessibilityLabel` gains `includesRef:` so the timeline speaks an event without its ref. The ref adds noise on a person's own timeline; the board can ask for it back.
- D-11: participants sit in an adaptive `LazyVGrid` at the tile's declared minimum width (the Gallery's precedent), not in Family's column formula.
- D-12: Family moves onto a shared section host in this PR, not in a preparatory one: the host is about 70 lines, and Family is its first caller.
- D-14: the accessibility-id scheme.
- M-2, M-4 and M-8 are placed in S13-32's scope (the term resolver, the Return modifier, the tile's key names), because S13-32 lands first and the tile shell owns them.

## Scope change after round 1 (user, 2026-11-02)

The user asked for a lifeline: a line to scale over the life, each event pointing to its place, gaps shown as empty stretches. Answers: an unknown birth or death is a placeholder with no invented date, placed by which event types may come before birth or after death (matching's lifespan window pads 110 years, so it is not reused); every event sits at its own date; undated events sit between the birth and death markers (R-24 changed, R-24a added); gaps are not labelled. Positions are Go's (new map row S13-33, Timeline positions, using `valuecompare.DateEnds`); the plan gains US-5, D-15, D-16, T-8, T-9. Round 2 reviews this text.

## Round 1

### M-1 · blocker · `PVPersonTile` in `Recipes/` would import the Conclusions feature
The tile reads `ConclusionListRow.title`, `ConclusionTitleSource`, `ConclusionDatePlaceLine`, `PersonTitleDisplay`, `PersonLifeDisplay`, `PersonGenderTreatment` and `L10n.Conclusions.family*`; `docs/design-system-layers.md` says lower layers do not import higher ones. This conflicts with the map rule "One person tile, built once".
**Answer:** To user → answered (see To user). D-4 rewritten; S13-32 added to the map.

### M-2 · blocker · `PersonTimelineRole.label` is a second term-label rule
`PropertyTermDisplay.name(key:origin:propertyKey:catalogLabel:)` already falls back to the key; T-5 contradicted it.
**Answer:** Fixed. D-5 uses `PropertyTermDisplay.name(_ value: CatalogConclusionValue, propertyKey:)`, which S13-32 adds; `PersonTimelineRole` is gone; T-5 and Edges now show the key. The older copy in `ReconciledValueDisplay.string(for:)` is logged to `docs/ideas/structural-debt.md` in STEP-7, and S13-32 moves `PersonGenderTreatment` off it.

### M-3 · blocker · `Row.spokenLabel` is a second spoken label for an event
**Answer:** Fixed. D-6 composes `EventsList.accessibilityLabel(event, orders:, includesRef: false)` with the roles; T-7 pins it.

### M-4 · major · The title button has no owner; Return handling would be a third hand copy
**Answer:** Fixed. D-10 names the title's style and uses the kit's Return modifier, which S13-32 adds and moves the tile and the fold row onto (S13-32 scope); D-9 uses it too. The brief's inventory names the style.

### M-5 · major · D-11 copied Family's column formula
**Answer:** Fixed. D-11: an adaptive grid at the minimum width declared on the tile shell (S13-32). Executability X-2 is closed by the same fix.

### M-6 · major · The section host is copied from Family, with S13-16 as the third copy
**Answer:** Fixed. D-12 `ConclusionPageSectionHost`; Family moves onto it in STEP-2.

### M-7 · major · Two copies of the preview and test data
**Answer:** Fixed. D-13 `PersonTimelinePreviewData`, shared by the previews and the tests, with people from `PersonFamilyPreviewData.summary`. Closes X-5.

### M-8 · major · The shared tile keeps Family-named keys
**Answer:** Fixed by moving the work: S13-32's scope renames the tile's "this person" keys. The Out row says so.

### M-9 · minor · `Participant` nearly copies `PersonFamilyContent.Tile`
**Answer:** Declined for now. S13-32 decides the person content type both hold, so a shared "tile with a destination" is reconsidered there; the plan says so under Codebase fit.

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
