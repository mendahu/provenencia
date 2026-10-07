import Foundation

/// The parent chain on a Place row. Nearest parent first, joined. Empty when
/// the place has no parents. Hierarchy truncation waits for S9-40.
enum PlaceChainDisplay {
    static func line(parents: [String]) -> String {
        parents
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .filter { !$0.isEmpty }
            .joined(separator: L10n.string(L10n.Conclusions.chainSeparator))
    }
}
