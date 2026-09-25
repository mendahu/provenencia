import Foundation
import Observation

enum PendingNavigation: Equatable {
    case location(WorkspaceLocation)
    case back
    case forward
    case index(Int)
}

@MainActor
protocol WorkspaceLeaveGuard: AnyObject {
    /// Return true to hold this navigation. The guard must later call
    /// `resumeHeldNavigation()` or `cancelHeldNavigation()`.
    func shouldHoldNavigation(_ pending: PendingNavigation) -> Bool
}

/// First-class workspace navigation history (`go(to:)` / Back / Forward).
/// Owned by `WorkspaceView` and injected via `.environment` so any destination
/// can commit or apply places without talking to sidebar chrome.
/// See `docs/deployment-plan/archive/spike-3/navigation-history.md`.
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
    /// Latest load/persist failure for chrome toast; cleared via `acknowledgeHistoryIssue()`.
    private(set) var lastHistoryIssue: NavigationHistoryIssue?
    /// Fired synchronously whenever a location is committed (including history restore).
    var onLocationCommit: ((WorkspaceLocation) -> Void)?
    /// Views never observe this; it only intercepts committed moves.
    @ObservationIgnored
    weak var leaveGuard: (any WorkspaceLeaveGuard)?
    private(set) var heldNavigation: PendingNavigation?

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
                lastHistoryIssue = .persistFailed
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
        refreshHistoryIssue()
    }

    /// Clears a surfaced history issue so it is not shown again until a new failure.
    func acknowledgeHistoryIssue() {
        history?.clearLastIssue()
        lastHistoryIssue = nil
    }

    /// Committed navigation. Coalesces identical locations; truncates forward.
    func go(to location: WorkspaceLocation) {
        if location == currentLocation {
            performGo(to: location)
            return
        }
        if hold(.location(location)) { return }
        performGo(to: location)
    }

    func goBack() {
        guard history != nil, canGoBack else { return }
        if hold(.back) { return }
        performBack()
    }

    func goForward() {
        guard history != nil, canGoForward else { return }
        if hold(.forward) { return }
        performForward()
    }

    func go(toIndex index: Int) {
        guard history != nil else { return }
        if hold(.index(index)) { return }
        performIndex(index)
    }

    func resumeHeldNavigation() {
        guard let pending = heldNavigation else { return }
        heldNavigation = nil
        switch pending {
        case .location(let location):
            performGo(to: location)
        case .back:
            performBack()
        case .forward:
            performForward()
        case .index(let index):
            performIndex(index)
        }
    }

    func cancelHeldNavigation() {
        heldNavigation = nil
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
        refreshHistoryIssue()
    }

    private func hold(_ pending: PendingNavigation) -> Bool {
        guard let leaveGuard, leaveGuard.shouldHoldNavigation(pending) else { return false }
        heldNavigation = pending
        return true
    }

    private func performGo(to location: WorkspaceLocation) {
        guard let history else {
            apply(location)
            return
        }
        apply(history.go(to: location))
        refreshHistoryIssue()
    }

    private func performBack() {
        guard let history, let location = history.goBack() else { return }
        apply(location)
        refreshHistoryIssue()
    }

    private func performForward() {
        guard let history, let location = history.goForward() else { return }
        apply(location)
        refreshHistoryIssue()
    }

    private func performIndex(_ index: Int) {
        guard let history, let location = history.go(toIndex: index) else { return }
        apply(location)
        refreshHistoryIssue()
    }

    private func apply(_ location: WorkspaceLocation) {
        #if DEBUG
        if ProcessInfo.processInfo.environment["PROVENENCIA_DEBUG_NAV_TIMING"] == "1" {
            print(
                "WorkspaceNavigation.apply section=\(location.section.rawValue) " +
                    "sourceId=\(location.sourceId ?? "nil") fieldId=\(location.fieldId ?? "nil") " +
                    "typeId=\(location.typeId ?? "nil") t=\(CFAbsoluteTimeGetCurrent())"
            )
        }
        #endif
        currentLocation = location
        selectedSection = location.section
        canGoBack = history?.canGoBack ?? false
        canGoForward = history?.canGoForward ?? false
        onLocationCommit?(location)
    }

    private func refreshHistoryIssue() {
        if let issue = history?.lastIssue {
            lastHistoryIssue = issue
        }
    }
}
