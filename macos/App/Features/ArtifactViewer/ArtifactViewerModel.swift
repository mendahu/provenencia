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

    /// Composer Find session (S8-04). Image Artifacts have no Find surface.
    var isFindPresented = false
    var findQuery = "" {
        didSet {
            guard isFindPresented, oldValue != findQuery else { return }
            runFind(resetToFirst: true)
        }
    }
    private(set) var findHasTextLayer = true
    private(set) var findSelections: [PDFSelection] = []
    private(set) var activeFindIndex: Int?
    private(set) var findNote: ArtifactFindNote = .idle
    /// Bumps when the active hit is (re)applied so `PDFView` re-scrolls.
    private(set) var findActivationID = 0

    var activeFindSelection: PDFSelection? {
        guard isFindPresented, let activeFindIndex,
              findSelections.indices.contains(activeFindIndex)
        else { return nil }
        return findSelections[activeFindIndex]
    }

    var findMatchCount: Int { findSelections.count }

    var findCountSuffix: String? {
        guard isFindPresented, findHasTextLayer else { return nil }
        let query = findQuery.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !query.isEmpty else { return nil }
        if findSelections.isEmpty { return "0" }
        guard let activeFindIndex else { return "0" }
        return L10n.ArtifactViewer.findMatchOf(
            current: activeFindIndex + 1,
            total: findSelections.count
        )
    }

    var findIsInvalid: Bool {
        guard isFindPresented, findHasTextLayer else { return false }
        let query = findQuery.trimmingCharacters(in: .whitespacesAndNewlines)
        return !query.isEmpty && findSelections.isEmpty
    }

    var findStatusNote: String { findNote.localizedString }

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
        resetFindSession()
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
        resetFindSession()
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
                findHasTextLayer = true
            case .pdf(let document):
                pageCount = max(document.pageCount, 1)
                page = 1
                displayImage = nil
                findHasTextLayer = Self.documentHasTextLayer(document)
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

    func openFind() {
        guard kind == .pdf else { return }
        isFindPresented = true
        refreshTextLayerFlag()
        runFind(resetToFirst: true)
    }

    func closeFind() {
        isFindPresented = false
        clearFindHits()
    }

    func toggleFind() {
        if isFindPresented {
            closeFind()
        } else {
            openFind()
        }
    }

    func findNext() {
        guard kind == .pdf else { return }
        if findSelections.isEmpty {
            runFind(resetToFirst: true)
            return
        }
        moveFind(delta: 1)
    }

    func findPrevious() {
        guard kind == .pdf else { return }
        if findSelections.isEmpty {
            runFind(resetToFirst: true)
            return
        }
        moveFind(delta: -1)
    }

    static func documentHasTextLayer(_ document: PDFDocument) -> Bool {
        for index in 0..<document.pageCount {
            guard let page = document.page(at: index),
                  let string = page.string
            else { continue }
            if !string.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                return true
            }
        }
        return false
    }

    private func unloadPreservingGeneration() {
        kind = .unsupported
        content = nil
        displayImage = nil
        emptyReason = .idle
        page = 1
        pageCount = 1
        zoom = 1
        resetFindSession()
    }

    private func resetFindSession() {
        isFindPresented = false
        findQuery = ""
        findSelections = []
        activeFindIndex = nil
        findNote = .idle
        findActivationID = 0
        findHasTextLayer = true
    }

    private func refreshTextLayerFlag() {
        guard let document = pdfDocument else {
            findHasTextLayer = true
            return
        }
        findHasTextLayer = Self.documentHasTextLayer(document)
    }

    private func runFind(resetToFirst: Bool) {
        guard kind == .pdf, let document = pdfDocument else {
            clearFindHits()
            return
        }
        refreshTextLayerFlag()
        if !findHasTextLayer {
            clearFindHits()
            findNote = .noTextLayer
            return
        }

        let query = findQuery.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !query.isEmpty else {
            clearFindHits()
            return
        }

        let hits = document.findString(query, withOptions: .caseInsensitive)
        findSelections = hits
        guard !hits.isEmpty else {
            activeFindIndex = nil
            findNote = .noMatches(pageCount: pageCount)
            return
        }

        if resetToFirst || activeFindIndex == nil {
            activateHit(0, wrapped: .none)
            return
        }
        let clamped = min(activeFindIndex ?? 0, hits.count - 1)
        activateHit(clamped, wrapped: .none)
    }

    private func moveFind(delta: Int) {
        guard !findSelections.isEmpty else { return }
        let count = findSelections.count
        let current = activeFindIndex ?? 0
        var index = (current + delta) % count
        if index < 0 { index += count }
        let wrapped: ArtifactFindWrap
        if delta > 0, current + delta >= count {
            wrapped = .first
        } else if delta < 0, current + delta < 0 {
            wrapped = .last
        } else {
            wrapped = .none
        }
        activateHit(index, wrapped: wrapped)
    }

    private func activateHit(_ index: Int, wrapped: ArtifactFindWrap) {
        guard findSelections.indices.contains(index), let document = pdfDocument else { return }
        activeFindIndex = index
        findActivationID += 1
        let previousPage = page
        let selection = findSelections[index]
        if let pageNumber = Self.pageNumber(for: selection, in: document) {
            setPage(pageNumber)
        }
        switch wrapped {
        case .first:
            findNote = .wrappedFirst(page: page)
        case .last:
            findNote = .wrappedLast(page: page)
        case .none:
            let nextIndex = (index + 1) % findSelections.count
            let nextPage = findSelections.indices.contains(nextIndex)
                ? Self.pageNumber(for: findSelections[nextIndex], in: document)
                : nil
            let nextMatchPage = (nextPage != page) ? nextPage : nil
            findNote = .match(
                page: page,
                jumped: page != previousPage,
                nextMatchPage: nextMatchPage
            )
        }
    }

    private func clearFindHits() {
        findSelections = []
        activeFindIndex = nil
        findActivationID += 1
        if findHasTextLayer {
            findNote = .idle
        }
    }

    private static func pageNumber(for selection: PDFSelection, in document: PDFDocument) -> Int? {
        guard let pdfPage = selection.pages.first else { return nil }
        let index = document.index(for: pdfPage)
        guard index != NSNotFound else { return nil }
        return index + 1
    }

}

/// Status line under the PDF Find field (S8-04).
enum ArtifactFindNote: Equatable {
    case idle
    case noTextLayer
    case noMatches(pageCount: Int)
    case match(page: Int, jumped: Bool, nextMatchPage: Int?)
    case wrappedFirst(page: Int)
    case wrappedLast(page: Int)

    var localizedString: String {
        switch self {
        case .idle:
            return ""
        case .noTextLayer:
            return String(localized: L10n.ArtifactViewer.findNoteNoTextLayer)
        case .noMatches(let pageCount):
            return L10n.ArtifactViewer.findNoteNoMatches(pageCount: pageCount)
        case .match(let page, let jumped, let nextMatchPage):
            var parts: [String] = []
            if jumped {
                parts.append(L10n.ArtifactViewer.findNoteJumped(page: page))
            }
            if let nextMatchPage {
                parts.append(L10n.ArtifactViewer.findNoteNextPage(page: nextMatchPage))
            }
            return parts.joined(separator: " ")
        case .wrappedFirst(let page):
            return L10n.ArtifactViewer.findNoteWrappedFirst(page: page)
        case .wrappedLast(let page):
            return L10n.ArtifactViewer.findNoteWrappedLast(page: page)
        }
    }
}

private enum ArtifactFindWrap {
    case none, first, last
}

private enum ArtifactViewerLoadError: Error {
    case decodeFailed
}
