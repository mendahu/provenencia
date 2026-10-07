import Foundation

/// The Place page, worded from one `CatalogConclusionDetail` (S9-27, board
/// S9-D7 frame 1f). Pure: views only lay it out. Period, Part of, Contains,
/// and Succession stay stated empty until S9-40.
struct PlaceDetailContent: ConclusionDetailBody {
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
        let line = PlaceChainDisplay.line(parents: [])
        chainIsRecorded = !line.isEmpty
        chain = chainIsRecorded ? line : L10n.string(L10n.Conclusions.placeNoParent)
        rows = [
            Self.namesRow(toponym, locale: locale),
            .empty(
                id: "period",
                label: L10n.string(L10n.Conclusions.placePeriod),
                style: .date,
                emptyText: L10n.string(L10n.Conclusions.emptyPlacePeriod)
            ),
        ]
        sections = Self.relationshipSections
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

    /// Part of, Contains, and Succession, each a header and an empty sentence.
    static let relationshipSections: [ConclusionDetailSection] = [
        ConclusionDetailSection(
            id: "partOf",
            title: L10n.Conclusions.placePartOf,
            aside: L10n.string(L10n.Conclusions.placePartOfAside),
            emptyText: L10n.string(L10n.Conclusions.placePartOfEmpty)
        ),
        ConclusionDetailSection(
            id: "contains",
            title: L10n.Conclusions.placeContains,
            aside: L10n.string(L10n.Conclusions.placeContainsAside),
            emptyText: L10n.string(L10n.Conclusions.placeContainsEmpty)
        ),
        ConclusionDetailSection(
            id: "succession",
            title: L10n.Conclusions.placeSuccession,
            aside: L10n.string(L10n.Conclusions.placeSuccessionAside),
            emptyText: L10n.string(L10n.Conclusions.placeSuccessionEmpty)
        ),
    ]
}
