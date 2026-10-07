import Foundation

/// The first named Place's first kept name, and how many other Places a walk
/// reached. A Place's other names (Montréal and Montreal) are one place, not
/// a disagreement. A chain is not part of the name.
enum DerivedPlace {
    static func name(_ places: [CatalogHeaderPlace]) -> String {
        for place in places {
            for name in place.names {
                let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
                if !trimmed.isEmpty { return trimmed }
            }
        }
        return ""
    }

    /// Places beyond the one `name` shows: the list's +N.
    static func extra(_ places: [CatalogHeaderPlace]) -> Int {
        max(0, places.count - (name(places).isEmpty ? 0 : 1))
    }
}
