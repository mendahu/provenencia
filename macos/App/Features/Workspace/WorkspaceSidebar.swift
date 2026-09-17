import SwiftUI

/// The workspace's leading sidebar: brand mark, primary nav destinations,
/// and a session-identity + collapse-toggle footer. Requirements W-1, W-2,
/// W-4, W-5, W-5b, W-7, W-8, W-13, W-16, W-19, W-20 — see
/// `docs/deployment-plan/archive/spike-2/design/archive/S2-01-workspace-chrome.md`.
///
/// The whole column (brand row, nav, footer) sits in one `ScrollView` so a
/// short window scrolls the sidebar instead of clipping destinations or
/// the session footer (W-5b), while a tall window still pins the footer to
/// the bottom via the `GeometryReader`-driven `minHeight`.
struct WorkspaceSidebar: View {
    let session: InstallIdentity?
    @Bindable var workspace: WorkspaceModel
    @Environment(WorkspaceNavigation.self) private var navigation
    /// `@Bindable` so badge publishes from Fields/Types invalidate this
    /// column — a plain `let` does not subscribe to `@Observable` writes.
    @Bindable var catalogCounts: CatalogCounts

    /// Matches `PVSpacing.widthSidebar` (264pt, the shared `--width-sidebar`
    /// token).
    private let expandedWidth = PVSpacing.widthSidebar
    /// No source token for the collapsed rail width: the stoplight
    /// cluster's rightmost edge (native right edge ~70pt, plus
    /// `TrafficLights.horizontalPadding`) plus a real trailing margin
    /// before the rail's own divider — not "the board's 78pt figure plus
    /// the padding," which just carries the board's already-tight, ~no
    /// margin figure along with the shifted lights instead of adding room
    /// past them.
    private let collapsedWidth: CGFloat = 96
    /// Leading padding before the logo, clearing the stoplights (shifted
    /// right by `TrafficLights.horizontalPadding`) with some margin
    /// (`ProvenenciaApp` hides the title bar, so this row's own content
    /// sits behind them — see Frame 7).
    private let trafficLightLeadingInset: CGFloat = 100

    private var items: [PVSidebarNavItem] {
        let configSections: [WorkspaceSection] = [
            .sourceTypes, .sourceFields, .subjectTypes, .subjectFields,
        ]
        let configChildren = configSections.map { section in
            PVSidebarNavItem(
                id: section.rawValue,
                label: section.label,
                icon: section.icon,
                accessibilityIdentifier: "workspace.nav.\(section.rawValue)"
            )
        }
        return [
            PVSidebarNavItem(
                id: WorkspaceSection.sources.rawValue,
                label: WorkspaceSection.sources.label,
                icon: WorkspaceSection.sources.icon,
                accessibilityIdentifier: "workspace.nav.sources",
                count: catalogCounts.badge(for: .sources),
                children: configChildren
            ),
        ]
    }

    var body: some View {
        GeometryReader { geometry in
            ScrollView {
                VStack(spacing: 0) {
                    brandHeader
                    PVSidebarNav(
                        items: items,
                        selection: navigation.selectedSection.rawValue,
                        collapsed: workspace.isSidebarCollapsed,
                        onSelect: { id in
                            guard let section = WorkspaceSection(rawValue: id) else { return }
                            navigation.go(to: .sectionRoot(section))
                        }
                    )
                    Spacer(minLength: 0)
                    footer
                }
                .frame(minHeight: geometry.size.height)
            }
        }
        .frame(width: workspace.isSidebarCollapsed ? collapsedWidth : expandedWidth)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .trailing) {
            PVDivider(axis: .vertical)
        }
        .pvAnimation(PVMotion.easeStandard, value: workspace.isSidebarCollapsed)
        .accessibilityIdentifier("workspace.sidebar")
    }

    @ViewBuilder
    private var brandHeader: some View {
        if workspace.isSidebarCollapsed {
            VStack(spacing: 0) {
                // Reserves the traffic lights their own row — the rail is
                // barely wider than the lights themselves, so an inline
                // logo doesn't fit beside them.
                Color.clear
                    .accessibilityHidden(true)
                    .pvWorkspaceHeaderRow()
                PVLogoMark(size: 18)
                    .padding(.top, PVSpacing.space5)
                    .padding(.bottom, PVSpacing.space2)
                    .frame(maxWidth: .infinity)
            }
        } else {
            HStack(spacing: PVSpacing.space4) {
                PVLogoMark(size: 18)
                Text(verbatim: "Provenencia")
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textDisplay)
                    .lineLimit(1)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.leading, trafficLightLeadingInset)
            .padding(.trailing, PVSpacing.space5)
            .pvWorkspaceHeaderRow()
        }
    }

    @ViewBuilder
    private var footer: some View {
        VStack(spacing: 0) {
            PVDivider()
            if workspace.isSidebarCollapsed {
                collapsedFooter
            } else {
                expandedFooter
            }
        }
        .background(PVColor.surfaceSunken)
    }

    private var contributorAccessibilityLabel: String {
        L10n.Onboarding.contributorOption(displayName: session?.displayName ?? "", ref: session?.ref ?? "")
    }

    @ViewBuilder
    private var expandedFooter: some View {
        HStack(spacing: PVSpacing.space4) {
            VStack(alignment: .leading, spacing: 1) {
                Text(session?.displayName ?? "")
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(1)
                if let ref = session?.ref, !ref.isEmpty {
                    Text(ref)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                }
            }
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(contributorAccessibilityLabel))
            .accessibilityIdentifier("workspace.sidebar.contributor")
            Spacer(minLength: 0)
            collapseToggle(label: L10n.Workspace.collapseSidebar)
        }
        .padding(.horizontal, PVSpacing.space5)
        .padding(.vertical, PVSpacing.space4)
    }

    @ViewBuilder
    private var collapsedFooter: some View {
        VStack(spacing: PVSpacing.space5) {
            PVIcon(.account, size: 22)
                .foregroundStyle(PVColor.textSecondary)
                .accessibilityElement(children: .ignore)
                .accessibilityLabel(Text(contributorAccessibilityLabel))
                .accessibilityIdentifier("workspace.sidebar.contributor")
                .help(Text(contributorAccessibilityLabel))
            collapseToggle(label: L10n.Workspace.expandSidebar)
        }
        .padding(.vertical, PVSpacing.space5)
    }

    private func collapseToggle(label: LocalizedStringResource) -> some View {
        // `PVIconButton` now owns the tooltip and accessibility label.
        PVIconButton(.sidebarToggle, label: label) {
            workspace.toggleSidebarCollapsed()
        }
        .accessibilityIdentifier("workspace.sidebar.collapseToggle")
    }
}
