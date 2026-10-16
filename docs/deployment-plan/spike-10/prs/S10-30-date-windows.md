# S10-30 — Date windows in matching

**Status:** Blocked — on S10-28
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-30--date-windows-in-matching)
**Brief:** — (no new chrome: the score changes, no view does)
**Board:** —
**Depends on:** S9-21 (landed; stores the date windows). Sequenced after S10-28 ([plan](S10-28-support-negatives.md), code not landed), whose `pairValues` this PR's period pooling sits in front of
**Blocks:** —
**Branch:** `s10-30-date-windows`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-date-value`](../../../../.agents/skills/add-date-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/date-values.mdc`
**Context:** [`name-matching-enhancements.md`](../name-matching-enhancements.md) ("Date precision below the year", "ABT on spans", "Cross-property dates"), [`docs/matching.md`](../../../matching.md), [`docs/structured-date-model.md`](../../../structured-date-model.md), [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §7.1, [S10-28 plan](S10-28-support-negatives.md) D-4, working rule "Dates in matching use the windows S9-21 already stores" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `ff14e3f` plus the S10-28 plan as merged. S10-28 code has not landed. Redo the rows marked **S10-28** in §4 on landing (see Stop conditions)
**Review:** [`S10-30-date-windows.review.md`](S10-30-date-windows.review.md) — round 1 of 2: Fits with changes / Not ready; every finding answered; round 2 follows

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---


## 1. Objective

Dates compare by the days they cover, not by their year numbers. December 1817 and January 1818 are a month apart, so they stop reading as "a year apart"; a range such as March to June 1817 keeps its months, so a September date no longer overlaps it; and an Event's `date` can be compared with another Event's `start_date` and `end_date` as a period.

**Why now.** S9-21 stored every date's day window and the reconciler already uses it. The matcher still counts whole years, so it ranks a boundary pair below a same-year pair that is months further apart. Slice 7 is the matching slice; this is the date half of it.

**What stays.** The tolerance still counts years, by the civil year numbers of the two dates. Every comparison between two whole-year dates, every pair of points in one calendar year, and every year-level range keeps its exact score (the 0.9 / 0.8 / 0.7 / 0.6 tiers, the span scores). Only a date finer than a year, compared across a year boundary or with a month-level range, moves, because the fall-off now counts days (two existing tests change, §8). `DateComparer`'s settings, the registry numbers, `Rank`, and Promote's scales are unchanged. The reconciler and its cache are untouched. Julian versus Gregorian and double dating stay out.

## 2. User stories

**US-1.** As a researcher, when two dates fall either side of a new year, I see them score by how many days apart they are, so that December 1817 and January 1818 do not read as a year apart.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `DEC 1817` against `JAN 1818` | 0.777 (`0.8 × (1 − (31/365)/3)`): above the old year-apart figure 0.533 and below 1 |
| AC-1.2 | `MAY 1817` against `SEP 1817` | 0.6, unchanged: points in one calendar year keep their tier |
| AC-1.3 | `1817` against `1818` (both year-only) | 0.533, unchanged |
| AC-1.4 | `DEC 1817` against `1818` (a month against a whole year) | 0.655: the centres are `1.5 − 349.5/365` = 0.542 years apart, so `0.8 × (1 − 0.542/3)` |
| AC-1.5 | `DEC 1817` against `JAN 1818`, `Tolerance` set to 0 | 0: the civil years differ by 1, past the tolerance |
| AC-1.6 | `14 MAY 1817` against `1818`, then against `1819`, then against `1820` | 0.498, 0.231, 0. The first two fall off by the day gap (1.134 and 2.134 years); the third is cut off because 1820 − 1817 = 3 years is past the tolerance of 2 |

**US-2.** As a researcher, when a date is a range of months or days, I see its months and days count, so that a September date is not "inside" a March-to-June range.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `BET MAR 1817 AND JUN 1817` against `14 MAY 1817` | 0.8: the point is inside the window; a window under a year scores as a one-year span |
| AC-2.2 | The same range against `14 SEP 1817` | 0.734 (`0.8 × (1 − (90.5/365)/3)`); it was 0.8 when both read as all of 1817 |
| AC-2.3 | `FROM 1815` (no end) against `1900` | 0.2, unchanged: an open span scores the floor |
| AC-2.4 | Every whole-year range case in `TestDateComparerScores` | Unchanged |

**US-3.** As a researcher, when a date is "about" a year and the other is a range or a bound, I see the normal tolerance apply.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ABT 1817` against `BET 1820 AND 1821` | 0: the civil years differ by 3 and the tolerance stays 2 |
| AC-3.2 | `ABT 1817` against `ABT 1820` | 0.32: two points, so the tolerance doubles to 4 |
| AC-3.3 | `ABT 1817` against `BEF 1820`, then against `BEF 1815` | 0.2 (overlap), then 0.067 (`0.2 × (1 − 2/3)`: the years differ by 2, inside the normal tolerance) |

**US-4.** As a researcher, when one Event has a `date` and another has only `start_date` and `end_date`, I see them compared, so that a 1817 baptism can match a residence from 1815 to 1820.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Event profile. Probe `date` 1817. Candidate `start_date` 1815 and `end_date` 1820, no `date` | The `date` feature scores 0.55 (the point inside a six-year span), contribution 3.3. The reason names `1817` and the period |
| AC-4.2 | The same two Events with the roles swapped | Contribution 3.3 |
| AC-4.3 | Probe `date` 1823 against the same period | The `date` feature contradicts: −4 (the civil years differ by 3, past the tolerance) |
| AC-4.4 | Both Events have a `date` | No period is built; `date` compares with `date` |
| AC-4.5 | Neither Event has a `date`, both have start and end | No period is built; only the `start_date` and `end_date` features run, as today |
| AC-4.6 | Candidate has `start_date` 1815 and no `end_date` | The period is an open `AFT 1815`: probe `date` 1817 scores 0.2 |
| AC-4.7 | `EvaluateWithNames` for AC-4.1 | The `date` comparison is comparable, not a conflict |
| AC-4.8 | Probe has `date`, `start_date`, and `end_date`; candidate has only `start_date` and `end_date` | No period is built: the start and end features compare directly and the `date` feature is skipped. Nothing counts three times |
| AC-4.9 | Candidate `start_date` is a negative Observation, or is `denied` (S10-28 fields), and `end_date` is 1820 | The period is `BEF 1820` built from the end alone; the negative start never becomes a period start |
| AC-4.10 | Candidate `start_date` 1822 and `end_date` 1815 (start after end) | No period is built; the `date` feature is skipped |

**US-5.** As a contributor, when I run the date tests, every existing case and invariant still holds.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | The existing cases of `TestDateComparerScores`, `TestDateComparerLadder`, and `TestDateComparerInvariants` | Pass with no edit to an expected value |
| AC-5.2 | Generated dates that now include months and days on range ends, and inverted ranges | Symmetric, within 0…1, and a date against itself is 1 |
| AC-5.3 | `TestRegistrySetsEverySetting` | Passes: this PR adds no comparer setting |
| AC-5.4 | `TestRankEventsAndPlaces` and `TestForSubjectEventsAndPlaces`, where `14 MAY 1817` meets year-only `1818` | `EVT-B` scores 7.0 (`4 + 6 × 0.498`), not 7.2. These two expectations are the only edits to a pre-existing test |

## 3. Scope

| In | Covered by |
| --- | --- |
| Day windows for a date's first and last end, with year coordinates, exported beside `DateBounds` | STEP-1, AC-1.1, AC-1.3, AC-2.1, AC-2.2, AC-5.2 |
| `DateComparer` measures gaps and widths from those windows; same-year points keep their tiers; the cutoff counts civil years | STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-5.1, AC-5.4 |
| ABT widens the tolerance only for two points (the behavior the code has; locked by tests) | STEP-2, AC-3.1, AC-3.2, AC-3.3 |
| A date can compare with the other side's start and end as a period, only when exactly one side has a `date` and no start or end | STEP-3, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, AC-4.6, AC-4.7, AC-4.8, AC-4.9, AC-4.10 |
| Docs: `matching.md`, the spike design note, the map, `completed.md` | STEP-4 |

| Out | Home |
| --- | --- |
| Julian versus Gregorian dates, and double dating (1717/18) | [`docs/ideas/international-names.md`](../../../ideas/international-names.md) lists calendars; not planned here |
| A new weight for the period comparison (it uses the `date` feature's weight) | not planned (D-7) |
| Time-of-day precision (a day is the finest window) | not planned |
| Qualifiers on a period's start or end (`BEF` on a `start_date`) | not planned (D-7) |
| Comparing `start_date` with `end_date`, or any other pair of different Properties | not planned |
| Exhibits and Promote's evidence sheet comparing across Properties: a period-derived `date` comparison shows a score with no exhibit line | not planned |
| Changing `windowOf`, `DateBounds`, or the reconciler's treatment of one-sided ranges | not planned (D-2) |
| A configurable days-per-year, a finer tier model for same-year points, or a continuous score across the year boundary | not planned (D-3) |

## 4. Current state

Verified at `ff14e3f`. Rows marked **S10-28** describe that plan's design, not code on `main`. Check them on landing.

| Area | Where | Today |
| --- | --- | --- |
| Date comparer | `core/match/dates.go` `DateComparer` (line 23), `spanOf` (line 64), `spanScore` (line 98), `yearsApart` (line 123), `Compare` (line 133) | Spans of whole years: `yearSpan{lo, hi, openLo, openHi, point, approx, month, day}` with `month` and `day` pointers. A range reads `StartYear`..`EndYear` only. Two points in one year take a tier by month and day; points in different years fall off from `YearsApartFrom` by the difference of year numbers; spans score by the wider span's width in years, less a gap term. A `Kind` other than range is a point |
| Date settings | `core/match/registry.go` `DefaultDates` (line 200), `DefaultProfile` event case (line 271) | `Tolerance` 2 years, `ApproxToleranceFactor` 2, tiers 0.9 / 0.8 / 0.7 / 0.6, `YearsApartFrom` 0.8, `SpanOneYear` 0.8, `SpanPerExtraYear` 0.05, `SpanFloor` 0.2. Event features `date` 6 / 4, `start_date` 3 / 2, `end_date` 3 / 2, each compared with the same Property only |
| Windows | `core/autoreconcile/dates.go` `windowOf` (line 287), `pointWindow` (line 296), `rangeWindow` (line 318), `spanOf` (line 332), `isLeap` (line 384), `daysFromCivil` (line 389), `DateBounds` (line 68), `PeriodWindow` (line 117) | Inclusive proleptic day numbers, Gregorian civil components. A range with one side missing is `!ok` there (no window). `DateBounds` returns the whole window's `lo` and `hi`, not each end. `PeriodWindow` returns a day window, not a `datevalues.Value` |
| Pair selection | `core/match/match.go` `Score` (line 115), `best` (line 136); `core/match/pair.go` `evaluate` (line 65), `compareProperty` (line 90) | Each Feature or meta compares the same Property on both sides. **S10-28** replaces `best` with `pairValues(cmp, probe, candidate []Value) PairResult` in a new file, called from `Score` and `compareProperty`, and adds `Value.Label`, `Reason.Probe`, `Reason.Candidate`, `Values.HasPositive`, and the leader definition (its D-3) |
| Callers of the pair seam | `core/database/matching/matching.go` (`Rank`), `core/database/promotealign/canon.go` and `core/graphalign/align.go` (`Rank`, `EvaluateWithNames`) | All reach `Score` or `evaluate`. `core/database/promotealign/exhibits.go` calls `CompareDiscount` itself and bypasses both |
| Promote's metas | `core/database/promotealign/layer.go` `loadMetas` (line 82) | Takes every profile Property, so `date`, `start_date`, and `end_date` are all metas of an Event comparison |
| Tests | `core/match/dates_test.go` `TestDateComparerScores`, `TestDateComparerLadder`, `TestDateComparerInvariants`; `core/match/match_test.go` `TestRankEventsAndPlaces` (line 176, expects `EVT-B=7.2`; `approx` at 1e-9); `core/database/matching/matching_test.go` `TestForSubjectEventsAndPlaces` (line 266, expects `7.2`) | Green at the base. The generated invariants use months and days on points but only years on ranges. The two named tests score `14 MAY 1817` against year-only `1818` |
| Benchmarks | none in `core/match`; `core/database/promotealign/bench_test.go` `BenchmarkProposeColdHub` | The closest existing benchmark |
| Reconciler date module | `core/autoreconcile/dates.go` `dateModule` | Unchanged by this PR |

**Precedent.** `core/autoreconcile/dates.go` already turns a date into a day window and the cache stores its `date_lo` and `date_hi` (S9-21). Copy that: derive each end's window with the same civil-day arithmetic, and do not add a second day-count routine. Do not copy `windowOf`'s refusal of a one-sided range: the matcher scores an open-ended range today.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The **fall-off** counts fractional years from year coordinates: a bound on civil day `d` of year `y` covers `[y + (doy−1)/n, y + doy/n)`, with `n` the days in `y`. A whole year is `[y, y+1)` and its centre is `y + 0.5`. The **cutoff** counts civil year numbers (D-3, D-4). Tolerance and every registry number stay in years. | Whole-year dates then give exactly the figures they give today, so no year-level score moves. Splitting the two jobs keeps a year-only date two calendar years from a month-level date inside the tolerance, which `TestDateComparerLadder` needs (`14 MAY 1817` against `1819` is 0.231, not 0). | Cutting off on the fractional gap — `14 MAY 1817` against `1819` is 2.13 years and would score 0, breaking the ladder. Dividing day counts by 365.25 — moves every year-level figure by about 0.0007 and breaks the 1e-9 test tolerance |
| D-2 | A new function `autoreconcile.DateEnds(d datevalues.Value) (first, last DateEnd, ok bool)` returns each end of a date. `DateEnd{Year, Month, Day int; Lo, Hi int; From, To float64; Open bool}` is the end as written (`Month` and `Day` are 0 when absent), its inclusive day span, and its year coordinates; an open end has `Open` true and zero fields. A `Kind` other than `range` is a point, with the same span at both ends; `BEF` has an open first end; `AFT` an open last end; a range has its start bound first and its end bound last, swapped when the start bound's `Lo` is greater than the end bound's `Lo`; a range with one side missing has an open end on that side. A point with no `StartYear` uses its `End*` fields, as `match.spanOf` does today. Day of year is `daysFromCivil(y, m, d) − daysFromCivil(y, 1, 1) + 1` over `n = daysFromCivil(y+1, 1, 1) − daysFromCivil(y, 1, 1)`. A month outside 1 to 12 or a day outside the month is not ok. Civil components are read as Gregorian, as `windowOf` does. The internal helper is named `boundEnd`, not `spanOf`. `windowOf`, `DateBounds`, and the reconciler are unchanged. | One day-count routine and one place that turns components into days. It tolerates one-sided ranges because `TestDateComparerScores` scores them. `datevalues` rejects a bad month or day on write, so the stricter check is unreachable from the catalog. | Calling `DateBounds` — gives the whole window, not each end, and has no year coordinates. Changing `windowOf` to accept one-sided ranges — changes what the reconciler treats as evidence |
| D-3 | Two **points** (exact or qualified, not `BEF` or `AFT`) in the same civil year keep their tier by month and day. Two points in different years score `YearsApartFrom × (1 − gap / (tol + 1))` when their civil years differ by at most `tol`, else 0, with `gap` the distance between the centres of their bounds in years and `tol` doubled when either is qualified. User answer 2026-10-16. | Every same-year and every year-level score holds. December and January score by days (0.777), May and September keep 0.6, which is the "outranks a same-year pair only when the days say so" the scope names. The score is discontinuous at the year boundary (31 December against 1 January is about 0.799, two days in one month 0.7): by decision. | Scoring every non-identical pair from the day distance — May and September would score 0.711 and the tiers and their tests would change |
| D-4 | For a pair where either side is a range or a bound, both **gaps** are 0 when the two day windows share a day (an open end reaches past every date). Otherwise the earlier date is the one whose last end is closed and ends before the other's first end begins; the **year gap** is the later first end's `Year` minus the earlier last end's `Year`, and the **gap** is the centre of the later first end minus the centre of the earlier last end. The score is 0 when the year gap exceeds the tolerance, else the span score times `1 − gap / (tol + 1)`. Two spans equal by value (`dateSpan` has no pointer fields) score 1. | A year-level range gives the difference of the nearest year numbers, as `yearsApart` does today (a point in 1822 against 1815–1820 is 2). Months and days now move the fall-off. | Edge-to-edge distance — a point in 1822 against 1815–1820 would be a gap of 1 and move today's score |
| D-5 | A span's **width** is `last.To − first.From` in years, and the span score uses `max(width, 1)` of the wider date. | A range of whole years gives the integer width it gives today. A window under a year scores as a one-year span (0.8), so the PR adds no number above 0.8. | Letting a short window score above 0.8 — a new number nobody tuned |
| D-6 | ABT widens the tolerance only for two points. The code already does this (the factor applies only in the point branch). This PR changes no line for it, adds tests that lock it, and rewords the `DefaultDates` comment. | The scope line is satisfied by today's behavior; a test stops a refactor from losing it. | Rewriting the branch — nothing to fix |
| D-7 | **Period pooling.** The registry gets `DatePeriods map[Property]PeriodProperties` *(new)*, with `PeriodProperties{Start, End Property}`: `date` → (`start_date`, `end_date`), product origin. In `Score` and `evaluate`, for a Property that is a key of `DatePeriods`, `poolPeriods` replaces one side's values for the Property with a single period exactly when: the other side has a positive value of the Property and no positive value of its start or end; and this side has no positive value of the Property and at least one positive, non-`Denied`, non-`Neutral` value of its start or end (**S10-28**: positive means not `Negative`; leader means its D-3). The period is built from the leader start value's first end and the leader end value's last end, through `autoreconcile.DateEnds`: both present → a range; start only → an `AFT` point; end only → a `BEF` point; the qualifiers on the sources are dropped. A start later than the end builds no period. The period is a plain `Value` (not `Counted`), so S10-28 weighs it 1. The `date` feature's weight and contradiction are used. User answer 2026-10-16, tightened so that a side with all three Properties never pools (nothing counts twice). | The scope says `date` may overlap `start_date` and `end_date` by window. One valid date value per side keeps the label (`CompactDisplay`), `Compatible`, and the comparer working on a value the catalog could store. Reusing the `date` weight adds no number. | Always comparing `date` against the period — counts start and end twice. One period per start × end pair — pairs values from unrelated Observations, so a start after an end is routine. A one-sided range `Value` — `datevalues` rejects it and its label is empty |
| D-8 | `poolPeriods(prop Property, probe, candidate Values) (a, b []Value)` lives in `core/match/periods.go` *(new)* and is called once in `Score` and once in `evaluate`, before the pair function (**S10-28**: `pairValues`). When the Property is not a `DatePeriods` key it returns the inputs unchanged and allocates nothing. | The pair seam takes value slices, so the Property rule has to run where both `Values` are in hand: the two entries, each calling one function. | Pooling in the loaders — Promote and merge hints load differently, and the profile would not know |
| D-9 | No setting is added to `DateComparer`. | The numbers are the registry's; nothing to tune that was not already. `TestRegistrySetsEverySetting` stays as it is. | A days-per-year setting — D-1 removes the need |
| D-10 | `dateSpan` replaces `yearSpan` with `first` and `last DateEnd`, `point`, and `approx`, and no pointer fields. `Compare` reads the civil year, month, and day of a point from its `DateEnd`. | One source for the tiers and the gaps, and `==` means equal by value. | Reading `datevalues.Value` fields again beside the ends — two sources of truth |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `autoreconcile.DateEnds`, `autoreconcile.DateEnd` *(new)* | func, type | `autoreconcile.DateBounds`, `Window`, `PeriodWindow` | `DateBounds` and `PeriodWindow` return one day window with no per-end span, year, or year coordinates; the new function adds those beside them and `PeriodWindow` is not reused because it returns day numbers, not a `Value` |
| `core/match/periods.go` *(new)*, `poolPeriods`, `DatePeriods`, `PeriodProperties` *(new)* | file, func, var, type | `core/match/registry.go` `DefaultProfile` | Pooling needs both sides' whole `Values`; the Property mapping is registry data, as profiles are |
| `dateSpan` in `core/match/dates.go` *(replaces `yearSpan`)* | type | `yearSpan` | It carries the two ends instead of two years |
| `BenchmarkDateComparer`, `BenchmarkPoolPeriods` *(new)* | benchmarks | `core/database/promotealign/bench_test.go` `BenchmarkProposeColdHub` | A date compare runs per pair per Feature; the new arithmetic must not allocate, and a non-pooling Property must cost a map lookup |

**Builds on.** S9-21's day windows and civil-day arithmetic in `core/autoreconcile`; the pair function S10-28 creates and its leader rule; the profile and meta lists that already name `date`, `start_date`, and `end_date`.
**Deletes.** `yearSpan`, `spanOf`, and `yearsApart` in `core/match/dates.go`, replaced by `dateSpan` and the gap helpers.
**Next adds.** Calendars: `datevalues.Value.Calendar` exists and `windowOf` and `DateEnds` ignore it, so a Julian date is converted to civil Gregorian components before `DateEnds`, and nothing after that changes. Another Property pair (a place relationship's `start_date` and `end_date`, which have no `date`) is one `DatePeriods` entry.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` `dates.go` | `DateEnd`, `DateEnds`, `boundEnd` (D-1, D-2) |
| `core/match` `dates.go` | `dateSpan` replaces `yearSpan`; `Compare` follows D-3, D-4, D-5, D-6, D-10; the gap helpers replace `yearsApart` |
| `core/match` `periods.go` *(new)* | `poolPeriods` and the period builder (D-7, D-8) |
| `core/match` `registry.go` | `DatePeriods`, `PeriodProperties`; `DefaultDates` comments say gaps and widths are in years and fall-off counts days, and state the ABT rule (D-6) |
| `core/match` `match.go`, `pair.go` | `Score` and `evaluate` call `poolPeriods` before the pair function (**S10-28**: `pairValues`) |

Errors: none new. A date `DateEnds` cannot place (no year, a bad month or day) is not comparable, as today.

### FFI

None — no message changes.

### macOS

None — no view changes. An Event suggestion can score differently when its dates sit across a year boundary, in a month-level range, or against a start and end; no copy, state, or layout changes.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Date ends.**
Files: `core/autoreconcile/dates.go`, `core/autoreconcile/dates_test.go`. Skill: `add-date-value`.
Do: `DateEnd`, `DateEnds`, and `boundEnd` per D-1 and D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile -run 'TestDateEnds'` passes (T-1), and the rest of `./core/autoreconcile` still passes.

**STEP-2 — `DateComparer` on date ends.**
Files: `core/match/dates.go`, `core/match/registry.go`, `core/match/dates_test.go`, `core/match/match_test.go`, `core/database/matching/matching_test.go`.
Do: replace `yearSpan` with `dateSpan`; D-3 through D-6 and D-10; update the `DateComparer` and `DefaultDates` comments; change the two expectations named in AC-5.4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match ./core/database/matching` passes, with every `dates_test.go` case unchanged and `EVT-B` at 7.0 (T-2, T-3, T-4, T-9). `CGO_ENABLED=1 go test -tags fts5 -bench BenchmarkDateComparer -benchmem -run '^$' ./core/match` reports 0 allocs/op (T-8).

**STEP-3 — Period pooling.**
Files: `core/match/periods.go` *(new)*, `core/match/registry.go`, `core/match/match.go`, `core/match/pair.go`, `core/match/periods_test.go` *(new)*, `core/match/match_test.go`, `core/match/pair_test.go`.
Do: `DatePeriods`, `poolPeriods`, and the two call sites (D-7, D-8).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match` passes (T-5, T-6, T-7). `CGO_ENABLED=1 go test -tags fts5 -bench BenchmarkPoolPeriods -benchmem -run '^$' ./core/match` reports 0 allocs/op for a Property that is not a `DatePeriods` key and for a pair of Events that does not pool (T-10).

**STEP-4 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/name-matching-enhancements.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-30-date-windows.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Score dates by the days they cover."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/dates_test.go` `TestDateEnds` *(new)* | Table-driven. `1817` → both ends `From` 1817.0, `To` 1818.0, `Year` 1817, `Lo`/`Hi` the year's first and last day. `MAY 1817` and leap-year `FEB 1816`: `From` and `To` from the day of year over 365 or 366. `14 MAY 1817`: one day, `Month` 5, `Day` 14. `BET MAR 1817 AND JUN 1817`: first end March, last end June. `BEF 1820`: first end open. `AFT 1820`: last end open. `FROM 1815` (no end): last end open. An inverted range (start `Lo` after end `Lo`) swaps. `ABT 1817` is a point. A `Kind` left empty is a point. Month 13 and a phrase with no year → not ok | AC-1.1, AC-2.1, AC-2.3, AC-5.2, D-1, D-2 |
| T-2 | Go | `core/match/dates_test.go` `TestDateComparerScores` | Every existing case unchanged. New rows: `DEC 1817` vs `JAN 1818` → `0.8 × (1 − (31/365)/3)`; same with `Tolerance` 0 → 0; `MAY 1817` vs `SEP 1817` → 0.6; `DEC 1817` vs `1818` → `0.8 × (1 − gap/3)` with `gap = 1.5 − 349.5/365`; `14 MAY 1817` vs `1818`, `1819`, `1820` → 0.498, 0.231, 0; range `BET MAR 1817 AND JUN 1817` vs `14 MAY 1817` → 0.8, vs `14 SEP 1817` → `0.8 × (1 − (90.5/365)/3)` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-5.1, D-3, D-4, D-5 |
| T-3 | Go | same file `TestDateComparerABT` *(new)* | `ABT 1817` vs `BET 1820 AND 1821` → 0; `ABT 1817` vs `ABT 1820` → 0.32; `ABT 1817` vs `BEF 1820` → 0.2; `ABT 1817` vs `BEF 1815` → `0.2 × (1 − 2/3)` | AC-3.1, AC-3.2, AC-3.3, D-6 |
| T-4 | Go | same file `TestDateComparerLadder`, `TestDateComparerInvariants` | Unchanged; the generator also draws months and days on range ends (a day only when it draws a month) and includes inverted ranges | AC-5.1, AC-5.2 |
| T-5 | Go | `core/match/periods_test.go` `TestPoolPeriods` *(new)* | Table-driven. Probe has `date`, candidate has `start_date` and `end_date` → candidate side gets one period. Both have `date` → no change. Neither has `date` → no change. Dated side also has a start → no change (AC-4.8). Only a start → an `AFT` point; only an end → a `BEF` point. Start after end → no change. Several start values → the first leader's. A property that is not a key of `DatePeriods` → no change. After S10-28: a negative or `Denied` start is skipped (period from the end alone); a probe whose only `date` is negative counts as having no `date` | AC-4.4, AC-4.5, AC-4.6, AC-4.8, AC-4.9, AC-4.10, D-7, D-8 |
| T-6 | Go | `core/match/match_test.go` `TestScoreDatePeriod` *(new)* | `DefaultProfile("event")`. Probe `date` 1817, candidate `start_date` 1815 and `end_date` 1820 → the `date` reason is similarity 0.55, contribution 3.3, and the start and end features are skipped. Swapped → 3.3. Probe 1823 → contribution −4. Both Events with start and end only → two reasons (start, end), no `date` reason. After S10-28 the reason labels name `1817` and the period | AC-4.1, AC-4.2, AC-4.3, AC-4.5 |
| T-7 | Go | `core/match/pair_test.go` `TestEvaluateDatePeriod` *(new)* | `Evaluate` over metas `date`, `start_date`, `end_date` for the AC-4.1 values → the `date` comparison is comparable with similarity 0.55 | AC-4.7 |
| T-8 | Go | `core/match/dates_test.go` `BenchmarkDateComparer` *(new)* | Calls `b.ReportAllocs()`. Two points, a point against a range, a range against a range: 0 allocs/op | D-1 |
| T-9 | Go | `core/match/match_test.go` `TestRankEventsAndPlaces`; `core/database/matching/matching_test.go` `TestForSubjectEventsAndPlaces`; `core/match/registry_test.go` `TestRegistrySetsEverySetting` | `EVT-B` expects `7.0` in both (the comment in the first reads `4 + 6 × 0.498`); the registry test passes with no new setting | AC-5.3, AC-5.4, D-3, D-9 |
| T-10 | Go | `core/match/periods_test.go` `BenchmarkPoolPeriods` *(new)* | Calls `b.ReportAllocs()`. A Property that is not a key, and an Event pair that does not pool: 0 allocs/op. A pooling pair is reported, with no bar | D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break are `core/match`, `core/database/matching`, and `core/database/promotealign` (date helper at `promotealign_test.go` line 345). A pre-existing test other than the two in AC-5.4 whose expected number changes is a stop-and-report.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host with no xcodebuild, the Go and plan-doc steps pass and the Xcode step is reported as not run
- [ ] The only edits to pre-existing test expectations are the two in AC-5.4; if another test needed a change, the work stopped and reported
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] BenchmarkDateComparer reports 0 allocs/op; BenchmarkPoolPeriods reports 0 allocs/op for the non-pooling cases
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-30 Plan line status; the S10-30 "Depends on" reads "S9-21; sequenced after S10-28"; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: the file table lists `periods.go` and describes `dates.go` as date ends and gaps; the `DateComparer` row says the cutoff counts civil years and the fall-off counts days, with the December and January example, the same-year tiers, the discontinuity at the year boundary, and the period rule; the "Not yet handled" bullet no longer names date windows (leave the other items in that bullet as they are, and rebase over S10-28's edit to the same bullet and table)
- [ ] name-matching-enhancements.md: "Date precision below the year", "ABT on spans", and "Cross-property dates" say they landed in S10-30
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice has no new chrome, so the check is a catalog run, not a screen. The tests assert every number; `-v` lists which cases ran.

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestDateComparerScores|TestScoreDatePeriod' -v ./core/match`. The December and January row, the month-level range rows, and the period rows are listed as passing.
2. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestRankEventsAndPlaces|TestForSubjectEventsAndPlaces' -v ./core/match ./core/database/matching`. `EVT-B` reads 7.0 in both.
3. Optional, in the app: in a project with a residence from 1815 to 1820, add a baptism dated 1817 and open Promote. The residence appears among the baptism's suggestions; the sheet shows the date row as a partial with no exhibit line (a period comparison has no exhibits).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-28 lands differently from its plan, so `Score`, `evaluate`, the leader rule, or the pair function differ from §4 | §4 marks every S10-28 row. The first action is to redo them against the landed code |
| A year-level score moves in the last decimal | D-1's coordinates give exact figures for whole years; `TestDateComparerScores` compares at 1e-9 and is the check |
| A date compare runs per pair per Feature over every candidate | `DateEnds` allocates nothing and T-8 and T-10 assert 0 allocs/op |
| A suggestion that relied on two dates "overlapping" in one year drops | Intended for ranges with months (AC-2.2). The tier pairs do not move |
| An Event with a period now gets date evidence it never had | Intended (US-4). The same weight as `date`, so a period cannot outweigh a real `date` match |
| Promote's sheet shows a period-derived date partial with no exhibit line | Accepted and listed in Out; exhibits pair Observations of one Property |
| The score is discontinuous at the year boundary | By decision (D-3); `docs/matching.md` says so |

**Stop and report if:**

- S10-28 has not landed, or `Score` and `compareProperty` do not both call one pair function. Redo §4 and tell the user what moved.
- Any existing case in `core/match/dates_test.go` needs a changed expected value to pass.
- A pre-existing test other than `TestRankEventsAndPlaces` and `TestForSubjectEventsAndPlaces` needs a changed expected value.
- `autoreconcile` has a caller that depends on `windowOf` accepting a one-sided range (it does not today; `DateEnds` must not change `windowOf`).
- `BenchmarkDateComparer` reports any allocation.

## 12. Open questions

None.

## 13. PR

**Title:** S10-30 — Date windows in matching
**Body:** Dates now compare by the days they cover, ranges keep their months and days, and an Event's `date` compares with another's start and end; what shipped against US-1 to US-5; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-30-date-windows.md`.
