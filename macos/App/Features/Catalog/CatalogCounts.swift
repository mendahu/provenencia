import Foundation
import Observation

/// The sidebar's Research badges: Sources and the unmerged Person, Event and
/// Place handles. Owned by `WorkspaceView`; a write whose `WriteEffects.lists`
/// has `counts` recounts through `refreshAll`. Configure rows draw no badge.
@MainActor
@Observable
final class CatalogCounts {
    /// Absent until a refresh or publish succeeds for that section —
    /// the sidebar treats absence as "no badge" rather than `0`.
    private(set) var sources: Int?
    /// Unmerged Person, Event, and Place handles.
    private(set) var persons: Int?
    private(set) var events: Int?
    private(set) var places: Int?
    /// Set when `refreshAll` fails; cleared on the next successful refresh.
    private(set) var lastRefreshError: String?

    private let projectDir: String
    private let store: any GenealogyStore
    /// Bumped by each `refreshAll`; a recount assigns only while it is the newest.
    private var refreshGeneration = 0

    init(projectDir: String, store: any GenealogyStore) {
        self.projectDir = projectDir
        self.store = store
    }

    func clearRefreshError() {
        lastRefreshError = nil
    }

    /// Sidebar badge for a nav destination.
    func badge(for section: WorkspaceSection) -> Int? {
        switch section {
        case .sources: sources
        case .persons: persons
        case .events: events
        case .places: places
        default: nil
        }
    }

    /// Full badge refresh via `GetWorkspaceNavCounts`. Safe to overlap with
    /// destination `load()` — Go `catalogsession` serializes catalog ops.
    /// Failures set `lastRefreshError` and leave badges unchanged (absent until
    /// a later success). A write whose `WriteEffects.lists` has `counts`
    /// recounts through here, called by `WorkspaceSession.apply` (wired in
    /// `WorkspaceView`). The newest recount wins: Go serializes catalog ops in
    /// call order, so the call started last read the newest catalog, and a
    /// slower earlier call that finishes after it assigns nothing.
    func refreshAll() async {
        refreshGeneration += 1
        let generation = refreshGeneration
        do {
            let nav = try await store.workspaceNavCounts(projectDir: projectDir)
            guard generation == refreshGeneration else { return }
            sources = nav.sources
            persons = nav.persons
            events = nav.events
            places = nav.places
            lastRefreshError = nil
        } catch {
            guard generation == refreshGeneration else { return }
            lastRefreshError = L10n.Errors.message(for: error)
        }
    }
}
