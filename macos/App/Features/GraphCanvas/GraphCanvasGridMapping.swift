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
        (saturatingFloor(gridPoint.x / spacing), saturatingFloor(gridPoint.y / spacing))
    }

    /// `cell` moved by `delta`, held at `Int64.min` / `Int64.max` instead of trapping.
    static func offsetCell(_ cell: Int64, by delta: Int64) -> Int64 {
        let (sum, overflow) = cell.addingReportingOverflow(delta)
        guard overflow else { return sum }
        return delta < 0 ? Int64.min : Int64.max
    }

    /// `floor(value)` as an `Int64`: NaN is 0, and a value past the `Int64` range
    /// (infinity included) is `Int64.min` / `Int64.max`.
    static func saturatingFloor(_ value: CGFloat) -> Int64 {
        if value.isNaN { return 0 }
        let floored = floor(Double(value))
        if floored >= 9_223_372_036_854_775_808.0 { return Int64.max }
        if floored < -9_223_372_036_854_775_808.0 { return Int64.min }
        return Int64(floored)
    }
}
