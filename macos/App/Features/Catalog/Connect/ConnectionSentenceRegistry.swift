import Foundation

/// Every connection sentence, as data. A bridge's kind, its role / type term
/// and (for a participation) the event's type pick one row; the row words the
/// connection. Adding wording for another event or role is one row here.
enum ConnectionSentenceRegistry {
    /// The two endpoint nouns in the kind's A/B order (participation A = person,
    /// B = event; location A = event, B = place; relationship and place
    /// relationship A = from, B = to) and the term's display label.
    struct Slots {
        var a: String
        var b: String
        var termLabel: String?
    }

    struct Row {
        var kind: EvidenceBridgeKind
        /// Role / relationship type term key; nil on the kind's default row.
        var term: String?
        /// Event type term key (participation only); nil matches any event.
        var eventType: String?
        /// Property key whose Observation carries the term (default rows only).
        var termProperty: String?
        /// The sentence when both endpoints are known.
        var sentence: (Slots) -> String
        /// The sentence when both endpoints are known but there is no term.
        var noTerm: ((Slots) -> String)?
        /// The relational phrase on a card, given the term label.
        var phrase: ((String?) -> String)?
    }

    private typealias Text = L10n.EvidenceGraph

    static let rows: [Row] = [
        Row(
            kind: .location,
            termProperty: SeededPropertyKey.role,
            sentence: { Text.bridgeSummaryLocation(event: $0.a, place: $0.b) },
            phrase: { _ in L10n.string(Text.bridgeSummaryLocationBare) }
        ),
        Row(
            kind: .relationship,
            termProperty: SeededPropertyKey.relationshipType,
            sentence: {
                Text.bridgeSummaryRelationship(person: $0.a, type: $0.termLabel ?? "", related: $0.b)
            },
            noTerm: { Text.bridgeSummaryRelationshipFallback(person: $0.a, related: $0.b) },
            phrase: { term in
                guard let term else { return L10n.string(Text.bridgeSummaryRelationshipBare) }
                return Text.bridgeSummaryRelationshipTypeOnly(type: term)
            }
        ),
        Row(
            kind: .placeRelationship,
            termProperty: SeededPropertyKey.placeRelationshipType,
            sentence: {
                Text.bridgeSummaryPlaceRelationship(from: $0.a, type: $0.termLabel ?? "", to: $0.b)
            },
            noTerm: { Text.bridgeSummaryPlaceRelationshipFallback(from: $0.a, to: $0.b) },
            phrase: { term in
                guard let term else { return L10n.string(Text.bridgeSummaryPlaceRelationshipBare) }
                return Text.bridgeSummaryPlaceRelationshipTypeOnly(type: term)
            }
        ),
        Row(
            kind: .participation,
            termProperty: SeededPropertyKey.role,
            sentence: {
                Text.bridgeSummaryParticipation(person: $0.a, role: $0.termLabel ?? "", event: $0.b)
            },
            noTerm: { Text.bridgeSummaryParticipationFallback(person: $0.a, event: $0.b) },
            phrase: { term in
                guard let term else { return L10n.string(Text.bridgeSummaryParticipationBare) }
                return Text.bridgeSummaryParticipationRoleOnly(role: term)
            }
        ),
        subjectRow(SeededEventTypeTerm.birth) { Text.bridgeSentenceBirthSubject(person: $0.a) },
        subjectRow(SeededEventTypeTerm.death) { Text.bridgeSentenceDeathSubject(person: $0.a) },
        subjectRow(SeededEventTypeTerm.baptism) { Text.bridgeSentenceBaptismSubject(person: $0.a) },
        subjectRow(SeededEventTypeTerm.marriage) { Text.bridgeSentenceMarriageSubject(person: $0.a, event: $0.b) },
        subjectRow(SeededEventTypeTerm.funeral) { Text.bridgeSentenceFuneralSubject(person: $0.a, event: $0.b) },
        subjectRow(SeededEventTypeTerm.burial) { Text.bridgeSentenceBurialSubject(person: $0.a) },
    ]

    private static func subjectRow(_ eventType: String, sentence: @escaping (Slots) -> String) -> Row {
        Row(kind: .participation, term: SeededRoleTerm.subject, eventType: eventType, sentence: sentence)
    }

    /// The most specific row: kind + term + event type, then kind + term, then
    /// the kind's default row (no term, no event type).
    static func lookup(kind: EvidenceBridgeKind, term: String?, eventType: String?) -> Row {
        let ofKind = rows.filter { $0.kind == kind }
        if let term, let eventType,
           let row = ofKind.first(where: { $0.term == term && $0.eventType == eventType }) {
            return row
        }
        if let term, let row = ofKind.first(where: { $0.term == term && $0.eventType == nil }) {
            return row
        }
        return defaultRow(kind)
    }

    static func defaultRow(_ kind: EvidenceBridgeKind) -> Row {
        // Every kind has a default row (`everyKindHasADefaultRow`).
        rows.first { $0.kind == kind && $0.term == nil && $0.eventType == nil }!
    }

    /// The property whose Observation carries the kind's term.
    static func termProperty(for kind: EvidenceBridgeKind) -> String {
        defaultRow(kind).termProperty ?? SeededPropertyKey.role
    }

    /// Relational phrase only; a specific row without its own phrase uses the
    /// kind default's.
    static func phrase(kind: EvidenceBridgeKind, term: String?) -> String {
        defaultRow(kind).phrase?(term) ?? ""
    }

    /// The sentence for a kind's endpoints (A/B as `Slots` documents). With an
    /// endpoint missing it is the relational phrase.
    static func sentence(
        kind: EvidenceBridgeKind,
        endpointA: String?,
        endpointB: String?,
        term: String?,
        termKey: String? = nil,
        eventType: String? = nil
    ) -> String {
        let fallback = defaultRow(kind)
        guard let endpointA, let endpointB else {
            return phrase(kind: kind, term: term)
        }
        let row = lookup(kind: kind, term: termKey, eventType: eventType)
        let slots = Slots(a: endpointA, b: endpointB, termLabel: term)
        if term != nil { return row.sentence(slots) }
        return (row.noTerm ?? fallback.noTerm ?? row.sentence)(slots)
    }
}
