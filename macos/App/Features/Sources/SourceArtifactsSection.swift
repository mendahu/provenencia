import Foundation
import Observation

/// Artifacts accordion on the Source page: expand/collapse, per-artifact
/// label/description edits, the Add-artifact dialog, and file attach/open.
@MainActor
@Observable
final class SourceArtifactsSection {
    struct Draft: Equatable {
        var label = ""
        var description = ""
        var filePath: String?
        var fileName: String?
    }

    var expandedIDs: Set<String> = []
    /// Local edit buffers for expanded artifact identity fields, keyed by id.
    var labels: [String: String] = [:]
    var descriptions: [String: String] = [:]
    /// Per-artifact validation for expanded edit fields.
    var fieldErrors: [String: String] = [:]
    /// Artifact id currently saving label/description from the accordion.
    private(set) var savingID: String?

    // MARK: Add dialog

    var isAdding = false
    var draft = Draft()
    var draftLabelError: String?
    private(set) var isSavingDraft = false

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var items: [CatalogArtifact] { context.workspace?.artifacts ?? [] }

    var canSubmitDraft: Bool {
        !isSavingDraft
            && !draft.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
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
    }

    // MARK: Accordion

    func toggleExpanded(_ id: String) {
        if expandedIDs.contains(id) {
            expandedIDs.removeAll()
        } else {
            // Accordion: only one Artifact expanded at a time.
            expandedIDs = [id]
            if let art = items.first(where: { $0.id == id }) {
                labels[id] = art.label
                descriptions[id] = art.description
                fieldErrors[id] = nil
            }
        }
    }

    // MARK: Expanded fields

    func fieldsDirty(_ id: String) -> Bool {
        guard let art = items.first(where: { $0.id == id }) else { return false }
        let label = (labels[id] ?? art.label).trimmingCharacters(in: .whitespacesAndNewlines)
        let desc = (descriptions[id] ?? art.description).trimmingCharacters(in: .whitespacesAndNewlines)
        return label != art.label || desc != art.description
    }

    func canSaveFields(_ id: String) -> Bool {
        fieldsDirty(id) && savingID == nil
    }

    func cancelFields(id: String) {
        guard savingID != id else { return }
        guard let art = items.first(where: { $0.id == id }) else { return }
        labels[id] = art.label
        descriptions[id] = art.description
        fieldErrors[id] = nil
    }

    func saveFields(id: String) async {
        guard let art = items.first(where: { $0.id == id }) else { return }
        guard savingID == nil else { return }
        let label = (labels[id] ?? art.label).trimmingCharacters(in: .whitespacesAndNewlines)
        let desc = (descriptions[id] ?? art.description).trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            fieldErrors[id] = String(localized: L10n.Sources.artifactLabelRequired)
            return
        }
        fieldErrors[id] = nil
        if label == art.label, desc == art.description { return }
        context.clearPageError()
        savingID = id
        defer { savingID = nil }
        do {
            let updated = try await context.store.updateArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id,
                label: label,
                description: desc
            )
            replace(updated)
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
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
            prompt: String(localized: L10n.Sources.filePickPrompt),
            message: String(localized: L10n.Sources.filePickMessage)
        )
        guard let path else { return }
        draft.filePath = path
        draft.fileName = URL(fileURLWithPath: path).lastPathComponent
    }

    func clearFile() {
        draft.filePath = nil
        draft.fileName = nil
    }

    func create() async {
        guard !isSavingDraft else { return }
        let label = draft.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if label.isEmpty {
            draftLabelError = String(localized: L10n.Sources.artifactLabelRequired)
            return
        }
        isSavingDraft = true
        defer { isSavingDraft = false }
        do {
            var art = try await context.store.createArtifact(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fileID: "",
                label: label,
                description: draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
            )
            if let path = draft.filePath {
                let ingested = try await context.store.ingestArtifactFile(
                    projectDir: context.projectDir,
                    userID: context.userID,
                    artifactID: art.id,
                    path: path
                )
                art = ingested.artifact
            }
            context.workspace?.artifacts.append(art)
            labels[art.id] = art.label
            descriptions[art.id] = art.description
            isAdding = false
            draftLabelError = nil
            context.publishCoverToList()
            context.toast = VocabularyToast(
                title: L10n.Sources.toastArtifactCreatedTitle(ref: art.ref),
                body: art.fileID.isEmpty
                    ? L10n.Sources.toastArtifactCreatedFileless
                    : L10n.Sources.toastArtifactCreatedWithFile,
                tone: .success
            )
        } catch {
            draftLabelError = L10n.Errors.message(for: error)
        }
    }

    // MARK: Files

    func addFile(to id: String) async {
        guard let art = items.first(where: { $0.id == id }), art.fileID.isEmpty else { return }
        let path = ProjectFiles.pickFileForIngest(
            prompt: String(localized: L10n.Sources.filePickPrompt),
            message: String(localized: L10n.Sources.filePickMessage)
        )
        guard let path else { return }
        context.clearPageError()
        do {
            let ingested = try await context.store.ingestArtifactFile(
                projectDir: context.projectDir,
                userID: context.userID,
                artifactID: id,
                path: path
            )
            replace(ingested.artifact)
            context.publishCoverToList()
            context.toast = VocabularyToast(
                title: L10n.Sources.toastFileAttachedTitle,
                body: L10n.Sources.toastFileAttachedBody(name: ingested.file.originalFilename),
                tone: .success
            )
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }

    func open(_ art: CatalogArtifact) {
        guard let file = art.file, !file.relPath.isEmpty else { return }
        context.clearPageError()
        if !ProjectFiles.openObject(projectDir: context.projectDir, relPath: file.relPath) {
            context.pageError = String(localized: L10n.Sources.fileOpenMissing)
        }
    }

    private func replace(_ updated: CatalogArtifact) {
        if let idx = context.workspace?.artifacts.firstIndex(where: { $0.id == updated.id }) {
            context.workspace?.artifacts[idx] = updated
        }
        labels[updated.id] = updated.label
        descriptions[updated.id] = updated.description
    }
}
