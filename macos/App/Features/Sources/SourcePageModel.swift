import Foundation
import Observation

/// State for the individual Source page (S2-18): hydrates the workspace from
/// `WorkspaceSession` and composes the section models (identity, credibility,
/// metadata, notes, artifacts).
@MainActor
@Observable
final class SourcePageModel {
    // `var` (not `let`) so SwiftUI can form writable key paths for
    // `$model.section.draft` bindings; the sections are never reassigned.
    var identity: SourceIdentitySection
    var credibility: SourceCredibilitySection
    var metadata: SourceMetadataSection
    var notes: SourceNotesSection
    var artifacts: SourceArtifactsSection

    private(set) var isLoading = false
    var loadError: Error?

    /// Session contributor name for the note composer byline.
    let sessionDisplayName: String

    private let context: SourcePageContext
    /// Skips redundant section resets when the handle republishes the same payload.
    private var appliedWorkspaceFingerprint: String?

    init(
        sourceID: String,
        session: WorkspaceSession,
        userID: String,
        sessionDisplayName: String = "",
        store: any GenealogyStore
    ) {
        let context = SourcePageContext(
            sourceID: sourceID,
            projectDir: session.projectKey.projectDir,
            userID: userID,
            store: store,
            session: session
        )
        self.context = context
        self.sessionDisplayName = sessionDisplayName
        identity = SourceIdentitySection(context: context)
        credibility = SourceCredibilitySection(context: context)
        metadata = SourceMetadataSection(context: context)
        notes = SourceNotesSection(context: context)
        artifacts = SourceArtifactsSection(context: context)
    }

    var workspace: CatalogSourceWorkspace? { context.workspace }

    var source: CatalogSource? { context.source }

    /// Project folder for resolving `objects/…` thumbnail paths.
    var pageProjectDir: String { context.projectDir }

    /// Non-field failure banner shared by the page sections.
    var pageError: String? {
        get { context.pageError }
        set { context.pageError = newValue }
    }

    var toast: VocabularyToast? {
        get { context.toast }
        set { context.toast = newValue }
    }

    /// Loads the workspace query handle and syncs section state (tests and previews).
    func warmFromSession() async {
        let session = context.session
        let handle: QueryHandle<CatalogSourceWorkspace> = session.query(
            CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: context.sourceID)
        )
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < 2_000_000_000 {
            if handle.status == .ready || handle.status == .error { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: step)
            waited += step
        }
        sync(from: handle, sourceID: context.sourceID)
    }

    /// Rebinds section state when history switches to another source id.
    func prepare(for sourceID: String) {
        guard context.sourceID != sourceID else { return }
        context.sourceID = sourceID
        context.workspace = nil
        context.pageError = nil
        appliedWorkspaceFingerprint = nil
        resetSections()
        isLoading = false
        loadError = nil
    }

    /// Mirrors `QueryHandle` state into the page observation graph.
    func sync(from handle: QueryHandle<CatalogSourceWorkspace>, sourceID: String) {
        guard context.sourceID == sourceID else { return }
        isLoading = handle.status == .loading && handle.value == nil
        if handle.status == .error, handle.value == nil {
            loadError = handle.error
        } else {
            loadError = nil
        }
        guard let workspace = handle.value else { return }
        let fingerprint = workspaceFingerprint(workspace)
        if fingerprint == appliedWorkspaceFingerprint {
            context.workspace = workspace
            return
        }
        appliedWorkspaceFingerprint = fingerprint
        apply(workspace)
    }

    private func apply(_ workspace: CatalogSourceWorkspace) {
        context.workspace = workspace
        context.pageError = nil
        identity.reset()
        credibility.resetDrafts()
        metadata.resetDrafts()
        notes.reset()
        artifacts.seedDrafts()
    }

    private func workspaceFingerprint(_ workspace: CatalogSourceWorkspace) -> String {
        [
            workspace.source.id,
            workspace.source.title,
            workspace.source.coverMode,
            workspace.source.primaryArtifactID,
            String(workspace.metadata.count),
            String(workspace.notes.count),
            String(workspace.artifacts.count),
        ].joined(separator: "|")
    }

    private func resetSections() {
        identity.reset()
        credibility.resetDrafts()
        metadata.resetDrafts()
        notes.reset()
        artifacts.seedDrafts()
    }
}
