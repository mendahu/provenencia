import Foundation
import Observation

/// What a board's Add reports to the add row that asked: written, refused by the
/// engine with a message for the value field, or failed with the host showing it.
enum MetadataBoardAddResult {
    case saved
    case invalid(String)
    case failed
}

/// Snapshot for the metadata-value delete confirm (`pvConfirm(item:)`), with
/// the board the clear runs on.
struct MetadataBoardDeleteItem: Identifiable, Equatable {
    var id: String { fieldID }
    var fieldID: String
    var label: String
    let board: MetadataBoardModel

    static func == (lhs: Self, rhs: Self) -> Bool {
        lhs.fieldID == rhs.fieldID && lhs.label == rhs.label && ObjectIdentifier(lhs.board) == ObjectIdentifier(rhs.board)
    }
}

/// A page's one Add-metadata dialog and one clear confirm. Each targets the
/// board that asked.
@MainActor
@Observable
final class MetadataBoardDialogs {
    /// The board the open Add dialog writes to.
    private(set) var addTarget: MetadataBoardModel?
    var isAdding: Bool { addTarget != nil }
    var addFieldError: String?
    /// Confirm target for clearing a saved value.
    var pendingDelete: MetadataBoardDeleteItem?
    private(set) var isClearing = false

    var fieldComboOptions: [PVComboBoxOption] { addTarget?.fieldComboOptions ?? [] }

    /// The Add dialog presents its target's add row; these read and write that row.
    var addFieldID: String {
        get { addTarget?.addRow?.fieldID ?? "" }
        set { addTarget?.addRow?.fieldID = newValue }
    }

    var addValue: String {
        get { addTarget?.addRow?.value ?? "" }
        set { addTarget?.addRow?.value = newValue }
    }

    var addValueError: String? {
        get { addTarget?.addRow?.error }
        set { addTarget?.addRow?.error = newValue }
    }

    var isSavingAdd: Bool { addTarget?.isSavingAddRow ?? false }

    var addSheetIsDirty: Bool {
        !addFieldID.isEmpty
            || !addValue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    var canSubmitAdd: Bool { addTarget?.canConfirmAddRow ?? false }

    func openAdd(on board: MetadataBoardModel) {
        if let other = addTarget, other !== board { other.closeAddRow() }
        board.closeAddRow()
        board.openAddRow()
        addTarget = board
        addFieldError = nil
    }

    func cancelAdd() {
        guard !isSavingAdd else { return }
        addTarget?.closeAddRow()
        addTarget = nil
    }

    func createFromAdd() async {
        guard let target = addTarget else { return }
        addFieldError = nil
        addValueError = nil
        if addFieldID.isEmpty {
            addFieldError = L10n.string(L10n.Sources.metadataFieldRequired)
            return
        }
        if addValue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            addValueError = L10n.string(L10n.Sources.metadataValueRequired)
            return
        }
        await target.confirmAddRow()
        if target.addRow == nil { addTarget = nil }
    }

    func askClear(on board: MetadataBoardModel, fieldID: String) {
        guard let entry = board.savedEntry(fieldID: fieldID) else { return }
        pendingDelete = MetadataBoardDeleteItem(fieldID: fieldID, label: entry.field.label, board: board)
    }

    func confirmClear() async {
        guard let item = pendingDelete else { return }
        guard !isClearing else { return }
        isClearing = true
        defer { isClearing = false }
        if await item.board.clear(fieldID: item.fieldID) {
            pendingDelete = nil
        }
    }
}
