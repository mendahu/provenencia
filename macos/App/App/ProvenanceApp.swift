import SwiftUI

@main
struct ProvenanceApp: App {
    var body: some Scene {
        WindowGroup {
            // TEMPORARY (Spike 1 PR2): replace with onboarding / workspace when those screens exist.
            VersionSmokeView()
        }
        .windowResizability(.contentSize)
    }
}
