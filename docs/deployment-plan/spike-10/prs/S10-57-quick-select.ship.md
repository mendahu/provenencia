# Ship log — S10-57 Hold Shift to select; Select disarms after one rectangle

**Plan:** [`S10-57-quick-select.md`](S10-57-quick-select.md) · **Branch:** `s10-57-quick-select` · **PR:** #428
**State:** review 1

## Round 0 — implement
Base: branch `s10-57-quick-select` cut from `main` @ `af50bd67`. The plan merged as S10-56 in #426, which reused the object store's number (#425). It is renumbered to S10-57 on this branch: files renamed, map row and scope entry updated (user, 2026-10-17).
Head `8b071471`. Gate PASS (orchestrator run). Report deviations:
- Code in two commits, not one per step; `shiftEngagesSelectWithoutArmingIt` lands in `ee4049b1` before the `inputMode` change in `2aaec3f3`, so that intermediate commit is not green on its own. HEAD is green.
- STEP-1's interim `wirePointer` closure skipped; final closures wired directly.
- AC-1.5, AC-1.7, the omnibar half of AC-1.6, and §10 (the in-app evidence that the `flagsChanged` monitor fires over the canvas) not run: they need a person at the app.
