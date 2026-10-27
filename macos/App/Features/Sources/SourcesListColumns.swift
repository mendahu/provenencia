import SwiftUI

/// The widths and insets the Sources list's caption band and its rows share:
/// Source (fills) · Type · Status · Evidence graph. Below `compactBelow` the Type
/// column goes (the type joins the Source's second line) and Status and the graph
/// column narrow.
struct SourcesListColumns: Equatable {
    let isCompact: Bool

    static let compactBelow: CGFloat = 880
    static let sourceTrailing = PVSpacing.space6
    static let cellTrailing = PVSpacing.space5
    static let graphLeading = PVSpacing.space6
    /// The Source cell's top and bottom padding around the cover (14 pt).
    static let sourceVertical = PVSpacing.space5 + PVSpacing.space1
    /// The status and graph stacks share these, so each badge stays level with
    /// its graph button (with `SourcesGraphButton.height`).
    static let slotSpacing = PVSpacing.space3
    static let stackVertical = PVSpacing.space4

    init(listWidth: CGFloat) {
        isCompact = listWidth < Self.compactBelow
    }

    /// Nil when compact: there is no Type column.
    var type: CGFloat? { isCompact ? nil : 150 }
    var status: CGFloat { isCompact ? 120 : 132 }
    var graphs: CGFloat { isCompact ? 240 : 300 }
}

/// One frame chain per column, so the caption band and the rows cannot drift.
extension View {
    func sourcesSourceCell() -> some View {
        padding(.trailing, SourcesListColumns.sourceTrailing)
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    func sourcesTypeCell(_ columns: SourcesListColumns) -> some View {
        padding(.trailing, SourcesListColumns.cellTrailing)
            .frame(width: columns.type, alignment: .leading)
    }

    func sourcesStatusCell(_ columns: SourcesListColumns) -> some View {
        padding(.trailing, SourcesListColumns.cellTrailing)
            .frame(width: columns.status, alignment: .leading)
    }

    func sourcesGraphCell(_ columns: SourcesListColumns) -> some View {
        frame(width: columns.graphs)
            .padding(.leading, SourcesListColumns.graphLeading)
    }
}
