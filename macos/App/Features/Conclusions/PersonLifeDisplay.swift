import Foundation

/// A Person's birth and death as one secondary line: dates, then places.
/// A missing half is left out. A birth with no death keeps the dash.
enum PersonLifeDisplay {
    struct Line: Equatable {
        var text: String
        /// Places beyond the ones in `text`. The list's +N.
        var extraPlaces: Int
    }

    static func line(_ header: CatalogPersonHeader, locale: Locale = .autoupdatingCurrent) -> Line {
        let dates = dateSpan(birth: header.birth.date, death: header.death.date, locale: locale)
        let places = placeSpan(birth: header.birth.places, death: header.death.places)
        let extra = DerivedPlace.extra(header.birth.places) + DerivedPlace.extra(header.death.places)
        let text: String
        switch (dates.isEmpty, places.isEmpty) {
        case (true, true):
            text = ""
        case (false, true):
            text = dates
        case (true, false):
            text = places
        case (false, false):
            text = "\(dates) · \(places)"
        }
        return Line(text: text, extraPlaces: extra)
    }

    /// The date a life row shows. Empty when none is recorded.
    static func dateText(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String {
        guard let date else { return "" }
        return DateRowDisplay.line(date: date, locale: locale)
    }

    private static func dateSpan(
        birth: CatalogDateValueInput?,
        death: CatalogDateValueInput?,
        locale: Locale
    ) -> String {
        let born = dateText(birth, locale: locale)
        let died = dateText(death, locale: locale)
        switch (born.isEmpty, died.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return "\(born) –"
        case (true, false):
            return died
        case (false, false):
            return "\(born) – \(died)"
        }
    }

    private static func placeSpan(birth: [CatalogHeaderPlace], death: [CatalogHeaderPlace]) -> String {
        let born = DerivedPlace.name(birth)
        let died = DerivedPlace.name(death)
        switch (born.isEmpty, died.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return born
        case (true, false):
            return died
        case (false, false):
            return "\(born) → \(died)"
        }
    }
}
