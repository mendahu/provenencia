import SwiftUI

/// The Places list (S9-26, board S9-D4). The Places place, query key, and
/// history entry stay Places. The chain line stays empty until S9-40.
typealias PlacesListView = ConclusionListPage<PlacesList>

enum PlacesList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .placesList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.placesTitle }
    static var identifierRoot: String { "places" }
    static var emptyIcon: PVSymbol { .mapPin }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.placesEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.placesEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.placeCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.placeCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectPlace }

    static func ref(_ header: CatalogPlaceHeader) -> String { header.entity.ref }

    static func titleSource(_ header: CatalogPlaceHeader) -> ConclusionTitleSource {
        PlaceTitleDisplay.titleSource(header.titleParts)
    }

    static func extraCount(_ header: CatalogPlaceHeader) -> Int { header.extraNameCount }

    static func accessibilityLabel(_ header: CatalogPlaceHeader) -> String {
        L10n.Workspace.placeRowAccessibility(
            title: titleSource(header).text,
            extra: header.extraNameCount,
            chain: PlaceChainDisplay.line(parents: header.parents),
            ref: header.entity.ref
        )
    }

    static func location(_ header: CatalogPlaceHeader) -> WorkspaceLocation {
        .placeDetail(entityId: header.entity.id, ref: header.entity.ref, title: titleSource(header).text)
    }

    @MainActor
    static func secondary(_ header: CatalogPlaceHeader) -> PlaceChainLine {
        PlaceChainLine(text: PlaceChainDisplay.line(parents: header.parents))
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
