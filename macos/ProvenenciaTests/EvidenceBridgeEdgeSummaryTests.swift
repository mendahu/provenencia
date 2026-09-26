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
        ref: String,
        observations: [CatalogObservation] = []
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
            isCited: !observations.isEmpty,
            observations: observations
        )
    }

    private func identityObservation(
        subjectID: String,
        propertyKey: String,
        valueText: String,
        polarity: ObservationPolarity = .positive,
        id: String = ""
    ) -> CatalogObservation {
        CatalogObservation(
            id: id.isEmpty ? "obs-\(subjectID)-\(propertyKey)-\(valueText)" : id,
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: subjectID,
            propertyID: "prop-\(propertyKey)",
            polarity: polarity.rawValue,
            valueText: valueText,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: propertyKey == "name" ? "nv-1" : "",
            nameForm: propertyKey == "name" ? valueText : "",
            valueSubjectID: "",
            valueTermID: propertyKey == "event_type" ? "term-1" : "",
            propertyKey: propertyKey,
            propertyLabel: propertyKey,
            propertyValueType: propertyKey == "name"
                ? PropertyValueType.name.rawValue
                : (propertyKey == "event_type"
                    ? PropertyValueType.term.rawValue
                    : PropertyValueType.text.rawValue)
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

    @Test func nameFormBeatsWorkingLabel() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "related_to", valueSubjectID: "p2"),
            observation(propertyKey: "relationship_type", valueText: "father", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(
                id: "p1",
                kind: .person,
                label: "Wm.",
                ref: "CPR-1",
                observations: [identityObservation(subjectID: "p1", propertyKey: "name", valueText: "Wm Robins")]
            ),
            primary(
                id: "p2",
                kind: .person,
                label: "John",
                ref: "CPR-2",
                observations: [identityObservation(subjectID: "p2", propertyKey: "name", valueText: "John Robins")]
            ),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryRelationship(
                    person: "Wm Robins",
                    type: "father",
                    related: "John Robins"
                )
        )
    }

    @Test func missingNameFallsBackToWorkingLabel() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "related_to", valueSubjectID: "p2"),
            observation(propertyKey: "relationship_type", valueText: "father", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(id: "p1", kind: .person, label: "Wm.", ref: "CPR-1"),
            primary(
                id: "p2",
                kind: .person,
                label: "John",
                ref: "CPR-2",
                observations: [identityObservation(subjectID: "p2", propertyKey: "name", valueText: "John Robins")]
            ),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryRelationship(
                    person: "Wm.",
                    type: "father",
                    related: "John Robins"
                )
        )
    }

    @Test func participationUsesEventType() {
        let placed = bridge(kind: .participation, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "event", valueSubjectID: "e1"),
            observation(propertyKey: "role", valueText: "head", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(
                id: "p1",
                kind: .person,
                label: "William",
                ref: "CPR-1",
                observations: [identityObservation(subjectID: "p1", propertyKey: "name", valueText: "Wm Robins")]
            ),
            primary(
                id: "e1",
                kind: .event,
                label: "1871 household",
                ref: "CEV-1",
                observations: [
                    identityObservation(
                        subjectID: "e1",
                        propertyKey: "event_type",
                        valueText: "Census Enumeration"
                    ),
                ]
            ),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryParticipation(
                    person: "Wm Robins",
                    role: "head",
                    event: "Census Enumeration"
                )
        )
    }

    @Test func locationUsesToponym() {
        let placed = bridge(kind: .location, observations: [
            observation(propertyKey: "event", valueSubjectID: "e1"),
            observation(propertyKey: "place", valueSubjectID: "pl1"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(
                id: "e1",
                kind: .event,
                label: "Household",
                ref: "CEV-1",
                observations: [
                    identityObservation(
                        subjectID: "e1",
                        propertyKey: "event_type",
                        valueText: "Census Enumeration"
                    ),
                ]
            ),
            primary(
                id: "pl1",
                kind: .place,
                label: "Erin",
                ref: "CPL-1",
                observations: [identityObservation(subjectID: "pl1", propertyKey: "toponym", valueText: "Erin")]
            ),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryLocation(
                    event: "Census Enumeration",
                    place: "Erin"
                )
        )
    }

    @Test func firstPositiveIdentityWins() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "related_to", valueSubjectID: "p2"),
            observation(propertyKey: "relationship_type", valueText: "father", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(
                id: "p1",
                kind: .person,
                label: "Working",
                ref: "CPR-1",
                observations: [
                    identityObservation(
                        subjectID: "p1",
                        propertyKey: "name",
                        valueText: "Wrong Name",
                        polarity: .negative,
                        id: "obs-neg"
                    ),
                    identityObservation(
                        subjectID: "p1",
                        propertyKey: "name",
                        valueText: "Wm Robins",
                        id: "obs-pos"
                    ),
                ]
            ),
            primary(id: "p2", kind: .person, label: "John Robins", ref: "CPR-2"),
        ])
        #expect(
            EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
                == L10n.EvidenceGraph.bridgeSummaryRelationship(
                    person: "Wm Robins",
                    type: "father",
                    related: "John Robins"
                )
        )
    }

    @Test func competingNamesAreNotJoined() {
        let placed = bridge(kind: .relationship, observations: [
            observation(propertyKey: "person", valueSubjectID: "p1"),
            observation(propertyKey: "related_to", valueSubjectID: "p2"),
            observation(propertyKey: "relationship_type", valueText: "father", valueType: "term"),
        ])
        let snap = snapshot(bridge: placed, subjects: [
            primary(
                id: "p1",
                kind: .person,
                label: "Working",
                ref: "CPR-1",
                observations: [
                    identityObservation(
                        subjectID: "p1",
                        propertyKey: "name",
                        valueText: "Wm Robins",
                        id: "obs-a"
                    ),
                    identityObservation(
                        subjectID: "p1",
                        propertyKey: "name",
                        valueText: "William Robins",
                        id: "obs-b"
                    ),
                ]
            ),
            primary(id: "p2", kind: .person, label: "John Robins", ref: "CPR-2"),
        ])
        let sentence = EvidenceBridgeEdgeSummary.sentence(for: placed, in: snap)
        #expect(sentence.contains("Wm Robins"))
        #expect(!sentence.contains("William Robins"))
    }
}
