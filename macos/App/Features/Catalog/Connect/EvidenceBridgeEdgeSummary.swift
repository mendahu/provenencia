import Foundation

/// Durable bridge copy: relational **phrase** on cards (S7-D3 §3.1) and full
/// **sentence** wherever a bridge is named (S8-D8 §2.4).
enum EvidenceBridgeEdgeSummary {
    /// Relational phrase only — “is the father of”, “participated as witness”.
    static func phrase(for placed: SourceGraphPlacedBridge) -> String {
        ConnectionSentenceRegistry.phrase(
            kind: placed.kind,
            term: termDisplay(kind: placed.kind, in: placed.observations)
        )
    }

    /// Snapshot-aware sentence. Nouns come from cited endpoints; unreadable
    /// edges fall back to the stored label or kind phrase · ref. Never empty.
    static func sentence(
        for bridge: SourceGraphPlacedBridge,
        in snapshot: SourceGraphSnapshot,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        let resolved = snapshot.bridges.first(where: { $0.id == bridge.id }) ?? bridge
        let endpointA = noun(
            subjectID: bridge.endpointAID ?? resolved.endpointAID,
            snapshot: snapshot,
            orders: orders
        )
        let endpointB = noun(
            subjectID: bridge.endpointBID ?? resolved.endpointBID,
            snapshot: snapshot,
            orders: orders
        )
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
                eventType: eventTypeKey(
                    subjectID: bridge.kind == .participation
                        ? (bridge.endpointBID ?? resolved.endpointBID)
                        : nil,
                    snapshot: snapshot
                )
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
        snapshot: SourceGraphSnapshot,
        orders: NameFormatOrders
    ) -> String? {
        guard let subjectID,
              let placed = snapshot.subjects.first(where: { $0.id == subjectID })
        else { return nil }
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

    /// The placed event's `event_type` term key, positive first, else nil.
    private static func eventTypeKey(subjectID: String?, snapshot: SourceGraphSnapshot) -> String? {
        guard let subjectID,
              let placed = snapshot.subjects.first(where: { $0.id == subjectID }),
              placed.kind == .event,
              let key = identityObservation(on: placed)?.valueTermKey,
              !key.isEmpty
        else { return nil }
        return key
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

    private static func termKey(kind: EvidenceBridgeKind, in observations: [CatalogObservation]) -> String? {
        guard let key = termObservation(kind: kind, in: observations)?.valueTermKey, !key.isEmpty else {
            return nil
        }
        return key
    }

    private static func termDisplay(
        kind: EvidenceBridgeKind,
        in observations: [CatalogObservation]
    ) -> String? {
        let key = ConnectionSentenceRegistry.termProperty(for: kind)
        guard let observation = termObservation(kind: kind, in: observations) else {
            return nil
        }
        let catalog = ObservationValueDisplay.string(for: observation)
            .trimmingCharacters(in: .whitespacesAndNewlines)
        let termKey = observation.valueTermKey
        if !termKey.isEmpty {
            return PropertyTermDisplay.name(
                key: termKey,
                propertyKey: key,
                catalogLabel: catalog
            )
        }
        return catalog.isEmpty ? nil : catalog
    }
}
