import Foundation

/// Compiled titles the name splitter may peel. Not a catalog table: matching
/// does not read it, and researchers cannot add to it. A token is a title only
/// when another word remains; the splitter decides which end.
enum NameTitleRegistry {
    /// Case and edge punctuation do not count (`Dr`, `Dr.`, `Jr.,`); dotted
    /// initials (`S.R.`) are never titles.
    static func isPrefix(_ token: String) -> Bool {
        !hasDottedInitials(token) && prefixes.contains(TextMatch.key(token, at: .same, locale: TextMatch.posix))
    }

    static func isSuffix(_ token: String) -> Bool {
        !hasDottedInitials(token) && suffixes.contains(TextMatch.key(token, at: .same, locale: TextMatch.posix))
    }

    /// A period with a letter on each side: `S.R.` reads as initials, not `Sr`.
    private static func hasDottedInitials(_ token: String) -> Bool {
        let chars = Array(token)
        guard chars.count >= 3 else { return false }
        return (1..<chars.count - 1).contains { chars[$0] == "." && chars[$0 - 1].isLetter && chars[$0 + 1].isLetter }
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
