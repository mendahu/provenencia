import Foundation

/// A Person's birth and death as one secondary line: dates, then places.
/// A birth with no death keeps the dash after it; a death with no birth gets
/// the dash (or the arrow) before it, so it never reads as a birth.
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
            text = L10n.Conclusions.lineJoin(first: dates, rest: places)
        }
        return Line(text: text, extraPlaces: extra)
    }

    /// The date a life row shows. Empty when none is recorded.
    static func dateText(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String {
        guard let date else { return "" }
        return DateRowDisplay.line(date: date, locale: locale)
    }

    /// A life's first and last dated events as one span: "14 May 1817 – 2 Jun 1890", and
    /// one date, once, when the two are the same event ("1843", not "1843 –").
    static func span(
        first: CatalogDateValueInput?,
        last: CatalogDateValueInput?,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        if first == last { return dateText(first, locale: locale) }
        return dateSpan(birth: first, death: last, locale: locale)
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
            return L10n.Conclusions.lifeSpanOpen(born: born)
        case (true, false):
            return L10n.Conclusions.lifeSpanDeathOnly(died: died)
        case (false, false):
            return L10n.Conclusions.lifeSpan(born: born, died: died)
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
            return L10n.Conclusions.lifePlaceDeathOnly(died: died)
        case (false, false):
            return L10n.Conclusions.lifePlaces(born: born, died: died)
        }
    }
}
