import Foundation

/// The Person page, worded from one `CatalogConclusionDetail` (S9-16, board
/// S9-D5). Pure: views only lay it out.
struct PersonDetailContent: ConclusionDetailBody {
    /// Name → label → ref (PD-1).
    var title: PVTitleSource
    var ref: String
    /// The ref is repeated at the right only when it isn't the title.
    var showsRef: Bool { if case .ref = title { false } else { true } }
    var members: String
    var lifeLines: [PersonLifeDisplay.LifeLine]
    var rows: [ReconciledValueRowModel]
    /// Nil when gender and sex at birth give nothing to show (S10-23).
    var treatment: PVGenderTreatment?
    /// The gallery of likenesses, earliest recorded first.
    var gallery: ConclusionDetailGallery
    /// Leads with the gallery's lead likeness.
    var thumbnail: ConclusionDetailThumbnail
    var badge: ConclusionDetailBadge? { treatment.map(Self.badge) }

    /// The header's gender badge: the value, with "at birth" when only sex at
    /// birth stands in, in the treatment's ink.
    static func badge(_ treatment: PVGenderTreatment) -> ConclusionDetailBadge {
        ConclusionDetailBadge(
            text: treatment.basis == .gender ? treatment.text : L10n.Conclusions.sexAtBirthBadge(treatment.text),
            ink: treatment.family.ink,
            accessibilityLabel: L10n.DesignSystem.spokenJoin(treatment.fieldLabel, treatment.text)
        )
    }

    init(
        detail: CatalogConclusionDetail,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        let summary = detail.summary?.person
        let nameField = detail.fields.first { $0.propertyKey == SeededPropertyKey.name }
        var leadName: CatalogNameValue?
        if case .name(let name)? = nameField?.displayedValues.first?.value { leadName = name }
        title = PersonTitleDisplay.titleSource(name: leadName, entity: detail.entity, style: .natural, orders: orders)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        lifeLines = PersonLifeDisplay.lifeLines(summary, locale: locale, missing: .stated)
        rows = Self.rows(detail.fields, summary: summary, derivedRows: detail.derivedRows, locale: locale, orders: orders)
        treatment = PVGenderTreatment.read(summary)
        gallery = ConclusionDetailGallery(
            title: L10n.Conclusions.likenessesSection,
            openTip: L10n.Conclusions.likenessViewHint,
            fallbackMark: .subjectPerson,
            fallbackColors: treatment?.thumbnailColors,
            tiles: ConclusionDetailGallery.tiles(
                from: detail,
                tileLabel: L10n.Conclusions.likenessLabel,
                leadTileLabel: L10n.Conclusions.likenessTileLabelLead,
                useAsLabel: L10n.Conclusions.useAsPortraitLabel,
                chosenBody: L10n.Conclusions.toastPortraitChosenBody,
                viewerLabel: L10n.Conclusions.likenessViewerLabel(index:count:line:)
            ),
            choose: .personPortrait,
            viewer: .likeness(title: L10n.Conclusions.likenessViewerTitle(title.text))
        )
        thumbnail = .leading(
            gallery: gallery,
            noImage: L10n.PersonCard.noLikeness,
            withImage: L10n.Conclusions.personLikenessThumb,
            leadLabel: L10n.Conclusions.likenessLabel
        )
    }

    /// A Person always shows its name and four life rows, stated empty when
    /// nothing is recorded. A life row is mixed when its value disagrees (two
    /// kept dates, two Places) or when more than one birth (or death) event
    /// survives. Every other field shows only once a record
    /// speaks to it (sex at birth, a custom birth weight), in binding order.
    static func rows(
        _ fields: [CatalogConclusionField],
        summary: CatalogPersonSummary? = nil,
        derivedRows: [CatalogDerivedRow] = [],
        locale: Locale,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> [ReconciledValueRowModel] {
        var rows: [ReconciledValueRowModel] = []
        if let name = fields.first(where: { $0.propertyKey == SeededPropertyKey.name }) {
            rows.append(ReconciledValueRowModel(field: name, locale: locale, orders: orders))
        } else {
            rows.append(.empty(
                id: "name", label: L10n.string(L10n.Conclusions.personName), face: .text,
                emptyText: L10n.string(L10n.Conclusions.emptyName)
            ))
        }
        let life: [(String, LocalizedStringResource, LocalizedStringResource, PVProvenanceFace, CatalogDerivedRowKind)] = [
            ("life.birthDate", L10n.Conclusions.personBirthDate, L10n.Conclusions.emptyBirthDate, .date, .birthDate),
            ("life.birthPlace", L10n.Conclusions.personBirthPlace, L10n.Conclusions.emptyBirthPlace, .place, .birthPlace),
            ("life.deathDate", L10n.Conclusions.personDeathDate, L10n.Conclusions.emptyDeathDate, .date, .deathDate),
            ("life.deathPlace", L10n.Conclusions.personDeathPlace, L10n.Conclusions.emptyDeathPlace, .place, .deathPlace),
        ]
        let facts: [CatalogLifeFacts?] = [summary?.birth, summary?.birth, summary?.death, summary?.death]
        for (index, entry) in life.enumerated() {
            let (id, label, empty, face, kind) = entry
            let side = facts[index]
            let lead: String
            let mixed: Bool
            let opens: ReconciledValueRowModel.Opens?
            let shownOwnerID: String?
            let competing = (side?.eventCount ?? 0) > 1
            if face == .date {
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
                id: id, label: L10n.string(label), face: face, lead: lead,
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
}
