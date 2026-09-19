import Foundation

/// Primary subject kinds drawn on the Evidence graph (S6-02).
/// Bridge / `source` types are filtered out of the canvas payload.
enum EvidencePrimaryKind: String, Sendable, Equatable, CaseIterable {
    case person
    case event
    case place
}

/// A primary subject with a persisted grid position for the Evidence graph.
struct SourceGraphPlacedSubject: Identifiable, Sendable, Equatable {
    var id: String { subject.id }
    var subject: CatalogSubject
    var kind: EvidencePrimaryKind
    var typeLabel: String
    var gridX: Int64
    var gridY: Int64
    /// True when the subject has at least one Observation. Always false until
    /// Observation wiring exists; previews/tests may set it for chrome contrast.
    var isCited: Bool
}

/// Source-scoped Evidence graph payload (subjects + positions).
struct SourceGraphSnapshot: Sendable, Equatable {
    var sourceId: String
    var subjects: [SourceGraphPlacedSubject]

    init(sourceId: String, subjects: [SourceGraphPlacedSubject] = []) {
        self.sourceId = sourceId
        self.subjects = subjects
    }

    /// Joins catalog rows into placed primaries. Bridges and unplaced subjects
    /// are omitted. Stable order: label, then ref.
    static func build(
        sourceId: String,
        subjects: [CatalogSubject],
        positions: [CatalogSubjectPosition],
        types: [CatalogSubjectType],
        isCited: (CatalogSubject) -> Bool = { _ in false }
    ) -> SourceGraphSnapshot {
        let typeByID = Dictionary(uniqueKeysWithValues: types.map { ($0.id, $0) })
        let positionBySubject = Dictionary(uniqueKeysWithValues: positions.map { ($0.subjectID, $0) })

        var placed: [SourceGraphPlacedSubject] = []
        placed.reserveCapacity(subjects.count)
        for subject in subjects {
            guard let type = typeByID[subject.subjectTypeID],
                  let kind = EvidencePrimaryKind(rawValue: type.key),
                  let position = positionBySubject[subject.id]
            else { continue }
            placed.append(
                SourceGraphPlacedSubject(
                    subject: subject,
                    kind: kind,
                    typeLabel: type.label,
                    gridX: position.gridX,
                    gridY: position.gridY,
                    isCited: isCited(subject)
                )
            )
        }
        placed.sort { lhs, rhs in
            let labelCompare = lhs.subject.label.localizedStandardCompare(rhs.subject.label)
            if labelCompare != .orderedSame { return labelCompare == .orderedAscending }
            return lhs.subject.ref.localizedStandardCompare(rhs.subject.ref) == .orderedAscending
        }
        return SourceGraphSnapshot(sourceId: sourceId, subjects: placed)
    }

    /// Returns a copy with one subject's grid cell updated (drag / arrow move).
    func updatingPosition(subjectID: String, gridX: Int64, gridY: Int64) -> SourceGraphSnapshot {
        var next = subjects
        guard let index = next.firstIndex(where: { $0.id == subjectID }) else { return self }
        next[index].gridX = gridX
        next[index].gridY = gridY
        return SourceGraphSnapshot(sourceId: sourceId, subjects: next)
    }
}
