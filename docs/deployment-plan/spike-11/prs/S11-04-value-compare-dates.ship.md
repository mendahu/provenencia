# Ship log — S11-04 Value compare: dates

**Plan:** [`S11-04-value-compare-dates.md`](S11-04-value-compare-dates.md) · **Branch:** `s11-04-value-compare-dates` · **PR:** #445
**State:** ready

Worktree: `.claude/worktrees/s11-04-value-compare-dates`, cut from `main` @ `79f7f042`. S11-03 is in flight on its own branch; plan risk row covers the shared files (additive; one `CauseBelowMinimum`).

## Round 0 — implement
Head `a3a6cac6`. Gate PASS (orchestrator run; Go results cached from the implementer's full run). Report deviations: `dateValue`/`ofDate` builders added in STEP-4 rather than STEP-2 so each commit compiles; extra test builders `optInt`, `dp`, `dateValue`; unexported `intPtr` in `datewindow.go`. Map row and plan Status already set to `Landed #445` by the implementer.

## Round 1 — review
Verdicts: CON changes needed · MNT changes needed · UX clean · SEC clean · PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `match.centre` duplicates `DateEnd.Mid` | Fix now → F-1 | verified: `core/match/dates.go:84` computes `(From+To)/2`; `YearLine` fills `Mid` with the same; L84 is already in the diff |
| CON-1 | CON | major | No Landed note; PR body says "Deviations: None" | Fix now → F-2 | verified: plan ends at `## 13. PR`; S11-02 plan has `## Landed`. PR body and manual check are corrected by the orchestrator at Finish |
| CON-2 | CON | minor | ABT day-with-no-month (D-6) not pinned in `TestAboutEnds` | Follow-up | code correct; fix plan does not touch `datewindow_test.go` |
| MNT-2 | MNT | minor | `YearLine` re-derives the year; one-civil-year rule only in a comment | Follow-up | fix plan does not touch `calendar.go`; no behavior change |
| MNT-3 | MNT | minor | `Validate` (day 1–31) vs `BoundDays` (1–`DaysInMonth`) | Follow-up | pre-existing divergence; idea file if the user agrees |
| MNT-4 | MNT | nit | `add-date-value` tree omits `calendar_test.go` | Logged | — |
| CON-3 | CON | nit | `spanOf` clock-field count rewritten as a loop | Logged | behavior same; listed in the Landed note (F-2) |

### Fix plan 1
**F-1 — Read `DateEnd.Mid` in match instead of `centre`** (from MNT-1)
Files: `core/match/dates.go`.
Failing test first: none — pure refactor; the arithmetic is identical, so the existing `core/match` tests are the check (AC-4.1: no expected value changes).
Do: delete `func centre`; replace `centre(x)` with `x.Mid` at the three call sites (L121, L123, L157). Keep the doc comments' wording true ("centres" still describes `Mid`).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match/...` passes unchanged; scripts/check-pr.sh PASS.

**F-2 — Add the plan's Landed note** (from CON-1, CON-3)
Files: `docs/deployment-plan/spike-11/prs/S11-04-value-compare-dates.md`.
Do: add a `## Landed` section after `## 13. PR`, in the shape of `S11-02-value-compare-core.md`'s, listing every difference from the plan: `dateValue`/`ofDate` builders added in STEP-4 not STEP-2 (so each commit compiles); extra test builders `optInt`, `dp`, `dateValue` beyond D-10; unexported `intPtr` in `datewindow.go`; `AboutEnds` cases in their own `TestAboutEnds`; the extra "below the near minimum" subtest; `spanOf`'s clock-field counting rewritten as a loop (same behavior); `match` reads `DateEnd.Mid` instead of its own `centre` (F-1).
Check: `python3 scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `ae089f50`. Gate PASS (orchestrator run). F-1 done (`48cb252a`); F-2 done (`ae089f50`). Implementer's first gate run hit a flaky `EvidenceCardLayoutTests measuredLayoutMatchesThePaint` (Swift, untouched code); unchanged rerun passed.

## Round 2 — review
Lenses: CON, MNT (the two with Fix-now findings). Scope `a3a6cac6...ae089f50`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | CON | — | CON-1 resolved; F-1 bit-identical; no new findings | — | Landed note claims checked against code |
| — | MNT | — | MNT-1 resolved; no new findings | — | `centre` gone; `Mid` from `YearLine` only |

No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: CON-2, MNT-2, MNT-3 · logged: MNT-4, CON-3 · to user: —
