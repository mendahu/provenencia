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
    /// Stable identity for the document (e.g. source id). Changing it resets
    /// magnification; keeping it stable preserves zoom.
    var contentID: AnyHashable
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

        let coordinator = context.coordinator
        coordinator.viewport.scrollView = scrollView
        scrollView.viewport = coordinator.viewport
        coordinator.pointer.viewport = coordinator.viewport

        let document = GraphCanvasDocumentView(
            pointer: coordinator.pointer,
            frame: CGRect(origin: .zero, size: contentSize)
        )
        document.installHostingRoot(hostedRoot(context: context))
        scrollView.documentView = document
        coordinator.documentView = document
        coordinator.installedContentID = contentID
        coordinator.installedContentSize = contentSize

        DispatchQueue.main.async {
            Self.centerDocument(in: scrollView, contentSize: contentSize)
        }
        return scrollView
    }

    func updateNSView(_ scrollView: GraphCanvasNSScrollView, context: Context) {
        let coordinator = context.coordinator
        coordinator.viewport.scrollView = scrollView
        scrollView.viewport = coordinator.viewport
        coordinator.pointer.viewport = coordinator.viewport

        let idChanged = coordinator.installedContentID != contentID
        let sizeChanged = coordinator.installedContentSize != contentSize
        coordinator.installedContentID = contentID
        coordinator.installedContentSize = contentSize

        // Paint updates every representable pass — safe now that AppKit owns gestures.
        coordinator.documentView?.installHostingRoot(hostedRoot(context: context))

        if sizeChanged {
            coordinator.documentView?.resizeDocument(to: contentSize)
        }
        if idChanged {
            scrollView.magnification = 1
            DispatchQueue.main.async {
                Self.centerDocument(in: scrollView, contentSize: contentSize)
            }
        }
    }

    private func hostedRoot(context: Context) -> AnyView {
        AnyView(
            content()
                .environment(\.graphCanvasViewport, context.coordinator.viewport)
                .environment(\.graphCanvasPointer, context.coordinator.pointer)
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

    private static func centerDocument(in scrollView: NSScrollView, contentSize: CGSize) {
        let visible = scrollView.contentView.bounds.size
        let origin = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: CGPoint(
                x: (contentSize.width - visible.width) / 2,
                y: (contentSize.height - visible.height) / 2
            ),
            documentSize: contentSize,
            visibleSize: visible
        )
        scrollView.contentView.scroll(to: origin)
        scrollView.reflectScrolledClipView(scrollView.contentView)
    }

    final class Coordinator {
        var documentView: GraphCanvasDocumentView?
        var installedContentID: AnyHashable?
        var installedContentSize: CGSize?
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

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        viewport?.observeWindow(window)
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
        let target = GraphCanvasCoordinates.clampMagnification(magnification * factor)
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
        let newMag = GraphCanvasCoordinates.clampMagnification(newMagnification)
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
    func contentPointUnderCursor() -> CGPoint? {
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

extension EnvironmentValues {
    var graphCanvasViewport: GraphCanvasViewportController? {
        get { self[GraphCanvasViewportKey.self] }
        set { self[GraphCanvasViewportKey.self] = newValue }
    }

    var graphCanvasPointer: GraphCanvasPointerController? {
        get { self[GraphCanvasPointerKey.self] }
        set { self[GraphCanvasPointerKey.self] = newValue }
    }
}
