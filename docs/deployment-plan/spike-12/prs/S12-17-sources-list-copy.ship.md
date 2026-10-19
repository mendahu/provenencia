# Ship log — S12-17 Copy Source from the Sources list

**Plan:** [`S12-17-sources-list-copy.md`](S12-17-sources-list-copy.md) · **Branch:** `s12-17-sources-list-copy` · **PR:** #488
**State:** ready

## Before round 0
The plan was refreshed on the branch first (`fc963b7e`): S12-02 had landed (#485), and board S12-D6 is agreed. The user asked on 2026-10-18 for the kit menu modifier to be built in this PR, so the plan now adds `PVActionMenuItem`, `PVActionMenuRows` and `pvRightClickMenu`, and moves the cover, portrait and Clear menus onto them. Refresh review: no blocker.

## Round 0 — implement
Head `ad0968db`. Gate PASS (orchestrator run). Report: no deviations. Note: the Clear region row now gets `index: nil` when disabled, as D-2's `PVActionMenuRows` specifies (before, it kept index 0). Manual checks §10 1–11 are for the user.

## Round 1 — review
Base `6ecfd88f`, head `ad0968db`. Reviewers: CON (sonnet), MNT, UX, SEC+PRF. Skipped: none. Security has no path match (no core, FFI, migrations or URL code); the combined reviewer runs because `macos/App/Features/Sources/SourcesModel.swift` is a Features model (performance).

Verdicts: CON changes needed · MNT clean · UX clean · SEC+PRF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 (+ MNT outside-lens) | CON | major | The plan has no **Landed** note | Fix now → F-1 | verified: `rg Landed` on the plan finds only the Status line; the DoD and STEP-6 require the note (S12-02 L4 shape) |
| MNT-1 | MNT | minor | README says every right-click menu is `.pvRightClickMenu`; the toolbar's Back/Forward history menu is a separate anchored right-click path | Fix now → F-2 (rides along: a line this PR adds, in a docs fix round) | verified: `README.md` L161; `WorkspaceToolbar.swift` L511–530 `PVRightClickCatcher` + `PVContextMenuPanel` on `HistoryJumpMenuModel` |
| UX-1 | UX | minor | Source cover is read as "No preview" while it shows an image, and now carries a VoiceOver action | Follow-up | predates this PR (`coverThumbnail` passes no `imageLabel`; `CachedThumbnail.swift` L58–63); the fix needs a new string, against D-8 |
| MNT (outside lens → UX) | UX | minor | Type-to-select can highlight the disabled Clear region row | Follow-up | unchanged from before this PR: `keyboardTitles` keeps every slot, as the old hand-written array did |
| MNT-2 | MNT | nit | `pvContextMenu` doc says `keyboardTitles` are "of a right-click menu" | Fix now → F-2 (same docs commit, one word) | `PVContextMenu.swift` L277 |
| MNT-3 | MNT | nit | README says "one VoiceOver action per item"; only enabled items get one | Fix now → F-2 | `README.md` L161 |

Nits (logged): `CON-2 · PVContextMenuTests.swift:218 · no all-enabled / mixed index case for PVActionMenuRows` · `UX-2 · Localizable.xcstrings:5382 · sources.copy.entry comment still says "Source page action"; now also a row menu item and VoiceOver action`

### Fix plan 1
**F-1 — Add the plan's Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md`.
Failing test first: none — docs.
Do: add a `**Landed:**` line right under **Status**, in the S12-02 plan's shape (its L4). Say: no code differences from the plan; one behavior note: the disabled Clear region row now has `index: nil` (as D-2's `PVActionMenuRows` specifies; before it kept index 0, which no code used for a disabled row); manual check §10 steps 1–11 are left to the user.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-17-sources-list-copy.md` reports no errors; scripts/check-pr.sh PASS.

**F-2 — Make the kit menu docs exact** (from MNT-1, MNT-2, MNT-3)
Files: `macos/App/DesignSystem/README.md` (ContextMenu row, L161), `macos/App/DesignSystem/Components/ContextMenu/PVContextMenu.swift` (doc comment, L277; comment only).
Do: in the README row, "a right-click menu is one `.pvRightClickMenu`" → "a right-click menu that opens at the pointer is one `.pvRightClickMenu`", and "one VoiceOver action per item" → "one VoiceOver action per enabled item"; append "(the toolbar's Back/Forward history menu anchors to its button and keeps its own presenter)". In `PVContextMenu.swift`, "the row titles of a right-click menu" → "the row titles of an action menu". No code change.
Check: scripts/check-pr.sh PASS.

Commit F-1 and F-2 as one docs commit: "Record the S12-17 Landed note and make the kit menu docs exact."

## Round 1 — fix
Head `9aac986c`. Gate PASS (orchestrator run). F-1 done; F-2 done. Orchestrator read F-2: MNT-1 **Resolved** (README row says "opens at the pointer" and names the history menu), MNT-2 **Resolved** (`PVContextMenu.swift` doc says "an action menu"), MNT-3 **Resolved** ("per enabled item").

## Round 2 — review
Reviewers: CON only (the only lens with a Fix-now major). Scope `ad0968db...9aac986c`.
Verdict: CON clean. CON-1 **Resolved**. No new findings.

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-1 (cover read as "No preview"), Clear-menu type-select on a disabled row · to user: — · CON miss: none
