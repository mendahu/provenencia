import SwiftUI

/// Secondary chip that opens a floating `PVContextMenuPanel` below itself.
/// Prefer this over SwiftUI `Menu` / `Picker` when the board specifies custom
/// popup rows (S5-D2 filter + sort).
struct PVPopupMenuButton<MenuContent: View>: View {
    let icon: PVSymbol
    let label: String
    let accessibilityLabel: LocalizedStringResource
    var accessibilityIdentifier: String?
    var menuWidth: CGFloat = 210
    var menuTitle: LocalizedStringResource?
    /// Number of keyboard-activatable rows. Pass matching `index:` on each item.
    var itemCount: Int = 0
    @ViewBuilder var menu: () -> MenuContent

    @State private var state = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive

    var body: some View {
        Button {
            if state.isPresented {
                state.dismiss()
            } else {
                keyboard = PVContextMenuKeyboard(itemCount: itemCount, activeIndex: -1)
                state.present(
                    at: CGPoint(x: 0, y: PVSpacing.controlHeightMedium + PVSpacing.space2)
                )
            }
        } label: {
            HStack(spacing: PVSpacing.space3) {
                PVIcon(icon, size: 12)
                Text(label)
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
                    .stroke(PVColor.borderDefault, lineWidth: 1)
            )
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(accessibilityLabel))
        .accessibilityAddIdentifiers(accessibilityIdentifier)
        .pvContextMenu($state, keyboard: itemCount > 0 ? $keyboard : nil) {
            PVContextMenuPanel(
                title: menuTitle,
                width: menuWidth,
                accessibilityIdentifier: accessibilityIdentifier.map { "\($0).menu" }
            ) {
                menu()
            }
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
