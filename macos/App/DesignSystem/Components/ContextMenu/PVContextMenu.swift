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
    /// Rows that cannot run. The highlight never lands on one. Empty means
    /// every row is enabled.
    var disabledRows: IndexSet = []
    var typeSelect = PVTypeSelectMatcher()

    static let inactive = PVContextMenuKeyboard(itemCount: 0, activeIndex: -1)

    /// The state of a menu that just opened: nothing highlighted, type-to-select over `titles`.
    static func opened(titles: [String], disabledRows: IndexSet = []) -> PVContextMenuKeyboard {
        PVContextMenuKeyboard(
            itemCount: titles.count, activeIndex: -1, itemTitles: titles, disabledRows: disabledRows
        )
    }

    /// Advances the highlight to the next enabled title matching the type-select buffer.
    /// Returns `true` when the highlight moved.
    @discardableResult
    mutating func applyTypeSelect(_ character: Character, now: Date = Date()) -> Bool {
        guard !itemTitles.isEmpty else { return false }
        let hit = typeSelect.advance(
            character, from: activeIndex, in: itemTitles,
            isDisabled: { disabledRows.contains($0) }, now: now
        )
        guard hit >= 0 else { return false }
        activeIndex = hit
        return true
    }

    /// What one key does to the open menu. `.move(to: activeIndex)` means the
    /// key is consumed and nothing changed (the menu stays open); `.activate`
    /// runs the highlighted row; `.ignore` is not taken.
    mutating func handle(_ key: KeyEquivalent, now: Date = Date()) -> PVListKeyAction {
        guard itemCount > 0 else { return .ignore }
        switch key {
        case .upArrow, .downArrow, .home, .end:
            guard let next = PVListKeyboard.moveIndex(
                for: key, option: false, index: activeIndex, count: itemCount,
                isDisabled: { disabledRows.contains($0) }
            ) else { return .ignore }
            activeIndex = next
            return .move(to: next)
        case .return:
            // Return on nothing, or on a disabled row, keeps the menu open and
            // lets no key through to a default button behind it.
            guard (0..<itemCount).contains(activeIndex), !disabledRows.contains(activeIndex) else {
                return .move(to: activeIndex)
            }
            return .activate(activeIndex)
        default:
            guard !itemTitles.isEmpty, PVTypeSelectMatcher.isTypeSelectCharacter(key.character) else {
                return .ignore
            }
            applyTypeSelect(key.character, now: now)
            return .move(to: activeIndex)
        }
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
/// A title is a localized key (`init(title:…)`) or data such as a ref, drawn as
/// given (`init(verbatimTitle:…)`); never both.
struct PVContextMenuPanel<Content: View>: View {
    private let heading: Text?
    var width: CGFloat
    var maxHeight: CGFloat?
    var accessibilityIdentifier: String?
    @ViewBuilder var content: () -> Content

    init(
        title: LocalizedStringResource? = nil,
        width: CGFloat = 180,
        maxHeight: CGFloat? = nil,
        accessibilityIdentifier: String? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.init(
            heading: title.map { Text($0) }, width: width, maxHeight: maxHeight,
            accessibilityIdentifier: accessibilityIdentifier, content: content
        )
    }

    init(
        verbatimTitle: String,
        width: CGFloat = 180,
        maxHeight: CGFloat? = nil,
        accessibilityIdentifier: String? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.init(
            heading: Text(verbatim: verbatimTitle), width: width, maxHeight: maxHeight,
            accessibilityIdentifier: accessibilityIdentifier, content: content
        )
    }

    /// For a host that already built its heading (`pvRightClickMenu`).
    init(
        heading: Text?,
        width: CGFloat = 180,
        maxHeight: CGFloat? = nil,
        accessibilityIdentifier: String? = nil,
        @ViewBuilder content: @escaping () -> Content
    ) {
        self.heading = heading
        self.width = width
        self.maxHeight = maxHeight
        self.accessibilityIdentifier = accessibilityIdentifier
        self.content = content
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if let heading {
                heading
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
        .frame(maxHeight: maxHeight, alignment: .top)
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

    private let labelContent: AnyView
    var index: Int?
    var isEnabled: Bool = true
    var isSelected: Bool = false
    /// When true, a leading checkmark gutter marks the committed row. Fill
    /// follows keyboard/hover highlight only — the two must not look equally
    /// selected (`PVSelect`).
    var showsSelectionMark: Bool = false
    var accessibilityIdentifier: String?
    var action: () -> Void

    init(
        _ title: LocalizedStringResource,
        index: Int? = nil,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        showsSelectionMark: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.labelContent = AnyView(
            Text(title)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
        )
        self.index = index
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.showsSelectionMark = showsSelectionMark
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    /// Dynamic catalog / runtime labels that are not String Catalog keys.
    init(
        plainTitle: String,
        index: Int? = nil,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        showsSelectionMark: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {}
    ) {
        self.labelContent = AnyView(
            Text(verbatim: plainTitle)
                .font(PVFont.body(size: PVTypeScale.bodySmall))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
        )
        self.index = index
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.showsSelectionMark = showsSelectionMark
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    /// Custom row body. The item still owns highlight, checkmark, and dismiss.
    init(
        index: Int? = nil,
        isEnabled: Bool = true,
        isSelected: Bool = false,
        showsSelectionMark: Bool = false,
        accessibilityIdentifier: String? = nil,
        action: @escaping () -> Void = {},
        @ViewBuilder label: () -> some View
    ) {
        self.labelContent = AnyView(label())
        self.index = index
        self.isEnabled = isEnabled
        self.isSelected = isSelected
        self.showsSelectionMark = showsSelectionMark
        self.accessibilityIdentifier = accessibilityIdentifier
        self.action = action
    }

    private var isKeyboardActive: Bool {
        guard let index, let keyboard else { return false }
        return keyboard.activeIndex == index
    }

    var body: some View {
        let chrome = PVSelectRowChrome.appearance(
            committed: isSelected,
            highlighted: isKeyboardActive
        )
        let fillHighlight = showsSelectionMark ? chrome.fillsHighlight : (isSelected || isKeyboardActive)
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
                    label(showsCheck: showsSelectionMark && chrome.showsCheckmark)
                }
                .buttonStyle(PVMenuRowButtonStyle(isSelected: fillHighlight))
                .accessibilityAddIdentifiers(accessibilityIdentifier)
                .accessibilityAddTraits(chrome.isSelectedTrait ? .isSelected : [])
            } else {
                label(showsCheck: showsSelectionMark && chrome.showsCheckmark)
                    .opacity(0.45)
                    .accessibilityAddIdentifiers(accessibilityIdentifier)
                    .accessibilityRemoveTraits(.isButton)
            }
        }
    }

    private func label(showsCheck: Bool) -> some View {
        HStack(alignment: .center, spacing: PVSpacing.space3) {
            if showsSelectionMark {
                PVIcon(.check, size: 13)
                    .foregroundStyle(PVColor.accent)
                    .opacity(showsCheck ? 1 : 0)
                    .frame(width: 22)
            }
            labelContent
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(.horizontal, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space3)
        .contentShape(Rectangle())
    }
}

// MARK: - View modifiers

extension View {
    /// Right-click / Ctrl-click target that opens `state` at the click point
    /// (local to this view).
    /// Feature code uses `.pvRightClickMenu`.
    func pvContextMenuTrigger(_ state: Binding<PVContextMenuState>) -> some View {
        overlay {
            PVRightClickCatcher { point in
                state.wrappedValue.present(at: point)
            }
        }
    }

    /// Positions a custom menu in a child `NSPanel` at `.popUpMenu` level (same
    /// approach as `PVComboBox`), so the panel floats above every SwiftUI
    /// sibling — forms, meta cards, scroll content — without per-call-site
    /// `zIndex` workarounds. It closes through `.pvPopup`, like every kit popup.
    ///
    /// Attach this to the view whose local coordinates match `state.origin`: the
    /// trigger view. Action menus use `.pvActionMenu` (`.pvRightClickMenu` for a
    /// right-click), which does this.
    ///
    /// `keyboard` feeds the rows' highlight and, unless `options.keysStayInField`,
    /// the menu's keys (↑/↓, Home/End, type-select, Return). A `PVSelect` keeps
    /// its keys in its focused field, so it passes `keysStayInField`.
    func pvContextMenu<Content: View>(
        _ state: Binding<PVContextMenuState>,
        keyboard: Binding<PVContextMenuKeyboard>? = nil,
        options: PVPopupOptions = PVPopupOptions(),
        @ViewBuilder content: @escaping () -> Content
    ) -> some View {
        modifier(PVContextMenuPresenter(
            state: state,
            keyboard: keyboard,
            options: options,
            content: content
        ))
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
    var options = PVPopupOptions()
    @ViewBuilder var content: () -> MenuContent
    @State private var actionRegistry = PVContextMenuActionRegistry()

    func body(content host: Content) -> some View {
        host
            .background(
                PVContextMenuPopupWindow(
                    isPresented: state.isPresented,
                    origin: state.origin,
                    content: menuRoot
                )
            )
            .pvPopup(
                isPresented: $state.isPresented,
                options: options,
                onKey: handleKey
            )
            .onChange(of: state.isPresented) { _, open in
                // Do not replace/clear the registry on open — items register
                // during body evaluation; clearing after first paint left
                // Enter calling an empty map.
                if !open { actionRegistry.clear() }
            }
            .onDisappear {
                state.dismiss()
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

    /// Applies `PVContextMenuKeyboard.handle`. The highlight is written back,
    /// and a Return runs its row, on the next main-queue turn, after the event
    /// is delivered.
    private func handleKey(_ key: KeyEquivalent) -> KeyPress.Result {
        guard let keyboard else { return .ignored }
        var next = keyboard.wrappedValue
        let action = next.handle(key)
        let registry = actionRegistry
        switch action {
        case .ignore:
            return .ignored
        case .move:
            DispatchQueue.main.async { keyboard.wrappedValue = next }
            return .handled
        case .activate(let index):
            DispatchQueue.main.async {
                keyboard.wrappedValue = next
                state.dismiss()
                registry.activate(index)
            }
            return .handled
        }
    }
}

// MARK: - Floating NSPanel host

/// Child-window popup at `.popUpMenu` level, shown through `PVPopupWindowHost`:
/// SwiftUI stacking cannot bury a separate AppKit window under later siblings.
private struct PVContextMenuPopupWindow<Content: View>: NSViewRepresentable {
    var isPresented: Bool
    var origin: CGPoint
    var content: Content

    func makeNSView(context: Context) -> NSView {
        let view = NSView(frame: .zero)
        context.coordinator.anchor = view
        return view
    }

    func updateNSView(_: NSView, context: Context) {
        let coordinator = context.coordinator
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
        weak var anchor: NSView? {
            didSet { host.anchor = anchor }
        }
        var origin: CGPoint = .zero
        var wantsPresented = false

        // Menu chrome already draws `pvShadow`; avoid a second AppKit shadow.
        private let host = PVPopupWindowHost(hasShadow: false)
        private var content: AnyView = AnyView(EmptyView())
        private var shownSize: CGSize?

        init() {
            host.onFollow = { [weak self] in self?.position(reframeIfSizeChanged: false) }
        }

        func setContent(_ content: some View) {
            self.content = AnyView(content)
            if host.isShown {
                host.rootView = self.content
            }
        }

        func applyPresentation() {
            guard wantsPresented else {
                hide()
                return
            }
            if !host.isShown {
                show()
            } else {
                position(reframeIfSizeChanged: true)
            }
        }

        private func show() {
            guard let frame = frame(reframeIfSizeChanged: true) else { return }
            host.show(at: frame)
        }

        private func hide() {
            shownSize = nil
            host.hide()
        }

        func tearDown() {
            host.tearDown()
        }

        /// The panel's screen frame from the content's fitted size. Nil without an anchor window.
        private func frame(reframeIfSizeChanged: Bool) -> CGRect? {
            guard let anchor, let window = anchor.window else { return nil }
            host.rootView = content
            let fitted = host.fittingSize
            let measured = CGSize(width: max(fitted.width, 1), height: max(fitted.height, 1))
            if shownSize == nil || (reframeIfSizeChanged && shownSize != measured) {
                shownSize = measured
            }
            let frameSize = shownSize ?? measured

            let anchorOnScreen = window.convertToScreen(anchor.convert(anchor.bounds, to: nil))
            return PVContextMenuPlacement.panelFrame(
                anchorOnScreen: anchorOnScreen,
                origin: origin,
                size: frameSize
            )
        }

        private func position(reframeIfSizeChanged: Bool) {
            guard host.isShown, let frame = frame(reframeIfSizeChanged: reframeIfSizeChanged) else { return }
            host.setFrame(frame)
        }
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
