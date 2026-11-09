import SwiftUI

/// A circular link diagram: each node on a circle, an arc per link between two nodes. Arcs
/// carry no data and take no pointer; a node highlights with its arcs and opens. Generic over
/// the caller's node id: it takes a dataset, the hover, the node that has the keyboard,
/// the style, and closures for open and each node's focus handle. It reads no hover itself:
/// its layout is built once per dataset, size and style.
struct CircularLinkDiagram<ID: Hashable>: View {
    let dataset: CircularLinkDataset<ID>
    var style: CircularLinkDiagramStyle = .standard
    /// The column's width.
    let width: CGFloat
    /// The height left for the drawing; the circle shrinks to it, never below the style's
    /// minimum. Nil: the circle fills the width.
    var fittingHeight: CGFloat?
    let hover: CircularLinkHover<ID>
    /// The node that has the keyboard; the hover wins while the pointer is on a node.
    var focusedID: ID?
    /// The group's spoken label.
    let accessibilityLabel: String
    /// Gives each node the identifier `<root>.node.<id>`; nil sets none.
    var accessibilityIdentifierRoot: String?
    var onOpen: (ID) -> Void = { _ in }
    /// Each node's focus handle; nil gives the nodes no focus binding.
    var focus: ((ID) -> PVFieldFocus)?

    var body: some View {
        let diameter = CircularLinkLayout<ID>.diameter(width: width, fitting: fittingHeight, style: style)
        let layout = CircularLinkLayout(dataset: dataset, width: width, diameter: diameter, style: style)
        CircularLinkHighlightedLayers(
            layout: layout, dataset: dataset, style: style, hover: hover, focusedID: focusedID,
            identifierRoot: accessibilityIdentifierRoot, open: onOpen, focus: focus
        )
        .frame(width: width, height: layout.size.height, alignment: .topLeading)
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(verbatim: accessibilityLabel))
    }
}

/// The circle, the arcs and the nodes: the one view that reads the hover.
private struct CircularLinkHighlightedLayers<ID: Hashable>: View {
    let layout: CircularLinkLayout<ID>
    let dataset: CircularLinkDataset<ID>
    let style: CircularLinkDiagramStyle
    let hover: CircularLinkHover<ID>
    let focusedID: ID?
    let identifierRoot: String?
    let open: (ID) -> Void
    let focus: ((ID) -> PVFieldFocus)?

    var body: some View {
        let nodes = dataset.nodes
        let highlightedID = hover.highlighted(focused: focusedID)
        let highlighted = highlightedID.flatMap { dataset.index(of: $0) }
        let neighbors = highlighted.map { dataset.neighbors(ofIndex: $0) } ?? []
        ZStack(alignment: .topLeading) {
            if nodes.count > 1 {
                Circle()
                    .strokeBorder(style.circleColor, lineWidth: style.circleLineWidth)
                    .frame(width: layout.diameter, height: layout.diameter)
                    .position(layout.center)
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
                // Drawn once for the layout; a highlight dims it with an opacity, a composite
                // and not a redraw.
                CircularLinkRestArcs(layout: layout, style: style)
                    .equatable()
                    .opacity(highlighted == nil ? 1 : style.dimmedArcOpacity)
                    .allowsHitTesting(false)
                    .accessibilityHidden(true)
                if let highlighted {
                    CircularLinkAccentArcs(layout: layout, style: style, node: highlighted)
                        .allowsHitTesting(false)
                        .accessibilityHidden(true)
                }
            }
            ForEach(Array(nodes.enumerated()), id: \.element.id) { index, node in
                let state: CircularLinkNodeState = highlighted == nil ? .rest
                    : index == highlighted ? .highlighted
                    : neighbors.contains(index) ? .neighbor : .receded
                let handle = focus?(node.id)
                CircularLinkNodeButton(
                    node: node,
                    point: layout.nodes[index].point,
                    side: layout.nodes[index].side,
                    showsLabel: nodes.count <= style.labelsShownUpTo
                        || state == .highlighted || (handle?.isActive() ?? false),
                    overlaysLabel: nodes.count > style.labelsShownUpTo,
                    state: state,
                    style: style,
                    identifierRoot: identifierRoot,
                    onHover: { hover.id = $0 ? node.id : nil },
                    open: open,
                    focus: handle
                )
            }
        }
    }
}

/// Every arc at rest, thinnest first. Equatable on the layout and the style, so a hover that
/// leaves them as they were does not run its drawing again.
private struct CircularLinkRestArcs<ID: Hashable>: View, Equatable {
    let layout: CircularLinkLayout<ID>
    let style: CircularLinkDiagramStyle

    var body: some View {
        Canvas { context, _ in
            for curve in layout.curves {
                context.stroke(
                    curve.path, with: .color(curve.color ?? style.arcColor),
                    style: StrokeStyle(lineWidth: curve.lineWidth, lineCap: .round)
                )
            }
        }
    }
}

/// The highlighted node's arcs alone, in the highlight colour, above the dimmed rest.
private struct CircularLinkAccentArcs<ID: Hashable>: View {
    let layout: CircularLinkLayout<ID>
    let style: CircularLinkDiagramStyle
    let node: Int

    var body: some View {
        Canvas { context, _ in
            for curve in layout.curves where curve.touches(node) {
                context.stroke(
                    curve.path, with: .color(style.highlightedArcColor),
                    style: StrokeStyle(lineWidth: curve.lineWidth, lineCap: .round)
                )
            }
        }
    }
}

private extension CircularLinkLayout.Curve {
    var path: Path {
        var path = Path()
        path.move(to: from)
        path.addQuadCurve(to: to, control: control)
        return path
    }
}

/// How a node is drawn while something is highlighted.
private enum CircularLinkNodeState {
    case rest, highlighted, neighbor, receded
}

/// One node: the dot and its label in a button, placed so the dot's centre is `point`.
private struct CircularLinkNodeButton<ID: Hashable>: View {
    let node: CircularLinkNode<ID>
    let point: CGPoint
    let side: CircularLinkLabelSide
    let showsLabel: Bool
    /// The label is drawn outside the button's layout (many nodes), so showing it moves nothing.
    let overlaysLabel: Bool
    let state: CircularLinkNodeState
    let style: CircularLinkDiagramStyle
    let identifierRoot: String?
    let onHover: (Bool) -> Void
    let open: (ID) -> Void
    let focus: PVFieldFocus?

    @State private var size: CGSize = .zero

    var body: some View {
        let button = Button { open(node.id) } label: { label }
            .buttonStyle(CircularLinkNodeStyle(
                hoverFill: style.nodeHoverFill, cornerRadius: style.nodeCornerRadius, onHover: onHover
            ))
            .disabled(!node.isEnabled)
            .pvActivatesOnReturn(when: node.isEnabled) { open(node.id) }
            .modifier(CircularLinkOptionalHelp(help: node.help))
            .accessibilityLabel(Text(verbatim: node.accessibilityLabel ?? node.label))
            .modifier(CircularLinkOptionalIdentifier(
                identifier: identifierRoot.map { "\($0).node.\(String(describing: node.id))" }
            ))
            .fixedSize()
            .onGeometryChange(for: CGSize.self) { $0.size } action: { size = $0 }
        let placed = focus.map { $0.apply(AnyView(button)) } ?? AnyView(button)
        placed
            .position(x: point.x - dotCenter.x + size.width / 2, y: point.y - dotCenter.y + size.height / 2)
    }

    /// Where the dot's centre sits inside the button.
    private var dotCenter: CGPoint {
        let inset = style.dotSize / 2
        let padX = style.nodePadding.width
        let padY = style.nodePadding.height
        switch side {
        case .right: return CGPoint(x: padX + inset, y: size.height / 2)
        case .left: return CGPoint(x: size.width - padX - inset, y: size.height / 2)
        case .top: return CGPoint(x: size.width / 2, y: size.height - padY - inset)
        case .bottom: return CGPoint(x: size.width / 2, y: padY + inset)
        }
    }

    private var opacity: Double { state == .receded ? style.recededNodeOpacity : 1 }

    private var dotView: some View {
        let fill: Color = state == .highlighted ? style.highlightedNodeColor : state == .neighbor ? style.neighborNodeFill : style.nodeColor
        let ring: Color = state == .highlighted ? style.highlightedNodeColor : state == .neighbor ? style.neighborNodeRing : style.nodeColor
        return Circle()
            .fill(fill)
            .overlay(Circle().strokeBorder(ring, lineWidth: style.dotRingWidth))
            .frame(width: style.dotSize, height: style.dotSize)
    }

    private var labelView: some View {
        let color: Color = state == .highlighted ? style.highlightedLabelColor : state == .neighbor ? style.neighborLabelColor : style.labelColor
        return Text(verbatim: node.label)
            .font(state == .highlighted ? style.highlightedLabelFont : style.labelFont)
            .foregroundStyle(color)
            .fixedSize()
    }

    @ViewBuilder
    private var label: some View {
        let gap = style.labelGap
        Group {
            if overlaysLabel {
                dotView.overlay(alignment: overlayAlignment) {
                    if showsLabel {
                        labelView
                            .padding(.horizontal, style.overlayLabelPadding)
                            .background(style.overlayLabelBackground, in: RoundedRectangle(cornerRadius: style.overlayLabelCornerRadius))
                            .alignmentGuide(overlayHorizontalGuide) { [side] d in CircularLinkLabelGuides.shift(d, side: side, gap: gap) }
                            .alignmentGuide(overlayVerticalGuide) { [side] d in CircularLinkLabelGuides.shiftVertical(d, side: side, gap: gap) }
                    }
                }
            } else {
                switch side {
                case .right: HStack(spacing: gap) { dotView; labelView }
                case .left: HStack(spacing: gap) { labelView; dotView }
                case .top: VStack(spacing: gap) { labelView; dotView }
                case .bottom: VStack(spacing: gap) { dotView; labelView }
                }
            }
        }
        .padding(.horizontal, style.nodePadding.width)
        .padding(.vertical, style.nodePadding.height)
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
}

/// The alignment shifts that put an overlaid label a gap beyond its dot.
private enum CircularLinkLabelGuides {
    nonisolated static func shift(_ d: ViewDimensions, side: CircularLinkLabelSide, gap: CGFloat) -> CGFloat {
        switch side {
        case .right: d[.leading] - gap
        case .left: d[.trailing] + gap
        case .top, .bottom: d[HorizontalAlignment.center]
        }
    }

    nonisolated static func shiftVertical(_ d: ViewDimensions, side: CircularLinkLabelSide, gap: CGFloat) -> CGFloat {
        switch side {
        case .top: d[.bottom] + gap
        case .bottom: d[.top] - gap
        case .right, .left: d[VerticalAlignment.center]
        }
    }
}

/// `.help` only when the node has a tooltip.
private struct CircularLinkOptionalHelp: ViewModifier {
    let help: String?

    @ViewBuilder
    func body(content: Content) -> some View {
        if let help { content.help(help) } else { content }
    }
}

/// `.accessibilityIdentifier` only when the diagram was given an identifier root.
private struct CircularLinkOptionalIdentifier: ViewModifier {
    let identifier: String?

    @ViewBuilder
    func body(content: Content) -> some View {
        if let identifier { content.accessibilityIdentifier(identifier) } else { content }
    }
}

/// The kit's hover and focus chrome on a node: a hover fill and the focus ring around the
/// whole button, reporting hover to the diagram through the one tracker.
private struct CircularLinkNodeStyle: ButtonStyle {
    let hoverFill: Color
    let cornerRadius: CGFloat
    let onHover: (Bool) -> Void

    func makeBody(configuration: Configuration) -> some View {
        CircularLinkNodeStyleBody(
            configuration: configuration, hoverFill: hoverFill, cornerRadius: cornerRadius, onHover: onHover
        )
    }
}

private struct CircularLinkNodeStyleBody: View {
    let configuration: ButtonStyleConfiguration
    let hoverFill: Color
    let cornerRadius: CGFloat
    let onHover: (Bool) -> Void

    /// Read here, in the button's focus scope (see `PVButtonBody`).
    @Environment(\.isFocused) private var isFocused

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed, onHoverChange: onHover) { showHover in
            configuration.label
                .background(
                    (showHover ? hoverFill : Color.clear),
                    in: RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                )
                .pvFocusRing(isFocused, cornerRadius: cornerRadius)
        }
    }
}

#if DEBUG
private struct CircularLinkDiagramPreview: View {
    var nodes: [CircularLinkNode<String>]
    var arcs: [CircularLinkArc<String>]
    var style: CircularLinkDiagramStyle = .standard
    var width: CGFloat = 340
    @State private var hover = CircularLinkHover<String>()

    var body: some View {
        CircularLinkDiagram(
            dataset: CircularLinkDataset(nodes: nodes, arcs: arcs), style: style, width: width, hover: hover,
            accessibilityLabel: "Preview"
        )
        .padding()
        .background(PVColor.surfacePage)
    }

    static func numbered(_ count: Int) -> [CircularLinkNode<String>] {
        (0..<count).map { CircularLinkNode(id: "n\($0)", label: "N\($0)", help: "Node \($0)") }
    }
}

#Preview("Six nodes") {
    let nodes = ["A", "B", "C", "D", "E", "F"].map { CircularLinkNode(id: $0, label: $0, help: "Node \($0)") }
    CircularLinkDiagramPreview(nodes: nodes, arcs: [
        CircularLinkArc(from: "A", to: "B", lineWidth: 3),
        CircularLinkArc(from: "B", to: "A", lineWidth: 1.25),
        CircularLinkArc(from: "A", to: "B", lineWidth: 1.25),
        CircularLinkArc(from: "A", to: "D", lineWidth: 3),
        CircularLinkArc(from: "B", to: "D", lineWidth: 1.25),
        CircularLinkArc(from: "C", to: "E", lineWidth: 1.25),
    ])
}

#Preview("One node") {
    CircularLinkDiagramPreview(nodes: [CircularLinkNode(id: "A", label: "A")], arcs: [])
}

#Preview("Custom style") {
    var style = CircularLinkDiagramStyle.standard
    style.arcColor = .orange.opacity(0.5)
    style.highlightedArcColor = .purple
    style.highlightedNodeColor = .purple
    style.dotSize = 14
    let nodes = ["A", "B", "C", "D", "E"].map { CircularLinkNode(id: $0, label: $0) }
    return CircularLinkDiagramPreview(nodes: nodes, arcs: [
        CircularLinkArc(from: "A", to: "C", lineWidth: 5),
        CircularLinkArc(from: "B", to: "E", lineWidth: 2, color: .teal),
        CircularLinkArc(from: "A", to: "D", lineWidth: 2),
    ], style: style)
}

#Preview("300 arcs, 40 nodes, Wide") {
    let arcs = (0..<300).map { index in
        CircularLinkArc(from: "n\(index % 40)", to: "n\((index * 7 + 3) % 40)", lineWidth: index % 5 == 0 ? 3 : 1.25)
    }
    CircularLinkDiagramPreview(nodes: CircularLinkDiagramPreview.numbered(40), arcs: arcs, width: 720)
}
#endif
