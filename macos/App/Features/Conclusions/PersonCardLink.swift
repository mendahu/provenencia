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
