# Ship log — S11-41 One link reading for bridges

**Plan:** [`S11-41-link-reading.md`](S11-41-link-reading.md) · **Branch:** `s11-41-link-reading` · **PR:** #561
**State:** ready

## Round 0 — implement
Head `aa22b779`. Gate PASS (orchestrator run). Report: deviations none; base moved `74bb416` → `660111c5` (plan merge only), Current state held.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes). SEC+PRF runs on the `core/database/**` performance path (`core/database/promotealign`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | — | — | No blocker, major, or minor findings from CON, MNT, or SEC+PRF | — | — |

Nits (logged): `CON-1 / MNT-1 · core/database/promotealign/canon.go:208, layer.go:184 · edited comments wrap segmentcompare.Keys in Markdown backticks and run past the comment width` · `CON-2 · core/segmentcompare/segmentcompare_test.go · TestLinkOf relationship rows pass EventBirth to show it is dropped, but no row name says so` · `MNT-2 · core/segmentcompare/key.go:49 · LinkOf exported with Keys as its only non-test caller; D-8 accepts it, revisit by S11-39`

No fix plan: nothing to fix. No round 2: no Fix-now blocker or major.

## Finish
Merged `origin/main` @ `bc673dd6` (S12-13 plan docs only). Gate PASS on the merged head.

## Result
Gate PASS · CI <pending> · open: none · follow-ups: none · to user: — · CON miss: none
