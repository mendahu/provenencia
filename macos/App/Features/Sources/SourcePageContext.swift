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
    let session: WorkspaceSession

    var workspace: CatalogSourceWorkspace?
    /// Non-field failure banner shared by the page sections: one value, so a title can never
    /// pair with another write's message. Written only through `showPageError`.
    struct PageError: Equatable {
        var title: String?
        var message: String
    }

    private(set) var pageError: PageError?
    /// How many times a page error has been shown; clearing does not change it. The page folds
    /// it into its scroll-reset key so a new error scrolls up to the banner.
    private(set) var pageErrorShownCount = 0
    var toast: VocabularyToast?
    /// The notice to show when a write found this Source deleted elsewhere; the page leaves
    /// for the Sources list while it is set.
    var gone: String?

    /// Clears the shared banner before a new section mutation.
    func clearPageError() {
        pageError = nil
    }

    /// Shows `message` (with an optional bold `title`) in the shared banner.
    func showPageError(_ message: String, title: String? = nil) {
        pageError = PageError(title: title, message: message)
        pageErrorShownCount += 1
    }

    /// The one door for a failed Source page write. A Source another window deleted sets `gone`
    /// and returns nil: the caller returns without writing any state, the page is leaving.
    /// Any other failure returns its message for the caller to show.
    func fail(_ error: Error) -> String? {
        if (error as? CoreInvokeError)?.engineCode == .sourceNotFound {
            gone = L10n.Sources.sourceDeletedNotice(title: source?.title ?? "")
            return nil
        }
        return L10n.Errors.message(for: error)
    }

    /// A failed write whose answer is the page banner: `fail`, then the banner, with `wrap`
    /// shaping the message (and `title` heading it). False when the Source was deleted
    /// elsewhere and the page is leaving, so the caller returns without writing state.
    @discardableResult
    func showFailure(
        _ error: Error,
        title: String? = nil,
        wrap: (String) -> String = { $0 }
    ) -> Bool {
        guard let message = fail(error) else { return false }
        showPageError(wrap(message), title: title)
        return true
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
    // owns each list, so adding a field on the Metadata page shows up here
    // without busting the page. Empty until the list loads — `PlaceRegistry`
    // warms all three alongside the workspace.

    var types: [CatalogSourceType] {
        let handle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(
            .sourceTypesList(project: session.projectKey)
        )
        return handle?.value ?? []
    }

    /// The Source's type icon, for Artifact thumbnails and covers that have no file.
    var sourceTypeIconKey: String? {
        types.iconKey(forTypeID: source?.sourceTypeID)
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
    /// type, so its suggested metadata rows and per-type counts refetch. The
    /// write's `effects` name its graphs and Source activity; the mutation
    /// carries the enriched row.
    func applySource(_ source: CatalogSource, effects: CatalogWriteEffects, typeChanged: Bool = false) {
        workspace?.source = source
        session.apply(effects, typeChanged ? .changedSourceType(source) : .updatedSource(source))
    }

    /// Applies a non-identity page write (note, artifact, credibility): the
    /// write's effects move the cached lists, and this page's workspace follows
    /// the mutation.
    func applyWrite(_ effects: CatalogWriteEffects) {
        session.apply(effects, .mutatedSourceWorkspace(sourceId: sourceID))
    }

    /// Same, plus the field list, whose `usedBy` counts move with every
    /// metadata value written or cleared here.
    func applyMetadataWrite(_ effects: CatalogWriteEffects) {
        session.apply(effects, .mutatedSourceMetadata(sourceId: sourceID))
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
            showFailure(error)
        }
    }

    /// Pins a file-bearing Artifact as cover, or reverts to the type icon.
    func setSourceCover(mode: String, primaryArtifactID: String = "") async -> Bool {
        clearPageError()
        do {
            let (updated, effects) = try await store.setSourceCover(
                projectDir: projectDir,
                userID: userID,
                sourceID: sourceID,
                coverMode: mode,
                primaryArtifactID: primaryArtifactID
            )
            applySource(updated, effects: effects)
            return true
        } catch {
            // A gone Source: the page is leaving, nothing is shown.
            showFailure(error)
            return false
        }
    }
}
