import Foundation
import Observation

/// State for the individual Source page: hydrates the workspace from
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
    let deleteImpact = DeleteImpactFlow()
    var pendingImpact: PVDeleteImpactRequest? {
        get { deleteImpact.request }
        set { deleteImpact.applyRequest(newValue) }
    }
    var isDeletingResource: Bool { deleteImpact.isRunning }

    /// Session contributor name for the note composer byline.
    let sessionDisplayName: String

    private let context: SourcePageContext
    /// Last workspace applied to section drafts. Slice-equal fields keep in-progress edits.
    private var appliedWorkspace: CatalogSourceWorkspace?

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
        appliedWorkspace = nil
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
        apply(workspace)
    }

    private func apply(_ workspace: CatalogSourceWorkspace) {
        let previous = appliedWorkspace
        context.workspace = workspace
        context.pageError = nil
        if previous?.source != workspace.source {
            identity.reset()
        }
        if previous?.credibility != workspace.credibility {
            credibility.resetDrafts()
        }
        if previous?.metadata != workspace.metadata {
            metadata.resetDrafts()
        }
        if previous?.notes != workspace.notes {
            notes.reset()
        }
        if previous?.artifacts != workspace.artifacts {
            artifacts.seedDrafts()
        }
        appliedWorkspace = workspace
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
        let kind = deleteImpact.request?.target.kind
        let erasedSource = await deleteImpact.confirm { target in
            if target.kind == "source" {
                try await context.store.deleteSource(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: target.id
                )
                context.session.apply(.deletedSource(id: target.id))
                return
            }
            try await context.store.deleteArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: target.id
            )
            if var workspace = context.workspace {
                workspace.artifacts.removeAll { $0.id == target.id }
                if workspace.source.primaryArtifactID == target.id {
                    workspace.source.primaryArtifactID = ""
                    workspace.source.coverMode = "type_icon"
                    workspace.source.thumbnailRelPath = ""
                }
                apply(workspace)
                context.session.setQueryValue(
                    CatalogQueryKey.sourceWorkspace(
                        project: context.session.projectKey,
                        sourceId: context.sourceID
                    ),
                    value: workspace
                )
            }
            context.toast = VocabularyToast(
                title: String(localized: L10n.Sources.toastArtifactDeletedTitle),
                body: L10n.Sources.toastArtifactDeletedBody(ref: target.ref),
                tone: .success
            )
            context.notifyWorkspaceMutated()
        }
        return erasedSource && kind == "source"
    }

    private func presentImpact(kind: String, id: String, ref: String, title: String) async {
        context.clearPageError()
        await deleteImpact.ask(kind: kind, id: id, ref: ref, title: title) {
            try await self.context.store.getDeleteImpact(
                projectDir: self.context.projectDir,
                kind: kind,
                id: id
            )
        }
        if deleteImpact.request == nil, let message = deleteImpact.error {
            context.pageError = message
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
