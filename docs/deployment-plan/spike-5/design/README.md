# Spike 5 — Claude Design briefs

**All UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Each is self-contained: objective, domain facts the UI must reflect, numbered requirements, screen inventory, out-of-scope, and acceptance checks.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Design rationale for the layer: [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S5-D2 | [`S5-D2-interpretation-sources-list.md`](S5-D2-interpretation-sources-list.md) | PR S5-08 | The Source picker; stub behind a row is fine |

**This spike designs no Node UI at all.** The interpretation graph is the only surface for Nodes, Citations, and Observations (design note §1.3), and it is Spike 6. Two briefs that existed here were deleted when that was settled — a Source-page entry control and a per-Source node list — and they are not re-homed. Spike 6 gets a fresh design track once this spike's results are in.

The destination behind a Source can be a plain "coming soon" stub. Early development — placeholders are fine; Spike 6 replaces the view.

## Completed

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S5-D1 | [`archive/S5-D1-interpretation-nav-entry.md`](archive/S5-D1-interpretation-nav-entry.md) | PR S5-07 | Interpretation sidebar destination — label, icon, placement |

## How to use

1. Open the existing Provenencia Claude Design project / design-system bundle (see `macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the visual language aligned with the shipped app (parchment neutrals, serif display, Spectral body, iron-gall accent). Do not invent a second brand.
4. Prefer existing components — `PVList`, `PVButton`, `PVIconButton`, `PVEmptyState`, `PVSelect`, `PVConfirm`, `PVField`, `PVInput`, `PVToast`. This spike should add **no** new design-system primitives; if a board seems to need one, that is a finding worth raising rather than drawing.
5. When the board is done, move the brief into `archive/` and add a step write-up in [`../completed.md`](../completed.md). The implementing PR still gates on the board.

## Shared product facts (all briefs)

- Offline-first macOS genealogy app. After onboarding the researcher is working inside a local `*.provenencia` project folder.
- The app has three layers. **Source** is evidence as filed (shipped). **Interpretation** is what a single Source *appears to say* (this spike, first slice). **Conclusion** is what the researcher believes across Sources (not built).
- **Interpretation is Source-scoped.** Every Node belongs to one home Source, and a work session is "sit with one Source and map what it appears to say." There is no cross-Source view.
- Short human refs are shown in mono and are never editable. Sources are `SRC-…`. Interpretation Nodes are **candidates** and carry their own prefix: a person Node is `CPR-7KD45`, an event `CEV-…`, a place `CPL-…`. The distinct prefix is meaningful — `CPR-…` marks "this is what one source seems to say," against the Conclusion layer's `PER-…`, "this is a person I concluded existed." Same shape, different prefix; nothing in the UI should try to derive one from the other.
- **Nothing in this spike is cited yet.** Citations and Observations arrive in a later spike. So a Node here has a type, a ref, and an optional working label — and no asserted facts at all. Boards must not imply otherwise.
- **This is the spike before the canvas.** Spike 6 is the spatial graph. Design the Sources list as the permanent picker into that graph; a stub behind a row is fine until the canvas exists. Do not design the canvas, and do not invent a node list or table view — the graph is the only Node surface.
- Existing boards to extend rather than reinvent: the workspace chrome and sidebar (Spike 2 S2-01), the Sources list (S2-04), and the Source page (S2-23, with a checked-in export at `archive/spike-2/design/boards/source-page.dc.html`).
