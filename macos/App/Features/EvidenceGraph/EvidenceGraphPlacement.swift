import CoreGraphics
import Foundation

/// Pure grid placement for composer-created subjects. No store or session access.
enum EvidenceGraphPlacement {
    static let contentSize = CGSize(width: 4_000, height: 4_000)

    static var spacing: CGFloat { GraphCanvasGridMapping.defaultSpacing }

    /// Clears the widest card plus one cell of gap.
    static var horizontalStep: Int64 {
        Int64(ceil((EvidenceSubjectCard.width + spacing) / spacing))
    }

    /// One new card plus room to grow.
    static var verticalStep: Int64 {
        Int64(ceil((2 * EvidenceSubjectCard.approximateHalfHeight + 2 * spacing) / spacing))
    }

    static var minCell: CatalogGridCell {
        CatalogGridCell(gridX: minGridX, gridY: minGridY)
    }

    static var maxCell: CatalogGridCell {
        CatalogGridCell(gridX: maxGridX, gridY: maxGridY)
    }

    /// Next empty landing cell to the right of the graph, stacking in one column.
    static func composerSlot(
        in snapshot: SourceGraphSnapshot,
        landingColumn: Int64?
    ) -> (cell: CatalogGridCell, landingColumn: Int64) {
        let placed = placedCells(in: snapshot)
        if placed.isEmpty {
            let cell = CatalogGridCell(gridX: minCell.gridX, gridY: minCell.gridY + 2)
            return (cell, minCell.gridX)
        }

        var column = landingColumn ?? min(placed.map(\.gridX).max()! + horizontalStep, maxCell.gridX)
        var row = rowInColumn(column, placed: placed)
        if row > maxCell.gridY, column < maxCell.gridX {
            column = min(column + horizontalStep, maxCell.gridX)
            row = rowInColumn(column, placed: placed)
        }
        if row > maxCell.gridY {
            row = maxCell.gridY
        }
        return (CatalogGridCell(gridX: column, gridY: row), column)
    }

    private static func placedCells(in snapshot: SourceGraphSnapshot) -> [CatalogGridCell] {
        snapshot.subjects.map { CatalogGridCell(gridX: $0.gridX, gridY: $0.gridY) }
            + snapshot.bridges.map { CatalogGridCell(gridX: $0.gridX, gridY: $0.gridY) }
    }

    private static func rowInColumn(_ column: Int64, placed: [CatalogGridCell]) -> Int64 {
        let inColumn = placed.filter { $0.gridX == column }
        if inColumn.isEmpty {
            return max(placed.map(\.gridY).min() ?? minCell.gridY, minCell.gridY)
        }
        return inColumn.map(\.gridY).max()! + verticalStep
    }

    /// Smallest x such that the subject card's left edge stays at least `spacing` from the origin.
    private static var minGridX: Int64 {
        let width = EvidenceSubjectCard.width
        let raw = (spacing + width / 2) / spacing - 0.5
        return Int64(ceil(raw - 1e-9))
    }

    /// Smallest y such that the card's top stays at least `spacing / 2` from the origin.
    private static var minGridY: Int64 {
        let half = EvidenceSubjectCard.approximateHalfHeight
        let raw = (spacing / 2 + half) / spacing - 0.5
        return Int64(ceil(raw - 1e-9))
    }

    /// Largest x such that the card's right edge stays inside `contentSize.width`.
    private static var maxGridX: Int64 {
        let width = EvidenceSubjectCard.width
        let raw = (contentSize.width - width / 2) / spacing - 0.5
        return Int64(floor(raw + 1e-9))
    }

    /// Largest y such that the card's bottom stays inside `contentSize.height`.
    private static var maxGridY: Int64 {
        let half = EvidenceSubjectCard.approximateHalfHeight
        let raw = (contentSize.height - half) / spacing - 0.5
        return Int64(floor(raw + 1e-9))
    }
}
