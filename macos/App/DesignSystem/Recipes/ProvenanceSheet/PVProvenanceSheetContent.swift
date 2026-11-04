import SwiftUI

/// How a value reads: names, text and terms in the body face, dates in mono,
/// places in italic. One type for the Vitals row, the provenance sheet and the
/// Why (S13-36 D-6).
enum PVProvenanceFace: Equatable {
    /// Names, text, terms, integers.
    case text
    /// Mono, like every date in the app.
    case date
    /// Italic, like every place.
    case place

    /// The face of a displayed value.
    var valueFont: Font {
        switch self {
        case .text: PVFont.body(size: PVTypeScale.body)
        case .date: PVFont.mono(size: 15)
        case .place: PVFont.body(size: PVTypeScale.body, italic: true)
        }
    }

    /// The face of a record's "Read as" cell: a phrase ("not …", "—") is italic.
    func readAsFont(isPhrase: Bool) -> Font {
        if isPhrase {
            return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
        switch self {
        case .text: return PVFont.body(size: PVTypeScale.bodySmall)
        case .date: return PVFont.mono(size: PVTypeScale.caption)
        case .place: return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
    }
}

/// What a provenance sheet says: a value, the records behind it, and how they
/// group. Pure; built from the wire records by the assembler below.
struct PVProvenanceSheetContent: Equatable {
    /// A displayed value as a line: its text and its Source count.
    struct Value: Equatable, Identifiable {
        var rank: Int
        var text: String
        var support: String
        var id: Int { rank }
    }

    /// A displayed value as a line: its text and its Source count.
    static func value(
        _ value: CatalogReconciledValue,
        propertyKey: String,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> Value {
        Value(
            rank: value.rank,
            text: ReconciledValueDisplay.string(for: value.value, propertyKey: propertyKey, locale: locale, orders: orders),
            support: L10n.Conclusions.sourceCount(value.support)
        )
    }
}
