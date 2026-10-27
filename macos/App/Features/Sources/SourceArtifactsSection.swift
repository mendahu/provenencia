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
        /// The unchecked Evidence graphs, shown when the Source has two or more.
        /// State is the unchecked set, so a graph made while the dialog is open
        /// starts checked.
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
        /// Graph id to the hint under its locked row.
        var locked: [String: String] = [:]
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
    /// The Evidence graph choices of the card in edit; nil when the Source has
    /// one graph (or the graph list has no value yet).
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

    /// The Source's Evidence graphs in the session's name order when there are
    /// two or more (the link checkboxes show); nil otherwise, so a one-graph
    /// Source (or a graph list still loading) shows none.
    var visibleGraphs: [CatalogEvidenceGraph]? {
        guard let graphs = context.session.evidenceGraphs(sourceID: context.sourceID),
              graphs.count >= 2
        else { return nil }
        return graphs
    }

    /// The graph ids the Add dialog links: the checked ones in list order.
    /// Empty when the group is hidden (the engine links every graph).
    var draftGraphIDs: [String] {
        guard let graphs = visibleGraphs else { return [] }
        return PVChecklistState.checkedIDs(
            ids: graphs.map(\.id), excluded: draft.excludedGraphIDs, locked: []
        )
    }

    /// The Artifact is not made yet, the group shows and every box is unchecked.
    /// After a refused file the Artifact exists and a retry does not link, so
    /// the boxes no longer matter.
    var draftNeedsGraph: Bool {
        draft.createdID == nil && visibleGraphs != nil && draftGraphIDs.isEmpty
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
            && !draftNeedsGraph
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
    /// When the Source has two or more graphs it also starts reading which of
    /// the Artifact's links the graphs' Citations hold.
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
        if let graphs = visibleGraphs {
            linkEdit = LinkEdit(
                artifactID: id,
                graphIDs: graphs.map(\.id),
                excluded: Set(graphs.filter { !$0.links(id) }.map(\.id))
            )
            linkLoadTask = Task { await loadLinkLocks() }
        }
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

    /// The edit's checked graph ids in list order; a locked graph is always checked.
    var checkedGraphIDs: [String] {
        guard let edit = linkEdit else { return [] }
        return PVChecklistState.checkedIDs(
            ids: liveGraphIDs, excluded: edit.excluded, locked: Set(edit.locked.keys)
        )
    }

    /// The graphs the card in edit offers, read from the session in the order
    /// they were captured; a graph deleted since then drops out.
    var linkEditGraphs: [CatalogEvidenceGraph] {
        (linkEdit?.graphIDs ?? []).compactMap { context.session.evidenceGraph(id: $0) }
    }

    private func linksDirty(_ id: String) -> Bool {
        guard let edit = linkEdit, edit.artifactID == id else { return false }
        return Set(checkedGraphIDs) != Set(linkedGraphIDs)
    }

    /// Reads which graphs must keep the Artifact. A graph whose only Artifact
    /// this is locks locally; every other graph that links it gets one unlink
    /// preflight, in order, and a block locks it with its Citation count. A
    /// result for a card that is no longer in edit is dropped.
    func loadLinkLocks() async {
        guard let edit = linkEdit else { return }
        let artifactID = edit.artifactID
        var locked: [String: String] = [:]
        var reads: [String] = []
        for graphID in edit.graphIDs {
            guard let graph = context.session.evidenceGraph(id: graphID),
                  graph.links(artifactID)
            else { continue }
            if graph.isOnlyArtifact(artifactID) {
                locked[graphID] = L10n.string(L10n.Sources.artifactGraphOnlyArtifact)
            } else {
                reads.append(graphID)
            }
        }
        if reads.isEmpty {
            linkEdit?.locked = locked
            linkEdit?.loadState = .ready
            return
        }
        linkEdit?.loadState = .loading
        do {
            for graphID in reads {
                let blocks = try await context.store.getEvidenceGraphUnlinkImpact(
                    projectDir: context.projectDir,
                    graphID: graphID,
                    artifactIDs: [artifactID]
                )
                if let block = blocks.first(where: { $0.artifactID == artifactID }) {
                    locked[graphID] = PVDeleteImpactCopy.heading(block.citations)
                }
            }
            guard linkEdit?.artifactID == artifactID else { return }
            linkEdit?.locked = locked
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

    /// One Save for the card: the label and description first, then the graph
    /// links (links before unlinks, so the Artifact never passes through zero
    /// graphs needlessly). The first refusal stops the save and shows on the
    /// card; every field and box keeps its value, and Save again sends only
    /// what is left.
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
        var toLink: [String] = []
        var toUnlink: [String] = []
        if let edit = linkEdit, edit.artifactID == id {
            guard edit.loadState == .ready else { return }
            let checked = Set(checkedGraphIDs)
            let linked = Set(linkedGraphIDs)
            let live = liveGraphIDs
            toLink = live.filter { checked.contains($0) && !linked.contains($0) }
            toUnlink = live.filter { !checked.contains($0) && linked.contains($0) }
        }
        let fieldsChanged = label != art.label || desc != art.description
        if !fieldsChanged, toLink.isEmpty, toUnlink.isEmpty { return }
        context.clearPageError()
        cardSaveError = nil
        savingID = id
        defer { savingID = nil }
        var attemptedLink = false
        var wroteAnything = false
        var refusedGraphID: String?
        // The effects of the link writes that succeeded, applied once: each
        // apply reloads the graph list.
        var linkEffects: CatalogWriteEffects?
        do {
            if fieldsChanged {
                let (updated, effects) = try await context.store.updateArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: id,
                    label: label,
                    description: desc
                )
                wroteAnything = true
                replace(updated, effects: effects)
            }
            // Graph links do not change the page payload: only the session's
            // effects apply, as the graph form's writes do.
            for graphID in toLink {
                attemptedLink = true
                refusedGraphID = graphID
                let effects = try await context.store.linkEvidenceGraphArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    graphID: graphID,
                    artifactID: id
                )
                wroteAnything = true
                linkEffects = linkEffects?.merging(effects) ?? effects
            }
            for graphID in toUnlink {
                attemptedLink = true
                refusedGraphID = graphID
                let effects = try await context.store.unlinkEvidenceGraphArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    graphID: graphID,
                    artifactID: id
                )
                wroteAnything = true
                linkEffects = linkEffects?.merging(effects) ?? effects
            }
            refusedGraphID = nil
            if let linkEffects { context.session.apply(linkEffects) }
            if attemptedLink { _ = await graphListReloaded() }
            if editingArtifactID == id {
                editingArtifactID = nil
                dropLinkEdit()
            }
        } catch {
            if let linkEffects { context.session.apply(linkEffects) }
            let message = L10n.Errors.message(for: error)
            // Only the card that saved shows the error: another card may have
            // opened edit while this one saved.
            guard editingArtifactID == id else { return }
            cardSaveError = saveErrorBody(
                message, wroteAnything: wroteAnything, refusedGraphID: refusedGraphID, id: id
            )
            guard attemptedLink, linkEdit?.artifactID == id else { return }
            // `session.apply` only starts the reload: wait for it, then read
            // the locks again from the new links, so a partial save, an
            // in-use Artifact and a last link all show.
            let reloaded = await graphListReloaded()
            let reloadError = context.session.evidenceGraphsHandle?.failure
            if reloaded == nil || reloadError != nil {
                linkEdit?.loadState = .failed(L10n.Errors.message(for: reloadError ?? error))
            } else {
                await loadLinkLocks()
            }
            guard editingArtifactID == id else { return }
            cardSaveError = saveErrorBody(
                message, wroteAnything: wroteAnything, refusedGraphID: refusedGraphID, id: id
            )
        }
    }

    /// The card's save-error body by what is left to save: the graph that
    /// refused when only its refusal remains; "kept" when nothing was written;
    /// the partial-save wording otherwise.
    private func saveErrorBody(
        _ message: String, wroteAnything: Bool, refusedGraphID: String?, id: String
    ) -> String {
        if let refusedGraphID, !fieldsDirty(id),
           let graph = context.session.evidenceGraph(id: refusedGraphID)
        {
            return L10n.Sources.artifactCardSaveErrorGraphKeeps(graph: graph.name, error: message)
        }
        if !wroteAnything {
            return L10n.Sources.artifactCardSaveErrorKept(message)
        }
        return L10n.Sources.artifactCardSaveErrorMessage(message)
    }

    /// Waits for the session's graph list to finish reloading after a link
    /// write; nil when the reload failed.
    private func graphListReloaded() async -> [CatalogEvidenceGraph]? {
        await context.session.readyValue(context.session.evidenceGraphsKey)
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
        guard draft.reject == nil, !draftNeedsGraph else { return }
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
                        description: description
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

    func open(_ art: CatalogArtifact) {
        guard let file = art.file, !file.relPath.isEmpty else { return }
        context.clearPageError()
        if !ProjectFiles.openObject(projectDir: context.projectDir, relPath: file.relPath) {
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
