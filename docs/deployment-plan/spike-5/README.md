# Spike 5 — Interpretation foundation (everything up to the canvas)

## Status

**Planned.** Finished steps: [`completed.md`](completed.md).

Lay the Interpretation-layer **data** track — candidate refs, Node vocabulary, Nodes, layout storage, FFI — and the **Sources-family** entry into an Evidence graph stub, so **Spike 6 can start drawing.** The canvas itself is out of scope.

> **No Node UI.** The Evidence graph is the only surface for Nodes, Citations, and Observations (design note §1.3); Spike 6 draws it. Product nav is one **Sources** family with nested config (§1.4) — no Interpretation sidebar section. Entry: dual action on the Sources list → Evidence graph stub.

Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §11.1 / §1.4. Schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4. Seed: [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1.

## Why this spike exists

The design note builds the canvas **first**; a canvas needs rows. Spike 5 answers: what does the first `nodes` INSERT require? (`sources` exists; `node_types` does not.) Everything else hangs off Observations and does not block a bubble.

## Goal

A researcher can:

1. See **Sources** as the primary sidebar item, with Source types / Source fields / Subject types / Subject fields as nested config (stubs OK for Subject*).
2. From the Sources list, open a Source page **or** an Evidence graph stub (when the Source has an Artifact).
3. Leave and return via Back/Forward with the place restored.

Everything else is proven by test: Node CRUD with candidate refs, positions, FakeStore RPCs.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, schema, gotchas, definition of done |
| [**Design briefs**](design/) | S5-D3 (nav), S5-D2 (list → Evidence graph); S5-D1 superseded |
| [**Completed**](completed.md) | Finished steps (S5-01…) |
| [Interpretation / Evidence graph UI](../../ideas/interpretation-graph-ui.md) | Design rationale |

## Relationship to the spikes around it

Spike 4 shipped place loading (`WorkspaceSession`, `PlaceRegistry`). Spike 5 extends Sources (discriminator, nested sections, graph place) rather than adding an Interpretation section.

**Spike 6** is the Evidence graph canvas — zero schema work; first PR replaces the stub.

## Out of scope

- The canvas (drawing, drag, snap, tray, connect).
- Any UI that creates or edits a Node.
- Observations, Citations, locator validation, NameValue, artifact viewer.
- Subject types / Subject fields **editors** (nav stubs only).
- Omnibar projection for Nodes.
- Source-to-source commentary on the Source page.
