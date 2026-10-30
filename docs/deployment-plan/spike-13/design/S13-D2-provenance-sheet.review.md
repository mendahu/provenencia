# Review — S13-D2 Provenance sheet

**Verdict:** Fits with changes — Not ready (round 1); every finding Fixed or answered below, no blocker open
**Rounds:** 1. **Reviewers:** maintainability (fresh-context Explore agent), executability (fresh-context Explore agent). Round 2 is folded into the S13-11 plan's review, which runs once the board is agreed.

## To user

- **The sheet's form is now bound, not left to the board** (M-5, X-12): a modal sheet at both widths, with no docked panel at Wide. S13-D1 had left Vitals' secondary column "for S13-D2's provenance sheet". The Event and Place pages, which reuse the sheet, have no secondary column, and a docked form would need a selected-row state and a presenter that switches at 1600 for one caller. Recommendation: keep the sheet. If you want a docked inspector at Wide, it goes back to the board as an option.

## Writer's own calls

- No new wire field: a Citation has no title on the wire, so the Citation link reads "Citation". A title would be a follow-up engine PR.
- A link closes the sheet, then navigates. Back lands on Vitals with the sheet closed, because sheets are not places.
- The summary line says only what the row knows (X-1): a Property row's state and disagree note, the record count, and on a life row the owner count. "Stated empty · 1 record denied" is dropped.
- Value groups cover Mixed and multi-valued rows. Every record with no value rank, negative records included, goes under "Other records" (X-5).
- The group-line rule follows the code: a line when there are several owners or the one owner is not the shown one (X-4).

## Round 1

### M-1 · blocker · Two record renderers that stay side by side indefinitely
The sheet's table and `ReconciliationReasoningView`'s grid would be two copies of the ink, face, outcome and reserved-column rules until the accordion is deleted.
**Answer:** Fixed. PRV-5 binds the table layout (option (b), one block per record, is dropped). The inventory adds "Record rows — Extend (extracted)": one rows view that the accordion's Card and the sheet both wrap. The plan's D-12 adds `ReconciliationRecordRows` *(new)*, extracted from the grid.

### M-2 · major · Summary and value lines need to reuse the existing count and state formatters
**Answer:** Fixed. Inventory row "Summary and value wording — Extend": `spokenState`, `againstLine`, `listedValues` and `otherValues`. PRV-4 and PRV-6 say the lines reuse the row's words. Plan D-14.

### M-3 · major · Value groups need to extend the existing record-group model
**Answer:** Fixed. Inventory row "Record groups — Extend": the grouping is done in the row model. Plan D-13 adds `ReconciliationRecordGroup.values`, tested by T-5.

### M-4 · major · The three links need to come from `ReconciliationRecord`, not the sheet
**Answer:** Fixed. Inventory row "Record — Extend". It was already plan D-4, which now also carries `valueRank`.

### M-5 · minor · Option (B) adds a second presenter without a named caller
**Answer:** Fixed. The form is bound to the sheet (domain fact "Form", PRV-3, plan D-11). It is listed under To user so you can reverse it.

### X-1 · blocker · The PRV-4 summary line has no source for life rows and no copy at all
**Answer:** Fixed. New domain fact "What a row can say about itself". PRV-4 sets out the summary per row kind and allows no other wording. Plan D-14 and T-7.

### X-2 · blocker · The inventory leaves out the model files the sheet needs
**Answer:** Fixed. The inventory now has rows for "Record" and "Record groups" (see M-3 and M-4).

### X-3 · major · Where the sheet's presentation state lives
**Answer:** Fixed. The "Sheet presenter" row now says one presenter per section, keyed by row id. PRV-11's "gone" means no row with that id, or no records left. Plan D-2.

### X-4 · major · PRV-7 contradicts the code on single-owner group lines
**Answer:** Fixed. PRV-7 now states the code's rule. PRV-8 marks the shown owner only when there are group lines. Plan D-15.

### X-5 · major · PRV-6 value groups don't say where some records go
**Answer:** Fixed. PRV-6 now covers Mixed and multi-valued rows, puts negative records under "Other records" (new copy, named), and shows no value lines when there is one value. Plan D-13.

### X-6 · major · PVPanel can't hold the PRV-4 header as specified
**Answer:** Fixed. The "Sheet chrome" row places the value, Badge and summary at the top of the body, above a ScrollView capped at the window's height, with the close control in `titleAccessory` or the footer. PRV-11 adds that the header stays put while the records scroll.

### X-7 · minor · Two domain facts are wrong
**Answer:** Fixed. Outcome phrases are now quoted lowercase, as in `L10n.swift`. "No evidence" is corrected to "no usable value", the "provisional" phrase is corrected to "provisional member", and the facts note that "no usable value" has no mark of its own.

### X-8 · minor · Record order isn't what the brief says
**Answer:** Declined, but the wording is clarified. Go sorts the outcomes: `core/database/conclusiondetails/conclusiondetails.go` `Field.Outcomes` is commented "by value rank (none last), then Observation id", and Swift keeps that order. The new domain fact "Order" states it.

### X-9 · minor · Frame 7 is self-contradictory
**Answer:** Fixed. Frame 7 and the domain fact now list two records: the denied one, and the negative record that denied it, which reads "kept".

### X-10 · minor · "Ghost" IconButton tone doesn't exist
**Answer:** Fixed. PRV-1 now says "neutral tone" (`PVIconButtonTone.neutral`).

### X-11 · minor · Are the refs spoken?
**Answer:** Fixed. PRV-13: the record element speaks the Subject ref, and each link names its target ("Observation OBS-7QH3D").

### X-12 · minor · Option (B) needs a selected-row state
**Answer:** Fixed by M-5: option (B) is gone.

### X-13 · nit · Sample refs reuse one token
**Answer:** Fixed. The Subject sample is now `PER-3HW8C`.
