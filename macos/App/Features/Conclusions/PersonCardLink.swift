import Foundation

/// One person card on a Person page section and where it goes: the card's
/// content and the location a click opens. `location` is nil for a card that
/// opens nothing (the person the page is about). Family's tiles and the
/// Timeline's participants are both this.
struct PersonCardLink: Equatable, Identifiable {
    /// Unique on the page.
    var id: String
    var content: PVPersonCardContent
    var location: WorkspaceLocation?
}

extension PersonCardLink {
    /// A card for another person that opens them on `section` of their page.
    static func link(
        id: String,
        summary: CatalogPersonSummary,
        role: String?,
        opensOn section: ConclusionPageSection,
        locale: Locale,
        orders: NameFormatOrders
    ) -> PersonCardLink {
        PersonCardLink(
            id: id,
            content: PVPersonCardContent(summary: summary, role: role, isFocus: false, locale: locale, orders: orders),
            location: PersonsList.location(summary, orders: orders).withPageSection(section)
        )
    }
}
