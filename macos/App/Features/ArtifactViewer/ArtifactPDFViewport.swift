@preconcurrency import AppKit
import PDFKit
import SwiftUI

/// Live PDFKit page for PDF Artifacts (S8-03).
///
/// - **Select:** I-beam; drag creates a `PDFSelection` (Preview).
/// - **Pan:** `PDFView` scroll view — trackpad, mouse wheel, scrollbars.
/// - **Zoom / page:** chrome drives `scaleFactor` and `go(to:)`. Single-page
///   mode so 1-based locators stay authoritative.
///
/// Image Artifacts stay on ``ArtifactMediaViewport``.
struct ArtifactPDFViewport: NSViewRepresentable {
    var document: PDFDocument
    var page: Int
    var zoom: CGFloat
    var contentID: AnyHashable
    var onZoomChange: (CGFloat) -> Void
    var overlayEnabled: Bool = false
    var overlayInput: ArtifactRegionOverlayInput?
    var onCommitRegion: ((ArtifactRegionDraft) -> Void)?
    var onDisarmRegionTool: (() -> Void)?
    var freeformDeleteTooltip: String = ""

    func makeCoordinator() -> Coordinator {
        Coordinator(onZoomChange: onZoomChange)
    }

    func makeNSView(context: Context) -> ArtifactPDFHostView {
        let host = ArtifactPDFHostView()
        context.coordinator.host = host
        context.coordinator.installedContentID = contentID
        apply(to: host, context: context, resetPage: true)
        host.startObserving()
        return host
    }

    func updateNSView(_ host: ArtifactPDFHostView, context: Context) {
        let coordinator = context.coordinator
        coordinator.host = host
        let idChanged = coordinator.installedContentID != contentID
        coordinator.installedContentID = contentID
        apply(to: host, context: context, resetPage: idChanged)
    }

    static func dismantleNSView(_ host: ArtifactPDFHostView, coordinator: Coordinator) {
        host.stopObserving()
        coordinator.host = nil
    }

    @MainActor
    final class Coordinator: NSObject {
        weak var host: ArtifactPDFHostView?
        var installedContentID: AnyHashable?
        var onZoomChange: (CGFloat) -> Void
        var onCommitRegion: ((ArtifactRegionDraft) -> Void)?
        var onDisarmRegionTool: (() -> Void)?

        init(onZoomChange: @escaping (CGFloat) -> Void) {
            self.onZoomChange = onZoomChange
        }
    }

    @MainActor
    private func apply(to host: ArtifactPDFHostView, context: Context, resetPage: Bool) {
        let coordinator = context.coordinator
        coordinator.onZoomChange = onZoomChange
        host.onZoomChange = { [weak coordinator] value in
            coordinator?.onZoomChange(value)
        }
        coordinator.onCommitRegion = onCommitRegion
        coordinator.onDisarmRegionTool = onDisarmRegionTool

        host.overlayView.isHidden = !overlayEnabled
        host.overlayView.armedTool = overlayInput?.armedTool
        host.overlayView.committed = overlayInput?.committed
        host.overlayView.toolTip = freeformDeleteTooltip
        host.overlayView.onCommit = { [weak coordinator] draft in
            coordinator?.onCommitRegion?(draft)
        }
        host.overlayView.onDisarm = { [weak coordinator] in
            coordinator?.onDisarmRegionTool?()
        }

        host.sync(
            document: document,
            page: page,
            zoom: zoom,
            resetPage: resetPage,
            selectionEnabled: overlayInput?.armedTool == nil
        )
    }
}

// MARK: - Host (PDFView + region overlay)

final class ArtifactPDFHostView: NSView {
    let pdfView = PDFView()
    let overlayView = ArtifactRegionOverlayView(frame: .zero)

    var onZoomChange: ((CGFloat) -> Void)?

    private var isObserving = false
    private var suppressZoomCallback = false
    private var observedScrollViews: [NSView] = []

    override init(frame frameRect: NSRect) {
        super.init(frame: frameRect)
        wantsLayer = true
        pdfView.translatesAutoresizingMaskIntoConstraints = false
        overlayView.translatesAutoresizingMaskIntoConstraints = false
        addSubview(pdfView)
        addSubview(overlayView)

        NSLayoutConstraint.activate([
            pdfView.leadingAnchor.constraint(equalTo: leadingAnchor),
            pdfView.trailingAnchor.constraint(equalTo: trailingAnchor),
            pdfView.topAnchor.constraint(equalTo: topAnchor),
            pdfView.bottomAnchor.constraint(equalTo: bottomAnchor),
            overlayView.leadingAnchor.constraint(equalTo: leadingAnchor),
            overlayView.trailingAnchor.constraint(equalTo: trailingAnchor),
            overlayView.topAnchor.constraint(equalTo: topAnchor),
            overlayView.bottomAnchor.constraint(equalTo: bottomAnchor),
        ])

        pdfView.displayMode = .singlePage
        pdfView.displayBox = .mediaBox
        pdfView.autoScales = false
        pdfView.minScaleFactor = ArtifactViewerModel.minZoom
        pdfView.maxScaleFactor = ArtifactViewerModel.maxZoom
        pdfView.backgroundColor = NSColor(PVColor.surfaceSunken)
        pdfView.pageShadowsEnabled = false
    }

    @available(*, unavailable)
    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override var isFlipped: Bool { false }

    override func layout() {
        super.layout()
        layoutPageOverlay()
    }

    override func scrollWheel(with event: NSEvent) {
        pdfView.scrollWheel(with: event)
    }

    func startObserving() {
        guard !isObserving else { return }
        isObserving = true
        let center = NotificationCenter.default
        center.addObserver(
            self,
            selector: #selector(pdfLayoutChanged(_:)),
            name: .PDFViewPageChanged,
            object: pdfView
        )
        center.addObserver(
            self,
            selector: #selector(pdfLayoutChanged(_:)),
            name: .PDFViewScaleChanged,
            object: pdfView
        )
        center.addObserver(
            self,
            selector: #selector(pdfLayoutChanged(_:)),
            name: .PDFViewVisiblePagesChanged,
            object: pdfView
        )
        observeScrollableViews()
    }

    func stopObserving() {
        guard isObserving else { return }
        NotificationCenter.default.removeObserver(self)
        observedScrollViews = []
        isObserving = false
    }

    func sync(
        document: PDFDocument,
        page: Int,
        zoom: CGFloat,
        resetPage: Bool,
        selectionEnabled: Bool
    ) {
        if pdfView.document !== document {
            pdfView.document = document
            observeScrollableViews()
        }

        let targetPageIndex = ArtifactViewerModel.clampPage(page, pageCount: max(document.pageCount, 1)) - 1
        if let pdfPage = document.page(at: targetPageIndex),
           pdfView.currentPage != pdfPage || resetPage
        {
            pdfView.go(to: pdfPage)
        }

        let targetZoom = ArtifactViewerModel.clampZoom(zoom)
        if abs(pdfView.scaleFactor - targetZoom) > 0.001 {
            suppressZoomCallback = true
            pdfView.scaleFactor = targetZoom
            suppressZoomCallback = false
        }

        if !selectionEnabled {
            pdfView.currentSelection = nil
        }

        layoutPageOverlay()
    }

    func layoutPageOverlay() {
        guard let page = pdfView.currentPage else {
            overlayView.mediaRectOverride = .zero
            overlayView.imageSize = .zero
            return
        }
        let pageBounds = page.bounds(for: .mediaBox)
        let viewRect = pdfView.convert(pageBounds, from: page)
        let hostRect = convert(viewRect, from: pdfView)
        let flipped = CGRect(
            x: hostRect.minX,
            y: bounds.height - hostRect.maxY,
            width: hostRect.width,
            height: hostRect.height
        )
        overlayView.mediaRectOverride = flipped
        overlayView.imageSize = flipped.size
    }

    @objc private func pdfLayoutChanged(_ notification: Notification) {
        layoutPageOverlay()
        if notification.name == .PDFViewScaleChanged, !suppressZoomCallback {
            onZoomChange?(ArtifactViewerModel.clampZoom(pdfView.scaleFactor))
        }
    }

    private func observeScrollableViews() {
        for view in observedScrollViews {
            NotificationCenter.default.removeObserver(
                self,
                name: NSView.boundsDidChangeNotification,
                object: view
            )
        }
        observedScrollViews = []

        var views: [NSView] = []
        if let documentView = pdfView.documentView {
            documentView.postsBoundsChangedNotifications = true
            views.append(documentView)
            if let clip = documentView.superview {
                clip.postsBoundsChangedNotifications = true
                views.append(clip)
            }
        }
        for view in views {
            NotificationCenter.default.addObserver(
                self,
                selector: #selector(pdfLayoutChanged(_:)),
                name: NSView.boundsDidChangeNotification,
                object: view
            )
        }
        observedScrollViews = views
    }
}
