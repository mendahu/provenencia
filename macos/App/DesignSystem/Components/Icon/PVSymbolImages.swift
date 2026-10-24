import AppKit

/// SF Symbol images for views that paint with AppKit, configured once per
/// symbol and point size. A reused `NSImage` lets AppKit keep its own
/// representation cache across paints. `PVPainter.symbol` still draws the
/// vector from these images, so icons stay crisp at any zoom; only the image
/// object is cached.
@MainActor
enum PVSymbolImages {
    private struct Key: Hashable {
        let symbol: PVSymbol
        let pointSize: CGFloat
    }

    private static var images: [Key: NSImage] = [:]

    /// The configured image, or nil for a name AppKit does not know (a custom
    /// `pv.` glyph), which is not stored.
    static func image(_ symbol: PVSymbol, pointSize: CGFloat) -> NSImage? {
        let key = Key(symbol: symbol, pointSize: pointSize)
        if let stored = images[key] { return stored }
        guard let base = NSImage(systemSymbolName: symbol.rawValue, accessibilityDescription: nil) else { return nil }
        let image = base.withSymbolConfiguration(
            NSImage.SymbolConfiguration(pointSize: pointSize, weight: .regular)
        ) ?? base
        images[key] = image
        return image
    }

    #if DEBUG
    static func isCachedForTesting(_ symbol: PVSymbol, pointSize: CGFloat) -> Bool {
        images[Key(symbol: symbol, pointSize: pointSize)] != nil
    }
    #endif
}
