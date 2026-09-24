# Spike 6 — Claude Design briefs

**All canvas chrome in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Design rationale: [`interpretation-graph-ui.md`](../../../../ideas/archive/interpretation-graph-ui.md).

## Open

_None._

## Completed

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S6-D1 | [`archive/S6-D1-canvas-bubbles.md`](archive/S6-D1-canvas-bubbles.md) | PRs S6-02, S6-03 | Primary cards + place/create + uncited shell — write-up in [`../completed.md`](../completed.md) |
| S6-D2 | [`archive/S6-D2-connect-edges.md`](archive/S6-D2-connect-edges.md) | PR S6-04 | Connect, lines, bridge cards (no citation modal) — write-up in [`../completed.md`](../completed.md) |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language and tokens. Prefer existing `PV*` where chrome touches app UI; canvas-native pieces (bubbles, edges) may be new but should feel like the same product.
4. When the board is done, archive the brief and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project.
- **Evidence graph** is the product name for this Source-scoped canvas (engine: Interpretation layer).
- Entry: Sources list → Evidence graph (Spike 5). No Interpretation sidebar section.
- Subjects on this canvas: Person / Event / Place roots; bridge mid-bubbles for connect. **`source` subjects never appear.**
- Do **not** use "claim." Candidate refs (`CPR-…`, etc.) may appear in a11y or inspector chrome later — not required on the bubble face for this spike.
- This spike is a **prototype** to validate the canvas. Prefer clarity over ornament. No minimap, multi-select, or auto-layout.
- Extend: Sources family chrome (S5), design tokens — not a parallel visual language.
