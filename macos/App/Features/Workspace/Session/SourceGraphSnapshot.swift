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
}

/// Catalog rows for one Source's Evidence graph. Types stay on
/// `subjectFieldsWorkspace` so a citation save does not reload vocabulary.
struct SourceGraphRows: Sendable, Equatable {
    var sourceId: String
    var subjects: [CatalogSubject]
    var positions: [CatalogSubjectPosition]
    var observations: [CatalogObservation]

    init(
        sourceId: String,
        subjects: [CatalogSubject] = [],
        positions: [CatalogSubjectPosition] = [],
        observations: [CatalogObservation] = []
    ) {
        self.sourceId = sourceId
        self.subjects = subjects
        self.positions = positions
        self.observations = observations
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
            observations: observations
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
        rules: [CatalogConnectRule] = []
    ) -> SourceGraphSnapshot {
        let typeByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0) })
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
                        observations: subjectObservations
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
                        endpointBID: ends.b
                    )
                )
            }
        }

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

