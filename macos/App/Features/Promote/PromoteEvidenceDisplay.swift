import Foundation

/// The evidence sheet's choices, made without a view: which boxes are
/// checked, whether a line disagrees, which bridge a relationship names, and
/// how a lifespan reads.
enum PromoteEvidenceDisplay {
    /// How much of a line's or relationship's pins are checked.
    enum CheckState: Equatable {
        case checked
        case unchecked
        case partial
    }

    /// All pins in `pins` is checked, none is unchecked, some is partial. No pins is unchecked.
    static func checkState(pinIDs: [String], pins: Set<String>) -> CheckState {
        guard !pinIDs.isEmpty else { return .unchecked }
        let checked = pinIDs.filter(pins.contains).count
        if checked == 0 { return .unchecked }
        return checked == pinIDs.count ? .checked : .partial
    }

    /// How many of the evidence's offered pins are checked.
    static func checkedCount(_ evidence: CatalogClaimEvidence, pins: Set<String>) -> Int {
        pins.intersection(evidence.offeredPinIDs).count
    }

    /// A line anywhere in the evidence, or the lifespan, disagrees. A relationship never does.
    static func hasConflict(_ evidence: CatalogClaimEvidence) -> Bool {
        if evidence.lifespan?.outcome == .conflict { return true }
        let related = evidence.relationships.flatMap(\.sections)
        return (evidence.sections + related).contains { section in
            section.lines.contains { $0.outcome == .conflict }
        }
    }

    /// The bridge on the page a relationship stands for: the first, in page order, of the
    /// relationship's kind that joins the row and the neighbor.
    static func bridge(
        for relationship: CatalogClaimEvidenceRelationship,
        row: String,
        in bridges: [PromoteFlow.BridgeFact]
    ) -> PromoteFlow.BridgeFact? {
        guard let kind = EvidenceBridgeKind(rawValue: relationship.bridgeType) else { return nil }
        let ends: Set<String?> = [row, relationship.neighborSubjectID]
        return bridges.first { $0.kind == kind && Set<String?>([$0.endA, $0.endB]) == ends }
    }

    /// The lifespan line's sentence. A side that contradicts itself is named before one that has
    /// no dates, this source before the record.
    static func lifespanSentence(_ lifespan: CatalogClaimEvidenceLifespan, locale: Locale = .autoupdatingCurrent) -> String {
        if lifespan.row.missing == .contradicts { return L10n.string(L10n.Promote.lifespanContradictsThisSource) }
        if lifespan.record.missing == .contradicts { return L10n.string(L10n.Promote.lifespanContradictsRecord) }
        if lifespan.row.missing != .none { return L10n.string(L10n.Promote.lifespanNoneThisSource) }
        if lifespan.record.missing != .none { return L10n.string(L10n.Promote.lifespanNoneRecord) }
        let here = PersonLifeDisplay.span(first: lifespan.row.first, last: lifespan.row.last, locale: locale)
        let there = PersonLifeDisplay.span(first: lifespan.record.first, last: lifespan.record.last, locale: locale)
        return lifespan.outcome == .conflict
            ? L10n.Promote.lifespanConflict(here: here, there: there)
            : L10n.Promote.lifespanFits(here: here, there: there)
    }
}
