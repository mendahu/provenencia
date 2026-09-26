import Foundation
import Observation

/// Snapshot for the metadata-value delete confirm (`pvConfirm(item:)`).
struct SourceMetadataDeleteItem: Identifiable, Equatable {
    var id: String { fieldID }
    var fieldID: String
    var label: String
}

/// Metadata rows (saved values + type suggestions) and the Add-metadata dialog.
@MainActor
@Observable
final class SourceMetadataSection {
    /// Edit buffers keyed by field id.
    var drafts: [String: String] = [:]
    /// Saved row currently in explicit edit mode (one at a time).
    var editingFieldID: String?
    private(set) var savingFieldID: String?
    /// Inline value error after `sourcemetadata.invalid` (cleared on next edit).
    var fieldError: String?
    var fieldErrorID: String?
    /// Confirm target for clearing a saved value.
    var pendingDelete: SourceMetadataDeleteItem?

    // MARK: Add dialog

    var isAdding = false
    var addFieldID = ""
    var addValue = ""
    var addFieldError: String?
    var addValueError: String?
    private(set) var isSavingAdd = false
    private(set) var isClearing = false

    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    var entries: [CatalogMetadataEntry] { context.workspace?.metadata ?? [] }

    /// Saved values only — drag-reorderable in the Metadata section.
    var saved: [CatalogMetadataEntry] {
        entries.filter(\.hasValue)
    }

    /// Type suggestions without a value — separate from the reorderable list.
    var suggested: [CatalogMetadataEntry] {
        entries.filter { !$0.hasValue && $0.suggested }
    }

    var fieldComboOptions: [PVComboBoxOption] {
        let used = Set(entries.map(\.field.id))
        return context.fields
            .filter { !used.contains($0.id) }
            .map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }

    var canSubmitAdd: Bool {
        !isSavingAdd
            && !addFieldID.isEmpty
            && !addValue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    /// Seeds edit buffers for saved values and drops buffers for rows that no
    /// longer exist (dismissed suggestions). Also exits any open edit mode.
    func resetDrafts() {
        editingFieldID = nil
        fieldError = nil
        fieldErrorID = nil
        syncDrafts()
    }

    private func syncDrafts() {
        for e in entries where drafts[e.field.id] == nil || e.hasValue {
            drafts[e.field.id] = e.valueText
        }
        let ids = Set(entries.map(\.field.id))
        drafts = drafts.filter { ids.contains($0.key) }
    }

    // MARK: Row edit

    func beginEdit(fieldID: String) {
        guard let entry = entries.first(where: { $0.field.id == fieldID }), entry.hasValue else { return }
        if let current = editingFieldID, current != fieldID {
            cancelEdit()
        }
        drafts[fieldID] = entry.valueText
        editingFieldID = fieldID
        if fieldErrorID == fieldID {
            fieldError = nil
            fieldErrorID = nil
        }
    }

    func cancelEdit() {
        guard let fieldID = editingFieldID else { return }
        guard savingFieldID == nil else { return }
        if let entry = entries.first(where: { $0.field.id == fieldID }) {
            drafts[fieldID] = entry.valueText
        }
        editingFieldID = nil
        if fieldErrorID == fieldID {
            fieldError = nil
            fieldErrorID = nil
        }
    }

    func save(fieldID: String) async {
        guard let entry = entries.first(where: { $0.field.id == fieldID }) else { return }
        guard savingFieldID == nil else { return }
        let value = (drafts[fieldID] ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else { return }
        if entry.hasValue, value == entry.valueText {
            if editingFieldID == fieldID {
                editingFieldID = nil
            }
            return
        }
        savingFieldID = fieldID
        defer { savingFieldID = nil }
        context.clearPageError()
        fieldError = nil
        fieldErrorID = nil
        do {
            let updated = try await context.store.setSourceMetadata(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fieldID: fieldID,
                valueText: value
            )
            replaceEntry(updated)
            if editingFieldID == fieldID {
                editingFieldID = nil
            }
        } catch {
            applySetError(error, fieldID: fieldID)
        }
    }

    func askClear(fieldID: String) {
        guard let entry = entries.first(where: { $0.field.id == fieldID }), entry.hasValue else { return }
        pendingDelete = SourceMetadataDeleteItem(fieldID: fieldID, label: entry.field.label)
    }

    func confirmClear() async {
        guard let item = pendingDelete else { return }
        guard !isClearing else { return }
        isClearing = true
        defer { isClearing = false }
        context.clearPageError()
        do {
            try await context.store.clearSourceMetadata(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fieldID: item.fieldID
            )
            applyCleared(fieldID: item.fieldID)
            pendingDelete = nil
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }

    func dismissSuggestion(fieldID: String) async {
        context.clearPageError()
        do {
            let updated = try await context.store.dismissSourceMetadataSuggestion(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fieldID: fieldID
            )
            context.workspace?.metadata = updated
            context.notifyWorkspaceMutated()
            syncDrafts()
        } catch {
            context.pageError = L10n.Errors.message(for: error)
        }
    }

    /// Reorders saved fields only; suggestions stay below in their current
    /// order. On failure the optimistic move is reverted.
    func moveSaved(from source: IndexSet, to destination: Int) async {
        let previous = entries
        var savedRows = saved
        savedRows.move(fromOffsets: source, toOffset: destination)
        let ordered = savedRows + suggested
        context.workspace?.metadata = ordered
        context.clearPageError()
        do {
            let updated = try await context.store.reorderSourceMetadata(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fieldIDs: ordered.map(\.field.id)
            )
            context.workspace?.metadata = updated
            context.notifyWorkspaceMutated()
            syncDrafts()
        } catch {
            context.pageError = L10n.Errors.message(for: error)
            context.workspace?.metadata = previous
        }
    }

    // MARK: Add dialog

    func openAdd() {
        addFieldID = ""
        addValue = ""
        addFieldError = nil
        addValueError = nil
        isAdding = true
    }

    func cancelAdd() {
        guard !isSavingAdd else { return }
        isAdding = false
    }

    func createFromAdd() async {
        addFieldError = nil
        addValueError = nil
        if addFieldID.isEmpty {
            addFieldError = String(localized: L10n.Sources.metadataFieldRequired)
            return
        }
        let value = addValue.trimmingCharacters(in: .whitespacesAndNewlines)
        if value.isEmpty {
            addValueError = String(localized: L10n.Sources.metadataValueRequired)
            return
        }
        isSavingAdd = true
        defer { isSavingAdd = false }
        context.clearPageError()
        do {
            let entry = try await context.store.setSourceMetadata(
                projectDir: context.projectDir,
                userID: context.userID,
                sourceID: context.sourceID,
                fieldID: addFieldID,
                valueText: value
            )
            replaceEntry(entry)
            isAdding = false
        } catch {
            applySetError(error, fieldID: nil)
        }
    }

    /// `sourcemetadata.invalid` stays on the value field; I/O goes to `pageError`.
    func applySetError(_ error: Error, fieldID: String?) {
        if case .coded(_, let code, _, _) = error as? CoreInvokeError,
           code == "sourcemetadata.invalid"
        {
            let message = L10n.Errors.message(for: error)
            if isAdding {
                addValueError = message
            } else if let fieldID {
                fieldError = message
                fieldErrorID = fieldID
            }
            return
        }
        context.pageError = L10n.Errors.message(for: error)
    }

    /// Patches one workspace row from a `setSourceMetadata` response.
    private func replaceEntry(_ entry: CatalogMetadataEntry) {
        if let idx = entries.firstIndex(where: { $0.field.id == entry.field.id }) {
            context.workspace?.metadata[idx] = entry
        } else {
            context.workspace?.metadata.append(entry)
        }
        drafts[entry.field.id] = entry.valueText
        context.notifyMetadataMutated()
    }

    private func applyCleared(fieldID: String) {
        guard var list = context.workspace?.metadata,
              let idx = list.firstIndex(where: { $0.field.id == fieldID })
        else { return }
        if list[idx].suggested {
            list[idx].valueText = ""
            list[idx].hasValue = false
        } else {
            list.remove(at: idx)
        }
        context.workspace?.metadata = list
        drafts[fieldID] = ""
        if editingFieldID == fieldID {
            editingFieldID = nil
        }
        context.notifyMetadataMutated()
        syncDrafts()
    }
}
