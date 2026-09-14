import Foundation
import Observation

/// Debounced SearchCatalog state for the toolbar omnibar results overlay.
@MainActor
@Observable
final class OmnibarResultsModel {
    static let minQueryLength = 2
    static let debounceNanoseconds: UInt64 = 180_000_000
    static let panelWidth: CGFloat = 640

    var query = ""
    var hits: [CatalogSearchHit] = []
    /// Source rows keyed by id — filled after each search for lead thumbnails.
    var sourcesByID: [String: CatalogSource] = [:]
    /// Type rows keyed by id — filled after each search for type/source icons.
    var typesByID: [String: CatalogSourceType] = [:]
    var isLoading = false
    var hasSearched = false
    /// Set when Esc / outside click closes the panel without clearing the query.
    var userDismissed = false
    var selectedIndex = 0
    var fieldFrame: CGRect = .zero

    private var searchGeneration = 0
    private var debounceTask: Task<Void, Never>?

    var trimmedQuery: String {
        query.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Panel is shown once the query is long enough and a search has run (or is in flight).
    var isPresented: Bool {
        !userDismissed && trimmedQuery.count >= Self.minQueryLength && (hasSearched || isLoading)
    }

    var showsEmptyState: Bool {
        hasSearched && !isLoading && hits.isEmpty
    }

    func closePanel() {
        userDismissed = true
        isLoading = false
    }

    func dismiss() {
        debounceTask?.cancel()
        debounceTask = nil
        searchGeneration += 1
        hits = []
        sourcesByID = [:]
        typesByID = [:]
        isLoading = false
        hasSearched = false
        userDismissed = false
        selectedIndex = 0
    }

    func clearAfterNavigate() {
        query = ""
        dismiss()
    }

    func scheduleSearch(
        projectDir: String,
        location: WorkspaceLocation,
        store: any GenealogyStore
    ) {
        debounceTask?.cancel()
        let q = trimmedQuery
        guard q.count >= Self.minQueryLength else {
            searchGeneration += 1
            hits = []
            sourcesByID = [:]
            typesByID = [:]
            isLoading = false
            hasSearched = false
            userDismissed = false
            selectedIndex = 0
            return
        }

        userDismissed = false
        isLoading = true
        let generation = searchGeneration + 1
        searchGeneration = generation
        debounceTask = Task { [weak self] in
            try? await Task.sleep(nanoseconds: Self.debounceNanoseconds)
            guard !Task.isCancelled else { return }
            await self?.runSearch(
                generation: generation,
                projectDir: projectDir,
                query: q,
                location: location,
                store: store
            )
        }
    }

    func moveSelection(delta: Int) {
        guard !hits.isEmpty else { return }
        let next = selectedIndex + delta
        selectedIndex = max(0, min(hits.count - 1, next))
    }

    func selectedHit() -> CatalogSearchHit? {
        guard hits.indices.contains(selectedIndex) else { return nil }
        return hits[selectedIndex]
    }

    private func runSearch(
        generation: Int,
        projectDir: String,
        query: String,
        location: WorkspaceLocation,
        store: any GenealogyStore
    ) async {
        do {
            async let search = store.searchCatalog(
                projectDir: projectDir,
                query: query,
                location: location
            )
            async let sources = store.listSources(projectDir: projectDir)
            async let types = store.listSourceTypes(projectDir: projectDir)
            let (results, sourceRows, typeRows) = try await (search, sources, types)
            guard generation == searchGeneration else { return }
            hits = results
            sourcesByID = Dictionary(uniqueKeysWithValues: sourceRows.map { ($0.id, $0) })
            typesByID = Dictionary(uniqueKeysWithValues: typeRows.map { ($0.id, $0) })
            hasSearched = true
            isLoading = false
            selectedIndex = 0
        } catch {
            guard generation == searchGeneration else { return }
            hits = []
            sourcesByID = [:]
            typesByID = [:]
            hasSearched = true
            isLoading = false
            selectedIndex = 0
        }
    }
}

/// Maps a SearchCatalog hit into omnibar row copy (kind-agnostic slots).
enum OmnibarHitPresentation {
    static func kindLabel(for kind: String) -> LocalizedStringResource {
        switch kind {
        case "source": L10n.Workspace.omnibarKindSource
        case "source_type": L10n.Workspace.omnibarKindType
        case "source_field": L10n.Workspace.omnibarKindField
        default: L10n.Workspace.omnibarKindSource
        }
    }

    static func showMatchContext(_ reason: String) -> Bool {
        let r = reason.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if r.isEmpty { return false }
        if r == "title" || r == "ref" || r == "label" || r == "key" { return false }
        return true
    }

    static func refAccent(for hit: CatalogSearchHit) -> Bool {
        hit.matchReason == "ref" && !hit.ref.isEmpty
    }

    static func leadSymbol(for kind: String) -> PVSymbol {
        switch kind {
        case "source": .scrollText
        case "source_type": .library
        case "source_field": .tag
        default: .search
        }
    }
}
