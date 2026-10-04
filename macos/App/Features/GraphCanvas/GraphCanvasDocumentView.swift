@preconcurrency import AppKit
import SwiftUI

/// Scroll document that owns canvas mouse sequences and hosts paint-only SwiftUI.
///
/// Hit-testing always returns this view so nested `NSHostingView`s never
/// steal pointer events. Empty-canvas drag pans via ``GraphCanvasPointerController``.
///
/// Two layers of content: the host's base SwiftUI view (grid, edges, overlays)
/// at the bottom, and ``items`` — one hosting view per card — above it. A card
/// drag moves its item view directly; no SwiftUI runs per pointer event.
final class GraphCanvasDocumentView: NSView {
    let pointer: GraphCanvasPointerController
    private(set) lazy var items = GraphCanvasItemLayer(container: self)
    private var hostingView: NSHostingView<AnyView>?
    private var trackingArea: NSTrackingArea?

    override var isFlipped: Bool { true }

    override var acceptsFirstResponder: Bool { false }

    init(pointer: GraphCanvasPointerController, frame: NSRect) {
        self.pointer = pointer
        super.init(frame: frame)
        wantsLayer = true
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    func installHostingRoot(_ root: AnyView) {
        if let hostingView {
            hostingView.rootView = root
            hostingView.frame = bounds
            return
        }
        let hosting = NSHostingView(rootView: root)
        hosting.frame = bounds
        hosting.autoresizingMask = [.width, .height]
        addSubview(hosting, positioned: .below, relativeTo: nil)
        hostingView = hosting
    }

    func resizeDocument(to size: CGSize) {
        setFrameSize(size)
        hostingView?.frame = bounds
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        updateTrackingAreas()
    }

    override func updateTrackingAreas() {
        super.updateTrackingAreas()
        if let trackingArea {
            removeTrackingArea(trackingArea)
        }
        let options: NSTrackingArea.Options = [
            .activeInKeyWindow,
            .mouseMoved,
            .mouseEnteredAndExited,
            .inVisibleRect,
        ]
        let area = NSTrackingArea(rect: .zero, options: options, owner: self, userInfo: nil)
        addTrackingArea(area)
        trackingArea = area
    }

    override func hitTest(_ point: NSPoint) -> NSView? {
        bounds.contains(point) ? self : nil
    }

    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }

    override func mouseDown(with event: NSEvent) {
        let doc = convert(event.locationInWindow, from: nil)
        pointer.mouseDown(documentPoint: doc, windowPoint: event.locationInWindow)
    }

    override func mouseDragged(with event: NSEvent) {
        let doc = convert(event.locationInWindow, from: nil)
        pointer.mouseDragged(documentPoint: doc, windowPoint: event.locationInWindow)
        items.applyDragOffsets(pointer.offsets)
    }

    override func mouseUp(with event: NSEvent) {
        let doc = convert(event.locationInWindow, from: nil)
        pointer.mouseUp(documentPoint: doc, windowPoint: event.locationInWindow)
        // The dropped item returns to its pushed origin; the host's drop
        // handler pushes the new origin in the same update.
        items.applyDragOffsets(pointer.offsets)
    }

    override func mouseMoved(with event: NSEvent) {
        let doc = convert(event.locationInWindow, from: nil)
        pointer.mouseMoved(documentPoint: doc)
    }

    override func mouseEntered(with event: NSEvent) {
        pointer.mouseEntered()
    }

    override func mouseExited(with event: NSEvent) {
        pointer.mouseExited()
    }

    // MARK: - Accessibility

    /// A group named by the host, so VoiceOver can interact with the canvas
    /// and use its item rotors.
    override func isAccessibilityElement() -> Bool { true }

    override func accessibilityRole() -> NSAccessibility.Role? { .group }

    override func accessibilityCustomRotors() -> [NSAccessibilityCustomRotor] {
        rotorSearches = items.rotorNames.map { GraphCanvasRotorSearch(name: $0, layer: items) }
        return rotorSearches.map { NSAccessibilityCustomRotor(label: $0.name, itemSearchDelegate: $0) }
    }

    /// Rotors hold their search delegate weakly; keep the current ones alive.
    private var rotorSearches: [GraphCanvasRotorSearch] = []
}
