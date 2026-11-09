import Testing
@testable import Provenencia

struct ConclusionFoldTests {
    @Test func foldRule() {
        let family = ConclusionFold(limit: 30, minimumHidden: 1)
        let thirtyOne = Array(1...31)
        let folded31 = family.shown(thirtyOne, pages: 0)
        #expect(folded31.shown.count == 30 && folded31.hidden == 1)

        let lines = ConclusionFold(limit: 5, minimumHidden: 2)
        let seven = Array(1...7)
        let folded = lines.shown(seven, pages: 0)
        #expect(folded.shown == [1, 2, 3, 4, 5] && folded.hidden == 2)

        let six = Array(1...6)
        let whole = lines.shown(six, pages: 0)
        #expect(whole.shown == six && whole.hidden == 0)

        let open = lines.shown(seven, pages: 1)
        #expect(open.shown == seven && open.hidden == 0)

        #expect(lines.firstRevealed(seven, pages: 0) == 6)
        #expect(lines.firstRevealed(six, pages: 0) == nil)
        #expect(lines.firstRevealed(seven, pages: 1) == nil)
    }

    @Test func pagedFold() {
        let fold = ConclusionFold(limit: 5, minimumHidden: 2)
        let list = Array(1...120)
        let shown = (0...3).map { fold.shown(list, pages: $0) }
        #expect(shown.map(\.shown.count) == [5, 55, 105, 120])
        #expect(fold.nextPageCount(hidden: shown[0].hidden) == 50)
        #expect(fold.nextPageCount(hidden: shown[1].hidden) == 50)
        #expect(fold.nextPageCount(hidden: shown[2].hidden) == 15)
        #expect(shown[3].hidden == 0)

        let fiftySix = Array(1...56)
        #expect(fold.shown(fiftySix, pages: 0).hidden == 51)
        #expect(fold.nextPageCount(hidden: 51) == 51)
        #expect(fold.shown(fiftySix, pages: 1).hidden == 0)
        #expect(fold.shown(Array(1...7), pages: 0).hidden == 2)
        #expect(fold.nextPageCount(hidden: 2) == 2)
        #expect(fold.shown(Array(1...6), pages: 0).hidden == 0)
        #expect(fold.shown(Array(1...20), pages: 9).shown.count == 20)
        // 57 lines: a page of 50, then the last two.
        #expect(fold.nextPageCount(hidden: 52) == 50)
        #expect(fold.nextPageCount(hidden: 2) == 2)

        let family = ConclusionFold(limit: 30, minimumHidden: 1)
        let kids = Array(1...100)
        #expect([0, 1, 2].map { family.shown(kids, pages: $0).shown.count } == [30, 80, 100])
        #expect(family.nextPageCount(hidden: 20) == 20)
    }

    @Test func foldStateAndPagesToShow() {
        var state = ConclusionFoldState<String>()
        #expect(state.pages("a") == 0)
        state.revealPage("a")
        state.revealPage("a")
        #expect(state.pages("a") == 2)
        state.reveal("a", pages: 1)
        #expect(state.pages("a") == 2)
        state.reveal("a", pages: 3)
        #expect(state.pages("a") == 3)
        #expect(state.pages("unknown") == 0)

        let events = ConclusionFold(limit: 50, minimumHidden: 1)
        #expect(events.pages(toShow: 49, count: 2000) == 0)
        #expect(events.pages(toShow: 50, count: 2000) == 1)
        #expect(events.pages(toShow: 199, count: 2000) == 3)
        let lines = ConclusionFold(limit: 5, minimumHidden: 2)
        #expect(lines.pages(toShow: 5, count: 6) == 0)
    }
}
