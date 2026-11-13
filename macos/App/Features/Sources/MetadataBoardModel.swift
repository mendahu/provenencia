import Foundation
import Observation

/// The row a researcher fills to add a field to a board: the picked field, the typed value,
/// and the refusal under the value.
struct MetadataBoardAddRow: Equatable {
    var fieldID = ""
    var value = ""
    var error: String?
}

/// One owner's metadata rows (saved values + type suggestions), over the host
/// that holds the rows and runs the writes.
@MainActor
@Observable
final class MetadataBoardModel {
    /// Edit buffers keyed by field id.
    var drafts: [String: String] = [:]
    /// Saved row currently in explicit edit mode (one at a time).
    var editingFieldID: String?
    private(set) var savingFieldID: String?
    /// Inline value error after `sourcemetadata.invalid`, from a row's own save or the owner's write
    /// (`showRefusal`); cleared on next edit.
    var fieldError: String?
    var fieldErrorID: String?
    /// The add row; nil is closed. Kept across `resetDrafts`.
    var addRow: MetadataBoardAddRow?
    private(set) var isSavingAddRow = false

    /// Read by `MetadataBoardDialogs` for the clear confirm and the Add dialog over the add row.
    let host: any MetadataBoardHost

    init(host: any MetadataBoardHost) {
        self.host = host
    }

    var entries: [CatalogMetadataEntry] { host.entries }

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
        return host.fields
            .filter { !used.contains($0.id) }
            .map { PVComboBoxOption(value: $0.id, label: $0.label, subtext: $0.key) }
    }

    /// Seeds edit buffers for saved values and drops buffers for rows that no
    /// longer exist (dismissed suggestions). Also exits any open edit mode. An open add row stays;
    /// a field it picked that the rows now hold is cleared from it, and its value stays.
    func resetDrafts() {
        editingFieldID = nil
        fieldError = nil
        fieldErrorID = nil
        syncDrafts()
        if let picked = addRow?.fieldID, entries.contains(where: { $0.field.id == picked }) {
            addRow?.fieldID = ""
        }
    }

    /// Resets only when the rows the workspace holds differ from the rows the
    /// board showed, so a reload or patch that carries the board's own rows
    /// keeps an edit begun since.
    func resetIfChanged(from old: [CatalogMetadataEntry]?, to new: [CatalogMetadataEntry]?) {
        if old != new { resetDrafts() }
    }

    private func syncDrafts() {
        for e in entries where drafts[e.field.id] == nil || e.hasValue {
            drafts[e.field.id] = e.valueText
        }
        let ids = Set(entries.map(\.field.id))
        drafts = drafts.filter { ids.contains($0.key) }
    }

    /// The row holding a saved value for `fieldID`; nil for an empty suggestion or no row.
    func savedEntry(fieldID: String) -> CatalogMetadataEntry? {
        entries.first { $0.field.id == fieldID && $0.hasValue }
    }

    // MARK: Row edit

    func beginEdit(fieldID: String) {
        guard let entry = savedEntry(fieldID: fieldID) else { return }
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
        host.clearFailure()
        fieldError = nil
        fieldErrorID = nil
        do {
            let (updated, effects) = try await host.set(fieldID: fieldID, valueText: value)
            applySet(updated, fieldID: fieldID, effects: effects)
            if editingFieldID == fieldID {
                editingFieldID = nil
            }
        } catch {
            if case .invalid(let message) = classifySetFailure(error) {
                placeRowError(message, fieldID: fieldID)
            }
        }
    }

    /// Asks a clear of a value: with no value for `fieldID`, false; on a host that confirms, true and
    /// nothing is written (the caller opens its confirm); otherwise clears at once and answers false.
    func requestClear(fieldID: String) async -> Bool {
        guard savedEntry(fieldID: fieldID) != nil else { return false }
        if host.confirmsClear { return true }
        _ = await clear(fieldID: fieldID)
        return false
    }

    /// Clears a saved value. True when it was cleared.
    func clear(fieldID: String) async -> Bool {
        host.clearFailure()
        do {
            let (updated, effects) = try await host.clear(fieldID: fieldID)
            host.entries = updated
            drafts[fieldID] = ""
            if editingFieldID == fieldID {
                editingFieldID = nil
            }
            host.didWrite(effects)
            syncDrafts()
            return true
        } catch {
            report(error)
            return false
        }
    }

    // MARK: Add row

    /// True when the open add row has a field, a value with text, and no write in flight.
    var canConfirmAddRow: Bool {
        guard let row = addRow, !isSavingAddRow else { return false }
        return !row.fieldID.isEmpty
            && !row.value.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    /// Opens an empty add row; an open one keeps its field, value and error.
    func openAddRow() {
        if addRow == nil { addRow = MetadataBoardAddRow() }
    }

    func closeAddRow() {
        guard !isSavingAddRow else { return }
        addRow = nil
    }

    /// Adds the row's field with its trimmed value. A saved value closes the row; a refusal stays
    /// on it; any other failure shows where the host shows failures and leaves the row as it is.
    func confirmAddRow() async {
        guard canConfirmAddRow, let row = addRow else { return }
        addRow?.error = nil
        isSavingAddRow = true
        defer { isSavingAddRow = false }
        let value = row.value.trimmingCharacters(in: .whitespacesAndNewlines)
        switch await add(fieldID: row.fieldID, valueText: value) {
        case .saved: addRow = nil
        case .invalid(let message): addRow?.error = message
        case .failed: break
        }
    }

    /// Saves a value for a field with no saved value yet (the add row's write).
    func add(fieldID: String, valueText: String) async -> MetadataBoardAddResult {
        host.clearFailure()
        do {
            let (updated, effects) = try await host.set(fieldID: fieldID, valueText: valueText)
            applySet(updated, fieldID: fieldID, effects: effects)
            return .saved
        } catch {
            if case .invalid(let message) = classifySetFailure(error) { return .invalid(message) }
            return .failed
        }
    }

    private enum SetFailure {
        case ownerGone
        case invalid(String)
        case shown
    }

    /// A failed set: `sourcemetadata.invalid` stays on the row or dialog; I/O goes
    /// to the host's failure place; nothing shows when the owner is gone.
    private func classifySetFailure(_ error: Error) -> SetFailure {
        guard let message = host.failureMessage(for: error) else { return .ownerGone }
        if isRowRefusal(error) { return .invalid(message) }
        host.showFailure(message)
        return .shown
    }

    /// The message for a `sourcemetadata.invalid` refusal, or nil for any other error.
    private func rowRefusalMessage(_ error: Error) -> String? {
        guard isRowRefusal(error) else { return nil }
        return host.failureMessage(for: error)
    }

    private func isRowRefusal(_ error: Error) -> Bool {
        (error as? CoreInvokeError)?.engineCode == .sourceMetadataInvalid
    }

    private func placeRowError(_ message: String, fieldID: String) {
        fieldError = message
        fieldErrorID = fieldID
    }

    /// Lands an owner's refused write on the row it names: the refused value's field id is
    /// `params[0]`. True when the row took it (its error placed); false when the error is no row
    /// refusal, names no row with a value, or another row holds typed text (that row stays in
    /// edit and the owner shows its general failure); nothing changes then. Typed text is never
    /// discarded: the row's own open edit keeps its buffer; otherwise the refused row opens for
    /// edit, where `PVInlineEdit` draws its error.
    func showRefusal(_ error: Error) -> Bool {
        guard let fieldID = (error as? CoreInvokeError)?.params.first,
              savedEntry(fieldID: fieldID) != nil,
              let message = rowRefusalMessage(error)
        else { return false }
        if editingFieldID != fieldID {
            if hasTypedEdit { return false }
            beginEdit(fieldID: fieldID)
        }
        placeRowError(message, fieldID: fieldID)
        return true
    }

    /// True when the row in edit holds text that is not its saved value.
    private var hasTypedEdit: Bool {
        guard let id = editingFieldID, let saved = savedEntry(fieldID: id) else { return false }
        return drafts[id] != saved.valueText
    }

    func dismissSuggestion(fieldID: String) async {
        host.clearFailure()
        do {
            let (updated, effects) = try await host.dismissSuggestion(fieldID: fieldID)
            host.entries = updated
            host.didWrite(effects)
            syncDrafts()
        } catch {
            report(error)
        }
    }

    /// Reorders saved fields only; suggestions stay below in their current
    /// order. On failure the optimistic move is reverted.
    func moveSaved(from source: IndexSet, to destination: Int) async {
        let previous = entries
        var savedRows = saved
        savedRows.move(fromOffsets: source, toOffset: destination)
        let ordered = savedRows + suggested
        host.entries = ordered
        host.clearFailure()
        do {
            let (updated, effects) = try await host.reorder(fieldIDs: ordered.map(\.field.id))
            host.entries = updated
            host.didWrite(effects)
            syncDrafts()
        } catch {
            report(error)
            host.entries = previous
        }
    }

    /// Takes the owner's rows from a `set` answer: the rows, the field's draft, then the effects.
    private func applySet(_ updated: [CatalogMetadataEntry], fieldID: String, effects: CatalogWriteEffects) {
        host.entries = updated
        drafts[fieldID] = updated.first { $0.field.id == fieldID }?.valueText ?? ""
        host.didWrite(effects)
    }

    /// Shows a failed write's message; shows nothing when the owner is gone.
    private func report(_ error: Error) {
        if let message = host.failureMessage(for: error) { host.showFailure(message) }
    }
}
