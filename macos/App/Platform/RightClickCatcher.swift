import AppKit
import SwiftUI

/// Transparent hit target that reports right-clicks (and Ctrl-click) without
/// consuming left-clicks — used for custom context menus that replace `NSMenu`.
struct RightClickCatcher: NSViewRepresentable {
    var onRightClick: (CGPoint) -> Void

    func makeNSView(context: Context) -> RightClickCatcherView {
        let view = RightClickCatcherView()
        view.onRightClick = onRightClick
        return view
    }

    func updateNSView(_ nsView: RightClickCatcherView, context: Context) {
        nsView.onRightClick = onRightClick
    }
}

final class RightClickCatcherView: NSView {
    var onRightClick: ((CGPoint) -> Void)?

    override var acceptsFirstResponder: Bool { false }

    override func hitTest(_ point: NSPoint) -> NSView? {
        // Only claim right-clicks / Ctrl-clicks so left-clicks pass through.
        guard let event = NSApp.currentEvent else { return nil }
        if event.type == .rightMouseDown { return self }
        if event.type == .leftMouseDown, event.modifierFlags.contains(.control) { return self }
        return nil
    }

    override func rightMouseDown(with event: NSEvent) {
        report(event)
    }

    override func mouseDown(with event: NSEvent) {
        if event.modifierFlags.contains(.control) {
            report(event)
        }
    }

    private func report(_ event: NSEvent) {
        // Local point in this view — callers position menus relative to the host.
        let point = convert(event.locationInWindow, from: nil)
        onRightClick?(point)
    }
}
