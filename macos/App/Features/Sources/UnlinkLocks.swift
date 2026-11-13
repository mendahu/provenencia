import Foundation
import Observation

/// The engine's refusals to unlink Artifacts from Evidence graphs, read once
/// for a form that offers the links as checkboxes: the Artifact sheet and the
/// Evidence graph form hold one each. A block locks its row; Swift decides no
/// lock. `lockHints(by:)` in `SourceArtifactGraphLinks.swift` words each lock.
@MainActor
@Observable
final class UnlinkLocks {
    /// The links the engine refuses to unlink, with why.
    private(set) var blocks: [CatalogEvidenceGraphUnlinkBlock] = []
    private(set) var state: SourceFormLoadState

    init(state: SourceFormLoadState = .loading) {
        self.state = state
    }

    /// Key (a graph or Artifact id) to the hint under its locked row.
    func locked(by key: KeyPath<CatalogEvidenceGraphUnlinkBlock, String>) -> [String: String] {
        blocks.lockHints(by: key)
    }

    /// One unlink preflight over `pairs`, keeping the blocks `include` accepts.
    /// No pairs has nothing to read. A failure goes through `onError`, which
    /// answers the message to show, or nil when the page is leaving and the
    /// state stays as it was.
    func load(
        store: any GenealogyStore,
        projectDir: String,
        pairs: [CatalogEvidenceGraphArtifactPair],
        include: (CatalogEvidenceGraphUnlinkBlock) -> Bool = { _ in true },
        onError: (Error) -> String?
    ) async {
        if pairs.isEmpty {
            blocks = []
            state = .ready
            return
        }
        state = .loading
        do {
            let read = try await store.getEvidenceGraphUnlinkImpact(projectDir: projectDir, pairs: pairs)
            blocks = read.filter(include)
            state = .ready
        } catch {
            guard let message = onError(error) else { return }
            state = .failed(message)
        }
    }
}
