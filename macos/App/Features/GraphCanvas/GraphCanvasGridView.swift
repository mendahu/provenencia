import SwiftUI

/// Empty content plane drawn under magnification — grid only, no bubbles (S6-01).
struct GraphCanvasGridView: View {
    /// Size of the pannable document in content points.
    var contentSize: CGSize
    /// Spacing between grid lines in content points (future snap unit).
    var gridSpacing: CGFloat = 40

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
