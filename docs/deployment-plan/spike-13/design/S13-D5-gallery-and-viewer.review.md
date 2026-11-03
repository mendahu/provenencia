# Review — S13-D5 Gallery section and likeness viewer (brief)

**Verdict:** Fits with changes — Ready with changes (round 1), every finding Fixed.
**Rounds:** 1. **Reviewers:** maintainability (Opus; UI rows Q1, Q3, Q5, Q6), executability (Sonnet; Briefs), shared with the S13-14 and S13-15 plan reviews.

## To user
None.

## Writer's own calls
- The brief recommends the kit sheet for the viewer (LV-1). The cap of the image at its pixel size ÷ 2 is drawn on the board for the user to accept or change (LV-2).

## Round 1

### M-6 · major · The inventory leaves out interaction kit pieces
**Answer:** Fixed. New inventory rows: Tile image button (`PVFocusRingButtonStyle`, `pvActivatesOnReturn`), Viewer panel (`PVPanel` in `pvSheet`), and Viewer image surface (`surfaceSunken`).

### M-7 · major · Form (B) would be a new modal presenter
**Answer:** Fixed. LV-1 recommends (A) and says (B) needs a new page-level presenter. The viewer's inventory row says the same.

### X-19 · minor · Wide empty and Wide large frames; viewer inventory; viewer title
**Answer:** Fixed. GA-W3 *(new)* covers the empty state at Wide; frame 10b adds the 40-likeness and empty frames at Wide; the § 4.3 row cites GA-W3. The viewer's inventory rows are added (M-6). The title is a plan decision (S13-15 D-9) the board may restyle, since it is inside the kit Panel's title row.

## Change after review (user, 2026-11-03)
GA-4 is now the provenance control and a one-record provenance sheet (new frame 2b). LV-4 puts the record row under the viewer's image. Open citation is dropped everywhere. The inventory gains the provenance control, the provenance sheet and the record rows.

## Change after review (user, 2026-11-03, later)
Tiles get no provenance control, no one-record sheet and no context menu. A likeness's provenance is its one record row under the viewer's image (the kit's Record row, S13-D13 PSK-6). Until S13-15 ships, a tile's click opens its Citation, as today. S13-14 no longer depends on S13-36, but S13-15 does. R-13b and R-32 are updated in `person-detail.md`. This is reviewed again on the board refresh.
