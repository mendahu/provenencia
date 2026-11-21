import AppKit
import SwiftUI

/// What differs between popups. A difference a popup really has is a field
/// here, never a second watcher.
struct PVPopupOptions: Equatable {
    /// The popup's keys belong to a field that stays focused (a `PVSelect`),
    /// so the watcher does not watch keys at all, Escape included.
    var keysStayInField = false
    /// A mouse event on the popup's anchor (its field) passes: the popup stays
    /// open on a click in its own field (a combo box's list).
    var staysOpenOnFieldClick = false
    /// Right-clicks pass, inside or outside (the history menu: they switch
    /// between Back and Forward).
    var ignoresRightClicks = false
}

enum PVPopupMouseButton: Equatable {
    case left, right, other
}

/// One event the watcher sees, reduced to what the rule reads.
enum PVPopupInput: Equatable {
    case key(KeyEquivalent)
    case mouseDown(PVPopupMouseButton, inside: Bool)
    case mouseUp(PVPopupMouseButton, inside: Bool)
    case windowResignedKey
    /// A left press outside ended and its release never reached the watcher: a
    /// nested tracking loop (a scroller or title-bar drag, a native control)
    /// consumed it. It closes as that release would.
    case pressEndedUnseen

    /// Nil for an event type, or a key, the watcher does not take.
    @MainActor
    init?(_ event: NSEvent) {
        let inside = Self.isPopupWindow(level: event.window?.level)
        switch event.type {
        case .keyDown:
            guard let key = Self.keyEquivalent(
                keyCode: event.keyCode,
                characters: event.charactersIgnoringModifiers,
                modifiers: event.modifierFlags
            ) else { return nil }
            self = .key(key)
        case .leftMouseDown: self = .mouseDown(.left, inside: inside)
        case .rightMouseDown: self = .mouseDown(.right, inside: inside)
        case .otherMouseDown: self = .mouseDown(.other, inside: inside)
        case .leftMouseUp: self = .mouseUp(.left, inside: inside)
        case .rightMouseUp: self = .mouseUp(.right, inside: inside)
        case .otherMouseUp: self = .mouseUp(.other, inside: inside)
        default: return nil
        }
    }

    /// The keys a popup takes, as the SwiftUI vocabulary the kit's lists,
    /// tables and selects already speak. A letter or digit with neither ⌘ nor
    /// ⌃ type-selects; anything else is not taken and the event passes.
    static func keyEquivalent(
        keyCode: UInt16,
        characters: String?,
        modifiers: NSEvent.ModifierFlags
    ) -> KeyEquivalent? {
        switch keyCode {
        case 53: return .escape
        case 126: return .upArrow
        case 125: return .downArrow
        case 115: return .home
        case 119: return .end
        case 36, 76: return .return
        default:
            return PVTypeSelectMatcher.typeSelectCharacter(
                characters, command: modifiers.contains(.command), control: modifiers.contains(.control)
            ).map { KeyEquivalent($0) }
        }
    }

    /// Every kit popup is a child panel at `.popUpMenu` level, so a window at
    /// or above that level is the popup (or a system menu over it), not the
    /// page. Sheets use the same test (`PVOutsideClick.isInsideSheet`).
    static func isPopupWindow(level: NSWindow.Level?) -> Bool {
        guard let level else { return false }
        return level.rawValue >= NSWindow.Level.popUpMenu.rawValue
    }
}

/// What the watcher does with one event. The whole closing rule, decided
/// without a window, in the shape of `PVSheetDismiss.action`.
enum PVPopupDecision: Equatable {
    /// The event goes on, untouched.
    case pass
    /// Close the popup; `consumesEvent` also swallows the event.
    case close(consumesEvent: Bool)
    /// The press that opened the popup is being released (`PVPopupOpeningPress`);
    /// the event goes on.
    case endOpeningPress
    /// Hand the key to the popup; it is swallowed when the popup handles it.
    case toPopup(KeyEquivalent)

    /// `onAnchor` is whether a mouse event landed on the popup's anchor;
    /// `pressBeganInside` is whether this release's press began inside the
    /// popup, or on the anchor with `staysOpenOnFieldClick`. A click is outside
    /// only when its press began outside.
    static func decide(
        _ input: PVPopupInput,
        openingPressPending: Bool,
        onAnchor: Bool = false,
        pressBeganInside: Bool = false,
        options: PVPopupOptions
    ) -> PVPopupDecision {
        let anchorKeepsOpen = options.staysOpenOnFieldClick && onAnchor
        switch input {
        case .key(let key):
            if options.keysStayInField { return .pass }
            return key == .escape ? .close(consumesEvent: true) : .toPopup(key)
        case .mouseUp(let button, let inside):
            if openingPressPending { return .endOpeningPress }
            guard button == .left, !inside, !anchorKeepsOpen, !pressBeganInside else { return .pass }
            return .close(consumesEvent: false)
        case .mouseDown(let button, let inside):
            if options.ignoresRightClicks, button == .right { return .pass }
            guard button != .left, !inside, !anchorKeepsOpen else { return .pass }
            return .close(consumesEvent: false)
        case .windowResignedKey:
            return .close(consumesEvent: false)
        case .pressEndedUnseen:
            return pressBeganInside ? .pass : .close(consumesEvent: false)
        }
    }
}

/// The press that opened a popup, and whether its release is still to come
/// (D-5). A release only ends it, and passes through, never closes the popup.
/// When the popup opened on a mouse-down, that press's `eventNumber` names the
/// release (a mouse-up carries its mouse-down's number), so a tap-to-click or
/// fast click whose release is still queued when the watcher arms is still
/// recognised. Otherwise the buttons held at arm time say it: the next
/// mouse-up of any button.
struct PVPopupOpeningPress {
    private var openingNumber: Int?
    private var isPending: Bool

    init(openingEvent: NSEvent?, pressedMouseButtons: Int) {
        switch openingEvent?.type {
        case .leftMouseDown, .rightMouseDown, .otherMouseDown:
            openingNumber = openingEvent?.eventNumber
            isPending = true
        default:
            openingNumber = nil
            isPending = pressedMouseButtons != 0
        }
    }

    /// True once, for the mouse-up that releases the opening press.
    mutating func endsOpeningPress(_ event: NSEvent) -> Bool {
        guard isPending else { return false }
        if let openingNumber, event.eventNumber != openingNumber { return false }
        isPending = false
        return true
    }

    /// For the decision table: nothing pending.
    static let none = PVPopupOpeningPress(openingEvent: nil, pressedMouseButtons: 0)
}

private struct PVPopupRepresentable: NSViewRepresentable {
    var isPresented: Bool
    var options: PVPopupOptions
    var onKey: (KeyEquivalent) -> KeyPress.Result
    var close: () -> Void

    func makeNSView(context: Context) -> PVPopupWatcherView {
        let view = PVPopupWatcherView()
        apply(to: view)
        return view
    }

    func updateNSView(_ nsView: PVPopupWatcherView, context: Context) {
        apply(to: nsView)
    }

    private func apply(to view: PVPopupWatcherView) {
        view.options = options
        view.onKey = onKey
        view.close = close
        view.isPresented = isPresented
    }
}

/// The one event watcher of a floating popup, attached to its anchor. It owns
/// no window: the popup keeps its own, and this only decides when it closes.
@MainActor
private final class PVPopupWatcherView: NSView {
    var options = PVPopupOptions()
    var onKey: (KeyEquivalent) -> KeyPress.Result = { _ in .ignored }
    var close: () -> Void = {}
    var isPresented = false {
        didSet { if isPresented != oldValue { sync() } }
    }

    /// Touched from `deinit`, which is not isolated to the main actor.
    nonisolated(unsafe) private var monitor: Any?
    nonisolated(unsafe) private var resignObserver: Any?
    private var isArming = false
    /// The event current when the popup turned presented (D-5).
    private var openingEvent: NSEvent?
    private var openingPress = PVPopupOpeningPress.none
    /// Whether the left press now held began inside the popup or on the field.
    private var pressBeganInside = false
    /// A left press outside that passed and has not been released in view; a
    /// tracking loop that swallows the release leaves it set (`watchPress`).
    private var outsidePressPending = false

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        sync()
    }

    deinit {
        if let monitor { NSEvent.removeMonitor(monitor) }
        if let resignObserver { NotificationCenter.default.removeObserver(resignObserver) }
    }

    private func sync() {
        if isPresented, window != nil {
            guard monitor == nil, !isArming else { return }
            // One main-queue turn after it opens, so the event that opened it
            // is already delivered.
            isArming = true
            openingEvent = NSApp.currentEvent
            DispatchQueue.main.async { [weak self] in
                guard let self else { return }
                isArming = false
                if isPresented, window != nil { arm() }
            }
        } else {
            disarm()
        }
    }

    private func arm() {
        guard monitor == nil, let window else { return }
        openingPress = PVPopupOpeningPress(
            openingEvent: openingEvent, pressedMouseButtons: NSEvent.pressedMouseButtons
        )
        var mask: NSEvent.EventTypeMask = [
            .leftMouseDown, .leftMouseUp, .rightMouseUp, .otherMouseUp, .rightMouseDown, .otherMouseDown,
        ]
        if !options.keysStayInField { mask.insert(.keyDown) }
        monitor = NSEvent.addLocalMonitorForEvents(matching: mask) { [weak self] event in
            let consumed = MainActor.assumeIsolated { self?.handle(event) ?? false }
            return consumed ? nil : event
        }
        resignObserver = NotificationCenter.default.addObserver(
            forName: NSWindow.didResignKeyNotification, object: window, queue: .main
        ) { [weak self] _ in
            MainActor.assumeIsolated {
                _ = self?.apply(.windowResignedKey)
            }
        }
    }

    private func disarm() {
        if let monitor { NSEvent.removeMonitor(monitor) }
        if let resignObserver { NotificationCenter.default.removeObserver(resignObserver) }
        monitor = nil
        resignObserver = nil
        openingEvent = nil
        openingPress = .none
        pressBeganInside = false
        outsidePressPending = false
    }

    /// True when the event is swallowed.
    private func handle(_ event: NSEvent) -> Bool {
        guard let input = PVPopupInput(event) else { return false }
        let onAnchor = isOnAnchor(event)
        let releasesOpeningPress: Bool
        var beganInside = false
        switch input {
        case .mouseUp(let button, _):
            releasesOpeningPress = openingPress.endsOpeningPress(event)
            if button == .left {
                beganInside = pressBeganInside
                pressBeganInside = false
                outsidePressPending = false
            }
        case .mouseDown(.left, let inside):
            releasesOpeningPress = false
            pressBeganInside = inside || (options.staysOpenOnFieldClick && onAnchor)
            if !pressBeganInside { watchPress() }
        default:
            releasesOpeningPress = false
        }
        return apply(
            input, openingPressPending: releasesOpeningPress,
            onAnchor: onAnchor, pressBeganInside: beganInside
        )
    }

    /// A local monitor is not called for events a nested tracking loop consumes
    /// (a scroller or title-bar drag, a native control), so a press outside
    /// would never see its release. The queued block runs in the default mode,
    /// so only once any such loop returns; if the button is up by then and no
    /// release was seen, the press closes the popup as its release would.
    private func watchPress() {
        outsidePressPending = true
        RunLoop.main.perform(inModes: [.default]) { [weak self] in
            MainActor.assumeIsolated {
                guard let self, self.outsidePressPending, NSEvent.pressedMouseButtons & 1 == 0 else { return }
                self.outsidePressPending = false
                _ = self.apply(.pressEndedUnseen)
            }
        }
    }

    /// The watcher is a `.background` of its anchor, so its bounds are the anchor's.
    private func isOnAnchor(_ event: NSEvent) -> Bool {
        guard let window, event.window === window else { return false }
        return bounds.contains(convert(event.locationInWindow, from: nil))
    }

    /// Applies the decision. True when the event is swallowed.
    private func apply(
        _ input: PVPopupInput,
        openingPressPending: Bool = false,
        onAnchor: Bool = false,
        pressBeganInside: Bool = false
    ) -> Bool {
        switch PVPopupDecision.decide(
            input, openingPressPending: openingPressPending,
            onAnchor: onAnchor, pressBeganInside: pressBeganInside, options: options
        ) {
        case .pass, .endOpeningPress:
            return false
        case .close(let consumesEvent):
            // After the event is delivered: a button that opened the popup
            // toggles on this same release and would reopen it.
            DispatchQueue.main.async { [weak self] in self?.close() }
            return consumesEvent
        case .toPopup(let key):
            return onKey(key) == .handled
        }
    }
}

extension View {
    /// The one presenter of a floating popup. Attach it to the popup's anchor.
    /// Escape, a left click outside (on release), a right-click outside (unless
    /// `options.ignoresRightClicks`) and an app switch close it; the press that
    /// opened it never does; a click inside never does; a click is outside only
    /// when its press began outside.
    /// Keys other than Escape go to `onKey`. See `PVPopupDecision.decide`.
    func pvPopup(
        isPresented: Binding<Bool>,
        options: PVPopupOptions = PVPopupOptions(),
        onKey: @escaping (KeyEquivalent) -> KeyPress.Result = { _ in .ignored }
    ) -> some View {
        background {
            PVPopupRepresentable(
                isPresented: isPresented.wrappedValue,
                options: options,
                onKey: onKey,
                close: { isPresented.wrappedValue = false }
            )
        }
    }
}
