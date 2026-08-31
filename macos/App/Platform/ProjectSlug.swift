import Foundation

/// Mirrors Go `onboarding.FolderName` for create-flow slug previews.
enum ProjectSlug {
    private static let suffix = ".provenencia"

    /// Returns `{kebab-slug}.provenencia` for a project label.
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
            return suffix
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
