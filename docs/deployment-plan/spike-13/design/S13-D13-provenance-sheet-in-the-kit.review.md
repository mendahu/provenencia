# Review — S13-D13 The provenance sheet in the design system (brief)

**Verdict:** Fits with changes — Ready with changes (round 1). Every finding is Fixed or answered by the user.
**Rounds:** 1. **Reviewers:** one fresh-context reviewer, maintainability (UI rows) and executability (Briefs).

## To user
- M-1 (answered 2026-11-03): the Event and Place Why accordion keeps its own copy of the record table until those pages are redesigned onto the kit sheet.

## Writer's own calls
- The content shape (§ 2) and the link callback ("record id, and Source, Citation or Observation"). The caller closes the sheet, then navigates, so the kit names no `WorkspaceLocation`.
- The one-record variant has no value line and no Read as column (PSK-4).
- The viewer's sheet takes the provenance sheet's widths, 880 and 1280 (PSK-6; S13-15 D-4).
- PSK-4 and PSK-6 are previews only in S13-36.
- The copy namespace `L10n.ProvenanceSheet`.
- Whether `ReconciledValueDisplay`'s record wording moves to `Platform/` is left to the S13-36 plan. The recipe takes display-ready strings either way.

## Round 1

Findings are abridged; the reviewer's full text was returned to the writer.

### M-1 · blocker · Moving only the sheet style forks the record table
**Answer:** To user. Answered (b): the accordion keeps its copy, as recorded debt (§ 1, § 3, non-goals; the map's working rule).

### M-2 · blocker · "Display-ready content" is undefined
**Answer:** Fixed. § 2 names the content: title, values with faces, Badge, summary, groups, owner lines, value headings and records. It also gives the link callback and who closes the sheet and navigates. § 6 names `PVProvenanceSheetContent` via the recipe row.

### M-3 · major · The inventory mislabels or misses pieces
**Answer:** Fixed. New rows: group lines, citation link, outcome mark, value faces, copy button, copy namespace. The Badge list is Merged, Mixed, Concluded.

### M-4 · minor · Type names, copy namespace and ids
**Answer:** Fixed. `PVProvenanceSheet`, `PVProvenanceRecordTable`, `L10n.ProvenanceSheet`, and an id prefix the caller passes.

### M-5 · minor · Say why this is a recipe
**Answer:** Fixed. § 1 "Why a recipe"; `ProvenanceSheetItem` stays with the caller (§ 6).

### X-1 · major · PSK-6 cannot be drawn at a known width
**Answer:** Fixed. The host Panel takes 880 and 1280, and S13-15 D-4 follows.

### X-2 · major · PSK-4 leaves a slot choice open
**Answer:** Fixed. There is no value line and no Read as column, and both are optional slots (PSK-1, PSK-5).

### X-3 · major · What ships without a caller
**Answer:** Fixed. § 3: PSK-4 and PSK-6 are previews only in S13-36.

### X-4 · minor · M-1's choice is not in § 3
**Answer:** Fixed. It is in § 3 "Stays out".

### X-5 · nit · Name the specimen's file
**Answer:** Fixed. PSK-1 names *Provenance Sheet.dc.html*.
