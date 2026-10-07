import Foundation
import Testing
@testable import Provenencia

struct EventDetailTests {
    private typealias D = ReconciledValueDisplayTests
    private let en = Locale(identifier: "en_US")

    private func detail(
        label: String = "",
        ref: String = "EVT-8PL22",
        fields: [CatalogConclusionField] = [],
        members: Int = 4
    ) -> CatalogConclusionDetail {
        CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: ref, subjectTypeID: "t", label: label),
            fields: fields,
            memberCount: members
        )
    }

    private func content(_ detail: CatalogConclusionDetail) -> EventDetailContent {
        EventDetailContent(detail: detail, locale: en)
    }

    private func dateField(
        _ key: String,
        _ date: CatalogDateValueInput,
        id: String = "rec"
    ) -> CatalogConclusionField {
        D.field(
            "single", key: key, label: key, valueType: "date",
            [D.value(1, "kept", .date(date))],
            outcomes: [D.outcome(id, "kept", title: "Census", recorded: .date(date))]
        )
    }

    @Test func recordedNameIsAPlainTitle() {
        let page = content(detail(fields: [
            D.field("single", key: "event_name", label: "Event name", valueType: "text", [D.value(1, "kept", .text("Fire at York"))]),
        ]))
        #expect(page.title == .name("Fire at York"))
        #expect(page.showsRef)
        #expect(page.ref == "EVT-8PL22")
    }

    @Test func typeOnlyBecomesUnspecified() {
        let type = D.field(
            "single", key: "event_type", label: "Event type", valueType: "term",
            [D.value(1, "kept", .term(id: "t", key: "birth", label: "Birth"))]
        )
        #expect(content(detail(fields: [type])).title == .name("Unspecified birth"))
    }

    @Test func refTitleHidesTheTrailingRef() {
        let page = content(detail())
        #expect(page.title == .ref("EVT-8PL22"))
        #expect(!page.showsRef)
    }

    @Test func dateFieldWinsOverASpan() {
        let point = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let page = content(detail(fields: [
            dateField("date", point, id: "point"),
            dateField("start_date", CatalogDateValueInput(kind: "point", startYear: 1849), id: "start"),
            dateField("end_date", CatalogDateValueInput(kind: "point", startYear: 1851), id: "end"),
        ]))
        #expect(page.rows.map(\.label) == ["Date", "Place"])
        #expect(page.rows[0].lead == "14 May 1817")
        #expect(page.rows[0].records.map(\.id) == ["point"])
        #expect(page.summaryDate == "14 May 1817")
    }

    @Test func spanLeadWhenThereIsNoDate() {
        let page = content(detail(fields: [
            dateField("start_date", CatalogDateValueInput(kind: "point", startYear: 1849), id: "start"),
            dateField("end_date", CatalogDateValueInput(kind: "point", startYear: 1851), id: "end"),
        ]))
        #expect(page.rows[0].lead == L10n.Dates.rowSpan(start: "1849", end: "1851", locale: en))
        #expect(page.rows[0].records.map(\.id) == ["start", "end"])
        #expect(page.rows[0].badge == nil)
        #expect(page.summaryDate == page.rows[0].lead)
    }

    @Test func headerSuppliesSubjectsAndPlace() {
        let james = CatalogCanonicalEntity(id: "p1", ref: "PER-1", subjectTypeID: "t", label: "")
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let toronto = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        let header = CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-8PL22", subjectTypeID: "t", label: ""),
            eventTypeKey: "birth",
            eventTypeLabel: "Birth",
            subjects: [
                CatalogEventSubject(
                    entity: james,
                    name: CatalogNameValue(form: "James Robins"),
                    nameValueCount: 1
                ),
            ],
            places: [
                CatalogHeaderPlace(entity: york, names: ["York"], nameCount: 1),
                CatalogHeaderPlace(entity: toronto, names: ["Toronto"], nameCount: 1),
            ]
        )
        let page = EventDetailContent(detail: detail(), header: header, locale: en)
        #expect(page.title == .name(L10n.EventTitle.ofOne(type: "Birth", subject: "James Robins", locale: en)))
        #expect(page.summaryPlace == "York")
        #expect(page.rows[1].lead == "York" && page.rows[1].badge == .mixed && page.rows[1].records.isEmpty)
        #expect(!page.rows[1].lead!.contains(","))
        #expect(page.rows[1].opens?.location == .placeDetail(entityId: "pl1", ref: "PLC-1", title: "York"))
    }

    @Test func placeRowIsEmpty() {
        let page = content(detail(fields: [
            dateField("date", CatalogDateValueInput(kind: "point", startYear: 1810), id: "d"),
        ]))
        let place = page.rows[1]
        #expect(place.label == "Place" && place.style == .place && place.lead == nil)
        #expect(place.emptyText == "No place recorded")
        #expect(place.records.isEmpty)
    }

    @Test func missingDateIsAnEmptyDateRow() {
        let page = content(detail())
        let date = page.rows[0]
        #expect(date.label == "Date" && date.style == .date && date.lead == nil)
        #expect(date.emptyText == "No date recorded")
        #expect(date.records.isEmpty)
        #expect(page.summaryDate == nil)
    }
}
