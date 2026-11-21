# Ship log — S16-53 The history menu uses the kit presenter and a kit menu row

**Plan:** [`S16-53-history-menu-kit-row.md`](S16-53-history-menu-kit-row.md) · **Branch:** `s16-53-history-menu-kit-row` · **PR:** #937
**State:** ready

## Round 0 — implement
Head `1d2836750`. Gate PASS (orchestrator run). PR #937 (draft). Report deviations: STEP-3's `rg "activeIndex"` keeps one hit, the `activeIndex: -1` D-6 requires in the keyboard seed; `check-plan-docs.py` on the plan flags two mentions of the deleted `HistoryJumpMenuModelTests.swift` (gate still passes); commits are per behavior. AC-1.1, AC-1.7, AC-2.1 – AC-2.7 and §10 are in-app checks, owed to the user's manual check.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`. SEC+PRF: no security path; dispatched for performance because the diff replaces a `macos/App/Features/**` model and the menu loops over the history list (unsure → dispatch).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | The idea file still names `rows(for:navigation:names:)` | Fix now → F-1 | verified: `docs/ideas/history-menu-labels.md:14`; head declares `HistoryJumpMenuHost.rows(_:names:)` |
| CON-2 | CON | minor | A Landed note bullet describes a check failure that does not happen | Fix now → F-2 | verified: `check-plan-docs.py` on the plan prints 0 errors, 0 warnings; plan line 309 says it flags |
| MNT-1 | MNT | minor | Two `contentWidth` comments are now wrong | Fix now → F-3 | verified: `WorkspaceToolbar.swift:17` ("omnibar list's width rule" only), `WorkspaceContent.swift:16` ("placing overlays only"; the overlay is gone) |
| MNT-2 | MNT | minor | `import AppKit` in `WorkspaceToolbar.swift` is unused | Fix now → F-4 | verified: `rg "NS[A-Z]"` in the file at head finds nothing |
| SEC/PRF | SEC+PRF | — | No findings | — | history rows capped at 15; no trust boundary touched |
| UX-1 | UX | minor | The history menu stays open after a click in a page's right-click menu | To user | AC-2.7 asks the menu to stay open on a page right-click; the later click lands in a popup-level window, which every kit popup counts as inside. Closing it there is a product call against AC-2.7, not a defect in the plan's terms. Minor, so it does not hold the PR; asked at finish |
| UX-2 | UX | minor | The leaf's tooltip now lives in a non-key panel; nothing shows it still appears | Fix now (manual check) | low confidence; covered by D-4's §11 stop condition. Added to the manual check in the PR body: hover a long-named row and check the tooltip appears |

Nits (logged): `MNT-3 · WorkspaceToolbar.swift:4 · HistoryJumpMenuSide can be private, drop explicit Hashable` · `MNT-4 · WorkspaceToolbar.swift:329 · rows(_:names:) can stay private static` · `MNT-5 · PVPopupTests.swift:116 · rightClicksPassWhenIgnored as table rows` · `UX-3 · HistoryJumpMenuHost.menu · 1×1 panel possible while anchor is nil; gate on anchor != nil` · `CON-3 · commit 525b59aec · subject hides the menu move`

### Fix plan 1
**F-1 — The idea file names the renamed row builder** (from CON-1)
Files: `docs/ideas/history-menu-labels.md`.
Do: line 14, replace `` `rows(for:navigation:names:)` `` with `` `HistoryJumpMenuHost.rows(_:names:)` ``.
Check: `rg -n "rows\(for:navigation" docs .agents .cursor macos` finds nothing outside `docs/deployment-plan/`.

**F-2 — The Landed note says what is true about the deleted test file** (from CON-2)
Files: `docs/deployment-plan/spike-16/prs/S16-53-history-menu-kit-row.md` (Landed note only).
Do: reword the bullet "This plan still names `HistoryJumpMenuModelTests.swift` … so `check-plan-docs.py` flags those lines." to: "This plan still names the deleted `HistoryJumpMenuModelTests.swift` in STEP-4 and Deletes; `check-plan-docs.py` accepts it."
Check: `python3 scripts/check-plan-docs.py <plan>` reports 0 errors.

**F-3 — The `contentWidth` comments match their uses** (from MNT-1)
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift:17`, `macos/App/Features/Workspace/WorkspaceContent.swift:16`.
Do: toolbar: "The content column's width, for the omnibar list's and the history menu's width rules." Content: "The content column's width, for the toolbar popups' width rules. Pages read \.pvWidthClass."
Check: scripts/check-pr.sh PASS.

**F-4 — Drop the unused AppKit import** (from MNT-2)
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift:1`.
Do: delete `import AppKit` (SwiftUI stays).
Check: the app builds; scripts/check-pr.sh PASS.

One commit for F-3 and F-4 ("Correct the toolbar's width comments and drop an unused import."), one for F-1 and F-2 ("Point the history label idea at the renamed row builder.").

## Round 1 — fix
Head `b7556f1e6`. Gate PASS (orchestrator run). F-1 – F-4 done, no deviations. MNT-1, MNT-2: **Resolved** (orchestrator read `1d2836750..b7556f1e6`: both comments name their uses, `import AppKit` gone). CON-2: **Resolved** (Landed note reworded; `check-plan-docs.py` 0 errors).

## Round 2 — review
Reviewers: CON (CON-1 was a Fix-now major). Not sent: MNT, UX (minors only, read above), SEC+PRF (no findings).
CON-1: **Resolved** (`history-menu-labels.md:14` names `HistoryJumpMenuHost.rows(_:names:)`; `rg "rows\(for:navigation"` outside the deployment plan finds nothing). New findings: none. No Fix plan 2.

## Result
Gate PASS (head `b7556f1e6`, `origin/main` unchanged since the branch was cut) · CI see PR · open: none · follow-ups: none · to user: UX-1 (close the history menu after a click in another popup, against AC-2.7) · manual check owed: plan §10 plus the UX-2 tooltip hover.
