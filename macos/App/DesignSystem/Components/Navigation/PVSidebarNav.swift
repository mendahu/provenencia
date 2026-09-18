import SwiftUI

/// Mirrors the S5-D3 `SidebarNav.jsx` proposal: flat items plus optional
/// nested config children (always expanded; no disclosure). Collapsed mode
/// shows the parent icon and a smaller child icon cluster under a hairline.
///
/// `.accessibilityIdentifier` is normally left to the call site
/// (`DesignSystem/README.md`), but this component renders every row
/// itself, so there's no single returned `View` a call site could chain an
/// identifier onto — `PVSidebarNavItem.accessibilityIdentifier` carries
/// that call-site-owned value as data instead.
struct PVSidebarNavItem: Identifiable, Equatable {
    let id: String
    let label: LocalizedStringResource
    let icon: PVSymbol
    let accessibilityIdentifier: String
    /// Trailing entity count (e.g. how many Sources exist). `nil` hides
    /// the badge entirely — used for a destination with no backing count
    /// query yet, not to mean zero. Nested children never show a count.
    var count: Int? = nil
    /// Nested config destinations under this item (S5-D3 Sources family).
    var children: [PVSidebarNavItem] = []
}

/// Provenencia's leading sidebar navigation list. Expanded mode shows an
/// optional group eyebrow plus icon + label rows; collapsed mode is an
/// icon-only rail exposing each destination's name via a system tooltip
/// and accessibility label. Both share the same idle / hover / selected /
/// disabled treatments (disabled follows the standard `.disabled(_:)`
/// environment, same as `PVButtonStyle`).
struct PVSidebarNav: View {
    let groupLabel: LocalizedStringResource?
    let items: [PVSidebarNavItem]
    let selection: String
    let collapsed: Bool
    let onSelect: (String) -> Void

    init(
        groupLabel: LocalizedStringResource? = nil,
        items: [PVSidebarNavItem],
        selection: String,
        collapsed: Bool,
        onSelect: @escaping (String) -> Void
    ) {
        self.groupLabel = groupLabel
        self.items = items
        self.selection = selection
        self.collapsed = collapsed
        self.onSelect = onSelect
    }

    var body: some View {
        VStack(alignment: collapsed ? .center : .leading, spacing: collapsed ? PVSpacing.space4 : PVSpacing.space1) {
            if let groupLabel, !collapsed {
                Text(groupLabel)
                    .pvMicroCaps()
                    .foregroundStyle(PVColor.textFaint)
                    .padding(.horizontal, PVSpacing.space4)
                    .padding(.bottom, PVSpacing.space2)
            }
            ForEach(items) { item in
                if collapsed {
                    collapsedItem(item)
                } else {
                    expandedItem(item)
                }
            }
        }
        .padding(collapsed ? PVSpacing.space4 : PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: collapsed ? .center : .leading)
    }

    @ViewBuilder
    private func expandedItem(_ item: PVSidebarNavItem) -> some View {
        let childActive = item.children.contains { $0.id == selection }
        let parentSelected = item.id == selection && !childActive

        VStack(alignment: .leading, spacing: PVSpacing.space2) {
            PVSidebarNavButton(
                item: item,
                isSelected: parentSelected,
                collapsed: false,
                isChild: false,
                action: { onSelect(item.id) }
            )

            if !item.children.isEmpty {
                // Guide rule sits on the leading edge of the indented block
                // (CSS `marginLeft: 21` + `borderLeft`), not in the margin gutter.
                VStack(alignment: .leading, spacing: 1) {
                    ForEach(item.children) { child in
                        PVSidebarNavButton(
                            item: child,
                            isSelected: child.id == selection,
                            collapsed: false,
                            isChild: true,
                            action: { onSelect(child.id) }
                        )
                    }
                }
                .overlay(alignment: .leading) {
                    Rectangle()
                        .fill(childActive ? PVColor.borderDefault : PVColor.borderSubtle)
                        .frame(width: 1)
                }
                .padding(.leading, 21)
            }
        }
    }

    @ViewBuilder
    private func collapsedItem(_ item: PVSidebarNavItem) -> some View {
        let childActive = item.children.contains { $0.id == selection }
        let parentSelected = item.id == selection && !childActive

        VStack(spacing: PVSpacing.space2) {
            PVSidebarNavButton(
                item: item,
                isSelected: parentSelected,
                collapsed: true,
                isChild: false,
                action: { onSelect(item.id) }
            )

            if !item.children.isEmpty {
                Rectangle()
                    .fill(PVColor.borderSubtle)
                    .frame(width: 16, height: 1)

                VStack(spacing: 1) {
                    ForEach(item.children) { child in
                        PVSidebarNavButton(
                            item: child,
                            isSelected: child.id == selection,
                            collapsed: true,
                            isChild: true,
                            action: { onSelect(child.id) }
                        )
                    }
                }
            }
        }
    }
}

/// Sidebar row button. Uses a `ButtonStyle` + `PVHoverEffect` so click,
/// hover, and pressed share one hit-testing surface (modifiers chained
/// after `.buttonStyle(.plain)` do not).
private struct PVSidebarNavButton: View {
    let item: PVSidebarNavItem
    let isSelected: Bool
    let collapsed: Bool
    let isChild: Bool
    let action: () -> Void

    private var iconSize: CGFloat {
        if collapsed {
            return isChild ? 14 : 16
        }
        return isChild ? 14 : 15
    }

    private var hitSize: CGFloat {
        collapsed ? (isChild ? 26 : 32) : 0
    }

    var body: some View {
        Button(action: action) {
            HStack(spacing: isChild ? PVSpacing.space4 : PVSpacing.space5) {
                PVIcon(item.icon, size: iconSize)
                if !collapsed {
                    Text(item.label)
                        .lineLimit(1)
                        .truncationMode(.tail)
                    Spacer(minLength: 0)
                    if !isChild, let count = item.count {
                        Text("\(count)")
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .foregroundStyle(PVColor.textFaint)
                    }
                }
            }
        }
        .buttonStyle(
            PVSidebarNavRowStyle(
                isSelected: isSelected,
                collapsed: collapsed,
                isChild: isChild
            )
        )
        .pvHelp(item.label, when: collapsed)
        .accessibilityLabel(Text(item.label))
        .accessibilityAddTraits(isSelected ? .isSelected : [])
        .accessibilityIdentifier(item.accessibilityIdentifier)
        .pvAccessibilityCount(collapsed || isChild ? nil : item.count)
    }
}

private struct PVSidebarNavRowStyle: ButtonStyle {
    let isSelected: Bool
    let collapsed: Bool
    let isChild: Bool

    func makeBody(configuration: Configuration) -> some View {
        PVSidebarNavRowBody(
            configuration: configuration,
            isSelected: isSelected,
            collapsed: collapsed,
            isChild: isChild
        )
    }
}

private struct PVSidebarNavRowBody: View {
    let configuration: ButtonStyleConfiguration
    let isSelected: Bool
    let collapsed: Bool
    let isChild: Bool

    private var hitSize: CGFloat {
        collapsed ? (isChild ? 26 : 32) : 0
    }

    var body: some View {
        PVHoverEffect(isPressed: configuration.isPressed) { showHover in
            let background: Color = isSelected
                ? PVColor.surfaceSelected
                : (showHover ? PVColor.surfaceHover : .clear)
            let textColor: Color = {
                if isSelected { return PVColor.textPrimary }
                return isChild ? PVColor.textMuted : PVColor.textSecondary
            }()
            let font = PVFont.body(
                size: isChild ? PVTypeScale.caption : PVTypeScale.bodySmall,
                weight: isSelected ? PVFontWeight.semibold : PVFontWeight.regular
            )

            configuration.label
                .frame(width: collapsed ? hitSize : nil, height: collapsed ? hitSize : nil)
                .frame(maxWidth: collapsed ? nil : .infinity, alignment: .leading)
                .font(font)
                .foregroundStyle(textColor)
                .padding(.horizontal, collapsed ? 0 : (isChild ? 10 : PVSpacing.space4))
                .padding(.vertical, collapsed ? 0 : (isChild ? 5 : PVSpacing.space4))
                .background {
                    if isChild, !collapsed {
                        HStack(spacing: 0) {
                            Rectangle()
                                .fill(isSelected ? PVColor.accent : .clear)
                                .frame(width: 2)
                            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                                .fill(background)
                        }
                    } else {
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .fill(background)
                    }
                }
        }
    }
}

private extension View {
    /// Attaches a system tooltip only in `collapsed` mode, where the
    /// destination's label isn't visible as text.
    @ViewBuilder
    func pvHelp(_ text: LocalizedStringResource, when condition: Bool) -> some View {
        if condition {
            help(Text(text))
        } else {
            self
        }
    }

    /// Exposes the trailing count badge to VoiceOver as the row's
    /// accessibility value, not just visual decoration. `nil` (no count,
    /// or collapsed — the badge isn't shown then either) attaches nothing.
    @ViewBuilder
    func pvAccessibilityCount(_ count: Int?) -> some View {
        if let count {
            accessibilityValue(Text("\(count)"))
        } else {
            self
        }
    }
}

#Preview {
    let config = [
        PVSidebarNavItem(
            id: "source-types",
            label: L10n.Workspace.sourceTypesTitle,
            icon: .tag,
            accessibilityIdentifier: "preview.nav.sourceTypes"
        ),
        PVSidebarNavItem(
            id: "source-fields",
            label: L10n.Workspace.sourceFieldsTitle,
            icon: .list,
            accessibilityIdentifier: "preview.nav.sourceFields"
        ),
        PVSidebarNavItem(
            id: "subject-types",
            label: L10n.Workspace.subjectTypesTitle,
            icon: .shapes,
            accessibilityIdentifier: "preview.nav.subjectTypes"
        ),
        PVSidebarNavItem(
            id: "subject-fields",
            label: L10n.Workspace.subjectFieldsTitle,
            icon: .listTree,
            accessibilityIdentifier: "preview.nav.subjectFields"
        ),
    ]
    let items = [
        PVSidebarNavItem(
            id: "sources",
            label: L10n.Workspace.sourcesTitle,
            icon: .library,
            accessibilityIdentifier: "preview.nav.sources",
            count: 12,
            children: config
        ),
    ]
    return HStack(alignment: .top, spacing: PVSpacing.space9) {
        PVSidebarNav(items: items, selection: "sources", collapsed: false, onSelect: { _ in })
            .frame(width: 220)
        PVSidebarNav(items: items, selection: "subject-types", collapsed: true, onSelect: { _ in })
            .frame(width: 78)
    }
    .padding(PVSpacing.space9)
    .background(PVColor.surfaceCard)
}
