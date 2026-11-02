import Testing
@testable import Provenencia

struct MemoTests {
    @Test func buildsOncePerKey() {
        var memo = Memo<Int, String>()
        var builds = 0
        func build() -> String {
            builds += 1
            return "a"
        }

        #expect(memo.value(for: 1, build: build) == "a")
        #expect(memo.value(for: 1, build: build) == "a")
        #expect(builds == 1)
        #expect(memo.value(for: 2, build: build) == "a")
        #expect(builds == 2)

        #expect(memo.value(for: 9) == nil)
        memo.store("b", for: 3)
        #expect(memo.value(for: 3) == "b")
        #expect(memo.value(for: 2) == nil)
    }
}
