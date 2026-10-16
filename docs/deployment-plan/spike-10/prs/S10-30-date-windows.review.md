# S10-30 review log

**Plan:** [`S10-30-date-windows.md`](S10-30-date-windows.md)
**Lint:** `python3 scripts/check-plan-docs.py` — run after each round's fixes
**Round 1:** maintainability reviewer: Fits with changes (two blockers). Executability reviewer: Not ready (two blockers). Both reviewed against `ff14e3f` with S10-28's plan as the assumed base for the rows marked **S10-28**.
**Round 2:** maintainability: Fits with changes, no blocker, two majors. Executability: Ready with changes, no blocker, three majors. Both confirmed every round-1 answer and re-derived the arithmetic; all 29 rows of `TestDateComparerScores`, the ladder, and the invariants hold under the new math, and no pre-existing test other than the two named in AC-5.4 changes.
**Verdict:** Fixed in the plan. No blocker is open. The plan stays `Blocked — on S10-28`: S10-28's code has not landed, and the rows marked **S10-28** need a redo on landing.

## Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | The tolerance cutoff on the fractional gap breaks `TestDateComparerLadder` (`14 MAY 1817` against `1819` is 2.13 years and would score 0) | **Fixed.** Cutoff counts civil year numbers; the fall-off counts days (D-1, D-3, D-4). AC-1.6 pins 0.498 / 0.231 / 0. The executability reviewer's ladder walk (0.232) agreed once the cutoff is by year numbers |
| M-2 | blocker | Period pooling reads negative, denied, and unweighted values; start × end pairs unrelated Observations | **Fixed.** D-7: positive non-`Denied` non-`Neutral` leader values only, one period per side, no period when start is after end, "no value" decided by positive values. AC-4.9, AC-4.10, T-5 rows (redo on S10-28 landing) |
| M-3 | major | "Counts nothing twice" is false when the dated side also has start and end | **Fixed.** D-7 pools only when the dated side has neither a start nor an end; AC-4.8 |
| M-4 | major | The period `Value` is not a value `datevalues` allows (one-sided range), has no label, and `PeriodWindow` is not cited | **Fixed.** D-7 builds a range, or an `AFT` / `BEF` point for a one-sided period; built through `DateEnds`. Codebase fit cites `PeriodWindow` and says why it is not reused |
| M-5 | major | Pooling as a pre-step at two call sites instead of the one seam | **Declined in part, fixed in part.** `pairValues` takes value slices and has no Property, so the rule runs where both `Values` are in hand; the two entries call one function (D-8). The hot-path cost is bounded by T-10 |
| M-6 | major | `DateEnd` lacks year, month, and day; `dateSpan` has two sources of truth | **Fixed.** D-2 adds `Year`, `Month`, `Day`; D-10 makes `dateSpan` pointer-free so `==` is by value |
| M-7 | minor | "Next adds" names a pair that does not exist; calendars not one line | **Fixed.** Next adds reworded: a place relationship's start and end; calendars convert before `DateEnds`, which reads Gregorian components |
| M-8 | minor | Contradictory Deletes text; `DatePeriods` / `datePeriods`; missing docs rows; map Depends-on | **Fixed.** `poolPeriods`; Deletes lists `yearsApart` as replaced; DoD adds the `matching.md` file-table rows and the Depends-on edit |
| M-9 | minor | Promote's sheet cannot explain a period score | **Fixed.** Out line, Risks row, Manual check 3 |
| M-10 | nit | The tier model is discontinuous | **Fixed.** D-3 states it; the DoD asks `matching.md` to say so |

## Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | Two existing tests change value (`EVT-B=7.2` → 7.0) and the plan did not predict it | **Fixed.** AC-5.4, T-9, and STEP-2 name `TestRankEventsAndPlaces` and `TestForSubjectEventsAndPlaces` as the only edits to a pre-existing test; "What stays" is narrowed |
| X-2 | blocker | Pooling reads `Negative`, `Denied`, and `Counted` values | **Fixed.** Same as M-2 |
| X-3 | major | `DateEnd` does not carry what `Compare` needs | **Fixed.** Same as M-6 |
| X-4 | major | Inverted-range rule undefined; the generator will produce inverted ranges | **Fixed.** D-2: swap when the start bound's `Lo` is greater than the end bound's `Lo`; the unanswerable stop condition is removed; T-4 says the generator draws a day only with a month and includes inverted ranges |
| X-5 | major | Identical non-point dates not stated | **Fixed.** D-4: two spans equal by value score 1; D-10 makes `dateSpan` comparable |
| X-6 | major | "Counts nothing twice" does not hold with all three Properties | **Fixed.** Same as M-3 |
| X-7 | major | Label of the period unspecified; open periods have no label | **Fixed.** Same as M-4; T-6 asserts the labels after S10-28 lands |
| X-8 | minor | `Kind` unset, open-end fields, later/earlier, day of year, name collision, bad months, Cartesian product | **Fixed.** D-2 (all), D-4 (earlier/later), D-7 (one period) |
| X-9 | minor | Benchmark claims, `-benchmem`, manual checks, `DatePeriods` type, DoD rows | **Fixed.** Codebase fit names the closest benchmark; both benchmarks call `ReportAllocs`; Manual check rewritten; `PeriodProperties` given; DoD rows added |
| X-10 | minor | Callers; sheet shows a period score with no exhibit | **Fixed.** §4 lists the callers; Out and Risks name the sheet |

## Round 2

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| R2-M1 / R2-X1 | major | D-2's inverted-range rule (start `Lo` after end `Lo`) swaps valid ranges such as `FROM MAR 1817 TO 1817`; `rangeWindow` swaps only when start `Lo` is after end `Hi` | **Fixed.** D-2 uses the `rangeWindow` rule; T-1 rows for `FROM MAR 1817 TO 1817` and `BET 14 MAY 1817 AND 1817` |
| R2-M2 / R2-X4 | major | D-7 contradicts S10-28 D-3 and D-4: its "positive, non-denied, non-neutral" gate omits provisional and zero-support values; the tie order is unstated | **Fixed.** D-7 uses S10-28's voting definition for the gate and the build, and takes ties in slice order |
| R2-X3 | major | Replacing the dated side's values drops a negative `date` that S10-28 R1 reads | **Fixed.** D-7 replaces only the positive values; T-5 row |
| R2-X2 | major | How an open source end (`BEF 1815` as a start), "start later than end", and the built `Value`'s calendar are decided | **Fixed.** D-7: an open source end counts as absent; later means start first end `Lo` after end last end `Hi`; `Calendar` is `gregorian`. AC-4.11, T-5 rows |
| R2-M4 / R2-X6 | minor | A range with neither year is not ok; the generator already draws inverted ranges | **Fixed.** D-2 and T-1; T-4 says what the generator adds |
| R2-X5 | minor | `boundEnd` has no signature | **Fixed.** D-2 gives it and says it calls the existing `spanOf` |
