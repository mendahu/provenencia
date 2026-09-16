@preconcurrency import AppKit
import SwiftUI

/// Where a `WindowDragRegion` sits relative to the view's own content.
enum PVWindowDragPlacement {
    /// On top of everything. Only for chrome with no interactive controls
    /// of its own (e.g. the sidebar brand row) — an overlay swallows the
    /// hover and clicks of anything beneath it.
    case overlay
    /// Behind the view's content, so controls keep their own clicks and
    /// only the space around them drags the window. What a row like
    /// `WorkspaceToolbar` needs: its buttons are ~26pt tall inside a 52pt
    /// row, and without this the strips above and below them are dead.
    case behindContent
}

/// Whether the press being handled right now has already moved the window.
///
/// A `.behindContent` drag surface and the controls painted on top of it
/// both act on the same mouse-down: AppKit hands it to `DragView`, which
/// blocks in `performDrag` until mouse-up, and SwiftUI then resumes its own
/// gesture for whatever control was under the cursor. So dragging the window
/// by the Back button also *pressed* the Back button, and a slow drag
/// tripped its long-press jump menu.
///
/// Every press over the surface reaches `DragView.mouseDown` before SwiftUI
/// acts on it — that is the same delivery order that lets you drag from a
/// control at all — so each press can clear the previous press's verdict.
/// No timers: `didDrag` is false again the instant the next press starts.
@MainActor
enum WindowDrag {
    private(set) static var didDrag = false

    /// Runs `action` unless the current press has moved the window. Controls
    /// sitting on a drag surface put their work behind this.
    static func unlessDragging(_ action: () -> Void) {
        guard !didDrag else { return }
        action()
    }

    /// Called by the drag surface on every mouse-down it receives, and
    /// internal (not `fileprivate`) so the press/drag sequence is testable
    /// without a real window — same escape hatch as `TrafficLights`.
    static func beginPress() { didDrag = false }

    static func noteWindowMoved() { didDrag = true }
}

/// Transparent AppKit surface that turns mouse-downs into window moves.
///
/// With `.windowStyle(.hiddenTitleBar)`, AppKit still only treats roughly
/// the system title-bar strip as a drag region. Our workspace header is
/// taller (`WorkspaceChrome.headerHeight`), so clicks in the lower part
/// of that visual row would otherwise do nothing. `WindowDragGesture` is
/// macOS 15+; this is the macOS 14-compatible escape hatch.
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
            guard let window else { return }
            WindowDrag.beginPress()
            let origin = window.frame.origin
            // Posted synchronously (`queue: nil`) so the flag flips *during*
            // the drag, in time for a long-press gesture that fires while
            // `performDrag` is still running its modal event loop.
            let observer = NotificationCenter.default.addObserver(
                forName: NSWindow.didMoveNotification,
                object: window,
                queue: nil
            ) { _ in
                MainActor.assumeIsolated { WindowDrag.noteWindowMoved() }
            }
            defer { NotificationCenter.default.removeObserver(observer) }
            window.performDrag(with: event)
            if window.frame.origin != origin {
                WindowDrag.noteWindowMoved()
            }
        }

        override func accessibilityIsIgnored() -> Bool { true }
    }
}

extension View {
    /// Makes the view's full bounds a window-drag surface (see
    /// `WindowDragRegion`). Use `.behindContent` on any row that has
    /// interactive SwiftUI controls; `.overlay` swallows their hits.
    func pvWindowDragRegion(_ placement: PVWindowDragPlacement = .overlay) -> some View {
        modifier(WindowDragRegionModifier(placement: placement))
    }
}

private struct WindowDragRegionModifier: ViewModifier {
    let placement: PVWindowDragPlacement

    func body(content: Content) -> some View {
        switch placement {
        case .overlay:
            content.overlay { region }
        case .behindContent:
            content.background { region }
        }
    }

    private var region: some View {
        WindowDragRegion()
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .accessibilityHidden(true)
    }
}
