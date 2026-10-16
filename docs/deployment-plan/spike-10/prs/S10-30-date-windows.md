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
**Review:** — (not reviewed yet)

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

**What stays.** Every comparison between whole-year dates, every same-year pair of points, and every year-level range keeps its exact score (the 0.9 / 0.8 / 0.7 / 0.6 tiers, the span scores, the tolerance in years). `DateComparer`'s settings, the registry numbers, `Rank`, and Promote's scales are unchanged. The reconciler and its cache are untouched. Julian versus Gregorian and double dating stay out.

## 2. User stories

**US-1.** As a researcher, when two dates fall either side of a new year, I see them score by how many days apart they are, so that December 1817 and January 1818 do not read as a year apart.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `DEC 1817` against `JAN 1818` | 0.777 (`0.8 × (1 − (31/365)/3)`): above the old year-apart figure 0.533 and below 1 |
| AC-1.2 | `MAY 1817` against `SEP 1817` | 0.6, unchanged: points in one calendar year keep their tier |
| AC-1.3 | `1817` against `1818` (both year-only) | 0.533, unchanged |
| AC-1.4 | `DEC 1817` against `1818` (a month against a whole year) | 0.655: the centres are `1.5 − 349.5/365` = 0.542 years apart, so `0.8 × (1 − 0.542/3)` |
| AC-1.5 | `DEC 1817` against `JAN 1818`, `Tolerance` set to 0 | 0: any gap beyond the tolerance is a disagreement |

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
| AC-3.1 | `ABT 1817` against `BET 1820 AND 1821` | 0: the gap is 3 years and the tolerance stays 2 |
| AC-3.2 | `ABT 1817` against `ABT 1820` | 0.32: two points, so the tolerance doubles to 4 |
| AC-3.3 | `ABT 1817` against `BEF 1820` | 0.2 (overlap); against `BEF 1815` the gap is 2 years, inside the normal tolerance |

**US-4.** As a researcher, when one Event has a `date` and another has only `start_date` and `end_date`, I see them compared, so that a 1817 baptism can match a residence from 1815 to 1820.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Event profile. Probe `date` 1817. Candidate `start_date` 1815 and `end_date` 1820, no `date` | The `date` feature scores 0.55 (the point inside a six-year span), contribution 3.3. The reason names `1817` and the period |
| AC-4.2 | The same two Events with the roles swapped | Contribution 3.3 |
| AC-4.3 | Probe `date` 1823 against the same period | The `date` feature contradicts: −4 (the gap is 3 years, past the tolerance) |
| AC-4.4 | Both Events have a `date` | No period is built; `date` compares with `date` |
| AC-4.5 | Neither Event has a `date`, both have start and end | No period is built; only the `start_date` and `end_date` features run, as today |
| AC-4.6 | Candidate has `start_date` 1815 and no `end_date` | The period is open-ended (`FROM 1815`): probe `date` 1817 scores 0.2 |
| AC-4.7 | `EvaluateWithNames` for AC-4.1 | The `date` comparison is comparable, not a conflict |

**US-5.** As a contributor, when I run the date tests, every existing case and invariant still holds.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | The existing cases of `TestDateComparerScores`, `TestDateComparerLadder`, and `TestDateComparerInvariants` | Pass with no edit to an expected value |
| AC-5.2 | Generated dates that now include months and days on ranges | Symmetric, within 0…1, and a date against itself is 1 |
| AC-5.3 | `TestRegistrySetsEverySetting` | Passes: this PR adds no comparer setting |

## 3. Scope

| In | Covered by |
| --- | --- |
| Day windows for a date's first and last end, with year coordinates, exported beside `DateBounds` | STEP-1, AC-1.1, AC-1.3, AC-2.1, AC-2.2, AC-5.2 |
| `DateComparer` measures gaps and widths from those windows; same-year points keep their tiers | STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-5.1 |
| ABT widens the tolerance only for two points (the behavior the code has; locked by tests) | STEP-2, AC-3.1, AC-3.2, AC-3.3 |
| A date can compare with the other side's start and end as a period, only when this side has no `date` | STEP-3, AC-4.1, AC-4.2, AC-4.3, AC-4.4, AC-4.5, AC-4.6, AC-4.7 |
| Docs: `matching.md`, the spike design note, the map, `completed.md` | STEP-4 |

| Out | Home |
| --- | --- |
| Julian versus Gregorian dates, and double dating (1717/18) | [`docs/ideas/international-names.md`](../../../ideas/international-names.md) lists calendars; not planned here |
| A new weight for the period comparison (it uses the `date` feature's weight) | not planned (D-7) |
| Time-of-day precision (a day is the finest window) | not planned |
| Qualifiers on a period's start or end (`BEF` on a `start_date`) | not planned (D-7) |
| Comparing `start_date` with `end_date`, or any other pair of different Properties | not planned |
| Exhibits and Promote's evidence sheet comparing across Properties | not planned: `exhibits.go` pairs Observations of one Property |
| Changing `windowOf`, `DateBounds`, or the reconciler's treatment of one-sided ranges | not planned (D-2) |
| A configurable days-per-year or a finer tier model for same-year points | not planned (D-3) |

## 4. Current state

Verified at `ff14e3f`. Rows marked **S10-28** describe that plan's design, not code on `main`. Check them on landing.

| Area | Where | Today |
| --- | --- | --- |
| Date comparer | `core/match/dates.go` `DateComparer` (line 23), `spanOf` (line 64), `spanScore` (line 98), `yearsApart` (line 123), `Compare` (line 133) | Spans of whole years: `yearSpan{lo, hi, openLo, openHi, point, approx, month, day}`. A range reads `StartYear`..`EndYear` only. Two points in one year take a tier by month and day; points in different years fall off from `YearsApartFrom` by the difference of year numbers; spans score by the wider span's width in years, less a gap term |
| Date settings | `core/match/registry.go` `DefaultDates` (line 200), `DefaultProfile` event case (line 271) | `Tolerance` 2 years, `ApproxToleranceFactor` 2, tiers 0.9 / 0.8 / 0.7 / 0.6, `YearsApartFrom` 0.8, `SpanOneYear` 0.8, `SpanPerExtraYear` 0.05, `SpanFloor` 0.2. Event features `date` 6 / 4, `start_date` 3 / 2, `end_date` 3 / 2, each compared with the same Property only |
| Windows | `core/autoreconcile/dates.go` `windowOf` (line 287), `pointWindow` (line 296), `rangeWindow` (line 318), `spanOf` (line 332), `daysFromCivil` (line 389), `DateBounds` (line 68) | Inclusive proleptic day numbers. A range with only one side present is `!ok` there (no window). `DateBounds` returns the whole window's `lo` and `hi`, not each end |
| Pair selection | `core/match/match.go` `Score` (line 115), `best` (line 136); `core/match/pair.go` `evaluate` (line 65), `compareProperty` (line 90) | Each Feature or meta compares the same Property on both sides. **S10-28** replaces `best` with `pairValues(cmp, probe, candidate []Value) PairResult` in a new file, called from `Score` and `compareProperty`, and adds `Value.Label`, `Reason.Probe`, `Reason.Candidate` |
| Promote's metas | `core/database/promotealign/layer.go` `loadMetas` (line 82) | Takes every profile Property, so `date`, `start_date`, and `end_date` are all metas of an Event comparison |
| Tests | `core/match/dates_test.go` `TestDateComparerScores`, `TestDateComparerLadder`, `TestDateComparerInvariants`; `core/match/match_test.go` (`approx` at 1e-9) | Green at the base. The generated invariants use months and days on points but only years on ranges |
| Reconciler date module | `core/autoreconcile/dates.go` `dateModule` | Unchanged by this PR |

**Precedent.** `core/autoreconcile/dates.go` already turns a date into a day window and the cache stores its `date_lo` and `date_hi` (S9-21, migration 000034 onward). Copy that: derive windows from the date's components with `spanOf` and `daysFromCivil`, and do not add a second day-count routine. Do not copy `windowOf`'s refusal of a one-sided range: the matcher scores an open-ended range today.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Gaps and widths are in **years as a fractional number**, from year coordinates: a bound on civil day `d` of year `y` covers `[y + (doy−1)/n, y + doy/n)`, with `n` the days in `y`. A whole year is `[y, y+1)` and its centre is `y + 0.5`. Tolerance and every registry number stay in years. | Whole-year dates then give exactly the figures they give today (`1817` against `1818` is a gap of 1.0), so no existing score moves, and months and days give fractions of a year. | Dividing day counts by 365.25 — moves every year-level figure by about 0.0007 and breaks the 1e-9 test tolerance |
| D-2 | A new function `autoreconcile.DateEnds(d datevalues.Value) (first, last DateEnd, ok bool)` returns each end of a date. `DateEnd{Lo, Hi int; From, To float64; Open bool}` is the inclusive day span of that bound and its year coordinates. A point has the same span at both ends; `BEF` has an open first end; `AFT` an open last end; a range has its start bound first and its end bound last (swapped if inverted); a range with one side missing has an open end on that side. A point with no `StartYear` uses its `End*` fields, as `spanOf` in `core/match/dates.go` does today. `windowOf`, `DateBounds`, and the reconciler are unchanged. | It reuses `spanOf` and `daysFromCivil` so there is one day-count routine. It tolerates one-sided ranges because `TestDateComparerScores` scores them. | Calling `DateBounds` — gives the whole window, not each end, and has no year coordinates. Changing `windowOf` to accept one-sided ranges — changes what the reconciler treats as evidence |
| D-3 | Two **points** (exact or qualified, not `BEF` or `AFT`) in the same civil year keep their tier by month and day. Two points in different years score `YearsApartFrom × (1 − gap / (tolerance + 1))` within the tolerance, else 0, with `gap` the distance between the centres of their bounds in years. User answer 2026-10-16. | Every same-year and every year-level score holds. December and January score by days (0.777), May and September keep 0.6, which is the "outranks a same-year pair only when the days say so" the scope names. | Scoring every non-identical pair from the day distance — May and September would score 0.711 and the 0.9 / 0.8 / 0.7 / 0.6 tiers and their tests would change |
| D-4 | For a pair where either side is a range or a bound, the **gap** is 0 when the two day windows share a day (an open end counts as reaching past every date), else the distance in years between the centre of the later date's first end and the centre of the earlier date's last end. The score is the span score times `1 − gap / (tolerance + 1)`, 0 when the gap passes the tolerance. | A year-level range gives a gap of the difference of the nearest year numbers, as `yearsApart` does today (a point in 1822 against 1815–1820 is 2). Months and days now move the gap. | Edge-to-edge distance — a point in 1822 against 1815–1820 would be a gap of 1 and move today's score |
| D-5 | A span's **width** is `last.To − first.From` in years, and the span score uses `max(width, 1)`. | A range of whole years gives the integer width it gives today. A window under a year scores as a one-year span (0.8), so the PR adds no number above 0.8. | Letting a short window score above 0.8 — a new number nobody tuned |
| D-6 | ABT widens the tolerance only for two points. The code already does this (the factor applies only in the point branch). This PR changes no line for it, adds tests that lock it, and rewords the `DefaultDates` comment. | The scope line is satisfied by today's behavior; a test stops a refactor from losing it. | Rewriting the branch — nothing to fix |
| D-7 | **Period pooling.** The registry gets `DatePeriods` *(new)*: `date` → (`start_date`, `end_date`), product origin. In `Score` and `evaluate`, for a Feature or meta whose Property is a key of `DatePeriods`: when one side has no value of that Property and the other has at least one, the first side's values for the Property are replaced by period values built from its `start_date` and `end_date` values. A period is a range `Value` with the start value's first bound's year, month, and day and the end value's last bound's year, month, and day; qualifiers are dropped. One period per start × end pair, in `Values` order; a missing start or end leaves that side open. A side with neither gets no period. The period is a plain `Value` (not `Counted`), so S10-28 weighs it 1. The `date` feature's weight and contradiction are used. User answer 2026-10-16. | The scope says `date` may overlap `start_date` and `end_date` by window. Pooling only when `date` is missing counts nothing twice. Reusing the `date` weight adds no number. | Always comparing `date` against the period — counts start and end twice on Events that have all three. A period weight — a new number |
| D-8 | Pooling is a function `datePeriods(prop Property, probe, candidate Values) (a, b []Value)` in `core/match/periods.go` *(new)*, called once in `Score` and once in `evaluate` before the pair function. | One place decides, as S10-27 and S10-28 decided for the pair itself. | Pooling in the loaders — Promote and merge hints load differently, and the profile would not know |
| D-9 | No setting is added to `DateComparer`. | The numbers are the registry's; nothing to tune that was not already. `TestRegistrySetsEverySetting` stays as it is. | A days-per-year setting — D-1 removes the need |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `autoreconcile.DateEnds`, `autoreconcile.DateEnd` *(new)* | func, type | `autoreconcile.DateBounds`, `Window` | `DateBounds` returns one window with no per-end span and no year coordinates, and refuses nothing the matcher needs; the new function adds both beside it |
| `core/match/periods.go` *(new)* | file | `core/match/pair.go` | Pooling needs both sides' whole `Values`, which `compareProperty` does not get |
| `datePeriods`, `DatePeriods` *(new)* | func, var | `core/match/registry.go` `DefaultProfile` | The Property mapping is registry data, as profiles are |
| `dateSpan` in `core/match/dates.go` *(replaces `yearSpan`)* | type | `yearSpan` | It carries the two ends instead of two years |
| `BenchmarkDateComparer` *(new)* | benchmark | `core/match/names_test.go` | A date compare runs per pair per Feature; the new arithmetic must not allocate |

**Builds on.** S9-21's day windows and `spanOf`/`daysFromCivil` in `core/autoreconcile`; the pair function S10-28 creates; the profile and meta lists that already name `date`, `start_date`, and `end_date`.
**Deletes.** `yearSpan`, `spanOf`, and `yearsApart` in `core/match/dates.go` (replaced).
**Next adds.** Calendars (Julian and Gregorian, double dating) become another input to `DateEnds`, and nothing else changes. Another Property pair (a place's `founded` and `dissolved`) is one `DatePeriods` entry.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` `dates.go` | `DateEnd`, `DateEnds` (D-1, D-2). Built from the existing `spanOf` and `daysFromCivil`; adds year coordinates and one-sided ranges |
| `core/match` `dates.go` | `dateSpan` replaces `yearSpan`; `Compare` follows D-3, D-4, D-5, D-6. `spanScore`, `sameOptional` stay in shape; `yearsApart` becomes a gap in years |
| `core/match` `periods.go` *(new)* | `datePeriods` and the period builder (D-7, D-8) |
| `core/match` `registry.go` | `DatePeriods`; `DefaultDates` comments say "years" for gaps and widths and state the ABT rule (D-6) |
| `core/match` `match.go`, `pair.go` | `Score` and `evaluate` call `datePeriods` before the pair function (**S10-28**: `pairValues`) |

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
Do: `DateEnd` and `DateEnds` per D-1 and D-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile -run 'TestDateEnds'` passes (T-1), and the rest of `./core/autoreconcile` still passes.

**STEP-2 — `DateComparer` on date ends.**
Files: `core/match/dates.go`, `core/match/registry.go`, `core/match/dates_test.go`.
Do: replace `yearSpan` with `dateSpan`; D-3 through D-6; update the `DateComparer` and `DefaultDates` comments.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match -run 'TestDateComparer'` passes with every existing case unchanged (T-2, T-3, T-4); `CGO_ENABLED=1 go test -tags fts5 -bench BenchmarkDateComparer -run '^$' ./core/match` reports 0 allocs/op (T-8).

**STEP-3 — Period pooling.**
Files: `core/match/periods.go` *(new)*, `core/match/registry.go`, `core/match/match.go`, `core/match/pair.go`, `core/match/periods_test.go` *(new)*, `core/match/match_test.go`, `core/match/pair_test.go`.
Do: `DatePeriods`, `datePeriods`, and the two call sites (D-7, D-8).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match` passes (T-5, T-6, T-7).

**STEP-4 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/name-matching-enhancements.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-30-date-windows.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Score dates by the days they cover."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/dates_test.go` `TestDateEnds` *(new)* | Table-driven. `1817` → both ends `From` 1817.0, `To` 1818.0, `Lo`/`Hi` the year's first and last day. `MAY 1817` and a leap year `FEB 1816`: `From` and `To` from the day of year over 365 or 366. `14 MAY 1817`: one day. `BET MAR 1817 AND JUN 1817`: first end March, last end June. `BEF 1820`: first end open. `AFT 1820`: last end open. `FROM 1815` (no end): last end open. Inverted range swaps. `ABT 1817` is a point. A phrase with no year → not ok | AC-1.1, AC-2.1, AC-2.3, D-1, D-2 |
| T-2 | Go | `core/match/dates_test.go` `TestDateComparerScores` | Every existing case unchanged. New rows: `DEC 1817` vs `JAN 1818` → `0.8 × (1 − (31/365)/3)`; same with `Tolerance` 0 → 0; `MAY 1817` vs `SEP 1817` → 0.6; `DEC 1817` vs `1818` → `0.8 × (1 − gap/3)` with `gap = 1.5 − 349.5/365`; range `BET MAR 1817 AND JUN 1817` vs `14 MAY 1817` → 0.8, vs `14 SEP 1817` → `0.8 × (1 − (90.5/365)/3)` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-5.1, D-3, D-4, D-5 |
| T-3 | Go | same file `TestDateComparerABT` *(new)* | `ABT 1817` vs `BET 1820 AND 1821` → 0; `ABT 1817` vs `ABT 1820` → 0.32; `ABT 1817` vs `BEF 1820` → 0.2; `ABT 1817` vs `BEF 1815` → `0.2 × (1 − 2/3)` | AC-3.1, AC-3.2, AC-3.3, D-6 |
| T-4 | Go | same file `TestDateComparerLadder`, `TestDateComparerInvariants` | Unchanged; the generator also draws months and days on range ends | AC-5.1, AC-5.2 |
| T-5 | Go | `core/match/periods_test.go` `TestDatePeriods` *(new)* | Table-driven. Probe has `date`, candidate has `start_date` and `end_date` → candidate side gets one period. Both have `date` → no change. Neither has `date` → no change. Only a start → open-ended period. Two start values and one end → two periods in `Values` order. Neither side has any date Property → no change. A property with no `DatePeriods` entry → no change | AC-4.4, AC-4.5, AC-4.6, D-7, D-8 |
| T-6 | Go | `core/match/match_test.go` `TestScoreDatePeriod` *(new)* | `DefaultProfile("event")`. Probe `date` 1817, candidate `start_date` 1815 and `end_date` 1820 → the `date` reason is similarity 0.55, contribution 3.3, and the start and end features are skipped. Swapped → 3.3. Probe 1823 → contribution −4. Both Events with start and end only → two reasons (start, end), no `date` reason | AC-4.1, AC-4.2, AC-4.3, AC-4.5 |
| T-7 | Go | `core/match/pair_test.go` `TestEvaluateDatePeriod` *(new)* | `Evaluate` over metas `date`, `start_date`, `end_date` for the AC-4.1 values → the `date` comparison is comparable with similarity 0.55 | AC-4.7 |
| T-8 | Go | `core/match/dates_test.go` `BenchmarkDateComparer` *(new)* | Two points, a point against a range, a range against a range: 0 allocs/op | D-1 |
| T-9 | Go | `core/match/registry_test.go` `TestRegistrySetsEverySetting` | Passes with no new comparer setting | AC-5.3, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. A pre-existing test whose expected number changes is a stop-and-report (see Stop conditions).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host with no xcodebuild, the Go and plan-doc steps pass and the Xcode step is reported as not run
- [ ] No pre-existing test needed a changed expectation; if one did, the work stopped and reported
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] BenchmarkDateComparer reports 0 allocs/op
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-30 Plan line status; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: the `DateComparer` row says spans and gaps are in days from the date windows, with the December and January example, the same-year tiers, and the period rule; the "Not yet handled" bullet no longer names date windows (leave the other items in that bullet as they are)
- [ ] name-matching-enhancements.md: "Date precision below the year", "ABT on spans", and "Cross-property dates" say they landed in S10-30
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice has no new chrome, so the check is a catalog run, not a screen.

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestDateComparerScores|TestScoreDatePeriod' -v ./core/match`. See `DEC 1817` against `JAN 1818` score 0.777, `MAY 1817` against `SEP 1817` stay 0.6, and an Event with `date` 1817 score 3.3 against one with start 1815 and end 1820.
2. In a project with two Events, one dated December 1817 and one dated January 1818, add a Source with an Event dated January 1818 and open Promote. The December Event is suggested at a score above a same-year Event dated May 1817 would get for a September 1817 Subject.
3. In a project with a residence from 1815 to 1820, add a baptism dated 1817 and open Promote. The baptism's suggestions list the residence with a date comparison that is a partial, not unknown.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-28 lands differently from its plan, so `Score`, `evaluate`, or the pair function differ from §4 | §4 marks every S10-28 row. The first action is to redo them against the landed code |
| A year-level score moves in the last decimal | D-1's coordinates give exact figures for whole years; `TestDateComparerScores` compares at 1e-9 and is the check |
| A date compare runs per pair per Feature over every candidate | `DateEnds` allocates nothing and T-8 asserts 0 allocs/op |
| A suggestion that relied on two dates "overlapping" in one year drops | Intended for ranges with months (AC-2.2). The tier pairs do not move |
| An Event with a period now gets date evidence it never had | Intended (US-4). The same weight as `date`, so a period cannot outweigh a real `date` match |

**Stop and report if:**

- S10-28 has not landed, or `Score` and `compareProperty` do not both call one pair function. Redo §4 and tell the user what moved.
- Any existing case in `core/match/dates_test.go` needs a changed expected value to pass.
- `autoreconcile` has a caller that depends on `windowOf` accepting a one-sided range (it does not today; `DateEnds` must not change `windowOf`).
- `Value.Date` can hold a range whose end bound is earlier than its start bound after the swap in D-2.
- A `Values` Property in `DatePeriods` is read as a `date` Property of another value type in the catalog.

## 12. Open questions

None.

## 13. PR

**Title:** S10-30 — Date windows in matching
**Body:** Dates now compare by the days they cover, ranges keep their months and days, and an Event's `date` compares with another's start and end; what shipped against US-1 to US-5; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-30-date-windows.md`.
