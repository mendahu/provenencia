# Ship log — S10-56 Object store stays inside objects/; deletes prune empty folders

**Plan:** [`S10-56-object-store-root.md`](S10-56-object-store-root.md) · **Branch:** `s10-56-object-store-root` · **PR:** #425
**State:** review 1

## Round 0 — implement
Base: branch `s10-56-object-store-root` cut from `s10-53-stage-image` @ `39ee4860` (#424, open); plan committed on it. PR targets `s10-53-stage-image` until #424 merges.
Head `6a7ab5db`. Gate PASS (orchestrator run). Report deviations:
- The prune cases are a separate `TestRemoveObjectPrunesEmptyShards` (the existing `TestRemoveObject` table's shape cannot assert folder removal); same cases as T-1.
- Plan Status written as "Landed in [#425](…) (2026-10-17)", the form `check-plan-docs.py` accepts.
- In-app manual check not run.
