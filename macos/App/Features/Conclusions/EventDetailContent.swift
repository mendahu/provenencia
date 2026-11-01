import Foundation

/// The Event page, worded from one `CatalogConclusionDetail` and the Event
/// summary it carries (S9-24, board S9-D6). The summary supplies the title and
/// places. The date's Why stays on this page's date row.
struct EventDetailContent: ConclusionDetailBody {
    var title: ConclusionTitleSource
    var ref: String
    var showsRef: Bool { if case .ref = title { false } else { true } }
    var members: String
    /// Compact date under the title. Nil when none is recorded.
    var secondaryDate: String?
    /// First place name under the title. Nil when none is recorded.
    var secondaryPlace: String?
    var rows: [ReconciledValueRowModel]
    /// The pictures recorded for the Event, earliest first.
    var gallery: ConclusionDetailGallery
    /// Leads with the gallery's lead picture.
    var thumbnail: ConclusionDetailThumbnail

    init(
        detail: CatalogConclusionDetail,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let summary = detail.summary?.event
        title = summary.map { EventTitleDisplay.titleSource($0, locale: locale, orders: orders) }
            ?? EventTitleDisplay.untitled(detail.entity)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        let line = summary.map { EventSecondaryDisplay.content($0, locale: locale) }
        secondaryDate = line.flatMap { $0.date.isEmpty ? nil : $0.date }
        secondaryPlace = line.flatMap { $0.place.isEmpty ? nil : $0.place }
        rows = [Self.dateRow(detail.fields, locale: locale), Self.placeRow(summary)]
        gallery = .pictures(from: detail, fallbackMark: .subjectEvent, choose: .eventThumbnail)
        thumbnail = .leading(
            gallery: gallery,
            noImage: L10n.Conclusions.eventNoImage,
            withImage: L10n.Conclusions.eventPictureThumb,
            leadLabel: L10n.Conclusions.pictureFrom
        )
    }

    /// The Date row. A `date` with records is that field. Otherwise one row
    /// whose lead is the start–end span and whose Why lists those records.
    /// Stated empty when none of them is set.
    static func dateRow(_ fields: [CatalogConclusionField], locale: Locale) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.eventDate)
        if let date = fields.first(where: { $0.propertyKey == SeededPropertyKey.date }), !date.outcomes.isEmpty {
            return labeled(ReconciledValueRowModel(field: date, locale: locale), label, field: date)
        }
        let span = [SeededPropertyKey.startDate, SeededPropertyKey.endDate].compactMap { key in
            fields.first { $0.propertyKey == key && !$0.outcomes.isEmpty }
        }
        let empty = L10n.string(L10n.Conclusions.emptyEventDate)
        guard !span.isEmpty else {
            return .empty(id: "date", label: label, style: .date, emptyText: empty)
        }
        if span.count == 1, let only = span.first {
            return labeled(ReconciledValueRowModel(field: only, locale: locale), label, field: only)
        }
        let lead = DateRowDisplay.line(
            date: nil,
            start: dateValue(fields, key: SeededPropertyKey.startDate),
            end: dateValue(fields, key: SeededPropertyKey.endDate),
            locale: locale
        )
        return .spanning(id: "date", label: label, lead: lead, emptyText: empty, fields: span, locale: locale)
    }

    /// The Place row. The first kept name, mixed when more survived. No Why:
    /// that stays on the Place, which the row opens.
    static func placeRow(_ summary: CatalogEventSummary?) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.eventPlace)
        let empty = L10n.string(L10n.Conclusions.emptyEventPlace)
        let places = summary?.places ?? []
        return .derived(
            id: "place", label: label, style: .place, lead: DerivedPlace.name(places),
            emptyText: empty, mixed: DerivedPlace.extra(places) > 0,
            opens: places.first.map(ReconciledValueRowModel.Opens.place)
        )
    }

    private static func labeled(
        _ row: ReconciledValueRowModel,
        _ label: String,
        field: CatalogConclusionField
    ) -> ReconciledValueRowModel {
        var row = row
        row.label = label
        row.accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: label, lead: row.lead, field: field)
        return row
    }

    private static func dateValue(_ field: CatalogConclusionField) -> CatalogDateValueInput? {
        guard case .date(let date)? = field.displayedValues.first?.value else { return nil }
        return date
    }

    private static func dateValue(_ fields: [CatalogConclusionField], key: String) -> CatalogDateValueInput? {
        guard let field = fields.first(where: { $0.propertyKey == key }) else { return nil }
        return dateValue(field)
    }
}
