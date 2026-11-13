import Foundation
import Observation

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

/// A page's one clear confirm. It targets the board that asked. Adding a field
/// is the board's own inline row (`MetadataBoardView`), not a dialog.
@MainActor
@Observable
final class MetadataBoardDialogs {
    /// Confirm target for clearing a saved value.
    var pendingDelete: MetadataBoardDeleteItem?
    private(set) var isClearing = false

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
