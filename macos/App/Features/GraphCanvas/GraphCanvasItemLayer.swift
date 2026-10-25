@preconcurrency import AppKit

/// One item the host draws and the canvas positions in AppKit, such as a card.
///
/// Each item is a ``GraphCanvasItemView`` the host builds with ``makeView``, sized by its
/// `contentSize` plus ``GraphCanvasItemView/paintOutset`` on every side,
/// and a subview of ``GraphCanvasDocumentView`` above the host's base SwiftUI
/// layer. AppKit owns where items are: a drag moves one view's frame. A view is
/// rebuilt only when ``model`` changes.
struct GraphCanvasItem {
    let id: String
    /// Grid-space top-leading corner of the item's content.
    var origin: CGPoint
    /// Drawn above every other item (the item being dragged or activated).
    var isRaised: Bool = false
    /// The VoiceOver rotor that lists this item, if any.
    var rotor: GraphCanvasRotorEntry?
    /// Everything the view draws. The view is rebuilt only when this changes,
    /// so a host can push every item on any state change.
    var model: any Equatable
    /// Builds the item's view for ``model``: a ``GraphCanvasItemView`` that
    /// paints in `paint()` and calls `invalidatePaint()` when its content
    /// changes. The view is paint plus accessibility; pointer input stays with
    /// ``GraphCanvasPointerController``.
    /// It sizes itself through `init(contentSize:)`.
    var makeView: @MainActor () -> GraphCanvasItemView
}

/// A translucent stand-in the host shows while it waits for the researcher,
/// such as the item a click would create. See ``GraphCanvasItemLayer/preview``.
struct GraphCanvasPreview {
    /// Grid-space top-leading corner of the preview's content.
    var origin: CGPoint
    /// Everything the view draws. The view is rebuilt only when this changes.
    var model: any Equatable
    /// Builds the preview's view for ``model``.
    var makeView: @MainActor () -> GraphCanvasItemView
}

/// An item's entry in a document-level VoiceOver rotor.
struct GraphCanvasRotorEntry: Equatable {
    /// The rotor's name ("Cards", say); items sharing it form one rotor.
    var rotor: String
    /// What VoiceOver reads for this item in the rotor.
    var label: String
}

/// Hosts ``GraphCanvasItem``s as AppKit subviews of the canvas document.
@MainActor
final class GraphCanvasItemLayer {
    /// The grid point the document's top-left stands on. Items are pushed at
    /// grid origins and placed at `origin − contentOrigin`; the document view
    /// wires this to its own ``GraphCanvasDocumentView/contentOrigin``.
    var contentOrigin: () -> CGPoint = { .zero }

    /// Called when VoiceOver moves to an item through a rotor.
    var onRotorSelect: ((String) -> Void)?

    private weak var container: NSView?
    private var views: [String: GraphCanvasItemView] = [:]
    private var models: [String: any Equatable] = [:]
    private var origins: [String: CGPoint] = [:]
    private var dragOffsets: [String: CGSize] = [:]
    /// Item ids in the host's order (bottom to top before raising).
    private var order: [String] = []
    /// Each item's place among the views, bottom to top; see ``stackIndex(of:)``.
    private var stackIndices: [String: Int] = [:]
    /// Item ids bottom to top, as last restacked.
    private var stack: [String] = []
    private var rotorEntries: [String: GraphCanvasRotorEntry] = [:]

    /// The scale every item's layer renders at. The scroll view sets it to the
    /// window's backing scale times the settled zoom past 1x. A first paint can
    /// be cached at 1x before the window's scale is known, so this puts every
    /// layer, even one AppKit reset, back; the new scale changes each view's paint
    /// key, so it repaints at it.
    var contentsScale: CGFloat = 1 {
        didSet {
            for view in views.values + (previewView.map { [$0] } ?? []) {
                guard let layer = view.layer,
                      abs(layer.contentsScale - contentsScale) > GraphCanvasCoordinates.magnificationTolerance
                else { continue }
                layer.contentsScale = contentsScale
                view.needsDisplay = true
            }
        }
    }

    /// The stand-in the host shows while it waits for the researcher (the item
    /// a click would create). It goes through the same build, scale and
    /// placement as an item, but is not one: it has no id, rotor entry or stack
    /// index, `update` never removes it, a drag never moves it, and it sits
    /// below every item, so a card covers it and its frame never lies over one. Its view is rebuilt only when its model changes, so
    /// moving it only moves the frame. Nil removes it.
    var preview: GraphCanvasPreview? {
        didSet { applyPreview(replacing: oldValue) }
    }

    /// The preview's view, for tests.
    private(set) var previewView: GraphCanvasItemView?

    init(container: NSView) {
        self.container = container
    }

    /// Replaces the item set: adds, removes, repositions, and rebuilds the
    /// view of items whose model changed.
    func update(_ items: [GraphCanvasItem]) {
        guard container != nil else { return }
        let ids = Set(items.map(\.id))
        let stack = items.filter { !$0.isRaised }.map(\.id) + items.filter(\.isRaised).map(\.id)
        for (id, view) in views where !ids.contains(id) {
            view.removeFromSuperview()
            views[id] = nil
            models[id] = nil
            origins[id] = nil
            dragOffsets[id] = nil
            rotorEntries[id] = nil
        }

        for item in items {
            let view: GraphCanvasItemView
            if let existing = views[item.id] {
                if let previous = models[item.id], !item.model.isEqual(to: previous) {
                    existing.removeFromSuperview()
                    view = insert(item.makeView(), forID: item.id, in: stack)
                } else {
                    view = existing
                }
            } else {
                view = insert(item.makeView(), forID: item.id, in: stack)
            }
            views[item.id] = view
            models[item.id] = item.model
            origins[item.id] = item.origin
            rotorEntries[item.id] = item.rotor
            place(item.id)
        }
        order = items.map(\.id)

        restack(stack)
    }

    /// Where `id` is drawn among the items, 0 at the bottom; nil for an id that
    /// is not an item. Follows the host's order with raised items last, so the
    /// pointer can pick the item drawn on top.
    func stackIndex(of id: String) -> Int? {
        stackIndices[id]
    }

    /// Puts the views in `stack` order (bottom to top), moving only the views
    /// that are out of place: the longest run already in order stays, so a
    /// raised item goes to the top and a lowered one back into place, and an
    /// unchanged order moves nothing.
    private func restack(_ stack: [String]) {
        guard let container else { return }
        self.stack = stack
        stackIndices = Dictionary(uniqueKeysWithValues: stack.enumerated().map { ($1, $0) })
        let desired = stack.compactMap { views[$0] }
        let currentIndex = Dictionary(
            uniqueKeysWithValues: container.subviews.enumerated().map { (ObjectIdentifier($1), $0) }
        )
        let positions = desired.map { currentIndex[ObjectIdentifier($0)] ?? Int.max }
        let stays = Set(Self.longestIncreasingRun(positions))
        for (index, view) in desired.enumerated() where !stays.contains(index) {
            if index > 0 {
                container.addSubview(view, positioned: .above, relativeTo: desired[index - 1])
            } else {
                container.addSubview(view, positioned: .below, relativeTo: desired.count > 1 ? desired[1] : nil)
            }
        }
        settlePreview()
    }

    /// The indices of a longest strictly increasing subsequence of `values`.
    private static func longestIncreasingRun(_ values: [Int]) -> [Int] {
        var tails: [Int] = []
        var previous = [Int?](repeating: nil, count: values.count)
        for (index, value) in values.enumerated() {
            var low = 0
            var high = tails.count
            while low < high {
                let mid = (low + high) / 2
                if values[tails[mid]] < value { low = mid + 1 } else { high = mid }
            }
            if low > 0 { previous[index] = tails[low - 1] }
            if low == tails.count { tails.append(index) } else { tails[low] = index }
        }
        var run: [Int] = []
        var cursor = tails.last
        while let at = cursor {
            run.append(at)
            cursor = previous[at]
        }
        return run
    }

    /// The lowest item view, which the preview sits directly below.
    private var lowestItemView: GraphCanvasItemView? {
        stack.first.flatMap { views[$0] }
    }

    /// Keeps the preview directly below the lowest item view, so cards cover it.
    private func settlePreview() {
        guard let container, let previewView, let lowest = lowestItemView,
              let at = container.subviews.firstIndex(of: lowest),
              at == 0 || container.subviews[at - 1] !== previewView
        else { return }
        container.addSubview(previewView, positioned: .below, relativeTo: lowest)
    }

    /// Builds, moves or removes the preview view for the new ``preview``.
    private func applyPreview(replacing previous: GraphCanvasPreview?) {
        guard let item = preview else {
            previewView?.removeFromSuperview()
            previewView = nil
            return
        }
        if let previewView, let previous, item.model.isEqual(to: previous.model) {
            place(previewView, at: item.origin, offset: .zero)
            return
        }
        previewView?.removeFromSuperview()
        let view = item.makeView()
        if let lowest = lowestItemView {
            container?.addSubview(view, positioned: .below, relativeTo: lowest)
        } else {
            container?.addSubview(view, positioned: .above, relativeTo: nil)
        }
        view.layer?.contentsScale = contentsScale
        previewView = view
        place(view, at: item.origin, offset: .zero)
    }

    /// Moves items by live drag offsets (a delta, the same in grid and document space). Items missing from
    /// `offsets` return to their pushed origin.
    func applyDragOffsets(_ offsets: [String: CGSize]) {
        let changed = Set(dragOffsets.keys).union(offsets.keys)
        dragOffsets = offsets
        for id in changed {
            place(id)
        }
    }

    /// Places every item again, after the content origin moved.
    func placeAll() {
        for id in views.keys {
            place(id)
        }
        if let previewView, let preview {
            place(previewView, at: preview.origin, offset: .zero)
        }
    }

    /// The view for an item, for accessibility targets and tests.
    func view(for id: String) -> GraphCanvasItemView? {
        views[id]
    }

    /// Rotor names in the order their first item appears.
    var rotorNames: [String] {
        var names: [String] = []
        for id in order {
            if let rotor = rotorEntries[id]?.rotor, !names.contains(rotor) {
                names.append(rotor)
            }
        }
        return names
    }

    /// Items listed in `rotor`, in host order.
    func rotorItems(named rotor: String) -> [(id: String, label: String)] {
        order.compactMap { id in
            guard let entry = rotorEntries[id], entry.rotor == rotor else { return nil }
            return (id, entry.label)
        }
    }

    /// Adds the item's view (its own layer, ``GraphCanvasItemView``, painted
    /// when built, when its pixel size or appearance changes, or on
    /// `invalidatePaint()`, so a pan or a zoom step composites cached layers)
    /// at the current ``contentsScale``, straight into its place in `stack`:
    /// below the next item already placed, else above the last one before it.
    private func insert(_ view: GraphCanvasItemView, forID id: String, in stack: [String]) -> GraphCanvasItemView {
        if let container, let position = stack.firstIndex(of: id) {
            func placed(_ ids: some Sequence<String>) -> GraphCanvasItemView? {
                ids.lazy.compactMap { self.views[$0] }.first { $0.superview === container }
            }
            if let next = placed(stack[(position + 1)...]) {
                container.addSubview(view, positioned: .below, relativeTo: next)
            } else {
                container.addSubview(view, positioned: .above, relativeTo: placed(stack[..<position].reversed()))
            }
        }
        view.layer?.contentsScale = contentsScale
        return view
    }

    private func place(_ id: String) {
        guard let view = views[id], let origin = origins[id] else { return }
        place(view, at: origin, offset: dragOffsets[id] ?? .zero)
    }

    /// Puts `view` at the grid `origin`, shown relative to the content origin.
    private func place(_ view: GraphCanvasItemView, at origin: CGPoint, offset: CGSize) {
        let document = GraphCanvasCoordinates.documentPoint(origin, origin: contentOrigin())
        view.setFrameOrigin(CGPoint(
            x: document.x + offset.width - GraphCanvasItemView.paintOutset,
            y: document.y + offset.height - GraphCanvasItemView.paintOutset
        ))
    }
}

private extension Equatable {
    func isEqual(to other: any Equatable) -> Bool {
        (other as? Self) == self
    }
}

/// Searches one document-level VoiceOver rotor over ``GraphCanvasItemLayer`` items.
///
/// AppKit calls rotor searches on the main thread; the `@preconcurrency`
/// conformance checks that at runtime.
@MainActor
final class GraphCanvasRotorSearch: NSObject, @preconcurrency NSAccessibilityCustomRotorItemSearchDelegate {
    let name: String
    private weak var layer: GraphCanvasItemLayer?

    init(name: String, layer: GraphCanvasItemLayer) {
        self.name = name
        self.layer = layer
    }

    func rotor(
        _ rotor: NSAccessibilityCustomRotor,
        resultFor searchParameters: NSAccessibilityCustomRotor.SearchParameters
    ) -> NSAccessibilityCustomRotor.ItemResult? {
        result(for: searchParameters)
    }

    func result(for searchParameters: NSAccessibilityCustomRotor.SearchParameters) -> NSAccessibilityCustomRotor.ItemResult? {
        guard let layer else { return nil }
        let query = TextMatch.Query(searchParameters.filterString)
        let items = layer.rotorItems(named: name).filter { query.matches($0.label) }
        guard !items.isEmpty else { return nil }

        let current = searchParameters.currentItem?.targetElement as? NSView
        let currentIndex = items.firstIndex { layer.view(for: $0.id) === current }
        let index: Int
        switch (searchParameters.searchDirection, currentIndex) {
        case (.next, nil): index = 0
        case (.previous, nil): index = items.count - 1
        case (.next, let i?): index = i + 1
        case (.previous, let i?): index = i - 1
        @unknown default: return nil
        }
        guard items.indices.contains(index), let view = layer.view(for: items[index].id) else { return nil }

        let result = NSAccessibilityCustomRotor.ItemResult(targetElement: view)
        result.customLabel = items[index].label
        layer.onRotorSelect?(items[index].id)
        return result
    }
}
