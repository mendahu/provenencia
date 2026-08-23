import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
struct FakeStore: GenealogyStore {
    func ping(_ message: String) async throws -> String {
        message
    }

    func version() async throws -> String {
        "0.0.0"
    }

    // TEMPORARY (Spike 1 PR3): dylib SQLite probe. Remove when project open/create exists.
    func sqliteProbe() async throws -> (ok: Bool, detail: String) {
        (true, "fake")
    }
}
