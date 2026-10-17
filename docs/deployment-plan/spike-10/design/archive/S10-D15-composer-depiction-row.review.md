# Review — S10-D15 Citation composer: depiction row

**Verdict:** round 1: Fits with changes — Not ready → every finding answered below. Round 2 runs with the S10-34 plan once the board is agreed.
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None open. The product calls (redraw on edit, the PDF page rule, Duplicate copies the region) were answered on 2026-10-17 and are in §2.

## Writer's own calls
- How drawing works is decided in the brief, not left to the board: Draw steps the dialog aside the way a custom term does; a finished shape or Cancel/Escape brings it back. The board draws the chrome only. The reviewers asked for this (X-1, M-2), and it matches the S10-34 plan's D-7.
- Sizes: the preview is 96pt and the saved-line thumbnail 32pt, both cover-fit, so the line keeps its height.
- After a save the dialog stays open with the property kept and the person cleared.

## Round 1 — maintainability

### M-1 · blocker · Keyed on the Likeness property instead of the `depiction` value type
**Answer:** Fixed — CDR-1, the inventory row, and the copy name the value type or the property's label. A non-goal forbids hard-coded "likeness" in controls.

### M-2 · major · The step-aside mechanism exists; state it
**Answer:** Fixed — "How drawing works" at the top. "Another flow" is a non-goal.

### M-3 · major · The Viewer row hides who owns the change
**Answer:** Fixed — the rows are split: Viewer (Ship, API unchanged) and Region target (`CitationComposerModel`, Extend).

### M-4 · major · Citation-only controls stay live while drawing
**Answer:** Fixed — CDR-3 and the "Clear menu and Set page" row: disabled in drawing mode.

### M-5 · major · No kit piece for the drawing hint
**Answer:** Fixed — the hint is `PVCallout` (info, or warning on the wrong page) with a ghost `PVButton` Cancel. A local banner is a non-goal.

### M-6 · minor · The saved-line row names only the view file
**Answer:** Fixed — Home names `CitationSavedLine` and `lineRow`.

### M-7 · minor · S10-D12 still describes the old depiction model
**Answer:** Declined for this brief — logged as an Out line in the S10-34 plan. The S10-35 plan reads S10-D12 and corrects it there.

### M-8 · nit · Three names for one PR
**Answer:** Fixed — the brief, the plan, and the branch all use `composer-depiction-row`.

## Round 1 — executability

### X-1 · blocker · The core interaction is left to the board
**Answer:** Fixed — see M-2.

### X-2 · blocker · No control ends drawing
**Answer:** Fixed — the inventory has Cancel drawing, and CDR-3 covers mouse-up commit, Escape, and the rectangle default. The Clear menu is disabled.

### X-3 · blocker · CDR-3 contradicts CDR-10 on the page
**Answer:** Fixed — §2 "PDF pages": a new Citation's page is set from the region, its region is unchanged, a discard does not undo the page, and the locator line shows the page (frame 8).

### X-4 · major · "Negated badge" is wrong
**Answer:** Fixed — Negative badge plus the italic *not*.

### X-5 · major · Saved-line thumbnail placement and fit
**Answer:** Fixed — CDR-7: 32pt, cover-fit, in place of the value text, tap opens Edit, "—" with no image.

### X-6 · major · "Next face" undefined
**Answer:** Fixed — CDR-6 and §2: the dialog stays open, as it does today; closing it ends the run.

### X-7 · major · PDF page states undefined
**Answer:** Fixed — §2 "PDF pages" and CDR-10: the viewer moves to the page, a shape on another page is not taken, and the warning copy and the no-page Callout placement are given.

### X-8 · major · Draft edge cases
**Answer:** Fixed — §2 "Changing the draft".

### X-9 · major · Edit and Duplicate scope
**Answer:** Fixed — §2 "Edit and duplicate": dirty includes the region, and a property change clears the value as for any property. §2 "A region the engine hides" and CDR-14 cover a page mismatch.

### X-10 · minor · Map and brief differ on which properties
**Answer:** Fixed — the brief now offers any depiction property, matching the map.

### X-11 · minor · Preview and Callout states loose
**Answer:** Fixed — CDR-4 gives 96pt, cover-fit, and the spoken label. The render is synchronous, so there is no loading state. CDR-10 says where the no-page Callout sits.

### X-12 · nit · Copy and fixtures
**Answer:** Fixed — the placeholder glyph is named, all three error strings are quoted, and frame 1 lists the seeded person properties.
