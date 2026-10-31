import Foundation

/// One record in a field's provenance: what it was read as, its Source, and what
/// the auto-reconciler did with it. In the Why accordion the Source opens the
/// record's Citation with its Observation in focus; the provenance sheet opens
/// the Source, the Citation or the Observation (`location`, `sourceLocation`,
/// `citationLocation`).
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
    /// The citation composer on the record's Citation, its Observation in focus.
    var location: WorkspaceLocation
    /// The Source page.
    var sourceLocation: WorkspaceLocation
    /// The citation composer on the record's Citation, no Observation in focus.
    var citationLocation: WorkspaceLocation
    var observationRef: String
    var subjectRef: String
    /// The displayed value the record counts under; 0 for none (`ReconciliationRecordGroup.ValueGroup`).
    var groupRank: Int
    /// "Source, Artifact, Citation, page" (`CitationLineDisplay`).
    var citationLine: String

    init(
        outcome: CatalogReconcilerOutcome,
        in field: CatalogConclusionField,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        id = outcome.observationID
        readAs = ReconciledValueDisplay.readAs(outcome, locale: locale, orders: orders)
        readAsIsPhrase = ReconciledValueDisplay.readAsIsPhrase(outcome)
        counted = ReconciledValueDisplay.counted(outcome, in: field)
        sourceTitle = Self.sourceTitle(for: outcome)
        phrase = ReconciledValueDisplay.outcomePhrase(outcome, in: field, locale: locale)
        mark = ReconciledValueDisplay.outcomeMark(outcome, in: field)
        location = Self.composerLocation(for: outcome)
        sourceLocation = .sourcePage(sourceId: outcome.sourceID, ref: nil, title: outcome.sourceTitle)
        citationLocation = Self.composerLocation(for: outcome, focusingObservation: false)
        observationRef = outcome.observationRef
        subjectRef = outcome.subjectRef
        groupRank = outcome.groupRank
        citationLine = CitationLineDisplay.line(outcome)
    }

    /// The Source's title, else the Observation ref when the Source has none.
    static func sourceTitle(for outcome: CatalogReconcilerOutcome) -> String {
        outcome.sourceTitle.isEmpty ? outcome.observationRef : outcome.sourceTitle
    }

    /// The citation composer on the record's Citation, its Observation in
    /// focus unless `focusingObservation` is false. Back returns to the page
    /// that opened it.
    static func composerLocation(
        for outcome: CatalogReconcilerOutcome,
        focusingObservation: Bool = true
    ) -> WorkspaceLocation {
        .citationComposer(
            sourceId: outcome.sourceID,
            graphId: outcome.graphID,
            subjectId: outcome.subjectID,
            citationId: outcome.citationID,
            artifactId: outcome.artifactID,
            observationId: focusingObservation ? outcome.observationID : nil,
            ref: outcome.subjectRef,
            sourceTitle: outcome.sourceTitle
        )
    }
}

/// The records of one owning entity in a derived row's Why. `owner` is nil
/// for a row with one owner, which needs no group line.
struct ReconciliationRecordGroup: Equatable, Identifiable {
    /// A group line: the owning entity as its list titles it.
    struct OwnerLine: Equatable {
        /// The owner's entity id.
        var id: String
        var title: String
        /// The owner's ref; nil when the title is the ref.
        var ref: String?
        var spoken: String
        /// The owner behind the row's shown value, when the row has several groups.
        var shown: Bool = false
    }

    /// The records of one run of the group that count under the same displayed
    /// value, in the order Go sent them. Only the provenance sheet reads these.
    struct ValueGroup: Equatable, Identifiable {
        enum Heading: Equatable {
            /// The displayed value the records went into.
            case value(ReconciledValueRowModel.OtherValue)
            /// Records that count under no displayed value.
            case otherRecords
        }

        var heading: Heading?
        var records: [ReconciliationRecord]

        var id: String { records.first?.id ?? "" }
    }

    var owner: OwnerLine?
    var records: [ReconciliationRecord]
    /// `records` in runs by value. One run with no heading unless the group needs headings.
    var values: [ValueGroup]

    var id: String { owner?.id ?? "" }

    init(owner: OwnerLine?, records: [ReconciliationRecord], values: [ValueGroup]? = nil) {
        self.owner = owner
        self.records = records
        self.values = values ?? (records.isEmpty ? [] : [ValueGroup(heading: nil, records: records)])
    }
}
