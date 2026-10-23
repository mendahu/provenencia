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

    // MARK: - What a row shares with a connected record

    private func share(
        _ kind: EvidenceBridgeKind,
        neighborIsA: Bool,
        termKey: String? = nil,
        term: String? = nil,
        eventTypeKey: String? = nil,
        eventType: String? = nil,
        name: String = "Mary Smith",
        similar: Bool = false
    ) -> String {
        ConnectionSentenceRegistry.share(
            ConnectionSentenceRegistry.ShareInput(
                kind: kind, neighborIsA: neighborIsA, termKey: termKey, term: term,
                eventTypeKey: eventTypeKey, eventType: eventType, neighborName: name
            ),
            similar: similar
        )
    }

    @Test func aNeighborEventIsNamedByItsType() {
        #expect(share(.participation, neighborIsA: false, eventTypeKey: "birth", eventType: "Birth") == "a birth event")
        #expect(share(.participation, neighborIsA: false, eventTypeKey: "birth", eventType: "Birth", similar: true) == "a similar birth event")
        #expect(share(.participation, neighborIsA: false) == "an event")
        #expect(share(.participation, neighborIsA: false, similar: true) == "a similar event")
    }

    @Test func aVowelInitialEventTypeTakesAn() {
        #expect(share(.participation, neighborIsA: false, eventTypeKey: "engagement", eventType: "Engagement") == "an engagement event")
        #expect(share(.participation, neighborIsA: false, eventTypeKey: "engagement", eventType: "Engagement", similar: true) == "a similar engagement event")
    }

    private static let eventTypes = [
        SeededEventTypeTerm.birth, SeededEventTypeTerm.death, SeededEventTypeTerm.marriage,
        SeededEventTypeTerm.engagement, SeededEventTypeTerm.baptism, SeededEventTypeTerm.burial,
        SeededEventTypeTerm.funeral, SeededEventTypeTerm.census, SeededEventTypeTerm.residence,
        SeededEventTypeTerm.migration, SeededEventTypeTerm.travel,
    ]
    private static let roles = [
        SeededRoleTerm.subject, SeededRoleTerm.father, SeededRoleTerm.mother, SeededRoleTerm.spouse,
        SeededRoleTerm.child, SeededRoleTerm.witness, SeededRoleTerm.informant, SeededRoleTerm.pallbearer,
    ]
    private static let relationships = [
        SeededRelationshipTypeTerm.spouse, SeededRelationshipTypeTerm.sibling, SeededRelationshipTypeTerm.cousin,
        SeededRelationshipTypeTerm.parent, SeededRelationshipTypeTerm.child, SeededRelationshipTypeTerm.grandparent,
        SeededRelationshipTypeTerm.grandchild, SeededRelationshipTypeTerm.pibling, SeededRelationshipTypeTerm.nibling,
        SeededRelationshipTypeTerm.guardian, SeededRelationshipTypeTerm.ward,
    ]

    @Test func everySeededTermHasItsOwnSharePhrase() {
        for key in Self.eventTypes {
            let phrase = share(.participation, neighborIsA: false, eventTypeKey: key, eventType: key)
            #expect(phrase.hasSuffix(" event") && !phrase.hasPrefix("a connection"), "\(key): \(phrase)")
        }
        for key in Self.roles {
            let phrase = share(.participation, neighborIsA: true, termKey: key, term: key, name: "X")
            #expect(phrase.hasSuffix(", X") && !phrase.hasPrefix("a connection"), "\(key): \(phrase)")
        }
        for key in Self.relationships {
            let phrase = share(.relationship, neighborIsA: true, termKey: key, term: key)
            #expect(!phrase.hasPrefix("a connection"), "\(key): \(phrase)")
        }
    }

    @Test func aNeighborPersonIsNamedByRoleAndName() {
        #expect(share(.participation, neighborIsA: true, termKey: "subject", term: "Subject") == "a subject, Mary Smith")
        #expect(share(.participation, neighborIsA: true, termKey: "subject", term: "Subject", similar: true) == "a similar subject, Mary Smith")
        #expect(share(.participation, neighborIsA: true, termKey: "informant", term: "Informant", name: "John Smith") == "an informant, John Smith")
        // No role to name: nothing to word.
        #expect(share(.participation, neighborIsA: true) == "a connection")
    }

    @Test func aRelationshipIsTheNeighborsPartRelativeToTheRow() {
        // The A end is the term itself; the B end is its inverse.
        #expect(share(.relationship, neighborIsA: true, termKey: "spouse", term: "Spouse") == "a spouse")
        #expect(share(.relationship, neighborIsA: true, termKey: "spouse", term: "Spouse", similar: true) == "a similar spouse")
        #expect(share(.relationship, neighborIsA: true, termKey: "parent", term: "Parent") == "a parent")
        #expect(share(.relationship, neighborIsA: false, termKey: "parent", term: "Parent") == "a child")
        #expect(share(.relationship, neighborIsA: false, termKey: "grandparent", term: "Grandparent") == "a grandchild")
        // Spouse, sibling and cousin are their own inverse.
        for key in ["spouse", "sibling", "cousin"] {
            #expect(share(.relationship, neighborIsA: false, termKey: key, term: key.capitalized) == "a \(key)")
        }
        // The seeded label "Aunt / uncle" reads "aunt or uncle" in a phrase, with its own article.
        #expect(share(.relationship, neighborIsA: true, termKey: "pibling", term: "Aunt / uncle") == "an aunt or uncle")
        #expect(share(.relationship, neighborIsA: true, termKey: "pibling", term: "Aunt / uncle", similar: true) == "a similar aunt or uncle")
        #expect(share(.relationship, neighborIsA: false, termKey: "nibling", term: "Niece / nephew") == "an aunt or uncle")
    }

    @Test func aResearchersTermReadsAsAConnectionNamingIt() {
        #expect(share(.relationship, neighborIsA: true, term: "Godparent") == "a connection (Godparent)")
        #expect(share(.relationship, neighborIsA: true, term: "Godparent", similar: true) == "a similar connection (Godparent)")
        #expect(share(.participation, neighborIsA: true, term: "Mentor") == "a connection (Mentor)")
        #expect(share(.participation, neighborIsA: false, eventType: "Cremation") == "a connection (Cremation)")
        // Read from the B end a researcher's term has no inverse to name.
        #expect(share(.relationship, neighborIsA: false, term: "Godparent") == "a connection")
    }

    @Test func aPlaceRelationshipIsParentWithinSuccessorOrPredecessor() {
        #expect(share(.placeRelationship, neighborIsA: false, termKey: "part_of") == "a parent place")
        #expect(share(.placeRelationship, neighborIsA: false, termKey: "part_of", similar: true) == "a similar parent place")
        #expect(share(.placeRelationship, neighborIsA: true, termKey: "part_of") == "a place within it")
        #expect(share(.placeRelationship, neighborIsA: false, termKey: "succeeded_by") == "a successor place")
        #expect(share(.placeRelationship, neighborIsA: true, termKey: "succeeded_by") == "a predecessor place")
        #expect(share(.placeRelationship, neighborIsA: true, termKey: "succeeded_by", similar: true) == "a similar predecessor place")
        #expect(share(.placeRelationship, neighborIsA: true) == "a connection")
    }

    @Test func aLocationNamesThePlace() {
        #expect(share(.location, neighborIsA: false, name: "Leeds") == "a place, Leeds")
        #expect(share(.location, neighborIsA: false, name: "Leeds", similar: true) == "a similar place, Leeds")
        #expect(share(.location, neighborIsA: true, name: "Leeds") == "a connection")
    }

    @Test func nothingNamedIsAConnection() {
        #expect(ConnectionSentenceRegistry.share(nil, similar: false) == "a connection")
        #expect(ConnectionSentenceRegistry.share(nil, similar: true) == "a similar connection")
    }
}
