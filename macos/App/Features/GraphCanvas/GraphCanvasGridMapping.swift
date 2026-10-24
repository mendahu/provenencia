import CoreGraphics
import Foundation

/// Pure grid-cell ↔ grid-point conversion for canvas documents.
///
/// Matches `GraphCanvasGridView` spacing so snap and placement share one
/// unit. Product hosts call these helpers; keep catalog types out of this
/// module.
enum GraphCanvasGridMapping {
    /// Default spacing — keep in sync with `GraphCanvasGridView.gridSpacing`.
    static let defaultSpacing: CGFloat = 40

    /// Grid-space center of the cell at integer `(gridX, gridY)`.
    ///
    /// Cell `(0, 0)` is centered at `(spacing/2, spacing/2)` so an item sits in
    /// the first grid square rather than on the origin corner.
    static func gridPoint(
        gridX: Int64,
        gridY: Int64,
        spacing: CGFloat = defaultSpacing
    ) -> CGPoint {
        CGPoint(
            x: (CGFloat(gridX) + 0.5) * spacing,
            y: (CGFloat(gridY) + 0.5) * spacing
        )
    }

    /// Nearest grid cell for a grid-space point (future place/drag snap).
    static func gridCell(
        gridPoint: CGPoint,
        spacing: CGFloat = defaultSpacing
    ) -> (gridX: Int64, gridY: Int64) {
        let x = Int64(floor(gridPoint.x / spacing))
        let y = Int64(floor(gridPoint.y / spacing))
        return (x, y)
    }
}
