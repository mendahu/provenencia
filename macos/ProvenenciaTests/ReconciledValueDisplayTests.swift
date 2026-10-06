import AppKit
import Foundation
import Testing
@testable import Provenencia

/// Fixtures mirror board S9-D5's frames (1b–1h): a merged name with a folded
/// initial, an outvoted spelling and a low-trust record; a mixed death date;
/// weak and denied records with the negative behind them.
struct ReconciledValueDisplayTests {
    static let jamesRobins = CatalogNameValue(form: "James Robins", parts: [
        CatalogNameValuePart(value: "James", type: "given"),
        CatalogNameValuePart(value: "Robins", type: "surname"),
    ])

    static func value(_ rank: Int, _ reason: String, support: Int = 1, against: Int = 0,
                      _ v: CatalogConclusionValue = .text("x")) -> CatalogReconciledValue {
        CatalogReconciledValue(rank: rank, reason: reason, support: support, against: against, value: v)
    }

    static func outcome(
        _ id: String,
        _ reason: String,
        rank: Int? = 1,
        source: String = "src-\(UUID().uuidString)",
        title: String = "Census",
        recorded: CatalogConclusionValue = .text("x"),
        deniedBy: String = "",
        credibilityOffset: Int = 0,
        uncertain: Bool = false,
        claimConfidenceOffset: Int = 0,
        vote: (Int, Int) = (0, 0)
    ) -> CatalogReconcilerOutcome {
        CatalogReconcilerOutcome(
            observationID: id, observationRef: "OBS-\(id)", reason: reason, valueRank: rank,
            deniedByObservationID: deniedBy, recorded: recorded, subjectID: "sub-\(id)", subjectRef: "CPR-\(id)",
            citationID: "cit-\(id)", artifactID: "art-\(id)", sourceID: source, sourceTitle: title, credibilityKey: "",
            transcriptionUncertain: uncertain, claimConfidenceKey: "",
            credibilityOffset: credibilityOffset, claimConfidenceOffset: claimConfidenceOffset,
            voteSupport: vote.0, voteTotal: vote.1
        )
    }

    static func field(
        _ state: String,
        key: String = "name",
        label: String = "Name",
        valueType: String = "name",
        _ values: [CatalogReconciledValue] = [],
        outcomes: [CatalogReconcilerOutcome] = []
    ) -> CatalogConclusionField {
        CatalogConclusionField(
            propertyID: "p-\(key)", propertyKey: key, label: label, valueType: valueType,
            state: state, values: values, outcomes: outcomes
        )
    }

    /// Frame 1c: the merged name and its four records.
    static var mergedName: CatalogConclusionField {
        field("merged", [
            value(1, "kept", support: 2, .name(jamesRobins)),
            value(2, "outvoted", .text("James Robbins")),
            value(3, "weak", .text("James Robins")),
        ], outcomes: [
            outcome("census", "kept", source: "s-census", title: "Census of Canada West, 1851", recorded: .name(jamesRobins)),
            outcome("baptism", "folded", source: "s-baptism", title: "Baptism register, St James, York, 1817",
                    recorded: .name(CatalogNameValue(form: "J. Robins"))),
            outcome("bible", "outvoted", rank: 2, source: "s-bible", title: "Family Bible, Robins family",
                    recorded: .name(CatalogNameValue(form: "James Robbins")), vote: (2, 3)),
            outcome("news", "weak", rank: 3, source: "s-news", title: "Death notice, The Globe, 1880",
                    recorded: .name(jamesRobins), credibilityOffset: -1),
        ])
    }

    @Test func badgesNameEveryShownState() {
        #expect(ReconciledValueDisplay.stateBadge(Self.field("")) == nil)
        #expect(ReconciledValueDisplay.stateBadge(Self.field("single", [Self.value(1, "kept")])) == nil)
        #expect(ReconciledValueDisplay.stateBadge(Self.mergedName) == .merged)
        #expect(ReconciledValueDisplay.stateBadge(Self.field("mixed")) == .mixed)
        #expect(ReconciledValueDisplay.stateBadge(Self.field("concluded")) == .concluded)
    }

    @Test func countLineCountsSourcesNotRecords() {
        #expect(ReconciledValueDisplay.countLine(Self.field("")) == nil)
        #expect(ReconciledValueDisplay.countLine(Self.field("single", [Self.value(1, "kept")])) == "1 Source")
        #expect(ReconciledValueDisplay.countLine(Self.mergedName) == "2 Sources")
        #expect(ReconciledValueDisplay.countLine(Self.field("concluded", [Self.value(1, "kept")])) == "by you")
        // Mixed: distinct Sources across the shown values; one Source behind
        // two values counts once.
        let mixed = Self.field("mixed", key: "deathDate", [
            Self.value(1, "kept", support: 2), Self.value(2, "kept", support: 2),
        ], outcomes: [
            Self.outcome("burial", "kept", rank: 1, source: "a"),
            Self.outcome("probate", "kept", rank: 1, source: "b"),
            Self.outcome("bible", "kept", rank: 2, source: "b"),
            Self.outcome("bible2", "kept", rank: 2, source: "c"),
            Self.outcome("census", "no_evidence", rank: nil, source: "d"),
        ])
        #expect(ReconciledValueDisplay.countLine(mixed) == "3 Sources")
    }

    @Test func againstLineSaysHowManyRecordsDisagree() {
        #expect(ReconciledValueDisplay.againstLine(Self.mergedName) == nil)
        #expect(ReconciledValueDisplay.againstLine(Self.field("merged", [Self.value(1, "kept", support: 2, against: 1)])) == "1 record disagrees")
        #expect(ReconciledValueDisplay.againstLine(Self.field("single", [Self.value(1, "kept", against: 2)])) == "2 records disagree")
    }

    @Test func onlyAMixedFieldDisclosesOtherValues() {
        let mixed = Self.field("mixed", [Self.value(1, "kept"), Self.value(2, "kept"), Self.value(3, "outvoted")])
        #expect(ReconciledValueDisplay.otherValues(mixed).map(\.rank) == [2])
        #expect(ReconciledValueDisplay.otherValuesLabel(mixed) == "1 other value")
        let three = Self.field("mixed", [Self.value(1, "kept"), Self.value(2, "kept"), Self.value(3, "kept")])
        #expect(ReconciledValueDisplay.otherValuesLabel(three) == "2 other values")
        #expect(ReconciledValueDisplay.otherValuesLabel(Self.mergedName) == nil)
    }

    @Test func emptyTextIsHonestPerField() {
        #expect(ReconciledValueDisplay.emptyText(propertyKey: "name") == "No name recorded")
        #expect(ReconciledValueDisplay.emptyText(propertyKey: "sex_at_birth") == "No sex at birth recorded")
        #expect(ReconciledValueDisplay.emptyText(propertyKey: "occupation") == "Nothing recorded")
    }

    @Test func valuesUseTheirOwnDisplays() {
        let en = Locale(identifier: "en_US")
        #expect(ReconciledValueDisplay.string(for: .none) == "")
        #expect(ReconciledValueDisplay.string(for: .text("  farmer ")) == "farmer")
        #expect(ReconciledValueDisplay.string(for: .integer(1890), locale: en) == "1890")
        #expect(ReconciledValueDisplay.string(for: .term(id: "t", key: "male", label: "Male")) == "Male")
        #expect(ReconciledValueDisplay.string(for: .term(id: "t", key: "male", label: "")) == "male")
        #expect(ReconciledValueDisplay.string(for: .name(Self.jamesRobins)) == "James Robins")
        let date = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        #expect(ReconciledValueDisplay.string(for: .date(date), locale: en) == EventTitleDisplay.dateLine(date: date, locale: en))
        #expect(ReconciledValueDisplay.string(for: .date(date), locale: en) == "14 May 1817")
    }

    @Test func frame1cPhrasesAndMarks() {
        let f = Self.mergedName
        func phrase(_ i: Int) -> String { ReconciledValueDisplay.outcomePhrase(f.outcomes[i], in: f) }
        func mark(_ i: Int) -> PVSymbol { ReconciledValueDisplay.outcomeMark(f.outcomes[i], in: f) }
        #expect(phrase(0) == "kept" && mark(0) == .check)
        #expect(phrase(1) == "folded into James Robins" && mark(1) == .gitMerge)
        #expect(phrase(2) == "outvoted (2 of 3 Sources)" && mark(2) == .scale)
        #expect(phrase(3) == "weak · low-trust Source" && mark(3) == .signalLow)
        #expect(ReconciledValueDisplay.readAs(f.outcomes[1]) == "J. Robins")
        // Kept and folded records went into the value; the rest didn't.
        #expect(f.outcomes.map { ReconciledValueDisplay.counted($0, in: f) } == [true, true, false, false])
    }

    @Test func outvotedWithoutAVoteSaysOutvoted() {
        let f = Self.field("single", [Self.value(1, "kept")], outcomes: [Self.outcome("a", "outvoted", rank: 2)])
        #expect(ReconciledValueDisplay.outcomePhrase(f.outcomes[0], in: f) == "outvoted")
        let one = Self.field("single", [Self.value(1, "kept")], outcomes: [Self.outcome("a", "outvoted", rank: 2, vote: (1, 1))])
        #expect(ReconciledValueDisplay.outcomePhrase(one.outcomes[0], in: one) == "outvoted (1 of 1 Source)")
    }

    /// Frame 1e: an uncertain transcription is weak; a deposition's negative
    /// denies James Robinson (and reads as kept); a negative that eliminated
    /// nothing disagrees.
    @Test func frame1eWeakDeniedAndAgainst() {
        let name = Self.field("merged", [Self.value(1, "kept", support: 2, .name(Self.jamesRobins))], outcomes: [
            Self.outcome("pass", "weak", rank: 2, title: "Passenger list index, Quebec, 1832", uncertain: true),
            Self.outcome("dir", "denied", rank: 3, title: "City directory, Toronto, 1856",
                         recorded: .text("James Robinson"), deniedBy: "depo"),
            Self.outcome("depo", "against", rank: 3, title: "Court deposition, 1862", recorded: .text("James Robinson")),
        ])
        func phrase(_ i: Int) -> String { ReconciledValueDisplay.outcomePhrase(name.outcomes[i], in: name) }
        #expect(phrase(0) == "weak · uncertain transcription")
        #expect(phrase(1) == "denied by Court deposition, 1862")
        #expect(ReconciledValueDisplay.outcomeMark(name.outcomes[1], in: name) == .ban)
        #expect(phrase(2) == "kept")
        #expect(ReconciledValueDisplay.outcomeMark(name.outcomes[2], in: name) == .check)
        #expect(ReconciledValueDisplay.readAs(name.outcomes[2]) == "not James Robinson")
        #expect(ReconciledValueDisplay.readAsIsPhrase(name.outcomes[2]))
        #expect(ReconciledValueDisplay.counted(name.outcomes[2], in: name))

        let place = Self.field("merged", key: "birthPlace", label: "Birth place", valueType: "text",
                               [Self.value(1, "kept", support: 2, against: 1, .text("York, Upper Canada"))],
                               outcomes: [Self.outcome("bible", "against", rank: 1, recorded: .text("York"))])
        #expect(ReconciledValueDisplay.outcomePhrase(place.outcomes[0], in: place) == "disagrees · did not eliminate")
        #expect(ReconciledValueDisplay.outcomeMark(place.outcomes[0], in: place) == .circleMinus)
        #expect(!ReconciledValueDisplay.counted(place.outcomes[0], in: place))
        #expect(ReconciledValueDisplay.againstLine(place) == "1 record disagrees")
    }

    @Test func deniedFallsBackToTheRefThenToAPlainPhrase() {
        let noTitle = Self.field("single", [Self.value(1, "kept")], outcomes: [
            Self.outcome("dir", "denied", rank: 2, deniedBy: "depo"),
            Self.outcome("depo", "against", rank: nil, title: ""),
        ])
        #expect(ReconciledValueDisplay.outcomePhrase(noTitle.outcomes[0], in: noTitle) == "denied by OBS-depo")
        let gone = Self.field("single", [Self.value(1, "kept")], outcomes: [Self.outcome("dir", "denied", rank: 2, deniedBy: "x")])
        #expect(ReconciledValueDisplay.outcomePhrase(gone.outcomes[0], in: gone) == "denied")
    }

    @Test func otherReasonsHaveAPhraseAndAMark() {
        let f = Self.field("single", [Self.value(1, "kept")])
        func both(_ o: CatalogReconcilerOutcome) -> (String, PVSymbol) {
            (ReconciledValueDisplay.outcomePhrase(o, in: f), ReconciledValueDisplay.outcomeMark(o, in: f))
        }
        #expect(both(Self.outcome("p", "provisional", rank: 2)) == ("provisional member", .circleDashed))
        let none = Self.outcome("n", "no_evidence", rank: nil, recorded: .none)
        #expect(both(none) == ("no usable value", .minus))
        #expect(ReconciledValueDisplay.readAs(none) == "—")
        #expect(ReconciledValueDisplay.readAsIsPhrase(none))
        #expect(ReconciledValueDisplay.outcomePhrase(Self.outcome("h", "someday"), in: f) == "")
    }

    @Test func weakSaysWhichEvidenceIsWeak() {
        let f = Self.field("single", [Self.value(1, "kept")])
        func phrase(_ o: CatalogReconcilerOutcome) -> String { ReconciledValueDisplay.outcomePhrase(o, in: f) }
        #expect(phrase(Self.outcome("a", "weak", claimConfidenceOffset: -1)) == "weak · low-confidence claim")
        #expect(phrase(Self.outcome("a", "weak")) == "weak")
        // Above-default grades are not causes.
        #expect(phrase(Self.outcome("a", "weak", credibilityOffset: 1, uncertain: true, claimConfidenceOffset: 1)) == "weak · uncertain transcription")
        let all = Self.outcome("a", "weak", credibilityOffset: -2, uncertain: true, claimConfidenceOffset: -1)
        #expect(ReconciledValueDisplay.weakCauses(all) == ["low-trust Source", "uncertain transcription", "low-confidence claim"])
        #expect(phrase(all) == "weak · " + ListFormatter.localizedString(byJoining: ReconciledValueDisplay.weakCauses(all)))
    }

    @Test func voiceOverReadsStateInWords() {
        #expect(ReconciledValueDisplay.accessibilityLabel(label: "Name", lead: "James Robins", field: Self.mergedName)
            == "Name, James Robins, merged from 2 Sources")
        let place = Self.field("merged", [Self.value(1, "kept", support: 2, against: 1)])
        #expect(ReconciledValueDisplay.accessibilityLabel(label: "Birth place", lead: "York", field: place)
            == "Birth place, York, merged from 2 Sources, 1 record disagrees")
        #expect(ReconciledValueDisplay.accessibilityLabel(label: "Name", lead: "Jim", field: Self.field("single", [Self.value(1, "kept")]))
            == "Name, Jim, single, 1 Source")
        #expect(ReconciledValueDisplay.accessibilityLabel(label: "Birth date", lead: nil, field: nil) == "Birth date, empty")
    }
}
