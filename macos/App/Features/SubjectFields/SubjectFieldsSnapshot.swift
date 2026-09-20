import Foundation
import SwiftUI

/// Session-cached Subject fields destination payload (properties + types + bindings + presentation).
struct SubjectFieldsSnapshot: Sendable, Equatable {
    var properties: [CatalogProperty]
    var types: [CatalogSubjectType]
    /// Bindings keyed by subject type id.
    var fieldsByTypeID: [String: [CatalogSubjectTypeField]]
    /// Presentation keyed by subject type key (person, event, …).
    var presentationsByKey: [String: CatalogSubjectTypePresentation]

    static let empty = SubjectFieldsSnapshot(
        properties: [],
        types: [],
        fieldsByTypeID: [:],
        presentationsByKey: [:]
    )

    /// Types in registry palette order (All strip uses this after the All card).
    var typesInPaletteOrder: [CatalogSubjectType] {
        types.sorted { lhs, rhs in
            let l = presentationsByKey[lhs.key]?.paletteSort ?? SubjectFieldsTypeChrome.fallbackSort(lhs.key)
            let r = presentationsByKey[rhs.key]?.paletteSort ?? SubjectFieldsTypeChrome.fallbackSort(rhs.key)
            if l != r { return l < r }
            return lhs.label.localizedCaseInsensitiveCompare(rhs.label) == .orderedAscending
        }
    }

    func presentation(for type: CatalogSubjectType) -> CatalogSubjectTypePresentation? {
        presentationsByKey[type.key]
    }

    func boundTypes(for propertyID: String) -> [CatalogSubjectType] {
        typesInPaletteOrder.filter { type in
            fieldsByTypeID[type.id]?.contains { $0.property.id == propertyID } == true
        }
    }

    func binding(propertyID: String, typeID: String) -> CatalogSubjectTypeField? {
        fieldsByTypeID[typeID]?.first { $0.property.id == propertyID }
    }

    func propertyCount(forTypeID typeID: String) -> Int {
        fieldsByTypeID[typeID]?.count ?? 0
    }
}

/// Maps registry presentation tokens / type keys onto S6 canvas chrome.
enum SubjectFieldsTypeChrome {
    static func fallbackSort(_ key: String) -> Int {
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

    static func iconKind(symbol: String?, typeKey: String) -> PVSubjectIconKind? {
        let raw = symbol ?? typeKey
        return PVSubjectIconKind(rawValue: raw)
    }

    static func isBridge(role: String?) -> Bool {
        role == "bridge"
    }

    /// Primaries carry pigment; bridges and unknown roles stay ink-neutral.
    static func ink(for presentation: CatalogSubjectTypePresentation?) -> Color {
        guard let presentation, !isBridge(role: presentation.role) else {
            return PVColor.textPrimary
        }
        switch presentation.inkToken {
        case "subjectPersonInk": return PVColor.subjectPersonInk
        case "subjectEventInk": return PVColor.subjectEventInk
        case "subjectPlaceInk": return PVColor.subjectPlaceInk
        default:
            switch presentation.typeKey {
            case "person", "source": return PVColor.subjectPersonInk
            case "event": return PVColor.subjectEventInk
            case "place": return PVColor.subjectPlaceInk
            default: return PVColor.textPrimary
            }
        }
    }

    static func tint(for presentation: CatalogSubjectTypePresentation?) -> Color {
        guard let presentation, !isBridge(role: presentation.role) else {
            return PVColor.surfaceRaised
        }
        switch presentation.tintToken {
        case "subjectPersonTint": return PVColor.subjectPersonTint
        case "subjectEventTint": return PVColor.subjectEventTint
        case "subjectPlaceTint": return PVColor.subjectPlaceTint
        default: return PVColor.surfaceRaised
        }
    }
}
