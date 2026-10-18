# Ship log — S10-66 Pure depiction value package

**Plan:** [`S10-66-pure-depiction-value.md`](S10-66-pure-depiction-value.md) · **Branch:** `s10-66-pure-depiction-value` · **PR:** #473
**State:** ready

Worktree `/Users/jakerobins/Repos/provenencia-s10-66`, cut from `origin/main` @ `ab70bd69`.

## Round 0 — implement
Head `c34646c0`. Gate PASS (run by the orchestrator). Report deviations: base was `ab70bd69`, not the planned `e97cf54` (no Current state row invalidated; docs edited by text, not line); `TestDepictionValues` takes its page-only `region` from `depictionvaluetest.Locator`; imports sorted with `gofmt -w`. Both are in the plan's Landed note.

## Round 1 — review
Five lenses on `ab70bd69...c34646c0`. PERF, UX, SEC: no findings.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1, CON-1 | MNT, CON | minor | `TestDepictionValues` stores the `depictionvaluetest.Locator` fixture under the name `region` | Fix now → F-1 (user, 2026-10-18) | verified: `depictionvalues_test.go:39` is `const region = depictionvaluetest.Locator`; `Region` exists beside it and passes `Validate` equally. Behavior unchanged (same literal as before), deviation already in the Landed note. Minor, and no fix plan touches the file, so not fixed now; user chose to fix it before merge |
| MNT-2 | MNT | nit | `resolveDepiction` trims `RegionJSON` again after `InsertTx` trims it | Logged | pre-existing; the plan's D-5 keeps it; code this PR does not change |

No blocker or major. After the first Result the user asked for MNT-1 to be fixed before merge, so Fix plan 1 has one item.

### Fix plan 1
**F-1 — `TestDepictionValues` uses the region fixture as its region** (from MNT-1, CON-1)
Files: `core/database/depictionvalues/depictionvalues_test.go`. Skill: —
Failing test first: none. The test passes before and after; this is a naming and fixture fix, so the check is that the test still passes.
Do: change line 39 `const region = depictionvaluetest.Locator` to `const region = depictionvaluetest.Region`. Change nothing else. Do not edit the plan's Landed note; the orchestrator records the outcome here.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/depictionvalues/` passes; scripts/check-pr.sh PASS. Commit: "Use the depiction region fixture as the region in TestDepictionValues."

## Round 1 — fix
Head `277c5825`. Gate PASS (run by the orchestrator; `go test -count=1 ./core/database/depictionvalues/` ok). F-1 done: one line, `const region = depictionvaluetest.Region`. Plan Landed note updated to match.

## Round 2 — review
MNT, CON on `0a2427eb...277c5825`. MNT-1/CON-1 resolved; no new findings from either lens.

## Result
Gate PASS · CI see PR · open: none · follow-ups: MNT-2 (one trim on the depiction write path) · to user: —
