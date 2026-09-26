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

    var onCommit: ((ArtifactRegionDraft) -> Void)?
    var onDisarm: (() -> Void)?

    private var dragStart: CGPoint?
    private var dragCurrent: CGPoint?
    private var inProgressPoints: [CGPoint] = []
    private var activeHandle: ArtifactRegionHandle?
    private var isEditingCommitted = false
    private var moveDragOrigin: CGPoint?
    private var moveDragPoints: [CGPoint]?
    private var tracking: NSTrackingArea?

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
        let pad: CGFloat = 8
        if let anchor = ArtifactRegionGeometry.moveAnchor(for: committed) {
            let point = ArtifactRegionGeometry.documentPoint(normalized: anchor, imageRect: imageRect)
            addCursorRect(cursorRect(around: point, pad: pad), cursor: .openHand)
        }
        let centroid = ArtifactRegionGeometry.moveAnchor(for: committed) ?? CGPoint(x: 0.5, y: 0.5)
        for handle in ArtifactRegionGeometry.handlePoints(for: committed, imageRect: imageRect) {
            let point = ArtifactRegionGeometry.documentPoint(normalized: handle, imageRect: imageRect)
            let kind = ArtifactRegionGeometry.resizeCursorKind(handle: handle, centroid: centroid)
            addCursorRect(cursorRect(around: point, pad: pad), cursor: ArtifactRegionCursor.resize(kind))
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
        if hitHandle(at: point) != nil { return self }
        if committed != nil, isNearCommittedStroke(point) { return self }
        return nil
    }

    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        let media = imageRect
        guard media.width > 0, media.height > 0 else { return }

        if let committed, armedTool == nil || isEditingCommitted {
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

        if let handle = hitHandle(at: document) {
            activeHandle = handle
            isEditingCommitted = true
            if handle == .move, let committed {
                moveDragOrigin = normalized
                moveDragPoints = committed.points
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

        if let handle = activeHandle, var draft = committed {
            switch handle {
            case .move:
                if let origin = moveDragOrigin, let start = moveDragPoints {
                    draft.points = ArtifactRegionGeometry.translate(
                        start,
                        by: CGPoint(x: normalized.x - origin.x, y: normalized.y - origin.y)
                    )
                }
            case .resize(let index):
                draft.points = moveHandle(draft: draft, handle: index, to: normalized)
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

        if isEditingCommitted, let draft = committed, draft.isValid {
            onCommit?(draft)
        }
        activeHandle = nil
        isEditingCommitted = false
        moveDragOrigin = nil
        moveDragPoints = nil
    }

    override func rightMouseDown(with event: NSEvent) {
        guard committed?.kind == .freeform else {
            super.rightMouseDown(with: event)
            return
        }
        let document = convert(event.locationInWindow, from: nil)
        guard let index = freeformVertexIndex(at: document),
              let points = ArtifactRegionGeometry.deleteFreeformVertex(
                points: committed?.points ?? [],
                index: index
              )
        else { return }
        var next = committed
        next?.points = points
        committed = next
        if let next, next.isValid {
            onCommit?(next)
        }
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
        inProgressPoints.append(normalized)
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

    private func commitAndDisarm(_ draft: ArtifactRegionDraft) {
        committed = draft
        dragStart = nil
        dragCurrent = nil
        inProgressPoints = []
        armedTool = nil
        onCommit?(draft)
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
        let threshold = handleHitSize
        if let anchor = ArtifactRegionGeometry.moveAnchor(for: committed) {
            let point = ArtifactRegionGeometry.documentPoint(normalized: anchor, imageRect: imageRect)
            if hypot(point.x - documentPoint.x, point.y - documentPoint.y) <= threshold + 1 {
                return .move
            }
        }
        let handles = ArtifactRegionGeometry.handlePoints(for: committed, imageRect: imageRect)
        for (index, normalized) in handles.enumerated() {
            let point = ArtifactRegionGeometry.documentPoint(normalized: normalized, imageRect: imageRect)
            if hypot(point.x - documentPoint.x, point.y - documentPoint.y) <= threshold {
                return .resize(index)
            }
        }
        return nil
    }

    private func freeformVertexIndex(at documentPoint: CGPoint) -> Int? {
        guard let committed, committed.kind == .freeform else { return nil }
        let threshold = handleHitSize
        for (index, normalized) in committed.points.enumerated() {
            let point = ArtifactRegionGeometry.documentPoint(normalized: normalized, imageRect: imageRect)
            if hypot(point.x - documentPoint.x, point.y - documentPoint.y) <= threshold {
                return index
            }
        }
        return nil
    }

    private func isNearCommittedStroke(_ documentPoint: CGPoint) -> Bool {
        guard let committed else { return false }
        let media = imageRect
        let docs = ArtifactRegionGeometry.documentPoints(normalized: committed.points, imageRect: media)
        guard docs.count >= 2 else { return false }
        let threshold = handleHitSize
        for i in 0..<docs.count {
            let a = docs[i]
            let b = docs[(i + 1) % docs.count]
            if distance(documentPoint, segmentFrom: a, to: b) <= threshold {
                return true
            }
        }
        return false
    }

    private var handleHitSize: CGFloat {
        7
    }

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

        for handle in ArtifactRegionGeometry.handlePoints(for: draft, imageRect: media) {
            drawHandle(at: ArtifactRegionGeometry.documentPoint(normalized: handle, imageRect: media))
        }
        if let anchor = ArtifactRegionGeometry.moveAnchor(for: draft) {
            drawMoveAnchor(
                at: ArtifactRegionGeometry.documentPoint(normalized: anchor, imageRect: media)
            )
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
        let size: CGFloat = 7
        let rect = CGRect(x: point.x - size / 2, y: point.y - size / 2, width: size, height: size)
        let path = NSBezierPath(ovalIn: rect)
        NSColor(PVColor.surfaceRaised).setFill()
        NSColor(PVColor.accent).setStroke()
        path.fill()
        path.lineWidth = 1.5
        path.stroke()
    }

    private func drawMoveAnchor(at point: CGPoint) {
        let size: CGFloat = 9
        let path = NSBezierPath()
        path.move(to: CGPoint(x: point.x, y: point.y - size / 2))
        path.line(to: CGPoint(x: point.x + size / 2, y: point.y))
        path.line(to: CGPoint(x: point.x, y: point.y + size / 2))
        path.line(to: CGPoint(x: point.x - size / 2, y: point.y))
        path.close()
        NSColor(PVColor.accent).setFill()
        NSColor(PVColor.surfaceRaised).setStroke()
        path.fill()
        path.lineWidth = 1.25
        path.stroke()
    }

    private func applyCursor(at documentPoint: CGPoint) {
        if armedTool != nil {
            NSCursor.crosshair.set()
            return
        }
        guard let committed else { return }
        switch hitHandle(at: documentPoint) {
        case .move:
            NSCursor.openHand.set()
        case .resize:
            let centroid = ArtifactRegionGeometry.moveAnchor(for: committed) ?? CGPoint(x: 0.5, y: 0.5)
            let handle = ArtifactRegionGeometry.normalize(documentPoint: documentPoint, imageRect: imageRect)
            ArtifactRegionCursor.resize(
                ArtifactRegionGeometry.resizeCursorKind(handle: handle, centroid: centroid)
            ).set()
        case nil:
            break
        }
    }

    private func cursorRect(around point: CGPoint, pad: CGFloat) -> CGRect {
        CGRect(x: point.x - pad, y: point.y - pad, width: pad * 2, height: pad * 2)
    }

    private func distance(_ point: CGPoint, segmentFrom a: CGPoint, to b: CGPoint) -> CGFloat {
        let dx = b.x - a.x
        let dy = b.y - a.y
        let length2 = dx * dx + dy * dy
        guard length2 > 0 else { return hypot(point.x - a.x, point.y - a.y) }
        let t = max(0, min(1, ((point.x - a.x) * dx + (point.y - a.y) * dy) / length2))
        let proj = CGPoint(x: a.x + t * dx, y: a.y + t * dy)
        return hypot(point.x - proj.x, point.y - proj.y)
    }
}

enum ArtifactRegionHandle: Equatable {
    case resize(Int)
    case move
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
}
