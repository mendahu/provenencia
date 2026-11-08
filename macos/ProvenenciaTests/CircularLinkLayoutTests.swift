import CoreGraphics
import Testing
@testable import Provenencia

struct CircularLinkLayoutTests {
    private func ids(_ count: Int) -> [String] { (0..<count).map { "n\($0)" } }

    private func arc(_ a: String, _ b: String, width: CGFloat = 1.25) -> CircularLinkArc<String> {
        CircularLinkArc(from: a, to: b, lineWidth: width)
    }

    private func distance(_ p: CGPoint, _ q: CGPoint) -> CGFloat {
        ((p.x - q.x) * (p.x - q.x) + (p.y - q.y) * (p.y - q.y)).squareRoot()
    }

    @Test func fourNodesSitTopRightBottomLeft() {
        let layout = CircularLinkLayout(nodes: ids(4), arcs: [], width: 340)
        #expect(layout.diameter == 172)
        #expect(layout.size == CGSize(width: 340, height: 232))
        #expect(layout.center == CGPoint(x: 170, y: 116))
        let expected = [CGPoint(x: 170, y: 30), CGPoint(x: 256, y: 116), CGPoint(x: 170, y: 202), CGPoint(x: 84, y: 116)]
        for (node, want) in zip(layout.nodes, expected) {
            #expect(distance(node.point, want) < 0.001)
        }
        #expect(layout.nodes.map(\.side) == [.top, .right, .bottom, .left])
    }

    @Test func zeroOneAndTwoNodes() {
        #expect(CircularLinkLayout<String>(nodes: [], arcs: [], width: 340).nodes.isEmpty)
        let one = CircularLinkLayout(nodes: ids(1), arcs: [], width: 340)
        #expect(one.nodes.map(\.point) == [CGPoint(x: 170, y: 24)])
        #expect(one.size == CGSize(width: 340, height: 64))
        let two = CircularLinkLayout(nodes: ids(2), arcs: [], width: 340)
        #expect(distance(two.nodes[0].point, CGPoint(x: 170, y: 30)) < 0.001)
        #expect(distance(two.nodes[1].point, CGPoint(x: 170, y: 202)) < 0.001)
        #expect(two.nodes.map(\.side) == [.top, .bottom])
    }

    @Test func eightNodesLabelDiagonalsBesideTheDot() {
        let layout = CircularLinkLayout(nodes: ids(8), arcs: [], width: 340)
        #expect(layout.nodes.map(\.side) == [.top, .right, .right, .right, .bottom, .left, .left, .left])
    }

    @Test func twoArcsOnOnePairFanBySpacing() {
        let layout = CircularLinkLayout(nodes: ids(4), arcs: [arc("n0", "n1"), arc("n0", "n1")], width: 340)
        #expect(layout.curves.count == 2)
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
        #expect(CircularLinkLayout<String>.spacing(radius: 86, slots: 2) == 7)
    }

    @Test func manyArcsOnOnePairStayInsideTheRadius() {
        let arcs = (0..<30).map { _ in arc("n0", "n2") }
        let layout = CircularLinkLayout(nodes: ids(4), arcs: arcs, width: 340)
        #expect(distance(layout.curves[0].control, layout.curves[29].control) <= layout.radius + 0.001)
    }

    @Test func arcsOnOnePairFanInInputOrder() {
        let arcs = [arc("A", "B"), arc("B", "A"), arc("A", "B")]
        let layout = CircularLinkLayout(nodes: ["A", "B", "C", "D"], arcs: arcs, width: 340)
        #expect(layout.curves.map(\.arcIndex) == [0, 1, 2])
        // Three slots of three, spaced by 7pt along the chord's normal: the middle arc is the unshifted one.
        let single = CircularLinkLayout(nodes: ["A", "B", "C", "D"], arcs: [arc("A", "B")], width: 340)
        #expect(distance(layout.curves[1].control, single.curves[0].control) < 0.001)
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
        #expect(abs(distance(layout.curves[1].control, layout.curves[2].control) - 7) < 0.001)
        // Pairs are unordered: B–A sits on A–B's side of the fan, between the others.
        #expect(distance(layout.curves[0].control, layout.curves[2].control) > 13.9)
        // Another pair does not share the fan.
        let apart = CircularLinkLayout(nodes: ["A", "B", "C", "D"], arcs: [arc("A", "B"), arc("C", "D")], width: 340)
        #expect(distance(apart.curves[0].control, single.curves[0].control) < 0.001)
    }

    @Test func curvesDrawThinnestFirst() {
        let arcs = [arc("A", "B", width: 3), arc("A", "C"), arc("B", "C", width: 3), arc("C", "D")]
        let layout = CircularLinkLayout(nodes: ["A", "B", "C", "D"], arcs: arcs, width: 340)
        #expect(layout.curves.map(\.arcIndex) == [1, 3, 0, 2])
    }

    @Test func arcsNamingNoNodeOrOneNodeAreSkipped() {
        let layout = CircularLinkLayout(nodes: ["A", "B"], arcs: [arc("A", "Z"), arc("B", "B")], width: 340)
        #expect(layout.curves.isEmpty)
    }

    @Test func eachCurveKeepsItsArcIndex() {
        let arcs = [arc("A", "Z"), CircularLinkArc(from: "B", to: "C", lineWidth: 3, color: .red)]
        let layout = CircularLinkLayout(nodes: ["A", "B", "C"], arcs: arcs, width: 340)
        #expect(layout.curves.count == 1)
        #expect(layout.curves[0].arcIndex == 1)
        #expect(layout.curves[0].lineWidth == 3)
        #expect(layout.curves[0].color == .red)
    }

    @Test func aStyleMovesTheGeometry() {
        var style = CircularLinkDiagramStyle.standard
        style.sideRoom = 50
        #expect(CircularLinkLayout<String>.diameter(width: 340, fitting: nil, style: style) == 240)
        #expect(CircularLinkLayout(nodes: ids(4), arcs: [], width: 340, style: style).diameter == 240)
    }

    @Test(arguments: [
        (340, nil, 172), (340, 300, 172), (340, 200, 172),
        (720, 500, 440), (720, nil, 552), (720, 100, 172),
    ] as [(CGFloat, CGFloat?, CGFloat)])
    func diameterShrinksToFitButNotBelowTheFloor(width: CGFloat, fitting: CGFloat?, diameter: CGFloat) {
        #expect(CircularLinkLayout<String>.diameter(width: width, fitting: fitting, style: .standard) == diameter)
    }

    @Test func aNarrowColumnNeverDrawsBeyondItsWidth() {
        #expect(CircularLinkLayout<String>.diameter(width: 300, fitting: 100, style: .standard) == 132)
        #expect(CircularLinkLayout<String>.diameter(width: 100, fitting: nil, style: .standard) == 0)
    }

    @Test func neighborSetsMatchNeighborsOfEachNode() {
        let arcs = [arc("A", "B"), arc("A", "B"), arc("B", "C"), arc("A", "D"), arc("E", "E")]
        let sets = CircularLinkArc.neighborSets(in: arcs)
        for id in ["A", "B", "C", "D", "E", "F"] {
            #expect(sets[id, default: []] == CircularLinkArc.neighbors(of: id, in: arcs))
        }
        #expect(sets["A"] == ["B", "D"])
        #expect(sets["F"] == nil)
    }
}
