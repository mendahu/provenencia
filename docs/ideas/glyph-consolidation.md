# Glyph consolidation

**Status:** idea, 2026-10-14. Not scheduled. Raised in the S10-19 plan review ([`S10-19-event-mark-sidebar.review.md`](../deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.review.md), M-3 and M-4).

S10-19 adds `PVGlyph` (an SF Symbol or a mark) and `PVGlyphView` in `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`. The sidebar, the history menus, and `PVEmptyState` use it. Two clean-ups are left.

## Older symbol-or-mark branches

These sites still choose between a mark and a symbol by hand:

- `PVThumbnail.Content`: `mark: PVMarkKey?` beside `icon: PVSymbol?`, with the mark winning.
- `PropertiesView.typeCard`: `if let mark { PVMark } else { Image(systemName:) }`.
- `OmnibarResultsHost`: a `switch` on hit kind that builds thumbnail content from a mark or a symbol.

The vocabulary panes' empty states also drift from their sidebar rows. Source types uses `.library` (`SourceTypesListPane`) where its row uses `.tag`, and Metadata uses `.tag` (`MetadataListPane`) where its row uses `.list`. They could read `WorkspaceSection.glyph`.

Moving them to `PVGlyph` leaves one shape for one concept. `PVThumbnail.Content` is the largest of these, because rows across the app build it.

## One subject-type → mark mapping

The Go registry (`core/database/subjectvocab/registry.go`) declares `IconSymbol` (`subject_person`, `subject_event`, …) for every subject type. Swift receives it as `CatalogSubjectTypePresentation.iconSymbol`, and nothing reads it. Swift repeats the mapping by hand in `PropertiesTypeChrome.stripMarkKey(typeKey:)`, `OmnibarHitPresentation.leadMark(for:)`, `EvidencePrimaryKind.markKey` / `EvidenceBridgeKind.markKey`, the three `ConclusionListKind.mark` implementations, and the detail views. One `PVMarkKey(subjectTypeKey:)` would replace those switches, so a new subject type would get its mark from its registry entry. The workspace sidebar still needs a static mapping, because it draws before any catalog read.
