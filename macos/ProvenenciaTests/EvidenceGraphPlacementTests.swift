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

    @Test func fullColumnWraps() {
        var subjects: [SourceGraphPlacedSubject] = [placedSubject(id: "origin", x: 0, y: 1)]
        let first = EvidenceGraphPlacement.composerSlot(
            in: snapshot(subjects: subjects, bridges: []),
            landingColumn: nil
        )
        var y = first.cell.gridY
        while y <= EvidenceGraphPlacement.maxCell.gridY {
            subjects.append(placedSubject(id: "col-\(y)", x: first.landingColumn, y: y))
            y += EvidenceGraphPlacement.verticalStep
        }
        let wrapped = EvidenceGraphPlacement.composerSlot(
            in: snapshot(subjects: subjects, bridges: []),
            landingColumn: first.landingColumn
        )
        #expect(wrapped.cell.gridX == first.landingColumn + 8)
        #expect(wrapped.cell.gridY == 1)
    }

    @Test func rightmostColumnClampsYAndX() {
        let nearEdge = snapshot(
            subjects: [placedSubject(id: "edge", x: 90, y: 1)],
            bridges: []
        )
        let slot = EvidenceGraphPlacement.composerSlot(in: nearEdge, landingColumn: nil)
        #expect(slot.cell.gridX == 96)

        var subjects = [placedSubject(id: "edge", x: 96, y: 1)]
        var y: Int64 = 1
        while y <= EvidenceGraphPlacement.maxCell.gridY {
            subjects.append(placedSubject(id: "full-\(y)", x: 96, y: y))
            y += EvidenceGraphPlacement.verticalStep
        }
        let clamped = EvidenceGraphPlacement.composerSlot(
            in: snapshot(subjects: subjects, bridges: []),
            landingColumn: 96
        )
        #expect(clamped.cell.gridX == 96)
        #expect(clamped.cell.gridY == 98)
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
