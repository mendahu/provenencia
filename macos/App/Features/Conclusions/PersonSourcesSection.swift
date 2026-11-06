import SwiftUI

/// The Person page's Sources section (S13-16, board S13-D6): the Sources the person's
/// records come from, one row each, with a line for every record on that Source's Evidence
/// graphs. It reads its own query (`ConclusionPageSection.queryKey`) and the Source types
/// list the place warms beside it, keeps the rows drawn while it reloads, opens a Source on
/// its page and a record on its graph, with that Subject selected.
struct PersonSourcesSection: View {
    let entityId: String

    @Environment(WorkspaceSession.self) private var session
    @Environment(WorkspaceNavigation.self) private var navigation

    /// The Sources whose lines the researcher has unfolded; another person's page is another view.
    @State private var expanded: Set<String> = []
    /// Whether the Sources past the fold are shown. Starts folded on every mount, so a restored section opens folded.
    @State private var rowsExpanded = false

    var body: some View {
        let orders = session.nameFormat
        let expanded = expanded
        let rowsExpanded = rowsExpanded
        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(
            CatalogQueryKey.sourceTypesList(project: session.projectKey)
        )
        ConclusionPageSectionHost<[CatalogPersonClaimGroup], PersonSourcesContent, PersonSourcesList>(
            section: .sources,
            entityId: entityId,
            errorTitle: L10n.Conclusions.sourcesErrorTitle,
            emptyTitle: L10n.Conclusions.sourcesEmptyTitle,
            emptyMessage: L10n.string(L10n.Conclusions.sourcesEmptyMessage),
            make: {
                PersonSourcesContent(
                    groups: $0,
                    sourceTypes: typesHandle?.value ?? [],
                    orders: orders,
                    expanded: expanded,
                    rowsExpanded: rowsExpanded
                )
            },
            isEmpty: \.isEmpty
        ) { content in
            PersonSourcesList(
                content: content,
                open: { navigation.go(to: $0) },
                expand: { self.expanded.insert($0) },
                expandRows: { self.rowsExpanded = true }
            )
        }
    }
}

/// The widths the header line and every Source row share: Citation (fills) · Type · Date ·
/// Ref, with one cell modifier per column so they cannot drift.
enum PersonSourcesColumns {
    static let spacing = PVProvenanceColumns.spacing
    static let typeWidth: CGFloat = 176
    static let dateWidth: CGFloat = 120
    static let refWidth: CGFloat = 104
    /// A graph line's indent under the title.
    static let lineIndent: CGFloat = 20
}

private extension View {
    func personSourcesCitationCell() -> some View {
        frame(maxWidth: .infinity, alignment: .leading)
    }

    func personSourcesTypeCell() -> some View {
        frame(width: PersonSourcesColumns.typeWidth, alignment: .leading)
    }

    func personSourcesDateCell() -> some View {
        frame(width: PersonSourcesColumns.dateWidth, alignment: .leading)
    }

    func personSourcesRefCell() -> some View {
        frame(width: PersonSourcesColumns.refWidth, alignment: .leading)
    }
}

/// The header line and the Source rows.
struct PersonSourcesList: View {
    let content: PersonSourcesContent
    let open: (WorkspaceLocation) -> Void
    let expand: (String) -> Void
    let expandRows: () -> Void

    /// The graph link with the keyboard.
    @FocusState private var focusedLine: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
            PVDivider(color: PVColor.borderStrong)
            ForEach(Array(content.rows.enumerated()), id: \.element.id) { index, row in
                if index > 0 { PVDivider() }
                PersonSourcesRowView(row: row, open: open, expand: { unfold(row) }, focused: $focusedLine)
            }
            if content.hiddenRowCount > 0 {
                ConclusionShowMoreRow(count: content.hiddenRowCount) {
                    expandRows()
                    Task { @MainActor in focusedLine = content.firstRevealedRowLineID }
                }
                .padding(.vertical, PVSpacing.space3)
                .accessibilityIdentifier("persons.detail.sources.more")
            }
        }
    }

    /// Unfolds a Source and, once its lines are drawn, focuses the first one it added.
    private func unfold(_ row: PersonSourcesContent.Row) {
        expand(row.id)
        Task { @MainActor in focusedLine = row.firstRevealedLineID }
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PersonSourcesColumns.spacing) {
            Text(L10n.Conclusions.sourcesColumnSource).personSourcesCitationCell()
            Text(L10n.Conclusions.sourcesColumnType).personSourcesTypeCell()
            Text(L10n.Conclusions.sourcesColumnDate).personSourcesDateCell()
            Text(L10n.Conclusions.sourcesColumnRef).personSourcesRefCell()
        }
        .pvMicroCaps()
        .foregroundStyle(PVColor.textMuted)
        .padding(.vertical, PVSpacing.space2)
        .accessibilityHidden(true)
    }
}

/// One Source: its title as a link, its type, date and ref, and under the title a line per record.
private struct PersonSourcesRowView: View {
    let row: PersonSourcesContent.Row
    let open: (WorkspaceLocation) -> Void
    let expand: () -> Void
    var focused: FocusState<String?>.Binding

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PersonSourcesColumns.spacing) {
            citation.personSourcesCitationCell()
            typeCell.personSourcesTypeCell().accessibilityHidden(true)
            Text(L10n.Conclusions.sourcesDateNone)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .personSourcesDateCell()
                .accessibilityHidden(true)
            Text(verbatim: row.ref)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .textSelection(.enabled)
                .personSourcesRefCell()
                .accessibilityHidden(true)
        }
        // The row as one element; the title link and the lines follow as controls.
        .overlay(alignment: .topLeading) {
            Color.clear.frame(width: 1, height: 1)
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(row.accessibilityLabel)
        }
        .padding(.vertical, PVSpacing.space3)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("persons.detail.sources.row.\(row.id)")
    }

    private var citation: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            PVInkLink(ink: .primary, action: { open(row.location) }) {
                PVTitleText(row.title, style: .rowLink, lineLimit: 2)
            }
            .accessibilityLabel(Text(verbatim: L10n.Conclusions.sourcesOpenSource(row.helpTitle)))
            .accessibilityIdentifier("persons.detail.sources.title.\(row.id)")
            .help(row.helpTitle)
            VStack(alignment: .leading, spacing: PVSpacing.space2) {
                ForEach(row.lines) { line in
                    PersonSourcesLineView(line: line, open: open, focused: focused)
                }
                if row.hiddenCount > 0 {
                    ConclusionShowMoreRow(count: row.hiddenCount, expand: expand)
                        .accessibilityIdentifier("persons.detail.sources.more.\(row.id)")
                }
            }
            .padding(.leading, PersonSourcesColumns.lineIndent)
        }
    }

    @ViewBuilder
    private var typeCell: some View {
        if let label = row.typeLabel {
            CatalogSourceTypePill(label: label, iconKey: row.typeIconKey)
        } else {
            Color.clear.frame(height: 1)
        }
    }
}

/// One record: the graph's name as a link, then "· graph ref · Subject name · Subject ref".
/// Every part stays on one line; a line too narrow cuts the Subject's name first.
private struct PersonSourcesLineView: View {
    let line: PersonSourcesContent.Line
    let open: (WorkspaceLocation) -> Void
    var focused: FocusState<String?>.Binding

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 0) {
            PVButton(line.graphTitle, variant: .link, size: .sm) { open(line.location) }
                .lineLimit(1)
                .truncationMode(.tail)
                .help(line.graphTitle)
                .focused(focused, equals: line.id)
                .pvActivatesOnReturn(when: focused.wrappedValue == line.id) { open(line.location) }
                .accessibilityLabel(line.accessibilityLabel)
                .accessibilityIdentifier("persons.detail.sources.graph.\(line.id)")
            Group {
                if let graphRef = line.graphRef {
                    Text(verbatim: " · ")
                    Text(verbatim: graphRef).font(PVFont.mono(size: PVTypeScale.caption))
                        .lineLimit(1).fixedSize(horizontal: true, vertical: false)
                }
                Text(verbatim: " · ")
                Text(verbatim: line.subjectName)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .layoutPriority(-1)
                    .help(line.subjectName)
                Text(verbatim: " · ")
                Text(verbatim: line.subjectRef).font(PVFont.mono(size: PVTypeScale.caption))
                    .lineLimit(1).fixedSize(horizontal: true, vertical: false)
            }
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textSecondary)
            .accessibilityHidden(true)
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("persons.detail.sources.line.\(line.id)")
    }
}

#if DEBUG
/// The Sources section over a fixture, with its folds working.
private struct PersonSourcesPreview: View {
    let groups: [CatalogPersonClaimGroup]
    var expandedAtStart: Set<String> = []

    @State private var expanded: Set<String> = []
    @State private var rowsExpanded = false

    var body: some View {
        PVPageColumns(primaryFill: true) {
            PersonSourcesList(
                content: PersonSourcesContent(
                    groups: groups,
                    sourceTypes: PersonSourcesPreviewData.types,
                    orders: NameFormatOrders(),
                    expanded: expanded,
                    rowsExpanded: rowsExpanded
                ),
                open: { _ in },
                expand: { expanded.insert($0) },
                expandRows: { rowsExpanded = true }
            )
        }
        .frame(width: 1102, height: 700)
        .background(PVColor.surfacePage)
        .onAppear { expanded = expandedAtStart }
    }
}

#Preview("Sources") { PersonSourcesPreview(groups: PersonSourcesPreviewData.regular) }
#Preview("Untitled Source") { PersonSourcesPreview(groups: PersonSourcesPreviewData.untitled) }
#Preview("Folded") { PersonSourcesPreview(groups: PersonSourcesPreviewData.folded) }
#Preview("Many Sources") { PersonSourcesPreview(groups: PersonSourcesPreviewData.manySources) }
#endif
