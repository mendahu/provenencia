import SwiftUI

/// The Places list (S9-26, board S9-D4). A configuration of
/// `ConclusionListPage`. The Places place, query key, and history entry stay
/// Places. The chain line stays empty until S9-40.
struct PlacesListView: View {
    let session: WorkspaceSession

    var body: some View {
        ConclusionListPage<CatalogPlaceHeader, PlaceChainLine>(
            session: session,
            key: .placesList(project: session.projectKey),
            title: L10n.Workspace.placesTitle,
            pageAccessibilityIdentifier: "places.list",
            emptyIcon: .mapPin,
            emptyTitle: L10n.Workspace.placesEmptyTitle,
            emptyMessage: L10n.Workspace.placesEmptyMessage,
            emptyAccessibilityIdentifier: "places.empty",
            countMeta: L10n.Workspace.placeCount,
            refreshingMeta: L10n.Workspace.placeCountRefreshing,
            mark: .subjectPlace,
            ref: { $0.entity.ref },
            rowAccessibilityIdentifier: { "places.row.\($0.entity.ref)" },
            titleSource: { PlaceTitleDisplay.titleSource($0.titleParts) },
            extraCount: { $0.extraNameCount },
            rows: { PlaceListOrder.sorted($0) },
            accessibilityLabel: Self.rowLabel,
            location: Self.location,
            secondary: Self.chainLine
        )
    }

    private static func rowLabel(_ header: CatalogPlaceHeader) -> String {
        let source = PlaceTitleDisplay.titleSource(header.titleParts)
        return L10n.Workspace.placeRowAccessibility(
            title: source.text,
            extra: header.extraNameCount,
            chain: PlaceChainDisplay.line(parents: header.parents),
            ref: header.entity.ref
        )
    }

    private static func location(_ header: CatalogPlaceHeader) -> WorkspaceLocation {
        .placeDetail(
            entityId: header.entity.id,
            ref: header.entity.ref,
            title: PlaceTitleDisplay.titleSource(header.titleParts).text
        )
    }

    private static func chainLine(_ header: CatalogPlaceHeader) -> PlaceChainLine {
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
private struct PlaceChainLine: View {
    let text: String

    var body: some View {
        Text(verbatim: text)
            .italic()
            .lineLimit(1)
            .frame(maxHeight: text.isEmpty ? 0 : nil)
            .accessibilityHidden(text.isEmpty)
    }
}
