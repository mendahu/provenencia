# Ship log — S10-60 The gallery reads only gallery records

**Plan:** [`S10-60-gallery-member-records.md`](S10-60-gallery-member-records.md) · **Branch:** `s10-60-gallery-member-records` · **PR:** #459
**State:** ready

Base: `origin/main` @ `e97cf54d` (plan Planned against `aa60b63`). Worktree: `.claude/worktrees/s10-60`.

## Round 0 — implement
Head `59c5f763`. Gate PASS (orchestrator run: go-test 123s, xcode-test 14s). Report: no code deviations. Status and map row use the checker's `Landed in [#459]…` form rather than the skill's literal `Landed #N`. No `completed.md` in this folder. §10 manual check left to the user.

Note: the implementer's Record commit (`59c5f763`) committed this ship log's stub, though its report called the file untracked. Harmless; the log belongs on the branch.

## Round 1 — review
Lenses: CON, MNT, UX, SEC, PERF on `e97cf54d...59c5f763`. CON, SEC, PERF clean (CON re-ran all seven new tests against `v.reason` in an export: all fail).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `ByObservation` and `LikenessTile` comments say "an Observation that went into it", the old any-outcome rule | Fix now → F-1 (Go comments only) | verified: `likenesses.go:113`, `conclusiondetails.go:104`; "went into" means value rank in `000040.sql:8`, `conclusiondetails.go:74`. The same words in `engine.proto:1356` and `GenealogyStore.swift:608` → Follow-up: the plan's Out line rules out any Swift, FFI, or proto change |
| UX-1 | UX | minor | A portrait stored before this PR at a non-gallery record silently falls back to the earliest face | Follow-up (PR body note) | D-4 and AC-1.5 accept it; re-pointing would reopen D-4. Raised with the user in the report |
| MNT-2 | MNT | nit | T-6 rebuilds the `negativeTwin` catalog inline in `conclusiondetails_test.go` | Logged | a test package cannot share another's fixture; lift into a shared builder if a third package needs it |

### Fix plan 1
**F-1 — Name gallery records in the two gallery doc comments** (from MNT-1)
Files: `core/database/likenesses/likenesses.go` (`ByObservation`), `core/database/conclusiondetails/conclusiondetails.go` (`LikenessTile`). Skill: none.
Failing test first: none (comment only).
Do: `ByObservation` → "returns the index of the gallery likeness whose gallery records include the Observation: the one rule for which likeness a record names. -1, false when none." `LikenessTile` → "is one likeness a gallery shows: its image, and the outcome of its earliest gallery record, which the tile opens." Do not touch `api/proto/engine.proto` or Swift (plan Out). One commit.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `cc2e0aac`. Gate PASS. F-1 done (comment only). First gate FAILED diff-check on a trailing blank line in this log (orchestrator edit), not in the fix; trimmed.

## Round 2 — review
MNT only, on `59c5f763...cc2e0aac`. MNT-1 resolved; no new findings. No Fix plan 2.

## Result
Gate PASS · CI see PR #459 · open: none · follow-ups: MNT-1 (proto and Swift doc comments), UX-1 (PR body note), MNT-2 (logged) · to user: UX-1 (accept D-4's silent fallback, or re-point in a later PR)
