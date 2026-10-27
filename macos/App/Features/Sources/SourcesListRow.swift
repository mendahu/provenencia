import SwiftUI

/// One Source in the Sources list: four columns (S12-D5). The Source link (cover,
/// title, and a second line of ref and last update), its type, and one line per
/// Evidence graph holding that graph's status badge and its button. A Source with
/// no Artifact shows one blocked block instead.
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

    /// One slot per entry, so each badge sits level with its graph button.
    private var statusStack: some View {
        VStack(spacing: SourcesListColumns.slotSpacing) {
            ForEach(entries, id: \.id) { entry in
                statusSlot(entry.kind)
                    .frame(height: SourcesGraphButton.height)
                    .sourcesStatusCell(columns)
            }
        }
        .padding(.vertical, SourcesListColumns.stackVertical)
    }

    /// One graph button per entry; the hairline is the stack's leading edge.
    private var graphStack: some View {
        VStack(spacing: SourcesListColumns.slotSpacing) {
            ForEach(entries, id: \.id) { entry in
                SourcesGraphButton(sourceID: source.id, kind: entry.kind)
                    .frame(height: SourcesGraphButton.height)
                    .sourcesGraphCell(columns)
            }
        }
        .padding(.vertical, SourcesListColumns.stackVertical)
        .overlay(alignment: .leading) {
            PVDivider(axis: .vertical)
        }
    }

    /// The graph's status badge, a rule while its counts load, or a faint dash
    /// where there is no status to show.
    @ViewBuilder
    private func statusSlot(_ kind: SourcesGraphButton.Kind) -> some View {
        if case .graph(_, let counts, _) = kind, let status = counts.status {
            status.badge
        } else if case .graph(_, .loading, _) = kind {
            PVSkeletonBar(width: 72, height: 6)
                .accessibilityHidden(true)
        } else {
            Text(verbatim: "—")
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textFaint)
                .accessibilityHidden(true)
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
                    typeLabel: columns.isCompact ? model.typeLabel(for: source) : "",
                    ref: source.ref,
                    updatedAt: source.updated.at
                )
                .padding(.vertical, SourcesListColumns.sourceVertical)
                .padding(.leading, PVSpacing.gutterPage)
                .sourcesSourceCell()
                if columns.type != nil {
                    Text(verbatim: model.typeLabel(for: source))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(1)
                        .truncationMode(.tail)
                        .help(model.typeLabel(for: source))
                        .sourcesTypeCell(columns)
                        .frame(minHeight: Self.firstSlot)
                }
                statusStack
            }
            .frame(maxHeight: .infinity, alignment: .top)
        }
        .buttonStyle(PVListRowStyle(focused: nil))
        .accessibilityLabel(Text(L10n.Sources.openSourcePage))
        .accessibilityValue(Text(verbatim: PVLastUpdated.label(
            "\(source.title), \(model.typeLabel(for: source)), \(source.ref)",
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

/// The Source link's content: cover, title, and a second line of the type (only
/// when the Type column is gone), ref and last update. The button's row style
/// draws the hover, press and focus.
private struct SourcesLinkLabel: View {
    let projectDir: String
    let thumbnailRelPath: String
    let typeIconKey: String?
    let title: String
    /// Empty when the list has a Type column.
    let typeLabel: String
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
                    .lineLimit(1)
                    .help(title)
                secondLine
            }
            .frame(minHeight: SourcesListRow.coverSize)
        }
    }

    /// Type (compact lists only, truncates first), ref, then the last update.
    private var secondLine: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            if !typeLabel.isEmpty {
                Text(verbatim: typeLabel)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .help(typeLabel)
                separator
            }
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
