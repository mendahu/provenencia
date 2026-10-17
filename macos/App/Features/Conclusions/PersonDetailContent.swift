import Foundation

/// The Person page, worded from one `CatalogConclusionDetail` (S9-16, board
/// S9-D5). Pure: views only lay it out.
struct PersonDetailContent: ConclusionDetailBody {
    /// One header line: *b. 14 May 1817 · York*. Date and place come from
    /// the Person header. Nil when that half is not recorded.
    struct Vital: Equatable {
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
    var vitals: [Vital]
    var rows: [ReconciledValueRowModel]
    /// Nil when gender and sex at birth give nothing to show (S10-23).
    var treatment: PersonGenderTreatment?
    /// The `objects/…` path of the likeness the Person header leads with; "" for none.
    var likenessImageRelPath: String
    /// The gallery, earliest recorded first.
    var likenesses: [ConclusionDetailLikeness]
    var thumbnail: ConclusionDetailThumbnail {
        .init(
            mark: .subjectPerson,
            label: likenessImageRelPath.isEmpty
                ? L10n.Conclusions.personNoLikeness : L10n.Conclusions.personLikenessThumb,
            imageRelPath: likenessImageRelPath,
            imageLabel: leadLabel,
            colors: treatment?.thumbnailColors
        )
    }
    var badge: ConclusionDetailBadge? { treatment?.badge }

    /// Spoken label of the header thumb while it shows the lead likeness: the
    /// gallery tile with that image gives its Source; nil when there is none.
    private var leadLabel: String? {
        guard !likenessImageRelPath.isEmpty else { return nil }
        if let tile = likenesses.first(where: { $0.imageRelPath == likenessImageRelPath }) {
            return L10n.Conclusions.likenessLabel(tile.sourceTitle)
        }
        return L10n.string(L10n.Conclusions.personLikenessThumb)
    }

    init(
        detail: CatalogConclusionDetail,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let header = detail.personHeader
        let nameField = detail.fields.first { $0.propertyKey == SeededPropertyKey.name }
        var leadName: CatalogNameValue?
        if case .name(let name)? = nameField?.displayedValues.first?.value { leadName = name }
        title = PersonHeaderDisplay.titleSource(name: leadName, entity: detail.entity, style: .natural, orders: orders)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        vitals = [
            Self.vital(L10n.Conclusions.personBornAbbr, spoken: L10n.Conclusions.a11yBorn, facts: header?.birth, locale: locale),
            Self.vital(L10n.Conclusions.personDiedAbbr, spoken: L10n.Conclusions.a11yDied, facts: header?.death, locale: locale),
        ]
        rows = Self.rows(detail.fields, header: header, locale: locale, orders: orders)
        treatment = PersonGenderTreatment.read(detail.fields)
        likenessImageRelPath = header?.likenessImageRelPath ?? ""
        likenesses = detail.likenesses.map {
            ConclusionDetailLikeness(
                id: $0.record.observationID,
                imageRelPath: $0.imageRelPath,
                sourceTitle: ReconciliationRecord.sourceTitle(for: $0.record),
                location: ReconciliationRecord.composerLocation(for: $0.record)
            )
        }
    }

    /// A Person always shows its name and four life rows, stated empty when
    /// nothing is recorded. A life row is mixed when its value disagrees (two
    /// kept dates, two Places) or when more than one birth (or death) event
    /// survives. Every other field shows only once a record
    /// speaks to it (sex at birth, a custom birth weight), in binding order.
    static func rows(
        _ fields: [CatalogConclusionField],
        header: CatalogPersonHeader? = nil,
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
        let life: [(String, LocalizedStringResource, LocalizedStringResource, ReconciledValueRowModel.ValueStyle)] = [
            ("life.birthDate", L10n.Conclusions.personBirthDate, L10n.Conclusions.emptyBirthDate, .date),
            ("life.birthPlace", L10n.Conclusions.personBirthPlace, L10n.Conclusions.emptyBirthPlace, .place),
            ("life.deathDate", L10n.Conclusions.personDeathDate, L10n.Conclusions.emptyDeathDate, .date),
            ("life.deathPlace", L10n.Conclusions.personDeathPlace, L10n.Conclusions.emptyDeathPlace, .place),
        ]
        let facts: [CatalogLifeFacts?] = [header?.birth, header?.birth, header?.death, header?.death]
        for (index, entry) in life.enumerated() {
            let (id, label, empty, style) = entry
            let side = facts[index]
            let lead: String
            let mixed: Bool
            let opens: ReconciledValueRowModel.Opens?
            let competing = (side?.eventCount ?? 0) > 1
            if style == .date {
                lead = PersonLifeDisplay.dateText(side?.date, locale: locale)
                mixed = competing || (side?.dateCount ?? 0) > 1
                opens = side?.event.map(ReconciledValueRowModel.Opens.event)
            } else {
                lead = DerivedPlace.name(side?.places ?? [])
                mixed = competing || DerivedPlace.extra(side?.places ?? []) > 0
                opens = side?.places.first.map(ReconciledValueRowModel.Opens.place)
            }
            rows.append(.derived(
                id: id, label: L10n.string(label), style: style, lead: lead,
                emptyText: L10n.string(empty), mixed: mixed, opens: opens
            ))
        }
        for field in fields where field.propertyKey != SeededPropertyKey.name && !field.outcomes.isEmpty {
            rows.append(ReconciledValueRowModel(field: field, locale: locale, orders: orders))
        }
        return rows
    }

    private static func vital(
        _ abbreviation: LocalizedStringResource,
        spoken: LocalizedStringResource,
        facts: CatalogLifeFacts?,
        locale: Locale
    ) -> Vital {
        let date = PersonLifeDisplay.dateText(facts?.date, locale: locale)
        let place = DerivedPlace.name(facts?.places ?? [])
        return Vital(
            abbreviation: L10n.string(abbreviation),
            date: date.isEmpty ? nil : date,
            place: place.isEmpty ? nil : place,
            accessibilityLabel: L10n.Conclusions.a11yVital(
                L10n.string(spoken),
                date: date.isEmpty ? L10n.string(L10n.Conclusions.dateUnknown) : date,
                place: place.isEmpty ? L10n.string(L10n.Conclusions.placeUnknown) : place
            )
        )
    }
}
