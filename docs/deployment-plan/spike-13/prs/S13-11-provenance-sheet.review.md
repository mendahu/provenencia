# Review — S13-11 Provenance sheet

**Verdict:** Round 1: Fits with changes — Not ready. Every finding is Fixed or Declined below; both To user items are answered. Round 2: see below.
**Rounds:** 2. **Reviewers:** maintainability (fresh-context Explore agent), executability (fresh-context Explore agent). The brief's own review is [`../design/S13-D2-provenance-sheet.review.md`](../design/S13-D2-provenance-sheet.review.md).

## To user

- **M-7: value grouping in Swift or in Go.** Answered 2026-10-31: **Go groups them.** Go now decides which displayed value each record counts under (`GroupRank` in `conclusiondetails.finish`, wire `group_rank = 25`), and Swift only places headings. D-8 and D-13; the working rule is followed, so **Rules this plan bends** stays None.
- **M-9: "image 37" in the citation line versus "Page 37" in the composer.** Answered 2026-10-31: **differ, on purpose.** The composer names which page of the file is open; the citation line is a citation. D-17 records why.

## Writer's own calls

- From the board: drop a value line that would only repeat its owner line (frame 6h); keyboard focus follows reading order, not the board note's order (frame 6l); outside click closes, as the Swift kit's `.pvSheet` does.
- The records scroll capped at 560pt (`ProvenanceSheet.recordsMaxHeight`), as `PVDeleteImpactGroupStack` caps its list.
- The sheet gets `navigation` explicitly through `.environment(navigation)`.

## Round 1

### M-1 · blocker · `locator.Page` adds a second page-selector reader beside the one already in the package
**Answer:** Fixed. D-16: one unexported `citationPage` is extracted from `ValidateDepictionRegion` and runs after `validateDoc` with the at-most-one rule; `Page` and `ValidateDepictionRegion` both call it. T-8 adds the cases for two page selectors and for a page selector with no `artifact_page`.

### M-2 · major · Two Bool switches on the shared rows view, beside an enum that already names the split
**Answer:** Fixed. D-12: `ReconciliationRecordRows` takes `provenance: ReconciledValueRow.Provenance` and `groups:`. D-6 no longer changes the accordion's input; owner lines stay driven by data.

### M-3 · major · The citation-line cell has two homes
**Answer:** Fixed. D-18: the cell lives in `ReconciliationRecordRows` under `.sheet`; `ProvenanceSheet` keeps the header, summary, footer and presentation.

### M-4 · major · The row model keeps the raw field so it can build the summary later
**Answer:** Fixed. D-14: `provenanceSummary` is computed once in each factory, as `count` and `against` are; no field is stored.

### M-5 · major · `ValueGroup.Heading.value(text:support:)` copies `OtherValue`
**Answer:** Fixed. D-13: `Heading.value(ReconciledValueRowModel.OtherValue)`.

### M-6 · major · `docs/macos-client-patterns.md` becomes false
**Answer:** Fixed. STEP-8 updates its row disclosure paragraph, and the DoD names it.

### M-7 · major (To user) · Value bucketing in Swift against "data, not layout"
**Answer:** To user, answered: Go groups them (above).

### M-8 · minor · A second composer-location builder
**Answer:** Fixed. D-4: `composerLocation(for:focusingObservation:)`.

### M-9 · minor · Two words for one locator page
**Answer:** To user, answered: differ on purpose (above); D-17.

### M-10 · minor · A default no-op closure on `.sheet`
**Answer:** Declined. `Provenance` is also the input of `ConclusionDetailRowsSection` and `ReconciliationRecordRows` (D-12), and neither has a closure to give. The one `.sheet` caller of the row is the section, which always passes it (D-3).

### X-1 · blocker · D-13's value buckets contradict PRV-6 and frames 6c/6e for outvoted records
**Answer:** Fixed. D-13: Go keys each record on the displayed value it went into (kept or folded into a displayed value), else 0. T-11 covers the outvoted record whose `ValueRank` is 2.

### X-2 · blocker · The summary copy and its owner noun are not decided
**Answer:** Fixed. §6 Strings is now a table of every new key, with plural variations. D-14 adds `ownerKind: ProvenanceOwnerKind` to `derived`.

### X-3 · major · No API is named for the testable lookup in T-4
**Answer:** Fixed. D-2: `ProvenanceSheetItem.row(in:)`, cleared in `.onChange(of: rows)`.

### X-4 · major · The sheet's Wide switch, scrolling and environment are not specified
**Answer:** Fixed. D-2 reads `pvWidthClass` in the section and passes `navigation`. D-12 puts a capped `ScrollView` around the rows, with the headings above it.

### X-5 · major · The T-9 fixtures cannot express the case
**Answer:** Fixed. T-9 uses its own Source and Citation, and extends `TestConclusionDetailProtoCarriesVoteArtifactAndMembers`.

### X-6 · minor · D-4's `citationLocation` lists only part of its arguments
**Answer:** Fixed. D-4: through `composerLocation(for:focusingObservation: false)`; `sourcePage` gets `ref: nil`.

### X-7 · minor · The new test file and its target are missing from the step files
**Answer:** Fixed. STEP-5 and STEP-7 list `ProvenanceSheetTests.swift` and its pbxproj target entry.

### X-8 · minor · The header status says Blocked
**Answer:** Fixed. Status is set from this review's outcome.

### X-9 · minor · Hover colour
**Answer:** Fixed. D-18: `PVColor.textLinkHover`.

### X-10 · nit · The Codebase fit row for `ReconciliationRecord` leaves out `valueRank`
**Answer:** Fixed. The row lists `groupRank`, which replaced `valueRank`, and `citationLine`.

## Round 2

Pending.
