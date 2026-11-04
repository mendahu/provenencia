# Add a boolean property value type

- **Date:** 2026-10-25
- **Type:** Feature request (data model)
- **Where:** Properties / property value types
- **Annoyance:** Yes/no facts from a record ("literate", "can sign name", "veteran", "owns home" on a census) have no natural home. Today they end up as free text ("yes") or as a two-term vocabulary per property, both of which are awkward to enter, compare, and reconcile.
- **Wanted:** A `boolean` value type (true / false) alongside `text`, `integer`, `date`, `name`, `subject`, `term`, `depiction`, researcher-creatable like `text` and `integer`, with a checkbox / toggle editor in the citation composer and a clear display on Conclusion details.

## Code notes (observed, not implemented)

Where the value types are defined, and the layers a new one touches. `integer` is the closest precedent to copy from at every layer.

- **Shared constants (Go).** `core/propertymeta/propertymeta.go` — `ValueType*` constants; add `ValueTypeBoolean = "boolean"`.
- **Catalog schema.** `core/database/migrations/000052.sql` is the latest recreation of `properties` with `CHECK (value_type IN ('text', 'integer', …))`; SQLite cannot alter a CHECK, so a new migration recreates the table with the expanded set (see `000024.sql` for the `term` precedent). Observations need a value column: either a new `value_boolean INTEGER` on `observations` (and the matching conclusion-side table) or a convention of reusing `value_integer` 0/1 — a product decision; a dedicated column is more honest. Follow `.agents/skills/add-catalog-migration/SKILL.md` and register any column in `core/database/catalogmodel` (`add-catalog-model`).
- **Properties domain (Go).** `core/database/properties/properties.go` — `valueTypeOK` and the researcher-creatable switch (the comment says it mirrors Swift's `researcherCreatable`); change-type guard around line 220 if type changes stay forbidden.
- **Observation storage (Go).** `core/database/observations/observations.go` (`sqlCols`, resolve switch around `ValueTypeInteger`), `observations/write.go`, `observations/update.go`, `graphcache/load.go`, `graphfill/load.go`, `conclusiondetails/conclusiondetails.go`, `deleteimpact/facets.go` + `projectors.go` — every place that lists `value_integer` needs the sibling column.
- **Reconciliation / matching (Go).** `core/valuecompare/valuecompare.go` (comparer registry keyed by value type), `core/autoreconcile/autoreconcile.go` and `modules.go` (sort key / module eligibility per type), and the match registry if booleans should score.
- **FFI / proto.** `api/proto` `engine.proto` — the value-type comment lists allowed strings; `ConclusionValue` oneof (`text | integer | term | date | name | depiction` in the generated `engine.pb.swift`) needs a `bool` arm; handlers that map observation drafts and conclusion values. `api/ffi/swift_vocabulary_test.go` asserts the Swift enum matches `propertymeta`, so it fails until Swift is updated.
- **Swift vocabulary.** `macos/App/Platform/InterpretationValueKinds.swift` — `PropertyValueType` enum + `label`; `macos/App/Platform/L10n.swift` `Properties.valueTypeBoolean` (plus `Localizable.xcstrings`, per `add-localized-string`). `GenealogyStore.swift` `CatalogConclusionValue` gains a `.boolean(Bool)` case, and `CatalogObservation` / `CatalogObservationDraft` a `valueBoolean` field.
- **Swift editors / display.** Citation composer: `CitationObservationValue.swift` (draft field, `isValid`, `apply` switch on `PropertyValueType.integer.rawValue` is the template), `CitationComposerAuthoringBox.swift` (value-field switch ~line 378 — a `PVCheckbox` / toggle instead of `PVInput`), `CitationComposerVocabulary.swift` `supportedValueTypes`. Properties page: `PropertiesSnapshot.swift` / `PropertiesModel.swift` `researcherCreatable` so the type is offered when creating a Property. Conclusion detail: `Conclusions/ReconciledValueDisplay.swift` `.integer` case sibling. Search: `core/database/searchindex` if boolean values should be findable (probably not).
- **Docs.** `docs/seeded-vocabulary.md` §3.2 (`value_type` list) and `docs/interpretation-layer-data-model.md`.
