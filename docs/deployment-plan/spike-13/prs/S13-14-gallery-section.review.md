# Review — S13-14 Gallery section

**Verdict:** Fits with changes — Ready with changes (rounds 1 and 2), every finding answered. The plan stays `Blocked — on board S13-D5`; it is reviewed again when the board's answers are copied in.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet). One reviewer per lens covered S13-D5, S13-14 and S13-15 together.

## To user
None.

## Writer's own calls
- The Event and Place picture tiles' VoiceOver label also gets the citation line and ", thumbnail" on the lead tile (D-3, D-4). Their look and click are unchanged (AC-1.4).
- The section is kind-blind, and both sections share one tile view, `GalleryTileView`, with new parameters (D-2, D-9). A board tile those parameters cannot draw is a stop condition, not a second tile.
- At Wide the Gallery fills the full width (D-8). The tile image keeps opening the Citation until S13-15 (brief § 3).

## Round 1

Findings are abridged; the reviewers' full text was returned to the writer.

### M-1 · blocker · The kind-blind section speaks Person-only strings
**Answer:** Fixed. D-4: `tiles(from:…)` takes the `leadTileLabel` and `openLabel` closures by kind; Event and Place get the `picture*` keys (§ Strings). The section reads only slot and tile words (D-2).

### M-3 · major · The citation parts are added one field at a time across two PRs
**Answer:** Fixed. D-3: `CitationLineDisplay.parts(_:)` returns one `Parts` value, `line(_:)` is rebuilt on it, and the tile carries one `citation` field. S13-15 reads the same field.

### M-4 · major · `fallbackColors` is passed beside the slot
**Answer:** Fixed. D-7: `ConclusionDetailGallery.fallbackColors` sits beside `fallbackMark`, `ConclusionDetailThumbnail.leading` reads it, and its `colors:` parameter goes (§ 6 Deletes; T-6).

### M-5 · major · The plan allows a private copy of the tile
**Answer:** Fixed. D-9: `GalleryTileView` becomes internal with `size`, `colors`, `help` and `captionIsMono` parameters. A second tile is a stop condition. The doc of `pictureCrop` is updated.

### M-6 · major · The inventory leaves out the tile's interaction kit pieces
**Answer:** Fixed. D-11 adds `PVFocusRingButtonStyle`, `pvActivatesOnReturn(when:)` and the hint; § 4 has a Tile button row; the brief's inventory has the matching row.

### M-9 · major · The tile-label Current state row is wrong
**Answer:** Fixed. § 4 Strings now quotes "Likeness from %@. Open citation". § Strings states the new defaults, with ". Open citation" moved to the hint (D-11), and the `Localizable.xcstrings` edit.

### M-11 · minor · T-3 repeats an existing test; the coexistence debt is not in the map
**Answer:** Fixed. T-3 is cut to the one-picture case. The map gains the working rule "The Person's gallery is redrawn on the Person page only", owned by the Event and Place redesign. Renaming the old section is Declined: it would touch Event and Place code this PR leaves alone (D-1), and the redesign deletes the section anyway.

### X-6 · major · Kind-blind section, Person wording
**Answer:** Fixed with M-1 (D-4). `%@` is the citation line in the tile labels and the Source title in the open labels (§ Strings).

### X-7 · major · The existing label wording conflicts with AC-1.2
**Answer:** Fixed. See M-9. T-4 pins the literal wording.

### X-8 · major · AC-1.4 contradicts D-3
**Answer:** Fixed. AC-1.4 now names the one VoiceOver change. T-5 *(new)* pins the Event label and nil colors.

### X-14 · minor · Fixture and test details
**Answer:** Fixed. T-1 names the fixture and the record fields to set. The whitespace-title edge is Declined: `titleIsRef` uses the same untrimmed `isEmpty` as `ReconciliationRecord.sourceTitle(for:)` (D-3), so the title and the flag never disagree. The line reference is now 387.

### X-15 · minor · Section header slot and icon type not named
**Answer:** Fixed. § 6 names `PVSectionHeader(title:meta:actions:)`. The icon cases are on `PVSymbol` (`PVIcon.swift` line 4), which `PVIconButton` takes.

### X-16 · minor · Checks an agent cannot run
**Answer:** Declined. Layout and focus across 1600 are checked by hand on this page, as for S13-12 and S13-13 (their § 10). The `primaryFill` path is S13-13's and already ships.

### X-17 · minor · Dangling review links
**Answer:** Fixed (this file).

## Round 2

Findings are abridged; the reviewers' full text was returned to the writer.

### M2-1 · major · `sourceTitle` stays beside `citation.title`
**Answer:** Fixed. D-3 deletes `ConclusionDetailPicture.sourceTitle`. `ConclusionDetailThumbnail.leading` and the old section's caption read `citation.title`. Added to § 6 Deletes; T-4 is updated.

### M2-2 · major · The picture keys keep ". Open citation" while the new section speaks it as a hint
**Answer:** Fixed. Both picture keys and `likenessTileLabel` drop ". Open citation", and both sections put it in the image's hint (D-9, D-11). AC-1.4 and T-5 quote the new strings.

### M2-3 · minor · Two keys with the same wording
**Answer:** Fixed. Both keys stay, since their arguments differ (the citation line and the Source title); `likenessTileLabel`'s comment says so (§ Strings).

### X2-5 · major · The new tile's caption text is not stated
**Answer:** Fixed. D-9 and § 6 macOS: `caption: tile.citation.title` in both sections.

### X2-6 · minor · The tile label is on two elements
**Answer:** Fixed. D-11: only the image button carries the label; the container is unlabelled `.contain`.

### X2-7 · minor · T-6 has no step; the menu's parameters; a short table row
**Answer:** Fixed. STEP-1 lists T-6; D-5 gives `pvRightClickMenu(title: nil, accessibilityIdentifier: "\(tileID).menu", items:)`; the Tile button row has its third cell.

## Change after review (user, 2026-11-03)
The tile's Open citation control is replaced by the Vitals rows' provenance control. It opens S13-11's `ProvenanceSheet` with only that likeness's record, and the record's links reach the Source, the Citation and the Observation. There is no direct Open citation anywhere on the tile (R-13b and R-32 updated in `person-detail.md`).
- D-5 builds each tile's one-record row: `ReconciledValueRowModel.narrowed(to:title:)` over the likeness field's own row.
- D-12 presents it with `ConclusionDetailRowsSection`'s sheet code; until S13-15, the image's click also opens it.
- `likenessOpenTip` is deleted. T-7 is new.

These Decisions have not been through `review-plan`. They are reviewed with the rest when the board's answers are copied in.
