import SwiftUI

/// One Source in the Sources list: two lines. The first is the Source link
/// (cover, title, type), when it was last updated, and its ref; the second is
/// one token per Evidence graph of the Source, each opening that graph. A
/// Source with no Artifact shows one blocked token instead (S12-D5).
struct SourcesListRow: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(WorkspaceSession.self) private var session
    let source: CatalogSource
    let model: SourcesModel

    static let coverSize: CGFloat = 44
    /// The title's x inside the row: the link's padding, the cover and the gap
    /// after it, so the token line starts under the title.
    static let tokenIndent: CGFloat = PVSpacing.space2 + coverSize + PVSpacing.space4

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space4) {
            HStack(spacing: PVSpacing.space5) {
                sourceLink
                Spacer(minLength: 0)
                PVListDetailCell(
                    detail: PVLastUpdated.detail(source.updated.at, in: .current),
                    width: PVLastUpdated.columnWidth
                )
                Text(source.ref)
                    .font(PVFont.mono(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
                    .fixedSize(horizontal: true, vertical: false)
            }
            PVFlowLayout(spacing: PVSpacing.space3, lineSpacing: PVSpacing.space3) {
                ForEach(tokens, id: \.id) { token in
                    SourcesGraphToken(sourceID: source.id, kind: token.kind)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.leading, Self.tokenIndent)
        }
        .padding(.vertical, PVSpacing.space5)
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

    private var sourceLink: some View {
        Button {
            navigation.go(to: SourcesListNavigation.pageLocation(for: source))
        } label: {
            SourcesLinkLabel(
                projectDir: model.pageProjectDir,
                thumbnailRelPath: source.thumbnailRelPath,
                typeIconKey: model.typeIconKey(for: source),
                title: source.title,
                typeLabel: model.typeLabel(for: source)
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

    private struct TokenEntry {
        let id: String
        let kind: SourcesGraphToken.Kind
    }

    /// What a graph token does. `graphLocation` is nil only when
    /// `graphBlockedReason` is non-nil (both read `graph.hasArtifact`), so the
    /// last case cannot happen today; it keeps the token inert if they diverge.
    private func open(_ graph: CatalogEvidenceGraph) -> SourcesGraphToken.Open {
        if let reason = SourcesListNavigation.graphBlockedReason(for: graph) {
            return .blocked(reason: reason)
        }
        if let location = SourcesListNavigation.graphLocation(for: source, graph: graph) {
            return .opens { navigation.go(to: location) }
        }
        return .blocked(reason: L10n.string(L10n.Sources.graphNeedsArtifactTooltip))
    }

    /// The row's tokens: one per graph, or the one state that stands for them.
    private var tokens: [TokenEntry] {
        guard source.hasArtifact else {
            return [TokenEntry(id: "needsArtifact", kind: .needsArtifact)]
        }
        switch session.evidenceGraphsRead(sourceID: source.id) {
        case .loading:
            return [TokenEntry(id: "listLoading", kind: .listLoading)]
        case .failed(let error):
            return [TokenEntry(id: "listFailed", kind: .listFailed(error))]
        case .loaded(let graphs):
            let countsLoading = session.evidenceGraphCountsLoading
            return graphs.map { graph in
                TokenEntry(
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

/// The Source link's content: cover, title and type. Highlighted, the title
/// takes the link color and an underline and the cover an outline.
private struct SourcesLinkLabel: View {
    let projectDir: String
    let thumbnailRelPath: String
    let typeIconKey: String?
    let title: String
    let typeLabel: String

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
                    .underline(highlighted)
                    .lineLimit(1)
                if !typeLabel.isEmpty {
                    Text(verbatim: typeLabel)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
            }
        }
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
