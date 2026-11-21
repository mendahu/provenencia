import SwiftUI

enum HistoryJumpMenuSide: Hashable {
    case back, forward
}

/// App Layout main-column toolbar: Back/Forward (+ jump menus), breadcrumbs, omnibar.
/// See `docs/deployment-plan/archive/spike-3/navigation-history.md` § Toolbar.
struct WorkspaceToolbar: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(OmnibarFocusCoordinator.self) private var omnibarFocus
    @Environment(WorkspaceSession.self) private var session
    @Bindable var omnibarResults: OmnibarResultsModel
    let projectDir: String
    let store: any GenealogyStore
    /// The content column's width, for the omnibar list's and the history menu's width rules.
    let contentWidth: CGFloat
    @FocusState private var omnibarFocused: Bool
    /// Which side's history menu is open, if either.
    @State private var jumpMenuSide: HistoryJumpMenuSide?

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
        // Drag surface sits *behind* the row: an overlay would swallow
        // Back/Forward, breadcrumb, and omnibar hits, but without one only
        // the system title-bar strip at the very top of the 52pt row drags.
        .pvWorkspaceHeaderRow(windowDrag: .behindContent)
        .onChange(of: omnibarFocus.focusGeneration) { _, _ in
            omnibarFocused = true
        }
        .onChange(of: navigation.currentLocation) { _, _ in
            jumpMenuSide = nil
        }
        // Losing focus (Tab, a click that takes focus) closes the results and
        // keeps the query, as `PVComboBox` does; a click on a hit row does not
        // move focus, the popup being a window that never becomes key.
        .onChange(of: omnibarFocused) { _, focused in
            if !focused { omnibarResults.closePanel() }
        }
        .onChange(of: omnibarResults.search.query) { _, _ in
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
        let menuPresented = Binding(
            get: { jumpMenuSide == side },
            set: { presented in
                if presented {
                    omnibarResults.closePanel()
                    jumpMenuSide = side
                } else if jumpMenuSide == side {
                    jumpMenuSide = nil
                }
            }
        )
        return HistoryNavControl(
            enabled: enabled,
            symbol: symbol,
            shortcut: shortcut,
            label: label,
            accessibilityIdentifier: side == .back
                ? "workspace.toolbar.back"
                : "workspace.toolbar.forward",
            menuPresented: menuPresented,
            hasJumpItems: !items.isEmpty,
            onStep: {
                jumpMenuSide = nil
                step()
            }
        )
        .modifier(HistoryJumpMenuHost(items: items, presented: menuPresented, contentWidth: contentWidth))
    }

    private var breadcrumbs: some View {
        PVBreadcrumbs(items: Self.breadcrumbItems(
            for: navigation.currentLocation,
            names: LocationNames(session: session),
            goTo: { location in
                WindowDrag.unlessDragging {
                    navigation.go(to: location)
                }
            }
        ))
        .lineLimit(1)
        .accessibilityIdentifier("workspace.toolbar.breadcrumbs")
    }

    private var omnibarField: some View {
        PVInput(
            text: Bindable(omnibarResults.search).query,
            size: .sm,
            prompt: L10n.Workspace.omnibarPlaceholder,
            icon: .search,
            suffix: "⌘K",
            focused: $omnibarFocused
        )
        .modifier(OmnibarResultsHost(
            results: omnibarResults,
            projectDir: projectDir,
            contentWidth: contentWidth,
            onActivate: { omnibarResults.activate($0, navigation: navigation) }
        ))
        .accessibilityLabel(Text(L10n.Workspace.omnibarAccessibilityLabel))
        .accessibilityIdentifier("workspace.toolbar.omnibar")
    }

    /// Builds toolbar crumbs.
    /// Graph (S12-11): `Sources › {Source title} › {graph name}`; the Source
    /// crumb goes to the Source page.
    /// Composer (S7-D4): `Sources › {Source title} › {graph name} › Citation for {scope}`
    /// with a tappable graph segment back to the graph's canvas.
    /// Promote (S9-11): `Sources › {Source title} › {graph name} › Promote {ref}`, the same way.
    static func breadcrumbItems(
        for location: WorkspaceLocation,
        names: LocationNames = .none,
        goTo: @escaping (WorkspaceLocation) -> Void
    ) -> [PVBreadcrumbItem] {
        let sectionLabel = L10n.string(location.section.label)
        let isDeep = location.sourceId != nil || location.fieldId != nil || location.typeId != nil
            || location.subjectTypeKey != nil || location.propertyId != nil || location.entityId != nil
        if !isDeep {
            return [
                PVBreadcrumbItem(id: "section-\(location.section.rawValue)", label: sectionLabel, action: nil),
            ]
        }

        var items: [PVBreadcrumbItem] = [
            PVBreadcrumbItem(
                id: "section-\(location.section.rawValue)",
                label: sectionLabel,
                action: { goTo(.sectionRoot(location.section)) }
            ),
        ]

        if let graphID = location.graphSurfaceID, let sourceID = location.sourceId {
            let onCanvas = location.sourceSurface == .graph
            let carriedTitle = onCanvas ? location.title : location.sourceTitle
            let sourceLabel = WorkspaceLocationLabels.sourceTitle(sourceID: sourceID, carried: carriedTitle, names: names)
            if let sourceLabel {
                items.append(
                    PVBreadcrumbItem(
                        id: "source-\(sourceID)",
                        label: sourceLabel,
                        action: { goTo(.sourcePage(sourceId: sourceID, title: sourceLabel)) }
                    )
                )
            }
            items.append(
                PVBreadcrumbItem(
                    id: "graph-\(graphID)",
                    label: WorkspaceLocationLabels.graphLabel(graphID: graphID, names: names),
                    action: onCanvas ? nil : {
                        goTo(.evidenceGraph(sourceId: sourceID, graphId: graphID, title: sourceLabel))
                    }
                )
            )
            if onCanvas { return items }
            let surface = WorkspaceLocationLabels.surfaceLabel(for: location) ?? ""
            items.append(
                PVBreadcrumbItem(
                    id: location.sourceSurface == .promote ? "promote-\(graphID)" : "composer-\(graphID)",
                    label: surface,
                    action: nil
                )
            )
            return items
        }

        let leaf: String
        if location.sourceSurface == .graph, location.sourceId != nil {
            leaf = L10n.string(L10n.Workspace.evidenceGraphTitle)
        } else if location.typeId != nil {
            leaf = WorkspaceLocationLabels.labels(for: location, names: names).leaf ?? "…"
        } else if location.entityId != nil {
            let labels = WorkspaceLocationLabels.labels(for: location)
            leaf = labels.leaf ?? "…"
            if let section = labels.section {
                // A section of the page: the record's crumb goes back to its first
                // section.
                items.append(
                    PVBreadcrumbItem(
                        id: "entity-\(location.section.rawValue)",
                        label: leaf,
                        action: { goTo(location.withPageSection(nil)) }
                    )
                )
                items.append(
                    PVBreadcrumbItem(id: "leaf-\(location.section.rawValue)", label: section, action: nil)
                )
                return items
            }
        } else {
            leaf = location.ref.flatMap { $0.trimmedNonEmpty }
                ?? location.title.flatMap { $0.trimmedNonEmpty }
                ?? "…"
        }
        items.append(
            PVBreadcrumbItem(id: "leaf-\(location.section.rawValue)", label: leaf, action: nil)
        )
        return items
    }
}

// MARK: - Jump menu host

/// Presents one history button's jump menu through `.pvContextMenu`, in the
/// kit menu window under the button. Right-clicks pass: they switch between
/// Back and Forward.
private struct HistoryJumpMenuHost: ViewModifier {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(WorkspaceSession.self) private var session
    let items: [(index: Int, location: WorkspaceLocation)]
    let presented: Binding<Bool>
    let contentWidth: CGFloat
    /// The button's frame in the content column.
    @State private var anchor: CGRect?
    @State private var keyboard = PVContextMenuKeyboard.inactive

    /// The panel's preferred width.
    static let panelWidth: CGFloat = 600

    func body(content: Content) -> some View {
        content
            .onGeometryChange(for: CGRect.self) {
                $0.frame(in: .named(WorkspaceContent.coordinateSpace))
            } action: {
                anchor = $0
            }
            .pvContextMenu(
                state,
                keyboard: $keyboard,
                options: PVPopupOptions(ignoresRightClicks: true)
            ) {
                menu
            }
            .onChange(of: presented.wrappedValue) { _, open in
                if open {
                    keyboard = PVContextMenuKeyboard(itemCount: items.count, activeIndex: -1)
                }
            }
    }

    private var state: Binding<PVContextMenuState> {
        Binding(
            get: {
                PVContextMenuState(
                    isPresented: presented.wrappedValue,
                    origin: CGPoint(x: 0, y: (anchor?.height ?? 0) + PVSpacing.space2)
                )
            },
            set: { presented.wrappedValue = $0.isPresented }
        )
    }

    /// Nothing while closed: the kit presenter builds its content on every render.
    @ViewBuilder
    private var menu: some View {
        if presented.wrappedValue, let anchor {
            HistoryJumpMenuPanel(
                items: Self.rows(items, names: LocationNames(session: session)),
                width: PVPopupListPlacement.fittedWidth(
                    preferred: Self.panelWidth, hangsFrom: anchor.minX,
                    alignment: .leading, columnWidth: contentWidth
                ),
                onSelect: { navigation.go(toIndex: $0) }
            )
        }
    }

    @MainActor
    static func rows(
        _ items: [(index: Int, location: WorkspaceLocation)],
        names: LocationNames
    ) -> [HistoryJumpRow] {
        items.map { row in
            let labels = WorkspaceLocationLabels.labels(for: row.location, names: names)
            let leaf = labels.leaf.flatMap { leaf in
                labels.section.map { L10n.Conclusions.pageSectionLeaf(leaf, $0) } ?? leaf
            }
            return HistoryJumpRow(index: row.index, location: row.location, leafLabel: leaf)
        }
    }
}

// MARK: - Jump menu rows

private struct HistoryJumpRow: Identifiable {
    let index: Int
    let location: WorkspaceLocation

    var id: Int { index }

    var glyph: PVGlyph { location.section.glyph }

    var rootLabel: String {
        L10n.string(location.section.label)
    }

    let leafLabel: String?

    var refLabel: String {
        location.ref?.trimmedNonEmpty ?? ""
    }

    init(index: Int, location: WorkspaceLocation, leafLabel: String?) {
        self.index = index
        self.location = location
        self.leafLabel = leafLabel
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
    let width: CGFloat
    let onSelect: (Int) -> Void

    var body: some View {
        PVContextMenuPanel(
            width: width,
            accessibilityIdentifier: "workspace.toolbar.jumpMenu"
        ) {
            ForEach(Array(items.enumerated()), id: \.element.id) { offset, item in
                PVContextMenuItem(index: offset, action: { onSelect(item.index) }) {
                    HistoryJumpRowLabel(item: item)
                }
            }
        }
        .accessibilityLabel(Text(L10n.Workspace.jumpMenuAccessibilityLabel))
    }
}

/// One history row's content: glyph, section root, `›`, leaf and ref. The kit
/// menu item around it supplies padding, highlight and dismiss.
private struct HistoryJumpRowLabel: View {
    let item: HistoryJumpRow

    var body: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            PVGlyphView(item.glyph, size: 15)
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
                        .help(Text(verbatim: leafLabel))
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
    }
}

#Preview {
    WorkspaceToolbar(
        omnibarResults: OmnibarResultsModel(),
        projectDir: "/tmp/preview.provenencia",
        store: FakeStore(),
        contentWidth: 900
    )
        .environment(WorkspaceNavigation())
        .environment(WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: FakeStore()))
        .environment(OmnibarFocusCoordinator())
        .frame(width: 900)
        .background(PVColor.surfacePage)
}
