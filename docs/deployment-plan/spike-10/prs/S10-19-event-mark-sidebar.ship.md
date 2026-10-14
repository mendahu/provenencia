# Ship log — S10-19 Event mark and sidebar icons

**Plan:** [`S10-19-event-mark-sidebar.md`](S10-19-event-mark-sidebar.md) · **Branch:** `s10-19-event-mark-sidebar` · **PR:** #393
**State:** ready

## Before start

`main` @ `c13e3d9d`. The plan was made at `8b742aec`. Since then S10-40 (#392) changed `ConclusionListPage.swift`, but only below the `.empty` case, by adding the scroll-anchor modifier. Every §4 row still holds: `emptyIcon` is at line 46 and the empty state at line 144, the two `section.icon` readers are unchanged, and `rg "PVEmptyState\(" macos/App` still counts 15.

## Round 0 — implement
Head `6b83344e`. Gate PASS: the orchestrator ran `scripts/check-pr.sh` against `c13e3d9d`. Report deviations:
- STEP-8's Files line named the briefs at their pre-archive paths, so `check-plan-docs` failed after the move. The implementer fixed that one plan line in a separate commit and recorded it in the Landed note.
- The STEP-8 commit was pushed before the gate result was seen. It was fixed forward, with no force-push.
- The manual check is not run. It is listed in the PR body for the reviewer.

## Round 1 — review

All five lenses were dispatched against `c13e3d9d...6b83344e`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | PR body says "Deviations: None" but the Landed note records one | Fixed by the orchestrator in the PR body | The orchestrator owns the PR body (ship-pr §8). Body Deviations now copies the Landed note. |
| MNT-1 + CON-2 | MNT, CON | minor | T-2's loop restates the production `switch` for the Conclude cases | Follow-up | Verified: the loop's Conclude arms equal `WorkspaceSection.glyph`, and the three literal `#expect`s after it pin the keys. The plan's T-2 asked for this shape. No fix round touches the file. |
| MNT-2 | MNT | nit | `PVEmptyState` header comment line no longer wraps at about 80 | Logged | — |
| UX | UX | — | No findings | — | — |
| SEC | SEC | — | No findings | — | — |
| PRF | PRF | — | No findings | — | — |

No Fix-now finding, so there is no Fix plan 1 and no round 2.

## Result
Gate PASS at `6b83344e`. CI: see the PR. Open: none. Follow-ups: MNT-1 / CON-2 (trim T-2 to its literal assertions). To user: the manual check in plan §10.
