import AppKit
import SwiftUI

/// Optional grid underlay for a canvas document view.
///
/// One repeated cell, painted as a layer pattern. The document can grow
/// without a bitmap that size. Product-agnostic — any graph host can use it
/// or supply different document content to `GraphCanvasScrollView`.
struct GraphCanvasGridView: View {
    /// Size of the pannable document in content points.
    var contentSize: CGSize
    /// Spacing between grid lines in content points (future snap unit).
    var gridSpacing: CGFloat = GraphCanvasGridMapping.defaultSpacing
    @Environment(\.graphCanvasViewport) private var viewport

    var body: some View {
        GraphCanvasGridPatternRepresentable(
            gridSpacing: gridSpacing,
            magnification: viewport?.magnification ?? 1
        )
        .frame(width: contentSize.width, height: contentSize.height)
        .background(PVColor.surfacePage)
    }
}

private struct GraphCanvasGridPatternRepresentable: NSViewRepresentable {
    var gridSpacing: CGFloat
    var magnification: CGFloat

    func makeNSView(context: Context) -> GraphCanvasGridPatternView {
        let view = GraphCanvasGridPatternView()
        view.gridSpacing = gridSpacing
        view.magnification = magnification
        return view
    }

    func updateNSView(_ view: GraphCanvasGridPatternView, context: Context) {
        let spacingChanged = view.gridSpacing != gridSpacing
        let magnificationChanged = abs(view.magnification - magnification) > 0.000_1
        guard spacingChanged || magnificationChanged else { return }
        view.gridSpacing = gridSpacing
        view.magnification = magnification
        view.needsDisplay = true
    }
}

/// Tiles one cell through the layer's background color. `wantsUpdateLayer`
/// keeps that off a document-sized backing store.
final class GraphCanvasGridPatternView: NSView {
    var gridSpacing: CGFloat = GraphCanvasGridMapping.defaultSpacing
    var magnification: CGFloat = 1

    override var isFlipped: Bool { true }
    override var isOpaque: Bool { false }
    override var wantsUpdateLayer: Bool { true }

    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        wantsLayer = true
    }

    required init?(coder: NSCoder) {
        nil
    }

    override func hitTest(_ point: NSPoint) -> NSView? {
        nil
    }

    override func updateLayer() {
        super.updateLayer()
        layer?.contents = nil
        layer?.backgroundColor = Self.patternColor(
            spacing: gridSpacing,
            magnification: magnification,
            appearance: effectiveAppearance
        )
        layer?.opacity = Float(
            GraphCanvasCoordinates.gridPatternOpacity(spacing: gridSpacing, magnification: magnification)
        )
    }

    override func viewDidChangeEffectiveAppearance() {
        super.viewDidChangeEffectiveAppearance()
        needsDisplay = true
    }

    private static func patternColor(
        spacing: CGFloat,
        magnification: CGFloat,
        appearance: NSAppearance
    ) -> CGColor? {
        let cell = max(spacing, 1)
        let lineWidth = min(
            GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: magnification),
            cell
        )
        var stroke = NSColor.gray.withAlphaComponent(0.7)
        appearance.performAsCurrentDrawingAppearance {
            stroke = NSColor(PVColor.borderSubtle).withAlphaComponent(0.7)
        }
        guard let color = stroke.usingColorSpace(.sRGB)?.cgColor else { return nil }
        return GraphCanvasGridPattern.color(spacing: cell, lineWidth: lineWidth, stroke: color)
    }
}

private enum GraphCanvasGridPattern {
    static func color(spacing: CGFloat, lineWidth: CGFloat, stroke: CGColor) -> CGColor? {
        let state = GridPatternDrawState(spacing: spacing, lineWidth: lineWidth, stroke: stroke)
        let info = Unmanaged.passRetained(state).toOpaque()
        var callbacks = CGPatternCallbacks(
            version: 0,
            drawPattern: { info, context in
                guard let info else { return }
                let state = Unmanaged<GridPatternDrawState>.fromOpaque(info).takeUnretainedValue()
                context.setStrokeColor(state.stroke)
                context.setLineWidth(state.lineWidth)
                // Keep the stroke inside the cell. A line centered on the
                // boundary is clipped by the pattern tile.
                let inset = state.lineWidth / 2
                context.move(to: CGPoint(x: inset, y: 0))
                context.addLine(to: CGPoint(x: inset, y: state.spacing))
                context.move(to: CGPoint(x: 0, y: inset))
                context.addLine(to: CGPoint(x: state.spacing, y: inset))
                context.strokePath()
            },
            releaseInfo: { info in
                guard let info else { return }
                Unmanaged<GridPatternDrawState>.fromOpaque(info).release()
            }
        )
        guard let pattern = CGPattern(
            info: info,
            bounds: CGRect(x: 0, y: 0, width: spacing, height: spacing),
            matrix: .identity,
            xStep: spacing,
            yStep: spacing,
            tiling: .constantSpacing,
            isColored: true,
            callbacks: &callbacks
        ) else {
            Unmanaged<GridPatternDrawState>.fromOpaque(info).release()
            return nil
        }
        guard let space = CGColorSpace(patternBaseSpace: nil) else { return nil }
        var alpha: [CGFloat] = [1]
        return CGColor(patternSpace: space, pattern: pattern, components: &alpha)
    }
}

private final class GridPatternDrawState {
    let spacing: CGFloat
    let lineWidth: CGFloat
    let stroke: CGColor

    init(spacing: CGFloat, lineWidth: CGFloat, stroke: CGColor) {
        self.spacing = spacing
        self.lineWidth = lineWidth
        self.stroke = stroke
    }
}
