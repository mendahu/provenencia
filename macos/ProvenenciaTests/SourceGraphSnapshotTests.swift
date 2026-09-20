import Foundation
import Testing
@testable import Provenencia

@Suite
struct SourceGraphSnapshotTests {
    private let personType = CatalogSubjectType(
        id: "type-person",
        key: "person",
        origin: "provenencia",
        label: "Person",
        description: "",
        refPrefix: "PER",
        candidateRefPrefix: "CPR"
    )
    private let eventType = CatalogSubjectType(
        id: "type-event",
        key: "event",
        origin: "provenencia",
        label: "Event",
        description: "",
        refPrefix: "EVT",
        candidateRefPrefix: "CEV"
    )
    private let locationType = CatalogSubjectType(
        id: "type-location",
        key: "location",
        origin: "provenencia",
        label: "Location",
        description: "",
        refPrefix: "LOC",
        candidateRefPrefix: "CLO"
    )

    @Test func buildMarksCitedFromObservations() {
        let alice = CatalogSubject(
            id: "s-alice",
            ref: "CPR-A",
            sourceID: "src-1",
            subjectTypeID: personType.id,
            label: "Alice",
            description: ""
        )
        let observation = CatalogObservation(
            id: "obs-1",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: alice.id,
            propertyID: "prop-1",
            polarity: "positive",
            valueText: "Boston",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "toponym",
            propertyLabel: "Toponym",
            propertyValueType: "text"
        )
        let snapshot = SourceGraphSnapshot.build(
            sourceId: "src-1",
            subjects: [alice],
            positions: [CatalogSubjectPosition(subjectID: alice.id, gridX: 0, gridY: 0)],
            types: [personType],
            observations: [observation]
        )
        #expect(snapshot.subjects.count == 1)
        #expect(snapshot.subjects[0].isCited)
        #expect(snapshot.subjects[0].observations.count == 1)
        #expect(snapshot.subjects[0].observations[0].valueText == "Boston")
    }

    @Test func buildKeepsPlacedPrimariesAndBridgesWithLinks() {
        let alice = CatalogSubject(
            id: "s-alice",
            ref: "CPR-A",
            sourceID: "src-1",
            subjectTypeID: personType.id,
            label: "Alice",
            description: ""
        )
        let birth = CatalogSubject(
            id: "s-birth",
            ref: "CEV-B",
            sourceID: "src-1",
            subjectTypeID: eventType.id,
            label: "Birth",
            description: ""
        )
        let bridge = CatalogSubject(
            id: "s-loc",
            ref: "CLO-L",
            sourceID: "src-1",
            subjectTypeID: locationType.id,
            label: "At home",
            description: ""
        )
        let unplaced = CatalogSubject(
            id: "s-bob",
            ref: "CPR-B",
            sourceID: "src-1",
            subjectTypeID: personType.id,
            label: "Bob",
            description: ""
        )

        let snapshot = SourceGraphSnapshot.build(
            sourceId: "src-1",
            subjects: [birth, alice, bridge, unplaced],
            positions: [
                CatalogSubjectPosition(subjectID: alice.id, gridX: 1, gridY: 2),
                CatalogSubjectPosition(subjectID: birth.id, gridX: 3, gridY: 4),
                CatalogSubjectPosition(subjectID: bridge.id, gridX: 5, gridY: 6),
            ],
            types: [personType, eventType, locationType],
            provisionalLinks: [
                EvidenceProvisionalLink(
                    bridgeSubjectID: bridge.id,
                    endpointAID: alice.id,
                    endpointBID: birth.id
                ),
            ]
        )

        #expect(snapshot.sourceId == "src-1")
        #expect(snapshot.subjects.map(\.id) == [alice.id, birth.id])
        #expect(snapshot.subjects[0].kind == .person)
        #expect(snapshot.subjects[0].gridX == 1)
        #expect(snapshot.subjects[0].isCited == false)
        #expect(snapshot.subjects[1].kind == .event)
        #expect(snapshot.bridges.count == 1)
        #expect(snapshot.bridges[0].id == bridge.id)
        #expect(snapshot.bridges[0].kind == .location)
        #expect(snapshot.bridges[0].endpointAID == alice.id)
        #expect(snapshot.bridges[0].endpointBID == birth.id)
    }

    @Test func attachingLinksFillsBridgeEndpoints() {
        let bridge = SourceGraphPlacedBridge(
            subject: CatalogSubject(
                id: "s-loc",
                ref: "CLO-L",
                sourceID: "src-1",
                subjectTypeID: locationType.id,
                label: "At home",
                description: ""
            ),
            kind: .location,
            typeLabel: "Location",
            gridX: 5,
            gridY: 6,
            endpointAID: nil,
            endpointBID: nil
        )
        let snapshot = SourceGraphSnapshot(sourceId: "src-1", bridges: [bridge])
        let next = snapshot.attaching(links: [
            EvidenceProvisionalLink(
                bridgeSubjectID: bridge.id,
                endpointAID: "a",
                endpointBID: "b"
            ),
        ])
        #expect(next.bridges[0].endpointAID == "a")
        #expect(next.bridges[0].endpointBID == "b")
    }

    @Test func bridgeKindInferenceTable() {
        #expect(EvidenceBridgeKindInference.kind(.person, .event) == .participation)
        #expect(EvidenceBridgeKindInference.kind(.event, .person) == .participation)
        #expect(EvidenceBridgeKindInference.kind(.person, .place) == .location)
        #expect(EvidenceBridgeKindInference.kind(.event, .place) == .location)
        #expect(EvidenceBridgeKindInference.kind(.person, .person) == .relationship)
        #expect(EvidenceBridgeKindInference.kind(.event, .event) == nil)
        #expect(EvidenceBridgeKindInference.kind(.place, .place) == nil)
    }

    @Test func updatingPositionChangesOnlyMatchingSubject() {
        let alice = CatalogSubject(
            id: "s-alice",
            ref: "CPR-A",
            sourceID: "src-1",
            subjectTypeID: personType.id,
            label: "Alice",
            description: ""
        )
        let snapshot = SourceGraphSnapshot(
            sourceId: "src-1",
            subjects: [
                SourceGraphPlacedSubject(
                    subject: alice,
                    kind: .person,
                    typeLabel: "Person",
                    gridX: 1,
                    gridY: 2,
                    isCited: false
                ),
            ]
        )
        let next = snapshot.updatingPosition(subjectID: alice.id, gridX: 9, gridY: 8)
        #expect(next.subjects[0].gridX == 9)
        #expect(next.subjects[0].gridY == 8)
    }

    @Test func accessibilityLabelIncludesUncited() {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: personType.id,
                label: "Wm Robins",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: false
        )
        let label = EvidenceSubjectCard.accessibilityLabel(for: placed)
        #expect(label.contains("Person"))
        #expect(label.contains("Wm Robins"))
        #expect(label.contains(String(localized: L10n.EvidenceGraph.uncitedAccessibility)))
    }

    @Test func accessibilityLabelIncludesCited() {
        let placed = SourceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: personType.id,
                label: "Wm Robins",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true
        )
        let label = EvidenceSubjectCard.accessibilityLabel(for: placed)
        #expect(label.contains(String(localized: L10n.EvidenceGraph.citedAccessibility)))
    }
}
