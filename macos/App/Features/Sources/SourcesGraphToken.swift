import SwiftUI

/// One Evidence graph (or the reason there is none to show) in a Sources list
/// row's token line: name and compact counts, opening the graph's canvas. One
/// view owns every token state; the states are `Kind`.
struct SourcesGraphToken: View {
    enum Kind {
        /// A graph of the Source: opens its canvas, or is blocked (dashed, inert,
        /// "needs an artifact") with the reason as its help and hint.
        case graph(CatalogEvidenceGraph, counts: SourcesGraphCounts, open: Open)
        /// The Source has no Artifact, so it has nothing to build a graph from.
        case needsArtifact
        /// The graph list has not loaded.
        case listLoading
        /// The graph list failed to load.
        case listFailed(Error)
    }

    enum Open {
        case opens(() -> Void)
        case blocked(reason: String)
    }

    let sourceID: String
    let kind: Kind

    private static let maxNameWidth: CGFloat = 300

    var body: some View {
        switch kind {
        case .graph(let graph, let counts, let open):
            graphToken(graph, counts: counts, open: open)
        case .needsArtifact:
            needsArtifactToken
        case .listLoading:
            SourcesGraphTokenFrame(look: .live) {
                PVSkeletonBar(width: 96, height: 6)
            }
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(L10n.DesignSystem.listLoading))
            .accessibilityIdentifier("sources.row.\(sourceID).graphsLoading")
        case .listFailed(let error):
            failedToken(error)
        }
    }

    @ViewBuilder
    private func graphToken(
        _ graph: CatalogEvidenceGraph,
        counts: SourcesGraphCounts,
        open: Open
    ) -> some View {
        switch open {
        case .blocked(let blockedReason):
            let label = L10n.Sources.graphNeedsArtifact(graph.name)
            Button {} label: {
                HStack(spacing: PVSpacing.space3) {
                    PVIcon(.network, size: 13)
                    Text(verbatim: label)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .lineLimit(1)
                        .frame(maxWidth: Self.maxNameWidth, alignment: .leading)
                }
                .foregroundStyle(PVColor.textFaint)
            }
            .buttonStyle(SourcesGraphTokenButtonStyle(look: .blocked))
            .help(L10n.Sources.graphTokenHelp(name: graph.name, detail: blockedReason))
            .accessibilityLabel(Text(verbatim: label))
            .accessibilityHint(Text(verbatim: blockedReason))
            .accessibilityIdentifier("sources.row.\(sourceID).graph.\(graph.ref)")
        case .opens(let onOpen):
            Button {
                onOpen()
            } label: {
                HStack(spacing: PVSpacing.space3) {
                    PVIcon(.network, size: 13)
                        .foregroundStyle(PVColor.textMuted)
                    Text(verbatim: graph.name)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                        .truncationMode(.tail)
                        .frame(maxWidth: Self.maxNameWidth, alignment: .leading)
                    countView(counts)
                }
            }
            .buttonStyle(SourcesGraphTokenButtonStyle(look: .live))
            .help(L10n.Sources.graphTokenHelp(name: graph.name, detail: L10n.Sources.graphCounts(counts, style: .full)))
            .accessibilityLabel(Text(verbatim:
                "\(L10n.Sources.openEvidenceGraphLabel(name: graph.name)), \(L10n.Sources.graphCounts(counts, style: .spoken))"
            ))
            .accessibilityIdentifier("sources.row.\(sourceID).graph.\(graph.ref)")
        }
    }

    @ViewBuilder
    private func countView(_ counts: SourcesGraphCounts) -> some View {
        switch counts {
        case .counts:
            Text(verbatim: L10n.Sources.graphCounts(counts, style: .compact))
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
        case .notStarted:
            Text(verbatim: L10n.Sources.graphCounts(counts, style: .compact))
                .font(PVFont.body(size: PVTypeScale.micro))
                .italic()
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
        case .loading:
            PVSkeletonBar(width: 32, height: 6)
        case .unavailable:
            PVIcon(.danger, size: 11)
                .foregroundStyle(PVColor.danger)
        }
    }

    private var needsArtifactToken: some View {
        Button {} label: {
            HStack(spacing: PVSpacing.space3) {
                PVIcon(.network, size: 13)
                Text(L10n.Sources.needsArtifact)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .italic()
                    .lineLimit(1)
            }
            .foregroundStyle(PVColor.textFaint)
        }
        .buttonStyle(SourcesGraphTokenButtonStyle(look: .blocked))
        .help(L10n.string(L10n.Sources.needsArtifactTooltip))
        .accessibilityLabel(Text(L10n.Sources.needsArtifact))
        .accessibilityHint(Text(L10n.Sources.needsArtifactTooltip))
        .accessibilityIdentifier("sources.row.\(sourceID).graphBlocked")
    }

    private func failedToken(_ error: Error) -> some View {
        let message = L10n.Errors.message(for: error)
        return Button {} label: {
            PVIcon(.danger, size: 13)
                .foregroundStyle(PVColor.danger)
        }
        .buttonStyle(SourcesGraphTokenButtonStyle(look: .inert))
        .help(L10n.Sources.graphTokenHelp(name: L10n.string(L10n.Sources.graphsUnavailable), detail: message))
        .accessibilityLabel(Text(L10n.Sources.graphsUnavailable))
        .accessibilityHint(Text(verbatim: message))
        .accessibilityIdentifier("sources.row.\(sourceID).graphsError")
    }
}

/// How a token looks: `live` opens a graph (or is a placeholder), `blocked` is
/// dashed and takes focus but no hover fill, `inert` (the list failed) fills
/// nothing at all.
private enum SourcesGraphTokenLook {
    case live, blocked, inert
}

/// The token's frame: height 26, 1 pt border, `PVRadius.sm`. A blocked token's
/// border is dashed.
private struct SourcesGraphTokenFrame<Content: View>: View {
    var look: SourcesGraphTokenLook
    var hovered = false
    @ViewBuilder var content: Content

    var body: some View {
        let shape = RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
        content
            .padding(.horizontal, PVSpacing.space4 + PVSpacing.spacePx)
            .frame(height: 26)
            .background(shape.fill(hovered ? PVColor.surfaceHover : PVColor.surfaceCard))
            .overlay {
                if look == .blocked {
                    shape.strokeBorder(PVColor.borderDefault, style: PVStroke.dashed)
                } else {
                    shape.strokeBorder(hovered ? PVColor.borderStrong : PVColor.borderDefault, lineWidth: 1)
                }
            }
    }
}

/// Kit hover and press on the token frame, with the kit focus ring. Only a
/// `live` token takes the hover fill.
private struct SourcesGraphTokenButtonStyle: ButtonStyle {
    let look: SourcesGraphTokenLook

    func makeBody(configuration: Configuration) -> some View {
        SourcesGraphTokenButtonBody(configuration: configuration, look: look)
    }
}

private struct SourcesGraphTokenButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let look: SourcesGraphTokenLook

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        let live = look == .live
        PVHoverEffect(isPressed: live && configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            SourcesGraphTokenFrame(look: look, hovered: live && (showHover || configuration.isPressed)) {
                configuration.label
            }
            .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
        }
    }
}
