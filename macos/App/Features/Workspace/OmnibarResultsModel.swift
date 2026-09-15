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
    var isLoading = false
    var hasSearched = false
    /// Set when Esc / outside click closes the panel without clearing the query.
    var userDismissed = false
    /// Non-nil when the last completed search failed (not “no matches”).
    var searchError: String?
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

    var showsLoadingState: Bool {
        isLoading && hits.isEmpty && searchError == nil
    }

    var showsErrorState: Bool {
        searchError != nil && !isLoading
    }

    var showsEmptyState: Bool {
        hasSearched && !isLoading && hits.isEmpty && searchError == nil
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
        isLoading = false
        hasSearched = false
        userDismissed = false
        searchError = nil
        selectedIndex = 0
    }

    func clearAfterNavigate() {
        query = ""
        dismiss()
    }

    /// Return / click on a hit: commit navigation and clear the results chrome.
    func activate(_ hit: CatalogSearchHit, navigation: WorkspaceNavigation) {
        navigation.go(to: hit.location)
        clearAfterNavigate()
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
            isLoading = false
            hasSearched = false
            userDismissed = false
            searchError = nil
            selectedIndex = 0
            return
        }

        userDismissed = false
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
            let results = try await store.searchCatalog(
                projectDir: projectDir,
                query: query,
                location: location
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
            searchError = String(localized: L10n.Workspace.omnibarSearchFailed)
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

    /// Whether the match-context slot should show for this stable field code.
    static func showsMatchContext(field: String) -> Bool {
        switch field.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() {
        case "notes", "metadata", "filename", "description":
            return true
        default:
            return false
        }
    }

    /// Localized match-context line from Go field code + optional raw snippet.
    static func matchContextText(field: String, snippet: String) -> String {
        let code = field.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard showsMatchContext(field: code) else { return "" }
        let trimmed = snippet.trimmingCharacters(in: .whitespacesAndNewlines)
        switch code {
        case "notes":
            return L10n.Workspace.omnibarMatchNote(snippet: trimmed)
        case "metadata":
            return L10n.Workspace.omnibarMatchMetadata(snippet: trimmed)
        case "filename":
            return L10n.Workspace.omnibarMatchFilename(snippet: trimmed)
        case "description":
            return String(localized: L10n.Workspace.omnibarMatchDescription)
        default:
            return ""
        }
    }

    static func refAccent(for hit: CatalogSearchHit) -> Bool {
        hit.matchReason == "ref" && !hit.ref.isEmpty
    }

    /// VoiceOver label for a result row — title, kind, subtitle, ref, match context.
    static func accessibilityLabel(for hit: CatalogSearchHit) -> String {
        var parts = [
            hit.title,
            String(localized: kindLabel(for: hit.kind)),
        ]
        let subtitle = hit.subtitle.trimmingCharacters(in: .whitespacesAndNewlines)
        if !subtitle.isEmpty {
            parts.append(subtitle)
        }
        let ref = hit.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if !ref.isEmpty {
            parts.append(ref)
        }
        let matchContext = matchContextText(field: hit.matchReason, snippet: hit.matchSnippet)
        if !matchContext.isEmpty {
            parts.append(matchContext)
        }
        let formatter = ListFormatter()
        formatter.locale = .current
        return formatter.string(from: parts) ?? parts.joined(separator: ", ")
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
