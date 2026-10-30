import Foundation

/// The first Place's display line (leaf plus chain when parents exist), and
/// how many other Places a walk reached. A Place's other names (Montréal and
/// Montreal) are one place, not a disagreement.
enum DerivedPlace {
    /// The Place the row shows: the first whose line is not empty. The label,
    /// the open arrow and the Why's ink all read this one Place.
    static func shown(_ places: [CatalogHeaderPlace]) -> CatalogHeaderPlace? {
        places.first { !PlaceChainDisplay.placeLine($0).isEmpty }
    }

    static func name(_ places: [CatalogHeaderPlace]) -> String {
        shown(places).map(PlaceChainDisplay.placeLine) ?? ""
    }

    /// Places beyond the one `name` shows: the list's +N.
    static func extra(_ places: [CatalogHeaderPlace]) -> Int {
        max(0, places.count - (shown(places) == nil ? 0 : 1))
    }
}
