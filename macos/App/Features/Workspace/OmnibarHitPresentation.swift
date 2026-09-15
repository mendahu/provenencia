import Foundation

/// Maps a SearchCatalog hit into omnibar row copy (kind-agnostic slots).
enum OmnibarHitPresentation {
    static func kindLabel(for kind: String) -> LocalizedStringResource {
        switch kind {
        case "source": L10n.Workspace.omnibarKindSource
        case "source_type": L10n.Workspace.omnibarKindType
        case "source_field": L10n.Workspace.omnibarKindField
        default: L10n.Workspace.omnibarKindSource
        }
    }

    /// Whether the match-context slot should show for this stable field code.
    static func showsMatchContext(field: String) -> Bool {
        switch field.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() {
        case "notes", "metadata", "filename", "description":
            return true
        default:
            return false
        }
    }

    /// Localized match-context line from Go field code + optional raw snippet.
    static func matchContextText(field: String, snippet: String) -> String {
        let code = field.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        guard showsMatchContext(field: code) else { return "" }
        let trimmed = snippet.trimmingCharacters(in: .whitespacesAndNewlines)
        switch code {
        case "notes":
            return L10n.Workspace.omnibarMatchNote(snippet: trimmed)
        case "metadata":
            return L10n.Workspace.omnibarMatchMetadata(snippet: trimmed)
        case "filename":
            return L10n.Workspace.omnibarMatchFilename(snippet: trimmed)
        case "description":
            return String(localized: L10n.Workspace.omnibarMatchDescription)
        default:
            return ""
        }
    }

    static func refAccent(for hit: CatalogSearchHit) -> Bool {
        hit.matchReason == "ref" && !hit.ref.isEmpty
    }

    /// VoiceOver label for a result row — title, kind, subtitle, ref, match context.
    static func accessibilityLabel(for hit: CatalogSearchHit) -> String {
        var parts = [
            hit.title,
            String(localized: kindLabel(for: hit.kind)),
        ]
        let subtitle = hit.subtitle.trimmingCharacters(in: .whitespacesAndNewlines)
        if !subtitle.isEmpty {
            parts.append(subtitle)
        }
        let ref = hit.ref.trimmingCharacters(in: .whitespacesAndNewlines)
        if !ref.isEmpty {
            parts.append(ref)
        }
        let matchContext = matchContextText(field: hit.matchReason, snippet: hit.matchSnippet)
        if !matchContext.isEmpty {
            parts.append(matchContext)
        }
        let formatter = ListFormatter()
        formatter.locale = .current
        return formatter.string(from: parts) ?? parts.joined(separator: ", ")
    }

    static func leadSymbol(for kind: String) -> PVSymbol {
        switch kind {
        case "source": .scrollText
        case "source_type": .library
        case "source_field": .tag
        default: .search
        }
    }
}
