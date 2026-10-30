# Source-level privacy toggle (not public)

- **Date:** 2026-10-27
- **Type:** Feature request (foundation for sharing)
- **Where:** Source detail page / Source data model
- **Annoyance (verbatim):** "We may already have this but just in case we don't, we need a privacy toggle at the source level to mark certain sources as not public (which we'll be able to use for different features later on down the line, such as sharing)."
- **Consequence:** There is no way to mark a Source as private today. Any future share / export / sync feature would have nothing to filter on, so researchers could not keep sensitive sources (living-person records, private correspondence, restricted archive copies) out of a package without deleting them.
- **Wanted:** A per-Source on/off "not public" switch (default public) persisted in the catalog and toggled on the Source page. It is a foundation for later features — share packages, exports, sync, web access — and does nothing else yet.

## Decisions (Jake, 2026-10-27)

Verbatim: "For now let's make it a simple on/off switch and it doesn't do anything yet. I just want the data to be in there and the ability to toggle it for now."

- **Shape:** boolean on/off. Not tri-state.
- **Behaviour:** none yet. Nothing is hidden, filtered, or warned about in-app; omnibar, Sources list, and composers are unchanged.
- **Cascade:** none for now. The flag lives on the Source row only; Artifacts, Citations, and grounded entities are not touched. Whether a future share feature cascades is that feature's decision.
- **Scope:** persist the field end to end, and a toggle on the Source page. Everything else below is marked out of scope.

## Does it already exist?

**No.** Searched the catalog schema (all migrations), the Go core and FFI handlers, `engine.proto`, every macOS feature folder, and `docs/`. There is no privacy, visibility, public/private, sensitive, restricted, or redaction flag on Sources, Artifacts, Evidence graphs, Subjects, canonical entities, or the project. `sources` today is `id, ref, source_type_id, title, description, cover_mode, primary_artifact_id` (`000006.sql`, `000011.sql`, `000015.sql`). No living-person privacy model exists either.

### Related horizon mentions (not mechanisms)

- `docs/ideas/share-packages.md` — "Share a Source" idea (not roadmapped) lists "Privacy / redaction: notes and files may contain living-person data; any warn-before-share UX?" as an open question. This toggle is the natural answer to that question's first half.
- `docs/narrative-layer-data-model.md` §9.5 / open question 7 — "Living people — Redaction and warn-before-share for narrative exports." Horizon only.
- `docs/user-identity-model.md` — frames future sync / sharing / hosted / web access as extension boundaries; explicitly has no permission or membership model today. A Source flag would sit below any of that.
- `docs/ideas/place-gazetteer-service.md` — a privacy note about an opt-in lookup service, unrelated to per-Source data.

## Likely code locations (observed, not implemented)

The closest precedent is `cover_mode` / `primary_artifact_id` — a small per-Source attribute added by one migration and carried end to end — so each layer below names the file where that attribute lives.

### In scope

- **Schema.** New migration: `ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1))`. Register in `core/database/catalogmodel`; follow `.agents/skills/add-catalog-migration/SKILL.md`. Existing rows default to public so nothing changes on upgrade.
- **Go.** `core/database/sources/sources.go` — `Source` struct, `Create`, `Update`, list / get SQL. `core/database/effects/registry.go` already treats `sources` row changes as a Source-scoped effect, so audit / last-activity pick it up for free.
- **FFI.** `api/proto/engine.proto` `Source` message plus `UpdateSourceRequest` (or a dedicated `SetSourcePrivacy` RPC, mirroring `METHOD_SET_SOURCE_COVER`); `api/ffi/handlers/sources.go`.
- **Swift model.** `macos/App/Platform/GenealogyStore.swift` `CatalogSource` (beside `coverMode`), `GoStore.swift` mapping, `FakeStore.swift`.
- **Swift UI.** Source page only: an on/off switch in the identity header or its own small section (`macos/App/Features/Sources/SourcePageIdentityHeader.swift` / `SourceIdentitySection.swift`; `PVSwitch` exists in `DesignSystem/Components/Switch`). Copy through `L10n` (`add-localized-string`).
- **Docs.** `docs/source-layer-data-model.md` (sources table + a one-line "not public" note), `docs/ideas/share-packages.md` (point its privacy question here).

### Out of scope for now (per the decisions above)

- Sources list "Private" glyph on `SourcesListRow.swift`, and any filter in `SourcesListSettings.swift` / `SourcesModel`.
- Omnibar / search hiding, composer or graph behaviour.
- Carrying the flag through Copy Source (`core/database/sourcecopy`, `SourceCopyDialog.swift`). A copy starts public unless decided otherwise later.
- Any cascade to Artifacts, Citations, or canonical entities.
- A `visibility` enum or `docs/seeded-vocabulary.md` entry (boolean chosen).

## Design questions

All resolved by the 2026-10-27 decisions above:

- ~~Boolean vs. tri-state~~ — boolean on/off.
- ~~Hide anything in-app today~~ — no; stored only.
- ~~Cascade to Artifacts / Citations / entities~~ — no; the Source row only, future consumers decide.
