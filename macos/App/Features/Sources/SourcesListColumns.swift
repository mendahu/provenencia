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

    init(listWidth: CGFloat) {
        isCompact = listWidth < Self.compactBelow
    }

    /// Nil when compact: there is no Type column.
    var type: CGFloat? { isCompact ? nil : 150 }
    var status: CGFloat { isCompact ? 120 : 132 }
    var graphs: CGFloat { isCompact ? 240 : 300 }
}
