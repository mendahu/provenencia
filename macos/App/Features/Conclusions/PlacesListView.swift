import SwiftUI

/// The Places list (S9-26, board S9-D4). The Places place, query key, and
/// history entry stay Places. The chain line reads today's parents from Go.
typealias PlacesListView = ConclusionListPage<PlacesList>

enum PlacesList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .placesList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.placesTitle }
    static var identifierRoot: String { "places" }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.placesEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.placesEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.placeCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.placeCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectPlace }

    static func ref(_ summary: CatalogPlaceSummary) -> String { summary.entity.ref }
    static func updatedAt(_ summary: CatalogPlaceSummary) -> String { summary.updated.at }
    static func thumbnailRelPath(_ summary: CatalogPlaceSummary) -> String { summary.pictureImageRelPath }

    static func titleSource(_ summary: CatalogPlaceSummary) -> ConclusionTitleSource {
        PlaceTitleDisplay.titleSource(summary.titleParts)
    }

    static func extraCount(_ summary: CatalogPlaceSummary) -> Int { summary.extraNameCount }

    static func accessibilityLabel(_ summary: CatalogPlaceSummary) -> String {
        L10n.Workspace.placeRowAccessibility(
            title: titleSource(summary).text,
            extra: summary.extraNameCount,
            chain: PlaceChainDisplay.line(
                parents: summary.parents,
                candidates: summary.parentsAreCandidates
            ),
            ref: summary.entity.ref
        )
    }

    static func location(_ summary: CatalogPlaceSummary) -> WorkspaceLocation {
        .placeDetail(entityId: summary.entity.id, ref: summary.entity.ref, title: PlaceTitleDisplay.title(summary))
    }

    @MainActor
    static func secondary(_ summary: CatalogPlaceSummary) -> PlaceChainLine {
        PlaceChainLine(text: PlaceChainDisplay.line(
            parents: summary.parents,
            candidates: summary.parentsAreCandidates
        ))
    }
}

/// The Places row's parent chain, italic, omitted when the place has none.
struct PlaceChainLine: View {
    let text: String

    /// Nothing when the place has no parents, so `PVList` lays out no second line.
    var body: some View {
        if !text.isEmpty {
            Text(verbatim: text)
                .italic()
                .lineLimit(1)
        }
    }
}
