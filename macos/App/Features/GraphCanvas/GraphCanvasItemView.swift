import AppKit

/// Base class of every ``GraphCanvasItem`` view.
///
/// A subclass draws in ``paint()``. The view is its own layer and shows an
/// image of that paint, made in ``updateLayer()`` only when `needsDisplay` is
/// set, so a pan, which changes only the visible rect, repaints nothing.
/// `needsDisplay` is the one repaint signal: appearance, window, backing scale,
/// the item layer's `contentsScale`, and a host whose content changed set it.
class GraphCanvasItemView: NSView {
    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        wantsLayer = true
        layerContentsRedrawPolicy = .onSetNeedsDisplay
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override var isFlipped: Bool { true }
    override var isOpaque: Bool { false }
    override var wantsUpdateLayer: Bool { true }

    /// Draws the item in the current `NSGraphicsContext`, flipped, in `bounds`.
    func paint() {}

    /// Whole points on both sides, so `bounds` is integral and the image maps
    /// one to one onto pixels at an integer scale.
    override func setFrameSize(_ newSize: NSSize) {
        super.setFrameSize(NSSize(width: ceil(newSize.width), height: ceil(newSize.height)))
    }

    /// Keeps `cacheDisplay`, printing and bitmap tests on the same paint.
    final override func draw(_ dirtyRect: NSRect) {
        paint()
    }

    /// `cacheDisplay` and printing paint directly through ``draw(_:)`` instead
    /// of sampling the layer's image, so they are exact at any scale. The context
    /// arrives unflipped, so this flips it to the view's coordinates. Item
    /// views have no subviews, so there is nothing for `super` to draw.
    final override func displayIgnoringOpacity(_ rect: NSRect, in context: NSGraphicsContext) {
        let cg = context.cgContext
        cg.saveGState()
        cg.translateBy(x: 0, y: bounds.height)
        cg.scaleBy(x: 1, y: -1)
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = NSGraphicsContext(cgContext: cg, flipped: true)
        draw(rect)
        NSGraphicsContext.restoreGraphicsState()
        cg.restoreGState()
    }

    /// The paint as an image `scale` pixels per point, top row first, or nil
    /// for an empty view. Goes through `cacheDisplay` into a point-sized
    /// bitmap, so AppKit scales shadows and the screen, tests and printing
    /// share one path.
    func renderImage(scale: CGFloat) -> CGImage? {
        guard bounds.width > 0, bounds.height > 0 else { return nil }
        guard let rep = NSBitmapImageRep(
            bitmapDataPlanes: nil,
            pixelsWide: max(1, Int((bounds.width * scale).rounded(.up))),
            pixelsHigh: max(1, Int((bounds.height * scale).rounded(.up))),
            bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
            colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0
        )
        else { return nil }
        rep.size = bounds.size
        effectiveAppearance.performAsCurrentDrawingAppearance {
            cacheDisplay(in: bounds, to: rep)
        }
        return rep.cgImage
    }

    override func updateLayer() {
        guard let layer else { return }
        layer.contents = renderImage(scale: layer.contentsScale)
        layer.contentsGravity = .resize
    }

    override func viewDidChangeEffectiveAppearance() {
        super.viewDidChangeEffectiveAppearance()
        needsDisplay = true
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        needsDisplay = true
    }

    override func viewDidChangeBackingProperties() {
        super.viewDidChangeBackingProperties()
        needsDisplay = true
    }
}
