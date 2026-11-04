import AppKit
import SwiftUI

/// Where a gallery stands while one picture is shown: the tile, its position and its neighbours (S13-15
/// D-2). A pure value over the tiles: the position is read from the shown id, never stored, and
/// stepping does not wrap.
struct ConclusionPictureStep: Equatable {
    let tiles: [ConclusionDetailPicture]
    /// Position of the shown tile in `tiles`.
    let index: Int

    /// The step showing `id`; nil when no tile has it.
    init?(tiles: [ConclusionDetailPicture], showing id: String) {
        guard let index = tiles.firstIndex(where: { $0.id == id }) else { return nil }
        self.tiles = tiles
        self.index = index
    }

    var tile: ConclusionDetailPicture { tiles[index] }
    var count: Int { tiles.count }
    var previousID: String? { index > 0 ? tiles[index - 1].id : nil }
    var nextID: String? { index + 1 < tiles.count ? tiles[index + 1].id : nil }
}

/// Which picture the viewer shows: the picture's id, resolved against the live tiles on every draw, so the
/// viewer follows its gallery across a page refresh (S13-15 D-3). The `Identifiable.id` is constant, so
/// stepping never presents the sheet again.
struct ConclusionPictureViewerItem: Identifiable, Equatable {
    let id = "viewer"
    var shownID: String

    /// This item against a changed gallery: the same picture while it is still there, else the one now
    /// at its old position (the new last when it was last), else nil when the gallery is empty.
    func resolved(from old: [ConclusionDetailPicture], in new: [ConclusionDetailPicture]) -> Self? {
        guard !new.isEmpty else { return nil }
        if new.contains(where: { $0.id == shownID }) { return self }
        let position = old.firstIndex { $0.id == shownID } ?? 0
        var moved = self
        moved.shownID = new[min(position, new.count - 1)].id
        return moved
    }
}

/// What the viewer's image area shows (S13-15 D-6). The sizes are the file's pixel sizes as shown (turned for
/// EXIF), so the frame follows the panel's width and the pane's height.
enum ConclusionPictureImage {
    /// A load is starting: the tile's own image, else a spinner.
    case waiting
    /// The header is read: the tile's image in the final frame while the full image decodes.
    case sized(CGSize)
    /// The full image.
    case decoded(NSImage, CGSize)
    /// No file, or one the decode refused: the kind's mark.
    case refused

    /// Height of the image band on the board, at Regular and at Wide.
    private static func boardHeight(_ widthClass: PVWidthClass) -> CGFloat {
        widthClass == .wide ? 950 : 440
    }

    /// The image band's height: the board's, less what a short pane lacks (the panel's chrome is about 280pt).
    static func bandHeight(widthClass: PVWidthClass, paneHeight: CGFloat) -> CGFloat {
        min(boardHeight(widthClass), max(320, paneHeight - 320))
    }

    /// The image's frame: half its pixels in points (a small crop is not blown up), fitted inside the
    /// band's middle column and the band's height less its 20pt of padding above and below, with the aspect kept.
    static func frame(pixelWidth: CGFloat, pixelHeight: CGFloat, widthClass: PVWidthClass, bandHeight: CGFloat) -> CGSize {
        guard pixelWidth > 0, pixelHeight > 0 else { return .zero }
        let panelWidth = PVProvenanceColumns(widthClass: widthClass).panelWidth
        let availableWidth = panelWidth - 2 * PVSpacing.space7 - 2 * stepColumn
        let availableHeight = bandHeight - 2 * PVSpacing.space7
        let wanted = CGSize(width: pixelWidth / 2, height: pixelHeight / 2)
        let scale = min(1, availableWidth / wanted.width, availableHeight / wanted.height)
        return CGSize(width: wanted.width * scale, height: wanted.height * scale)
    }

    /// Width of the previous and next columns on either side of the image.
    static let stepColumn: CGFloat = 56
}

/// The likeness viewer: one picture large, stepping through the gallery, with its provenance record, Use as
/// the lead, and a way to close (S13-15). Kind-blind: it takes the gallery's words and the kind's copy, and
/// names no navigation type. Present it through `pvSheet`; it closes itself through `\.pvSheetClose`.
struct ConclusionPictureViewer: View {
    let projectDir: String
    let pageIdentifier: String
    let gallery: ConclusionDetailGallery
    let viewer: ConclusionDetailGallery.ViewerCopy
    @Bindable var lead: ConclusionLeadModel
    let step: ConclusionPictureStep
    let widthClass: PVWidthClass
    let paneHeight: CGFloat
    let show: (String) -> Void
    let open: (PVProvenanceSheetContent.Record, PVProvenanceLink) -> Void

    @Environment(\.pvSheetClose) private var sheetClose
    @State private var image: ConclusionPictureImage = .waiting
    @FocusState private var imageFocused: Bool

    private var columns: PVProvenanceColumns {
        PVProvenanceColumns(showsReadAs: false, widthClass: widthClass)
    }

    private var bandHeight: CGFloat {
        ConclusionPictureImage.bandHeight(widthClass: widthClass, paneHeight: paneHeight)
    }

    private var identifier: String { "\(pageIdentifier).viewer" }

    var body: some View {
        PVPanel(
            title: Text(verbatim: viewer.title).accessibilityHeading(.h1),
            width: columns.panelWidth
        ) {
            VStack(alignment: .leading, spacing: PVSpacing.space6) {
                band
                record
            }
            .onKeyPress(.leftArrow) { move(to: step.previousID) }
            .onKeyPress(.rightArrow) { move(to: step.nextID) }
        } footer: {
            footer
        } titleAccessory: {
            titleAccessory
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier(identifier)
        .vocabularyToastOverlay($lead.toast, identifier: "\(identifier).toast")
        .onAppear { FocusClaim.claim($imageFocused, true) }
        .task(id: step.tile.id) { await load() }
    }

    // MARK: Title row

    private var titleAccessory: some View {
        HStack(spacing: PVSpacing.space4) {
            Spacer(minLength: 0)
            if step.count > 1 {
                Text(verbatim: L10n.DesignSystem.selectOptionPosition(current: step.index + 1, count: step.count))
                    .font(PVFont.mono(size: PVTypeScale.bodySmall))
                    .foregroundStyle(PVColor.textSecondary)
                    .accessibilityHidden(true)
                    .accessibilityIdentifier("\(identifier).counter")
            }
            PVIconButton(.dismiss, label: viewer.close, size: .md) { sheetClose() }
                .accessibilityIdentifier("\(identifier).close")
        }
    }

    // MARK: Image band

    private var band: some View {
        HStack(spacing: 0) {
            stepColumn {
                PVIconButton(.chevronBack, label: viewer.previous, size: .lg) {
                    if let id = step.previousID { show(id) }
                }
                .disabled(step.previousID == nil)
                .accessibilityIdentifier("\(identifier).previous")
            }
            imageArea
            stepColumn {
                PVIconButton(.chevronForward, label: viewer.next, size: .lg) {
                    if let id = step.nextID { show(id) }
                }
                .disabled(step.nextID == nil)
                .accessibilityIdentifier("\(identifier).next")
            }
        }
        .frame(height: bandHeight)
        .frame(maxWidth: .infinity)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm))
        .overlay {
            RoundedRectangle(cornerRadius: PVRadius.sm).stroke(PVColor.borderSubtle, lineWidth: 1)
        }
    }

    /// One of the two columns beside the image; a gallery of one leaves them empty so the image keeps its cap.
    private func stepColumn<Content: View>(@ViewBuilder _ content: () -> Content) -> some View {
        ZStack {
            if step.count > 1 { content() }
        }
        .frame(width: ConclusionPictureImage.stepColumn)
        .frame(maxHeight: .infinity)
    }

    private var imageArea: some View {
        ZStack { imageContent }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(.vertical, PVSpacing.space7)
            .contentShape(Rectangle())
            .focusable()
            .focused($imageFocused)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: step.tile.viewerLabel))
            .accessibilityAddTraits(.isImage)
            .pvAnnounce(step.tile.viewerLabel)
            .accessibilityIdentifier("\(identifier).image")
    }

    @ViewBuilder
    private var imageContent: some View {
        switch image {
        case .waiting:
            placeholder(frame: nil)
        case .sized(let pixels):
            placeholder(frame: fitted(pixels))
        case .decoded(let decoded, let pixels):
            let frame = fitted(pixels)
            Image(nsImage: decoded)
                .resizable()
                .scaledToFit()
                .frame(width: frame.width, height: frame.height)
        case .refused:
            CachedThumbnail(
                projectDir: projectDir,
                relPath: "",
                size: 160,
                fallbackMark: gallery.fallbackMark,
                fallbackColors: gallery.fallbackColors
            )
        }
    }

    private func fitted(_ pixels: CGSize) -> CGSize {
        ConclusionPictureImage.frame(
            pixelWidth: pixels.width, pixelHeight: pixels.height, widthClass: widthClass, bandHeight: bandHeight
        )
    }

    /// The tile's own image while the full one loads, in the final frame once it is known (at most 256pt before).
    @ViewBuilder
    private func placeholder(frame: CGSize?) -> some View {
        if let cached = ThumbnailCache.shared.cachedImage(projectDir: projectDir, relPath: step.tile.imageRelPath) {
            if let frame {
                Image(nsImage: cached).resizable().scaledToFit().frame(width: frame.width, height: frame.height)
            } else {
                Image(nsImage: cached).resizable().scaledToFit().frame(maxWidth: 256, maxHeight: 256)
            }
        } else {
            ProgressView().controlSize(.small)
        }
    }

    // MARK: Record and footer

    private var record: some View {
        VStack(spacing: 0) {
            PVDivider(color: PVColor.borderDefault)
            PVProvenanceRecordRow(
                record: PVProvenanceSheetContent.Record(outcome: step.tile.outcome, in: nil),
                columns: columns,
                idPrefix: "\(identifier).record",
                open: { record, link in sheetClose { open(record, link) } }
            )
            PVDivider(color: PVColor.borderSubtle)
        }
    }

    @ViewBuilder
    private var footer: some View {
        HStack {
            Spacer(minLength: 0)
            if step.tile.isLead {
                PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("\(identifier).leadBadge")
            } else {
                PVButton(gallery.choose.useAs, variant: .primary) {
                    Task { _ = await lead.choose(step.tile, copy: gallery.choose) }
                }
                .disabled(lead.isSaving)
                .accessibilityIdentifier("\(identifier).useAsLead")
            }
        }
    }

    // MARK: Behavior

    private func move(to id: String?) -> KeyPress.Result {
        guard let id else { return .ignored }
        show(id)
        return .handled
    }

    /// Loads the shown picture: the tile's image first, then the file's declared size, then the full decode.
    /// A step more than 120 ms later cancels the wait before any file opens; a decode already started
    /// runs to its end and is dropped (S13-15 D-5).
    private func load() async {
        image = .waiting
        do { try await Task.sleep(for: .milliseconds(120)) } catch { return }
        let relPath = step.tile.imageRelPath
        guard let url = ProjectFiles.objectURL(projectDir: projectDir, relPath: relPath) else {
            image = .refused
            return
        }
        let types = Array(ThumbnailDecodePolicy.allowedTypes)
        let maxEdge = ThumbnailDecodePolicy.maxDeclaredEdge
        let declared = await Task.detached {
            ArtifactImageDecode.declaredSize(url: url, typeIdentifiers: types, maxEdge: maxEdge)
        }.value
        guard !Task.isCancelled else { return }
        guard case .success(let pixels) = declared else {
            image = .refused
            return
        }
        let size = CGSize(width: pixels.width, height: pixels.height)
        image = .sized(size)
        let decoded = await Task.detached {
            ArtifactImageDecode.decode(url: url, typeIdentifiers: types, maxEdge: maxEdge)
        }.value
        guard !Task.isCancelled else { return }
        switch decoded {
        case .success(let result): image = .decoded(result.image, size)
        case .failure: image = .refused
        }
    }
}

#Preview("Viewer, no image") {
    let outcome = CatalogReconcilerOutcome(
        observationID: "o1", observationRef: "OBS-1", reason: .kept, valueRank: 1, deniedByObservationID: "",
        recorded: .none, subjectID: "s1", subjectRef: "CPR-1", citationID: "c1", sourceID: "src1",
        sourceTitle: "Hamburger Adressbuch 1861", credibilityKey: "", transcriptionUncertain: false,
        artifactLabel: "Volume 3", citationLabel: "Entry 214", artifactPage: 23, pageLabel: "23"
    )
    let citation = CitationLineDisplay.parts(outcome)
    let tiles = (1...3).map { number in
        var tile = ConclusionDetailPicture(
            imageRelPath: "", outcome: outcome, citation: citation, isLead: number == 1,
            accessibilityLabel: "Likeness", useAsLabel: "Use as portrait", chosenToastBody: "",
            viewerLabel: L10n.Conclusions.likenessViewerLabel(index: number, count: 3, line: citation.line)
        )
        tile.outcome.observationID = "o\(number)"
        return tile
    }
    let gallery = ConclusionDetailGallery(
        title: L10n.Conclusions.likenessesSection,
        openTip: L10n.Conclusions.likenessViewHint,
        fallbackMark: .subjectPerson,
        tiles: tiles,
        choose: .personPortrait,
        viewer: .likeness(title: L10n.Conclusions.likenessViewerTitle("Johann Weber"))
    )
    let session = WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: FakeStore())
    return ConclusionPictureViewer(
        projectDir: "/tmp/preview.provenencia",
        pageIdentifier: "preview",
        gallery: gallery,
        viewer: gallery.viewer!,
        lead: ConclusionLeadModel(session: session, userID: "u1", entityId: "e1"),
        step: ConclusionPictureStep(tiles: tiles, showing: "o2")!,
        widthClass: .regular,
        paneHeight: 900,
        show: { _ in },
        open: { _, _ in }
    )
}
