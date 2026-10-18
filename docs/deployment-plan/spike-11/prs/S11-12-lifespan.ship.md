# Ship log — S11-12 Lifespan for people

**Plan:** [`S11-12-lifespan.md`](S11-12-lifespan.md) · **Branch:** `s11-12-lifespan` · **PR:** #480
**State:** ready

## Round 0 — implement
Head `73faf105`. Gate PASS (orchestrator run). Report deviations: STEP-1 leap case used offset 110 (plan's offset 10 was arithmetically wrong; matches AC-2.10, noted in Landed note); STEP-4 and STEP-5 one commit (signature changes force both); AC-2.11 nil-Dates and phrase-date events share one row; `TestEntityCompare` passes the person lifespan rule for every row as the plan says.

## Round 1 — review
UX, SEC, PRF: clean. CON: clean (two nits). MNT: changes needed.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Skills don't name `AddYears`, `ShiftYears`, `Earliest` / `Latest`, or the `subject.Lifespans` slot | Fix now → F-1 | verified: `add-date-value/SKILL.md:22` lists calendar.go without `AddYears`; `rg 'Earliest\|ShiftYears'` over both skills finds nothing; `add-match-rule` "Add a Subject property or pair" steps 2–3 have no Lifespans entry |
| MNT-2 | MNT | minor → major | `AboutEnds` year branch moves years by hand beside the new `ShiftYears` | Fix now → F-2 | verified `core/valuecompare/datewindow.go:258-261`. Re-rated: this PR adds `ShiftYears` as the year move (D-6) and leaves a parallel copy in the same file; "it predates the PR" does not lower it once the PR adds the shared home |
| MNT-3 | MNT | minor | `LifeEvent` repeats `EventRole`'s two fields | To user (non-blocking) | D-11 spells out separate fields; the embed looks right and is cheaper before S11-13's loader is written against it. Minor, so the loop does not pause; raised in the report |
| MNT-4 | MNT | nit | AC-3.5 invariant reads the registry again instead of using the group | Logged | test-only |
| CON-1 | CON | nit | `TestShiftYears` has no `!ok` row | Logged | T-1 does not ask for it |
| CON-2 | CON | nit | Test helper `with` not named in D-13 | Logged | test-only, no behaviour |
| SEC note | SEC | — | A contradicting side leaves any existing value under `lifespan` in place | Rejected | only reachable if a loader breaks the product-origin contract (`core/database/matching/matching.go:70,78`); `lifespan` is not seeded |
| UX note | UX | — | S11-13 should say why a person has no lifespan | Follow-up | S11-13's UX scope; not code in this PR |

### Fix plan 1

**F-1 — Name the shared date helpers and the Lifespans slot in the skills** (from MNT-1)
Files: `.agents/skills/add-date-value/SKILL.md`, `.agents/skills/add-match-rule/SKILL.md`.
Failing test first: none (docs); `python3 scripts/check-plan-docs.py` stays at 0 errors.
Do: add `AddYears` to the `calendar.go` line in `add-date-value`. In both skills' datewindow sentences name `ShiftYears` (year move on an end, then `boundEnd`) and `Earliest` / `Latest` as the one earliest / latest pick of date ends, beside `EndsDate` in `add-match-rule`'s "one way back" line. In `add-match-rule` "Add a Subject property or pair": step 2 adds "or a `Lifespans` entry"; step 3 adds the inventory paths `subject.Lifespans.<kind>.<lead>.Span`, `.MaxYears`, `.Starts`, `.Ends` (match the exact path form `TestKnobInventoryMatchesRegistry` and `docs/matching-model.md` use); add one line: a new life-fixing event is a keyed `EventRole` in `Starts` or `Ends`.
Check: scripts/check-pr.sh PASS.

**F-2 — Move `AboutEnds`' year branch onto `ShiftYears`** (from MNT-2)
Files: `core/valuecompare/datewindow.go`.
Failing test first: none (refactor with no behaviour change); the existing `AboutEnds` tests and `TestShiftYears` must stay green.
Do: in the `m == 0` branch of `AboutEnds`, replace the two `boundEnd(intPtr(y±units), nil, nil)` calls with `ShiftYears(first, -units)` and `ShiftYears(last, units)`, shifting from the ends `DateEnds` returned (keep the originals in locals so the second shift does not start from the moved first end). Drop `intPtr` only if nothing else uses it.
Check: `go test ./core/valuecompare/... ./core/subjectcompare/... ./core/subjectprofile/...` pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a72660b0`. Gate PASS (orchestrator run; the first run failed `diff-check` on a trailing blank line in this ship log, the orchestrator's own edit). F-1 done (`a72660b0`), F-2 done (`b46926eb`). Note: round 0's `65921efc` committed an early copy of this ship log; harmless, the final commit replaces it.

## Round 2 — review
MNT only (the one lens with Fix-now findings), scope `73faf105...a72660b0`. MNT-1 and MNT-2 resolved. No new blockers.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-R2-1 | MNT (CON note) | minor | An ABT year-precision point with a day but no month now keeps `DateEnd.Day` after widening (`ShiftYears` → `endPart(d)`); the old code dropped it. `Lo`/`Hi` unchanged; `EndsDate` now gives the day back | Follow-up | verified: `core/datevalue/datevalue.go:139` allows day-without-month gaps. Consistent with `DateEnds` for non-ABT values; round 2 takes only blockers. Pin it with a test |
| MNT-R2-2 | MNT | nit | `lo, hi` locals in `AboutEnds` are unneeded | Logged | — |

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-R2-1, MNT-4, CON-1, UX note (S11-13) · to user: MNT-3
