import SwiftUI

/// Which row's provenance sheet is open: the row's id, not a snapshot, so the
/// sheet follows its row across a page refresh (S13-11 AC-4.1, AC-4.2).
struct ProvenanceSheetItem: Identifiable, Equatable {
    let id: String

    /// The row with this id, or nil when it is gone or has no records.
    func row(in rows: [ReconciledValueRowModel]) -> ReconciledValueRowModel? {
        rows.first { $0.id == id && $0.recordCount > 0 }
    }
}
