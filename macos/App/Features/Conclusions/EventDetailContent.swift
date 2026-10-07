import Foundation

/// The Event page, worded from one `CatalogConclusionDetail` and, when the
/// page has it, the Event header (S9-24, board S9-D6). The header supplies
/// the subjects and places. The date's Why stays on this page's date row.
struct EventDetailContent: ConclusionDetailBody {
    var title: ConclusionTitleSource
    var ref: String
    var showsRef: Bool { if case .ref = title { false } else { true } }
    var members: String
    /// Compact date under the title. Nil when none is recorded.
    var summaryDate: String?
    /// First place name under the title. Nil when none is recorded.
    var summaryPlace: String?
    var rows: [ReconciledValueRowModel]

    init(
        detail: CatalogConclusionDetail,
        header: CatalogEventHeader? = nil,
        locale: Locale = .autoupdatingCurrent
    ) {
        let parts = header?.titleParts ?? Self.parts(detail)
        title = EventTitleDisplay.titleSource(parts, locale: locale)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        summaryDate = Self.summaryDate(detail.fields, locale: locale)
        let place = header.map { DerivedPlace.name($0.places) } ?? ""
        summaryPlace = place.isEmpty ? nil : place
        rows = [Self.dateRow(detail.fields, locale: locale), Self.placeRow(header)]
    }

    private static func parts(_ detail: CatalogConclusionDetail) -> EventTitleParts {
        let type = term(detail.fields, propertyKey: "event_type")
        return EventTitleParts(
            recordedName: text(detail.fields, key: "event_name"),
            label: detail.entity.label,
            ref: detail.entity.ref,
            typeKey: type.key,
            typeLabel: type.label
        )
    }

    /// A recorded `date` wins. Otherwise the start–end span. Empty when
    /// none of them is displayed.
    static func summaryDate(_ fields: [CatalogConclusionField], locale: Locale) -> String? {
        if let date = fields.first(where: { $0.propertyKey == "date" }), !date.outcomes.isEmpty {
            guard let value = dateValue(date) else { return nil }
            let line = EventTitleDisplay.dateLine(date: value, locale: locale)
            return line.isEmpty ? nil : line
        }
        let line = EventTitleDisplay.dateLine(
            date: nil,
            start: dateValue(fields, key: "start_date"),
            end: dateValue(fields, key: "end_date"),
            locale: locale
        )
        return line.isEmpty ? nil : line
    }

    /// The Date row. A `date` with records is that field. Otherwise one row
    /// whose lead is the start–end span and whose Why lists those records.
    /// Stated empty when none of them is set.
    static func dateRow(_ fields: [CatalogConclusionField], locale: Locale) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.eventDate)
        if let date = fields.first(where: { $0.propertyKey == "date" }), !date.outcomes.isEmpty {
            return labeled(ReconciledValueRowModel(field: date, locale: locale), label, field: date)
        }
        let span = ["start_date", "end_date"].compactMap { key in
            fields.first { $0.propertyKey == key && !$0.outcomes.isEmpty }
        }
        let empty = L10n.string(L10n.Conclusions.emptyEventDate)
        guard !span.isEmpty else {
            return .empty(id: "date", label: label, style: .date, emptyText: empty)
        }
        if span.count == 1, let only = span.first {
            return labeled(ReconciledValueRowModel(field: only, locale: locale), label, field: only)
        }
        let lead = EventTitleDisplay.dateLine(
            date: nil,
            start: dateValue(fields, key: "start_date"),
            end: dateValue(fields, key: "end_date"),
            locale: locale
        )
        return .spanning(id: "date", label: label, lead: lead, emptyText: empty, fields: span, locale: locale)
    }

    /// The Place row. The first kept name, mixed when more survived. No Why:
    /// that stays on the Place, which the row opens.
    static func placeRow(_ header: CatalogEventHeader?) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.eventPlace)
        let empty = L10n.string(L10n.Conclusions.emptyEventPlace)
        let places = header?.places ?? []
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

    private static func text(_ fields: [CatalogConclusionField], key: String) -> String {
        guard let field = fields.first(where: { $0.propertyKey == key }),
              case .text(let text)? = field.displayedValues.first?.value
        else { return "" }
        return text
    }

    private static func term(_ fields: [CatalogConclusionField], propertyKey: String) -> (key: String, label: String) {
        guard let field = fields.first(where: { $0.propertyKey == propertyKey }),
              case .term(_, let key, let label)? = field.displayedValues.first?.value
        else { return ("", "") }
        return (key, label)
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
