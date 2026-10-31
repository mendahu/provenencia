# Ship log — S13-25 The pin table goes

**Plan:** [`S13-25-the-pin-table-goes.md`](S13-25-the-pin-table-goes.md) · **Branch:** `s13-25-the-pin-table-goes` · **PR:** #644
**State:** ready

## Round 0 — implement

Base `3f6af69f` (`main`). Plan planned against `c54f34bf`. Dispatched `pr-implementer`.
Head `795535fd`. Gate PASS (orchestrator re-ran `scripts/check-pr.sh`). PR #644 (draft). Report: deviations — three conclusion-layer lines the plan did not name were corrected because the removal made them false (§2, §4, invariant 20); conclusion §7 `reconciliation_claims.confidence_grade_id` DDL lost its `REFERENCES` clause; prose sweeps keep D-9's named hits; manual check and upgrade run not done (plan §10); two earlier xcode-test runs hit `waitUntil` timeouts in untouched suites, passed on rerun.

## Round 1 — review

Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`, `Localizable.xcstrings`, `core/apperr`. SEC+PRF: migration `000058.sql`, `core/database/**`.

UX first returned without reading the diff; resumed the same round-1 reviewer to finish (not a new round).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `rebuiltSnapshot` has no caller left; its orphaned comment sits above `type outcome` | Fix now → F-1 | verified: `rg -n rebuiltSnapshot core` → only `autoreconciler_test.go:333` (comment) and `:368` (definition); the one base caller was in the block this PR replaced. Same-pass cleanup: a copy of the rebuild check the new `autoreconcilertest.AssertCacheMatchesRebuild` owns. Supersedes D-7's "`rebuiltSnapshot` … stay (other tests read them)", whose premise is false; `outcomesSnapshot` stays (used) |

CON: clean. UX: clean. SEC+PRF: clean.

Nits (logged): `CON-1 · docs/conclusion-layer-data-model.md:332-333 · "Observation pins" in the badge table means Reconciliation pins; could say so` · `MNT-2 · core/database/promote/unpromote.go:20 · rewrap the long doc-comment line`

### Fix plan 1

**F-1 — Delete the dead `rebuiltSnapshot`** (from MNT-1)
Files: `core/database/autoreconciler/autoreconciler_test.go`.
Failing test first: none (dead-code removal; the compiler and the existing suite are the check).
Do: delete `func (f *fixture) rebuiltSnapshot() []row` (line 368) and its two-line comment at lines 333-334 ("rebuiltSnapshot rebuilds inside a rolled-back transaction, so upkeep state / is left as it was."). Keep `snapshot`, `f.rows()`, `outcomesSnapshot` and `type outcome`. Remove any import the deletion leaves unused. Add one sentence to the plan's Landed note: F-1 superseded D-7's keep of `rebuiltSnapshot` (no caller remained).
Check: `rg -n rebuiltSnapshot core` prints nothing; `go test -tags fts5 ./core/database/autoreconciler/...` passes; scripts/check-pr.sh PASS.

## Round 1 — fix

Head `d19176ed`. Gate PASS (orchestrator re-ran). F-1 done.

## Round 2 — review

Reviewers: MNT (had a Fix-now major). Scope: `795535fd...d19176ed`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) dead `rebuiltSnapshot` | Resolved | `rg -n rebuiltSnapshot core` empty; comment now above `type outcome`; package vets and tests clean |

No new findings.

## Finish

`origin/main` unchanged since the branch was cut (`3f6af69f`); no merge needed. Gate PASS at `d19176ed`.

## Result
Gate PASS · CI green · open: none · follow-ups: none · to user: — · CON miss: none
