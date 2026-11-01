import Foundation

/// A conclusion summary's title, secondary line and ref, whatever its kind: the one switch over
/// person, event and place summaries, so search hits, Promote's menus and the evidence sheet's
/// pivots name a record the same way. Not a view.
enum ConclusionSummaryDisplay {
    static func title(
        _ summary: CatalogConclusionSummary,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        switch summary {
        case .person(let person): PersonTitleDisplay.title(person, style: style, orders: orders)
        case .event(let event): EventTitleDisplay.title(event, orders: orders)
        case .place(let place): PlaceTitleDisplay.title(place)
        }
    }

    /// The line under the title: a person's life, an event's date and place, a place's chain.
    static func secondary(_ summary: CatalogConclusionSummary) -> String {
        switch summary {
        case .person(let person): PersonLifeDisplay.line(person).text
        case .event(let event): EventSecondaryDisplay.line(event)
        case .place(let place):
            PlaceChainDisplay.line(parents: place.parents, candidates: place.parentsAreCandidates)
        }
    }

    /// The summary's record ref.
    static func ref(_ summary: CatalogConclusionSummary) -> String {
        switch summary {
        case .person(let person): person.entity.ref
        case .event(let event): event.entity.ref
        case .place(let place): place.entity.ref
        }
    }
}
