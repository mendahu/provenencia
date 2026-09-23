import SwiftUI

/// One option in a `PVSelect` — `value` is the stable identifier (what a
/// binding stores), `label` is what's shown. Plain `String`, not
/// `LocalizedStringResource`: options are almost always data (a project
/// filename, a person's name), not static UI copy — see
/// `docs/macos-client-patterns.md` §6.
struct PVSelectOption: Identifiable {
    let id: String
    let label: String
    var accessibilityIdentifier: String?

    init(value: String, label: String, accessibilityIdentifier: String? = nil) {
        self.id = value
        self.label = label
        self.accessibilityIdentifier = accessibilityIdentifier
    }
}

/// A styled dropdown select built on the floating `PVContextMenu` kit
/// (shared panel chrome + ↑/↓/⏎/Esc + type-to-select). Replaces SwiftUI
/// `Menu` / `Picker` so label chrome and keyboard policy stay on `PV*`
/// tokens. A focused trigger opens on ↑/↓; letters jump to matching options.
///
/// - **Field** (default): full-width control matching `PVInput` rest border.
/// - **Chip** (`icon:` set, `fillsWidth: false`): compact toolbar select
///   (Sources filter / sort).
struct PVSelect: View {
    @Binding private var selection: String
    private let options: [PVSelectOption]
    private let size: PVControlSize
    private let icon: PVSymbol?
    private let displayLabel: String?
    private let menuWidth: CGFloat
    private let fillsWidth: Bool
    private let accessibilityLabelResource: LocalizedStringResource?
    private let accessibilityIdentifier: String?

    @State private var menuState = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive
    @FocusState private var isFocused: Bool

    init(
        selection: Binding<String>,
        options: [PVSelectOption],
        size: PVControlSize = .md,
        icon: PVSymbol? = nil,
        displayLabel: String? = nil,
        menuWidth: CGFloat = 210,
        fillsWidth: Bool = true,
        accessibilityLabel: LocalizedStringResource? = nil,
        accessibilityIdentifier: String? = nil
    ) {
        self._selection = selection
        self.options = options
        self.size = size
        self.icon = icon
        self.displayLabel = displayLabel
        self.menuWidth = menuWidth
        self.fillsWidth = fillsWidth
        self.accessibilityLabelResource = accessibilityLabel
        self.accessibilityIdentifier = accessibilityIdentifier
    }

    private var triggerLabel: String {
        if let displayLabel { return displayLabel }
        return options.first(where: { $0.id == selection })?.label ?? ""
    }

    var body: some View {
        Button {
            toggleMenu()
        } label: {
            trigger
        }
        .buttonStyle(.plain)
        .focused($isFocused)
        .onMoveCommand(perform: handleMoveCommand)
        .onKeyPress(action: handleTriggerKey)
        .accessibilityLabel(accessibilityLabelText)
        .accessibilityValue(Text(verbatim: triggerLabel))
        .accessibilityAddIdentifiers(accessibilityIdentifier)
        .pvContextMenu($menuState, keyboard: options.isEmpty ? nil : $keyboard) {
            PVContextMenuPanel(
                width: menuWidth,
                accessibilityIdentifier: accessibilityIdentifier.map { "\($0).menu" }
            ) {
                ForEach(Array(options.enumerated()), id: \.element.id) { offset, option in
                    PVContextMenuItem(
                        plainTitle: option.label,
                        index: offset,
                        isSelected: option.id == selection,
                        accessibilityIdentifier: option.accessibilityIdentifier
                    ) {
                        selection = option.id
                    }
                }
            }
        }
    }

    private var selectedIndex: Int {
        options.firstIndex(where: { $0.id == selection }) ?? -1
    }

    private func toggleMenu() {
        if menuState.isPresented {
            menuState.dismiss()
        } else {
            openMenu(highlight: selectedIndex)
        }
    }

    private func openMenu(highlight: Int) {
        guard !options.isEmpty else { return }
        keyboard = PVContextMenuKeyboard(
            itemCount: options.count,
            activeIndex: highlight,
            itemTitles: options.map(\.label)
        )
        menuState.present(at: CGPoint(x: 0, y: triggerHeight + PVSpacing.space2))
    }

    private func handleMoveCommand(_ direction: MoveCommandDirection) {
        guard !menuState.isPresented, !options.isEmpty else { return }
        switch direction {
        case .down:
            openMenu(highlight: selectedIndex >= 0 ? selectedIndex : 0)
        case .up:
            openMenu(highlight: selectedIndex >= 0 ? selectedIndex : options.count - 1)
        default:
            break
        }
    }

    private func handleTriggerKey(_ press: KeyPress) -> KeyPress.Result {
        guard !options.isEmpty, !menuState.isPresented else { return .ignored }
        switch press.key {
        case .downArrow:
            openMenu(highlight: selectedIndex >= 0 ? selectedIndex : 0)
            return .handled
        case .upArrow:
            openMenu(highlight: selectedIndex >= 0 ? selectedIndex : options.count - 1)
            return .handled
        default:
            guard press.characters.count == 1,
                  let character = press.characters.first,
                  character.isLetter || character.isNumber,
                  !press.modifiers.contains(.command),
                  !press.modifiers.contains(.control)
            else { return .ignored }
            if !menuState.isPresented {
                openMenu(highlight: selectedIndex)
            }
            if keyboard.applyTypeSelect(character),
               options.indices.contains(keyboard.activeIndex)
            {
                selection = options[keyboard.activeIndex].id
            }
            return .handled
        }
    }

    private var accessibilityLabelText: Text {
        if let accessibilityLabelResource {
            Text(accessibilityLabelResource)
        } else {
            Text(triggerLabel)
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
            Text(triggerLabel)
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
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
        HStack(spacing: PVSpacing.space3) {
            if let icon {
                PVIcon(icon, size: 12)
            }
            Text(triggerLabel)
                .lineLimit(1)
            PVIcon(.chevronDown, size: 11)
        }
        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
        .foregroundStyle(PVColor.textPrimary)
        .padding(.horizontal, PVSpacing.space4)
        .frame(height: PVSpacing.controlHeightMedium)
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
