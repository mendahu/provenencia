import AppKit

/// Base class of every ``GraphCanvasItem`` view.
///
/// A subclass draws in ``paint()``. The view is its own layer and shows an
/// image of that paint. ``updateLayer()`` runs on every pan and zoom step, so it
/// reuses the last image and paints again only when the pixel size at the
/// layer's `contentsScale` or the appearance changed, or after
/// ``invalidatePaint()``. A subclass whose content changed calls
/// ``invalidatePaint()``; `needsDisplay` alone shows the last image again.
///
/// The view is its content plus ``paintOutset`` on every side, so shadows and
/// halos that paint outside the content are not clipped. ``paint()`` draws in
/// content coordinates: (0, 0) is the content's top-left, not the view's.
class GraphCanvasItemView: NSView {
    /// Transparent margin around the content on every side.
    static let paintOutset: CGFloat = 24

    /// The size of the content, without the margin.
    let contentSize: CGSize

    private struct PaintKey: Equatable {
        var pixelWidth: Int
        var pixelHeight: Int
        var appearance: NSAppearance.Name
    }

    private var image: CGImage?
    private var imageKey: PaintKey?

    init(contentSize: CGSize) {
        self.contentSize = contentSize
        super.init(frame: NSRect(origin: .zero, size: Self.frameSize(for: contentSize)))
        wantsLayer = true
        layerContentsRedrawPolicy = .onSetNeedsDisplay
    }

    /// The content plus the margin on every side, in whole points.
    private static func frameSize(for contentSize: CGSize) -> NSSize {
        NSSize(
            width: ceil(contentSize.width + 2 * paintOutset),
            height: ceil(contentSize.height + 2 * paintOutset)
        )
    }

    @available(*, unavailable)
    override init(frame frameRect: NSRect) {
        fatalError("init(frame:) has not been implemented; use init(contentSize:)")
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override var isFlipped: Bool { true }
    override var isOpaque: Bool { false }
    override var wantsUpdateLayer: Bool { true }

    /// The content plus the margin on every side.
    final override var intrinsicContentSize: NSSize {
        Self.frameSize(for: contentSize)
    }

    /// Draws the item in the current `NSGraphicsContext`, flipped, in content
    /// coordinates: (0, 0) is the content's top-left and `contentSize` its extent.
    func paint() {}

    /// Whole points on both sides, so `bounds` is integral and the image maps
    /// one to one onto pixels at an integer scale.
    override func setFrameSize(_ newSize: NSSize) {
        super.setFrameSize(NSSize(width: ceil(newSize.width), height: ceil(newSize.height)))
    }

    /// Keeps `cacheDisplay`, printing and bitmap tests on the same paint.
    final override func draw(_ dirtyRect: NSRect) {
        guard let context = NSGraphicsContext.current else { return }
        context.saveGraphicsState()
        context.cgContext.translateBy(x: Self.paintOutset, y: Self.paintOutset)
        paint()
        context.restoreGraphicsState()
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

    /// The paint, margin included, as an image `scale` pixels per point, top row
    /// first, or nil for empty content. Goes through `cacheDisplay` into a point-sized
    /// bitmap, so AppKit scales shadows and the screen, tests and printing
    /// share one path.
    func renderImage(scale: CGFloat) -> CGImage? {
        guard let size = pixelSize(scale: scale) else { return nil }
        guard let rep = NSBitmapImageRep(
            bitmapDataPlanes: nil,
            pixelsWide: size.width,
            pixelsHigh: size.height,
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

    /// The image's size in pixels at `scale`; nil for empty content.
    private func pixelSize(scale: CGFloat) -> (width: Int, height: Int)? {
        guard contentSize.width > 0, contentSize.height > 0 else { return nil }
        return (
            max(1, Int((bounds.width * scale).rounded(.up))),
            max(1, Int((bounds.height * scale).rounded(.up)))
        )
    }

    /// Shows the last image, painting a new one only when there is none or the
    /// pixel size or appearance it was made for changed.
    override func updateLayer() {
        guard let layer else { return }
        let key = pixelSize(scale: layer.contentsScale).map {
            PaintKey(pixelWidth: $0.width, pixelHeight: $0.height, appearance: effectiveAppearance.name)
        }
        if image == nil || key != imageKey {
            image = renderImage(scale: layer.contentsScale)
            imageKey = image == nil ? nil : key
        }
        layer.contents = image
        layer.contentsGravity = .resize
    }

    /// Drops the image and asks for a repaint. A subclass calls it when the
    /// content it paints changed.
    final func invalidatePaint() {
        image = nil
        imageKey = nil
        needsDisplay = true
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
