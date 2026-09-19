@preconcurrency import AppKit
import SwiftUI

/// Scroll document that owns canvas mouse sequences and hosts paint-only SwiftUI.
///
/// Hit-testing always returns this view so the nested `NSHostingView` never
/// steals pointer events. Empty-canvas drag pans via ``GraphCanvasPointerController``.
final class GraphCanvasDocumentView: NSView {
    let pointer: GraphCanvasPointerController
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
        addSubview(hosting)
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
    }

    override func mouseUp(with event: NSEvent) {
        let doc = convert(event.locationInWindow, from: nil)
        pointer.mouseUp(documentPoint: doc, windowPoint: event.locationInWindow)
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
}
