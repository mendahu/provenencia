import AppKit
import SwiftUI

/// Draws kit-token shapes, text, symbols and marks into the current
/// `NSGraphicsContext`, for views that paint themselves (a canvas item) where
/// SwiftUI is too slow to host thousands of nodes.
///
/// Colors are kit `Color` tokens resolved for the appearance being drawn, so
/// a cached or tinted image never sticks on the default (black or accent).
/// Radii are continuous, like `PVRadius` shapes in SwiftUI. Text sizes with
/// `PVTextMetrics`. A host that measures without drawing passes no painter.
@MainActor
final class PVPainter {
    init() {}

    /// Resolves a token to the appearance being drawn. Use it for any color
    /// baked into a cached image or layer.
    static func color(_ color: Color, in appearance: NSAppearance = .currentDrawing()) -> NSColor {
        var resolved = NSColor(color)
        appearance.performAsCurrentDrawingAppearance {
            resolved = NSColor(cgColor: NSColor(color).cgColor) ?? resolved
        }
        return resolved
    }

    /// A continuous (squircle) rounded rectangle as a bezier path.
    static func continuous(_ rect: CGRect, radius: CGFloat) -> NSBezierPath {
        let path = NSBezierPath()
        var current = CGPoint.zero
        Path(roundedRect: rect, cornerRadius: radius, style: .continuous).forEach { element in
            switch element {
            case .move(let point):
                path.move(to: point)
                current = point
            case .line(let point):
                path.line(to: point)
                current = point
            case .quadCurve(let point, let control):
                // A quad curve as a cubic: each control sits two thirds of the
                // way from its end to the quad's control.
                path.curve(
                    to: point,
                    controlPoint1: CGPoint(
                        x: current.x + 2 / 3 * (control.x - current.x),
                        y: current.y + 2 / 3 * (control.y - current.y)
                    ),
                    controlPoint2: CGPoint(
                        x: point.x + 2 / 3 * (control.x - point.x),
                        y: point.y + 2 / 3 * (control.y - point.y)
                    )
                )
                current = point
            case .curve(let point, let control1, let control2):
                path.curve(to: point, controlPoint1: control1, controlPoint2: control2)
                current = point
            case .closeSubpath:
                path.close()
            }
        }
        return path
    }

    func fill(_ rect: CGRect, _ color: Color) {
        Self.color(color).setFill()
        rect.fill()
    }

    func fillRound(_ rect: CGRect, radius: CGFloat, _ color: Color, shadow: PVShadowLayer? = nil) {
        let path = Self.continuous(rect, radius: radius)
        if let shadow {
            NSGraphicsContext.saveGraphicsState()
            let shade = NSShadow()
            shade.shadowBlurRadius = shadow.radius
            shade.shadowOffset = NSSize(width: shadow.x, height: shadow.y)
            shade.shadowColor = Self.color(shadow.color)
            shade.set()
            Self.color(color).setFill()
            path.fill()
            NSGraphicsContext.restoreGraphicsState()
        } else {
            Self.color(color).setFill()
            path.fill()
        }
    }

    func beginClip(_ rect: CGRect, radius: CGFloat) {
        NSGraphicsContext.saveGraphicsState()
        Self.continuous(rect, radius: radius).setClip()
    }

    func endClip() {
        NSGraphicsContext.restoreGraphicsState()
    }

    func strokeRound(_ rect: CGRect, radius: CGFloat, _ color: Color, width: CGFloat, dash: [CGFloat] = []) {
        let path = Self.continuous(rect, radius: radius)
        path.lineWidth = width
        if !dash.isEmpty { path.setLineDash(dash, count: dash.count, phase: 0) }
        Self.color(color).setStroke()
        path.stroke()
    }

    func text(
        _ string: String,
        in rect: CGRect,
        font: NSFont,
        color: Color,
        lines: Int,
        kern: CGFloat = 0,
        uppercased: Bool = false
    ) {
        let style = NSMutableParagraphStyle()
        style.lineBreakMode = .byWordWrapping
        let attributes: [NSAttributedString.Key: Any] = [
            .font: font,
            .foregroundColor: Self.color(color),
            .paragraphStyle: style,
            .kern: kern,
        ]
        var box = rect
        box.size.height = min(box.height, PVTextMetrics.maxHeight(font: font, lines: lines))
        NSAttributedString(string: uppercased ? string.uppercased() : string, attributes: attributes)
            .draw(with: box, options: [.usesLineFragmentOrigin, .truncatesLastVisibleLine])
    }

    func symbol(_ symbol: PVSymbol, in rect: CGRect, _ color: Color) {
        guard let image = PVSymbolImages.image(symbol, pointSize: min(rect.width, rect.height)) else { return }
        drawTinted(image, in: rect, color: Self.color(color))
    }

    /// An aspect-fit image in a rounded tile; the placeholder fill when not loaded.
    func tile(_ image: NSImage?, in rect: CGRect, radius: CGFloat) {
        NSGraphicsContext.saveGraphicsState()
        Self.continuous(rect, radius: radius).addClip()
        Self.color(PVColor.surfaceHover).setFill()
        rect.fill()
        if let image, image.size.width > 0, image.size.height > 0 {
            let fit = min(rect.width / image.size.width, rect.height / image.size.height)
            let size = CGSize(width: image.size.width * fit, height: image.size.height * fit)
            let drawRect = CGRect(
                x: rect.midX - size.width / 2, y: rect.midY - size.height / 2,
                width: size.width, height: size.height
            )
            image.draw(in: drawRect, from: .zero, operation: .sourceOver, fraction: 1, respectFlipped: true, hints: nil)
        }
        NSGraphicsContext.restoreGraphicsState()
    }

    /// Recipe artwork in a component slot, the same exception `PVThumbnail`
    /// makes: the mark's asset belongs to the product, the drawing to the kit.
    func mark(_ key: PVMarkKey, in rect: CGRect, _ color: Color) {
        guard let image = NSImage(named: key.assetName) else { return }
        drawTinted(image, in: rect, color: Self.color(color))
    }

    /// Recolors a template image in the view's current context.
    ///
    /// An intermediate `NSImage` rasterizes the SVG mark at one pixel per
    /// point. The canvas renders items at up to 2.5x (backing scale times the
    /// zoom), so that bitmap would be scaled up and the mark would go soft. A
    /// transparency layer tints the vector at the same scale as the rest of the
    /// item.
    private func drawTinted(_ image: NSImage, in rect: CGRect, color: NSColor) {
        guard let context = NSGraphicsContext.current else { return }
        context.saveGraphicsState()
        context.cgContext.beginTransparencyLayer(in: rect, auxiliaryInfo: nil)
        image.draw(
            in: rect,
            from: .zero,
            operation: .sourceOver,
            fraction: 1,
            respectFlipped: true,
            hints: [.interpolation: NSImageInterpolation.high]
        )
        color.setFill()
        rect.fill(using: .sourceAtop)
        context.cgContext.endTransparencyLayer()
        context.restoreGraphicsState()
    }
}
