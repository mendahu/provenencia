#if DEBUG
import Foundation
import SwiftUI

/// Timelines for the Timeline section's `#Preview`s and tests (board S13-D4,
/// brief § 5): the sample life of Johann Weber as Go returns it, in date order
/// with each row's span, undated last, and the birth and death markers.
enum PersonTimelinePreviewData {
    // MARK: - When

    /// When an event happened, with the span S13-33 gives it: a whole year is
    /// [y, y+1), every year is 365 days.
    enum When {
        case day(Int32, Int32, Int32)
        case month(Int32, Int32)
        case year(Int32)
        case years(Int32, Int32)
        case before(Int32)

        private static let monthStart: [Int32] = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365]

        private static func fraction(month: Int32, day: Int32 = 1) -> Double {
            Double(monthStart[Int(month) - 1] + day - 1) / 365
        }

        var span: CatalogTimelineSpan {
            switch self {
            case .day(let y, let m, let d):
                let from = Double(y) + Self.fraction(month: m, day: d)
                return CatalogTimelineSpan(from: from, to: from + 1.0 / 365, openStart: false, openEnd: false)
            case .month(let y, let m):
                let from = Double(y) + Self.fraction(month: m)
                let to = Double(y) + Self.fraction(month: m + 1)
                return CatalogTimelineSpan(from: from, to: to, openStart: false, openEnd: false)
            case .year(let y):
                return CatalogTimelineSpan(from: Double(y), to: Double(y) + 1, openStart: false, openEnd: false)
            case .years(let a, let b):
                return CatalogTimelineSpan(from: Double(a), to: Double(b) + 1, openStart: false, openEnd: false)
            case .before(let y):
                return CatalogTimelineSpan(from: Double(y), to: Double(y) + 1, openStart: true, openEnd: false)
            }
        }

        func apply(to summary: inout CatalogEventSummary) {
            switch self {
            case .day(let y, let m, let d):
                summary.date = CatalogDateValueInput(kind: "point", startYear: y, startMonth: m, startDay: d)
                summary.dateCount = 1
            case .month(let y, let m):
                summary.date = CatalogDateValueInput(kind: "point", startYear: y, startMonth: m)
                summary.dateCount = 1
            case .year(let y):
                summary.date = CatalogDateValueInput(kind: "point", startYear: y)
                summary.dateCount = 1
            case .years(let a, let b):
                summary.startDate = CatalogDateValueInput(kind: "point", startYear: a)
                summary.endDate = CatalogDateValueInput(kind: "point", startYear: b)
                summary.startDateCount = 1
                summary.endDateCount = 1
            case .before(let y):
                summary.date = CatalogDateValueInput(kind: "point", qualifier: "BEF", startYear: y)
                summary.dateCount = 1
            }
        }
    }

    // MARK: - People

    static func person(_ id: String, _ name: String, born: Int32? = nil) -> CatalogPersonSummary {
        PersonFamilyPreviewData.summary(id, name, born: born)
    }

    static let bremen = CatalogChainedPlace(
        entity: CatalogCanonicalEntity(id: "pl-bremen", ref: "PLC-BREMEN", subjectTypeID: "t", label: ""),
        names: ["Bremen"]
    )

    /// Born 3 May 1790 at Bremen, died 1851.
    static var anna: CatalogPersonSummary {
        var summary = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "anna", ref: "PER-ANNA", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "Anna Weber"),
            nameValueCount: 1
        )
        summary.birth = CatalogLifeFacts(
            date: CatalogDateValueInput(kind: "point", startYear: 1790, startMonth: 5, startDay: 3),
            dateCount: 1,
            places: [bremen]
        )
        summary.death = CatalogLifeFacts(date: CatalogDateValueInput(kind: "point", startYear: 1851), dateCount: 1)
        return summary
    }

    static let heinrich = person("heinrich", "Heinrich Weber")
    static let kohl = person("kohl", "Friedrich Kohl")
    static let margarethe = person("margarethe", "Margarethe Weber", born: 1815)
    static let friedrich = person("friedrich", "Friedrich Weber", born: 1820)
    static let wilhelm = person("wilhelm", "Wilhelm Weber")
    static let maria = person("maria", "Maria Schulz")
    static let karl = person("karl", "Karl Weber")
    static let clara = person("clara", "Clara Lenz")
    static let sophie = person("sophie", "Sophie Brandt")
    static let otto = person("otto", "Otto Weber")

    // MARK: - Events

    /// A product role term ("father" reads "Father").
    static func role(_ key: String, origin: String = CatalogOrigin.provenencia) -> CatalogConclusionValue {
        .term(id: "role-\(key)", key: key, label: key.capitalized, origin: origin)
    }

    static func participant(_ person: CatalogPersonSummary, _ roles: String...) -> CatalogEventParticipant {
        CatalogEventParticipant(person: person, roles: roles.map { role($0) })
    }

    /// One event of a timeline. `subjects` are the names the title reads; the
    /// rule is chosen by their count (one: "Birth of X", two: "Marriage of X and
    /// Y", none and a place: "Census at Bremen", else the bare type).
    static func event(
        _ id: String,
        type: String,
        subjects: [String] = [],
        when: When? = nil,
        place: String? = nil,
        picture: Bool = false,
        roles: [String] = ["subject"],
        participants: [CatalogEventParticipant] = []
    ) -> CatalogPersonEvent {
        var title = CatalogEventTitle(
            rule: .type,
            typeKey: type,
            typeOrigin: CatalogOrigin.provenencia,
            subjects: subjects.map { CatalogNameValue(form: $0) }
        )
        switch (subjects.count, place) {
        case (1, _): title.rule = .subject
        case (2, _): title.rule = .couple
        case (0, let place?): title.rule = .typeAtPlace; title.place = place
        default: break
        }
        var summary = CatalogEventSummary(
            entity: CatalogCanonicalEntity(id: id, ref: "EVT-\(id.uppercased())", subjectTypeID: "t", label: ""),
            eventTypeKey: type,
            eventTypeOrigin: CatalogOrigin.provenencia,
            title: title
        )
        when?.apply(to: &summary)
        if let place {
            summary.places = [CatalogChainedPlace(
                entity: CatalogCanonicalEntity(id: "pl-\(place)", ref: "PLC-\(place)", subjectTypeID: "t", label: ""),
                names: [place]
            )]
        }
        if picture { summary.pictureImageRelPath = "objects/\(id).jpg" }
        return CatalogPersonEvent(
            event: summary,
            roles: roles.map { role($0) },
            participants: participants,
            span: when?.span
        )
    }

    // MARK: - Johann

    /// Brief § 5, events 1 – 13: death 18 Sep 1880, burial 22 Sep 1880.
    static var johannEvents: [CatalogPersonEvent] {
        [
            event("1", type: "birth", subjects: ["Johann Weber"], when: .day(1817, 3, 12), place: "Bremen", picture: true,
                  participants: [participant(heinrich, "father"), participant(anna, "mother")]),
            event("2", type: "baptism", subjects: ["Johann Weber"], when: .day(1817, 3, 16), place: "St. Ansgarii, Bremen", picture: true,
                  participants: [participant(heinrich, "father"), participant(anna, "mother"), participant(kohl, "witness")]),
            event("3", type: "census", when: .year(1835), place: "Bremen",
                  participants: [participant(heinrich, "subject"), participant(anna, "subject"),
                                 participant(margarethe, "subject"), participant(friedrich, "subject"),
                                 participant(wilhelm, "subject")]),
            event("4", type: "marriage", subjects: ["Johann Weber", "Maria Schulz"], when: .day(1844, 6, 4), place: "Bremen", picture: true,
                  participants: [participant(maria, "subject"), participant(friedrich, "witness")]),
            event("5", type: "baptism", subjects: ["Karl Weber"], when: .day(1845, 2, 2), place: "Hamburg", picture: true,
                  roles: ["father"], participants: [participant(karl, "subject"), participant(maria, "mother")]),
            event("6", type: "census", when: .year(1846), place: "Hamburg",
                  participants: [participant(maria, "subject"), participant(karl, "subject")]),
            event("7", type: "marriage", subjects: ["Friedrich Weber", "Clara Lenz"], when: .year(1850), place: "Bremen",
                  roles: ["witness"], participants: [participant(friedrich, "subject"), participant(clara, "subject")]),
            event("8", type: "death", subjects: ["Maria Schulz"], when: .year(1859), place: "Hamburg",
                  roles: ["witness", "informant"], participants: [participant(maria, "subject")]),
            event("9", type: "marriage", subjects: ["Johann Weber", "Sophie Brandt"], when: .year(1861), place: "Hamburg", picture: true,
                  participants: [participant(sophie, "subject")]),
            event("10", type: "death", subjects: ["Johann Weber"], when: .day(1880, 9, 18), place: "Hamburg", picture: true,
                  participants: [participant(otto, "informant")]),
            event("11", type: "burial", subjects: ["Johann Weber"], when: .day(1880, 9, 22), place: "Ohlsdorf, Hamburg"),
            event("12", type: "residence", place: "Altona"),
            event("13", type: "migration", participants: [participant(sophie, "subject")]),
        ]
    }

    static var johann: CatalogPersonTimeline {
        CatalogPersonTimeline(
            events: johannEvents,
            birth: CatalogLifeMarker(eventId: "1"),
            death: CatalogLifeMarker(eventId: "10")
        )
    }

    /// Johann without events 1 and 10: a birth placeholder beside the baptism, a
    /// death placeholder beside the 1861 marriage.
    static var unknownEnds: CatalogPersonTimeline {
        CatalogPersonTimeline(
            events: johannEvents.filter { $0.id != "1" && $0.id != "10" },
            birth: CatalogLifeMarker(besideEventId: "2"),
            death: CatalogLifeMarker(besideEventId: "9")
        )
    }

    /// Born 3 Apr 1845, baptised in December, died 14 Feb 1848.
    static var ernst: CatalogPersonTimeline {
        CatalogPersonTimeline(
            events: [
                event("e1", type: "birth", subjects: ["Ernst Weber"], when: .day(1845, 4, 3), place: "Hamburg"),
                event("e2", type: "baptism", subjects: ["Ernst Weber"], when: .month(1845, 12), place: "Hamburg"),
                event("e3", type: "death", subjects: ["Ernst Weber"], when: .day(1848, 2, 14), place: "Hamburg"),
            ],
            birth: CatalogLifeMarker(eventId: "e1"),
            death: CatalogLifeMarker(eventId: "e3")
        )
    }

    /// One burial with 41 other participants: 4 Child, 1 Witness, 1 Informant, 6
    /// Pallbearer, and 29 with the researcher's role "Mourner", in S13-33's order.
    static var largeEvent: CatalogPersonTimeline {
        var people: [CatalogEventParticipant] = []
        func add(_ count: Int, _ prefix: String, _ roles: CatalogConclusionValue) {
            for k in 1...count {
                people.append(CatalogEventParticipant(person: person("\(prefix)\(k)", "\(prefix.capitalized) \(k) Weber"), roles: [roles]))
            }
        }
        add(4, "child", role("child"))
        add(1, "witness", role("witness"))
        add(1, "informant", role("informant"))
        add(6, "pallbearer", role("pallbearer"))
        add(29, "mourner", role("mourner", origin: CatalogOrigin.user))
        let burial = event("big", type: "burial", subjects: ["Johann Weber"], when: .day(1880, 9, 22), place: "Ohlsdorf, Hamburg",
                           participants: people)
        return CatalogPersonTimeline(events: [burial], birth: CatalogLifeMarker(), death: CatalogLifeMarker(eventId: "big"))
    }

    /// Born 1801, died 1899: 120 dated events, the death the 118th, a residence
    /// range every few years, a burial "bef. 1900" past the death, and two
    /// undated events.
    static var lise: CatalogPersonTimeline {
        var events: [CatalogPersonEvent] = [
            event("l1", type: "birth", subjects: ["Lise Weber"], when: .day(1801, 5, 9), place: "Bremen"),
        ]
        for k in 2...117 {
            let year = Int32(1801 + (k - 1) * 98 / 117)
            if k % 4 == 0 {
                events.append(event("l\(k)", type: "residence", when: .years(year, year + 2), place: "Bremen"))
            } else {
                events.append(event("l\(k)", type: "census", when: .year(year), place: "Hamburg"))
            }
        }
        events.append(event("l118", type: "death", subjects: ["Lise Weber"], when: .day(1899, 11, 2), place: "Hamburg"))
        events.append(event("l119", type: "burial", subjects: ["Lise Weber"], when: .before(1900), place: "Ohlsdorf, Hamburg"))
        events.append(event("l120", type: "census", when: .year(1901), place: "Hamburg"))
        events.append(event("lu1", type: "residence", place: "Altona"))
        events.append(event("lu2", type: "migration"))
        return CatalogPersonTimeline(
            events: events,
            birth: CatalogLifeMarker(eventId: "l1"),
            death: CatalogLifeMarker(eventId: "l118")
        )
    }

    /// Two undated events and no markers.
    static var ida: CatalogPersonTimeline {
        CatalogPersonTimeline(
            events: [
                event("i1", type: "residence", place: "Altona"),
                event("i2", type: "migration", participants: [participant(sophie, "subject")]),
            ],
            birth: CatalogLifeMarker(),
            death: CatalogLifeMarker()
        )
    }

    /// No events.
    static let ludwig = CatalogPersonTimeline()
}
#endif
