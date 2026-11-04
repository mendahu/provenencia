import SwiftUI

/// The fixed columns of a provenance record row and table: Read as · Citation ·
/// Outcome · Record · a reserved column. The Citation column takes the rest, so
/// a row placed alone and the rows of a table line up when they share one
/// `PVProvenanceColumns`.
struct PVProvenanceColumns: Equatable {
    /// Draw the Read as cell. Off, the row drops it and the record's one spoken
    /// line moves to the row's leading edge, so the record still reads as one element.
    var showsReadAs = true
    /// At `.wide` Read as and Outcome are wider and the Citation column takes the rest.
    var widthClass: PVWidthClass = .regular

    var readAsWidth: CGFloat { widthClass == .wide ? 220 : 140 }
    var outcomeWidth: CGFloat { widthClass == .wide ? 280 : 190 }
    /// The sheet panel's width around the columns.
    var panelWidth: CGFloat { widthClass == .wide ? 1280 : 880 }

    var spacing: CGFloat { Self.spacing }

    static let spacing: CGFloat = 14
    static let recordWidth: CGFloat = 112
    /// Reserved for a future "conclude this value" action, so adding it won't relayout the table (PD-8).
    static let actionWidth: CGFloat = 88
    /// The copy button's track beside the citation line, and the gap before it.
    static let copyTrackWidth: CGFloat = 26
    static let copyTrackSpacing: CGFloat = 6
}
