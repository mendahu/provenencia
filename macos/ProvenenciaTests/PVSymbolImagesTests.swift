import AppKit
import Testing
@testable import Provenencia

/// Symbol images are configured once per symbol and point size.
@Suite
@MainActor
struct PVSymbolImagesTests {
    @Test func cachesBySymbolAndPointSize() {
        #expect(!PVSymbolImages.isCachedForTesting(.penLine, pointSize: 12.5))
        let first = PVSymbolImages.image(.penLine, pointSize: 12.5)
        let second = PVSymbolImages.image(.penLine, pointSize: 12.5)
        #expect(first != nil)
        #expect(first === second)
        #expect(PVSymbolImages.isCachedForTesting(.penLine, pointSize: 12.5))
        #expect(!PVSymbolImages.isCachedForTesting(.trash, pointSize: 12.5))
    }

    @Test func customGlyphIsNilAndNotCached() {
        #expect(PVSymbolImages.image(.imageUp, pointSize: 12.5) == nil)
        #expect(!PVSymbolImages.isCachedForTesting(.imageUp, pointSize: 12.5))
    }
}
