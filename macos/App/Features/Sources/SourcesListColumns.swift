import SwiftUI

/// The widths and insets the Sources list's caption band and its rows share:
/// Source (fills) · Type · Evidence graph. Below `compactBelow` the Type column
/// goes (the type joins the Source's second line) and the graph column narrows.
struct SourcesListColumns: Equatable {
    let isCompact: Bool

    static let compactBelow: CGFloat = 880
    static let sourceTrailing = PVSpacing.space6
    static let cellTrailing = PVSpacing.space5
    static let graphLeading = PVSpacing.space6
    /// The Source cell's top and bottom padding around the cover (14 pt).
    static let sourceVertical = PVSpacing.space5 + PVSpacing.space1
    /// The graph stack's spacing and vertical padding (with
    /// `SourcesGraphButton.height`).
    static let slotSpacing = PVSpacing.space3
    static let stackVertical = PVSpacing.space4

    init(listWidth: CGFloat) {
        isCompact = listWidth < Self.compactBelow
    }

    /// Nil when compact: there is no Type column.
    var type: CGFloat? { isCompact ? nil : 170 }
    var graphs: CGFloat { isCompact ? 380 : 420 }
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

    func sourcesGraphCell(_ columns: SourcesListColumns) -> some View {
        frame(width: columns.graphs)
            .padding(.leading, SourcesListColumns.graphLeading)
    }
}
