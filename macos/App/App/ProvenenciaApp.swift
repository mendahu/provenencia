import SwiftUI

@main
struct ProvenenciaApp: App {
    init() {
        PVFontRegistration.registerBundledFontsIfNeeded()
    }

    var body: some Scene {
        WindowGroup {
            OnboardingView()
        }
        .windowResizability(.contentSize)
    }
}
