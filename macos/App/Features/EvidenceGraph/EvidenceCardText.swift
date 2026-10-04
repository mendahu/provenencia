import Foundation

/// Display and VoiceOver copy for one graph card, formatted once per graph
/// snapshot rather than in view bodies.
///
/// The graph's document body runs again on every drag frame. Formatting card
/// sentences and accessibility names there repeated the same work for every
/// card on every frame; building it with the snapshot (see
/// ``EvidenceGraphModel/displayText(rows:types:)``) keeps formatting out of
/// the per-frame path. Views read these strings and format nothing.
struct EvidenceCardText: Equatable, Sendable {
    /// VoiceOver label for the card and its rotor entry.
    var accessibilityLabel: String
    /// Name of the card's delete accessibility action.
    var deleteActionName: String
    /// Primary cards: the footer action ("Promote …" / "Open person …").
    var footerActionName: String?
    /// Bridge cards: the edge sentence shown when the bridge is cited.
    var summary: String?

    init(subject placed: SourceGraphPlacedSubject) {
        accessibilityLabel = EvidenceSubjectCard.accessibilityLabel(for: placed)
        deleteActionName = L10n.EvidenceGraph.deleteAccessibility(
            kind: placed.kind.rawValue,
            label: placed.subject.label.isEmpty ? placed.subject.ref : placed.subject.label,
            ref: placed.subject.ref
        )
        footerActionName = EvidenceSubjectCard.footerAccessibilityActionName(for: placed)
        summary = nil
    }

    init(bridge placed: SourceGraphPlacedBridge, in snapshot: SourceGraphSnapshot) {
        let sentence = EvidenceBridgeEdgeSummary.sentence(for: placed, in: snapshot)
        accessibilityLabel = EvidenceBridgeCard.accessibilityLabel(for: placed, in: snapshot)
        deleteActionName = L10n.EvidenceGraph.deleteAccessibility(
            kind: placed.kind.rawValue,
            label: sentence,
            ref: placed.subject.ref
        )
        footerActionName = nil
        summary = sentence
    }
}

/// ``EvidenceCardText`` for every card in one graph snapshot, keyed by card id.
struct EvidenceGraphText: Equatable, Sendable {
    private var cards: [String: EvidenceCardText]

    init(snapshot: SourceGraphSnapshot) {
        var cards: [String: EvidenceCardText] = [:]
        cards.reserveCapacity(snapshot.subjects.count + snapshot.bridges.count)
        for placed in snapshot.subjects {
            cards[placed.id] = EvidenceCardText(subject: placed)
        }
        for placed in snapshot.bridges {
            cards[placed.id] = EvidenceCardText(bridge: placed, in: snapshot)
        }
        self.cards = cards
    }

    /// Text for a placed subject; built on the spot only if the snapshot lacked it.
    subscript(subject placed: SourceGraphPlacedSubject) -> EvidenceCardText {
        cards[placed.id] ?? EvidenceCardText(subject: placed)
    }

    /// Text for a placed bridge; built on the spot only if the snapshot lacked it.
    subscript(bridge placed: SourceGraphPlacedBridge, in snapshot: SourceGraphSnapshot) -> EvidenceCardText {
        cards[placed.id] ?? EvidenceCardText(bridge: placed, in: snapshot)
    }
}
