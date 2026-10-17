# Ship log — S10-57 Hold Shift to select; Select disarms after one rectangle

**Plan:** [`S10-57-quick-select.md`](S10-57-quick-select.md) · **Branch:** `s10-57-quick-select` · **PR:** #428
**State:** review 2

## Round 0 — implement
Base: branch `s10-57-quick-select` cut from `main` @ `af50bd67`. The plan merged as S10-56 in #426, which reused the object store's number (#425). It is renumbered to S10-57 on this branch: files renamed, map row and scope entry updated (user, 2026-10-17).
Head `8b071471`. Gate PASS (orchestrator run). Report deviations:
- Code in two commits, not one per step; `shiftEngagesSelectWithoutArmingIt` lands in `ee4049b1` before the `inputMode` change in `2aaec3f3`, so that intermediate commit is not green on its own. HEAD is green.
- STEP-1's interim `wirePointer` closure skipped; final closures wired directly.
- AC-1.5, AC-1.7, the omnibar half of AC-1.6, and §10 (the in-app evidence that the `flagsChanged` monitor fires over the canvas) not run: they need a person at the app.

## Round 1 — review
Lenses: all five. SEC, PERF clean; CON one major; MNT, UX minors.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | No Landed note; PR body says "Deviations: None" though the plan was renumbered and §10 not run | Fix now → F-1 | verified: no `## Landed` in the plan; PR body checked at finish |
| UX-3 | UX | minor | The plan's Out line sends "keyboard / VoiceOver way to add one card" to the dogfood log; no entry was written | Fix now → F-1 (rides along) | verified: `docs/dogfood/ux.md` has no entry; same docs round, a few lines |
| CON-2, UX-4 | CON, UX | minor | Shift held makes `inputMode` `.selecting`, so Shift+Return / Shift+Space on a selected card do nothing (before: edit / activate) | Follow-up; F-1 records it in the Landed note | verified: `EvidenceGraphModel.swift:499` guards `inputMode == .idle`; plan silent on keys under Shift; minor in a file the fix plan does not touch |
| UX-2 | UX | minor | Shift-press on a card action still paints its pressed state, but the action never runs | Follow-up | minor; `GraphCanvasPointerController.mouseDown` is not in the fix plan |
| UX-1 | UX | minor | Select lights for any Shift in the window (Shift-Tab, Cmd-Shift shortcuts), not only over the graph | Follow-up | D-2 chose window scope; pointer-inside gating is a later call |
| MNT-1 | MNT | minor | Fourth local-monitor + resign-key copy; the shared-helper follow-up has no home in `docs/ideas/` | Follow-up | plan Out line; idea file if the user agrees |
| MNT-2 | MNT | minor | `wirePointer()`'s Shift re-seed has no comment | Follow-up | minor; file not in the fix plan |
| MNT-3 | MNT | nit | `onToggleSelect` vs model `toggleSelect()` | Logged | plan chose the name |
| MNT-O1 | MNT (outside) | nit | `docs/ideas/source-structure.md:58` still says "Box-select is still S10-20" | Logged | stale before this PR; docs sweep |

### Fix plan 1

**F-1 — Landed note; the deferred keyboard gap gets its dogfood entry** (from CON-1, UX-3, CON-2/UX-4)
Files: `docs/deployment-plan/spike-10/prs/S10-57-quick-select.md`, `docs/dogfood/ux.md`.
Do: add `## Landed` under the plan header, in the shape of S10-56's and S10-20's ("Differences from this plan: …"), listing: the plan merged as S10-56 in #426 and was renumbered S10-57 on this branch (user, 2026-10-17; files, map row, scope heading, anchors); while Shift is held, Shift+Return and Shift+Space on a selected card do nothing, because `inputMode` is `.selecting` (follow-up); the implementer did not run §10, so AC-1.5, AC-1.7, the omnibar half of AC-1.6, and the canvas `flagsChanged` monitor are unchecked and left to the user; otherwise the code follows §6. In `docs/dogfood/ux.md` under **Open**, below the `<!-- add below this line -->` marker, add a stub in the file's shape: title "No keyboard way to add one card to a multi-selection", Date 2026-10-17, Where Evidence graph, Annoyance: Shift-click adds or removes one card, but keyboard and VoiceOver users have no way to do the same (deferred from S10-57).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-57-quick-select.md` reports no errors; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `455ab6fd`. Gate PASS (orchestrator run). F-1 done (docs only; no failing test applies). Deviations: none.
