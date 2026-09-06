import Foundation

/// Mirrors Go `onboarding.FolderName` for create-flow slug previews.
/// Empty / unslugifiable labels return `""` (Go returns `ErrInvalidFamilyName`).
enum ProjectSlug {
    private static let suffix = ".provenencia"

    /// Shared success cases with `core/onboarding.TestFolderName` (keep in sync).
    static let folderNameFixtures: [(label: String, folder: String)] = [
        ("Robins Family", "robins-family.provenencia"),
        ("Robins/Family", "robins-family.provenencia"),
        ("Robins\\Family", "robins-family.provenencia"),
        ("Robins:Family", "robins-family.provenencia"),
        ("Robins*Family", "robins-family.provenencia"),
        ("Robins?Family", "robins-family.provenencia"),
        ("Robins\"Family", "robins-family.provenencia"),
        ("Robins<Family>", "robins-family.provenencia"),
        ("Robins|Family", "robins-family.provenencia"),
        ("...Robins", "robins.provenencia"),
        ("Robins Family.provenencia", "robins-family.provenencia"),
        ("  Robins   Family  ", "robins-family.provenencia"),
        ("García Family", "garcía-family.provenencia"),
        ("A/B", "a-b.provenencia"),
    ]

    /// Returns `{kebab-slug}.provenencia`, or `""` when the label cannot form a slug.
    static func folderName(from label: String) -> String {
        var s = label.trimmingCharacters(in: .whitespacesAndNewlines)
        if s.lowercased().hasSuffix(suffix) {
            s = String(s.dropLast(suffix.count)).trimmingCharacters(in: .whitespacesAndNewlines)
        }
        var out = ""
        var prevHyphen = false
        for ch in s.lowercased() {
            if ch.isLetter || ch.isNumber {
                out.append(ch)
                prevHyphen = false
                continue
            }
            if !out.isEmpty && !prevHyphen {
                out.append("-")
                prevHyphen = true
            }
        }
        while out.hasPrefix("-") { out.removeFirst() }
        while out.hasSuffix("-") { out.removeLast() }
        if out.isEmpty {
            return ""
        }
        return out + suffix
    }

    static func labelFromFolder(_ path: String) -> String {
        var base = URL(fileURLWithPath: path).lastPathComponent
        if base.lowercased().hasSuffix(suffix) {
            base = String(base.dropLast(suffix.count))
        }
        return base.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
