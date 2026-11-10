import Foundation
import Testing
@testable import Provenencia

struct PersonTimelineContentTests {
    private let en = Locale(identifier: "en_US")
    private typealias Data = PersonTimelinePreviewData

    private func content(
        _ timeline: CatalogPersonTimeline,
        pages: Int = 0,
        of unfolded: [PersonTimelineContent.Fold] = [],
        widthClass: PVWidthClass = .regular
    ) -> PersonTimelineContent {
        var folds = ConclusionFoldState<PersonTimelineContent.Fold>()
        // Every list in `unfolded` is open through `pages` pages (a hundred shows any list here).
        for fold in unfolded { folds.reveal(fold, pages: max(pages, 100)) }
        return PersonTimelineContent(timeline: timeline, folds: folds, locale: en, widthClass: widthClass)
    }

    private func paged(
        _ timeline: CatalogPersonTimeline,
        _ revealed: [PersonTimelineContent.Fold: Int],
        widthClass: PVWidthClass = .regular
    ) -> PersonTimelineContent {
        var folds = ConclusionFoldState<PersonTimelineContent.Fold>()
        for (fold, pages) in revealed { folds.reveal(fold, pages: pages) }
        return PersonTimelineContent(timeline: timeline, folds: folds, locale: en, widthClass: widthClass)
    }

    private func ids(_ content: PersonTimelineContent) -> [String] { content.items.map(\.id) }

    private func row(_ content: PersonTimelineContent, _ id: String) -> PersonTimelineContent.Row? {
        content.items.compactMap(\.eventRow).first { $0.id == id }
    }

    @Test func itemsInOrder() {
        let johann = content(Data.johann)
        #expect(ids(johann) == ["1", "2", "3", "4", "5", "6", "7", "8", "9", "undated", "12", "13", "10", "11"])
        guard case .marker(let birth) = johann.items[0], case .marker(let death) = johann.items[12] else {
            Issue.record("birth and death are markers in place")
            return
        }
        #expect(birth.kind == .birth && birth.row?.id == "1" && death.kind == .death && death.row?.id == "10")

        let marriage = row(johann, "4")
        #expect(marriage?.title.text == "Marriage of Johann Weber and Maria Schulz")
        #expect(marriage?.line.date == "4 Jun 1844" && marriage?.line.place == "Bremen")
        #expect(marriage?.roles == [PersonTimelineContent.RoleBadge(text: "Subject", isSubject: true)])
        #expect(marriage?.participants.map(\.content.title.text) == ["Maria Schulz", "Friedrich Weber"])
        #expect(marriage?.participants.map(\.content.role) == ["Subject", "Witness"])
        #expect(row(johann, "1")?.imageRelPath.isEmpty == false)
        #expect(row(johann, "11")?.imageRelPath.isEmpty == true)
        #expect(row(johann, "13")?.line == EventSecondaryDisplay.Content(date: "", place: "", extraPlaces: 0))

        let unknown = content(Data.unknownEnds)
        #expect(ids(unknown) == ["birth", "2", "3", "4", "5", "6", "7", "8", "9", "undated", "12", "13", "death", "11"])
        #expect(unknown.items.first.map { if case .marker(let m) = $0 { m.spokenLabel } else { "" } } == "Birth unknown")

        let ida = content(Data.ida)
        #expect(ids(ida) == ["birth", "undated", "i1", "i2", "death"])

        // A recorded birth with no span, beside event 2: that row first, as the
        // birth marker, not under Undated.
        var timeline = Data.johann
        timeline.events = timeline.events.map { event in
            var event = event
            if event.id == "1" { event.span = nil }
            return event
        }
        timeline.birth = CatalogLifeMarker(eventId: "1", place: .before(eventId: "2"))
        let recorded = content(timeline)
        #expect(ids(recorded) == ["1", "2", "3", "4", "5", "6", "7", "8", "9", "undated", "12", "13", "10", "11"])
        guard case .marker = recorded.items[0] else {
            Issue.record("a recorded birth with no span is a marker")
            return
        }
        #expect(recorded.allSpans.map(\.id) == ["2", "3", "4", "5", "6", "7", "8", "9", "10", "11"])
    }

    @Test func foldsAndOrder() {
        let lise = content(Data.lise)
        let rows = lise.items.compactMap(\.eventRow).map(\.id)
        #expect(rows.prefix(50).allSatisfy { $0.hasPrefix("l") })
        #expect(lise.items.count == 50 + 1 + 3 + 1 + 2)
        #expect(lise.items[50] == .more(.dated, count: 67))
        #expect(Array(ids(lise).suffix(6)) == ["undated", "lu1", "lu2", "l118", "l119", "l120"])
        #expect(lise.allSpans.count == 120)

        let open = content(Data.lise, of: [.dated])
        #expect(open.items.count == 120 + 3 && !ids(open).contains("more.dated"))
        #expect(Array(ids(open).dropLast(2).suffix(1)) == ["l118"])

        let large = content(Data.largeEvent)
        let burial = large.items.compactMap(\.eventRow).first
        #expect(burial?.participants.count == 5 && burial?.participantsHidden == 36)
        let wide = content(Data.largeEvent, widthClass: .wide)
        #expect(wide.items.compactMap(\.eventRow).first?.participants.count == 12)
        #expect(wide.items.compactMap(\.eventRow).first?.participantsHidden == 29)
        #expect(PersonTimelineContent.participantFold(.regular).limit == 5 && PersonTimelineContent.participantFold(.wide).limit == 12)
        let all = content(Data.largeEvent, of: [.participants(eventId: "big")])
        #expect(all.items.compactMap(\.eventRow).first?.participants.count == 41)
        #expect(all.items.compactMap(\.eventRow).first?.participantsHidden == 0)
        // Go's order is kept: Child, Witness, Informant, Pallbearer, then Mourner.
        #expect(all.items.compactMap(\.eventRow).first?.participants.prefix(6).map(\.content.role)
            == ["Child", "Child", "Child", "Child", "Witness", "Informant"])

        // The first item a fold reveals, for the focus after "Show N more".
        let datedRows = open.items.compactMap(\.eventRow).prefix(51).map(\.id)
        #expect(lise.firstRevealed(by: .dated) == datedRows.last)
        // Fully unfolded, a fold hides nothing to reveal.
        #expect(open.firstRevealed(by: .dated) == nil)
        #expect(lise.firstRevealed(by: .undated) == nil)
        #expect(large.firstRevealed(by: .participants(eventId: "big")) == all.items.compactMap(\.eventRow).first?.participants[5].id)
        #expect(wide.firstRevealed(by: .participants(eventId: "big")) == all.items.compactMap(\.eventRow).first?.participants[12].id)
        #expect(all.firstRevealed(by: .participants(eventId: "big")) == nil)
        #expect(large.firstRevealed(by: .participants(eventId: "other")) == nil)

        // One person on two events: distinct ids.
        let johann = content(Data.johann)
        let sophie = johann.items.compactMap(\.eventRow).flatMap(\.participants).filter { $0.content.title.text == "Sophie Brandt" }
        #expect(sophie.count == 2 && Set(sophie.map(\.id)).count == 2)

        // The undated rows fold at 50.
        var timeline = Data.ida
        timeline.events = (0..<52).map { Data.event("u\($0)", type: "residence") }
        let undated = content(timeline)
        #expect(undated.items.contains(.more(.undated, count: 2)))
        #expect(content(timeline, of: [.undated]).items.compactMap(\.eventRow).count == 52)
        #expect(undated.firstRevealed(by: .undated) == "u50")
    }

    @Test func markerInsideHiddenRunStaysShown() {
        var timeline = Data.ida
        timeline.events = (1...60).map { Data.event("d\($0)", type: "census", when: .year(Int32(1800 + $0))) }
        // A birth with no recorded event, placed beside the 55th row: past the 50th.
        timeline.birth = CatalogLifeMarker(place: .before(eventId: "d55"))
        timeline.death = CatalogLifeMarker()

        let folded = content(timeline)
        let rows = folded.items.compactMap(\.eventRow)
        #expect(rows.count == 50 && rows.first?.id == "d1" && rows.last?.id == "d50")
        #expect(folded.items.contains(.more(.dated, count: 10)))

        // The marker stays after "Show N more", where it was before the fold moved to the shared rule.
        let foldedOrder = folded.items.map(\.id)
        #expect(foldedOrder.contains("birth"))
        #expect(foldedOrder.firstIndex(of: "more.dated").map { $0 + 1 } == foldedOrder.firstIndex(of: "birth"))

        let open = content(timeline, of: [.dated])
        let order = open.items.map(\.id)
        let birthAt = order.firstIndex(of: "birth")
        #expect(birthAt != nil && birthAt == order.firstIndex(of: "d55").map { $0 - 1 })

        #expect(folded.firstRevealed(by: .dated) == "d51")
    }

    private func datedEvents(_ count: Int) -> [CatalogPersonEvent] {
        (1...count).map { Data.event("d\($0)", type: "census", when: .year(Int32(1800 + $0 / 20))) }
    }

    @Test func twoThousandEventsUnfoldInPages() {
        var timeline = Data.ida
        timeline.events = datedEvents(2000)
        timeline.birth = CatalogLifeMarker()
        timeline.death = CatalogLifeMarker()
        for pages in 0..<39 {
            let open = paged(timeline, [.dated: pages])
            #expect(open.items.compactMap(\.eventRow).count == 50 + 50 * pages)
            #expect(open.items.contains(.more(.dated, count: 1950 - 50 * pages)))
            #expect(open.firstRevealed(by: .dated) == "d\(50 * (pages + 1) + 1)")
        }
        let all = paged(timeline, [.dated: 39])
        #expect(all.items.compactMap(\.eventRow).count == 2000)
        #expect(!all.items.map(\.id).contains("more.dated"))
        #expect(all.firstRevealed(by: .dated) == nil)
    }

    @Test func foldedMarkersStayAfterPagedRun() {
        var timeline = Data.ida
        timeline.events = datedEvents(120)
        timeline.birth = CatalogLifeMarker(place: .before(eventId: "d110"))
        timeline.death = CatalogLifeMarker()
        let order = paged(timeline, [.dated: 1]).items.map(\.id)
        #expect(order.firstIndex(of: "d100").map { $0 + 1 } == order.firstIndex(of: "more.dated"))
        #expect(order.firstIndex(of: "more.dated").map { $0 + 1 } == order.firstIndex(of: "birth"))
        #expect(order.last == "death")
    }

    @Test func participantsPageAtBothWidths() {
        var timeline = Data.largeEvent
        timeline.events[0].participants = (1...500).map {
            CatalogEventParticipant(person: Data.person("p\($0)", "P \($0) Weber"), roles: [Data.role("mourner")])
        }
        let fold = PersonTimelineContent.Fold.participants(eventId: "big")
        let regular = paged(timeline, [fold: 1])
        let wide = paged(timeline, [fold: 1], widthClass: .wide)
        #expect(regular.items.compactMap(\.eventRow).first?.participants.count == 55)
        #expect(wide.items.compactMap(\.eventRow).first?.participants.count == 62)
        #expect(regular.firstRevealed(by: fold) == "big-p56")
        #expect(wide.firstRevealed(by: fold) == "big-p63")
        #expect(paged(timeline, [:]).firstRevealed(by: fold) == "big-p6")
    }

    @Test func aCardNamesItsRowAndNothingElseDoes() {
        var timeline = Data.largeEvent
        timeline.events[0].participants = (1...20).map {
            CatalogEventParticipant(person: Data.person("p\($0)", "P \($0) Weber"), roles: [Data.role("mourner")])
        }
        let shown = content(timeline)
        #expect(shown.row(holdingCard: "big-p1") == "big")
        #expect(shown.row(holdingCard: "big") == nil)
        #expect(shown.row(holdingCard: "big-p20") == nil)
        #expect(shown.row(holdingCard: "unknown") == nil)
    }

    @Test func pointOnFoldedPageOpensThroughIt() {
        var timeline = Data.ida
        timeline.events = datedEvents(2000) + [Data.event("u1", type: "residence")]
        timeline.birth = CatalogLifeMarker()
        timeline.death = CatalogLifeMarker()
        let folded = content(timeline)
        #expect(folded.pages(toShow: "d180") == 3)
        #expect(folded.pages(toShow: "d1") == 0)
        let open = paged(timeline, [.dated: 3])
        #expect(open.isShown("d180") && !open.isShown("d201"))
        #expect(folded.pages(toShow: "u1") == nil && folded.pages(toShow: "death") == nil)
        #expect(folded.pages(toShow: "unknown") == nil)
    }

    @Test func pointsAreNamedByTheirRow() {
        for timeline in [Data.johann, Data.lise, Data.unknownEnds] {
            let all = content(timeline, of: [.dated, .undated])
            #expect(!all.allSpans.isEmpty)
            let line = PersonTimelineLifeline(content: all, pointsPerYear: 40)
            for item in all.items {
                if let row = item.eventRow { #expect(item.spokenLabel == row.spokenLabel) }
                // A lifeline's target is named by its item.
                if let mark = line.mark(item.id) {
                    #expect(line.targets.first { $0.mark.id == mark.id }?.label == item.spokenLabel)
                }
            }
            #expect(all.items.allSatisfy { item in
                switch item {
                case .undatedHeader, .more: item.spokenLabel == nil
                case .row, .marker: item.spokenLabel != nil
                }
            })
        }
        // A fold hides a row and its point has no name or target of its own; a placeholder cap is named by its text.
        let lise = content(Data.lise)
        let line = PersonTimelineLifeline(content: lise, pointsPerYear: 41)
        #expect(line.mark("l60") != nil && !line.targets.contains { $0.mark.id == "l60" })
        let unknown = content(Data.unknownEnds)
        #expect(unknown.items.first?.spokenLabel == "Birth unknown")
        #expect(PersonTimelineLifeline(content: unknown, pointsPerYear: 43).targets.first { $0.mark.id == "birth" }?.label == "Birth unknown")
    }

    @Test func buildsOnlyShownRowsAndCards() {
        let crowd = (1...500).map {
            CatalogEventParticipant(person: Data.person("p\($0)", "P \($0) Weber"), roles: [Data.role("mourner")])
        }
        var timeline = Data.ida
        timeline.events = (1...2000).map {
            Data.event("d\($0)", type: "census", when: .year(Int32(1800 + $0 / 20)), participants: crowd)
        }
        timeline.birth = CatalogLifeMarker()
        timeline.death = CatalogLifeMarker()
        let folded = paged(timeline, [:], widthClass: .wide)
        let rows = folded.items.compactMap(\.eventRow)
        #expect(rows.count == 50 && rows.allSatisfy { $0.participants.count == 12 })
        #expect(folded.allSpans.count == 2000)
        #expect(folded.items.compactMap(\.spokenLabel).count == 52)
        #expect(!folded.isShown("d51") && !folded.isShown("d2000"))
        #expect(paged(timeline, [.dated: 1], widthClass: .wide).items.compactMap(\.eventRow).count == 100)
    }

    @Test func deathNeverFollowsAnAfterDeathEvent() {
        // Only a burial dated, two undated events: the audit's case.
        let burial = Data.timeline(
            [Data.event("b1", type: "burial", when: .year(1850)), Data.event("i1", type: "residence"), Data.event("i2", type: "migration")],
            birth: CatalogLifeMarker(place: .before(eventId: "b1")), death: CatalogLifeMarker(place: .before(eventId: "b1"))
        )
        #expect(ids(content(burial)) == ["birth", "undated", "i1", "i2", "death", "b1"])

        // A baptism and a burial: the death sits between them.
        let baptism = Data.timeline(
            [Data.event("bap", type: "baptism", when: .year(1820)), Data.event("bur", type: "burial", when: .year(1850))],
            birth: CatalogLifeMarker(place: .before(eventId: "bap")), death: CatalogLifeMarker(place: .before(eventId: "bur"))
        )
        #expect(ids(content(baptism)) == ["birth", "bap", "death", "bur"])

        // A recorded death with no date, before a dated funeral: its row is the marker.
        let recorded = Data.timeline(
            [Data.event("c", type: "census", when: .year(1851)), Data.event("f", type: "funeral", when: .year(1880)), Data.event("d", type: "death")],
            birth: CatalogLifeMarker(place: .before(eventId: "c")), death: CatalogLifeMarker(eventId: "d", place: .before(eventId: "f"))
        )
        let recordedContent = content(recorded)
        #expect(ids(recordedContent) == ["birth", "c", "d", "f"])
        if case .marker(let marker) = recordedContent.items[2] {
            #expect(marker.kind == .death && marker.row?.id == "d" && marker.place == .before(eventId: "f"))
        } else { Issue.record("a recorded death with no date is a marker") }

        // Both after the same row: the birth, then the death.
        let both = Data.timeline(
            [Data.event("x", type: "census", when: .year(1850))],
            birth: CatalogLifeMarker(place: .after(eventId: "x")), death: CatalogLifeMarker(place: .after(eventId: "x"))
        )
        #expect(ids(content(both)) == ["x", "birth", "death"])

        // A place that is not a dated row counts as none.
        let unknown = Data.timeline(
            [Data.event("x", type: "census", when: .year(1850))],
            birth: CatalogLifeMarker(place: .before(eventId: "nope")), death: CatalogLifeMarker(place: .before(eventId: "nope"))
        )
        let unknownContent = content(unknown)
        #expect(ids(unknownContent) == ["birth", "x", "death"])
        if case .marker(let death) = unknownContent.items[2] { #expect(death.place == nil) } else { Issue.record("death marker") }
    }

    @Test func rolesLocationsAndSpoken() {
        let johann = content(Data.johann)
        #expect(row(johann, "8")?.roles.map(\.text) == ["Witness", "Informant"])
        #expect(row(johann, "8")?.roles.allSatisfy { !$0.isSubject } == true)
        #expect(row(johann, "1")?.roles == [PersonTimelineContent.RoleBadge(text: "Subject", isSubject: true)])

        // A researcher's term keyed "subject" is neutral, and all roles empty
        // draw no Badge and no "as" clause.
        var timeline = Data.johann
        timeline.events = timeline.events.map { event in
            var event = event
            if event.id == "1" { event.roles = [Data.role("subject", origin: CatalogOrigin.user)] }
            if event.id == "2" { event.roles = [] }
            if event.id == "13" { event.roles = [.none, .text("")] }
            return event
        }
        let changed = content(timeline)
        #expect(row(changed, "1")?.roles == [PersonTimelineContent.RoleBadge(text: "Subject", isSubject: false)])
        #expect(row(changed, "2")?.roles.isEmpty == true)
        #expect(row(changed, "2")?.spokenLabel == "Baptism of Johann Weber, 16 Mar 1817, St. Ansgarii, Bremen, age 0")
        let bare = row(changed, "13")?.spokenLabel ?? ""
        #expect(!bare.isEmpty && !bare.contains(", as ") && !bare.contains(",,") && !bare.hasSuffix(","))
        #expect(bare == row(changed, "13")?.title.text)

        let birth = row(johann, "1")
        #expect(birth?.spokenLabel == "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject, age 0")
        if case .marker(let marker) = johann.items[0] { #expect(marker.spokenLabel == birth?.spokenLabel) }
        #expect(row(johann, "8")?.spokenLabel.hasSuffix(", as Witness, Informant, age 41 to 42") == true)

        // The age flag: a number, a range, or "?" with no age, and its tooltip and spoken tail.
        #expect(row(johann, "1")?.ageText == "0" && row(johann, "4")?.ageText == "27")
        #expect(row(johann, "3")?.ageText == "17–18" && row(johann, "12")?.ageText == "?")
        #expect(row(johann, "4")?.ageHelp == "Age 27" && row(johann, "12")?.ageHelp == "Age unknown")
        // Only one side known: before or after, the year included.
        let atMost = PersonTimelineContent.ageWords(CatalogAgeRange(lo: 0, hi: 63, openLo: true))
        #expect(atMost.flag == "≤ 63" && atMost.help == "Age at most 63" && atMost.spoken == "age at most 63")
        let atLeast = PersonTimelineContent.ageWords(CatalogAgeRange(lo: 32, hi: 0, openHi: true))
        #expect(atLeast.flag == "≥ 32" && atLeast.help == "Age at least 32" && atLeast.spoken == "age at least 32")
        #expect(PersonTimelineContent.ageWords(nil).spoken == nil)
        #expect(row(johann, "4")?.spokenLabel.hasSuffix(", as Subject, age 27") == true)
        #expect(row(johann, "3")?.spokenLabel.hasSuffix("age 17 to 18") == true)
        #expect(row(johann, "12")?.spokenLabel.contains("age") == false)

        // The product key of a value.
        #expect(Data.role("subject").productKey == "subject")
        #expect(Data.role("subject", origin: CatalogOrigin.user).productKey == nil)
        #expect(CatalogConclusionValue.text("subject").productKey == nil)
        #expect(CatalogConclusionValue.none.productKey == nil)

        // Cards: role caption, spoken label, location.
        let anna = row(johann, "1")?.participants.last
        #expect(anna?.content.role == "Mother")
        #expect(anna?.content.spokenLabel == "Mother, Anna Weber, born 3 May 1790, Bremen, died 1851")
        #expect(anna?.location == PersonsList.location(Data.anna).withPageSection(.timeline))
        #expect(anna?.id == "1-anna")
        let event = Data.johannEvents[3].event
        #expect(row(johann, "4")?.location == EventsList.location(event))

        // The spoken line has no ref; the Events list's label is unchanged.
        let spoken = EventsList.spokenLine(event, orders: NameFormatOrders(), locale: en)
        #expect(spoken == "Marriage of Johann Weber and Maria Schulz, 4 Jun 1844, Bremen")
        #expect(!spoken.contains(event.entity.ref))
        let label = EventsList.accessibilityLabel(event, orders: NameFormatOrders(), locale: en)
        #expect(label.hasPrefix(L10n.Workspace.eventRowAccessibility(title: "Marriage of Johann Weber and Maria Schulz", date: "4 Jun 1844", ref: event.entity.ref)))
        #expect(label.hasSuffix(", Bremen"))
    }
}
