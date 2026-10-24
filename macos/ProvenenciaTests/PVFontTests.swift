import AppKit
import Testing
@testable import Provenencia

/// AppKit faces are resolved once per family, size, weight and italic.
@Suite
struct PVFontTests {
    @Test func bodyFontIsCached() {
        PVFontRegistration.registerBundledFontsIfNeeded()
        func cached(_ size: CGFloat) -> Bool {
            PVFont.nsIsCachedForTest(family: PVFontFamily.body, size: size, weight: PVFontWeight.regular, italic: false)
        }
        #expect(!cached(12.25))
        let first = PVFont.nsBody(size: 12.25)
        let second = PVFont.nsBody(size: 12.25)
        #expect(first === second)
        #expect(cached(12.25))
        #expect(!cached(13.75))
    }

    @Test func monoResolvesPlexMono() {
        PVFontRegistration.registerBundledFontsIfNeeded()
        #expect(PVFont.nsMono(size: 9).familyName == PVFontFamily.mono)
    }
}
