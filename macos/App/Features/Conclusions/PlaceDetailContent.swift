import Foundation

/// The Place page, worded from one `CatalogConclusionDetail` (S9-27 / S9-40,
/// board S9-D7). Pure: views only lay it out.
struct PlaceDetailContent: ConclusionDetailBody {
    /// Soft cap for Contains before showing a remainder count.
    static let containsVisibleLimit = 12

    var title: ConclusionTitleSource
    var ref: String
    /// The ref is repeated at the right only when it isn't the title.
    var showsRef: Bool { if case .ref = title { false } else { true } }
    var members: String
    /// Joined parent chain, or the sentence that none is recorded.
    var chain: String
    var chainIsRecorded: Bool
    var rows: [ReconciledValueRowModel]
    var sections: [ConclusionDetailSection]

    init(detail: CatalogConclusionDetail, locale: Locale = .autoupdatingCurrent) {
        let toponym = detail.fields.first { $0.propertyKey == SeededPropertyKey.toponym }
        let names = toponym?.displayedValues.compactMap { value -> String? in
            let text = ReconciledValueDisplay.string(for: value.value, locale: locale)
            return text.isEmpty ? nil : text
        } ?? []
        title = PlaceTitleDisplay.titleSource(
            PlaceTitleParts(names: names, label: detail.entity.label, ref: detail.entity.ref)
        )
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        let header = detail.placeHeader
        let line = PlaceChainDisplay.line(
            parents: header?.parents ?? [],
            candidates: header?.parentsAreCandidates ?? false
        )
        chainIsRecorded = !line.isEmpty
        chain = chainIsRecorded ? line : L10n.string(L10n.Conclusions.placeNoParent)
        rows = [
            Self.namesRow(toponym, locale: locale),
            Self.periodRow(detail: detail, header: header, locale: locale),
        ]
        sections = Self.relationshipSections(header: header)
    }

    /// The Names row is the `toponym` field, relabeled. Several kept names
    /// are all listed. A Place with no toponym states that none is recorded.
    static func namesRow(_ field: CatalogConclusionField?, locale: Locale) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.placeNames)
        let empty = L10n.string(L10n.Conclusions.emptyPlaceNames)
        guard let field, !field.outcomes.isEmpty else {
            return .empty(id: "names", label: label, style: .text, emptyText: empty)
        }
        var row = ReconciledValueRowModel(field: field, locale: locale)
        row.label = label
        row.emptyText = empty
        guard row.listedValues.count > 1 else {
            row.accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(
                label: label, lead: row.lead, field: field
            )
            return row
        }
        let names = row.listedValues.map(\.text).joined(separator: "; ")
        row.whyTitle = L10n.Conclusions.whyThese(row.listedValues.count, label: label)
        row.accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(
            label: label, lead: names, field: field
        )
        return row
    }

    static func periodRow(
        detail: CatalogConclusionDetail,
        header: CatalogPlaceHeader?,
        locale: Locale
    ) -> ReconciledValueRowModel {
        let label = L10n.string(L10n.Conclusions.placePeriod)
        let empty = L10n.string(L10n.Conclusions.emptyPlacePeriod)
        let lead = PlacePeriodDisplay.span(
            start: header?.startDate,
            end: header?.endDate,
            locale: locale
        )
        let fields = detail.fields.filter {
            $0.propertyKey == SeededPropertyKey.startDate || $0.propertyKey == SeededPropertyKey.endDate
        }
        return .spanning(
            id: "period",
            label: label,
            lead: lead.isEmpty ? nil : lead,
            emptyText: empty,
            fields: fields,
            locale: locale
        )
    }

    static func relationshipSections(header: CatalogPlaceHeader?) -> [ConclusionDetailSection] {
        let partOf = header?.partOf ?? []
        let contains = header?.contains ?? []
        let predecessors = header?.predecessors ?? []
        let successors = header?.successors ?? []
        let visibleContains = Array(contains.prefix(containsVisibleLimit))
        let omitted = max(0, contains.count - visibleContains.count)

        var successionRows: [ConclusionDetailRelationship] = []
        for r in predecessors {
            successionRows.append(ConclusionDetailRelationship(
                relationship: r,
                role: L10n.string(L10n.Conclusions.placeSucceeded)
            ))
        }
        for r in successors {
            successionRows.append(ConclusionDetailRelationship(
                relationship: r,
                role: L10n.string(L10n.Conclusions.placeSucceededBy)
            ))
        }

        return [
            ConclusionDetailSection(
                id: "partOf",
                title: L10n.Conclusions.placePartOf,
                aside: L10n.string(L10n.Conclusions.placePartOfAside),
                emptyText: L10n.string(L10n.Conclusions.placePartOfEmpty),
                relationships: partOf.map { ConclusionDetailRelationship(relationship: $0) }
            ),
            ConclusionDetailSection(
                id: "contains",
                title: L10n.Conclusions.placeContains,
                aside: L10n.string(L10n.Conclusions.placeContainsAside),
                emptyText: L10n.string(L10n.Conclusions.placeContainsEmpty),
                relationships: visibleContains.map { ConclusionDetailRelationship(relationship: $0) },
                omittedCount: omitted
            ),
            ConclusionDetailSection(
                id: "succession",
                title: L10n.Conclusions.placeSuccession,
                aside: L10n.string(L10n.Conclusions.placeSuccessionAside),
                emptyText: L10n.string(L10n.Conclusions.placeSuccessionEmpty),
                relationships: successionRows
            ),
        ]
    }
}
