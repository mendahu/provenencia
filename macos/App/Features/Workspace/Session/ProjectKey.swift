import Foundation

/// Stable project scope for catalog query keys (open `.provenencia` directory).
struct ProjectKey: Hashable, Sendable {
    let projectDir: String

    init(projectDir: String) {
        self.projectDir = projectDir
    }
}
