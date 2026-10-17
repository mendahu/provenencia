import Foundation

/// The draft of one depiction value: the region drawn on the Artifact and the
/// saved value it came from. Same shape as `DateValueDraft` / `NameValueDraft`.
/// The region rides in a `CitationLocatorDraft`, which owns validation, the PDF
/// page stamp, and the JSON round trip.
struct DepictionValueDraft: Equatable {
    var locator: CitationLocatorDraft
    /// `depiction_values.id` when the value is saved; empty for a new one.
    var valueDepictionID: String
    /// `objects/…` path of the saved image; empty for a new one.
    var imageRelPath: String

    static func empty() -> DepictionValueDraft {
        DepictionValueDraft(locator: .artifactOnly(), valueDepictionID: "", imageRelPath: "")
    }

    init(locator: CitationLocatorDraft, valueDepictionID: String, imageRelPath: String) {
        self.locator = locator
        self.valueDepictionID = valueDepictionID
        self.imageRelPath = imageRelPath
    }

    init(from observation: CatalogObservation) {
        self.locator = observation.depiction.map { CitationLocatorDraft.decode($0.regionJSON) }
            ?? .artifactOnly()
        self.valueDepictionID = observation.valueDepictionID
        self.imageRelPath = observation.depiction?.imageRelPath ?? ""
    }

    var hasRegion: Bool { locator.region?.isValid == true }

    /// A saved value whose region Go hid, until a region is drawn again.
    var regionHidden: Bool { !valueDepictionID.isEmpty && !hasRegion }
}
