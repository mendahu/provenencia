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
            if oldValue != imageSize { needsDisplay = true }
        }
    }

    var onCommit: ((ArtifactRegionDraft) -> Void)?
    var onDisarm: (() -> Void)?

    private var dragStart: CGPoint?
    private var dragCurrent: CGPoint?
    private var inProgressPoints: [CGPoint] = []
    private var activeHandle: Int?
    private var isEditingCommitted = false

    override var isFlipped: Bool { true }
    override var acceptsFirstResponder: Bool { armedTool != nil || committed != nil }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        if armedTool != nil {
            window?.makeFirstResponder(self)
        }
    }

    override func resetCursorRects() {
        if armedTool != nil {
            addCursorRect(bounds, cursor: .crosshair)
        } else {
            addCursorRect(bounds, cursor: .arrow)
        }
    }

    override func hitTest(_ point: NSPoint) -> NSView? {
        if armedTool != nil { return self }
        if handleIndex(at: point) != nil { return self }
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

        if let handle = handleIndex(at: document) {
            activeHandle = handle
            isEditingCommitted = true
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
            draft.points = moveHandle(draft: draft, handle: handle, to: normalized)
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
        ArtifactRegionGeometry.imageRect(documentSize: bounds.size, imageSize: imageSize)
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
                points = ArtifactRegionGeometry.circleRing(from: start, to: current)
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
            return ArtifactRegionGeometry.moveCircleRadius(points: draft.points, to: location)
        case .freeform:
            return ArtifactRegionGeometry.moveFreeformVertex(
                points: draft.points,
                index: handle,
                to: location
            )
        }
    }

    private func handleIndex(at documentPoint: CGPoint) -> Int? {
        guard let committed else { return nil }
        let handles = ArtifactRegionGeometry.handlePoints(for: committed)
        let threshold = handleHitSize
        for (index, normalized) in handles.enumerated() {
            let point = ArtifactRegionGeometry.documentPoint(normalized: normalized, imageRect: imageRect)
            if hypot(point.x - documentPoint.x, point.y - documentPoint.y) <= threshold {
                return index
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
        let docs = ArtifactRegionGeometry.documentPoints(normalized: draft.points, imageRect: media)
        guard docs.count >= 3, let path = polygonPath(docs) else { return }

        let veil = NSBezierPath(rect: media)
        veil.append(path)
        veil.windingRule = .evenOdd
        NSColor(PVColor.surfaceOverlay).setFill()
        veil.fill()

        NSColor(PVColor.accent).setStroke()
        path.lineWidth = 2
        path.lineJoinStyle = .round
        path.stroke()

        for handle in ArtifactRegionGeometry.handlePoints(for: draft) {
            drawHandle(at: ArtifactRegionGeometry.documentPoint(normalized: handle, imageRect: media))
        }
    }

    private func drawInProgress(_ draft: ArtifactRegionDraft, imageRect media: CGRect) {
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

/// Host-facing overlay bindings (spatial kinds only).
struct ArtifactRegionOverlayInput: Equatable {
    var armedTool: ArtifactRegionTool?
    var committed: ArtifactRegionDraft?

    static func == (lhs: ArtifactRegionOverlayInput, rhs: ArtifactRegionOverlayInput) -> Bool {
        lhs.armedTool == rhs.armedTool && lhs.committed == rhs.committed
    }
}
