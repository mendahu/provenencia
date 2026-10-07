import Foundation

/// The first Place's display line (leaf plus chain when parents exist), and
/// how many other Places a walk reached. A Place's other names (Montréal and
/// Montreal) are one place, not a disagreement.
enum DerivedPlace {
    static func name(_ places: [CatalogHeaderPlace]) -> String {
        for place in places {
            let line = PlaceChainDisplay.placeLine(place)
            if !line.isEmpty { return line }
        }
        return ""
    }

    /// Places beyond the one `name` shows: the list's +N.
    static func extra(_ places: [CatalogHeaderPlace]) -> Int {
        max(0, places.count - (name(places).isEmpty ? 0 : 1))
    }
}
