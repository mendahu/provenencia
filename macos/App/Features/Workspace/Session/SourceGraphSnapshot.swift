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

    /// Joins catalog rows into placed primaries and bridges. Unplaced subjects
    /// and `source` types are omitted. Endpoint ids come from cited
    /// `value_subject_id` on edge Properties (`person` / `event` / `place` /
    /// `related_to`). Leftover uncited JSON links do not draw lines.
    static func build(
        sourceId: String,
        subjects: [CatalogSubject],
        positions: [CatalogSubjectPosition],
        types: [CatalogSubjectType],
        provisionalLinks: [EvidenceProvisionalLink] = [],
        observations: [CatalogObservation] = []
    ) -> SourceGraphSnapshot {
        let typeByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0) })
        let positionBySubject = Dictionary(uniqueKeysWithValues: positions.map { ($0.subjectID, $0) })
        let observationsBySubject = Dictionary(grouping: observations, by: \.subjectID)
        _ = provisionalLinks

        var placed: [SourceGraphPlacedSubject] = []
        var placedBridges: [SourceGraphPlacedBridge] = []
        placed.reserveCapacity(subjects.count)
        placedBridges.reserveCapacity(subjects.count)

        for subject in subjects {
            guard let type = typeByID[subject.subjectTypeID],
                  let position = positionBySubject[subject.id]
            else { continue }

            if let kind = EvidencePrimaryKind(rawValue: type.key) {
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
                let ends = Self.citedEndpoints(kind: kind, observations: subjectObservations)
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

    /// Cited A/B endpoints from edge `value_subject_id` (S7-10).
    static func citedEndpoints(
        kind: EvidenceBridgeKind,
        observations: [CatalogObservation]
    ) -> (a: String?, b: String?) {
        func subjectID(for key: String) -> String? {
            let id = observations.first(where: { $0.propertyKey == key })?.valueSubjectID
                .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
            return id.isEmpty ? nil : id
        }
        switch kind {
        case .participation:
            return (subjectID(for: "person"), subjectID(for: "event"))
        case .relationship:
            return (subjectID(for: "person"), subjectID(for: "related_to"))
        case .location:
            return (subjectID(for: "event"), subjectID(for: "place"))
        }
    }

    /// Card row order: property label A→Z, then key, then Observation ref.
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

    /// Returns a copy with provisional endpoint ids applied to matching bridges.
    func attaching(links: [EvidenceProvisionalLink]) -> SourceGraphSnapshot {
        guard !bridges.isEmpty else { return self }
        let byBridge = Dictionary(uniqueKeysWithValues: links.map { ($0.bridgeSubjectID, $0) })
        let next = bridges.map { bridge -> SourceGraphPlacedBridge in
            guard let link = byBridge[bridge.id] else { return bridge }
            var copy = bridge
            copy.endpointAID = link.endpointAID
            copy.endpointBID = link.endpointBID
            return copy
        }
        return SourceGraphSnapshot(sourceId: sourceId, subjects: subjects, bridges: next)
    }

    /// Returns a copy with one subject's or bridge's grid cell updated.
    func updatingPosition(subjectID: String, gridX: Int64, gridY: Int64) -> SourceGraphSnapshot {
        if let index = subjects.firstIndex(where: { $0.id == subjectID }) {
            var next = subjects
            next[index].gridX = gridX
            next[index].gridY = gridY
            return SourceGraphSnapshot(sourceId: sourceId, subjects: next, bridges: bridges)
        }
        if let index = bridges.firstIndex(where: { $0.id == subjectID }) {
            var next = bridges
            next[index].gridX = gridX
            next[index].gridY = gridY
            return SourceGraphSnapshot(sourceId: sourceId, subjects: subjects, bridges: next)
        }
        return self
    }
}

