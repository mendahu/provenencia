# Review — S13-12 Family section

**Verdict:** Fits with changes — Not ready (round 1); every finding Fixed. The plan stays `Blocked — on board S13-D3, and on S13-04 landing`; round 2 runs when the board is agreed and its answers are in the plan.
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet). Both read the S13-04 branch @ `b12a7fe` with the plan docs copied in.

## To user

- None. The three product calls were asked before the plan was written (2026-11-01): family tiles open the relative on Family; children in birth order, sorted in Go inside S13-12; recorded and implied drawn alike.

## Writer's own calls

- D-3: the person goes into the children of their parent and sibling units in Go, and `GetPersonFamilyResponse` gains the person's summary (a proto field). It is the only way to place the focus in birth order without a Swift date key.
- D-6: the tile leaves out a missing half ("d. 1851"), while the header keeps "place unknown". The brief and the plan agree; the board can overrule.
- D-8: one `ConclusionPageSection.queryKey` mapping, which Timeline and Sources extend.
- D-12: line geometry (partner line, stem, bus, drops).
- D-13: defaults until the board: units stacked, leading rows, focus raised and repeated per unit.

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
