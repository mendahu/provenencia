import SwiftUI

@main
struct ProvenenciaApp: App {
    var body: some Scene {
        WindowGroup {
            OnboardingView()
        }
        .windowResizability(.contentSize)
    }
}
