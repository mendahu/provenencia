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

    /// What the row's records are built from, not the records: a sheet or a
    /// Why builds them when it opens (`provenanceContent()`).
    private enum RecordsSource: Equatable {
        case field(CatalogConclusionField)
        case spanning([CatalogConclusionField])
        case derived(owners: [CatalogOwnerField], shownOwnerID: String?, kind: CatalogDerivedRowKind?)
    }

    var id: String
    var label: String
    /// What the row's value line shows: the lead, every kept value of a
    /// multi-valued field, the state and the empty text.
    var shown: PVProvenanceSheetContent.Shown
    var count: String?
    var against: String?
    var otherValuesLabel: String?
    var otherValues: [PVProvenanceSheetContent.Value]
    var whyTitle: String
    var accessibilityLabel: String
    /// A derived row opens the Event or Place it is read from. Nil otherwise.
    var opens: Opens?
    private var source: RecordsSource
    private var locale: Locale
    private var orders: NameFormatOrders

    /// How many records sit behind the row.
    var recordCount: Int {
        switch source {
        case .field(let field): field.outcomes.count
        case .spanning(let fields): fields.reduce(0) { $0 + $1.outcomes.count }
        case .derived(let owners, _, _): owners.reduce(0) { $0 + $1.field.outcomes.count }
        }
    }

    init(
        field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let shown = PVProvenanceSheetContent.Shown(field: field, locale: locale, orders: orders)
        let spokenLead = shown.listedValues.count > 1 ? shown.listedValues.map(\.text).joined(separator: "; ") : shown.lead
        id = field.propertyID
        label = field.label
        self.shown = shown
        count = ReconciledValueDisplay.countLine(field)
        against = ReconciledValueDisplay.againstLine(field)
        otherValuesLabel = ReconciledValueDisplay.otherValuesLabel(field)
        otherValues = ReconciledValueDisplay.otherValues(field).map {
            PVProvenanceSheetContent.value($0, propertyKey: field.propertyKey, locale: locale, orders: orders)
        }
        whyTitle = shown.listedValues.count > 1
            ? L10n.Conclusions.whyThese(shown.listedValues.count, label: field.label)
            : field.valueType == .name
                ? L10n.Conclusions.whyTitleName(shown.lead ?? field.label)
                : L10n.Conclusions.whyTitle(shown.lead ?? field.label)
        accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: field.label, lead: spokenLead, field: field)
        source = .field(field)
        self.locale = locale
        self.orders = orders
    }

    /// The sheet or Why content of the row, built from its records now. The
    /// row's label and shown value go in, so a relabelled row and its sheet
    /// cannot disagree.
    func provenanceContent() -> PVProvenanceSheetContent {
        switch source {
        case .field(let field):
            PVProvenanceSheetContent(field: field, title: label, shown: shown, locale: locale, orders: orders)
        case .spanning(let fields):
            .spanning(fields: fields, title: label, shown: shown, locale: locale, orders: orders)
        case .derived(let owners, let shownOwnerID, let kind):
            .derived(
                owners: owners, shownOwnerID: shownOwnerID, kind: kind,
                title: label, shown: shown, locale: locale, orders: orders
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
        let lead = lead.flatMap { $0.isEmpty ? nil : $0 }
        let spoken = lead.map { L10n.DesignSystem.spokenJoin(label, $0) }
            ?? ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        return Self(
            id: id, label: label,
            shown: .init(face: .date, lead: lead, emptyText: emptyText, badge: nil),
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            whyTitle: lead.map { L10n.Conclusions.whyTitle($0) } ?? "",
            accessibilityLabel: spoken,
            source: .spanning(fields), locale: locale, orders: NameFormatOrders()
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
        let lead = lead.isEmpty ? nil : lead
        let spoken: String
        if let lead {
            let parts = [label, lead] + (mixed ? [L10n.string(L10n.ProvenanceSheet.badgeMixed)] : [])
            spoken = L10n.DesignSystem.spokenJoin(parts) ?? label
        } else {
            spoken = ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        }
        var row = Self(
            id: id, label: label,
            shown: .init(face: face, lead: lead, emptyText: emptyText, badge: mixed && lead != nil ? .mixed : nil),
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            whyTitle: "", accessibilityLabel: spoken,
            source: .derived(owners: owners, shownOwnerID: shownOwnerID, kind: kind),
            locale: locale, orders: orders
        )
        if row.recordCount > 0 { row.whyTitle = L10n.Conclusions.whyTitle(lead ?? label) }
        row.opens = opens
        return row
    }

    /// A row with nothing behind it yet.
    static func empty(id: String, label: String, face: PVProvenanceFace, emptyText: String) -> Self {
        Self(
            id: id, label: label,
            shown: .init(face: face, lead: nil, emptyText: emptyText, badge: nil),
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [], whyTitle: "",
            accessibilityLabel: ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil),
            source: .spanning([]), locale: .autoupdatingCurrent, orders: NameFormatOrders()
        )
    }

    private init(
        id: String, label: String, shown: PVProvenanceSheetContent.Shown, count: String?, against: String?,
        otherValuesLabel: String?, otherValues: [PVProvenanceSheetContent.Value], whyTitle: String,
        accessibilityLabel: String, source: RecordsSource, locale: Locale, orders: NameFormatOrders
    ) {
        self.id = id
        self.label = label
        self.shown = shown
        self.count = count
        self.against = against
        self.otherValuesLabel = otherValuesLabel
        self.otherValues = otherValues
        self.whyTitle = whyTitle
        self.accessibilityLabel = accessibilityLabel
        self.source = source
        self.locale = locale
        self.orders = orders
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
