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
    /// has no structured header.
    static func title(for hit: CatalogSearchHit, orders: NameFormatOrders = NameFormatOrders()) -> String {
        switch hit.header {
        case .person(let header): return PersonHeaderDisplay.title(header, style: .sorted, orders: orders)
        case .event(let header): return EventTitleDisplay.title(header, orders: orders)
        case .place(let header): return PlaceTitleDisplay.title(header)
        case nil: return hit.title
        }
    }

    /// Secondary line the list would show.
    static func secondary(for hit: CatalogSearchHit) -> String {
        switch hit.header {
        case .person(let header): return PersonLifeDisplay.line(header).text
        case .event(let header): return EventSecondaryDisplay.line(header)
        case .place(let header):
            return PlaceChainDisplay.line(parents: header.parents, candidates: header.parentsAreCandidates)
        case nil: return hit.subtitle
        }
    }

    /// The lead thumbnail, or nil for a `source` hit, which `CatalogSearchHitRow`
    /// draws from the Source's own image.
    static func leadThumbnail(for hit: CatalogSearchHit) -> PVThumbnail.Content? {
        switch hit.searchKind {
        case .source:
            return nil
        case .person:
            return ConclusionListRow.thumbnail(mark: EvidencePrimaryKind.person.markKey)
        case .event:
            return ConclusionListRow.thumbnail(mark: EvidencePrimaryKind.event.markKey)
        case .place:
            return ConclusionListRow.thumbnail(mark: EvidencePrimaryKind.place.markKey)
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
