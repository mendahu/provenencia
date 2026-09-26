import AppKit
import Foundation
import Observation
import PDFKit

/// Isolated Artifact document session (S7-06). Load via ``ProjectFiles`` only.
@Observable
@MainActor
final class ArtifactViewerModel {
    static let minZoom: CGFloat = 0.5
    static let maxZoom: CGFloat = 4.0
    static let zoomStep: CGFloat = 0.25

    private(set) var kind: ArtifactViewerKind = .unsupported
    private(set) var content: ArtifactViewerContent?
    private(set) var emptyReason: ArtifactViewerEmptyReason = .idle
    private(set) var isLoading = false

    /// 1-based page index when ``supportsPages``.
    private(set) var page: Int = 1
    private(set) var pageCount: Int = 1
    private(set) var zoom: CGFloat = 1

    /// Raster for the image spatial viewport. PDF Artifacts keep ``pdfDocument``
    /// and do not flatten a page into this (S8-03).
    private(set) var displayImage: NSImage?

    /// Live PDFKit document when ``kind`` is ``ArtifactViewerKind/pdf``.
    var pdfDocument: PDFDocument? {
        guard case .pdf(let document) = content else { return nil }
        return document
    }

    /// Media-box size of the current PDF page in points (y-up page space).
    var pdfPageMediaSize: CGSize? {
        guard let document = pdfDocument,
              let pdfPage = document.page(at: page - 1)
        else { return nil }
        let bounds = pdfPage.bounds(for: .mediaBox)
        guard bounds.width > 0, bounds.height > 0 else { return nil }
        return bounds.size
    }

    var supportsPages: Bool { kind.supportsPages }
    var supportsSpatialZoom: Bool { kind.supportsSpatialZoom }
    var locatorCapabilities: ArtifactLocatorCapabilities { kind.locatorCapabilities }

    var zoomPercentLabel: String {
        let pct = Int((zoom * 100).rounded())
        return "\(pct)%"
    }

    private var loadGeneration = 0

    func unload() {
        loadGeneration += 1
        kind = .unsupported
        content = nil
        displayImage = nil
        emptyReason = .idle
        isLoading = false
        page = 1
        pageCount = 1
        zoom = 1
    }

    /// Test hook: install an in-memory image raster without going through ``load``.
    func installImageRasterForTesting(_ image: NSImage) {
        kind = .image
        content = .image(image)
        displayImage = image
        emptyReason = .idle
        isLoading = false
        page = 1
        pageCount = 1
        zoom = 1
    }

    /// Sole entry point for hosts (composer, future Source sheet, …).
    func load(_ source: ArtifactViewerSource) async {
        loadGeneration += 1
        let generation = loadGeneration
        unloadPreservingGeneration()
        isLoading = true

        let kind = ArtifactViewerKind.classify(mediaType: source.mediaType)
        self.kind = kind

        guard kind.isRenderableInS706 else {
            emptyReason = (kind == .audio || kind == .video)
                ? .mediaNotYetAvailable
                : .unsupportedKind
            isLoading = false
            return
        }

        guard let url = ProjectFiles.objectURL(
            projectDir: source.projectDir,
            relPath: source.relPath
        ) else {
            emptyReason = .missingFile
            isLoading = false
            return
        }

        guard FileManager.default.fileExists(atPath: url.path) else {
            emptyReason = .missingFile
            isLoading = false
            return
        }

        let loaded: Result<ArtifactViewerContent, Error> = await Task.detached(priority: .userInitiated) {
            switch kind {
            case .image:
                guard let image = NSImage(contentsOf: url) else {
                    return .failure(ArtifactViewerLoadError.decodeFailed)
                }
                return .success(.image(image))
            case .pdf:
                guard let document = PDFDocument(url: url) else {
                    return .failure(ArtifactViewerLoadError.decodeFailed)
                }
                return .success(.pdf(document))
            case .audio, .video, .unsupported:
                return .failure(ArtifactViewerLoadError.decodeFailed)
            }
        }.value

        guard generation == loadGeneration else { return }

        switch loaded {
        case .success(let content):
            self.content = content
            emptyReason = .idle
            zoom = 1
            switch content {
            case .image(let image):
                page = 1
                pageCount = 1
                displayImage = image
            case .pdf(let document):
                pageCount = max(document.pageCount, 1)
                page = 1
                displayImage = nil
            }
        case .failure:
            content = nil
            displayImage = nil
            emptyReason = .loadFailed
        }
        isLoading = false
    }

    func goToPreviousPage() {
        guard supportsPages, page > 1 else { return }
        setPage(page - 1)
    }

    func goToNextPage() {
        guard supportsPages, page < pageCount else { return }
        setPage(page + 1)
    }

    /// Clamps and applies a 1-based page; no-op when pages unsupported.
    func setPage(_ newPage: Int) {
        guard supportsPages else { return }
        let clamped = Self.clampPage(newPage, pageCount: pageCount)
        guard clamped != page else { return }
        page = clamped
    }

    func zoomIn() {
        guard supportsSpatialZoom else { return }
        setZoom(zoom + Self.zoomStep)
    }

    func zoomOut() {
        guard supportsSpatialZoom else { return }
        setZoom(zoom - Self.zoomStep)
    }

    func setZoom(_ value: CGFloat) {
        guard supportsSpatialZoom else { return }
        zoom = Self.clampZoom(value)
    }

    static func clampZoom(_ value: CGFloat) -> CGFloat {
        min(max(value, minZoom), maxZoom)
    }

    static func clampPage(_ page: Int, pageCount: Int) -> Int {
        let count = max(pageCount, 1)
        return min(max(page, 1), count)
    }

    private func unloadPreservingGeneration() {
        kind = .unsupported
        content = nil
        displayImage = nil
        emptyReason = .idle
        page = 1
        pageCount = 1
        zoom = 1
    }

}

private enum ArtifactViewerLoadError: Error {
    case decodeFailed
}
