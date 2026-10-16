# Ship log — S10-30 Date windows in matching

**Plan:** [`S10-30-date-windows.md`](S10-30-date-windows.md) · **Branch:** `s10-30-date-windows` · **PR:** #415
**State:** ready

## Round 0 — implement
Head `ba22f75b`. Gate PASS (orchestrator run, against `4ea7483b`). Report deviations:
- No Landed note added ("the code does not differ from the plan") — see CON-2.
- AC-1.6 rows sit in `TestDateComparerScores` ("a day against …"); `TestDateComparerABT` is the new ABT test.
- First commit amended once before any push.

## Round 1 — review
Head `ba22f75b`. Verdicts: CON changes needed · MNT clean · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | A point with a day but no month scores as identical to any date in its year | Fix now → F-1 | verified: `boundEnd` (`core/autoreconcile/dates.go:150-155`) sets `Day` only inside `if month != nil`, so `core/match/dates.go:162` takes the identical tier; `datevalues` accepts day without month; reviewer's test fails at head, passes on base |
| CON-2 | CON | major | The plan has no Landed note | Fix now → F-2 | verified: plan diff changes only `**Status:**`; DoD requires the note; S10-28/S10-50 have one |
| CON-3 | CON | minor | Enhancements note marks precision landed but body still says whole years | Fix now → F-3 | verified: `name-matching-enhancements.md:29-31` unchanged present tense; docs-only, a few lines, and it is now wrong |
| MNT-1 | MNT | minor | `DateEnds` and `windowOf` are two date→day dispatchers that already differ | Follow-up | D-2 fixed this shape; unify before calendars land |
| MNT-2 | MNT | nit | `match.spanOf` rederives point/bound after `DateEnds` | Logged | same behavior as base; goes with MNT-1 |
| MNT-3 | MNT | nit | `year(y)` duplicates `date(y,nil,nil)`; `span2` is a weak name | Logged | test helpers only |
| MNT-4 | MNT | nit | `"gregorian"` literal in two places | Logged | add a constant when calendars land |
| MNT-5 | MNT | nit | `DatePeriods` comment overstates the extension slot | Logged | `registry.go` not in the fix plan |
| UX-1 | UX | minor | Pooled period shows on the Promote sheet as a plain Date line, qualifiers dropped | Follow-up | D-7 drops qualifiers; marking it derived needs a proto field and copy |
| SEC-1 | SEC | minor | Unbounded catalog years overflow date math; scores outside [0, 1] | Follow-up | pre-existing at base (68 pairs at base vs 43 at head); hardening PR |
| PRF-1 | PRF | minor | Pooling and day-window compares add a constant cost per pair in `Rank` (+45% dated, 17 allocs pooling) | Follow-up | plan's alloc bars (D-8, T-8, T-10) met; cache per side / per Rank later |

### Fix plan 1

**F-1 — Keep a day with no month in `DateEnd`** (from CON-1)
Files: `core/autoreconcile/dates.go`, `core/match/dates_test.go`. Skill: `add-date-value`.
Failing test first: add these rows to `TestDateComparerScores` (or as `TestDateComparerDayWithoutMonth` if the table's shape cannot take a raw `Value`), built from `dayOnly := func(y, d int) Value { return Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &y, StartDay: &d}} }`:
- "different days, no month": `dayOnly(1817, 14)` vs `dayOnly(1817, 20)` → 0.8
- "a day with no month against the year": `dayOnly(1817, 14)` vs `date(1817, nil, nil)` → 0.8
Both fail at head (got 1).
Do: in `boundEnd`, set `e.Day = *day` whenever `day != nil`, not only inside `if month != nil`. The window (`Lo`/`Hi`, `From`/`To`) stays the whole year, as `spanOf` already places it. Update the `DateEnd` doc comment if it implies Day needs Month. Do not change the tier switch in `core/match/dates.go`: with `Day` kept, `ea.Month == eb.Month && ea.Day == eb.Day` no longer matches different days, and `sameMonth` already requires `Month != 0`. Confirm `endFields` in `core/match/periods.go` now carries the day through (no code change expected).
Check: the new rows pass; every existing `dates_test.go`, `periods_test.go` and `TestDateEnds` expectation is unchanged; `scripts/check-pr.sh` PASS.

**F-2 — Add the plan's Landed note** (from CON-2)
Files: `docs/deployment-plan/spike-10/prs/S10-30-date-windows.md`.
Do: add `**Landed:**` after `**Review:**`, in the shape of `S10-28-support-negatives.md:18` and `S10-50-promote-sheet-values.md:18`. List every difference from the plan:
- test helpers `denied`, `year`, `span2` added beside the plan's named ones;
- extra tests `TestPoolPeriodsKeepsNonVoting`, `TestPoolPeriodsMonthsAndDays`;
- extra "both dated" and "neither dated" cases in `BenchmarkPoolPeriods`; the "pooling" case allocates 6 times per op;
- T-7 calls `Evaluate` where AC-4.7 names `EvaluateWithNames` (same `evaluate` path);
- AC-1.6 rows live in `TestDateComparerScores`;
- F-1: a day without a month is kept on `DateEnd` (review round 1, CON-1), with its test.
Check: `scripts/check-plan-docs.py` clean; `scripts/check-pr.sh` PASS.

**F-3 — Put the enhancements note in the landed tense** (from CON-3)
Files: `docs/deployment-plan/spike-10/name-matching-enhancements.md` (lines 29-31).
Do: rewrite the "Date precision below the year" body so it no longer says spans are counted in whole years in the present tense: say spans and gaps are now measured in days from the stored windows, with the landed figures (December 1817 vs January 1818 0.777; "BET MAR 1817 AND JUN 1817" against a later month 0.734, per AC-1/AC-2). Keep the "S9-21 stored the windows" line accurate.
Check: `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `2b58f8b6`. Gate PASS (orchestrator run, against `4ea7483b`). F-1 done (`TestDateComparerDayWithoutMonth` red first, then green; a separate function because the table's `date()` helper cannot build a day without a month); F-2 done; F-3 done. `DateEnd` doc comment left as is (already says Month and Day are 0 when not named).

## Round 2 — review
Scope `81de1124...2b58f8b6`, Conformance only (the one lens with Fix-now findings). Verdict clean: CON-1, CON-2, CON-3 resolved; `endFields` now carries a day without a month into a pooled value, as stored. One wording nit logged: the enhancements note says "against a later month" where AC-2.2's comparand is the day 14 SEP 1817. No blocker, so no Fix plan 2.

## Result
Gate PASS · CI see PR #415 · open: none · follow-ups: MNT-1, UX-1, SEC-1, PRF-1 · to user: —
