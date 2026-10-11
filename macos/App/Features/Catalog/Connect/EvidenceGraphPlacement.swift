import CoreGraphics
import Foundation

/// The scroll document for one evidence graph: a grid-space origin and a size.
///
/// Coordinates below ``origin`` are the unrendered reserve. The document's
/// top-left is `origin`, so a card at grid point `p` is drawn at `p - origin`.
struct EvidenceGraphCanvas: Equatable, Sendable {
    var origin: CGPoint
    var size: CGSize

    var maxX: CGFloat { origin.x + size.width }
    var maxY: CGFloat { origin.y + size.height }

    static let downtown = EvidenceGraphCanvas(
        origin: CGPoint(
            x: EvidenceGraphPlacement.originReserve,
            y: EvidenceGraphPlacement.originReserve
        ),
        size: EvidenceGraphPlacement.contentSize
    )
}

/// Pure grid placement for composer-created subjects. No store or session access.
enum EvidenceGraphPlacement {
    /// Resting downtown document. A graph grows past this when a card approaches an edge.
    static let contentSize = CGSize(width: 4_000, height: 4_000)
    /// One growth step: 20 grid cells.
    static let growthStep: CGFloat = 800
    /// One resting screen, held back from grid zero so the document can grow left and up.
    /// Stored positions are shifted by ``originReserveCells`` to match.
    static let originReserve: CGFloat = 4_000

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

    /// Grid cells in ``originReserve`` (40-point spacing).
    static var originReserveCells: Int64 {
        Int64(originReserve / spacing)
    }

    /// New cards are born downtown, one reserve in from grid zero.
    static var minCell: CatalogGridCell {
        CatalogGridCell(
            gridX: minGridX + originReserveCells,
            gridY: minGridY + originReserveCells
        )
    }

    /// Largest cell on the resting downtown board.
    static var maxCell: CatalogGridCell {
        maxCell(on: .downtown)
    }

    /// Largest cell whose card still sits inside `canvas`.
    static func maxCell(on canvas: EvidenceGraphCanvas) -> CatalogGridCell {
        CatalogGridCell(
            gridX: maxGridX(documentEdge: canvas.maxX),
            gridY: maxGridY(documentEdge: canvas.maxY)
        )
    }

    /// Downtown board, grown by ``growthStep`` on any edge a frame approaches.
    /// Left and top move toward grid zero and stop there. Right and bottom do not.
    static func canvas(fitting frames: [CGRect]) -> EvidenceGraphCanvas {
        var originX = originReserve
        var originY = originReserve
        var maxX = originReserve + contentSize.width
        var maxY = originReserve + contentSize.height
        let margin = subjectCardWidth
        let downtownRight = originReserve + contentSize.width
        let downtownBottom = originReserve + contentSize.height
        for frame in frames {
            if frame.minX - originX < margin {
                let needed = frame.minX - margin
                let steps = ceil((originReserve - needed) / growthStep)
                originX = min(originX, max(0, originReserve - steps * growthStep))
            }
            if frame.minY - originY < margin {
                let needed = frame.minY - margin
                let steps = ceil((originReserve - needed) / growthStep)
                originY = min(originY, max(0, originReserve - steps * growthStep))
            }
            if maxX - frame.maxX < margin {
                let needed = frame.maxX + margin
                let steps = ceil((needed - downtownRight) / growthStep)
                maxX = max(maxX, downtownRight + steps * growthStep)
            }
            if maxY - frame.maxY < margin {
                let needed = frame.maxY + margin
                let steps = ceil((needed - downtownBottom) / growthStep)
                maxY = max(maxY, downtownBottom + steps * growthStep)
            }
        }
        return EvidenceGraphCanvas(
            origin: CGPoint(x: originX, y: originY),
            size: CGSize(width: maxX - originX, height: maxY - originY)
        )
    }

    /// Document size for `frames`. See ``canvas(fitting:)``.
    static func contentSize(fitting frames: [CGRect]) -> CGSize {
        canvas(fitting: frames).size
    }

    /// The smaller origin and the further far edge, so a live drag can extend a fitted board.
    static func union(_ a: EvidenceGraphCanvas, _ b: EvidenceGraphCanvas) -> EvidenceGraphCanvas {
        let minX = min(a.origin.x, b.origin.x)
        let minY = min(a.origin.y, b.origin.y)
        let maxX = max(a.maxX, b.maxX)
        let maxY = max(a.maxY, b.maxY)
        return EvidenceGraphCanvas(
            origin: CGPoint(x: minX, y: minY),
            size: CGSize(width: maxX - minX, height: maxY - minY)
        )
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
        let board = canvas(fitting: cardFrames(in: snapshot) + [landing])
        let limit = maxCell(on: board)
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

    /// Largest x such that the card's right edge stays inside `documentEdge` (grid space).
    private static func maxGridX(documentEdge: CGFloat) -> Int64 {
        let width = EvidenceSubjectCard.width
        let raw = (documentEdge - width / 2) / spacing - 0.5
        return Int64(floor(raw + 1e-9))
    }

    /// Largest y such that the card's bottom stays inside `documentEdge` (grid space).
    private static func maxGridY(documentEdge: CGFloat) -> Int64 {
        let half = EvidenceSubjectCard.approximateHalfHeight
        let raw = (documentEdge - half) / spacing - 0.5
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
