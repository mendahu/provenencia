# Rethink source metadata and source-type field suggestions

- **Date:** 2026-10-25
- **Type:** Product rethink / possible descope
- **Where:** Sources / source metadata / source-type → metadata field suggestions
- **Annoyance:** Jake wants to rethink source metadata and the association between source types and the metadata fields offered as suggestions. In practice he is finding it isn't very useful and thinks it may be possible to descope it.
- **Wanted:** A decision on whether source metadata and type suggestions stay, shrink, or go. Nothing is proposed here; the section below is a factual map of what exists today so the blast radius of any descope is clear.

## Current footprint

Observational only. Nothing below is a recommendation.

### Storage (catalog schema)

Four tables, all registered in `core/database/catalogmodel`:

- `source_metadata_fields` — vocabulary rows: `key`, `origin` (`provenencia` / `user`), `label`, `data_type` (`text` | `url` only, since migration `000028` dropped `date`), `description`. `UNIQUE (key, origin)`. Created `000005`, reshaped `000027` (url) and `000028` (date removed).
- `source_type_metadata_fields` — the type → suggested field join (`type_id`, `field_id`, `sort_order`). No origin of its own. Created `000005`.
- `source_metadata` — one value per `(source_id, field_id)`, `value_text` only. Created `000009`, flattened to text-only in `000028`.
- `source_metadata_layout` — per-Source presentation state: `sort_order` and `dismissed` per field (which suggestions this Source has waved off, and row order). Created `000013`.

Metadata values are filing text only. They never resolve to Interpretation subjects, canonical entities, or DateValues (`docs/seeded-vocabulary.md` §2.2, `000028.sql` header).

### Go

- `core/database/metadatafields/` — field vocabulary CRUD: `Upsert`, `Create`, `Update`, `GetByID`, `Lookup`, `List`, `UsedBy`, `CountByOrigin`, `Delete` (goes through `deleteimpact`; refused while any `source_metadata` row references the field).
- `core/database/sourcemetadata/` — per-Source values and layout: `Set`, `CreateAll`, `Clear`, `DismissSuggestion`, `Reorder`, `ListBySource`, `ListWorkspace` (merges saved values with the type's undismissed suggestions for the Source page board), `ListLayoutTx`.
- `core/database/sourcevocab/` — the create-time seed registry (`registry.go`) and `Install`, plus the suggestion join ops used by the Source types admin: `CountSuggestions`, `AppendSuggestion`, `DeleteSuggestion`.
- `core/database/sourcecopy/` — Copy Source (Spike 12) copies metadata values **and** layout rows verbatim when the "metadata" checkbox is on.
- `core/database/deleteimpact/` — `KindMetadataField` registered with count/list probes (`register.go`), facets and projectors; Source delete cascades `source_metadata` and `source_metadata_layout`.
- `core/database/effects/registry.go` — `source_metadata` and `source_metadata_layout` are effect entities (feeds audit scopes and "last activity").
- `core/database/searchindex/` + `core/search/` — `KindMetadataField` is an omnibar hit kind (field key / label / description indexed; `ReprojectMetadataField`), and metadata values are folded into the Source's FTS body under `BodyTagMetadata` (`metadata:\t`).
- `api/ffi/handlers/source_defs.go` and `sources.go` — 11 RPCs: `ListMetadataFields`, `CreateMetadataField`, `UpdateMetadataField`, `DeleteMetadataField`, `ListTypeSuggestions`, `AssignTypeField`, `RemoveTypeField`, `SetSourceMetadata`, `ClearSourceMetadata`, `DismissSourceMetadataSuggestion`, `ReorderSourceMetadata`. `GetSourceWorkspace` returns the merged metadata board. `workspace_counts.go` reports metadata-field origin counts for the sidebar.

### Seeded data

- Shipped seed (`core/database/sourcevocab/registry.go`) is deliberately small: one type (`birth_certificate`), three fields (`document_number`, `record_date`, `issue_date`), three suggestions binding them. Installed at catalog create only (`core/onboarding/ready.go`); existing catalogs are never healed.
- The documented horizon (`docs/seeded-vocabulary.md` §2.1–2.3) is much larger: nine source types, a long field list (`creator`, `author`, `photographer`, `jurisdiction`, `certificate_number`, `repository`, `collection`, …), and per-type suggestion lists. Only the slice above is implemented.

### UI surfaces (macOS)

- **Source page** — `Features/Sources/SourcePageMetadataView.swift` + `SourceMetadataSection.swift` (model): the "Metadata" board beside Credibility: saved rows (reorderable, editable, clearable), a "Suggestions" group from the type with per-row Add / Dismiss, an Add dialog listing every field, and empty state. Layout constants in `SourcePageChrome.swift`.
- **Copy Source dialog** — `SourceCopyDialog.swift` / `SourceCopyModel.swift`: "Copy metadata" checkbox (`includeMetadata`, `metadataAvailable`).
- **Metadata admin** — sidebar destination `WorkspaceSection.metadata` (renamed from `source-fields` in S9-07b): `Features/Metadata/` (`MetadataView`, `MetadataListPane`, `MetadataDetailPane`, `MetadataModel`) — create / edit / delete fields, shows used-by counts; registered in `PlaceRegistry`, `CatalogQueryKey.metadataFieldsList`, `WorkspaceDestinationHost`, `WorkspacePresentationID`.
- **Source types admin** — `Features/SourceTypes/` (`SourceTypesDetailPane` etc.): per-type "Suggested fields" assign / remove, backed by `ListTypeSuggestions` / `AssignTypeField` / `RemoveTypeField`.
- **Omnibar** — metadata fields appear as search hits (`KindMetadataField`); Source hits can match on metadata values.
- **Localization** — ~235 `metadata` mentions in `Platform/L10n.swift` (Sources board copy, Metadata admin, Source types "suggested fields", Copy dialog).
- **Tests** — Go: `metadatafields_test.go`, `sourcemetadata_test.go`, `sourcevocab_test.go`, plus deleteimpact / searchindex / sourcecopy coverage. Swift: `PlaceRegistryTests`, `WorkspaceNavigationTests`, `CatalogSearchTests`, `PVDeleteImpactTests` and others reference the metadata place or kind.

### Docs that describe it

- `docs/source-layer-data-model.md` §5 (`5.1`–`5.4`: fields, type suggestions, values, layout) and the Copy Source section.
- `docs/seeded-vocabulary.md` §2.2–2.3 (field horizon and per-type suggestions) and §1.1 origin rules.
- `docs/catalog-deletes.md`, `docs/audit-revision-history.md` (metadata as a delete-impact kind and audit scope), `docs/macos-client-patterns.md`, `docs/structured-date-model.md` (why metadata dates are text).
- Skills: `.agents/skills/add-seeded-vocabulary/SKILL.md` (sourcevocab registry is its primary example).
- Ideas: `docs/ideas/registry-source-import.md` says a registry record's collection details "for now go in the Source metadata"; `docs/ideas/settled-list-count-publish.md` audits `MetadataModel` count publishing.

### What depends on it

- **Copy Source** (Spike 12) — copies values + layout; the checkbox and `sourcecopy` draft carry metadata.
- **Search** — a hit kind (`metadata_field`) and part of the Source FTS body; removing either changes the omnibar registry and FTS projection.
- **Delete impact** — `KindMetadataField` and the Source → metadata cascade are part of the registered impact graph and `TestPragmaHonesty`-style model checks.
- **Audit / last activity** — `source_metadata` and `source_metadata_layout` writes are effect entities that scope to a Source.
- **Sidebar counts** — vocabulary origin counts include metadata fields.
- **Workspace navigation** — `metadata` is a persisted `WorkspaceSection` with a legacy id alias; removing it needs a `legacyIDs` fallback so restored history does not dangle.
- **Citation formatting / reference output** — nothing today. No code formats citations or exports from metadata; `core/database/citations` and the citation composer do not read `source_metadata`.
- **Import / export** — nothing today. There is no GEDCOM / export path reading metadata; `registry-source-import.md` is the only forward reference (idea stage).
