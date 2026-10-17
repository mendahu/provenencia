import AppKit
import Foundation

/// One rendered preview, keyed by what it was cut from.
struct DepictionPreviewMemo {
    var key: String
    var image: NSImage
}

/// The region could not be cut from the open file, or its image could not be written.
struct DepictionRenderFailed: Error {}

/// Depiction authoring on the composer: drawing a region on the viewer, the
/// PDF page rule, the preview, and the temporary image sent with a save.
/// The Citation's own region stays `setRegion(_:)`; a depiction's region is
/// held on the draft and shown on the viewer while a depiction dialog is open.
extension CitationComposerModel {
    /// Locator capabilities from the selected Artifact's media type, so the
    /// Property list does not change while the viewer is still loading.
    var depictionCapabilities: ArtifactLocatorCapabilities {
        ArtifactViewerKind.classify(mediaType: selectedArtifact?.file?.mediaType ?? "").locatorCapabilities
    }

    var isDrawingDepiction: Bool {
        if case .drawingDepiction = presented { return true }
        return false
    }

    /// Values of the open observation dialog when its Property is a depiction.
    private var currentDepictionValues: ObservationRowValues? {
        guard presented != nil else { return nil }
        switch dialogBase() {
        case .authoring:
            if case .property(let draft) = authoring,
               vocabulary.isDepiction(propertyID: draft.values.propertyID)
            {
                return draft.values
            }
        case .edit:
            if let session = editSession,
               case .property(let values, _, _) = session.kind,
               vocabulary.isDepiction(propertyID: values.propertyID)
            {
                return values
            }
        case .none:
            break
        }
        return nil
    }

    var currentDepictionDraft: DepictionValueDraft? {
        currentDepictionValues?.depictionDraft
    }

    /// What the viewer shows as the committed region: the open depiction
    /// draft's, else the Citation's.
    var viewerRegion: ArtifactRegionDraft? {
        if let draft = currentDepictionDraft { return draft.locator.region }
        return locator.region
    }

    private var isExistingCitation: Bool {
        context.citationID?.isEmpty == false
    }

    /// The page a depiction must be drawn on: the saved Citation's page, or
    /// the unsaved Citation's page when it already has one.
    var depictionPage: Int? {
        isExistingCitation ? fields.baseline.locator.page : fields.values.locator.page
    }

    /// An existing PDF Citation saved with no page cannot take a region.
    var depictionNeedsPage: Bool {
        isExistingCitation
            && depictionCapabilities.supportsPageLocator
            && fields.baseline.locator.page == nil
    }

    /// Drawing on a PDF page other than the Citation's.
    var depictionWrongPage: Bool {
        guard isDrawingDepiction, depictionCapabilities.supportsPageLocator,
              let page = depictionPage
        else { return false }
        return artifactViewer.page != page
    }

    /// True when the open edit has a region unlike the saved one.
    var isDepictionRedrawn: Bool {
        guard let session = editSession, case .property(let values, _, _) = session.kind else {
            return false
        }
        return values.depictionDraft.locator != session.propertyBaseline.depictionDraft.locator
    }

    var depictionCaption: LocalizedStringResource {
        let draft = currentDepictionDraft
        if depictionNeedsPage { return L10n.CitationComposer.depictionCaptionNeedsPage }
        if draft?.regionHidden == true { return L10n.CitationComposer.depictionCaptionHidden }
        if draft?.hasRegion != true { return L10n.CitationComposer.depictionCaptionEmpty }
        if dialogBase() == .edit, isDepictionRedrawn {
            return L10n.CitationComposer.depictionCaptionRedrawn
        }
        return L10n.CitationComposer.depictionCaptionBox
    }

    /// The image a save would send, derived so it always matches the draft.
    var depictionPreview: NSImage? {
        guard let draft = currentDepictionDraft, draft.hasRegion, let region = draft.locator.region else {
            return nil
        }
        let key = [
            selectedArtifact?.id ?? "",
            artifactViewer.kind.rawValue,
            String(artifactViewer.isLoading),
            draft.locator.encodeJSON(),
        ].joined(separator: "|")
        if let memo = depictionPreviewMemo, memo.key == key { return memo.image }
        guard let rendered = ArtifactRegionCrop.render(artifactViewer, page: draft.locator.page, region: region)
        else { return nil }
        let image = NSImage(cgImage: rendered, size: .zero)
        depictionPreviewMemo = DepictionPreviewMemo(key: key, image: image)
        return image
    }

    // MARK: - Drawing

    func startDepictionDrawing() {
        guard !isTranscribing, !depictionNeedsPage, currentDepictionDraft != nil else { return }
        let base = dialogBase()
        guard base == .authoring || base == .edit else { return }
        if depictionCapabilities.supportsPageLocator, let page = depictionPage {
            artifactViewer.setPage(page)
        }
        artifactViewer.closeFind()
        ignoreNextDisarm = false
        if armedRegionTool == nil { armedRegionTool = .rectangle }
        show(.drawingDepiction(over: base))
    }

    func cancelDepictionDrawing() {
        guard case .drawingDepiction(let over) = presented else { return }
        ignoreNextDisarm = false
        armedRegionTool = nil
        restore(over)
    }

    /// The viewer's region commit. With a depiction dialog open the region goes
    /// to its draft and the dialog returns; otherwise it is the Citation's.
    @discardableResult
    func commitViewerRegion(_ region: ArtifactRegionDraft) -> Bool {
        guard var draft = currentDepictionDraft else { return setRegion(region) }
        guard !isTranscribing else { return false }
        let capabilities = depictionCapabilities
        if capabilities.supportsPageLocator {
            if let page = depictionPage {
                guard artifactViewer.page == page else {
                    // The viewer reports a disarm right after a commit; keep drawing.
                    ignoreNextDisarm = true
                    return false
                }
            } else if isExistingCitation {
                return false
            } else {
                _ = setPage(artifactViewer.page)
            }
        }
        guard draft.locator.setRegion(
            region,
            capabilities: capabilities,
            viewerPage: artifactViewer.page
        ) else { return false }
        writeDepictionDraft(draft)
        armedRegionTool = nil
        if case .drawingDepiction(let over) = presented {
            restore(over)
        }
        return true
    }

    private func writeDepictionDraft(_ draft: DepictionValueDraft) {
        switch dialogBase() {
        case .authoring:
            mutateProperty { $0.values.depictionDraft = draft }
        case .edit:
            mutateEdit { session in
                guard case .property(var values, let id, let ref) = session.kind else { return }
                values.depictionDraft = draft
                session.kind = .property(values: values, persistedID: id, persistedRef: ref)
            }
        case .none:
            break
        }
    }

    // MARK: - Save

    /// Cuts the drawn region to a temporary PNG, runs `perform` with it, and
    /// removes the file whether `perform` succeeds or throws. The cut comes
    /// first so a failed render creates nothing.
    func withDepictionInput<T>(
        for values: ObservationRowValues,
        perform: (CatalogDepictionValueInput) async throws -> T
    ) async throws -> T {
        let draft = values.depictionDraft
        guard let region = draft.locator.region,
              let rendered = ArtifactRegionCrop.render(artifactViewer, page: draft.locator.page, region: region)
        else { throw DepictionRenderFailed() }
        let url: URL
        do {
            url = try ArtifactRegionCrop.writeTemporaryPNG(rendered)
        } catch {
            throw DepictionRenderFailed()
        }
        defer { try? FileManager.default.removeItem(at: url) }
        return try await perform(
            CatalogDepictionValueInput(regionJSON: draft.locator.encodeJSON(), imagePath: url.path)
        )
    }
}
