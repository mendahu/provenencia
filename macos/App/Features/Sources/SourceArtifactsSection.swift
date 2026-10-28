import Foundation
import Observation

/// Artifacts on the Source page: one card in edit at a time, the Add-artifact
/// dialog, and file attach/open.
@MainActor
@Observable
final class SourceArtifactsSection {
    struct Draft: Equatable {
        var label = ""
        var description = ""
        /// Absolute path when validation passed; nil when rejected or cleared.
        var filePath: String?
        var fileName: String?
        var reject: L10n.Errors.IngestCallout?
        /// The Artifact an earlier confirm created before its file was
        /// refused. A retry saves onto it instead of creating another.
        var createdID: String?
        /// The unchecked Evidence graphs, shown whenever the session has the
        /// Source's graphs. State is the unchecked set, so a graph made while
        /// the dialog is open starts checked.
        var excludedGraphIDs: Set<String> = []
    }

    /// The Evidence graph choices of the card in edit: ids and choices only.
    /// Rows, names, refs and current links are read from the session's graph
    /// list each time, so a write elsewhere shows at once.
    struct LinkEdit: Equatable {
        let artifactID: String
        /// The Source's graphs when edit opened, in the session's name order.
        let graphIDs: [String]
        /// The unchecked graph ids, seeded with the graphs that do not link the Artifact.
        var excluded: Set<String>
        /// The engine's refusals to unlink this Artifact, from its preflight. Each
        /// locks its graph; Swift decides no lock.
        var blocks: [CatalogEvidenceGraphUnlinkBlock] = []
        var loadState: SourceFormLoadState = .loading
    }

    struct AttachDraft: Equatable {
        var artifactID = ""
        var artifactRef = ""
        var filePath: String?
        var fileName: String?
        var reject: L10n.Errors.IngestCallout?
    }

    /// The card whose label and description are fields. One at a time.
    var editingArtifactID: String?
    /// Local edit buffers for the card in edit, keyed by id.
    var labels: [String: String] = [:]
    var descriptions: [String: String] = [:]
    /// Per-artifact validation for the card label.
    var fieldErrors: [String: String] = [:]
    /// Artifact id currently saving label/description from the card.
    private(set) var savingID: String?
    /// The Evidence graph choices of the card in edit; nil when the session has
    /// no graphs for the Source (the list is loading, failed or not refreshed).
    var linkEdit: LinkEdit?
    /// Why the last card save stopped, shown on the card; nil otherwise.
    var cardSaveError: String?
    /// The card's graph link read, so a test can await it.
    @ObservationIgnored var linkLoadTask: Task<Void, Never>?

    // MARK: Add dialog

    var isAdding = false
    var draft = Draft()
    var draftLabelError: String?
    private(set) var isSavingDraft = false

    // MARK: Add file dialog

    var isAttaching = false
    var attachDraft = AttachDraft()
    private(set) var isSavingAttach = false

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var items: [CatalogArtifact] { context.workspace?.artifacts ?? [] }

    /// The Source's Evidence graphs in the session's name order when the
    /// session has them (the link checkboxes show); nil otherwise (the list is
    /// loading, failed, or not refreshed), so no group shows.
    var visibleGraphs: [CatalogEvidenceGraph]? {
        guard case .loaded(let graphs) = context.session.evidenceGraphsRead(sourceID: context.sourceID)
        else { return nil }
        return graphs
    }

    /// The graph ids the Add dialog links: the checked ones in list order,
    /// possibly none. Nil when the group is hidden (the engine links every graph).
    var draftGraphIDs: [String]? {
        guard let graphs = visibleGraphs else { return nil }
        return PVChecklistState.checkedIDs(
            ids: graphs.map(\.id), excluded: draft.excludedGraphIDs, locked: []
        )
    }

    /// The Artifact is not made yet and no box is checked: it will be in no
    /// graph. After a refused file the Artifact exists and a retry does not
    /// link, so the boxes no longer matter.
    var draftInNoGraph: Bool {
        draft.createdID == nil && draftGraphIDs == []
    }

    /// The Add dialog's graph boxes are read-only: a retry after a refused
    /// file saves onto the Artifact already made, which keeps its links.
    var draftGraphsLocked: Bool {
        draft.createdID != nil
    }

    var canSubmitDraft: Bool {
        !isSavingDraft
            && !draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            && draft.reject == nil
    }

    var canSubmitAttach: Bool {
        !isSavingAttach
            && attachDraft.filePath != nil
            && attachDraft.reject == nil
    }

    var addSheetIsDirty: Bool {
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        return !label.isEmpty || !description.isEmpty || draft.filePath != nil || draftGraphsChanged
    }

    /// A box was unchecked since the dialog opened.
    private var draftGraphsChanged: Bool {
        guard let graphs = visibleGraphs else { return false }
        return !PVChecklistState.liveExcluded(
            ids: graphs.map(\.id), excluded: draft.excludedGraphIDs, locked: []
        ).isEmpty
    }

    /// The artifact the file will attach to is a seed. A chosen file is an edit.
    var attachSheetIsDirty: Bool {
        attachDraft.filePath != nil
    }

    /// The page's Artifacts that none of the Source's graphs links; empty while
    /// the session has no graphs for the Source, so no card is marked.
    var inNoGraph: Set<String> {
        context.session.artifactIDsInNoGraph(sourceID: context.sourceID, artifactIDs: items.map(\.id)) ?? []
    }

    /// Seeds edit buffers for artifacts that do not have one yet.
    func seedDrafts() {
        for art in items {
            if labels[art.id] == nil {
                labels[art.id] = art.label
            }
            if descriptions[art.id] == nil {
                descriptions[art.id] = art.description
            }
        }
        if let id = editingArtifactID, !items.contains(where: { $0.id == id }) {
            editingArtifactID = nil
            dropLinkEdit()
        }
    }

    // MARK: Card edit

    /// Opens edit on one card and restores any other card to its stored text.
    /// When the session has the Source's graphs it also starts reading which of
    /// the Artifact's links the engine says its graphs must keep.
    func beginEdit(_ id: String) {
        guard let art = items.first(where: { $0.id == id }) else { return }
        if let current = editingArtifactID, current != id {
            cancelFields(id: current)
        }
        editingArtifactID = id
        labels[id] = art.label
        descriptions[id] = art.description
        fieldErrors[id] = nil
        cardSaveError = nil
        dropLinkEdit()
        startLinkEdit(id: id)
    }

    /// Builds the card's graph choices and starts its lock read, when the
    /// session has the Source's graphs.
    private func startLinkEdit(id: String) {
        guard let graphs = visibleGraphs else { return }
        linkEdit = LinkEdit(
            artifactID: id,
            graphIDs: graphs.map(\.id),
            excluded: Set(graphs.filter { !$0.links(id) }.map(\.id))
        )
        linkLoadTask = Task { await loadLinkLocks() }
    }

    /// True while the card `id` is in edit and the graph list has not arrived,
    /// so its graph group shows the loading state.
    func awaitingGraphs(id: String) -> Bool {
        guard editingArtifactID == id, linkEdit == nil else { return false }
        if case .loading = context.session.evidenceGraphsRead(sourceID: context.sourceID) { return true }
        return false
    }

    /// The graph list changed: a card opened in edit before it arrived gets
    /// its group now, as `beginEdit` would have built it.
    func graphsChanged() {
        guard let id = editingArtifactID, linkEdit == nil else { return }
        startLinkEdit(id: id)
    }

    private func dropLinkEdit() {
        linkLoadTask?.cancel()
        linkLoadTask = nil
        linkEdit = nil
    }

    /// The edit's graph ids whose session graph links the Artifact now.
    var linkedGraphIDs: [String] {
        guard let edit = linkEdit else { return [] }
        return edit.graphIDs.filter {
            context.session.evidenceGraph(id: $0)?.links(edit.artifactID) == true
        }
    }

    /// The edit's graph ids that are still in the session's graph list; a
    /// graph deleted since edit opened drops out of every count and write.
    private var liveGraphIDs: [String] {
        (linkEdit?.graphIDs ?? []).filter { context.session.evidenceGraph(id: $0) != nil }
    }

    /// Graph id to the hint under its locked row: one lock per graph, from the
    /// first block that names it.
    var lockedGraphs: [String: String] {
        (linkEdit?.blocks ?? []).lockHints(by: \.graphID)
    }

    /// The edit's checked graph ids in list order; a locked graph is always checked.
    var checkedGraphIDs: [String] {
        guard let edit = linkEdit else { return [] }
        return PVChecklistState.checkedIDs(
            ids: liveGraphIDs, excluded: edit.excluded, locked: Set(lockedGraphs.keys)
        )
    }

    /// The graphs the card in edit offers, read from the session in the order
    /// they were captured; a graph deleted since then drops out.
    var linkEditGraphs: [CatalogEvidenceGraph] {
        (linkEdit?.graphIDs ?? []).compactMap { context.session.evidenceGraph(id: $0) }
    }

    /// The card in edit has its graph boxes read and none is checked: saving
    /// leaves the Artifact in no graph.
    func cardInNoGraph(id: String) -> Bool {
        guard let edit = linkEdit, edit.artifactID == id, edit.loadState == .ready else { return false }
        return checkedGraphIDs.isEmpty
    }

    private func linksDirty(_ id: String) -> Bool {
        guard let edit = linkEdit, edit.artifactID == id else { return false }
        return Set(checkedGraphIDs) != Set(linkedGraphIDs)
    }

    /// Reads which graphs must keep the Artifact: one unlink preflight over a
    /// pair for every graph that links it. Every block the engine reports locks
    /// its graph. A result for a card that is no longer in edit is dropped.
    func loadLinkLocks() async {
        guard let edit = linkEdit else { return }
        let artifactID = edit.artifactID
        let pairs = edit.graphIDs.compactMap { graphID -> CatalogEvidenceGraphArtifactPair? in
            guard context.session.evidenceGraph(id: graphID)?.links(artifactID) == true else { return nil }
            return CatalogEvidenceGraphArtifactPair(graphID: graphID, artifactID: artifactID)
        }
        if pairs.isEmpty {
            linkEdit?.blocks = []
            linkEdit?.loadState = .ready
            return
        }
        linkEdit?.loadState = .loading
        do {
            let blocks = try await context.store.getEvidenceGraphUnlinkImpact(
                projectDir: context.projectDir,
                pairs: pairs
            )
            guard linkEdit?.artifactID == artifactID else { return }
            linkEdit?.blocks = blocks.filter { $0.artifactID == artifactID }
            linkEdit?.loadState = .ready
        } catch {
            guard linkEdit?.artifactID == artifactID else { return }
            linkEdit?.loadState = .failed(L10n.Errors.message(for: error))
        }
    }

    func fieldsDirty(_ id: String) -> Bool {
        guard let art = items.first(where: { $0.id == id }) else { return false }
        return textChanged(id, art) || linksDirty(id)
    }

    private func textChanged(_ id: String, _ art: CatalogArtifact) -> Bool {
        let label = (labels[id] ?? art.label).trimmingCharacters(in: .whitespacesAndNewlines)
        let desc = (descriptions[id] ?? art.description).trimmingCharacters(in: .whitespacesAndNewlines)
        return label != art.label || desc != art.description
    }

    func canSaveFields(_ id: String) -> Bool {
        guard savingID == nil, fieldsDirty(id) else { return false }
        if let edit = linkEdit, edit.artifactID == id, edit.loadState != .ready { return false }
        let label = (labels[id] ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        return !label.isEmpty
    }

    /// Shown while the card is in edit and the label has no text.
    func cardLabelError(id: String) -> String? {
        guard editingArtifactID == id else { return fieldErrors[id] }
        let label = (labels[id] ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            return L10n.string(L10n.Sources.artifactCardLabelRequired)
        }
        return fieldErrors[id]
    }

    func cancelFields(id: String) {
        guard savingID != id else { return }
        if linkEdit?.artifactID == id { dropLinkEdit() }
        if editingArtifactID == id { cardSaveError = nil }
        guard let art = items.first(where: { $0.id == id }) else {
            if editingArtifactID == id { editingArtifactID = nil }
            return
        }
        labels[id] = art.label
        descriptions[id] = art.description
        fieldErrors[id] = nil
        if editingArtifactID == id { editingArtifactID = nil }
    }

    /// The full set of graph ids a links Save sends: the checked ones in list
    /// order, then the graphs made since edit opened that link the Artifact (the
    /// card never showed them, so a Save must not drop them).
    private var savedGraphIDs: [String] {
        guard let edit = linkEdit else { return [] }
        let shown = Set(edit.graphIDs)
        let added = (context.session.evidenceGraphs(sourceID: context.sourceID) ?? [])
            .filter { !shown.contains($0.id) && $0.links(edit.artifactID) }
            .map(\.id)
        return checkedGraphIDs + added
    }

    /// One Save for the card: one `updateArtifact` with the label, the
    /// description and, when a box changed, every link, so the engine applies
    /// all of it or none. A refusal leaves the card in edit with every field
    /// and box as the researcher set it and nothing half applied.
    func saveFields(id: String) async {
        guard let art = items.first(where: { $0.id == id }) else { return }
        guard savingID == nil else { return }
        let label = (labels[id] ?? art.label).trimmingCharacters(in: .whitespacesAndNewlines)
        let desc = (descriptions[id] ?? art.description).trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            fieldErrors[id] = L10n.string(L10n.Sources.artifactCardLabelRequired)
            return
        }
        fieldErrors[id] = nil
        var graphIDs: [String]?
        if let edit = linkEdit, edit.artifactID == id {
            guard edit.loadState == .ready else { return }
            if linksDirty(id) { graphIDs = savedGraphIDs }
        }
        let fieldsChanged = label != art.label || desc != art.description
        if !fieldsChanged, graphIDs == nil { return }
        context.clearPageError()
        cardSaveError = nil
        savingID = id
        defer { savingID = nil }
        do {
            let (updated, effects) = try await context.store.updateArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id,
                label: label,
                description: desc,
                graphIDs: graphIDs
            )
            // Its effects reload the graph list when links moved.
            replace(updated, effects: effects)
            if editingArtifactID == id {
                editingArtifactID = nil
                dropLinkEdit()
            }
        } catch {
            let message = L10n.Errors.message(for: error)
            // Read the locks again: a Citation may have taken the Artifact
            // since edit opened, and the engine's preflight says which graph keeps it.
            if graphIDs != nil, linkEdit?.artifactID == id {
                await loadLinkLocks()
            }
            // Only the card that saved shows the error: another card may have
            // opened edit while this one saved.
            guard editingArtifactID == id else { return }
            cardSaveError = saveErrorBody(message)
        }
    }

    /// The card's save-error body. Nothing was written, so it never speaks of a
    /// partial save: it names the graph the engine says keeps the Artifact when
    /// the researcher unchecked it, else says the changes are kept.
    private func saveErrorBody(_ message: String) -> String {
        if let edit = linkEdit, edit.loadState == .ready {
            let locked = lockedGraphs
            let kept = edit.graphIDs.first { edit.excluded.contains($0) && locked[$0] != nil }
            if let kept, let graph = context.session.evidenceGraph(id: kept) {
                return L10n.Sources.artifactCardSaveErrorGraphKeeps(graph: graph.name, error: message)
            }
        }
        return L10n.Sources.artifactCardSaveErrorKept(message)
    }

    /// Opens Add Artifact with the dropped file staged. Nothing is stored until confirm.
    func stagePageDrop(_ urls: [URL]) {
        guard !urls.isEmpty, savingID == nil, !isSavingDraft else { return }
        if let id = editingArtifactID {
            cancelFields(id: id)
        }
        openAdd()
        applyDroppedURLs(urls, into: .create)
    }

    // MARK: Add dialog

    func openAdd() {
        draft = Draft()
        draftLabelError = nil
        isAdding = true
    }

    func cancelAdd() {
        guard !isSavingDraft else { return }
        isAdding = false
        draftLabelError = nil
    }

    func pickFile() {
        let path = ProjectFiles.pickFileForIngest(
            prompt: L10n.string(L10n.Sources.filePickPrompt),
            message: L10n.string(L10n.Sources.filePickMessage)
        )
        guard let path else { return }
        applyPickedFile(url: URL(fileURLWithPath: path), into: .create)
    }

    func applyDroppedURLs(_ urls: [URL], into target: FilePickTarget) {
        if urls.count > 1 {
            setReject(
                L10n.Errors.ingestCallout(reason: .multiFile),
                fileName: urls.first?.lastPathComponent,
                into: target
            )
            return
        }
        guard let url = urls.first else { return }
        applyPickedFile(url: url, into: target)
    }

    func applyPickedFile(url: URL, into target: FilePickTarget) {
        let name = url.lastPathComponent
        switch IngestMediaPolicy.validate(url: url) {
        case .ok:
            switch target {
            case .create:
                draft.filePath = url.path
                draft.fileName = name
                draft.reject = nil
            case .attach:
                attachDraft.filePath = url.path
                attachDraft.fileName = name
                attachDraft.reject = nil
            }
        case .reject(let reason):
            let size = reason == .tooLarge
                ? IngestMediaPolicy.formatByteSize(IngestMediaPolicy.byteSize(of: url))
                : nil
            setReject(
                L10n.Errors.ingestCallout(reason: reason, sizeLabel: size),
                fileName: name,
                into: target
            )
        }
    }

    func clearFile() {
        draft.filePath = nil
        draft.fileName = nil
        draft.reject = nil
    }

    func create() async {
        guard !isSavingDraft else { return }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            draftLabelError = L10n.string(L10n.Sources.artifactLabelRequired)
            return
        }
        guard draft.reject == nil else { return }
        isSavingDraft = true
        defer { isSavingDraft = false }
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        do {
            var art: CatalogArtifact
            if let id = draft.createdID, let existing = items.first(where: { $0.id == id }) {
                art = existing
                if label != art.label || description != art.description {
                    let (updated, effects) = try await context.store.updateArtifact(
                        projectDir: context.projectDir,
                        userID: context.userID,
                        artifactID: id,
                        label: label,
                        description: description,
                        graphIDs: nil
                    )
                    art = updated
                    show(art, effects: effects)
                }
            } else {
                let (created, effects) = try await context.store.createArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: context.sourceID,
                    fileID: "",
                    label: label,
                    description: description,
                    // Nil when the boxes are hidden: the engine links every graph.
                    graphIDs: draftGraphIDs
                )
                art = created
                // Shown before the file is ingested: if the file is refused,
                // the Artifact Go created still reads on the page, and its
                // graph links (which the ingest does not report) still apply.
                show(art, effects: effects)
                draft.createdID = art.id
            }
            if let path = draft.filePath {
                let ingested = try await context.store.ingestArtifactFile(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: art.id,
                    path: path
                )
                art = ingested.artifact
                show(art, effects: ingested.effects)
            }
            isAdding = false
            draftLabelError = nil
            draft.reject = nil
            draft.createdID = nil
            await context.refreshCoverFromStore()
            context.toast = VocabularyToast(
                title: L10n.Sources.toastArtifactCreatedTitle(ref: art.ref),
                body: art.fileID.isEmpty
                    ? L10n.Sources.toastArtifactCreatedFileless
                    : L10n.Sources.toastArtifactCreatedWithFile,
                tone: .success
            )
        } catch {
            if let coded = error as? CoreInvokeError, case .coded(_, let code, _, let params) = coded,
               let callout = L10n.Errors.ingestCallout(code: code, params: params)
            {
                draft.reject = callout
                draft.filePath = nil
                draftLabelError = nil
            } else {
                draftLabelError = L10n.Errors.message(for: error)
            }
        }
    }

    // MARK: Cover

    func isCover(_ art: CatalogArtifact) -> Bool {
        guard let source = context.source else { return false }
        return source.coverMode == "artifact" && source.primaryArtifactID == art.id
    }

    func canUseAsThumbnail(_ art: CatalogArtifact) -> Bool {
        !art.thumbnailRelPath.isEmpty && !isCover(art)
    }

    func useAsThumbnail(_ art: CatalogArtifact) async {
        guard canUseAsThumbnail(art) else { return }
        let ok = await context.setSourceCover(mode: "artifact", primaryArtifactID: art.id)
        guard ok else { return }
        context.toast = VocabularyToast(
            title: L10n.Sources.toastThumbnailUpdatedTitle,
            body: L10n.Sources.toastThumbnailUpdatedArtifactBody(ref: art.ref),
            tone: .success
        )
    }

    func revertCoverToTypeIcon() async {
        guard context.source?.coverMode != "type_icon" else { return }
        let ok = await context.setSourceCover(mode: "type_icon")
        guard ok else { return }
        context.toast = VocabularyToast(
            title: L10n.Sources.toastThumbnailUpdatedTitle,
            body: L10n.Sources.toastThumbnailUpdatedTypeIconBody,
            tone: .success
        )
    }

    // MARK: Files

    enum FilePickTarget {
        case create
        case attach
    }

    func openAttach(to id: String) {
        guard let art = items.first(where: { $0.id == id }), art.fileID.isEmpty else { return }
        attachDraft = AttachDraft(artifactID: id, artifactRef: art.ref)
        isAttaching = true
    }

    func cancelAttach() {
        guard !isSavingAttach else { return }
        isAttaching = false
        attachDraft = AttachDraft()
    }

    func pickAttachFile() {
        let path = ProjectFiles.pickFileForIngest(
            prompt: L10n.string(L10n.Sources.filePickPrompt),
            message: L10n.string(L10n.Sources.filePickMessage)
        )
        guard let path else { return }
        applyPickedFile(url: URL(fileURLWithPath: path), into: .attach)
    }

    func clearAttachFile() {
        attachDraft.filePath = nil
        attachDraft.fileName = nil
        attachDraft.reject = nil
    }

    func confirmAttach() async {
        guard canSubmitAttach else { return }
        guard let path = attachDraft.filePath else { return }
        let id = attachDraft.artifactID
        isSavingAttach = true
        defer { isSavingAttach = false }
        context.clearPageError()
        do {
            let ingested = try await context.store.ingestArtifactFile(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id,
                path: path
            )
            replace(ingested.artifact, effects: ingested.effects)
            isAttaching = false
            attachDraft = AttachDraft()
            await context.refreshCoverFromStore()
            context.toast = VocabularyToast(
                title: L10n.Sources.toastFileAttachedTitle,
                body: L10n.Sources.toastFileAttachedBody(name: ingested.file.originalFilename),
                tone: .success
            )
        } catch {
            if let coded = error as? CoreInvokeError, case .coded(_, let code, _, let params) = coded,
               let callout = L10n.Errors.ingestCallout(code: code, params: params)
            {
                attachDraft.reject = callout
                attachDraft.filePath = nil
            } else {
                context.pageError = L10n.Errors.message(for: error)
                isAttaching = false
            }
        }
    }

    /// Opens the Artifact's file in its own app. A file with no path (Go returns none
    /// for a type Provenencia does not store) or one `ProjectFiles` refuses shows the
    /// shared refusal naming the file; a missing one keeps its own message.
    func open(_ art: CatalogArtifact) {
        guard let file = art.file else { return }
        context.clearPageError()
        let name = art.fileDisplayName
        guard !file.relPath.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            context.pageError = L10n.ObjectOpen.refused(filename: name)
            return
        }
        switch ProjectFiles.openObject(projectDir: context.projectDir, relPath: file.relPath) {
        case nil:
            break
        case .unsafe:
            context.pageError = L10n.ObjectOpen.refused(filename: name)
        case .missing:
            context.pageError = L10n.string(L10n.Sources.fileOpenMissing)
        }
    }

    private func setReject(
        _ callout: L10n.Errors.IngestCallout,
        fileName: String?,
        into target: FilePickTarget
    ) {
        switch target {
        case .create:
            draft.filePath = nil
            draft.fileName = fileName
            draft.reject = callout
        case .attach:
            attachDraft.filePath = nil
            attachDraft.fileName = fileName
            attachDraft.reject = callout
        }
    }

    /// Puts a written Artifact on the page, appended when new. The list is
    /// patched here, so only the effects apply: the Source's row moves and
    /// its graph reads follow.
    private func show(_ art: CatalogArtifact, effects: CatalogWriteEffects) {
        if let idx = context.workspace?.artifacts.firstIndex(where: { $0.id == art.id }) {
            context.workspace?.artifacts[idx] = art
        } else {
            context.workspace?.artifacts.append(art)
        }
        labels[art.id] = art.label
        descriptions[art.id] = art.description
        context.session.apply(effects)
    }

    private func replace(_ updated: CatalogArtifact, effects: CatalogWriteEffects) {
        if let idx = context.workspace?.artifacts.firstIndex(where: { $0.id == updated.id }) {
            context.workspace?.artifacts[idx] = updated
        }
        labels[updated.id] = updated.label
        descriptions[updated.id] = updated.description
        context.applyWrite(effects)
    }
}
