import Testing
@testable import Provenencia

@Suite
struct PVArtifactChecklistTests {
    struct Case: Sendable, CustomTestStringConvertible {
        let name: String
        let check: @Sendable () -> Bool
        var testDescription: String { name }
    }

    private static let ids = ["a1", "a2", "a3"]

    private static let cases: [Case] = [
        Case(name: "All off clears every id but the locked one") {
            let excluded = PVChecklistState.toggledAll(on: false, ids: ids, locked: ["a2"])
            return excluded == ["a1", "a3"]
                && PVChecklistState.checkedCount(ids: ids, excluded: excluded, locked: ["a2"]) == 1
                && !PVChecklistState.isAllChecked(ids: ids, excluded: excluded, locked: ["a2"])
                && PVChecklistState.isAllIndeterminate(ids: ids, excluded: excluded, locked: ["a2"])
        },
        Case(name: "All on clears the unchecked set") {
            PVChecklistState.toggledAll(on: true, ids: ids, locked: ["a2"]).isEmpty
        },
        Case(name: "Every id unchecked with none locked is not mixed") {
            !PVChecklistState.isAllIndeterminate(ids: ids, excluded: Set(ids), locked: [])
        },
        Case(name: "A stale excluded id does not change the count") {
            PVChecklistState.checkedCount(ids: ids, excluded: ["gone"], locked: []) == 3
        },
        Case(name: "Checked ids keep list order and keep a locked id even if excluded") {
            PVChecklistState.checkedIDs(ids: ids, excluded: ["a1", "a2"], locked: ["a2"]) == ["a2", "a3"]
        },
    ]

    @Test("stateHonorsLockedRows", arguments: cases)
    func stateHonorsLockedRows(_ testCase: Case) {
        #expect(testCase.check())
    }
}
