import Foundation

/// Mirrors Go `core/slug.Kebab` for the Metadata add-flow's live key
/// preview. Empty / unslugifiable labels return `""` (Go returns `""` too —
/// the FFI call then fails with `metadatafields.invalid`).
enum FieldSlug {
    /// Shared success cases with `core/slug.TestKebab` (keep in sync).
    static let fixtures: [(label: String, key: String)] = [
        ("Certificate number", "certificate-number"),
        ("Grandma's album code", "grandmas-album-code"),
        ("Grandma\u{2019}s album code", "grandmas-album-code"),
        ("A/B", "a-b"),
        ("  Scanned   at church  ", "scanned-at-church"),
    ]

    /// Returns a kebab-case slug of `label`, or `""` when nothing survives.
    static func kebab(_ label: String) -> String {
        var out = ""
        var prevHyphen = false
        for ch in label.lowercased() {
            if ch == "'" || ch == "\u{2019}" {
                continue
            }
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
        return out
    }
}
