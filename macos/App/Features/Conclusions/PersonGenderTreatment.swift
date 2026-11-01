import SwiftUI

/// How the Person summary shows gender (S10-23, board S10-D12): a Badge above
/// the name and a tint on the thumb. Gender wins; sex at birth stands in only
/// when the summary shows no gender value, and says so. Pure: the view
/// only lays it out.
struct PersonGenderTreatment: Equatable {
    enum Family: Equatable {
        case man, woman, nonBinary, other

        var tint: Color {
            switch self {
            case .man: PVColor.accentSoft
            case .woman: PVColor.personWomanTint
            case .nonBinary: PVColor.infoSoft
            case .other: PVColor.surfaceSunken
            }
        }

        var line: Color {
            switch self {
            case .man: PVColor.accentLine
            case .woman: PVColor.lineMaternal
            case .nonBinary: PVColor.info
            case .other: PVColor.borderInked
            }
        }

        var ink: Color {
            switch self {
            case .man: PVColor.accentSoftForeground
            case .woman: PVColor.personWomanInk
            case .nonBinary: PVColor.infoForeground
            case .other: PVColor.textPrimary
            }
        }
    }

    enum Basis: Equatable {
        case gender, sexAtBirth
    }

    var family: Family
    var basis: Basis
    /// The value as the Details row words it.
    var text: String
    /// The Property's own label ("Gender", "Sex at birth").
    var fieldLabel: String

    var badge: ConclusionDetailBadge {
        ConclusionDetailBadge(
            text: basis == .gender ? text : L10n.Conclusions.sexAtBirthBadge(text),
            ink: family.ink,
            accessibilityLabel: L10n.Conclusions.a11yList(fieldLabel, rest: text)
        )
    }

    var thumbnailColors: PVThumbnail.Colors {
        PVThumbnail.Colors(tint: family.tint, line: family.line)
    }

    /// Reads the Person summary's terms (S13-02). Nil when there is nothing to
    /// show: no summary, no value, a disagreement, or a value that is not a
    /// readable term.
    static func read(_ summary: CatalogPersonSummary?) -> PersonGenderTreatment? {
        guard let summary else { return nil }
        if summary.gender.count > 0 { return treatment(summary.gender, basis: .gender) }
        return treatment(summary.sexAtBirth, basis: .sexAtBirth)
    }

    private static func treatment(_ facts: CatalogTermFacts, basis: Basis) -> PersonGenderTreatment? {
        guard facts.count == 1, let value = facts.term, case .term(_, let key, _, let origin) = value else { return nil }
        let text = ReconciledValueDisplay.string(for: value)
        guard !text.isEmpty else { return nil }
        return PersonGenderTreatment(family: family(CatalogOrigin.isProduct(origin) ? key : "", basis: basis), basis: basis, text: text, fieldLabel: facts.propertyLabel)
    }

    private static func family(_ key: String, basis: Basis) -> Family {
        switch (basis, key) {
        case (.gender, SeededGenderTerm.man), (.sexAtBirth, SeededSexAtBirthTerm.male): .man
        case (.gender, SeededGenderTerm.woman), (.sexAtBirth, SeededSexAtBirthTerm.female): .woman
        case (.gender, SeededGenderTerm.nonBinary): .nonBinary
        default: .other
        }
    }
}
