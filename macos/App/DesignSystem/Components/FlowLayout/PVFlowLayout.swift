import SwiftUI

/// Wraps children onto new lines when the row runs out of width. Children are
/// measured at their ideal width, so a caller caps a long child's width.
struct PVFlowLayout: Layout {
    var spacing: CGFloat
    var lineSpacing: CGFloat

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let rows = rows(in: proposal.width ?? 0, subviews: subviews)
        let height = rows.reduce(CGFloat(0)) { $0 + $1.height } + lineSpacing * CGFloat(max(rows.count - 1, 0))
        return CGSize(width: proposal.width ?? 0, height: height)
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let rows = rows(in: bounds.width, subviews: subviews)
        var y = bounds.minY
        for row in rows {
            var x = bounds.minX
            for item in row.items {
                let originY = y + (row.height - item.size.height) / 2
                item.view.place(
                    at: CGPoint(x: x, y: originY),
                    anchor: .topLeading,
                    proposal: ProposedViewSize(item.size)
                )
                x += item.size.width + spacing
            }
            y += row.height + lineSpacing
        }
    }

    private func rows(in width: CGFloat, subviews: Subviews) -> [Row] {
        var rows: [Row] = []
        var items: [Item] = []
        var rowWidth: CGFloat = 0
        var rowHeight: CGFloat = 0
        for view in subviews {
            let size = view.sizeThatFits(.unspecified)
            let next = rowWidth == 0 ? size.width : rowWidth + spacing + size.width
            if rowWidth > 0, next > width {
                rows.append(Row(items: items, height: rowHeight))
                items = []
                rowWidth = 0
                rowHeight = 0
            }
            items.append(Item(view: view, size: size))
            rowWidth = rowWidth == 0 ? size.width : rowWidth + spacing + size.width
            rowHeight = max(rowHeight, size.height)
        }
        if !items.isEmpty {
            rows.append(Row(items: items, height: rowHeight))
        }
        return rows
    }

    private struct Item {
        var view: LayoutSubview
        var size: CGSize
    }

    private struct Row {
        var items: [Item]
        var height: CGFloat
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
