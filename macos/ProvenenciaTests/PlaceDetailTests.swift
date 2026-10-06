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
        members: Int = 1
    ) -> CatalogConclusionDetail {
        CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: ref, subjectTypeID: "t", label: label),
            fields: fields,
            memberCount: members
        )
    }

    private func content(_ detail: CatalogConclusionDetail) -> PlaceDetailContent {
        PlaceDetailContent(detail: detail, locale: en)
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
        #expect(page.sections.map(\.title) == ["Part of", "Contains", "Succession"])
        #expect(page.sections.map(\.aside) == [
            "Grouped by type · span is when the link holds",
            "Direct children only",
            "Renames and mergers · not part of the hierarchy",
        ])
        #expect(page.sections.map(\.emptyText) == [
            "Not part of any recorded place",
            "No places recorded as part of this one",
            "No predecessor or successor recorded",
        ])
    }

    @Test func refTitleHidesTheTrailingRef() {
        let page = content(detail())
        #expect(page.title == .ref("PLC-2MT40"))
        #expect(!page.showsRef)
        #expect(page.members == "1 member")
    }
}
