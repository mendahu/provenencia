import SwiftUI

/// The graph the pointer is on, on a node or on any part of a line. Its own observable, so a
/// hover redraws only the views that read it: the rows, the nodes and the highlighted arcs,
/// and never the read's content or the diagram's layout.
@MainActor @Observable
final class RecordLinksHover {
    var graphID: String?
}

/// The record links diagram (S13-28, board S13-D10): each record on a circle, an arc per
/// shared person between two records. Arcs carry no data and take no pointer; a node
/// highlights with its arcs and opens its record. Person-free: it takes nodes and arcs, the
/// hover, the graph of what has the keyboard, and closures for open and the focus handle of
/// each node. It reads no hover itself: its layout is built once per nodes, arcs and size.
struct RecordLinksDiagram: View {
    let nodes: [RecordLinksNode]
    let arcs: [RecordLinksArc]
    let hover: RecordLinksHover
    /// The graph of what has the keyboard; the hover wins while the pointer is on a graph.
    let focusedGraphID: String?
    /// The column's width; the circle is `RecordLinksLayout.diameter(width:fitting:)` unless
    /// `diameter` is given.
    let width: CGFloat
    var diameter: CGFloat?
    /// The accessibility identifier root of the nodes: `<idRoot>.node.<graph id>`.
    let idRoot: String
    let open: (String) -> Void
    let focus: (String) -> PVFieldFocus

    /// Past this many nodes only the highlighted or focused node prints its ref.
    static let refHidingThreshold = 30

    var body: some View {
        let layout = RecordLinksLayout(nodeCount: nodes.count, arcs: arcs, width: width, diameter: diameter)
        RecordLinksHighlightedLayers(
            layout: layout, nodes: nodes, arcs: arcs, hover: hover, focusedGraphID: focusedGraphID,
            idRoot: idRoot, open: open, focus: focus
        )
        .frame(width: width, height: layout.size.height, alignment: .topLeading)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: L10n.Conclusions.recordLinksGroupAccessibility(count: nodes.count)))
    }
}

/// The circle, the arcs and the nodes: the one view that reads the hover.
private struct RecordLinksHighlightedLayers: View {
    let layout: RecordLinksLayout
    let nodes: [RecordLinksNode]
    let arcs: [RecordLinksArc]
    let hover: RecordLinksHover
    let focusedGraphID: String?
    let idRoot: String
    let open: (String) -> Void
    let focus: (String) -> PVFieldFocus

    var body: some View {
        let highlightedID = hover.graphID ?? focusedGraphID
        let highlighted = highlightedID.flatMap { id in nodes.firstIndex { $0.id == id } }
        let neighbors = highlighted.map { RecordLinksArc.neighbors(of: $0, in: arcs) } ?? []
        ZStack(alignment: .topLeading) {
            if nodes.count > 1 {
                Circle()
                    .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                    .frame(width: layout.diameter, height: layout.diameter)
                    .position(layout.center)
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
                // Drawn once for the layout; a highlight dims it with an opacity, a composite
                // and not a redraw.
                RecordLinksRestArcs(layout: layout)
                    .equatable()
                    .opacity(highlighted == nil ? 1 : 0.18)
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
                if let highlighted {
                    RecordLinksAccentArcs(layout: layout, node: highlighted)
                        .allowsHitTesting(false)
                        .accessibilityHidden(true)
                }
            }
            ForEach(Array(nodes.enumerated()), id: \.element.id) { index, node in
                let state: RecordLinksNodeState = highlighted == nil ? .rest
                    : index == highlighted ? .highlighted
                    : neighbors.contains(index) ? .neighbor : .receded
                let handle = focus(node.id)
                RecordLinksNodeButton(
                    node: node,
                    point: layout.nodes[index].point,
                    side: layout.nodes[index].side,
                    showsRef: nodes.count <= RecordLinksDiagram.refHidingThreshold
                        || state == .highlighted || handle.isActive(),
                    overlaysRef: nodes.count > RecordLinksDiagram.refHidingThreshold,
                    state: state,
                    idRoot: idRoot,
                    onHover: { hover.graphID = $0 },
                    open: open,
                    focus: handle
                )
            }
        }
    }
}

/// Every arc at rest, thin under thick, in `borderStrong`. Equatable on the layout alone, so a
/// hover that leaves the layout as it was does not run its drawing again.
private struct RecordLinksRestArcs: View, Equatable {
    let layout: RecordLinksLayout

    var body: some View {
        Canvas { context, _ in
            for strong in [false, true] {
                for curve in layout.curves where curve.arc.strong == strong {
                    context.stroke(
                        curve.path, with: .color(PVColor.borderStrong),
                        style: StrokeStyle(lineWidth: strong ? 3 : 1.25, lineCap: .round)
                    )
                }
            }
        }
    }
}

/// The highlighted node's arcs alone, in the accent line, above the dimmed rest.
private struct RecordLinksAccentArcs: View {
    let layout: RecordLinksLayout
    let node: Int

    var body: some View {
        Canvas { context, _ in
            let touching = layout.curves.filter { $0.arc.touches(node) }
            for strong in [false, true] {
                for curve in touching where curve.arc.strong == strong {
                    context.stroke(
                        curve.path, with: .color(PVColor.accentLine),
                        style: StrokeStyle(lineWidth: strong ? 3 : 1.25, lineCap: .round)
                    )
                }
            }
        }
    }
}

private extension RecordLinksLayout.Curve {
    var path: Path {
        var path = Path()
        path.move(to: from)
        path.addQuadCurve(to: to, control: control)
        return path
    }
}

/// How a node is drawn while something is highlighted.
private enum RecordLinksNodeState {
    case rest, highlighted, neighbor, receded
}

/// One node: the dot and its ref in a button, placed so the dot's centre is `point`.
private struct RecordLinksNodeButton: View {
    let node: RecordLinksNode
    let point: CGPoint
    let side: RecordLinksLayout.LabelSide
    let showsRef: Bool
    /// The ref is drawn outside the button's layout (many nodes), so showing it moves nothing.
    let overlaysRef: Bool
    let state: RecordLinksNodeState
    let idRoot: String
    let onHover: (String?) -> Void
    let open: (String) -> Void
    let focus: PVFieldFocus

    @State private var size: CGSize = .zero

    private static let dot: CGFloat = 8
    private static let padX: CGFloat = 4
    private static let padY: CGFloat = 2

    var body: some View {
        let button = Button { open(node.id) } label: { label }
            .buttonStyle(RecordLinksNodeStyle(onHover: { onHover($0 ? node.id : nil) }))
            .disabled(node.location == nil)
            .pvActivatesOnReturn(when: node.location != nil) { open(node.id) }
            .help(node.help)
            .accessibilityLabel(Text(verbatim: node.accessibilityLabel))
            .accessibilityIdentifier("\(idRoot).node.\(node.id)")
            .fixedSize()
            .onGeometryChange(for: CGSize.self) { $0.size } action: { size = $0 }
        focus.apply(AnyView(button))
            .position(x: point.x - dotCenter.x + size.width / 2, y: point.y - dotCenter.y + size.height / 2)
    }

    /// Where the dot's centre sits inside the button.
    private var dotCenter: CGPoint {
        let inset = Self.dot / 2
        switch side {
        case .right: return CGPoint(x: Self.padX + inset, y: size.height / 2)
        case .left: return CGPoint(x: size.width - Self.padX - inset, y: size.height / 2)
        case .top: return CGPoint(x: size.width / 2, y: size.height - Self.padY - inset)
        case .bottom: return CGPoint(x: size.width / 2, y: Self.padY + inset)
        }
    }

    private var opacity: Double { state == .receded ? 0.35 : 1 }

    private var dotView: some View {
        let fill: Color = state == .highlighted ? PVColor.accent : state == .neighbor ? PVColor.accentSoft : PVColor.textSecondary
        let ring: Color = state == .highlighted ? PVColor.accent : state == .neighbor ? PVColor.accentLine : PVColor.textSecondary
        return Circle()
            .fill(fill)
            .overlay(Circle().strokeBorder(ring, lineWidth: 1.5))
            .frame(width: Self.dot, height: Self.dot)
    }

    private var refView: some View {
        let color: Color = state == .highlighted ? PVColor.accent : state == .neighbor ? PVColor.textPrimary : PVColor.textSecondary
        return Text(verbatim: node.label)
            .font(PVFont.mono(size: PVTypeScale.caption, weight: state == .highlighted ? PVFontWeight.semibold : PVFontWeight.regular))
            .foregroundStyle(color)
            .fixedSize()
    }

    @ViewBuilder
    private var label: some View {
        let gap = PVSpacing.space3
        Group {
            if overlaysRef {
                dotView.overlay(alignment: overlayAlignment) {
                    if showsRef {
                        refView
                            .padding(.horizontal, 3)
                            .background(PVColor.surfacePage, in: RoundedRectangle(cornerRadius: PVRadius.sm))
                            .alignmentGuide(overlayHorizontalGuide) { [side] d in Self.shift(d, side: side, gap: gap) }
                            .alignmentGuide(overlayVerticalGuide) { [side] d in Self.shiftVertical(d, side: side, gap: gap) }
                    }
                }
            } else {
                switch side {
                case .right: HStack(spacing: gap) { dotView; refView }
                case .left: HStack(spacing: gap) { refView; dotView }
                case .top: VStack(spacing: gap) { refView; dotView }
                case .bottom: VStack(spacing: gap) { dotView; refView }
                }
            }
        }
        .padding(.horizontal, Self.padX)
        .padding(.vertical, Self.padY)
        .opacity(opacity)
    }

    private var overlayAlignment: Alignment {
        switch side {
        case .right: .trailing
        case .left: .leading
        case .top: .top
        case .bottom: .bottom
        }
    }

    private var overlayHorizontalGuide: HorizontalAlignment {
        side == .left ? .leading : .trailing
    }

    private var overlayVerticalGuide: VerticalAlignment {
        side == .top ? .top : .bottom
    }

    private nonisolated static func shift(_ d: ViewDimensions, side: RecordLinksLayout.LabelSide, gap: CGFloat) -> CGFloat {
        switch side {
        case .right: d[.leading] - gap
        case .left: d[.trailing] + gap
        case .top, .bottom: d[HorizontalAlignment.center]
        }
    }

    private nonisolated static func shiftVertical(_ d: ViewDimensions, side: RecordLinksLayout.LabelSide, gap: CGFloat) -> CGFloat {
        switch side {
        case .top: d[.bottom] + gap
        case .bottom: d[.top] - gap
        case .right, .left: d[VerticalAlignment.center]
        }
    }
}

/// The kit's hover and focus chrome on a node: a hover fill and the focus ring around the
/// whole button, reporting hover to the diagram through the one tracker.
private struct RecordLinksNodeStyle: ButtonStyle {
    let onHover: (Bool) -> Void

    func makeBody(configuration: Configuration) -> some View {
        RecordLinksNodeStyleBody(configuration: configuration, onHover: onHover)
    }
}

private struct RecordLinksNodeStyleBody: View {
    let configuration: ButtonStyleConfiguration
    let onHover: (Bool) -> Void

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed, onHoverChange: onHover) { showHover in
            configuration.label
                .background(
                    (showHover ? PVColor.surfaceHover : Color.clear),
                    in: RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                )
                .pvFocusRing(isFocused, cornerRadius: PVRadius.md)
        }
    }
}

#if DEBUG
private struct RecordLinksDiagramPreview: View {
    let links: CatalogPersonRecordLinks
    var width: CGFloat = 340
    @State private var hover = RecordLinksHover()

    var body: some View {
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        RecordLinksDiagram(
            nodes: content.nodes,
            arcs: content.arcs,
            hover: hover,
            focusedGraphID: nil,
            width: width,
            idRoot: "preview",
            open: { _ in },
            focus: { _ in PVFieldFocus(isActive: { false }, focus: {}, resign: {}, apply: { $0 }) }
        )
        .padding()
        .background(PVColor.surfacePage)
    }
}

#Preview("Johann") {
    RecordLinksDiagramPreview(links: PersonSourcesPreviewData.johannRecordLinks)
}

#Preview("One graph") {
    RecordLinksDiagramPreview(links: CatalogPersonRecordLinks(
        graphs: [PersonSourcesPreviewData.johannRecordLinks.graphs[0]], pairs: []
    ))
}

#Preview("50 household graphs, Wide") {
    RecordLinksDiagramPreview(links: PersonSourcesPreviewData.householdRecordLinks(graphs: 50), width: 720)
}

#Preview("50 graphs, Wide") {
    RecordLinksDiagramPreview(links: PersonSourcesPreviewData.manyRecordLinks(graphs: 50), width: 720)
}
#endif
