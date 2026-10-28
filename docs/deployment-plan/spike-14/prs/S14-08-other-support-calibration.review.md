# Review — S14-08 Other-link support calibration

**Verdict:** Fits — Ready
**Rounds:** 2. **Reviewers:** maintainability (inherit), executability (inherit)

## To user

None.


## Writer's own calls

- D-1 value **0.75** / leave `OtherCredit` (user option A).
- Calibration test in `graphalign` rather than only `segmentcompare`.
- Patch VERSION for researcher-visible bands.
- Inventory Start column updated to shipped value (S11-43 MinSimilarity pattern).

## Round 1

### M-1 · minor · Put the dogfood decision comment above the field in the `decided: user` form
**Plan:** STEP-1 trailing comment form.
**Codebase:** `alignment.go` `MinSimilarity` / `kinship.go` `MaxPivotEnd` use above-field `decided: user, …`.
**Answer:** Fixed — STEP-1 now requires the above-field `decided: user, 2026-10-28 (S14-08): …` form.

### X-1 · blocker · STEP-3 indexes nil `Patterns` and panics
**Plan:** STEP-3 assign into `cfg.Value.Names.Patterns[…]`.
**Codebase:** `Default().Value.Names.Patterns` is nil; empty NamePattern falls back to Western.
**Answer:** Fixed — STEP-3 / D-3 drop Patterns/NamePattern; use `givenProp` / `nameProp`; assert bands + `OtherSupport == 0.75`.

### X-2 · major · AC-1.3 is not locked by any named test
**Plan:** T-3 → `TestSegmentSection` (positive only).
**Answer:** Fixed — T-1 asserts `OtherSupport == 0.75`; T-3 row removed.

### X-3 · minor · STEP-5 under-specifies four `MARKETING_VERSION` lines
**Plan:** “both configurations”.
**Answer:** Fixed — STEP-5 names all four `MARKETING_VERSION` lines in `project.pbxproj`.

## Round 2

No new blockers. Round-1 Fixed answers verified against code (`givenProp`/`nameProp`, nil `Patterns`, Western fallback, four `MARKETING_VERSION` lines). Revised STEP-3 implementable without further decisions.
