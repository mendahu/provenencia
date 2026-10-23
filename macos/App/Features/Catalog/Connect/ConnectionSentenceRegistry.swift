import Foundation

/// Every connection sentence, as data. Each kind has one `KindEntry` (its
/// default wording and where its term lives); a `Sentence` words one role /
/// type term key, and for a participation one event type, more specifically.
/// Adding wording for another event or role is one `Sentence` here.
enum ConnectionSentenceRegistry {
    /// The two endpoint nouns in the kind's A/B order (participation A = person,
    /// B = event; location A = event, B = place; relationship and place
    /// relationship A = from, B = to) and the term's display label.
    struct Slots: Sendable {
        var a: String
        var b: String
        var termLabel: String?
    }

    /// What a kind always has: where its term lives and its default wording.
    struct KindEntry: Sendable {
        var kind: EvidenceBridgeKind
        /// Property key whose Observation carries the role / type term.
        var termProperty: String
        /// The sentence when both endpoints and a term label are known.
        var sentence: @Sendable (Slots) -> String
        /// The sentence when both endpoints are known but there is no term
        /// (the kind's `sentence` when nil).
        var noTerm: (@Sendable (Slots) -> String)?
        /// The relational phrase on a card, given the term label.
        var phrase: @Sendable (String?) -> String
        /// Whether the kind has a term a reason can name.
        var namesTerm: Bool
        /// What a row shares with a connected record, after "they share": "a spouse", "a birth
        /// event". nil when the input names nothing this kind can word.
        var share: @Sendable (ShareInput, _ similar: Bool) -> String?
    }

    /// What the engine reports of a connection between a row and a neighbor, read from the
    /// neighbor's side (`segmentcompare.ReadFrom`).
    struct ShareInput: Sendable, Equatable {
        var kind: EvidenceBridgeKind
        /// The endpoint property the neighbor is at ("person", "event", "to"); "" for a symmetric link.
        var neighborEnd: String
        /// The term the neighbor carries; nil when the connection has none.
        var neighborTerm: CatalogEvidenceTerm?
        /// The term names the neighbor ("the neighbor is the parent"), not the row.
        var namesNeighbor: Bool
        /// A participation's event type.
        var eventType: CatalogEvidenceTerm?
        var neighborName: String

        init(
            kind: EvidenceBridgeKind,
            neighborEnd: String,
            neighborTerm: CatalogEvidenceTerm? = nil,
            namesNeighbor: Bool = true,
            eventType: CatalogEvidenceTerm? = nil,
            neighborName: String
        ) {
            self.kind = kind
            self.neighborEnd = neighborEnd
            self.neighborTerm = neighborTerm
            self.namesNeighbor = namesNeighbor
            self.eventType = eventType
            self.neighborName = neighborName
        }

        /// A relationship on the evidence sheet; nil for a bridge type this build cannot word.
        init?(_ relationship: CatalogClaimEvidenceRelationship, neighborName: String) {
            guard let kind = EvidenceBridgeKind(rawValue: relationship.bridgeType) else { return nil }
            self.init(
                kind: kind,
                neighborEnd: relationship.neighborEnd,
                neighborTerm: relationship.neighborTerm,
                namesNeighbor: relationship.neighborTermNamesNeighbor,
                eventType: relationship.eventType,
                neighborName: neighborName
            )
        }
    }

    /// Wording for one term key (and optionally one event type) of a kind.
    struct Sentence: Sendable {
        var kind: EvidenceBridgeKind
        /// Role / relationship type term key.
        var term: String
        /// Event type term key (participation only); nil matches any event.
        var eventType: String?
        var sentence: @Sendable (Slots) -> String
        /// The card phrase; the kind's when nil.
        var phrase: (@Sendable (String?) -> String)?
    }

    private typealias Text = L10n.EvidenceGraph

    /// The term label as every sentence reads it mid-line ("spouse"):
    /// lowercased for the current locale; blank is nil. The one place a term
    /// is lowercased.
    private static func termInLowercase(_ term: String?) -> String? {
        guard let term = term?.trimmingCharacters(in: .whitespacesAndNewlines), !term.isEmpty else { return nil }
        return term.lowercased(with: .autoupdatingCurrent)
    }

    // MARK: Share wording

    private typealias ShareTerm = L10n.EvidenceGraph.ShareTerm

    /// Each seeded term's own share phrase, article included ("an informant"). A role, an event
    /// type and a relationship term of one key (child, spouse) read in their own table.
    private static let roleShares: [String: ShareTerm] = [
        SeededRoleTerm.subject: Text.shareRoleSubject,
        SeededRoleTerm.father: Text.shareRoleFather,
        SeededRoleTerm.mother: Text.shareRoleMother,
        SeededRoleTerm.spouse: Text.shareRoleSpouse,
        SeededRoleTerm.child: Text.shareRoleChild,
        SeededRoleTerm.witness: Text.shareRoleWitness,
        SeededRoleTerm.informant: Text.shareRoleInformant,
        SeededRoleTerm.pallbearer: Text.shareRolePallbearer,
    ]

    private static let eventShares: [String: ShareTerm] = [
        SeededEventTypeTerm.birth: Text.shareEventBirth,
        SeededEventTypeTerm.death: Text.shareEventDeath,
        SeededEventTypeTerm.marriage: Text.shareEventMarriage,
        SeededEventTypeTerm.engagement: Text.shareEventEngagement,
        SeededEventTypeTerm.baptism: Text.shareEventBaptism,
        SeededEventTypeTerm.burial: Text.shareEventBurial,
        SeededEventTypeTerm.funeral: Text.shareEventFuneral,
        SeededEventTypeTerm.census: Text.shareEventCensus,
        SeededEventTypeTerm.residence: Text.shareEventResidence,
        SeededEventTypeTerm.migration: Text.shareEventMigration,
        SeededEventTypeTerm.travel: Text.shareEventTravel,
    ]

    private static let relationshipShares: [String: ShareTerm] = [
        SeededRelationshipTypeTerm.spouse: Text.shareRelationshipSpouse,
        SeededRelationshipTypeTerm.sibling: Text.shareRelationshipSibling,
        SeededRelationshipTypeTerm.cousin: Text.shareRelationshipCousin,
        SeededRelationshipTypeTerm.parent: Text.shareRelationshipParent,
        SeededRelationshipTypeTerm.child: Text.shareRelationshipChild,
        SeededRelationshipTypeTerm.grandparent: Text.shareRelationshipGrandparent,
        SeededRelationshipTypeTerm.grandchild: Text.shareRelationshipGrandchild,
        SeededRelationshipTypeTerm.pibling: Text.shareRelationshipPibling,
        SeededRelationshipTypeTerm.nibling: Text.shareRelationshipNibling,
        SeededRelationshipTypeTerm.guardian: Text.shareRelationshipGuardian,
        SeededRelationshipTypeTerm.ward: Text.shareRelationshipWard,
    ]

    /// A term as the sheet names it: a product term by its localized name, a researcher's by its label.
    private static func termName(_ term: CatalogEvidenceTerm, property: String) -> String {
        PropertyTermDisplay.name(key: term.key, origin: term.origin, propertyKey: property, catalogLabel: term.label)
    }

    /// The seeded wording of a term key; nil for a key `table` does not know.
    private static func seededShare(_ key: String?, in table: [String: ShareTerm], similar: Bool) -> String? {
        key.flatMap { table[$0] }.map { Text.sharePhrase($0, similar: similar) }
    }

    /// A term with no seeded wording (a researcher's), named as it is: "a connection (Godparent)".
    /// nil when it has no label.
    private static func namedShare(_ label: String?, similar: Bool) -> String? {
        guard let label = label?.trimmingCharacters(in: .whitespacesAndNewlines), !label.isEmpty else { return nil }
        return Text.shareConnectionNamed(term: label, similar: similar)
    }

    static let entries: [KindEntry] = [
        KindEntry(
            kind: .location,
            termProperty: SeededPropertyKey.role,
            sentence: { Text.bridgeSummaryLocation(event: $0.a, place: $0.b) },
            phrase: { _ in L10n.string(Text.bridgeSummaryLocationBare) },
            namesTerm: false,
            share: { input, similar in
                // The neighbor is a place only when it is the place endpoint.
                guard input.neighborEnd == SeededPropertyKey.place else { return nil }
                return Text.shareLocation(place: input.neighborName, similar: similar)
            }
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
            },
            namesTerm: true,
            share: { input, similar in
                // The engine read the term from the neighbor's side. A term that names the
                // row (a researcher's, with no inverse) is only named: "a connection (Godparent)".
                guard let term = input.neighborTerm else { return nil }
                if input.namesNeighbor,
                   let phrase = seededShare(term.productKey, in: relationshipShares, similar: similar) {
                    return phrase
                }
                return namedShare(termName(term, property: SeededPropertyKey.relationshipType), similar: similar)
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
            },
            namesTerm: true,
            share: { input, similar in
                // The term names the row when the neighbor is the "to" end ("Leeds is part of
                // Yorkshire": from Leeds, Yorkshire is a parent place) and the neighbor otherwise.
                switch (input.neighborTerm?.productKey, input.namesNeighbor) {
                case (SeededPlaceRelationshipTypeTerm.partOf, false): return Text.shareParentPlace(similar: similar)
                case (SeededPlaceRelationshipTypeTerm.partOf, true): return Text.sharePlaceWithin(similar: similar)
                case (SeededPlaceRelationshipTypeTerm.succeededBy, false): return Text.shareSuccessorPlace(similar: similar)
                case (SeededPlaceRelationshipTypeTerm.succeededBy, true): return Text.sharePredecessorPlace(similar: similar)
                default: return nil
                }
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
            },
            namesTerm: true,
            share: { input, similar in
                // A person neighbor is worded by its role and name; an event neighbor, by its type.
                if input.neighborEnd == SeededPropertyKey.person {
                    guard let term = input.neighborTerm else { return nil }
                    if let role = seededShare(term.productKey, in: roleShares, similar: similar) {
                        return Text.shareNamed(phrase: role, name: input.neighborName)
                    }
                    return namedShare(termName(term, property: SeededPropertyKey.role), similar: similar)
                }
                return seededShare(input.eventType?.productKey, in: eventShares, similar: similar)
                    ?? namedShare(input.eventType.map { termName($0, property: SeededPropertyKey.eventType) }, similar: similar)
                    ?? Text.shareEventUntyped(similar: similar)
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
        sentence: @escaping @Sendable (Slots) -> String
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
        phrase(kind: kind, lowercasedTerm: termInLowercase(term), termKey: termKey, eventType: eventType, in: sentences)
    }

    private static func phrase(
        kind: EvidenceBridgeKind,
        lowercasedTerm: String?,
        termKey: String?,
        eventType: String?,
        in sentences: [Sentence]
    ) -> String {
        let specific = lookup(kind: kind, term: termKey, eventType: eventType, in: sentences)
        return (specific?.phrase ?? entry(kind).phrase)(lowercasedTerm)
    }

    /// The term as a reason names the connection: "spouse", "subject". nil
    /// when the kind or the term gives nothing to name.
    static func reasonRole(kind: EvidenceBridgeKind, term: String?) -> String? {
        entry(kind).namesTerm ? termInLowercase(term) : nil
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
        let term = termInLowercase(term)
        guard let endpointA, let endpointB else {
            return phrase(kind: kind, lowercasedTerm: term, termKey: termKey, eventType: eventType, in: sentences)
        }
        let slots = Slots(a: endpointA, b: endpointB, termLabel: term)
        if let specific = lookup(kind: kind, term: termKey, eventType: eventType, in: sentences) {
            return specific.sentence(slots)
        }
        let entry = entry(kind)
        if term != nil { return entry.sentence(slots) }
        return (entry.noTerm ?? entry.sentence)(slots)
    }

    /// What a row shares with its neighbor, after "they share": "a spouse", "a similar birth
    /// event". "a connection" when nothing names it (no connection on the page, or one this kind
    /// cannot word from that side).
    static func share(_ input: ShareInput?, similar: Bool) -> String {
        guard let input, let phrase = entry(input.kind).share(input, similar) else {
            return Text.shareConnection(similar: similar)
        }
        return phrase
    }
}
