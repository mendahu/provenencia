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
        return !label.isEmpty || !description.isEmpty || draft.filePath != nil
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
        }
    }

    // MARK: Card edit

    /// Opens edit on one card and restores any other card to its stored text.
    func beginEdit(_ id: String) {
        guard let art = items.first(where: { $0.id == id }) else { return }
        if let current = editingArtifactID, current != id {
            cancelFields(id: current)
        }
        editingArtifactID = id
        labels[id] = art.label
        descriptions[id] = art.description
        fieldErrors[id] = nil
    }

    func fieldsDirty(_ id: String) -> Bool {
        guard let art = items.first(where: { $0.id == id }) else { return false }
        let label = (labels[id] ?? art.label).trimmingCharacters(in: .whitespacesAndNewlines)
        let desc = (descriptions[id] ?? art.description).trimmingCharacters(in: .whitespacesAndNewlines)
        return label != art.label || desc != art.description
    }

    func canSaveFields(_ id: String) -> Bool {
        guard savingID == nil, fieldsDirty(id) else { return false }
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
        guard let art = items.first(where: { $0.id == id }) else {
            if editingArtifactID == id { editingArtifactID = nil }
            return
        }
        labels[id] = art.label
        descriptions[id] = art.description
        fieldErrors[id] = nil
        if editingArtifactID == id { editingArtifactID = nil }
    }

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
        if label == art.label, desc == art.description { return }
        context.clearPageError()
        savingID = id
        defer { savingID = nil }
        do {
            let (updated, effects) = try await context.store.updateArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id,
                label: label,
                description: desc
            )
            replace(updated, effects: effects)
            if editingArtifactID == id { editingArtifactID = nil }
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
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
        do {
            var (art, created) = try await context.store.createArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fileID: "",
                label: label,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines),
                graphIDs: []
            )
            // Each write's effects apply: the create linked the Artifact to the
            // Source's graphs, which the file ingest does not report again.
            var effects = [created]
            if let path = draft.filePath {
                let ingested = try await context.store.ingestArtifactFile(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: art.id,
                    path: path
                )
                art = ingested.artifact
                effects.append(ingested.effects)
            }
            context.workspace?.artifacts.append(art)
            labels[art.id] = art.label
            descriptions[art.id] = art.description
            isAdding = false
            draftLabelError = nil
            draft.reject = nil
            // The artifact list is appended above, so only the effects apply:
            // the Source's row moves and its graph reads follow.
            for write in effects {
                context.session.apply(write)
            }
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

    private func replace(_ updated: CatalogArtifact, effects: CatalogWriteEffects) {
        if let idx = context.workspace?.artifacts.firstIndex(where: { $0.id == updated.id }) {
            context.workspace?.artifacts[idx] = updated
        }
        labels[updated.id] = updated.label
        descriptions[updated.id] = updated.description
        context.applyWrite(effects)
    }
}
