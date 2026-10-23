import Foundation

/// Durable bridge copy: relational **phrase** on cards (S7-D3 §3.1) and full
/// **sentence** wherever a bridge is named (S8-D8 §2.4).
enum EvidenceBridgeEdgeSummary {
    /// Relational phrase only — “is the father of”, “participated as witness”.
    static func phrase(for placed: SourceGraphPlacedBridge) -> String {
        ConnectionSentenceRegistry.phrase(
            kind: placed.kind,
            term: termDisplay(kind: placed.kind, in: placed.observations),
            termKey: termKey(kind: placed.kind, in: placed.observations)
        )
    }

    /// The bridge's term as a reason names it: "spouse", "subject". nil when
    /// the bridge has none.
    static func reasonRole(for placed: SourceGraphPlacedBridge) -> String? {
        ConnectionSentenceRegistry.reasonRole(
            kind: placed.kind,
            term: termDisplay(kind: placed.kind, in: placed.observations)
        )
    }

    /// The bridge's term key, only when it is a product term; nil otherwise.
    static func productTermKey(for placed: SourceGraphPlacedBridge) -> String? {
        termKey(kind: placed.kind, in: placed.observations)
    }

    /// The product key of a participation's event type ("birth"), from the event the bridge
    /// reaches; nil for another kind, an event with no type, or a researcher's type.
    static func eventTypeKey(
        for bridge: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        subjects: [String: SourceGraphPlacedSubject]? = nil
    ) -> String? {
        guard bridge.kind == .participation else { return nil }
        return productKey(of: eventTypeObservation(
            endpointAID: bridge.endpointAID, endpointBID: bridge.endpointBID, subjects: subjects ?? Self.subjects(of: snapshot)
        ))
    }

    /// The label of a participation's event type ("birth"), from the event the bridge reaches; nil
    /// for another kind or an event with no type.
    static func eventTypeLabel(
        for bridge: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        subjects: [String: SourceGraphPlacedSubject]? = nil
    ) -> String? {
        guard bridge.kind == .participation,
              let observation = eventTypeObservation(
                endpointAID: bridge.endpointAID, endpointBID: bridge.endpointBID, subjects: subjects ?? Self.subjects(of: snapshot)
              )
        else { return nil }
        return termLabel(of: observation, propertyKey: SeededPropertyKey.eventType)
    }

    /// The snapshot's placed Subjects by id: build once, then pass to every lookup of one pass.
    static func subjects(of snapshot: SourceGraphSnapshot) -> [String: SourceGraphPlacedSubject] {
        Dictionary(snapshot.subjects.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
    }

    /// Snapshot-aware sentence. Nouns come from cited endpoints; unreadable
    /// edges fall back to the stored label or kind phrase · ref. Never empty.
    static func sentence(
        for bridge: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        orders: NameFormatOrders = NameFormatOrders(),
        subjects: [String: SourceGraphPlacedSubject]? = nil
    ) -> String {
        let subjects = subjects ?? Self.subjects(of: snapshot)
        let resolved = snapshot.bridges.first(where: { $0.id == bridge.id }) ?? bridge
        let endpointAID = bridge.endpointAID ?? resolved.endpointAID
        let endpointBID = bridge.endpointBID ?? resolved.endpointBID
        let endpointA = noun(subjectID: endpointAID, subjects: subjects, orders: orders)
        let endpointB = noun(subjectID: endpointBID, subjects: subjects, orders: orders)
        let term = termDisplay(kind: bridge.kind, in: bridge.observations)
        if endpointA == nil || endpointB == nil {
            return nonEmpty(wholeNameFallback(for: bridge))
        }
        return nonEmpty(
            ConnectionSentenceRegistry.sentence(
                kind: bridge.kind,
                endpointA: endpointA,
                endpointB: endpointB,
                term: term,
                termKey: termKey(kind: bridge.kind, in: bridge.observations),
                eventType: productKey(of: eventTypeObservation(
                    endpointAID: endpointAID, endpointBID: endpointBID, subjects: subjects
                ))
            ),
            fallback: wholeNameFallback(for: bridge)
        )
    }

    static func identityPropertyKey(for kind: EvidencePrimaryKind) -> String {
        switch kind {
        case .person: SeededPropertyKey.name
        case .event: SeededPropertyKey.eventType
        case .place: SeededPropertyKey.toponym
        }
    }

    private static func noun(
        subjectID: String?,
        subjects: [String: SourceGraphPlacedSubject],
        orders: NameFormatOrders
    ) -> String? {
        guard let subjectID, let placed = subjects[subjectID] else { return nil }
        if placed.kind == .event, let title = placed.eventTitle {
            let text = EventTitleDisplay.title(title, orders: orders)
                .trimmingCharacters(in: .whitespacesAndNewlines)
            if !text.isEmpty { return text }
        }
        if let identity = identityNoun(on: placed, orders: orders) {
            return identity
        }
        let label = placed.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if !label.isEmpty { return label }
        let typed = L10n.EvidenceGraph.bridgeNounTypeAndRef(
            type: placed.typeLabel,
            ref: placed.subject.ref
        )
        return typed.isEmpty ? nil : typed
    }

    /// First positive identity Observation, else first, else nil (GV-17).
    private static func identityNoun(
        on placed: SourceGraphPlacedSubject,
        orders: NameFormatOrders
    ) -> String? {
        guard let chosen = identityObservation(on: placed) else { return nil }
        let display = ObservationValueDisplay.string(for: chosen, orders: orders)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        return display.isEmpty ? nil : display
    }

    private static func identityObservation(on placed: SourceGraphPlacedSubject) -> CatalogObservation? {
        let key = identityPropertyKey(for: placed.kind)
        let matches = placed.observations.filter { $0.propertyKey == key }
        return matches.first(where: isPositive) ?? matches.first
    }

    /// The `event_type` Observation of the event a bridge reaches: the A end's, else the B end's.
    /// The one choice of endpoint behind an event type's key and its label.
    private static func eventTypeObservation(
        endpointAID: String?,
        endpointBID: String?,
        subjects: [String: SourceGraphPlacedSubject]
    ) -> CatalogObservation? {
        eventTypeObservation(subjectID: endpointAID, subjects: subjects)
            ?? eventTypeObservation(subjectID: endpointBID, subjects: subjects)
    }

    private static func eventTypeObservation(
        subjectID: String?,
        subjects: [String: SourceGraphPlacedSubject]
    ) -> CatalogObservation? {
        guard let subjectID, let placed = subjects[subjectID], placed.kind == .event else { return nil }
        return identityObservation(on: placed)
    }

    private static func isPositive(_ observation: CatalogObservation) -> Bool {
        let polarity = observation.polarity
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .lowercased()
        return polarity.isEmpty || polarity == ObservationPolarity.positive.rawValue
    }

    private static func wholeNameFallback(for bridge: SourceGraphPlacedBridge) -> String {
        let stored = bridge.subject.label.trimmingCharacters(in: .whitespacesAndNewlines)
        if !stored.isEmpty { return stored }
        return L10n.EvidenceGraph.bridgeNameKindAndRef(
            phrase: phrase(for: bridge),
            ref: bridge.subject.ref
        )
    }

    private static func nonEmpty(_ value: String, fallback: String = "") -> String {
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        if !trimmed.isEmpty { return trimmed }
        let next = fallback.trimmingCharacters(in: .whitespacesAndNewlines)
        if !next.isEmpty { return next }
        return "—"
    }

    private static func termObservation(
        kind: EvidenceBridgeKind,
        in observations: [CatalogObservation]
    ) -> CatalogObservation? {
        let key = ConnectionSentenceRegistry.termProperty(for: kind)
        return observations.first(where: { $0.propertyKey == key })
    }

    /// The bridge's term key, only when it is a product term: the sentence
    /// registry matches product keys, never a researcher's.
    private static func termKey(kind: EvidenceBridgeKind, in observations: [CatalogObservation]) -> String? {
        productKey(of: termObservation(kind: kind, in: observations))
    }

    private static func productKey(of observation: CatalogObservation?) -> String? {
        guard let observation, CatalogOrigin.isProduct(observation.valueTermOrigin),
              !observation.valueTermKey.isEmpty
        else { return nil }
        return observation.valueTermKey
    }

    private static func termDisplay(
        kind: EvidenceBridgeKind,
        in observations: [CatalogObservation]
    ) -> String? {
        guard let observation = termObservation(kind: kind, in: observations) else {
            return nil
        }
        return termLabel(of: observation, propertyKey: ConnectionSentenceRegistry.termProperty(for: kind))
    }

    private static func termLabel(of observation: CatalogObservation, propertyKey key: String) -> String? {
        let catalog = ObservationValueDisplay.string(for: observation)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let termKey = observation.valueTermKey
        if !termKey.isEmpty {
            return PropertyTermDisplay.name(
                key: termKey,
                origin: observation.valueTermOrigin,
                propertyKey: key,
                catalogLabel: catalog
            )
        }
        return catalog.isEmpty ? nil : catalog
    }
}
