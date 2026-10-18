# Ship log — S11-11 Subject profile and entity compare

**Plan:** [`S11-11-subject-profile.md`](S11-11-subject-profile.md) · **Branch:** `s11-11-subject-profile` · **PR:** #476
**State:** ready

## Round 0 — implement
Head `c3904b60` (base `origin/main` @ `f39e2e14`). Gate PASS. Report: STEP-2 and STEP-3 in one commit (`Build` written with `Options.Excluding` from the start); test expectations use the real `Labels` format (`sex_at_birth={male}`) rather than the plan's schematic form.

## Round 1 — review
Reviewers at `c3904b60`. Verdicts: CON Clean (2 minor, 1 nit) · MNT Clean (3 minor, 1 nit) · UX Clean · SEC Clean · PERF Clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor → major | AC-1.9 test does not check the `Name` pointer is shared | Fix now → F-1 | Re-rated: AC-1.9 says "`Name` is the reconciled value's `Name` pointer" and T-1 claims AC-1.9 and D-5; the row checks only `Form`, so the AC has no evidence. Code this PR adds |
| CON-2, MNT-3 | CON, MNT | minor → major | Map working rule and design table still say "concluded" / "researcher-denied" | Fix now → F-2 | Verified: `deployment-plan.md:63` "concluded values positive, researcher-denied values negative"; `comparison-layers.md:403` "A value the researcher denied". Code reads `Displayed()` / `ReasonDenied` (`autoreconcile.go:58`). Re-rated: a binding working rule, in a sentence this PR edited, now contradicts the code and the PR's own new paragraph; S11-12's planner reads it |
| MNT-1 | MNT | minor | Entity fixtures and plain-candidate helper written twice across the package's tests | Fix now → F-3 (rides along) | F-1 already touches `subjectprofile_test.go`; tests only; S11-12 adds rows to both tests (plan Out row) |
| MNT-2 | MNT | minor | `Entity` carries `NamePattern` as a field; `ForSubject` takes it positionally | Follow-up | D-15 and §6 fix the `ForSubject` shape; no callers outside tests, so S11-12 (which adds an argument to both entries per §6 Next adds) is the place to pick one input shape. Noted for the user |
| UX (note) | UX | — | `Build` returns `Reconcile`'s error unwrapped; it does not name the Property | Follow-up | D-7 decides unwrapped; matches `autoreconciler.go:326`, `truth.go:46`; fires only on a broken loader. S11-13 maps it to a stable code when Promote calls it |
| CON-3 | CON | nit | STEP-6 split into docs + landed-record commits | Logged | Behavior unaffected |
| MNT-4 | MNT | nit | `entityOf` lacks `t.Helper()` | Logged | `invariants_test.go:422` |

### Fix plan 1

**F-1 — AC-1.9 proves the `Name` pointer is shared** (from CON-1)
Files: `core/subjectprofile/subjectprofile_test.go`.
Failing test first: none possible — the code already shares the pointer; this strengthens a weak check. Confirm the new check fails if `value()` copies `Name` (temporarily, not committed), then restore.
Do: in the AC-1.9 row, build the entity once so the `check` can reach it, and assert `p[nameProp][0].Name == e.Properties[nameProp].Values[0].Value.Name` by identity (the way AC-5.2 does on the Subject side), keeping the existing `Form`, `NamePattern`, `Type`, `Ref`, negative and neutral checks.
Check: `go test ./core/subjectprofile/...` passes; scripts/check-pr.sh PASS.

**F-2 — Docs say displayed / denied, not concluded / researcher-denied** (from CON-2, MNT-3)
Files: `docs/deployment-plan/spike-11/deployment-plan.md` (line 63 working rule), `docs/deployment-plan/spike-11/comparison-layers.md` (§Entities profile table, ~line 398–403).
Do: working rule → "The profile is strict: displayed values positive, denied values negative, nothing else." Table: "The concluded value" → "A displayed value"; "A value the researcher denied" → "A value a stronger negative Observation denied (reason `denied`)". Match the wording of the PR's new "Built in S11-11" paragraph.
Check: `rg -n 'researcher-denied|researcher denied|concluded values positive' docs/deployment-plan/spike-11` prints nothing outside `prs/`; `scripts/check-plan-docs.py` 0 errors; scripts/check-pr.sh PASS.

**F-3 — One copy of the shared entity fixtures and candidate helpers** (from MNT-1)
Files: `core/subjectprofile/helpers_test.go`, `subjectprofile_test.go`, `entity_test.go`.
Do: move the `maleFemale` and `denied` entity closures from `TestBuild` and `TestEntityCompare` into `helpers_test.go` as `maleFemale(t)` and `denied(t)`; add one helper for an `autoreconcile.Candidate` from a value with a negative flag (replacing `pos` in `entity_test.go`, `c`/`neg` in `TestForSubject`, and AC-5.5's inline negative). No test row changes its inputs or expectations.
Check: test count in `core/subjectprofile` unchanged (`go test -v ./core/subjectprofile/... | grep -c '^    --- PASS'` before and after); scripts/check-pr.sh PASS.

## Round 1 — fix
Head `7c7b2c3a`. Gate PASS. F-1 done (`5ef3e8d5`; mutation of `value()` to copy `Name` made AC-1.9 and AC-5.2 fail, then restored); F-3 done (`2f2d993c`; subtest count 40 before and after); F-2 done (`7c7b2c3a`). Deviation: AC-1.9 captures its entity in a `named` variable so the `check` can reach it.

## Round 2 — review
Scope `c3904b60...7c7b2c3a`, CON and MNT only. CON: CON-1 and CON-2 resolved, no new findings, Clean. MNT: MNT-1 and MNT-3 resolved, no new findings, Clean. No Fix plan 2.

## Result
Gate PASS · CI see PR · open: none · follow-ups: MNT-2 (S11-12: one input shape for `Build`'s `Entity` and `ForSubject`), UX note (S11-13: stable code for `Build`'s reconcile error) · to user: —
