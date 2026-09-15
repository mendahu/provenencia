import Foundation
import Observation

/// Research notes on the Source page: composer draft plus explicit
/// edit/save/cancel for existing rows (same pattern as description).
@MainActor
@Observable
final class SourceNotesSection {
    /// New-note composer text.
    var draft = ""
    private(set) var isSaving = false

    /// Note currently in explicit edit mode (one at a time).
    private(set) var editingNoteID: String?
    var bodyDraft = ""
    var bodyError: String?

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var items: [CatalogSourceNote] { context.workspace?.notes ?? [] }

    /// Clears any open edit session after a workspace (re)load.
    func reset() {
        editingNoteID = nil
        bodyDraft = ""
        bodyError = nil
    }

    func add() async {
        let body = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !body.isEmpty, !isSaving else { return }
        isSaving = true
        defer { isSaving = false }
        context.clearPageError()
        do {
            let note = try await context.store.addSourceNote(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                body: body
            )
            draft = ""
            context.workspace?.notes.append(note)
            context.notifyWorkspaceMutated()
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }

    // MARK: Edit session

    func beginEdit(id: String) {
        guard let note = items.first(where: { $0.id == id }) else { return }
        if let current = editingNoteID, current != id {
            cancelEdit()
        }
        editingNoteID = id
        bodyDraft = note.body
        bodyError = nil
    }

    func cancelEdit() {
        guard !isSaving else { return }
        editingNoteID = nil
        bodyDraft = ""
        bodyError = nil
    }

    func saveEdit() async {
        guard let id = editingNoteID else { return }
        guard !isSaving else { return }
        let trimmed = bodyDraft.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            bodyError = String(localized: L10n.Sources.noteBodyRequired)
            return
        }
        guard let note = items.first(where: { $0.id == id }) else { return }
        if trimmed == note.body {
            editingNoteID = nil
            bodyDraft = ""
            bodyError = nil
            return
        }
        bodyError = nil
        isSaving = true
        defer { isSaving = false }
        do {
            let updated = try await context.store.updateSourceNote(
                projectDir: context.projectDir,
                userID: context.userID,
                noteID: id,
                body: trimmed
            )
            if let idx = context.workspace?.notes.firstIndex(where: { $0.id == id }) {
                context.workspace?.notes[idx] = updated
            }
            editingNoteID = nil
            bodyDraft = ""
            context.notifyWorkspaceMutated()
        } catch {
            bodyError = L10n.Errors.message(for: error)
        }
    }

    func delete(id: String) async {
        guard !isSaving else { return }
        context.clearPageError()
        do {
            try await context.store.deleteSourceNote(
                projectDir: context.projectDir,
                userID: context.userID,
                noteID: id
            )
            if editingNoteID == id {
                editingNoteID = nil
                bodyDraft = ""
                bodyError = nil
            }
            context.workspace?.notes.removeAll { $0.id == id }
            context.notifyWorkspaceMutated()
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }
}
