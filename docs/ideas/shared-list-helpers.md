# Shared list helpers

**Status:** idea only — not roadmapped. Small duplication found while planning S10-41 (the Sources list filters). Each item is a cleanup, not a feature.

## In-list text matching

Three filter-as-you-type lists use three rules:

- `PropertiesModel.swift` (search field): `lowercased()` then `contains`. Misses accents: `jose` does not find `José`.
- `GraphCanvasItemLayer.swift`: `label.lowercased().contains(filter)`.
- S10-41 `SourcesModel.visibleSources`: `localizedStandardContains` (case- and accent-insensitive).

`PVComboBoxMatch.normalize` in `DesignSystem/Components/ComboBox/PVComboBox.swift` already folds case and diacritics. One helper would give every list the same rule and fix the Properties accent miss. Properties also draws its own “no results” state instead of `PVEmptyState`.

## Primary Subject kinds

`person`, `event`, `place` is written out in several places: `promote.primaryKinds` (`core/database/promote/promote.go`), `core/match/registry.go` `ProfiledKinds`, `core/database/promotealign/layer.go`, and four literals in `macos/App/Platform/FakeStore.swift` (S10-41 adds `FakeStore.primaryTypeKeys` for its own site). A home in `core/database/subjecttypes` would also let `sourcegraphprogress` drop its import of `promote` (a large dependency for a two-line check).

## Source graph mutations

`WorkspaceSession.refreshSourceGraphProgress(for:)` switches over the graph mutations by hand; `CatalogQueryRegistry` lists the same ones for `.sourceGraph`. After S10-41 the lists differ only by `.deletedSource`, which would also be correct to patch (`sourcegraphprogress.Get` returns zeros for a gone Source, and `isZero` removes the entry). A `CatalogMutation.graphSourceId` read by both would keep them in step.
