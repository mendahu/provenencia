import SwiftUI

/// The workspace's leading sidebar: brand mark, titled nav sections (S9-D1:
/// Source and Conclude at the top, Configure bottom-aligned above the
/// footer), and a session-identity + collapse-toggle footer.
/// Workspace chrome: `docs/deployment-plan/archive/spike-2/README.md`.
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

    private var sections: [WorkspaceSidebarSection] {
        WorkspaceSidebarSections.make(counts: catalogCounts)
    }

    /// Minimum gap between the research sections and Configure (S9-D1): the
    /// flexible space collapses to this on a short window.
    private let sectionGapFloor: CGFloat = 24

    var body: some View {
        GeometryReader { geometry in
            ScrollView {
                VStack(spacing: 0) {
                    brandHeader
                    let top = sections.filter { $0.placement == .top }
                    ForEach(Array(top.enumerated()), id: \.element.id) { index, section in
                        if index > 0 && workspace.isSidebarCollapsed {
                            sectionRule
                        }
                        nav(section)
                    }
                    Spacer(minLength: sectionGapFloor)
                    ForEach(sections.filter { $0.placement == .bottom }, id: \.id) { section in
                        sectionRule
                        nav(section)
                    }
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

    private func nav(_ section: WorkspaceSidebarSection) -> some View {
        PVSidebarNav(
            groupLabel: section.title,
            items: section.items,
            selection: navigation.selectedSection.rawValue,
            collapsed: workspace.isSidebarCollapsed,
            onSelect: { id in
                guard let section = WorkspaceSection(id: id) else { return }
                navigation.go(to: .sectionRoot(section))
            }
        )
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(section.title))
        .accessibilityIdentifier("workspace.sidebar.section.\(section.id)")
    }

    /// The hairline between sections (S9-D1): inset to the rows when
    /// expanded, a short 24pt mark centred on the collapsed rail, where it
    /// stands in for the dropped titles.
    @ViewBuilder
    private var sectionRule: some View {
        if workspace.isSidebarCollapsed {
            PVDivider()
                .frame(width: 24)
                .padding(.vertical, PVSpacing.space1)
                .frame(maxWidth: .infinity)
                .accessibilityHidden(true)
        } else {
            PVDivider()
                .padding(.horizontal, PVSpacing.space5)
                .accessibilityHidden(true)
        }
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
                if let name = session?.displayName, !name.isEmpty {
                    Text(verbatim: name)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                }
                if let ref = session?.ref, !ref.isEmpty {
                    Text(verbatim: ref)
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
