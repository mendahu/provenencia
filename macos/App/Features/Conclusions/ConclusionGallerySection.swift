import SwiftUI

/// A page's Gallery section: the heading with the count, then one 160pt tile per picture in a
/// wrapping grid, each captioned with its citation (S13-14, S13-D5). Kind-blind: its words come
/// from `gallery`, its ids from the page's identifier. A click on a tile's image opens the Citation
/// it was cut from, until the likeness viewer takes the click (S13-15).
///
/// The Event and Place Pictures still draw `ConclusionDetailGallerySection`; the Event and Place
/// redesign adopts this section and deletes that one.
struct ConclusionGallerySection: View {
    let projectDir: String
    let pageIdentifier: String
    let gallery: ConclusionDetailGallery
    let lead: ConclusionLeadModel

    @Environment(WorkspaceNavigation.self) private var navigation
    @FocusState private var focusedID: String?

    private static let look = GalleryTileLook.large

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            PVSectionHeader(title: gallery.title, meta: String(gallery.tiles.count))
            LazyVGrid(
                columns: [GridItem(
                    .adaptive(minimum: Self.look.column, maximum: Self.look.column),
                    spacing: PVSpacing.space8,
                    alignment: .top
                )],
                alignment: .leading,
                spacing: PVSpacing.space8
            ) {
                ForEach(gallery.tiles) { tile in
                    tileView(tile)
                }
            }
            .padding(.top, PVSpacing.space6)
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(L10n.string(gallery.title))
    }

    private func tileView(_ tile: ConclusionDetailPicture) -> some View {
        let tileID = "\(pageIdentifier).picture.\(tile.id)"
        return VStack(alignment: .leading, spacing: PVSpacing.space3) {
            GalleryTileView(
                projectDir: projectDir,
                tile: tile,
                look: Self.look,
                fallbackMark: gallery.fallbackMark,
                colors: gallery.fallbackColors
            ) {
                Button {
                    navigation.go(to: tile.location)
                } label: {
                    Color.clear
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                        .contentShape(Rectangle())
                }
                .buttonStyle(PVFocusRingButtonStyle(cornerRadius: PVRadius.sm))
                .focused($focusedID, equals: tile.id)
                .pvActivatesOnReturn(when: focusedID == tile.id) { navigation.go(to: tile.location) }
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
                        if await lead.choose(tile, copy: gallery.choose) { focusedID = tile.id }
                    }
                }
                .disabled(lead.isSaving)
                .fixedSize(horizontal: true, vertical: false)
                .accessibilityLabel(tile.useAsLabel)
                .accessibilityIdentifier("\(tileID).useAsLead")
            }
        }
        .frame(width: Self.look.column, alignment: .leading)
        .accessibilityElement(children: .contain)
    }
}
