import Foundation

/// Compiled titles the name splitter may peel. Not a catalog table: matching
/// does not read it, and researchers cannot add to it. A token is a title only
/// when another word remains; the splitter decides which end.
enum NameTitleRegistry {
    /// Case-insensitive, and one trailing period does not count (`Dr` and `Dr.`).
    static func lookupKey(_ token: String) -> String {
        var key = token.trimmingCharacters(in: .whitespacesAndNewlines)
        if key.hasSuffix(".") {
            key.removeLast()
        }
        return key.lowercased()
    }

    static func isPrefix(_ token: String) -> Bool {
        prefixes.contains(lookupKey(token))
    }

    static func isSuffix(_ token: String) -> Bool {
        suffixes.contains(lookupKey(token))
    }

    /// Leading titles. `sr` here is Sister; a trailing `Sr` is Senior.
    static let prefixes: Set<String> = [
        "mr", "mrs", "ms", "miss", "mx",
        "dr", "rev", "hon", "sir", "dame", "prof", "fr",
        "capt", "col", "gen",
        "sr",
    ]

    /// Trailing generation markers. `sr` here is Senior.
    static let suffixes: Set<String> = [
        "jr", "sr", "junior", "senior", "esq",
        "i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
    ]
}
