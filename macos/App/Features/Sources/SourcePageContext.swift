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
    var sourceID: String
    let projectDir: String
    let userID: String
    let store: any GenealogyStore
    let session: WorkspaceSession

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
        session: WorkspaceSession
    ) {
        self.sourceID = sourceID
        self.projectDir = projectDir
        self.userID = userID
        self.store = store
        self.session = session
    }

    var source: CatalogSource? { workspace?.source }

    // MARK: Shared vocabulary

    // Read from the shared list caches rather than the page payload: one cache
    // owns each list, so adding a field on the Source Fields page shows up here
    // without busting the page. Empty until the list loads — `PlaceRegistry`
    // warms all three alongside the workspace.

    var types: [CatalogSourceType] {
        let handle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(
            .sourceTypesList(project: session.projectKey)
        )
        return handle?.value ?? []
    }

    var grades: [CatalogCredibilityGrade] {
        let handle: QueryHandle<[CatalogCredibilityGrade]>? = session.queryHandle(
            .credibilityGradesList(project: session.projectKey)
        )
        return handle?.value ?? []
    }

    var fields: [CatalogMetadataField] {
        let handle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            .metadataFieldsList(project: session.projectKey)
        )
        return handle?.value ?? []
    }

    /// Applies an enriched Source (cover fields included) to the workspace and
    /// list cache. Pass `typeChanged` when the write moved the Source to another
    /// type, so its suggested metadata rows and per-type counts refetch.
    func applySource(_ source: CatalogSource, typeChanged: Bool = false) {
        workspace?.source = source
        session.apply(typeChanged ? .changedSourceType(source) : .updatedSource(source))
    }

    /// Busts the workspace cache after non-identity page edits.
    func notifyWorkspaceMutated() {
        session.apply(.mutatedSourceWorkspace(sourceId: sourceID))
    }

    /// Busts the workspace cache plus the field list, whose `usedBy` counts move
    /// with every metadata value written or cleared here.
    func notifyMetadataMutated() {
        session.apply(.mutatedSourceMetadata(sourceId: sourceID))
    }

    /// Reloads Source cover fields after Artifact create/ingest (raster may appear).
    func refreshCoverFromStore() async {
        clearPageError()
        do {
            let ws = try await store.getSourceWorkspace(projectDir: projectDir, sourceID: sourceID)
            workspace = ws
            session.setQueryValue(
                CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: sourceID),
                value: ws
            )
            session.apply(.updatedSource(ws.source))
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
