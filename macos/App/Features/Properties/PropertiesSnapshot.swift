import Foundation
import SwiftUI

/// Session-cached Properties destination payload.
struct PropertiesSnapshot: Sendable, Equatable {
    var properties: [CatalogProperty]
    var types: [CatalogSubjectType]
    var propertiesByTypeID: [String: [CatalogSubjectTypeProperty]]
    var presentationsByKey: [String: CatalogSubjectTypePresentation]

    static let empty = PropertiesSnapshot(
        properties: [],
        types: [],
        propertiesByTypeID: [:],
        presentationsByKey: [:]
    )

    /// Board TYPES order: Person → Event → Place → Relationship → Participation → Location → Source.
    var typesInPaletteOrder: [CatalogSubjectType] {
        types.sorted {
            PropertiesTypeChrome.sortIndex($0.key) < PropertiesTypeChrome.sortIndex($1.key)
        }
    }

    func presentation(for type: CatalogSubjectType) -> CatalogSubjectTypePresentation? {
        presentationsByKey[type.key]
    }

    func boundTypes(for propertyID: String) -> [CatalogSubjectType] {
        typesInPaletteOrder.filter { type in
            propertiesByTypeID[type.id]?.contains { $0.property.id == propertyID } == true
        }
    }

    func binding(propertyID: String, typeID: String) -> CatalogSubjectTypeProperty? {
        propertiesByTypeID[typeID]?.first { $0.property.id == propertyID }
    }

    func propertyCount(forTypeID typeID: String) -> Int {
        propertiesByTypeID[typeID]?.count ?? 0
    }
}

// MARK: - Board bind control (16×16 filled box — not a native checkbox)

/// Matches the board `box(on, locked)` control used in On column + inspector binds.
struct PropertiesBindBox: View {
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

struct PropertiesValueTypePill: View {
    let valueType: String

    var body: some View {
        // Board shows the raw value-type key in mono + VT_COLOR, not a title-cased label.
        Text(verbatim: valueType)
            .font(PVFont.mono(size: PVTypeScale.micro))
            .foregroundStyle(PropertiesTypeChrome.valueTypeForeground(valueType))
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
struct PropertiesOriginCell: View {
    let origin: String

    var body: some View {
        if origin == CatalogOrigin.user {
            PVBadge(L10n.Properties.originUserShort, tone: .warning, subtle: true)
        } else if origin == CatalogOrigin.provenencia {
            HStack(spacing: 5) {
                Circle()
                    .fill(PVColor.accentLine)
                    .frame(width: 5, height: 5)
                Text(L10n.Properties.originSeededShort)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        } else {
            OriginPill(origin: origin)
        }
    }
}

struct PropertiesBoundChip: View {
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
