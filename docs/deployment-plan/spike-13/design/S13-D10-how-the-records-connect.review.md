# Review — S13-D10 How the records connect

Brief: [`S13-D10-how-the-records-connect.md`](archive/S13-D10-how-the-records-connect.md). Reviewed with its plan; plan log: [`../prs/S13-28-how-the-records-connect.review.md`](../prs/S13-28-how-the-records-connect.review.md). Placement (beside the rows at both widths, pinned) answered by the user before review, 2026-11-07.

## Round 1 (2026-11-07)

Maintainability (opus): **Fits with changes**. Executability (sonnet): **Ready with changes**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| XB-1 | blocker | No rule for the diagram column while the section itself shows a state | **Fixed.** RL-14 and frame 8: no diagram column; the section's states full width |
| B-1 | major | The node's "button behaviour from the focus ring style" does not exist there; no kit cursor | **Fixed.** Node row names `PVHoverEffect`, `pvFocusRing`, `pvActivatesOnReturn`; RL-6 drops the cursor |
| B-2 | major | The split and pin need one home outside the Sources section | **Fixed.** Inventory rows "List with a pinned side column" (New, second call site Event and Place) and "Visible-area reading" (Extend) |
| XB-2 | major | RL-1 leaves the mark and diameter with no default | **Fixed.** RL-1: an 8pt dot drawn by the diagram, label room, label avoidance recommended |
| XB-3 | major | `check-plan-docs.py` warns on New rows | **Declined.** The script warns on every New row whatever the note (`check-plan-docs.py` line 319); each New row names its second call site |
| XB-4 | major | No Wide frames for 50 graphs and the pin; no `-W` for states | **Fixed.** Frame 7 adds both; RL-W4 |
| XB-5 | major | The pin's gap unstated | **Fixed.** RL-13: 24pt, the page's vertical padding |
| B-3 | minor | `accentSoft` is the Badge fill | **Fixed.** `surfaceSelected` |
| B-4 | minor | Warning Callout, a second look for a failed read | **Fixed.** Danger, compact |
| XB-6 | minor | RL-3 has no default | **Fixed.** Recommended: the missing arcs are the mark |
| XB-7 | minor | Scope matches the map | — |
| XB-8 | minor | Broken review link | **Fixed.** This log |
| XB-9 | minor | Dark frames omit 1b | **Fixed** |

Round 2 runs because round 1 had blockers.

## Round 2 (2026-11-07)

Maintainability (opus): **Fits**. Executability (sonnet): **Ready with changes**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| XB-10 | blocker | No requirement for narrow panes | **Fixed.** RL-W5 and frame 9, with a recommendation; the user answers it on the board |
| XB-11 | blocker | RL-W2 covers the title, not the graph lines | **Fixed.** RL-W2 asks for the lines at that width, with a recommendation |
| XB-12 | minor | No floor for the short window | **Fixed.** RL-13: a 240pt circle, then the column scrolls |
| XB-13 | minor | Label cut length | **Fixed.** Cut to a ref's width; RL-1 decides |
| B (minor) | minor | The node tooltip is not "the same as" the line's | **Fixed.** RL-7 says why it adds the Source title |

## Verdict

Ready for the board. Open for the board: every "Decide" item, RL-W2's lines and RL-W5.
