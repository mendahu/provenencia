import AppKit
import Foundation
import Testing
@testable import Provenencia

@Suite @MainActor
struct AppLaunchTests {
    @Test func hostRunsTheTestApp() {
        #expect(ProvenenciaMain.isTestHost)
        #expect(NSApplication.shared.activationPolicy() == .accessory)
    }

    @Test func noWindowIsVisible() {
        let visible = NSApplication.shared.windows.filter(\.isVisible)
        #expect(visible.isEmpty)
    }
}
