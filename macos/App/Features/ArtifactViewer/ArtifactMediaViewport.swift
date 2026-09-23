@preconcurrency import AppKit
import SwiftUI

/// Spatial pan/zoom shell for image / PDF-page rasters (document range 0.5…4).
///
/// - **Pan:** click-drag when content is larger than the clip (also trackpad
///   two-finger scroll). Mouse wheel zooms toward the cursor.
/// - **Center:** when the image is smaller than the visible area (at the
///   current magnification), it is centered in the clip.
///
/// Not used for future audio/video players — those get their own chrome inside
/// ``ArtifactViewer``.
struct ArtifactMediaViewport: NSViewRepresentable {
    var image: NSImage?
    var zoom: CGFloat
    var contentID: AnyHashable
    var onZoomChange: (CGFloat) -> Void

    func makeCoordinator() -> Coordinator {
        Coordinator(onZoomChange: onZoomChange)
    }

    func makeNSView(context: Context) -> ArtifactMediaNSScrollView {
        let scrollView = ArtifactMediaNSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.allowsMagnification = true
        scrollView.minMagnification = ArtifactViewerModel.minZoom
        scrollView.maxMagnification = ArtifactViewerModel.maxZoom
        scrollView.magnification = ArtifactViewerModel.clampZoom(zoom)
        scrollView.drawsBackground = false
        scrollView.borderType = .noBorder
        scrollView.coordinator = context.coordinator
        scrollView.backgroundColor = NSColor(PVColor.surfaceSunken)

        let document = ArtifactMediaDocumentView(frame: .zero)
        document.scrollView = scrollView
        scrollView.documentView = document
        context.coordinator.documentView = document
        context.coordinator.scrollView = scrollView
        context.coordinator.installedContentID = contentID
        context.coordinator.installedImage = image

        document.image = image
        scrollView.layoutDocumentCentered(resetScroll: true)
        scrollView.startObservingClipBounds()

        return scrollView
    }

    func updateNSView(_ scrollView: ArtifactMediaNSScrollView, context: Context) {
        let coordinator = context.coordinator
        scrollView.coordinator = coordinator
        coordinator.onZoomChange = onZoomChange
        coordinator.scrollView = scrollView
        coordinator.documentView?.scrollView = scrollView

        let idChanged = coordinator.installedContentID != contentID
        let imageChanged = !imagesIdentical(coordinator.installedImage, image)
        coordinator.installedContentID = contentID
        coordinator.installedImage = image

        if imageChanged {
            coordinator.documentView?.image = image
        }

        let target = ArtifactViewerModel.clampZoom(zoom)
        if abs(scrollView.magnification - target) > 0.001 {
            coordinator.suppressZoomCallback = true
            scrollView.magnification = target
            coordinator.suppressZoomCallback = false
        }

        if idChanged || imageChanged {
            scrollView.layoutDocumentCentered(resetScroll: true)
        } else {
            // Clip resize is handled via bounds-size notifications; avoid
            // re-entering layout on every SwiftUI refresh (e.g. zoom % label).
            scrollView.layoutDocumentCentered(resetScroll: false, allowScrollAdjust: false)
        }
    }

    static func dismantleNSView(_ scrollView: ArtifactMediaNSScrollView, coordinator: Coordinator) {
        scrollView.stopObservingClipBounds()
    }

    private func imagesIdentical(_ a: NSImage?, _ b: NSImage?) -> Bool {
        switch (a, b) {
        case (nil, nil): return true
        case (nil, _), (_, nil): return false
        case (let x?, let y?): return x === y
        }
    }

    @MainActor
    final class Coordinator: NSObject {
        var documentView: ArtifactMediaDocumentView?
        weak var scrollView: ArtifactMediaNSScrollView?
        var installedContentID: AnyHashable?
        var installedImage: NSImage?
        var onZoomChange: (CGFloat) -> Void
        var suppressZoomCallback = false

        init(onZoomChange: @escaping (CGFloat) -> Void) {
            self.onZoomChange = onZoomChange
        }
    }
}

// MARK: - Document (image + centering pad)

/// Document whose frame is at least the visible clip (in document space) so a
/// small image can sit centered; grows with the image when content overflows.
final class ArtifactMediaDocumentView: NSView {
    weak var scrollView: ArtifactMediaNSScrollView?

    var image: NSImage? {
        didSet { needsDisplay = true }
    }

    override var isFlipped: Bool { true }

    override func draw(_ dirtyRect: NSRect) {
        super.draw(dirtyRect)
        guard let image else { return }
        let size = image.size
        let origin = CGPoint(
            x: (bounds.width - size.width) / 2,
            y: (bounds.height - size.height) / 2
        )
        image.draw(
            in: CGRect(origin: origin, size: size),
            from: .zero,
            operation: .sourceOver,
            fraction: 1,
            respectFlipped: true,
            hints: [.interpolation: NSImageInterpolation.high]
        )
    }

    override func resetCursorRects() {
        if scrollView?.canScroll == true {
            addCursorRect(bounds, cursor: .openHand)
        } else {
            addCursorRect(bounds, cursor: .arrow)
        }
    }

    override func mouseDown(with event: NSEvent) {
        scrollView?.beginDragPan(with: event) ?? super.mouseDown(with: event)
    }

    override func mouseDragged(with event: NSEvent) {
        scrollView?.continueDragPan(with: event) ?? super.mouseDragged(with: event)
    }

    override func mouseUp(with event: NSEvent) {
        scrollView?.endDragPan(with: event) ?? super.mouseUp(with: event)
    }
}

// MARK: - Scroll view (wheel zoom + drag pan)

final class ArtifactMediaNSScrollView: NSScrollView {
    weak var coordinator: ArtifactMediaViewport.Coordinator?

    private(set) var isDragPanning = false
    private var dragStartViewPoint: CGPoint = .zero
    private var dragStartOrigin: CGPoint = .zero
    private var isLayingOut = false
    /// Clip bounds *size* only — origin changes while panning must not re-layout.
    private var lastClipSize: CGSize = .zero
    private var isObservingClipBounds = false

    func startObservingClipBounds() {
        guard !isObservingClipBounds else { return }
        contentView.postsBoundsChangedNotifications = true
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(clipBoundsDidChange(_:)),
            name: NSView.boundsDidChangeNotification,
            object: contentView
        )
        lastClipSize = contentView.bounds.size
        isObservingClipBounds = true
    }

    func stopObservingClipBounds() {
        guard isObservingClipBounds else { return }
        NotificationCenter.default.removeObserver(
            self,
            name: NSView.boundsDidChangeNotification,
            object: contentView
        )
        isObservingClipBounds = false
    }

    @objc private func clipBoundsDidChange(_ notification: Notification) {
        guard !isDragPanning else { return }
        let size = contentView.bounds.size
        guard abs(size.width - lastClipSize.width) > 0.5
            || abs(size.height - lastClipSize.height) > 0.5
        else { return }
        lastClipSize = size
        layoutDocumentCentered(resetScroll: false)
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        layoutDocumentCentered(resetScroll: true)
    }

    override func setFrameSize(_ newSize: NSSize) {
        let sizeChanged = abs(newSize.width - frame.width) > 0.5
            || abs(newSize.height - frame.height) > 0.5
        super.setFrameSize(newSize)
        // Only re-layout when *our* frame size changes — scrolling must not
        // re-enter layoutDocumentCentered (that was the pan crash).
        if sizeChanged {
            layoutDocumentCentered(resetScroll: false)
        }
    }

    /// Sizes the document to `max(image, visibleDoc)` and optionally centers / clamps.
    ///
    /// - Parameter allowScrollAdjust: When false, only resize the document pad
    ///   (used from SwiftUI `updateNSView` so zoom-label refreshes do not scroll).
    func layoutDocumentCentered(resetScroll: Bool = false, allowScrollAdjust: Bool = true) {
        guard !isLayingOut else { return }
        guard let document = documentView as? ArtifactMediaDocumentView else { return }
        isLayingOut = true
        defer { isLayingOut = false }

        // Under magnification, `contentView.bounds.size` is already the visible
        // rect in document coordinates — do not divide by magnification again.
        let visibleDoc = contentView.bounds.size
        let imageSize = document.image?.size ?? .zero
        let docSize = CGSize(
            width: max(imageSize.width, visibleDoc.width),
            height: max(imageSize.height, visibleDoc.height)
        )

        if abs(document.frame.width - docSize.width) > 0.5
            || abs(document.frame.height - docSize.height) > 0.5
        {
            document.setFrameSize(docSize)
            document.needsDisplay = true
        }

        lastClipSize = visibleDoc

        guard allowScrollAdjust else {
            window?.invalidateCursorRects(for: document)
            return
        }

        if resetScroll || !canScroll {
            centerContent()
        } else {
            clampScrollOrigin()
        }
        window?.invalidateCursorRects(for: document)
    }

    /// True when the document extends past the visible clip on either axis.
    var canScroll: Bool {
        guard let document = documentView else { return false }
        let clip = contentView.bounds.size
        let doc = document.frame.size
        return doc.width > clip.width + 0.5 || doc.height > clip.height + 0.5
    }

    func centerContent() {
        guard let document = documentView else { return }
        let clip = contentView.bounds.size
        let doc = document.frame.size
        let origin = CGPoint(
            x: max(0, (doc.width - clip.width) / 2),
            y: max(0, (doc.height - clip.height) / 2)
        )
        scrollToOriginIfNeeded(origin)
    }

    func clampScrollOrigin() {
        scrollToClamped(contentView.bounds.origin)
    }

    // MARK: Wheel zoom

    override func scrollWheel(with event: NSEvent) {
        if event.hasPreciseScrollingDeltas {
            super.scrollWheel(with: event)
            return
        }
        let factor: CGFloat = event.scrollingDeltaY > 0 ? 1.1 : (1 / 1.1)
        let target = ArtifactViewerModel.clampZoom(magnification * factor)
        guard abs(target - magnification) > 0.000_1 else { return }
        let point = contentView.convert(event.locationInWindow, from: nil)
        setMagnification(target, centeredAt: point)
        layoutDocumentCentered(resetScroll: false)
        reportZoomIfNeeded()
    }

    override var magnification: CGFloat {
        didSet {
            guard abs(oldValue - magnification) > 0.000_1 else { return }
            layoutDocumentCentered(resetScroll: false)
            reportZoomIfNeeded()
        }
    }

    private func reportZoomIfNeeded() {
        guard !(coordinator?.suppressZoomCallback ?? false) else { return }
        coordinator?.onZoomChange(ArtifactViewerModel.clampZoom(magnification))
    }

    // MARK: Click-drag pan

    func beginDragPan(with event: NSEvent) {
        guard canScroll, documentView != nil else { return }
        isDragPanning = true
        // View-space delta ÷ magnification → document-space pan (grab-hand).
        dragStartViewPoint = convert(event.locationInWindow, from: nil)
        dragStartOrigin = contentView.bounds.origin
        NSCursor.closedHand.push()
        window?.disableCursorRects()
    }

    func continueDragPan(with event: NSEvent) {
        guard isDragPanning else { return }
        let current = convert(event.locationInWindow, from: nil)
        let dx = current.x - dragStartViewPoint.x
        let dy = current.y - dragStartViewPoint.y
        let mag = max(magnification, 0.000_1)
        var origin = dragStartOrigin
        origin.x -= dx / mag
        origin.y -= dy / mag
        scrollToClamped(origin)
    }

    func endDragPan(with event: NSEvent) {
        guard isDragPanning else { return }
        isDragPanning = false
        NSCursor.pop()
        window?.enableCursorRects()
        if let document = documentView {
            window?.invalidateCursorRects(for: document)
        }
    }

    private func scrollToClamped(_ proposed: CGPoint) {
        guard let document = documentView else { return }
        let clip = contentView.bounds.size
        let doc = document.frame.size
        var origin = proposed
        let maxX = max(0, doc.width - clip.width)
        let maxY = max(0, doc.height - clip.height)
        origin.x = min(max(origin.x, 0), maxX)
        origin.y = min(max(origin.y, 0), maxY)
        scrollToOriginIfNeeded(origin)
    }

    private func scrollToOriginIfNeeded(_ origin: CGPoint) {
        let current = contentView.bounds.origin
        guard abs(current.x - origin.x) > 0.5 || abs(current.y - origin.y) > 0.5 else { return }
        contentView.scroll(to: origin)
        reflectScrolledClipView(contentView)
    }
}
