import Foundation

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
    }

    /// The Source's title, else the Observation ref when the Source has none.
    static func sourceTitle(for outcome: CatalogReconcilerOutcome) -> String {
        outcome.sourceTitle.isEmpty ? outcome.observationRef : outcome.sourceTitle
    }

    /// The citation composer on the record's Citation, its Observation in
    /// focus. Back returns to the page that opened it.
    static func composerLocation(for outcome: CatalogReconcilerOutcome) -> WorkspaceLocation {
        .citationComposer(
            sourceId: outcome.sourceID,
            graphId: outcome.graphID,
            subjectId: outcome.subjectID,
            citationId: outcome.citationID,
            artifactId: outcome.artifactID,
            observationId: outcome.observationID,
            ref: outcome.subjectRef,
            sourceTitle: outcome.sourceTitle
        )
    }
}
