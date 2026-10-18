import Foundation

/// Compiled titles the name splitter may peel. Not a catalog table: matching
/// does not read it, and researchers cannot add to it. A token is a title only
/// when another word remains; the splitter decides which end.
enum NameTitleRegistry {
    /// Case and punctuation do not count (`Dr`, `Dr.`, `D.r.`).
    static func isPrefix(_ token: String) -> Bool {
        prefixes.contains(TextMatch.key(token, at: .same, locale: TextMatch.posix))
    }

    static func isSuffix(_ token: String) -> Bool {
        suffixes.contains(TextMatch.key(token, at: .same, locale: TextMatch.posix))
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
