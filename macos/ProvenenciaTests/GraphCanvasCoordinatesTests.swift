import Foundation
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasCoordinatesTests {
    @Test func detailBandKeepsTheCurrentFaceBetweenCutoffs() {
        let band = GraphCanvasDetailBand(cut: EvidenceGraphCardDetail.cut)
        #expect(abs(band.overviewBelow - 0.675) < 0.000_001)
        #expect(abs(band.readingAbove - 0.725) < 0.000_001)
        let inside = EvidenceGraphCardDetail.cut
        #expect(band.next(.reading, magnification: band.overviewBelow - 0.01) == .overview)
        #expect(band.next(.reading, magnification: inside) == .reading)
        #expect(band.next(.overview, magnification: inside) == .overview)
        #expect(band.next(.overview, magnification: band.readingAbove + 0.01) == .reading)
        #expect(band.next(.overview, magnification: GraphCanvasCamera.minMagnification) == .overview)
        #expect(band.initial(magnification: GraphCanvasCamera.minMagnification) == .overview)
        #expect(band.initial(magnification: inside) == .reading)
    }

    @Test func clampMagnificationRespectsBounds() {
        #expect(GraphCanvasCoordinates.clampMagnification(0.1) == GraphCanvasCamera.minMagnification)
        #expect(GraphCanvasCoordinates.clampMagnification(10) == GraphCanvasCamera.maxMagnification)
        #expect(GraphCanvasCoordinates.clampMagnification(1.5) == GraphCanvasCamera.maxMagnification)
        #expect(GraphCanvasCoordinates.clampMagnification(1.1) == 1.1)
        #expect(GraphCanvasCoordinates.clampMagnification(0.3, minimum: 0.5) == 0.5)
    }

    @Test func magnificationFloorKeepsTheDocumentCoveringTheWindow() {
        let document = CGSize(width: 4_000, height: 4_000)
        #expect(
            GraphCanvasCoordinates.magnificationThatFills(
                documentSize: document,
                viewportSize: CGSize(width: 2_000, height: 1_200)
            ) == 0.5
        )
        #expect(
            GraphCanvasCoordinates.magnificationThatFills(
                documentSize: document,
                viewportSize: CGSize(width: 800, height: 600)
            ) == GraphCanvasCamera.minMagnification
        )
        #expect(
            GraphCanvasCoordinates.magnificationThatFills(
                documentSize: document,
                viewportSize: CGSize(width: 6_000, height: 5_000)
            ) == GraphCanvasCamera.maxMagnification
        )
    }

    @Test func cameraInitClampsMagnification() {
        let low = GraphCanvasCamera(magnification: 0.01, contentOffset: .zero)
        #expect(low.magnification == GraphCanvasCamera.minMagnification)
        let high = GraphCanvasCamera(magnification: 99, contentOffset: .zero)
        #expect(high.magnification == GraphCanvasCamera.maxMagnification)
    }

    @Test func identityMagnificationMapsViewToContentWithOffset() {
        let camera = GraphCanvasCamera(magnification: 1, contentOffset: CGPoint(x: 100, y: 40))
        let content = GraphCanvasCoordinates.contentPoint(
            fromViewPoint: CGPoint(x: 20, y: 10),
            camera: camera
        )
        #expect(content == CGPoint(x: 120, y: 50))
    }

    @Test func magnifiedCameraScalesViewDeltaInContent() {
        let camera = GraphCanvasCamera(magnification: 1.25, contentOffset: CGPoint(x: 10, y: 20))
        let content = GraphCanvasCoordinates.contentPoint(
            fromViewPoint: CGPoint(x: 40, y: 60),
            camera: camera
        )
        #expect(content == CGPoint(x: 42, y: 68))
    }

    @Test func viewAndContentRoundTrip() {
        let camera = GraphCanvasCamera(magnification: 1.25, contentOffset: CGPoint(x: 80, y: 12))
        let original = CGPoint(x: 200, y: 90)
        let view = GraphCanvasCoordinates.viewPoint(fromContentPoint: original, camera: camera)
        let back = GraphCanvasCoordinates.contentPoint(fromViewPoint: view, camera: camera)
        #expect(abs(back.x - original.x) < 0.0001)
        #expect(abs(back.y - original.y) < 0.0001)
    }

    @Test func halfMagnificationDoublesViewDeltaInContent() {
        let camera = GraphCanvasCamera(magnification: 0.5, contentOffset: .zero)
        let content = GraphCanvasCoordinates.contentPoint(
            fromViewPoint: CGPoint(x: 50, y: 25),
            camera: camera
        )
        #expect(content == CGPoint(x: 100, y: 50))
    }

    @Test func clampedContentOriginStaysInsideDocument() {
        let clamped = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: CGPoint(x: -40, y: 500),
            documentSize: CGSize(width: 400, height: 300),
            visibleSize: CGSize(width: 200, height: 100)
        )
        #expect(clamped == CGPoint(x: 0, y: 200))
    }

    @Test func clampedContentOriginZeroWhenDocumentFits() {
        let clamped = GraphCanvasCoordinates.clampedContentOrigin(
            proposed: CGPoint(x: 50, y: 50),
            documentSize: CGSize(width: 100, height: 80),
            visibleSize: CGSize(width: 200, height: 100)
        )
        #expect(clamped == .zero)
    }

    @Test func trackpadStyleScrollClassification() {
        #expect(
            GraphCanvasScrollInput.isTrackpadStyleScroll(
                hasPreciseScrollingDeltas: true,
                hasPhase: false,
                hasMomentumPhase: false
            )
        )
        #expect(
            GraphCanvasScrollInput.isTrackpadStyleScroll(
                hasPreciseScrollingDeltas: false,
                hasPhase: true,
                hasMomentumPhase: false
            )
        )
        #expect(
            GraphCanvasScrollInput.isTrackpadStyleScroll(
                hasPreciseScrollingDeltas: false,
                hasPhase: false,
                hasMomentumPhase: true
            )
        )
        #expect(
            !GraphCanvasScrollInput.isTrackpadStyleScroll(
                hasPreciseScrollingDeltas: false,
                hasPhase: false,
                hasMomentumPhase: false
            )
        )
    }

    @Test func screenPixelWidthTracksMagnification() {
        #expect(GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: 1) == 1)
        #expect(GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: 0.25) == 4)
        #expect(GraphCanvasCoordinates.screenPixelDocumentWidth(magnification: 0.5) == 2)
    }

    @Test func gridPatternFadesOnlyWhenACellIsAboutOnePixel() {
        let spacing = GraphCanvasGridMapping.defaultSpacing
        #expect(GraphCanvasCoordinates.gridPatternOpacity(spacing: spacing, magnification: 1) == 1)
        #expect(GraphCanvasCoordinates.gridPatternOpacity(spacing: spacing, magnification: 0.25) == 1)
        #expect(GraphCanvasCoordinates.gridPatternOpacity(spacing: 40, magnification: 1.5 / 40) == 0.5)
        #expect(GraphCanvasCoordinates.gridPatternOpacity(spacing: 40, magnification: 1 / 40) == 0)
        #expect(GraphCanvasCoordinates.gridPatternOpacity(spacing: 40, magnification: 0.01) == 0)
    }
}
