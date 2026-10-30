import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasGridMappingTests {
    @Test func gridPointCentersCellAtHalfSpacing() {
        let point = GraphCanvasGridMapping.gridPoint(gridX: 0, gridY: 0, spacing: 40)
        #expect(point.x == 20)
        #expect(point.y == 20)

        let next = GraphCanvasGridMapping.gridPoint(gridX: 2, gridY: 3, spacing: 40)
        #expect(next.x == 100)
        #expect(next.y == 140)
    }

    @Test func gridCellFloorsGridPoint() {
        let cell = GraphCanvasGridMapping.gridCell(
            gridPoint: CGPoint(x: 99, y: 141),
            spacing: 40
        )
        #expect(cell.gridX == 2)
        #expect(cell.gridY == 3)
    }

    @Test func roundTripNearestCell() {
        let origin = GraphCanvasGridMapping.gridPoint(gridX: 5, gridY: 7)
        let cell = GraphCanvasGridMapping.gridCell(gridPoint: origin)
        #expect(cell.gridX == 5)
        #expect(cell.gridY == 7)
    }

    @Test func saturatesExtremeCells() {
        #expect(GraphCanvasGridMapping.offsetCell(Int64.max, by: 1) == Int64.max)
        #expect(GraphCanvasGridMapping.offsetCell(Int64.min, by: -1) == Int64.min)
        #expect(GraphCanvasGridMapping.offsetCell(5, by: -7) == -2)

        let infinite = GraphCanvasGridMapping.gridCell(gridPoint: CGPoint(x: CGFloat.infinity, y: CGFloat.nan))
        #expect(infinite.gridX == Int64.max)
        #expect(infinite.gridY == 0)
        let far = GraphCanvasGridMapping.gridCell(gridPoint: CGPoint(x: -1e30, y: 40))
        #expect(far.gridX == Int64.min)
        #expect(far.gridY == 1)

        #expect(GraphCanvasGridMapping.saturatingFloor(.nan) == 0)
        #expect(GraphCanvasGridMapping.saturatingFloor(2.5) == 2)
        #expect(GraphCanvasGridMapping.saturatingFloor(-2.5) == -3)
    }
}
