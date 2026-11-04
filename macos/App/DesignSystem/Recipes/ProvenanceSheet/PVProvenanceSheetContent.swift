import SwiftUI

/// How a value reads: names, text and terms in the body face, dates in mono,
/// places in italic. One type for the Vitals row, the provenance sheet and the
/// Why (S13-36 D-6).
enum PVProvenanceFace: Equatable {
    /// Names, text, terms, integers.
    case text
    /// Mono, like every date in the app.
    case date
    /// Italic, like every place.
    case place

    /// The face of a displayed value.
    var valueFont: Font {
        switch self {
        case .text: PVFont.body(size: PVTypeScale.body)
        case .date: PVFont.mono(size: 15)
        case .place: PVFont.body(size: PVTypeScale.body, italic: true)
        }
    }

    /// The face of a record's "Read as" cell: a phrase ("not …", "—") is italic.
    func readAsFont(isPhrase: Bool) -> Font {
        if isPhrase {
            return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
        switch self {
        case .text: return PVFont.body(size: PVTypeScale.bodySmall)
        case .date: return PVFont.mono(size: PVTypeScale.caption)
        case .place: return PVFont.body(size: PVTypeScale.bodySmall, italic: true)
        }
    }
}

/// What a record's link opens: its Source, its Citation, or its Observation.
/// The recipe names no place; the caller maps a link to one.
enum PVProvenanceLink: Equatable {
    case source
    case citation
    case observation
}

/// What a provenance sheet says: a value, the records behind it, and how they
/// group. Pure; built from the wire records by the assembler below, so a caller
/// holds the wire records and asks for a sheet when one opens (S13-36).
///
/// No view code goes in an extension of this type: its nested `Group` shadows
/// SwiftUI's.
struct PVProvenanceSheetContent: Equatable {
    /// A displayed value as a line: its text and its Source count.
    struct Value: Equatable, Identifiable {
        var rank: Int
        var text: String
        var support: String
        var id: Int { rank }
    }

    /// One record in a field's provenance: what it was read as, its Source, and
    /// what the auto-reconciler did with it.
    struct Record: Equatable, Identifiable {
        var id: String
        var readAs: String
        /// "not …" or "—": set as a phrase, not in the value's face.
        var readAsIsPhrase: Bool
        /// Went into the shown value (primary ink); the rest are secondary.
        var counted: Bool
        var phrase: String
        var mark: PVSymbol
        var observationRef: String
        var subjectRef: String
        /// The wire record the rest is read from.
        var outcome: CatalogReconcilerOutcome
        /// "Source, Artifact, Citation, page" (`CitationLineDisplay`).
        var citation: CitationLineDisplay.Parts

        /// The record as one VoiceOver line: what it was read as, its citation
        /// line, the outcome, then its Observation and Subject refs. The links
        /// follow as separate controls.
        var spoken: String {
            L10n.DesignSystem.spokenJoin([readAs, citation.line, phrase, observationRef, subjectRef]) ?? ""
        }
    }

    /// The records of one owning entity in a derived row. `owner` is nil for a
    /// row with one owner, which needs no group line.
    struct Group: Equatable, Identifiable {
        /// A group line: the owning entity as its list titles it.
        struct OwnerLine: Equatable {
            /// The owner's entity id.
            var id: String
            var title: String
            /// The owner's ref; nil when the title is the ref.
            var ref: String?
            var spoken: String
            /// The owner behind the row's shown value, when the row has several groups.
            var shown: Bool = false
        }

        /// The records of one run of the group that count under the same
        /// displayed value, in the order Go sent them.
        struct ValueGroup: Equatable {
            enum Heading: Equatable {
                /// The displayed value the records went into.
                case value(Value)
                /// Records that count under no displayed value.
                case otherRecords
            }

            var heading: Heading?
            var records: [Record]
        }

        var owner: OwnerLine?
        /// The group's records in runs by value. One run with no heading unless
        /// the group needs headings.
        var values: [ValueGroup]
        /// Every record of the group, in order.
        var records: [Record] { values.flatMap(\.records) }

        var id: String { owner?.id ?? "" }

        init(owner: OwnerLine?, values: [ValueGroup]) {
            self.owner = owner
            self.values = values
        }

        /// A group whose records form one run with no heading.
        init(owner: OwnerLine?, records: [Record]) {
            self.init(owner: owner, values: records.isEmpty ? [] : [ValueGroup(heading: nil, records: records)])
        }
    }

    /// What a value line shows: the lead, every kept value of a multi-valued
    /// field, else the field's empty text in italic, and the state. The one
    /// derivation of it, for the Vitals row and the sheet header alike.
    struct Shown: Equatable {
        var face: PVProvenanceFace
        /// The lead value; nil when the field is empty.
        var lead: String?
        /// Every displayed value when the field keeps several (`multiple`). Empty
        /// otherwise, including a mixed field, which still discloses the rest.
        var listedValues: [Value]
        var emptyText: String
        var badge: ReconciledValueDisplay.StateBadge?
    }

    var title: String
    var shown: Shown
    /// What the content knows about the records behind it. Nil when there is nothing to say.
    var summary: String?
    var groups: [Group]

    /// Every record, in order.
    var records: [Record] { groups.flatMap(\.records) }
}

// MARK: - Assembler

extension PVProvenanceSheetContent.Record {
    init(
        outcome: CatalogReconcilerOutcome,
        in field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let citation = CitationLineDisplay.parts(outcome)
        self.init(
            id: outcome.observationID,
            readAs: ReconciledValueDisplay.readAs(outcome, propertyKey: field.propertyKey, locale: locale, orders: orders),
            readAsIsPhrase: ReconciledValueDisplay.readAsIsPhrase(outcome, propertyKey: field.propertyKey),
            counted: ReconciledValueDisplay.counted(outcome, in: field),
            phrase: ReconciledValueDisplay.outcomePhrase(outcome, in: field, locale: locale),
            mark: ReconciledValueDisplay.outcomeMark(outcome, in: field),
            observationRef: outcome.observationRef,
            subjectRef: outcome.subjectRef,
            outcome: outcome,
            citation: citation
        )
    }
}

extension PVProvenanceSheetContent.Shown {
    /// What one reconciled field shows. A weak spelling is not displayed, so it
    /// stays out of the listed values.
    init(
        field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let lead = field.displayedValues.first
            .map { ReconciledValueDisplay.string(for: $0.value, propertyKey: field.propertyKey, locale: locale, orders: orders) }
            .flatMap { $0.isEmpty ? nil : $0 }
        let listed = field.state == .multiple
            ? field.displayedValues.compactMap { value -> PVProvenanceSheetContent.Value? in
                let listed = PVProvenanceSheetContent.value(value, propertyKey: field.propertyKey, locale: locale, orders: orders)
                return listed.text.isEmpty ? nil : listed
            }
            : []
        self.init(
            face: field.valueType == .date ? .date : .text,
            lead: lead,
            listedValues: listed,
            emptyText: ReconciledValueDisplay.emptyText(propertyKey: field.propertyKey),
            badge: ReconciledValueDisplay.stateBadge(field)
        )
    }
}

extension PVProvenanceSheetContent {
    /// One field's provenance: its state line, what it counts against, and its
    /// records in runs by displayed value.
    init(
        field: CatalogConclusionField,
        title: String,
        shown: Shown? = nil,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let records = Self.records([field], locale: locale, orders: orders)
        self.init(
            title: title,
            shown: shown ?? Shown(field: field, locale: locale, orders: orders),
            summary: Self.joinSummary([
                field.state == .single
                    ? L10n.Conclusions.provenanceStateSingle(L10n.Conclusions.sourceCount(ReconciledValueDisplay.sourceCount(field)))
                    : ReconciledValueDisplay.spokenState(field),
                ReconciledValueDisplay.againstLine(field),
                records.isEmpty ? nil : L10n.Conclusions.provenanceRecordCount(records.count),
            ]),
            groups: [Group(
                owner: nil,
                values: Self.valueGroups(records, field: field, locale: locale, orders: orders)
            )]
        )
    }

    /// Every record of several fields in one group with no heading, as a row
    /// whose value spans them (a start–end date).
    static func spanning(
        fields: [CatalogConclusionField],
        title: String,
        shown: Shown,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> Self {
        let records = Self.records(fields, locale: locale, orders: orders)
        return Self(
            title: title,
            shown: shown,
            summary: records.isEmpty ? nil : L10n.Conclusions.provenanceRecordCount(records.count),
            groups: [Group(owner: nil, records: records)]
        )
    }

    /// A row read off the canonical graph. Its records are the owning handles'
    /// fields: the shown owner's first, each under a group line when there are
    /// several or when the group is not the shown owner's, and only the shown
    /// owner's records in primary ink.
    static func derived(
        owners: [CatalogOwnerField],
        shownOwnerID: String?,
        kind: CatalogDerivedRowKind?,
        title: String,
        shown: Shown,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> Self {
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
            let memberRecords = members.map { member -> [Record] in
                var records = Self.records([member.field], locale: locale, orders: orders)
                if ownerID != shownOwnerID {
                    for index in records.indices { records[index].counted = false }
                }
                return records
            }
            return Group(
                owner: ownerIDs.count > 1 || ownerID != shownOwnerID
                    ? Self.ownerLine(
                        members[0], shown: ownerIDs.count > 1 && ownerID == shownOwnerID,
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
        return Self(title: title, shown: shown, summary: summary, groups: groups)
    }

    /// A displayed value as a line: its text and its Source count. The one
    /// builder of a `Value`.
    static func value(
        _ value: CatalogReconciledValue,
        propertyKey: String,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> Value {
        Value(
            rank: value.rank,
            text: ReconciledValueDisplay.string(for: value.value, propertyKey: propertyKey, locale: locale, orders: orders),
            support: L10n.Conclusions.sourceCount(value.support)
        )
    }

    /// A group's records in runs of equal `groupRank`, in the order Go sent them.
    /// Headings show when the group has two or more runs, and over a group whose
    /// only run counts under no displayed value; one value run needs none.
    static func valueGroups(
        _ records: [Record],
        field: CatalogConclusionField,
        locale: Locale,
        orders: NameFormatOrders
    ) -> [Group.ValueGroup] {
        let runs = Self.runs(of: records)
        let showsHeadings = runs.count > 1 || runs.first?.rank == 0
        return runs.map { run in
            let heading: Group.ValueGroup.Heading?
            if !showsHeadings {
                heading = nil
            } else if run.rank > 0, let value = field.displayedValues.first(where: { $0.rank == run.rank }) {
                heading = .value(Self.value(value, propertyKey: field.propertyKey, locale: locale, orders: orders))
            } else {
                heading = .otherRecords
            }
            return .init(heading: heading, records: run.records)
        }
    }

    /// The records in runs of equal `groupRank`, in order.
    private static func runs(of records: [Record]) -> [(rank: Int, records: [Record])] {
        var runs: [(rank: Int, records: [Record])] = []
        for record in records {
            if runs.last?.rank == record.outcome.groupRank {
                runs[runs.count - 1].records.append(record)
            } else {
                runs.append((record.outcome.groupRank, [record]))
            }
        }
        return runs
    }

    /// The records of several links to one Place, each link's own runs merged by
    /// displayed value: runs of the same value join in link order under one
    /// heading counting the distinct Sources of their outcomes, and records under no displayed
    /// value join last under one "other records" heading.
    private static func pooledValueGroups(
        _ members: [(field: CatalogConclusionField, records: [Record])],
        locale: Locale,
        orders: NameFormatOrders
    ) -> [Group.ValueGroup] {
        var valueRuns: [(value: Value, support: Int, outcomes: [CatalogReconcilerOutcome], records: [Record])] = []
        var other: [Record] = []
        for member in members {
            var counted: Set<String> = []
            for run in Self.runs(of: member.records) {
                guard run.rank > 0, let value = member.field.displayedValues.first(where: { $0.rank == run.rank }) else {
                    other += run.records
                    continue
                }
                let shown = Self.value(value, propertyKey: member.field.propertyKey, locale: locale, orders: orders)
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
        var groups = valueRuns.map { run -> Group.ValueGroup in
            var value = run.value
            value.support = L10n.Conclusions.sourceCount(
                ReconciledValueDisplay.distinctSourceCount(of: run.outcomes, fallback: run.support)
            )
            return .init(heading: showsHeadings ? .value(value) : nil, records: run.records)
        }
        if !other.isEmpty { groups.append(.init(heading: showsHeadings ? .otherRecords : nil, records: other)) }
        return groups
    }

    /// The group line of a derived row's owner, titled as its list titles it:
    /// an Event by its summary (else its label or ref), a Place by the kept
    /// names of the Place its link names (else its label or ref).
    private static func ownerLine(
        _ owner: CatalogOwnerField,
        shown: Bool,
        locale: Locale,
        orders: NameFormatOrders
    ) -> Group.OwnerLine {
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
        let line = ref.map { L10n.DesignSystem.spokenJoin(title, $0) } ?? title
        let spoken = shown ? L10n.Conclusions.provenanceOwnerShown(line) : line
        return .init(
            id: identity.id,
            title: title,
            ref: ref,
            spoken: spoken,
            shown: shown
        )
    }

    /// Every outcome of the fields, as records.
    private static func records(
        _ fields: [CatalogConclusionField],
        locale: Locale,
        orders: NameFormatOrders
    ) -> [Record] {
        fields.flatMap { field in
            field.outcomes.map { Record(outcome: $0, in: field, locale: locale, orders: orders) }
        }
    }

    /// The summary's parts, in order, joined with the summary separator; nil when none.
    private static func joinSummary(_ parts: [String?]) -> String? {
        let parts = parts.compactMap { $0 }
        guard let first = parts.first else { return nil }
        return parts.dropFirst().reduce(first) { L10n.Conclusions.provenanceSummaryJoin($0, $1) }
    }
}

private extension CatalogOwnerField {
    /// What the content names, orders and inks the owner by: the Place, else the entity.
    var ownerID: String { place?.entity.id ?? entity.id }
}
