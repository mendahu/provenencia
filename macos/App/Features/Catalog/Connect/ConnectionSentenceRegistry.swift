import Foundation

/// Every connection sentence, as data. Each kind has one `KindEntry` (its
/// default wording and where its term lives); a `Sentence` words one role /
/// type term key, and for a participation one event type, more specifically.
/// Adding wording for another event or role is one `Sentence` here.
enum ConnectionSentenceRegistry {
    /// The two endpoint nouns in the kind's A/B order (participation A = person,
    /// B = event; location A = event, B = place; relationship and place
    /// relationship A = from, B = to) and the term's display label.
    struct Slots {
        var a: String
        var b: String
        var termLabel: String?
    }

    /// What a kind always has: where its term lives and its default wording.
    struct KindEntry {
        var kind: EvidenceBridgeKind
        /// Property key whose Observation carries the role / type term.
        var termProperty: String
        /// The sentence when both endpoints and a term label are known.
        var sentence: (Slots) -> String
        /// The sentence when both endpoints are known but there is no term
        /// (the kind's `sentence` when nil).
        var noTerm: ((Slots) -> String)?
        /// The relational phrase on a card, given the term label.
        var phrase: (String?) -> String
    }

    /// Wording for one term key (and optionally one event type) of a kind.
    struct Sentence {
        var kind: EvidenceBridgeKind
        /// Role / relationship type term key.
        var term: String
        /// Event type term key (participation only); nil matches any event.
        var eventType: String?
        var sentence: (Slots) -> String
        /// The card phrase; the kind's when nil.
        var phrase: ((String?) -> String)?
    }

    private typealias Text = L10n.EvidenceGraph

    static let entries: [KindEntry] = [
        KindEntry(
            kind: .location,
            termProperty: SeededPropertyKey.role,
            sentence: { Text.bridgeSummaryLocation(event: $0.a, place: $0.b) },
            phrase: { _ in L10n.string(Text.bridgeSummaryLocationBare) }
        ),
        KindEntry(
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
        KindEntry(
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
        KindEntry(
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
    ]

    static let sentences: [Sentence] = [
        subjectSentence(SeededEventTypeTerm.birth) { Text.bridgeSentenceBirthSubject(person: $0.a) },
        subjectSentence(SeededEventTypeTerm.death) { Text.bridgeSentenceDeathSubject(person: $0.a) },
        subjectSentence(SeededEventTypeTerm.baptism) { Text.bridgeSentenceBaptismSubject(person: $0.a) },
        subjectSentence(SeededEventTypeTerm.marriage) {
            Text.bridgeSentenceMarriageSubject(person: $0.a, event: $0.b)
        },
        subjectSentence(SeededEventTypeTerm.funeral) {
            Text.bridgeSentenceFuneralSubject(person: $0.a, event: $0.b)
        },
        subjectSentence(SeededEventTypeTerm.burial) { Text.bridgeSentenceBurialSubject(person: $0.a) },
    ]

    private static func subjectSentence(
        _ eventType: String,
        sentence: @escaping (Slots) -> String
    ) -> Sentence {
        Sentence(kind: .participation, term: SeededRoleTerm.subject, eventType: eventType, sentence: sentence)
    }

    /// The kind's entry; `everyKindHasAnEntry` proves none is missing.
    static func entry(_ kind: EvidenceBridgeKind) -> KindEntry {
        guard let entry = entries.first(where: { $0.kind == kind }) else {
            preconditionFailure("ConnectionSentenceRegistry has no entry for \(kind)")
        }
        return entry
    }

    /// The most specific wording: kind + term key + event type, then kind +
    /// term key. nil means the kind's entry words it.
    static func lookup(
        kind: EvidenceBridgeKind,
        term: String?,
        eventType: String?,
        in sentences: [Sentence] = Self.sentences
    ) -> Sentence? {
        guard let term else { return nil }
        let ofKind = sentences.filter { $0.kind == kind && $0.term == term }
        if let eventType, let found = ofKind.first(where: { $0.eventType == eventType }) {
            return found
        }
        return ofKind.first { $0.eventType == nil }
    }

    /// The kind a Connect rule's bridge type names; an unknown type is worded
    /// as a participation.
    static func kind(bridgeTypeKey: String) -> EvidenceBridgeKind {
        EvidenceBridgeKind(rawValue: bridgeTypeKey) ?? .participation
    }

    /// The property whose Observation carries the kind's term.
    static func termProperty(for kind: EvidenceBridgeKind) -> String {
        entry(kind).termProperty
    }

    /// Relational phrase only; a specific sentence's own phrase wins.
    static func phrase(
        kind: EvidenceBridgeKind,
        term: String?,
        termKey: String? = nil,
        eventType: String? = nil,
        in sentences: [Sentence] = Self.sentences
    ) -> String {
        let specific = lookup(kind: kind, term: termKey, eventType: eventType, in: sentences)
        return (specific?.phrase ?? entry(kind).phrase)(term)
    }

    /// The sentence for a kind's endpoints (A/B as `Slots` documents). With an
    /// endpoint missing it is the relational phrase.
    static func sentence(
        kind: EvidenceBridgeKind,
        endpointA: String?,
        endpointB: String?,
        term: String?,
        termKey: String? = nil,
        eventType: String? = nil,
        in sentences: [Sentence] = Self.sentences
    ) -> String {
        guard let endpointA, let endpointB else {
            return phrase(kind: kind, term: term, termKey: termKey, eventType: eventType, in: sentences)
        }
        let slots = Slots(a: endpointA, b: endpointB, termLabel: term)
        if let specific = lookup(kind: kind, term: termKey, eventType: eventType, in: sentences) {
            return specific.sentence(slots)
        }
        let entry = entry(kind)
        if term != nil { return entry.sentence(slots) }
        return (entry.noTerm ?? entry.sentence)(slots)
    }
}
