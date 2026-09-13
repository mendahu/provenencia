import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct WorkspaceModelTests {
    private func makeDefaults() -> UserDefaults {
        let suiteName = "WorkspaceModelTests.\(UUID().uuidString)"
        let defaults = UserDefaults(suiteName: suiteName)!
        defaults.removePersistentDomain(forName: suiteName)
        return defaults
    }

    private func makeModel(defaults: UserDefaults? = nil) -> WorkspaceModel {
        WorkspaceModel(defaults: defaults ?? makeDefaults())
    }

    @Test func defaultsToExpanded() {
        let model = makeModel()
        #expect(!model.isSidebarCollapsed)
    }

    @Test func toggleCollapsesAndPersists() {
        let defaults = makeDefaults()
        let model = makeModel(defaults: defaults)
        model.toggleSidebarCollapsed()
        #expect(model.isSidebarCollapsed)

        let reloaded = makeModel(defaults: defaults)
        #expect(reloaded.isSidebarCollapsed)
    }

    @Test func toggleTwiceReturnsToExpanded() {
        let model = makeModel()
        model.toggleSidebarCollapsed()
        model.toggleSidebarCollapsed()
        #expect(!model.isSidebarCollapsed)
    }
}
