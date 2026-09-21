import SwiftUI

/// Kind chrome for Evidence graph primary cards (tinge wash, chip, ink).
/// Prefer registry presentation tokens when available (S7-09).
struct EvidenceSubjectKindStyle: Sendable {
    var ink: Color
    var tint: Color
    var line: Color
    var chip: Color

    static func forKind(_ kind: EvidencePrimaryKind) -> EvidenceSubjectKindStyle {
        resolve(typeKey: kind.rawValue, presentation: nil)
    }

    static func resolve(
        typeKey: String,
        presentation: CatalogSubjectTypePresentation?
    ) -> EvidenceSubjectKindStyle {
        EvidenceSubjectKindStyle(
            ink: SubjectFieldsTypeChrome.ink(typeKey: typeKey, presentation: presentation),
            tint: color(token: presentation?.tintToken, fallbackTypeKey: typeKey, role: .tint),
            line: color(token: presentation?.lineToken, fallbackTypeKey: typeKey, role: .line),
            chip: color(token: presentation?.chipToken, fallbackTypeKey: typeKey, role: .chip)
        )
    }

    private enum TokenRole {
        case tint, line, chip
    }

    private static func color(
        token: String?,
        fallbackTypeKey: String,
        role: TokenRole
    ) -> Color {
        switch token ?? "" {
        case "subjectPersonTint": return PVColor.subjectPersonTint
        case "subjectEventTint": return PVColor.subjectEventTint
        case "subjectPlaceTint": return PVColor.subjectPlaceTint
        case "subjectPersonLine": return PVColor.subjectPersonLine
        case "subjectEventLine": return PVColor.subjectEventLine
        case "subjectPlaceLine": return PVColor.subjectPlaceLine
        case "subjectPersonChip": return PVColor.subjectPersonChip
        case "subjectEventChip": return PVColor.subjectEventChip
        case "subjectPlaceChip": return PVColor.subjectPlaceChip
        default:
            switch (fallbackTypeKey, role) {
            case ("event", .tint), ("participation", .tint): return PVColor.subjectEventTint
            case ("place", .tint), ("location", .tint): return PVColor.subjectPlaceTint
            case (_, .tint): return PVColor.subjectPersonTint
            case ("event", .line), ("participation", .line): return PVColor.subjectEventLine
            case ("place", .line), ("location", .line): return PVColor.subjectPlaceLine
            case (_, .line): return PVColor.subjectPersonLine
            case ("event", .chip), ("participation", .chip): return PVColor.subjectEventChip
            case ("place", .chip), ("location", .chip): return PVColor.subjectPlaceChip
            case (_, .chip): return PVColor.subjectPersonChip
            }
        }
    }
}
