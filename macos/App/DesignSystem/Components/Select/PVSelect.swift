import AppKit
import SwiftUI

/// One option in a `PVSelect` — `value` is the stable identifier (what a
/// binding stores), `label` is what's shown. Plain `String`, not
/// `LocalizedStringResource`: options are almost always data (a project
/// filename, a person's name), not static UI copy — see
/// `docs/macos-client-patterns.md` §6.
struct PVSelectOption: Identifiable, Equatable {
    let id: String
    let label: String
    var accessibilityIdentifier: String?

    init(value: String, label: String, accessibilityIdentifier: String? = nil) {
        self.id = value
        self.label = label
        self.accessibilityIdentifier = accessibilityIdentifier
    }
}

/// A styled dropdown select built on the floating `PVContextMenu` kit.
/// Interaction matches a native macOS popup button (`NSPopUpButton`): custom
/// Frost chrome, native closed/open contract. The view forwards keys and
/// pointer hit targets into `PVSelectSession`.
///
/// - **Field** (default): full-width control matching `PVInput` rest border.
/// - **Chip** (`icon:` set, `fillsWidth: false`): compact toolbar select.
/// - **Icon-only chip** (`icon` + `iconOnly`): table column filter; pass
///   `accessibilityLabel`.
struct PVSelect: View {
    @Binding private var selection: String
    private let options: [PVSelectOption]
    private let size: PVControlSize
    private let icon: PVSymbol?
    private let iconOnly: Bool
    private let displayLabel: String?
    private let placeholder: String?
    private let menuWidth: CGFloat
    private let fillsWidth: Bool
    private let maxVisibleRows: Int
    private let isDisabled: Bool
    private let accessibilityLabelResource: LocalizedStringResource?
    private let accessibilitySpokenLabel: String?
    private let accessibilityIdentifier: String?

    @State private var session: PVSelectSession
    @State private var menuState = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive
    @State private var triggerScreen = CGRect.zero
    @State private var visibleScreen = CGRect.zero
    @State private var menuScreen = CGRect.zero
    @State private var isTracking = false
    @FocusState private var isFocused: Bool

    init(
        selection: Binding<String>,
        options: [PVSelectOption],
        size: PVControlSize = .md,
        icon: PVSymbol? = nil,
        iconOnly: Bool = false,
        displayLabel: String? = nil,
        placeholder: String? = nil,
        menuWidth: CGFloat = PVSelectPlacement.defaultMenuWidth,
        fillsWidth: Bool = true,
        maxVisibleRows: Int = PVSelectPlacement.defaultMaxVisibleRows,
        isDisabled: Bool = false,
        accessibilityLabel: LocalizedStringResource? = nil,
        accessibilitySpokenLabel: String? = nil,
        accessibilityIdentifier: String? = nil
    ) {
        self._selection = selection
        self.options = options
        self.size = size
        self.icon = icon
        self.iconOnly = iconOnly
        self.displayLabel = displayLabel
        self.placeholder = placeholder
        self.menuWidth = menuWidth
        self.fillsWidth = fillsWidth
        self.maxVisibleRows = maxVisibleRows
        self.isDisabled = isDisabled
        self.accessibilityLabelResource = accessibilityLabel
        self.accessibilitySpokenLabel = accessibilitySpokenLabel
        self.accessibilityIdentifier = accessibilityIdentifier
        self._session = State(initialValue: PVSelectSession(
            options: options,
            selection: selection.wrappedValue,
            isDisabled: isDisabled
        ))
    }

    private var spoken: PVSelectAccessibility.Spoken {
        PVSelectAccessibility.spoken(
            session: session,
            placeholder: placeholder,
            displayLabel: displayLabel
        )
    }

    private var triggerText: String {
        PVSelectAccessibility.triggerLabel(
            session: session,
            displayLabel: displayLabel,
            placeholder: placeholder
        )
    }

    private var showsPlaceholder: Bool {
        displayLabel == nil && session.committedLabel == nil
    }

    var body: some View {
        trigger
            .opacity(isDisabled ? 0.45 : 1)
            .focusable(!isDisabled)
            .focused($isFocused)
            .disabled(isDisabled)
            .onMoveCommand(perform: handleMoveCommand)
            .onKeyPress(action: handleTriggerKey)
            .background(anchorReader)
            .overlay(pressTracker)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(accessibilityLabelText)
            .accessibilityValue(Text(verbatim: spoken.value))
            .accessibilityAddIdentifiers(accessibilityIdentifier)
            .accessibilityAdjustableAction { direction in
                switch direction {
                case .increment: applySession { $0.increment() }
                case .decrement: applySession { $0.decrement() }
                default: break
                }
            }
            .accessibilityCustomContent(
                Text(L10n.DesignSystem.selectState),
                Text(spoken.isExpanded
                     ? L10n.DesignSystem.selectExpanded
                     : L10n.DesignSystem.selectCollapsed)
            )
            .modifier(PVSelectPositionContent(position: spoken.position))
            .pvContextMenu(
                $menuState,
                keyboard: options.isEmpty ? nil : $keyboard,
                stealKeys: false,
                dismissOnMouseUp: false
            ) {
                menu
            }
            .onChange(of: selection) { _, newValue in
                if session.selection != newValue {
                    session.selection = newValue
                }
            }
            .onChange(of: options.map(\.id)) { _, _ in
                session.syncOptions(options, selection: selection, isDisabled: isDisabled)
            }
            .onChange(of: isDisabled) { _, newValue in
                session.syncOptions(options, selection: selection, isDisabled: newValue)
            }
            .onChange(of: session.selection) { _, newValue in
                if selection != newValue {
                    selection = newValue
                }
            }
            .onChange(of: session.isOpen) { _, open in
                if open {
                    isFocused = true
                    presentMenu()
                } else {
                    menuState.dismiss()
                    keyboard = .inactive
                }
            }
            .onChange(of: session.highlightIndex) { _, index in
                keyboard.activeIndex = index
            }
            .onChange(of: menuState.isPresented) { _, presented in
                if !presented, session.isOpen {
                    applySession { $0.dismissRestoring() }
                }
            }
    }

    private var menu: some View {
        let height = max(menuScreen.height, 1)
        return PVContextMenuPanel(
            width: max(menuScreen.width, menuWidth),
            maxHeight: height,
            accessibilityIdentifier: PVSelectAccessibility.menuIdentifier(accessibilityIdentifier)
        ) {
            ScrollViewReader { proxy in
                ScrollView {
                    VStack(alignment: .leading, spacing: 0) {
                        ForEach(Array(options.enumerated()), id: \.element.id) { offset, option in
                            PVContextMenuItem(
                                plainTitle: option.label,
                                index: offset,
                                isSelected: option.id == session.selection,
                                showsSelectionMark: true,
                                accessibilityIdentifier: option.accessibilityIdentifier
                            ) {
                                applySession { $0.commit(index: offset, close: true) }
                            }
                            .id(offset)
                        }
                    }
                }
                .frame(maxHeight: height)
                .onChange(of: session.highlightIndex) { _, index in
                    guard index >= 0 else { return }
                    proxy.scrollTo(index, anchor: .center)
                }
            }
        }
    }

    private func presentMenu() {
        let contentHeight = PVSelectPlacement.contentHeight(
            optionCount: options.count,
            maxVisibleRows: maxVisibleRows
        )
        let frame = PVSelectPlacement.popupFrame(
            anchor: triggerScreen,
            contentHeight: contentHeight,
            visibleFrame: visibleScreen,
            menuWidth: menuWidth
        )
        menuScreen = frame
        keyboard = PVContextMenuKeyboard(
            itemCount: options.count,
            activeIndex: session.highlightIndex,
            itemTitles: options.map(\.label)
        )
        menuState.present(at: PVSelectPlacement.menuOrigin(anchor: triggerScreen, frame: frame))
    }

    private func applySession(_ body: (inout PVSelectSession) -> Void) {
        var next = session
        body(&next)
        session = next
    }

    private func handleMoveCommand(_ direction: MoveCommandDirection) {
        switch direction {
        case .down:
            applySession { $0.handleKey(.down) }
        case .up:
            applySession { $0.handleKey(.up) }
        default:
            break
        }
    }

    private func handleTriggerKey(_ press: KeyPress) -> KeyPress.Result {
        guard !isDisabled else { return .ignored }
        if press.modifiers.contains(.command) || press.modifiers.contains(.control) {
            return .ignored
        }
        let key: PVSelectKey?
        switch press.key {
        case .upArrow, .downArrow:
            // `onMoveCommand` already applied ↑/↓.
            return .ignored
        case .home: key = .home
        case .end: key = .end
        case .escape: key = .escape
        case .space: key = .space
        case .return: key = .return
        default:
            guard press.characters.count == 1,
                  let character = press.characters.first,
                  character.isLetter || character.isNumber
            else { return .ignored }
            key = .character(character)
        }
        guard let key else { return .ignored }
        applySession { $0.handleKey(key) }
        return .handled
    }

    private var accessibilityLabelText: Text {
        if let accessibilityLabelResource {
            Text(accessibilityLabelResource)
        } else if let accessibilitySpokenLabel {
            Text(verbatim: accessibilitySpokenLabel)
        } else {
            Text(verbatim: triggerText)
        }
    }

    private var triggerHeight: CGFloat {
        icon == nil ? size.height : PVSpacing.controlHeightMedium
    }

    @ViewBuilder
    private var trigger: some View {
        if icon != nil {
            chipTrigger
        } else {
            fieldTrigger
        }
    }

    private var fieldTrigger: some View {
        HStack(spacing: PVSpacing.space3) {
            triggerLabel
            Spacer(minLength: 0)
            PVIcon(.chevronDown, size: 14)
                .foregroundStyle(PVColor.textFaint)
        }
        .font(size.font)
        .padding(.horizontal, PVSpacing.space3 + PVSpacing.space1)
        .frame(height: size.height)
        .frame(maxWidth: fillsWidth ? .infinity : nil)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(isFocused ? PVColor.borderFocus : PVColor.borderDefault, lineWidth: 1)
        )
        .pvInsetShadow(cornerRadius: PVRadius.sm, visible: !isFocused)
        .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
    }

    private var chipTrigger: some View {
        HStack(spacing: iconOnly ? PVSpacing.space1 : PVSpacing.space3) {
            if let icon {
                PVIcon(icon, size: iconOnly ? 13 : 12)
            }
            if !iconOnly {
                triggerLabel
            }
            PVIcon(.chevronDown, size: iconOnly ? 9 : 11)
        }
        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, iconOnly ? PVSpacing.space3 : PVSpacing.space4)
        .frame(height: PVSpacing.controlHeightMedium)
        .frame(minWidth: iconOnly ? PVSpacing.controlHeightMedium : nil)
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(isFocused ? PVColor.borderFocus : PVColor.borderDefault, lineWidth: 1)
        )
        .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
    }

    @ViewBuilder
    private var triggerLabel: some View {
        Text(verbatim: triggerText)
            .foregroundStyle(showsPlaceholder ? PVColor.textFaint : PVColor.textPrimary)
            .lineLimit(1)
    }

    private var anchorReader: some View {
        PVSelectScreenAnchor { trigger, visible in
            triggerScreen = trigger
            visibleScreen = visible
        }
        .allowsHitTesting(false)
    }

    private var pressTracker: some View {
        PVSelectPressTracker(
            isEnabled: !isDisabled,
            triggerScreen: triggerScreen,
            menuScreen: menuScreen,
            onPress: { point in
                isFocused = true
                isTracking = true
                applySession { $0.handlePointer(.press(target(at: point))) }
            },
            onDrag: { point in
                applySession { $0.handlePointer(.drag(target(at: point))) }
            },
            onRelease: { point in
                applySession { $0.handlePointer(.release(target(at: point))) }
                isTracking = false
            },
            onClickAway: {
                guard session.isOpen, !isTracking else { return }
                applySession { $0.handlePointer(.clickAway) }
            },
            isMenuOpen: session.isOpen
        )
    }

    private func target(at point: CGPoint) -> PVSelectPointerTarget {
        PVSelectPointerTracking.hitTest(
            point: point,
            trigger: triggerScreen,
            rows: PVSelectPointerTracking.rowFrames(
                panel: menuScreen,
                count: options.count
            )
        )
    }
}

private struct PVSelectPositionContent: ViewModifier {
    var position: (current: Int, count: Int)?

    func body(content: Content) -> some View {
        if let position {
            content.accessibilityCustomContent(
                Text(L10n.DesignSystem.selectPosition),
                Text(verbatim: L10n.DesignSystem.selectOptionPosition(
                    current: position.current,
                    count: position.count
                ))
            )
        } else {
            content
        }
    }
}

private struct PVSelectScreenAnchor: NSViewRepresentable {
    var onChange: (CGRect, CGRect) -> Void

    func makeNSView(context: Context) -> NSView {
        NSView(frame: .zero)
    }

    func updateNSView(_ view: NSView, context: Context) {
        DispatchQueue.main.async {
            guard let window = view.window else { return }
            let screen = window.convertToScreen(view.convert(view.bounds, to: nil))
            let visible = window.screen?.visibleFrame ?? screen
            onChange(screen, visible)
        }
    }
}

/// Mouse-down tracking on the trigger, plus click-away while the menu is open
/// from the keyboard (not mid-drag).
private struct PVSelectPressTracker: NSViewRepresentable {
    var isEnabled: Bool
    var triggerScreen: CGRect
    var menuScreen: CGRect
    var onPress: (CGPoint) -> Void
    var onDrag: (CGPoint) -> Void
    var onRelease: (CGPoint) -> Void
    var onClickAway: () -> Void
    var isMenuOpen: Bool

    func makeNSView(context: Context) -> PVSelectPressView {
        let view = PVSelectPressView()
        apply(to: view)
        return view
    }

    func updateNSView(_ view: PVSelectPressView, context: Context) {
        apply(to: view)
        view.installClickAwayIfNeeded()
    }

    private func apply(to view: PVSelectPressView) {
        view.isEnabled = isEnabled
        view.triggerScreen = triggerScreen
        view.menuScreen = menuScreen
        view.onPress = onPress
        view.onDrag = onDrag
        view.onRelease = onRelease
        view.onClickAway = onClickAway
        view.isMenuOpen = isMenuOpen
    }
}

private final class PVSelectPressView: NSView {
    var isEnabled = true
    var isMenuOpen = false
    var triggerScreen = CGRect.zero
    var menuScreen = CGRect.zero
    var onPress: ((CGPoint) -> Void)?
    var onDrag: ((CGPoint) -> Void)?
    var onRelease: ((CGPoint) -> Void)?
    var onClickAway: (() -> Void)?
    /// Opaque `NSEvent` tokens. `deinit` is nonisolated, so these stay
    /// `nonisolated(unsafe)` the same way other kit caches do.
    nonisolated(unsafe) private var clickAwayMonitor: Any?
    nonisolated(unsafe) private var trackMonitor: Any?
    private var isTracking = false

    override var acceptsFirstResponder: Bool { false }

    override func hitTest(_ point: NSPoint) -> NSView? {
        guard isEnabled else { return nil }
        return self
    }

    override func mouseDown(with event: NSEvent) {
        guard isEnabled else { return }
        isTracking = true
        onPress?(screenPoint(event))
        if let trackMonitor {
            NSEvent.removeMonitor(trackMonitor)
        }
        trackMonitor = NSEvent.addLocalMonitorForEvents(matching: [.leftMouseDragged, .leftMouseUp]) { [weak self] next in
            guard let self else { return next }
            switch next.type {
            case .leftMouseDragged:
                self.onDrag?(self.screenPoint(next))
            case .leftMouseUp:
                self.onRelease?(self.screenPoint(next))
                self.finishTracking()
            default:
                break
            }
            return nil
        }
    }

    private func finishTracking() {
        isTracking = false
        if let trackMonitor {
            NSEvent.removeMonitor(trackMonitor)
            self.trackMonitor = nil
        }
    }

    func installClickAwayIfNeeded() {
        if isMenuOpen {
            guard clickAwayMonitor == nil else { return }
            clickAwayMonitor = NSEvent.addLocalMonitorForEvents(matching: .leftMouseDown) { [weak self] event in
                guard let self, self.isMenuOpen, !self.isTracking else { return event }
                let point = self.screenPoint(event)
                if self.triggerScreen.contains(point) || self.menuScreen.contains(point) {
                    return event
                }
                self.onClickAway?()
                return event
            }
        } else if let clickAwayMonitor {
            NSEvent.removeMonitor(clickAwayMonitor)
            self.clickAwayMonitor = nil
        }
    }

    deinit {
        if let clickAwayMonitor {
            NSEvent.removeMonitor(clickAwayMonitor)
        }
        if let trackMonitor {
            NSEvent.removeMonitor(trackMonitor)
        }
    }

    private func screenPoint(_ event: NSEvent) -> CGPoint {
        guard let window = event.window ?? window else { return .zero }
        return window.convertToScreen(NSRect(origin: event.locationInWindow, size: .zero)).origin
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

#Preview("Field") {
    PVSelect(
        selection: .constant("smith"),
        options: [
            PVSelectOption(value: "smith", label: "smith-family.provenencia"),
            PVSelectOption(value: "halvorsen", label: "halvorsen-line.provenencia"),
            PVSelectOption(value: "baca", label: "baca-oaxaca.provenencia")
        ]
    )
    .padding(PVSpacing.space9)
    .frame(width: 320)
    .background(PVColor.surfacePage)
}

#Preview("Chip") {
    PVSelect(
        selection: .constant(""),
        options: [
            PVSelectOption(value: "", label: "All types"),
            PVSelectOption(value: "census", label: "Census")
        ],
        icon: .filter,
        fillsWidth: false,
        accessibilityLabel: "Filter"
    )
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}

#Preview("Icon-only") {
    PVSelect(
        selection: .constant(""),
        options: [
            PVSelectOption(value: "", label: "All data types"),
            PVSelectOption(value: "text", label: "Text")
        ],
        icon: .filter,
        iconOnly: true,
        fillsWidth: false,
        accessibilityLabel: "Filter Data type column"
    )
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}

#Preview("Disabled") {
    PVSelect(
        selection: .constant("gregorian"),
        options: [
            PVSelectOption(value: "gregorian", label: "Gregorian"),
            PVSelectOption(value: "julian", label: "Julian")
        ],
        isDisabled: true
    )
    .padding(PVSpacing.space9)
    .frame(width: 280)
    .background(PVColor.surfacePage)
}

#Preview("Empty") {
    PVSelect(
        selection: .constant(""),
        options: [],
        placeholder: "No projects in Documents"
    )
    .padding(PVSpacing.space9)
    .frame(width: 280)
    .background(PVColor.surfacePage)
}
