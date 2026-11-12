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
    var metadata: MetadataBoardModel
    var notes: SourceNotesSection
    var artifacts: SourceArtifactsSection
    var graphs: SourceGraphsSection
    /// The page's one Add-metadata dialog and clear confirm, for every board on it.
    let metadataDialogs = MetadataBoardDialogs()

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
        metadata = MetadataBoardModel(host: SourcePageMetadataHost(context: context))
        notes = SourceNotesSection(context: context)
        artifacts = SourceArtifactsSection(context: context)
        graphs = SourceGraphsSection(context: context)
    }

    var workspace: CatalogSourceWorkspace? { context.workspace }

    var source: CatalogSource? { context.source }

    /// Project folder for resolving `objects/…` thumbnail paths.
    var pageProjectDir: String { context.projectDir }

    /// Non-field failure banner shared by the page sections.
    var pageError: SourcePageContext.PageError? { context.pageError }
    var pageErrorShownCount: Int { context.pageErrorShownCount }

    /// The notice a write set when it found this Source deleted elsewhere; the page leaves while it is set.
    var gone: String? { context.gone }

    /// The notice to show as the page leaves: the write path's (`context.gone`), else the
    /// "no longer exists" copy when the workspace read failed with `sources.not_found`. Pure and
    /// read at body time, so a Back visit to a deleted Source already has it on the first pass.
    func goneNotice(read: QueryRead<CatalogSourceWorkspace>) -> String? {
        if let gone = context.gone { return gone }
        if case .failed(let error) = read,
           (error as? CoreInvokeError)?.engineCode == .sourceNotFound
        {
            return L10n.string(L10n.Errors.sourcesNotFound)
        }
        return nil
    }

    var toast: VocabularyToast? {
        get { context.toast }
        set { context.toast = newValue }
    }

    /// The content pane takes a file drop when no sheet is already up.
    var acceptsArtifactPageDrop: Bool {
        !artifacts.isAdding
            && !artifacts.isAttaching
            && !metadataDialogs.isAdding
            && metadataDialogs.pendingDelete == nil
            && notes.pendingDelete == nil
            && pendingImpact == nil
            && artifacts.pendingEditSwitch == nil
    }

    /// Loads the workspace and page vocabulary, then syncs section state
    /// (tests and previews; navigation warms the same keys via `PlaceRegistry`).
    func warmFromSession() async {
        let session = context.session
        let key = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: context.sourceID)
        let handle: QueryHandle<CatalogSourceWorkspace> = session.reread(key)
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
        sync(from: handle)
    }

    /// Mirrors `QueryHandle` state into the page observation graph.
    func sync(from handle: QueryHandle<CatalogSourceWorkspace>) {
        isLoading = handle.status == .loading && handle.value == nil
        if case .failed(let error) = handle.read {
            loadError = error
        } else {
            loadError = nil
        }
        guard let workspace = handle.value else { return }
        apply(workspace)
    }

    private func apply(_ workspace: CatalogSourceWorkspace) {
        let previous = appliedWorkspace
        // The board writes its rows into `context.workspace` itself, so
        // "differs from what the board shows" is read from it, not from `previous`.
        let shown = context.workspace
        context.workspace = workspace
        context.clearPageError()
        identity.resetIfChanged(from: previous?.source, to: workspace.source)
        if previous?.credibility != workspace.credibility {
            credibility.resetDrafts()
        }
        let sourceOwner = MetadataOwner(level: .source, id: context.sourceID)
        metadata.resetIfChanged(
            from: shown?.metadataRows(for: sourceOwner), to: workspace.metadataRows(for: sourceOwner)
        )
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
        guard let graph = graphs.ownGraph(id: id) else { return }
        await presentImpact(kind: "evidence_graph", id: graph.id, ref: graph.ref, title: graph.name)
    }

    func askDeleteArtifact(id: String) async {
        guard let art = artifacts.items.first(where: { $0.id == id }) else { return }
        await presentImpact(kind: "artifact", id: art.id, ref: art.ref, title: art.displayTitle)
    }

    /// Returns true when the Source itself was erased. `leave` runs once, inside the write's
    /// application, when it was; a failed write does not run it.
    @discardableResult
    func confirmResourceDelete(leaving leave: () -> Void) async -> Bool {
        let kind = deleteImpact.request?.target.kind
        let erasedSource = await deleteImpact.confirm { target in
            switch target.kind {
            case "source":
                let effects = try await context.store.deleteSource(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: target.id
                )
                context.session.apply(effects, .deletedSource(id: target.id), leavingPage: leave)
            case "artifact":
                let effects = try await context.store.deleteArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: target.id
                )
                if var workspace = context.workspace {
                    workspace.artifacts.removeAll { $0.id == target.id }
                    apply(workspace)
                    context.setWorkspace(workspace)
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
            context.showPageError(message)
        }
    }
}
