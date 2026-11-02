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

    /// One life line: *b. 14 May 1817 · York*. Date and place come from the
    /// Person summary. Nil when that half is not recorded.
    struct LifeLine: Equatable {
        var abbreviation: String
        var date: String?
        var place: String?
        var accessibilityLabel: String
    }

    /// What a line says of a half that is not recorded. `.stated` is the
    /// header: both lines always, the half "unknown", spoken in full
    /// ("Born 3 May 1790, Bremen"). `.omitted` is a tile: only lines with a
    /// date or a place, the half left out, spoken mid-sentence in lowercase
    /// ("born 3 May 1790, Bremen"; "died 1851").
    enum Missing: Equatable { case stated, omitted }

    /// The *b.* and *d.* lines of a Person.
    static func lifeLines(
        _ summary: CatalogPersonSummary?,
        locale: Locale = .autoupdatingCurrent,
        missing: Missing
    ) -> [LifeLine] {
        let born = lifeLine(
            L10n.Conclusions.personBornAbbr, facts: summary?.birth, locale: locale, missing: missing,
            spoken: missing == .stated ? L10n.Conclusions.a11yBorn : L10n.Conclusions.a11yBornInline
        )
        let died = lifeLine(
            L10n.Conclusions.personDiedAbbr, facts: summary?.death, locale: locale, missing: missing,
            spoken: missing == .stated ? L10n.Conclusions.a11yDied : L10n.Conclusions.a11yDiedInline
        )
        return [born, died].compactMap { $0 }
    }

    private static func lifeLine(
        _ abbreviation: LocalizedStringResource,
        facts: CatalogLifeFacts?,
        locale: Locale,
        missing: Missing,
        spoken: LocalizedStringResource
    ) -> LifeLine? {
        let date = dateText(facts?.date, locale: locale)
        let place = DerivedPlace.name(facts?.places ?? [])
        let word = L10n.string(spoken)
        let label: String
        switch missing {
        case .stated:
            label = L10n.Conclusions.a11yLifeLine(
                word,
                date: date.isEmpty ? L10n.string(L10n.Conclusions.dateUnknown) : date,
                place: place.isEmpty ? L10n.string(L10n.Conclusions.placeUnknown) : place
            )
        case .omitted:
            switch (date.isEmpty, place.isEmpty) {
            case (true, true): return nil
            case (false, false): label = L10n.Conclusions.a11yLifeLine(word, date: date, place: place)
            case (false, true): label = L10n.Conclusions.a11yLifeLineKnown(word, part: date)
            case (true, false): label = L10n.Conclusions.a11yLifeLinePlaceOnly(word, place: place)
            }
        }
        return LifeLine(
            abbreviation: L10n.string(abbreviation),
            date: date.isEmpty ? nil : date,
            place: place.isEmpty ? nil : place,
            accessibilityLabel: label
        )
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
