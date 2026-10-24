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

    /// Keeps `cacheDisplay`, printing and bitmap tests on the same paint.
    override func draw(_ dirtyRect: NSRect) {
        paint()
    }

    /// `cacheDisplay` and printing paint directly through ``draw(_:)`` instead
    /// of sampling the layer's image, so they are exact at any scale.
    override func displayIgnoringOpacity(_ rect: NSRect, in context: NSGraphicsContext) {
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = context
        draw(rect)
        NSGraphicsContext.restoreGraphicsState()
    }

    /// The paint as an image `scale` pixels per point, top row first, or nil
    /// for an empty view.
    func renderImage(scale: CGFloat) -> CGImage? {
        guard bounds.width > 0, bounds.height > 0 else { return nil }
        let w = max(1, Int(ceil(bounds.width * scale)))
        let h = max(1, Int(ceil(bounds.height * scale)))
        guard let space = CGColorSpace(name: CGColorSpace.sRGB),
              let context = CGContext(
                  data: nil, width: w, height: h, bitsPerComponent: 8, bytesPerRow: 0,
                  space: space, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
              )
        else { return nil }
        context.translateBy(x: 0, y: CGFloat(h))
        context.scaleBy(x: CGFloat(w) / bounds.width, y: -CGFloat(h) / bounds.height)
        let graphics = NSGraphicsContext(cgContext: context, flipped: true)
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = graphics
        effectiveAppearance.performAsCurrentDrawingAppearance { paint() }
        NSGraphicsContext.restoreGraphicsState()
        return context.makeImage()
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
