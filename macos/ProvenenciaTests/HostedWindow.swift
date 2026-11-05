import AppKit
import SwiftUI

/// Hosts `view` in a borderless window so layout, `.task` and geometry
/// callbacks run. The hosting view is sized after it becomes the content view.
/// The caller holds the returned window for the test's length, by calling
/// `withExtendedLifetime(window) {}` after the last assertion.
@MainActor
func hostInBorderlessWindow<V: View>(
    _ view: V,
    windowSize: CGSize,
    hostSize: CGSize? = nil,
    orderFront: Bool = true
) -> (NSWindow, NSHostingView<V>) {
    let host = NSHostingView(rootView: view)
    let window = borderlessWindow(size: windowSize)
    window.contentView = host
    host.frame = NSRect(origin: .zero, size: hostSize ?? windowSize)
    if orderFront { window.orderFrontRegardless() }
    return (window, host)
}

/// Hosts an AppKit `view` as the content view of a borderless window of `size`.
/// The caller holds the returned window for the test's length.
@MainActor
func hostInBorderlessWindow(nsView view: NSView, size: CGSize) -> NSWindow {
    let window = borderlessWindow(size: size)
    window.contentView = view
    window.orderFrontRegardless()
    return window
}

@MainActor
private func borderlessWindow(size: CGSize) -> NSWindow {
    NSWindow(
        contentRect: NSRect(origin: .zero, size: size),
        styleMask: .borderless,
        backing: .buffered,
        defer: false
    )
}
