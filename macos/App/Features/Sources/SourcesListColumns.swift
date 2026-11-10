import SwiftUI

/// The widths and insets the Sources list's caption band and its rows share:
/// Source (fills) · Type · (Published, at Wide) · Evidence graph. Below `compactBelow`
/// the Type column goes (the type and the date join a record line under the title)
/// and the graph column narrows.
struct SourcesListColumns: Equatable {
    let isCompact: Bool
    let widthClass: PVWidthClass

    /// Where a row draws its published date.
    enum DatePlacement: Equatable {
        case underType, column, recordLine
    }

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

    init(listWidth: CGFloat, widthClass: PVWidthClass) {
        isCompact = listWidth < Self.compactBelow
        self.widthClass = widthClass
    }

    var datePlacement: DatePlacement {
        if isCompact { return .recordLine }
        return widthClass == .wide ? .column : .underType
    }

    /// Nil when compact: there is no Type column.
    var type: CGFloat? {
        switch datePlacement {
        case .recordLine: nil
        case .column: 150
        case .underType: 170
        }
    }

    /// Nil unless the date has its own column (Wide).
    var published: CGFloat? { datePlacement == .column ? 150 : nil }
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

    func sourcesPublishedCell(_ columns: SourcesListColumns) -> some View {
        padding(.trailing, SourcesListColumns.cellTrailing)
            .frame(width: columns.published, alignment: .leading)
    }

    func sourcesGraphCell(_ columns: SourcesListColumns) -> some View {
        frame(width: columns.graphs)
            .padding(.leading, SourcesListColumns.graphLeading)
    }
}
