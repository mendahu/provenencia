import Foundation
import Observation

/// State for the post-onboarding app workspace: which top-level
/// destination is selected, sidebar collapse, and first-class navigation
/// history (`go(to:)` / Back / Forward). See
/// `docs/deployment-plan/spike-3/navigation-history.md`.
///
/// Nav-row counts live on `CatalogCounts` (environment), not here — so
/// vocabulary panes can publish totals without talking to workspace chrome.
@MainActor
@Observable
final class WorkspaceModel {
    /// Spike 2's top-level destinations (W-13, W-16), plus Files — added to
    /// the design board after the written brief, alongside Source content
    /// (its Artifact ingest / file preview arrives in S2-03). Raw values
    /// match the design board's kebab-case section ids.
    enum Section: String, CaseIterable {
        case sources
        case sourceTypes = "source-types"
        case sourceFields = "source-fields"
        case files

        var label: LocalizedStringResource {
            switch self {
            case .sources: L10n.Workspace.sourcesTitle
            case .sourceTypes: L10n.Workspace.sourceTypesTitle
            case .sourceFields: L10n.Workspace.sourceFieldsTitle
            case .files: L10n.Workspace.filesTitle
            }
        }

        var placeholderNote: LocalizedStringResource {
            // Only Files still uses the workspace placeholder host; Sources /
            // types / fields mount real destinations.
            L10n.Workspace.filesPlaceholderNote
        }

        var icon: PVSymbol {
            switch self {
            case .sources: .library
            case .sourceTypes: .tag
            case .sourceFields: .list
            case .files: .folderOpen
            }
        }
    }

    /// W-5: the researcher chooses this; it's never derived from window
    /// width. Persisted per install via `UserDefaults`, as the design brief
    /// suggests.
    private static let sidebarCollapsedDefaultsKey = "pv.sidebarCollapsed"

    /// Derived from history; updated by `go` / Back / Forward / apply.
    private(set) var selectedSection: Section
    /// Current history leaf (section + deep ids). Views apply deep state from this.
    private(set) var currentLocation: WorkspaceLocation
    var isSidebarCollapsed: Bool

    private let defaults: UserDefaults
    private var history: NavigationHistoryStore?
    private var projectUuid: String = ""

    var canGoBack: Bool { history?.canGoBack ?? false }
    var canGoForward: Bool { history?.canGoForward ?? false }

    init(
        selectedSection: Section = .sources,
        defaults: UserDefaults = .standard
    ) {
        self.selectedSection = selectedSection
        self.currentLocation = .sectionRoot(selectedSection)
        self.defaults = defaults
        isSidebarCollapsed = defaults.bool(forKey: Self.sidebarCollapsedDefaultsKey)
    }

    /// Load or create persisted history for this catalog project UUID.
    func attachProject(
        uuid: String,
        fileURL: URL? = nil,
        fileManager: FileManager = .default
    ) {
        let trimmed = uuid.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        if trimmed == projectUuid, history != nil { return }
        projectUuid = trimmed
        let url: URL
        if let fileURL {
            url = fileURL
        } else {
            do {
                url = try InstallPaths.navigationFile(projectUuid: trimmed, fileManager: fileManager)
            } catch {
                #if DEBUG
                assertionFailure("navigation file URL failed: \(error)")
                #endif
                return
            }
        }
        let store = NavigationHistoryStore(
            projectUuid: trimmed,
            fileURL: url,
            fileManager: fileManager,
            seed: .sectionRoot(.sources)
        )
        history = store
        apply(store.current)
    }

    func toggleSidebarCollapsed() {
        isSidebarCollapsed.toggle()
        defaults.set(isSidebarCollapsed, forKey: Self.sidebarCollapsedDefaultsKey)
    }

    /// Committed navigation. Coalesces identical locations; truncates forward.
    func go(to location: WorkspaceLocation) {
        guard let history else {
            apply(location)
            return
        }
        apply(history.go(to: location))
    }

    func goBack() {
        guard let history, let location = history.goBack() else { return }
        apply(location)
    }

    func goForward() {
        guard let history, let location = history.goForward() else { return }
        apply(location)
    }

    func go(toIndex index: Int) {
        guard let history, let location = history.go(toIndex: index) else { return }
        apply(location)
    }

    /// When a deep id is missing after catalog load, land on the section list
    /// and rewrite the current stack entry so persistence stays honest.
    func fallbackToSectionRoot() {
        let root = WorkspaceLocation.sectionRoot(currentLocation.section)
        guard let history else {
            apply(root)
            return
        }
        apply(history.replaceCurrent(with: root))
    }

    private func apply(_ location: WorkspaceLocation) {
        currentLocation = location
        selectedSection = location.section
    }
}
