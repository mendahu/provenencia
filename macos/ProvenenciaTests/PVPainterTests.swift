import AppKit
import Testing
@testable import Provenencia

/// The drawing toolkit resolves colors for an appearance it is given, builds
/// rounded paths without a SwiftUI host, and measures text with one rule.
@Suite
@MainActor
struct PVPainterTests {
    @Test func colorResolvesPerAppearance() throws {
        let light = PVPainter.color(PVColor.textMuted, in: try #require(NSAppearance(named: .aqua)))
        let dark = PVPainter.color(PVColor.textMuted, in: try #require(NSAppearance(named: .darkAqua)))
        #expect(light != dark)
    }

    @Test func continuousRoundedRectContainsItsCentreAndNotACornerOutside() {
        let rect = CGRect(x: 0, y: 0, width: 100, height: 60)
        let path = PVPainter.continuous(rect, radius: 16)
        #expect(path.contains(CGPoint(x: 50, y: 30)))
        #expect(!path.contains(CGPoint(x: 1, y: 1)))
        #expect(path.contains(CGPoint(x: 20, y: 1)))
    }

    @Test func emptyTextMeasuresOneLine() {
        let font = PVFont.nsBody(size: 12)
        #expect(PVTextMetrics.textHeight("", width: 100, font: font, maxLines: 3) == PVTextMetrics.fontLineHeight(font))
    }

    @Test func maxHeightIsLinesOfTheCeiledLineHeightAndTreatsZeroAsOne() {
        let font = PVFont.nsBody(size: 12)
        let line = ceil(PVTextMetrics.fontLineHeight(font))
        #expect(PVTextMetrics.maxHeight(font: font, lines: 2) == line * 2)
        #expect(PVTextMetrics.maxHeight(font: font, lines: 1) == line)
        #expect(PVTextMetrics.maxHeight(font: font, lines: 0) == line)
    }
}
