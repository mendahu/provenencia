@preconcurrency import AppKit
import SwiftUI

/// Transparent AppKit overlay that turns mouse-downs into window moves.
///
/// With `.windowStyle(.hiddenTitleBar)`, AppKit still only treats roughly
/// the system title-bar strip as a drag region. Our workspace header is
/// taller (`WorkspaceChrome.headerHeight`), so clicks in the lower part
/// of that visual row would otherwise do nothing. `WindowDragGesture` is
/// macOS 15+; this is the macOS 14-compatible escape hatch.
///
/// **Only** attach over chrome with no interactive controls of its own
/// (e.g. the sidebar brand row). A full-row overlay on `WorkspaceToolbar`
/// steals hover/clicks from Back/Forward, breadcrumbs, and the omnibar.
private struct WindowDragRegion: NSViewRepresentable {
    @MainActor
    func makeNSView(context: Context) -> NSView {
        DragView()
    }

    @MainActor
    func updateNSView(_: NSView, context: Context) {}

    @MainActor
    private final class DragView: NSView {
        override func mouseDown(with event: NSEvent) {
            window?.performDrag(with: event)
        }

        override func accessibilityIsIgnored() -> Bool { true }
    }
}

extension View {
    /// Makes the view's full bounds a window-drag surface (see
    /// `WindowDragRegion`). Attach only to header chrome that has **no**
    /// interactive SwiftUI controls — the overlay sits on top and will
    /// otherwise swallow hits.
    func pvWindowDragRegion() -> some View {
        overlay {
            WindowDragRegion()
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .accessibilityHidden(true)
        }
    }
}
