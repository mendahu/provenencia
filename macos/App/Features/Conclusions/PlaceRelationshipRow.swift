import SwiftUI

/// One related Place on a Place detail: title, optional role, span, opens
/// the related page (S9-D7 / S9-40).
struct PlaceRelationshipRow: View {
    let relationship: CatalogPlaceRelationship
    /// Optional role line under the title (Succeeded / Succeeded by).
    var role: String? = nil
    @Environment(WorkspaceNavigation.self) private var navigation

    var body: some View {
        Button {
            navigation.go(to: .placeDetail(
                entityId: relationship.entity.id,
                ref: relationship.entity.ref,
                title: relationship.title.isEmpty ? nil : relationship.title
            ))
        } label: {
            HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                VStack(alignment: .leading, spacing: PVSpacing.space1) {
                    Text(verbatim: displayTitle)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .multilineTextAlignment(.leading)
                    if let role, !role.isEmpty {
                        Text(verbatim: role)
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textSecondary)
                    }
                    if !spanText.isEmpty {
                        Text(verbatim: spanText)
                            .font(PVFont.mono(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                    }
                }
                Spacer(minLength: 0)
                PVIcon(.chevronForward, size: 12)
                    .foregroundStyle(PVColor.textMuted)
            }
            .padding(.vertical, PVSpacing.space4)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: accessibilityLabel))
        .accessibilityIdentifier("places.detail.relationship.\(relationship.id)")
    }

    private var displayTitle: String {
        let t = relationship.title.trimmingCharacters(in: .whitespacesAndNewlines)
        return t.isEmpty ? relationship.entity.ref : t
    }

    private var spanText: String {
        PlacePeriodDisplay.span(start: relationship.startDate, end: relationship.endDate)
    }

    private var accessibilityLabel: String {
        var parts = [displayTitle]
        if let role, !role.isEmpty { parts.append(role) }
        if !spanText.isEmpty { parts.append(spanText) }
        return parts.joined(separator: ", ")
    }
}

/// Formats a Place period or membership span for detail rows.
enum PlacePeriodDisplay {
    static func span(
        start: CatalogDateValueInput?,
        end: CatalogDateValueInput?,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        let a = start.map { DateValueDisplay.string(for: $0, locale: locale) } ?? ""
        let b = end.map { DateValueDisplay.string(for: $0, locale: locale) } ?? ""
        switch (a.isEmpty, b.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return L10n.Conclusions.placePeriodOpen(start: a)
        case (true, false):
            return L10n.Conclusions.placePeriodUntil(end: b)
        case (false, false):
            return L10n.Conclusions.lifeSpan(born: a, died: b)
        }
    }
}
