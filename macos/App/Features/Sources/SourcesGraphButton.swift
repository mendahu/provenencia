import SwiftUI

/// One Evidence graph (or the reason there is none to show) in a Sources list
/// row's graph column: its name and counts, opening the graph's canvas. One view
/// owns every state; the states are `Kind`.
struct SourcesGraphButton: View {
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

    /// The button's minimum height; a wrapped name makes it taller.
    static let height: CGFloat = 52
    /// The live button's status column, so the badges share a left edge down a
    /// stack.
    static let statusSlotWidth: CGFloat = 112

    var body: some View {
        switch kind {
        case .graph(let graph, let counts, let open):
            graphButton(graph, counts: counts, open: open)
        case .needsArtifact:
            SourcesGraphButtonFrame(look: .blocked) {
                content(
                    glyph: { PVIcon(.network, size: 15).foregroundStyle(PVColor.textFaint) },
                    name: { nameText(L10n.string(L10n.Sources.noGraphYet), color: PVColor.textFaint) },
                    detail: { italicDetail(L10n.Sources.needsArtifact) },
                    chevron: false
                )
            }
            .help(L10n.string(L10n.Sources.needsArtifactTooltip))
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphBlockLabel(
                title: L10n.string(L10n.Sources.noGraphYet),
                detail: L10n.string(L10n.Sources.needsArtifact)
            )))
            .accessibilityHint(Text(L10n.Sources.needsArtifactTooltip))
            .accessibilityIdentifier("sources.row.\(sourceID).graphBlocked")
        case .listLoading:
            SourcesGraphButtonFrame(look: .inert) {
                content(
                    glyph: { PVIcon(.network, size: 15).foregroundStyle(PVColor.textFaint) },
                    name: { PVSkeletonBar(width: 72, height: 6) },
                    detail: { PVSkeletonBar(width: 132, height: 6) },
                    chevron: false,
                    lineSpacing: PVSpacing.space4 + PVSpacing.spacePx
                )
            }
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(L10n.Sources.graphsLoading))
            .accessibilityIdentifier("sources.row.\(sourceID).graphsLoading")
        case .listFailed(let error):
            let message = L10n.Errors.message(for: error)
            SourcesGraphButtonFrame(look: .inert) {
                content(
                    glyph: { PVIcon(.circleAlert, size: 15).foregroundStyle(PVColor.danger) },
                    name: { nameText(L10n.string(L10n.Sources.graphsFailedTitle), color: PVColor.textSecondary) },
                    detail: { italicDetail(L10n.Sources.graphsFailedDetail) },
                    chevron: false
                )
            }
            .help(message)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphBlockLabel(
                title: L10n.string(L10n.Sources.graphsFailedTitle),
                detail: L10n.string(L10n.Sources.graphsFailedDetail)
            )))
            .accessibilityHint(Text(verbatim: message))
            .accessibilityIdentifier("sources.row.\(sourceID).graphsError")
        }
    }

    @ViewBuilder
    private func graphButton(
        _ graph: CatalogEvidenceGraph,
        counts: SourcesGraphCounts,
        open: Open
    ) -> some View {
        switch open {
        case .blocked(let reason):
            Button {} label: {
                content(
                    glyph: { PVIcon(.network, size: 15).foregroundStyle(PVColor.textFaint) },
                    name: { nameText(graph.name, color: PVColor.textFaint) },
                    detail: { italicDetail(L10n.Sources.graphNeedsArtifactToOpen) },
                    chevron: false
                )
            }
            .buttonStyle(SourcesGraphButtonStyle(look: .blocked))
            .help(L10n.Sources.graphButtonHelp(name: graph.name, detail: reason))
            .accessibilityLabel(Text(verbatim: graph.name))
            .accessibilityHint(Text(verbatim: reason))
            .accessibilityIdentifier("sources.row.\(sourceID).graph.\(graph.ref)")
        case .opens(let onOpen):
            Button {
                onOpen()
            } label: {
                content(
                    glyph: { PVIcon(.network, size: 15) },
                    name: { nameText(graph.name, color: PVColor.textPrimary) },
                    detail: { SourcesGraphCountLine(counts: counts, style: .button) },
                    chevron: true,
                    status: counts
                )
            }
            .buttonStyle(SourcesGraphButtonStyle(look: .live))
            .help(L10n.Sources.graphButtonHelp(
                name: graph.name,
                detail: L10n.Sources.graphCounts(counts, style: .full, notStarted: .zeroCounts)
            ))
            .accessibilityLabel(Text(verbatim: L10n.Sources.graphButtonLabel(
                open: L10n.Sources.openEvidenceGraphLabel(name: graph.name),
                counts: L10n.Sources.graphCounts(counts, style: .spoken, notStarted: .zeroCounts),
                status: counts.status.map { L10n.string($0.label) }
            )))
            .accessibilityIdentifier("sources.row.\(sourceID).graph.\(graph.ref)")
        }
    }

    /// Glyph, then the name and detail lines, then (for a graph that opens) the
    /// status slot and the chevron.
    private func content<Glyph: View, Name: View, Detail: View>(
        glyph: () -> Glyph,
        name: () -> Name,
        detail: () -> Detail,
        chevron: Bool,
        status: SourcesGraphCounts? = nil,
        lineSpacing: CGFloat = PVSpacing.spacePx
    ) -> some View {
        HStack(spacing: PVSpacing.space4 + PVSpacing.space1) {
            glyph()
            VStack(alignment: .leading, spacing: lineSpacing) {
                name()
                detail()
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if let status {
                statusSlot(status)
            }
            if chevron {
                PVIcon(.chevronForward, size: 14)
            }
        }
    }

    /// The graph's status badge, a rule while its counts load, or nothing.
    @ViewBuilder
    private func statusSlot(_ counts: SourcesGraphCounts) -> some View {
        Group {
            if let status = counts.status {
                status.badge
            } else if counts == .loading {
                PVSkeletonBar(width: 72, height: 6)
                    .accessibilityHidden(true)
            }
        }
        .frame(width: Self.statusSlotWidth, alignment: .trailing)
    }

    private func nameText(_ name: String, color: Color) -> some View {
        Text(verbatim: name)
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .foregroundStyle(color)
            .fixedSize(horizontal: false, vertical: true)
    }

    private func italicDetail(_ text: LocalizedStringResource) -> some View {
        Text(text)
            .font(PVFont.body(size: PVTypeScale.micro))
            .italic()
            .foregroundStyle(PVColor.textMuted)
            .lineLimit(1)
            .truncationMode(.tail)
    }
}

/// How a graph button looks: `live` opens a graph, `blocked` is dashed and takes
/// focus but no hover fill, `inert` (loading, failed) fills nothing at all.
private enum SourcesGraphButtonLook {
    case live, blocked, inert
}

/// The button's frame: minimum height 52, 1 pt border, `PVRadius.sm`. A blocked button's
/// border is dashed.
private struct SourcesGraphButtonFrame<Content: View>: View {
    var look: SourcesGraphButtonLook
    var hovered = false
    @ViewBuilder var content: Content

    var body: some View {
        let shape = RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
        content
            .padding(.vertical, PVSpacing.space3)
            .padding(.horizontal, PVSpacing.space4 + PVSpacing.space1)
            .frame(maxWidth: .infinity)
            .frame(minHeight: SourcesGraphButton.height)
            .background(shape.fill(hovered ? PVColor.surfaceHover : PVColor.surfaceCard))
            .overlay {
                switch look {
                case .blocked:
                    shape.strokeBorder(PVColor.borderDefault, style: PVStroke.dashed)
                case .inert:
                    shape.strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                case .live:
                    shape.strokeBorder(hovered ? PVColor.borderStrong : PVColor.borderDefault, lineWidth: 1)
                }
            }
    }
}

/// Kit hover and press on the button frame, with the kit focus ring. Only a
/// `live` button takes the hover fill and the darker glyphs.
private struct SourcesGraphButtonStyle: ButtonStyle {
    let look: SourcesGraphButtonLook

    func makeBody(configuration: Configuration) -> some View {
        SourcesGraphButtonBody(configuration: configuration, look: look)
    }
}

private struct SourcesGraphButtonBody: View {
    let configuration: ButtonStyleConfiguration
    let look: SourcesGraphButtonLook

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        let live = look == .live
        PVHoverEffect(isPressed: live && configuration.isPressed, hoverAnimation: PVMotion.fastStandard) { showHover in
            let hovered = live && (showHover || configuration.isPressed)
            SourcesGraphButtonFrame(look: look, hovered: hovered) {
                configuration.label
                    .foregroundStyle(hovered ? PVColor.textPrimary : PVColor.textMuted)
            }
            .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
        }
    }
}
