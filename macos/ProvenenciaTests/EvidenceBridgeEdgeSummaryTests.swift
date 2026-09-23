import Foundation
import Testing
@testable import Provenencia

@Suite
struct EvidenceBridgeEdgeSummaryTests {
    private func observation(
        propertyKey: String,
        valueText: String,
        valueType: String = "term"
    ) -> CatalogObservation {
        CatalogObservation(
            id: "obs-\(propertyKey)",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "bridge-1",
            propertyID: "prop-\(propertyKey)",
            polarity: "positive",
            valueText: valueText,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: valueType == "subject" ? "subj-\(propertyKey)" : "",
            valueTermID: valueType == "term" && !valueText.isEmpty ? "term-1" : "",
            propertyKey: propertyKey,
            propertyLabel: propertyKey,
            propertyValueType: valueType
        )
    }

    private func bridge(
        kind: EvidenceBridgeKind,
        observations: [CatalogObservation]
    ) -> SourceGraphPlacedBridge {
        SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "bridge-1",
                ref: "CPA-1",
                sourceID: "src-1",
                subjectTypeID: "type-\(kind.rawValue)",
                label: "Working label",
                description: ""
            ),
            kind: kind,
            typeLabel: kind.rawValue.capitalized,
            gridX: 0,
            gridY: 0,
            isCited: !observations.isEmpty,
            observations: observations
        )
    }

    @Test func locationPhraseIsRelationalOnly() {
        let placed = bridge(kind: .location, observations: [
            observation(propertyKey: "event", valueText: "Marriage", valueType: "subject"),
            observation(propertyKey: "place", valueText: "Leeds", valueType: "subject"),
        ])
        #expect(EvidenceBridgeEdgeSummary.phrase(for: placed) == String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare))
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryLocation(event: "Marriage", place: "Leeds")
        )
    }

    @Test func locationFallsBackWithoutEndpoints() {
        let phrase = EvidenceBridgeEdgeSummary.phrase(
            for: bridge(kind: .location, observations: [
                observation(propertyKey: "event", valueText: "Marriage", valueType: "subject"),
            ])
        )
        #expect(phrase == String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare))
    }

    @Test func relationshipPhraseUsesTypeOnly() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueText: "John", valueType: "subject"),
            observation(propertyKey: "related_to", valueText: "Mary", valueType: "subject"),
            observation(propertyKey: "relationship_type", valueText: "father"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.phrase(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryRelationshipTypeOnly(type: "father")
        )
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryRelationship(
                    person: "John",
                    type: "father",
                    related: "Mary"
                )
        )
    }

    @Test func relationshipFallsBackWithoutType() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueText: "Alice", valueType: "subject"),
            observation(propertyKey: "related_to", valueText: "Bob", valueType: "subject"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.phrase(for: placed)
                == String(localized: L10n.EvidenceGraph.bridgeSummaryRelationshipBare)
        )
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryRelationshipFallback(
                    person: "Alice",
                    related: "Bob"
                )
        )
    }

    @Test func participationPhraseUsesRoleOnly() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueText: "Margt.", valueType: "subject"),
            observation(propertyKey: "event", valueText: "1851 census", valueType: "subject"),
            observation(propertyKey: "role", valueText: "head of household"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.phrase(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: "head of household")
        )
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryParticipation(
                    person: "Margt.",
                    role: "head of household",
                    event: "1851 census"
                )
        )
    }

    @Test func participationFallsBackWithoutRole() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueText: "Alice", valueType: "subject"),
            observation(propertyKey: "event", valueText: "Birth", valueType: "subject"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.phrase(for: placed)
                == String(localized: L10n.EvidenceGraph.bridgeSummaryParticipationBare)
        )
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed)
                == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                    person: "Alice",
                    event: "Birth"
                )
        )
    }

    @Test func participationBareWithoutEndpoints() {
        let phrase = EvidenceBridgeEdgeSummary.phrase(
            for: bridge(kind: .participation, observations: [
                observation(propertyKey: "note", valueText: "x", valueType: "text"),
            ])
        )
        #expect(phrase == String(localized: L10n.EvidenceGraph.bridgeSummaryParticipationBare))
    }
}
