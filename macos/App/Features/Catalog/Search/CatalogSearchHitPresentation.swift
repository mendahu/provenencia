import Foundation

/// Maps a SearchCatalog hit into result row copy (kind-agnostic slots).
enum CatalogSearchHitPresentation {
    /// A kind this build doesn't know (nil) reads as a Source.
    static func kindLabel(for kind: CatalogSearchKind?) -> LocalizedStringResource {
        switch kind {
        case .source, nil: L10n.Workspace.omnibarKindSource
        case .sourceType: L10n.Workspace.omnibarKindType
        case .metadataField: L10n.Workspace.omnibarKindMetadataField
        case .person: L10n.Workspace.omnibarKindPerson
        case .event: L10n.Workspace.omnibarKindEvent
        case .place: L10n.Workspace.omnibarKindPlace
        case .evidenceGraph: L10n.Workspace.omnibarKindEvidenceGraph
        }
    }

    /// Whether the match-context slot should show for this stable field code.
    static func showsMatchContext(field: String) -> Bool {
        switch field.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() {
        case "notes", "metadata", "filename", "description", "name", "place":
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
            return L10n.string(L10n.Workspace.omnibarMatchDescription)
        case "name":
            return L10n.Workspace.omnibarMatchName(snippet: trimmed)
        case "place":
            return L10n.Workspace.omnibarMatchPlace(snippet: trimmed)
        default:
            return ""
        }
    }

    static func refAccent(for hit: CatalogSearchHit) -> Bool {
        hit.matchReason == "ref" && !hit.ref.isEmpty
    }

    /// VoiceOver label for a result row — title, kind, subtitle, ref, match context.
    static func accessibilityLabel(for hit: CatalogSearchHit, orders: NameFormatOrders = NameFormatOrders()) -> String {
        var parts = [
            title(for: hit, orders: orders),
            L10n.string(kindLabel(for: hit.searchKind)),
        ]
        let subtitle = secondary(for: hit).trimmingCharacters(in: .whitespacesAndNewlines)
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

    /// Title the list would show. Falls back to the indexed title when the hit
    /// has no structured summary.
    static func title(for hit: CatalogSearchHit, orders: NameFormatOrders = NameFormatOrders()) -> String {
        hit.summary.map { ConclusionSummaryDisplay.title($0, style: .sorted, orders: orders) } ?? hit.title
    }

    /// Secondary line the list would show.
    static func secondary(for hit: CatalogSearchHit) -> String {
        hit.summary.map(ConclusionSummaryDisplay.secondary) ?? hit.subtitle
    }

    /// The image path `CatalogSearchHitRow` loads when `leadThumbnail` is nil:
    /// a conclusion's lead picture (what its list row shows), else the Source cover.
    static func leadRelPath(for hit: CatalogSearchHit) -> String {
        let picture = hit.summary?.pictureImageRelPath ?? ""
        return picture.isEmpty ? hit.thumbnailRelPath : picture
    }

    /// The mark a conclusion hit falls back to when its picture cannot load.
    static func fallbackMark(for hit: CatalogSearchHit) -> PVMarkKey? {
        switch hit.searchKind {
        case .person: EvidencePrimaryKind.person.markKey
        case .event: EvidencePrimaryKind.event.markKey
        case .place: EvidencePrimaryKind.place.markKey
        default: nil
        }
    }

    /// The lead thumbnail, or nil when `CatalogSearchHitRow` draws an image:
    /// a `source` hit's cover, or a person/event/place hit's lead picture.
    static func leadThumbnail(for hit: CatalogSearchHit) -> PVThumbnail.Content? {
        if !(hit.summary?.pictureImageRelPath ?? "").trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return nil
        }
        switch hit.searchKind {
        case .source:
            return nil
        case .person, .event, .place:
            return fallbackMark(for: hit).map { ConclusionListRow.thumbnail(mark: $0) }
        case .sourceType:
            if hit.iconKey.isEmpty { return PVThumbnail.Content(glyph: WorkspaceSection.sourceTypes.glyph) }
            return PVThumbnail.Content(
                glyph: .mark(PVMarkKey(catalogKey: hit.iconKey)),
                label: kindLabel(for: .sourceType)
            )
        case .metadataField:
            return PVThumbnail.Content(glyph: WorkspaceSection.metadata.glyph)
        case .evidenceGraph:
            return PVThumbnail.Content(glyph: .symbol(.gitBranch))
        case nil:
            return PVThumbnail.Content(glyph: .symbol(.search))
        }
    }
}
