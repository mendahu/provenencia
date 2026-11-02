# Review — S13-12 Family section

**Verdict:** Round 1 Fits with changes — Not ready; round 2 (after the board was agreed) Fits with changes — Not ready. Every finding of both rounds is Fixed; none is open. The plan is `Planned`.
**Rounds:** 2 (the review-plan cap). **Reviewers:** maintainability (Opus), executability (Sonnet), fresh each round. Round 1 read the S13-04 branch @ `b12a7fe`; round 2 read `main` with S13-04 merged and the board's answers in the plan.

## To user

- None open. The three product calls were asked before the plan was written (2026-11-01): family tiles open the relative on Family; children in birth order, sorted in Go inside S13-12; recorded and implied drawn alike. The board was agreed by the user on 2026-11-01.

## Writer's own calls

- D-3: Go returns the person's own summary and the origin siblings (each sibling once, the person among them, the board's FC-3 order) as two new response fields, so the merge and order stay in Go (round 2 M2-3).
- D-5: the focus tile is the kit's raised, elevated card with the "This person" Badge and a semibold name; the board's darker inset fill is not drawn, because no `PVCard` tone has it and the brief rules out a new card style for one caller.
- D-6: the tile leaves out a missing half ("d. 1851"), as the board draws; the header keeps "place unknown".
- D-12: the board's `layTree3` measures, written out as one pure layout; the rail rule generalises the board's two sets to any number.
- D-13: the board's two open notes (long places cut on narrow tiles; the right column top-aligned) are taken as drawn.
- D-16: the artifact card's hover lift moves to the kit and is shared with the tile.
- D-1: the tile opens the relative through `PersonsList.location`, so the crumb shows the same title as any Person link (the board's mock shows "Anna Weber").

## Round 1

### M-1 · blocker · Declare each page section's query key once, not a preference tunnel plus per-section branches
**Answer:** Fixed — D-8 is `ConclusionPageSection.queryKey`, read by `PlaceRegistry`, `ConclusionDetailLoaded` and the section; the preference, its `@State` mirror, old T-10 and its stop condition are gone. STEP-7 records the pattern in `macos-client-patterns.md` (M-10).

### M-2 · blocker · A third loading/failed/content mechanism
**Answer:** Fixed — D-10 makes `ConclusionDetailPresentation.init` generic over the read's value; the section uses it.

### M-3 · blocker · "Missing half left out" forces a second builder and view
**Answer:** Fixed — D-6 adds `PersonLifeDisplay.Missing` (`.stated` / `.omitted`) to the one builder and to `ConclusionDatePlaceLine`; T-4. The brief's wrong "as in the page header" sentence is corrected.

### M-4 · blocker · `BornBefore` re-implements `eventKeyLess`
**Answer:** Fixed — D-2: `eventKey` / `eventKeyLess` renamed `dateKey` / `dateKeyLess`, `birthKeyOf`, `SortByBirth` beside `SortByTitle` with keys computed once; T-1 tests `SortByBirth`.

### M-5 · major · The tile's hover and press have no owner
**Answer:** Fixed — D-5 `PersonTileButtonStyle` copies `SourcesGraphButtonStyle`; the brief's inventory gains a Tile interaction row.

### M-6 · major · Two-line names need a `lineLimit` parameter
**Answer:** Fixed — D-5: `ConclusionListRow.title` gains `lineLimit: Int = 1`; the brief's Name row is Extend.

### M-7 · major · Build the tile's location with `PersonsList.location`
**Answer:** Fixed — D-1, T-7, AC-3.1 (sorted title in the crumbs). The Evidence graph's hand-built copy is an Out line.

### M-8 · minor · Read the session from the environment
**Answer:** Fixed — D-11. `PersonDetailSections` gains only `entityId` (X-3).

### M-9 · minor · The focus tile's inked border has no kit option
**Answer:** Fixed — D-13 and the brief: `PVCard(tone: .raised, elevated: true)`.

### M-10 · minor · Record the pattern in `macos-client-patterns.md`
**Answer:** Fixed — STEP-7 and the DoD.

### M-11 · nit · `familyTileAccessibility` repeats `a11yList`
**Answer:** Fixed — the key is gone; D-7 composes with `a11yList`.

### X-1 · blocker · `ConclusionListRow.title` cannot draw two lines
**Answer:** Fixed — as M-6.

### X-2 · blocker · The focus's place among siblings needs a birth sort Swift lacks
**Answer:** Fixed — D-3: Go places the person among the children before D-2's sort; AC-1.1, AC-1.4, T-2, T-3.

### X-3 · blocker · The section cannot get the focus summary or the entity id
**Answer:** Fixed — D-3 (`CatalogPersonFamily.person`), D-11 (`entityId` passed to `PersonDetailSections`); `person` nil draws no focus (Reach, T-6).

### X-4 · major · `LifeLine` names and the wrong file in STEP-2
**Answer:** Fixed — STEP-2 lists `PersonDetailView.swift` (`PersonDetailLifeLines`) and not `ConclusionDetailPage.swift`. The name stays `PersonLifeDisplay.LifeLine`: it lives beside `PersonLifeDisplay.Line` (the list's one-line span) in the same type, and `PersonLifeLine` is the list's view; the doc comment on each says which.

### X-5 · minor · The review log is not in the branch
**Answer:** Fixed — this file.

### X-6 · blocker · T-2 cannot be built with the fixtures
**Answer:** Fixed — D-15 `Canon.Date`; T-2 is its own test with its own ordered display; the zero-query case stays in `TestPersonFamilyQueryCount`, named in STEP-1's check.

### X-7 · major · T-1's expected order for `ABT 1820`
**Answer:** Fixed — the case is dropped; T-1 lists concrete orders and the equal-key tie.

### X-8 · major · Tile content, role type, spoken label and group labels underspecified
**Answer:** Fixed — D-5 (`PersonTileRole`), D-7 (every member named; spoken label composition; group labels, "Partner %@ and children"; the focus's label), D-1 (the location's title), § 6 Strings (`a11yLifeLineKnown`).

### X-9 · major · Refresh and section states leave choices open
**Answer:** Fixed — D-8 (b): only a handle that holds a value and is fetching shows in the switcher; D-10 ids for the error and empty states; `Tile.id` includes the unit index; the preference and its test are gone (M-1).

### X-10 · blocker (board) · Layout choices D-13 leaves open
**Answer:** Fixed where the board is not needed — D-12 names the line geometry, D-13 the card options and leading rows, D-7 the `.oneParent` focus (alone above the children). The rest stays § 12's open question, which is why the plan is Blocked on the board.

### Order and other points
**Answer:** Fixed — L10n keys are added in the step that first uses them (§ 6 Strings); STEP-1 names the generated files, and the header says it needs `protoc`.

## Refresh after S13-04 landed (2026-11-01)

S13-04 merged in #671 (`main` @ `be51621`). Since the planning base `b12a7fe`, only `core/kinship/families.go` changed: past `Rules.MaxPivotEnd` parents, a sibling sits with the first parent it shares. Every § 4 path and symbol was checked on `main`. The S13-04 Landed note (5) says that at or under the bound a sibling can sit under more than one parent set; the plan's Edges "Assumptions" line, T-6 and the brief's domain facts and frame 3 said each relative appears once, and now say a sibling can repeat. The status drops "and on S13-04 landing"; the plan stays Blocked on the board. No Decision changed.

## Board agreed (2026-11-01)

The user agreed the S13-D3 board (turn 11, frames 11a – 11q) and attached the project archive. The writer rendered every Family frame in Chromium (React and Babel served from npm; unpkg is not reachable here) and checked them against the brief: FS-1 – FS-14 and FS-W1 – FS-W3 are each drawn or answered. The board replaced the brief's stacked units with two columns (FC-1 – FC-5); the plan's D-5, D-6, D-7, D-10, D-12, D-13, § 6 macOS (frames in words), Strings and tests were rewritten to it, and § 12 was closed.

## Round 2

### M2-1 · blocker · A second grow-on-hover mechanism beside `ArtifactCardHover`
**Answer:** Fixed — D-16 moves `ArtifactCardHover` and `GrownHitShape` to `pvLiftHover(scale:resting:)` in `DesignSystem/Components/HoverEffect/`; the artifact card and the tile both use it; the tile's button style keeps only the focus ring. The brief's Tile interaction row names it.

### M2-2 · blocker · `PersonTileRole` and the spoken label are Family-shaped
**Answer:** Fixed — D-5: `PersonTileContent` takes `role: String?` and composes its own spoken label and the "this person" suffix; `PersonFamilyRole` lives in `PersonFamilyContent.swift`; T-5 holds the label cases.

### M2-3 · blocker · D-7 merges and reorders siblings in Swift, against "Family groups in Go"
**Answer:** Fixed — option (a): D-3 has Go return `origin_siblings` (each once, with its `parent_sets`, the person among them, FC-3 order); D-7 only lays it out; T-2 is the Go table test. No working rule is bent.

### M2-4 · blocker · No source for `wide`
**Answer:** Fixed — D-12 takes `widthClass: PVWidthClass`, read from `\.pvWidthClass`; T-12 runs width 1180 at both classes.

### Below the bar
**Answer:** Fixed — stale `PVFlowLayout` rows gone; a line with both halves speaks through the existing `a11yLifeLine`, only a one-half line uses `a11yLifeLineKnown`; `familyGroupThisPerson` dropped for `familyThisPerson`. The writer's calls above are refreshed.

### X2-1 · blocker · Spoken "born" and "died" are capitalised in the code
**Answer:** Fixed — D-6 and Strings: `.omitted` speaks with new `a11yBornInline` / `a11yDiedInline` ("born", "died"); the header keeps `a11yBorn` / `a11yDied`.

### X2-2 · blocker · Existing Go tests break when D-3 lands
**Answer:** Fixed — D-3 no longer puts the person into the units, so their members are unchanged; `kinWorld.show` sorts names, so the new child order does not show (§ 6 Edges). STEP-1's check names the three tests, which change only at `PersonFamily`'s call sites.

### X2-3 · major · Rail rule for a later parent set
**Answer:** Fixed — D-12: a set's rail runs to the last sibling in that set; it draws horizontals only to siblings not in an earlier set, whose horizontal crosses it; a set with no own sibling still draws its rail to its last shared sibling.

### X2-4 · major · The bus and a group with no partner
**Answer:** Fixed — D-12: the bus spans the minimum to the maximum of the focus's centre and every group's join y; a group with no partner joins at its first child's centre; its children sit at `c2 + 32`; one child joins straight from the bus, more through the rail. T-12 checks the bus and Paul.

### X2-5 · major · `siblingsCaptioned` and the lone focus
**Answer:** Fixed — D-7 `siblingsCaption` needs a sibling besides the focus; D-12 places a lone focus at (20, 26); T-8 and T-12 check 11h.

### X2-6 · major · A flat layout cannot give VoiceOver groups
**Answer:** Fixed — D-12's layout returns groups (label, frame, tiles in group coordinates) in reading order; the view puts one container per group.

### X2-7 · major · `PersonTileContent` undefined
**Answer:** Fixed — D-5 lists its init and fields, the natural title style, the thumbnail's inputs, and `Equatable` on every content type.

### X2-8 · major · The Swift `person` field breaks call sites; no preview data
**Answer:** Fixed — `person = nil` and `originSiblings = []` defaults keep `FakeStore` and `WorkspaceSessionTests` compiling; § 6 macOS puts the previews' `CatalogPersonFamily` values in `PersonFamilySection.swift`.

### X2-9 · minor · Height, clipping, semibold, fonts
**Answer:** Fixed — D-5: `.frame(height: 92)` and `.clipped()`; `ConclusionListRow.title` gains `emphasized`; fonts are the kit's, stated.

### X2-10 · minor · Sources for the section's inputs
**Answer:** Fixed — § 6 macOS: a nil handle draws `.loading`; locale `.autoupdatingCurrent`, orders `session.nameFormat`, `projectDir` from `session.projectKey`; `widthClass` from the environment (M2-4).

### X2-11 · minor · The bracket at x 40
**Answer:** Fixed — D-12: the crossing is drawn as it falls, as on the board.

### X2-12 · nit · Status header
**Answer:** Fixed — the plan and the map row are `Planned`.
