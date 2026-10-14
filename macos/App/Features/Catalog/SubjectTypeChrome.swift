import SwiftUI

/// Strip / chip chrome for Subject types — TYPES + VT_COLOR tables shared by
/// Properties, Evidence graph, and the composer.
enum PropertiesTypeChrome {
    /// Micro-label “BRIDGE” on Relationship / Participation / Location strip cards.
    /// Source is a reification, not a bridge — do not label it as one.
    static func showsBridgeLabel(presentation: CatalogSubjectTypePresentation?) -> Bool {
        presentation?.role == "bridge"
    }

    static func ink(typeKey: String, presentation: CatalogSubjectTypePresentation?) -> Color {
        // Bridges stay muted; source (reification) and roots follow presentation / kind ink.
        if showsBridgeLabel(presentation: presentation) { return PVColor.textMuted }
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
