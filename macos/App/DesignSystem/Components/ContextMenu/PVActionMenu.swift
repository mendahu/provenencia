import SwiftUI

// Action menus, in three layers over `PVContextMenu`.
//
// Items: `PVActionMenuItem` lists are written once. `PVActionMenuRows` draws
// them as menu rows (indexes follow position), and `PVActionMenuItem.keyboard`
// gives the menu keyboard its titles and which rows are disabled.
//
// Menu: `.pvActionMenu` is the one recipe for an action menu on any view: its
// keyboard state, its seed on open, its panel and rows. The keys skip disabled
// rows, and it closes through `.pvPopup`. A menu opened from a button is one
// `.pvActionMenu` on that button.
//
// Right-click: `.pvRightClickMenu` adds the menu's state, the right-click
// trigger and one VoiceOver action per enabled item on top of `.pvActionMenu`.
// `PVSelect` and the toolbar's history menu use `.pvContextMenu` directly: `PVSelect`'s keys stay in its field, and the history menu's rows are not action items.

/// One row of an action menu.
struct PVActionMenuItem {
    let title: LocalizedStringResource
    var isEnabled: Bool = true
    var accessibilityIdentifier: String? = nil
    var action: () -> Void = {}

    /// The keyboard of an action menu that just opened: every title in order
    /// (disabled rows keep their slot so indexes match positions) and which
    /// rows are disabled, or nil when no item is enabled.
    static func keyboard(_ items: [PVActionMenuItem]) -> PVContextMenuKeyboard? {
        guard items.contains(where: \.isEnabled) else { return nil }
        return .opened(
            titles: items.map { L10n.string($0.title) },
            disabledRows: IndexSet(items.indices.filter { !items[$0].isEnabled })
        )
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
            heading: title.map { Text($0) },
            accessibilityIdentifier: accessibilityIdentifier,
            items: items
        ))
    }

    /// The same menu with a title that is data (a ref), drawn as given.
    func pvRightClickMenu(
        verbatimTitle: String,
        accessibilityIdentifier: String?,
        items: [PVActionMenuItem]
    ) -> some View {
        modifier(PVRightClickMenuModifier(
            heading: Text(verbatim: verbatimTitle),
            accessibilityIdentifier: accessibilityIdentifier,
            items: items
        ))
    }
}

private struct PVRightClickMenuModifier: ViewModifier {
    let heading: Text?
    let accessibilityIdentifier: String?
    let items: [PVActionMenuItem]

    @State private var menu = PVContextMenuState()

    func body(content: Content) -> some View {
        content
            .accessibilityActions {
                ForEach(Array(items.enumerated()).filter { $0.element.isEnabled }, id: \.offset) { _, item in
                    Button(action: item.action) { Text(item.title) }
                }
            }
            .pvContextMenuTrigger($menu)
            .pvActionMenu(
                $menu, heading: heading, accessibilityIdentifier: accessibilityIdentifier, items: items
            )
    }
}

/// Holds an action menu's keyboard and seeds it each time the menu opens.
private struct PVActionMenuModifier: ViewModifier {
    @Binding var state: PVContextMenuState
    let heading: Text?
    let accessibilityIdentifier: String?
    let items: [PVActionMenuItem]

    @State private var keyboard = PVContextMenuKeyboard.inactive

    func body(content: Content) -> some View {
        content
            .onChange(of: state.isPresented) { _, open in
                if open { keyboard = PVActionMenuItem.keyboard(items) ?? .inactive }
            }
            // A menu with no enabled item gets no keyboard: only Escape is taken.
            .pvContextMenu($state, keyboard: items.contains(where: \.isEnabled) ? $keyboard : nil) {
                PVContextMenuPanel(heading: heading, accessibilityIdentifier: accessibilityIdentifier) {
                    PVActionMenuRows(items: items)
                }
            }
    }
}

extension View {
    /// The one action-menu recipe: `state`'s menu on this view, listing `items`,
    /// with ↑/↓, Home/End, type-select and Return that skip disabled rows.
    /// The menu opens at `state.origin`, in this view's coordinates; the trigger
    /// (a button, or `.pvContextMenuTrigger`) sets it.
    func pvActionMenu(
        _ state: Binding<PVContextMenuState>,
        heading: Text? = nil,
        accessibilityIdentifier: String?,
        items: [PVActionMenuItem]
    ) -> some View {
        modifier(PVActionMenuModifier(
            state: state, heading: heading, accessibilityIdentifier: accessibilityIdentifier, items: items
        ))
    }
}
