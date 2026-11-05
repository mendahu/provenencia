import Testing
@testable import Provenencia

@Suite
@MainActor
struct PVFieldFocusTests {
    private func handle(
        next: (@MainActor () -> Void)?,
        previous: (@MainActor () -> Void)?
    ) -> PVFieldFocus {
        PVFieldFocus(
            isActive: { false },
            focus: {},
            resign: {},
            apply: { $0 },
            next: next,
            previous: previous
        )
    }

    @Test func tabClaimPicksTheNeighbour() {
        var nextCalls = 0
        var previousCalls = 0
        let both = handle(next: { nextCalls += 1 }, previous: { previousCalls += 1 })

        both.tabClaim(shift: false)?()
        #expect(nextCalls == 1)
        #expect(previousCalls == 0)
        both.tabClaim(shift: true)?()
        #expect(nextCalls == 1)
        #expect(previousCalls == 1)

        let onlyNext = handle(next: { nextCalls += 1 }, previous: nil)
        #expect(onlyNext.tabClaim(shift: false) != nil)
        #expect(onlyNext.tabClaim(shift: true) == nil)

        let none = handle(next: nil, previous: nil)
        #expect(none.tabClaim(shift: false) == nil)
        #expect(none.tabClaim(shift: true) == nil)
    }

    @Test func tabPairOrder() {
        #expect(PVTabPairField.first.next == .second)
        #expect(PVTabPairField.first.previous == nil)
        #expect(PVTabPairField.second.previous == .first)
        #expect(PVTabPairField.second.next == nil)

        var claimed: [PVTabPairField] = []
        let neighbours = PVFieldFocus.neighbours(
            next: PVTabPairField.second,
            previous: nil
        ) { claimed.append($0) }
        #expect(neighbours.previous == nil)
        #expect(neighbours.next != nil)
        neighbours.next?()
        #expect(claimed == [.second])
    }
}
