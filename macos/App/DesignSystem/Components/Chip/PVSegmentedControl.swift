import SwiftUI

/// A few exclusive options in one sunken track, chosen inline: the Swift form
/// of the web kit's segmented `ChipGroup`. The selection is a value and a
/// `select` action, so a caller that navigates, sets a model field or refuses
/// while locked uses it the same way. Each segment is a `PVChip` (so it shows
/// keyboard focus and reads as selected); the group reads its `label` to
/// VoiceOver. A `focus:` handle lands on the selected segment, or on the first
/// enabled one when the selected one is disabled, so a host can Tab or write
/// focus into the control like any other kit field. With a
/// handle the control is one Tab stop, with or without Full Keyboard Access:
/// the segments without the handle leave the key loop. Space
/// chooses, and ← / → move the selection to the neighbouring enabled segment
/// (no wrap), as a native segmented control does. The control does not claim
/// focus after a choice: a host whose handle follows the selection re-claims
/// it, so while the keyboard is on the control (the handle is focused) Space,
/// an arrow or a ⌘ shortcut keep focus on the newly selected segment. The
/// control does not read the handle's `next` / `previous` neighbours. With no
/// handle it adds no Tab stop (as today).
///
/// Meaning-free like `PVChip`: the options and what choosing one does belong
/// to the caller.
struct PVSegmentedControl<Value: Hashable>: View {
    struct Item: Identifiable {
        enum Title {
            case key(LocalizedStringResource)
            case text(String)
        }

        var value: Value
        var title: Title
        var glyph: PVGlyph?
        var isDisabled: Bool
        /// ⌘ plus this key chooses the item while the control is on screen.
        var shortcut: KeyEquivalent?
        /// Every segment carries one, so a UI test finds it.
        var accessibilityIdentifier: String
        var id: String { accessibilityIdentifier }

        /// The tooltip that names the shortcut; nil when the item has none.
        fileprivate var shortcutHelp: String? {
            guard let shortcut else { return nil }
            let name: String
            switch title {
            case .key(let key): name = L10n.string(key)
            case .text(let text): name = text
            }
            return L10n.DesignSystem.segmentShortcutHelp(name, key: String(shortcut.character))
        }

        /// Fixed UI copy.
        init(
            _ title: LocalizedStringResource,
            value: Value,
            glyph: PVGlyph? = nil,
            isDisabled: Bool = false,
            shortcut: KeyEquivalent? = nil,
            accessibilityIdentifier: String
        ) {
            self.value = value
            self.title = .key(title)
            self.glyph = glyph
            self.isDisabled = isDisabled
            self.shortcut = shortcut
            self.accessibilityIdentifier = accessibilityIdentifier
        }

        /// Data label (a catalog value type's name).
        init(
            text: String,
            value: Value,
            glyph: PVGlyph? = nil,
            isDisabled: Bool = false,
            shortcut: KeyEquivalent? = nil,
            accessibilityIdentifier: String
        ) {
            self.value = value
            self.title = .text(text)
            self.glyph = glyph
            self.isDisabled = isDisabled
            self.shortcut = shortcut
            self.accessibilityIdentifier = accessibilityIdentifier
        }
    }

    private let label: LocalizedStringResource
    private let items: [Item]
    private let selection: Value
    private let size: PVControlSize
    private let focus: PVFieldFocus?
    private let select: (Value) -> Void

    init(
        _ label: LocalizedStringResource,
        items: [Item],
        selection: Value,
        size: PVControlSize = .md,
        focus: PVFieldFocus? = nil,
        select: @escaping (Value) -> Void
    ) {
        self.label = label
        self.items = items
        self.selection = selection
        self.size = size
        self.focus = focus
        self.select = select
    }

    var body: some View {
        PVChipGroup(style: .segmented) {
            ForEach(items) { item in
                segment(item)
                    .disabled(item.isDisabled)
                    .modifier(PVSegmentShortcut(shortcut: item.shortcut))
                    .help(item.shortcutHelp ?? "")
                    .accessibilityIdentifier(item.accessibilityIdentifier)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(label))
    }

    /// The id of the segment that carries the host's focus handle: the selected
    /// one when it is enabled, else the first enabled one, else none.
    static func focusTarget(items: [Item], selection: Value) -> String? {
        if let selected = items.first(where: { $0.value == selection && !$0.isDisabled }) {
            return selected.id
        }
        return items.first(where: { !$0.isDisabled })?.id
    }

    /// The value an arrow key moves to from `selection`: the next (or previous)
    /// enabled item, skipping disabled ones, with no wrap; nil at either end.
    static func arrowTarget(items: [Item], selection: Value, forward: Bool) -> Value? {
        guard let index = items.firstIndex(where: { $0.value == selection }) else { return nil }
        let candidates = forward ? Array(items[(index + 1)...]) : Array(items[..<index].reversed())
        return candidates.first(where: { !$0.isDisabled })?.value
    }

    /// The chip. With a host handle only the focus target is a Tab stop: it
    /// carries the handle, Space chooses it, and ← / → move the selection.
    /// Every other segment is a clickable chip that leaves the key loop
    /// (not focusable), with or without Full Keyboard Access.
    @ViewBuilder
    private func segment(_ item: Item) -> some View {
        if let focus, !item.isDisabled, item.id == Self.focusTarget(items: items, selection: selection) {
            focus.apply(AnyView(
                chip(item)
                    .focusable()
                    .onKeyPress(.space) {
                        select(item.value)
                        return .handled
                    }
                    .onKeyPress(.leftArrow) { arrow(forward: false) }
                    .onKeyPress(.rightArrow) { arrow(forward: true) }
            ))
        } else if focus != nil {
            chip(item).focusable(false)
        } else {
            chip(item)
        }
    }

    private func arrow(forward: Bool) -> KeyPress.Result {
        guard let value = Self.arrowTarget(items: items, selection: selection, forward: forward) else {
            return .ignored
        }
        select(value)
        return .handled
    }

    @ViewBuilder
    private func chip(_ item: Item) -> some View {
        let isSelected = item.value == selection
        switch item.title {
        case .key(let key):
            PVChip(
                key,
                isSelected: isSelected,
                size: size,
                glyph: item.glyph,
                expands: true,
                selectionLift: true,
                action: { select(item.value) }
            )
        case .text(let text):
            PVChip(
                text: text,
                isSelected: isSelected,
                size: size,
                glyph: item.glyph,
                expands: true,
                selectionLift: true,
                action: { select(item.value) }
            )
        }
    }
}

private struct PVSegmentShortcut: ViewModifier {
    let shortcut: KeyEquivalent?

    @ViewBuilder
    func body(content: Content) -> some View {
        if let shortcut {
            content.keyboardShortcut(shortcut, modifiers: .command)
        } else {
            content
        }
    }
}

#Preview("Segmented control") {
    PVSegmentedControl(
        "Kind",
        items: [
            .init("Point", value: 0, accessibilityIdentifier: "preview.point"),
            .init("Range", value: 1, accessibilityIdentifier: "preview.range"),
        ],
        selection: 0,
        select: { _ in }
    )
    .frame(width: 280)
    .padding(PVSpacing.space9)
    .background(PVColor.surfacePage)
}
