import SwiftUI

/// Kind chrome for Evidence graph primary cards (tinge wash, chip, ink).
struct EvidenceSubjectKindStyle: Sendable {
    var ink: Color
    var tint: Color
    var line: Color
    var chip: Color

    static func forKind(_ kind: EvidencePrimaryKind) -> EvidenceSubjectKindStyle {
        switch kind {
        case .person:
            return EvidenceSubjectKindStyle(
                ink: PVColor.subjectPersonInk,
                tint: PVColor.subjectPersonTint,
                line: PVColor.subjectPersonLine,
                chip: PVColor.subjectPersonChip
            )
        case .event:
            return EvidenceSubjectKindStyle(
                ink: PVColor.subjectEventInk,
                tint: PVColor.subjectEventTint,
                line: PVColor.subjectEventLine,
                chip: PVColor.subjectEventChip
            )
        case .place:
            return EvidenceSubjectKindStyle(
                ink: PVColor.subjectPlaceInk,
                tint: PVColor.subjectPlaceTint,
                line: PVColor.subjectPlaceLine,
                chip: PVColor.subjectPlaceChip
            )
        }
    }
}
