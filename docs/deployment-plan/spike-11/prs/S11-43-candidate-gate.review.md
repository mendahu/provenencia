# Review — S11-43 Lower the candidate gate

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (fresh agent, strongest model), executability (fresh agent, implementing model)

No blocker is left open after round 1. M-1 was the only blocker, and it was fixed by moving the new fixture to `namevaluetest`. That change is a fixture-home choice, not a change to the plan's shape, so no round 2 was run.

## To user

None.

## Writer's own calls

- The registry comment and the inventory row call 0.4 "decided: user, 2026-10-24 (S11-43)", not an estimate (D-1, D-2).
- T-2 asserts that the row is not placed, has no band, and has the record as its one alternative. It does not assert Skip `weak`, so S11-47 can change unplaced rows without touching it (D-4).
- Tests are in `core/graphalign` only, with no `promotealign` test (D-3).
- The `name-candidates.md` case 1 table ("Candidate at 0.5") is kept as the diagnosis (D-5).
- `namevaluetest.Given` is added, but the four private given-only builders in database and catalogsession tests are not migrated to it (Out).

## Round 1

### M-1 · blocker · `givenVals` is another private name builder; the shared `namevaluetest` package should get the given-only builder
The plan added `givenVals` to `align_test.go`. `core/namevalue/namevaluetest` is the shared fixture home, and four packages already hand-roll a given-only name. The "Next adds" claim did not hold: S11-46 tests in `core/valuecompare` and S11-45 in `core/valuecompare/name_test.go`, and neither can import a `graphalign_test` helper.
**Answer:** Fixed. D-6 now adds `namevaluetest.Given(form)`, and `givenVals` is dropped. The handles use `Western` through `person` / `nameProp`. Migrating the existing private builders is an Out line with its reason: they build values with no `Form` and with a pattern.

### M-2 · major · T-1 ignores the fixtures `TestCandidates` already uses
**Answer:** Fixed. T-1 now uses the subtest's `person`, `metas` (`subjectcomparetest.Metas()`) and `refs`. Similarity under those metas was re-measured at 2 / 4.5 and 0.375 (2026-10-27).

### M-3 · minor · T-2 adds another "one Subject against one handle" setup; lift `alignOne`
**Answer:** Declined, D-7. Lifting the closure edits an existing test, and AC-1.4's evidence is that no existing test changes. The inline setup is four lines. S11-47, which rewrites those rows, is the natural place to lift it.

### M-4 · minor · The docs sweep copies the number into prose that names knobs by symbol
**Answer:** Fixed. D-5: `matching.md` and `promote-graph-alignment.md` name `alignment.MinSimilarity` with no number. The idea file and `name-candidates.md`, which are about the value, keep 0.4.

### M-5 · nit · The source comment should follow the registry's wording for a user decision
**Answer:** Fixed. D-1 uses `// decided: user, 2026-10-24 (S11-43): …` on the line above the field, as in `defaultKinship`. D-2 uses the same words.

### X-1 · major · T-1 clause (a) fails at 0.5, but STEP-1's check says it passes
**Answer:** Fixed. T-1 is split into (a1) the value and (a2) the gate comparison. STEP-1 names the three assertions expected to fail at 0.5 and makes only (a1) a stop condition.

### X-2 · major · STEP-4 and the DoD checks cannot run on this platform
**Answer:** Declined. The check-pr DoD line is the template's, word for word, and every landed Spike 11 plan carries it. `ship-pr` runs `scripts/check-pr.sh`, which needs macOS for the Xcode tests, and the manual check is the reviewer's step in the app. An implementer on another platform reports the unmet lines under Implementer rule 7.

### X-3 · major · STEP-5 and the DoD need a PR number that does not exist yet
**Answer:** Partly fixed. STEP-5 now says it runs once the PR number is known. The `completed.md` paragraph follows the S11-42 entry's voice, and the map row's Plan link and Planned status are wired now, with this review. The `Landed in #N` wording is the template's and is filled in at landing, as on every landed plan.

### X-4 · minor · T-2 is a single-case test; `go-tests.mdc` says table-driven
**Answer:** Fixed. D-4 says T-2 is a plain test because it has one case.

### X-5 · minor · STEP-1 command has no `-tags fts5` / `CGO_ENABLED`
**Answer:** Fixed. §7 opens with the one command form, and every check uses it.

### X-6 · minor · `givenVals` helper wording leaves small choices
**Answer:** Moot after M-1. D-6 states `Given`'s `Form` and parts exactly, with no surname part.

### X-7 · minor · STEP-3's rg check proves less than it appears to
**Answer:** Fixed. The check is labelled a leftover-number check, and D-5 fixes the wording.
