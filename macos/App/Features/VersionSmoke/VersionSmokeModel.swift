import Foundation
import Observation

// TEMPORARY (Spike 1 PR3): proves GetVersion and SQLite-in-dylib over FFI.
// Delete this type when a real root/workspace model exists.
@Observable
final class VersionSmokeModel {
    var versionText = ""
    var sqliteText = ""
    var errorText: String?

    private let store: any GenealogyStore

    init(store: any GenealogyStore) {
        self.store = store
    }

    func loadVersion() async {
        do {
            versionText = try await store.version()
            let probe = try await store.sqliteProbe()
            sqliteText = probe.ok ? "sqlite: \(probe.detail)" : "sqlite failed: \(probe.detail)"
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }
}
