import SwiftUI

/// One Source in the Sources list: three columns (S12-D5). The Source link (cover,
/// title, and a second line of ref and last update), its type, and a stack of
/// Evidence graph buttons, each carrying its own status badge. A Source with no
/// Artifact shows one blocked block instead.
struct SourcesListRow: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(WorkspaceSession.self) private var session
    let source: CatalogSource
    let model: SourcesModel
    let columns: SourcesListColumns

    static let coverSize: CGFloat = 40
    /// The Source cell's height: the first graph line, so the Source and Type
    /// centre on it (8 + a 52 pt button + 8).
    static let firstSlot: CGFloat = PVSpacing.space4 + SourcesGraphButton.height + PVSpacing.space4

    var body: some View {
        HStack(alignment: .top, spacing: 0) {
            sourceLink
            graphStack
        }
        .fixedSize(horizontal: false, vertical: true)
        .padding(.trailing, PVSpacing.gutterPage)
        .frame(maxWidth: .infinity, alignment: .leading)
        .overlay(alignment: .bottom) {
            PVDivider()
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: source.title))
        .accessibilityIdentifier("sources.row.\(source.id)")
    }

    /// One graph button per entry; the hairline is the stack's leading edge.
    private var graphStack: some View {
        VStack(spacing: SourcesListColumns.slotSpacing) {
            ForEach(entries, id: \.id) { entry in
                SourcesGraphButton(sourceID: source.id, kind: entry.kind)
                    .sourcesGraphCell(columns)
            }
        }
        .padding(.vertical, SourcesListColumns.stackVertical)
        .overlay(alignment: .leading) {
            PVDivider(axis: .vertical)
        }
    }

    private var sourceLink: some View {
        Button {
            navigation.go(to: SourcesListNavigation.pageLocation(for: source))
        } label: {
            HStack(alignment: .top, spacing: 0) {
                SourcesLinkLabel(
                    projectDir: model.pageProjectDir,
                    thumbnailRelPath: source.thumbnailRelPath,
                    typeIconKey: model.typeIconKey(for: source),
                    title: source.title,
                    typeLabel: showsRecordLine ? model.typeLabel(for: source) : "",
                    date: source.date,
                    showsRecordLine: showsRecordLine,
                    ref: source.ref,
                    updatedAt: source.updated.at
                )
                .padding(.vertical, SourcesListColumns.sourceVertical)
                .padding(.leading, PVSpacing.gutterPage)
                .sourcesSourceCell()
                typeAndDateCells
            }
            .frame(maxHeight: .infinity, alignment: .top)
        }
        .buttonStyle(PVListRowStyle(focused: nil))
        .accessibilityLabel(Text(L10n.Sources.openSourcePage))
        .accessibilityValue(Text(verbatim: PVLastUpdated.label(
            Self.spokenBase(
                title: source.title, type: model.typeLabel(for: source), ref: source.ref, date: source.date
            ),
            updatedAt: source.updated.at,
            in: .current
        )))
        .accessibilityIdentifier("sources.row.\(source.id).page")
        .pvRightClickMenu(
            accessibilityIdentifier: "sources.row.\(source.id).menu",
            items: [
                PVActionMenuItem(
                    title: L10n.Sources.copySourceEntry,
                    accessibilityIdentifier: "sources.row.\(source.id).copy",
                    action: { model.openCopy(source) }
                ),
            ]
        )
    }

    private var showsRecordLine: Bool { columns.datePlacement == .recordLine }

    /// What VoiceOver reads before the last update: the record's facts, then its ref.
    static func spokenBase(
        title: String, type: String, ref: String, date: CatalogDateValueInput?,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        L10n.DesignSystem.spokenJoin([title, type, SourcesRowDate.spoken(date, locale: locale), ref]) ?? title
    }

    private var typeText: some View {
        Text(verbatim: model.typeLabel(for: source))
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textSecondary)
            .lineLimit(1)
            .truncationMode(.tail)
            .help(model.typeLabel(for: source))
    }

    /// The Type cell, and at Wide the Published cell after it; none when compact.
    @ViewBuilder
    private var typeAndDateCells: some View {
        switch columns.datePlacement {
        case .underType:
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                typeText
                SourcesPublishedDate(date: source.date, style: .labeled)
            }
            .sourcesTypeCell(columns)
            .frame(minHeight: Self.firstSlot)
        case .column:
            typeText
                .sourcesTypeCell(columns)
                .frame(minHeight: Self.firstSlot)
            SourcesPublishedDate(date: source.date, style: .bare)
                .sourcesPublishedCell(columns)
                .frame(minHeight: Self.firstSlot)
        case .recordLine:
            EmptyView()
        }
    }

    private struct GraphEntry {
        let id: String
        let kind: SourcesGraphButton.Kind
    }

    /// What a graph button does. `graphLocation` is nil only when
    /// `graphBlockedReason` is non-nil (both read `graph.hasArtifact`), so the
    /// last case cannot happen today; it keeps the button inert if they diverge.
    private func open(_ graph: CatalogEvidenceGraph) -> SourcesGraphButton.Open {
        if let reason = SourcesListNavigation.graphBlockedReason(for: graph) {
            return .blocked(reason: reason)
        }
        if let location = SourcesListNavigation.graphLocation(for: source, graph: graph) {
            return .opens { navigation.go(to: location) }
        }
        return .blocked(reason: L10n.string(L10n.Sources.graphNeedsArtifactTooltip))
    }

    /// The row's entries: one per graph, or the one state that stands for them.
    private var entries: [GraphEntry] {
        guard source.hasArtifact else {
            return [GraphEntry(id: "needsArtifact", kind: .needsArtifact)]
        }
        switch session.evidenceGraphsRead(sourceID: source.id) {
        case .loading:
            return [GraphEntry(id: "listLoading", kind: .listLoading)]
        case .failed(let error):
            return [GraphEntry(id: "listFailed", kind: .listFailed(error))]
        case .loaded(let graphs):
            let countsLoading = session.evidenceGraphCountsLoading
            return graphs.map { graph in
                GraphEntry(
                    id: graph.id,
                    kind: .graph(
                        graph,
                        counts: SourcesGraphCounts(
                            progress: session.evidenceGraphProgress(graphID: graph.id),
                            countsLoading: countsLoading,
                            countsError: session.evidenceGraphCountsError
                        ),
                        open: open(graph)
                    )
                )
            }
        }
    }
}

// MARK: Source link

/// The Source link's content: cover, title, a record line of the type and the
/// published date (only when the Type column is gone), and a line of ref and last
/// update. The button's row style draws the hover, press and focus.
private struct SourcesLinkLabel: View {
    let projectDir: String
    let thumbnailRelPath: String
    let typeIconKey: String?
    let title: String
    /// Empty unless the record line shows (the list has no Type column).
    let typeLabel: String
    let date: CatalogDateValueInput?
    let showsRecordLine: Bool
    let ref: String
    let updatedAt: String

    var body: some View {
        HStack(spacing: PVSpacing.space4) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: thumbnailRelPath,
                typeIconKey: typeIconKey,
                size: SourcesListRow.coverSize
            )
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .fixedSize(horizontal: false, vertical: true)
                if showsRecordLine {
                    recordLine
                }
                activityLine
            }
            .frame(minHeight: SourcesListRow.coverSize)
        }
    }

    /// Type, then the published date; the type truncates first.
    private var recordLine: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            if !typeLabel.isEmpty {
                Text(verbatim: typeLabel)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .help(typeLabel)
                    .layoutPriority(0)
                separator
            }
            SourcesPublishedDate(date: date, style: .labeled)
                .layoutPriority(1)
        }
        .lineLimit(1)
    }

    /// Ref, then the last update.
    private var activityLine: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: ref)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
                .fixedSize(horizontal: true, vertical: false)
            separator
            PVListDetailCell(
                detail: PVLastUpdated.detail(updatedAt, in: .current),
                width: nil,
                iconSize: 12
            )
        }
        .lineLimit(1)
    }

    private var separator: some View {
        Text(verbatim: "·")
            .font(PVFont.body(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textFaint)
            .accessibilityHidden(true)
    }
}

// MARK: Published date

/// The text a row says about the Source's date, in one place: a date whose line is
/// empty counts as undated.
enum SourcesRowDate {
    static func line(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String? {
        guard let date else { return nil }
        let line = DateRowDisplay.line(date: date, locale: locale)
        return line.isEmpty ? nil : line
    }

    static func help(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String? {
        line(date, locale: locale).map(L10n.Sources.publishedHelp(date:))
    }

    static func spoken(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String {
        line(date, locale: locale).map(L10n.Sources.publishedSpoken(date:))
            ?? L10n.string(L10n.Sources.noPublishedDateSpoken)
    }
}

/// A Source's published date on a list row: `.labeled` ("Published" then the date,
/// or "No published date") under the type or on the record line; `.bare` (the date
/// alone, or "Undated") in the Published column.
private struct SourcesPublishedDate: View {
    enum Style { case labeled, bare }

    let date: CatalogDateValueInput?
    let style: Style

    var body: some View {
        if let line = SourcesRowDate.line(date) {
            dated(line)
        } else {
            undated
        }
    }

    @ViewBuilder
    private func dated(_ line: String) -> some View {
        switch style {
        case .labeled:
            HStack(spacing: PVSpacing.space2) {
                Text(L10n.Sources.publishedLabel)
                    .foregroundStyle(PVColor.textMuted)
                    .fixedSize()
                value(line)
            }
            .font(PVFont.body(size: PVTypeScale.caption))
            .help(SourcesRowDate.help(date) ?? "")
        case .bare:
            value(line)
                .font(PVFont.body(size: PVTypeScale.caption))
                .help(SourcesRowDate.help(date) ?? "")
        }
    }

    private func value(_ line: String) -> some View {
        Text(verbatim: line)
            .foregroundStyle(PVColor.textSecondary)
            .lineLimit(1)
            .truncationMode(.tail)
    }

    private var undated: some View {
        Group {
            switch style {
            case .labeled: Text(L10n.Sources.noPublishedDate)
            case .bare: Text(L10n.Dates.rowUndated)
            }
        }
        .font(PVFont.body(size: PVTypeScale.caption))
        .foregroundStyle(PVColor.textFaint)
        .lineLimit(1)
    }
}

// MARK: Status badge

extension SourcesModel.EvidenceStatus {
    var badgeTone: PVBadgeTone {
        switch self {
        case .empty: .neutral
        case .inProgress: .warning
        case .promoted: .success
        }
    }

    var badgeIcon: PVSymbol {
        switch self {
        case .empty: .circleDashed
        case .inProgress: .circleDot
        case .promoted: .circleCheck
        }
    }

    /// Only a graph with nothing on it draws an outline instead of a fill.
    var badgeSubtle: Bool { self == .empty }

    /// The status as a Sources list badge. Hidden from VoiceOver: the graph
    /// button's label carries the status.
    @MainActor var badge: some View {
        PVBadge(label, tone: badgeTone, icon: badgeIcon, subtle: badgeSubtle)
            .accessibilityHidden(true)
    }
}
