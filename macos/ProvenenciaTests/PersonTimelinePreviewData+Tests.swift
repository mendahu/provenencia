import Foundation
@testable import Provenencia

extension PersonTimelinePreviewData {
    /// Ida's timeline with these events and markers, for the Timeline tests.
    static func timeline(_ events: [CatalogPersonEvent], birth: CatalogLifeMarker, death: CatalogLifeMarker) -> CatalogPersonTimeline {
        var timeline = ida
        timeline.events = events
        timeline.birth = birth
        timeline.death = death
        return timeline
    }
}
