import Foundation
import Testing
@testable import Provenencia

@Suite
struct EvidenceBridgeEdgeSummaryTests {
    private func observation(
        propertyKey: String,
        valueText: String = "",
        valueSubjectID: String = "",
        valueType: String = "subject"
    ) -> CatalogObservation {
        CatalogObservation(
            id: "obs-\(propertyKey)",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "bridge-1",
            propertyID: "prop-\(propertyKey)",
            polarity: ObservationPolarity.positive.rawValue,
            valueText: valueText,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: valueSubjectID,
            valueTermID: valueType == "term" && !valueText.isEmpty ? "term-1" : "",
            propertyKey: propertyKey,
            propertyLabel: propertyKey,
            propertyValueType: valueType
        )
    }

    private func primary(
        id: String,
        kind: EvidencePrimaryKind,
        label: String,
        ref: String
    ) -> SourceGraphPlacedSubject {
        SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: id,
                ref: ref,
                sourceID: "src-1",
                subjectTypeID: "type-\(kind.rawValue)",
                label: label,
                description: ""
            ),
            kind: kind,
            typeLabel: kind.rawValue.capitalized,
            gridX: 0,
            gridY: 0,
            isCited: false
        )
    }

    private func bridge(
        kind: EvidenceBridgeKind,
        observations: [CatalogObservation],
        label: String = "Working label",
        ref: String = "CPA-1"
    ) -> SourceGraphPlacedBridge {
        let ends = SourceGraphSnapshot.citedEndpoints(
            kind: kind,
            observations: observations,
            rules: CatalogConnectRule.productMatrix
        )
        return SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "bridge-1",
                ref: ref,
                sourceID: "src-1",
                subjectTypeID: "type-\(kind.rawValue)",
                label: label,
                description: ""
            ),
            kind: kind,
            typeLabel: kind.rawValue.capitalized,
            gridX: 0,
            gridY: 0,
            isCited: !observations.isEmpty,
            observations: observations,
            endpointAID: ends.a,
            endpointBID: ends.b
        )
    }

    private func snapshot(
        bridge: SourceGraphPlacedBridge,
        subjects: [SourceGraphPlacedSubject]
    ) -> SourceGraphSnapshot {
        SourceGraphSnapshot(sourceId: "src-1", subjects: subjects, bridges: [bridge])
    }

    @Test func locationPhraseIsRelationalOnly() {
        let placed = bridge(kind: .location, observations: [
            observation(propertyKey: "event", valueSubjectID: "e1"),
            observation(propertyKey: "place", valueSubjectID: "pl1"),
        ])
        #expect(EvidenceBridgeEdgeSummary.phrase(for: placed) == String(localized: L10n.EvidenceGraph.bridgeSummaryLocationBare))
    }

    @Test func workingLabelBecomesNoun() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "event", valueSubjectID: "e1"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "p1", kind: .person, label: "Margt.", ref: "CPR-1"),
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                    person: "Margt.",
                    event: "Birth"
                )
        )
    }

    @Test func blankLabelFallsBackToTypeAndRef() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "event", valueSubjectID: "e1"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "p1", kind: .person, label: "   ", ref: "CPR-F4N2P"),
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                    person: L10n.EvidenceGraph.bridgeNounTypeAndRef(type: "Person", ref: "CPR-F4N2P"),
                    event: "Birth"
                )
        )
    }

    @Test func missingEndpointUsesStoredLabel() {
        let placed = bridge(
            kind: .participation,
            observations: [
                observation(propertyKey: "person", valueSubjectID: "missing"),
                observation(propertyKey: "event", valueSubjectID: "e1"),
            ],
            label: "Mary's baptism"
        )
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap) == "Mary's baptism")
    }

    @Test func missingEndpointAndLabelUsesKindPhraseAndRef() {
        let placed = bridge(
            kind: .participation,
            observations: [
                observation(propertyKey: "person", valueSubjectID: "missing"),
                observation(propertyKey: "event", valueSubjectID: "e1"),
            ],
            label: ""
        )
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeNameKindAndRef(
                    phrase: String(localized: L10n.EvidenceGraph.bridgeSummaryParticipationBare),
                    ref: "CPA-1"
                )
        )
    }

    @Test func missingRoleUsesRolelessTemplate() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "event", valueSubjectID: "e1"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "p1", kind: .person, label: "Alice", ref: "CPR-1"),
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                    person: "Alice",
                    event: "Birth"
                )
        )
    }

    @Test func sentenceNeverEmpty() {
        let placed = bridge(kind: .relationship, observations: [], label: "", ref: "")
        let snap = snapshot(bridge: placed, subjects: [])
        #expect(!EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap).isEmpty)
    }

    @Test func citedCardSentenceKeepsEndpointNames() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "event", valueSubjectID: "e1"),
            observation(propertyKey: "role", valueText: "subject", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "p1", kind: .person, label: "Jerry", ref: "CPR-1"),
            primary(id: "e1", kind: .event, label: "Birth", ref: "CEV-1"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryParticipation(
                    person: "Jerry",
                    role: "subject",
                    event: "Birth"
                )
        )
        #expect(
            EvidenceBridgeCard.accessibilityLabel(for: placed, in: snap)
                == "Participation, CPA-1, \(EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap))"
        )
    }
}
