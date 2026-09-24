# Interpretation graph (archived)

**Not a backlog.** Shipped as [Spike 5](../../deployment-plan/archive/spike-5/), [Spike 6](../../deployment-plan/archive/spike-6/), and [Spike 7](../../deployment-plan/archive/spike-7/). Schema: [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md). Leftovers live in Spike 8, [`docs/dogfood/ux.md`](../../dogfood/ux.md), or other [`docs/ideas/`](../) files.

## Decisions

| | Decision |
| --- | --- |
| Surface | Source-scoped **Evidence graph**, not list/detail per table. The graph is the only Subject / Citation / Observation surface. |
| Entry | Sources list dual-action (page \| graph) and the Source page. **No Interpretation sidebar.** Config (types / fields) is nested under Sources. |
| Subjects | Seeded person / event / place / bridges / `source`. `source` subjects **never render on the canvas**. Node → **Subject**. |
| Connect | Atomic bridge + edges; disambiguation on the graph; durable cite in Spike 7. |
| Uncited | No schema flag — subjects with zero Observations. `subjects.label` is the working handle. |
| Citation | One-to-many Observations. Cross-source Observations stay legal; the canvas is `source_id = ?`. |
| Composer | Navigable **place** (Option B), not a sheet. Spike 7 shipped subject-locked; Spike 8 rethinks flexibility. |
| Layout | Unaudited `subject_positions` keyed by `subject_id`. Positions travel with the project; camera does not. CASCADE on subject delete. Unplaced **tray descoped** — create always writes a position. |
| Canvas | AppKit `NSScrollView`. Neutral `GraphCanvas` geometry; position tables stay per-layer. |
| Types | Product-seeded + first-class only. **Subject fields** are the extensible surface. Capabilities / connect / term Properties live in one registry. |
| Values | `text`, `integer`, `date`, `name`, `subject`, `term`. Kind/edge Properties use `property_terms`. |
| Refs | One format, two prefixes (`ref_prefix` + `candidate_ref_prefix`). |
| Commentary | Source-to-source is a Source-page concern — [`source-to-source-relationships.md`](../source-to-source-relationships.md). |

## Leftovers (where they went)

| Theme | Home |
| --- | --- |
| Composer rethink (reuse, multi-subject, empty Save) | Spike 8 **S8-D7** |
| Image Auto Transcribe; PDF Find / paste; page ⇄ graph; list counts; delete paths | Spike 8 |
| Density filters on a huge graph | [`docs/dogfood/ux.md`](../../dogfood/ux.md) |
| `text_quote` locators | [`text-quote-locators.md`](../text-quote-locators.md) |
| Audio / video Sources | [`audio-video-sources.md`](../audio-video-sources.md) |
| Collapse/expand, undo, tray, minimap, QuickLook, Change type, adopt/import | Descoped |
| Conclusion / Narrative (sameness, family tree) | Those layers’ model docs |
