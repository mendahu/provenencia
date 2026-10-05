import Foundation
import Testing
@testable import Provenencia

struct PersonDetailTests {
    private typealias D = ReconciledValueDisplayTests

    private func detail(
        name: CatalogConclusionField? = D.mergedName,
        label: String = "",
        members: Int = 4,
        extra: [CatalogConclusionField] = []
    ) -> CatalogConclusionDetail {
        let sex = D.field("", key: "sex_at_birth", label: "Sex at birth", valueType: "term")
        return CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: label),
            fields: (name.map { [$0] } ?? []) + [sex] + extra,
            memberCount: members
        )
    }

    @Test func presentationFollowsTheHandle() {
        #expect(PersonDetailPresentation(value: nil, isFetching: true, status: .loading, error: nil) == .loading)
        let failed = PersonDetailPresentation(
            value: nil, isFetching: false, status: .error,
            error: CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: [])
        )
        #expect(failed == .failed("This record no longer exists. It may have been merged into another."))
        let content = PersonDetailContent(detail: detail())
        #expect(PersonDetailPresentation(value: detail(), isFetching: false, status: .ready, error: nil)
            == .content(content, refreshing: false))
        #expect(PersonDetailPresentation(value: detail(), isFetching: true, status: .ready, error: nil)
            == .content(content, refreshing: true))
    }

    /// Name and the life rows always show; any other field only once a
    /// record speaks to it.
    @Test func rowsAreNameThenLifeThenRecordedFields() {
        let weak = D.outcome("w", "weak", rank: 1, recorded: .integer(8))
        let content = PersonDetailContent(detail: detail(extra: [
            D.field("", key: "occupation", label: "Occupation", valueType: "text", outcomes: [D.outcome("o", "no_evidence", rank: nil)]),
            D.field("single", key: "birth_weight", label: "Birth weight", valueType: "integer", [D.value(1, "weak", .integer(8))], outcomes: [weak]),
            D.field("", key: "height", label: "Height", valueType: "integer"),
        ]))
        // Sex at birth and Height have no records, so they don't show.
        #expect(content.rows.map(\.label) == [
            "Name", "Birth date", "Birth place", "Death date", "Death place", "Occupation", "Birth weight",
        ])
        #expect(content.rows[1...4].map(\.style) == [.date, .place, .date, .place])
        #expect(content.rows[1...4].map(\.emptyText) == [
            "No birth date recorded", "No birth place recorded", "No death date recorded", "No death place recorded",
        ])
        #expect(content.rows[1...4].allSatisfy { $0.lead == nil && $0.records.isEmpty && $0.count == nil })
        // A recorded field with nothing to show says so, and keeps its Why.
        #expect(content.rows[5].lead == nil && content.rows[5].emptyText == "Nothing recorded" && content.rows[5].records.count == 1)
    }

    @Test func aMergedNameRowCarriesItsStateAndWhy() {
        let row = PersonDetailContent(detail: detail()).rows[0]
        #expect(row.lead == "James Robins" && row.style == .text)
        #expect(row.badge == .merged && row.count == "2 Sources" && row.against == nil)
        #expect(row.otherValuesLabel == nil && row.otherValues.isEmpty)
        #expect(row.whyTitle == "Why “James Robins”")
        #expect(row.records.map(\.phrase) == [
            "kept", "folded into James Robins", "outvoted (2 of 3 Sources)", "weak · low-trust Source",
        ])
        #expect(row.records.map(\.sourceTitle).first == "Census of Canada West, 1851")
        #expect(row.accessibilityLabel == "Name, James Robins, merged from 2 Sources")
    }

    @Test func aMixedRowDisclosesItsOtherValues() {
        let date = CatalogDateValueInput(kind: "point", startYear: 1878)
        let mixed = D.field("mixed", key: "x", label: "Occupation", valueType: "text", [
            D.value(1, "kept", support: 2, .text("farmer")), D.value(2, "kept", .text("labourer")),
        ])
        let row = ReconciledValueRowModel(field: mixed)
        #expect(row.lead == "farmer" && row.badge == .mixed)
        #expect(row.otherValuesLabel == "1 other value")
        #expect(row.otherValues == [.init(rank: 2, text: "labourer", support: "1 Source")])
        let dated = ReconciledValueRowModel(field: D.field("single", key: "d", label: "Date", valueType: "date", [D.value(1, "kept", .date(date))]))
        #expect(dated.style == .date)
    }

    @Test func headerFallsBackNameLabelRef() {
        let named = PersonDetailContent(detail: detail())
        #expect(named.title == .name("James Robins") && named.showsRef && named.members == "4 members")
        let labelled = PersonDetailContent(detail: detail(name: D.field("", [], outcomes: []), label: "Grandpa", members: 1))
        #expect(labelled.title == .label("Grandpa") && labelled.showsRef && labelled.members == "1 member")
        let bare = PersonDetailContent(detail: detail(name: nil, members: 1))
        #expect(bare.title == .ref("PER-7KD45") && !bare.showsRef)
        // No name field at all: the Name row still shows, empty.
        #expect(bare.rows.first?.label == "Name" && bare.rows.first?.lead == nil)
        #expect(bare.rows.first?.emptyText == "No name recorded")
        #expect(bare.rows.first?.accessibilityLabel == "Name, empty")
    }

    @Test func vitalsSayUnknownUntilLifeEventsArrive() {
        let vitals = PersonDetailContent(detail: detail()).vitals
        #expect(vitals.map(\.abbreviation) == ["b.", "d."])
        #expect(vitals.allSatisfy { $0.date == nil && $0.place == nil })
        #expect(vitals.map(\.accessibilityLabel) == ["Born date unknown, place unknown", "Died date unknown, place unknown"])
    }

    @Test func aSourceOpensTheRecordsCitationInTheComposer() throws {
        let outcome = D.outcome("bible", "outvoted", rank: 2, source: "s-bible", title: "Family Bible")
        let location = ReconciliationRecord.composerLocation(for: outcome)
        #expect(location.section == .sources && location.sourceSurface == .citationComposer)
        let entry = try #require(CitationComposerEntry(location: location))
        guard case let .edit(sourceID, subjectID, citationID, artifactID, observationID) = entry else {
            Issue.record("expected an edit entry, got \(entry)")
            return
        }
        #expect(sourceID == "s-bible" && subjectID == "sub-bible" && citationID == "cit-bible")
        #expect(artifactID == "art-bible" && observationID == "bible")
    }
}
