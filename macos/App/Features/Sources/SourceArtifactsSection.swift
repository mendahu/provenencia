import Foundation
import Observation

/// Artifacts on the Source page: the Artifact sheet, the Add-artifact
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

    struct AttachDraft: Equatable {
        var artifactID = ""
        var artifactRef = ""
        var filePath: String?
        var fileName: String?
        var reject: L10n.Errors.IngestCallout?
    }

    // MARK: Add dialog

    var isAdding = false
    var draft = Draft()
    var draftLabelError: String?
    private(set) var isSavingDraft = false

    // MARK: Add file dialog

    var isAttaching = false
    var attachDraft = AttachDraft()
    private(set) var isSavingAttach = false

    /// The Artifact sheet; nil when closed.
    var sheet: ArtifactSheetModel?

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var items: [CatalogArtifact] { context.workspace?.artifacts ?? [] }

    /// The Source's Evidence graphs when the session has them (the link
    /// checkboxes show); nil otherwise, so no group shows.
    var visibleGraphs: [CatalogEvidenceGraph]? { context.evidenceGraphs }

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

    /// The Artifact's saved copy details, read from the workspace the page already holds.
    func copyDetails(for artifactID: String) -> [CatalogMetadataEntry] {
        context.workspace?.metadataRows(for: MetadataOwner(level: .artifact, id: artifactID)) ?? []
    }

    /// The page's workspace changed: a sheet whose Artifact is no longer on the
    /// page closes with its draft and says so. An Artifact still there keeps its draft.
    func workspaceChanged() {
        guard let sheet, !items.contains(where: { $0.id == sheet.artifact.id }) else { return }
        self.sheet = nil
        context.toast = VocabularyToast(
            title: L10n.string(L10n.Sources.toastArtifactGoneTitle),
            body: L10n.Sources.toastArtifactGoneBody(label: sheet.openedLabel),
            tone: .info
        )
    }

    /// Opens the Artifact sheet at `section`. Nothing happens while a sheet is
    /// open or when the Artifact is not on the page.
    func openSheet(artifactID: String, at section: ArtifactSheetModel.Section = .top) {
        guard sheet == nil, let art = items.first(where: { $0.id == artifactID }) else { return }
        sheet = ArtifactSheetModel(
            context: context,
            mode: .edit(art),
            opensAt: section,
            onSaved: { [weak self] art, metadata, effects in
                self?.show(art, metadata: metadata, effects: effects)
            }
        )
    }

    /// Opens Add Artifact with the dropped file staged. Nothing is stored until confirm.
    func stagePageDrop(_ urls: [URL]) {
        guard !urls.isEmpty, sheet == nil, !isSavingDraft else { return }
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
                    let (updated, _, effects) = try await context.store.updateArtifact(
                        projectDir: context.projectDir,
                        userID: context.userID,
                        artifactID: id,
                        label: label,
                        description: description,
                        graphIDs: nil,
                        metadata: nil
                    )
                    art = updated
                    show(art, effects: effects)
                }
            } else {
                let (created, _, effects) = try await context.store.createArtifact(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    sourceID: context.sourceID,
                    fileID: "",
                    label: label,
                    description: description,
                    // Nil when the boxes are hidden: the engine links every graph.
                    graphIDs: draftGraphIDs,
                    metadata: nil
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
            if let coded = error as? CoreInvokeError, let code = coded.code,
               let callout = L10n.Errors.ingestCallout(code: code, params: coded.params)
            {
                draft.reject = callout
                draft.filePath = nil
                draftLabelError = nil
            } else if let message = context.fail(error) {
                draftLabelError = message
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
            show(ingested.artifact, effects: ingested.effects)
            isAttaching = false
            attachDraft = AttachDraft()
            await context.refreshCoverFromStore()
            context.toast = VocabularyToast(
                title: L10n.Sources.toastFileAttachedTitle,
                body: L10n.Sources.toastFileAttachedBody(name: ingested.file.originalFilename),
                tone: .success
            )
        } catch {
            if let coded = error as? CoreInvokeError, let code = coded.code,
               let callout = L10n.Errors.ingestCallout(code: code, params: coded.params)
            {
                attachDraft.reject = callout
                attachDraft.filePath = nil
            } else if context.showFailure(error) {
                isAttaching = false
            }
        }
    }

    /// Opens the Artifact's file in its own app. A file with no path (Go returns none
    /// for a type Provenencia does not store) or one `ProjectFiles` refuses shows the
    /// shared refusal naming the file; a missing one keeps its own message.
    func open(_ art: CatalogArtifact) async {
        guard art.file != nil else { return }
        context.clearPageError()
        let name = art.fileDisplayName
        guard let rel = art.openableRelPath else {
            context.showPageError(L10n.ObjectOpen.refused(filename: name))
            return
        }
        guard let refusal = await ProjectFiles.openObject(projectDir: context.projectDir, relPath: rel) else { return }
        context.showPageError(L10n.ObjectOpen.message(
            for: refusal,
            filename: name,
            missing: L10n.string(L10n.Sources.fileOpenMissing)
        ))
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

    /// Puts a written Artifact on the page, appended when new: the one apply
    /// path for every Artifact write. The list is patched here (the page and the
    /// cached workspace; no slot patches the Artifact list). A write that carried
    /// the Artifact's copy details also puts its rows on the page, and the
    /// `.wroteMetadataBoard` slot writes them into the cache, with no reload.
    private func show(
        _ art: CatalogArtifact,
        metadata: [CatalogMetadataEntry]? = nil,
        effects: CatalogWriteEffects
    ) {
        guard var workspace = context.workspace else { return }
        if let idx = workspace.artifacts.firstIndex(where: { $0.id == art.id }) {
            workspace.artifacts[idx] = art
        } else {
            workspace.artifacts.append(art)
        }
        context.setWorkspace(workspace)
        let owner = MetadataOwner(level: .artifact, id: art.id)
        if let metadata {
            // The page only: the slot below writes the rows into the cached key once.
            context.workspace?.setMetadataRows(metadata, for: owner)
            context.session.apply(
                effects, .wroteMetadataBoard(sourceId: context.sourceID, owner: owner, entries: metadata)
            )
        } else {
            context.session.apply(effects)
        }
    }
}
