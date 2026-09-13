import Foundation
import Observation

/// Workspace chrome that is not navigation history: sidebar collapse only.
/// Place / Back / Forward live on `WorkspaceNavigation` (environment).
@MainActor
@Observable
final class WorkspaceModel {
    /// W-5: the researcher chooses this; it's never derived from window
    /// width. Persisted per install via `UserDefaults`, as the design brief
    /// suggests.
    private static let sidebarCollapsedDefaultsKey = "pv.sidebarCollapsed"

    var isSidebarCollapsed: Bool

    private let defaults: UserDefaults

    init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
        isSidebarCollapsed = defaults.bool(forKey: Self.sidebarCollapsedDefaultsKey)
    }

    func toggleSidebarCollapsed() {
        isSidebarCollapsed.toggle()
        defaults.set(isSidebarCollapsed, forKey: Self.sidebarCollapsedDefaultsKey)
    }
}
