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
            observations: [observation],
            rules: CatalogConnectRule.productMatrix
        )
        #expect(snapshot.subjects.count == 1)
        #expect(snapshot.subjects[0].isCited)
        #expect(snapshot.subjects[0].observations.count == 1)
        #expect(snapshot.subjects[0].observations[0].valueText == "Boston")
    }

    @Test func buildSortsObservationsByPropertyLabel() {
        let alice = CatalogSubject(
            id: "s-alice",
            ref: "CPR-A",
            sourceID: "src-1",
            subjectTypeID: personType.id,
            label: "Alice",
            description: ""
        )
        let sex = CatalogObservation(
            id: "obs-sex",
            ref: "OBS-Z",
            citationID: "cit-1",
            subjectID: alice.id,
            propertyID: "prop-sex",
            polarity: "positive",
            valueText: "Female",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "sex_at_birth",
            propertyLabel: "Sex at Birth",
            propertyValueType: "term"
        )
        let name = CatalogObservation(
            id: "obs-name",
            ref: "OBS-A",
            citationID: "cit-1",
            subjectID: alice.id,
            propertyID: "prop-name",
            polarity: "positive",
            valueText: "Ada",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "name",
            propertyLabel: "Name",
            propertyValueType: "name"
        )
        let snapshot = SourceGraphSnapshot.build(
            sourceId: "src-1",
            subjects: [alice],
            positions: [CatalogSubjectPosition(subjectID: alice.id, gridX: 0, gridY: 0)],
            types: [personType],
            observations: [sex, name],
            rules: CatalogConnectRule.productMatrix
        )
        #expect(snapshot.subjects[0].observations.map(\.propertyLabel) == ["Name", "Sex at Birth"])
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
            observations: [
                CatalogObservation(
                    id: "obs-event",
                    ref: "OBS-E",
                    citationID: "cit-1",
                    subjectID: bridge.id,
                    propertyID: "prop-event",
                    polarity: "positive",
                    valueText: "Birth",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: birth.id,
                    valueTermID: "",
                    propertyKey: "event",
                    propertyLabel: "Event",
                    propertyValueType: "subject"
                ),
                CatalogObservation(
                    id: "obs-place",
                    ref: "OBS-P",
                    citationID: "cit-1",
                    subjectID: bridge.id,
                    propertyID: "prop-place",
                    polarity: "positive",
                    valueText: "Alice",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: alice.id,
                    valueTermID: "",
                    propertyKey: "place",
                    propertyLabel: "Place",
                    propertyValueType: "subject"
                ),
            ],
            rules: CatalogConnectRule.productMatrix
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
        #expect(snapshot.bridges[0].endpointAID == birth.id)
        #expect(snapshot.bridges[0].endpointBID == alice.id)
    }

    @Test func citedEndpointsComeFromRuleEdges() {
        let custom = CatalogConnectRule(
            fromTypeKey: "event",
            toTypeKey: "place",
            bridgeTypeKey: "location",
            edgePropertyKeys: ["place", "event"],
            disambiguation: "none",
            refuse: false,
            edges: [
                CatalogConnectEdge(propertyKey: "place", endpointTypeKey: "place"),
                CatalogConnectEdge(propertyKey: "event", endpointTypeKey: "event"),
            ]
        )
        let ends = SourceGraphSnapshot.citedEndpoints(
            kind: .location,
            observations: [
                CatalogObservation(
                    id: "obs-event",
                    ref: "OBS-E",
                    citationID: "cit-1",
                    subjectID: "s-loc",
                    propertyID: "prop-event",
                    polarity: ObservationPolarity.positive.rawValue,
                    valueText: "",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "s-birth",
                    valueTermID: "",
                    propertyKey: "event",
                    propertyLabel: "Event",
                    propertyValueType: PropertyValueType.subject.rawValue
                ),
                CatalogObservation(
                    id: "obs-place",
                    ref: "OBS-P",
                    citationID: "cit-1",
                    subjectID: "s-loc",
                    propertyID: "prop-place",
                    polarity: ObservationPolarity.positive.rawValue,
                    valueText: "",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    valueSubjectID: "s-alice",
                    valueTermID: "",
                    propertyKey: "place",
                    propertyLabel: "Place",
                    propertyValueType: PropertyValueType.subject.rawValue
                ),
            ],
            rules: [custom]
        )
        #expect(ends.a == "s-alice")
        #expect(ends.b == "s-birth")
    }

    @Test func legacyConnectKeysDecodeAndAreIgnored() throws {
        let json = """
        {
          "section":"sources",
          "sourceId":"s1",
          "connectFromSubjectId":"p1",
          "connectToSubjectId":"e1",
          "connectBridgeTypeKey":"participation",
          "connectDisambiguationTermId":"term-witness",
          "connectGridX":2,
          "connectGridY":3,
          "sourceSurface":"citationComposer"
        }
        """
        let decoded = try JSONDecoder().decode(WorkspaceLocation.self, from: Data(json.utf8))
        #expect(decoded.connectFromSubjectId == "p1")
        #expect(decoded.connectToSubjectId == "e1")
        #expect(decoded.connectBridgeTypeKey == "participation")
        #expect(decoded.isConnectPrefill)
        #expect(decoded.sourceSurface == .citationComposer)
    }

    @Test func connectRulesRefusePersonPlace() {
        let rules = CatalogConnectRule.productMatrix
        #expect(CatalogConnectRule.match(from: "person", to: "event", in: rules).bridgeTypeKey == "participation")
        #expect(CatalogConnectRule.match(from: "event", to: "person", in: rules).bridgeTypeKey == "participation")
        #expect(CatalogConnectRule.match(from: "person", to: "place", in: rules).refuse)
        #expect(CatalogConnectRule.match(from: "place", to: "person", in: rules).refuse)
        #expect(CatalogConnectRule.match(from: "event", to: "place", in: rules).bridgeTypeKey == "location")
        #expect(CatalogConnectRule.match(from: "person", to: "person", in: rules).bridgeTypeKey == "relationship")
        #expect(CatalogConnectRule.match(from: "event", to: "event", in: rules).refuse)
        #expect(CatalogConnectRule.match(from: "place", to: "place", in: rules).refuse)
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
        let rows = SourceGraphRows(
            sourceId: "src-1",
            subjects: [alice],
            positions: [CatalogSubjectPosition(subjectID: alice.id, gridX: 1, gridY: 2)]
        )
        let next = rows.updatingPosition(subjectID: alice.id, gridX: 9, gridY: 8)
        #expect(next.positions[0].gridX == 9)
        #expect(next.positions[0].gridY == 8)
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
