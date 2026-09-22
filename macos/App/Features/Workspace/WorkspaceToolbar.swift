import AppKit
import SwiftUI

/// Shared open/anchor state for the Back/Forward jump menu. Owned by
/// `WorkspaceContent` so the panel overlays the full content column (not the
/// clipped 52pt toolbar row) and can receive clicks over the page below.
@Observable
final class HistoryJumpMenuModel {
    static let contentCoordinateSpace = "workspace.content"

    var side: HistoryJumpMenuSide?
    var anchors: [HistoryJumpMenuSide: CGRect] = [:]

    func present(_ side: HistoryJumpMenuSide) {
        self.side = side
    }

    func dismiss() {
        side = nil
    }

    var isPresented: Bool { side != nil }
}

enum HistoryJumpMenuSide: Hashable {
    case back, forward
}

/// App Layout main-column toolbar: Back/Forward (+ jump menus), breadcrumbs, omnibar.
/// See `docs/deployment-plan/archive/spike-3/navigation-history.md` § Toolbar.
struct WorkspaceToolbar: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(OmnibarFocusCoordinator.self) private var omnibarFocus
    var jumpMenu: HistoryJumpMenuModel
    @Bindable var omnibarResults: OmnibarResultsModel
    let projectDir: String
    let store: any GenealogyStore
    @FocusState private var omnibarFocused: Bool

    /// Omnibar width at a comfortable window size, and the floor it
    /// compresses to before the breadcrumbs give up any more room.
    private static let omnibarWidth: CGFloat = 420
    private static let omnibarMinWidth: CGFloat = 200

    var body: some View {
        HStack(alignment: .center, spacing: PVSpacing.space6) {
            navCluster
            breadcrumbs
                .frame(maxWidth: .infinity, alignment: .leading)
                .layoutPriority(0)
            omnibarField
                // Flexible, not a hard 420: a fixed width made the content
                // column's minimum wider than a narrow window, and the
                // overflow was centered — pushing the sidebar (and its logo)
                // off the left edge, under the traffic lights.
                .frame(minWidth: Self.omnibarMinWidth, maxWidth: Self.omnibarWidth)
                .layoutPriority(1)
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .frame(maxWidth: .infinity)
        .onPreferenceChange(HistoryJumpMenuAnchorKey.self) { jumpMenu.anchors = $0 }
        .onPreferenceChange(OmnibarFieldAnchorKey.self) { omnibarResults.fieldFrame = $0 }
        // Drag surface sits *behind* the row: an overlay would swallow
        // Back/Forward, breadcrumb, and omnibar hits, but without one only
        // the system title-bar strip at the very top of the 52pt row drags.
        .pvWorkspaceHeaderRow(windowDrag: .behindContent)
        .onChange(of: omnibarFocus.focusGeneration) { _, _ in
            omnibarFocused = true
        }
        .onChange(of: navigation.currentLocation) { _, _ in
            jumpMenu.dismiss()
        }
        .onChange(of: omnibarResults.query) { _, _ in
            omnibarResults.scheduleSearch(
                projectDir: projectDir,
                location: navigation.currentLocation,
                store: store
            )
        }
        .accessibilityIdentifier("workspace.toolbar")
    }

    private var navCluster: some View {
        HStack(spacing: PVSpacing.space4) {
            historyButton(
                side: .back,
                enabled: navigation.canGoBack,
                symbol: .chevronBack,
                shortcut: "⌘[",
                label: L10n.Workspace.goBack,
                step: { navigation.goBack() },
                items: navigation.backJumpItems()
            )
            historyButton(
                side: .forward,
                enabled: navigation.canGoForward,
                symbol: .chevronForward,
                shortcut: "⌘]",
                label: L10n.Workspace.goForward,
                step: { navigation.goForward() },
                items: navigation.forwardJumpItems()
            )
        }
        .layoutPriority(2)
    }

    private func historyButton(
        side: HistoryJumpMenuSide,
        enabled: Bool,
        symbol: PVSymbol,
        shortcut: String,
        label: LocalizedStringResource,
        step: @escaping () -> Void,
        items: [(index: Int, location: WorkspaceLocation)]
    ) -> some View {
        HistoryNavControl(
            enabled: enabled,
            symbol: symbol,
            shortcut: shortcut,
            label: label,
            accessibilityIdentifier: side == .back
                ? "workspace.toolbar.back"
                : "workspace.toolbar.forward",
            menuPresented: Binding(
                get: { jumpMenu.side == side },
                set: { presented in
                    if presented {
                        jumpMenu.present(side)
                    } else if jumpMenu.side == side {
                        jumpMenu.dismiss()
                    }
                }
            ),
            hasJumpItems: !items.isEmpty,
            onStep: {
                jumpMenu.dismiss()
                step()
            }
        )
        .background {
            GeometryReader { geo in
                Color.clear.preference(
                    key: HistoryJumpMenuAnchorKey.self,
                    value: [
                        side: geo.frame(in: .named(HistoryJumpMenuModel.contentCoordinateSpace)),
                    ]
                )
            }
        }
    }

    private var breadcrumbs: some View {
        PVBreadcrumbs(items: Self.breadcrumbItems(
            for: navigation.currentLocation,
            goToSectionRoot: { section in
                // Guarded here rather than in `PVBreadcrumbs`: the toolbar is
                // the only place crumbs sit on a window-drag surface, and the
                // design system has no business knowing about window chrome.
                WindowDrag.unlessDragging {
                    navigation.go(to: .sectionRoot(section))
                }
            }
        ))
        .lineLimit(1)
        .accessibilityIdentifier("workspace.toolbar.breadcrumbs")
    }

    private var omnibarField: some View {
        PVInput(
            text: $omnibarResults.query,
            size: .sm,
            prompt: L10n.Workspace.omnibarPlaceholder,
            icon: .search,
            suffix: "⌘K",
            focused: $omnibarFocused
        )
        .background {
            GeometryReader { geo in
                Color.clear.preference(
                    key: OmnibarFieldAnchorKey.self,
                    value: geo.frame(in: .named(HistoryJumpMenuModel.contentCoordinateSpace))
                )
            }
        }
        .accessibilityLabel(Text(L10n.Workspace.omnibarAccessibilityLabel))
        .accessibilityIdentifier("workspace.toolbar.omnibar")
    }

    /// Builds toolbar crumbs: section root alone, or section (navigable) + leaf.
    static func breadcrumbItems(
        for location: WorkspaceLocation,
        goToSectionRoot: @escaping (WorkspaceSection) -> Void
    ) -> [PVBreadcrumbItem] {
        let sectionLabel = String(localized: location.section.label)
        let isDeep = location.sourceId != nil || location.fieldId != nil || location.typeId != nil
        if !isDeep {
            return [
                PVBreadcrumbItem(id: "section-\(location.section.rawValue)", label: sectionLabel, action: nil),
            ]
        }
        let leaf: String
        if location.sourceSurface == .citationComposer, location.sourceId != nil {
            let scope = location.title.flatMap { $0.nilIfEmpty }
                ?? location.ref.flatMap { $0.nilIfEmpty }
                ?? "…"
            leaf = L10n.CitationComposer.breadcrumbCitationFor(scope: scope)
        } else if location.sourceSurface == .graph, location.sourceId != nil {
            leaf = String(localized: L10n.Workspace.evidenceGraphTitle)
        } else {
            leaf = location.ref.flatMap { $0.nilIfEmpty }
                ?? location.title.flatMap { $0.nilIfEmpty }
                ?? "…"
        }
        return [
            PVBreadcrumbItem(
                id: "section-\(location.section.rawValue)",
                label: sectionLabel,
                action: { goToSectionRoot(location.section) }
            ),
            PVBreadcrumbItem(id: "leaf-\(location.section.rawValue)", label: leaf, action: nil),
        ]
    }
}

// MARK: - Jump menu host (content-column overlay)

/// Renders and dismisses the jump menu over the workspace content column.
struct HistoryJumpMenuHost: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    var jumpMenu: HistoryJumpMenuModel
    @State private var dismissMonitor: Any?
    @State private var activeIndex = -1

    var body: some View {
        Group {
            if let side = jumpMenu.side,
               let anchor = jumpMenu.anchors[side] {
                let items = Self.rows(for: side, navigation: navigation)
                if !items.isEmpty {
                    HistoryJumpMenuPanel(
                        items: items,
                        activeIndex: activeIndex,
                        onSelect: { index in
                            jumpMenu.dismiss()
                            navigation.go(toIndex: index)
                        }
                    )
                    .offset(x: anchor.minX, y: anchor.maxY + 4)
                }
            }
        }
        .onChange(of: jumpMenu.side) { _, open in
            if open != nil {
                activeIndex = -1
                DispatchQueue.main.async { armDismissMonitor() }
            } else {
                removeDismissMonitor()
            }
        }
        .onDisappear {
            jumpMenu.dismiss()
            removeDismissMonitor()
        }
    }

    private static func rows(
        for side: HistoryJumpMenuSide,
        navigation: WorkspaceNavigation
    ) -> [HistoryJumpRow] {
        let raw: [(index: Int, location: WorkspaceLocation)] = switch side {
        case .back: navigation.backJumpItems()
        case .forward: navigation.forwardJumpItems()
        }
        return raw.map(HistoryJumpRow.init)
    }

    private func armDismissMonitor() {
        removeDismissMonitor()
        // Wait out the opening gesture (long-press release / right-click up)
        // before arming, or the same mouse-up closes the menu immediately.
        if NSEvent.pressedMouseButtons != 0 {
            dismissMonitor = NSEvent.addLocalMonitorForEvents(
                matching: [.leftMouseUp, .rightMouseUp, .otherMouseUp]
            ) { event in
                DispatchQueue.main.async { armDismissMonitor() }
                return event
            }
            return
        }

        dismissMonitor = NSEvent.addLocalMonitorForEvents(
            // Right-click is owned by the Back/Forward catchers (present /
            // switch). Listening for rightMouseDown here raced that present and
            // closed the menu when switching buttons.
            matching: [.leftMouseUp, .keyDown]
        ) { event in
            if event.type == .keyDown {
                return handleKeyDown(event)
            }
            // Dismiss on mouse*Up* (not Down): Button actions fire on mouseUp,
            // and tearing the panel down on mouseDown prevents the item from
            // running.
            DispatchQueue.main.async { jumpMenu.dismiss() }
            return event
        }
    }

    private func handleKeyDown(_ event: NSEvent) -> NSEvent? {
        if event.keyCode == 53 { // Escape
            DispatchQueue.main.async { jumpMenu.dismiss() }
            return nil
        }
        guard let side = jumpMenu.side else { return event }
        let items = Self.rows(for: side, navigation: navigation)
        guard !items.isEmpty else { return event }

        switch event.keyCode {
        case 125: // Down
            DispatchQueue.main.async {
                activeIndex = PVFloatingMenuSelection.moveIndex(
                    from: activeIndex, delta: 1, count: items.count
                )
            }
            return nil
        case 126: // Up
            DispatchQueue.main.async {
                activeIndex = PVFloatingMenuSelection.moveIndex(
                    from: activeIndex, delta: -1, count: items.count
                )
            }
            return nil
        case 36, 76: // Return / keypad Enter
            let highlight = activeIndex
            DispatchQueue.main.async {
                guard items.indices.contains(highlight) else { return }
                jumpMenu.dismiss()
                navigation.go(toIndex: items[highlight].index)
            }
            return nil
        default:
            return event
        }
    }

    private func removeDismissMonitor() {
        if let dismissMonitor {
            NSEvent.removeMonitor(dismissMonitor)
            self.dismissMonitor = nil
        }
    }
}

// MARK: - Jump menu rows

private struct HistoryJumpMenuAnchorKey: PreferenceKey {
    static var defaultValue: [HistoryJumpMenuSide: CGRect] { [:] }

    static func reduce(
        value: inout [HistoryJumpMenuSide: CGRect],
        nextValue: () -> [HistoryJumpMenuSide: CGRect]
    ) {
        value.merge(nextValue(), uniquingKeysWith: { _, new in new })
    }
}

private struct HistoryJumpRow: Identifiable {
    let index: Int
    let location: WorkspaceLocation

    var id: Int { index }

    var icon: PVSymbol { location.section.icon }

    var rootLabel: String {
        String(localized: location.section.label)
    }

    var leafLabel: String? {
        if let title = location.title?.nilIfEmpty { return title }
        if let ref = location.ref?.nilIfEmpty { return ref }
        return nil
    }

    var refLabel: String {
        location.ref?.nilIfEmpty ?? ""
    }

    init(index: Int, location: WorkspaceLocation) {
        self.index = index
        self.location = location
    }
}

// MARK: - Back / Forward control

private struct HistoryNavControl: View {
    let enabled: Bool
    let symbol: PVSymbol
    let shortcut: String
    let label: LocalizedStringResource
    let accessibilityIdentifier: String
    @Binding var menuPresented: Bool
    let hasJumpItems: Bool
    let onStep: () -> Void

    @State private var suppressStep = false
    @State private var hovering = false

    var body: some View {
        Button {
            // A press that moved the window was a window drag that happened
            // to start here, not a click on Back/Forward.
            WindowDrag.unlessDragging {
                if suppressStep {
                    suppressStep = false
                    return
                }
                guard enabled else { return }
                onStep()
            }
        } label: {
            HStack(spacing: 3) {
                PVIcon(symbol, size: 14)
                Text(shortcut)
                    .font(PVFont.mono(size: 10))
                    .accessibilityHidden(true)
            }
            .foregroundStyle(enabled ? (hovering ? PVColor.textPrimary : PVColor.textSecondary) : PVColor.textFaint)
            .padding(.horizontal, 6)
            .frame(height: 26)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(hovering && enabled ? PVColor.surfaceHover : Color.clear)
            )
        }
        .buttonStyle(.plain)
        .disabled(!enabled)
        .opacity(enabled ? 1 : 0.45)
        .help(Text(label))
        .accessibilityLabel(Text(label))
        .accessibilityIdentifier(accessibilityIdentifier)
        .onHover { hovering = $0 }
        .simultaneousGesture(
            LongPressGesture(minimumDuration: 0.4)
                .onEnded { _ in
                    // Both statements stay inside the guard: skipping the
                    // menu but still arming `suppressStep` would eat the
                    // user's next real click.
                    WindowDrag.unlessDragging {
                        guard enabled, hasJumpItems else { return }
                        suppressStep = true
                        menuPresented = true
                    }
                }
        )
        .overlay {
            PVRightClickCatcher { _ in
                guard enabled, hasJumpItems else { return }
                menuPresented = true
            }
        }
    }
}

private struct HistoryJumpMenuPanel: View {
    let items: [HistoryJumpRow]
    var activeIndex: Int
    let onSelect: (Int) -> Void

    var body: some View {
        PVContextMenuPanel(
            width: 360,
            accessibilityIdentifier: "workspace.toolbar.jumpMenu"
        ) {
            ForEach(Array(items.enumerated()), id: \.element.id) { offset, item in
                Button {
                    onSelect(item.index)
                } label: {
                    HStack(alignment: .center, spacing: PVSpacing.space5) {
                        PVIcon(item.icon, size: 15)
                            .foregroundStyle(PVColor.textSecondary)
                            .frame(width: 15)
                        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
                            Text(item.rootLabel)
                                .font(PVFont.body(size: PVTypeScale.caption))
                                .foregroundStyle(PVColor.textMuted)
                                .lineLimit(1)
                                .layoutPriority(0)
                            if let leafLabel = item.leafLabel {
                                Text(verbatim: "›")
                                    .font(PVFont.body(size: PVTypeScale.caption))
                                    .foregroundStyle(PVColor.textFaint)
                                Text(verbatim: leafLabel)
                                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                                    .foregroundStyle(PVColor.textPrimary)
                                    .lineLimit(1)
                                    .truncationMode(.tail)
                                    .layoutPriority(1)
                            }
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        if !item.refLabel.isEmpty {
                            Text(item.refLabel)
                                .font(PVFont.mono(size: PVTypeScale.micro))
                                .foregroundStyle(PVColor.textFaint)
                                .fixedSize(horizontal: true, vertical: false)
                        }
                    }
                    .padding(.horizontal, 9)
                    .padding(.vertical, 6)
                    .contentShape(Rectangle())
                }
                .buttonStyle(HistoryJumpRowButtonStyle(isSelected: offset == activeIndex))
            }
        }
        .accessibilityLabel(Text(L10n.Workspace.jumpMenuAccessibilityLabel))
    }
}

private struct HistoryJumpRowButtonStyle: ButtonStyle {
    var isSelected: Bool = false

    func makeBody(configuration: Configuration) -> some View {
        HistoryJumpRowButtonBody(configuration: configuration, isSelected: isSelected)
    }
}

private struct HistoryJumpRowButtonBody: View {
    let configuration: ButtonStyleConfiguration
    var isSelected: Bool
    @State private var hovering = false

    var body: some View {
        configuration.label
            .background {
                RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                    .fill(rowFill)
            }
            .contentShape(Rectangle())
            .onHover { hovering = $0 }
            .pvAnimation(PVMotion.instantStandard, value: hovering)
    }

    private var rowFill: Color {
        if configuration.isPressed { return PVColor.surfaceActive }
        if isSelected { return PVColor.surfaceSelected }
        if hovering { return PVColor.surfaceHover }
        return .clear
    }
}

private extension String {
    var nilIfEmpty: String? {
        let trimmed = trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}

private extension Optional where Wrapped == String {
    var nilIfEmpty: String? {
        self?.nilIfEmpty
    }
}

#Preview {
    WorkspaceToolbar(
        jumpMenu: HistoryJumpMenuModel(),
        omnibarResults: OmnibarResultsModel(),
        projectDir: "/tmp/preview.provenencia",
        store: FakeStore()
    )
        .environment(WorkspaceNavigation())
        .environment(OmnibarFocusCoordinator())
        .frame(width: 900)
        .background(PVColor.surfacePage)
}
