import Foundation

/// One titled sidebar section: a `PVSidebarNav` group (S9-D1).
struct WorkspaceSidebarSection: Equatable {
    enum Placement: Equatable {
        /// Research sections, stacked from the top.
        case top
        /// Configuration, bottom-aligned above the session footer.
        case bottom
    }

    let id: String
    let title: LocalizedStringResource
    let placement: Placement
    let items: [PVSidebarNavItem]
}

/// The sidebar's sections, in order (S9-D1): **Source** and **Conclude** at
/// the top, **Configure** at the bottom. Every destination is a top-level row.
/// Research rows carry counts (zero reads `0`); Configure rows carry none — the
/// missing count is one of the cues that separates the halves.
@MainActor
enum WorkspaceSidebarSections {
    static func make(counts: CatalogCounts) -> [WorkspaceSidebarSection] {
        [
            WorkspaceSidebarSection(
                id: "source",
                title: L10n.Workspace.sidebarSourceTitle,
                placement: .top,
                items: [.sources].map { item($0, count: counts.badge(for: $0)) }
            ),
            WorkspaceSidebarSection(
                id: "conclude",
                title: L10n.Workspace.sidebarConcludeTitle,
                placement: .top,
                items: [.persons, .events, .places].map { item($0, count: counts.badge(for: $0)) }
            ),
            // Narrate goes here (after Conclude, top) once the Narrative layer
            // gives it a destination; it takes its height from the flexible
            // space, so nothing else moves. Hidden until then.
            WorkspaceSidebarSection(
                id: "configure",
                title: L10n.Workspace.sidebarConfigureTitle,
                placement: .bottom,
                items: [.sourceTypes, .metadata, .properties].map { item($0, count: nil) }
            ),
        ]
    }

    private static func item(_ section: WorkspaceSection, count: Int?) -> PVSidebarNavItem {
        PVSidebarNavItem(
            id: section.rawValue,
            label: section.label,
            icon: section.icon,
            accessibilityIdentifier: "workspace.nav.\(section.rawValue)",
            count: count
        )
    }
}
