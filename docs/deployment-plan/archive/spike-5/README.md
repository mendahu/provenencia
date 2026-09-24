# Spike 5 — Interpretation foundation (up to the canvas)

**Done.** Subjects, layout, vocabulary seed, FFI, and a Sources-family graph **stub**. No Subject UI — the first visible subject is a Spike 6 bubble. Schema: [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md). Themes: [`interpretation-graph-ui.md`](../../../ideas/archive/interpretation-graph-ui.md).

## Decisions

- **One Sources family.** Filing and the source-scoped graph are one product layer. No Interpretation sidebar section.
- The **graph is the only Subject surface** (no co-equal list). Spike 5 ships data + stub only.
- `subjects` / `subject_types` / `subject_positions`. Positions are unaudited, keyed by `subject_id`, **CASCADE** on subject delete. Evidence FKs stay **NO ACTION**.
- **One ref format, two prefixes** (`ref_prefix` + `candidate_ref_prefix` on `subject_types`).
- Node was renamed **Subject**. Bridge rows are subjects too. Do not use “claim” (Conclusion owns that word).
- Config under Sources: Source types / fields and Subject types / fields are nested, de-emphasized.
- Floating menus unified to one kit (`PVContextMenu` / `PVSelect` / jump menu). Omnibar and ComboBox keep their own hosts.
