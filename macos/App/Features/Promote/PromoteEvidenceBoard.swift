import Foundation

/// The evidence sheet's content, worded once when the evidence loads. A line's words, its
/// Source titles and its spoken label are made here from the catalog evidence, not again each
/// time the sheet draws, so a keystroke in the argument box or a ticked box leaves every other
/// line's value equal and its body unevaluated. The board holds no check state: which boxes
/// are checked is the row's pins, read beside the board.
struct EvidenceBoard: Equatable {
    /// Every pin the evidence offers, and the ones that start checked.
    var offeredPinIDs: Set<String>
    var checkedPinIDs: Set<String>
    /// The Subject's own Cards in Go's order, one per property; a Card's id is its property key.
    var cards: [Group]
    /// One Card per relationship, after them.
    var relationships: [RelationshipCard]

    /// Loaded, and nothing to compare: no line and no relationship.
    var hasNothingToCompare: Bool { cards.isEmpty && relationships.isEmpty }

    /// What a value column draws.
    enum Value: Equatable {
        case none
        case text(String)
        /// A likeness: the stored crops' paths, this Source's first.
        case crops([String])
    }

    /// One record behind a line, shown under it when the line is open.
    struct Member: Equatable, Identifiable {
        /// Go's pin id.
        var id: String
        var value: Value
        /// Its Source's title, or the word for a record whose Source has none.
        var source: String
        /// How VoiceOver speaks its box.
        var spoken: String
    }

    /// One incoming Observation against the records that say something about its property.
    struct Line: Equatable, Identifiable {
        var id: String
        /// The pins the line's box checks together.
        var pinIDs: [String]
        /// The first record's value, or a likeness's two crops.
        var value: Value
        var sentence: String
        /// "This source: …", when the Card does not say it once in its header; never for a likeness.
        var thisSource: String?
        /// The first record's Source title, or the word for one with no title.
        var source: String
        var outcome: EvidenceOutcome
        var cause: ValueCause
        var isLikeness: Bool
        /// How VoiceOver speaks the line's box.
        var spoken: String
        var members: [Member]

        /// Opens and closes the records behind it, and names its boxes.
        var key: String { pinIDs.first ?? id }
        /// Several records stand behind it: the Source column offers to show them.
        var collapses: Bool { pinIDs.count > 1 }
    }

    /// The lines for one property under a heading, with what this Source has when every line agrees
    /// on it: a Card of the Subject's own, or a group under a relationship.
    struct Group: Equatable, Identifiable {
        var id: String
        /// The Card header's words, without the ellipsis it is drawn after.
        var header: String
        var thisSource: String?
        var lines: [Line]
    }

    /// A connection the row and the record share: its header, its box and its neighbor's own lines.
    struct RelationshipCard: Equatable, Identifiable {
        var id: String
        var neighborSubjectID: String
        var key: String
        /// "they share a spouse (CEV-3B5RT and PER-7QH3D)", without the ellipsis it is drawn after.
        var header: String
        var pinIDs: [String]
        var outcome: EvidenceOutcome
        var spoken: String
        /// The neighbor's other properties that have lines.
        var groups: [Group]

        /// More than one record stands behind it: the header says how many.
        var recordCount: Int { pinIDs.count }
    }

    /// How many of the offered pins are checked in `pins`.
    func checkedCount(pins: Set<String>) -> Int {
        PromoteEvidenceDisplay.checkedCount(offered: offeredPinIDs, pins: pins)
    }

    /// Words the evidence once. `neighbor` is the page row a connected Subject is; `labels` and
    /// `worded` are the model's property labels and its values in the researcher's name order.
    static func build(
        evidence: CatalogClaimEvidence,
        neighbor: (String) -> PromoteFlow.Row?,
        labels: (String) -> String,
        worded: (CatalogConclusionValue) -> String
    ) -> EvidenceBoard {
        func sourceTitle(_ pin: CatalogClaimEvidencePin) -> String? {
            guard let title = evidence.sourceTitles[pin.sourceID], !title.isEmpty else { return nil }
            return title
        }

        func member(_ pin: CatalogClaimEvidencePin) -> PromoteEvidenceDisplay.Member {
            PromoteEvidenceDisplay.Member(source: sourceTitle(pin), value: worded(pin.value))
        }

        func incomingText(_ line: CatalogClaimEvidenceLine) -> String {
            let text = worded(line.incoming)
            return line.incomingNegative ? L10n.Promote.negatedValue(text) : text
        }

        func crop(_ value: CatalogConclusionValue) -> [String] {
            if case .depiction(let depiction) = value { return [depiction.imageRelPath] }
            return []
        }

        func memberRow(_ pin: CatalogClaimEvidencePin) -> Member {
            let isLikeness = pin.value.isDepiction
            // A likeness with no Source title is "a record", not "Likeness".
            let spokenSource = isLikeness
                ? PromoteEvidenceDisplay.Member(
                    source: PromoteEvidenceDisplay.recordSource(title: sourceTitle(pin)), value: worded(pin.value)
                )
                : member(pin)
            return Member(
                id: pin.id,
                value: isLikeness ? .crops(crop(pin.value)) : .text(worded(pin.value)),
                source: PromoteEvidenceDisplay.recordSource(title: sourceTitle(pin)),
                spoken: PromoteEvidenceDisplay.spokenMember(spokenSource)
            )
        }

        func lineRow(_ line: CatalogClaimEvidenceLine, showsThisSource: Bool) -> Line {
            let isLikeness = line.incoming.isDepiction
            let sentence = L10n.string(L10n.Promote.lineSentence(
                relation: line.relation, cause: line.cause, propertyCause: line.propertyCause, value: line.incoming
            ))
            let first = line.members.first
            let value: Value
            if isLikeness {
                value = .crops(crop(line.incoming) + (first.map { crop($0.value) } ?? []))
            } else if let first {
                value = .text(worded(first.value))
            } else {
                value = .none
            }
            let firstSource = PromoteEvidenceDisplay.recordSource(title: first.flatMap(sourceTitle))
            let spoken: String
            if isLikeness {
                spoken = L10n.Promote.spokenLikeness(source: firstSource, sentence: sentence)
            } else {
                spoken = PromoteEvidenceDisplay.spokenLabel(
                    property: labels(line.propertyKey),
                    sentence: sentence,
                    here: incomingText(line),
                    count: line.members.count,
                    member: first.map(member)
                )
            }
            return Line(
                id: line.id,
                pinIDs: line.members.map(\.id),
                value: value,
                sentence: sentence,
                thisSource: showsThisSource && !isLikeness ? L10n.Promote.thisSource(incomingText(line)) : nil,
                source: firstSource,
                outcome: line.outcome,
                cause: line.cause,
                isLikeness: isLikeness,
                spoken: spoken,
                members: line.members.map(memberRow)
            )
        }

        func group(_ id: String, _ section: CatalogClaimEvidenceSection) -> Group {
            let placement = PromoteEvidenceDisplay.thisSourcePlacement(values: section.lines.map { line in
                line.incoming.isDepiction ? nil : incomingText(line)
            })
            return Group(
                id: id,
                header: PromoteEvidenceDisplay.similarityHeader(property: labels(section.propertyKey)),
                thisSource: placement.headerValue,
                lines: section.lines.map { lineRow($0, showsThisSource: placement == .lines) }
            )
        }

        func relationshipCard(_ relationship: CatalogClaimEvidenceRelationship) -> RelationshipCard {
            let neighborRow = neighbor(relationship.neighborSubjectID)
            let phrase = PromoteEvidenceDisplay.sharePhrase(
                relationship: relationship, neighborName: neighborRow?.name ?? ""
            )
            let header = PromoteEvidenceDisplay.shareHeader(
                phrase: phrase, ownRef: neighborRow?.ref, recordRef: neighborRow?.target.handleRef
            )
            let groups = relationship.sections.filter { !$0.lines.isEmpty }.enumerated().map { index, section in
                group("\(relationship.id).\(section.propertyKey).\(index)", section)
            }
            return RelationshipCard(
                id: relationship.id,
                neighborSubjectID: relationship.neighborSubjectID,
                key: relationship.key,
                header: header,
                pinIDs: relationship.bundles.map(\.id),
                outcome: relationship.outcome,
                spoken: L10n.Promote.spokenRelationship(
                    sentence: header, link: L10n.string(L10n.Promote.outcome(relationship.outcome))
                ),
                groups: groups
            )
        }

        return EvidenceBoard(
            offeredPinIDs: evidence.offeredPinIDs,
            checkedPinIDs: evidence.checkedPinIDs,
            cards: evidence.sections.filter { !$0.lines.isEmpty }.map { group($0.propertyKey, $0) },
            relationships: evidence.relationships.map(relationshipCard)
        )
    }
}
