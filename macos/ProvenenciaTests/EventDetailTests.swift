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
        members: Int = 4,
        summary: CatalogEventSummary? = nil
    ) -> CatalogConclusionDetail {
        CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: ref, subjectTypeID: "t", label: label),
            fields: fields,
            memberCount: members,
            summary: summary.map { .event($0) }
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

    private func dated(
        date: CatalogDateValueInput? = nil,
        start: CatalogDateValueInput? = nil,
        end: CatalogDateValueInput? = nil
    ) -> CatalogEventSummary {
        var summary = CatalogEventSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-8PL22", subjectTypeID: "t", label: ""),
            title: CatalogEventTitle(rule: .ref, ref: "EVT-8PL22")
        )
        summary.date = date
        summary.startDate = start
        summary.endDate = end
        return summary
    }

    private func summary(_ title: CatalogEventTitle) -> CatalogEventSummary {
        CatalogEventSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-8PL22", subjectTypeID: "t", label: ""),
            title: title
        )
    }

    @Test func titleIsTheSummaryChosenTitle() {
        let page = EventDetailContent(
            detail: detail(summary: summary(CatalogEventTitle(rule: .recordedName, recordedName: "Fire at York", ref: "EVT-8PL22"))),
            locale: en
        )
        #expect(page.title == .name("Fire at York"))
        #expect(page.showsRef)
        #expect(page.ref == "EVT-8PL22")
    }

    @Test func withNoSummaryTheTitleIsTheLabelThenTheRef() {
        // A recorded name in the detail does not choose a title; only Go's summary does.
        let named = content(detail(label: "Grandpa's fire", fields: [
            D.field("single", key: "event_name", label: "Event name", valueType: "text", [D.value(1, "kept", .text("Fire at York"))]),
        ]))
        #expect(named.title == .label("Grandpa's fire"))
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
        ], summary: dated(date: point)))
        #expect(page.rows.map(\.label) == ["Date", "Place"])
        #expect(page.rows[0].lead == "14 May 1817")
        #expect(page.rows[0].records.map(\.id) == ["point"])
        #expect(page.secondaryDate == "14 May 1817")
    }

    @Test func spanLeadWhenThereIsNoDate() {
        let start = CatalogDateValueInput(kind: "point", startYear: 1849)
        let end = CatalogDateValueInput(kind: "point", startYear: 1851)
        let page = content(detail(fields: [
            dateField("start_date", start, id: "start"),
            dateField("end_date", end, id: "end"),
        ], summary: dated(start: start, end: end)))
        #expect(page.rows[0].lead == L10n.Dates.rowSpan(start: "1849", end: "1851", locale: en))
        #expect(page.rows[0].records.map(\.id) == ["start", "end"])
        #expect(page.rows[0].badge == nil)
        #expect(page.secondaryDate == page.rows[0].lead)
    }

    @Test func bandLineReadsTheSummary() {
        let point = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let withDate = content(detail(summary: dated(date: point)))
        #expect(withDate.secondaryDate == "14 May 1817")
        #expect(withDate.secondaryDate == EventSecondaryDisplay.content(dated(date: point), locale: en).date)

        // A `date` field with records but no displayed value: the band shows the kept span, as the list row does.
        let start = CatalogDateValueInput(kind: "point", startYear: 1849)
        let end = CatalogDateValueInput(kind: "point", startYear: 1851)
        let span = dated(start: start, end: end)
        let spanned = content(detail(
            fields: [D.field("single", key: SeededPropertyKey.date, label: "Date", valueType: "date", [])],
            summary: span
        ))
        #expect(spanned.secondaryDate == EventSecondaryDisplay.content(span, locale: en).date)
        #expect(spanned.secondaryDate == L10n.Dates.rowSpan(start: "1849", end: "1851", locale: en))
    }

    @Test func summarySuppliesSubjectsAndPlace() {
        let james = CatalogCanonicalEntity(id: "p1", ref: "PER-1", subjectTypeID: "t", label: "")
        let york = CatalogCanonicalEntity(id: "pl1", ref: "PLC-1", subjectTypeID: "t", label: "")
        let toronto = CatalogCanonicalEntity(id: "pl2", ref: "PLC-2", subjectTypeID: "t", label: "")
        let summary = CatalogEventSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-8PL22", subjectTypeID: "t", label: ""),
            eventTypeKey: "birth",
            eventTypeOrigin: CatalogOrigin.provenencia,
            eventTypeLabel: "Birth",
            subjects: [
                CatalogEventSubject(
                    entity: james,
                    name: CatalogNameValue(form: "James Robins"),
                    nameValueCount: 1
                ),
            ],
            places: [
                CatalogChainedPlace(entity: york, names: ["York"]),
                CatalogChainedPlace(entity: toronto, names: ["Toronto"]),
            ],
            title: CatalogEventTitle(
                rule: .subject, ref: "EVT-8PL22", typeKey: "birth", typeOrigin: CatalogOrigin.provenencia, typeLabel: "Birth",
                subjects: [CatalogNameValue(form: "James Robins")], place: "York"
            )
        )
        let page = EventDetailContent(detail: detail(summary: summary), locale: en)
        #expect(page.title == .name(L10n.EventTitle.ofOne(type: "Birth", subject: "James Robins", locale: en)))
        #expect(page.secondaryPlace == "York")
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
        let page = content(detail(summary: dated()))
        let date = page.rows[0]
        #expect(date.label == "Date" && date.style == .date && date.lead == nil)
        #expect(date.emptyText == "No date recorded")
        #expect(date.records.isEmpty)
        #expect(page.secondaryDate == nil)
    }

    @Test func noPictureLeavesTheMarkAndNoGallery() {
        let page = content(detail())
        #expect(page.gallery.tiles.isEmpty)
        #expect(!page.gallery.offersEarliest)
        #expect(page.thumbnail == .init(mark: .subjectEvent, label: L10n.Conclusions.eventNoImage))
    }

    @Test func aSinglePictureLeadsTheHeaderWithoutAMenu() {
        var source = detail()
        source.pictures = [D.picture("a", path: "objects/ab/cd.png", title: "Wedding photograph", leads: true)]
        let page = content(source)
        #expect(page.thumbnail.imageRelPath == "objects/ab/cd.png")
        #expect(page.thumbnail.mark == .subjectEvent)
        #expect(page.thumbnail.label == L10n.Conclusions.eventPictureThumb)
        #expect(page.thumbnail.imageLabel == "Picture from Wedding photograph")
        #expect(!page.gallery.offersEarliest)
    }

    @Test func aChosenLaterPictureOffersTheEarliest() {
        var source = detail()
        source.pictures = [
            D.picture("a", path: "objects/ab/cd.png", title: "Wedding photograph", leads: false),
            D.picture("b", path: "objects/ef/01.png", title: "Family album", leads: true),
        ]
        let gallery = content(source).gallery
        #expect(gallery.title == L10n.Conclusions.picturesSection)
        #expect(gallery.tiles.map(\.id) == ["a", "b"])
        #expect(gallery.tiles.map(\.isLead) == [false, true])
        #expect(gallery.offersEarliest)
        #expect(gallery.choose == .eventThumbnail)
        #expect(gallery.tiles[0].useAsLabel == "Use the picture from Wedding photograph as thumbnail")
        #expect(gallery.tiles[1].useAsLabel == "Use the picture from Family album as thumbnail")
        #expect(gallery.tiles[1].chosenToastBody == "Using the picture from Family album as the thumbnail.")
        #expect(gallery.tiles[0].accessibilityLabel == "Picture from Wedding photograph. Open citation")
    }

    @Test func twoTilesSharingAPathNameTheLeadingSource() {
        var source = detail()
        source.pictures = [
            D.picture("a", path: "objects/ab/cd.png", title: "First copy", leads: false),
            D.picture("b", path: "objects/ab/cd.png", title: "Second copy", leads: true),
        ]
        let page = content(source)
        #expect(page.thumbnail.imageLabel == "Picture from Second copy")
        #expect(page.gallery.offersEarliest)
    }
}
