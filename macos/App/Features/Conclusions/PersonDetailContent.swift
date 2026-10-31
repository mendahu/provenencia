import Foundation

/// The Person page, worded from one `CatalogConclusionDetail` (S9-16, board
/// S9-D5). Pure: views only lay it out.
struct PersonDetailContent: ConclusionDetailBody {
    /// One header line: *b. 14 May 1817 · York*. Date and place come from
    /// the Person header. Nil when that half is not recorded.
    struct LifeLine: Equatable {
        var abbreviation: String
        var date: String?
        var place: String?
        var accessibilityLabel: String
    }

    /// Name → label → ref (PD-1).
    var title: PersonHeaderDisplay.TitleSource
    var ref: String
    /// The ref is repeated at the right only when it isn't the title.
    var showsRef: Bool { if case .ref = title { false } else { true } }
    var members: String
    var lifeLines: [LifeLine]
    var rows: [ReconciledValueRowModel]
    /// Nil when gender and sex at birth give nothing to show (S10-23).
    var treatment: PersonGenderTreatment?
    /// The gallery of likenesses, earliest recorded first.
    var gallery: ConclusionDetailGallery
    /// Leads with the gallery's lead likeness.
    var thumbnail: ConclusionDetailThumbnail
    var badge: ConclusionDetailBadge? { treatment?.badge }

    init(
        detail: CatalogConclusionDetail,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let header = detail.header?.person
        let nameField = detail.fields.first { $0.propertyKey == SeededPropertyKey.name }
        var leadName: CatalogNameValue?
        if case .name(let name)? = nameField?.displayedValues.first?.value { leadName = name }
        title = PersonHeaderDisplay.titleSource(name: leadName, entity: detail.entity, style: .natural, orders: orders)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        lifeLines = [
            Self.lifeLine(L10n.Conclusions.personBornAbbr, spoken: L10n.Conclusions.a11yBorn, facts: header?.birth, locale: locale),
            Self.lifeLine(L10n.Conclusions.personDiedAbbr, spoken: L10n.Conclusions.a11yDied, facts: header?.death, locale: locale),
        ]
        rows = Self.rows(detail.fields, header: header, derivedRows: detail.derivedRows, locale: locale, orders: orders)
        treatment = PersonGenderTreatment.read(detail.fields)
        gallery = ConclusionDetailGallery(
            title: L10n.Conclusions.likenessesSection,
            openTip: L10n.Conclusions.likenessOpenTip,
            fallbackMark: .subjectPerson,
            tiles: ConclusionDetailGallery.tiles(
                from: detail,
                tileLabel: L10n.Conclusions.likenessTileLabel,
                useAsLabel: L10n.Conclusions.useAsPortraitLabel,
                chosenBody: L10n.Conclusions.toastPortraitChosenBody
            ),
            choose: .personPortrait
        )
        thumbnail = .leading(
            gallery: gallery,
            noImage: L10n.Conclusions.personNoLikeness,
            withImage: L10n.Conclusions.personLikenessThumb,
            leadLabel: L10n.Conclusions.likenessLabel,
            colors: treatment?.thumbnailColors
        )
    }

    /// A Person always shows its name and four life rows, stated empty when
    /// nothing is recorded. A life row is mixed when its value disagrees (two
    /// kept dates, two Places) or when more than one birth (or death) event
    /// survives. Every other field shows only once a record
    /// speaks to it (sex at birth, a custom birth weight), in binding order.
    static func rows(
        _ fields: [CatalogConclusionField],
        header: CatalogPersonHeader? = nil,
        derivedRows: [CatalogDerivedRow] = [],
        locale: Locale,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> [ReconciledValueRowModel] {
        var rows: [ReconciledValueRowModel] = []
        if let name = fields.first(where: { $0.propertyKey == SeededPropertyKey.name }) {
            rows.append(ReconciledValueRowModel(field: name, locale: locale, orders: orders))
        } else {
            rows.append(.empty(
                id: "name", label: L10n.string(L10n.Conclusions.personName), style: .text,
                emptyText: L10n.string(L10n.Conclusions.emptyName)
            ))
        }
        let life: [(String, LocalizedStringResource, LocalizedStringResource, ReconciledValueRowModel.ValueStyle, CatalogDerivedRowKind)] = [
            ("life.birthDate", L10n.Conclusions.personBirthDate, L10n.Conclusions.emptyBirthDate, .date, .birthDate),
            ("life.birthPlace", L10n.Conclusions.personBirthPlace, L10n.Conclusions.emptyBirthPlace, .place, .birthPlace),
            ("life.deathDate", L10n.Conclusions.personDeathDate, L10n.Conclusions.emptyDeathDate, .date, .deathDate),
            ("life.deathPlace", L10n.Conclusions.personDeathPlace, L10n.Conclusions.emptyDeathPlace, .place, .deathPlace),
        ]
        let facts: [CatalogLifeFacts?] = [header?.birth, header?.birth, header?.death, header?.death]
        for (index, entry) in life.enumerated() {
            let (id, label, empty, style, kind) = entry
            let side = facts[index]
            let lead: String
            let mixed: Bool
            let opens: ReconciledValueRowModel.Opens?
            let shownOwnerID: String?
            let competing = (side?.eventCount ?? 0) > 1
            if style == .date {
                lead = PersonLifeDisplay.dateText(side?.date, locale: locale)
                mixed = competing || (side?.dateCount ?? 0) > 1
                opens = side?.event.map(ReconciledValueRowModel.Opens.event)
                shownOwnerID = side?.event?.id
            } else {
                let shown = DerivedPlace.shown(side?.places ?? [])
                lead = DerivedPlace.name(side?.places ?? [])
                mixed = competing || DerivedPlace.extra(side?.places ?? []) > 0
                opens = (shown ?? side?.places.first).map(ReconciledValueRowModel.Opens.place)
                shownOwnerID = shown?.entity.id
            }
            rows.append(.derived(
                id: id, label: L10n.string(label), style: style, lead: lead,
                emptyText: L10n.string(empty), mixed: mixed, opens: opens,
                owners: derivedRows.first { $0.kind == kind }?.owners ?? [],
                shownOwnerID: shownOwnerID, kind: kind, locale: locale, orders: orders
            ))
        }
        for field in fields where field.propertyKey != SeededPropertyKey.name && !field.outcomes.isEmpty {
            rows.append(ReconciledValueRowModel(field: field, locale: locale, orders: orders))
        }
        return rows
    }

    private static func lifeLine(
        _ abbreviation: LocalizedStringResource,
        spoken: LocalizedStringResource,
        facts: CatalogLifeFacts?,
        locale: Locale
    ) -> LifeLine {
        let date = PersonLifeDisplay.dateText(facts?.date, locale: locale)
        let place = DerivedPlace.name(facts?.places ?? [])
        return LifeLine(
            abbreviation: L10n.string(abbreviation),
            date: date.isEmpty ? nil : date,
            place: place.isEmpty ? nil : place,
            accessibilityLabel: L10n.Conclusions.a11yLifeLine(
                L10n.string(spoken),
                date: date.isEmpty ? L10n.string(L10n.Conclusions.dateUnknown) : date,
                place: place.isEmpty ? L10n.string(L10n.Conclusions.placeUnknown) : place
            )
        )
    }
}
