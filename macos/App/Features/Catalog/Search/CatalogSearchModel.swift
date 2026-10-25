import Foundation
import Observation

/// Debounced SearchCatalog state shared by every find-a-record surface: the
/// toolbar omnibar and pickers. `kinds` is fixed for the model's life (empty:
/// every searchable kind); the location is per search (nil: no section nudge).
@MainActor
@Observable
final class CatalogSearchModel {
    static let minQueryLength = 2
    static let debounceNanoseconds: UInt64 = 180_000_000

    let kinds: [String]

    var query = ""
    var hits: [CatalogSearchHit] = []
    var isLoading = false
    var hasSearched = false
    /// Non-nil when the last completed search failed (not "no matches").
    var searchError: String?
    var selectedIndex = 0

    private var searchGeneration = 0
    private var debounceTask: Task<Void, Never>?

    init(kinds: [String] = []) {
        self.kinds = kinds
    }

    var trimmedQuery: String {
        query.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// The query is too short to search.
    var isBelowMinimumQuery: Bool {
        trimmedQuery.count < Self.minQueryLength
    }

    /// Results show once the query is long enough and a search has run (or is in flight).
    var showsResults: Bool {
        !isBelowMinimumQuery && (hasSearched || isLoading)
    }

    var showsLoadingState: Bool {
        isLoading && hits.isEmpty && searchError == nil
    }

    var showsErrorState: Bool {
        searchError != nil && !isLoading
    }

    var showsEmptyState: Bool {
        hasSearched && !isLoading && hits.isEmpty && searchError == nil
    }

    /// Stops any pending or in-flight search; keeps the query and the hits on screen.
    func cancel() {
        debounceTask?.cancel()
        debounceTask = nil
        searchGeneration += 1
        isLoading = false
    }

    func dismiss() {
        cancel()
        hits = []
        hasSearched = false
        searchError = nil
        selectedIndex = 0
    }

    func scheduleSearch(
        projectDir: String,
        location: WorkspaceLocation?,
        store: any GenealogyStore
    ) {
        debounceTask?.cancel()
        let q = trimmedQuery
        guard !isBelowMinimumQuery else {
            dismiss()
            return
        }

        searchError = nil
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
        selectedIndex = PVFloatingMenuSelection.moveIndex(
            from: selectedIndex, delta: delta, count: hits.count
        )
    }

    func selectedHit() -> CatalogSearchHit? {
        guard hits.indices.contains(selectedIndex) else { return nil }
        return hits[selectedIndex]
    }

    private func runSearch(
        generation: Int,
        projectDir: String,
        query: String,
        location: WorkspaceLocation?,
        store: any GenealogyStore
    ) async {
        do {
            let results = try await store.searchCatalog(
                projectDir: projectDir,
                query: query,
                location: location,
                kinds: kinds
            )
            guard generation == searchGeneration else { return }
            hits = results
            hasSearched = true
            searchError = nil
            isLoading = false
            selectedIndex = 0
        } catch {
            guard generation == searchGeneration else { return }
            hits = []
            hasSearched = true
            searchError = L10n.Errors.message(for: error)
            isLoading = false
            selectedIndex = 0
        }
    }
}
