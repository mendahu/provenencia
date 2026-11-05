# Review — S14-32 Place relationship rows read from the far end

**Verdict:** Fits with changes — Ready with changes (round 1). Round 2 pending.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user

- **Far-end wording for `succeeded_by` (planning).** "Preceded by" (map, dogfood note) or "Succeeded" (Place page). Recommended "Preceded by". **Answered 2026-11-05: "Succeeded"** (D-1; working rule, bar 27 and scope entry reworded).
- **M-1 — one place-term wording table.** Fold the composer's three place-term switches and the planned fourth into one table with a Go guard, or keep a fourth switch because the term set is closed. Recommended one table. **Answered 2026-11-05: one table** (D-4, D-5, D-8, STEP-2, STEP-3).

## Writer's own calls

- New composer keys `containsWord` / `succeededWord` rather than reusing the Place page's `L10n.Conclusions` keys (D-3); the comments cross-reference them.
- The guard is a Go subtest in `TestSwiftVocabularyMatchesGo`, not a Swift test over a hand-kept list (D-8): it runs off a Mac and reads the same constants Go pins to the seed.
- The near end reads `PropertyTermDisplay` in this PR; the same bypass in the relationship and participation rows is logged, not fixed (Out).
- Version: patch; the implementer takes the next PATCH at merge time because S14-31 is also a patch (STEP-5).

## Round 1

### M-1 · blocker · A fourth hand-written switch over the two place terms; the place-term wording table is due
The composer words `part_of` / `succeeded_by` in `endTitles`, `readingSentence` and `CitationBridgeSentence.displayLabel`; the plan added a fourth (`placeFarEndLabel`) with no guard, and §6 Next adds undercounted the cost of a new term. Better: one per-term wording entry read by all four, plus a test that every seeded place term has one.
**Answer:** Fixed (user's answer, 2026-11-05) — D-4/D-5 add `CitationPlaceTermWording` and `placeTermWordings`; STEP-2 rewrites the four readers; D-8/STEP-3/T-4 add the Go guard; §6 Deletes and Next adds rewritten.

### M-2 · major · The two ends of one row get their label from different places
The near end read raw `term.label`; the far end new L10n copy. Better: near end through `PropertyTermDisplay.name(term:propertyKey:)`.
**Answer:** Fixed — D-6; T-1 compares the near end with `L10n.PropertyTerm.placeRelationshipPartOf` / `…SucceededBy`. The same bypass in the relationship and participation cases is an Out line (dogfood log).

### M-3 · minor · T-3 nearly copies an existing test
**Answer:** Fixed — the researcher-term case extends `aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel` (T-2); the missing-term case is in T-1.

### M-4 · minor · A full `ConnectionRow` literal again, plus overwriting `subjectsByID` by hand
**Answer:** Fixed — D-9: `termVocabulary` gains `endLabels`; `savedRow` helper for the new tests. Moving older literals is Out.

### M-5 · nit · Two keys for wording the plan says must match the Place page
**Answer:** Fixed — D-3 comments name `conclusions.place.contains` / `conclusions.place.succeeded`.

### X-1 · major · Version bump has no step and no number
**Answer:** Fixed — STEP-5 names the three places, 0.3.3 → 0.3.4 at the base, and the rebase rule for a collision with S14-31.

### X-2 · major · Xcode and string-catalog checks cannot run here
**Answer:** Fixed — STEP-4 says what to run and report off a Mac; the Go guard (STEP-3) and the Python checks run anywhere.

### X-3 · minor · Where the String Catalog entries go
**Answer:** Fixed — STEP-1 / §6 Strings: `containsWord` then `succeededWord`, immediately after `citationComposer.siteOfWord`, same shape.

### X-4 · minor · STEP-3 test construction leaves choices open
**Answer:** Fixed — STEP-4 gives `savedRow`'s full argument list in declaration order with `id: UUID()`, the term labels, term-id scheme and property id, and says to find groups by `subjectID`.

### X-5 · minor · Status field contradicts the plan's own readiness claims
**Answer:** Fixed — status set after review (see Verdict).

### X-6 · nit · The test section note
**Answer:** Logged — no change needed.
