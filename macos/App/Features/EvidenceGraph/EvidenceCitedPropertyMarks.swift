import Foundation

/// Conflict / negated marks and extra-row filtering for cited graph cards (S8-06).
enum EvidenceCitedPropertyMarks {
    /// Reserved leading gutter on every cited property row so values align
    /// whether a bracket is painted or not (board: 13px shell + 11px = 24px).
    static let conflictGutterWidth: CGFloat = 11

    /// Leading inset for a cited row. The gutter is added to the shell, not
    /// substituted for it, and is reserved on every row.
    static func leadingInset(shell: CGFloat) -> CGFloat {
        shell + conflictGutterWidth
    }

    /// How many times each `propertyKey` appears on this card.
    static func conflictCounts(in observations: [CatalogObservation]) -> [String: Int] {
        var counts: [String: Int] = [:]
        for observation in observations {
            let key = observation.propertyKey
            guard !key.isEmpty else { continue }
            counts[key, default: 0] += 1
        }
        return counts
    }

    static func isConflicted(_ observation: CatalogObservation, counts: [String: Int]) -> Bool {
        counts[observation.propertyKey, default: 0] >= 2
    }

    static func cardHasConflict(in observations: [CatalogObservation]) -> Bool {
        conflictCounts(in: observations).values.contains { $0 >= 2 }
    }

    static func isNegated(_ observation: CatalogObservation) -> Bool {
        observation.polarity
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .lowercased() == ObservationPolarity.negative.rawValue
    }

    /// Consecutive runs whose `propertyKey` appears ≥ 2 times on the card.
    static func conflictRuns(in observations: [CatalogObservation]) -> [Range<Int>] {
        let counts = conflictCounts(in: observations)
        var runs: [Range<Int>] = []
        var index = 0
        while index < observations.count {
            let key = observations[index].propertyKey
            var end = index + 1
            while end < observations.count, observations[end].propertyKey == key {
                end += 1
            }
            if counts[key, default: 0] >= 2 {
                runs.append(index..<end)
            }
            index = end
        }
        return runs
    }

    static func graphHasConflict(
        subjects: [SourceGraphPlacedSubject],
        bridges: [SourceGraphPlacedBridge]
    ) -> Bool {
        if subjects.contains(where: { cardHasConflict(in: $0.observations) }) {
            return true
        }
        return bridges.contains { cardHasConflict(in: extraObservations(in: $0.observations)) }
    }

    /// Observations that are not the connect sentence (edges + role / type).
    static func extraObservations(
        in observations: [CatalogObservation],
        rules: [CatalogConnectRule] = CatalogConnectRule.productMatrix
    ) -> [CatalogObservation] {
        let sentence = sentencePropertyKeys(rules: rules)
        return observations.filter { !sentence.contains($0.propertyKey) }
    }

    static func sentencePropertyKeys(
        rules: [CatalogConnectRule] = CatalogConnectRule.productMatrix
    ) -> Set<String> {
        var keys = Set<String>()
        for rule in rules where !rule.refuse {
            for edge in rule.edges {
                keys.insert(edge.propertyKey)
            }
            let disambiguation = rule.disambiguation
                .trimmingCharacters(in: .whitespacesAndNewlines)
            if !disambiguation.isEmpty, disambiguation != "none" {
                keys.insert(disambiguation)
            }
        }
        return keys
    }

    static func accessibilityLabel(
        propertyLabel: String,
        spokenValue: String,
        conflictCount: Int?,
        isNegated: Bool
    ) -> String {
        var parts = [propertyLabel, spokenValue]
        if let conflictCount, conflictCount >= 2 {
            parts.append(L10n.EvidenceGraph.conflictOneOf(count: conflictCount))
        }
        if isNegated {
            parts.append(String(localized: L10n.EvidenceGraph.negatedAccessibility))
        }
        return "\(parts.joined(separator: ", ")). \(String(localized: L10n.EvidenceGraph.citedRowEditHint))"
    }
}
