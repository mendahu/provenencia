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

    @Test func renderImageIsScaledAndTopLeftIsRed() throws {
        let image = try #require(view().renderImage(scale: 2))
        #expect(image.width == 80)
        #expect(image.height == 60)

        let color = try #require(NSBitmapImageRep(cgImage: image).colorAt(x: 0, y: 0))
        #expect(color.redComponent > 0.9)
        #expect(color.greenComponent < 0.1)
        #expect(color.blueComponent < 0.1)
    }

    @Test func emptyViewRendersNothing() {
        #expect(view(frame: .zero).renderImage(scale: 2) == nil)
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
        let view = ShadowSquareView(frame: NSRect(x: 0, y: 0, width: 40, height: 40))
        let image = try #require(view.renderImage(scale: scale))
        let rep = NSBitmapImageRep(cgImage: image)
        let column = Int(15 * scale)
        var lowest = -1
        for y in 0..<rep.pixelsHigh where (rep.colorAt(x: column, y: y)?.alphaComponent ?? 0) > 0.5 {
            lowest = y
        }
        // The square ends at y = 20; its shadow ends 4 pt below.
        #expect(CGFloat(lowest + 1) / scale == 24)
    }

    @Test func fractionalFrameHeightRoundsUpAndLineStaysCrisp() throws {
        let view = LineView(frame: NSRect(x: 0, y: 0, width: 40, height: 30))
        view.setFrameSize(NSSize(width: 40, height: 30.4))
        #expect(view.bounds.height == 31)
        let image = try #require(view.renderImage(scale: 1))
        let rep = NSBitmapImageRep(cgImage: image)
        func alpha(_ y: Int) -> CGFloat { rep.colorAt(x: 5, y: y)?.alphaComponent ?? 0 }
        #expect(alpha(20) > 250.0 / 255)
        #expect(alpha(19) < 5.0 / 255)
        #expect(alpha(21) < 5.0 / 255)
    }
}
