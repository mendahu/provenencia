import Foundation
import Observation

/// Origin split for a vocabulary destination. The sidebar badge uses
/// `total`; Source fields / Source types headers use the same numbers for
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
    private(set) var sourceFields: CatalogCountSummary?
    private(set) var sourceTypes: CatalogCountSummary?
    private(set) var subjectFields: CatalogCountSummary?
    /// Set when `refreshAll` fails; cleared on the next successful refresh.
    private(set) var lastRefreshError: String?

    private let projectDir: String
    private let store: any GenealogyStore

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
        case .sourceFields: sourceFields?.total
        case .subjectFields: subjectFields?.total
        case .subjectTypes: nil
        }
    }

    /// Writes a vocabulary summary the feature model already computed
    /// from its in-memory rows (after load, create, or delete).
    func publishSourceFields(_ summary: CatalogCountSummary) {
        sourceFields = summary
    }

    func publishSourceTypes(_ summary: CatalogCountSummary) {
        sourceTypes = summary
    }

    func publishSubjectFields(_ summary: CatalogCountSummary) {
        subjectFields = summary
    }

    /// Writes the Sources total the feature model already knows from its
    /// in-memory list (after load or create).
    func publishSources(_ count: Int) {
        sources = count
    }

    /// Full badge refresh via `GetWorkspaceNavCounts`. Safe to overlap with
    /// destination `load()` — Go `catalogsession` serializes catalog ops.
    /// Failures set `lastRefreshError` and leave badges unchanged (absent until
    /// a later success). Mutations use `publish*` instead of recounting.
    func refreshAll() async {
        do {
            let nav = try await store.workspaceNavCounts(projectDir: projectDir)
            sources = nav.sources
            sourceTypes = CatalogCountSummary(nav.sourceTypes)
            sourceFields = CatalogCountSummary(nav.sourceFields)
            lastRefreshError = nil
        } catch {
            lastRefreshError = L10n.Errors.message(for: error)
        }
    }
}
