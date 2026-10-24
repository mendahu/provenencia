import AppKit
import Testing
@testable import Provenencia

@Suite
@MainActor
struct GraphCanvasItemViewTests {
    private final class RedSquareView: GraphCanvasItemView {
        var paints = 0

        override func paint() {
            paints += 1
            NSColor.red.setFill()
            NSRect(x: 0, y: 0, width: 1, height: 1).fill()
        }
    }

    private func view(frame: NSRect = NSRect(x: 0, y: 0, width: 40, height: 30)) -> RedSquareView {
        let view = RedSquareView(frame: frame)
        view.layer?.contentsScale = 1
        return view
    }

    @Test func updateLayerPaintsOncePerCallAndSetsContents() {
        let view = view()
        view.updateLayer()
        #expect(view.paints == 1)
        #expect(view.layer?.contents != nil)
        view.updateLayer()
        #expect(view.paints == 2)
    }

    @Test func renderImageIsScaledAndTopLeftIsRed() throws {
        let image = try #require(view().renderImage(scale: 2))
        #expect(image.width == 80)
        #expect(image.height == 60)

        var pixels = [UInt8](repeating: 0, count: 80 * 60 * 4)
        let space = try #require(CGColorSpace(name: CGColorSpace.sRGB))
        try pixels.withUnsafeMutableBytes { buffer in
            let context = try #require(CGContext(
                data: buffer.baseAddress, width: 80, height: 60, bitsPerComponent: 8,
                bytesPerRow: 80 * 4, space: space,
                bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
            ))
            context.draw(image, in: CGRect(x: 0, y: 0, width: 80, height: 60))
        }
        #expect(pixels[0] > 230)
        #expect(pixels[1] < 25)
        #expect(pixels[2] < 25)
    }

    @Test func emptyViewRendersNothing() {
        #expect(view(frame: .zero).renderImage(scale: 2) == nil)
    }
}
