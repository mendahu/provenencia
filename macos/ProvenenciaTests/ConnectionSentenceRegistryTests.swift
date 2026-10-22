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
                participation(termKey: "subject", label: "Subject", eventType: eventType, event: event)
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
        #expect(sentence == "John Smith participated as Witness at Baptism of Don Smith")
    }

    @Test func lookupPrefersTheMostSpecificRow() {
        let specific = ConnectionSentenceRegistry.lookup(kind: .participation, term: "subject", eventType: "birth")
        #expect(specific.term == "subject" && specific.eventType == "birth")

        let unknownEvent = ConnectionSentenceRegistry.lookup(kind: .participation, term: "subject", eventType: "census")
        #expect(unknownEvent.term == nil && unknownEvent.eventType == nil)

        let noTerm = ConnectionSentenceRegistry.lookup(kind: .participation, term: nil, eventType: "birth")
        #expect(noTerm.term == nil && noTerm.eventType == nil)

        let otherKind = ConnectionSentenceRegistry.lookup(kind: .relationship, term: "subject", eventType: "birth")
        #expect(otherKind.kind == .relationship && otherKind.term == nil)
    }

    @Test func everyKindHasADefaultRow() {
        for kind in EvidenceBridgeKind.allCases {
            let defaults = ConnectionSentenceRegistry.rows.filter {
                $0.kind == kind && $0.term == nil && $0.eventType == nil
            }
            #expect(defaults.count == 1)
            let row = ConnectionSentenceRegistry.defaultRow(kind)
            #expect(row.phrase != nil)
            #expect(ConnectionSentenceRegistry.lookup(kind: kind, term: nil, eventType: nil).kind == kind)
        }
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
                == L10n.EvidenceGraph.bridgeSummaryParticipationRoleOnly(role: "Witness")
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
