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

    static func ordered(_ headers: [CatalogPlaceHeader]) -> [CatalogPlaceHeader] {
        PlaceListOrder.sorted(headers)
    }

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

/// Displayed-title order for the Places list. A name or label sorts
/// case- and diacritic-insensitively; a ref-only row sorts last, by ref.
enum PlaceListOrder {
    static func sorted(_ headers: [CatalogPlaceHeader], locale: Locale = .current) -> [CatalogPlaceHeader] {
        headers.sorted { a, b in
            switch (sortTitle(a), sortTitle(b)) {
            case (nil, nil):
                return refOrder(a, b, locale: locale)
            case (.some, .none):
                return true
            case (.none, .some):
                return false
            case let (left?, right?):
                let order = left.compare(right, options: [.caseInsensitive, .diacriticInsensitive], locale: locale)
                if order == .orderedSame {
                    return refOrder(a, b, locale: locale)
                }
                return order == .orderedAscending
            }
        }
    }

    /// The name or label the row shows. Nil when the title is the ref.
    private static func sortTitle(_ header: CatalogPlaceHeader) -> String? {
        switch PlaceTitleDisplay.titleSource(header.titleParts) {
        case .name(let text), .label(let text):
            return text
        case .ref:
            return nil
        }
    }

    private static func refOrder(_ a: CatalogPlaceHeader, _ b: CatalogPlaceHeader, locale: Locale) -> Bool {
        a.entity.ref.compare(b.entity.ref, options: .caseInsensitive, locale: locale) == .orderedAscending
    }
}

/// The Places row's parent chain, italic, omitted when the place has none.
struct PlaceChainLine: View {
    let text: String

    var body: some View {
        Text(verbatim: text)
            .italic()
            .lineLimit(1)
            .frame(maxHeight: text.isEmpty ? 0 : nil)
            .accessibilityHidden(text.isEmpty)
    }
}
