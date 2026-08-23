import Foundation

protocol GenealogyStore: Sendable {
    func ping(_ message: String) async throws -> String
    func version() async throws -> String
    // TEMPORARY (Spike 1 PR3): dylib SQLite probe. Remove when project open/create exists.
    func sqliteProbe() async throws -> (ok: Bool, detail: String)
}
