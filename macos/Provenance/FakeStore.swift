import Foundation

struct FakeStore: GenealogyStore {
    func ping(_ message: String) async throws -> String {
        message
    }

    func version() async throws -> String {
        "0.0.0"
    }
}
