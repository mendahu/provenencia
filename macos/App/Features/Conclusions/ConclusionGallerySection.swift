import SwiftUI

/// A page's Gallery section: the heading with the count, then one 160pt tile per picture in a
/// wrapping grid, each captioned with its citation (S13-14, S13-D5). Kind-blind: its words come
/// from `gallery`, its ids from the page's identifier. A click on a tile's image opens the likeness
/// viewer when the kind has one (`gallery.viewer`, the Person's), else the Citation it was cut from (S13-15).
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
    @Environment(\.pvPageViewportSize) private var viewportSize
    @FocusState private var focused: TileFocus?
    /// Keeps the keyboard on a tile across a width change, a lazy grid dropping it and the viewer closing.
    @State private var keeper = FocusKeeper<TileFocus>()
    @State private var scrollTarget: String?
    @State private var scrollRequest = 0
    /// The picture the likeness viewer shows; nil while it is closed. Not a navigation place.
    @State private var viewing: ConclusionPictureViewerItem?

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
        .focusKeeper(
            $keeper,
            focus: $focused,
            isShown: { stop in gallery.tiles.contains { $0.id == stop.id } },
            fallback: { _ in nil },
            reveal: reveal
        )
        .pvSheet(item: $viewing) { item in
            // The shown picture has just left the gallery: draw its first tile until the item catches up.
            if let viewer = gallery.viewer,
               let step = ConclusionPictureStep(tiles: gallery.tiles, showing: item.shownID)
                   ?? gallery.tiles.first.flatMap({ ConclusionPictureStep(tiles: gallery.tiles, showing: $0.id) }) {
                ConclusionPictureViewer(
                    projectDir: projectDir,
                    pageIdentifier: pageIdentifier,
                    gallery: gallery,
                    viewer: viewer,
                    lead: lead,
                    step: step,
                    widthClass: widthClass,
                    paneHeight: viewportSize.height,
                    show: { viewing?.shownID = $0 },
                    open: { navigation.go(to: .provenance($0, $1)) }
                )
            }
        }
        // The viewer follows its gallery: the picture that leaves is replaced by its neighbour.
        .onChange(of: gallery.tiles) { old, new in
            viewing = viewing?.resolved(from: old, in: new)
        }
        // A close puts the keyboard back on the tile last shown, scrolled into view.
        .onChange(of: viewing) { old, new in
            if let old, new == nil {
                // An emptied gallery swaps this section for its EmptyState: there is no tile to focus.
                if !gallery.tiles.isEmpty {
                    FocusKeeper.hand(.image(old.shownID), keeper: $keeper, focus: $focused, reveal: reveal)
                }
            }
        }
        .onDisappear { lead.viewerClosed() }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.string(gallery.title))
    }

    /// Scrolls the stop's tile into view and names the stop that takes the keyboard: Use as portrait is
    /// not there to take it when the tile is now the lead or the lead is saving, so the keyboard goes to
    /// the tile's image.
    private func reveal(_ stop: TileFocus) -> TileFocus {
        scrollTarget = stop.id
        scrollRequest += 1
        let tile = gallery.tiles.first { $0.id == stop.id }
        if case .useAs(let id) = stop, tile?.isLead == true || lead.isSaving { return .image(id) }
        return stop
    }

    private func tileView(_ tile: ConclusionDetailPicture) -> some View {
        let tileID = "\(pageIdentifier).picture.\(tile.id)"
        let open = {
            if gallery.viewer != nil {
                viewing = ConclusionPictureViewerItem(shownID: tile.id)
            } else {
                navigation.go(to: tile.location)
            }
        }
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
            if keeper.pending?.id == tile.id {
                FocusKeeper.reclaim(keeper: $keeper, focus: $focused, reveal: reveal)
            }
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

    /// Return opens the focused tile's image (the Person's Gallery).
    var opensOnReturn: Bool { self == .large }

    /// The image button carries the open tooltip (the Event and Place Pictures).
    var hasOpenTooltip: Bool { self == .small }

    /// The tile is one accessibility group of its image, badge and button (the Person's Gallery).
    var groupsTile: Bool { self == .large }
}

/// One picture: its crop over a caption of its citation. The gallery passes the button that opens
/// the picture (the viewer or its Citation) as `overlay` (drawn over the crop only). `look` holds every measure of the two
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
