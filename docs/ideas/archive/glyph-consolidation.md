# Glyph consolidation (archived)

**Status:** **Done** in S10-49. The leftovers below are not scheduled. Raised in the S10-19 plan review ([`S10-19-event-mark-sidebar.review.md`](../../deployment-plan/spike-10/prs/S10-19-event-mark-sidebar.review.md), M-3 and M-4).

S10-19 added `PVGlyph` (an SF Symbol or a mark) and `PVGlyphView` in `macos/App/DesignSystem/Recipes/Marks/PVMark.swift`. The sidebar, the history menus, and `PVEmptyState` used it. Two clean-ups were left, and S10-49 did them.

## Older symbol-or-mark branches

These sites chose between a mark and a symbol by hand:

- `PVThumbnail.Content`: `mark: PVMarkKey?` beside `icon: PVSymbol?`, with the mark winning.
- `PropertiesView.typeCard`: `if let mark { PVMark } else { Image(systemName:) }`.
- `OmnibarResultsHost`: a `switch` on hit kind that builds thumbnail content from a mark or a symbol.

The vocabulary panes' empty states also drifted from their sidebar rows. Source types used `.library` (`SourceTypesListPane`) where its row uses `.tag`, and Metadata used `.tag` (`MetadataListPane`) where its row uses `.list`. They now read `WorkspaceSection.glyph`.

Moving them to `PVGlyph` left one shape for one concept. `PVThumbnail.Content` was the largest of these, because rows across the app build it.

## One subject-type → mark mapping

The Go registry (`core/database/subjectvocab/registry.go`) declares `IconSymbol` (`subject_person`, `subject_event`, …) for every subject type. Swift receives it as `CatalogSubjectTypePresentation.iconSymbol`, and nothing read it. Swift repeated the mapping by hand in `PropertiesTypeChrome.stripMarkKey(typeKey:)`, `OmnibarHitPresentation.leadMark(for:)`, `EvidencePrimaryKind.markKey` / `EvidenceBridgeKind.markKey`, the three `ConclusionListKind.mark` implementations, and the detail views. One `PVMarkKey(subjectTypeKey:)` replaced those switches, so a new subject type gets its mark from its registry entry. The workspace sidebar still needs a static mapping, because it draws before any catalog read.

## Leftovers

A follow-up round on S10-49 did the rest. The omnibar lead symbols for Source types and Metadata fields now read `WorkspaceSection.glyph`. `PropertiesTypeChrome.sortIndex` and `showsBridgeLabel` are gone: Properties reads `paletteSort` and `role` from the presentation, and `ink` mutes on that role. `CitationComposerAuthoring.typeRank` reads `EvidencePrimaryKind`.

One stays: `FakeStore.syntheticPresentation` switches over type keys on purpose. It is the test double for the Go registry, so its table is the fake's data.
