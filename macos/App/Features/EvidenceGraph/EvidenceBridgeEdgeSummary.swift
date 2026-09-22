import Foundation

/// Durable bridge body copy from cited Properties (S7-D3 §3.1 + endpoint labels).
///
/// Preferred form joins the two connected subjects’ working labels with the
/// relational phrase — same full-sentence shape as the citation-composer crumb
/// (S7-D4): “Alice participated in Birth”, “John is the father of Mary”.
/// Working `subjects.label` on the bridge itself is for uncited chrome only.
enum EvidenceBridgeEdgeSummary {
    /// Preferred edge-summary phrase for a cited bridge.
    static func phrase(for placed: SourceGraphPlacedBridge) -> String {
        switch placed.kind {
        case .location:
            return locationPhrase(in: placed.observations)
        case .relationship:
            return relationshipPhrase(in: placed.observations)
        case .participation:
            return participationPhrase(in: placed.observations)
        }
    }

    private static func locationPhrase(in observations: [CatalogObservation]) -> String {
        let event = subjectDisplay(propertyKey: "event", in: observations)
        let place = subjectDisplay(propertyKey: "place", in: observations)
        if let event, let place {
            return L10n.EvidenceGraph.bridgeSummaryLocation(event: event, place: place)
        }
        return String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare)
    }

    private static func relationshipPhrase(in observations: [CatalogObservation]) -> String {
        let person = subjectDisplay(propertyKey: "person", in: observations)
        let related = subjectDisplay(propertyKey: "related_to", in: observations)
        let type = termDisplay(propertyKey: "relationship_type", in: observations)
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
        if let type {
            return L10n.EvidenceGraph.bridgeSummaryRelationshipTypeOnly(type: type)
        }
        return String(localized: L10n.EvidenceGraph.bridgeSummaryRelationshipBare)
    }

    private static func participationPhrase(in observations: [CatalogObservation]) -> String {
        let person = subjectDisplay(propertyKey: "person", in: observations)
        let event = subjectDisplay(propertyKey: "event", in: observations)
        let role = termDisplay(propertyKey: "role", in: observations)
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
        if let role {
            return L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: role)
        }
        return String(localized: L10n.EvidenceGraph.bridgeSummaryParticipationBare)
    }

    private static func subjectDisplay(
        propertyKey: String,
        in observations: [CatalogObservation]
    ) -> String? {
        display(propertyKey: propertyKey, in: observations)
    }

    private static func termDisplay(
        propertyKey: String,
        in observations: [CatalogObservation]
    ) -> String? {
        display(propertyKey: propertyKey, in: observations)
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
