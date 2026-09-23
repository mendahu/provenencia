import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite
struct ArtifactRegionGeometryTests {
    @Test func imageRectIgnoresCenteringPad() {
        let rect = ArtifactRegionGeometry.imageRect(
            documentSize: CGSize(width: 200, height: 300),
            imageSize: CGSize(width: 100, height: 80)
        )
        #expect(rect.origin.x == 50)
        #expect(rect.origin.y == 110)
        #expect(rect.size.width == 100)
        #expect(rect.size.height == 80)
    }

    @Test func normalizeClampsToUnitSquare() {
        let image = CGRect(x: 10, y: 20, width: 100, height: 200)
        #expect(ArtifactRegionGeometry.normalize(documentPoint: CGPoint(x: 10, y: 20), imageRect: image) == .zero)
        #expect(ArtifactRegionGeometry.normalize(documentPoint: CGPoint(x: 110, y: 220), imageRect: image) == CGPoint(x: 1, y: 1))
        let mid = ArtifactRegionGeometry.normalize(documentPoint: CGPoint(x: 60, y: 120), imageRect: image)
        #expect(abs(mid.x - 0.5) < 0.000_1)
        #expect(abs(mid.y - 0.5) < 0.000_1)
        let outside = ArtifactRegionGeometry.normalize(documentPoint: CGPoint(x: -50, y: 999), imageRect: image)
        #expect(outside.x == 0)
        #expect(outside.y == 1)
    }

    @Test func rectangleHasFourAxisAlignedPoints() {
        let points = ArtifactRegionGeometry.rectanglePoints(
            from: CGPoint(x: 0.2, y: 0.8),
            to: CGPoint(x: 0.6, y: 0.3)
        )
        #expect(points.count == 4)
        #expect(ArtifactRegionGeometry.isAxisAlignedRectangle(points))
        #expect(ArtifactRegionGeometry.isValidPolygon(points))
        #expect(ArtifactRegionKind.infer(fromNormalized: points) == .rectangle)
    }

    @Test func lShapesHaveSixPointsAndStayConstrained() {
        for kind: ArtifactRegionKind in [
            .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft,
        ] {
            let points = ArtifactRegionGeometry.lShapePoints(
                kind: kind,
                from: CGPoint(x: 0.1, y: 0.1),
                to: CGPoint(x: 0.7, y: 0.7)
            )
            #expect(points.count == 6)
            #expect(ArtifactRegionGeometry.isValidPolygon(points))
            #expect(ArtifactRegionKind.infer(fromNormalized: points) == kind)

            let moved = ArtifactRegionGeometry.moveLHandle(
                kind: kind,
                points: points,
                handle: 2,
                to: CGPoint(x: 0.9, y: 0.95)
            )
            #expect(moved.count == 6)
            #expect(ArtifactRegionKind.infer(fromNormalized: moved) == kind)
        }
    }

    @Test func circleRingHasThirtyTwoPoints() {
        let ring = ArtifactRegionGeometry.circleRing(
            center: CGPoint(x: 0.5, y: 0.5),
            radius: 0.2
        )
        #expect(ring.count == 32)
        #expect(ArtifactRegionGeometry.looksLikeCircleRing(ring))
        #expect(ArtifactRegionGeometry.isValidPolygon(ring))
        #expect(ArtifactRegionKind.infer(fromNormalized: ring) == .circle)

        let resized = ArtifactRegionGeometry.moveCircleRadius(
            points: ring,
            to: CGPoint(x: 0.8, y: 0.5)
        )
        #expect(resized.count == 32)
        #expect(ArtifactRegionGeometry.looksLikeCircleRing(resized))
    }

    @Test func freeformKeepsAtLeastThreeVertices() {
        let points = [
            CGPoint(x: 0.1, y: 0.1),
            CGPoint(x: 0.4, y: 0.1),
            CGPoint(x: 0.2, y: 0.4),
        ]
        #expect(ArtifactRegionGeometry.deleteFreeformVertex(points: points, index: 1) == nil)
        let four = points + [CGPoint(x: 0.1, y: 0.3)]
        #expect(ArtifactRegionGeometry.deleteFreeformVertex(points: four, index: 1)?.count == 3)
        let moved = ArtifactRegionGeometry.moveFreeformVertex(
            points: points,
            index: 2,
            to: CGPoint(x: 1.4, y: -0.2)
        )
        #expect(moved[2] == CGPoint(x: 1, y: 0))
    }
}
