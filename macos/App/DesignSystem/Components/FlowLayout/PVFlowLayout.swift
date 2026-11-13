import SwiftUI

/// Wraps children onto new lines when the row runs out of width. Children are
/// measured at their ideal width; one wider than the row gets the row's width.
struct PVFlowLayout: Layout {
    var spacing: CGFloat
    var lineSpacing: CGFloat

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let width = proposal.width ?? 0
        let result = Self.arrange(
            width: width, spacing: spacing, lineSpacing: lineSpacing, count: subviews.count,
            fit: { subviews[$0].sizeThatFits(ProposedViewSize(width: $1, height: nil)) }
        )
        return CGSize(width: width, height: result.height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let result = Self.arrange(
            width: bounds.width, spacing: spacing, lineSpacing: lineSpacing, count: subviews.count,
            fit: { subviews[$0].sizeThatFits(ProposedViewSize(width: $1, height: nil)) }
        )
        for (index, frame) in result.frames.enumerated() {
            subviews[index].place(
                at: CGPoint(x: bounds.minX + frame.minX, y: bounds.minY + frame.minY),
                anchor: .topLeading,
                proposal: ProposedViewSize(frame.size)
            )
        }
    }

    /// The frames of `count` items in a row `width` wide, and the height they fill. `fit(i, nil)`
    /// is item `i`'s ideal size; `fit(i, w)` its size when offered `w`. An item wider than the
    /// row is measured again at the row's width, and one still wider is placed at the row's
    /// width, so no frame passes the row's edge. A width of 0 or not finite measures nothing
    /// again: 0 puts one item on each line, infinity puts them all on one.
    static func arrange(
        width: CGFloat,
        spacing: CGFloat,
        lineSpacing: CGFloat,
        count: Int,
        fit: (Int, CGFloat?) -> CGSize
    ) -> (frames: [CGRect], height: CGFloat) {
        var lines: [[(index: Int, size: CGSize)]] = []
        var line: [(index: Int, size: CGSize)] = []
        var lineWidth: CGFloat = 0
        for index in 0..<count {
            var size = fit(index, nil)
            if width.isFinite, width > 0, size.width > width {
                size = fit(index, width)
                size.width = min(size.width, width)
            }
            let next = lineWidth == 0 ? size.width : lineWidth + spacing + size.width
            if lineWidth > 0, next > width {
                lines.append(line)
                line = []
                lineWidth = 0
            }
            line.append((index, size))
            lineWidth = lineWidth == 0 ? size.width : lineWidth + spacing + size.width
        }
        if !line.isEmpty {
            lines.append(line)
        }
        var frames = [CGRect](repeating: .zero, count: count)
        var y: CGFloat = 0
        for line in lines {
            let lineHeight = line.reduce(CGFloat(0)) { max($0, $1.size.height) }
            var x: CGFloat = 0
            for item in line {
                frames[item.index] = CGRect(
                    origin: CGPoint(x: x, y: y + (lineHeight - item.size.height) / 2),
                    size: item.size
                )
                x += item.size.width + spacing
            }
            y += lineHeight + lineSpacing
        }
        return (frames, lines.isEmpty ? 0 : y - lineSpacing)
    }
}

#Preview("PVFlowLayout") {
    PVFlowLayout(spacing: PVSpacing.space3, lineSpacing: PVSpacing.space3) {
        ForEach(["Default", "Jones household", "Smith household", "Hendricks family", "Pryke", "Unlinked"], id: \.self) { name in
            Text(verbatim: name)
                .padding(.horizontal, PVSpacing.space4)
                .padding(.vertical, PVSpacing.space2)
                .overlay(RoundedRectangle(cornerRadius: PVRadius.sm).strokeBorder(PVColor.borderDefault))
        }
    }
    .frame(width: 240)
    .padding()
}
