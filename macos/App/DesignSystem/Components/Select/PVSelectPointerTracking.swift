import CoreGraphics

/// Press / drag / release hit target for a Select tracking loop.
enum PVSelectPointerTarget: Equatable {
    case trigger
    case row(Int)
    case outside
}

/// Pointer events the view reports into `PVSelectSession`.
enum PVSelectPointerEvent: Equatable {
    case press(PVSelectPointerTarget)
    case drag(PVSelectPointerTarget)
    case release(PVSelectPointerTarget)
    case clickAway
}

/// Hit-testing for press–drag–release. The view supplies screen (or any shared)
/// rectangles; this type does not talk to AppKit.
enum PVSelectPointerTracking {
    static func hitTest(
        point: CGPoint,
        trigger: CGRect,
        rows: [CGRect]
    ) -> PVSelectPointerTarget {
        for (index, row) in rows.enumerated() where row.contains(point) {
            return .row(index)
        }
        if trigger.contains(point) {
            return .trigger
        }
        return .outside
    }

    /// Equal slices of the panel's inner height, top-down in AppKit space
    /// (first row is the highest `maxY`).
    static func rowFrames(
        panel: CGRect,
        count: Int,
        padding: CGFloat = PVSelectPlacement.panelPadding / 2
    ) -> [CGRect] {
        guard count > 0 else { return [] }
        let inner = panel.insetBy(dx: padding, dy: padding)
        guard inner.height > 0, inner.width > 0 else { return [] }
        let rowHeight = inner.height / CGFloat(count)
        return (0..<count).map { index in
            let maxY = inner.maxY - CGFloat(index) * rowHeight
            return CGRect(
                x: inner.minX,
                y: maxY - rowHeight,
                width: inner.width,
                height: rowHeight
            )
        }
    }
}
