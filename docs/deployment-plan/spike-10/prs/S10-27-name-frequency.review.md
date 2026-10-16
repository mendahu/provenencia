# S10-27 review log

**Plan:** [`S10-27-name-frequency.md`](S10-27-name-frequency.md)
**Lint:** `python3 scripts/check-plan-docs.py` — 0 errors, 0 warnings, run after the round-1 fixes
**Round 1:** maintainability reviewer: Fits with changes. Executability reviewer: Ready with changes. Round 2 not run: no blocker remained open after the fixes.
**Verdict:** Planned. No open blocker. Nothing for the user.

## Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | A new graph slot, package, and SQL scan duplicate `Stats` | **Fixed.** Counts are `Stats.Words` (D-3, STEP-5). `rank` counts the candidates it already loaded. The `namefrequency` package, the graph slot, T-8, and T-10 are gone. Reading `g.Kind` nodes for the count was declined: it loads every Person into memory, which `computeStats`'s SQL avoids |
| M-2 | major | `ScoreSimilarity` gains a positional float with a zero-value sentinel | **Fixed in part.** The field is now `Discount`, where 0 means none (D-6, D-9). The sentinel rule is gone. A shared pair-result struct was **declined**: S10-28's fields are not known, and it can introduce the struct when it needs a second one |
| M-3 | major | `StrengthComparer` has one conformer; tie-break stacks keys | **Fixed in part.** `best` picks by `similarity × (1 − discount)` (D-9), so there is no tie-break. The interface stays as `DiscountComparer` (D-8) because it is the existing `Comparer` pattern and avoids a type switch on `NameComparer` in `best`. The "Next adds" claim about S10-28 is reworded |
| M-4 | major | Count and lookup derive keys from different rules | **Fixed.** `NameComparer.CountWords` shares `Words.resolve` with the lookup (D-4) |
| M-5 | minor | A third comparer-building step is repeated at two call sites | **Declined.** With M-1 each caller adds one line that differs (`stats.Words` versus the loaded candidates). The Load → aliases pair is S10-25's and is out of scope |
| M-6 | minor | Population query is not type-scoped | **Fixed.** The query joins `subject_types` to `person` (D-2) |
| M-7 | minor | Initial credit and `MinScore` comments | **Fixed.** D-5 states the initial rule. D-11 and STEP-2 update the `DefaultProfile` comment |
| M-8 | nit | A fourth `Querier` interface | **Fixed** by M-1 |

## Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | major | T-6 cannot reach a discount through `Evaluate` | **Fixed.** T-5 uses `EvaluateWithNames` and `Profile.WithNameComparer` |
| X-2 | major | The `MinScore` cut sees scaled scores; `canon.go` and `align.go` are unlisted readers | **Fixed.** D-11, AC-1.10, T-5, and §4 "Ranking readers" |
| X-3 | major | The benchmark cannot show the count's cost | **Fixed.** `BenchmarkLoadStatsNames` *(new)* with name rows; an absolute 250 ms/op bar replaces the 15% rise |
| X-4 | minor | Types and signatures left open | **Fixed.** D-7 (`*WordCounts`, `noWordCounts`), D-8 (return shape, value receiver), D-9 (contradiction discount 0). The `Querier` and shared-map questions are gone with M-1 or answered in `cloneStats` (D-10) |
| X-5 | minor | Fixtures leave choices open | **Fixed.** The 60-person fixture is specified in §8; T-7 names the SQL insert; T-8 and T-10 are removed |
| X-6 | minor | `rank` loads counts for every profile | **Fixed.** D-3 and T-11: only when the profile has a `name` Feature, from candidates already loaded |
| X-7 | minor | Fold keys can disagree | **Fixed.** D-4 and T-1 |
| X-8 | minor | Exhibit pairing ignores the discount | **Fixed.** D-9 and the `exhibits.go` row |
| X-9 | nit | Map line, `matching.md` wording, weak skill reference | **Fixed.** Skill and rule dropped from the header; DoD names the bullet to edit and what to leave |
