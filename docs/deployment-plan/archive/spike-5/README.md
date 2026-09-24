# Spike 5 — Interpretation foundation (everything up to the canvas)

## Status

**Done.** Spike archived after S5-01…S5-10. Finished steps: [`completed.md`](completed.md). Product foundation closed in S5-09; floating-menu unify landed in S5-10.

Interpretation foundation: candidate refs, subject vocabulary, Subjects, layout storage, FFI, and Sources-family entry into a graph stub. Authoritative design: [`interpretation-graph-ui.md`](../../../ideas/archive/interpretation-graph-ui.md) §11.1 / §1.4. Authoritative schema: [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) §4.

> **This spike ships no Subject UI.** The graph is the only surface for Subjects, Citations, and Observations (design note §1.3), and the graph is Spike 6. Product nav keeps **one Sources family** — no Interpretation sidebar section (§1.4).

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, schema, gotchas, definition of done |
| [**Completed**](completed.md) | Finished steps (S5-01…S5-10, S5-D*) |
| [Design briefs](design/) | Claude Design summaries; archived briefs in [`design/archive/`](design/archive/) |

## Relationship to Spike 4 / Spike 6

Spike 4 shipped place loading (`WorkspaceSession`, `PlaceRegistry`). Spike 5 extends Sources (discriminator, nested sections, graph place) rather than adding an Interpretation section.

**Spike 6** is the Evidence graph canvas — zero schema work; first PR replaces the stub.

## Out of scope (for this spike)

- Canvas / Subject UI (Spike 6)
- Source-page Evidence graph button
- Subject type / field editors
- Product SemVer bump for docs/menu hygiene
