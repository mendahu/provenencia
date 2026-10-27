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

    /// The record in between that a side records a link through: a marriage, a birth or a parent.
    /// A Source's is a page row; the record's is a header from the catalog.
    struct Pivot: Equatable {
        var name: String
        /// An event's date as the Events list lists it, without the source's words for it; empty for
        /// anything else, or when undated.
        var date: String
        var ref: String

        init(name: String, date: String, ref: String) {
            self.name = name
            self.date = date
            self.ref = ref
        }

        /// The record's pivot, titled as every header is; nil when it has none.
        init?(header: CatalogConclusionHeader?, orders: NameFormatOrders) {
            guard let header else { return nil }
            self.init(
                name: ConclusionHeaderDisplay.title(header, orders: orders),
                date: header.event.map {
                    DateRowDisplay.line(date: $0.date, start: $0.startDate, end: $0.endDate, includesPhrase: false)
                } ?? "",
                ref: ConclusionHeaderDisplay.ref(header)
            )
        }
    }

    /// A connection the row and the record share: its header, its box and its neighbor's own lines.
    struct RelationshipCard: Equatable, Identifiable {
        var id: String
        var neighborSubjectID: String
        var key: String
        /// "they share a spouse (CEV-3B5RT and PER-7QH3D)", without the ellipsis it is drawn after.
        var header: String
        /// "This source: … · The record: …", when either side records the link through a marriage,
        /// a birth or a parent; nil when both record it directly.
        var recorded: String?
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

    /// Words the evidence once. `neighbor` is the page row a Subject is (a connected row, or a
    /// relationship's Source pivot); `labels` and `worded` are the model's property labels and its
    /// values in the researcher's name order, which `orders` is for the record's pivot.
    static func build(
        evidence: CatalogClaimEvidence,
        neighbor: (String) -> PromoteFlow.Row?,
        labels: (String) -> String,
        worded: (CatalogConclusionValue) -> String,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> EvidenceBoard {
        func sourceTitle(_ pin: CatalogClaimEvidencePin) -> String? {
            guard let title = evidence.sourceTitles[pin.sourceID], !title.isEmpty else { return nil }
            return title
        }

        func recordSource(_ pin: CatalogClaimEvidencePin) -> String {
            PromoteEvidenceDisplay.recordSource(title: sourceTitle(pin))
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
            let source = recordSource(pin)
            let text = worded(pin.value)
            return Member(
                id: pin.id,
                value: pin.value.isDepiction ? .crops(crop(pin.value)) : .text(text),
                source: source,
                spoken: PromoteEvidenceDisplay.spokenMember(source: source, value: text)
            )
        }

        func lineRow(_ line: CatalogClaimEvidenceLine, showsThisSource: Bool) -> Line {
            let isLikeness = line.incoming.isDepiction
            let sentence = L10n.string(L10n.Promote.lineSentence(
                relation: line.relation, cause: line.cause, propertyCause: line.propertyCause, value: line.incoming
            ))
            let first = line.members.first
            let firstSource = first.map(recordSource) ?? PromoteEvidenceDisplay.recordSource(title: nil)
            let firstText = first.map { worded($0.value) }
            let value: Value
            if isLikeness {
                value = .crops(crop(line.incoming) + (first.map { crop($0.value) } ?? []))
            } else if let firstText {
                value = .text(firstText)
            } else {
                value = .none
            }
            let spoken: String
            if isLikeness {
                spoken = PromoteEvidenceDisplay.spokenLikeness(
                    source: firstSource, sentence: sentence, count: line.members.count
                )
            } else {
                spoken = PromoteEvidenceDisplay.spokenLabel(
                    property: labels(line.propertyKey),
                    sentence: sentence,
                    here: incomingText(line),
                    count: line.members.count,
                    member: firstText.map { (source: firstSource, value: $0) }
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
            let layerPivot = (relationship.layer.pivotID.isEmpty ? nil : neighbor(relationship.layer.pivotID)).map {
                Pivot(name: $0.name, date: $0.date, ref: $0.ref)
            }
            let recorded = PromoteEvidenceDisplay.recorded(
                layer: relationship.layer.pattern, layerPivot: layerPivot,
                canon: relationship.canon.pattern,
                canonPivot: Pivot(header: relationship.canon.pivot, orders: orders),
                term: relationship.neighborTerm?.productKey
            )
            let spokenHeader = recorded.map {
                L10n.Promote.spokenRecorded(header: header, source: $0.source, record: $0.record)
            } ?? header
            return RelationshipCard(
                id: relationship.id,
                neighborSubjectID: relationship.neighborSubjectID,
                key: relationship.key,
                header: header,
                recorded: recorded.map { L10n.Promote.recorded(source: $0.source, record: $0.record) },
                pinIDs: relationship.bundles.map(\.id),
                outcome: relationship.outcome,
                spoken: PromoteEvidenceDisplay.spokenRelationship(
                    header: spokenHeader, outcome: relationship.outcome, count: relationship.bundles.count
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
