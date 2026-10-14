# Ship log — S10-20 Open a file and box-select cards

**Plan:** [`S10-20-open-file-box-select.md`](S10-20-open-file-box-select.md) · **Branch:** `s10-20-open-file-box-select` · **PR:** #396
**State:** ready

## Round 0 — implement
Head `1984251c`. Gate PASS (`scripts/check-pr.sh` against `09c5b8d4`). Draft PR #396.

Deviations reported: `closedHelp` uses `displayFilename`; `openSelectedFile` stores the resolved id so the missing-file callout can show; arrow keys call `moveSubjects` directly because it is not async; `snapAndPatch` returns a non-optional tuple; the marquee is a child of the base layer so cards paint above it; Select strings landed with the header strings; manual check §10 was not run.

## Round 1 — review

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1, PRF-1 | SEC, PRF | major | Menu thumbnails decode the original file when there is no derivative | Fix now → F-4 | User agreed 2026-10-14: fall back to the glyph, not the original file. Differs from D-4's openable-path fallback. Record it in the Landed note. Do not edit D-4. |
| CON-1 | CON | major | Missing-file branch of Open is untested | Fix now → F-1 | Verified: no test calls `openSelectedFile`. `ProjectFiles.openObject` returns false when the file is absent, before `NSWorkspace`. |
| UX-1 | UX | major | Multi-select has no keyboard or VoiceOver path | Follow-up | §3 Out: Shift-click add is not planned. One card at a time still moves from the keyboard. |
| CON-2 | CON | minor | Pan test does not prove the selection stays | Fix now → F-2 | The pan branch was re-routed in this PR. A few lines in the test the PR already added. |
| UX-2 | UX | minor | A one-file Source cannot dismiss the missing-file callout, so the armed banner stays hidden | Fix now → F-3 | D-6 hides the banner while the callout shows. Clearing the error when Place, Connect, or Select arms keeps that idle placement and gives the one-file case a way out. |
| MNT-1 | MNT | minor | Four bump-and-patch sites | Follow-up | More than a few lines. Behavior matches D-11 and D-12. |
| MNT-2 | MNT | minor | Resolved file row looked up three times | Follow-up | Works. A `resolvedFileRow` helper can wait. |
| MNT-3 | MNT | minor | `commitDrag` and `moveSubject` have no production caller | Follow-up | D-11 and D-12 keep the one-card APIs. |
| MNT-4 | MNT | minor | Callout depends on two stored ids matching | Follow-up | The Landed note records why `openSelectedFile` stores the id. CON-1 locks it. |
| MNT-5 | MNT | minor | Third armed-tool flag | Follow-up | D-13 and Next adds defer `armedTool`. |
| MNT-6 | MNT | minor | Cover and openable-file copies remain | Follow-up | Next adds already names this. |
| UX-3 | UX | minor | "File not found" also covers "no app can open this" | Follow-up | `ProjectFiles.openObject` is shared with the Source page and is not in this PR's files. |
| UX-4 | UX | minor | Menu can read "1 citations" | Follow-up | §3 Out: do not add a plural to the shared citation key. |
| UX-5 | UX | minor | A long label has no full-text help | Follow-up | AC-1.6 makes help the filename. Medium confidence. |
| SEC-2 | SEC | minor | `objectURL` does not follow symlinks | Follow-up | Pre-existing. Same guard as the Source page. |
| SEC-3 | SEC | minor | Arrow move can overflow `Int64` | Follow-up | Same arithmetic as the existing one-card move. |
| PRF-2 | PRF | minor | `fileRows` rebuilt several times per header body | Follow-up | Under a frame. Medium confidence. |
| PRF-3 | PRF | minor | Group move is one write per card | Follow-up | D-11 rejected one store transaction for the group. |
| MNT-7 | MNT | nit | `PaletteHelp` for one tooltip | Logged | D-14 names the help parameter. |
| MNT-8 | MNT | nit | Redundant graph-handle check | Logged | |
| UX-6 | UX | nit | Select's help names Shift, and Select has no banner | Logged | EG-4 and D-13. |

### Fix plan 1

**F-1 — Lock the missing-file Open branch** (from CON-1)
Files: `macos/ProvenenciaTests/EvidenceGraphModelTests.swift`.
Do: Add `openSelectedFileReportsAMissingFileAndKeepsTheSelection`. It is expected to pass on the current `openSelectedFile`: a workspace whose object is not on disk, `openSelectedFile()` sets `fileOpenError` to the Page 7 sentence and `selectedFileArtifactID` to that artifact, and `selectFile` of the other artifact clears the error. Commit the test. Do not add a file-opener seam. Do not change production code unless the test fails.
Check: the test passes.

**F-2 — Pan does not change the selection** (from CON-2)
Files: `macos/ProvenenciaTests/GraphCanvasPointerHitTestingTests.swift`.
Do: On the Shift-up empty drag in `shiftDragOnEmptyCanvasMarquees`, set `onSelect` and `onDeselect` and expect neither runs.
Check: `GraphCanvasPointerPressTests` passes.

**F-3 — Arming a tool clears the missing-file callout** (from UX-2)
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`.
Do: When `toggleArm` or `toggleConnect` arms, and when `toggleSelect` arms, set `fileOpenError = nil`. Leave the callout as the `else if` after `deleteError`.
Check: a model test that arms Place after a missing-file error sees `fileOpenError == nil`.

**F-4 — Menu thumbnails use the glyph when there is no derivative** (from SEC-1, PRF-1)
Files: `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `EvidenceGraphFileMenuRow`, the plan's **Landed** note.
Do: Pass `relPath: row.thumbnailRelPath` only. `CachedThumbnail` already shows the MIME glyph when that path is empty. Do not pass `row.relPath`. Append one sentence to the Landed note: the menu does not open the original object when `thumbnailRelPath` is empty; the user chose the glyph over D-4's openable-path fallback. Do not edit D-4.
Check: `scripts/check-pr.sh` PASS.

## Round 1 — fix
Head `dd4d5cfe`. Gate PASS. F-1, F-2, F-3, and F-4 done.

## Round 2 — review
Lenses: conformance, user experience, security, performance. Scope `1984251c...dd4d5cfe`. Maintainability had no Fix-now finding, so it did not return.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | CON, UX, SEC, PRF | — | No new blockers | — | CON-1, CON-2, UX-2, SEC-1, and PRF-1 are resolved. |

No Fix plan 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-1, MNT-1, MNT-2, MNT-3, MNT-5, MNT-6, UX-3, UX-4, UX-5, SEC-2, SEC-3, PRF-2, PRF-3 · to user: —
