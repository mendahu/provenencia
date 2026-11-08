import CoreGraphics
import Testing
@testable import Provenencia

struct CircularLinkDataTests {
    private func node(_ id: String) -> CircularLinkNode<String> { CircularLinkNode(id: id, label: id) }

    private func arc(_ a: String, _ b: String, width: CGFloat = 1.25) -> CircularLinkArc<String> {
        CircularLinkArc(from: a, to: b, lineWidth: width)
    }

    @Test func duplicateNodeIDsKeepTheFirst() {
        let first = CircularLinkNode(id: "A", label: "first")
        let second = CircularLinkNode(id: "A", label: "second")
        let dataset = CircularLinkDataset(nodes: [first, node("B"), second], arcs: [])
        #expect(dataset.nodes.map(\.label) == ["first", "B"])
        #expect(dataset.index(of: "A") == 0)
    }

    @Test func neighborsSkipArcsToNoNode() {
        let dataset = CircularLinkDataset(nodes: [node("A"), node("B")], arcs: [arc("A", "Z"), arc("B", "B")])
        #expect(dataset.resolvedArcs.isEmpty)
        #expect(dataset.neighbors(ofIndex: 0).isEmpty)
        #expect(dataset.neighborCount(of: "A") == 0)
        #expect(dataset.neighborCount(of: "Z") == 0)
    }

    @Test func neighborsAreTheOtherEndOfEachArc() {
        let arcs = [arc("A", "B"), arc("A", "B"), arc("B", "C"), arc("A", "D"), arc("E", "E")]
        let dataset = CircularLinkDataset(nodes: ["A", "B", "C", "D", "E", "F"].map(node), arcs: arcs)
        #expect(dataset.neighbors(ofIndex: 0) == [1, 3])
        #expect(dataset.neighborCount(of: "B") == 2)
        #expect(dataset.neighborCount(of: "E") == 0)
        #expect(dataset.neighborCount(of: "F") == 0)
    }

    @Test func slotsAreAssignedPerUnorderedPairInInputOrder() {
        let arcs = [arc("A", "B"), arc("C", "D"), arc("B", "A"), arc("A", "B")]
        let dataset = CircularLinkDataset(nodes: ["A", "B", "C", "D"].map(node), arcs: arcs)
        #expect(dataset.resolvedArcs.map(\.arcIndex) == [0, 1, 2, 3])
        #expect(dataset.resolvedArcs.map(\.slot) == [0, 0, 1, 2])
        #expect(dataset.resolvedArcs.map(\.slots) == [3, 1, 3, 3])
        // The first arc sets the pair's reference direction; B to A runs against it.
        #expect(dataset.resolvedArcs.map(\.flipped) == [false, false, true, false])
    }
}
