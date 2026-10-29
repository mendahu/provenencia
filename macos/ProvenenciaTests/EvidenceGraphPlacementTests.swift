import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct EvidenceGraphPlacementTests {
    @Test func constantsMatchFormulas() {
        #expect(EvidenceGraphPlacement.horizontalStep == 8)
        #expect(EvidenceGraphPlacement.verticalStep == 4)
        #expect(EvidenceGraphPlacement.minCell.gridX == 104)
        #expect(EvidenceGraphPlacement.minCell.gridY == 101)
        #expect(EvidenceGraphPlacement.maxCell.gridX == 196)
        #expect(EvidenceGraphPlacement.maxCell.gridY == 198)
    }

    @Test func emptyGraphLandsAtFourThree() {
        let snapshot = EvidenceGraphSnapshot(graphId: "src")
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX == 104)
        #expect(slot.cell.gridY == 103)
        #expect(slot.landingColumn == 104)
    }

    @Test func firstCreateIsRightOfRightmostAndLevelWithTop() {
        let snapshot = snapshot(
            subjects: [placedSubject(id: "p1", x: downtown(0), y: downtown(2))],
            bridges: [placedBridge(id: "b1", x: downtown(8), y: downtown(6))]
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX == 116)
        #expect(slot.cell.gridY == 102)
        #expect(slot.landingColumn == 116)
        #expect(framesIntersect(slot.cell, existing: snapshot) == false)
    }

    @Test func newCardDoesNotIntersectSubjectOrBridge() {
        let subjectRight = snapshot(subjects: [placedSubject(id: "p1", x: downtown(10), y: downtown(4))], bridges: [])
        let subjectSlot = EvidenceGraphPlacement.composerSlot(in: subjectRight, landingColumn: nil)
        #expect(framesIntersect(subjectSlot.cell, existing: subjectRight) == false)

        let bridgeRight = snapshot(subjects: [], bridges: [placedBridge(id: "b1", x: downtown(12), y: downtown(5))])
        let bridgeSlot = EvidenceGraphPlacement.composerSlot(in: bridgeRight, landingColumn: nil)
        #expect(framesIntersect(bridgeSlot.cell, existing: bridgeRight) == false)
    }

    @Test func secondCreateStacksInLandingColumn() {
        let first = snapshot(subjects: [placedSubject(id: "p1", x: downtown(0), y: downtown(1))], bridges: [])
        let slot1 = EvidenceGraphPlacement.composerSlot(in: first, landingColumn: nil)
        let second = snapshot(
            subjects: [
                placedSubject(id: "p1", x: downtown(0), y: downtown(1)),
                placedSubject(id: "p2", x: slot1.cell.gridX, y: slot1.cell.gridY),
            ],
            bridges: []
        )
        let slot2 = EvidenceGraphPlacement.composerSlot(in: second, landingColumn: slot1.landingColumn)
        #expect(slot2.cell.gridX == slot1.landingColumn)
        #expect(slot2.cell.gridY == slot1.cell.gridY + 4)
    }

    @Test func columnPastTheOldFloorGrowsDownward() {
        var subjects: [EvidenceGraphPlacedSubject] = [placedSubject(id: "origin", x: downtown(40), y: downtown(50))]
        let first = EvidenceGraphPlacement.composerSlot(
            in: snapshot(subjects: subjects, bridges: []),
            landingColumn: nil
        )
        var y = first.cell.gridY
        var lastY = y
        while y <= EvidenceGraphPlacement.maxCell.gridY {
            subjects.append(placedSubject(id: "col-\(y)", x: first.landingColumn, y: y))
            lastY = y
            y += EvidenceGraphPlacement.verticalStep
        }
        let grown = snapshot(subjects: subjects, bridges: [])
        let next = EvidenceGraphPlacement.composerSlot(
            in: grown,
            landingColumn: first.landingColumn
        )
        #expect(next.cell.gridX == first.landingColumn)
        #expect(next.cell.gridY == lastY + EvidenceGraphPlacement.verticalStep)
        #expect(next.cell.gridY > EvidenceGraphPlacement.maxCell.gridY)
        let size = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(in: grown)
        )
        #expect(size.height == EvidenceGraphPlacement.contentSize.height + EvidenceGraphPlacement.growthStep)
        #expect(size.width == EvidenceGraphPlacement.contentSize.width)
    }

    @Test func cardNearTheRightEdgeLandsPastTheOldMax() {
        let nearEdge = snapshot(
            subjects: [placedSubject(id: "edge", x: downtown(90), y: downtown(50))],
            bridges: []
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: nearEdge, landingColumn: nil)
        #expect(slot.cell.gridX == 198)
        let landing = EvidenceGraphPlacement.subjectFrame(gridX: slot.cell.gridX, gridY: slot.cell.gridY)
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: nearEdge) + [landing]
        )
        #expect(board.size.width == EvidenceGraphPlacement.contentSize.width + EvidenceGraphPlacement.growthStep)
        #expect(board.origin.x == EvidenceGraphPlacement.originReserve)
        let limit = EvidenceGraphPlacement.maxCell(on: board)
        #expect(slot.cell.gridX <= limit.gridX)
        #expect(limit.gridX > EvidenceGraphPlacement.maxCell.gridX)
    }

    @Test func furthestCardOpensOnALargerDocument() {
        let packed = snapshot(
            subjects: [placedSubject(id: "edge", x: downtown(96), y: downtown(98))],
            bridges: []
        )
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: packed)
        )
        #expect(board.size.width == 4_800)
        #expect(board.size.height == 4_800)
        #expect(board.origin == EvidenceGraphCanvas.downtown.origin)
        let limit = EvidenceGraphPlacement.maxCell(on: board)
        #expect(limit.gridX > 196)
        #expect(limit.gridY > 198)
        #expect(EvidenceGraphPlacement.maxCell.gridX == 196)
        #expect(EvidenceGraphPlacement.maxCell.gridY == 198)
    }

    @Test func smallGraphStaysOnTheBaseDocument() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: downtown(50), y: downtown(50))], bridges: [])
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(board == EvidenceGraphCanvas.downtown)
    }

    @Test func dragTowardTheEdgeGrowsBeforeTheDrop() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: downtown(80), y: downtown(50))], bridges: [])
        let atRest = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(atRest == EvidenceGraphPlacement.contentSize)
        let dragged = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(
                in: snapshot,
                dragOffsets: ["p1": CGSize(width: 700, height: 0)]
            )
        )
        #expect(dragged.width == EvidenceGraphPlacement.contentSize.width + EvidenceGraphPlacement.growthStep)
    }

    @Test func cardAtTheDowntownMinimumOpensOneStepIntoTheReserve() {
        let snapshot = snapshot(subjects: [placedSubject(id: "edge", x: downtown(4), y: downtown(50))], bridges: [])
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(board.origin.x == EvidenceGraphPlacement.originReserve - EvidenceGraphPlacement.growthStep)
        #expect(board.origin.y == EvidenceGraphPlacement.originReserve)
        #expect(board.size.width == EvidenceGraphPlacement.contentSize.width + EvidenceGraphPlacement.growthStep)
        #expect(board.maxX == EvidenceGraphPlacement.originReserve + EvidenceGraphPlacement.contentSize.width)
    }

    @Test func cardNearTheTopEdgePullsTheOriginUp() {
        let snapshot = snapshot(subjects: [placedSubject(id: "edge", x: downtown(50), y: downtown(1))], bridges: [])
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(board.origin.y == EvidenceGraphPlacement.originReserve - EvidenceGraphPlacement.growthStep)
        #expect(board.origin.x == EvidenceGraphPlacement.originReserve)
        #expect(board.size.height == EvidenceGraphPlacement.contentSize.height + EvidenceGraphPlacement.growthStep)
        #expect(board.maxY == EvidenceGraphPlacement.originReserve + EvidenceGraphPlacement.contentSize.height)
    }

    @Test func dragTowardTheLeftEdgeGrowsBeforeTheDrop() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: downtown(12), y: downtown(50))], bridges: [])
        let atRest = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(atRest.origin.x == EvidenceGraphPlacement.originReserve)
        let dragged = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(
                in: snapshot,
                dragOffsets: ["p1": CGSize(width: -200, height: 0)]
            )
        )
        #expect(dragged.origin.x == EvidenceGraphPlacement.originReserve - EvidenceGraphPlacement.growthStep)
        #expect(dragged.size.width == EvidenceGraphPlacement.contentSize.width + EvidenceGraphPlacement.growthStep)
    }

    @Test func originStopsAtZero() {
        let snapshot = snapshot(subjects: [placedSubject(id: "far", x: downtown(-100), y: downtown(50))], bridges: [])
        let board = EvidenceGraphPlacement.canvas(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(board.origin.x == 0)
        #expect(board.origin.y == EvidenceGraphPlacement.originReserve)
        #expect(board.maxX == EvidenceGraphPlacement.originReserve + EvidenceGraphPlacement.contentSize.width)
    }

    @Test func resultStaysInsideMinMax() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: downtown(4), y: downtown(1))], bridges: [])
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX >= EvidenceGraphPlacement.minCell.gridX)
        #expect(slot.cell.gridX <= EvidenceGraphPlacement.maxCell.gridX)
        #expect(slot.cell.gridY >= EvidenceGraphPlacement.minCell.gridY)
        #expect(slot.cell.gridY <= EvidenceGraphPlacement.maxCell.gridY)
    }

    @Test func extremePositionsDoNotTrap() {
        let limit = EvidenceGraphPlacement.cellLimit
        let snapshot = snapshot(
            subjects: [placedSubject(id: "p1", x: Int64.max, y: Int64.min)],
            bridges: [placedBridge(id: "b1", x: Int64.min, y: Int64.max)]
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        let stacked = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: slot.landingColumn)
        let canvas = EvidenceGraphPlacement.canvas(fitting: EvidenceGraphPlacement.cardFrames(in: snapshot))
        let maxCell = EvidenceGraphPlacement.maxCell(on: canvas)
        for cell in [slot.cell, stacked.cell, maxCell] {
            #expect(abs(cell.gridX) <= limit)
            #expect(abs(cell.gridY) <= limit)
        }
        #expect(abs(slot.landingColumn) <= limit)
        #expect(EvidenceGraphPlacement.clampedCell(Int64.max) == limit)
        #expect(EvidenceGraphPlacement.clampedCell(Int64.min) == -limit)
        #expect(EvidenceGraphPlacement.clampedCell(7) == 7)
    }

    /// Hand-written cells are the pre-reserve layout. Composer results are already absolute.
    private func downtown(_ cell: Int64) -> Int64 {
        cell + EvidenceGraphPlacement.originReserveCells
    }

    private func snapshot(
        subjects: [EvidenceGraphPlacedSubject],
        bridges: [EvidenceGraphPlacedBridge]
    ) -> EvidenceGraphSnapshot {
        EvidenceGraphSnapshot(graphId: "src", subjects: subjects, bridges: bridges)
    }

    private func placedSubject(id: String, x: Int64, y: Int64) -> EvidenceGraphPlacedSubject {
        EvidenceGraphPlacedSubject(
            subject: CatalogSubject(
                id: id, ref: "CPR-1", sourceID: "src", subjectTypeID: "type-person",
                label: id, description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: x,
            gridY: y,
            isCited: false
        )
    }

    private func placedBridge(id: String, x: Int64, y: Int64) -> EvidenceGraphPlacedBridge {
        EvidenceGraphPlacedBridge(
            subject: CatalogSubject(
                id: id, ref: "PTN-1", sourceID: "src", subjectTypeID: "type-participation",
                label: id, description: ""
            ),
            kind: .participation,
            typeLabel: "Participation",
            gridX: x,
            gridY: y
        )
    }

    private func framesIntersect(_ cell: CatalogGridCell, existing: EvidenceGraphSnapshot) -> Bool {
        let incoming = cardFrame(x: cell.gridX, y: cell.gridY)
        let others = existing.subjects.map { cardFrame(x: $0.gridX, y: $0.gridY) }
            + existing.bridges.map { cardFrame(x: $0.gridX, y: $0.gridY) }
        return others.contains { $0.intersects(incoming) }
    }

    private func cardFrame(x: Int64, y: Int64) -> CGRect {
        let center = GraphCanvasGridMapping.gridPoint(gridX: x, gridY: y)
        return CGRect(
            x: center.x - EvidenceSubjectCard.width / 2,
            y: center.y - EvidenceSubjectCard.approximateHalfHeight,
            width: EvidenceSubjectCard.width,
            height: 2 * EvidenceSubjectCard.approximateHalfHeight
        )
    }
}
