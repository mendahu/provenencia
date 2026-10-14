import Testing
@testable import Provenencia

@Suite
struct PVGlyphTests {
    @Test func subjectTypeKeysMapToSubjectMarks() {
        let expected: [(String, PVMarkKey)] = [
            ("person", .subjectPerson),
            ("event", .subjectEvent),
            ("place", .subjectPlace),
            ("relationship", .subjectRelationship),
            ("participation", .subjectParticipation),
            ("location", .subjectLocation),
            ("place_relationship", .subjectPlaceRelationship),
            ("source", .subjectSource),
        ]
        for (key, mark) in expected {
            #expect(PVMarkKey(subjectTypeKey: key) == mark)
        }
        for key in ["", "source_type", "widget"] {
            #expect(PVMarkKey(subjectTypeKey: key) == nil)
        }
    }

    @Test func evidenceKindsAgreeWithTheSubjectRule() {
        for kind in EvidencePrimaryKind.allCases {
            #expect(kind.markKey == PVMarkKey(subjectTypeKey: kind.rawValue))
        }
        for kind in EvidenceBridgeKind.allCases {
            #expect(kind.markKey == PVMarkKey(subjectTypeKey: kind.rawValue))
        }
    }
}
