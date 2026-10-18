# Review — S11-11 Subject profile and entity compare

**Verdict:** Fits with changes — Ready with changes → fixed
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

Round 1 raised two blockers (M-1, M-2), both fixed. Round 2 was skipped: neither fix reshapes the plan (M-1 adds one small entry point beside `Build`, M-2 changes a wording to call an existing method), and the plan stayed Blocked on S10-66 until it landed.

**Refresh (2026-10-18, `main` @ `f6d3ac4`).** S10-66 landed as its plan said: `autoreconcile.Value.Depiction` is `*depictionvalue.Value`, `go list -deps -test ./core/autoreconcile` reaches nothing under `core/database`, and `layerPackages` gained the two `core/depictionvalue` rows. No reconciler logic changed, so every §2 value holds. The S10-66 plan went with the Spike 10 follow-ups folder, so its link and STEP-6's edit of its header were removed. Status → Planned.

## To user

- None.

## Writer's own calls

- **S10-66 is a dependency.** `core/autoreconcile` reaches `core/database` through `depictionvalues.Value` today. The builder must import the reconciler (to translate its output, and to re-run it for leave-one-out), so S11-11 waits for S10-66 instead of moving the type itself. Map row: Depends on S11-10, S10-66.
- **`core/subjectprofile` is the one layer package that imports `core/autoreconcile`.** The skill and `match-rules.mdc` gain that exception (D-2).
- **No concluded-value rule** (after M-3): the builder reads auto-reconciled values only; Reconciliation Claims are laid over them by the PR that ships claims.
- **Leave-one-out re-runs `Reconcile` only on properties a left-out member recorded**, with a nil concluded value (D-6).
- **Negative Observations that deny nothing are not in the profile**: the reconciler leaves no value for them, and the profile reads only its output.
- **The S11-10 review's M-8** (the test command written in two places) is resolved here: the rule points at the skill's Check (D-12).

## Round 1

### M-1 · blocker · A plain Subject has no path through the profile builder, but the design and S11-12 send both sides through it
`comparison-layers.md` lifespan: "for a person Subject or person entity"; map S11-13: "the loader that feeds `propertycompare` and `subjectprofile`". As planned, S11-13 would write a second Observation → layer 2 conversion and S11-12 would fork its lifespan hook.
**Answer:** Fixed. D-1, D-15: `ForSubject(map[vocabkey.PropertyKey][]autoreconcile.Candidate, namePattern) subjectcompare.Values`, sharing D-5's `value` with `Build`. US-5 (AC-5.1 – AC-5.5); US-3 builds side B with it. The Out row became S11-13's loader only.

### M-2 · blocker · D-4 restates the "displayed" rule beside `Reason.Displayed()`
**Answer:** Fixed. D-4 and AC-4.4 use `ReconciledValue.Displayed()`; the skill bullet names it. The `conclusiondetails` string copy (`core/database/conclusiondetails/conclusiondetails.go` `ReconciledValue.Displayed`) is logged here as an Out row, not this PR.

### M-3 · major · D-4's `Concluded` branch serves an input no caller produces, in a shape the domain doc contradicts
`conclusion-reconciliation.md`: claims are laid over the cache, never written into it; concluded values replace the list, one claim per value on a multiple property.
**Answer:** Fixed. The concluded branch, AC-1.7 and AC-2.5 (old) are gone; `Property` holds `Values []autoreconcile.ReconciledValue` instead of `autoreconcile.Result` (D-3); D-6 passes a nil concluded value. New AC-1.7 (kept beside denied) and AC-2.5 (leaving out the denying member) take the IDs. Out row: Reconciliation Claims → the claims PR, as a replacing list. Also answers X-5.

### M-4 · major · A functional-options pattern, new to the repo, for one option
**Answer:** Fixed. `Options{Excluding [][]byte}` as `Build`'s last argument (D-6), matching `core/database/matching` `Options` and `core/raster` `Options`. STEP-6 renames `Excluding(member)` to `Options.Excluding` in `comparison-layers.md`, the map's working rule and scope entry, the spike README and `docs/ideas/identity-claim-review.md`.

### M-5 · major · The `labels` test helper copies `subjectcompare`'s private `labels`
**Answer:** Fixed. D-10, STEP-1: `subjectcomparetest.Labels` with that body; `core/subjectcompare/invariants_test.go`'s `labels` is deleted and its two calls use it. The DoD now freezes only `core/subjectcompare`'s non-test files.

### M-6 · major · "Empty cardinality reads as single" becomes another hand-rolled copy
**Answer:** Fixed. D-14: `propertymeta.Meta.CardinalityOrSingle()`, used by `Build`, tested by `TestCardinalityOrSingle`. The other copies are an Out row: `core/match/pair.go` goes in S11-13, `promotealign/exhibits.go` with `pairExhibits` in S11-25, `properties.go` when next edited.

### M-7 · minor · Next adds omits that S11-12's lifespan needs a rules argument on `Build`
**Answer:** Fixed. §6 Next adds: S11-12 gives `Build` and `ForSubject` the subject section's lifespan rules as a new argument.

### M-8 · minor · `Build`'s `value` parameter shares its name with `Compare`'s, but takes different rules
**Answer:** Fixed. D-8: the parameter is `reconcile`; its doc says `matchregistry.Default().Value`, not `valuerules.Load`'s.

### M-9 · minor · The "only `subjectprofile` imports `autoreconcile`" rule stays prose, though the guard already walks imports
**Answer:** Declined for this PR, logged. The sibling rule ("no `core/matchregistry`") is also prose, and `core/graphalign` imports `matchregistry` today, so a guard now would need an exception that S11-16 removes. Out row: logged for S11-16's planner, where the rewrite of `core/graphalign` first meets the rule.

### X-1 · major · The skill's "Add a layer package" paragraph contradicts D-12 and is not in STEP-5
**Answer:** Fixed. STEP-6 (was STEP-5) rewrites it to "and its path to the Check command below."

### X-2 · major · The AC-4.4 check has no defined mapping between profile values and Result values
**Answer:** Fixed. AC-4.4 now counts: positives equal the `Displayed()` values, negatives the `denied` ones, per property. No second labeler.

### X-3 · minor · Test helper details D-10 leaves open
**Answer:** Fixed. D-10 says what each builder sets (`text` sets `HasText`; every other field zero), that AC-4.1 – AC-4.3 compare `subjectcomparetest.Labels`, gives `TestEntityCompare`'s `check` signature, lists the rows with a `check` (AC-1.8, AC-1.9 – AC-1.12, AC-2.7 – AC-2.9), and AC-2.4 says its `Candidates` are set to nil by hand.

### X-4 · minor · The Excluding error order is nondeterministic across properties
**Answer:** Fixed. D-7: `Build` walks keys in ascending order and returns the first error.

### X-5 · minor · D-4's concluded rule differs from the map and the design text
**Answer:** Fixed by M-3: no concluded rule; the map's scope entry now says "displayed values positive" (STEP-6).

### X-6 · minor · The STEP-5 greps are loosely worded
**Answer:** Fixed. STEP-6's Check gives exact counts: `rg -c 'core/autoreconcile'` prints 2 for the skill and 1 for the rule.

### X-7 · nit · Smaller points
Logged. Commits are now one per step; the skill paragraph rewraps to 80 columns; STEP-1 was split (shared fixtures first, then the package).
