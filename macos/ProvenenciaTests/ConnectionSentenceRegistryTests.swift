import Foundation
import Testing
@testable import Provenencia

@Suite
struct ConnectionSentenceRegistryTests {
    private func participation(
        termKey: String?,
        label: String?,
        eventType: String?,
        event: String
    ) -> String {
        ConnectionSentenceRegistry.sentence(
            kind: .participation,
            endpointA: "Don Smith",
            endpointB: event,
            term: label,
            termKey: termKey,
            eventType: eventType
        )
    }

    @Test func eachEventRowReadsPlainly() {
        let cases: [(String, String, String)] = [
            ("birth", "Birth of Don Smith", "Don Smith was born"),
            ("death", "Death of Don Smith", "Don Smith died"),
            ("baptism", "Baptism of Don Smith", "Don Smith was baptized"),
            (
                "marriage",
                "Marriage of Don Smith and Marion Brown",
                "Don Smith married (Marriage of Don Smith and Marion Brown)"
            ),
            ("funeral", "Funeral of Don Smith", "Don Smith's funeral (Funeral of Don Smith)"),
            ("burial", "Burial of Don Smith", "Don Smith was buried"),
        ]
        for (eventType, event, expected) in cases {
            #expect(
                participation(termKey: SeededRoleTerm.subject, label: "Subject", eventType: eventType, event: event)
                    == expected
            )
        }
    }

    @Test func anotherRoleKeepsTheGenericSentence() {
        let sentence = ConnectionSentenceRegistry.sentence(
            kind: .participation,
            endpointA: "John Smith",
            endpointB: "Baptism of Don Smith",
            term: "Witness",
            termKey: "witness",
            eventType: "baptism"
        )
        #expect(sentence == "John Smith participated as witness at Baptism of Don Smith")
    }

    @Test func aTermReadsLowercaseInsideEverySentenceAndPhrase() {
        func whole(_ kind: EvidenceBridgeKind, _ a: String, _ b: String, _ term: String) -> String {
            ConnectionSentenceRegistry.sentence(kind: kind, endpointA: a, endpointB: b, term: term)
        }
        #expect(whole(.relationship, "John Smith", "Mary Smith", "Spouse") == "John Smith is the spouse of Mary Smith")
        #expect(whole(.placeRelationship, "Leeds", "Yorkshire", "Part of") == "Leeds is part of Yorkshire")
        #expect(
            whole(.participation, "Mary Smith", "Baptism of John Smith", "Witness")
                == "Mary Smith participated as witness at Baptism of John Smith"
        )
        #expect(ConnectionSentenceRegistry.phrase(kind: .relationship, term: "Spouse") == "Is the spouse of")
        #expect(ConnectionSentenceRegistry.phrase(kind: .participation, term: "Witness") == "Participated as witness")
        #expect(ConnectionSentenceRegistry.reasonRole(kind: .relationship, term: "Spouse") == "spouse")
    }

    @Test func lookupPrefersTheMostSpecificRow() {
        let specific = ConnectionSentenceRegistry.lookup(
            kind: .participation,
            term: SeededRoleTerm.subject,
            eventType: SeededEventTypeTerm.birth
        )
        #expect(specific?.term == SeededRoleTerm.subject && specific?.eventType == SeededEventTypeTerm.birth)
        #expect(
            ConnectionSentenceRegistry.lookup(kind: .participation, term: SeededRoleTerm.subject, eventType: "census")
                == nil
        )
        #expect(
            ConnectionSentenceRegistry.lookup(kind: .participation, term: nil, eventType: SeededEventTypeTerm.birth)
                == nil
        )
        #expect(
            ConnectionSentenceRegistry.lookup(kind: .relationship, term: SeededRoleTerm.subject, eventType: "birth")
                == nil
        )
    }

    @Test func aRoleOnlyRowBeatsTheDefault() {
        let rows = [
            ConnectionSentenceRegistry.Sentence(
                kind: .participation, term: "witness", eventType: nil, sentence: { _ in "role-only" }
            ),
            ConnectionSentenceRegistry.Sentence(
                kind: .participation, term: "witness", eventType: "baptism", sentence: { _ in "role-and-event" }
            ),
        ]
        func read(_ eventType: String?) -> String {
            ConnectionSentenceRegistry.sentence(
                kind: .participation,
                endpointA: "A",
                endpointB: "B",
                term: "Witness",
                termKey: "witness",
                eventType: eventType,
                in: rows
            )
        }
        #expect(read("baptism") == "role-and-event")
        #expect(read("census") == "role-only")
        #expect(read(nil) == "role-only")
    }

    @Test func aSpecificRowsPhraseWins() {
        let rows = [
            ConnectionSentenceRegistry.Sentence(
                kind: .participation,
                term: "witness",
                eventType: nil,
                sentence: { _ in "x" },
                phrase: { _ in "witnessed it" }
            ),
        ]
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .participation, term: "Witness", termKey: "witness", in: rows)
                == "witnessed it"
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .participation, term: "Informant", termKey: "informant", in: rows)
                == L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: "informant")
        )
    }

    @Test func aKeyWithoutALabelStillReadsPlainly() {
        #expect(
            participation(
                termKey: SeededRoleTerm.subject,
                label: nil,
                eventType: SeededEventTypeTerm.birth,
                event: "Birth of Don Smith"
            ) == "Don Smith was born"
        )
    }

    @Test func rowsAreUnique() {
        var seen = Set<String>()
        for row in ConnectionSentenceRegistry.sentences {
            let id = "\(row.kind)|\(row.term)|\(row.eventType ?? "-")"
            #expect(seen.insert(id).inserted, "duplicate row \(id)")
        }
        var kinds = Set<EvidenceBridgeKind>()
        for entry in ConnectionSentenceRegistry.entries {
            #expect(kinds.insert(entry.kind).inserted)
        }
    }

    @Test func everyKindHasAnEntry() {
        for kind in EvidenceBridgeKind.allCases {
            #expect(ConnectionSentenceRegistry.entries.filter { $0.kind == kind }.count == 1)
            #expect(ConnectionSentenceRegistry.entry(kind).kind == kind)
        }
    }

    @Test func eachKindEntryKeepsTodaysWording() {
        func whole(
            _ kind: EvidenceBridgeKind,
            _ a: String?,
            _ b: String?,
            _ term: String?,
            _ termKey: String?,
            _ eventType: String?
        ) -> String {
            ConnectionSentenceRegistry.sentence(
                kind: kind,
                endpointA: a,
                endpointB: b,
                term: term,
                termKey: termKey,
                eventType: eventType
            )
        }
        #expect(
            whole(.relationship, "Wm", "John", "father", nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryRelationship(person: "Wm", type: "father", related: "John")
        )
        #expect(
            whole(.relationship, "Wm", "John", nil, nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryRelationshipFallback(person: "Wm", related: "John")
        )
        #expect(
            whole(.placeRelationship, "A", "B", "within", nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryPlaceRelationship(from: "A", type: "within", to: "B")
        )
        #expect(
            whole(.placeRelationship, "A", "B", nil, nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryPlaceRelationshipFallback(from: "A", to: "B")
        )
        #expect(
            whole(.location, "Birth", "Leeds", nil, nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryLocation(event: "Birth", place: "Leeds")
        )
        #expect(
            whole(.participation, "Ada", "Birth", nil, nil, nil)
                == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(person: "Ada", event: "Birth")
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .relationship, term: "father")
                == L10n.EvidenceGraph.bridgeSummaryRelationshipTypeOnly(type: "father")
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .relationship, term: nil)
                == L10n.string(L10n.EvidenceGraph.bridgeSummaryRelationshipBare)
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .placeRelationship, term: "within")
                == L10n.EvidenceGraph.bridgeSummaryPlaceRelationshipTypeOnly(type: "within")
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .placeRelationship, term: nil)
                == L10n.string(L10n.EvidenceGraph.bridgeSummaryPlaceRelationshipBare)
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .participation, term: "Witness")
                == L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: "witness")
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .participation, term: nil)
                == L10n.string(L10n.EvidenceGraph.bridgeSummaryParticipationBare)
        )
        #expect(
            ConnectionSentenceRegistry.phrase(kind: .location, term: "x")
                == L10n.string(L10n.EvidenceGraph.bridgeSummaryLocationBare)
        )
    }
}
