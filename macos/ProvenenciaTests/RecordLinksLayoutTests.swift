import CoreGraphics
import Testing
@testable import Provenencia

struct RecordLinksLayoutTests {
    private func arc(_ a: Int, _ b: Int, slot: Int = 0, slots: Int = 1) -> RecordLinksArc {
        RecordLinksArc(a: a, b: b, strong: false, slot: slot, slots: slots)
    }

    private func distance(_ p: CGPoint, _ q: CGPoint) -> CGFloat {
        ((p.x - q.x) * (p.x - q.x) + (p.y - q.y) * (p.y - q.y)).squareRoot()
    }

    @Test func fourNodesSitTopRightBottomLeft() {
        let layout = RecordLinksLayout(nodeCount: 4, arcs: [], width: 340)
        #expect(layout.diameter == 172)
        #expect(layout.size == CGSize(width: 340, height: 232))
        #expect(layout.center == CGPoint(x: 170, y: 116))
        let points = layout.nodes.map(\.point)
        let expected = [CGPoint(x: 170, y: 30), CGPoint(x: 256, y: 116), CGPoint(x: 170, y: 202), CGPoint(x: 84, y: 116)]
        for (point, want) in zip(points, expected) {
            #expect(distance(point, want) < 0.001)
        }
        #expect(layout.nodes.map(\.side) == [.top, .right, .bottom, .left])
    }

    @Test func zeroOneAndTwoNodes() {
        #expect(RecordLinksLayout(nodeCount: 0, arcs: [], width: 340).nodes.isEmpty)
        let one = RecordLinksLayout(nodeCount: 1, arcs: [], width: 340)
        #expect(one.nodes.map(\.point) == [CGPoint(x: 170, y: 24)])
        #expect(one.size == CGSize(width: 340, height: 64))
        let two = RecordLinksLayout(nodeCount: 2, arcs: [], width: 340)
        #expect(distance(two.nodes[0].point, CGPoint(x: 170, y: 30)) < 0.001)
        #expect(distance(two.nodes[1].point, CGPoint(x: 170, y: 202)) < 0.001)
        #expect(two.nodes.map(\.side) == [.top, .bottom])
    }

    @Test func eightNodesLabelDiagonalsBesideTheDot() {
        let layout = RecordLinksLayout(nodeCount: 8, arcs: [], width: 340)
        #expect(layout.nodes.map(\.side) == [.top, .right, .right, .right, .bottom, .left, .left, .left])
    }

    @Test func twoArcsOnOnePairFanBySpacing() {
        let layout = RecordLinksLayout(nodeCount: 4, arcs: [arc(0, 1, slot: 0, slots: 2), arc(0, 1, slot: 1, slots: 2)], width: 340)
        #expect(layout.curves.count == 2)
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
        #expect(RecordLinksLayout.spacing(radius: 86, slots: 2) == 7)
    }

    @Test func manyArcsOnOnePairStayInsideTheRadius() {
        let arcs = (0..<30).map { arc(0, 2, slot: $0, slots: 30) }
        let layout = RecordLinksLayout(nodeCount: 4, arcs: arcs, width: 340)
        let first = layout.curves[0].control
        let last = layout.curves[29].control
        #expect(distance(first, last) <= layout.radius + 0.001)
    }

    @Test func arcsWithUnknownNodesAreSkipped() {
        let layout = RecordLinksLayout(nodeCount: 2, arcs: [arc(0, 5), arc(1, 1)], width: 340)
        #expect(layout.curves.isEmpty)
    }

    @Test func eachCurveKeepsItsArc() {
        let kept = RecordLinksArc(a: 1, b: 2, strong: true, slot: 0, slots: 1)
        let layout = RecordLinksLayout(nodeCount: 3, arcs: [arc(0, 5), kept], width: 340)
        #expect(layout.curves.count == 1)
        #expect(layout.curves[0].arc == kept)
    }

    @Test(arguments: [
        (340, nil, 172), (340, 300, 172), (340, 200, 172),
        (720, 500, 440), (720, nil, 552), (720, 100, 172),
    ] as [(CGFloat, CGFloat?, CGFloat)])
    func diameterShrinksToFitButNotBelowTheFloor(width: CGFloat, fitting: CGFloat?, diameter: CGFloat) {
        #expect(RecordLinksLayout.diameter(width: width, fitting: fitting) == diameter)
    }

    @Test func aNarrowColumnNeverDrawsBeyondItsWidth() {
        #expect(RecordLinksLayout.diameter(width: 300, fitting: 100) == 132)
        #expect(RecordLinksLayout.diameter(width: 100, fitting: nil) == 0)
    }
}
