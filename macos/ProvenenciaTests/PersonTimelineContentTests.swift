import Foundation
import Testing
@testable import Provenencia

struct PersonTimelineContentTests {
    private let en = Locale(identifier: "en_US")
    private typealias Data = PersonTimelinePreviewData

    private func content(
        _ timeline: CatalogPersonTimeline,
        expanded: Set<PersonTimelineContent.Fold> = []
    ) -> PersonTimelineContent {
        PersonTimelineContent(timeline: timeline, expanded: expanded, locale: en)
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
        timeline.birth = CatalogLifeMarker(eventId: "1", besideEventId: "2")
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

        let open = content(Data.lise, expanded: [.dated])
        #expect(open.items.count == 120 + 3 && !ids(open).contains("more.dated"))
        #expect(Array(ids(open).dropLast(2).suffix(1)) == ["l118"])

        let large = content(Data.largeEvent)
        let burial = large.items.compactMap(\.eventRow).first
        #expect(burial?.participants.count == 15 && burial?.participantsHidden == 26)
        let all = content(Data.largeEvent, expanded: [.participants(eventId: "big")])
        #expect(all.items.compactMap(\.eventRow).first?.participants.count == 41)
        #expect(all.items.compactMap(\.eventRow).first?.participantsHidden == 0)
        // Go's order is kept: Child, Witness, Informant, Pallbearer, then Mourner.
        #expect(all.items.compactMap(\.eventRow).first?.participants.prefix(6).map(\.content.role)
            == ["Child", "Child", "Child", "Child", "Witness", "Informant"])

        // The first item a fold reveals, for the focus after "Show N more".
        let datedRows = open.items.compactMap(\.eventRow).prefix(51).map(\.id)
        #expect(lise.firstRevealed(by: .dated) == datedRows.last)
        #expect(open.firstRevealed(by: .dated) == datedRows.last)
        #expect(lise.firstRevealed(by: .undated) == nil)
        #expect(large.firstRevealed(by: .participants(eventId: "big")) == all.items.compactMap(\.eventRow).first?.participants[15].id)
        #expect(all.firstRevealed(by: .participants(eventId: "big")) == large.firstRevealed(by: .participants(eventId: "big")))
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
        #expect(content(timeline, expanded: [.undated]).items.compactMap(\.eventRow).count == 52)
        #expect(undated.firstRevealed(by: .undated) == "u50")
    }

    @Test func pointsAreNamedByTheirRow() {
        for timeline in [Data.johann, Data.lise, Data.unknownEnds] {
            let all = content(timeline, expanded: [.dated, .undated])
            #expect(!all.allSpans.isEmpty)
            for span in all.allSpans {
                #expect(span.spokenLabel == row(all, span.id)?.spokenLabel)
                #expect(all.pointLabels[span.id] == span.spokenLabel)
            }
        }
        // A fold hides a row but its point keeps its name; a placeholder cap is named by its text.
        let lise = content(Data.lise)
        #expect(lise.pointLabels["l60"] == lise.allSpans.first { $0.id == "l60" }?.spokenLabel)
        #expect(content(Data.unknownEnds).pointLabels["birth"] == "Birth unknown")
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
        #expect(row(changed, "2")?.spokenLabel == "Baptism of Johann Weber, 16 Mar 1817, St. Ansgarii, Bremen")
        let bare = row(changed, "13")?.spokenLabel ?? ""
        #expect(!bare.isEmpty && !bare.contains(", as ") && !bare.contains(",,") && !bare.hasSuffix(","))
        #expect(bare == row(changed, "13")?.title.text)

        let birth = row(johann, "1")
        #expect(birth?.spokenLabel == "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject")
        if case .marker(let marker) = johann.items[0] { #expect(marker.spokenLabel == birth?.spokenLabel) }
        #expect(row(johann, "8")?.spokenLabel.hasSuffix(", as Witness, Informant") == true)

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
