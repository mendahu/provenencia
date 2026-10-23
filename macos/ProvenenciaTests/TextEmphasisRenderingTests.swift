import SwiftUI
import Testing
@testable import Provenencia

/// The evidence sheet sets a name by weight: an `AttributedString` run marked strongly emphasized
/// inside a `Text` takes the font of the text around it, heavier. This draws both and compares
/// them, so a font that ignores the emphasis fails here rather than going unseen on the sheet.
/// Emphasis cannot make a semibold display face heavier, so a `PVSectionHeader` given an
/// `AttributedString` title draws its base at regular weight and its emphasized runs carry the
/// weight.
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

    private func headerBitmap(_ title: AttributedString) -> Data? {
        let renderer = ImageRenderer(
            content: PVSectionHeader(title: title)
                .frame(width: 400)
                .padding(4)
                .background(Color.white)
        )
        renderer.scale = 2
        guard let image = renderer.cgImage,
              let data = image.dataProvider?.data as Data? else { return nil }
        return data
    }

    @Test func anEmphasizedRunInASectionHeaderTitleDrawsHeavier() throws {
        var name = AttributedString("John Smith")
        name.inlinePresentationIntent = .stronglyEmphasized
        let strong = name + AttributedString(" in Leeds")
        let plain = try #require(headerBitmap(AttributedString("John Smith in Leeds")))
        let emphasized = try #require(headerBitmap(strong))
        #expect(plain != emphasized)
    }
}
