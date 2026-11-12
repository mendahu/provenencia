import Foundation
import Observation

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
    /// Inline value error after `sourcemetadata.invalid` (cleared on next edit).
    var fieldError: String?
    var fieldErrorID: String?

    /// Read by `MetadataBoardDialogs` for the Add dialog and clear confirm that target this board.
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
            // `sourcemetadata.invalid` stays on the row; I/O goes to the host's failure place.
            guard let message = host.failureMessage(for: error) else { return }
            if (error as? CoreInvokeError)?.engineCode == .sourceMetadataInvalid {
                fieldError = message
                fieldErrorID = fieldID
            } else {
                host.showFailure(message)
            }
        }
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

    /// Saves a value for a field with no saved value yet (the Add dialog's write).
    func add(fieldID: String, valueText: String) async -> MetadataBoardAddResult {
        host.clearFailure()
        do {
            let (updated, effects) = try await host.set(fieldID: fieldID, valueText: valueText)
            applySet(updated, fieldID: fieldID, effects: effects)
            return .saved
        } catch {
            guard let message = host.failureMessage(for: error) else { return .failed }
            if (error as? CoreInvokeError)?.engineCode == .sourceMetadataInvalid {
                return .invalid(message)
            }
            host.showFailure(message)
            return .failed
        }
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
