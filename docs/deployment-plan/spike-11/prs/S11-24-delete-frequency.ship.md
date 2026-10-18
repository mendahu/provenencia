# Ship log — S11-24 Delete frequency weighting

**Plan:** [`S11-24-delete-frequency.md`](S11-24-delete-frequency.md) · **Branch:** `s11-24-delete-frequency` · **PR:** #483
**State:** ready

## Round 0 — implement
Base `23cdea1`. Head `ca8df05`. Gate PASS (run by orchestrator). Report: deviations none; 4 commits (361315d, ea458dc, 2abcd81, ca8df05).

## Round 1 — review
CON, SEC, PRF: clean, no findings.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | minor | Promote notice still says "weighted by how rare each value is" | Rejected | `macos/App/Features/Sources/Promote/` does not exist; `git grep -i 'rare\|frequen'` over `macos/App` and `*.xcstrings` finds no such copy |
| MNT-1 | MNT | minor | `PropertyScale` map keys are string literals while lookups use `vocabkey` constants | Follow-up | verified `core/graphalign/registry.go:58-61`; no fix plan touches the file, and S11-13 replaces this map |
| MNT-2 | MNT | nit | New comments are single long lines among ~80-column neighbors | Logged | — |

No blocker or major; no fix plan, no round 2.

## Result
Gate PASS · CI see PR · open: none · follow-ups: MNT-1 (S11-13) · to user: —
