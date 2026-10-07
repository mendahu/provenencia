import Foundation
import Testing
@testable import Provenencia

struct PlaceDetailTests {
    private typealias D = ReconciledValueDisplayTests
    private let en = Locale(identifier: "en_US")

    private func detail(
        label: String = "",
        ref: String = "PLC-2MT40",
        fields: [CatalogConclusionField] = [],
        members: Int = 1,
        placeHeader: CatalogPlaceHeader? = nil
    ) -> CatalogConclusionDetail {
        CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: ref, subjectTypeID: "t", label: label),
            fields: fields,
            memberCount: members,
            header: placeHeader.map { .place($0) }
        )
    }

    private func content(_ detail: CatalogConclusionDetail) -> PlaceDetailContent {
        PlaceDetailContent(detail: detail, locale: en)
    }

    private func rel(
        id: String,
        title: String,
        kind: String,
        start: Int? = nil,
        end: Int? = nil
    ) -> CatalogPlaceRelationship {
        CatalogPlaceRelationship(
            entity: CatalogCanonicalEntity(id: id, ref: "PLC-\(id)", subjectTypeID: "t", label: ""),
            title: title,
            kind: kind,
            startDate: start.map { CatalogDateValueInput(kind: "point", startYear: Int32($0)) },
            endDate: end.map { CatalogDateValueInput(kind: "point", startYear: Int32($0)) }
        )
    }

    /// Two kept toponyms are both listed, each with its Source count. A weak
    /// spelling stays in the Why and is not a displayed name.
    @Test func keptNamesAreAllListedAndAWeakSpellingStaysInTheWhy() {
        let field = D.field(
            "multiple", key: "toponym", label: "Toponym", valueType: "text",
            [
                D.value(1, "kept", support: 3, .text("Toronto")),
                D.value(2, "kept", support: 1, .text("Tkaronto")),
                D.value(3, "weak", .text("Torento")),
            ],
            outcomes: [
                D.outcome("census", "kept", source: "s-census", recorded: .text("Toronto")),
                D.outcome("map", "kept", rank: 2, source: "s-map", recorded: .text("Tkaronto")),
                D.outcome("letter", "weak", rank: 3, source: "s-letter", title: "Letter, 1793",
                          recorded: .text("Torento"), credibilityOffset: -1),
            ]
        )
        let page = content(detail(label: "York", fields: [field]))
        let names = page.rows[0]
        #expect(page.title == .name("Toronto"))
        #expect(names.label == "Names")
        #expect(names.listedValues.map(\.text) == ["Toronto", "Tkaronto"])
        #expect(names.listedValues.map(\.support) == ["3 Sources", "1 Source"])
        #expect(names.badge == nil)
        #expect(names.otherValuesLabel == nil)
        #expect(names.whyTitle == "Why these 2 names")
        #expect(names.accessibilityLabel == "Names, Toronto; Tkaronto, 2 values from 2 Sources")
        #expect(names.records.map(\.id).contains("letter"))
        #expect(!names.listedValues.map(\.text).contains("Torento"))
        #expect(names.records.first { $0.id == "letter" }?.phrase.contains("weak") == true)
    }

    @Test func periodAndRelationshipsAreStatedEmpty() {
        let page = content(detail(label: "York"))
        #expect(page.rows.map(\.label) == ["Names", "Period"])
        #expect(page.rows[0].lead == nil)
        #expect(page.rows[0].emptyText == "No names recorded")
        #expect(page.rows[1].lead == nil)
        #expect(page.rows[1].emptyText == "No period recorded")
        #expect(page.chain == "No parent place recorded")
        #expect(!page.chainIsRecorded)
        #expect(page.sections.map { L10n.string($0.title) } == ["Part of", "Contains", "Succession"])
        #expect(page.sections.map(\.aside) == [
            "Span is when the link holds",
            "Direct children only",
            "Renames and mergers · not part of the hierarchy",
        ])
        #expect(page.sections.map(\.emptyText) == [
            "Not part of any recorded place",
            "No places recorded as part of this one",
            "No predecessor or successor recorded",
        ])
        #expect(page.sections.map(\.relationships) == [[], [], []])
    }

    @Test func periodAndRelationshipsFillFromTheHeader() {
        let header = CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PLC-2MT40", subjectTypeID: "t", label: ""),
            names: ["Toronto"],
            startDate: CatalogDateValueInput(kind: "point", startYear: 1834 as Int32),
            parents: ["Ontario", "Canada"],
            partOf: [
                rel(id: "uc", title: "Upper Canada", kind: "part_of", start: 1791, end: 1841),
                rel(id: "on", title: "Ontario", kind: "part_of", start: 1867),
            ],
            contains: (1...14).map { rel(id: "c\($0)", title: "Ward \($0)", kind: "contains") },
            predecessors: [rel(id: "york", title: "York", kind: "predecessor", start: 1793, end: 1834)],
            successors: []
        )
        let page = content(detail(label: "Toronto", placeHeader: header))
        #expect(page.chain == "Ontario, Canada")
        #expect(page.chainIsRecorded)
        #expect(page.rows[1].lead == "1834 –")
        #expect(page.sections[0].relationships.map(\.relationship.title) == ["Upper Canada", "Ontario"])
        #expect(page.sections[0].relationships.map { PlacePeriodDisplay.span(
            start: $0.relationship.startDate, end: $0.relationship.endDate, locale: en
        ) } == ["1791 – 1841", "1867 –"])
        #expect(page.sections[1].relationships.count == PlaceDetailContent.containsVisibleLimit)
        #expect(page.sections[1].omittedCount == 2)
        #expect(page.sections[2].relationships.map(\.relationship.title) == ["York"])
        #expect(page.sections[2].relationships.map(\.role) == ["Succeeded"])
    }

    @Test func refTitleHidesTheTrailingRef() {
        let page = content(detail())
        #expect(page.title == .ref("PLC-2MT40"))
        #expect(!page.showsRef)
        #expect(page.members == "1 member")
    }
}
