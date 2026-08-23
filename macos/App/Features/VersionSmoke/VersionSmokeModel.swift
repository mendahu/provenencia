import Foundation
import Observation

// TEMPORARY (Spike 1 PR2): only exists to prove GetVersion over FFI.
// Delete this type when a real root/workspace model exists.
@Observable
final class VersionSmokeModel {
    var versionText = ""
    var errorText: String?

    private let store: any GenealogyStore

    init(store: any GenealogyStore) {
        self.store = store
    }

    func loadVersion() async {
        do {
            versionText = try await store.version()
            errorText = nil
        } catch {
            errorText = error.localizedDescription
        }
    }
}
