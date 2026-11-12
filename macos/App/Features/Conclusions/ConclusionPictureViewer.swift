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
    /// The ids whose decoded images the viewer holds: the shown tile and its two neighbours (S13-15 F-15).
    var heldIDs: Set<String> { Set([tile.id, previousID, nextID].compactMap { $0 }) }
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
/// EXIF), so the frame follows the panel's width and the pane's height. Each state past `waiting` carries
/// the id of the tile it was loaded for, so a stale one never draws under another tile's counter.
enum ConclusionPictureImage {
    /// A load is starting or running: an empty band, with a spinner once the load is slow.
    case waiting
    /// The full image.
    case decoded(String, NSImage, CGSize)
    /// No file, or one the decode refused: the kind's mark.
    case refused(String)

    /// This state for the tile `id`; `waiting` while it belongs to another tile.
    func shown(for id: String) -> Self {
        switch self {
        case .waiting: self
        case .decoded(let owner, _, _), .refused(let owner):
            owner == id ? self : .waiting
        }
    }

    /// Height of the image band on the board, at Regular and at Wide.
    private static func boardHeight(_ widthClass: PVWidthClass) -> CGFloat {
        widthClass == .wide ? 950 : 440
    }

    /// The least the image band is drawn.
    static let minBandHeight: CGFloat = 320
    /// What the panel's other rows (title, record, footer) are allowed of the pane's height.
    static let chromeAllowance: CGFloat = 320

    /// The image band's height: the board's, less what a short pane lacks after the chrome's allowance.
    static func bandHeight(widthClass: PVWidthClass, paneHeight: CGFloat) -> CGFloat {
        min(boardHeight(widthClass), max(minBandHeight, paneHeight - chromeAllowance))
    }

    /// The image's frame: half its pixels in points (a small crop is not blown up), fitted inside the
    /// band's middle column and the band's height less its 20pt of padding above and below, with the aspect kept.
    static func frame(pixelWidth: CGFloat, pixelHeight: CGFloat, widthClass: PVWidthClass, bandHeight: CGFloat) -> CGSize {
        guard pixelWidth > 0, pixelHeight > 0 else { return .zero }
        let panelWidth = PVProvenanceColumns(widthClass: widthClass).panelWidth
        let availableWidth = panelWidth - 2 * PVPanelMetrics.bodyInset - 2 * stepColumn
        let availableHeight = bandHeight - 2 * PVSpacing.space7
        let wanted = CGSize(width: pixelWidth / 2, height: pixelHeight / 2)
        let scale = min(1, availableWidth / wanted.width, availableHeight / wanted.height)
        return CGSize(width: wanted.width * scale, height: wanted.height * scale)
    }

    /// Width of the previous and next columns on either side of the image.
    static let stepColumn: CGFloat = 56
}

/// A keyboard stop in the viewer: the image band, or one of the two step buttons.
private enum ViewerStop: Hashable {
    case image
    case previous
    case next
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
    /// Set once a load has taken 250 ms, so a quick decode shows no spinner and stepping does not flash.
    @State private var showsSpinner = false
    /// The decoded images of the shown tile and its neighbours, by tile id; every other id is dropped on a step.
    @State private var decoded: [String: (image: NSImage, pixels: CGSize)] = [:]
    @State private var decodeTask: Task<Result<BoundedImageDecode.Decoded, BoundedImageDecode.Refusal>, Never>?
    @FocusState private var focus: ViewerStop?

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
        } footer: {
            footer
        } titleAccessory: {
            titleAccessory
        }
        .onKeyPress(.leftArrow) { move(to: step.previousID) }
        .onKeyPress(.rightArrow) { move(to: step.nextID) }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier(identifier)
        .vocabularyToastOverlay(lead.toastInViewer ? $lead.toast : .constant(nil), identifier: "\(identifier).toast")
        .onAppear {
            lead.viewerOpened()
            FocusClaim.claim($focus, .image)
        }
        .onDisappear { lead.viewerClosed() }
        .onChange(of: step) { reclaimFocusIfStranded() }
        .task(id: step.tile.id) { await load() }
        .task(id: step.tile.id) {
            showsSpinner = false
            do { try await Task.sleep(for: .milliseconds(250)) } catch { return }
            showsSpinner = true
        }
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
                .focused($focus, equals: .previous)
                .disabled(step.previousID == nil)
                .accessibilityIdentifier("\(identifier).previous")
            }
            imageArea
            stepColumn {
                PVIconButton(.chevronForward, label: viewer.next, size: .lg) {
                    if let id = step.nextID { show(id) }
                }
                .focused($focus, equals: .next)
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
            .focused($focus, equals: .image)
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(Text(verbatim: step.tile.viewerLabel))
            .accessibilityValue(isRefused ? Text(L10n.Conclusions.likenessViewerImageUnavailable) : Text(verbatim: ""))
            .accessibilityAddTraits(.isImage)
            .pvAnnounce(step.tile.viewerLabel)
            .accessibilityIdentifier("\(identifier).image")
    }

    private var isRefused: Bool {
        if case .refused = image.shown(for: step.tile.id) { true } else { false }
    }

    @ViewBuilder
    private var imageContent: some View {
        switch image.shown(for: step.tile.id) {
        case .waiting:
            if showsSpinner { ProgressView().controlSize(.small) }
        case .decoded(_, let decoded, let pixels):
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
                PVButton(gallery.choose.useAs, variant: .primary, loading: lead.savingID == step.tile.id) {
                    Task {
                        _ = await lead.choose(step.tile, copy: gallery.choose)
                        FocusClaim.claim($focus, .image)
                    }
                }
                .disabled(lead.isSaving)
                .accessibilityLabel(step.tile.useAsLabel)
                .accessibilityIdentifier("\(identifier).useAsLead")
            }
        }
    }

    // MARK: Behavior

    private func move(to id: String?) -> KeyPress.Result {
        if let id { show(id) }
        return .handled
    }

    /// After a step, or a change of the gallery's size, the keyboard stays in the viewer: a focused step
    /// button that is now disabled or gone, or no focus at all, hands it to the image.
    private func reclaimFocusIfStranded() {
        let stranded = (focus == .previous && step.previousID == nil) || (focus == .next && step.nextID == nil)
        if focus == nil || stranded { FocusClaim.claim($focus, .image) }
    }

    /// Loads the shown picture, then its neighbours' ahead of a step. A held image shows at once; any other
    /// waits 120 ms, so a step more than 120 ms later cancels the wait before any file opens. One decode
    /// is in flight (`decodeTask`), and a step cancels a prefetch not yet started (S13-15 D-5, F-3, F-15).
    private func load() async {
        let step = step
        let id = step.tile.id
        decoded = decoded.filter { step.heldIDs.contains($0.key) }
        if let held = decoded[id] {
            image = .decoded(id, held.image, held.pixels)
        } else {
            image = .waiting
            do { try await Task.sleep(for: .milliseconds(120)) } catch { return }
            guard let result = await decode(step.tile) else { return }
            switch result {
            case .success(let done): store(done, for: id)
            case .failure: image = .refused(id)
            }
        }
        await prefetch(step)
    }

    private func prefetch(_ step: ConclusionPictureStep) async {
        for id in [step.nextID, step.previousID].compactMap({ $0 }) where decoded[id] == nil {
            guard let tile = step.tiles.first(where: { $0.id == id }), let result = await decode(tile),
                case .success(let done) = result, !Task.isCancelled
            else { continue }
            decoded[id] = (done.image, CGSize(width: done.pixelSize.width, height: done.pixelSize.height))
        }
    }

    /// Shows a decoded image for the tile `id` and holds it. The size is the file's pixel size as shown.
    private func store(_ done: BoundedImageDecode.Decoded, for id: String) {
        let pixels = CGSize(width: done.pixelSize.width, height: done.pixelSize.height)
        decoded[id] = (done.image, pixels)
        image = .decoded(id, done.image, pixels)
    }

    /// The bounded decode of one tile's file, after the one in flight ends; nil when this task was cancelled
    /// before the file opened or while it decoded.
    private func decode(_ tile: ConclusionDetailPicture) async -> Result<BoundedImageDecode.Decoded, BoundedImageDecode.Refusal>? {
        _ = await decodeTask?.value
        guard !Task.isCancelled else { return nil }
        guard let url = ProjectFiles.objectURL(projectDir: projectDir, relPath: tile.imageRelPath) else {
            return .failure(.unreadable)
        }
        let limits = ThumbnailDecodePolicy.limits
        let work = Task.detached(priority: .userInitiated) {
            BoundedImageDecode.decode(url: url, limits: limits)
        }
        decodeTask = work
        let result = await work.value
        return Task.isCancelled ? nil : result
    }
}

#Preview("Viewer, no image") {
    let outcome = CatalogReconcilerOutcome(
        observationID: "o1", observationRef: "OBS-1", reason: .kept, valueRank: 1, deniedByObservationID: "",
        recorded: .none, subjectID: "s1", subjectRef: "CPR-1", citationID: "c1", sourceID: "src1",
        sourceTitle: "Hamburger Adressbuch 1861", credibilityKey: "", transcriptionUncertain: false,
        artifactLabel: "Volume 3", citationLabel: "Entry 214", artifactPage: 23, pageLabel: "23"
    )
    let tiles = (1...3).map { number in
        var tile = ConclusionDetailPicture(
            imageRelPath: "", outcome: outcome, isLead: number == 1,
            accessibilityLabel: "Likeness", useAsLabel: "Use as portrait", chosenToastBody: "",
            viewerLabel: L10n.Conclusions.likenessViewerLabel(index: number, count: 3, line: CitationLineDisplay.parts(outcome).line)
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
