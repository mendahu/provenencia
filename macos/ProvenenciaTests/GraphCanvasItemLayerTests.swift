import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@Suite
@MainActor
struct GraphCanvasItemLayerTests {
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
        size: CGSize = CGSize(width: 100, height: 40),
        model: String = "v1",
        raised: Bool = false,
        staysMounted: Bool = false,
        rotor: String? = nil,
        counter: BuildCounter? = nil
    ) -> GraphCanvasItem {
        GraphCanvasItem(
            id: id,
            origin: origin,
            size: size,
            staysMounted: staysMounted,
            isRaised: raised,
            rotor: rotor.map { GraphCanvasRotorEntry(rotor: $0, label: "\(id) label") },
            model: model,
            content: {
                counter?.counts[id, default: 0] += 1
                return AnyView(Color.red.frame(width: size.width, height: size.height))
            }
        )
    }

    /// A rect that contains every card these tests place near the origin.
    private func showAll(_ layer: GraphCanvasItemLayer) {
        layer.setVisibleDocumentRect(
            CGRect(x: -10_000, y: -10_000, width: 40_000, height: 40_000),
            magnification: 1
        )
    }

    /// Lets deferred slices run. One turn builds two cards.
    private func drain() async {
        for _ in 0..<8 {
            await withCheckedContinuation { (continuation: CheckedContinuation<Void, Never>) in
                DispatchQueue.main.async { continuation.resume() }
            }
        }
    }

    @Test func placesItemsAtTheirOriginOutsetForPaint() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        layer.update([item("a", at: CGPoint(x: 300, y: 200))])

        let view = try #require(layer.view(for: "a"))
        let outset = GraphCanvasItemLayer.paintOutset
        #expect(view.superview === container)
        #expect(view.frame.origin == CGPoint(x: 300 - outset, y: 200 - outset))
        #expect(view.frame.width >= 100 + 2 * outset)
    }

    @Test func rebuildsContentOnlyWhenTheModelChanges() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        let counter = BuildCounter()
        layer.update([item("a", at: .zero, counter: counter)])
        layer.update([item("a", at: CGPoint(x: 50, y: 0), counter: counter)])
        #expect(counter.counts["a"] == 1)

        layer.update([item("a", at: CGPoint(x: 50, y: 0), model: "v2", counter: counter)])
        #expect(counter.counts["a"] == 2)
    }

    @Test func removesItemsNoLongerPushed() async {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        layer.update([item("a", at: .zero), item("b", at: .zero)])
        await drain()
        let b = layer.view(for: "b")
        layer.update([item("a", at: .zero)])
        #expect(layer.view(for: "b") == nil)
        #expect(b?.superview == nil)
    }

    @Test func dragOffsetsMoveTheViewAndClearingReturnsIt() throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        layer.update([item("a", at: CGPoint(x: 100, y: 100))])
        let view = try #require(layer.view(for: "a"))
        let rest = view.frame.origin

        layer.applyDragOffsets(["a": CGSize(width: 30, height: -10)])
        #expect(view.frame.origin == CGPoint(x: rest.x + 30, y: rest.y - 10))

        layer.applyDragOffsets([:])
        #expect(view.frame.origin == rest)
    }

    @Test func raisedItemsDrawAboveTheRest() async throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        layer.update([item("a", at: .zero), item("b", at: .zero)])
        await drain()
        #expect(container.subviews.last === layer.view(for: "b"))

        layer.update([item("a", at: .zero, raised: true), item("b", at: .zero)])
        #expect(container.subviews.last === layer.view(for: "a"))
    }

    @Test func rotorsListItemsInHostOrderAndSelectOnMove() async throws {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        layer.update([
            item("link", at: .zero, rotor: "Links"),
            item("s1", at: .zero, rotor: "Subjects"),
            item("s2", at: .zero, rotor: "Subjects"),
        ])
        await drain()
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

    @Test func largeContentSwapKeepsTheOldViewsUntilTheNewOnesAreReady() async {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        showAll(layer)
        let counter = BuildCounter()
        let ids = (0..<6).map { "c\($0)" }
        layer.update(ids.map { item($0, at: .zero, model: "old", counter: counter) })
        await drain()
        let oldViews = ids.map { layer.view(for: $0) }
        #expect(counter.counts.values.reduce(0, +) == 6)

        layer.update(
            ids.map { item($0, at: CGPoint(x: 10, y: 0), model: "new", counter: counter) }
                + [item("outside", at: CGPoint(x: 100_000, y: 0), model: "new", counter: counter)]
        )
        #expect(counter.counts.values.reduce(0, +) == 6)
        #expect(layer.view(for: "c0") === oldViews[0])
        #expect(layer.view(for: "outside") == nil)
        let outset = GraphCanvasItemLayer.paintOutset
        #expect(layer.view(for: "c0")?.frame.origin.x == 10 - outset)

        await drain()
        #expect(counter.counts.values.reduce(0, +) == 12)
        #expect(layer.view(for: "c0") !== oldViews[0])
        #expect(oldViews[0]?.superview == nil)
        #expect(layer.view(for: "c0")?.superview === container)
        #expect(layer.view(for: "outside") == nil)
    }

    @Test func visibleRectMountsNearbyCardsAndAPinnedOne() async {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.setVisibleDocumentRect(CGRect(x: 0, y: 0, width: 200, height: 200), magnification: 10_000)
        layer.update([
            item("near", at: CGPoint(x: 10, y: 10)),
            item("far", at: CGPoint(x: 5_000, y: 10)),
            item("pinned", at: CGPoint(x: 5_000, y: 400), staysMounted: true),
        ])
        await drain()
        #expect(layer.view(for: "near") != nil)
        #expect(layer.view(for: "far") == nil)
        #expect(layer.view(for: "pinned") != nil)
    }

    @Test func movingTheVisibleRectSwapsWhichCardIsMounted() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.setVisibleDocumentRect(CGRect(x: 0, y: 0, width: 150, height: 150), magnification: 10_000)
        layer.update([
            item("a", at: CGPoint(x: 0, y: 0)),
            item("b", at: CGPoint(x: 2_000, y: 0)),
        ])
        #expect(layer.view(for: "a") != nil)
        #expect(layer.view(for: "b") == nil)

        layer.setVisibleDocumentRect(CGRect(x: 1_900, y: 0, width: 200, height: 150), magnification: 10_000)
        #expect(layer.view(for: "a") == nil)
        #expect(layer.view(for: "b") != nil)
    }

    @Test func expandingTheVisibleRectDoesNotBuildEveryCardImmediately() async {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.setVisibleDocumentRect(CGRect(x: 0, y: 0, width: 80, height: 80), magnification: 10_000)
        let ids = (0..<6).map { "c\($0)" }
        layer.update(ids.map { index in
            let n = Int(index.dropFirst()) ?? 0
            return item(index, at: CGPoint(x: CGFloat(n) * 400, y: 0), size: CGSize(width: 50, height: 40))
        })
        #expect(layer.view(for: "c0") != nil)
        #expect(layer.view(for: "c1") == nil)

        layer.setVisibleDocumentRect(CGRect(x: 0, y: 0, width: 3_000, height: 200), magnification: 10_000)
        #expect(layer.view(for: "c5") == nil)
        await drain()
        #expect(layer.view(for: "c5") != nil)
        #expect(layer.view(for: "c0") != nil)
    }

    @Test func revealMountsACardOutsideTheVisibleRect() {
        let container = flippedContainer()
        let layer = GraphCanvasItemLayer(container: container)
        layer.setVisibleDocumentRect(CGRect(x: 0, y: 0, width: 120, height: 120), magnification: 10_000)
        layer.update([
            item("near", at: CGPoint(x: 0, y: 0)),
            item("far", at: CGPoint(x: 4_000, y: 0)),
        ])
        #expect(layer.view(for: "far") == nil)
        layer.reveal("far")
        #expect(layer.view(for: "far") != nil)
    }
}
