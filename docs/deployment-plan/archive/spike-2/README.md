# Spike 2 — Source layer catalog

**Done.** Validated Source-layer schema, audited CRUD, ingest, and the macOS workspace. Live models: [`source-layer-data-model.md`](../../../source-layer-data-model.md), [`artifact-file-storage.md`](../../../artifact-file-storage.md), [`audit-revision-history.md`](../../../audit-revision-history.md), [`seeded-vocabulary.md`](../../../seeded-vocabulary.md).

## Decisions

- First researched mutation **writes audit** in the same SQLite transaction. No `created_at` / `updated_by` on Source tables.
- **Workspace chrome:** sidebar + content host. Destinations: Sources (list → Source page), Source types, Source fields. Sign Out is an app-menu action.
- **File bytes never over protobuf.** Ingest writes content-addressed `objects/{hh}/{hh}/{sha256}`; a better scan is a **new Artifact**, not a pointer swap.
- Derivatives belong to **Files**, not Artifacts, and need not audit.
- **Seed small**; grow types/fields from use. Create-time starter, not the whole vocabulary horizon.
- Project-wide **Files list** was planned and **descoped**. Ingest and open stay on the Source page.

Swift stays thin; Go owns schema, ingest, and validation.
