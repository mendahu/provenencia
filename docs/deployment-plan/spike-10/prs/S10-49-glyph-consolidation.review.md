# Review — S10-49 Glyph consolidation

**Verdict:** Fits with changes — Ready with changes (applied)
**Rounds:** 1. **Reviewers:** maintainability (opus; the first run failed on a network error and was rerun), executability (sonnet)

## To user

- None open. The scope (all three idea parts, the T-2 trim, one PR) was the user's call on 2026-10-14.

## Writer's own calls

- S10-49 is the next free ID. It sits in Slice 5 after S10-19.
- D-7: the Properties All card's grid symbol goes from a 14pt medium-weight font glyph to `PVIcon` at 14pt in the same 16pt frame. The weight shift is accepted.
- Left out, with the reason recorded in the archived idea: the omnibar lead symbols for Source types and Metadata fields still differ from their sidebar glyphs, because that is a visible change the user has not chosen.
- No round 2. Round 1 had one blocker (X-1, a check that could not pass), fixed mechanically and tested against the planning base. It caused no reshape.

## Round 1

### M-1 · major · Check the Swift subject marks against Go with the existing drift test
**Answer:** Fixed. D-2 and T-7 add a `TestSwiftVocabularyMatchesGo` row. Swift's `subject_*` cases must equal the registry's `IconSymbol`s.

### M-2 · major · The FakeStore edit reintroduces the string rule for a field nobody reads
**Answer:** Fixed. The `FakeStore` change is dropped (D-2 and §3 Out). The `rg` patterns no longer match `L10n.swift` (X-1).

### M-3 · minor · Evidence kinds trade a compiler-checked switch for a string lookup
**Answer:** Fixed. The evidence switches stay (D-3 and §3 Out). T-3 now only checks that they agree with the rule.

### M-4 · minor · The omnibar lead builds the Conclusion row thumbnail a second way
**Answer:** Fixed. D-6 calls `ConclusionListRow.thumbnail(mark:)`.

### M-5 · minor · Dead `source` branch; omnibar symbol drift
**Answer:** Fixed. The dead branch is removed (D-6). The drift goes to the archived idea's Leftovers (STEP-9), not this PR.

### M-6 · minor · The add-a-subject-type skill does not learn the mark step
**Answer:** Fixed. STEP-9 adds a paragraph to `add-seeded-vocabulary`, and Next adds is corrected.

### M-7 · minor · Archiving as Done drops the sibling per-type switches
**Answer:** Fixed. STEP-9 writes a Leftovers section in the archived note.

### M-8 · nit · `leadThumbnail(kind:iconKey:)` breaks the siblings' signature
**Answer:** Fixed. The function is now `leadThumbnail(for hit:)`.

### X-1 · blocker · STEP-2 and AC-2.4 `rg` checks cannot print what the plan says
**Answer:** Fixed. Both patterns now match only built `"subject_\(` strings and the four bridge literals. They were tested at `e150a76a`: 20 lines before the PR, so "prints nothing" is a real post-condition.

### X-2 · major · The Properties All card changes size and weight
**Answer:** Fixed. D-7 switches on the glyph to keep the 16pt named mark and a 14pt symbol in a 16pt frame. The weight shift is named and accepted.

### X-3 · major · Unwrapping `PVMarkKey(subjectTypeKey:)` in D-6 is not stated
**Answer:** Fixed. D-6 uses `?? .fallback`.

### X-4 · major · `source` case and `ConclusionListRow.thumbnail` unresolved
**Answer:** Fixed. D-6 has no `source` case. `ConclusionListRow.thumbnail(mark:)` keeps its signature and gets a new body.

### X-5 · major · STEP-9 breaks the S10-19 plan's links
**Answer:** Fixed. STEP-9 updates the four links, and its check requires no `ideas/glyph-consolidation` path outside the archive.

### X-6 · major · Plan and map edit are uncommitted
**Answer:** Fixed in process. The orchestrator commits the plan, this log, and the map row as the first commit on the PR branch before dispatch, and records it in the ship log.

### X-7 · major · Plan status is Draft
**Answer:** Fixed. The status is now Planned, after this log.

### X-8 · minor · Composer row omits lines 321 and 463
**Answer:** Fixed in §4.

### X-9 · minor · `pbxproj` wiring ID scheme not given
**Answer:** Fixed. D-10 names the group, the Sources phase, and the `add-swift-test` ID scheme.

### X-10 · minor · The STEP-3 `rg` misses a multi-line `Content(`
**Answer:** Fixed. STEP-3 and the stop condition use `rg -U` with `\s*`.

### X-11 · nit · Temptations
**Answer:** Logged. §3 Out and D-4 already close each one.
