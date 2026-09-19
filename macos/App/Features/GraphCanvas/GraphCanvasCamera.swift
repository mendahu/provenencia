import CoreGraphics

/// Ephemeral pan/zoom camera for a graph canvas (design note §5.2).
///
/// Reusable canvas state — not research data, never written to the catalog.
/// Callers decide persistence (product hosts typically keep camera in-memory).
struct GraphCanvasCamera: Equatable, Sendable {
    /// Pinch / magnification factor applied by `NSScrollView`.
    var magnification: CGFloat
    /// Top-left of the visible rect in **content** (document) space.
    var contentOffset: CGPoint

    /// Default zoom-out floor for overview of a large document.
    /// Hosts may clamp differently; these are GraphCanvas defaults only.
    static let minMagnification: CGFloat = 0.25
    /// Cap zoom-in near identity — magnification is for overview, not to
    /// enlarge hosted views (layer zoom would look soft past ~1.25 anyway).
    static let maxMagnification: CGFloat = 1.25

    static let `default` = GraphCanvasCamera(magnification: 1, contentOffset: .zero)

    init(magnification: CGFloat, contentOffset: CGPoint) {
        self.magnification = GraphCanvasCoordinates.clampMagnification(magnification)
        self.contentOffset = contentOffset
    }
}
