import SwiftUI

// Action menus, in two layers over `PVContextMenu`.
//
// Items: `PVActionMenuItem` lists are written once. `PVActionMenuRows` draws
// them as menu rows (indexes follow position), and `PVActionMenuItem.keyboardTitles`
// gives the keyboard titles for a menu opened from a button (`.pvContextMenu`).
//
// Right-click: `.pvRightClickMenu` adds the menu's state, the right-click
// trigger and one VoiceOver action per enabled item on top of the items.
// `PVSelect` keeps using `.pvContextMenu` directly: it seeds its own highlight.

/// One row of an action menu.
struct PVActionMenuItem {
    let title: LocalizedStringResource
    var isEnabled: Bool = true
    var accessibilityIdentifier: String? = nil
    var action: () -> Void = {}

    /// Keyboard titles for `PVContextMenuKeyboard.opened(titles:)`: every
    /// item's title in order (disabled rows keep their slot so indexes match
    /// positions), or nil when no item is enabled.
    static func keyboardTitles(_ items: [PVActionMenuItem]) -> [String]? {
        guard items.contains(where: \.isEnabled) else { return nil }
        return items.map { L10n.string($0.title) }
    }
}

/// The rows of an action menu; place inside `PVContextMenuPanel`.
struct PVActionMenuRows: View {
    let items: [PVActionMenuItem]

    var body: some View {
        ForEach(Array(items.enumerated()), id: \.offset) { i, item in
            PVContextMenuItem(
                item.title,
                index: item.isEnabled ? i : nil,
                isEnabled: item.isEnabled,
                accessibilityIdentifier: item.accessibilityIdentifier,
                action: item.action
            )
        }
    }
}

extension View {
    /// A right-click (or Ctrl-click) action menu on this view: opens at the
    /// pointer, has ↑/↓/⏎ navigation, and lists each enabled item as a
    /// VoiceOver action. The trigger and panel share this view's coordinates.
    func pvRightClickMenu(
        title: LocalizedStringResource? = nil,
        accessibilityIdentifier: String?,
        items: [PVActionMenuItem]
    ) -> some View {
        modifier(PVRightClickMenuModifier(
            title: title,
            accessibilityIdentifier: accessibilityIdentifier,
            items: items
        ))
    }
}

private struct PVRightClickMenuModifier: ViewModifier {
    let title: LocalizedStringResource?
    let accessibilityIdentifier: String?
    let items: [PVActionMenuItem]

    @State private var menu = PVContextMenuState()
    @State private var keyboard = PVContextMenuKeyboard.inactive

    func body(content: Content) -> some View {
        let titles = PVActionMenuItem.keyboardTitles(items)
        content
            .accessibilityActions {
                ForEach(Array(items.enumerated()).filter { $0.element.isEnabled }, id: \.offset) { _, item in
                    Button(action: item.action) { Text(item.title) }
                }
            }
            .pvContextMenuTrigger($menu)
            .pvContextMenu(
                $menu,
                keyboard: titles == nil ? nil : $keyboard,
                keyboardTitles: titles
            ) {
                PVContextMenuPanel(title: title, accessibilityIdentifier: accessibilityIdentifier) {
                    PVActionMenuRows(items: items)
                }
            }
    }
}
