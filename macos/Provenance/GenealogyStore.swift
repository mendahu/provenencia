import Foundation

protocol GenealogyStore: Sendable {
    func ping(_ message: String) async throws -> String
    func version() async throws -> String
}
