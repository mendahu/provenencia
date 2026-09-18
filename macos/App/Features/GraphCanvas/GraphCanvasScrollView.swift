@preconcurrency import AppKit
import SwiftUI

/// AppKit pan/zoom shell for any graph-style canvas (design note §7.2).
///
/// Product-agnostic tooling: SwiftUI `ScrollView` has no magnification, so
/// hosts compose this bridge with their own document content. Hit-testing
/// math stays in `GraphCanvasCoordinates` — do not trust raw SwiftUI gesture
/// locations under zoom.
struct GraphCanvasScrollView<Content: View>: NSViewRepresentable {
    var contentSize: CGSize
    @ViewBuilder var content: () -> Content

    func makeCoordinator() -> Coordinator {
        Coordinator()
    }

    func makeNSView(context: Context) -> NSScrollView {
        let scrollView = NSScrollView()
        scrollView.hasVerticalScroller = true
        scrollView.hasHorizontalScroller = true
        scrollView.autohidesScrollers = true
        scrollView.allowsMagnification = true
        scrollView.minMagnification = GraphCanvasCamera.minMagnification
        scrollView.maxMagnification = GraphCanvasCamera.maxMagnification
        scrollView.magnification = 1
        scrollView.drawsBackground = false
        scrollView.borderType = .noBorder

        let hosting = NSHostingView(rootView: content())
        hosting.frame = CGRect(origin: .zero, size: contentSize)
        scrollView.documentView = hosting
        context.coordinator.hostingView = hosting

        // Start near the center of the content plane so an empty canvas feels open.
        DispatchQueue.main.async {
            Self.centerDocument(in: scrollView, contentSize: contentSize)
        }
        return scrollView
    }

    func updateNSView(_ scrollView: NSScrollView, context: Context) {
        context.coordinator.hostingView?.rootView = content()
        if let document = scrollView.documentView, document.frame.size != contentSize {
            document.frame = CGRect(origin: .zero, size: contentSize)
        }
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
        let x = max(0, (contentSize.width - visible.width) / 2)
        let y = max(0, (contentSize.height - visible.height) / 2)
        scrollView.contentView.scroll(to: CGPoint(x: x, y: y))
        scrollView.reflectScrolledClipView(scrollView.contentView)
    }

    final class Coordinator {
        var hostingView: NSHostingView<Content>?
    }
}
