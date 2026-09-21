import SwiftUI

/// The left pane of `SourceTypesView` (T-1/T-2/T-3/T-4): the shared
/// `VocabularyListPane` chrome around this destination's columns — label
/// with origin pill, mono key, suggested-field count — sortable by all three.
struct SourceTypesListPane: View {
    @Bindable var model: SourceTypesModel
    @Environment(WorkspaceNavigation.self) private var navigation

    /// Column widths live only here; `PVTable` shares them between the
    /// header and every row, so they are never restated.
    private static let keyColumnWidth: CGFloat = 160
    private static let fieldsColumnWidth: CGFloat = 132

    var body: some View {
        VocabularyListPane(
            totalCount: model.types.count,
            rows: model.visibleTypes,
            isLoading: model.isLoading,
            columns: columns,
            selection: selection,
            sort: PVTableSort(
                columnID: model.sortColumn.rawValue,
                direction: model.sortAscending ? .ascending : .descending
            ),
            onSortChange: { model.sortBy($0) },
            strings: VocabularyListStrings(
                tableLabel: L10n.Workspace.sourceTypesTitle,
                emptyIcon: .library,
                emptyTitle: L10n.SourceTypes.emptyProjectTitle,
                emptyBody: L10n.SourceTypes.emptyProjectBody,
                resultLine: { L10n.SourceTypes.resultLine(shown: $0, total: $1) }
            ),
            identifierPrefix: "sourceTypes"
        )
    }

    private var columns: [PVTableColumn<CatalogSourceType>] {
        [
            PVTableColumn(
                id: SourceTypesModel.SortColumn.label.rawValue,
                title: L10n.SourceTypes.columnLabel,
                sortable: true
            ) { type in
                HStack(spacing: PVSpacing.space2) {
                    PVMark(
                        PVMarkKey(catalogKey: type.iconKey),
                        size: .inline,
                        decorative: true
                    )
                    Text(type.label)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    OriginPill(origin: type.origin)
                }
            },
            PVTableColumn(
                id: SourceTypesModel.SortColumn.key.rawValue,
                title: L10n.SourceTypes.columnKey,
                width: Self.keyColumnWidth,
                sortable: true
            ) { type in
                Text(type.key)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .truncationMode(.middle)
            },
            PVTableColumn(
                id: SourceTypesModel.SortColumn.fields.rawValue,
                title: L10n.SourceTypes.columnFields,
                width: Self.fieldsColumnWidth,
                sortable: true
            ) { type in
                Text(
                    type.suggestedFieldCount > 0
                        ? "\(type.suggestedFieldCount)"
                        : String(localized: L10n.SourceTypes.fieldCountNone)
                )
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(type.suggestedFieldCount > 0 ? PVColor.textSecondary : PVColor.textFaint)
            },
        ]
    }

    /// Selection is the table's; workspace history owns committed place.
    private var selection: Binding<String?> {
        Binding(
            get: { model.isAdding ? nil : model.selectedType?.id },
            set: { id in
                guard let id else { return }
                let type = model.types.first(where: { $0.id == id })
                navigation.go(to: WorkspaceLocation(
                    section: .sourceTypes,
                    typeId: id,
                    title: type?.label
                ))
            }
        )
    }
}
