import CoreGraphics

/// Ephemeral pan/zoom camera for a graph canvas (design note §5.2).
///
/// Not research state — never written to the catalog. Spike 6-01 keeps this
/// in memory only; leaving the Evidence graph discards it.
struct GraphCanvasCamera: Equatable, Sendable {
    /// Pinch / magnification factor applied by `NSScrollView`.
    var magnification: CGFloat
    /// Top-left of the visible rect in **content** (document) space.
    var contentOffset: CGPoint

    static let minMagnification: CGFloat = 0.25
    static let maxMagnification: CGFloat = 4

    static let `default` = GraphCanvasCamera(magnification: 1, contentOffset: .zero)

    init(magnification: CGFloat, contentOffset: CGPoint) {
        self.magnification = GraphCanvasCoordinates.clampMagnification(magnification)
        self.contentOffset = contentOffset
    }
}
