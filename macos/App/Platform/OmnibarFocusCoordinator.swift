import Observation

/// Bridges ⌘K / menu “Focus search” to the workspace toolbar omnibar field.
/// `ProvenenciaApp.commands` lives outside the view that owns `@FocusState`.
@MainActor
@Observable
final class OmnibarFocusCoordinator {
    /// Incremented to request focus; the toolbar observes and focuses the field.
    private(set) var focusGeneration = 0

    func focusOmnibar() {
        focusGeneration += 1
    }
}
