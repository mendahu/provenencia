import Foundation
import Testing
@testable import Provenencia

/// The evidence sheet's choices and words, tested without a view.
@Suite
struct PromoteEvidenceDisplayTests {

    // MARK: - Fixtures

    private func pin(_ id: String, source: String = "src-1", value: String = "John Smith") -> CatalogClaimEvidencePin {
        CatalogClaimEvidencePin(
            id: id, memberSubjectID: "m-\(id)", observationID: "o-\(id)", sourceID: source,
            value: .name(CatalogNameValue(form: value))
        )
    }

    private func line(
        _ key: String = "name",
        outcome: EvidenceOutcome = .agree,
        relation: ValueRelation = .same,
        pins: [String] = ["p1"]
    ) -> CatalogClaimEvidenceLine {
        CatalogClaimEvidenceLine(
            propertyKey: key, outcome: outcome, relation: relation, pinnable: outcome != .conflict, checked: true,
            incomingObservationID: "in-1", incoming: .name(CatalogNameValue(form: "John Smith")),
            members: pins.map { pin($0) }
        )
    }

    private func relationship(
        neighbor: String = "n1",
        type: String = "participation",
        sections: [CatalogClaimEvidenceSection] = []
    ) -> CatalogClaimEvidenceRelationship {
        CatalogClaimEvidenceRelationship(
            neighborSubjectID: neighbor, key: "k", bridgeType: type, outcome: .agree, pinnable: true, checked: true,
            bundles: [CatalogClaimEvidenceBundle(id: "b1", memberSubjectID: "m", sourceID: "s")], sections: sections
        )
    }

    private func date(_ year: Int32, _ month: Int32? = nil, _ day: Int32? = nil) -> CatalogDateValueInput {
        CatalogDateValueInput(kind: "point", startYear: year, startMonth: month, startDay: day)
    }

    private func bridge(_ id: String, _ kind: EvidenceBridgeKind, _ a: String, _ b: String) -> PromoteFlow.BridgeFact {
        PromoteFlow.BridgeFact(id: id, sentence: "", endA: a, endB: b, kind: kind)
    }

    // MARK: - T-1

    @Test func checkStateFollowsThePins() {
        let ids = ["p1", "p2", "p3", "p4", "p5"]
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: Set(ids)) == .checked)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: []) == .unchecked)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: ["p1", "p3", "p4", "p5"]) == .partial)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: [], pins: ["p1"]) == .unchecked)
    }

    @Test func checkedCountIsOfferedPinsOnly() {
        let evidence = CatalogClaimEvidence(offers: [
            CatalogClaimEvidenceOffer(id: "p1", checked: true),
            CatalogClaimEvidenceOffer(id: "p2", checked: false),
        ])
        #expect(PromoteEvidenceDisplay.checkedCount(evidence, pins: ["p1", "x"]) == 1)
        #expect(PromoteEvidenceDisplay.checkedCount(evidence, pins: ["p1", "p2"]) == 2)
        #expect(PromoteEvidenceDisplay.checkedCount(evidence, pins: []) == 0)
    }

    @Test func aConflictAnywhereShowsTheCallout() {
        let conflict = CatalogClaimEvidenceSection(propertyKey: "date", lines: [line("date", outcome: .conflict, relation: .different)])
        let fine = CatalogClaimEvidenceSection(propertyKey: "name", lines: [line()])
        let similar = CatalogClaimEvidenceSection(propertyKey: "name", lines: [line(outcome: .partial, relation: .contains)])
        #expect(!PromoteEvidenceDisplay.hasConflict(CatalogClaimEvidence(sections: [fine])))
        #expect(PromoteEvidenceDisplay.hasConflict(CatalogClaimEvidence(sections: [fine, conflict])))
        // Only inside a relationship's sections.
        #expect(PromoteEvidenceDisplay.hasConflict(CatalogClaimEvidence(
            sections: [fine], relationships: [relationship(sections: [conflict])]
        )))
        // A similar relationship with no conflicting line is not a disagreement.
        var partial = relationship(sections: [similar])
        partial.outcome = .partial
        #expect(!PromoteEvidenceDisplay.hasConflict(CatalogClaimEvidence(sections: [fine], relationships: [partial])))
    }

    @Test func bridgeForARelationship() {
        let participation = bridge("b-p", .participation, "john", "birth")
        let spouse = bridge("b-r", .relationship, "birth", "john")
        let other = bridge("b-o", .participation, "mary", "birth")
        let bridges = [other, spouse, participation]
        let found = PromoteEvidenceDisplay.bridge(for: relationship(neighbor: "birth", type: "participation"), row: "john", in: bridges)
        #expect(found?.id == "b-p")
        // Ends in either order; the first of the kind in page order.
        let twice = [bridge("first", .participation, "birth", "john")] + bridges
        #expect(PromoteEvidenceDisplay.bridge(for: relationship(neighbor: "birth", type: "participation"), row: "john", in: twice)?.id == "first")
        #expect(PromoteEvidenceDisplay.bridge(for: relationship(neighbor: "birth", type: "relationship"), row: "john", in: bridges)?.id == "b-r")
        #expect(PromoteEvidenceDisplay.bridge(for: relationship(neighbor: "death", type: "participation"), row: "john", in: bridges) == nil)
        #expect(PromoteEvidenceDisplay.bridge(for: relationship(neighbor: "birth", type: "nonsense"), row: "john", in: bridges) == nil)
    }

    // MARK: - T-2

    private func sentence(
        _ relation: ValueRelation,
        cause: ValueCause = .none,
        propertyCause: EvidencePropertyCause = .none,
        _ value: CatalogConclusionValue
    ) -> String {
        L10n.Promote.lineSentence(relation: relation, cause: cause, propertyCause: propertyCause, value: value).key
    }

    @Test func lineSentencesFollowTheTable() {
        let name = CatalogConclusionValue.name(CatalogNameValue(form: "John Smith"))
        let day = CatalogConclusionValue.date(date(1817, 5, 14))
        let term = CatalogConclusionValue.term(id: "t", key: "male", label: "Male")
        let likeness = CatalogConclusionValue.depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: "objects/a.png"))
        let key = { (r: LocalizedStringResource) in r.key }

        // same, any cause: the value's type picks the wording.
        #expect(sentence(.same, cause: .exact, name) == key(L10n.Promote.lineSameName))
        #expect(sentence(.same, cause: .letterCase, name) == key(L10n.Promote.lineSameName))
        #expect(sentence(.same, day) == key(L10n.Promote.lineSameDate))
        #expect(sentence(.same, term) == key(L10n.Promote.lineSame))
        #expect(sentence(.same, likeness) == key(L10n.Promote.lineSameLikeness))
        for value in [name, day, term] {
            #expect(sentence(.equivalent, cause: .fold, value) == key(L10n.Promote.lineEquivalentFold))
            #expect(sentence(.equivalent, cause: .alias, value) == key(L10n.Promote.lineEquivalentAlias))
            #expect(sentence(.equivalent, cause: .abbreviation, value) == key(L10n.Promote.lineEquivalentAbbreviation))
            #expect(sentence(.equivalent, cause: .none, value) == key(L10n.Promote.lineEquivalentOther))
            #expect(sentence(.contains, cause: .precision, value) == key(L10n.Promote.lineContainsPrecision))
            #expect(sentence(.contains, cause: .missingPart, value) == key(L10n.Promote.lineContainsMissingPart))
            #expect(sentence(.contains, cause: .initial, value) == key(L10n.Promote.lineContainsInitial))
            #expect(sentence(.contains, cause: .none, value) == key(L10n.Promote.lineContainsOther))
            #expect(sentence(.near, cause: .spelling, value) == key(L10n.Promote.lineNearSpelling))
            #expect(sentence(.near, cause: .gap, value) == key(L10n.Promote.lineNearGap))
            #expect(sentence(.near, cause: .none, value) == key(L10n.Promote.lineSimilar))
        }
        #expect(sentence(.overlaps, day) == key(L10n.Promote.lineOverlapsDates))
        #expect(sentence(.overlaps, term) == key(L10n.Promote.lineOverlapsOther))
        // Different, or a negative that contradicts.
        #expect(sentence(.different, cause: .beyondTolerance, day) == key(L10n.Promote.lineDifferentDates))
        #expect(sentence(.different, name) == key(L10n.Promote.lineDifferentNames))
        #expect(sentence(.different, term) == key(L10n.Promote.lineDisagrees))
        #expect(sentence(.same, propertyCause: .negative, term) == key(L10n.Promote.lineDisagrees))
        #expect(sentence(.same, propertyCause: .negative, name) == key(L10n.Promote.lineDifferentNames))
        // Two images the layer cannot relate; codes this build does not know.
        #expect(sentence(.unknown, cause: .unjudged, likeness) == key(L10n.Promote.lineCompareLikenesses))
        #expect(sentence(.unrecognized("x"), name) == key(L10n.Promote.lineSimilar))
        #expect(sentence(.none, name) == key(L10n.Promote.notCompared))
    }

    @Test func badgesFollowTheOutcome() {
        #expect(L10n.Promote.outcome(.agree).key == "promote.sheet.agrees")
        #expect(L10n.Promote.outcome(.partial).key == "promote.sheet.similar")
        #expect(L10n.Promote.outcome(.conflict).key == "promote.sheet.conflicts")
        #expect(L10n.Promote.outcome(.unknown, cause: .unjudged).key == L10n.Promote.judgeByEye.key)
        #expect(L10n.Promote.outcome(.unknown).key == L10n.Promote.notCompared.key)
        #expect(L10n.Promote.outcome(.unrecognized("x")).key == L10n.Promote.notCompared.key)
        #expect(L10n.string(L10n.Promote.judgeByEye) == "Judge by eye")
        #expect(L10n.string(L10n.Promote.notCompared) == "Not compared")
    }

    // MARK: - T-3

    @Test func relationshipSentences() {
        #expect(L10n.string(L10n.Promote.relationshipSentence(.agree)) == "They share a matching connection")
        #expect(L10n.string(L10n.Promote.relationshipSentence(.partial)) == "They share a similar connection")
        #expect(
            L10n.Promote.connectionLink(first: "John Smith", phrase: "subject of", second: "Birth of John Smith")
                == "John Smith · subject of · Birth of John Smith"
        )
        #expect(L10n.Promote.relationshipWith("Mary Smith") == "with Mary Smith")
    }

    @Test func countsArePlural() {
        #expect(L10n.Promote.recordCount(1) == "· 1 record")
        #expect(L10n.Promote.recordCount(5) == "· 5 records")
        #expect(L10n.Promote.checkedCount(4) == "4 checked")
        #expect(L10n.Promote.showRecords(5) == "Show 5 records")
        #expect(L10n.Promote.evidenceHeading(ref: "PER-7QH3D", title: "John Smith") == "What's the evidence that this is PER-7QH3D John Smith?")
    }

    // MARK: - T-7, T-6

    @Test func aLineDetailJoinsItsParts() {
        let here = "This source: John Smith"
        let member = "1871 England Census — Leeds, Briggate: J. Smith"
        #expect(L10n.Promote.lineDetail(parts: [here, member]) == "\(here) · \(member)")
        #expect(L10n.Promote.lineDetail(parts: [here]) == here)
        #expect(L10n.Promote.lineDetail(parts: []) == "")
    }

    @Test func anOptionalFieldSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.optionalMarker) == "(optional)")
    }

    @Test func aPartlyCheckedBoxSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.checkboxPartlyChecked) == "partly checked")
    }

    // MARK: - T-5

    @Test func spokenLabelsReadTheLine() {
        #expect(
            L10n.Promote.spokenLineCount(property: "Name", sentence: "Same name", count: 5, here: "John Smith")
                == "Name. Same name, 5 records. This source: John Smith."
        )
        #expect(L10n.Promote.spokenLine(property: "Name", sentence: "Same name (nickname)", here: "John Smith") == "Name. Same name (nickname). This source: John Smith.")
        #expect(L10n.Promote.spokenMember(source: "Census 1851", value: "John Smith") == "Census 1851: John Smith.")
        #expect(
            L10n.Promote.spokenRelationship(
                sentence: L10n.string(L10n.Promote.relationshipAgrees),
                link: "John Smith · subject of · Birth of John Smith"
            ) == "They share a matching connection. John Smith · subject of · Birth of John Smith."
        )
    }

    @Test func twoLikenessLinesNeverReadAlike() {
        let same = L10n.Promote.spokenLikeness(source: "Funeral card", sentence: L10n.string(L10n.Promote.lineSameLikeness))
        let unjudged = L10n.Promote.spokenLikeness(source: "Smith family photograph album", sentence: L10n.string(L10n.Promote.lineCompareLikenesses))
        #expect(same == "Likeness from this source and Funeral card. Same likeness.")
        #expect(unjudged == "Likeness from this source and Smith family photograph album. Compare these likenesses.")
        #expect(same != unjudged)
        // Two different images from one Source still differ by their outcome.
        let again = L10n.Promote.spokenLikeness(source: "Funeral card", sentence: L10n.string(L10n.Promote.lineCompareLikenesses))
        #expect(same != again)
    }

    // MARK: - Round 1

    @Test func bridgeBetweenIgnoresKindWhenNoneIsGiven() {
        let participation = bridge("b-p", .participation, "john", "birth")
        let spouse = bridge("b-r", .relationship, "birth", "john")
        let bridges = [participation, spouse]
        #expect(PromoteEvidenceDisplay.bridge(between: "john", and: "birth", in: bridges)?.id == "b-p")
        #expect(PromoteEvidenceDisplay.bridge(between: "john", and: "birth", kind: nil, in: bridges.reversed())?.id == "b-r")
        #expect(PromoteEvidenceDisplay.bridge(between: "john", and: "birth", kind: .relationship, in: bridges)?.id == "b-r")
        #expect(PromoteEvidenceDisplay.bridge(between: "john", and: "death", in: bridges) == nil)
    }

    @Test func memberBoxIsDisabledUnderANonPinnableLine() {
        #expect(PromoteEvidenceDisplay.memberDisabled(line: line(outcome: .conflict, relation: .different)))
        #expect(!PromoteEvidenceDisplay.memberDisabled(line: line()))
    }

    @Test func spokenLabelOfACollapsedLineCounts() {
        let label = PromoteEvidenceDisplay.spokenLabel(
            property: "Name", sentence: "Same name", here: "John Smith", count: 5, member: nil
        )
        #expect(label == "Name. Same name, 5 records. This source: John Smith.")
    }

    @Test func spokenLabelOfOneMemberNamesItsSource() {
        let label = PromoteEvidenceDisplay.spokenLabel(
            property: "Name", sentence: "Same name", here: "John Smith", count: 1,
            member: PromoteEvidenceDisplay.Member(source: "Census 1851", value: "J. Smith")
        )
        #expect(label == "Name. Same name. This source: John Smith. Census 1851: J. Smith.")
    }

    @Test func spokenLabelWithoutASourceTitleSpeaksTheValue() {
        for source in [nil, ""] as [String?] {
            let label = PromoteEvidenceDisplay.spokenLabel(
                property: "Name", sentence: "Same name", here: "John Smith", count: 1,
                member: PromoteEvidenceDisplay.Member(source: source, value: "J. Smith")
            )
            #expect(label == "Name. Same name. This source: John Smith. J. Smith.")
            #expect(PromoteEvidenceDisplay.spokenMember(PromoteEvidenceDisplay.Member(source: source, value: "J. Smith")) == "J. Smith.")
        }
    }

    @Test func detailPartsOfOneMemberLine() {
        let member = PromoteEvidenceDisplay.Member(source: "Census 1851", value: "J. Smith")
        #expect(PromoteEvidenceDisplay.detail(here: "John Smith", member: member) == "This source: John Smith · Census 1851: J. Smith")
        #expect(PromoteEvidenceDisplay.detail(here: "John Smith", member: nil) == "This source: John Smith")
        let untitled = PromoteEvidenceDisplay.Member(source: nil, value: "J. Smith")
        #expect(PromoteEvidenceDisplay.detail(here: "John Smith", member: untitled) == "This source: John Smith · J. Smith")
    }

    @Test func relationshipHeaderIsTheBridgeSentence() {
        var found = bridge("b", .relationship, "marion", "don")
        found.sentence = "Marion Smith is the mother of Don Smith"
        let agrees = PromoteEvidenceDisplay.relationshipHeader(bridge: found, outcome: .agree, neighborName: "Don")
        #expect(agrees.title == "Marion Smith is the mother of Don Smith")
        #expect(agrees.subtitle == nil)
        #expect(agrees.spoken == "Marion Smith is the mother of Don Smith. Agrees.")
        let similar = PromoteEvidenceDisplay.relationshipHeader(bridge: found, outcome: .partial, neighborName: "Don")
        #expect(similar.spoken == "Marion Smith is the mother of Don Smith. Similar.")

        let named = PromoteEvidenceDisplay.relationshipHeader(bridge: nil, outcome: .agree, neighborName: "Birth")
        #expect(named.title == "They share a matching connection")
        #expect(named.subtitle == "with Birth")
        #expect(named.spoken == L10n.Promote.spokenRelationship(sentence: "They share a matching connection", link: "with Birth"))
        let unnamed = PromoteEvidenceDisplay.relationshipHeader(bridge: nil, outcome: .partial, neighborName: nil)
        #expect(unnamed.title == "They share a similar connection")
        #expect(unnamed.subtitle == "with a connected record")
    }

    @Test func nothingToCompareWhenNoLineOrRelationship() {
        #expect(PromoteEvidenceDisplay.hasNothingToCompare(CatalogClaimEvidence()))
        #expect(PromoteEvidenceDisplay.hasNothingToCompare(CatalogClaimEvidence(sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [])])))
        #expect(!PromoteEvidenceDisplay.hasNothingToCompare(CatalogClaimEvidence(sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [line()])])))
        #expect(!PromoteEvidenceDisplay.hasNothingToCompare(CatalogClaimEvidence(relationships: [relationship()])))
    }

    @Test func likenessMemberWithoutTitleUsesFallback() {
        #expect(PromoteEvidenceDisplay.likenessSource(title: "Funeral card") == "Funeral card")
        #expect(PromoteEvidenceDisplay.likenessSource(title: nil) == "a record")
        #expect(PromoteEvidenceDisplay.likenessSource(title: "") == "a record")
    }
}
