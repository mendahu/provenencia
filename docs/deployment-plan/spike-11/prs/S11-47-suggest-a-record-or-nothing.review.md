# Review — S11-47 Promote suggests a record or nothing

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

Round 1's blockers (M-1, X-1, X-3) were each answered Fixed with plan text checked against the code at `a70932e`. No round 2: the fixes take the shape the reviewers proposed, and the writer checked each against the code (`noteDuplicates` resets and re-reads `proposedDuplicateOf` after every `select`; the guard regex reads `case "":`; the pb.go / pb.swift comment lines).

## To user

- Answered before review (2026-10-27): an unplaced row's Updated caption reads "No record suggested" (D-4).

## Writer's own calls

- D-1: no Go `Target` const for the empty row target (Swift types it as `PromoteTargetCode.unplaced`).
- D-2: a named `ReasonUnplaced = ""` instead of a bare empty reason.
- D-10: the alike-New set is decided New plus unplaced rows with `ReasonUnplaced`. Taken and empty rows stay out, as they did when they were Skip.
- D-13: the row target becomes a typed Swift wire code in this PR. The send side (`Fixed`, batch rows) stays strings, Out.
- D-5: the unknown-code fallback stays `.empty`.

## Round 1

### M-1 · blocker · Type the row target in Swift
`.cursor/rules/macos-client.mdc`: wire vocabulary is spelled once as a typed enum with a guard row. `CatalogPromoteGraphAlignmentRow.target` is a raw `String`, compared at `PromoteFlow.swift` 694/733/748/756 and spelled in `FakeStore` 1799/1816/1830. This PR changes its meaning with no guard.
**Answer:** Fixed — D-13 adds `PromoteTargetCode` (decoded in `GoStore`, guard row "promote row targets" with `""` listed literally), §3 In, STEP-4, STEP-5, T-7, T-8. The encode side (`Fixed`, batch rows) is an Out line with its reason: the guard reads decode cases only.

### M-2 · major · Name the "would be a new record" set once
**Answer:** Fixed — D-10 `wouldBeNew`, used for both `i` and `j`; taken and empty rows excluded by name.

### M-3 · major · T-1 duplicates `TestRowReasons` fixtures
**Answer:** Fixed — D-16: T-1 is `TestRowReasons` with a `wantTarget` column and a "taken" row. T-4 covers only medium and strong.

### M-4 · major · Swift's default reason vs the wire's zero value
**Answer:** Fixed — D-5: `CatalogPromoteGraphAlignmentRow.reason` defaults to `.unplaced`. The alternative default and the unknown-code fallback stay `.empty`, with the reason given (Go's `alternative` always sets a reason). `GenealogyStore.swift` is added to STEP-4.

### M-5 · minor · Record the row-reason path in `add-match-rule`
**Answer:** Fixed — D-15, STEP-6.

### M-6 · minor · Comments made false
**Answer:** Fixed — D-14 (`Target.unset`, `targetsChosen`, `PromoteFlowTests` line 191, `rows.go`, `alikeNew`).

### X-1 · blocker · `duplicateNotesNameTheOtherRowAndItsRef` breaks
**Answer:** Fixed — STEP-5 sets rows `c` and `d` to New before the `.alsoNew` assertions. AC-2.6, T-9, plus a stop condition.

### X-2 · major · STEP-5's fallback check could not print what it said
**Answer:** Fixed — every fixture line is listed (including 830–832 and 904–910), and the check names the lines that should remain.

### X-3 · major · Which rows enter the alike check
**Answer:** Fixed — answered by M-2's `wouldBeNew` (D-10). The "scores 0" claim is gone.

### X-4 · major · Hand-edit wrapping of the proto comments
**Answer:** Fixed — D-6 gives the exact lines for `engine.proto`, `engine.pb.go` and `engine.pb.swift`.

### X-5 · minor · Swift checks cannot run off macOS
**Answer:** Fixed — STEP-7 and the risks table say to report the Swift half as not run; the reviewer runs check-pr on macOS.

### X-6 · minor · Commit boundary leaves the guard red
**Answer:** Fixed — STEP-1 to STEP-5 are one commit.

### X-7 · minor · An alternative's `""` reason would decode to `.unplaced`
**Answer:** Fixed — D-5 and §4 "Alternatives": Go never sends `""` on an alternative.

### X-8 · nit · Small inaccuracies
**Answer:** Fixed — the `TestLifespanRulesOutAPerson` AC reference was dropped, the `TestGracieObituaryGolden` first message was added, the rename list was rewritten in T-1, and STEP-1's vet set includes `promotealign`. The Status/Planned wording is as the template intends.
