# Spike 7 — Citations, Observations, composer

**Done.** Citation → Observation pipeline on the Evidence graph. Schema: [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md). Themes: [`interpretation-graph-ui.md`](../../../ideas/archive/interpretation-graph-ui.md). Composer rethink is Spike 8 (**S8-D7**).

## Decisions

- **Composer is Option B:** a navigable workspace place (artifact viewer \| form). Not a sheet over the graph, not a companion window. Back returns to the graph.
- **Subject types stay product-seeded** (person / event / place / bridges / source). No types editor. Behavior lives in one Interpretation subject registry (capabilities, locked bindings, connect matrix, term Properties).
- **Subject fields** (`properties` + bindings) remain the extensible config surface.
- Observation value types for v1: `text`, `integer`, `date`, `name`, `subject`, **`term`**. Not `real` / `boolean`.
- Kind/edge Properties use **`property_terms`**. Term-typed Properties are registry-driven; the composer picker may Add custom **term rows**. No Event types / Roles admin destinations.
- Connect writes a **cited bridge** (subject + position + Citation + edge Observations) in one transaction. Endpoint Properties are system-owned; Add property excludes them.
- Locator: default `artifact`, optional PDF `page`, optional `region` polygon (one region).
- NameValue is its own editor, hosted by the composer for `value_type = name`.
- The shipped composer was **subject-locked** (one card per trip; Save required ≥1 Observation). That policy is what Spike 8 rethinks — the schema already allows one Citation × many subjects and an empty reading.
