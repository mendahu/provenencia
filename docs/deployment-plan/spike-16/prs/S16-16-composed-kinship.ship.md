# Ship log — S16-16 Kinship answers are composed from the pivots, and the Family read is one pass

**Plan:** [`S16-16-composed-kinship.md`](S16-16-composed-kinship.md) · **Branch:** `s16-16-composed-kinship` · **PR:** #TBD
**State:** implementing (resumed)

## Round 0 — implement
Base `main` @ `aa150de0f`. Dispatched `pr-implementer`, round 0.
Local head `e883317d4` (2 commits, not pushed). Gate not run.

**Stopped on** §11: `warm-ms` at 16,000 is 231.6 vs base 137.9 (1.68×, limit 1.25×). Same M1, load 7–9; a re-run still slow.

| Measure | Base | After | Ratio |
| --- | --- | --- | --- |
| warm-ms 4,000 | 115.9 | 180.3 | 1.56 |
| warm-ms 16,000 | 137.9 | 231.6 | 1.68 |
| after-commit-ms 4,000 | 621.5 | 825.9 | 1.33 |
| after-commit-ms 16,000 | 1399 | 1289 | 0.92 |
| TestLargeSourceBudgets warm log | 131.8 ms | 258 / 199 ms | 1.5–1.96 |
| kinship-B 4,000 / 16,000 | 1,345,832 / 4,534,264 | 0 / 0 (HeapDelta artefact, needs re-run) | — |

Budgets still met. Cause as §6 Scale predicted: each warm proposal recomposes `For` per walked person. Done: STEP-1–3, part of STEP-5. Not done: docs rows, map, completed.md, gate, push, PR.
Deviation: `warmProposal` uses `databasetest.Warm` (main already had it), same behaviour.

**State:** stopped — memoising `For` is the user's call (plan §11, rule "No stored answer without a measurement").

**User, 2026-11-13:** accept the warm slowdown as the new normal. Plan §3, STEP-4 Check and §11 amended; memoising `For` parked in `docs/ideas/promote-ranking.md` § "Memoise a person's composed kinship". Raise `budgetWarm` to match: CI measured 383 ms before (S16-13) × the worst ratio seen (1.96) = 751 ms, × 1.25 headroom → 950 ms. Other budgets unchanged; a red CI on one of them is a gate failure with the CI figure as evidence.

## Round 0 — resume
Dispatched `pr-implementer` from STEP-4 (memory re-measure) onward.
