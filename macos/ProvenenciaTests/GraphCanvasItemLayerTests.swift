import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@Suite
@MainActor
struct GraphCanvasItemLayerTests {
    private final class SizedView: NSView {
        override var intrinsicContentSize: NSSize {
            NSSize(
                width: 100 + 2 * GraphCanvasItemLayer.paintOutset,
                height: 40 + 2 * GraphCanvasItemLayer.paintOutset
            )
        }
    }

    private final class BuildCounter {
        var counts: [String: Int] = [:]
    }

    /// The layer holds its container weakly (the document view owns the layer),
    /// so tests keep the container alive themselves.
    private func flippedContainer() -> NSView {
        final class Flipped: NSView {
            override var isFlipped: Bool { true }
        }
        return Flipped(frame: CGRect(x: 0, y: 0, width: 2_000, height: 2_000))
    }

    private func item(
        _ id: String,
        at origin: CGPoint,
        model: String = "v1",
        raised: Bool = false,
        rotor: String? = nil,
        counter: BuildCounter? = nil
    ) -> GraphCanvasItem {
        GraphCanvasItem(
            id: id,
            origin: origin,
            isRaised: raised,
            rotor: rotor.map { GraphCanvasRotorEntry(rotor: $0, label: "\(id) label") },
            model: model,
            makeView: {
                counter?.counts[id, default: 0] += 1
                return SizedView()
            }
        )
    }

    @Test func placesItemsAtTheirOriginOutsetForPaint() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: CGPoint(x: 300, y: 200))])

        let view = try #require(layer.view(for: "a"))
        let outset = GraphCanvasItemLayer.paintOutset
        #expect(view.superview === container)
        #expect(view.frame.origin == CGPoint(x: 300 - outset, y: 200 - outset))
        #expect(view.frame.width >= 100 + 2 * outset)
    }

    @Test func eachItemIsItsOwnLayerAtTheLayersScale() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: .zero)])

        let first = try #require(layer.view(for: "a"))
        #expect(first.wantsLayer)
        #expect(first.layerContentsRedrawPolicy == .onSetNeedsDisplay)
        #expect(first.layer?.contentsScale == 1)

        layer.contentsScale = 2
        #expect(first.layer?.contentsScale == 2)

        layer.update([item("a", at: .zero), item("b", at: CGPoint(x: 200, y: 0))])
        let second = try #require(layer.view(for: "b"))
        #expect(second.layer?.contentsScale == 2)
    }

    @Test func rebuildsContentOnlyWhenTheModelChanges() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        let counter = BuildCounter()
        layer.update([item("a", at: .zero, counter: counter)])
        layer.update([item("a", at: CGPoint(x: 50, y: 0), counter: counter)])
        #expect(counter.counts["a"] == 1)

        layer.update([item("a", at: CGPoint(x: 50, y: 0), model: "v2", counter: counter)])
        #expect(counter.counts["a"] == 2)
    }

    @Test func removesItemsNoLongerPushed() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: .zero), item("b", at: .zero)])
        let b = layer.view(for: "b")
        layer.update([item("a", at: .zero)])
        #expect(layer.view(for: "b") == nil)
        #expect(b?.superview == nil)
    }

    @Test func dragOffsetsMoveTheViewAndClearingReturnsIt() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: CGPoint(x: 100, y: 100))])
        let view = try #require(layer.view(for: "a"))
        let rest = view.frame.origin

        layer.applyDragOffsets(["a": CGSize(width: 30, height: -10)])
        #expect(view.frame.origin == CGPoint(x: rest.x + 30, y: rest.y - 10))

        layer.applyDragOffsets([:])
        #expect(view.frame.origin == rest)
    }

    @Test func raisedItemsDrawAboveTheRest() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: .zero), item("b", at: .zero)])
        #expect(container.subviews.last === layer.view(for: "b"))

        layer.update([item("a", at: .zero, raised: true), item("b", at: .zero)])
        #expect(container.subviews.last === layer.view(for: "a"))
    }

    @Test func rotorFilterIgnoresCaseAndAccents() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([
            item("José", at: .zero, rotor: "Subjects"),
            item("Smith", at: .zero, rotor: "Subjects"),
        ])
        let search = GraphCanvasRotorSearch(name: "Subjects", layer: layer)

        let jose = NSAccessibilityCustomRotor.SearchParameters()
        jose.searchDirection = .next
        jose.filterString = "jose"
        let hit = try #require(search.result(for: jose))
        #expect(hit.targetElement as? NSView === layer.view(for: "José"))

        let none = NSAccessibilityCustomRotor.SearchParameters()
        none.searchDirection = .next
        none.filterString = "x"
        #expect(search.result(for: none) == nil)
    }

    @Test func rotorsListItemsInHostOrderAndSelectOnMove() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([
            item("link", at: .zero, rotor: "Links"),
            item("s1", at: .zero, rotor: "Subjects"),
            item("s2", at: .zero, rotor: "Subjects"),
        ])
        #expect(layer.rotorNames == ["Links", "Subjects"])

        var selected: [String] = []
        layer.onRotorSelect = { selected.append($0) }
        let search = GraphCanvasRotorSearch(name: "Subjects", layer: layer)

        let first = NSAccessibilityCustomRotor.SearchParameters()
        first.searchDirection = .next
        let s1 = try #require(search.result(for: first))
        #expect(s1.targetElement as? NSView === layer.view(for: "s1"))
        #expect(s1.customLabel == "s1 label")

        let next = NSAccessibilityCustomRotor.SearchParameters()
        next.searchDirection = .next
        next.currentItem = s1
        let s2 = try #require(search.result(for: next))
        #expect(s2.targetElement as? NSView === layer.view(for: "s2"))

        next.currentItem = s2
        #expect(search.result(for: next) == nil)
        #expect(selected == ["s1", "s2"])
    }
}
