import Testing
@testable import Provenencia

struct ConclusionFoldTests {
    @Test func foldRule() {
        let thirtyOne = Array(1...31)
        let family = ConclusionFold.shown(thirtyOne, limit: 30, minimumHidden: 1, expanded: false)
        #expect(family.shown.count == 30 && family.hidden == 1)

        let seven = Array(1...7)
        let folded = ConclusionFold.shown(seven, limit: 5, minimumHidden: 2, expanded: false)
        #expect(folded.shown == [1, 2, 3, 4, 5] && folded.hidden == 2)

        let six = Array(1...6)
        let whole = ConclusionFold.shown(six, limit: 5, minimumHidden: 2, expanded: false)
        #expect(whole.shown == six && whole.hidden == 0)

        let open = ConclusionFold.shown(seven, limit: 5, minimumHidden: 2, expanded: true)
        #expect(open.shown == seven && open.hidden == 0)

        #expect(ConclusionFold.firstRevealed(seven, limit: 5, minimumHidden: 2) == 6)
        #expect(ConclusionFold.firstRevealed(six, limit: 5, minimumHidden: 2) == nil)
    }
}
