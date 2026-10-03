import SwiftUI

/// The left pane of `MetadataView` (F-1/F-2/F-4/F-5): the shared
/// `VocabularyListPane` chrome around this destination's columns — label
/// with origin pill, mono key, data-type badge — sorted by label only.
struct MetadataListPane: View {
    @Bindable var model: MetadataModel
    @Environment(WorkspaceNavigation.self) private var navigation

    /// Column widths live only here; `PVTable` shares them between the
    /// header and every row, so they are never restated.
    private static let keyColumnWidth: CGFloat = 160
    private static let dataTypeColumnWidth: CGFloat = 116

    private static let labelColumnID = "label"

    var body: some View {
        VocabularyListPane(
            totalCount: model.fields.count,
            rows: model.visibleFields,
            isLoading: model.isLoading,
            columns: columns,
            selection: selection,
            sort: PVTableSort(
                columnID: Self.labelColumnID,
                direction: model.sortAscending ? .ascending : .descending
            ),
            onSortChange: { _ in model.toggleLabelSort() },
            strings: VocabularyListStrings(
                tableLabel: L10n.Workspace.metadataTitle,
                emptyIcon: .tag,
                emptyTitle: L10n.Metadata.emptyProjectTitle,
                emptyBody: L10n.Metadata.emptyProjectBody,
                resultLine: { L10n.Metadata.resultLine(shown: $0, total: $1) }
            ),
            identifierPrefix: "metadata"
        )
    }

    private var columns: [PVTableColumn<CatalogMetadataField>] {
        [
            PVTableColumn(id: Self.labelColumnID, title: L10n.Metadata.columnLabel, sortable: true) { field in
                HStack(spacing: PVSpacing.space2) {
                    Text(field.label)
                        .font(PVFont.body(size: PVTypeScale.bodySmall))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    OriginPill(origin: field.origin)
                }
            },
            PVTableColumn(id: "key", title: L10n.Metadata.columnKey, width: Self.keyColumnWidth) { field in
                Text(field.key)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .truncationMode(.middle)
            },
            PVTableColumn(id: "dataType", title: L10n.Metadata.columnDataType, width: Self.dataTypeColumnWidth) { field in
                CatalogFieldDataTypeBadge(dataType: field.dataType)
            },
        ]
    }

    /// Selection is the table's; workspace history owns committed place.
    private var selection: Binding<String?> {
        Binding(
            get: { model.isAdding ? nil : model.selectedField?.id },
            set: { id in
                guard let id else { return }
                let field = model.fields.first(where: { $0.id == id })
                navigation.go(to: WorkspaceLocation(
                    section: .metadata,
                    fieldId: id,
                    title: field?.label
                ))
            }
        )
    }
}
