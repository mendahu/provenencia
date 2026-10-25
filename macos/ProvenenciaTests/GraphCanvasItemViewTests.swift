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

    private func view(contentSize: CGSize = CGSize(width: 40, height: 30)) -> RedSquareView {
        let view = RedSquareView(contentSize: contentSize)
        view.layer?.contentsScale = 1
        return view
    }

    @Test func intrinsicContentSizeIsTheFrameSize() {
        let view = view(contentSize: CGSize(width: 40, height: 30.4))
        #expect(view.intrinsicContentSize == view.frame.size)
    }

    @Test func updateLayerPaintsOncePerCallAndSetsContents() {
        let view = view()
        view.updateLayer()
        #expect(view.paints == 1)
        #expect(view.layer?.contents != nil)
        view.updateLayer()
        #expect(view.paints == 1)
        #expect(view.layer?.contents != nil)
    }

    @Test func updateLayerRepaintsOnlyWhenTheKeyOrTheContentChanges() {
        let view = view()
        view.appearance = NSAppearance(named: .aqua)
        var counts: [Int] = []
        func step(_ change: () -> Void = {}) {
            change()
            view.updateLayer()
            counts.append(view.paints)
            #expect(view.layer?.contents != nil)
        }
        step()
        step()
        step { view.layer?.contentsScale = 2 }
        step { view.invalidatePaint() }
        step { view.appearance = NSAppearance(named: .darkAqua) }
        #expect(counts == [1, 1, 2, 3, 4])
    }

    @Test func renderImageIsScaledAndTheContentOriginIsRed() throws {
        let outset = GraphCanvasItemView.paintOutset
        let image = try #require(view().renderImage(scale: 2))
        #expect(image.width == Int((40 + 2 * outset) * 2))
        #expect(image.height == Int((30 + 2 * outset) * 2))

        let rep = NSBitmapImageRep(cgImage: image)
        let at = Int(outset * 2)
        let color = try #require(rep.colorAt(x: at, y: at))
        #expect(color.redComponent > 0.9)
        #expect(color.greenComponent < 0.1)
        #expect(color.blueComponent < 0.1)
        #expect((rep.colorAt(x: at - 1, y: at - 1)?.alphaComponent ?? 0) < 0.05)
    }

    @Test func emptyContentRendersNothing() {
        #expect(view(contentSize: .zero).renderImage(scale: 2) == nil)
    }

    @Test func sizeIsTheContentPlusTheOutsetOnEverySide() throws {
        let outset = GraphCanvasItemView.paintOutset
        let view = view()
        #expect(view.intrinsicContentSize == NSSize(width: 40 + 2 * outset, height: 30 + 2 * outset))
        #expect(view.frame.size == view.intrinsicContentSize)

        let image = try #require(view.renderImage(scale: 1))
        let pixel = try #require(NSBitmapImageRep(cgImage: image).colorAt(x: Int(outset), y: Int(outset)))
        #expect(pixel.redComponent > 0.9)
        #expect(pixel.greenComponent < 0.1)
    }

    private final class ShadowSquareView: GraphCanvasItemView {
        override func paint() {
            let shadow = NSShadow()
            shadow.shadowOffset = NSSize(width: 0, height: -4)
            shadow.shadowBlurRadius = 0
            shadow.shadowColor = .black
            shadow.set()
            NSColor.red.setFill()
            NSRect(x: 10, y: 10, width: 10, height: 10).fill()
        }
    }

    private final class LineView: GraphCanvasItemView {
        override func paint() {
            NSColor.black.setFill()
            NSRect(x: 0, y: 20, width: 40, height: 1).fill()
        }
    }

    @Test(arguments: [CGFloat(1), CGFloat(2)])
    func shadowOffsetKeepsItsPointSizeAtAnyScale(scale: CGFloat) throws {
        let outset = GraphCanvasItemView.paintOutset
        let view = ShadowSquareView(contentSize: CGSize(width: 40, height: 40))
        let image = try #require(view.renderImage(scale: scale))
        let rep = NSBitmapImageRep(cgImage: image)
        let column = Int((15 + outset) * scale)
        var lowest = -1
        for y in 0..<rep.pixelsHigh where (rep.colorAt(x: column, y: y)?.alphaComponent ?? 0) > 0.5 {
            lowest = y
        }
        // The square ends at y = 20; its shadow ends 4 pt below.
        #expect(CGFloat(lowest + 1) / scale == 24 + outset)
    }

    @Test func fractionalFrameHeightRoundsUpAndLineStaysCrisp() throws {
        let outset = GraphCanvasItemView.paintOutset
        let view = LineView(contentSize: CGSize(width: 40, height: 30.4))
        #expect(view.bounds.height == ceil(30.4 + 2 * outset))
        let image = try #require(view.renderImage(scale: 1))
        let rep = NSBitmapImageRep(cgImage: image)
        let line = 20 + Int(outset)
        func alpha(_ y: Int) -> CGFloat { rep.colorAt(x: Int(outset) + 5, y: y)?.alphaComponent ?? 0 }
        #expect(alpha(line) > 250.0 / 255)
        #expect(alpha(line - 1) < 5.0 / 255)
        #expect(alpha(line + 1) < 5.0 / 255)
    }
}
