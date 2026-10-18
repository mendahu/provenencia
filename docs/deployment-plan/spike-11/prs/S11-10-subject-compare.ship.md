# Ship log — S11-10 Subject compare

**Plan:** [`S11-10-subject-compare.md`](S11-10-subject-compare.md) · **Branch:** `s11-10-subject-compare` · **PR:** #471
**State:** ready

Base `f4bde7ad` (main).

## Round 0 — implement
Head `326dbbab`. Gate PASS (orchestrator run). PR #471 (draft).
Report deviations: none in code or behavior. AC-1.14 asserts cause `relation` (the AC names none; recorded what `Compare` returns with no meta). STEP-3–5 package code and registry section in one commit (`234e74f0`), full `Compare` instead of the D-4 stub; tests in later commits.

## Round 1 — review
Reviewers at `326dbbab`: CON changes needed, MNT changes needed, UX clean, SEC clean, PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | AC-2.7 test passes with D-7's "no positive lead value" guard deleted | Fix now → F-1 | reviewer mutated `pairs.go:20` in a scratch copy; `TestCompare` still green. Carries its failing-under-mutation row |
| CON-4 (from SEC outside-lens) | CON | major | `EndsDate` drops a year of 0: `endPart(first.Year)` → nil, so year-0 ends rebuild open / lose their year | Fix now → F-2 | verified `core/valuecompare/datewindow.go:79-99`; `datevalue` has no year bound, `BoundDays` accepts 0. AC-4.4's round trip breaks; S11-12 reuses `EndsDate` |
| CON-2 | CON | minor | Landed note says "None" but round 0 had two differences | Fix now → F-3 | ship log round 0 lists them; DoD asks for every difference. Few lines in the plan file |
| CON-3 | CON | nit | No row covers `period` skipping a negative start/end | Fix now (rides on F-1) | same test file as F-1, one row |
| MNT-1 | MNT | major | `match.period`/`endFields` duplicates the new `EndsDate` | Rejected | plan §1 "What stays": `core/match` (`period`, `endFields`) untouched; D-8 reason and Codebase fit row: "`match`'s private copy goes in S11-13". The skill line is plan-mandated text (STEP-9). Already scheduled in S11-13 |
| MNT-2 | MNT | minor | `invariants_test.go` `kindGroups` hand-copies the registry's leads/pools/periods | Follow-up | S11-12 adds lifespan as a lead; tie `kindGroups` to `matchregistry.Default().Subject` there (or a check that they match) |
| MNT-3 | MNT | minor | `period` repeats its start/end selection loop | Follow-up | fix plan does not touch `pairs.go`; fold into S11-12, which touches period building |
| MNT-4 | MNT | nit | `endPart` duplicates `datevaluetest.Part` | Logged | |
| PRF outside-lens | CON | — | `split` now uses `Positive()`; neutral terms no longer positive? | Rejected | CON verified: neutrals are dropped with `continue` before it (`propertycompare.go:166-175`); PASS count unchanged 61 → 61 |

### Fix plan 1

**F-1 — AC-2.7: a recorded date keeps its period out** (from CON-1, CON-3)
Files: `core/subjectcompare/subjectcompare_test.go`.
Failing test first: there is no code bug; prove the row bites by mutation. Add to `TestCompare`, beside the existing AC-2.7 row:
```go
{name: "AC-2.7 a recorded date keeps its period out", kind: vocabkey.Event,
	a: Vals{date: {year(1817)}, start: {year(1815)}, end: {year(1820)}}, b: Vals{date: {pt.Neg(year(1818))}}, sim: 0, ev: 0,
	check: is(date, unknown, noVals, 0)},
```
(adapt names to the file's helpers). Temporarily remove `&& !slices.ContainsFunc(out, propertycompare.Value.Positive)` at `pairs.go:20`, confirm the row fails, restore it, confirm it passes. Also add one row for CON-3: an event with start `{¬1815}` (via `pt.Neg`), end `{1820}`, no date, against date `{1817}`; the period must read `BEF 1820`. Assert what the code returns, after confirming the row fails when the `!v.Positive()` skip in `period`'s start loop is removed (restore it after).
Do: tests only. No product code change.
Check: both rows pass at head and fail under their mutations (report the output); scripts/check-pr.sh PASS.

**F-2 — `EndsDate` keeps a year of 0** (from CON-4)
Files: `core/valuecompare/datewindow.go`, `core/valuecompare/datewindow_test.go`.
Failing test first: `TestEndsDate` round-trip case for year 0 (a point `0`, and a range `0 – 5`, built with `datevaluetest`): `DateEnds(EndsDate(DateEnds(d)))` equals `DateEnds(d)`. Today `endPart(0)` makes `StartYear` nil.
Do: a closed `DateEnd` always has a year, so set the year pointer unconditionally in `EndsDate` and `endsPoint` (take the address of a copy of `e.Year`); keep `endPart` for month and day only, and update its doc comment if needed. No other change.
Check: the new case fails before, passes after; `TestEndsDate` all pass; scripts/check-pr.sh PASS.

**F-3 — Landed note lists round 0's differences** (from CON-2)
Files: `docs/deployment-plan/spike-11/prs/S11-10-subject-compare.md` (the **Landed:** line).
Do: replace "None — the code is as planned." with the two differences: STEP-3 to STEP-5 landed in one commit (`234e74f0`) with the full `Compare` instead of STEP-3's D-4-only skeleton; the AC-1.14 row asserts cause `relation`, which the AC does not name. Add a third if F-2 is described as a fix to `EndsDate` (it is not a plan deviation; skip it). Keep the note's existing style.
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `441297a2`. Gate PASS (orchestrator run). F-1 done (`23baca48`; both rows fail under their mutations, pass restored). F-2 done (`e2eb1906`; year-0 point and range cases failed first). F-3 done (`441297a2`).

## Round 2 — review
CON only (the one lens with Fix-now findings), scope `326dbbab...441297a2`: clean. Each round-1 fix re-checked by mutation in a scratch export: CON-1, CON-3 and CON-4 rows fail with the code they cover broken; CON-2 Landed note lists both differences. No new blockers.

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-2, MNT-3 (in S11-12's map scope) · to user: —
