import SwiftUI

/// The copy a `VocabularyListPane` renders — all of it owned by the feature's
/// L10n namespace, so the two destinations stay independently localizable.
struct VocabularyListStrings {
    /// Accessibility label for the table itself.
    let tableLabel: LocalizedStringResource
    let emptyIcon: PVSymbol
    let emptyTitle: LocalizedStringResource
    let emptyBody: LocalizedStringResource
    let resultLine: (_ shown: Int, _ total: Int) -> String
}

/// The left pane both vocabulary destinations share: a `PVTable` of the rows
/// with a result-count footer. Find lives in the workspace omnibar (S3-10).
struct VocabularyListPane<Row: CatalogVocabularyRow>: View {
    /// The vocabulary size — decides empty-project vs loaded table.
    let totalCount: Int
    /// Already sorted by the model.
    let rows: [Row]
    let isLoading: Bool
    let columns: [PVTableColumn<Row>]
    let selection: Binding<String?>
    let sort: PVTableSort
    let onSortChange: (String) -> Void
    let strings: VocabularyListStrings
    /// Prefixes every identifier this pane mints: `.list`, `.row.<id>`,
    /// `.sortBy.<column>`.
    let identifierPrefix: String

    var body: some View {
        VStack(spacing: 0) {
            table
            footer
        }
    }

    private var table: some View {
        PVTable(
            rows: rows,
            columns: columns,
            selection: selection,
            primaryText: \.label,
            sort: sort,
            onSortChange: onSortChange,
            label: strings.tableLabel,
            rowAccessibilityIdentifier: { "\(identifierPrefix).row.\($0.id)" },
            sortAccessibilityIdentifier: { "\(identifierPrefix).sortBy.\($0)" }
        ) {
            placeholder
        }
        .accessibilityIdentifier("\(identifierPrefix).list")
    }

    /// Loading / empty chrome, rendered inside the table body so the column
    /// header stays put — the same slot the web board uses.
    @ViewBuilder
    private var placeholder: some View {
        if isLoading && totalCount == 0 {
            ProgressView()
                .tint(PVColor.accent)
                .frame(maxWidth: .infinity)
                .padding(PVSpacing.space9)
        } else if totalCount == 0 {
            PVEmptyState(
                icon: strings.emptyIcon,
                title: strings.emptyTitle,
                message: String(localized: strings.emptyBody)
            )
            .padding(PVSpacing.space9)
        }
    }

    private var footer: some View {
        HStack {
            Text(strings.resultLine(rows.count, totalCount))
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .padding(.vertical, PVSpacing.space3)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .top) {
            PVDivider()
        }
    }
}
