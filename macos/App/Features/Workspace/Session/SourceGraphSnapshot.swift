import Foundation

/// Primary subject kinds drawn on the Evidence graph (S6-02).
enum EvidencePrimaryKind: String, Sendable, Equatable, CaseIterable {
    case person
    case event
    case place
}

/// Bridge / association kinds drawn as mid-cards (S6-04).
enum EvidenceBridgeKind: String, Sendable, Equatable, CaseIterable {
    case relationship
    case participation
    case location
}

/// A primary subject with a persisted grid position for the Evidence graph.
struct SourceGraphPlacedSubject: Identifiable, Sendable, Equatable {
    var id: String { subject.id }
    var subject: CatalogSubject
    var kind: EvidencePrimaryKind
    var typeLabel: String
    var gridX: Int64
    var gridY: Int64
    /// True when the subject has at least one Observation on this Source.
    var isCited: Bool
    /// Observations about this subject (property summaries for card rows).
    var observations: [CatalogObservation] = []
    /// The handle this subject belongs to (accepted Identity Claim), or nil.
    var membership: CatalogSubjectMembership?

    /// Subject names and the first place name, filled from this Source's
    /// participation and location bridges when the snapshot is built. People
    /// stay on their own name, then label.
    var subjectNames: [String] = []
    var placeName: String = ""

    /// What the subject is called where it is promoted: its first asserted
    /// naming Observation (a Person's `name` form, a Place's `toponym`), else
    /// its working label, else its ref. An Event uses `EventTitleDisplay`.
    /// Subject and place parts come from the participation and location
    /// bridges on this graph. The date is not part of the title.
    var displayName: String {
        if kind == .event {
            return EventTitleDisplay.title(eventTitleParts)
        }
        if let named = observations.lazy.compactMap(Self.namingValue(for: kind)).first {
            return named
        }
        let label = subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? subject.ref : label
    }

    private var eventTitleParts: EventTitleParts {
        var parts = EventTitleParts(label: subject.label, ref: subject.ref)
        for observation in observations where observation.polarity != "negative" {
            let text = observation.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
            switch observation.propertyKey {
            case "event_name" where parts.recordedName.isEmpty && !text.isEmpty:
                parts.recordedName = text
            case "event_type" where parts.typeLabel.isEmpty && parts.typeKey.isEmpty:
                parts.typeLabel = text
                parts.typeKey = observation.valueTermKey
            default:
                break
            }
        }
        parts.subjects = subjectNames
        parts.place = placeName
        return parts
    }

    /// The value of an asserted Observation that names a subject of this kind.
    private static func namingValue(for kind: EvidencePrimaryKind) -> (CatalogObservation) -> String? {
        { observation in
            guard observation.polarity != "negative" else { return nil }
            let value: String
            switch (kind, observation.propertyKey) {
            case (.person, "name"): value = observation.nameForm
            case (.place, "toponym"): value = observation.valueText
            default: return nil
            }
            let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
            return trimmed.isEmpty ? nil : trimmed
        }
    }
}

/// A bridge subject with position and provisional endpoint ids (S6-04).
struct SourceGraphPlacedBridge: Identifiable, Sendable, Equatable {
    var id: String { subject.id }
    var subject: CatalogSubject
    var kind: EvidenceBridgeKind
    var typeLabel: String
    var gridX: Int64
    var gridY: Int64
    /// True when the bridge subject has at least one Observation on this Source.
    var isCited: Bool = false
    /// Observations about this bridge (edge Properties + disambiguation terms).
    var observations: [CatalogObservation] = []
    /// Provisional A/B endpoint subject ids (app-local until Citations exist).
    var endpointAID: String?
    var endpointBID: String?
    /// The association handle this bridge belongs to, or nil (filed from S9-28).
    var membership: CatalogSubjectMembership?
}

/// Catalog rows for one Source's Evidence graph. Types stay on
/// `propertiesWorkspace` so a citation save does not reload vocabulary.
struct SourceGraphRows: Sendable, Equatable {
    var sourceId: String
    var subjects: [CatalogSubject]
    var positions: [CatalogSubjectPosition]
    var observations: [CatalogObservation]
    /// Accepted handles of promoted subjects; absent subjects are unpromoted.
    var memberships: [CatalogSubjectMembership]

    init(
        sourceId: String,
        subjects: [CatalogSubject] = [],
        positions: [CatalogSubjectPosition] = [],
        observations: [CatalogObservation] = [],
        memberships: [CatalogSubjectMembership] = []
    ) {
        self.sourceId = sourceId
        self.subjects = subjects
        self.positions = positions
        self.observations = observations
        self.memberships = memberships
    }

    /// Returns a copy with one subject's grid cell updated.
    func updatingPosition(subjectID: String, gridX: Int64, gridY: Int64) -> SourceGraphRows {
        var next = positions
        if let index = next.firstIndex(where: { $0.subjectID == subjectID }) {
            next[index].gridX = gridX
            next[index].gridY = gridY
        } else {
            next.append(CatalogSubjectPosition(subjectID: subjectID, gridX: gridX, gridY: gridY))
        }
        return SourceGraphRows(
            sourceId: sourceId,
            subjects: subjects,
            positions: next,
            observations: observations,
            memberships: memberships
        )
    }
}

/// Source-scoped Evidence graph payload (primaries + bridges + positions).
struct SourceGraphSnapshot: Sendable, Equatable {
    var sourceId: String
    var subjects: [SourceGraphPlacedSubject]
    var bridges: [SourceGraphPlacedBridge]

    init(
        sourceId: String,
        subjects: [SourceGraphPlacedSubject] = [],
        bridges: [SourceGraphPlacedBridge] = []
    ) {
        self.sourceId = sourceId
        self.subjects = subjects
        self.bridges = bridges
    }

    /// Joins catalog rows with subject types. Unplaced subjects and `source`
    /// types are omitted. Endpoint ids come from cited `value_subject_id` on
    /// the matching connect rule's edge Properties.
    static func build(
        rows: SourceGraphRows,
        types: [CatalogSubjectType],
        rules: [CatalogConnectRule]
    ) -> SourceGraphSnapshot {
        build(
            sourceId: rows.sourceId,
            subjects: rows.subjects,
            positions: rows.positions,
            types: types,
            observations: rows.observations,
            memberships: rows.memberships,
            rules: rules
        )
    }

    /// Joins catalog rows into placed primaries and bridges. Unplaced subjects
    /// and `source` types are omitted. Endpoint ids come from cited
    /// `value_subject_id` on the matching connect rule's edges.
    ///
    /// Each card's Observations are sorted with ``observationDisplayOrder``
    /// (label, then key, then ref). Do not change that order: conflict
    /// brackets join adjacent same-key rows, and a different sort splits
    /// a competing Property into one-row marks.
    static func build(
        sourceId: String,
        subjects: [CatalogSubject],
        positions: [CatalogSubjectPosition],
        types: [CatalogSubjectType],
        observations: [CatalogObservation] = [],
        memberships: [CatalogSubjectMembership] = [],
        rules: [CatalogConnectRule] = []
    ) -> SourceGraphSnapshot {
        let typeByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0) })
        let membershipBySubject = Dictionary(
            memberships.map { ($0.subjectID, $0) },
            uniquingKeysWith: { first, _ in first }
        )
        let positionBySubject = Dictionary(uniqueKeysWithValues: positions.map { ($0.subjectID, $0) })
        let observationsBySubject = Dictionary(grouping: observations, by: \.subjectID)

        var placed: [SourceGraphPlacedSubject] = []
        var placedBridges: [SourceGraphPlacedBridge] = []
        placed.reserveCapacity(subjects.count)
        placedBridges.reserveCapacity(subjects.count)

        for subject in subjects {
            guard let type = typeByID[subject.subjectTypeID],
                  let position = positionBySubject[subject.id]
            else { continue }

            if let kind = EvidencePrimaryKind(rawValue: type.key) {
                // Keep this sort. Conflict brackets join adjacent same-key
                // rows; changing order can split a run into one-row marks.
                let subjectObservations = (observationsBySubject[subject.id] ?? [])
                    .sorted(by: Self.observationDisplayOrder)
                placed.append(
                    SourceGraphPlacedSubject(
                        subject: subject,
                        kind: kind,
                        typeLabel: type.label,
                        gridX: position.gridX,
                        gridY: position.gridY,
                        isCited: !subjectObservations.isEmpty,
                        observations: subjectObservations,
                        membership: membershipBySubject[subject.id]
                    )
                )
            } else if let kind = EvidenceBridgeKind(rawValue: type.key) {
                let subjectObservations = (observationsBySubject[subject.id] ?? [])
                    .sorted(by: Self.observationDisplayOrder)
                let ends = Self.citedEndpoints(
                    kind: kind,
                    observations: subjectObservations,
                    rules: rules
                )
                placedBridges.append(
                    SourceGraphPlacedBridge(
                        subject: subject,
                        kind: kind,
                        typeLabel: type.label,
                        gridX: position.gridX,
                        gridY: position.gridY,
                        isCited: !subjectObservations.isEmpty,
                        observations: subjectObservations,
                        endpointAID: ends.a,
                        endpointBID: ends.b,
                        membership: membershipBySubject[subject.id]
                    )
                )
            }
        }

        placed = Self.eventTitles(subjects: placed, bridges: placedBridges)
        placed.sort { lhs, rhs in
            let labelCompare = lhs.subject.label.localizedStandardCompare(rhs.subject.label)
            if labelCompare != .orderedSame { return labelCompare == .orderedAscending }
            return lhs.subject.ref.localizedStandardCompare(rhs.subject.ref) == .orderedAscending
        }
        placedBridges.sort { lhs, rhs in
            let labelCompare = lhs.subject.label.localizedStandardCompare(rhs.subject.label)
            if labelCompare != .orderedSame { return labelCompare == .orderedAscending }
            return lhs.subject.ref.localizedStandardCompare(rhs.subject.ref) == .orderedAscending
        }
        return SourceGraphSnapshot(sourceId: sourceId, subjects: placed, bridges: placedBridges)
    }

    /// Event cards take their subject and place parts from this graph's
    /// bridges: a subject-role participation names a person, a location names
    /// a place. People keep their own naming rule.
    private static func eventTitles(
        subjects: [SourceGraphPlacedSubject],
        bridges: [SourceGraphPlacedBridge]
    ) -> [SourceGraphPlacedSubject] {
        let byID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0) })
        return subjects.map { subject in
            guard subject.kind == .event else { return subject }
            var people: [(bridgeRef: String, personRef: String, name: String)] = []
            var places: [(ref: String, name: String)] = []
            for bridge in bridges {
                guard let otherID = counterpart(bridge, of: subject.id),
                      let other = byID[otherID]
                else { continue }
                switch (bridge.kind, other.kind) {
                case (.participation, .person) where isSubjectRole(bridge):
                    people.append((bridge.subject.ref, other.subject.ref, citedText(other, key: "name", name: true)))
                case (.location, .place):
                    places.append((other.subject.ref, citedText(other, key: "toponym", name: false)))
                default:
                    break
                }
            }
            people.sort { lhs, rhs in
                if lhs.bridgeRef != rhs.bridgeRef { return lhs.bridgeRef < rhs.bridgeRef }
                return lhs.personRef < rhs.personRef
            }
            places.sort { $0.ref < $1.ref }
            var copy = subject
            copy.subjectNames = people.map(\.name)
            copy.placeName = places.first?.name ?? ""
            return copy
        }
    }

    private static func counterpart(_ bridge: SourceGraphPlacedBridge, of id: String) -> String? {
        if bridge.endpointAID == id { return bridge.endpointBID }
        if bridge.endpointBID == id { return bridge.endpointAID }
        return nil
    }

    private static func isSubjectRole(_ bridge: SourceGraphPlacedBridge) -> Bool {
        bridge.observations.contains {
            $0.propertyKey == "role" && $0.polarity != "negative" && $0.valueTermKey == "subject"
        }
    }

    /// The first asserted name form, or toponym text. Empty when the subject
    /// has none, so an Event title can still say "unnamed person".
    private static func citedText(_ subject: SourceGraphPlacedSubject, key: String, name: Bool) -> String {
        for observation in subject.observations where observation.polarity != "negative" && observation.propertyKey == key {
            let value = (name ? observation.nameForm : observation.valueText)
                .trimmingCharacters(in: .whitespacesAndNewlines)
            if !value.isEmpty { return value }
        }
        return ""
    }

    /// Cited A/B endpoints from the matching rule's `edges` (A then B).
    static func citedEndpoints(
        kind: EvidenceBridgeKind,
        observations: [CatalogObservation],
        rules: [CatalogConnectRule]
    ) -> (a: String?, b: String?) {
        func subjectID(for key: String) -> String? {
            let id = observations.first(where: { $0.propertyKey == key })?.valueSubjectID
                .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            return id.isEmpty ? nil : id
        }
        let edges = rules.first { $0.bridgeTypeKey == kind.rawValue && !$0.refuse }?.edges ?? []
        let a = edges.first.map { subjectID(for: $0.propertyKey) } ?? nil
        let b = edges.dropFirst().first.map { subjectID(for: $0.propertyKey) } ?? nil
        return (a.flatMap { $0 }, b.flatMap { $0 })
    }

    /// Card row order: property label A→Z, then key, then Observation ref.
    /// Same-key rows must stay adjacent so `EvidenceCitedPropertyMarks.conflictRuns`
    /// can paint one bracket per competing Property.
    private static func observationDisplayOrder(
        _ lhs: CatalogObservation,
        _ rhs: CatalogObservation
    ) -> Bool {
        let labelCompare = lhs.propertyLabel.localizedStandardCompare(rhs.propertyLabel)
        if labelCompare != .orderedSame { return labelCompare == .orderedAscending }
        let keyCompare = lhs.propertyKey.localizedStandardCompare(rhs.propertyKey)
        if keyCompare != .orderedSame { return keyCompare == .orderedAscending }
        return lhs.ref.localizedStandardCompare(rhs.ref) == .orderedAscending
    }

}

