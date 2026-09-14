import Foundation
import Observation

/// First-class workspace navigation history (`go(to:)` / Back / Forward).
/// Owned by `WorkspaceView` and injected via `.environment` so any destination
/// can commit or apply places without talking to sidebar chrome.
/// See `docs/deployment-plan/spike-3/navigation-history.md`.
@MainActor
@Observable
final class WorkspaceNavigation {
    /// Derived from history; updated by `go` / Back / Forward / apply.
    private(set) var selectedSection: WorkspaceSection
    /// Current history leaf (section + deep ids). Views apply deep state from this.
    private(set) var currentLocation: WorkspaceLocation
    /// Stored (not computed) so `@Observable` publishes when the stack index moves.
    private(set) var canGoBack = false
    private(set) var canGoForward = false

    private var history: NavigationHistoryStore?
    private var projectUuid: String = ""

    init(selectedSection: WorkspaceSection = .sources) {
        self.selectedSection = selectedSection
        self.currentLocation = .sectionRoot(selectedSection)
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

    /// Nearest history entries before current (nearest first) for the Back jump menu.
    func backJumpItems(limit: Int = 15) -> [(index: Int, location: WorkspaceLocation)] {
        history?.backJumpItems(limit: limit) ?? []
    }

    /// History entries after current (stack order) for the Forward jump menu.
    func forwardJumpItems(limit: Int = 15) -> [(index: Int, location: WorkspaceLocation)] {
        history?.forwardJumpItems(limit: limit) ?? []
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
        canGoBack = history?.canGoBack ?? false
        canGoForward = history?.canGoForward ?? false
    }
}
