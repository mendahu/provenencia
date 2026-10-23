import Foundation
import Testing
@testable import Provenencia

/// The evidence sheet's choices and words, tested without a view.
@Suite
struct PromoteEvidenceDisplayTests {

    // MARK: - Fixtures

    private func pin(_ id: String, source: String = "src-1", value: String = "John Smith") -> CatalogClaimEvidencePin {
        CatalogClaimEvidencePin(
            id: id, sourceID: source,
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
            propertyKey: key, outcome: outcome, relation: relation,
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
            neighborSubjectID: neighbor, key: "k", bridgeType: type, outcome: .agree,
            bundles: [CatalogClaimEvidenceBundle(id: "b1")], sections: sections
        )
    }

    private func date(_ year: Int32, _ month: Int32? = nil, _ day: Int32? = nil) -> CatalogDateValueInput {
        CatalogDateValueInput(kind: "point", startYear: year, startMonth: month, startDay: day)
    }


    @Test func checkStateFollowsThePins() {
        let ids = ["p1", "p2", "p3", "p4", "p5"]
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: Set(ids)) == .checked)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: []) == .unchecked)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: ids, pins: ["p1", "p3", "p4", "p5"]) == .partial)
        #expect(PromoteEvidenceDisplay.checkState(pinIDs: [], pins: ["p1"]) == .unchecked)
    }

    @Test func checkedCountCountsCheckedOffers() {
        let offered: Set<String> = ["p1", "p2"]
        #expect(PromoteEvidenceDisplay.checkedCount(offered: offered, pins: ["p1", "x"]) == 1)
        #expect(PromoteEvidenceDisplay.checkedCount(offered: offered, pins: ["p1", "p2"]) == 2)
        #expect(PromoteEvidenceDisplay.checkedCount(offered: offered, pins: []) == 0)
    }


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
        let term = CatalogConclusionValue.term(id: "t", key: "male", label: "Male", origin: CatalogOrigin.provenencia)
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


    @Test func countsArePlural() {
        #expect(L10n.Promote.recordCount(1) == "· 1 record")
        #expect(L10n.Promote.recordCount(5) == "· 5 records")
        #expect(L10n.Promote.checkedCount(4) == "4 checked")
        #expect(L10n.Promote.showRecords(5) == "Show 5 records")
        #expect(L10n.Promote.hideRecords(1) == "Hide 1 record")
        #expect(L10n.Promote.hideRecords(4) == "Hide 4 records")
    }


    @Test func anOptionalFieldSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.optionalMarker) == "(optional)")
    }

    @Test func aPartlyCheckedBoxSaysSo() {
        #expect(L10n.string(L10n.DesignSystem.checkboxPartlyChecked) == "partly checked")
    }


    @Test func spokenLabelsReadTheLine() {
        #expect(
            L10n.Promote.spokenLineCount(property: "Name", sentence: "Same name", count: 5, here: "John Smith")
                == "Name. Same name, 5 records. This source: John Smith."
        )
        #expect(L10n.Promote.spokenLine(property: "Name", sentence: "Same name (nickname)", here: "John Smith") == "Name. Same name (nickname). This source: John Smith.")
        #expect(L10n.Promote.spokenMember(source: "Census 1851", value: "John Smith") == "Census 1851: John Smith.")
        #expect(
            L10n.Promote.spokenRelationship(
                sentence: "…they share a spouse (CEV-3B5RT and PER-7QH3D)",
                outcome: "Agrees"
            ) == "…they share a spouse (CEV-3B5RT and PER-7QH3D). Agrees."
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

    @Test func spokenLabelOfACollapsedLineCounts() {
        let label = PromoteEvidenceDisplay.spokenLabel(
            property: "Name", sentence: "Same name", here: "John Smith", count: 5, member: nil
        )
        #expect(label == "Name. Same name, 5 records. This source: John Smith.")
    }

    @Test func spokenLabelOfOneMemberNamesItsSource() {
        let label = PromoteEvidenceDisplay.spokenLabel(
            property: "Name", sentence: "Same name", here: "John Smith", count: 1,
            member: (source: "Census 1851", value: "J. Smith")
        )
        #expect(label == "Name. Same name. This source: John Smith. Census 1851: J. Smith.")
    }

    @Test func spokenLabelWithoutASourceTitleSpeaksAnUntitledRecord() {
        // The one rule: a Source with no title is "a record", on the line and on its member.
        let source = PromoteEvidenceDisplay.recordSource(title: nil)
        let label = PromoteEvidenceDisplay.spokenLabel(
            property: "Name", sentence: "Same name", here: "John Smith", count: 1,
            member: (source: source, value: "J. Smith")
        )
        #expect(label == "Name. Same name. This source: John Smith. a record: J. Smith.")
        #expect(PromoteEvidenceDisplay.spokenMember(source: source, value: "J. Smith") == "a record: J. Smith.")
    }

    @Test func aRelationshipAndALikenessSpeakTheirCountOnlyWhenTheyBundleSeveral() {
        #expect(PromoteEvidenceDisplay.spokenRelationship(header: "they share a spouse", outcome: .agree, count: 1) == "they share a spouse. Agrees.")
        #expect(PromoteEvidenceDisplay.spokenRelationship(header: "they share a spouse", outcome: .agree, count: 3) == "they share a spouse. Agrees, 3 records.")
        #expect(PromoteEvidenceDisplay.spokenLikeness(source: "Funeral card", sentence: "Same likeness", count: 1)
            == "Likeness from this source and Funeral card. Same likeness.")
        #expect(PromoteEvidenceDisplay.spokenLikeness(source: "Funeral card", sentence: "Same likeness", count: 2)
            == "Likeness from this source and Funeral card. Same likeness, 2 records.")
    }

    @Test func aSourceWithoutATitleIsARecord() {
        #expect(PromoteEvidenceDisplay.recordSource(title: "Funeral card") == "Funeral card")
        #expect(PromoteEvidenceDisplay.recordSource(title: nil) == "a record")
        #expect(PromoteEvidenceDisplay.recordSource(title: "") == "a record")
    }

    // MARK: - Round 16

    private func plain(_ text: AttributedString) -> String { String(text.characters) }

    @Test func theTitleJoinsTheRowToItsRecord() {
        let title = PromoteEvidenceDisplay.title(name: "John Smith", recordTitle: "Johnny Smith", ref: "PER-7QH3D")
        #expect(plain(title) == "John Smith joins Johnny Smith PER-7QH3D")
        // The connecting word is its own run; the ref is another.
        let runs = title.runs.map { String(title.characters[$0.range]) }
        #expect(runs.contains(" joins "))
        #expect(runs.contains("PER-7QH3D"))
        #expect(role(of: " joins ", in: title) == .muted)
        #expect(role(of: "PER-7QH3D", in: title) == .ref)
    }

    private func role(of text: String, in styled: AttributedString) -> PromoteEvidenceDisplay.Role? {
        styled.runs.first { String(styled.characters[$0.range]) == text }?[PromoteEvidenceDisplay.RoleAttribute.self]
    }

    @Test func theDisplayHelperNamesRolesAndStylesNothing() {
        let subtitle = PromoteEvidenceDisplay.subtitle(
            name: "John Smith", sourceTitle: "1871 Census", kind: .person, recordTitle: "Johnny Smith"
        )
        #expect(role(of: "John Smith", in: subtitle) == .name)
        #expect(role(of: "1871 Census", in: subtitle) == .source)
        #expect(role(of: "Johnny Smith", in: subtitle) == .name)
        let heading = PromoteEvidenceDisplay.evidenceHeading(
            name: "John Smith", sourceTitle: "1871 Census", kind: .person, recordTitle: "Johnny Smith"
        )
        #expect(role(of: "John Smith", in: heading) == .name)
        #expect(role(of: "1871 Census", in: heading) == .source)
        #expect(role(of: "Johnny Smith", in: heading) == .name)
        // No font or color: the sheet owns them.
        for styled in [heading, subtitle] {
            for run in styled.runs {
                #expect(run[AttributeScopes.SwiftUIAttributes.FontAttribute.self] == nil)
                #expect(run[AttributeScopes.SwiftUIAttributes.ForegroundColorAttribute.self] == nil)
            }
        }
        let untitled = PromoteEvidenceDisplay.evidenceHeading(name: "Leeds", sourceTitle: nil, kind: .place, recordTitle: "Leeds, Yorkshire")
        #expect(role(of: "Leeds", in: untitled) == .name)
    }

    @Test func theSubtitleNamesTheSourceAndTheConcludedRecord() {
        let subtitle = PromoteEvidenceDisplay.subtitle(
            name: "John Smith", sourceTitle: "1871 Census", kind: .person, recordTitle: "Johnny Smith"
        )
        #expect(plain(subtitle) == "This is an identity claim to join John Smith from 1871 Census into the concluded person Johnny Smith. Check the box on each observation below that supports the claim. Observations that clearly match are already checked — review them before you file.")

        let untitled = PromoteEvidenceDisplay.subtitle(name: "Birth", sourceTitle: nil, kind: .event, recordTitle: "Birth of John")
        #expect(plain(untitled).hasPrefix("This is an identity claim to join Birth from this source into the concluded event Birth of John."))
    }

    @Test func theHeadingSaysWhyTheRowBelongs() {
        let heading = PromoteEvidenceDisplay.evidenceHeading(
            name: "John Smith", sourceTitle: "1871 Census", kind: .person, recordTitle: "Johnny Smith"
        )
        #expect(plain(heading) == "John Smith from 1871 Census belongs in the concluded person Johnny Smith because…")
        let untitled = PromoteEvidenceDisplay.evidenceHeading(name: "Leeds", sourceTitle: nil, kind: .place, recordTitle: "Leeds, Yorkshire")
        #expect(plain(untitled) == "Leeds from this source belongs in the concluded place Leeds, Yorkshire because…")
    }

    @Test func aCardHeaderReadsTheLowercasePropertyLabel() {
        #expect(PromoteEvidenceDisplay.similarityHeader(property: "Name") == "they have a similarity in name")
        #expect(PromoteEvidenceDisplay.similarityHeader(property: "Sex at birth") == "they have a similarity in sex at birth")
    }

    @Test func thisSourceIsSaidOnceWhenEveryLineAgreesOnIt() {
        #expect(PromoteEvidenceDisplay.thisSourcePlacement(values: ["John Smith"]) == .header("John Smith"))
        #expect(PromoteEvidenceDisplay.thisSourcePlacement(values: ["John Smith", "John Smith"]) == .header("John Smith"))
        #expect(PromoteEvidenceDisplay.thisSourcePlacement(values: ["John Smith", "J. Smith"]) == .lines)
        // A likeness says it nowhere; an empty Card has nothing to say.
        #expect(PromoteEvidenceDisplay.thisSourcePlacement(values: [nil]) == .nowhere)
        #expect(PromoteEvidenceDisplay.thisSourcePlacement(values: []) == .nowhere)
        #expect(PromoteEvidenceDisplay.ThisSourcePlacement.header("x").headerValue == "x")
        #expect(PromoteEvidenceDisplay.ThisSourcePlacement.lines.headerValue == nil)
    }

    @Test func aRelationshipHeaderNamesTheRefsItKnows() {
        let phrase = "a spouse"
        #expect(PromoteEvidenceDisplay.shareHeader(phrase: phrase, ownRef: "CEV-3B5RT", recordRef: "PER-7QH3D")
            == "they share a spouse (CEV-3B5RT and PER-7QH3D)")
        // A neighbor not yet filed on a record has only its own ref.
        #expect(PromoteEvidenceDisplay.shareHeader(phrase: phrase, ownRef: "CEV-3B5RT", recordRef: nil)
            == "they share a spouse (CEV-3B5RT)")
        #expect(PromoteEvidenceDisplay.shareHeader(phrase: phrase, ownRef: "CEV-3B5RT", recordRef: "CEV-3B5RT")
            == "they share a spouse (CEV-3B5RT)")
        #expect(PromoteEvidenceDisplay.shareHeader(phrase: phrase, ownRef: nil, recordRef: "PER-7QH3D")
            == "they share a spouse (PER-7QH3D)")
        #expect(PromoteEvidenceDisplay.shareHeader(phrase: phrase, ownRef: nil, recordRef: nil) == "they share a spouse")
    }

    /// A relationship as the engine reports it from the neighbor's side.
    private func shared(
        type: String,
        end: String = "",
        term: CatalogEvidenceTerm? = nil,
        names: Bool = true,
        eventType: CatalogEvidenceTerm? = nil,
        outcome: EvidenceOutcome = .agree
    ) -> CatalogClaimEvidenceRelationship {
        var r = relationship(type: type)
        r.outcome = outcome
        r.neighborEnd = end
        r.neighborTerm = term
        r.neighborTermNamesNeighbor = names
        r.eventType = eventType
        return r
    }

    private func product(_ key: String) -> CatalogEvidenceTerm {
        CatalogEvidenceTerm(key: key, label: key.capitalized, origin: CatalogOrigin.provenencia)
    }

    @Test func relationshipShareFromTheEngine() {
        func phrase(_ r: CatalogClaimEvidenceRelationship, _ name: String = "Mary") -> String {
            PromoteEvidenceDisplay.sharePhrase(relationship: r, neighborName: name)
        }
        // Spouses who are also cousins: one phrase for each of the two cards.
        #expect(phrase(shared(type: "relationship", term: product("spouse"))) == "a spouse")
        #expect(phrase(shared(type: "relationship", term: product("cousin"))) == "a cousin")
        #expect(phrase(shared(type: "relationship", term: product("spouse"), outcome: .partial)) == "a similar spouse")
        // Mary is the parent of John: from John's sheet she is "a parent".
        #expect(phrase(shared(type: "relationship", end: "related_to", term: product("parent"))) == "a parent")
        // A researcher's directed term with no inverse names the connection.
        let godparent = CatalogEvidenceTerm(key: "godparent", label: "Godparent", origin: CatalogOrigin.user)
        #expect(phrase(shared(type: "relationship", end: "related_to", term: godparent, names: false)) == "a connection (Godparent)")
        #expect(phrase(shared(type: "relationship", end: "related_to", term: godparent, names: false, outcome: .partial)) == "a similar connection (Godparent)")
        // The neighbor is John's birth event.
        #expect(phrase(shared(type: "participation", end: "event", term: product("subject"), names: false, eventType: product("birth")), "Birth of John") == "a birth event")
        // Leeds is part of Yorkshire: Yorkshire is a parent place, and from its sheet Leeds is within it.
        #expect(phrase(shared(type: "place_relationship", end: "to", term: product("part_of"), names: false)) == "a parent place")
        #expect(phrase(shared(type: "place_relationship", end: "from", term: product("part_of"))) == "a place within it")
        // A bridge type this build cannot word, or no reading at all, is a connection.
        #expect(phrase(shared(type: "nonsense")) == "a connection")
        #expect(phrase(shared(type: "relationship")) == "a connection")
    }

    @Test func aVowelInitialRoleOrEventReadsWithItsOwnArticle() {
        let informant = shared(type: "participation", end: "person", term: product("informant"))
        #expect(PromoteEvidenceDisplay.sharePhrase(relationship: informant, neighborName: "John Smith") == "an informant, John Smith")
        let engagement = shared(type: "participation", end: "event", eventType: product("engagement"), outcome: .partial)
        #expect(PromoteEvidenceDisplay.sharePhrase(relationship: engagement, neighborName: "x") == "a similar engagement event")
    }

    @Test func aCardHeaderIsDrawnAfterAnEllipsisAndSpokenWithout() {
        let header = PromoteEvidenceDisplay.shareHeader(phrase: "a spouse", ownRef: "CEV-3B5RT", recordRef: "PER-7QH3D")
        #expect(L10n.Promote.cardHeaderLead(header) == "…they share a spouse (CEV-3B5RT and PER-7QH3D)")
        #expect(!header.hasPrefix("…"))
        #expect(L10n.Promote.cardHeaderLead(PromoteEvidenceDisplay.similarityHeader(property: "Name")) == "…they have a similarity in name")
    }

    @Test func subtitleAndHeadingPerKind() {
        for (kind, word) in [(EvidencePrimaryKind.person, "person"), (.event, "event"), (.place, "place")] {
            // One key per kind and per titled or untitled Source, each a whole sentence.
            #expect(L10n.Promote.subtitle(kind).key == "promote.sheet.subtitle.\(word)")
            #expect(L10n.Promote.subtitleUntitled(kind).key == "promote.sheet.subtitleUntitled.\(word)")
            #expect(L10n.Promote.evidenceHeading(kind).key == "promote.sheet.evidenceHeading.\(word)")
            #expect(L10n.Promote.evidenceHeadingUntitled(kind).key == "promote.sheet.evidenceHeadingUntitled.\(word)")
            let titled = PromoteEvidenceDisplay.subtitle(name: "A", sourceTitle: "Census", kind: kind, recordTitle: "B")
            #expect(plain(titled).contains("from Census into the concluded \(word) B."))
            let untitled = PromoteEvidenceDisplay.evidenceHeading(name: "A", sourceTitle: nil, kind: kind, recordTitle: "B")
            #expect(plain(untitled) == "A from this source belongs in the concluded \(word) B because…")
        }
    }

    @Test func titleHoldsJoins() {
        let title = PromoteEvidenceDisplay.title(name: "John Smith", recordTitle: "John Smith", ref: "PER-7QH3D")
        #expect(plain(title) == "John Smith joins John Smith PER-7QH3D")
        #expect(L10n.Promote.titleHandle.key == "promote.sheet.titleHandle")
        // "joins" lives inside the one format, not in a second key.
        #expect(role(of: " joins ", in: title) == .muted)
    }
}
