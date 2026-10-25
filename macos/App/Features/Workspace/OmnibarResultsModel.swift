import Foundation
import Observation

/// The toolbar omnibar's results panel: dismissal and anchor around a shared `CatalogSearchModel`.
@MainActor
@Observable
final class OmnibarResultsModel {
    static let panelWidth: CGFloat = 640

    let search = CatalogSearchModel()
    /// Set when Esc / outside click closes the panel without clearing the query.
    var userDismissed = false
    var fieldFrame: CGRect = .zero

    /// Panel is shown once the query is long enough and a search has run (or is in flight).
    var isPresented: Bool {
        !userDismissed && search.showsResults
    }

    func closePanel() {
        userDismissed = true
        search.cancel()
    }

    func dismiss() {
        search.dismiss()
        userDismissed = false
    }

    func clearAfterNavigate() {
        search.query = ""
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
        userDismissed = false
        search.scheduleSearch(projectDir: projectDir, location: location, store: store)
    }
}
