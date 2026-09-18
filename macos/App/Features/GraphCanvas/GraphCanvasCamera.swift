import CoreGraphics

/// Ephemeral pan/zoom camera for a graph canvas (design note §5.2).
///
/// Reusable canvas state — not research data, never written to the catalog.
/// Callers decide persistence (Evidence graph S6-01: in-memory only).
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
