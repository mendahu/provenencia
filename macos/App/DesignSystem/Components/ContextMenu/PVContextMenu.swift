import AppKit
import SwiftUI

// MARK: - Presentation

/// Open/origin state for a custom (non-`NSMenu`) context menu.
struct PVContextMenuState: Equatable {
    var isPresented = false
    var origin: CGPoint = .zero

    mutating func present(at origin: CGPoint) {
        self.origin = origin
        isPresented = true
    }

    mutating func dismiss() {
        isPresented = false
    }
}

/// Optional ↑/↓/⏎ navigation (and type-to-select) for an open overlay menu.
struct PVContextMenuKeyboard: Equatable {
    var itemCount: Int
    var activeIndex: Int
    /// Row titles for type-to-select. Empty disables letter jumping.
    var itemTitles: [String] = []
    var typeSelect = PVTypeSelectMatcher()

    static let inactive = PVContextMenuKeyboard(itemCount: 0, activeIndex: -1)

    /// Advances the highlight to the next title matching the type-select buffer.
    /// Returns `true` when the highlight moved.
    mutating func applyTypeSelect(_ character: Character, now: Date = Date()) -> Bool {
        guard !itemTitles.isEmpty else { return false }
        let buffer = typeSelect.append(character, now: now)
        let from = buffer.count == 1 ? activeIndex + 1 : max(0, activeIndex)
        let hit = PVTypeSelectMatcher.index(in: itemTitles, prefix: buffer, fromIndex: from)
        guard hit >= 0 else { return false }
        activeIndex = hit
        return true
    }
}

/// Screen-space frame for the child `NSPanel`.
///
/// `anchorOnScreen` is AppKit screen coordinates (y grows up). `origin` is the
/// SwiftUI top-leading offset inside the anchor (y grows down) — the same
/// local point `PVContextMenuState.present(at:)` stores.
enum PVContextMenuPlacement {
    static func panelFrame(
        anchorOnScreen: CGRect,
        origin: CGPoint,
        size: CGSize
    ) -> CGRect {
        let width = max(size.width, 1)
        let height = max(size.height, 1)
        let x = anchorOnScreen.minX + origin.x
        let top = anchorOnScreen.maxY - origin.y
        let y = top - height
        return CGRect(x: x, y: y, width: width, height: height)
    }
}

// MARK: - Panel + items

/// Floating menu chrome: card surface, subtle border, overlay shadow.
/// Optional caps section title matches Source-thumbnail / filter-style menus.
struct PVContextMenuPanel<Content: View>: View {
    var title: LocalizedStringResource?
    var width: CGFloat = 180
    var accessibilityIdentifier: String?
    @ViewBuilder var content: () -> Content

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let title {
                Text(title)
                    .font(PVFont.body(size: PVTypeScale.micro, weight: PVFontWeight.semibold))
                    .tracking(PVTypeScale.micro * PVTracking.caps)
                    .textCase(.uppercase)
                    .foregroundStyle(PVColor.textFaint)
                    .padding(.horizontal, PVSpacing.space5)
                    .padding(.top, PVSpacing.space3)
                    .padding(.bottom, PVSpacing.space2)
            }
            content()
        }
        .padding(PVSpacing.space2)
        .frame(width: width, alignment: .leading)
        .background(PVColor.surfaceCard)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
        .pvShadow(PVElevation.overlay)
        .accessibilityElement(children: .contain)
        .accessibilityAddIdentifiers(accessibilityIdentifier)
    }
}

/// A single menu row. Enabled items dismiss the hosting menu, then run `action`.
/// Pass `index` when the host provides keyboard navigation so ↑/↓ highlight works.
struct PVContextMenuItem: View {
    @Environment(\.pvContextMenuDismiss) private var dismiss
    @Environment(\.pvContextMenuKeyboardContext) private var keyboard

    private let titleResource: LocalizedStringResource?
    private let titleString: String?
    var index: Int?
    var isEnabled: Bool = true
    var isSelected: Bool = false
    var accessibilityIdentifier: String?
    var action: () -> Void

    init(
        _ title: LocalizedStringResource,
        index: Int? = nil,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.titleResource = title
        self.titleString = nil
        self.index = index
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    /// Dynamic catalog / runtime labels that are not String Catalog keys.
    init(
        plainTitle: String,
        index: Int? = nil,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.titleResource = nil
        self.titleString = plainTitle
        self.index = index
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    private var isKeyboardActive: Bool {
        guard let index, let keyboard else { return false }
        return keyboard.activeIndex == index
    }

    var body: some View {
        let highlighted = isSelected || isKeyboardActive
        // Register during body evaluation (not only onAppear) so Enter always
        // hits the live registry after open/arrow-key re-renders.
        if isEnabled, let index {
            keyboard?.register(index, action)
        }
        return Group {
            if isEnabled {
                Button {
                    dismiss()
                    action()
                } label: {
                    label(foreground: PVColor.textPrimary)
                }
                .buttonStyle(PVContextMenuItemButtonStyle(isSelected: highlighted))
                .accessibilityAddIdentifiers(accessibilityIdentifier)
                .accessibilityAddTraits(highlighted ? .isSelected : [])
            } else {
                label(foreground: PVColor.textFaint)
                    .accessibilityAddIdentifiers(accessibilityIdentifier)
                    .accessibilityRemoveTraits(.isButton)
            }
        }
    }

    private func label(foreground: Color) -> some View {
        Group {
            if let titleResource {
                Text(titleResource)
            } else if let titleString {
                Text(titleString)
            }
        }
        .font(PVFont.body(size: PVTypeScale.bodySmall))
        .foregroundStyle(foreground)
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space4)
        .contentShape(Rectangle())
    }
}

// MARK: - View modifiers

extension View {
    /// Right-click / Ctrl-click target that opens `state` at the click point
    /// (local to this view).
    func pvContextMenuTrigger(_ state: Binding<PVContextMenuState>) -> some View {
        overlay {
            PVRightClickCatcher { point in
                state.wrappedValue.present(at: point)
            }
        }
    }

    /// Positions and dismisses a custom menu in a child `NSPanel` at
    /// `.popUpMenu` level (same approach as `PVComboBox`), so the panel floats
    /// above every SwiftUI sibling — forms, meta cards, scroll content — without
    /// per-call-site `zIndex` workarounds.
    ///
    /// Attach this to the view whose local coordinates match `state.origin`
    /// (the trigger for `PVSelect`, or a larger ancestor for right-click menus
    /// so the click point and panel share a coordinate space).
    func pvContextMenu<Content: View>(
        _ state: Binding<PVContextMenuState>,
        keyboard: Binding<PVContextMenuKeyboard>? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) -> some View {
        modifier(PVContextMenuPresenter(state: state, keyboard: keyboard, content: content))
    }
}

// MARK: - Internals

private struct PVContextMenuDismissKey: EnvironmentKey {
    nonisolated(unsafe) static let defaultValue: () -> Void = {}
}

private struct PVContextMenuKeyboardContextKey: EnvironmentKey {
    nonisolated(unsafe) static let defaultValue: PVContextMenuKeyboardContext? = nil
}

/// Registers row actions for Enter activation. Highlight uses `activeIndex` (value).
final class PVContextMenuActionRegistry {
    private var actions: [Int: () -> Void] = [:]

    func register(_ index: Int, _ action: @escaping () -> Void) {
        actions[index] = action
    }

    func activate(_ index: Int) {
        actions[index]?()
    }

    func clear() {
        actions.removeAll(keepingCapacity: true)
    }
}

private struct PVContextMenuKeyboardContext {
    var activeIndex: Int
    var registry: PVContextMenuActionRegistry

    func register(_ index: Int, _ action: @escaping () -> Void) {
        registry.register(index, action)
    }
}

extension EnvironmentValues {
    var pvContextMenuDismiss: () -> Void {
        get { self[PVContextMenuDismissKey.self] }
        set { self[PVContextMenuDismissKey.self] = newValue }
    }

    fileprivate var pvContextMenuKeyboardContext: PVContextMenuKeyboardContext? {
        get { self[PVContextMenuKeyboardContextKey.self] }
        set { self[PVContextMenuKeyboardContextKey.self] = newValue }
    }
}

private struct PVContextMenuPresenter<MenuContent: View>: ViewModifier {
    @Binding var state: PVContextMenuState
    var keyboard: Binding<PVContextMenuKeyboard>?
    @ViewBuilder var content: () -> MenuContent
    @State private var dismissMonitor: Any?
    @State private var actionRegistry = PVContextMenuActionRegistry()

    func body(content host: Content) -> some View {
        host
            .background(
                PVContextMenuPopupWindow(
                    isPresented: state.isPresented,
                    origin: state.origin,
                    onDismiss: { state.dismiss() },
                    content: menuRoot
                )
            )
            .onChange(of: state.isPresented) { _, open in
                if open {
                    // Defer so the opening click does not immediately dismiss.
                    // Do not replace/clear the registry here — items register
                    // during body evaluation; clearing after first paint left
                    // Enter calling an empty map.
                    DispatchQueue.main.async { installDismissMonitor() }
                } else {
                    removeDismissMonitor()
                    actionRegistry.clear()
                }
            }
            .onDisappear {
                state.dismiss()
                removeDismissMonitor()
            }
    }

    private var menuRoot: some View {
        content()
            .environment(\.pvContextMenuDismiss) {
                state.dismiss()
            }
            .environment(\.pvContextMenuKeyboardContext, keyboardContext)
            .fixedSize()
    }

    private var keyboardContext: PVContextMenuKeyboardContext? {
        guard let keyboard else { return nil }
        return PVContextMenuKeyboardContext(
            activeIndex: keyboard.wrappedValue.activeIndex,
            registry: actionRegistry
        )
    }

    private func installDismissMonitor() {
        removeDismissMonitor()
        // Dismiss on mouse*Up* (not Down): Button actions fire on mouseUp, and
        // tearing the panel down on mouseDown prevents the item from running.
        dismissMonitor = NSEvent.addLocalMonitorForEvents(
            matching: [.leftMouseUp, .rightMouseDown, .otherMouseDown, .keyDown]
        ) { event in
            if event.type == .keyDown {
                return handleKeyDown(event)
            }
            DispatchQueue.main.async { state.dismiss() }
            return event
        }
    }

    private func handleKeyDown(_ event: NSEvent) -> NSEvent? {
        if event.keyCode == 53 { // Escape
            DispatchQueue.main.async { state.dismiss() }
            return nil
        }
        guard var nav = keyboard?.wrappedValue, nav.itemCount > 0 else {
            return event
        }
        switch event.keyCode {
        case 125: // Down
            DispatchQueue.main.async {
                nav.activeIndex = PVFloatingMenuSelection.moveIndex(
                    from: nav.activeIndex, delta: 1, count: nav.itemCount
                )
                keyboard?.wrappedValue = nav
            }
            return nil
        case 126: // Up
            DispatchQueue.main.async {
                nav.activeIndex = PVFloatingMenuSelection.moveIndex(
                    from: nav.activeIndex, delta: -1, count: nav.itemCount
                )
                keyboard?.wrappedValue = nav
            }
            return nil
        case 36, 76: // Return / keypad Enter
            let index = nav.activeIndex
            let registry = actionRegistry
            DispatchQueue.main.async {
                guard index >= 0 else { return }
                state.dismiss()
                registry.activate(index)
            }
            return nil
        default:
            guard let characters = event.charactersIgnoringModifiers,
                  characters.count == 1,
                  let character = characters.first,
                  character.isLetter || character.isNumber,
                  !event.modifierFlags.contains(.command),
                  !event.modifierFlags.contains(.control),
                  !nav.itemTitles.isEmpty
            else { return event }
            DispatchQueue.main.async {
                guard var next = keyboard?.wrappedValue else { return }
                _ = next.applyTypeSelect(character)
                keyboard?.wrappedValue = next
            }
            return nil
        }
    }

    private func removeDismissMonitor() {
        if let dismissMonitor {
            NSEvent.removeMonitor(dismissMonitor)
            self.dismissMonitor = nil
        }
    }
}

// MARK: - Floating NSPanel host

/// Child-window popup at `.popUpMenu` level. Mirrors `PVComboBoxPopupWindow`:
/// SwiftUI stacking cannot bury a separate AppKit window under later siblings.
private struct PVContextMenuPopupWindow<Content: View>: NSViewRepresentable {
    var isPresented: Bool
    var origin: CGPoint
    var onDismiss: () -> Void
    var content: Content

    func makeNSView(context: Context) -> NSView {
        let view = NSView(frame: .zero)
        context.coordinator.anchor = view
        return view
    }

    func updateNSView(_: NSView, context: Context) {
        let coordinator = context.coordinator
        coordinator.onDismiss = onDismiss
        coordinator.origin = origin
        coordinator.wantsPresented = isPresented
        coordinator.setContent(content)
        DispatchQueue.main.async {
            coordinator.applyPresentation()
        }
    }

    static func dismantleNSView(_: NSView, coordinator: Coordinator) {
        coordinator.tearDown()
    }

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    @MainActor
    final class Coordinator {
        weak var anchor: NSView?
        var onDismiss: () -> Void = {}
        var origin: CGPoint = .zero
        var wantsPresented = false

        private var panel: PVContextMenuNSPanel?
        private var hosting: PVContextMenuHostingView?
        private var content: AnyView = AnyView(EmptyView())
        private var observers: [NSObjectProtocol] = []
        private var shownSize: CGSize?

        func setContent(_ content: some View) {
            self.content = AnyView(content)
            if let hosting, let panel, panel.parent != nil {
                hosting.rootView = self.content
            }
        }

        func applyPresentation() {
            guard wantsPresented else {
                hide()
                return
            }
            if panel?.parent == nil {
                show()
            } else {
                position(reframeIfSizeChanged: true)
            }
        }

        private func show() {
            guard let anchor, let parent = anchor.window else { return }
            if hosting == nil {
                hosting = PVContextMenuHostingView(rootView: content)
            }
            guard let hosting else { return }
            if panel == nil {
                let created = PVContextMenuNSPanel(
                    contentRect: NSRect(x: 0, y: 0, width: 100, height: 100),
                    styleMask: [.borderless, .nonactivatingPanel],
                    backing: .buffered,
                    defer: false
                )
                created.isOpaque = false
                created.backgroundColor = .clear
                // Menu chrome already draws `pvShadow`; avoid a second AppKit shadow.
                created.hasShadow = false
                created.level = .popUpMenu
                created.animationBehavior = .none
                created.contentView = hosting
                panel = created
            }
            guard let panel, panel.parent == nil else { return }
            position(reframeIfSizeChanged: true)
            parent.addChildWindow(panel, ordered: .above)
            startWatching(parent: parent)
        }

        private func hide() {
            shownSize = nil
            guard let panel, panel.parent != nil else { return }
            stopWatching()
            panel.parent?.removeChildWindow(panel)
            panel.orderOut(nil)
        }

        func tearDown() {
            stopWatching()
            if let panel {
                panel.parent?.removeChildWindow(panel)
                panel.orderOut(nil)
            }
            panel = nil
            hosting = nil
        }

        private func position(reframeIfSizeChanged: Bool) {
            guard let panel, let anchor, let window = anchor.window, let hosting else { return }
            hosting.rootView = content
            let fitted = hosting.fittingSize
            let measured = CGSize(width: max(fitted.width, 1), height: max(fitted.height, 1))
            if shownSize == nil || (reframeIfSizeChanged && shownSize != measured) {
                shownSize = measured
            }
            let frameSize = shownSize ?? measured

            let anchorOnScreen = window.convertToScreen(anchor.convert(anchor.bounds, to: nil))
            let next = PVContextMenuPlacement.panelFrame(
                anchorOnScreen: anchorOnScreen,
                origin: origin,
                size: frameSize
            )
            if panel.frame != next {
                panel.setFrame(next, display: true)
            }
        }

        private func startWatching(parent: NSWindow) {
            stopWatching()
            let center = NotificationCenter.default
            for name in [NSWindow.didMoveNotification, NSWindow.didResizeNotification] {
                observers.append(center.addObserver(forName: name, object: parent, queue: .main) { [weak self] _ in
                    MainActor.assumeIsolated { self?.position(reframeIfSizeChanged: false) }
                })
            }
            observers.append(
                center.addObserver(forName: NSWindow.didResignKeyNotification, object: parent, queue: .main) { [weak self] _ in
                    MainActor.assumeIsolated { self?.dismissFromAppKit() }
                }
            )
            if let clip = anchor?.enclosingScrollView?.contentView {
                clip.postsBoundsChangedNotifications = true
                observers.append(
                    center.addObserver(forName: NSView.boundsDidChangeNotification, object: clip, queue: .main) { [weak self] _ in
                        MainActor.assumeIsolated { self?.position(reframeIfSizeChanged: false) }
                    }
                )
            }
        }

        private func stopWatching() {
            observers.forEach(NotificationCenter.default.removeObserver)
            observers.removeAll()
        }

        private func dismissFromAppKit() {
            wantsPresented = false
            hide()
            onDismiss()
        }
    }
}

private final class PVContextMenuNSPanel: NSPanel {
    override var canBecomeKey: Bool { false }
    override var canBecomeMain: Bool { false }
}

private final class PVContextMenuHostingView: NSHostingView<AnyView> {
    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }

    @available(*, unavailable)
    required init(coder: NSCoder) {
        fatalError("init(coder:) is not used — the popup is built in code")
    }

    @MainActor
    required init(rootView: AnyView) {
        super.init(rootView: rootView)
    }
}

private struct PVContextMenuItemButtonStyle: ButtonStyle {
    var isSelected: Bool = false

    func makeBody(configuration: Configuration) -> some View {
        PVContextMenuItemButtonBody(configuration: configuration, isSelected: isSelected)
    }
}

private struct PVContextMenuItemButtonBody: View {
    let configuration: ButtonStyleConfiguration
    var isSelected: Bool
    @State private var hovering = false

    var body: some View {
        configuration.label
            .background {
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(rowFill)
            }
            .contentShape(Rectangle())
            .onHover { hovering = $0 }
            .pvAnimation(PVMotion.instantStandard, value: hovering)
    }

    private var rowFill: Color {
        if configuration.isPressed { return PVColor.surfaceActive }
        if isSelected { return PVColor.surfaceSelected }
        if hovering { return PVColor.surfaceHover }
        return .clear
    }
}

/// Transparent hit target that reports right-clicks (and Ctrl-click) without
/// consuming left-clicks. Shared by context-menu triggers and history jump.
struct PVRightClickCatcher: NSViewRepresentable {
    var onRightClick: (CGPoint) -> Void

    func makeNSView(context: Context) -> PVRightClickCatcherView {
        let view = PVRightClickCatcherView()
        view.onRightClick = onRightClick
        return view
    }

    func updateNSView(_ nsView: PVRightClickCatcherView, context: Context) {
        nsView.onRightClick = onRightClick
    }
}

final class PVRightClickCatcherView: NSView {
    var onRightClick: ((CGPoint) -> Void)?

    override var acceptsFirstResponder: Bool { false }

    override func hitTest(_ point: NSPoint) -> NSView? {
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
        onRightClick?(convert(event.locationInWindow, from: nil))
    }
}

private extension View {
    @ViewBuilder
    func accessibilityAddIdentifiers(_ id: String?) -> some View {
        if let id {
            accessibilityIdentifier(id)
        } else {
            self
        }
    }
}
