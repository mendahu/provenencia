import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@Suite
@MainActor
struct GraphCanvasItemLayerTests {
    private final class SizedView: GraphCanvasItemView {}

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
                return SizedView(contentSize: CGSize(width: 100, height: 40))
            }
        )
    }

    @Test func placesItemsAtTheirOriginOutsetForPaint() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.update([item("a", at: CGPoint(x: 300, y: 200))])

        let view = try #require(layer.view(for: "a"))
        let outset = GraphCanvasItemView.paintOutset
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

    /// The item ids in the container's subview order, bottom to top.
    private func stackOrder(_ container: NSView, _ layer: GraphCanvasItemLayer, _ ids: [String]) -> [String] {
        container.subviews.compactMap { subview in ids.first { layer.view(for: $0) === subview } }
    }

    @Test func itemsKeepHostOrderAcrossARebuildAndARaise() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        let ids = ["a", "b", "c"]
        let origins = [CGPoint(x: 0, y: 0), CGPoint(x: 10, y: 0), CGPoint(x: 20, y: 0)]
        func push(model b: String = "v1", raised: String? = nil) {
            layer.update(zip(ids, origins).map { id, origin in
                item(id, at: origin, model: id == "b" ? b : "v1", raised: id == raised)
            })
        }

        push()
        #expect(stackOrder(container, layer, ids) == ["a", "b", "c"])

        // A rebuilt view goes back where it was, not on top.
        let before = layer.view(for: "b")
        push(model: "v2")
        #expect(layer.view(for: "b") !== before)
        #expect(stackOrder(container, layer, ids) == ["a", "b", "c"])

        push(model: "v2", raised: "c")
        #expect(stackOrder(container, layer, ids) == ["a", "b", "c"])
        push(model: "v2", raised: "a")
        #expect(stackOrder(container, layer, ids) == ["b", "c", "a"])
        #expect(ids.map { layer.stackIndex(of: $0) } == [2, 0, 1])

        // A lowered item comes back down.
        push(model: "v2")
        #expect(stackOrder(container, layer, ids) == ["a", "b", "c"])
        #expect(ids.map { layer.stackIndex(of: $0) } == [0, 1, 2])
        #expect(layer.stackIndex(of: "missing") == nil)
    }

    @Test func restackingPaintsNothing() throws {
        final class Counting: GraphCanvasItemView {
            var paints = 0
            override func paint() { paints += 1 }
        }
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        func push(raised: String?) {
            layer.update(["a", "b", "c"].map { id in
                GraphCanvasItem(id: id, origin: .zero, isRaised: id == raised, model: "v1", makeView: {
                    Counting(contentSize: CGSize(width: 10, height: 10))
                })
            })
        }
        push(raised: nil)
        for id in ["a", "b", "c"] {
            let view = try #require(layer.view(for: id) as? Counting)
            view.layer?.contentsScale = 1
            view.updateLayer()
        }
        func paints() -> [Int] {
            ["a", "b", "c"].map { (layer.view(for: $0) as? Counting)?.paints ?? -1 }
        }
        #expect(paints() == [1, 1, 1])

        push(raised: "a")
        push(raised: nil)
        for id in ["a", "b", "c"] { layer.view(for: id)?.updateLayer() }
        #expect(paints() == [1, 1, 1])
    }

    @Test func hitTestingPicksTheTopmostPaintedItem() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        let frame = CGRect(x: 0, y: 0, width: 50, height: 50)
        let targets = ["a", "b", "c"].map { GraphCanvasHitTarget(id: $0, frame: frame, acceptsConnect: true) }
        func hit() -> String? {
            GraphCanvasPointerHitTesting.topmostTarget(
                at: CGPoint(x: 10, y: 10), in: targets, stackIndex: { layer.stackIndex(of: $0) }
            )?.id
        }
        func push(raised: String?, bModel: String = "v1") {
            layer.update(["a", "b", "c"].map { id in
                item(id, at: .zero, model: id == "b" ? bModel : "v1", raised: id == raised)
            })
        }
        push(raised: nil)
        #expect(hit() == "c")
        push(raised: nil, bModel: "v2")
        #expect(hit() == "c")
        push(raised: "a")
        #expect(hit() == "a")
        push(raised: nil)
        #expect(hit() == "c")
    }

    @Test func gridOriginsAreShownRelativeToTheDocumentsContentOrigin() throws {
        let pointer = GraphCanvasPointerController()
        let document = GraphCanvasDocumentView(pointer: pointer, frame: CGRect(x: 0, y: 0, width: 600, height: 600))
        let outset = GraphCanvasItemView.paintOutset
        document.contentOrigin = CGPoint(x: 100, y: 50)
        document.items.update([item("a", at: CGPoint(x: 130, y: 70))])
        let view = try #require(document.items.view(for: "a"))
        #expect(view.frame.origin == CGPoint(x: 30 - outset, y: 20 - outset))

        #expect(document.gridPoint(forDocumentPoint: CGPoint(x: 35, y: 25)) == CGPoint(x: 135, y: 75))
        pointer.hitTargets = [
            GraphCanvasHitTarget(id: "a", frame: CGRect(x: 130, y: 70, width: 40, height: 30), acceptsConnect: true),
        ]
        var selected: String?
        pointer.onSelect = { selected = $0 }
        let grid = document.gridPoint(forDocumentPoint: CGPoint(x: 35, y: 25))
        pointer.mouseDown(gridPoint: grid, windowPoint: .zero)
        pointer.mouseUp(gridPoint: grid, windowPoint: .zero)
        #expect(selected == "a")

        // Moving the origin re-places the item without a new push.
        document.contentOrigin = CGPoint(x: 0, y: 0)
        #expect(view.frame.origin == CGPoint(x: 130 - outset, y: 70 - outset))
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
