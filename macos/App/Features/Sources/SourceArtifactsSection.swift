import Foundation
import Observation

/// One Artifact row's graph line: none for an Artifact in no graph (the badge and the
/// link), else the graphs it is in (`shown` the first two names, `more` the rest).
enum ArtifactGraphLine: Equatable {
    case none
    case graphs(all: [String], shown: [String], more: Int)
}

/// Artifacts on the Source page: the Artifact sheet (edit and add), the Add file
/// dialog, and file open.
@MainActor
@Observable
final class SourceArtifactsSection {
    struct AttachDraft: Equatable {
        var artifactID = ""
        var artifactRef = ""
        var file = IngestFileChoice()
    }

    // MARK: Artifact sheet

    /// True while `openAdd` reads the type's suggestions, so a second click or drop opens nothing.
    private(set) var isOpeningAdd = false

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

    /// The Source's Evidence graphs when the session has them (they feed
    /// `graphLines`); nil otherwise, so no line shows.
    var visibleGraphs: [CatalogEvidenceGraph]? { context.evidenceGraphs }

    var canSubmitAttach: Bool {
        !isSavingAttach
            && attachDraft.file.filePath != nil
            && attachDraft.file.reject == nil
    }

    /// The artifact the file will attach to is a seed. A chosen file is an edit.
    var attachSheetIsDirty: Bool {
        attachDraft.file.filePath != nil
    }

    /// What an Artifact row's graph line shows, by Artifact id. Empty until the session has the
    /// Source's graphs and has settled which Artifacts are in none: then an Artifact in no
    /// graph is `.none` and every other has its graph names, in the Graphs section's order.
    /// An id with no entry draws no graph line.
    var graphLines: [String: ArtifactGraphLine] {
        guard let noGraph = context.session.artifactIDsInNoGraph(
            sourceID: context.sourceID, artifactIDs: items.map(\.id)
        ), let graphs = visibleGraphs else { return [:] }
        var names: [String: [String]] = [:]
        for graph in graphs {
            for id in graph.artifactIDs {
                names[id, default: []].append(graph.name)
            }
        }
        var lines: [String: ArtifactGraphLine] = [:]
        for item in items {
            if noGraph.contains(item.id) {
                lines[item.id] = ArtifactGraphLine.none
            } else {
                let all = names[item.id] ?? []
                let shown = Array(all.prefix(2))
                lines[item.id] = .graphs(all: all, shown: shown, more: all.count - shown.count)
            }
        }
        return lines
    }

    /// The type mark an Artifact with no file shows in its thumbnail: the Source type's.
    func thumbnailTypeIconKey(for art: CatalogArtifact) -> String? {
        art.thumbnailTypeIconKey(sourceTypeIconKey: context.sourceTypeIconKey)
    }

    /// The Artifact's saved copy details, read from the workspace the page already holds.
    func copyDetails(for artifactID: String) -> [CatalogMetadataEntry] {
        context.workspace?.metadataRows(for: MetadataOwner(level: .artifact, id: artifactID)) ?? []
    }

    /// The page's workspace changed: a sheet whose Artifact is no longer on the
    /// page closes with its draft and says so. An Artifact still there keeps its draft.
    func workspaceChanged() {
        guard let sheet, let id = sheet.artifactID, !items.contains(where: { $0.id == id }) else { return }
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
            onSaved: sheetSaved()
        )
    }

    /// Opens the Artifact sheet in add mode, its copy details board starting as the Source type's
    /// `artifact` suggestions, empty. Answers the sheet, or nil when none opened: one is open or
    /// opening, or the suggestion read failed (the page banner shows the message).
    @discardableResult
    func openAdd() async -> ArtifactSheetModel? {
        guard sheet == nil, !isOpeningAdd else { return nil }
        isOpeningAdd = true
        defer { isOpeningAdd = false }
        let suggestions: [CatalogMetadataEntry]
        do {
            suggestions = try await context.unsavedMetadataRows(level: .artifact)
        } catch {
            context.showFailure(error)
            return nil
        }
        // A sheet opened while the read ran (an edit from a pencil) is not replaced.
        guard sheet == nil else { return nil }
        let created = ArtifactSheetModel(
            context: context,
            mode: .create(suggestions: suggestions),
            opensAt: .top,
            onSaved: sheetSaved()
        )
        sheet = created
        return created
    }

    /// Opens the Artifact sheet in add mode with the dropped file staged. Nothing is stored until Save.
    func stagePageDrop(_ urls: [URL]) async {
        guard !urls.isEmpty, sheet == nil, !isOpeningAdd else { return }
        guard let sheet = await openAdd() else { return }
        sheet.file.choose(urls)
    }

    /// What every Artifact write the sheet makes does with the answer: puts it on the page.
    private func sheetSaved() -> @MainActor (CatalogArtifact, [CatalogMetadataEntry]?, CatalogWriteEffects) -> Void {
        { [weak self] art, metadata, effects in
            self?.show(art, metadata: metadata, effects: effects)
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

    func confirmAttach() async {
        guard canSubmitAttach else { return }
        guard let path = attachDraft.file.filePath else { return }
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
            // A refused file keeps the dialog open with the refusal under the row.
            if !attachDraft.file.refuse(ifIngestRefusal: error), context.showFailure(error) {
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
