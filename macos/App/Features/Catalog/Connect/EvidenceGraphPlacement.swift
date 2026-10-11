import CoreGraphics
import Foundation

/// Pure grid placement for composer-created subjects. No store or session access.
enum EvidenceGraphPlacement {
    /// Resting document. A graph grows past this only when a card approaches an edge.
    static let contentSize = CGSize(width: 4_000, height: 4_000)
    /// One growth step: 20 grid cells.
    static let growthStep: CGFloat = 800

    static var spacing: CGFloat { GraphCanvasGridMapping.defaultSpacing }

    /// Keep in sync with `EvidenceSubjectCard.width` / `approximateHalfHeight`.
    static let subjectCardWidth: CGFloat = 264
    static let subjectCardHalfHeight: CGFloat = 36

    /// Clears the widest card plus one cell of gap.
    static var horizontalStep: Int64 {
        Int64(ceil((subjectCardWidth + spacing) / spacing))
    }

    /// One new card plus room to grow.
    static var verticalStep: Int64 {
        Int64(ceil((2 * subjectCardHalfHeight + 2 * spacing) / spacing))
    }

    static var minCell: CatalogGridCell {
        CatalogGridCell(gridX: minGridX, gridY: minGridY)
    }

    static var maxCell: CatalogGridCell {
        maxCell(fitting: contentSize)
    }

    /// Largest cell whose card still sits inside `size`.
    static func maxCell(fitting size: CGSize) -> CatalogGridCell {
        CatalogGridCell(
            gridX: maxGridX(documentWidth: size.width),
            gridY: maxGridY(documentHeight: size.height)
        )
    }

    /// Document that keeps every frame at least one card-width inside the edge.
    /// Starts at ``contentSize`` and extends an edge by ``growthStep`` until
    /// that card is clear. A frame already inside the margin does not grow it.
    static func contentSize(fitting frames: [CGRect]) -> CGSize {
        var width = contentSize.width
        var height = contentSize.height
        let margin = subjectCardWidth
        for frame in frames {
            if width - frame.maxX < margin {
                let needed = frame.maxX + margin
                let steps = ceil((needed - contentSize.width) / growthStep)
                width = max(width, contentSize.width + steps * growthStep)
            }
            if height - frame.maxY < margin {
                let needed = frame.maxY + margin
                let steps = ceil((needed - contentSize.height) / growthStep)
                height = max(height, contentSize.height + steps * growthStep)
            }
        }
        return CGSize(width: width, height: height)
    }

    static func subjectFrame(gridX: Int64, gridY: Int64, dragOffset: CGSize = .zero) -> CGRect {
        cardFrame(
            gridX: gridX,
            gridY: gridY,
            width: subjectCardWidth,
            halfHeight: subjectCardHalfHeight,
            dragOffset: dragOffset
        )
    }

    static func bridgeFrame(gridX: Int64, gridY: Int64, dragOffset: CGSize = .zero) -> CGRect {
        cardFrame(
            gridX: gridX,
            gridY: gridY,
            width: EvidenceBridgeCard.width,
            halfHeight: EvidenceBridgeCard.approximateHalfHeight,
            dragOffset: dragOffset
        )
    }

    static func cardFrames(
        in snapshot: SourceGraphSnapshot,
        dragOffsets: [String: CGSize] = [:]
    ) -> [CGRect] {
        snapshot.subjects.map {
            subjectFrame(gridX: $0.gridX, gridY: $0.gridY, dragOffset: dragOffsets[$0.id] ?? .zero)
        } + snapshot.bridges.map {
            bridgeFrame(gridX: $0.gridX, gridY: $0.gridY, dragOffset: dragOffsets[$0.id] ?? .zero)
        }
    }

    /// Next empty landing cell to the right of the graph, stacking in one column.
    /// The document grows so that cell is not clamped to the 4000-point board.
    static func composerSlot(
        in snapshot: SourceGraphSnapshot,
        landingColumn: Int64?
    ) -> (cell: CatalogGridCell, landingColumn: Int64) {
        let placed = placedCells(in: snapshot)
        if placed.isEmpty {
            let cell = CatalogGridCell(gridX: minCell.gridX, gridY: minCell.gridY + 2)
            return (cell, minCell.gridX)
        }

        var column = landingColumn ?? placed.map(\.gridX).max()! + horizontalStep
        var row = rowInColumn(column, placed: placed)
        let landing = subjectFrame(gridX: column, gridY: row)
        let size = contentSize(fitting: cardFrames(in: snapshot) + [landing])
        let limit = maxCell(fitting: size)
        if row > limit.gridY, column < limit.gridX {
            column = min(column + horizontalStep, limit.gridX)
            row = rowInColumn(column, placed: placed)
        }
        column = min(column, limit.gridX)
        if row > limit.gridY {
            row = limit.gridY
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

    /// Largest x such that the card's right edge stays inside `documentWidth`.
    private static func maxGridX(documentWidth: CGFloat) -> Int64 {
        let width = EvidenceSubjectCard.width
        let raw = (documentWidth - width / 2) / spacing - 0.5
        return Int64(floor(raw + 1e-9))
    }

    /// Largest y such that the card's bottom stays inside `documentHeight`.
    private static func maxGridY(documentHeight: CGFloat) -> Int64 {
        let half = EvidenceSubjectCard.approximateHalfHeight
        let raw = (documentHeight - half) / spacing - 0.5
        return Int64(floor(raw + 1e-9))
    }

    private static func cardFrame(
        gridX: Int64,
        gridY: Int64,
        width: CGFloat,
        halfHeight: CGFloat,
        dragOffset: CGSize
    ) -> CGRect {
        let center = GraphCanvasGridMapping.contentPoint(gridX: gridX, gridY: gridY)
        return CGRect(
            x: center.x - width / 2 + dragOffset.width,
            y: center.y - halfHeight + dragOffset.height,
            width: width,
            height: halfHeight * 2
        )
    }
}
