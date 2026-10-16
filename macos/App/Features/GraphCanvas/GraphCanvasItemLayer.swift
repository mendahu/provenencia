@preconcurrency import AppKit

/// One host-drawn item the canvas positions in AppKit — a graph card.
///
/// Each item is an `NSView` the host builds with ``makeView``, sized by its
/// `intrinsicContentSize` (which includes ``GraphCanvasItemLayer/paintOutset``),
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
    /// Builds the item's view for ``model``. The view is paint plus
    /// accessibility; pointer input stays with ``GraphCanvasPointerController``.
    /// It sizes itself through `intrinsicContentSize`, outset margin included.
    var makeView: @MainActor () -> NSView
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
    /// Transparent margin around each item's content, so shadows and selection
    /// halos that paint outside the content are not clipped by its view.
    static let paintOutset: CGFloat = 24

    /// Called when VoiceOver moves to an item through a rotor.
    var onRotorSelect: ((String) -> Void)?

    private weak var container: NSView?
    private var views: [String: NSView] = [:]
    private var models: [String: any Equatable] = [:]
    private var origins: [String: CGPoint] = [:]
    private var dragOffsets: [String: CGSize] = [:]
    /// Item ids in the host's order (bottom to top before raising).
    private(set) var order: [String] = []
    private var rotorEntries: [String: GraphCanvasRotorEntry] = [:]

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
            let view: NSView
            if let existing = views[item.id] {
                if let previous = models[item.id], !item.model.isEqual(to: previous) {
                    existing.removeFromSuperview()
                    let created = makeView(item)
                    container.addSubview(created, positioned: .above, relativeTo: nil)
                    view = created
                } else {
                    view = existing
                }
            } else {
                let created = makeView(item)
                container.addSubview(created, positioned: .above, relativeTo: nil)
                view = created
            }
            views[item.id] = view
            models[item.id] = item.model
            origins[item.id] = item.origin
            rotorEntries[item.id] = item.rotor
            place(item.id)
        }
        order = items.map(\.id)

        for item in items where item.isRaised {
            guard let view = views[item.id], container.subviews.last !== view else { continue }
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
    func view(for id: String) -> NSView? {
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

    /// Creates the item's view and sizes it to its intrinsic size.
    private func makeView(_ item: GraphCanvasItem) -> NSView {
        let view = item.makeView()
        view.frame.size = view.fittingSize
        return view
    }

    private func place(_ id: String) {
        guard let view = views[id], let origin = origins[id] else { return }
        let offset = dragOffsets[id] ?? .zero
        view.setFrameOrigin(CGPoint(
            x: origin.x + offset.width - Self.paintOutset,
            y: origin.y + offset.height - Self.paintOutset
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
        let filter = searchParameters.filterString.lowercased()
        let items = layer.rotorItems(named: name).filter {
            filter.isEmpty || $0.label.lowercased().contains(filter)
        }
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
