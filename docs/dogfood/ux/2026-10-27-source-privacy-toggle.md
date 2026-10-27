# Source-level privacy toggle (not public)

- **Date:** 2026-10-27
- **Type:** Feature request (foundation for sharing)
- **Where:** Source detail page / Sources list / Source data model
- **Annoyance (verbatim):** "We may already have this but just in case we don't, we need a privacy toggle at the source level to mark certain sources as not public (which we'll be able to use for different features later on down the line, such as sharing)."
- **Consequence:** There is no way to mark a Source as private today. Any future share / export / sync feature would have nothing to filter on, so researchers could not keep sensitive sources (living-person records, private correspondence, restricted archive copies) out of a package without deleting them.
- **Wanted:** A per-Source "not public" toggle (default public) stored in the catalog and surfaced on the Source page, intended as a foundation that later features — share packages, exports, sync, web access — can honor. No behaviour beyond storing and showing the flag is asked for now.

## Does it already exist?

**No.** Searched the catalog schema (all migrations), the Go core and FFI handlers, `engine.proto`, every macOS feature folder, and `docs/`. There is no privacy, visibility, public/private, sensitive, restricted, or redaction flag on Sources, Artifacts, Evidence graphs, Subjects, canonical entities, or the project. `sources` today is `id, ref, source_type_id, title, description, cover_mode, primary_artifact_id` (`000006.sql`, `000011.sql`, `000015.sql`). No living-person privacy model exists either.

### Related horizon mentions (not mechanisms)

- `docs/ideas/share-packages.md` — "Share a Source" idea (not roadmapped) lists "Privacy / redaction: notes and files may contain living-person data; any warn-before-share UX?" as an open question. This toggle is the natural answer to that question's first half.
- `docs/narrative-layer-data-model.md` §9.5 / open question 7 — "Living people — Redaction and warn-before-share for narrative exports." Horizon only.
- `docs/user-identity-model.md` — frames future sync / sharing / hosted / web access as extension boundaries; explicitly has no permission or membership model today. A Source flag would sit below any of that.
- `docs/ideas/place-gazetteer-service.md` — a privacy note about an opt-in lookup service, unrelated to per-Source data.

## Likely code locations (observed, not implemented)

The closest precedent is `cover_mode` / `primary_artifact_id` — a small per-Source attribute added by one migration and carried end to end — so each layer below names the file where that attribute lives.

- **Schema.** New migration: `ALTER TABLE sources ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1 CHECK (is_public IN (0, 1))` (or a `visibility TEXT` enum if more than two states are foreseen). Register in `core/database/catalogmodel`; follow `.agents/skills/add-catalog-migration/SKILL.md`. Existing rows default to public so nothing changes on upgrade.
- **Go.** `core/database/sources/sources.go` — `Source` struct, `Create`, `Update`, list / get SQL. `core/database/sourcecopy` — decide whether Copy Source carries the flag (probably yes, with a checkbox like metadata / credibility). `core/database/effects/registry.go` already treats `sources` row changes as a Source-scoped effect, so audit / last-activity pick it up for free.
- **FFI.** `api/proto/engine.proto` `Source` message plus `UpdateSourceRequest` (or a dedicated `SetSourcePrivacy` RPC, mirroring `METHOD_SET_SOURCE_COVER`); `api/ffi/handlers/sources.go`.
- **Swift model.** `macos/App/Platform/GenealogyStore.swift` `CatalogSource` (beside `coverMode`), `GoStore.swift` mapping, `FakeStore.swift`.
- **Swift UI.** Source page: a toggle in the identity header or its own small section (`macos/App/Features/Sources/SourcePageIdentityHeader.swift` / `SourceIdentitySection.swift`; `PVSwitch` exists in `DesignSystem/Components/Switch`). Sources list: a "Private" glyph on `SourcesListRow.swift` and optionally a filter in `SourcesListSettings.swift` / `SourcesModel`. Copy dialog: `SourceCopyDialog.swift`. All copy through `L10n` (`add-localized-string`).
- **Docs.** `docs/source-layer-data-model.md` (sources table + a "visibility" note), `docs/ideas/share-packages.md` (close its privacy question by pointing here), `docs/seeded-vocabulary.md` only if a vocabulary rather than a boolean is chosen.

## Open questions for the design

- Boolean (`is_public`) vs. tri-state (`public` / `private` / `restricted`)? The verbatim ask is binary.
- Should a private Source hide anything in-app today (omnibar hits, Sources list by default), or is it purely stored for later consumers? The ask implies the latter.
- Does privacy cascade (a private Source's Artifacts, Citations, and the canonical entities they ground) when a future share feature filters on it, or is it the share feature's job to decide?
