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

    /// Raster for the spatial viewport (full image or current PDF page).
    private(set) var displayImage: NSImage?

    var supportsPages: Bool { kind.supportsPages }
    var supportsSpatialZoom: Bool { kind.supportsSpatialZoom }

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
                displayImage = Self.renderPDFPage(document: document, pageIndex: 0)
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
        refreshPDFPageDisplay()
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

    private func refreshPDFPageDisplay() {
        guard case .pdf(let document) = content else { return }
        displayImage = Self.renderPDFPage(document: document, pageIndex: page - 1)
    }

    private static func renderPDFPage(document: PDFDocument, pageIndex: Int) -> NSImage? {
        guard pageIndex >= 0, pageIndex < document.pageCount,
              let page = document.page(at: pageIndex)
        else { return nil }
        let bounds = page.bounds(for: .mediaBox)
        guard bounds.width > 0, bounds.height > 0 else { return nil }
        let scale: CGFloat = 2
        let size = CGSize(width: bounds.width * scale, height: bounds.height * scale)
        return page.thumbnail(of: size, for: .mediaBox)
    }
}

private enum ArtifactViewerLoadError: Error {
    case decodeFailed
}
