@preconcurrency import AppKit
import SwiftUI

/// Spatial pan/zoom shell for image / PDF-page rasters (document range 0.5…4).
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

        let imageView = NSImageView(frame: .zero)
        imageView.imageScaling = .scaleProportionallyUpOrDown
        imageView.imageAlignment = .alignCenter
        imageView.animates = false
        scrollView.documentView = imageView
        context.coordinator.imageView = imageView
        context.coordinator.installedContentID = contentID
        applyImage(image, to: imageView, in: scrollView)
        return scrollView
    }

    func updateNSView(_ scrollView: ArtifactMediaNSScrollView, context: Context) {
        scrollView.coordinator = context.coordinator
        context.coordinator.onZoomChange = onZoomChange
        let coordinator = context.coordinator
        let idChanged = coordinator.installedContentID != contentID
        coordinator.installedContentID = contentID

        if let imageView = coordinator.imageView {
            applyImage(image, to: imageView, in: scrollView)
        }

        let target = ArtifactViewerModel.clampZoom(zoom)
        if abs(scrollView.magnification - target) > 0.001 {
            scrollView.magnification = target
        }
        if idChanged {
            scrollView.magnification = ArtifactViewerModel.clampZoom(zoom)
            DispatchQueue.main.async {
                Self.centerDocument(in: scrollView)
            }
        }
    }

    private func applyImage(
        _ image: NSImage?,
        to imageView: NSImageView,
        in scrollView: NSScrollView
    ) {
        imageView.image = image
        guard let image else {
            imageView.frame = .zero
            return
        }
        let size = image.size
        let frame = CGRect(origin: .zero, size: size)
        imageView.frame = frame
        scrollView.documentView?.setFrameSize(size)
    }

    private static func centerDocument(in scrollView: NSScrollView) {
        guard let document = scrollView.documentView else { return }
        let clip = scrollView.contentView.bounds.size
        let doc = document.bounds.size
        let origin = CGPoint(
            x: max(0, (doc.width - clip.width) / 2),
            y: max(0, (doc.height - clip.height) / 2)
        )
        scrollView.contentView.scroll(to: origin)
        scrollView.reflectScrolledClipView(scrollView.contentView)
    }

    final class Coordinator: NSObject {
        var imageView: NSImageView?
        var installedContentID: AnyHashable?
        var onZoomChange: (CGFloat) -> Void

        init(onZoomChange: @escaping (CGFloat) -> Void) {
            self.onZoomChange = onZoomChange
        }
    }
}

/// Wheel-zoom + magnification-reporting scroll view for artifact media.
final class ArtifactMediaNSScrollView: NSScrollView {
    weak var coordinator: ArtifactMediaViewport.Coordinator?

    override func scrollWheel(with event: NSEvent) {
        if event.hasPreciseScrollingDeltas {
            super.scrollWheel(with: event)
            return
        }
        // Discrete mouse wheel → zoom toward cursor (same idea as GraphCanvas).
        let factor: CGFloat = event.scrollingDeltaY > 0 ? 1.1 : (1 / 1.1)
        let target = ArtifactViewerModel.clampZoom(magnification * factor)
        guard abs(target - magnification) > 0.000_1 else { return }
        setMagnification(target, centeredAt: contentView.convert(event.locationInWindow, from: nil))
        coordinator?.onZoomChange(magnification)
    }

    override var magnification: CGFloat {
        didSet {
            coordinator?.onZoomChange(ArtifactViewerModel.clampZoom(magnification))
        }
    }
}
