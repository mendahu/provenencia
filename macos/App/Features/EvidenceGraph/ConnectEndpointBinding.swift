import Foundation

/// Binds Connect endpoints to a rule's ordered edges and names a pending sentence.
enum ConnectEndpointBinding {
    struct Bound: Equatable {
        var propertyKey: String
        var endpointTypeKey: String
        var subjectID: String
        var label: String
    }

    /// Pair each rule edge with one of the two endpoints.
    static func bind(
        rule: CatalogConnectRule,
        fromID: String,
        fromTypeKey: String,
        fromLabel: String,
        toID: String,
        toTypeKey: String,
        toLabel: String
    ) -> [Bound]? {
        guard !rule.refuse, rule.edges.count >= 2 else { return nil }
        let distinctTypes = rule.edges[0].endpointTypeKey != rule.edges[1].endpointTypeKey
        var bound: [Bound] = []
        for (index, edge) in rule.edges.enumerated() {
            let useFrom: Bool
            if distinctTypes {
                if edge.endpointTypeKey == fromTypeKey, fromTypeKey != toTypeKey {
                    useFrom = true
                } else if edge.endpointTypeKey == toTypeKey, fromTypeKey != toTypeKey {
                    useFrom = false
                } else if edge.endpointTypeKey == fromTypeKey {
                    useFrom = true
                } else if edge.endpointTypeKey == toTypeKey {
                    useFrom = false
                } else {
                    return nil
                }
            } else {
                useFrom = index == 0
            }
            bound.append(
                Bound(
                    propertyKey: edge.propertyKey,
                    endpointTypeKey: edge.endpointTypeKey,
                    subjectID: useFrom ? fromID : toID,
                    label: useFrom ? fromLabel : toLabel
                )
            )
        }
        return bound
    }

    /// Pending Connect sentence from the rule and both endpoints (no term yet).
    static func pendingSentence(
        rule: CatalogConnectRule,
        fromID: String,
        fromTypeKey: String,
        fromLabel: String,
        toID: String,
        toTypeKey: String,
        toLabel: String
    ) -> String {
        let kind = EvidenceBridgeKind(rawValue: rule.bridgeTypeKey) ?? .participation
        let bound = bind(
            rule: rule,
            fromID: fromID,
            fromTypeKey: fromTypeKey,
            fromLabel: fromLabel,
            toID: toID,
            toTypeKey: toTypeKey,
            toLabel: toLabel
        )
        return sentence(kind: kind, endpointA: bound?[0].label, endpointB: bound?[1].label, term: nil)
    }

    /// Kind-ordered A/B nouns. A is the first rule edge; B is the second.
    static func sentence(
        kind: EvidenceBridgeKind,
        endpointA: String?,
        endpointB: String?,
        term: String?
    ) -> String {
        switch kind {
        case .location:
            return EvidenceBridgeEdgeSummary.sentence(
                kind: kind,
                person: nil,
                related: nil,
                event: endpointA,
                place: endpointB,
                term: term
            )
        case .relationship:
            return EvidenceBridgeEdgeSummary.sentence(
                kind: kind,
                person: endpointA,
                related: endpointB,
                event: nil,
                place: nil,
                term: term
            )
        case .participation:
            return EvidenceBridgeEdgeSummary.sentence(
                kind: kind,
                person: endpointA,
                related: nil,
                event: endpointB,
                place: nil,
                term: term
            )
        }
    }
}
