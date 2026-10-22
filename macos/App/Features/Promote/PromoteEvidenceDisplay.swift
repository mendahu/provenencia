import Foundation

/// The evidence sheet's choices, made without a view: which boxes are
/// checked, whether a line disagrees, and which bridge a relationship names.
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

    /// A line anywhere in the evidence disagrees. A relationship never does.
    static func hasConflict(_ evidence: CatalogClaimEvidence) -> Bool {
        let related = evidence.relationships.flatMap(\.sections)
        return (evidence.sections + related).contains { section in
            section.lines.contains { $0.outcome == .conflict }
        }
    }

    /// The bridge on the page that joins the row and a neighbor: the first, in page order, with
    /// those two ends in either order and, when `kind` is given, of that kind.
    static func bridge(
        between row: String,
        and neighbor: String?,
        kind: EvidenceBridgeKind? = nil,
        in bridges: [PromoteFlow.BridgeFact]
    ) -> PromoteFlow.BridgeFact? {
        let ends: Set<String?> = [row, neighbor]
        return bridges.first { bridge in
            (kind == nil || bridge.kind == kind) && Set<String?>([bridge.endA, bridge.endB]) == ends
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
        return bridge(between: row, and: relationship.neighborSubjectID, kind: kind, in: bridges)
    }

    /// A record's own box under a line is off when the line cannot be pinned.
    static func memberDisabled(line: CatalogClaimEvidenceLine) -> Bool {
        !line.pinnable
    }

    /// Loaded, and nothing to compare: no line and no relationship.
    static func hasNothingToCompare(_ evidence: CatalogClaimEvidence) -> Bool {
        evidence.sections.allSatisfy { $0.lines.isEmpty } && evidence.relationships.isEmpty
    }

    /// A record behind a line: its Source's title when it has one, and its worded value.
    struct Member: Equatable {
        var source: String?
        var value: String

        var title: String? {
            guard let source, !source.isEmpty else { return nil }
            return source
        }
    }

    /// "Source: value", or the value alone when the Source has no title.
    static func memberText(_ member: Member) -> String {
        guard let title = member.title else { return member.value }
        return L10n.Promote.onSource(source: title, value: member.value)
    }

    /// The muted line under a sentence: this source's value, then, for one record, its Source and value.
    static func detail(here: String, member: Member?) -> String {
        var parts = [L10n.Promote.thisSource(here)]
        if let member { parts.append(memberText(member)) }
        return L10n.Promote.lineDetail(parts: parts)
    }

    /// One record as VoiceOver speaks it; the value alone when its Source has no title.
    static func spokenMember(_ member: Member) -> String {
        guard let title = member.title else { return L10n.Promote.spokenMemberUntitled(value: member.value) }
        return L10n.Promote.spokenMember(source: title, value: member.value)
    }

    /// A line as VoiceOver speaks it: with its record count when several stand behind it, else with
    /// its one record.
    static func spokenLabel(property: String, sentence: String, here: String, count: Int, member: Member?) -> String {
        if count > 1 {
            return L10n.Promote.spokenLineCount(property: property, sentence: sentence, count: count, here: here)
        }
        let line = L10n.Promote.spokenLine(property: property, sentence: sentence, here: here)
        guard let member else { return line }
        return L10n.Promote.spokenLineMember(line: line, member: spokenMember(member))
    }

    /// A likeness record's Source title, or a word for it when the Source has none.
    static func likenessSource(title: String?) -> String {
        guard let title, !title.isEmpty else { return L10n.string(L10n.Promote.untitledRecord) }
        return title
    }

    /// A relationship card's heading: what the card shows and how VoiceOver speaks it.
    struct RelationshipHeader: Equatable {
        var title: String
        var subtitle: String?
        var spoken: String
    }

    /// The bridge's sentence with the outcome as a badge; with no bridge, the outcome's sentence
    /// over the neighbor's name.
    static func relationshipHeader(
        bridge: PromoteFlow.BridgeFact?,
        outcome: EvidenceOutcome,
        neighborName: String?
    ) -> RelationshipHeader {
        if let bridge {
            return RelationshipHeader(
                title: bridge.sentence,
                subtitle: nil,
                spoken: L10n.Promote.spokenRelationship(sentence: bridge.sentence, link: L10n.string(L10n.Promote.outcome(outcome)))
            )
        }
        let title = L10n.string(L10n.Promote.relationshipSentence(outcome))
        let link = L10n.Promote.relationshipWith(neighborName ?? L10n.string(L10n.Promote.sheetNeighbor))
        return RelationshipHeader(
            title: title,
            subtitle: link,
            spoken: L10n.Promote.spokenRelationship(sentence: title, link: link)
        )
    }
}
