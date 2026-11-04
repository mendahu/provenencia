import Foundation

/// One field row on a Conclusion detail page (board S9-D5
/// `ReconciledValueRow`): label · lead value · state · support · disclosures.
struct ReconciledValueRowModel: Equatable, Identifiable {
    /// The handle a derived value is read from.
    struct Opens: Equatable {
        var location: WorkspaceLocation
        /// Tooltip and spoken label of the open control.
        var label: String
    }

    var id: String
    var label: String
    var face: PVProvenanceFace
    /// The lead value; nil when the field is empty.
    var lead: String?
    var emptyText: String
    var badge: ReconciledValueDisplay.StateBadge?
    var count: String?
    var against: String?
    var otherValuesLabel: String?
    var otherValues: [PVProvenanceSheetContent.Value]
    /// Every displayed value when the field keeps several (`multiple`). Empty
    /// otherwise, including a mixed field, which still discloses the rest.
    var listedValues: [PVProvenanceSheetContent.Value]
    var whyTitle: String
    /// The Why's records, grouped by the entity that owns them.
    var recordGroups: [PVProvenanceSheetContent.Group]
    var records: [PVProvenanceSheetContent.Record] { recordGroups.flatMap(\.records) }
    var accessibilityLabel: String
    /// The provenance sheet's summary line: what the row knows about the records
    /// behind it. Nil when there is nothing to say.
    var provenanceSummary: String?
    /// A derived row opens the Event or Place it is read from. Nil otherwise.
    var opens: Opens?

    init(
        field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let lead = field.displayedValues.first
            .map { ReconciledValueDisplay.string(for: $0.value, propertyKey: field.propertyKey, locale: locale, orders: orders) }
            .flatMap { $0.isEmpty ? nil : $0 }
        let listed = Self.listedValues(of: field, locale: locale, orders: orders)
        let spokenLead = listed.count > 1 ? listed.map(\.text).joined(separator: "; ") : lead
        id = field.propertyID
        label = field.label
        face = field.valueType == .date ? .date : .text
        self.lead = lead
        emptyText = ReconciledValueDisplay.emptyText(propertyKey: field.propertyKey)
        badge = ReconciledValueDisplay.stateBadge(field)
        count = ReconciledValueDisplay.countLine(field)
        against = ReconciledValueDisplay.againstLine(field)
        otherValuesLabel = ReconciledValueDisplay.otherValuesLabel(field)
        otherValues = ReconciledValueDisplay.otherValues(field).map { PVProvenanceSheetContent.value($0, propertyKey: field.propertyKey, locale: locale, orders: orders) }
        listedValues = listed
        whyTitle = listed.count > 1
            ? L10n.Conclusions.whyThese(listed.count, label: field.label)
            : field.valueType == .name
                ? L10n.Conclusions.whyTitleName(lead ?? field.label)
                : L10n.Conclusions.whyTitle(lead ?? field.label)
        let content = PVProvenanceSheetContent(field: field, title: field.label, locale: locale, orders: orders)
        recordGroups = content.groups
        accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: field.label, lead: spokenLead, field: field)
        provenanceSummary = content.summary
    }

    /// Every kept value of a multi-valued field, lead included. A weak
    /// spelling is not displayed, so it stays out of this list.
    private static func listedValues(
        of field: CatalogConclusionField,
        locale: Locale,
        orders: NameFormatOrders
    ) -> [PVProvenanceSheetContent.Value] {
        guard field.state == .multiple else { return [] }
        return field.displayedValues.compactMap { value in
            let listed = PVProvenanceSheetContent.value(value, propertyKey: field.propertyKey, locale: locale, orders: orders)
            return listed.text.isEmpty ? nil : listed
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
        let content = PVProvenanceSheetContent.spanning(fields: fields, title: label, locale: locale)
        let spoken = shown.map { L10n.DesignSystem.spokenJoin(label, $0) }
            ?? ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        return Self(
            id: id, label: label, face: .date, lead: shown, emptyText: emptyText,
            badge: nil, count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            listedValues: [],
            whyTitle: shown.map { L10n.Conclusions.whyTitle($0) } ?? "",
            recordGroups: content.groups,
            accessibilityLabel: spoken,
            provenanceSummary: content.summary
        )
    }

    /// A date or place read off the canonical graph. A kept count above 1 is
    /// the mixed badge. Its Why lists the owning handles' fields: the shown
    /// owner's first, each under a group line when there are several or when
    /// the group is not the shown owner's, and only the shown owner's records
    /// in primary ink.
    static func derived(
        id: String,
        label: String,
        face: PVProvenanceFace,
        lead: String,
        emptyText: String,
        mixed: Bool,
        opens: Opens? = nil,
        owners: [CatalogOwnerField] = [],
        shownOwnerID: String? = nil,
        kind: CatalogDerivedRowKind? = nil,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> Self {
        let shown = lead.isEmpty ? nil : lead
        let spoken: String
        if let shown {
            let parts = [label, shown] + (mixed ? [L10n.string(L10n.Conclusions.badgeMixed)] : [])
            spoken = L10n.DesignSystem.spokenJoin(parts) ?? label
        } else {
            spoken = ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        }
        let content = PVProvenanceSheetContent.derived(
            owners: owners, shownOwnerID: shownOwnerID, kind: kind, title: label, locale: locale, orders: orders
        )
        var row = Self(
            id: id, label: label, face: face, lead: shown, emptyText: emptyText,
            badge: mixed && shown != nil ? .mixed : nil,
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [],
            whyTitle: "", recordGroups: content.groups, accessibilityLabel: spoken, provenanceSummary: content.summary
        )
        if !row.records.isEmpty { row.whyTitle = L10n.Conclusions.whyTitle(shown ?? label) }
        row.opens = opens
        return row
    }

    /// A row with nothing behind it yet.
    static func empty(id: String, label: String, face: PVProvenanceFace, emptyText: String) -> Self {
        Self(
            id: id, label: label, face: face, lead: nil, emptyText: emptyText, badge: nil, count: nil,
            against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [], whyTitle: "", recordGroups: [],
            accessibilityLabel: ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil),
            provenanceSummary: nil
        )
    }

    private init(
        id: String, label: String, face: PVProvenanceFace, lead: String?, emptyText: String,
        badge: ReconciledValueDisplay.StateBadge?, count: String?, against: String?,
        otherValuesLabel: String?, otherValues: [PVProvenanceSheetContent.Value], listedValues: [PVProvenanceSheetContent.Value], whyTitle: String,
        recordGroups: [PVProvenanceSheetContent.Group], accessibilityLabel: String, provenanceSummary: String?
    ) {
        self.id = id
        self.label = label
        self.face = face
        self.lead = lead
        self.emptyText = emptyText
        self.badge = badge
        self.count = count
        self.against = against
        self.otherValuesLabel = otherValuesLabel
        self.otherValues = otherValues
        self.listedValues = listedValues
        self.whyTitle = whyTitle
        self.recordGroups = recordGroups
        self.accessibilityLabel = accessibilityLabel
        self.provenanceSummary = provenanceSummary
    }
}

extension ReconciledValueRowModel.Opens {
    /// The Event a birth or death date is read from.
    static func event(_ entity: CatalogCanonicalEntity) -> Self {
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return Self(
            location: .eventDetail(entityId: entity.id, ref: entity.ref, title: label.isEmpty ? nil : label),
            label: L10n.string(L10n.Conclusions.openEvent)
        )
    }

    /// The Place a place name is read from.
    static func place(_ place: CatalogChainedPlace) -> Self {
        let title = PlaceTitleDisplay.titleSource(place.titleParts).text
        return Self(
            location: .placeDetail(entityId: place.entity.id, ref: place.entity.ref, title: title),
            label: L10n.Conclusions.openPlace(title)
        )
    }
}
