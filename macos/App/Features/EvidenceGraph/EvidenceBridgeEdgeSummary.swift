import Foundation

/// Durable bridge copy: relational **phrase** on cards (S7-D3 §3.1) and full
/// **sentence** on the composer crumb (S7-D4).
enum EvidenceBridgeEdgeSummary {
    /// Relational phrase only — “is the father of”, “participated as witness”.
    static func phrase(for placed: SourceGraphPlacedBridge) -> String {
        phrase(kind: placed.kind, term: termDisplay(kind: placed.kind, in: placed.observations))
    }

    /// Full sentence — “John is the father of Mary”.
    static func sentence(for placed: SourceGraphPlacedBridge) -> String {
        sentence(
            kind: placed.kind,
            person: subjectDisplay(propertyKey: "person", in: placed.observations),
            related: subjectDisplay(propertyKey: "related_to", in: placed.observations),
            event: subjectDisplay(propertyKey: "event", in: placed.observations),
            place: subjectDisplay(propertyKey: "place", in: placed.observations),
            term: termDisplay(kind: placed.kind, in: placed.observations)
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

    private static func subjectDisplay(
        propertyKey: String,
        in observations: [CatalogObservation]
    ) -> String? {
        display(propertyKey: propertyKey, in: observations)
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

    private static func display(
        propertyKey: String,
        in observations: [CatalogObservation]
    ) -> String? {
        guard let observation = observations.first(where: { $0.propertyKey == propertyKey })
        else { return nil }
        let rendered = ObservationValueDisplay.string(for: observation)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        return rendered.isEmpty ? nil : rendered
    }
}
