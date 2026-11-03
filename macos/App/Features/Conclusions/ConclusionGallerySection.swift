import SwiftUI

/// A page's Gallery section: the heading with the count, then one 160pt tile per picture in a
/// wrapping grid, each captioned with its citation (S13-14, S13-D5). Kind-blind: its words come
/// from `gallery`, its ids from the page's identifier. A click on a tile's image opens the Citation
/// it was cut from, until the likeness viewer takes the click (S13-15).
///
/// One section for every kind: `look` holds the measures and the few behaviors that differ. The
/// Person's Gallery draws `.large`; the Event and Place Pictures keep `.small` until their redesign.
struct ConclusionGallerySection: View {
    let look: GalleryTileLook
    let projectDir: String
    let pageIdentifier: String
    let gallery: ConclusionDetailGallery
    let lead: ConclusionLeadModel

    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(\.pvWidthClass) private var widthClass
    @FocusState private var focused: TileFocus?
    /// The last stop that had the keyboard, kept to the next main-actor turn after it goes, so a width
    /// change in the same update still sees it while a click elsewhere clears it.
    @State private var lastFocused: TileFocus?
    /// The stop to give the keyboard back once its tile is realized again after the lazy grid let it go.
    @State private var pendingFocus: TileFocus?
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: gallery.title, meta: String(gallery.tiles.count))
            LazyVGrid(
                columns: [GridItem(
                    .adaptive(minimum: look.column, maximum: look.column),
                    spacing: look.columnGap,
                    alignment: .top
                )],
                alignment: .leading,
                spacing: look.rowGap
            ) {
                ForEach(gallery.tiles) { tile in
                    tileView(tile)
                        .id(tile.id)
                }
            }
            .pvScrollsRowIntoView(scrollTarget, animated: false, whenRequested: scrollRequest)
            .padding(.top, PVSpacing.space6)
        }
        .onChange(of: focused) {
            if let stop = focused {
                lastFocused = stop
                // The pending stop landed, or the keyboard went to another one: either way it is done.
                pendingFocus = nil
            } else if pendingFocus != nil {
                // The tile dropped after the width update: bring it back and try again.
                requestScroll()
                applyPending()
            } else if lastFocused != nil {
                Task { @MainActor in
                    if focused == nil { lastFocused = nil }
                }
            }
        }
        .onChange(of: widthClass) { keepFocus() }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.string(gallery.title))
    }

    /// Crossing the Regular/Wide line rebuilds the lazy grid and drops the focused tile: scroll it back
    /// into view, then give its stop the keyboard once the tile exists. The pending stop ends a moment
    /// later, so a tile that never dropped does not take the keyboard back after a click elsewhere.
    private func keepFocus() {
        guard let stop = focused ?? lastFocused else { return }
        pendingFocus = stop
        requestScroll()
        Task { @MainActor in
            await Task.yield()
            applyPending()
            try? await Task.sleep(for: .milliseconds(400))
            if pendingFocus == stop { pendingFocus = nil }
        }
    }

    private func requestScroll() {
        guard let id = pendingFocus?.id else { return }
        scrollTarget = id
        scrollRequest += 1
    }

    /// Writes the pending stop. Use as portrait is not there to take it when the tile is now the lead or
    /// the lead is saving, so the keyboard goes to the tile's image.
    private func applyPending() {
        guard let pending = pendingFocus else { return }
        let tile = gallery.tiles.first { $0.id == pending.id }
        if case .useAs(let id) = pending, tile?.isLead == true || lead.isSaving {
            focused = .image(id)
        } else {
            focused = pending
        }
    }

    private func tileView(_ tile: ConclusionDetailPicture) -> some View {
        let tileID = "\(pageIdentifier).picture.\(tile.id)"
        let open = { navigation.go(to: tile.location) }
        return VStack(alignment: .leading, spacing: PVSpacing.space3) {
            GalleryTileView(
                projectDir: projectDir,
                tile: tile,
                look: look,
                fallbackMark: gallery.fallbackMark,
                fallbackColors: gallery.fallbackColors
            ) {
                Button(action: open) {
                    Color.clear
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                        .contentShape(Rectangle())
                }
                .buttonStyle(PVFocusRingButtonStyle(cornerRadius: PVRadius.sm))
                .focused($focused, equals: .image(tile.id))
                .pvActivatesOnReturn(when: look.opensOnReturn && focused == .image(tile.id), open)
                .help(look.hasOpenTooltip ? Text(gallery.openTip) : Text(verbatim: ""))
                .accessibilityLabel(tile.accessibilityLabel)
                .accessibilityHint(L10n.string(gallery.openTip))
                .accessibilityIdentifier(tileID)
            }
            if tile.isLead {
                PVBadge(gallery.choose.badge, tone: .accent, icon: .photo)
                    .accessibilityIdentifier("\(tileID).leadBadge")
            } else {
                PVButton(gallery.choose.useAs, variant: .ghost, size: .sm) {
                    Task {
                        _ = await lead.choose(tile, copy: gallery.choose)
                        focused = .image(tile.id)
                    }
                }
                .focused($focused, equals: .useAs(tile.id))
                .disabled(lead.isSaving)
                .fixedSize(horizontal: true, vertical: false)
                .accessibilityLabel(tile.useAsLabel)
                .accessibilityIdentifier("\(tileID).useAsLead")
            }
        }
        .frame(width: look.column, alignment: .leading)
        .pvGroupsAccessibility(look.groupsTile)
        .onAppear {
            if pendingFocus?.id == tile.id { applyPending() }
        }
    }
}

/// A keyboard stop in a tile: its image button, or Use as portrait.
private enum TileFocus: Hashable {
    case image(String)
    case useAs(String)

    var id: String {
        switch self {
        case .image(let id), .useAs(let id): id
        }
    }
}

private extension View {
    @ViewBuilder
    func pvGroupsAccessibility(_ groups: Bool) -> some View {
        if groups {
            accessibilityElement(children: .contain)
        } else {
            self
        }
    }
}

/// Every measure of a gallery tile. `.small` is the Event and Place Pictures' tile; `.large` is
/// the Person's Gallery tile.
@MainActor
enum GalleryTileLook {
    case small
    case large

    /// The crop's side.
    var crop: CGFloat {
        switch self {
        case .small: CachedThumbnail.pictureCrop
        case .large: 160
        }
    }

    /// The tile's column, which the caption keeps to.
    var column: CGFloat {
        switch self {
        case .small: 120
        case .large: 160
        }
    }

    /// The gap between columns.
    var columnGap: CGFloat {
        switch self {
        case .small: PVSpacing.space6
        case .large: PVSpacing.space8
        }
    }

    /// The gap between rows.
    var rowGap: CGFloat {
        switch self {
        case .small: PVSpacing.space7
        case .large: PVSpacing.space8
        }
    }

    /// Return opens the focused tile's Citation (the Person's Gallery).
    var opensOnReturn: Bool { self == .large }

    /// The image button carries the open tooltip (the Event and Place Pictures).
    var hasOpenTooltip: Bool { self == .small }

    /// The tile is one accessibility group of its image, badge and button (the Person's Gallery).
    var groupsTile: Bool { self == .large }
}

/// One picture: its crop over a caption of its citation. The gallery passes the button that opens
/// the Citation as `overlay` (drawn over the crop only). `look` holds every measure of the two
/// galleries.
struct GalleryTileView<Overlay: View>: View {
    let projectDir: String
    let tile: ConclusionDetailPicture
    let look: GalleryTileLook
    let fallbackMark: PVMarkKey
    let fallbackColors: PVThumbnail.Colors?
    private let overlay: Overlay

    init(
        projectDir: String,
        tile: ConclusionDetailPicture,
        look: GalleryTileLook,
        fallbackMark: PVMarkKey,
        fallbackColors: PVThumbnail.Colors? = nil,
        @ViewBuilder overlay: () -> Overlay
    ) {
        self.projectDir = projectDir
        self.tile = tile
        self.look = look
        self.fallbackMark = fallbackMark
        self.fallbackColors = fallbackColors
        self.overlay = overlay()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            crop
            caption
        }
    }

    @ViewBuilder
    private var crop: some View {
        let image = CachedThumbnail(
            projectDir: projectDir,
            relPath: tile.imageRelPath,
            size: look.crop,
            fallbackMark: fallbackMark,
            fallbackColors: fallbackColors
        )
        .accessibilityHidden(true)
        .overlay { overlay }
        switch look {
        case .small:
            image
        case .large:
            image.pvLiftHover(scale: 1.02)
        }
    }

    @ViewBuilder
    private var caption: some View {
        switch look {
        case .small:
            Text(verbatim: tile.citation.title)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(2)
                .truncationMode(.tail)
                .frame(maxWidth: look.column, alignment: .leading)
                .help(Text(verbatim: tile.citation.title))
                .accessibilityHidden(true)
        case .large:
            VStack(alignment: .leading, spacing: PVSpacing.space1) {
                Text(verbatim: tile.citation.title)
                    .font(tile.citation.titleIsRef
                        ? PVFont.mono(size: PVTypeScale.caption)
                        : PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(2)
                    .truncationMode(.tail)
                if !tile.citation.detail.isEmpty {
                    Text(verbatim: tile.citation.detail)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                        .truncationMode(.tail)
                }
            }
            .frame(maxWidth: look.column, alignment: .leading)
            .help(Text(verbatim: tile.citation.line))
            .accessibilityHidden(true)
        }
    }
}
