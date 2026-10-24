@preconcurrency import AppKit

/// One host-drawn item the canvas positions in AppKit — a graph card.
///
/// Each item is a ``GraphCanvasItemView`` the host builds with ``makeView``, sized by its
/// `contentSize` plus ``GraphCanvasItemView/paintOutset`` on every side,
/// and a subview of ``GraphCanvasDocumentView`` above the host's base SwiftUI
/// layer. AppKit owns where items are: a drag moves one view's frame. A view is
/// rebuilt only when ``model`` changes.
struct GraphCanvasItem {
    let id: String
    /// Document-space top-leading corner of the item's content.
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

/// An item's entry in a document-level VoiceOver rotor.
struct GraphCanvasRotorEntry: Equatable {
    /// The rotor's name ("Subjects"); items sharing it form one rotor.
    var rotor: String
    /// What VoiceOver reads for this item in the rotor.
    var label: String
}

/// Hosts ``GraphCanvasItem``s as AppKit subviews of the canvas document.
@MainActor
final class GraphCanvasItemLayer {
    /// Called when VoiceOver moves to an item through a rotor.
    var onRotorSelect: ((String) -> Void)?

    private weak var container: NSView?
    private var views: [String: GraphCanvasItemView] = [:]
    private var models: [String: any Equatable] = [:]
    private var origins: [String: CGPoint] = [:]
    private var dragOffsets: [String: CGSize] = [:]
    /// Item ids in the host's order (bottom to top before raising).
    private(set) var order: [String] = []
    /// Each item's place among the views, bottom to top; see ``stackIndex(of:)``.
    private var stackIndices: [String: Int] = [:]
    private var rotorEntries: [String: GraphCanvasRotorEntry] = [:]

    /// The scale every item's layer renders at. The scroll view sets it to the
    /// window's backing scale times the settled zoom past 1x. A first paint can
    /// be cached at 1x before the window's scale is known, so this puts every
    /// layer, even one AppKit reset, back; the new scale changes each view's paint
    /// key, so it repaints at it.
    var contentsScale: CGFloat = 1 {
        didSet {
            for view in views.values {
                guard let layer = view.layer,
                      abs(layer.contentsScale - contentsScale) > 0.000_1
                else { continue }
                layer.contentsScale = contentsScale
                view.needsDisplay = true
            }
        }
    }

    init(container: NSView) {
        self.container = container
    }

    /// Replaces the item set: adds, removes, repositions, and rebuilds the
    /// view of items whose model changed.
    func update(_ items: [GraphCanvasItem]) {
        guard let container else { return }
        let ids = Set(items.map(\.id))
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
                    view = insert(item)
                } else {
                    view = existing
                }
            } else {
                view = insert(item)
            }
            views[item.id] = view
            models[item.id] = item.model
            origins[item.id] = item.origin
            rotorEntries[item.id] = item.rotor
            place(item.id)
        }
        order = items.map(\.id)

        restack(items.filter { !$0.isRaised }.map(\.id) + items.filter(\.isRaised).map(\.id))
    }

    /// Where `id` is drawn among the items, 0 at the bottom; nil for an id that
    /// is not an item. Follows the host's order with raised items last, so the
    /// pointer can pick the card drawn on top.
    func stackIndex(of id: String) -> Int? {
        stackIndices[id]
    }

    /// Puts the views in `stack` order (bottom to top), adding only the views
    /// from the first one that is out of place, so an unchanged order moves
    /// nothing.
    private func restack(_ stack: [String]) {
        guard let container else { return }
        stackIndices = Dictionary(uniqueKeysWithValues: stack.enumerated().map { ($1, $0) })
        let desired = stack.compactMap { views[$0] }
        let known = Set(desired.map(ObjectIdentifier.init))
        let current = container.subviews.filter { known.contains(ObjectIdentifier($0)) }
        guard let first = desired.indices.first(where: { current[$0] !== desired[$0] }) else { return }
        for view in desired[first...] {
            container.addSubview(view, positioned: .above, relativeTo: nil)
        }
    }

    /// Moves items by live drag offsets (document space). Items missing from
    /// `offsets` return to their pushed origin.
    func applyDragOffsets(_ offsets: [String: CGSize]) {
        let changed = Set(dragOffsets.keys).union(offsets.keys)
        dragOffsets = offsets
        for id in changed {
            place(id)
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

    /// Builds the item's view, which is its own layer (``GraphCanvasItemView``)
    /// painted when built, when its pixel size or appearance changes, or on
    /// `invalidatePaint()`, so a pan or a zoom step composites cached layers.
    /// It goes above the other items at the current ``contentsScale``.
    private func insert(_ item: GraphCanvasItem) -> GraphCanvasItemView {
        let view = item.makeView()
        container?.addSubview(view, positioned: .above, relativeTo: nil)
        view.layer?.contentsScale = contentsScale
        return view
    }

    private func place(_ id: String) {
        guard let view = views[id], let origin = origins[id] else { return }
        let offset = dragOffsets[id] ?? .zero
        view.setFrameOrigin(CGPoint(
            x: origin.x + offset.width - GraphCanvasItemView.paintOutset,
            y: origin.y + offset.height - GraphCanvasItemView.paintOutset
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
