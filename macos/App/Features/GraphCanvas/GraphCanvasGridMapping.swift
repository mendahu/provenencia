import CoreGraphics
import Foundation

/// Pure grid-cell ↔ content-point conversion for canvas documents.
///
/// Matches `GraphCanvasGridView` spacing so snap and card placement share one
/// unit. Product hosts (Evidence graph, later trees) call these helpers; keep
/// catalog types out of this module.
enum GraphCanvasGridMapping {
    /// Default spacing — keep in sync with `GraphCanvasGridView.gridSpacing`.
    static let defaultSpacing: CGFloat = 40

    /// Content-space center of the cell at integer `(gridX, gridY)`.
    ///
    /// Cell `(0, 0)` is centered at `(spacing/2, spacing/2)` so a card sits in
    /// the first grid square rather than on the origin corner.
    static func contentPoint(
        gridX: Int64,
        gridY: Int64,
        spacing: CGFloat = defaultSpacing
    ) -> CGPoint {
        CGPoint(
            x: (CGFloat(gridX) + 0.5) * spacing,
            y: (CGFloat(gridY) + 0.5) * spacing
        )
    }

    /// Nearest grid cell for a content-space point (future place/drag snap).
    static func gridCell(
        contentPoint: CGPoint,
        spacing: CGFloat = defaultSpacing
    ) -> (gridX: Int64, gridY: Int64) {
        let x = Int64(floor(contentPoint.x / spacing))
        let y = Int64(floor(contentPoint.y / spacing))
        return (x, y)
    }
}
