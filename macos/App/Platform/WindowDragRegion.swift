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
/// Dragging the window by the Back button used to step Back too, and a slow
/// drag also tripped its long-press jump menu. The window moves there without
/// `WindowDragRegion` being involved at all: a `.hiddenTitleBar` window still
/// carries AppKit's title-bar strip across the top ~28pt of the content, and
/// that strip moves the window on a drag *while* passing the click through to
/// the SwiftUI control underneath. Our 52pt header row overlaps it, so its
/// controls get both.
///
/// Watching the window instead of any one drag mechanism catches that strip
/// and `WindowDragRegion`'s own `performDrag` alike, since both post
/// `didMove`. A press counts as a drag once the window moves while the button
/// is held, and the next mouse-down clears the verdict — no timers.
@MainActor
enum WindowDrag {
    private(set) static var didDrag = false
    private static var pressMonitor: Any?

    /// Installed once for the app's lifetime from `ProvenenciaApp`, ahead of
    /// any press: arming this from a view's `onAppear` would miss presses on
    /// whatever chrome appears first.
    static func startTrackingPresses() {
        guard pressMonitor == nil else { return }
        // A local monitor sees the event before it reaches any view, so the
        // reset always lands ahead of the control's own handling of it.
        pressMonitor = NSEvent.addLocalMonitorForEvents(matching: [.leftMouseDown]) { event in
            MainActor.assumeIsolated { pressBegan() }
            return event
        }
        NotificationCenter.default.addObserver(
            forName: NSWindow.didMoveNotification,
            object: nil,
            queue: nil
        ) { _ in
            MainActor.assumeIsolated {
                // Programmatic moves (restore, zoom) are not drags; only a
                // move made with the button held is the user dragging.
                guard NSEvent.pressedMouseButtons != 0 else { return }
                windowMovedUnderPress()
            }
        }
    }

    /// Runs `action` unless the press in flight has moved the window.
    /// Controls that sit on window chrome put their work behind this.
    static func unlessDragging(_ action: () -> Void) {
        guard !didDrag else { return }
        action()
    }

    /// The two transitions, internal rather than private so the press/drag
    /// sequence is testable without synthesizing real drags — the same escape
    /// hatch `TrafficLights.verticalOrigin` takes.
    static func pressBegan() { didDrag = false }

    static func windowMovedUnderPress() { didDrag = true }
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
        // The resulting move is picked up by `WindowDrag`, which watches every
        // window rather than this one drag path.
        override func mouseDown(with event: NSEvent) {
            window?.performDrag(with: event)
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
