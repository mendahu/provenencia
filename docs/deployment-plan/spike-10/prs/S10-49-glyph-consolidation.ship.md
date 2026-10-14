# Ship log — S10-49 Glyph consolidation

**Plan:** [`S10-49-glyph-consolidation.md`](S10-49-glyph-consolidation.md) · **Branch:** `s10-49-glyph-consolidation` · **PR:** #394
**State:** ready

## Before start

`main` @ `e150a76a`, the planning base. The branch already exists. Its first commit holds the plan, the review log, and the map entry, because the plan was not on `main` (review X-6). The implementer continues on this branch.

## Round 0 — implement
Head `2e3637a3`. PR #394 (draft). Gate PASS: the orchestrator ran `scripts/check-pr.sh` against `e150a76a`. Report deviations:
- AC-2.4's `rg` still matches the four bridge raw values in `enum PVMarkKey` (`PVMark.swift` 61–64). The plan's pattern was wrong: it matches the case definitions themselves. No other site matches. This is a plan defect, not a code defect.
- `ConclusionDetailPage.swift:232` built `.init(mark:label:)`, a caller §4 missed. It was moved to `glyph:` so the build passes.
- The `pbxproj` IDs for `PVGlyphTests.swift` are `…09E0` / `…09E1`. The implementer first tried `03C0`, which collided with `PVContextMenu.swift`.
- The S10-49 plan's Context and STEP-9 Files lines were repointed to `ideas/archive/` so `check-plan-docs` passes, the same as S10-19 did for its briefs.
- The manual check is not run. It is for the reviewer.

## Round 1 — review

All five lenses were dispatched against `e150a76a...2e3637a3`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | minor | `PVThumbnail` decides its glyph twice (`state` then `if case` re-match) | Follow-up | Verified at `PVThumbnail.swift` lines 57–65 and 83–91. D-4 kept the private `State` cases. No fix round touches the file. |
| CON-1 | CON | nit | AC-2.4's `rg` prints the 4 `PVMarkKey` raw-value lines | Logged | Plan defect, already in the Landed note and in the round-0 deviations |
| MNT-2 | MNT | nit | Unreachable `?? .fallback` on the omnibar person/event/place lead | Logged | Matches D-6 |
| MNT-3 | MNT | nit | Archived idea body still reads in the present tense | Logged | STEP-9 asked only for the title, status, links, and leftovers |
| CON-2 | CON | nit | Blank-line placement around the S10-49 map table | Logged | Renders correctly |
| UX | UX | — | No findings | — | — |
| SEC | SEC | — | No findings | — | — |
| PRF | PRF | — | No findings | — | — |

No Fix-now finding, so there is no Fix plan 1 and no round 2.

## Result
Gate PASS at `2e3637a3`. CI: see the PR. Open: none. Follow-ups: MNT-1 (`PVThumbnail` private `State` with associated values). To user: the manual check in plan §10.
