import Foundation

/// One field row on a Conclusion detail page (board S9-D5
/// `ReconciledValueRow`): label · lead value · state · support · disclosures.
struct ReconciledValueRowModel: Equatable, Identifiable {
    enum ValueStyle: Equatable {
        /// Names, text, terms, integers.
        case text
        /// Mono, like every date in the app.
        case date
        /// Italic, like every place.
        case place
    }

    /// The handle whose page owns a derived value's Why.
    struct Opens: Equatable {
        var location: WorkspaceLocation
        /// Tooltip and spoken label of the open control.
        var label: String
    }

    struct OtherValue: Equatable, Identifiable {
        var rank: Int
        var text: String
        var support: String
        var id: Int { rank }
    }

    var id: String
    var label: String
    var style: ValueStyle
    /// The lead value; nil when the field is empty.
    var lead: String?
    var emptyText: String
    var badge: ReconciledValueDisplay.StateBadge?
    var count: String?
    var against: String?
    var otherValuesLabel: String?
    var otherValues: [OtherValue]
    /// Every displayed value when the field keeps several (`multiple`). Empty
    /// otherwise, including a mixed field, which still discloses the rest.
    var listedValues: [OtherValue]
    var whyTitle: String
    var records: [ReconciliationRecord]
    var accessibilityLabel: String
    /// A derived row opens the Event or Place that owns its Why. Nil
    /// otherwise.
    var opens: Opens?

    init(field: CatalogConclusionField, locale: Locale = .autoupdatingCurrent) {
        let lead = field.displayedValues.first
            .map { ReconciledValueDisplay.string(for: $0.value, locale: locale) }
            .flatMap { $0.isEmpty ? nil : $0 }
        let listed = Self.listedValues(of: field, locale: locale)
        let spokenLead = listed.count > 1 ? listed.map(\.text).joined(separator: "; ") : lead
        id = field.propertyID
        label = field.label
        style = field.valueType == .date ? .date : .text
        self.lead = lead
        emptyText = ReconciledValueDisplay.emptyText(propertyKey: field.propertyKey)
        badge = ReconciledValueDisplay.stateBadge(field)
        count = ReconciledValueDisplay.countLine(field)
        against = ReconciledValueDisplay.againstLine(field)
        otherValuesLabel = ReconciledValueDisplay.otherValuesLabel(field)
        otherValues = ReconciledValueDisplay.otherValues(field).map {
            OtherValue(
                rank: $0.rank,
                text: ReconciledValueDisplay.string(for: $0.value, locale: locale),
                support: L10n.Conclusions.sourceCount($0.support)
            )
        }
        listedValues = listed
        whyTitle = listed.count > 1
            ? L10n.Conclusions.whyThese(listed.count, label: field.label)
            : L10n.Conclusions.whyTitle(lead ?? field.label)
        records = field.outcomes.map { ReconciliationRecord(outcome: $0, in: field, locale: locale) }
        accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: field.label, lead: spokenLead, field: field)
    }

    /// Every kept value of a multi-valued field, lead included. A weak
    /// spelling is not displayed, so it stays out of this list.
    private static func listedValues(of field: CatalogConclusionField, locale: Locale) -> [OtherValue] {
        guard field.state == .multiple else { return [] }
        return field.displayedValues.compactMap { value in
            let text = ReconciledValueDisplay.string(for: value.value, locale: locale)
            guard !text.isEmpty else { return nil }
            return OtherValue(
                rank: value.rank,
                text: text,
                support: L10n.Conclusions.sourceCount(value.support)
            )
        }
    }

    /// One Date row whose lead is a start–end span. Why lists every record
    /// on those properties. The span is not one property's reconciled state,
    /// so it carries no badge.
    static func spanning(
        id: String,
        label: String,
        lead: String?,
        emptyText: String,
        fields: [CatalogConclusionField],
        locale: Locale
    ) -> Self {
        let shown = lead.flatMap { $0.isEmpty ? nil : $0 }
        let records = fields.flatMap { field in
            field.outcomes.map { ReconciliationRecord(outcome: $0, in: field, locale: locale) }
        }
        let spoken = shown.map { L10n.Conclusions.a11yList(label, rest: $0) }
            ?? ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        return Self(
            id: id, label: label, style: .date, lead: shown, emptyText: emptyText,
            badge: nil, count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            listedValues: [],
            whyTitle: shown.map { L10n.Conclusions.whyTitle($0) } ?? "",
            records: records,
            accessibilityLabel: spoken
        )
    }

    /// A date or place read off the canonical graph. A kept count above 1 is
    /// the mixed badge. The Why stays on the Event or Place that owns it,
    /// which `opens` navigates to.
    static func derived(
        id: String,
        label: String,
        style: ValueStyle,
        lead: String,
        emptyText: String,
        mixed: Bool,
        opens: Opens? = nil
    ) -> Self {
        let shown = lead.isEmpty ? nil : lead
        let spoken: String
        if let shown {
            let parts = [label, shown] + (mixed ? [L10n.string(L10n.Conclusions.badgeMixed)] : [])
            spoken = parts.dropFirst().reduce(parts[0]) { L10n.Conclusions.a11yList($0, rest: $1) }
        } else {
            spoken = ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        }
        var row = Self(
            id: id, label: label, style: style, lead: shown, emptyText: emptyText,
            badge: mixed && shown != nil ? .mixed : nil,
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [],
            whyTitle: "", records: [], accessibilityLabel: spoken
        )
        row.opens = opens
        return row
    }

    /// A row with nothing behind it yet.
    static func empty(id: String, label: String, style: ValueStyle, emptyText: String) -> Self {
        Self(
            id: id, label: label, style: style, lead: nil, emptyText: emptyText, badge: nil, count: nil,
            against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [], whyTitle: "", records: [],
            accessibilityLabel: ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        )
    }

    private init(
        id: String, label: String, style: ValueStyle, lead: String?, emptyText: String,
        badge: ReconciledValueDisplay.StateBadge?, count: String?, against: String?,
        otherValuesLabel: String?, otherValues: [OtherValue], listedValues: [OtherValue], whyTitle: String,
        records: [ReconciliationRecord], accessibilityLabel: String
    ) {
        self.id = id
        self.label = label
        self.style = style
        self.lead = lead
        self.emptyText = emptyText
        self.badge = badge
        self.count = count
        self.against = against
        self.otherValuesLabel = otherValuesLabel
        self.otherValues = otherValues
        self.listedValues = listedValues
        self.whyTitle = whyTitle
        self.records = records
        self.accessibilityLabel = accessibilityLabel
    }
}

extension ReconciledValueRowModel.Opens {
    /// The Event page a birth or death date is read from.
    static func event(_ entity: CatalogCanonicalEntity) -> Self {
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return Self(
            location: .eventDetail(entityId: entity.id, ref: entity.ref, title: label.isEmpty ? nil : label),
            label: L10n.string(L10n.Conclusions.openEvent)
        )
    }

    /// The Place page a place name is read from.
    static func place(_ place: CatalogHeaderPlace) -> Self {
        let title = PlaceTitleDisplay.titleSource(
            PlaceTitleParts(names: place.names, label: place.entity.label, ref: place.entity.ref)
        ).text
        return Self(
            location: .placeDetail(entityId: place.entity.id, ref: place.entity.ref, title: title),
            label: L10n.Conclusions.openPlace(title)
        )
    }
}
