import Foundation

/// A conclusion header's title, secondary line and ref, whatever its kind: the one switch over
/// person, event and place headers, so search hits, Promote's menus and the evidence sheet's
/// pivots name a record the same way. Not a view.
enum ConclusionHeaderDisplay {
    static func title(
        _ header: CatalogConclusionHeader,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        switch header {
        case .person(let person): PersonHeaderDisplay.title(person, style: style, orders: orders)
        case .event(let event): EventTitleDisplay.title(event, orders: orders)
        case .place(let place): PlaceTitleDisplay.title(place)
        }
    }

    /// The line under the title: a person's life, an event's date and place, a place's chain.
    static func secondary(_ header: CatalogConclusionHeader) -> String {
        switch header {
        case .person(let person): PersonLifeDisplay.line(person).text
        case .event(let event): EventSecondaryDisplay.line(event)
        case .place(let place):
            PlaceChainDisplay.line(parents: place.parents, candidates: place.parentsAreCandidates)
        }
    }

    /// The header's record ref.
    static func ref(_ header: CatalogConclusionHeader) -> String {
        switch header {
        case .person(let person): person.entity.ref
        case .event(let event): event.entity.ref
        case .place(let place): place.entity.ref
        }
    }
}
