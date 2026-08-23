import SwiftUI

@main
struct ProvenanceApp: App {
    var body: some Scene {
        WindowGroup {
            OnboardingView()
        }
        .windowResizability(.contentSize)
    }
}
