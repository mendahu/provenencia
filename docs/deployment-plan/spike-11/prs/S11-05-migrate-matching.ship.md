# Ship log — S11-05 Migrate matching onto value compare

**Plan:** [`S11-05-migrate-matching.md`](S11-05-migrate-matching.md) · **Branch:** `s11-05-migrate-matching` · **PR:** #451
**State:** fixing 1

Worktree: `.claude/worktrees/s11-05-migrate-matching`, cut from `main` @ `75ed63c`.

## Round 0 — implement
Head `31b94d6`. Gate PASS (orchestrator run). Report deviations: also removed the names "Worked examples" table and `names_test.go` paragraph from `docs/matching.md` (describe deleted code); a few extra tests (AC-1.5 / AC-2.1 subtests, `surnameVals`, a `valuerules` check in `TestNoStoredDefaultStampsWestern`). Base was `75ed63c`, not the plan's `b6d1d48`; no §4 row affected.
Not done: plan Status → Landed and its Landed note — `check-plan-docs.py` fails on backticked paths the PR deletes (§4 L115–117, L132; D-3; D-11; "Deletes" line). Held for the user at finish. Manual check §10 steps 3–4 (in-app Promote) left for the user.

## Round 1 — review
Verdicts: CON Changes needed · MNT Clean · UX Clean · SEC Clean · PERF Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Plan not marked Landed; `check-plan-docs.py` fails on paths this PR deletes (also turns S11-02 and S11-03 plans red) | To user → skip landed plans → F-1 | confirmed: `check_paths` errors on every non-`(new)` missing path; the fix touches either the checker or D-3/D-11 lines |
| UX-1 | UX | minor | D-16 says the researcher sees no change, but the Promote menu no longer lists a record whose first name differs (Skip+alternatives → New, `weak` → `no_match`) | To user → accept, document → F-2 | confirmed `align.go` no-assignment branch; contradicts a Decision's stated rationale, behaviour itself is the design's |
| MNT-1 | MNT | minor | `pairValues` keeps `top` and `topSim`, now always equal | Fix now → F-3 (user: fold minors in) | confirmed `core/match/pairs.go:334-363` |
| MNT-2 | MNT | minor | Test comment names deleted `DefaultText` | Fix now → F-4 | `core/graphalign/align_test.go:289` |
| CON-2 | CON | minor | AC-1.4 test doesn't assert `ReasonNoMatch` | Fix now → F-4 | `align_test.go:545` checks Target and Assessment only |
| MNT-3 | MNT | nit | `matchregistry/value.go` comments name deleted identifiers | Logged | plan forbids touching `matchregistry`; S11-16 |
| CON-3 | CON | nit | PR body Deviations omit base drift and held Landed note | Logged | folded into finish PR body |
| SEC | SEC | — | No findings | — | — |
| PERF | PERF | — | No findings; Rank ~2× faster, Propose drops a full name scan | — | — |

User answers (2026-10-17): CON-1 → teach the checker to skip path existence on landed plans; UX-1 → accept the behaviour, document it, add a `docs/ideas/` follow-up; minors → fix now.

### Fix plan 1

**F-1 — Landed plans skip the path-existence check; mark S11-05 Landed** (from CON-1, CON-3)
Files: `scripts/check-plan-docs.py`, `docs/deployment-plan/spike-11/prs/S11-05-migrate-matching.md`.
Failing test first: none (script); run `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-0{2,3}-*.md` and see the "path does not exist" errors first.
Do: in `check_plan`, call `check_paths` only when Status does not match `Landed in …` (a landed plan is a historical record; its paths may be deleted by later PRs). Find where `check_paths` is invoked and gate it there; update the module docstring line "every repo path in backticks exists, unless its line says (new)" to add "(not checked once the plan is Landed)". Then do the plan's STEP-6 edit you reverted in round 0: Status → `Landed in [#451](https://github.com/mendahu/provenencia/pull/451) (2026-10-17)` and a `## Landed` note in the S11-03 / S11-04 shape listing every difference from the plan: base `75ed63c` not `b6d1d48`; the extra `docs/matching.md` cuts; the extra tests; `valueType()` checked on both sides; the checker change itself; and the UX-1 correction from F-2. Do not edit any Decision, Scope or AC line.
Check: `python3 scripts/check-plan-docs.py` on S11-02, S11-03 and S11-05 plans reports 0 errors; scripts/check-pr.sh PASS.

**F-2 — Record D-16's real user effect** (from UX-1)
Files: the S11-05 plan's `## Landed` note; a new `docs/ideas/` file following that folder's existing format.
Do: in the Landed note, say that D-16's "the researcher sees no difference" is not quite right: a record whose first given name or surname differs is no longer offered in the Promote target menu (the row opens New / `no_match` with no alternatives instead of Skip / `weak` with alternatives), so linking such a record from Promote needs another path. The behaviour is the design's; the user accepted it. Add a short idea doc: a "find another record…" entry in the Promote target menu, so a researcher can link a record `Rank` did not offer. No code.
Check: scripts/check-pr.sh PASS (plan-docs).

**F-3 — Drop the duplicate `topSim` in `pairValues`** (from MNT-1)
Files: `core/match/pairs.go` (around L334–363).
Do: `top` and `topSim` now always hold the same value; keep one (`top`), return `Similarity: top`. Behaviour unchanged.
Check: `go test ./core/match/...` passes unchanged; scripts/check-pr.sh PASS.

**F-4 — AC-1.4 test asserts the reason; fix a stale comment** (from CON-2, MNT-2)
Files: `core/graphalign/align_test.go`.
Failing test first: n/a (asserts current behaviour) — add `wantReason` to the table in the subtest "a different first name is no candidate" (`TestUnreachableFallbackScoresOnTheWalkScale`, ~L517–548): `graphalign.ReasonNoMatch` for that case and whatever the second case produces today; compare `row.Reason` in the `t.Fatalf` condition. At L289 reword "toponym + DefaultText Rank" to "toponym + text value compare in Rank".
Check: `go test ./core/graphalign/...` passes; scripts/check-pr.sh PASS.
