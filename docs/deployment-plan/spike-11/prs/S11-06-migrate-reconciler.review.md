# Review — S11-06 Migrate the reconciler onto value compare

**Verdict:** Fits with changes — Ready with changes → all fixed
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

Plan: [`S11-06-migrate-reconciler.md`](S11-06-migrate-reconciler.md). Lint: `check-plan-docs.py` clean but for this log's link before round 1.

Before the review, a prototype of D-1 – D-9 ran the whole Go suite at `75ed63c` (S11-05 not landed; Promote's two call sites passed `matchregistry.Default().Value`). Its failures are §8 "Expected changes". M-1's fix was prototyped after the review: `core/autoreconcile` stays green and AC-2.6 holds.

Round 2 was not run: no finding asked for a reshape, and the one blocker that changes code (M-1) was fixed and verified on the prototype.

## To user

- None. The three product calls (grouping on `Same` / `Contains`, no text fold, ABT as a range) were answered on 2026-10-17 before the plan was written.

## Writer's own calls

- D-1: the caller passes the value rules; the cache passes `matchregistry.Default().Value`, never catalog data (D-8), and `CacheVersion` becomes 16.
- D-3: terms, subjects and depictions keep catalog-id identity (`property_terms` is unique by key *and* origin).
- D-4: name units compare as untyped words, so no part type weighs more than another; only `Same` and initial pairs fold.
- D-6: `valuecompare.DateWindow` (as written) for sort keys, `date_lo` / `date_hi` and place periods.
- D-7: `PeriodWindow` / `LinkMembership` move to `conclusionheaders`, not into value compare; `comparison-layers.md` L454 is corrected.
- Status stays `Blocked — on S11-05` until S11-05 lands and the "after S11-05" rows are rechecked.

## Round 1

### M-1 · blocker · The name module ends up with two word models
Grouping would read value compare's tokens (particles rejoined) and outvoting `strings.Fields(NormalizeForm)`; `Mc Donnald` would escape a `McDonald` majority.
**Answer:** Fixed. D-4: `readName` is a `nameModule` method using `valuecompare.SplitWords` with the rules' particles; AC-2.6 and a T-4 row cover it. Verified on the prototype.

### M-2 · major · Four ways to turn a reconciler value into a `valuecompare.Value`
**Answer:** Fixed. D-3: one `Value.compareValue(valueType)` adapter in `modules.go`, shared by text, integer and date; `dateValue` dropped; `nameUnit.compareValue` built once in `split` and stored (D-4).

### M-3 · major · `laterStart` / `earlierEnd` read raw `DateEnds` beside `DateWindow`
**Answer:** Fixed. D-7 reads `valuecompare.DateWindow` and compares `Lo` / `Hi`, nil meaning open.

### M-4 · minor · `ideas/international-names.md` L37 not updated
**Answer:** Fixed. STEP-4 replaces the bullet with the decision.

### M-5 · minor · `words.go` header and `IsInitial` export go stale
**Answer:** Fixed (header in STEP-4; Stop condition widened to allow it). Unexporting `IsInitial` is an Out line, not planned.

### M-6 · minor · A new file for one test variable
**Answer:** Fixed. `testRules` lives in `autoreconcile_test.go`; no test mutates it. A package variable is kept over 43 local copies.

### M-7 · nit · Rules argument order
**Answer:** Fixed. `moduleFor(rules, valueType)`.

### X-1 · blocker · The review log does not exist
**Answer:** Fixed. This file.

### X-2 · major · `ptrDate` claim false; moved test's helpers not placed
**Answer:** Fixed. D-10: `ymd`, `qualified`, `between`, `ip` copied into `placechains_internal_test.go`; `ptrDate` moves with the test and is deleted from `dates_test.go`.

### X-3 · major · STEP-4's grep check contradicts its own edit
**Answer:** Fixed. The check asserts the old wording is gone and the policy line appears once.

### X-4 · major · STEP-2 too large
**Answer:** Fixed. Place periods and membership by day are STEP-2 (own commit, checked by `conclusionheaders` tests); the reconciler is STEP-3. D-13 updated.

### X-5 · minor · `dateValue` unspecified
**Answer:** Fixed by M-2 (one adapter, specified in D-3).

### X-6 · minor · `laterStart` / `earlierEnd` edge cases open
**Answer:** Fixed. D-7 states nil, unreadable, tie and both-unreadable outcomes.

### X-7 · minor · Three name subtests describe the old rule
**Answer:** Fixed. D-10 renames two titles (cases unchanged) and keeps the third, whose title stays true; the `names.go` header bullets are rewritten.

### X-8 · minor · No group "words" in `TestReconcileNames`
**Answer:** Fixed. T-4 names `TestReconcileNamesWords`.

### X-9 · minor · Test file count
**Answer:** Fixed by M-6 (row removed).

### X-10 · minor · Checks that cannot run before S11-05
**Answer:** Declined as a change: recorded by the status `Blocked — on S11-05` and the Stop condition on S11-05's shapes.
