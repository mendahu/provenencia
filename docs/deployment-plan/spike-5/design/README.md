# Spike 5 — Claude Design briefs

**All UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Design rationale: [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S5-D3 | [`S5-D3-sources-section-nav.md`](S5-D3-sources-section-nav.md) | PR S5-07 | Nested Sources family: Sources primary; four config children; Subject types / fields stubs |
| S5-D2 | [`S5-D2-sources-list-graph-entry.md`](S5-D2-sources-list-graph-entry.md) | PR S5-08 | Dual action on Sources list → Source page or **Evidence graph** stub |

Run **S5-D3 before S5-D2**.

## Completed

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S5-D1 | [`archive/S5-D1-interpretation-nav-entry.md`](archive/S5-D1-interpretation-nav-entry.md) | — | **Superseded.** Flat Interpretation item — do not implement. Replaced by S5-D3. |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language. Prefer existing `PV*` components; no new design-system primitives unless raised as a finding.
4. When the board is done, archive the brief and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project.
- **Product IA:** Sources is primary work. Source types / Source fields / Subject types / Subject fields are **nested config** under Sources. No Interpretation sidebar section. Conclusion is the separate belief layer (not built).
- **Evidence graph** is the product name for the Source-scoped canvas (engine: Interpretation layer — Citations / Observations / Nodes).
- **Subject types / Subject fields** map to `node_types` / `properties`. Do **not** use "claim."
- Candidate refs: `CPR-…` etc. vs concluded `PER-…`.
- This spike: no Node UI; Evidence graph may be a stub. Do not design the canvas or a node list.
- Extend: sidebar (S2-01), Sources list (S2-04), Source page (S2-23).
