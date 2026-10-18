import AppKit
import Foundation

/// One rendered cut, keyed by what it was cut from. Save reuses the bitmap
/// the preview already drew. `cut` is nil when the cut failed for that key.
struct DepictionPreviewMemo {
    var key: String
    var cut: (cgImage: CGImage, image: NSImage)?
}

/// What the depiction preview shows, read from stored values only.
enum DepictionPreviewState: Equatable {
    case none
    case cutting
    case cut(NSImage)
    case failed
    case saved(relPath: String)
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

    /// The Citation locator a depiction is checked against: the saved one on an
    /// existing Citation (the save checks the saved locator), else the unsaved one.
    private var depictionCitationLocator: CitationLocatorDraft {
        isExistingCitation ? fields.baseline.locator : fields.values.locator
    }

    /// The page a depiction must be drawn on: the saved Citation's page, or
    /// the unsaved Citation's page when it already has one.
    var depictionPage: Int? {
        depictionCitationLocator.page
    }

    /// The Citation's own region, when it has a valid one, from the same locator
    /// as ``depictionPage`` so the two cannot disagree.
    var depictionCitationRegion: ArtifactRegionDraft? {
        guard let region = depictionCitationLocator.region, region.isValid else { return nil }
        return region
    }

    /// The title of the button that sets the region to the whole file: a page
    /// on a PDF, a photo otherwise.
    var depictionWholeFileTitle: LocalizedStringResource {
        depictionCapabilities.supportsPageLocator
            ? L10n.CitationComposer.depictionUseWholePage
            : L10n.CitationComposer.depictionUseWholePhoto
    }

    /// How the PDF page rule treats a depiction region drawn now.
    private enum DepictionPageCheck {
        case ok
        case needsPage
        case wrongPage(Int)
        case stampViewerPage
    }

    private var depictionPageCheck: DepictionPageCheck {
        guard depictionCapabilities.supportsPageLocator else { return .ok }
        if let page = depictionPage {
            return artifactViewer.page == page ? .ok : .wrongPage(page)
        }
        return isExistingCitation ? .needsPage : .stampViewerPage
    }

    /// An existing PDF Citation saved with no page cannot take a region.
    var depictionNeedsPage: Bool {
        if case .needsPage = depictionPageCheck { return true }
        return false
    }

    /// Drawing on a PDF page other than the Citation's.
    var depictionWrongPage: Bool {
        guard isDrawingDepiction, case .wrongPage = depictionPageCheck else { return false }
        return true
    }

    /// The tool the viewer overlay sees: unarmed while drawing could not commit
    /// (wrong page, no page), so the overlay never offers a refused region.
    var viewerArmedRegionTool: ArtifactRegionTool? {
        get {
            isDrawingDepiction && (depictionWrongPage || depictionNeedsPage) ? nil : armedRegionTool
        }
        set { armedRegionTool = newValue }
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

    /// The saved image path of an edit whose region is unchanged (empty when Go
    /// hid an unreadable image); nil otherwise. An unchanged edit sends no new
    /// image, so it never cuts.
    private var unchangedDepictionImagePath: String? {
        guard dialogBase() == .edit, !isDepictionRedrawn,
              let draft = currentDepictionDraft, draft.hasRegion
        else { return nil }
        return draft.imageRelPath
    }

    /// What a cut of this draft depends on; nil when the draft has no region.
    /// Built from the page and the region's points, not `encodeJSON()`, whose
    /// key order changes between calls.
    private func depictionCutKey(for draft: DepictionValueDraft) -> String? {
        guard draft.hasRegion, let region = draft.locator.region else { return nil }
        return [
            selectedArtifact?.id ?? "",
            artifactViewer.kind.rawValue,
            String(artifactViewer.isLoading),
            draft.locator.page.map(String.init) ?? "",
            region.points.map { "\($0.x),\($0.y)" }.joined(separator: ";"),
        ].joined(separator: "|")
    }

    /// The key of the cut the preview needs now; nil when it needs none.
    var depictionPreviewKey: String? {
        guard let draft = currentDepictionDraft, unchangedDepictionImagePath == nil else { return nil }
        return depictionCutKey(for: draft)
    }

    /// What the preview shows, derived from stored values so `body` never cuts.
    var depictionPreviewState: DepictionPreviewState {
        if let path = unchangedDepictionImagePath {
            return path.isEmpty ? .none : .saved(relPath: path)
        }
        guard let key = depictionPreviewKey else { return .none }
        guard let memo = depictionPreviewMemo, memo.key == key else { return .cutting }
        if let cut = memo.cut { return .cut(cut.image) }
        return .failed
    }

    /// The cut of a draft's region from the open file, shared by the preview and
    /// Save. The result is stored only while the draft is still the one on screen.
    func depictionImage(for draft: DepictionValueDraft) async -> CGImage? {
        guard let key = depictionCutKey(for: draft), let region = draft.locator.region else { return nil }
        if let memo = depictionPreviewMemo, memo.key == key { return memo.cut?.cgImage }
        let rendered = await ArtifactRegionCrop.render(artifactViewer, page: draft.locator.page, region: region)
        if depictionPreviewKey == key {
            depictionPreviewMemo = DepictionPreviewMemo(
                key: key,
                cut: rendered.map { ($0, NSImage(cgImage: $0, size: .zero)) }
            )
        }
        return rendered
    }

    /// Cuts the preview for the current draft when it needs one.
    func renderDepictionPreview() async {
        guard depictionPreviewKey != nil, let draft = currentDepictionDraft else { return }
        _ = await depictionImage(for: draft)
    }

    // MARK: - Drawing

    /// What keeps Draw from starting whatever the draft: a transcription in
    /// flight, or a PDF Citation saved with no page. The Tab order reads this
    /// alone because the Property may not be committed to the draft yet.
    var depictionDrawBlocked: Bool { isTranscribing || depictionNeedsPage }

    /// The one rule for setting a depiction region: Draw, the two region
    /// buttons, and ``prepareDepictionRegion()`` read it, and the Tab order
    /// reads ``depictionDrawBlocked``.
    var canStartDepictionDrawing: Bool {
        currentDepictionDraft != nil && !depictionDrawBlocked
    }

    /// What Draw and the region buttons share: nil unless a region may be set
    /// now; otherwise the viewer is on the Citation's page (on a PDF with one)
    /// and the dialog base is returned.
    private func prepareDepictionRegion() -> ComposerBase? {
        guard canStartDepictionDrawing else { return nil }
        let base = dialogBase()
        guard base == .authoring || base == .edit else { return nil }
        if depictionCapabilities.supportsPageLocator, let page = depictionPage {
            artifactViewer.setPage(page)
        }
        return base
    }

    /// Sets the open depiction draft's region without drawing it, through the
    /// same page rule as a drawn region. False when Draw could not start.
    @discardableResult
    func useDepictionRegion(_ region: ArtifactRegionDraft) -> Bool {
        prepareDepictionRegion() != nil && commitViewerRegion(region)
    }

    func startDepictionDrawing() {
        guard let base = prepareDepictionRegion() else { return }
        artifactViewer.closeFind()
        if armedRegionTool == nil { armedRegionTool = .rectangle }
        show(.drawingDepiction(over: base))
    }

    func cancelDepictionDrawing() {
        guard case .drawingDepiction(let over) = presented else { return }
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
        switch depictionPageCheck {
        case .ok:
            break
        case .needsPage, .wrongPage:
            return false
        case .stampViewerPage:
            _ = setPage(artifactViewer.page)
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
        guard let rendered = await depictionImage(for: values.depictionDraft)
        else { throw DepictionRenderFailed() }
        let url: URL
        do {
            url = try await Task.detached { try ArtifactRegionCrop.writeTemporaryPNG(rendered) }.value
        } catch {
            throw DepictionRenderFailed()
        }
        defer { try? FileManager.default.removeItem(at: url) }
        return try await perform(
            CatalogDepictionValueInput(
                regionJSON: values.depictionDraft.locator.encodeJSON(),
                imagePath: url.path
            )
        )
    }
}
