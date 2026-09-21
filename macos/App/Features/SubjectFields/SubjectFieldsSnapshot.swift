import Foundation
import SwiftUI

/// Session-cached Subject fields destination payload.
struct SubjectFieldsSnapshot: Sendable, Equatable {
    var properties: [CatalogProperty]
    var types: [CatalogSubjectType]
    var fieldsByTypeID: [String: [CatalogSubjectTypeField]]
    var presentationsByKey: [String: CatalogSubjectTypePresentation]

    static let empty = SubjectFieldsSnapshot(
        properties: [],
        types: [],
        fieldsByTypeID: [:],
        presentationsByKey: [:]
    )

    /// Board TYPES order: Person → Event → Place → Relationship → Participation → Location → Source.
    var typesInPaletteOrder: [CatalogSubjectType] {
        types.sorted {
            SubjectFieldsTypeChrome.sortIndex($0.key) < SubjectFieldsTypeChrome.sortIndex($1.key)
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

/// S7-D2 strip / chip chrome — matches the board's TYPES + VT_COLOR tables.
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

// MARK: - Board bind control (16×16 filled box — not a native checkbox)

/// Matches the board `box(on, locked)` control used in On column + inspector binds.
struct SubjectFieldsBindBox: View {
    var on: Bool
    var locked: Bool

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                .fill(fill)
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                        .strokeBorder(stroke, lineWidth: 1)
                )
            if on {
                Image(systemName: locked ? "lock.fill" : "checkmark")
                    .font(.system(size: 9, weight: .bold))
                    .foregroundStyle(PVColor.accentForeground)
            }
        }
        .frame(width: 16, height: 16)
        .accessibilityHidden(true)
    }

    private var fill: Color {
        guard on else { return PVColor.surfaceCard }
        return locked ? PVPalette.paper500 : PVColor.accent
    }

    private var stroke: Color {
        on ? .clear : PVColor.borderDefault
    }
}

struct SubjectFieldsValueTypePill: View {
    let valueType: String

    var body: some View {
        // Board shows the raw value-type key in mono + VT_COLOR, not a title-cased label.
        Text(verbatim: valueType)
            .font(PVFont.mono(size: PVTypeScale.micro))
            .foregroundStyle(SubjectFieldsTypeChrome.valueTypeForeground(valueType))
            .padding(.horizontal, 6)
            .frame(height: 18)
            .background(PVColor.surfaceSunken)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous)
                    .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.xs, style: .continuous))
            .accessibilityLabel(Text(SubjectPropertyValueType.label(valueType)))
    }
}

/// Board origin column: warning badge for user; accent dot + “seeded” for provenencia.
struct SubjectFieldsOriginCell: View {
    let origin: String

    var body: some View {
        if origin == CatalogOrigin.user {
            PVBadge(L10n.SubjectFields.originUserShort, tone: .warning, subtle: true)
        } else if origin == CatalogOrigin.provenencia {
            HStack(spacing: 5) {
                Circle()
                    .fill(PVColor.accentLine)
                    .frame(width: 5, height: 5)
                Text(L10n.SubjectFields.originSeededShort)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        } else {
            OriginPill(origin: origin)
        }
    }
}

struct SubjectFieldsBoundChip: View {
    let label: String
    let locked: Bool
    var emphasized: Bool = false

    var body: some View {
        HStack(spacing: 3) {
            if locked {
                Image(systemName: "lock.fill")
                    .font(.system(size: 9, weight: .semibold))
            }
            Text(verbatim: label)
                .font(PVFont.body(size: PVTypeScale.micro))
                .lineLimit(1)
        }
        .foregroundStyle(emphasized ? PVColor.accentSoftForeground : PVColor.textSecondary)
        .padding(.horizontal, 6)
        .frame(height: 18)
        .background(emphasized ? PVColor.accentSoft : Color.clear)
        .overlay(
            Capsule(style: .continuous)
                .strokeBorder(emphasized ? PVColor.accentLine : PVColor.borderSubtle, lineWidth: 1)
        )
        .clipShape(Capsule(style: .continuous))
    }
}
