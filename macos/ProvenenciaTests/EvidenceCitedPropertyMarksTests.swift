import Foundation
import Testing
@testable import Provenencia

@Suite
struct EvidenceCitedPropertyMarksTests {
    private func observation(
        id: String,
        key: String,
        value: String,
        polarity: ObservationPolarity = .positive,
        valueType: PropertyValueType = .text
    ) -> CatalogObservation {
        CatalogObservation(
            id: id,
            ref: "OBS-\(id)",
            citationID: "cit-1",
            subjectID: "s1",
            propertyID: "prop-\(key)",
            polarity: polarity.rawValue,
            valueText: value,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: key,
            propertyLabel: key,
            propertyValueType: valueType.rawValue
        )
    }

    @Test func twoLanguagesSpokenAreNotConflicted() {
        let rows = [
            observation(id: "a", key: "languages_spoken", value: "English", valueType: .term),
            observation(id: "b", key: "languages_spoken", value: "French", valueType: .term),
        ]
        let map = ["languages_spoken": PropertyCardinality.multiple]
        let counts = EvidenceCitedPropertyMarks.conflictCounts(in: rows, cardinalityByKey: map)
        #expect(counts["languages_spoken"] == nil)
        #expect(!EvidenceCitedPropertyMarks.cardHasConflict(in: rows, cardinalityByKey: map))
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows, cardinalityByKey: map).isEmpty)
        #expect(!EvidenceCitedPropertyMarks.isConflicted(rows[0], counts: counts))
        #expect(!EvidenceCitedPropertyMarks.isConflicted(rows[1], counts: counts))
    }

    @Test func twoLikenessesAreNotConflicted() {
        let rows = [
            observation(id: "a", key: "likeness", value: "img-a", valueType: .depiction),
            observation(id: "b", key: "likeness", value: "img-b", valueType: .depiction),
        ]
        let map = ["likeness": PropertyCardinality.multiple]
        #expect(!EvidenceCitedPropertyMarks.cardHasConflict(in: rows, cardinalityByKey: map))
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows, cardinalityByKey: map).isEmpty)
    }

    @Test func graphHasConflictIgnoresMultipleCardinality() {
        let subject = EvidenceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [
                observation(id: "a", key: "languages_spoken", value: "English", valueType: .term),
                observation(id: "b", key: "languages_spoken", value: "French", valueType: .term),
            ]
        )
        let multipleMap = ["languages_spoken": PropertyCardinality.multiple]
        #expect(!EvidenceCitedPropertyMarks.graphHasConflict(subjects: [subject], bridges: [], cardinalityByKey: multipleMap))

        let conflicted = EvidenceGraphPlacedSubject(
            subject: subject.subject,
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [
                observation(id: "a", key: "name", value: "Wm Robins", valueType: .name),
                observation(id: "b", key: "name", value: "William Robins", valueType: .name),
            ]
        )
        #expect(EvidenceCitedPropertyMarks.graphHasConflict(subjects: [conflicted], bridges: []))
        #expect(EvidenceCitedPropertyMarks.graphHasConflict(
            subjects: [conflicted],
            bridges: [],
            cardinalityByKey: ["name": PropertyCardinality.single]
        ))
    }

    @Test func mixedMultipleAndSingleOnlyMarksSingle() {
        let rows = [
            observation(id: "a", key: "name", value: "Wm Robins", valueType: .name),
            observation(id: "b", key: "name", value: "William Robins", valueType: .name),
            observation(id: "c", key: "languages_spoken", value: "English", valueType: .term),
            observation(id: "d", key: "languages_spoken", value: "French", valueType: .term),
        ]
        let map = [
            "name": PropertyCardinality.single,
            "languages_spoken": PropertyCardinality.multiple,
        ]
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows, cardinalityByKey: map) == [0..<2])
    }

    @Test func missingCardinalityStillConflicts() {
        let rows = [
            observation(id: "a", key: "name", value: "Wm Robins", valueType: .name),
            observation(id: "b", key: "name", value: "William Robins", valueType: .name),
        ]
        #expect(EvidenceCitedPropertyMarks.cardHasConflict(in: rows, cardinalityByKey: [:]))
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows, cardinalityByKey: [:]) == [0..<2])
    }

    @Test func accessibilityOmitsConflictForMultipleCounts() {
        let rows = [
            observation(id: "a", key: "languages_spoken", value: "English", valueType: .term),
            observation(id: "b", key: "languages_spoken", value: "French", valueType: .term),
        ]
        let multipleCounts = EvidenceCitedPropertyMarks.conflictCounts(
            in: rows,
            cardinalityByKey: ["languages_spoken": PropertyCardinality.multiple]
        )
        let multipleLabel = EvidenceCitedPropertyMarks.accessibilityLabel(
            propertyLabel: "Languages spoken",
            spokenValue: "English",
            conflictCount: multipleCounts["languages_spoken"],
            isNegated: false
        )
        #expect(!multipleLabel.contains(L10n.EvidenceGraph.conflictOneOf(count: 2)))

        let nameRows = [
            observation(id: "a", key: "name", value: "Wm Robins", valueType: .name),
            observation(id: "b", key: "name", value: "William Robins", valueType: .name),
        ]
        let nameCounts = EvidenceCitedPropertyMarks.conflictCounts(in: nameRows, cardinalityByKey: ["name": .single])
        let nameLabel = EvidenceCitedPropertyMarks.accessibilityLabel(
            propertyLabel: "Name",
            spokenValue: "Wm Robins",
            conflictCount: nameCounts["name"],
            isNegated: false
        )
        #expect(nameLabel.contains(L10n.EvidenceGraph.conflictOneOf(count: 2)))
    }

    @Test func citedRowInsetAlwaysReservesGutter() {
        #expect(EvidenceCitedPropertyMarks.leadingInset(shell: 13) == 24)
        #expect(EvidenceCitedPropertyMarks.leadingInset(shell: 11) == 22)
    }

    @Test func twoNamesAreBothConflicted() {
        let rows = [
            observation(id: "a", key: "name", value: "Wm Robins"),
            observation(id: "b", key: "name", value: "William Robins"),
            observation(id: "c", key: "occupation", value: "Farmer"),
        ]
        let counts = EvidenceCitedPropertyMarks.conflictCounts(in: rows)
        #expect(EvidenceCitedPropertyMarks.isConflicted(rows[0], counts: counts))
        #expect(EvidenceCitedPropertyMarks.isConflicted(rows[1], counts: counts))
        #expect(!EvidenceCitedPropertyMarks.isConflicted(rows[2], counts: counts))
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows) == [0..<2])
    }

    @Test func sameStringStillConflicts() {
        let rows = [
            observation(id: "a", key: "name", value: "Wm Robins"),
            observation(id: "b", key: "name", value: "Wm Robins"),
        ]
        #expect(EvidenceCitedPropertyMarks.cardHasConflict(in: rows))
        #expect(EvidenceCitedPropertyMarks.conflictCounts(in: rows)["name"] == 2)
    }

    @Test func singletonIsUnmarked() {
        let rows = [observation(id: "a", key: "name", value: "Wm Robins")]
        #expect(!EvidenceCitedPropertyMarks.cardHasConflict(in: rows))
        #expect(!EvidenceCitedPropertyMarks.isNegated(rows[0]))
        #expect(EvidenceCitedPropertyMarks.conflictRuns(in: rows).isEmpty)
    }

    @Test func negativeSingletonIsNegatedOnly() {
        let rows = [
            observation(id: "a", key: "birth_place", value: "Ireland", polarity: .negative),
        ]
        #expect(EvidenceCitedPropertyMarks.isNegated(rows[0]))
        #expect(!EvidenceCitedPropertyMarks.cardHasConflict(in: rows))
    }

    @Test func negativeAndPositiveSameKeyStack() {
        let rows = [
            observation(id: "a", key: "birth_place", value: "Ireland", polarity: .negative),
            observation(id: "b", key: "birth_place", value: "Canada"),
        ]
        let counts = EvidenceCitedPropertyMarks.conflictCounts(in: rows)
        #expect(EvidenceCitedPropertyMarks.isConflicted(rows[0], counts: counts))
        #expect(EvidenceCitedPropertyMarks.isConflicted(rows[1], counts: counts))
        #expect(EvidenceCitedPropertyMarks.isNegated(rows[0]))
        #expect(!EvidenceCitedPropertyMarks.isNegated(rows[1]))
    }

    @Test func accessibilityIncludesConflictAndNegated() {
        let conflict = EvidenceCitedPropertyMarks.accessibilityLabel(
            propertyLabel: "Name",
            spokenValue: "Wm Robins",
            conflictCount: 2,
            isNegated: false
        )
        #expect(conflict.contains("Name"))
        #expect(conflict.contains("Wm Robins"))
        #expect(conflict.contains(L10n.EvidenceGraph.conflictOneOf(count: 2)))
        #expect(conflict.contains(L10n.string(L10n.EvidenceGraph.citedRowEditHint)))

        let negated = EvidenceCitedPropertyMarks.accessibilityLabel(
            propertyLabel: "Birth place",
            spokenValue: "Not Ireland",
            conflictCount: 2,
            isNegated: true
        )
        #expect(negated.contains("Not Ireland"))
        #expect(negated.contains(L10n.string(L10n.EvidenceGraph.negatedAccessibility)))
    }

    @Test func extraRowsOmitSentenceKeys() {
        let rows = [
            observation(id: "person", key: "person", value: "p1"),
            observation(id: "related", key: "related_to", value: "p2"),
            observation(id: "type", key: "relationship_type", value: "father"),
            observation(id: "note", key: "remark", value: "Named in the household"),
        ]
        let extras = EvidenceCitedPropertyMarks.extraObservations(in: rows)
        #expect(extras.map(\.propertyKey) == ["remark"])
        #expect(EvidenceCitedPropertyMarks.sentencePropertyKeys().isSuperset(of: [
            "person", "event", "place", "related_to", "role", "relationship_type",
        ]))
    }

    @Test func graphHasConflictIgnoresNegationOnly() {
        let subject = EvidenceGraphPlacedSubject(
            subject: CatalogSubject(
                id: "s1",
                ref: "CPR-1",
                sourceID: "src",
                subjectTypeID: "t",
                label: "A",
                description: ""
            ),
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [observation(id: "a", key: "occupation", value: "Farmer")]
        )
        #expect(!EvidenceCitedPropertyMarks.graphHasConflict(subjects: [subject], bridges: []))

        let denied = EvidenceGraphPlacedSubject(
            subject: subject.subject,
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [observation(id: "a", key: "occupation", value: "Farmer", polarity: .negative)]
        )
        #expect(!EvidenceCitedPropertyMarks.graphHasConflict(subjects: [denied], bridges: []))

        let conflicted = EvidenceGraphPlacedSubject(
            subject: subject.subject,
            kind: .person,
            typeLabel: "Person",
            gridX: 0,
            gridY: 0,
            isCited: true,
            observations: [
                observation(id: "a", key: "name", value: "Wm Robins"),
                observation(id: "b", key: "name", value: "William Robins"),
            ]
        )
        #expect(EvidenceCitedPropertyMarks.graphHasConflict(subjects: [conflicted], bridges: []))
    }
}
