import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct EvidenceGraphPlacementTests {
    @Test func constantsMatchFormulas() {
        #expect(EvidenceGraphPlacement.horizontalStep == 8)
        #expect(EvidenceGraphPlacement.verticalStep == 4)
        #expect(EvidenceGraphPlacement.minCell.gridX == 4)
        #expect(EvidenceGraphPlacement.minCell.gridY == 1)
        #expect(EvidenceGraphPlacement.maxCell.gridX == 96)
        #expect(EvidenceGraphPlacement.maxCell.gridY == 98)
    }

    @Test func emptyGraphLandsAtFourThree() {
        let snapshot = SourceGraphSnapshot(sourceId: "src")
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX == 4)
        #expect(slot.cell.gridY == 3)
        #expect(slot.landingColumn == 4)
    }

    @Test func firstCreateIsRightOfRightmostAndLevelWithTop() {
        let snapshot = snapshot(
            subjects: [placedSubject(id: "p1", x: 0, y: 2)],
            bridges: [placedBridge(id: "b1", x: 8, y: 6)]
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX == 16)
        #expect(slot.cell.gridY == 2)
        #expect(slot.landingColumn == 16)
        #expect(framesIntersect(slot.cell, existing: snapshot) == false)
    }

    @Test func newCardDoesNotIntersectSubjectOrBridge() {
        let subjectRight = snapshot(subjects: [placedSubject(id: "p1", x: 10, y: 4)], bridges: [])
        let subjectSlot = EvidenceGraphPlacement.composerSlot(in: subjectRight, landingColumn: nil)
        #expect(framesIntersect(subjectSlot.cell, existing: subjectRight) == false)

        let bridgeRight = snapshot(subjects: [], bridges: [placedBridge(id: "b1", x: 12, y: 5)])
        let bridgeSlot = EvidenceGraphPlacement.composerSlot(in: bridgeRight, landingColumn: nil)
        #expect(framesIntersect(bridgeSlot.cell, existing: bridgeRight) == false)
    }

    @Test func secondCreateStacksInLandingColumn() {
        let first = snapshot(subjects: [placedSubject(id: "p1", x: 0, y: 1)], bridges: [])
        let slot1 = EvidenceGraphPlacement.composerSlot(in: first, landingColumn: nil)
        let second = snapshot(
            subjects: [
                placedSubject(id: "p1", x: 0, y: 1),
                placedSubject(id: "p2", x: slot1.cell.gridX, y: slot1.cell.gridY),
            ],
            bridges: []
        )
        let slot2 = EvidenceGraphPlacement.composerSlot(in: second, landingColumn: slot1.landingColumn)
        #expect(slot2.cell.gridX == slot1.landingColumn)
        #expect(slot2.cell.gridY == slot1.cell.gridY + 4)
    }

    @Test func columnPastTheOldFloorGrowsDownward() {
        var subjects: [SourceGraphPlacedSubject] = [placedSubject(id: "origin", x: 0, y: 1)]
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
            subjects: [placedSubject(id: "edge", x: 90, y: 1)],
            bridges: []
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: nearEdge, landingColumn: nil)
        #expect(slot.cell.gridX == 98)
        let landing = EvidenceGraphPlacement.subjectFrame(gridX: slot.cell.gridX, gridY: slot.cell.gridY)
        let size = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(in: nearEdge) + [landing]
        )
        #expect(size.width == EvidenceGraphPlacement.contentSize.width + EvidenceGraphPlacement.growthStep)
        let limit = EvidenceGraphPlacement.maxCell(fitting: size)
        #expect(slot.cell.gridX <= limit.gridX)
        #expect(limit.gridX > EvidenceGraphPlacement.maxCell.gridX)
    }

    @Test func furthestCardOpensOnALargerDocument() {
        let packed = snapshot(
            subjects: [placedSubject(id: "edge", x: 96, y: 98)],
            bridges: []
        )
        let size = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(in: packed)
        )
        #expect(size.width == 4_800)
        #expect(size.height == 4_800)
        let limit = EvidenceGraphPlacement.maxCell(fitting: size)
        #expect(limit.gridX > 96)
        #expect(limit.gridY > 98)
        #expect(EvidenceGraphPlacement.maxCell.gridX == 96)
        #expect(EvidenceGraphPlacement.maxCell.gridY == 98)
    }

    @Test func smallGraphStaysOnTheBaseDocument() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: 4, y: 3)], bridges: [])
        let size = EvidenceGraphPlacement.contentSize(
            fitting: EvidenceGraphPlacement.cardFrames(in: snapshot)
        )
        #expect(size == EvidenceGraphPlacement.contentSize)
    }

    @Test func dragTowardTheEdgeGrowsBeforeTheDrop() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: 80, y: 4)], bridges: [])
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

    @Test func resultStaysInsideMinMax() {
        let snapshot = snapshot(subjects: [placedSubject(id: "p1", x: 4, y: 1)], bridges: [])
        let slot = EvidenceGraphPlacement.composerSlot(in: snapshot, landingColumn: nil)
        #expect(slot.cell.gridX >= 4 && slot.cell.gridX <= 96)
        #expect(slot.cell.gridY >= 1 && slot.cell.gridY <= 98)
    }

    private func snapshot(
        subjects: [SourceGraphPlacedSubject],
        bridges: [SourceGraphPlacedBridge]
    ) -> SourceGraphSnapshot {
        SourceGraphSnapshot(sourceId: "src", subjects: subjects, bridges: bridges)
    }

    private func placedSubject(id: String, x: Int64, y: Int64) -> SourceGraphPlacedSubject {
        SourceGraphPlacedSubject(
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

    private func placedBridge(id: String, x: Int64, y: Int64) -> SourceGraphPlacedBridge {
        SourceGraphPlacedBridge(
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

    private func framesIntersect(_ cell: CatalogGridCell, existing: SourceGraphSnapshot) -> Bool {
        let incoming = cardFrame(x: cell.gridX, y: cell.gridY)
        let others = existing.subjects.map { cardFrame(x: $0.gridX, y: $0.gridY) }
            + existing.bridges.map { cardFrame(x: $0.gridX, y: $0.gridY) }
        return others.contains { $0.intersects(incoming) }
    }

    private func cardFrame(x: Int64, y: Int64) -> CGRect {
        let center = GraphCanvasGridMapping.contentPoint(gridX: x, gridY: y)
        return CGRect(
            x: center.x - EvidenceSubjectCard.width / 2,
            y: center.y - EvidenceSubjectCard.approximateHalfHeight,
            width: EvidenceSubjectCard.width,
            height: 2 * EvidenceSubjectCard.approximateHalfHeight
        )
    }
}
