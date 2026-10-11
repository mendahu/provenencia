@preconcurrency import AppKit
import SwiftUI

/// One host-drawn item the canvas positions in AppKit — a graph card.
///
/// Each item is its own `NSHostingView`, a subview of
/// ``GraphCanvasDocumentView`` above the host's base SwiftUI layer. AppKit owns
/// where items are: a drag moves one view's frame and runs no SwiftUI. SwiftUI
/// only draws an item's content, and only rebuilds it when ``model`` changes.
struct GraphCanvasItem {
    let id: String
    /// Document-space top-leading corner of the item's content.
    var origin: CGPoint
    /// Document-space size of the content. The layer uses it to decide which
    /// cards are near the window before a hosting view exists.
    var size: CGSize = CGSize(width: 100, height: 40)
    /// Stay mounted even when the card's rect is outside the window (selected
    /// or dragged). The layer does not know why.
    var staysMounted: Bool = false
    /// Drawn above every other item (the item being dragged or activated).
    var isRaised: Bool = false
    /// The VoiceOver rotor that lists this item, if any.
    var rotor: GraphCanvasRotorEntry?
    /// Everything the content draws. The hosted view is rebuilt only when this
    /// changes, so a host can push every item on any state change.
    var model: any Equatable
    /// Builds the content for ``model``. Content is paint plus accessibility;
    /// pointer input stays with ``GraphCanvasPointerController``.
    var content: @MainActor () -> AnyView
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
    /// halos that paint outside the content are not clipped by its hosting view.
    static let paintOutset: CGFloat = 24

    /// Called when VoiceOver moves to an item through a rotor.
    var onRotorSelect: ((String) -> Void)?

    /// Screen points of prefetch around the window. Divided by magnification
    /// so a pan still finds the next card already drawn at any zoom.
    static let screenMountMargin: CGFloat = 240

    private weak var container: NSView?
    private var catalog: [String: GraphCanvasItem] = [:]
    private var views: [String: NSHostingView<AnyView>] = [:]
    /// Model the on-screen view was built with. Absent when the card has no view.
    private var displayedModels: [String: any Equatable] = [:]
    private var origins: [String: CGPoint] = [:]
    private var dragOffsets: [String: CGSize] = [:]
    /// Item ids in the host's order (bottom to top before raising).
    private(set) var order: [String] = []
    private var rotorEntries: [String: GraphCanvasRotorEntry] = [:]
    private var raisedIDs: Set<String> = []
    /// Last document rect the scroll view reported. Zero until it has laid out,
    /// which mounts nothing.
    private var visibleDocumentRect: CGRect = .zero
    private var magnification: CGFloat = 1

    /// Replacements wait here until every one of them is built, then swap together.
    private var replaceQueue: [String] = []
    /// Mounts are inserted as each slice finishes. They are not part of the swap.
    private var mountQueue: [String] = []
    /// Model each queued card is being built for. A pan does not clear the others.
    private var queuedModels: [String: any Equatable] = [:]
    /// Replacement views not yet on screen.
    private var prepared: [String: NSHostingView<AnyView>] = [:]
    private var pumpScheduled = false

    /// One card can be built immediately. More than that is sliced so a zoom
    /// through the face cut is not stuck while views lay out.
    private static let synchronousContentLimit = 1
    private static let contentChunkSize = 2

    init(container: NSView) {
        self.container = container
    }

    /// Replaces the item set. Every card is remembered; a hosting view is
    /// created only for cards near the window, plus any card flagged
    /// ``GraphCanvasItem/staysMounted``.
    ///
    /// One card is built immediately, so a pan finds it already drawn. When
    /// many cards need views at once (a zoom through the face cut, or a zoom
    /// that reveals the whole graph), they are laid out a couple at a time.
    /// Cards already on screen keep their old face until every replacement is
    /// ready, then swap together. Cards that had no view are inserted as each
    /// slice finishes. Leaving the window removes a view immediately and drops
    /// only that card's pending work.
    func update(_ items: [GraphCanvasItem]) {
        let ids = Set(items.map(\.id))
        for id in catalog.keys where !ids.contains(id) {
            drop(id)
        }
        for item in items {
            catalog[item.id] = item
            origins[item.id] = item.origin
            rotorEntries[item.id] = item.rotor
        }
        order = items.map(\.id)
        raisedIDs = Set(items.filter(\.isRaised).map(\.id))
        for id in views.keys {
            place(id)
        }
        raiseRaisedViews()
        reconcile()
    }

    /// Document rect currently on screen, and the magnification that produced it.
    /// A zero rect mounts nothing; the scroll view calls this again once it has laid out.
    func setVisibleDocumentRect(_ rect: CGRect, magnification: CGFloat) {
        guard rect != visibleDocumentRect || magnification != self.magnification else { return }
        visibleDocumentRect = rect
        self.magnification = magnification
        reconcile()
    }

    /// Scrolls `id` into view and mounts it now, so a rotor entry off screen
    /// has a hosting view to land on.
    func reveal(_ id: String) {
        guard let item = catalog[id], let container else { return }
        container.scrollToVisible(cardRect(for: item))
        guard views[id] == nil || !(displayedModels[id]?.isEqual(to: item.model) ?? false) else { return }
        cancelWork(id)
        if views[id] != nil {
            install(item)
        } else {
            insert(item)
        }
        place(id)
        raiseRaisedViews()
    }

    private func reconcile() {
        let mount = mountIDs()
        for id in Array(views.keys) where !mount.contains(id) {
            unmount(id)
        }
        replaceQueue.removeAll { !mount.contains($0) }
        mountQueue.removeAll { !mount.contains($0) }
        for id in Array(prepared.keys) where !mount.contains(id) {
            prepared[id] = nil
            queuedModels[id] = nil
        }

        for id in order where mount.contains(id) {
            guard let item = catalog[id] else { continue }
            if let shown = displayedModels[id], views[id] != nil, shown.isEqual(to: item.model) {
                cancelWork(id)
                continue
            }
            let kind: WorkKind = views[id] == nil ? .mount : .replace
            enqueue(id, kind: kind, model: item.model)
        }

        let pending = replaceQueue.count + mountQueue.count
        if pending == 0 {
            if !prepared.isEmpty {
                commitReplacements()
            }
            return
        }
        if pending <= Self.synchronousContentLimit, prepared.isEmpty, !pumpScheduled {
            performOneSync()
            return
        }
        schedulePump()
    }

    private enum WorkKind {
        case replace
        case mount
    }

    private func mountIDs() -> Set<String> {
        guard visibleDocumentRect.width > 0, visibleDocumentRect.height > 0 else { return [] }
        let margin = Self.screenMountMargin / max(magnification, 0.000_1)
        let expanded = visibleDocumentRect.insetBy(dx: -margin, dy: -margin)
        var ids = Set<String>()
        for (id, item) in catalog {
            if item.staysMounted || expanded.intersects(cardRect(for: item)) {
                ids.insert(id)
            }
        }
        return ids
    }

    private func cardRect(for item: GraphCanvasItem) -> CGRect {
        let offset = dragOffsets[item.id] ?? .zero
        return CGRect(
            x: item.origin.x + offset.width,
            y: item.origin.y + offset.height,
            width: item.size.width,
            height: item.size.height
        )
    }

    private func enqueue(_ id: String, kind: WorkKind, model: any Equatable) {
        if let queued = queuedModels[id], queued.isEqual(to: model) {
            switch kind {
            case .replace where replaceQueue.contains(id) || prepared[id] != nil:
                return
            case .mount where mountQueue.contains(id):
                return
            default:
                break
            }
        }
        prepared[id] = nil
        queuedModels[id] = model
        replaceQueue.removeAll { $0 == id }
        mountQueue.removeAll { $0 == id }
        switch kind {
        case .replace: replaceQueue.append(id)
        case .mount: mountQueue.append(id)
        }
    }

    private func cancelWork(_ id: String) {
        replaceQueue.removeAll { $0 == id }
        mountQueue.removeAll { $0 == id }
        prepared[id] = nil
        queuedModels[id] = nil
    }

    private func performOneSync() {
        let id = replaceQueue.first ?? mountQueue.first
        guard let id, let item = catalog[id] else { return }
        cancelWork(id)
        if views[id] != nil {
            install(item)
        } else {
            insert(item)
        }
        place(id)
        raiseRaisedViews()
    }

    private func schedulePump() {
        guard !pumpScheduled else { return }
        pumpScheduled = true
        DispatchQueue.main.async { [weak self] in
            self?.prepareSlice()
        }
    }

    private func prepareSlice() {
        var budget = Self.contentChunkSize
        while budget > 0, !replaceQueue.isEmpty {
            let id = replaceQueue.removeFirst()
            guard let item = catalog[id], views[id] != nil else {
                queuedModels[id] = nil
                continue
            }
            prepared[id] = makeView(for: item)
            queuedModels[id] = item.model
            budget -= 1
        }
        if replaceQueue.isEmpty, !prepared.isEmpty {
            commitReplacements()
        }
        while budget > 0, !mountQueue.isEmpty {
            let id = mountQueue.removeFirst()
            guard let item = catalog[id], views[id] == nil else {
                queuedModels[id] = nil
                continue
            }
            insert(item)
            queuedModels[id] = nil
            place(id)
            budget -= 1
        }
        raiseRaisedViews()
        if !replaceQueue.isEmpty || !mountQueue.isEmpty {
            DispatchQueue.main.async { [weak self] in
                self?.prepareSlice()
            }
        } else {
            pumpScheduled = false
        }
    }

    private func commitReplacements() {
        guard let container else { return }
        for (id, newView) in prepared {
            guard views[id] != nil, let item = catalog[id] else { continue }
            views[id]?.removeFromSuperview()
            container.addSubview(newView, positioned: .above, relativeTo: nil)
            views[id] = newView
            displayedModels[id] = item.model
            queuedModels[id] = nil
            place(id)
        }
        prepared = [:]
        raiseRaisedViews()
    }

    private func insert(_ item: GraphCanvasItem) {
        guard let container else { return }
        let view = makeView(for: item)
        container.addSubview(view, positioned: .above, relativeTo: nil)
        views[item.id] = view
        displayedModels[item.id] = item.model
    }

    private func install(_ item: GraphCanvasItem) {
        guard let container, let view = views[item.id] else { return }
        let newView = makeView(for: item)
        view.removeFromSuperview()
        container.addSubview(newView, positioned: .above, relativeTo: nil)
        views[item.id] = newView
        displayedModels[item.id] = item.model
    }

    private func unmount(_ id: String) {
        views[id]?.removeFromSuperview()
        views[id] = nil
        displayedModels[id] = nil
        cancelWork(id)
    }

    private func drop(_ id: String) {
        unmount(id)
        catalog[id] = nil
        origins[id] = nil
        dragOffsets[id] = nil
        rotorEntries[id] = nil
    }

    private func makeView(for item: GraphCanvasItem) -> NSHostingView<AnyView> {
        let view = NSHostingView(rootView: Self.padded(item.content()))
        view.frame.size = view.fittingSize
        return view
    }

    private func raiseRaisedViews() {
        guard let container else { return }
        for id in order where raisedIDs.contains(id) {
            guard let view = views[id], container.subviews.last !== view else { continue }
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

    /// The hosting view for an item, for accessibility targets and tests.
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

    private func place(_ id: String) {
        guard let view = views[id], let origin = origins[id] else { return }
        let offset = dragOffsets[id] ?? .zero
        view.setFrameOrigin(CGPoint(
            x: origin.x + offset.width - Self.paintOutset,
            y: origin.y + offset.height - Self.paintOutset
        ))
    }

    private static func padded(_ content: AnyView) -> AnyView {
        AnyView(
            content
                .fixedSize()
                .padding(paintOutset)
        )
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
        guard items.indices.contains(index) else { return nil }
        let id = items[index].id
        layer.reveal(id)
        guard let view = layer.view(for: id) else { return nil }

        let result = NSAccessibilityCustomRotor.ItemResult(targetElement: view)
        result.customLabel = items[index].label
        layer.onRotorSelect?(id)
        return result
    }
}
