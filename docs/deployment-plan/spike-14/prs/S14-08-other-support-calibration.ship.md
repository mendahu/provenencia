# Ship log — S14-08 Other-link support calibration

**Plan:** [`S14-08-other-support-calibration.md`](S14-08-other-support-calibration.md) · **Branch:** `s14-08-other-support-calibration` · **PR:** #613
**State:** ready

## Round 0 — implement

Head `e2756f9b`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). Report: deviation — `TestProposeKinshipShapes/AC-2.2` expects Medium (3.375) after OtherSupport 0.75; manual §10 not run in app.

## Round 1 — review

Reviewers: CON, MNT, SEC+PRF (composer-2.5-fast after Opus/Sonnet limit). Skipped: UX (no `macos/App/**`, Localizable, or new error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | §10 manual checks not run | Logged | User runs §10 at merge; T-1 locks band math |
| CON-2 | CON | minor | AC-1.2 `Unplaced` not exercised | Fix now → F-1 | Test file this PR added; cheap second case |

Nits (logged): `README Batch 2 not in STEP-6` · `OneSupport comment column align` · `MNT-1 OneSupport column` · `MNT-2 table vs linear test`

### Fix plan 1

**F-1 — Exercise Unplaced in the calibration test** (from CON-2)
Files: `core/graphalign/align_test.go`. Skill: —.
Failing test first: extend `TestPartialNameWithGrandparentClearsWeak` with a subtest or second call using `segmentcompare.Unplaced` (Carol neighbors already ≥ `UnplacedAlike`); assert same Weak≤score&lt;Medium window.
Do: keep the Together case; add Unplaced beside it (table or `t.Run`).
Check: `go test ./core/graphalign/ -run TestPartialNameWithGrandparentClearsWeak` passes; `scripts/check-pr.sh` PASS.

## Round 1 — fix

Head `5d1c4ce2`. Gate PASS (orchestrator). F-1 done — Together and Unplaced subtests in `TestPartialNameWithGrandparentClearsWeak`. CON-2 Resolved (orchestrator read the fix). No round 2 (no Fix-now blocker or major).

## Result

Gate PASS · CI pending · open: none · follow-ups: none · to user: §10 manual check · CON miss: none
