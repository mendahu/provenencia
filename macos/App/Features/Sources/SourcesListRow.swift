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
                .padding(.trailing, SourcesListColumns.sourceTrailing)
                .frame(maxWidth: .infinity, minHeight: Self.firstSlot, alignment: .leading)
            if let typeWidth = columns.type {
                Text(verbatim: model.typeLabel(for: source))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .padding(.trailing, SourcesListColumns.cellTrailing)
                    .frame(width: typeWidth, alignment: .leading)
                    .frame(minHeight: Self.firstSlot)
            }
            graphStack
        }
        .padding(.leading, PVSpacing.gutterPage - PVSpacing.space2)
        .padding(.trailing, PVSpacing.gutterPage)
        .frame(maxWidth: .infinity, alignment: .leading)
        .overlay(alignment: .bottom) {
            PVDivider()
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: source.title))
        .accessibilityIdentifier("sources.row.\(source.id)")
    }

    /// One line per entry, status slot then button, so each badge sits level with
    /// its button. The hairline between the columns is drawn once.
    private var graphStack: some View {
        VStack(spacing: PVSpacing.space3) {
            ForEach(entries, id: \.id) { entry in
                HStack(spacing: 0) {
                    statusSlot(entry.kind)
                        .padding(.trailing, SourcesListColumns.cellTrailing)
                        .frame(width: columns.status, alignment: .leading)
                    SourcesGraphButton(sourceID: source.id, kind: entry.kind)
                        .frame(width: columns.graphs)
                        .padding(.leading, SourcesListColumns.graphLeading)
                }
            }
        }
        .padding(.vertical, PVSpacing.space4)
        .overlay(alignment: .leading) {
            PVDivider(axis: .vertical)
                .padding(.leading, columns.status)
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
            SourcesLinkLabel(
                projectDir: model.pageProjectDir,
                thumbnailRelPath: source.thumbnailRelPath,
                typeIconKey: model.typeIconKey(for: source),
                title: source.title,
                typeLabel: columns.isCompact ? model.typeLabel(for: source) : "",
                ref: source.ref,
                updatedAt: source.updated.at
            )
        }
        .buttonStyle(SourcesLinkButtonStyle())
        .accessibilityLabel(Text(L10n.Sources.openSourcePage))
        .accessibilityValue(Text(verbatim: PVLastUpdated.label(
            "\(source.title), \(source.ref)",
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

private struct SourcesLinkHighlightedKey: EnvironmentKey {
    static let defaultValue = false
}

private extension EnvironmentValues {
    /// True while the pointer is on the Source link or it is pressed.
    var sourcesLinkHighlighted: Bool {
        get { self[SourcesLinkHighlightedKey.self] }
        set { self[SourcesLinkHighlightedKey.self] = newValue }
    }
}

/// The Source link's content: cover, title, and a second line of the type (only
/// when the Type column is gone), ref and last update. Highlighted, the title
/// takes the link color and an underline and the cover an outline.
private struct SourcesLinkLabel: View {
    let projectDir: String
    let thumbnailRelPath: String
    let typeIconKey: String?
    let title: String
    /// Empty when the list has a Type column.
    let typeLabel: String
    let ref: String
    let updatedAt: String

    @Environment(\.sourcesLinkHighlighted) private var highlighted

    var body: some View {
        HStack(spacing: PVSpacing.space4) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: thumbnailRelPath,
                typeIconKey: typeIconKey,
                size: SourcesListRow.coverSize
            )
            .overlay {
                if highlighted {
                    RoundedRectangle(cornerRadius: PVRadius.sm)
                        .strokeBorder(PVColor.borderStrong, lineWidth: 1)
                }
            }
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: title)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                    .foregroundStyle(highlighted ? PVColor.textLinkHover : PVColor.textPrimary)
                    // Always underlined, only the color changes: toggling the
                    // underline under the hover animation crossfades two text
                    // layouts and shows two underlines mid-fade.
                    .underline(true, color: highlighted ? PVColor.textLinkHover : .clear)
                    .lineLimit(1)
                secondLine
            }
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

/// Kit hover, press and focus for the Source link, and the highlight state its
/// label reads. A style cannot recolor a label that sets its own color, so the
/// label takes the state from the environment.
private struct SourcesLinkButtonStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        SourcesLinkButtonBody(configuration: configuration)
    }
}

private struct SourcesLinkButtonBody: View {
    let configuration: ButtonStyleConfiguration

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            configuration.label
                .environment(\.sourcesLinkHighlighted, showHover || configuration.isPressed)
                .padding(PVSpacing.space2)
                .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
        }
    }
}
