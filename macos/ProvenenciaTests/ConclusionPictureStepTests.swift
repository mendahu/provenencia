import Foundation
import Testing
@testable import Provenencia

@Suite
struct ConclusionPictureStepTests {
    private typealias D = ReconciledValueDisplayTests

    private func tiles(_ ids: [String]) -> [ConclusionDetailPicture] {
        var detail = CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            fields: [], memberCount: 1, summary: nil, derivedRows: []
        )
        detail.pictures = ids.map { D.picture($0, path: "objects/\($0)/x.png", title: "Source \($0)") }
        return ConclusionDetailGallery.tiles(
            from: detail,
            tileLabel: L10n.Conclusions.likenessLabel,
            leadTileLabel: L10n.Conclusions.likenessTileLabelLead,
            useAsLabel: L10n.Conclusions.useAsPortraitLabel,
            chosenBody: L10n.Conclusions.toastPortraitChosenBody
        )
    }

    private let seven = ["a", "b", "c", "d", "e", "f", "g"]

    @Test func stepsThroughTheGalleryWithoutWrapping() throws {
        let all = tiles(seven)
        let middle = try #require(ConclusionPictureStep(tiles: all, showing: "b"))
        #expect(middle.index == 1 && middle.count == 7)
        #expect(middle.previousID == "a" && middle.nextID == "c")
        #expect(middle.tile.id == "b")

        let first = try #require(ConclusionPictureStep(tiles: all, showing: "a"))
        #expect(first.previousID == nil && first.nextID == "b")
        let last = try #require(ConclusionPictureStep(tiles: all, showing: "g"))
        #expect(last.previousID == "f" && last.nextID == nil)

        let only = try #require(ConclusionPictureStep(tiles: tiles(["a"]), showing: "a"))
        #expect(only.previousID == nil && only.nextID == nil && only.count == 1)

        #expect(ConclusionPictureStep(tiles: all, showing: "zzz") == nil)
    }

    @Test func holdsTheShownTileAndItsNeighbours() throws {
        let all = tiles(seven)
        #expect(try #require(ConclusionPictureStep(tiles: all, showing: "d")).heldIDs == ["c", "d", "e"])
        #expect(try #require(ConclusionPictureStep(tiles: all, showing: "a")).heldIDs == ["a", "b"])
        #expect(try #require(ConclusionPictureStep(tiles: all, showing: "g")).heldIDs == ["f", "g"])
        #expect(try #require(ConclusionPictureStep(tiles: tiles(["a"]), showing: "a")).heldIDs == ["a"])
    }

    @Test func followsTheGalleryWhenTheShownLikenessLeaves() {
        let all = tiles(seven)
        let withoutE = tiles(seven.filter { $0 != "e" })
        let withoutG = tiles(seven.filter { $0 != "g" })

        let onE = ConclusionPictureViewerItem(shownID: "e")
        #expect(onE.resolved(from: all, in: withoutE)?.shownID == "f")
        #expect(ConclusionPictureViewerItem(shownID: "g").resolved(from: all, in: withoutG)?.shownID == "f")
        #expect(ConclusionPictureViewerItem(shownID: "a").resolved(from: tiles(["a"]), in: []) == nil)

        let unchanged = onE.resolved(from: all, in: all)
        #expect(unchanged == onE)
        #expect(unchanged?.id == onE.id)
        #expect(onE.resolved(from: all, in: withoutE)?.id == onE.id)
    }

    @Test func framesCapAtHalfThePixelsAndFitTheBand() {
        func frame(_ width: CGFloat, _ height: CGFloat, _ widthClass: PVWidthClass, band: CGFloat) -> CGSize {
            ConclusionPictureImage.frame(pixelWidth: width, pixelHeight: height, widthClass: widthClass, bandHeight: band)
        }
        for widthClass in [PVWidthClass.regular, .wide] {
            #expect(frame(300, 380, widthClass, band: 440) == CGSize(width: 150, height: 190))
        }

        let regular = frame(1400, 2048, .regular, band: 440)
        #expect(regular.height == 400 && abs(regular.width - 400 * 1400 / 2048) < 0.01)
        let wide = frame(1400, 2048, .wide, band: 950)
        #expect(wide.height == 910 && abs(wide.width - 910 * 1400 / 2048) < 0.01)

        #expect(frame(4000, 1000, .regular, band: 440) == CGSize(width: 728, height: 182))
        #expect(frame(4000, 1000, .wide, band: 950) == CGSize(width: 1128, height: 282))

        #expect(ConclusionPictureImage.bandHeight(widthClass: .wide, paneHeight: 1000) == 680)
        #expect(ConclusionPictureImage.bandHeight(widthClass: .wide, paneHeight: 1440) == 950)
        #expect(ConclusionPictureImage.bandHeight(widthClass: .wide, paneHeight: 500) == 320)
        #expect(ConclusionPictureImage.bandHeight(widthClass: .regular, paneHeight: 1440) == 440)
    }
}
