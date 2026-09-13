import Foundation
import Observation

/// Shared dependencies and workspace state for the Source page section
/// models (identity, credibility, metadata, notes, artifacts).
///
/// The workspace is the single source of truth: sections read saved values
/// from it and mutate it directly after successful writes. Non-field
/// failures surface through `pageError`; section-specific validation stays
/// on the owning section.
@MainActor
@Observable
final class SourcePageContext {
    let sourceID: String
    let projectDir: String
    let userID: String
    let store: any GenealogyStore
    /// Notifies the Sources list when identity or list-cover thumbnails change.
    let onSourceUpdated: ((CatalogSource) -> Void)?

    var workspace: CatalogSourceWorkspace?
    /// Non-field failure banner shared by the page sections.
    var pageError: String?
    var toast: VocabularyToast?

    /// Clears the shared banner before a new section mutation.
    func clearPageError() {
        pageError = nil
    }

    init(
        sourceID: String,
        projectDir: String,
        userID: String,
        store: any GenealogyStore,
        onSourceUpdated: ((CatalogSource) -> Void)?
    ) {
        self.sourceID = sourceID
        self.projectDir = projectDir
        self.userID = userID
        self.store = store
        self.onSourceUpdated = onSourceUpdated
    }

    var source: CatalogSource? { workspace?.source }

    /// Applies an enriched Source (cover fields included) to the workspace and list.
    func applySource(_ source: CatalogSource) {
        workspace?.source = source
        onSourceUpdated?(source)
    }

    /// Reloads Source cover fields after Artifact create/ingest (server may auto-pin).
    func refreshCoverFromStore() async {
        do {
            let ws = try await store.getSourceWorkspace(projectDir: projectDir, sourceID: sourceID)
            applySource(ws.source)
        } catch {
            pageError = L10n.Errors.message(for: error)
        }
    }

    /// Pins a file-bearing Artifact as cover, or reverts to the type icon.
    func setSourceCover(mode: String, primaryArtifactID: String = "") async -> Bool {
        clearPageError()
        do {
            let updated = try await store.setSourceCover(
                projectDir: projectDir,
                userID: userID,
                sourceID: sourceID,
                coverMode: mode,
                primaryArtifactID: primaryArtifactID
            )
            applySource(updated)
            return true
        } catch {
            pageError = L10n.Errors.message(for: error)
            return false
        }
    }
}
