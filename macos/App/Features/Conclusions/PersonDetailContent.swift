import Foundation

/// The Person page, worded from one `CatalogConclusionDetail` (S9-16, board
/// S9-D5). Pure: views only lay it out.
struct PersonDetailContent: ConclusionDetailBody {
    /// One header line: *b. 14 May 1817 · York, Upper Canada*. Date and
    /// place come from the birth / death Events (S9-32); nil until then.
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

    init(detail: CatalogConclusionDetail, locale: Locale = .autoupdatingCurrent) {
        let nameField = detail.fields.first { $0.propertyKey == "name" }
        var leadName: CatalogNameValue?
        if case .name(let name)? = nameField?.displayedValues.first?.value { leadName = name }
        title = PersonHeaderDisplay.titleSource(name: leadName, entity: detail.entity)
        ref = detail.entity.ref
        members = L10n.Conclusions.memberCount(detail.memberCount)
        vitals = [
            Self.vital(L10n.Conclusions.personBornAbbr, spoken: L10n.Conclusions.a11yBorn),
            Self.vital(L10n.Conclusions.personDiedAbbr, spoken: L10n.Conclusions.a11yDied),
        ]
        rows = Self.rows(detail.fields, locale: locale)
    }

    /// A Person always shows its name and four life rows, stated empty when
    /// nothing is recorded. Every other field shows only once a record
    /// speaks to it (sex at birth, a custom birth weight), in binding order.
    static func rows(_ fields: [CatalogConclusionField], locale: Locale) -> [ReconciledValueRowModel] {
        var rows: [ReconciledValueRowModel] = []
        if let name = fields.first(where: { $0.propertyKey == "name" }) {
            rows.append(ReconciledValueRowModel(field: name, locale: locale))
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
        for (id, label, empty, style) in life {
            rows.append(.empty(id: id, label: L10n.string(label), style: style, emptyText: L10n.string(empty)))
        }
        for field in fields where field.propertyKey != "name" && !field.outcomes.isEmpty {
            rows.append(ReconciledValueRowModel(field: field, locale: locale))
        }
        return rows
    }

    private static func vital(_ abbreviation: LocalizedStringResource, spoken: LocalizedStringResource) -> Vital {
        Vital(
            abbreviation: L10n.string(abbreviation),
            date: nil,
            place: nil,
            accessibilityLabel: L10n.Conclusions.a11yVital(
                L10n.string(spoken),
                date: L10n.string(L10n.Conclusions.personDateUnknown),
                place: L10n.string(L10n.Conclusions.personPlaceUnknown)
            )
        )
    }
}

/// One field row on a Conclusion detail page (board S9-D5
/// `ReconciledValueRow`): label · lead value · state · support · disclosures.
struct ReconciledValueRowModel: Equatable, Identifiable {
    enum ValueStyle: Equatable {
        /// Names, text, terms, integers.
        case text
        /// Mono, like every date in the app.
        case date
        /// Italic, like every place.
        case place
    }

    struct OtherValue: Equatable, Identifiable {
        var rank: Int
        var text: String
        var support: String
        var id: Int { rank }
    }

    var id: String
    var label: String
    var style: ValueStyle
    /// The lead value; nil when the field is empty.
    var lead: String?
    var emptyText: String
    var badge: ReconciledValueDisplay.StateBadge?
    var count: String?
    var against: String?
    var otherValuesLabel: String?
    var otherValues: [OtherValue]
    var whyTitle: String
    var records: [ReconciliationRecord]
    var accessibilityLabel: String

    init(field: CatalogConclusionField, locale: Locale = .autoupdatingCurrent) {
        let lead = field.displayedValues.first
            .map { ReconciledValueDisplay.string(for: $0.value, locale: locale) }
            .flatMap { $0.isEmpty ? nil : $0 }
        id = field.propertyID
        label = field.label
        style = field.valueType == "date" ? .date : .text
        self.lead = lead
        emptyText = ReconciledValueDisplay.emptyText(propertyKey: field.propertyKey)
        badge = ReconciledValueDisplay.stateBadge(field)
        count = ReconciledValueDisplay.countLine(field)
        against = ReconciledValueDisplay.againstLine(field)
        otherValuesLabel = ReconciledValueDisplay.otherValuesLabel(field)
        otherValues = ReconciledValueDisplay.otherValues(field).map {
            OtherValue(
                rank: $0.rank,
                text: ReconciledValueDisplay.string(for: $0.value, locale: locale),
                support: L10n.Conclusions.sourceCount($0.support)
            )
        }
        whyTitle = L10n.Conclusions.whyTitle(lead ?? field.label)
        records = field.outcomes.map { ReconciliationRecord(outcome: $0, in: field, locale: locale) }
        accessibilityLabel = ReconciledValueDisplay.accessibilityLabel(label: field.label, lead: lead, field: field)
    }

    /// One Date row whose lead is a start–end span. Why lists every record
    /// on those properties. The span is not one property's reconciled state,
    /// so it carries no badge.
    static func spanning(
        id: String,
        label: String,
        lead: String?,
        emptyText: String,
        fields: [CatalogConclusionField],
        locale: Locale
    ) -> Self {
        let shown = lead.flatMap { $0.isEmpty ? nil : $0 }
        let records = fields.flatMap { field in
            field.outcomes.map { ReconciliationRecord(outcome: $0, in: field, locale: locale) }
        }
        let spoken = shown.map { L10n.Conclusions.a11yList(label, rest: $0) }
            ?? ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        return Self(
            id: id, label: label, style: .date, lead: shown, emptyText: emptyText,
            badge: nil, count: nil, against: nil, otherValuesLabel: nil, otherValues: [],
            whyTitle: shown.map { L10n.Conclusions.whyTitle($0) } ?? "",
            records: records,
            accessibilityLabel: spoken
        )
    }

    /// A row with nothing behind it yet (the life rows until S9-32).
    static func empty(id: String, label: String, style: ValueStyle, emptyText: String) -> Self {
        Self(
            id: id, label: label, style: style, lead: nil, emptyText: emptyText, badge: nil, count: nil,
            against: nil, otherValuesLabel: nil, otherValues: [], whyTitle: "", records: [],
            accessibilityLabel: ReconciledValueDisplay.accessibilityLabel(label: label, lead: nil, field: nil)
        )
    }

    private init(
        id: String, label: String, style: ValueStyle, lead: String?, emptyText: String,
        badge: ReconciledValueDisplay.StateBadge?, count: String?, against: String?,
        otherValuesLabel: String?, otherValues: [OtherValue], whyTitle: String,
        records: [ReconciliationRecord], accessibilityLabel: String
    ) {
        self.id = id
        self.label = label
        self.style = style
        self.lead = lead
        self.emptyText = emptyText
        self.badge = badge
        self.count = count
        self.against = against
        self.otherValuesLabel = otherValuesLabel
        self.otherValues = otherValues
        self.whyTitle = whyTitle
        self.records = records
        self.accessibilityLabel = accessibilityLabel
    }
}

/// One record in a field's Why: what it was read as, its Source, and what the
/// auto-reconciler did with it. The Source opens the record's Citation.
struct ReconciliationRecord: Equatable, Identifiable {
    var id: String
    var readAs: String
    /// "not …" or "—": set as a phrase, not in the value's face.
    var readAsIsPhrase: Bool
    /// Went into the shown value (primary ink); the rest are secondary.
    var counted: Bool
    var sourceTitle: String
    var phrase: String
    var mark: PVSymbol
    var location: WorkspaceLocation

    init(outcome: CatalogReconcilerOutcome, in field: CatalogConclusionField, locale: Locale = .autoupdatingCurrent) {
        id = outcome.observationID
        readAs = ReconciledValueDisplay.readAs(outcome, locale: locale)
        readAsIsPhrase = ReconciledValueDisplay.readAsIsPhrase(outcome)
        counted = ReconciledValueDisplay.counted(outcome, in: field)
        sourceTitle = outcome.sourceTitle.isEmpty ? outcome.observationRef : outcome.sourceTitle
        phrase = ReconciledValueDisplay.outcomePhrase(outcome, in: field, locale: locale)
        mark = ReconciledValueDisplay.outcomeMark(outcome, in: field)
        location = Self.composerLocation(for: outcome)
    }

    /// The citation composer on the record's Citation, its Observation in
    /// focus. Back returns to the page that opened it.
    static func composerLocation(for outcome: CatalogReconcilerOutcome) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: outcome.sourceID,
            subjectId: outcome.subjectID,
            citationId: outcome.citationID,
            artifactId: outcome.artifactID,
            observationId: outcome.observationID,
            sourceSurface: .citationComposer,
            ref: outcome.subjectRef,
            sourceTitle: outcome.sourceTitle
        )
    }
}
