# Review — S15-D14 Source page Artifacts

**Verdict:** Fits with changes — Ready with changes; every blocker Fixed
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- Trim #802 before merge so its per-Artifact chrome never lands? **Yes** (user, 2026-11-12).
- Draft board as its own preparatory PR (S15-36)? **Yes** (user, 2026-11-12).
- Thumbnail: square crop or keep shape? **Square crop** (user, 2026-11-12).
- Add field in the sheet: dialog on the sheet or inline row? **Inline row on every board, the Source's own included** (user, 2026-11-12); now SA-13 and the rule "A board adds a field in an inline row".

## Writer's own calls
- Clearing a row in the sheet needs no confirm; Cancel restores it (SA-8).
- Rows are the default list layout; compact cards only as a drawn alternative (SA-1).
- The sheet is 560pt wide (between the graph form's 520 and room for the board's label column).
- An entry shows at most four copy details then "+n more"; graphs by name, first two then "+n" (SA-3, SA-4).
- The census type's Artifact suggestions on the board: Provider, URL, Accessed, Image number (board data; the seed is S15-19's).
- The edit sheet shows no file row (SA-6).
- An Artifact deleted or moved elsewhere closes its sheet with a new "Artifact no longer here" Toast (SA-12).
- Thumbnail cap 80pt, not 128pt (round 2 M2-1): the largest square the 256-pixel thumbnail fills sharply at 2x; follows the user's "just display a smaller thumbnail", so no larger derivative.
- Escape closes an inner row edit before the sheet; Return in a row confirms the row (SA-15), owned by the kit.
- The Source Metadata section's header Add field button opens the inline row; its empty state loses its button; the sheet's board shows no empty state (SA-13).
- Rows have no hover of their own.

## Round 1

### X-1 · blocker · Add field and row edit/clear in a draft sheet are undecided
**Answer:** Fixed — SA-8 (a row's ✓, fill, dismiss, clear, reorder and add change the draft only; no clear confirm), SA-13 (inline Add field, user), frame 5.

### X-2 · blocker · Suggestion counts contradict between frames
**Answer:** Fixed — Domain facts name the four census suggestions; frames 4 and 9 agree.

### X-3 · major · Missing data for frames
**Answer:** Fixed — §5 Data names the graphs, the twelve Artifacts, the cover, the PDF, the fileless one, and the 8-value set.

### X-4 · major · `ART-0615` is not a real ref shape
**Answer:** Fixed — `ART-7QK2M` (`core/ref` `PrefixArtifact` + five-character token).

### X-5 · major · Wide behavior conditional, no threshold
**Answer:** Fixed — SA-W1: copy details sit in a right column of the row at both widths; Wide widens it. Cards keep their size and add columns.

### X-6 · major · Board's-call items that change scope are unbounded
**Answer:** Fixed — rows are the default (SA-1); visible actions named (SA-5); description two lines (SA-3); no empty line for copy details (SA-4).

### X-7 · major · Sheet scrolling vs `PVFormDialog`
**Answer:** Fixed — inventory: `PVFormDialog` Extend with a scrolling body and fixed footer (also M-5); width 560 (SA-6).

### X-8 · major · Add flow contradictions
**Answer:** Fixed — SA-7: File first, optional; Add enabled once Label has text; graph boxes locked after a refused file; SA-6: no file row in edit mode.

### X-9 · major · SA-8 / SA-12 edge cases
**Answer:** Fixed — SA-8 says a value put back counts as unchanged; SA-12 now says what happens when the Artifact goes while the sheet is open.

### X-10 · minor · "As today" copy
**Answer:** Fixed — §2.1 Copy table.

### X-11 · minor · Thumbnail sizing
**Answer:** Fixed — square `.tile`, one size up to 128pt (user: square crop; also M-4).

### X-12 · minor · Scope and gate mismatches
**Answer:** Fixed — §3 moves SA-4 into S15-34 (map K-37), so cards show copy details between S15-34 and S15-35; Move stays S15-25's.

### X-13 · nit
Logged.

### M-1 · blocker · A per-row refusal needs a field the engine does not name
**Answer:** Fixed — S15-33 makes `Apply` refuse with `ErrInvalid.WithParams(fieldID)`; S15-36 maps it to the row; SA-9 keeps the message under the row.

### M-2 · major · The sheet should copy the Evidence graph form's shape
**Answer:** Fixed — inventory: the sheet's model follows `SourceGraphFormModel`; SA-6 adds the graph locks and their load states.

### M-3 · major · Draft board behavior is left to the board
**Answer:** Fixed — SA-8, SA-13; the board uses `.regular`; `.compact` leaves with the #802 trim.

### M-4 · major · Keeping the image's aspect has no kit mode
**Answer:** Fixed — user chose the square crop; `.tile`.

### M-5 · major · The sheet's scrolling needs a kit change
**Answer:** Fixed — `PVFormDialog` Extend; second caller S15-11.

### M-6 · major · A second copy of the board's value display
**Answer:** Fixed — one metadata value view pulled out of `displayValue`; URL as `PVButton(.link)`.

### M-7 · minor · Name the right-click menu piece
**Answer:** Fixed — `.pvRightClickMenu` + `PVActionMenuItem`; Set cover reuses `L10n.Sources.useAsThumbnail`.

### M-8 · minor · Remove rows miss what the sheet makes dead
**Answer:** Fixed — inventory Remove rows for card edit mode and the Add Artifact dialog.

### M-9 · nit · Entry hover
**Answer:** Fixed — `PVHoverEffect` for a row, `pvLiftHover` for a card.

## Round 2

Fresh maintainability (Opus) and executability (Sonnet) reviewers. Round-1 answers checked against the code; two did not hold (M-6's `PVButton(.link)`, M-9's `PVHoverEffect`) and are corrected below.

### X2-1 · blocker · SA-12 relies on a notice that does not exist for an Artifact
**Answer:** Fixed — new Toast "Artifact no longer here" in §2.1; SA-12; frame 7.

### X2-2 · blocker · What "Link to a graph" does once card edit goes
**Answer:** Fixed — SA-14: it and "+n more" open the sheet at their section, in S15-34.

### X2-3 · major · No Wide frame for the inline Add field row
**Answer:** Fixed — frame 11a / 11b, the side column at 480pt.

### X2-4 · minor · Baseline unclear (disclosure, `.compact`)
**Answer:** Fixed — §3: S15-34 starts from #802 as trimmed.

### X2 minors (graph lock in frame 4, PDF and fileless Artifact, ref in title, empty ✓, hover)
**Answer:** Fixed — frame 4 names the graphs; Sheet 19 PDF, Sheet 24 no file; title "Edit ART-7QK2M"; ✓ disabled until filled; no hover-only action.

### M2-1 · blocker · A square crop of the 256-pixel thumbnail is not sharp at 128pt
**Answer:** Fixed — cap 80pt in §1, SA-2, the inventory, the map rule and S15-35.

### M2-2 · blocker · Escape and Return have two owners in the sheet
**Answer:** Fixed — SA-15; inventory `PVFormDialog` / `PVSheet` Extend; S15-34 scope.

### M2-3 · blocker · The inline row leaves the other Add entry points undecided
**Answer:** Fixed — SA-13: the header button opens the row, the Source section's empty state loses its button, the sheet has no empty state; `MetadataBoardStyle` leaves with the trim; frame 13.

### M-6 (round 1) did not hold · `PVButton(.link)` changes the board's value font
**Answer:** Fixed — the shared value view keeps today's mono text and plain link button.

### M-9 (round 1) did not hold · `PVHoverEffect` is button chrome
**Answer:** Fixed — rows have no hover of their own.

### Minors (graph-keeps message, opt-in scrolling)
**Answer:** Fixed — SA-9 keeps the message naming the graph; the scrolling body is opt-in.
