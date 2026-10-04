import AppKit
import Foundation
import Observation

/// OCR and paste-transcription commands for the citation composer.
@MainActor
@Observable
final class CitationTranscriptionModel {
    private let fields: CitationFieldsDraft
    private let artifactViewer: ArtifactViewerModel
    private let ocrEngine: any OCREngine

    var isTranscribing = false
    var ocrMessage: String?
    private var lastPastePage: Int?
    private var lastPasteText = ""
    private var isApplyingPaste = false

    init(
        fields: CitationFieldsDraft,
        artifactViewer: ArtifactViewerModel,
        ocrEngine: any OCREngine
    ) {
        self.fields = fields
        self.artifactViewer = artifactViewer
        self.ocrEngine = ocrEngine
    }

    /// Image raster only — PDF Artifacts use a live `PDFDocument`, not `displayImage`.
    var imageRaster: NSImage? {
        guard artifactViewer.kind == .image else { return nil }
        return artifactViewer.displayImage
    }

    var canAutoTranscribe: Bool {
        imageRaster != nil && !isTranscribing
    }

    var canPasteTranscription: Bool {
        artifactViewer.kind == .pdf
            && artifactViewer.findHasTextLayer
            && artifactViewer.hasUserSelection
    }

    var needsWholePageWarning: Bool {
        guard let image = imageRaster, fields.locator.region == nil else { return false }
        return OCRImage.isOversized(image.size)
    }

    var actionHint: String {
        if artifactViewer.kind == .pdf {
            return pasteHint
        }
        return autoTranscribeHint
    }

    var autoTranscribeHint: String {
        switch artifactViewer.kind {
        case .pdf:
            return pasteHint
        case .audio:
            return L10n.string(L10n.CitationComposer.autoTranscribeHintAudio)
        case .video:
            return L10n.string(L10n.CitationComposer.autoTranscribeHintVideo)
        case .unsupported:
            return L10n.string(L10n.CitationComposer.autoTranscribeHintNoRaster)
        case .image:
            break
        }
        if imageRaster == nil {
            return artifactViewer.emptyReason == .missingFile
                ? L10n.string(L10n.CitationComposer.autoTranscribeHintMissingFile)
                : L10n.string(L10n.CitationComposer.autoTranscribeHintNoRaster)
        }
        if fields.locator.hasRegion {
            return L10n.string(L10n.CitationComposer.autoTranscribeHintRegion)
        }
        return L10n.string(L10n.CitationComposer.autoTranscribeHintWholeImage)
    }

    private var pasteHint: String {
        if !artifactViewer.findHasTextLayer {
            return L10n.string(L10n.CitationComposer.pasteHintNoTextLayer)
        }
        if let page = lastPastePage, artifactViewer.userSelectionText == lastPasteText {
            return L10n.CitationComposer.pasteHintAfter(page: page)
        }
        if artifactViewer.hasUserSelection, let page = artifactViewer.userSelectionPage {
            return L10n.CitationComposer.pasteHintSelected(
                lines: artifactViewer.userSelectionLineCount,
                page: page
            )
        }
        return L10n.string(L10n.CitationComposer.pasteHintSelect)
    }

    func noteUserEditedTranscription() {
        guard !isApplyingPaste else { return }
        ocrMessage = nil
        lastPastePage = nil
        lastPasteText = ""
    }

    /// Returns a confirm payload when replacing text or warning about whole-page OCR.
    func autoTranscribeConfirmIfNeeded() -> CitationComposerModel.PendingTranscriptionConfirm? {
        guard canAutoTranscribe else { return nil }
        ocrMessage = nil
        let hasText = !fields.transcription.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        let wholePage = needsWholePageWarning
        if hasText || wholePage {
            return CitationComposerModel.PendingTranscriptionConfirm(
                existingText: fields.transcription,
                includeWholePage: wholePage
            )
        }
        return nil
    }

    func pasteConfirmIfNeeded() -> CitationComposerModel.PendingTranscriptionConfirm? {
        guard canPasteTranscription else { return nil }
        let hasText = !fields.transcription.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        guard hasText else { return nil }
        return CitationComposerModel.PendingTranscriptionConfirm(
            existingText: fields.transcription,
            includeWholePage: false,
            pastePage: artifactViewer.userSelectionPage,
            pasteLineCount: artifactViewer.userSelectionLineCount
        )
    }

    func applyPasteTranscription() {
        guard canPasteTranscription else { return }
        let text = artifactViewer.userSelectionText
            .trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        isApplyingPaste = true
        lastPastePage = artifactViewer.userSelectionPage
        lastPasteText = artifactViewer.userSelectionText
        fields.transcription = text
        isApplyingPaste = false
        ocrMessage = nil
    }

    func dismissOCRMessage() {
        ocrMessage = nil
    }

    func runAutoTranscribe() async {
        guard canAutoTranscribe, let nsImage = imageRaster else { return }
        guard let cgImage = OCRImage.cgImage(from: nsImage) else {
            ocrMessage = String(localized: L10n.CitationComposer.autoTranscribeFailed)
            return
        }
        let source: CGImage
        if let region = fields.locator.region, region.isValid {
            let bounds = ArtifactRegionGeometry.boundingRect(of: region.points)
            guard let cropped = OCRImage.crop(cgImage, normalizedRect: bounds) else {
                ocrMessage = String(localized: L10n.CitationComposer.autoTranscribeFailed)
                return
            }
            source = cropped
        } else {
            source = cgImage
        }
        isTranscribing = true
        defer { isTranscribing = false }
        do {
            let text = try await ocrEngine.recognizeText(in: source)
                .trimmingCharacters(in: .whitespacesAndNewlines)
            if text.isEmpty {
                ocrMessage = String(localized: L10n.CitationComposer.autoTranscribeNothingFound)
                return
            }
            fields.transcription = text
            ocrMessage = nil
        } catch {
            ocrMessage = String(localized: L10n.CitationComposer.autoTranscribeFailed)
        }
    }
}
