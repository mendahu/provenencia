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

    @Test func segmentEndpointsLieInsideRectsAfterTuck() {
        let a = CGRect(x: 0, y: 0, width: 100, height: 50)
        let b = CGRect(x: 200, y: 100, width: 100, height: 50)
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: a, toRect: b)
        // CGRect.contains is max-exclusive; require a clear inset past the border.
        #expect(seg.start.x > a.minX + 1 && seg.start.x < a.maxX - 1)
        #expect(seg.start.y > a.minY + 1 && seg.start.y < a.maxY - 1)
        #expect(seg.end.x > b.minX + 1 && seg.end.x < b.maxX - 1)
        #expect(seg.end.y > b.minY + 1 && seg.end.y < b.maxY - 1)
        #expect(seg.path.isEmpty == false)
    }

    @Test func bottomEdgeTuckMovesStraightUp() {
        let rect = CGRect(x: 0, y: 0, width: 100, height: 60)
        let onBottom = CGPoint(x: 70, y: 60)
        let tucked = GraphCanvasEdgeGeometry.tuckInside(onBottom, rect: rect, distance: 16)
        #expect(abs(tucked.x - 70) < 0.01)
        #expect(abs(tucked.y - 44) < 0.01)
        #expect(rect.contains(tucked))
    }

    @Test func diagonalApproachStillTucksPastBottomEdge() {
        let a = CGRect(x: 0, y: 0, width: 100, height: 60)
        let b = CGRect(x: 80, y: 200, width: 100, height: 60)
        let seg = GraphCanvasEdgeGeometry.segment(fromRect: a, toRect: b)
        // Start was on/near bottom of A — after tuck it must sit above maxY.
        #expect(seg.start.y < a.maxY - 8)
        #expect(a.contains(seg.start))
    }

    @Test func primaryEdgeFrameSharesLayoutTop() {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 2,
            gridY: 3,
            isCited: false
        )
        let frame = EvidenceSubjectCard.edgeFrame(for: placed)
        let offset = EvidenceSubjectCard.topLeadingOffset(gridX: 2, gridY: 3)
        #expect(abs(frame.minX - offset.width) < 0.01)
        #expect(abs(frame.minY - offset.height) < 0.01)
        #expect(frame.height >= EvidenceSubjectCard.edgeLayoutHeight)
        #expect(frame.width == EvidenceSubjectCard.width)
    }
}