# Spike 5 — Interpretation foundation (everything up to the canvas)

## Status

**Planned.** No steps landed. Finished steps will be recorded in [`completed.md`](completed.md).

Lay the first Interpretation-layer track — candidate refs, Node vocabulary, Nodes, layout storage, FFI, and a workspace place — so that **Spike 6 can open a file and start drawing.** The canvas itself is deliberately out of scope: this spike ends exactly where it begins.

Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §11.1 ("the load-bearing minimum"). Authoritative schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) §4. Seed vocabulary: [`seeded-vocabulary.md`](../../seeded-vocabulary.md) §3.1.

## Why this spike exists

The design note argues for building the canvas **first**, because it is the only part of the Interpretation layer with no prior art and no fallback. The constraint is that a canvas needs rows to place. So the question this spike answers is narrow:

> What does the first `nodes` INSERT actually require?

The answer is small. `nodes` has exactly two foreign keys — `sources`, which exists, and `node_types`, which does not. Everything else in the layer (`properties`, `observations`, `citations`, NameValue, the artifact viewer) hangs off **Observations**, not Nodes, and none of it blocks a bubble on a grid. That is the whole scoping argument, and it is why this spike is shorter than it looks like it should be.

## Goal

A researcher can:

1. Open a Source, click **Interpretation**, and land on a graph destination scoped to that Source.
2. Create `person`, `event`, and `place` Nodes there, see their candidate refs (`PER-C-7KD45`), rename them, and delete them.
3. Leave, navigate elsewhere, and return via Back/Forward or the sidebar with the place restored and the session cache warm.
4. See every Node write in the audit log.

Nothing here is genealogically useful yet — there are no Observations, so nothing is cited. The point is that the rail is laid.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, schema, gotchas, definition of done |
| [**Completed**](completed.md) | Finished steps (S5-01…) |
| [Interpretation graph UI](../../ideas/interpretation-graph-ui.md) | Design rationale and the decision log this spike implements |

## Relationship to the spikes around it

Spike 4 shipped **how places load and render** (`WorkspaceSession`, `CatalogQueryRegistry`, `PlaceRegistry`). Spike 5 is the first real customer of that extension point that is not Sources-shaped — if `add-workspace-place` works as advertised, this spike is mostly registry rows.

**Spike 6 is the canvas**, and it starts with zero schema work: migration, Go, FFI, store, place, and navigation are all done here. Its first PR draws bubbles.

## Out of scope

Everything below is deferred on purpose. None of it blocks the canvas.

- **The canvas.** No `NSScrollView` bridge, no drawing, no drag, no snap-to-grid, no tray, no connect tool.
- **Observations and Citations.** No `properties`, `node_type_properties`, `observations`, `citations`, or locator validation.
- **Bridge Node macros.** `relationship` / `participation` / `location` types are *seeded*, but nothing creates them — they need Observations to mean anything.
- **The artifact viewer** and NameValue (§4.4 of the design note — its own chunk of work, on the Observation path).
- **The vocabulary browser UI.** Node Types are seeded and read-only; researcher-defined types and `ref_prefix` validation UX come with the browser in Spike 7.
- **Search.** Nodes are explicitly **not** projected into the omnibar — see the deployment plan's gotcha list for the six-file surface that would entail.
- **Source-to-source commentary.** The `source` Node Type is seeded, but the Source-page flow it feeds (design note §4.6) is unspecified and unscheduled.
