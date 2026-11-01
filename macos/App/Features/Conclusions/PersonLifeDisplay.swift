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

    static func line(_ summary: CatalogPersonSummary, locale: Locale = .autoupdatingCurrent) -> Line {
        let dates = dateSpan(birth: summary.birth.date, death: summary.death.date, locale: locale)
        let places = placeSpan(birth: summary.birth.places, death: summary.death.places)
        let extra = DerivedPlace.extra(summary.birth.places) + DerivedPlace.extra(summary.death.places)
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

    private static func placeSpan(birth: [CatalogChainedPlace], death: [CatalogChainedPlace]) -> String {
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
