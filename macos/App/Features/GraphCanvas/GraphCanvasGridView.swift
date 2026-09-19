import SwiftUI

/// Optional grid underlay for a canvas document view.
///
/// Product-agnostic — any graph host can use it or supply different document
/// content to `GraphCanvasScrollView`.
struct GraphCanvasGridView: View {
    /// Size of the pannable document in content points.
    var contentSize: CGSize
    /// Spacing between grid lines in content points (future snap unit).
    var gridSpacing: CGFloat = GraphCanvasGridMapping.defaultSpacing

    var body: some View {
        Canvas { context, size in
            let lineColor = PVColor.borderSubtle.opacity(0.7)
            var path = Path()
            var x: CGFloat = 0
            while x <= size.width {
                path.move(to: CGPoint(x: x, y: 0))
                path.addLine(to: CGPoint(x: x, y: size.height))
                x += gridSpacing
            }
            var y: CGFloat = 0
            while y <= size.height {
                path.move(to: CGPoint(x: 0, y: y))
                path.addLine(to: CGPoint(x: size.width, y: y))
                y += gridSpacing
            }
            context.stroke(path, with: .color(lineColor), lineWidth: 1)
        }
        .frame(width: contentSize.width, height: contentSize.height)
        .background(PVColor.surfacePage)
    }
}
