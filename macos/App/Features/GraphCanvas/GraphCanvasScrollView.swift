@preconcurrency import AppKit
import Observation
import SwiftUI

/// AppKit pan/zoom shell for any graph-style canvas (design note §7.2).
///
/// Product-agnostic tooling: SwiftUI `ScrollView` has no magnification, so
/// hosts compose this bridge with their own document content.
///
/// **Pointer ownership**
/// ``GraphCanvasDocumentView`` owns all canvas mouse sequences (select, drag,
/// place, empty-canvas pan, connect). Hosted SwiftUI is paint-only — update its
/// `rootView` freely; there is no SwiftUI `DragGesture` to cancel.
///
/// **Coordinate seams**
/// - Grid space: what hosts speak — item origins, hit targets, gesture points,
///   base-layer drawing input. Two hosts agree on a grid point whatever the
///   document's size or position.
/// - Document space: the AppKit document view's own coordinates. The document's
///   top-left stands on grid point ``contentOrigin``; the document view converts
///   each mouse event once, and places each item at `origin − contentOrigin`.
/// - Live pointer: AppKit `convert` on the document view
/// - Pure layout / unit tests: ``GraphCanvasCoordinates``
///
/// **Input mapping**
/// - Trackpad (precise scroll): two-finger pan (system default)
/// - Trackpad pinch: zoom (system magnification)
/// - Mouse wheel (line scroll): zoom toward the cursor
/// - Click-drag on empty document: pan (via ``GraphCanvasPointerController``)
struct GraphCanvasScrollView<Content: View>: NSViewRepresentable {
    var contentSize: CGSize
    /// Grid-space point the document's top-left stands on. Moving it shifts the
    /// scroll offset by the same delta so the viewport stays on the same content.
    var contentOrigin: CGPoint = .zero
    /// Stable identity for the document (e.g. source id). Changing it restores
    /// `restoredCamera`, or centers at 1× when that is nil. Keeping it stable
    /// preserves zoom, including when the document grows.
    var contentID: AnyHashable
    /// Applied when `contentID` changes. Nil opens centered at 1×.
    /// A content-size change does not apply this and does not recenter.
    var restoredCamera: GraphCanvasCamera? = nil
    /// Live pan and zoom. Hosts persist this; the scroll view does not.
    var onCameraChange: (GraphCanvasCamera) -> Void = { _ in }
    /// VoiceOver name for the canvas group that holds the items and their rotors.
    var accessibilityLabel: String = ""
    @ViewBuilder var content: () -> Content

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeNSView(context: Context) -> GraphCanvasNSScrollView {
        let scrollView = GraphCanvasNSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.allowsMagnification = true
        scrollView.minMagnification = GraphCanvasCamera.minMagnification
        scrollView.maxMagnification = GraphCanvasCamera.maxMagnification
        scrollView.magnification = 1
        scrollView.drawsBackground = false
        scrollView.borderType = .noBorder
        scrollView.onCameraChange = { [weak coordinator = context.coordinator] camera in
            coordinator?.onCameraChange(camera)
        }

        let coordinator = context.coordinator
        coordinator.onCameraChange = onCameraChange
        coordinator.viewport.scrollView = scrollView
        scrollView.viewport = coordinator.viewport
        coordinator.pointer.viewport = coordinator.viewport

        let document = GraphCanvasDocumentView(
            pointer: coordinator.pointer,
            frame: CGRect(origin: .zero, size: contentSize)
        )
        document.contentOrigin = contentOrigin
        document.setAccessibilityLabel(accessibilityLabel)
        document.installHostingRoot(hostedRoot(context: context, items: document.items, origin: contentOrigin))
        scrollView.documentView = document
        scrollView.items = document.items
        coordinator.documentView = document
        coordinator.installedContentID = contentID
        coordinator.installedContentSize = contentSize

        let restored = restoredCamera
        DispatchQueue.main.async {
            scrollView.applyOpeningCamera(restored, contentSize: contentSize)
        }
        return scrollView
    }

    func updateNSView(_ scrollView: GraphCanvasNSScrollView, context: Context) {
        let coordinator = context.coordinator
        coordinator.onCameraChange = onCameraChange
        scrollView.onCameraChange = { [weak coordinator] camera in
            coordinator?.onCameraChange(camera)
        }
        coordinator.viewport.scrollView = scrollView
        scrollView.viewport = coordinator.viewport
        coordinator.pointer.viewport = coordinator.viewport

        let idChanged = coordinator.installedContentID != contentID
        let sizeChanged = coordinator.installedContentSize != contentSize
        let previousOrigin = coordinator.documentView?.contentOrigin ?? contentOrigin
        let originDelta = CGSize(
            width: previousOrigin.x - contentOrigin.x,
            height: previousOrigin.y - contentOrigin.y
        )
        coordinator.installedContentID = contentID
        coordinator.installedContentSize = contentSize

        // The origin is set before the root, so the root and the items are
        // built against the new one. Paint updates every representable pass —
        // safe now that AppKit owns gestures.
        if let document = coordinator.documentView {
            document.contentOrigin = contentOrigin
            document.setAccessibilityLabel(accessibilityLabel)
            document.installHostingRoot(hostedRoot(context: context, items: document.items, origin: contentOrigin))
        }

        if sizeChanged {
            coordinator.documentView?.resizeDocument(to: contentSize)
        }
        // Identity change restores that document's camera. Growing the
        // document must not recenter or reset zoom. Moving the origin keeps
        // the same grid point in view.
        if originDelta != .zero, !idChanged {
            Self.shiftVisibleOrigin(in: scrollView, by: originDelta, contentSize: contentSize)
        }
        if idChanged {
            scrollView.applyOpeningCamera(restoredCamera, contentSize: contentSize)
        }
    }

    private func hostedRoot(context: Context, items: GraphCanvasItemLayer, origin: CGPoint) -> AnyView {
        AnyView(
            content()
                .environment(\.graphCanvasContentOrigin, origin)
                .environment(\.graphCanvasViewport, context.coordinator.viewport)
                .environment(\.graphCanvasPointer, context.coordinator.pointer)
                .environment(\.graphCanvasItems, items)
                .allowsHitTesting(false)
        )
    }

    /// Reads the live camera from an `NSScrollView` for conversion / future tools.
    static func camera(from scrollView: NSScrollView) -> GraphCanvasCamera {
        let origin = scrollView.documentVisibleRect.origin
        return GraphCanvasCamera(
            magnification: scrollView.magnification,
            contentOffset: origin
        )
    }

    /// Keeps the same grid point in view when the document's origin moves.
    private static func shiftVisibleOrigin(
        in scrollView: NSScrollView,
        by delta: CGSize,
        contentSize: CGSize
    ) {
        let clip = scrollView.contentView
        let origin = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: CGPoint(
                x: clip.bounds.origin.x + delta.width,
                y: clip.bounds.origin.y + delta.height
            ),
            documentSize: contentSize,
            visibleSize: clip.bounds.size
        )
        clip.scroll(to: origin)
        scrollView.reflectScrolledClipView(clip)
    }

    final class Coordinator {
        var documentView: GraphCanvasDocumentView?
        var installedContentID: AnyHashable?
        var installedContentSize: CGSize?
        var onCameraChange: (GraphCanvasCamera) -> Void = { _ in }
        let viewport: GraphCanvasViewportController
        let pointer: GraphCanvasPointerController

        @MainActor
        init() {
            viewport = GraphCanvasViewportController()
            pointer = GraphCanvasPointerController()
            pointer.viewport = viewport
        }
    }
}

// MARK: - Mouse wheel zoom

/// `NSScrollView` that zooms on discrete mouse-wheel ticks while leaving
/// trackpad two-finger scrolling as pan.
final class GraphCanvasNSScrollView: NSScrollView {
    weak var viewport: GraphCanvasViewportController?
    var onCameraChange: ((GraphCanvasCamera) -> Void)?
    private var suppressCameraReport = false
    private var isUpdatingMagnificationFloor = false
    /// The item layer whose render scale follows the settled zoom.
    weak var items: GraphCanvasItemLayer?
    /// How long the zoom must hold still before the items re-render at it.
    static let itemScaleSettleDelay: TimeInterval = 0.15
    private var itemScaleSettle: DispatchWorkItem?
    nonisolated(unsafe) private var liveMagnifyEndObserver: Any?

    deinit {
        if let liveMagnifyEndObserver { NotificationCenter.default.removeObserver(liveMagnifyEndObserver) }
    }

    override var magnification: CGFloat {
        didSet {
            guard abs(oldValue - magnification) > 0.000_1 else { return }
            scheduleItemContentsScale()
            let value = magnification
            let viewport = viewport
            MainActor.assumeIsolated {
                viewport?.magnification = value
            }
            reportCamera()
        }
    }

    /// Restores `camera`, or centers at 1× when it is nil. Both stay at or above
    /// the magnification that still covers the window.
    func applyOpeningCamera(_ camera: GraphCanvasCamera?, contentSize: CGSize) {
        suppressCameraReport = true
        defer {
            suppressCameraReport = false
            reportCamera()
        }
        let floor = GraphCanvasCoordinates.magnificationThatFills(
            documentSize: contentSize,
            viewportSize: contentView.frame.size
        )
        minMagnification = floor
        if let camera {
            magnification = GraphCanvasCoordinates.clampMagnification(
                camera.magnification,
                minimum: floor
            )
            let visible = contentView.bounds.size
            let origin = GraphCanvasCoordinates.clampedContentOrigin(
                proposed: camera.contentOffset,
                documentSize: contentSize,
                visibleSize: visible
            )
            contentView.scroll(to: origin)
            reflectScrolledClipView(contentView)
        } else {
            magnification = GraphCanvasCoordinates.clampMagnification(1, minimum: floor)
            let visible = contentView.bounds.size
            let origin = GraphCanvasCoordinates.clampedContentOrigin(
                proposed: CGPoint(
                    x: (contentSize.width - visible.width) / 2,
                    y: (contentSize.height - visible.height) / 2
                ),
                documentSize: contentSize,
                visibleSize: visible
            )
            contentView.scroll(to: origin)
            reflectScrolledClipView(contentView)
        }
    }

    override func reflectScrolledClipView(_ clipView: NSClipView) {
        super.reflectScrolledClipView(clipView)
        reportCamera()
    }

    override func layout() {
        super.layout()
        updateMagnificationFloor()
    }

    /// Raises the zoom-out floor so the document keeps covering the clip.
    /// A larger document lowers the floor again without changing the current zoom.
    private func updateMagnificationFloor() {
        guard !isUpdatingMagnificationFloor, let document = documentView else { return }
        isUpdatingMagnificationFloor = true
        defer { isUpdatingMagnificationFloor = false }

        let floor = GraphCanvasCoordinates.magnificationThatFills(
            documentSize: document.frame.size,
            viewportSize: contentView.frame.size
        )
        if abs(minMagnification - floor) > 0.000_1 {
            minMagnification = floor
        }
        guard magnification + 0.000_1 < floor else { return }
        let clip = contentView
        let center = clip.convert(CGPoint(x: clip.bounds.midX, y: clip.bounds.midY), to: nil)
        zoom(to: floor, windowLocation: center)
    }

    private func reportCamera() {
        guard !suppressCameraReport, let onCameraChange else { return }
        onCameraChange(
            GraphCanvasCamera(
                magnification: magnification,
                contentOffset: documentVisibleRect.origin
            )
        )
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        viewport?.observeWindow(window)
        if liveMagnifyEndObserver == nil {
            liveMagnifyEndObserver = NotificationCenter.default.addObserver(
                forName: NSScrollView.didEndLiveMagnifyNotification,
                object: self,
                queue: .main
            ) { [weak self] _ in
                MainActor.assumeIsolated { self?.scheduleItemContentsScale() }
            }
        }
        applyItemContentsScale()
        scheduleItemContentsScale()
    }

    override func viewDidChangeBackingProperties() {
        super.viewDidChangeBackingProperties()
        applyItemContentsScale()
        scheduleItemContentsScale()
    }

    /// Applies the items' render scale once the zoom has held still for
    /// ``itemScaleSettleDelay``. A pinch, a wheel step and the spring-back all
    /// change `magnification`, and the end of a pinch is not a reliable event,
    /// so each change restarts a timer instead of waiting for an end. A live
    /// pinch does not call the `magnification` setter, so its end comes from
    /// `NSScrollView.didEndLiveMagnifyNotification`; wheel steps and the
    /// spring-back come through the setter. Every
    /// zoom source and a backing change share this one settle path, and it
    /// runs after AppKit's own backing-scale update.
    private func scheduleItemContentsScale() {
        itemScaleSettle?.cancel()
        let work = DispatchWorkItem { [weak self] in
            MainActor.assumeIsolated {
                self?.applyItemContentsScale()
            }
        }
        itemScaleSettle = work
        DispatchQueue.main.asyncAfter(deadline: .now() + Self.itemScaleSettleDelay, execute: work)
    }

    /// Sets the items' render scale at once. A window or backing-scale change
    /// calls this and then ``scheduleItemContentsScale()``, so a layer AppKit
    /// resets after this call is put back when the timer fires.
    private func applyItemContentsScale() {
        guard let window else { return }
        items?.contentsScale = GraphCanvasCoordinates.itemContentsScale(
            backingScale: window.backingScaleFactor,
            magnification: magnification
        )
    }

    override func scrollWheel(with event: NSEvent) {
        if Self.isTrackpadOrMagicMouseScroll(event) {
            super.scrollWheel(with: event)
            return
        }
        guard allowsMagnification, documentView != nil else {
            super.scrollWheel(with: event)
            return
        }
        let sensitivity: CGFloat = 0.08
        let factor = exp(event.scrollingDeltaY * sensitivity)
        let target = GraphCanvasCoordinates.clampMagnification(
            magnification * factor,
            minimum: minMagnification
        )
        guard abs(target - magnification) > 0.000_1 else { return }
        zoom(to: target, windowLocation: event.locationInWindow)
    }

    /// Zooms so the document point under `windowLocation` stays under the cursor.
    func zoom(to newMagnification: CGFloat, windowLocation: CGPoint) {
        guard documentView != nil else {
            magnification = newMagnification
            return
        }
        let oldMag = magnification
        let newMag = GraphCanvasCoordinates.clampMagnification(
            newMagnification,
            minimum: minMagnification
        )
        guard abs(newMag - oldMag) > 0.000_1 else { return }

        let clip = contentView
        let anchor = clip.convert(windowLocation, from: nil)

        magnification = newMag

        let after = clip.convert(windowLocation, from: nil)
        var origin = clip.bounds.origin
        origin.x -= after.x - anchor.x
        origin.y -= after.y - anchor.y

        let visible = clip.bounds.size
        let docSize = documentView?.frame.size ?? .zero
        origin = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: origin,
            documentSize: docSize,
            visibleSize: visible
        )

        clip.scroll(to: origin)
        reflectScrolledClipView(clip)
    }

    static func isTrackpadOrMagicMouseScroll(_ event: NSEvent) -> Bool {
        GraphCanvasScrollInput.isTrackpadStyleScroll(
            hasPreciseScrollingDeltas: event.hasPreciseScrollingDeltas,
            hasPhase: event.phase != [],
            hasMomentumPhase: event.momentumPhase != []
        )
    }
}

// MARK: - Viewport bridge

/// Bridges pan deltas into the hosting `NSScrollView`.
@MainActor
@Observable
final class GraphCanvasViewportController: NSObject {
    weak var scrollView: NSScrollView?
    /// Live `NSScrollView` magnification. Grid fade and screen-space strokes read this.
    var magnification: CGFloat = 1
    /// True while the window first responder is inside this canvas document.
    private(set) var documentContainsKeyboardFocus = false
    private var observedWindow: NSWindow?

    /// Pans by a delta measured in **window / view** points (Y down).
    /// Divides by magnification so one screen pixel of drag moves one screen
    /// pixel of content.
    func panByViewDelta(_ delta: CGSize) {
        guard let scrollView, let document = scrollView.documentView else { return }
        let mag = max(scrollView.magnification, 0.000_1)
        let clip = scrollView.contentView
        var origin = clip.bounds.origin
        origin.x -= delta.width / mag
        origin.y -= delta.height / mag

        origin = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: origin,
            documentSize: document.frame.size,
            visibleSize: clip.bounds.size
        )

        clip.scroll(to: origin)
        scrollView.reflectScrolledClipView(clip)
    }

    /// Document-space point under the cursor via AppKit conversion.
    func documentPointUnderCursor() -> CGPoint? {
        guard let scrollView,
              let window = scrollView.window,
              let document = scrollView.documentView
        else { return nil }
        let windowPoint = window.mouseLocationOutsideOfEventStream
        return document.convert(windowPoint, from: nil)
    }

    func observeWindow(_ window: NSWindow?) {
        if let observedWindow {
            observedWindow.removeObserver(self, forKeyPath: "firstResponder")
            self.observedWindow = nil
        }
        refreshDocumentKeyboardFocus()
        guard let window else { return }
        window.addObserver(self, forKeyPath: "firstResponder", options: [.new], context: nil)
        observedWindow = window
    }

    nonisolated override func observeValue(
        forKeyPath keyPath: String?,
        of object: Any?,
        change: [NSKeyValueChangeKey: Any]?,
        context: UnsafeMutableRawPointer?
    ) {
        guard keyPath == "firstResponder" else {
            super.observeValue(forKeyPath: keyPath, of: object, change: change, context: context)
            return
        }
        Task { @MainActor in
            self.refreshDocumentKeyboardFocus()
        }
    }

    private func refreshDocumentKeyboardFocus() {
        guard let scrollView,
              let window = scrollView.window,
              let document = scrollView.documentView
        else {
            documentContainsKeyboardFocus = false
            return
        }
        let responder = window.firstResponder as? NSView
        let inside = responder === document || (responder?.isDescendant(of: document) ?? false)
        if documentContainsKeyboardFocus != inside {
            documentContainsKeyboardFocus = inside
        }
    }
}

private struct GraphCanvasViewportKey: EnvironmentKey {
    static let defaultValue: GraphCanvasViewportController? = nil
}

private struct GraphCanvasPointerKey: EnvironmentKey {
    static let defaultValue: GraphCanvasPointerController? = nil
}

private struct GraphCanvasContentOriginKey: EnvironmentKey {
    static let defaultValue: CGPoint = .zero
}

private struct GraphCanvasItemsKey: EnvironmentKey {
    static let defaultValue: GraphCanvasItemLayer? = nil
}

extension EnvironmentValues {
    var graphCanvasViewport: GraphCanvasViewportController? {
        get { self[GraphCanvasViewportKey.self] }
        set { self[GraphCanvasViewportKey.self] = newValue }
    }

    var graphCanvasPointer: GraphCanvasPointerController? {
        get { self[GraphCanvasPointerKey.self] }
        set { self[GraphCanvasPointerKey.self] = newValue }
    }

    /// The grid point the canvas document's top-left stands on. The host's base
    /// view draws grid-space input at `grid − origin`
    /// (``GraphCanvasCoordinates/documentRect(_:origin:)``).
    var graphCanvasContentOrigin: CGPoint {
        get { self[GraphCanvasContentOriginKey.self] }
        set { self[GraphCanvasContentOriginKey.self] = newValue }
    }

    /// The canvas item layer, so the host's base view can push its items.
    var graphCanvasItems: GraphCanvasItemLayer? {
        get { self[GraphCanvasItemsKey.self] }
        set { self[GraphCanvasItemsKey.self] = newValue }
    }
}
