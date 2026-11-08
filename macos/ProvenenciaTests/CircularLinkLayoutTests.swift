import CoreGraphics
import Testing
@testable import Provenencia

struct CircularLinkLayoutTests {
    private func ids(_ count: Int) -> [String] { (0..<count).map { "n\($0)" } }

    private func arc(_ a: String, _ b: String, width: CGFloat = 1.25) -> CircularLinkArc<String> {
        CircularLinkArc(from: a, to: b, lineWidth: width)
    }

    private func make(
        _ nodes: [String], _ arcs: [CircularLinkArc<String>] = [], width: CGFloat = 340,
        style: CircularLinkDiagramStyle = .standard
    ) -> CircularLinkLayout<String> {
        CircularLinkLayout(
            dataset: CircularLinkDataset(nodes: nodes.map { CircularLinkNode(id: $0, label: $0) }, arcs: arcs),
            width: width, style: style
        )
    }

    private func distance(_ p: CGPoint, _ q: CGPoint) -> CGFloat {
        ((p.x - q.x) * (p.x - q.x) + (p.y - q.y) * (p.y - q.y)).squareRoot()
    }

    @Test func fourNodesSitTopRightBottomLeft() {
        let layout = make(ids(4), [])
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
        #expect(make([], []).nodes.isEmpty)
        let one = make(ids(1), [])
        #expect(one.nodes.map(\.point) == [CGPoint(x: 170, y: 24)])
        #expect(one.size == CGSize(width: 340, height: 64))
        let two = make(ids(2), [])
        #expect(distance(two.nodes[0].point, CGPoint(x: 170, y: 30)) < 0.001)
        #expect(distance(two.nodes[1].point, CGPoint(x: 170, y: 202)) < 0.001)
        #expect(two.nodes.map(\.side) == [.top, .bottom])
    }

    @Test func eightNodesLabelDiagonalsBesideTheDot() {
        let layout = make(ids(8), [])
        #expect(layout.nodes.map(\.side) == [.top, .right, .right, .right, .bottom, .left, .left, .left])
    }

    @Test func twoArcsOnOnePairFanBySpacing() {
        let layout = make(ids(4), [arc("n0", "n1"), arc("n0", "n1")])
        #expect(layout.curves.count == 2)
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
        #expect(CircularLinkLayout<String>.spacing(radius: 86, slots: 2) == 7)
    }

    @Test func manyArcsOnOnePairStayInsideTheRadius() {
        let arcs = (0..<30).map { _ in arc("n0", "n2") }
        let layout = make(ids(4), arcs)
        #expect(distance(layout.curves[0].control, layout.curves[29].control) <= layout.radius + 0.001)
    }

    @Test func arcsOnOnePairFanInInputOrder() {
        let arcs = [arc("A", "B"), arc("B", "A"), arc("A", "B")]
        let layout = make(["A", "B", "C", "D"], arcs)
        #expect(layout.curves.map(\.arcIndex) == [0, 1, 2])
        // Three slots of three, spaced by 7pt along the chord's normal: the middle arc is the unshifted one.
        let single = make(["A", "B", "C", "D"], [arc("A", "B")])
        #expect(distance(layout.curves[1].control, single.curves[0].control) < 0.001)
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
        #expect(abs(distance(layout.curves[1].control, layout.curves[2].control) - 7) < 0.001)
        // Slots follow the pair's reference direction (A to B), so the reversed B–A arc keeps its own place between the others.
        #expect(distance(layout.curves[0].control, layout.curves[2].control) > 13.9)
        // Another pair does not share the fan.
        let apart = make(["A", "B", "C", "D"], [arc("A", "B"), arc("C", "D")])
        #expect(distance(apart.curves[0].control, single.curves[0].control) < 0.001)
    }

    @Test func reversedArcsOnOnePairFanApart() {
        let layout = make(["A", "B", "C", "D"], [arc("A", "B"), arc("B", "A")])
        #expect(abs(distance(layout.curves[0].control, layout.curves[1].control) - 7) < 0.001)
    }

    @Test func curvesDrawThinnestFirst() {
        let arcs = [arc("A", "B", width: 3), arc("A", "C"), arc("B", "C", width: 3), arc("C", "D")]
        let layout = make(["A", "B", "C", "D"], arcs)
        #expect(layout.curves.map(\.arcIndex) == [1, 3, 0, 2])
    }

    @Test func eachCurveKeepsItsArcIndex() {
        let arcs = [arc("A", "Z"), CircularLinkArc(from: "B", to: "C", lineWidth: 3, color: .red)]
        let layout = make(["A", "B", "C"], arcs)
        #expect(layout.curves.count == 1)
        #expect(layout.curves[0].arcIndex == 1)
        #expect(layout.curves[0].lineWidth == 3)
        #expect(layout.curves[0].color == .red)
    }

    @Test func aStyleMovesTheGeometry() {
        var style = CircularLinkDiagramStyle.standard
        style.sideRoom = 50
        #expect(CircularLinkLayout<String>.diameter(width: 340, fitting: nil, style: style) == 240)
        #expect(make(ids(4), [], style: style).diameter == 240)
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

    @Test func johannGeometryIsUnchanged() {
        // Johann's two-line pairs at width 340, from the formula of the first circular link diagram: the
        // control point pulled 0.3 toward the chord's midpoint, then shifted 7pt apart along the chord's normal.
        let arcs = [
            arc("n0", "n1", width: 3), arc("n0", "n1"), arc("n0", "n3", width: 3), arc("n0", "n3", width: 3),
            arc("n1", "n3"), arc("n0", "n4", width: 3), arc("n4", "n5"), arc("n4", "n5"),
        ]
        let layout = make(ids(7), arcs)
        let controls: [Int: CGPoint] = [
            0: CGPoint(x: 181.604, y: 91.904), 1: CGPoint(x: 178.567, y: 98.21),
            2: CGPoint(x: 179.009, y: 113.944), 3: CGPoint(x: 172.185, y: 115.501),
            6: CGPoint(x: 149.09, y: 132.675), 7: CGPoint(x: 154.563, y: 128.311),
        ]
        for (arcIndex, want) in controls {
            let curve = layout.curves.first { $0.arcIndex == arcIndex }
            #expect(curve.map { distance($0.control, want) < 0.002 } == true)
        }
    }
}
