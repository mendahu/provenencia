import SwiftUI

/// Process entry. Xcode hosts the unit tests inside Provenencia.app, so a test
/// run launches `TestHostApp` (no window, no identity, no project) instead of
/// the researcher's scene.
@main
enum ProvenenciaMain {
    static var isTestHost: Bool {
        ProcessInfo.processInfo.environment["XCTestConfigurationFilePath"] != nil
    }

    static func main() {
        if isTestHost {
            TestHostApp.main()
        } else {
            ProvenenciaApp.main()
        }
    }
}

/// The app Xcode hosts the unit tests in: no window, no identity, no project.
struct TestHostApp: App {
    init() {
        PVFontRegistration.registerBundledFontsIfNeeded()
        NSApplication.shared.setActivationPolicy(.accessory)
    }

    var body: some Scene { Settings { EmptyView() } }
}
