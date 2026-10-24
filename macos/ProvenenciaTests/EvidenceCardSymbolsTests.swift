import AppKit
import Testing
@testable import Provenencia

/// Card symbols are configured once per symbol and point size.
@Suite
@MainActor
struct EvidenceCardSymbolsTests {
    @Test func cachesBySymbolAndPointSize() {
        #expect(!EvidenceCardSymbols.isCachedForTest(.penLine, pointSize: 12.5))
        let first = EvidenceCardSymbols.image(.penLine, pointSize: 12.5)
        let second = EvidenceCardSymbols.image(.penLine, pointSize: 12.5)
        #expect(first != nil)
        #expect(first === second)
        #expect(EvidenceCardSymbols.isCachedForTest(.penLine, pointSize: 12.5))
        #expect(!EvidenceCardSymbols.isCachedForTest(.trash, pointSize: 12.5))
    }

    @Test func customGlyphIsNilAndNotCached() {
        #expect(EvidenceCardSymbols.image(.imageUp, pointSize: 12.5) == nil)
        #expect(!EvidenceCardSymbols.isCachedForTest(.imageUp, pointSize: 12.5))
    }
}
