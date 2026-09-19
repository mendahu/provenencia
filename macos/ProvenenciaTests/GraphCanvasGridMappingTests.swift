import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasGridMappingTests {
    @Test func contentPointCentersCellAtHalfSpacing() {
        let point = GraphCanvasGridMapping.contentPoint(gridX: 0, gridY: 0, spacing: 40)
        #expect(point.x == 20)
        #expect(point.y == 20)

        let next = GraphCanvasGridMapping.contentPoint(gridX: 2, gridY: 3, spacing: 40)
        #expect(next.x == 100)
        #expect(next.y == 140)
    }

    @Test func gridCellFloorsContentPoint() {
        let cell = GraphCanvasGridMapping.gridCell(
            contentPoint: CGPoint(x: 99, y: 141),
            spacing: 40
        )
        #expect(cell.gridX == 2)
        #expect(cell.gridY == 3)
    }

    @Test func roundTripNearestCell() {
        let origin = GraphCanvasGridMapping.contentPoint(gridX: 5, gridY: 7)
        let cell = GraphCanvasGridMapping.gridCell(contentPoint: origin)
        #expect(cell.gridX == 5)
        #expect(cell.gridY == 7)
    }
}
