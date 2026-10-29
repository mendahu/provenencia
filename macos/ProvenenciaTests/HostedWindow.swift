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
    let window = NSWindow(
        contentRect: NSRect(origin: .zero, size: windowSize),
        styleMask: .borderless,
        backing: .buffered,
        defer: false
    )
    window.contentView = host
    host.frame = NSRect(origin: .zero, size: hostSize ?? windowSize)
    if orderFront { window.orderFrontRegardless() }
    return (window, host)
}
