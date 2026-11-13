import Foundation
import Observation

/// What a board's Add reports to the dialog that asked: written, refused by the
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
    var addFieldID = ""
    var addValue = ""
    var addFieldError: String?
    var addValueError: String?
    private(set) var isSavingAdd = false
    /// Confirm target for clearing a saved value.
    var pendingDelete: MetadataBoardDeleteItem?
    private(set) var isClearing = false

    var fieldComboOptions: [PVComboBoxOption] { addTarget?.fieldComboOptions ?? [] }

    var addSheetIsDirty: Bool {
        !addFieldID.isEmpty
            || !addValue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    var canSubmitAdd: Bool {
        !isSavingAdd
            && !addFieldID.isEmpty
            && !addValue.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    func openAdd(on board: MetadataBoardModel) {
        addFieldID = ""
        addValue = ""
        addFieldError = nil
        addValueError = nil
        addTarget = board
    }

    func cancelAdd() {
        guard !isSavingAdd else { return }
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
        let value = addValue.trimmingCharacters(in: .whitespacesAndNewlines)
        if value.isEmpty {
            addValueError = L10n.string(L10n.Sources.metadataValueRequired)
            return
        }
        isSavingAdd = true
        defer { isSavingAdd = false }
        switch await target.add(fieldID: addFieldID, valueText: value) {
        case .saved: addTarget = nil
        case .invalid(let message): addValueError = message
        case .failed: break
        }
    }

    func askClear(on board: MetadataBoardModel, fieldID: String) {
        guard let entry = board.entries.first(where: { $0.field.id == fieldID }), entry.hasValue else { return }
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
