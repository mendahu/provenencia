import SwiftUI
import Testing
@testable import Provenencia

@MainActor
struct SourcePageOverviewLayoutTests {
    /// Arranges four subviews of these heights at `width`.
    private func arrange(
        width: CGFloat, widthClass: PVWidthClass, errorHeight: CGFloat
    ) -> (frames: [CGRect], height: CGFloat) {
        let heights: [CGFloat] = [errorHeight, 100, 140, 80]
        return SourcePageOverviewLayout.arrange(width: width, widthClass: widthClass) { index, width in
            CGSize(width: width, height: heights[index])
        }
    }

    @Test func regularPlacesDescriptionBesideMetadataThenTheRest() {
        let frames = arrange(width: 1038, widthClass: .regular, errorHeight: 30).frames
        let (error, body, side, rest) = (frames[0], frames[1], frames[2], frames[3])
        let half: CGFloat = (1038 - PVSpacing.space11) / 2
        #expect(error.minX == 0 && error.minY == 0 && error.width == 1038)
        #expect(body.width == half && side.width == half)
        #expect(body.minY == side.minY && body.minY == 30 + PVSpacing.space9)
        #expect(side.minX == half + PVSpacing.space11)
        // Both are as tall as the taller, so Credibility sits at the bottom of its column.
        #expect(body.height == 140 && side.height == 140)
        #expect(rest.minX == 0 && rest.width == 1038 && rest.minY == body.maxY + PVSpacing.space9)
    }

    @Test func anEmptyErrorAddsNoGap() {
        let frames = arrange(width: 1038, widthClass: .regular, errorHeight: 0).frames
        #expect(frames[1].minY == 0)
    }

    @Test func wideStacksDescriptionAndCredibilityAboveMetadataInTheSideColumn() {
        let group = PVPageColumnsMetrics.split(
            available: 2296 - 2 * PVSpacing.gutterPage, sideMax: PVSpacing.widthSecondaryMax
        )
        let frames = arrange(width: group.group, widthClass: .wide, errorHeight: 30).frames
        let (error, body, side, rest) = (frames[0], frames[1], frames[2], frames[3])
        #expect(group.main == 1180 && group.side == 480)
        // The main column: the error, then graphs, Artifacts and Notes.
        #expect(error.minX == 0 && error.minY == 0 && error.width == 1180)
        #expect(rest.minX == 0 && rest.width == 1180 && rest.minY == 30 + PVSpacing.space9)
        // The side column: Description and Credibility, then Metadata under them.
        let sideX = 1180 + PVSpacing.space9
        #expect(body.minX == sideX && body.minY == 0 && body.width == 480 && body.height == 100)
        #expect(side.minX == sideX && side.width == 480 && side.minY == body.maxY + PVSpacing.space11)
    }

    @Test func wideBelowItsFullWidthShrinksTheMainColumnNotTheSide() {
        // A pane just over the Wide breakpoint: the side column keeps its full width.
        let width = PVSpacing.widthWideBreakpoint - 2 * PVSpacing.gutterPage
        let frames = arrange(width: width, widthClass: .wide, errorHeight: 0).frames
        let (body, side, rest) = (frames[1], frames[2], frames[3])
        let main = width - PVSpacing.space9 - PVSpacing.widthSecondaryMax
        #expect(rest.minX == 0 && rest.width == main)
        #expect(body.width == PVSpacing.widthSecondaryMax && side.width == PVSpacing.widthSecondaryMax)
        #expect(body.minX == main + PVSpacing.space9 && side.maxX == width)
    }
}
