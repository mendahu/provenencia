import SwiftUI

/// Strip / chip chrome for Subject types — TYPES + VT_COLOR tables shared by
/// Subject Fields, Evidence graph, and the composer.
enum SubjectFieldsTypeChrome {
    static func sortIndex(_ key: String) -> Int {
        switch key {
        case "person": return 0
        case "event": return 1
        case "place": return 2
        case "relationship": return 3
        case "participation": return 4
        case "location": return 5
        case "source": return 6
        default: return 99
        }
    }

    /// One curated mark per Subject type key (including distinct bridge marks).
    static func stripMarkKey(typeKey: String) -> PVMarkKey? {
        switch typeKey {
        case "person": return .subjectPerson
        case "event": return .subjectEvent
        case "place": return .subjectPlace
        case "relationship": return .subjectRelationship
        case "participation": return .subjectParticipation
        case "location": return .subjectLocation
        case "source": return .subjectSource
        default: return nil
        }
    }

    /// Micro-label “BRIDGE” on Relationship / Participation / Location strip cards.
    /// Source is a reification, not a bridge — do not label it as one.
    static func showsBridgeLabel(typeKey: String) -> Bool {
        switch typeKey {
        case "relationship", "participation", "location": return true
        default: return false
        }
    }

    static func ink(typeKey: String, presentation: CatalogSubjectTypePresentation?) -> Color {
        // Bridges stay muted; source (reification) and roots follow presentation / kind ink.
        if showsBridgeLabel(typeKey: typeKey) { return PVColor.textMuted }
        switch presentation?.inkToken ?? "" {
        case "subjectPersonInk": return PVColor.subjectPersonInk
        case "subjectEventInk": return PVColor.subjectEventInk
        case "subjectPlaceInk": return PVColor.subjectPlaceInk
        default:
            switch typeKey {
            case "person": return PVColor.subjectPersonInk
            case "event": return PVColor.subjectEventInk
            case "place": return PVColor.subjectPlaceInk
            default: return PVColor.textPrimary
            }
        }
    }

    /// Board VT_COLOR — mono pill foreground.
    static func valueTypeForeground(_ valueType: String) -> Color {
        switch valueType {
        case "text": return PVPalette.paper600
        case "integer": return PVPalette.copper700
        case "date": return PVPalette.lapis700
        case "name": return PVPalette.iron700
        case "subject": return PVPalette.plum700
        case "term": return PVPalette.ochre700
        default: return PVColor.textSecondary
        }
    }
}
