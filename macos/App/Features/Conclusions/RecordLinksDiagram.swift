import SwiftUI

/// The record links diagram (S13-28, board S13-D10): each record on a circle, an arc per
/// shared person between two records. Arcs carry no data and take no pointer; a node
/// highlights with its arcs and opens its record. Person-free: it takes nodes and arcs, the
/// highlighted node's id, and closures for hover, open and the focus handle of each node.
struct RecordLinksDiagram: View {
    let nodes: [RecordLinksNode]
    let arcs: [RecordLinksArc]
    /// The graph id to highlight: the hovered or focused node, or the line the pointer is on.
    let highlightedGraphID: String?
    /// The column's width; the circle is `RecordLinksLayout.diameter(width:fitting:)` unless
    /// `diameter` is given.
    let width: CGFloat
    var diameter: CGFloat?
    let onHover: (String?) -> Void
    let open: (String) -> Void
    let focus: (String) -> PVFieldFocus

    /// Past this many nodes only the highlighted or focused node prints its ref.
    static let refHidingThreshold = 30
    /// Height of the region a single node is drawn in, and where its dot sits in it.
    static let singleHeight: CGFloat = 64
    private static let singleDotY: CGFloat = 24

    private var layout: RecordLinksLayout {
        RecordLinksLayout(nodeCount: nodes.count, arcs: arcs, width: width, diameter: diameter)
    }

    var body: some View {
        let layout = layout
        let highlighted = highlightedGraphID.flatMap { id in nodes.firstIndex { $0.id == id } }
        let neighbors = highlighted.map { index in
            Set(arcs.compactMap { arc -> Int? in
                if arc.a == index { return arc.b }
                return arc.b == index ? arc.a : nil
            })
        } ?? []
        ZStack(alignment: .topLeading) {
            if nodes.count > 1 {
                Canvas { context, _ in
                    draw(layout, highlighted: highlighted, into: &context)
                }
                .allowsHitTesting(false)
                .accessibilityHidden(true)
            }
            ForEach(Array(nodes.enumerated()), id: \.element.id) { index, node in
                let state: RecordLinksNodeState = highlighted == nil ? .rest
                    : index == highlighted ? .highlighted
                    : neighbors.contains(index) ? .neighbor : .receded
                RecordLinksNodeButton(
                    node: node,
                    point: nodes.count == 1 ? CGPoint(x: width / 2, y: Self.singleDotY) : layout.nodes[index].point,
                    side: nodes.count == 1 ? .right : layout.nodes[index].side,
                    showsRef: nodes.count <= Self.refHidingThreshold || state == .highlighted,
                    overlaysRef: nodes.count > Self.refHidingThreshold,
                    state: state,
                    onHover: onHover,
                    open: open,
                    focus: focus(node.id)
                )
            }
        }
        .frame(
            width: width,
            height: nodes.count == 1 ? Self.singleHeight : layout.size.height,
            alignment: .topLeading
        )
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: L10n.Conclusions.recordLinksGroupAccessibility(count: nodes.count)))
    }

    private func draw(_ layout: RecordLinksLayout, highlighted: Int?, into context: inout GraphicsContext) {
        context.stroke(
            Path(ellipseIn: CGRect(
                x: layout.center.x - layout.radius, y: layout.center.y - layout.radius,
                width: layout.diameter, height: layout.diameter
            )),
            with: .color(PVColor.borderSubtle),
            lineWidth: 1
        )
        // Back to front: the rest at rest (or receded), thin under thick, a highlighted node's
        // arcs last in the accent line.
        let indexed = Array(zip(arcs, layout.curves).enumerated())
        let touching = { (arc: RecordLinksArc) in highlighted.map { arc.a == $0 || arc.b == $0 } ?? false }
        let order = indexed.sorted { lhs, rhs in
            let l = (touching(lhs.element.0) ? 2 : 0) + (lhs.element.0.strong ? 1 : 0)
            let r = (touching(rhs.element.0) ? 2 : 0) + (rhs.element.0.strong ? 1 : 0)
            return l == r ? lhs.offset < rhs.offset : l < r
        }
        for (_, pair) in order {
            let (arc, curve) = pair
            var path = Path()
            path.move(to: curve.from)
            path.addQuadCurve(to: curve.to, control: curve.control)
            let color: Color
            if touching(arc) {
                color = PVColor.accentLine
            } else {
                color = highlighted == nil ? PVColor.borderStrong : PVColor.borderStrong.opacity(0.18)
            }
            context.stroke(
                path, with: .color(color),
                style: StrokeStyle(lineWidth: arc.strong ? 3 : 1.25, lineCap: .round)
            )
        }
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
            .pvActivatesOnReturn(when: true) { open(node.id) }
            .help(node.help)
            .accessibilityLabel(Text(verbatim: node.accessibilityLabel))
            .accessibilityIdentifier("persons.detail.sources.node.\(node.id)")
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
    @State private var hovered: String?

    var body: some View {
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        RecordLinksDiagram(
            nodes: content.nodes,
            arcs: content.arcs,
            highlightedGraphID: hovered,
            width: width,
            onHover: { hovered = $0 },
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

#Preview("50 graphs, Wide") {
    RecordLinksDiagramPreview(links: PersonSourcesPreviewData.manyRecordLinks(graphs: 50), width: 720)
}
#endif
