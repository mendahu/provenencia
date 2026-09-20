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

/// Optional ↑/↓/⏎ navigation for an open overlay menu.
struct PVContextMenuKeyboard: Equatable {
    var itemCount: Int
    var activeIndex: Int

    static let inactive = PVContextMenuKeyboard(itemCount: 0, activeIndex: -1)
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

    /// Positions and dismisses a custom menu.
    ///
    /// Attach this to an ancestor large enough that the panel stays inside the
    /// hit-testing bounds — overlaying only a small trigger (e.g. a 72pt
    /// thumbnail) clips clicks on overflow rows.
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
            .overlay(alignment: .topLeading) {
                if state.isPresented {
                    content()
                        .environment(\.pvContextMenuDismiss) {
                            state.dismiss()
                        }
                        .environment(\.pvContextMenuKeyboardContext, keyboardContext)
                        .fixedSize()
                        .offset(x: state.origin.x, y: state.origin.y)
                }
            }
            // Menu is an overlay on this host. Later siblings in a VStack/ZStack
            // paint above us by default (e.g. onboarding project meta under a
            // PVSelect) — lift the open host so the panel wins hit-testing too.
            .zIndex(state.isPresented ? 1 : 0)
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
            return event
        }
    }

    private func removeDismissMonitor() {
        if let dismissMonitor {
            NSEvent.removeMonitor(dismissMonitor)
            self.dismissMonitor = nil
        }
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
