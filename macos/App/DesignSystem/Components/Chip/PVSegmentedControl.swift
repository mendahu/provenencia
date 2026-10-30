import SwiftUI

/// A few exclusive options in one sunken track, chosen inline: the Swift form
/// of the web kit's segmented `ChipGroup`. The selection is a value and a
/// `select` action, so a caller that navigates, sets a model field or refuses
/// while locked uses it the same way. Each segment is a `PVChip` (so it shows
/// keyboard focus and reads as selected); the group reads its `label` to
/// VoiceOver.
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
    private let select: (Value) -> Void

    init(
        _ label: LocalizedStringResource,
        items: [Item],
        selection: Value,
        size: PVControlSize = .md,
        select: @escaping (Value) -> Void
    ) {
        self.label = label
        self.items = items
        self.selection = selection
        self.size = size
        self.select = select
    }

    var body: some View {
        PVChipGroup(style: .segmented) {
            ForEach(items) { item in
                chip(item)
                    .disabled(item.isDisabled)
                    .modifier(PVSegmentShortcut(shortcut: item.shortcut))
                    .help(item.shortcutHelp ?? "")
                    .accessibilityIdentifier(item.accessibilityIdentifier)
            }
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(label))
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
