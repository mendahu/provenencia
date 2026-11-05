@preconcurrency import AppKit
import SwiftUI

/// AppKit overlay in document space: draw / edit / dim-outside (spatial kinds only).
final class ArtifactRegionOverlayView: NSView {
    var armedTool: ArtifactRegionTool? {
        didSet {
            if oldValue != armedTool {
                inProgressPoints = []
                dragStart = nil
                dragCurrent = nil
                needsDisplay = true
                window?.invalidateCursorRects(for: self)
            }
        }
    }

    var committed: ArtifactRegionDraft? {
        didSet {
            if oldValue != committed {
                needsDisplay = true
                window?.invalidateCursorRects(for: self)
            }
        }
    }

    var imageSize: CGSize = .zero {
        didSet {
            if oldValue != imageSize {
                needsDisplay = true
                window?.invalidateCursorRects(for: self)
            }
        }
    }

    /// When set (PDF remount), media box in this view's flipped coordinates.
    /// Image viewport leaves this nil and uses ``imageSize`` + centering pad.
    var mediaRectOverride: CGRect? {
        didSet {
            if oldValue != mediaRectOverride {
                needsDisplay = true
                window?.invalidateCursorRects(for: self)
            }
        }
    }

    /// The scroll view's zoom (image viewport). Handles keep one screen-point size at any zoom;
    /// the PDF host never magnifies the overlay and leaves this at 1.
    var magnification: CGFloat = 1 {
        didSet {
            if oldValue != magnification {
                needsDisplay = true
                window?.invalidateCursorRects(for: self)
            }
        }
    }

    /// Returns whether the host accepted the region; a refused edit is put back.
    var onCommit: ((ArtifactRegionDraft) -> Bool)?
    var onDisarm: (() -> Void)?

    private var dragStart: CGPoint?
    private var dragCurrent: CGPoint?
    private var inProgressPoints: [CGPoint] = [] {
        didSet {
            if inProgressPoints.count < ArtifactRegionGeometry.maxRegionPoints { capAnnounced = false }
        }
    }
    /// VoiceOver has heard the freehand cap for this shape; re-armed when the shape shrinks or restarts.
    private var capAnnounced = false
    /// The drag in flight on the committed region: what is grabbed, where the pointer went down,
    /// and the region as it was, so an invalid result can put it back. `moved` latches once the
    /// pointer travels past the drag threshold; a press that never does edits nothing.
    private var edit: (handle: ArtifactRegionHandle, origin: CGPoint, start: ArtifactRegionDraft, moved: Bool)?
    private var tracking: NSTrackingArea?
    /// A cursor from `applyCursor` is showing, so leaving the hit gives the arrow back.
    private var didSetCursor = false

    override var isFlipped: Bool { true }
    override var acceptsFirstResponder: Bool { armedTool != nil || committed != nil }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if armedTool != nil {
            window?.makeFirstResponder(self)
        }
    }

    override func updateTrackingAreas() {
        super.updateTrackingAreas()
        if let tracking {
            removeTrackingArea(tracking)
        }
        let area = NSTrackingArea(
            rect: bounds,
            options: [.mouseMoved, .cursorUpdate, .activeInKeyWindow, .inVisibleRect],
            owner: self
        )
        addTrackingArea(area)
        tracking = area
    }

    override func resetCursorRects() {
        if armedTool != nil {
            addCursorRect(bounds, cursor: .crosshair)
            return
        }
        guard let committed else { return }
        let centroid = cursorCentroid(for: committed)
        for target in ArtifactRegionGeometry.handleTargets(for: committed, imageRect: imageRect) {
            addCursorRect(
                cursorRect(around: target.document, pad: handleHitRadius),
                cursor: cursor(for: target, centroid: centroid)
            )
        }
    }

    override func cursorUpdate(with event: NSEvent) {
        applyCursor(at: convert(event.locationInWindow, from: nil))
    }

    override func mouseMoved(with event: NSEvent) {
        applyCursor(at: convert(event.locationInWindow, from: nil))
    }

    override func hitTest(_ point: NSPoint) -> NSView? {
        if armedTool != nil { return self }
        // AppKit hands `hitTest` a point in the superview's coordinates.
        let local = convert(point, from: superview)
        return hitHandle(at: local) != nil ? self : nil
    }

    override func draw(_ dirtyRect: NSRect) {
        // The veil is translucent, so a new citation's region would otherwise
        // paint over the previous one and leave that polygon on screen.
        dirtyRect.fill(using: .clear)
        let media = imageRect
        guard media.width > 0, media.height > 0 else { return }

        if let committed, armedTool == nil || edit != nil {
            drawCommitted(committed, imageRect: media)
        }

        if armedTool != nil, let preview = previewDraft() {
            drawInProgress(preview, imageRect: media)
        }
    }

    override func mouseDown(with event: NSEvent) {
        let document = convert(event.locationInWindow, from: nil)
        let normalized = ArtifactRegionGeometry.normalize(documentPoint: document, imageRect: imageRect)

        if let tool = armedTool {
            switch tool {
            case .freeform:
                handleFreeformClick(normalized)
            case .rectangle, .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft, .circle:
                dragStart = normalized
                dragCurrent = normalized
                needsDisplay = true
            }
            window?.makeFirstResponder(self)
            return
        }

        if let handle = hitHandle(at: document), let committed {
            edit = (handle, normalized, committed, false)
            if handle == .move {
                NSCursor.closedHand.set()
            }
            window?.makeFirstResponder(self)
            return
        }
    }

    override func mouseDragged(with event: NSEvent) {
        let document = convert(event.locationInWindow, from: nil)
        let normalized = ArtifactRegionGeometry.normalize(documentPoint: document, imageRect: imageRect)

        if armedTool != nil, armedTool != .freeform {
            dragCurrent = normalized
            needsDisplay = true
            return
        }

        if var current = edit, var draft = committed {
            if !current.moved {
                let press = ArtifactRegionGeometry.documentPoint(normalized: current.origin, imageRect: imageRect)
                guard GraphCanvasPointerHitTesting.isDrag(
                    from: press,
                    to: document,
                    threshold: GraphCanvasPointerHitTesting.dragThreshold * screenPoint
                ) else { return }
                current.moved = true
                edit = current
            }
            let delta = CGPoint(x: normalized.x - current.origin.x, y: normalized.y - current.origin.y)
            switch current.handle {
            case .move:
                draft.points = ArtifactRegionGeometry.translate(current.start.points, by: delta)
            case .resize(let index):
                // The vertex keeps the offset it was grabbed at instead of jumping to the pointer.
                let grabbed = ArtifactRegionGeometry.handlePoints(for: current.start, imageRect: imageRect)
                guard grabbed.indices.contains(index) else { return }
                let target = CGPoint(
                    x: min(max(grabbed[index].x + delta.x, 0), 1),
                    y: min(max(grabbed[index].y + delta.y, 0), 1)
                )
                draft.points = moveHandle(draft: draft, handle: index, to: target)
            }
            committed = draft
            needsDisplay = true
        }
    }

    override func mouseUp(with event: NSEvent) {
        if let tool = armedTool, tool != .freeform {
            if let draft = previewDraft(), draft.isValid {
                commitAndDisarm(draft)
            } else {
                dragStart = nil
                dragCurrent = nil
                needsDisplay = true
            }
            return
        }

        if let edit, let draft = committed {
            finishEdit(draft, start: edit.start)
        }
        edit = nil
    }

    override func rightMouseDown(with event: NSEvent) {
        guard let committed, committed.kind == .freeform else {
            super.rightMouseDown(with: event)
            return
        }
        let document = convert(event.locationInWindow, from: nil)
        guard let index = freeformVertexIndex(at: document),
              let points = ArtifactRegionGeometry.deleteFreeformVertex(points: committed.points, index: index)
        else { return }
        var next = committed
        next.points = points
        finishEdit(next, start: committed)
    }

    override func keyDown(with event: NSEvent) {
        guard armedTool == .freeform else {
            super.keyDown(with: event)
            return
        }
        if event.charactersIgnoringModifiers == String(UnicodeScalar(NSDeleteCharacter)!)
            || event.keyCode == 51
        {
            if !inProgressPoints.isEmpty {
                inProgressPoints.removeLast()
                needsDisplay = true
            }
            return
        }
        if event.keyCode == 53 {
            inProgressPoints = []
            onDisarm?()
            needsDisplay = true
            return
        }
        super.keyDown(with: event)
    }

    // MARK: - Private

    private var imageRect: CGRect {
        if let mediaRectOverride { return mediaRectOverride }
        return ArtifactRegionGeometry.imageRect(documentSize: bounds.size, imageSize: imageSize)
    }

    private func handleFreeformClick(_ normalized: CGPoint) {
        if let first = inProgressPoints.first,
           inProgressPoints.count >= ArtifactRegionGeometry.freeformMinimumVertices,
           ArtifactRegionGeometry.closeEnoughToOrigin(normalized, origin: first)
        {
            let draft = ArtifactRegionDraft(kind: .freeform, points: inProgressPoints)
            if draft.isValid {
                commitAndDisarm(draft)
            }
            return
        }
        guard let next = ArtifactRegionGeometry.appendFreeformVertex(points: inProgressPoints, normalized) else {
            NSSound.beep()
            if !capAnnounced {
                capAnnounced = true
                AccessibilityAnnouncement.post(
                    L10n.ArtifactViewer.freeformCapReached(ArtifactRegionGeometry.maxRegionPoints)
                )
            }
            return
        }
        inProgressPoints = next
        needsDisplay = true
    }

    private func previewDraft() -> ArtifactRegionDraft? {
        guard let tool = armedTool else { return nil }
        switch tool {
        case .freeform:
            guard inProgressPoints.count >= 1 else { return nil }
            return ArtifactRegionDraft(kind: .freeform, points: inProgressPoints)
        case .rectangle, .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft, .circle:
            guard let start = dragStart, let current = dragCurrent else { return nil }
            let points: [CGPoint]
            switch tool {
            case .rectangle:
                points = ArtifactRegionGeometry.rectanglePoints(from: start, to: current)
            case .circle:
                points = ArtifactRegionGeometry.circleRing(
                    centerNormalized: start,
                    throughNormalized: current,
                    imageRect: imageRect
                )
            default:
                points = ArtifactRegionGeometry.lShapePoints(
                    kind: tool.regionKind,
                    from: start,
                    to: current
                )
            }
            return ArtifactRegionDraft(kind: tool.regionKind, points: points)
        }
    }

    /// Ends an edit of the committed region: unchanged does nothing, a valid region the host
    /// accepts is kept, anything else puts `start` back, beeps and announces.
    private func finishEdit(_ draft: ArtifactRegionDraft, start: ArtifactRegionDraft) {
        if draft == start { return }
        if draft.isValid, onCommit?(draft) == true {
            committed = draft
        } else {
            committed = start
            refuseEdit()
        }
    }

    private func refuseEdit() {
        NSSound.beep()
        AccessibilityAnnouncement.post(L10n.string(L10n.ArtifactViewer.regionEditRefused))
    }

    private func commitAndDisarm(_ draft: ArtifactRegionDraft) {
        committed = draft
        dragStart = nil
        dragCurrent = nil
        inProgressPoints = []
        armedTool = nil
        // The host's model update runs through `onDisarm` and corrects a refused region.
        _ = onCommit?(draft)
        onDisarm?()
        needsDisplay = true
    }

    private func moveHandle(draft: ArtifactRegionDraft, handle: Int, to location: CGPoint) -> [CGPoint] {
        switch draft.kind {
        case .rectangle:
            return ArtifactRegionGeometry.moveRectangleHandle(
                points: draft.points,
                handle: handle,
                to: location
            )
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            return ArtifactRegionGeometry.moveLHandle(
                kind: draft.kind,
                points: draft.points,
                handle: handle,
                to: location
            )
        case .circle:
            return ArtifactRegionGeometry.moveCircleRadius(
                points: draft.points,
                to: location,
                imageRect: imageRect
            )
        case .freeform:
            return ArtifactRegionGeometry.moveFreeformVertex(
                points: draft.points,
                index: handle,
                to: location
            )
        }
    }

    private func hitHandle(at documentPoint: CGPoint) -> ArtifactRegionHandle? {
        guard let committed else { return nil }
        return ArtifactRegionGeometry.hitHandle(
            at: documentPoint,
            draft: committed,
            imageRect: imageRect,
            radius: handleHitRadius,
            anchorRadius: moveAnchorHitRadius
        )
    }

    private func freeformVertexIndex(at documentPoint: CGPoint) -> Int? {
        guard let committed, committed.kind == .freeform else { return nil }
        let targets = ArtifactRegionGeometry.handleTargets(for: committed, imageRect: imageRect)
        guard case .resize(let index)? = ArtifactRegionGeometry.nearestHandle(
            at: documentPoint, in: targets, radius: handleHitRadius
        ) else { return nil }
        return index
    }

    /// One screen point in overlay coordinates.
    private var screenPoint: CGFloat {
        GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: magnification)
    }

    /// Hit radius in overlay points: 7 screen points at any zoom.
    var handleHitRadius: CGFloat { 7 * screenPoint }

    /// The move anchor reaches one screen point past a handle so it wins where they overlap.
    var moveAnchorHitRadius: CGFloat { 8 * screenPoint }

    /// Handle diameter in overlay points: 7 screen points at any zoom.
    var handleDrawSize: CGFloat { 7 * screenPoint }

    private func drawCommitted(_ draft: ArtifactRegionDraft, imageRect media: CGRect) {
        guard let path = shapePath(draft, imageRect: media) else { return }

        let veil = NSBezierPath(rect: media)
        veil.append(path)
        veil.windingRule = .evenOdd
        NSColor(PVColor.surfaceOverlay).setFill()
        veil.fill()

        NSColor(PVColor.accent).setStroke()
        path.lineWidth = 2
        path.lineJoinStyle = .round
        path.stroke()

        // Resize handles first, the move anchor last so it stays on top.
        let targets = ArtifactRegionGeometry.handleTargets(for: draft, imageRect: media)
        for target in targets where target.handle != .move {
            drawHandle(at: target.document)
        }
        for target in targets where target.handle == .move {
            drawMoveAnchor(at: target.document)
        }
    }

    private func drawInProgress(_ draft: ArtifactRegionDraft, imageRect media: CGRect) {
        if draft.kind == .circle, let path = circlePath(draft, imageRect: media) {
            NSColor(PVColor.accent).setStroke()
            path.lineWidth = 2
            path.stroke()
            if let center = ArtifactRegionGeometry.circleCenter(draft.points) {
                drawHandle(at: ArtifactRegionGeometry.documentPoint(normalized: center, imageRect: media))
            }
            for handle in ArtifactRegionGeometry.handlePoints(for: draft, imageRect: media) {
                drawHandle(at: ArtifactRegionGeometry.documentPoint(normalized: handle, imageRect: media))
            }
            return
        }
        let docs = ArtifactRegionGeometry.documentPoints(normalized: draft.points, imageRect: media)
        guard docs.count >= 1 else { return }
        let path = NSBezierPath()
        path.move(to: docs[0])
        for point in docs.dropFirst() {
            path.line(to: point)
        }
        let isFreeformPreview = armedTool == .freeform
        if !isFreeformPreview, docs.count >= 3 {
            path.close()
        }
        NSColor(PVColor.accent).setStroke()
        path.lineWidth = 2
        if isFreeformPreview {
            path.setLineDash([5, 4], count: 2, phase: 0)
        }
        path.stroke()
        for point in docs {
            drawHandle(at: point)
        }
    }

    private func shapePath(_ draft: ArtifactRegionDraft, imageRect media: CGRect) -> NSBezierPath? {
        if draft.kind == .circle {
            return circlePath(draft, imageRect: media)
        }
        let docs = ArtifactRegionGeometry.documentPoints(normalized: draft.points, imageRect: media)
        return polygonPath(docs)
    }

    /// True circle in image pixels — a unit-square ring would stretch into an oval.
    private func circlePath(_ draft: ArtifactRegionDraft, imageRect media: CGRect) -> NSBezierPath? {
        guard let center = ArtifactRegionGeometry.circleCenter(draft.points),
              let radius = ArtifactRegionGeometry.circleRadiusDocument(draft.points, imageRect: media),
              radius > 0
        else { return nil }
        let origin = ArtifactRegionGeometry.documentPoint(normalized: center, imageRect: media)
        return NSBezierPath(
            ovalIn: CGRect(x: origin.x - radius, y: origin.y - radius, width: radius * 2, height: radius * 2)
        )
    }

    private func polygonPath(_ points: [CGPoint]) -> NSBezierPath? {
        guard let first = points.first, points.count >= 3 else { return nil }
        let path = NSBezierPath()
        path.move(to: first)
        for point in points.dropFirst() {
            path.line(to: point)
        }
        path.close()
        return path
    }

    private func drawHandle(at point: CGPoint) {
        let size = handleDrawSize
        let rect = CGRect(x: point.x - size / 2, y: point.y - size / 2, width: size, height: size)
        let path = NSBezierPath(ovalIn: rect)
        NSColor(PVColor.surfaceRaised).setFill()
        NSColor(PVColor.accent).setStroke()
        path.fill()
        path.lineWidth = 1.5 * screenPoint
        path.stroke()
    }

    private func drawMoveAnchor(at point: CGPoint) {
        let size = 9 * screenPoint
        let path = NSBezierPath()
        path.move(to: CGPoint(x: point.x, y: point.y - size / 2))
        path.line(to: CGPoint(x: point.x + size / 2, y: point.y))
        path.line(to: CGPoint(x: point.x, y: point.y + size / 2))
        path.line(to: CGPoint(x: point.x - size / 2, y: point.y))
        path.close()
        NSColor(PVColor.accent).setFill()
        NSColor(PVColor.surfaceRaised).setStroke()
        path.fill()
        path.lineWidth = 1.25 * screenPoint
        path.stroke()
    }

    private func applyCursor(at documentPoint: CGPoint) {
        if armedTool != nil {
            NSCursor.crosshair.set()
            return
        }
        guard let committed else { return }
        guard let handle = hitHandle(at: documentPoint) else {
            // Give back only a cursor this overlay set; the pan hand and I-beam are not ours.
            if didSetCursor {
                NSCursor.arrow.set()
                didSetCursor = false
            }
            return
        }
        let targets = ArtifactRegionGeometry.handleTargets(for: committed, imageRect: imageRect)
        if let target = targets.first(where: { $0.handle == handle }) {
            cursor(for: target, centroid: cursorCentroid(for: committed)).set()
        } else {
            // The outline, which moves the shape and has no target of its own.
            NSCursor.openHand.set()
        }
        didSetCursor = true
    }

    private func cursorCentroid(for draft: ArtifactRegionDraft) -> CGPoint {
        ArtifactRegionGeometry.moveAnchor(for: draft) ?? CGPoint(x: 0.5, y: 0.5)
    }

    private func cursor(for target: ArtifactRegionTarget, centroid: CGPoint) -> NSCursor {
        switch target.handle {
        case .move:
            .openHand
        case .resize:
            ArtifactRegionCursor.resize(
                ArtifactRegionGeometry.resizeCursorKind(handle: target.normalized, centroid: centroid)
            )
        }
    }

    private func cursorRect(around point: CGPoint, pad: CGFloat) -> CGRect {
        CGRect(x: point.x - pad, y: point.y - pad, width: pad * 2, height: pad * 2)
    }
}

enum ArtifactRegionCursor {
    static func resize(_ kind: ArtifactResizeCursorKind) -> NSCursor {
        switch kind {
        case .east: .resizeRight
        case .west: .resizeLeft
        case .south: .resizeDown
        case .north: .resizeUp
        case .southEast, .northWest: diagonal(nwse: true)
        case .southWest, .northEast: diagonal(nwse: false)
        }
    }

    private static func diagonal(nwse: Bool) -> NSCursor {
        let size: CGFloat = 20
        let image = NSImage(size: NSSize(width: size, height: size), flipped: true) { rect in
            let inset = rect.insetBy(dx: 3.5, dy: 3.5)
            let start: CGPoint
            let end: CGPoint
            if nwse {
                start = CGPoint(x: inset.minX, y: inset.minY)
                end = CGPoint(x: inset.maxX, y: inset.maxY)
            } else {
                start = CGPoint(x: inset.minX, y: inset.maxY)
                end = CGPoint(x: inset.maxX, y: inset.minY)
            }
            let path = NSBezierPath()
            path.move(to: start)
            path.line(to: end)
            path.lineWidth = 1.6
            path.lineCapStyle = .round
            NSColor.white.setStroke()
            path.lineWidth = 3.2
            path.stroke()
            NSColor.black.setStroke()
            path.lineWidth = 1.6
            path.stroke()
            drawArrowHead(at: end, from: start)
            drawArrowHead(at: start, from: end)
            return true
        }
        return NSCursor(image: image, hotSpot: NSPoint(x: size / 2, y: size / 2))
    }

    private static func drawArrowHead(at tip: CGPoint, from tail: CGPoint) {
        let dx = tip.x - tail.x
        let dy = tip.y - tail.y
        let length = hypot(dx, dy)
        guard length > 0 else { return }
        let ux = dx / length
        let uy = dy / length
        let head: CGFloat = 4.5
        let wing: CGFloat = 2.4
        let base = CGPoint(x: tip.x - ux * head, y: tip.y - uy * head)
        let left = CGPoint(x: base.x - uy * wing, y: base.y + ux * wing)
        let right = CGPoint(x: base.x + uy * wing, y: base.y - ux * wing)
        let path = NSBezierPath()
        path.move(to: tip)
        path.line(to: left)
        path.line(to: right)
        path.close()
        NSColor.white.setStroke()
        path.lineWidth = 1.4
        path.stroke()
        NSColor.black.setFill()
        path.fill()
    }
}

/// Host-facing overlay bindings (spatial kinds only).
struct ArtifactRegionOverlayInput: Equatable {
    var armedTool: ArtifactRegionTool?
    var committed: ArtifactRegionDraft?

    static func == (lhs: ArtifactRegionOverlayInput, rhs: ArtifactRegionOverlayInput) -> Bool {
        lhs.armedTool == rhs.armedTool && lhs.committed == rhs.committed
    }

    /// Paint the committed region only on the locator page. Images have no page
    /// layer, so they always show. Locator JSON is unchanged.
    static func committedOnCurrentPage(
        _ committed: ArtifactRegionDraft?,
        locatorPage: Int?,
        viewerPage: Int,
        supportsPageLocator: Bool
    ) -> ArtifactRegionDraft? {
        guard let committed else { return nil }
        if supportsPageLocator, locatorPage != viewerPage {
            return nil
        }
        return committed
    }
}
