@preconcurrency import AppKit
import Observation
import SwiftUI

/// AppKit pan/zoom shell for any graph-style canvas (design note §7.2).
///
/// Product-agnostic tooling: SwiftUI `ScrollView` has no magnification, so
/// hosts compose this bridge with their own document content.
///
/// **Coordinate seams**
/// - Live pointer hit-test (place / ghost): AppKit `convert` via
///   ``GraphCanvasViewportController/contentPointUnderCursor()``
/// - Pure layout / unit tests: ``GraphCanvasCoordinates`` camera math + clamp
///
/// **Input mapping**
/// - Trackpad (precise scroll): two-finger pan (system default)
/// - Trackpad pinch: zoom (system magnification)
/// - Mouse wheel (line scroll): zoom toward the cursor
/// - Click-drag on empty document: pan (via ``GraphCanvasViewportController``)
///
/// `updateNSView` only replaces the hosting root when `contentID` or
/// `contentSize` changes. Live `@Observable` / `@GestureState` updates must
/// happen **inside** the hosted tree — rewriting `rootView` every parent
/// render destroys in-flight drags.
struct GraphCanvasScrollView<Content: View>: NSViewRepresentable {
    var contentSize: CGSize
    /// Stable identity for the document (e.g. source id). Changing it rebuilds
    /// the hosted root and resets magnification; keeping it stable preserves
    /// gesture state and zoom.
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

        context.coordinator.viewport.scrollView = scrollView
        scrollView.viewport = context.coordinator.viewport

        // Must accept first responder so `.focusable()` cards inside the
        // document join the window tab loop (plain `NSHostingView` does not).
        let hosting = FocusableHostingView(rootView: hostedRoot(context: context))
        hosting.frame = CGRect(origin: .zero, size: contentSize)
        scrollView.documentView = hosting
        context.coordinator.hostingView = hosting
        context.coordinator.installedContentID = contentID
        context.coordinator.installedContentSize = contentSize

        // Start near the center of the content plane so an empty canvas feels open.
        DispatchQueue.main.async {
            Self.centerDocument(in: scrollView, contentSize: contentSize)
        }
        return scrollView
    }

    func updateNSView(_ scrollView: GraphCanvasNSScrollView, context: Context) {
        let coordinator = context.coordinator
        coordinator.viewport.scrollView = scrollView
        scrollView.viewport = coordinator.viewport
        let idChanged = coordinator.installedContentID != contentID
        let sizeChanged = coordinator.installedContentSize != contentSize
        if idChanged || sizeChanged {
            coordinator.installedContentID = contentID
            coordinator.installedContentSize = contentSize
            coordinator.hostingView?.rootView = hostedRoot(context: context)
            if let document = scrollView.documentView {
                document.frame = CGRect(origin: .zero, size: contentSize)
            }
            if idChanged {
                scrollView.magnification = 1
                DispatchQueue.main.async {
                    Self.centerDocument(in: scrollView, contentSize: contentSize)
                }
            }
        }
    }

    private func hostedRoot(context: Context) -> AnyView {
        AnyView(
            content()
                .environment(\.graphCanvasViewport, context.coordinator.viewport)
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
        var hostingView: NSHostingView<AnyView>?
        var installedContentID: AnyHashable?
        var installedContentSize: CGSize?
        let viewport: GraphCanvasViewportController

        @MainActor
        init() {
            viewport = GraphCanvasViewportController()
        }
    }
}

/// `NSHostingView` that stays in the window key-view loop so SwiftUI
/// `.focusable()` document content (Evidence subject cards) can be Tab targets.
///
/// It must **not** become first responder on mouse-down. `NSScrollView`
/// plus a first-responder document view swallows the drag before SwiftUI
/// `DragGesture` sees it. Tab still reaches hosted `.focusable()` views
/// via `acceptsFirstResponder`; pointer stays a pointer.
private final class FocusableHostingView<Content: View>: NSHostingView<Content> {
    override var acceptsFirstResponder: Bool { true }

    override func acceptsFirstMouse(for event: NSEvent?) -> Bool { false }

    override func becomeFirstResponder() -> Bool {
        // Keyboard (Tab) may focus hosted content. A mouse click that made
        // this view first responder is what kills card drag.
        guard NSApp.currentEvent?.type != .leftMouseDown else { return false }
        return super.becomeFirstResponder()
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
        // Discrete mouse wheel → zoom toward the cursor (scroll up = zoom in).
        let sensitivity: CGFloat = 0.08
        let factor = exp(event.scrollingDeltaY * sensitivity)
        let target = GraphCanvasCoordinates.clampMagnification(magnification * factor)
        guard abs(target - magnification) > 0.000_1 else { return }
        zoom(to: target, windowLocation: event.locationInWindow)
    }

    /// Zooms so the document point under `windowLocation` stays under the cursor.
    ///
    /// Avoids `setMagnification(_:centeredAt:)`, which mis-anchors against an
    /// `NSHostingView` document. Anchor math uses the clip view’s bounds space
    /// (already document coordinates under magnification).
    func zoom(to newMagnification: CGFloat, windowLocation: CGPoint) {
        guard documentView != nil else {
            magnification = newMagnification
            return
        }
        let oldMag = magnification
        let newMag = GraphCanvasCoordinates.clampMagnification(newMagnification)
        guard abs(newMag - oldMag) > 0.000_1 else { return }

        let clip = contentView
        // Clip-view bounds space == document coordinates of the point under the cursor.
        let anchor = clip.convert(windowLocation, from: nil)

        magnification = newMag

        let after = clip.convert(windowLocation, from: nil)
        // If the cursor now maps to a higher document point than before, scroll
        // origin down so the original anchor returns under the cursor.
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

    /// Precise / phase-bearing events come from trackpads and Magic Mouse —
    /// keep those as pan. Traditional mouse wheels are line-delta only.
    static func isTrackpadOrMagicMouseScroll(_ event: NSEvent) -> Bool {
        GraphCanvasScrollInput.isTrackpadStyleScroll(
            hasPreciseScrollingDeltas: event.hasPreciseScrollingDeltas,
            hasPhase: event.phase != [],
            hasMomentumPhase: event.momentumPhase != []
        )
    }
}

// MARK: - Click-drag pan (from SwiftUI document background)

/// Bridges empty-canvas drag gestures into the hosting `NSScrollView`.
@MainActor
@Observable
final class GraphCanvasViewportController: NSObject {
    weak var scrollView: NSScrollView?
    /// True while the window first responder is inside this canvas document.
    private(set) var documentContainsKeyboardFocus = false
    private var observedWindow: NSWindow?

    /// Pans by a delta measured in **window / view** points (e.g. SwiftUI
    /// `.global` drag translation). Divides by magnification so one screen
    /// pixel of drag moves one screen pixel of content.
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
    ///
    /// This is the **live hit-test seam**. Prefer it over SwiftUI hover/tap
    /// locations under magnification — `NSScrollView` does not map those into
    /// the hosted document correctly (design note §7.2). Pure
    /// `GraphCanvasCoordinates` helpers remain for layout and unit tests.
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

extension EnvironmentValues {
    var graphCanvasViewport: GraphCanvasViewportController? {
        get { self[GraphCanvasViewportKey.self] }
        set { self[GraphCanvasViewportKey.self] = newValue }
    }
}
