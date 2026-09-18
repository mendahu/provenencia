import CoreGraphics

/// Pure view ↔ content coordinate conversion under magnification.
///
/// `NSScrollView` magnification does not correctly map SwiftUI gesture
/// locations into document space on its own (design note §7.2). All canvas
/// hit-testing and drag math must go through this seam.
enum GraphCanvasCoordinates {
    static func clampMagnification(_ value: CGFloat) -> CGFloat {
        min(max(value, GraphCanvasCamera.minMagnification), GraphCanvasCamera.maxMagnification)
    }

    /// Converts a point in the scroll view's visible/clip space into content
    /// (document) coordinates.
    ///
    /// - Parameters:
    ///   - viewPoint: Location in the unmagnified clip view (points from the
    ///     visible top-leading corner).
    ///   - camera: Current magnification and content-space visible origin.
    static func contentPoint(fromViewPoint viewPoint: CGPoint, camera: GraphCanvasCamera) -> CGPoint {
        let m = clampMagnification(camera.magnification)
        return CGPoint(
            x: camera.contentOffset.x + viewPoint.x / m,
            y: camera.contentOffset.y + viewPoint.y / m
        )
    }

    /// Converts a content (document) point into the scroll view's visible/clip space.
    static func viewPoint(fromContentPoint contentPoint: CGPoint, camera: GraphCanvasCamera) -> CGPoint {
        let m = clampMagnification(camera.magnification)
        return CGPoint(
            x: (contentPoint.x - camera.contentOffset.x) * m,
            y: (contentPoint.y - camera.contentOffset.y) * m
        )
    }
}
