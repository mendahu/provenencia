import SwiftUI

/// One picture tile in a page's gallery: the image, its citation, the composer location its
/// Citation opens at, and the Strings the section and the lead model speak for it (S10-35, S14-02).
struct ConclusionDetailPicture: Equatable, Identifiable {
    var imageRelPath: String
    /// The picture's record: its earliest gallery Observation, `kept` or `weak`. The viewer reads its
    /// provenance from it; the id and the Citation location follow from it.
    var outcome: CatalogReconcilerOutcome
    /// The tile the header leads with wears the lead Badge; exactly one tile.
    var isLead: Bool
    /// Spoken label of the tile.
    var accessibilityLabel: String
    /// Spoken label of the button that makes this tile the lead.
    var useAsLabel: String
    /// Toast body after this tile is chosen as the lead.
    var chosenToastBody: String
    /// Spoken label of the viewer's image: the tile's position, how many there are, its citation
    /// line and whether it leads; the citation line alone for a kind with no viewer.
    var viewerLabel: String = ""

    /// Where the picture comes from, as the caption and the spoken label read it.
    var citation: CitationLineDisplay.Parts { CitationLineDisplay.parts(outcome) }
    /// The Observation id.
    var id: String { outcome.observationID }
    /// The composer location of the Citation the picture was cut from.
    var location: WorkspaceLocation { .composer(for: outcome) }
}

/// A page's gallery, one slot on the body for every kind: the section's words, the mark a missing
/// image falls back to, the tiles (earliest recorded first), and the copy of choosing a lead
/// (S14-02 D-1, D-2).
struct ConclusionDetailGallery: Equatable {
    var title: LocalizedStringResource
    var openTip: LocalizedStringResource
    var fallbackMark: PVMarkKey
    /// The tint of the fallback mark (a Person's gender colors); nil for none.
    var fallbackColors: PVThumbnail.Colors? = nil
    var tiles: [ConclusionDetailPicture]
    var choose: LeadCopy
    /// The static words of the likeness viewer; nil for a kind with no viewer, whose tile image
    /// opens the Citation (S13-15).
    var viewer: ViewerCopy? = nil

    /// The tile the header leads with.
    var lead: ConclusionDetailPicture? { tiles.first(where: \.isLead) }
    /// The header leads with a picture other than the earliest, so its menu offers the earliest
    /// (S10-39, user 2026-10-17).
    var offersEarliest: Bool { tiles.first.map { !$0.isLead } ?? false }

    /// The static words of the viewer; the per-tile spoken line lives on the tile (`viewerLabel`).
    struct ViewerCopy: Equatable {
        /// The finished title of the panel, with the owner's name in it.
        var title: String
        var previous: LocalizedStringResource
        var next: LocalizedStringResource
        var close: LocalizedStringResource

        static func likeness(title: String) -> ViewerCopy {
            ViewerCopy(
                title: title,
                previous: L10n.Conclusions.likenessViewerPrevious,
                next: L10n.Conclusions.likenessViewerNext,
                close: L10n.Conclusions.likenessViewerClose
            )
        }
    }

    /// The static copy of choosing a lead; the per-tile Strings live on the tile.
    struct LeadCopy: Equatable {
        var badge: LocalizedStringResource
        var useAs: LocalizedStringResource
        var useEarliest: LocalizedStringResource
        var menuTitle: LocalizedStringResource
        var menuHint: LocalizedStringResource
        var toastTitle: String
        var toastEarliestBody: String

        static let personPortrait = LeadCopy(
            badge: L10n.Conclusions.portraitBadge,
            useAs: L10n.Conclusions.useAsPortrait,
            useEarliest: L10n.Conclusions.useEarliestLikeness,
            menuTitle: L10n.Conclusions.portraitMenuTitle,
            menuHint: L10n.Conclusions.portraitMenuHint,
            toastTitle: L10n.Conclusions.toastPortraitUpdatedTitle,
            toastEarliestBody: L10n.Conclusions.toastPortraitEarliestBody
        )
        static let eventThumbnail = LeadCopy(
            badge: L10n.Conclusions.thumbnailBadge,
            useAs: L10n.Conclusions.useAsThumbnail,
            useEarliest: L10n.Conclusions.useEarliestPicture,
            menuTitle: L10n.Conclusions.eventThumbnailMenuTitle,
            menuHint: L10n.Conclusions.thumbnailMenuHint,
            toastTitle: L10n.Conclusions.toastThumbnailUpdatedTitle,
            toastEarliestBody: L10n.Conclusions.toastThumbnailEarliestBody
        )
        /// A Place's thumbnail copy is an Event's with its own menu title.
        static let placeThumbnail: LeadCopy = {
            var copy = eventThumbnail
            copy.menuTitle = L10n.Conclusions.placeThumbnailMenuTitle
            return copy
        }()
    }

    /// The gallery of an Event or a Place: the shared section words and tile Strings, with the
    /// kind's fallback mark and lead copy.
    static func pictures(
        from detail: CatalogConclusionDetail,
        fallbackMark: PVMarkKey,
        choose: LeadCopy
    ) -> ConclusionDetailGallery {
        ConclusionDetailGallery(
            title: L10n.Conclusions.picturesSection,
            openTip: L10n.Conclusions.pictureOpenTip,
            fallbackMark: fallbackMark,
            tiles: tiles(
                from: detail,
                tileLabel: L10n.Conclusions.pictureFrom,
                leadTileLabel: L10n.Conclusions.pictureTileLabelLead,
                useAsLabel: L10n.Conclusions.useAsThumbnailLabel,
                chosenBody: L10n.Conclusions.toastThumbnailChosenBody
            ),
            choose: choose
        )
    }

    /// The tiles of `detail.pictures`, in wire order. `tileLabel` and `leadTileLabel` take the
    /// picture's citation line; `useAsLabel` and `chosenBody` its Source title (the Observation ref
    /// when the title is empty); `viewerLabel` its position, the count, its citation line and whether it leads. Each
    /// returns the kind's wording.
    static func tiles(
        from detail: CatalogConclusionDetail,
        tileLabel: (String) -> String,
        leadTileLabel: (String) -> String,
        useAsLabel: (String) -> String,
        chosenBody: (String) -> String,
        viewerLabel: (_ position: Int, _ count: Int, _ citationLine: String, _ leads: Bool) -> String = { _, _, line, _ in line }
    ) -> [ConclusionDetailPicture] {
        detail.pictures.enumerated().map { index, picture in
            let citation = CitationLineDisplay.parts(picture.record)
            let source = citation.title
            return ConclusionDetailPicture(
                imageRelPath: picture.imageRelPath,
                outcome: picture.record,
                isLead: picture.leads,
                accessibilityLabel: picture.leads ? leadTileLabel(citation.line) : tileLabel(citation.line),
                useAsLabel: useAsLabel(source),
                chosenToastBody: chosenBody(source),
                viewerLabel: viewerLabel(index + 1, detail.pictures.count, citation.line, picture.leads)
            )
        }
    }
}
