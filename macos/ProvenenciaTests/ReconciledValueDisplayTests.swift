import Foundation
import Testing
@testable import Provenencia

struct ReconciledValueDisplayTests {
    private let jim = CatalogNameValue(form: "James Robins", parts: [
        CatalogNameValuePart(value: "James", type: "given"),
        CatalogNameValuePart(value: "Robins", type: "surname"),
    ])

    private func value(_ rank: Int, _ reason: String, support: Int = 1, _ v: CatalogConclusionValue = .text("x")) -> CatalogReconciledValue {
        CatalogReconciledValue(rank: rank, reason: reason, support: support, against: 0, value: v)
    }

    private func outcome(
        _ id: String,
        _ reason: String,
        rank: Int? = 1,
        deniedBy: String = "",
        credibilityOffset: Int = 0,
        uncertain: Bool = false,
        claimConfidenceOffset: Int = 0
    ) -> CatalogReconcilerOutcome {
        CatalogReconcilerOutcome(
            observationID: id, observationRef: "OBS-\(id)", reason: reason, valueRank: rank,
            deniedByObservationID: deniedBy, recorded: .text("x"), subjectID: "s", subjectRef: "CPR-1",
            citationID: "c", sourceID: "src", sourceTitle: "Census", credibilityKey: "",
            transcriptionUncertain: uncertain, claimConfidenceKey: "",
            credibilityOffset: credibilityOffset, claimConfidenceOffset: claimConfidenceOffset
        )
    }

    private func field(
        _ state: String,
        _ values: [CatalogReconciledValue] = [],
        outcomes: [CatalogReconcilerOutcome] = []
    ) -> CatalogConclusionField {
        CatalogConclusionField(
            propertyID: "p", propertyKey: "name", label: "Name", valueType: "name",
            state: state, values: values, outcomes: outcomes
        )
    }

    @Test func stateLineWordsEachStateWithSourcePlurals() {
        #expect(ReconciledValueDisplay.stateLine(field("")) == "Nothing recorded")
        #expect(ReconciledValueDisplay.stateLine(field("single", [value(1, "kept")])) == "1 Source")
        #expect(ReconciledValueDisplay.stateLine(field("merged", [value(1, "kept", support: 3)])) == "merged · 3 Sources")
        #expect(ReconciledValueDisplay.stateLine(field("mixed", [value(1, "kept"), value(2, "kept")])) == "mixed")
        #expect(ReconciledValueDisplay.stateLine(field("concluded", [value(1, "kept")])) == "concluded")
        // Support is of the displayed value, not of a dropped one ranked first.
        #expect(ReconciledValueDisplay.stateLine(field("merged", [value(1, "weak", support: 5), value(2, "kept", support: 2)])) == "merged · 2 Sources")
    }

    @Test func additionalCountsOnlyDisplayedValuesBeyondTheFirst() {
        let one = field("single", [value(1, "kept"), value(2, "outvoted")])
        #expect(ReconciledValueDisplay.additionalCount(one) == 0)
        #expect(ReconciledValueDisplay.additionalLabel(one) == nil)
        let three = field("mixed", [value(1, "kept"), value(2, "kept"), value(3, "weak"), value(4, "kept")])
        #expect(ReconciledValueDisplay.additionalCount(three) == 2)
        #expect(ReconciledValueDisplay.additionalLabel(three) == "+2")
        #expect(ReconciledValueDisplay.additionalCount(field("")) == 0)
    }

    @Test func valuesUseTheirOwnDisplays() {
        let en = Locale(identifier: "en_US")
        #expect(ReconciledValueDisplay.string(for: .none) == "")
        #expect(ReconciledValueDisplay.string(for: .text("  farmer ")) == "farmer")
        #expect(ReconciledValueDisplay.string(for: .integer(1890), locale: en) == "1890")
        #expect(ReconciledValueDisplay.string(for: .term(id: "t", key: "male", label: "Male")) == "Male")
        #expect(ReconciledValueDisplay.string(for: .term(id: "t", key: "male", label: "")) == "male")
        #expect(ReconciledValueDisplay.string(for: .name(jim)) == "James Robins")
        let date = CatalogDateValueInput(kind: "point", startYear: 1890)
        #expect(ReconciledValueDisplay.string(for: .date(date), locale: en) == DateValueDisplay.string(for: date, locale: en))
    }

    @Test func everyOutcomeReasonHasAPhrase() {
        let f = field("mixed", [value(1, "kept", .name(jim)), value(2, "kept")], outcomes: [
            outcome("a", "kept"),
            outcome("neg", "against", rank: nil),
        ])
        func phrase(_ o: CatalogReconcilerOutcome) -> String { ReconciledValueDisplay.outcomePhrase(o, in: f) }
        #expect(phrase(outcome("a", "kept")) == "kept")
        #expect(phrase(outcome("b", "folded", rank: 1)) == "folded into James Robins")
        #expect(phrase(outcome("b", "folded", rank: 9)) == "folded")
        #expect(phrase(outcome("c", "outvoted", rank: 3)) == "outvoted")
        #expect(phrase(outcome("d", "weak", rank: 3)) == "weak")
        #expect(phrase(outcome("e", "denied", rank: nil, deniedBy: "neg")) == "denied by OBS-neg")
        #expect(phrase(outcome("e", "denied", rank: nil, deniedBy: "gone")) == "denied")
        #expect(phrase(outcome("f", "provisional", rank: nil)) == "provisional member")
        #expect(phrase(outcome("g", "no_evidence", rank: nil)) == "no usable value")
        #expect(phrase(outcome("neg", "against", rank: nil)) == "counts against")
        #expect(phrase(outcome("h", "someday", rank: nil)) == "")
    }

    @Test func weakSaysWhichEvidenceIsWeak() {
        let f = field("single", [value(1, "kept")])
        func phrase(_ o: CatalogReconcilerOutcome) -> String { ReconciledValueDisplay.outcomePhrase(o, in: f) }
        #expect(phrase(outcome("a", "weak", credibilityOffset: -1)) == "weak · low-trust Source")
        #expect(phrase(outcome("a", "weak", uncertain: true)) == "weak · uncertain transcription")
        #expect(phrase(outcome("a", "weak", claimConfidenceOffset: -1)) == "weak · low-confidence claim")
        // Above-default grades are not causes.
        #expect(phrase(outcome("a", "weak", credibilityOffset: 1, uncertain: true, claimConfidenceOffset: 1)) == "weak · uncertain transcription")
        let all = outcome("a", "weak", credibilityOffset: -2, uncertain: true, claimConfidenceOffset: -1)
        #expect(ReconciledValueDisplay.weakCauses(all) == ["low-trust Source", "uncertain transcription", "low-confidence claim"])
        #expect(phrase(all) == "weak · " + ListFormatter.localizedString(byJoining: ReconciledValueDisplay.weakCauses(all)))
    }
}
