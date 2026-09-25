import Foundation

/// Durable bridge copy: relational **phrase** on cards (S7-D3 §3.1) and full
/// **sentence** wherever a bridge is named (S8-D8 §2.4).
enum EvidenceBridgeEdgeSummary {
    /// Relational phrase only — “is the father of”, “participated as witness”.
    static func phrase(for placed: SourceGraphPlacedBridge) -> String {
        phrase(kind: placed.kind, term: termDisplay(kind: placed.kind, in: placed.observations))
    }

    /// Snapshot-aware sentence. Nouns come from cited endpoints; unreadable
    /// edges fall back to the stored label or kind phrase · ref. Never empty.
    static func sentence(for bridge: SourceGraphPlacedBridge, in snapshot: SourceGraphSnapshot) -> String {
        let person = noun(propertyKey: "person", in: bridge, snapshot: snapshot)
        let related = noun(propertyKey: "related_to", in: bridge, snapshot: snapshot)
        let event = noun(propertyKey: "event", in: bridge, snapshot: snapshot)
        let place = noun(propertyKey: "place", in: bridge, snapshot: snapshot)
        let term = termDisplay(kind: bridge.kind, in: bridge.observations)
        let nouns: [String?]
        switch bridge.kind {
        case .location:
            nouns = [event, place]
        case .relationship:
            nouns = [person, related]
        case .participation:
            nouns = [person, event]
        }
        if nouns.contains(where: { $0 == nil }) {
            return nonEmpty(wholeNameFallback(for: bridge))
        }
        return nonEmpty(
            sentence(
                kind: bridge.kind,
                person: person,
                related: related,
                event: event,
                place: place,
                term: term
            ),
            fallback: wholeNameFallback(for: bridge)
        )
    }

    static func phrase(kind: EvidenceBridgeKind, term: String?) -> String {
        switch kind {
        case .location:
            return String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare)
        case .relationship:
            if let term {
                return L10n.EvidenceGraph.bridgeSummaryRelationshipTypeOnly(type: term)
            }
            return String(localized: L10n.EvidenceGraph.bridgeSummaryRelationshipBare)
        case .participation:
            if let term {
                return L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: term)
            }
            return String(localized: L10n.EvidenceGraph.bridgeSummaryParticipationBare)
        }
    }

    static func sentence(
        kind: EvidenceBridgeKind,
        person: String?,
        related: String?,
        event: String?,
        place: String?,
        term: String?
    ) -> String {
        switch kind {
        case .location:
            return locationSentence(event: event, place: place)
        case .relationship:
            return relationshipSentence(person: person, related: related, type: term)
        case .participation:
            return participationSentence(person: person, event: event, role: term)
        }
    }

    private static func locationSentence(event: String?, place: String?) -> String {
        if let event, let place {
            return L10n.EvidenceGraph.bridgeSummaryLocation(event: event, place: place)
        }
        return String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare)
    }

    private static func relationshipSentence(person: String?, related: String?, type: String?) -> String {
        if let person, let related, let type {
            return L10n.EvidenceGraph.bridgeSummaryRelationship(
                person: person,
                type: type,
                related: related
            )
        }
        if let person, let related {
            return L10n.EvidenceGraph.bridgeSummaryRelationshipFallback(
                person: person,
                related: related
            )
        }
        return phrase(kind: .relationship, term: type)
    }

    private static func participationSentence(person: String?, event: String?, role: String?) -> String {
        if let person, let event, let role {
            return L10n.EvidenceGraph.bridgeSummaryParticipation(
                person: person,
                role: role,
                event: event
            )
        }
        if let person, let event {
            return L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                person: person,
                event: event
            )
        }
        return phrase(kind: .participation, term: role)
    }

    private static func noun(
        propertyKey: String,
        in bridge: SourceGraphPlacedBridge,
        snapshot: SourceGraphSnapshot
    ) -> String? {
        guard let subjectID = subjectID(propertyKey: propertyKey, in: bridge.observations),
              let placed = snapshot.subjects.first(where: { $0.id == subjectID })
        else { return nil }
        let label = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if !label.isEmpty { return label }
        let typed = L10n.EvidenceGraph.bridgeNounTypeAndRef(
            type: placed.typeLabel,
            ref: placed.subject.ref
        )
        return typed.isEmpty ? nil : typed
    }

    private static func wholeNameFallback(for bridge: SourceGraphPlacedBridge) -> String {
        let stored = bridge.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if !stored.isEmpty { return stored }
        return L10n.EvidenceGraph.bridgeNameKindAndRef(
            phrase: phrase(for: bridge),
            ref: bridge.subject.ref
        )
    }

    private static func nonEmpty(_ value: String, fallback: String = "") -> String {
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty { return trimmed }
        let next = fallback.trimmingCharacters(in: .whitespacesAndNewlines)
        if !next.isEmpty { return next }
        return "—"
    }

    private static func subjectID(
        propertyKey: String,
        in observations: [CatalogObservation]
    ) -> String? {
        let id = observations.first(where: { $0.propertyKey == propertyKey })?.valueSubjectID
            .trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return id.isEmpty ? nil : id
    }

    private static func termDisplay(
        kind: EvidenceBridgeKind,
        in observations: [CatalogObservation]
    ) -> String? {
        let key = kind == .relationship ? "relationship_type" : "role"
        guard let observation = observations.first(where: { $0.propertyKey == key }) else {
            return nil
        }
        let catalog = ObservationValueDisplay.string(for: observation)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let termKey = observation.valueTermKey
        if !termKey.isEmpty {
            return PropertyTermDisplay.name(
                key: termKey,
                propertyKey: key,
                catalogLabel: catalog
            )
        }
        return catalog.isEmpty ? nil : catalog
    }
}
