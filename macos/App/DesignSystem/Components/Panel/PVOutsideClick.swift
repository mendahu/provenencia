import AppKit
import SwiftUI

/// Whether a mouse-down belongs to the sheet. A system open or save panel
/// counts as inside: its clicks are not a click away from the form.
@MainActor
enum PVOutsideClick {
    static func isInsideSheet(
        eventWindow: NSWindow?,
        sheet: NSWindow,
        modalWindow: NSWindow?
    ) -> Bool {
        if modalWindow is NSSavePanel { return true }
        guard let eventWindow else { return false }
        if eventWindow.level.rawValue >= NSWindow.Level.popUpMenu.rawValue { return true }
        var current: NSWindow? = eventWindow
        while let window = current {
            if window === sheet || window is NSSavePanel { return true }
            current = window.parent
        }
        return false
    }
}

/// Sees a mouse-down that lands outside this sheet. A macOS sheet does not
/// deliver that click to the parent window, so the presenter watches the
/// event stream instead of adding a scrim.
private struct PVOutsideClickRepresentable: NSViewRepresentable {
    var isEnabled: Bool
    var onOutsideClick: () -> Void

    func makeNSView(context: Context) -> PVOutsideClickView {
        let view = PVOutsideClickView()
        view.isEnabled = isEnabled
        view.onOutsideClick = onOutsideClick
        return view
    }

    func updateNSView(_ nsView: PVOutsideClickView, context: Context) {
        nsView.isEnabled = isEnabled
        nsView.onOutsideClick = onOutsideClick
    }
}

private final class PVOutsideClickView: NSView {
    var isEnabled = true
    var onOutsideClick: () -> Void = {}
    /// Touched from `deinit`, which is not isolated to the main actor.
    nonisolated(unsafe) private var monitor: Any?

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if window == nil {
            removeMonitor()
        } else {
            installMonitor()
        }
    }

    deinit {
        if let monitor {
            NSEvent.removeMonitor(monitor)
        }
    }

    private func installMonitor() {
        guard monitor == nil else { return }
        monitor = NSEvent.addLocalMonitorForEvents(
            matching: [.leftMouseDown, .rightMouseDown]
        ) { [weak self] event in
            MainActor.assumeIsolated {
                self?.handle(event)
            }
            return event
        }
    }

    private func removeMonitor() {
        if let monitor {
            NSEvent.removeMonitor(monitor)
        }
        monitor = nil
    }

    private func handle(_ event: NSEvent) {
        guard isEnabled, let sheet = window else { return }
        if PVOutsideClick.isInsideSheet(
            eventWindow: event.window,
            sheet: sheet,
            modalWindow: NSApp.modalWindow
        ) { return }
        onOutsideClick()
    }
}

extension View {
    /// A click in the sheet's window, outside its content, runs `perform`. Put
    /// it on a sheet's content. Menus and popovers are not this path, and a
    /// system open or save panel counts as inside. `isEnabled` false pauses it.
    func pvOutsideClick(isEnabled: Bool = true, perform: @escaping () -> Void) -> some View {
        background {
            PVOutsideClickRepresentable(isEnabled: isEnabled, onOutsideClick: perform)
        }
    }
}
