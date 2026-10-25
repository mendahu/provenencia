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

/// Default one-line option label. Marks, a second line, or other rich
/// row chrome go through `PVSelect`'s `row:` builder — same slot as
/// `PVComboBox`.
struct PVSelectPlainRow: View {
    let text: String

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.body(size: PVTypeScale.bodySmall))
            .foregroundStyle(PVColor.textPrimary)
            .lineLimit(1)
    }
}

/// A styled dropdown select built on the floating `PVContextMenu` kit.
/// Click the field to open, click a row to commit, click away or Escape
/// to dismiss. Keys go to `PVSelectSession`. There is no press-drag.
///
/// - **Field** (default): full-width control matching `PVInput` rest border.
/// - **Chip** (`icon:` set, `fillsWidth: false`): compact toolbar select.
/// - **Icon-only chip** (`icon` + `iconOnly`): table column filter; pass
///   `accessibilityLabel`.
/// - **`row:`** optional custom option body. Keyboard, checkmark, and
///   placement stay on the select.
struct PVSelect<Row: View>: View {
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
    private let rowHeight: CGFloat
    private let isDisabled: Bool
    private let isInvalid: Bool
    private let accessibilityLabelResource: LocalizedStringResource?
    private let accessibilitySpokenLabel: String?
    private let accessibilityIdentifier: String?
    private let closedHelp: String?
    private let row: (PVSelectOption) -> Row

    @State private var session: PVSelectSession
    @State private var menuState = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive
    @State private var triggerScreen = CGRect.zero
    @State private var visibleScreen = CGRect.zero
    @State private var menuScreen = CGRect.zero
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
        rowHeight: CGFloat = PVSelectPlacement.rowHeight,
        isDisabled: Bool = false,
        isInvalid: Bool = false,
        accessibilityLabel: LocalizedStringResource? = nil,
        accessibilitySpokenLabel: String? = nil,
        accessibilityIdentifier: String? = nil,
        closedHelp: String? = nil,
        @ViewBuilder row: @escaping (PVSelectOption) -> Row
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
        self.rowHeight = rowHeight
        self.isDisabled = isDisabled
        self.isInvalid = isInvalid
        self.accessibilityLabelResource = accessibilityLabel
        self.accessibilitySpokenLabel = accessibilitySpokenLabel
        self.accessibilityIdentifier = accessibilityIdentifier
        self.closedHelp = closedHelp
        self.row = row
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
            .overlay {
                PVSelectClickCatcher(isEnabled: !isDisabled, onClick: toggleMenu)
            }
            .opacity(isDisabled ? 0.45 : 1)
            .focusable(!isDisabled)
            .focused($isFocused)
            .disabled(isDisabled)
            .onMoveCommand(perform: handleMoveCommand)
            .onKeyPress(action: handleTriggerKey)
            .background(anchorReader)
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
            .pvExpandedState(spoken.isExpanded)
            .modifier(PVSelectPositionContent(position: spoken.position))
            .pvContextMenu(
                $menuState,
                keyboard: options.isEmpty ? nil : $keyboard,
                stealKeys: false,
                dismissOnMouseUp: false,
                dismissOnClickAway: true
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
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    ForEach(Array(options.enumerated()), id: \.element.id) { offset, option in
                        PVContextMenuItem(
                            index: offset,
                            isSelected: option.id == session.selection,
                            showsSelectionMark: true,
                            accessibilityIdentifier: option.accessibilityIdentifier,
                            action: { applySession { $0.commit(index: offset, close: true) } }
                        ) {
                            row(option)
                        }
                        .id(offset)
                    }
                }
            }
            .frame(maxHeight: height)
            .pvScrollsRowIntoView(
                session.highlightIndex >= 0 ? session.highlightIndex : nil,
                anchor: .center,
                animated: false,
                scrollsOnAppear: true
            )
        }
    }

    private func presentMenu() {
        let contentHeight = PVSelectPlacement.contentHeight(
            optionCount: options.count,
            maxVisibleRows: maxVisibleRows,
            rowHeight: rowHeight
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

    private func toggleMenu() {
        isFocused = true
        applySession { $0.toggle() }
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

    @ViewBuilder
    private var trigger: some View {
        if icon != nil {
            chipTrigger
        } else {
            fieldTrigger
        }
    }

    private var fieldTrigger: some View {
        closedLabel
            .frame(minWidth: 0, maxWidth: fillsWidth ? .infinity : nil, alignment: .leading)
            .modifier(PVInputChrome(
                size: size,
                isFocused: isFocused,
                isInvalid: isInvalid,
                trailingSuffixWidth: PVSelectFieldChrome.chevronSize
            ))
            .overlay(alignment: .trailing) {
                PVIcon(.chevronDown, size: PVSelectFieldChrome.chevronSize)
                    .foregroundStyle(PVColor.textFaint)
                    .padding(.trailing, PVInputChrome.horizontalInset)
                    .allowsHitTesting(false)
            }
    }

    private var chipTrigger: some View {
        HStack(spacing: iconOnly ? PVSpacing.space1 : PVSpacing.space3) {
            if let icon {
                PVIcon(icon, size: iconOnly ? 13 : 12)
            }
            if !iconOnly {
                closedLabel
            }
            PVIcon(.chevronDown, size: iconOnly ? 9 : 11)
                .fixedSize()
                .layoutPriority(1)
        }
        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, iconOnly ? PVSpacing.space3 : PVSpacing.space4)
        .frame(height: PVSpacing.controlHeightMedium)
        .frame(
            minWidth: iconOnly ? PVSpacing.controlHeightMedium : 0,
            maxWidth: fillsWidth ? .infinity : nil,
            alignment: .leading
        )
        .background(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .fill(PVColor.surfaceRaised)
        )
        .overlay(triggerBorder)
        .pvFocusRing(isFocused, cornerRadius: PVRadius.sm)
    }

    private var closedLabel: some View {
        Text(verbatim: triggerText)
            .foregroundStyle(showsPlaceholder ? PVColor.textFaint : PVColor.textPrimary)
            .lineLimit(1)
            .truncationMode(.tail)
            .frame(minWidth: 0, alignment: .leading)
            .help(Text(verbatim: closedHelp ?? triggerText))
    }

    private var triggerBorder: some View {
        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
            .stroke(triggerBorderColor, lineWidth: 1)
            .allowsHitTesting(false)
            .pvAnimation(PVMotion.fastStandard, value: isFocused)
            .pvAnimation(PVMotion.fastStandard, value: isInvalid)
    }

    private var triggerBorderColor: Color {
        if isInvalid { return PVColor.danger }
        return isFocused ? PVColor.borderFocus : PVColor.borderDefault
    }

    private var anchorReader: some View {
        PVSelectScreenAnchor { trigger, visible in
            triggerScreen = trigger
            visibleScreen = visible
        }
        .allowsHitTesting(false)
    }

}

extension PVSelect where Row == PVSelectPlainRow {
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
        rowHeight: CGFloat = PVSelectPlacement.rowHeight,
        isDisabled: Bool = false,
        isInvalid: Bool = false,
        accessibilityLabel: LocalizedStringResource? = nil,
        accessibilitySpokenLabel: String? = nil,
        accessibilityIdentifier: String? = nil,
        closedHelp: String? = nil
    ) {
        self.init(
            selection: selection,
            options: options,
            size: size,
            icon: icon,
            iconOnly: iconOnly,
            displayLabel: displayLabel,
            placeholder: placeholder,
            menuWidth: menuWidth,
            fillsWidth: fillsWidth,
            maxVisibleRows: maxVisibleRows,
            rowHeight: rowHeight,
            isDisabled: isDisabled,
            isInvalid: isInvalid,
            accessibilityLabel: accessibilityLabel,
            accessibilitySpokenLabel: accessibilitySpokenLabel,
            accessibilityIdentifier: accessibilityIdentifier,
            closedHelp: closedHelp
        ) { option in
            PVSelectPlainRow(text: option.label)
        }
    }
}

enum PVSelectFieldChrome {
    static let chevronSize: CGFloat = 14
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

/// Catches the field click, including when the window is not yet key.
/// It has no chrome: SwiftUI owns the trigger's size.
private struct PVSelectClickCatcher: NSViewRepresentable {
    var isEnabled: Bool
    var onClick: () -> Void

    func makeNSView(context: Context) -> PVSelectClickCatcherView {
        let view = PVSelectClickCatcherView()
        view.isEnabled = isEnabled
        view.onClick = onClick
        return view
    }

    func updateNSView(_ view: PVSelectClickCatcherView, context: Context) {
        view.isEnabled = isEnabled
        view.onClick = onClick
    }

    func sizeThatFits(
        _ proposal: ProposedViewSize,
        nsView: PVSelectClickCatcherView,
        context: Context
    ) -> CGSize? {
        CGSize(
            width: proposal.width?.isFinite == true ? proposal.width ?? 0 : 0,
            height: proposal.height?.isFinite == true ? proposal.height ?? 0 : 0
        )
    }
}

private final class PVSelectClickCatcherView: NSView {
    var isEnabled = true
    var onClick: (() -> Void)?

    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        setAccessibilityElement(false)
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) is not used")
    }

    override var acceptsFirstResponder: Bool { false }

    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { true }

    override func hitTest(_ point: NSPoint) -> NSView? {
        guard isEnabled, bounds.contains(point) else { return nil }
        return self
    }

    override func mouseDown(with event: NSEvent) {
        guard isEnabled else { return }
        onClick?()
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

#Preview("Rich rows") {
    PVSelect(
        selection: .constant("cit-1"),
        options: [
            PVSelectOption(value: "", label: "New citation"),
            PVSelectOption(value: "cit-1", label: "CIT-F4N2P"),
        ],
        displayLabel: "CIT-F4N2P",
        menuWidth: 400,
        fillsWidth: false,
        rowHeight: 64
    ) { option in
        VStack(alignment: .leading, spacing: 2) {
            Text(verbatim: option.label)
                .font(PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
            if option.id == "cit-1" {
                Text(verbatim: "Margt. Alderwick, Farmer")
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(2)
            }
        }
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
