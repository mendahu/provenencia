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
struct PVContextMenuItem: View {
    @Environment(\.pvContextMenuDismiss) private var dismiss

    private let titleResource: LocalizedStringResource?
    private let titleString: String?
    var isEnabled: Bool = true
    var isSelected: Bool = false
    var accessibilityIdentifier: String?
    var action: () -> Void

    init(
        _ title: LocalizedStringResource,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.titleResource = title
        self.titleString = nil
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    /// Dynamic catalog / runtime labels that are not String Catalog keys.
    init(
        plainTitle: String,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.titleResource = nil
        self.titleString = plainTitle
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    var body: some View {
        if isEnabled {
            Button {
                dismiss()
                action()
            } label: {
                label(foreground: PVColor.textPrimary)
            }
            .buttonStyle(PVContextMenuItemButtonStyle(isSelected: isSelected))
            .accessibilityAddIdentifiers(accessibilityIdentifier)
            .accessibilityAddTraits(isSelected ? .isSelected : [])
        } else {
            label(foreground: PVColor.textFaint)
                .accessibilityAddIdentifiers(accessibilityIdentifier)
                .accessibilityRemoveTraits(.isButton)
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
        @ViewBuilder content: @escaping () -> Content
    ) -> some View {
        modifier(PVContextMenuPresenter(state: state, content: content))
    }
}

// MARK: - Internals

private struct PVContextMenuDismissKey: EnvironmentKey {
    nonisolated(unsafe) static let defaultValue: () -> Void = {}
}

extension EnvironmentValues {
    var pvContextMenuDismiss: () -> Void {
        get { self[PVContextMenuDismissKey.self] }
        set { self[PVContextMenuDismissKey.self] = newValue }
    }
}

private struct PVContextMenuPresenter<MenuContent: View>: ViewModifier {
    @Binding var state: PVContextMenuState
    @ViewBuilder var content: () -> MenuContent
    @State private var dismissMonitor: Any?

    func body(content host: Content) -> some View {
        host
            .overlay(alignment: .topLeading) {
                if state.isPresented {
                    content()
                        .environment(\.pvContextMenuDismiss) {
                            state.dismiss()
                        }
                        .fixedSize()
                        .offset(x: state.origin.x, y: state.origin.y)
                }
            }
            .onChange(of: state.isPresented) { _, open in
                if open {
                    // Defer so the opening right-click does not immediately dismiss.
                    DispatchQueue.main.async { installDismissMonitor() }
                } else {
                    removeDismissMonitor()
                }
            }
            .onDisappear {
                state.dismiss()
                removeDismissMonitor()
            }
    }

    private func installDismissMonitor() {
        removeDismissMonitor()
        // Dismiss on mouse*Up* (not Down): Button actions fire on mouseUp, and
        // tearing the panel down on mouseDown prevents the item from running.
        dismissMonitor = NSEvent.addLocalMonitorForEvents(
            matching: [.leftMouseUp, .rightMouseDown, .otherMouseDown, .keyDown]
        ) { event in
            if event.type == .keyDown {
                if event.keyCode == 53 { // Escape
                    DispatchQueue.main.async { state.dismiss() }
                    return nil
                }
                return event
            }
            DispatchQueue.main.async { state.dismiss() }
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
/// consuming left-clicks.
private struct PVRightClickCatcher: NSViewRepresentable {
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

private final class PVRightClickCatcherView: NSView {
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
