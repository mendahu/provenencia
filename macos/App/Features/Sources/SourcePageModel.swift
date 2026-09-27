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
    var pendingImpact: PVDeleteImpactRequest?
    var isDeletingResource = false
    private var pendingResourceID: String?

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

    /// Loads the workspace and page vocabulary, then syncs section state
    /// (tests and previews; navigation warms the same keys via `PlaceRegistry`).
    func warmFromSession() async {
        let session = context.session
        let key = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: context.sourceID)
        session.invalidate(key)
        let handle: QueryHandle<CatalogSourceWorkspace> = session.query(key)
        let types: QueryHandle<[CatalogSourceType]> = session.query(
            .sourceTypesList(project: session.projectKey)
        )
        let grades: QueryHandle<[CatalogCredibilityGrade]> = session.query(
            .credibilityGradesList(project: session.projectKey)
        )
        let fields: QueryHandle<[CatalogMetadataField]> = session.query(
            .metadataFieldsList(project: session.projectKey)
        )
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < 2_000_000_000 {
            if handle.status == .error { break }
            if handle.status == .ready, !handle.isFetching,
               types.status != .loading, grades.status != .loading, fields.status != .loading
            {
                break
            }
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

    func askDeleteSource() async {
        guard let source else { return }
        await presentImpact(kind: "source", id: source.id, ref: source.ref, title: source.title)
    }

    func askDeleteArtifact(id: String) async {
        guard let art = artifacts.items.first(where: { $0.id == id }) else { return }
        let title = art.label.isEmpty ? art.ref : art.label
        await presentImpact(kind: "artifact", id: art.id, ref: art.ref, title: title)
    }

    /// Returns true when the Source itself was erased (leave the page).
    @discardableResult
    func confirmResourceDelete() async -> Bool {
        guard let request = pendingImpact, request.report.allowed, let id = pendingResourceID else {
            return false
        }
        guard !isDeletingResource else { return false }
        isDeletingResource = true
        defer { isDeletingResource = false }
        context.clearPageError()
        do {
            if request.target.kind == "source" {
                try await context.store.deleteSource(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: id
                )
                context.session.apply(.deletedSource(id: id))
                pendingImpact = nil
                pendingResourceID = nil
                return true
            }
            try await context.store.deleteArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id
            )
            if var workspace = context.workspace {
                workspace.artifacts.removeAll { $0.id == id }
                if workspace.source.primaryArtifactID == id {
                    workspace.source.primaryArtifactID = ""
                    workspace.source.coverMode = "type_icon"
                    workspace.source.thumbnailRelPath = ""
                }
                context.workspace = workspace
            }
            context.toast = VocabularyToast(
                title: String(localized: L10n.Sources.toastArtifactDeletedTitle),
                body: L10n.Sources.toastArtifactDeletedBody(ref: request.target.ref),
                tone: .success
            )
            context.notifyWorkspaceMutated()
            pendingImpact = nil
            pendingResourceID = nil
            return false
        } catch {
            context.pageError = L10n.Errors.message(for: error)
            return false
        }
    }

    private func presentImpact(kind: String, id: String, ref: String, title: String) async {
        context.clearPageError()
        do {
            let report = try await context.store.getDeleteImpact(
                projectDir: context.projectDir,
                kind: kind,
                id: id
            )
            pendingResourceID = id
            pendingImpact = PVDeleteImpactRequest(
                target: PVDeleteImpactTarget(kind: kind, ref: ref, title: title),
                report: report
            )
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }

    private func resetSections() {
        identity.reset()
        credibility.resetDrafts()
        metadata.resetDrafts()
        notes.reset()
        artifacts.seedDrafts()
    }
}
