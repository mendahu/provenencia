# Ship log — S11-05 Migrate matching onto value compare

**Plan:** [`S11-05-migrate-matching.md`](S11-05-migrate-matching.md) · **Branch:** `s11-05-migrate-matching` · **PR:** #451
**State:** review 1 — paused for user (CON-1, UX-1)

Worktree: `.claude/worktrees/s11-05-migrate-matching`, cut from `main` @ `75ed63c`.

## Round 0 — implement
Head `31b94d6`. Gate PASS (orchestrator run). Report deviations: also removed the names "Worked examples" table and `names_test.go` paragraph from `docs/matching.md` (describe deleted code); a few extra tests (AC-1.5 / AC-2.1 subtests, `surnameVals`, a `valuerules` check in `TestNoStoredDefaultStampsWestern`). Base was `75ed63c`, not the plan's `b6d1d48`; no §4 row affected.
Not done: plan Status → Landed and its Landed note — `check-plan-docs.py` fails on backticked paths the PR deletes (§4 L115–117, L132; D-3; D-11; "Deletes" line). Held for the user at finish. Manual check §10 steps 3–4 (in-app Promote) left for the user.

## Round 1 — review
Verdicts: CON Changes needed · MNT Clean · UX Clean · SEC Clean · PERF Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan not marked Landed; `check-plan-docs.py` fails on paths this PR deletes (also turns S11-02 and S11-03 plans red) | To user | confirmed: `check_paths` errors on every non-`(new)` missing path; the fix touches either the checker or D-3/D-11 lines |
| UX-1 | UX | minor | D-16 says the researcher sees no change, but the Promote menu no longer lists a record whose first name differs (Skip+alternatives → New, `weak` → `no_match`) | To user | confirmed `align.go` no-assignment branch; contradicts a Decision's stated rationale, behaviour itself is the design's |
| MNT-1 | MNT | minor | `pairValues` keeps `top` and `topSim`, now always equal | Pending user (rides along if a fix round runs) | confirmed `core/match/pairs.go:334-363` |
| MNT-2 | MNT | minor | Test comment names deleted `DefaultText` | Pending user (rides along) | `core/graphalign/align_test.go:289` |
| CON-2 | CON | minor | AC-1.4 test doesn't assert `ReasonNoMatch` | Pending user (rides along) | `align_test.go:545` checks Target and Assessment only |
| MNT-3 | MNT | nit | `matchregistry/value.go` comments name deleted identifiers | Logged | plan forbids touching `matchregistry`; S11-16 |
| CON-3 | CON | nit | PR body Deviations omit base drift and held Landed note | Logged | folded into finish PR body |
| SEC | SEC | — | No findings | — | — |
| PERF | PERF | — | No findings; Rank ~2× faster, Propose drops a full name scan | — | — |
