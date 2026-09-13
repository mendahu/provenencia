import Observation

/// Bridges workspace Back/Forward to the macOS app menu / keyboard commands.
/// `ProvenenciaApp.commands` lives outside the view hierarchy that owns
/// `WorkspaceNavigation` — same pattern as `SignOutCoordinator`.
@MainActor
@Observable
final class NavigationCoordinator {
    var canGoBack = false
    var canGoForward = false
    var goBack: () -> Void = {}
    var goForward: () -> Void = {}
}
