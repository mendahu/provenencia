import Foundation
import Observation

@Observable
final class RootModel {
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
