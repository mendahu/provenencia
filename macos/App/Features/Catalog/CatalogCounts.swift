import Foundation
import Observation

/// Origin split for a vocabulary destination. The sidebar badge uses
/// `total`; Metadata / Source types headers use the same numbers for
/// their "N · seeded · yours · plugin" line. `total` is always
/// `seeded + user + plugin`.
struct CatalogCountSummary: Equatable {
    var total: Int
    var seeded: Int
    var user: Int
    var plugin: Int

    static let zero = CatalogCountSummary(total: 0, seeded: 0, user: 0, plugin: 0)

    init(total: Int, seeded: Int, user: Int, plugin: Int) {
        self.total = total
        self.seeded = seeded
        self.user = user
        self.plugin = plugin
    }

    init(_ counts: WorkspaceNavOriginCounts) {
        self.init(
            total: counts.total,
            seeded: counts.seeded,
            user: counts.user,
            plugin: counts.plugin
        )
    }

    /// Derives the summary from rows the caller already loaded — no extra
    /// catalog round trip.
    static func from<Row: CatalogVocabularyRow>(_ rows: [Row]) -> CatalogCountSummary {
        CatalogCountSummary(
            total: rows.count,
            seeded: rows.seededCount,
            user: rows.userCount,
            plugin: rows.pluginCount
        )
    }
}

/// Project-scoped catalog totals shared by the workspace sidebar and the
/// vocabulary destination headers. Owned by `WorkspaceView` and injected
/// via the environment — feature models publish after load / create /
/// delete instead of prop-drilling refresh callbacks.
///
/// Sources only need a total (no origin split). Vocabulary sections keep
/// the full summary so the header line and the nav badge stay one write apart.
@MainActor
@Observable
final class CatalogCounts {
    /// Absent until a refresh or publish succeeds for that section —
    /// the sidebar treats absence as "no badge" rather than `0`.
    private(set) var sources: Int?
    private(set) var metadata: CatalogCountSummary?
    private(set) var sourceTypes: CatalogCountSummary?
    private(set) var properties: CatalogCountSummary?
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
        case .sourceTypes: sourceTypes?.total
        case .metadata: metadata?.total
        case .properties: properties?.total
        case .persons: persons
        case .events: events
        case .places: places
        }
    }

    /// Writes a vocabulary summary the feature model already computed
    /// from its in-memory rows (after load, create, or delete).
    func publishMetadata(_ summary: CatalogCountSummary) {
        metadata = summary
    }

    func publishSourceTypes(_ summary: CatalogCountSummary) {
        sourceTypes = summary
    }

    func publishProperties(_ summary: CatalogCountSummary) {
        properties = summary
    }

    /// Full badge refresh via `GetWorkspaceNavCounts`. Safe to overlap with
    /// destination `load()` — Go `catalogsession` serializes catalog ops.
    /// Failures set `lastRefreshError` and leave badges unchanged (absent until
    /// a later success). Vocabulary mutations use `publish*` instead of
    /// recounting; a write whose `WriteEffects.lists` has `counts` recounts
    /// through here, called by `WorkspaceSession.apply` (wired in
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
            sourceTypes = CatalogCountSummary(nav.sourceTypes)
            metadata = CatalogCountSummary(nav.metadataFields)
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
