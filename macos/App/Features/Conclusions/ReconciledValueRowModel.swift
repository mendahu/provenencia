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

    /// The handle a derived value is read from.
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
    /// The Why's records, grouped by the entity that owns them.
    var recordGroups: [ReconciliationRecordGroup]
    var records: [ReconciliationRecord] { recordGroups.flatMap(\.records) }
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
            .map { ReconciledValueDisplay.string(for: $0.value, locale: locale, orders: orders) }
            .flatMap { $0.isEmpty ? nil : $0 }
        let listed = Self.listedValues(of: field, locale: locale, orders: orders)
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
        otherValues = ReconciledValueDisplay.otherValues(field).map { Self.otherValue($0, locale: locale, orders: orders) }
        listedValues = listed
        whyTitle = listed.count > 1
            ? L10n.Conclusions.whyThese(listed.count, label: field.label)
            : field.valueType == .name
                ? L10n.Conclusions.whyTitleName(lead ?? field.label)
                : L10n.Conclusions.whyTitle(lead ?? field.label)
        let records = Self.records([field], locale: locale, orders: orders)
        recordGroups = [ReconciliationRecordGroup(
            owner: nil,
            values: Self.valueGroups(records, field: field, locale: locale, orders: orders)
        )]
        accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: field.label, lead: spokenLead, field: field)
        provenanceSummary = Self.joinSummary([
            field.state == .single
                ? L10n.Conclusions.provenanceStateSingle(L10n.Conclusions.sourceCount(ReconciledValueDisplay.sourceCount(field)))
                : ReconciledValueDisplay.spokenState(field),
            ReconciledValueDisplay.againstLine(field),
            records.isEmpty ? nil : L10n.Conclusions.provenanceRecordCount(records.count),
        ])
    }

    /// The summary's parts, in order, joined with the summary separator; nil when none.
    private static func joinSummary(_ parts: [String?]) -> String? {
        let parts = parts.compactMap { $0 }
        guard let first = parts.first else { return nil }
        return parts.dropFirst().reduce(first) { L10n.Conclusions.provenanceSummaryJoin($0, $1) }
    }

    /// A displayed value as a line: its text and its Source count.
    static func otherValue(
        _ value: CatalogReconciledValue,
        locale: Locale,
        orders: NameFormatOrders
    ) -> OtherValue {
        OtherValue(
            rank: value.rank,
            text: ReconciledValueDisplay.string(for: value.value, locale: locale, orders: orders),
            support: L10n.Conclusions.sourceCount(value.support)
        )
    }

    /// A group's records in runs of equal `groupRank`, in the order Go sent them.
    /// Headings show when the group has two or more runs, and over a group whose
    /// only run counts under no displayed value; one value run needs none.
    static func valueGroups(
        _ records: [ReconciliationRecord],
        field: CatalogConclusionField,
        locale: Locale,
        orders: NameFormatOrders
    ) -> [ReconciliationRecordGroup.ValueGroup] {
        var runs: [(rank: Int, records: [ReconciliationRecord])] = []
        for record in records {
            if runs.last?.rank == record.groupRank {
                runs[runs.count - 1].records.append(record)
            } else {
                runs.append((record.groupRank, [record]))
            }
        }
        let showsHeadings = runs.count > 1 || runs.first?.rank == 0
        return runs.map { run in
            let heading: ReconciliationRecordGroup.ValueGroup.Heading?
            if !showsHeadings {
                heading = nil
            } else if run.rank > 0, let value = field.displayedValues.first(where: { $0.rank == run.rank }) {
                heading = .value(otherValue(value, locale: locale, orders: orders))
            } else {
                heading = .otherRecords
            }
            return .init(heading: heading, records: run.records)
        }
    }

    /// The records of several links to one Place, each link's own runs merged by
    /// displayed value: runs of the same value join in link order under one
    /// heading counting the distinct Sources of their outcomes, and records under no displayed
    /// value join last under one "other records" heading.
    private static func pooledValueGroups(
        _ members: [(field: CatalogConclusionField, records: [ReconciliationRecord])],
        locale: Locale,
        orders: NameFormatOrders
    ) -> [ReconciliationRecordGroup.ValueGroup] {
        var valueRuns: [(value: OtherValue, support: Int, outcomes: [CatalogReconcilerOutcome], records: [ReconciliationRecord])] = []
        var other: [ReconciliationRecord] = []
        for member in members {
            var counted: Set<String> = []
            var runs: [(rank: Int, records: [ReconciliationRecord])] = []
            for record in member.records {
                if runs.last?.rank == record.groupRank {
                    runs[runs.count - 1].records.append(record)
                } else {
                    runs.append((record.groupRank, [record]))
                }
            }
            for run in runs {
                guard run.rank > 0, let value = member.field.displayedValues.first(where: { $0.rank == run.rank }) else {
                    other += run.records
                    continue
                }
                let shown = otherValue(value, locale: locale, orders: orders)
                let outcomes = member.field.outcomes.filter { $0.groupRank == run.rank }
                if let index = valueRuns.firstIndex(where: { $0.value.text == shown.text }) {
                    valueRuns[index].records += run.records
                    valueRuns[index].outcomes += outcomes
                    if counted.insert(shown.text).inserted { valueRuns[index].support += value.support }
                } else {
                    counted.insert(shown.text)
                    valueRuns.append((shown, value.support, outcomes, run.records))
                }
            }
        }
        let showsHeadings = valueRuns.count + (other.isEmpty ? 0 : 1) > 1 || valueRuns.isEmpty
        var groups = valueRuns.map { run -> ReconciliationRecordGroup.ValueGroup in
            var value = run.value
            value.support = L10n.Conclusions.sourceCount(
                ReconciledValueDisplay.distinctSourceCount(of: run.outcomes, fallback: run.support)
            )
            return .init(heading: showsHeadings ? .value(value) : nil, records: run.records)
        }
        if !other.isEmpty { groups.append(.init(heading: showsHeadings ? .otherRecords : nil, records: other)) }
        return groups
    }

    /// Every kept value of a multi-valued field, lead included. A weak
    /// spelling is not displayed, so it stays out of this list.
    private static func listedValues(
        of field: CatalogConclusionField,
        locale: Locale,
        orders: NameFormatOrders
    ) -> [OtherValue] {
        guard field.state == .multiple else { return [] }
        return field.displayedValues.compactMap { value in
            let listed = Self.otherValue(value, locale: locale, orders: orders)
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
        let records = Self.records(fields, locale: locale, orders: NameFormatOrders())
        let spoken = shown.map { L10n.Conclusions.a11yList(label, rest: $0) }
            ?? ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        return Self(
            id: id, label: label, style: .date, lead: shown, emptyText: emptyText,
            badge: nil, count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            listedValues: [],
            whyTitle: shown.map { L10n.Conclusions.whyTitle($0) } ?? "",
            recordGroups: [ReconciliationRecordGroup(owner: nil, records: records)],
            accessibilityLabel: spoken,
            provenanceSummary: records.isEmpty ? nil : L10n.Conclusions.provenanceRecordCount(records.count)
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
        style: ValueStyle,
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
            spoken = parts.dropFirst().reduce(parts[0]) { L10n.Conclusions.a11yList($0, rest: $1) }
        } else {
            spoken = ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        }
        // One group per Place (or per other owner): the links to one Place
        // pool their records, in link order, the shown owner's group first.
        let ordered = owners.filter { $0.ownerID == shownOwnerID } + owners.filter { $0.ownerID != shownOwnerID }
        var ownerIDs: [String] = []
        var membersOf: [String: [CatalogOwnerField]] = [:]
        for owner in ordered {
            if membersOf[owner.ownerID] == nil { ownerIDs.append(owner.ownerID) }
            membersOf[owner.ownerID, default: []].append(owner)
        }
        let groups = ownerIDs.map { ownerID in
            let members = membersOf[ownerID] ?? []
            let memberRecords = members.map { member -> [ReconciliationRecord] in
                var records = Self.records([member.field], locale: locale, orders: orders)
                if ownerID != shownOwnerID {
                    for index in records.indices { records[index].counted = false }
                }
                return records
            }
            return ReconciliationRecordGroup(
                owner: ownerIDs.count > 1 || ownerID != shownOwnerID
                    ? Self.ownerLine(
                        members[0], style: style, shown: ownerIDs.count > 1 && ownerID == shownOwnerID,
                        locale: locale, orders: orders
                    )
                    : nil,
                values: members.count == 1
                    ? Self.valueGroups(memberRecords[0], field: members[0].field, locale: locale, orders: orders)
                    : Self.pooledValueGroups(
                        zip(members.map(\.field), memberRecords).map { ($0, $1) }, locale: locale, orders: orders
                    )
            )
        }
        let recordCount = groups.reduce(0) { $0 + $1.records.count }
        let summary: String?
        if recordCount == 0 {
            summary = nil
        } else if groups.count > 1, let kind {
            summary = L10n.Conclusions.provenanceRecordsFromOwners(records: recordCount, owners: groups.count, kind: kind)
        } else {
            summary = L10n.Conclusions.provenanceRecordCount(recordCount)
        }
        var row = Self(
            id: id, label: label, style: style, lead: shown, emptyText: emptyText,
            badge: mixed && shown != nil ? .mixed : nil,
            count: nil, against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [],
            whyTitle: "", recordGroups: groups, accessibilityLabel: spoken, provenanceSummary: summary
        )
        if !row.records.isEmpty { row.whyTitle = L10n.Conclusions.whyTitle(shown ?? label) }
        row.opens = opens
        return row
    }

    /// The group line of a derived row's owner, titled as its list titles it:
    /// an Event by its header (else its label or ref), a Place by the kept
    /// names of the Place its link names (else its label or ref).
    private static func ownerLine(
        _ owner: CatalogOwnerField,
        style: ValueStyle,
        shown: Bool,
        locale: Locale,
        orders: NameFormatOrders
    ) -> ReconciliationRecordGroup.OwnerLine {
        let title: String
        if let event = owner.event {
            title = EventTitleDisplay.title(event, locale: locale, orders: orders)
        } else if let place = owner.place {
            title = PlaceTitleDisplay.titleSource(place.titleParts).text
        } else {
            title = EventTitleDisplay.untitled(owner.entity).text
        }
        let identity = owner.place?.entity ?? owner.entity
        let ref = title == identity.ref ? nil : identity.ref
        let line = ref.map { L10n.Conclusions.a11yList(title, rest: $0) } ?? title
        let spoken = shown ? L10n.Conclusions.provenanceOwnerShown(line) : line
        return .init(
            id: identity.id,
            title: title,
            ref: ref,
            spoken: spoken,
            shown: shown
        )
    }

    /// Every outcome of the fields, as Why records.
    private static func records(
        _ fields: [CatalogConclusionField],
        locale: Locale,
        orders: NameFormatOrders
    ) -> [ReconciliationRecord] {
        fields.flatMap { field in
            field.outcomes.map { ReconciliationRecord(outcome: $0, in: field, locale: locale, orders: orders) }
        }
    }

    /// A row with nothing behind it yet.
    static func empty(id: String, label: String, style: ValueStyle, emptyText: String) -> Self {
        Self(
            id: id, label: label, style: style, lead: nil, emptyText: emptyText, badge: nil, count: nil,
            against: nil, otherValuesLabel: nil, otherValues: [], listedValues: [], whyTitle: "", recordGroups: [],
            accessibilityLabel: ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil),
            provenanceSummary: nil
        )
    }

    private init(
        id: String, label: String, style: ValueStyle, lead: String?, emptyText: String,
        badge: ReconciledValueDisplay.StateBadge?, count: String?, against: String?,
        otherValuesLabel: String?, otherValues: [OtherValue], listedValues: [OtherValue], whyTitle: String,
        recordGroups: [ReconciliationRecordGroup], accessibilityLabel: String, provenanceSummary: String?
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
    static func place(_ place: CatalogHeaderPlace) -> Self {
        let title = PlaceTitleDisplay.titleSource(place.titleParts).text
        return Self(
            location: .placeDetail(entityId: place.entity.id, ref: place.entity.ref, title: title),
            label: L10n.Conclusions.openPlace(title)
        )
    }
}

private extension CatalogOwnerField {
    /// What the row names, orders and inks the owner by: the Place, else the entity.
    var ownerID: String { place?.entity.id ?? entity.id }
}
