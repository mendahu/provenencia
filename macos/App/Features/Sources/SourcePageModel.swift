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
    var graphs: SourceGraphsSection

    /// The Copy Source form while it is open.
    var copying: SourceCopyModel?

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
        graphs = SourceGraphsSection(context: context)
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

    /// The content pane takes a file drop when no sheet is already up.
    var acceptsArtifactPageDrop: Bool {
        !artifacts.isAdding
            && !artifacts.isAttaching
            && !metadata.isAdding
            && metadata.pendingDelete == nil
            && notes.pendingDelete == nil
            && pendingImpact == nil
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
        let _: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
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

    /// Opens the Copy Source form for this page's Source.
    func openCopy() {
        guard let source else { return }
        copying = SourceCopyModel(
            original: source,
            session: context.session,
            userID: context.userID,
            store: context.store
        )
    }

    func askDeleteSource() async {
        guard let source else { return }
        await presentImpact(kind: "source", id: source.id, ref: source.ref, title: source.title)
    }

    func askDeleteGraph(id: String) async {
        guard let graph = graphs.rows?.first(where: { $0.id == id }) else { return }
        await presentImpact(kind: "evidence_graph", id: graph.id, ref: graph.ref, title: graph.name)
    }

    func askDeleteArtifact(id: String) async {
        guard let art = artifacts.items.first(where: { $0.id == id }) else { return }
        await presentImpact(kind: "artifact", id: art.id, ref: art.ref, title: art.displayTitle)
    }

    /// Returns true when the Source itself was erased (leave the page).
    @discardableResult
    func confirmResourceDelete() async -> Bool {
        let kind = deleteImpact.request?.target.kind
        let erasedSource = await deleteImpact.confirm { target in
            switch target.kind {
            case "source":
                try await context.store.deleteSource(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: target.id
                )
                context.session.apply(.deletedSource(id: target.id))
            case "artifact":
                let effects = try await context.store.deleteArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: target.id
                )
                if var workspace = context.workspace {
                    workspace.artifacts.removeAll { $0.id == target.id }
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
                    title: L10n.string(L10n.Sources.toastArtifactDeletedTitle),
                    body: L10n.Sources.toastArtifactDeletedBody(ref: target.ref),
                    tone: .success
                )
                // The refresh below is the one read of the Source's workspace: it
                // carries the cover and `hasArtifact`, which the activity patch
                // cannot.
                context.session.apply(effects)
                await context.refreshCoverFromStore()
            case "evidence_graph":
                // The effects refresh the graph list; a graph changes no
                // workspace field and no cover.
                let effects = try await context.store.deleteEvidenceGraph(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    graphID: target.id
                )
                context.toast = VocabularyToast(
                    title: L10n.string(L10n.Sources.toastGraphDeletedTitle),
                    body: L10n.Sources.toastGraphDeletedBody(ref: target.ref),
                    tone: .success
                )
                context.session.apply(effects)
            default:
                return
            }
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
