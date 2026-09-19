import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasEdgeGeometryTests {
    @Test func attachmentPointHitsHorizontalEdge() {
        let rect = CGRect(x: 0, y: 0, width: 100, height: 40)
        let point = GraphCanvasEdgeGeometry.attachmentPoint(
            fromRect: rect,
            toward: CGPoint(x: 200, y: 20)
        )
        #expect(abs(point.x - 100) < 0.01)
        #expect(abs(point.y - 20) < 0.01)
    }

    @Test func attachmentPointHitsVerticalEdge() {
        let rect = CGRect(x: 0, y: 0, width: 80, height: 60)
        let point = GraphCanvasEdgeGeometry.attachmentPoint(
            fromRect: rect,
            toward: CGPoint(x: 40, y: 200)
        )
        #expect(abs(point.x - 40) < 0.01)
        #expect(abs(point.y - 60) < 0.01)
    }

    @Test func segmentEndpointsLieOnRectBoundaries() {
        let a = CGRect(x: 0, y: 0, width: 100, height: 50)
        let b = CGRect(x: 200, y: 100, width: 100, height: 50)
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: a, toRect: b)
        #expect(abs(seg.start.x - 100) < 0.5 || abs(seg.start.y - 50) < 0.5)
        #expect(abs(seg.end.x - 200) < 0.5 || abs(seg.end.y - 100) < 0.5)
        #expect(seg.path.isEmpty == false)
    }

    @Test func cubicControlsBiasHorizontally() {
        let from = CGPoint(x: 0, y: 10)
        let to = CGPoint(x: 100, y: 50)
        let (c1, c2) = GraphCanvasEdgeGeometry.cubicControls(from: from, to: to)
        #expect(c1.y == from.y)
        #expect(c2.y == to.y)
        #expect(abs(c1.x - 50) < 0.01)
        #expect(abs(c2.x - 50) < 0.01)
    }
}
