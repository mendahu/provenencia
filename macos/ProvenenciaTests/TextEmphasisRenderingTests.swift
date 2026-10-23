import SwiftUI
import Testing
@testable import Provenencia

/// The evidence sheet sets a name by weight: an `AttributedString` run marked strongly emphasized
/// inside a `Text` takes the font of the text around it, heavier. This draws both and compares
/// them, so a font that ignores the emphasis fails here rather than going unseen on the sheet.
/// Emphasis cannot make a semibold display face heavier: in a semibold title a name reads at the
/// weight of the text around it.
@Suite
@MainActor
struct TextEmphasisRenderingTests {
    private func bitmap(_ text: Text, font: Font) -> Data? {
        let renderer = ImageRenderer(content: text.font(font).foregroundStyle(Color.black).padding(4).background(Color.white))
        renderer.scale = 2
        guard let image = renderer.cgImage,
              let data = image.dataProvider?.data as Data? else { return nil }
        return data
    }

    private func emphasized(_ intent: InlinePresentationIntent?) -> AttributedString {
        var text = AttributedString("John Smith")
        if let intent {
            text.inlinePresentationIntent = intent
        }
        return text
    }

    private func drawsHeavier(_ font: Font) throws -> Bool {
        let plain = try #require(bitmap(Text(emphasized(nil)), font: font))
        let strong = try #require(bitmap(Text(emphasized(.stronglyEmphasized)), font: font))
        return plain != strong
    }

    @Test func inTheBodyFont() throws {
        #expect(try drawsHeavier(PVFont.body(size: PVTypeScale.bodySmall)))
    }

    @Test func inTheDisplayFontAtRegularWeight() throws {
        #expect(try drawsHeavier(PVFont.display(size: PVTypeScale.h3, weight: PVFontWeight.regular)))
    }
}
