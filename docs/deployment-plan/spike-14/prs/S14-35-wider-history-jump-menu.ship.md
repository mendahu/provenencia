# Ship log — S14-35 Wider history jump menu

**Plan:** [`S14-35-wider-history-jump-menu.md`](S14-35-wider-history-jump-menu.md) · **Branch:** `s14-35-wider-history-jump-menu` · **PR:** #730
**State:** ready

## Round 0 — implement
Base `f971f406`. §4 files unchanged since the planned base `f48846c`.
Head `f0bb09dc`. Gate PASS (orchestrator run; base `f971f406`). Report deviations: none. Implementer saw one `xcode-test` failure in `ThumbnailCacheTests`, a suite this PR does not touch, which passed on rerun unchanged. Manual check not run.

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/**`), SEC+PRF (`macos/App/Features/**` model; layout code reads a measured width with no bound). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1, MNT outside-lens | CON, MNT | major | Plan has no Landed note | Fix now → F-1 | verified: no `**Landed` line in the plan; DoD line asks for it |
| CON-2, UX-1, SEC outside-lens | CON, UX | minor | Back sits at x = 32, not the plan's 12; caps are 388pt and about 338pt at the minimum window, and Forward is narrower than the old 360pt | Fix now → F-2 | verified: `.padding(.horizontal, PVSpacing.gutterPage)` at `WorkspaceToolbar.swift:71`, `gutterPage` 32. The rule and D-2 stand and the right edge stays inside; D-5's rejected floor still brings back the overflow, so the Decision is not touched. Tests and Landed note take the real numbers; the user judges about 338pt at §10 step 3 |
| CON-3 | CON | minor | Right-clicking Back with omnibar results open puts the wider menu under the results | Fix now → F-3 | verified: `OmnibarResultsHost` monitor is `.leftMouseUp`, `.keyDown`; the results overlay is later in `WorkspaceContent.body`, so it paints on top. This PR makes the overlap larger |
| UX-2 | UX | minor | A truncated leaf cannot be read in full | Fix now → F-4 | minor in a file this PR changes; reviewer wrote Follow-up. A tooltip changes no label, truncation, line count or Out line. `.help` precedent `PVSelect.swift:402` |
| PRF note (MNT outside-lens) | PRF | minor | `contentWidth` re-evaluates `WorkspaceContent.body` per live-resize tick | Rejected | count: one `@State` write per tick; body builds 2 children and 2 overlays; `HistoryJumpMenuHost` exits early while closed; at most 15 rows while open. Independent of catalog size. `PVPageColumns.swift:180` already writes a size per tick on every page. SEC+PRF reviewer reached the same call |

Nits (logged): `MNT-1 · WorkspaceToolbar.swift:4 · doc comment still says "Shared open/anchor state"` · `MNT-2 · WorkspaceToolbar.swift:25 · group panelWidth with contentCoordinateSpace`

SEC+PRF: no findings. MNT: no findings above nit.

### Fix plan 1

**F-1 — Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-14/prs/S14-35-wider-history-jump-menu.md`.
Do: add a `**Landed:** [#730](https://github.com/mendahu/provenencia/pull/730), 2026-11-06.` line under the header, in the shape of S14-34's. List every difference from the plan: Back's anchor is x = 32 and Forward's about 82, so at the 436pt column the caps are 388pt and about 338pt, not AC-1.3's 408pt and D-5's about 376pt, and Forward is narrower there than the old 360pt; T-1 gained the rows from F-2; opening the menu closes omnibar results (F-3); the leaf has a tooltip (F-4). State that the §10 manual check is not run by the implementer. Do not edit D-5, AC-1.3 or any other Decision or AC.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-35-wider-history-jump-menu.md` reports no errors; scripts/check-pr.sh PASS.

**F-2 — T-1 uses the real anchors** (from CON-2, UX-1)
Files: `macos/ProvenenciaTests/HistoryJumpMenuModelTests.swift`.
Do: add rows `(436, 32) → 388` and `(436, 82) → 338`. Keep the five existing rows.
Check: `HistoryJumpMenuModelTests` passes.

**F-3 — Opening the jump menu closes omnibar results** (from CON-3)
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `historyButton`'s `menuPresented` setter.
Do: in the `if presented` branch call `omnibarResults.closePanel()` before `jumpMenu.present(side)`.
Check: scripts/check-pr.sh PASS. No unit test reaches a view binding; the PR body's manual check gains the step.

**F-4 — A tooltip reads the whole leaf** (from UX-2)
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `HistoryJumpMenuPanel` leaf `Text`.
Do: add `.help(Text(verbatim: leafLabel))` to the leaf `Text`, as `PVSelect.swift:402` does.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `047bb78e`. Gate PASS (orchestrator run). F-1 to F-4 done. F-2 adds rows and F-4 is a tooltip, so neither ran red first; F-3 has no unit seam, so the PR body's manual check gains the step.

## Round 2 — review
Reviewers: CON (CON-1 was a Fix-now major). Not sent back: UX. UX-1 and UX-2 were minor; read in the fix commits and **Resolved** (T-1 rows at `HistoryJumpMenuModelTests.swift`, `.help` on the leaf).
CON-1 **Resolved** (reviewer confirmed the Landed note and that no Decision or AC changed). No new blockers. Verdict: Clean.

## Result
Gate PASS at `047bb78e` · CI pending at this commit · open: none · follow-ups: none · to user: Forward's about 338pt at the minimum window, judged at §10 step 3. Manual check not run; it is the user's.
