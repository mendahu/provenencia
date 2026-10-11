import CoreGraphics

/// Pure view ↔ content coordinate helpers for layout and unit tests.
///
/// **Live hit-testing** under a magnified `NSScrollView` uses AppKit
/// `NSView.convert(_:from:)` via `GraphCanvasViewportController.contentPointUnderCursor()`
/// — do not route pointer place/ghost through these helpers alone (design note §7.2).
/// This enum stays the pure math seam for camera transforms, magnification clamp,
/// and scroll-origin clamp shared by pan + zoom.
enum GraphCanvasCoordinates {
    static func clampMagnification(_ value: CGFloat) -> CGFloat {
        min(max(value, GraphCanvasCamera.minMagnification), GraphCanvasCamera.maxMagnification)
    }

    /// Document-space width of one screen point at `magnification`.
    ///
    /// `NSScrollView` scales the document, so a stroke of `1 / magnification`
    /// stays one screen point at every zoom.
    static func screenPixelDocumentWidth(magnification: CGFloat) -> CGFloat {
        1 / max(magnification, 0.000_1)
    }

    /// Opacity for a repeated grid cell. The pattern stays fully visible while
    /// a cell is clearly larger than a screen point, and fades out as a whole
    /// once the cell itself is about one point — hairlines must not pile into
    /// a solid field.
    static func gridPatternOpacity(spacing: CGFloat, magnification: CGFloat) -> CGFloat {
        let cell = spacing * max(magnification, 0)
        if cell >= 2 { return 1 }
        if cell <= 1 { return 0 }
        return cell - 1
    }

    /// Clamps a proposed content-view scroll origin so the visible rect stays
    /// inside the document. Shared by click-drag pan and mouse-wheel zoom.
    static func clampedContentOrigin(
        proposed: CGPoint,
        documentSize: CGSize,
        visibleSize: CGSize
    ) -> CGPoint {
        let maxX = max(0, documentSize.width - visibleSize.width)
        let maxY = max(0, documentSize.height - visibleSize.height)
        return CGPoint(
            x: min(max(proposed.x, 0), maxX),
            y: min(max(proposed.y, 0), maxY)
        )
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

/// Classifies scroll-wheel input for GraphCanvas (pan vs zoom).
enum GraphCanvasScrollInput {
    /// Precise / phase-bearing events come from trackpads and Magic Mouse —
    /// keep those as pan. Traditional mouse wheels are line-delta only.
    static func isTrackpadStyleScroll(
        hasPreciseScrollingDeltas: Bool,
        hasPhase: Bool,
        hasMomentumPhase: Bool
    ) -> Bool {
        if hasPreciseScrollingDeltas { return true }
        if hasPhase || hasMomentumPhase { return true }
        return false
    }
}
