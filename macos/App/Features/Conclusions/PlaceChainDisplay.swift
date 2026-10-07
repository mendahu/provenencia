import Foundation

/// The parent chain on a Place row or life/event place line. Nearest parent
/// first. Several simultaneous parents join with "or"; a path joins with
/// commas. Long chains drop ancestors from the top (PLL-3).
enum PlaceChainDisplay {
    /// Soft limit on how many names fit one secondary line before dropping
    /// the furthest ancestors (nearest parents stay).
    static let maxSegments = 4

    static func line(
        parents: [String],
        candidates: Bool = false,
        maxSegments: Int = maxSegments
    ) -> String {
        let parts = parents
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
        guard !parts.isEmpty else { return "" }
        let shown: [String]
        if candidates || parts.count <= maxSegments {
            shown = parts
        } else {
            // Truncate from the top of the hierarchy (furthest ancestors).
            shown = Array(parts.prefix(maxSegments))
        }
        if candidates {
            return joinCandidates(shown)
        }
        return shown.joined(separator: L10n.string(L10n.Conclusions.chainSeparator))
    }

    /// Leaf name plus parent chain for a folded Location (Person / Event rows).
    static func placeLine(_ place: CatalogHeaderPlace, maxSegments: Int = maxSegments) -> String {
        let leaf = place.names
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .first { !$0.isEmpty } ?? ""
        let chain = line(
            parents: place.parents,
            candidates: place.parentsAreCandidates,
            maxSegments: max(1, maxSegments - (leaf.isEmpty ? 0 : 1))
        )
        switch (leaf.isEmpty, chain.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return leaf
        case (true, false):
            return chain
        case (false, false):
            return leaf + L10n.string(L10n.Conclusions.chainSeparator) + chain
        }
    }

    private static func joinCandidates(_ names: [String]) -> String {
        guard names.count > 1 else { return names.first ?? "" }
        if names.count == 2 {
            return L10n.Conclusions.chainOrPair(first: names[0], second: names[1])
        }
        let head = names.dropLast().joined(separator: L10n.string(L10n.Conclusions.chainSeparator))
        return L10n.Conclusions.chainOrPair(first: head, second: names[names.count - 1])
    }
}
