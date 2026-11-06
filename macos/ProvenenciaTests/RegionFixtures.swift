import CoreGraphics
import Foundation
@testable import Provenencia

/// Region geometry for tests: rings of any point count and the locator JSON that holds one.
enum RegionFixtures {
    /// `count` points on a circle of radius 0.4 around (0.5, 0.5), counter-clockwise from angle 0.
    static func ring(count: Int) -> [CGPoint] {
        (0..<count).map { index in
            let angle = 2 * Double.pi * Double(index) / Double(count)
            return CGPoint(x: 0.5 + 0.4 * cos(angle), y: 0.5 + 0.4 * sin(angle))
        }
    }

    /// A small square in the top-left quarter of the unit square.
    static let smallSquare = [
        CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.4, y: 0.1), CGPoint(x: 0.4, y: 0.4), CGPoint(x: 0.1, y: 0.4),
    ]

    /// A 32-point polygon whose radius alternates 0.4 and 0.25: the point count of a circle, not a circle.
    static let bumpyRing: [CGPoint] = (0..<ArtifactRegionGeometry.circleRingCount).map { index in
        let angle = 2 * Double.pi * Double(index) / Double(ArtifactRegionGeometry.circleRingCount)
        let radius = index.isMultiple(of: 2) ? 0.4 : 0.25
        return CGPoint(x: 0.5 + radius * cos(angle), y: 0.5 + radius * sin(angle))
    }

    /// A six-point polygon that is not an L.
    static let hexagon = [
        CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.5, y: 0.1), CGPoint(x: 0.6, y: 0.3),
        CGPoint(x: 0.5, y: 0.5), CGPoint(x: 0.1, y: 0.5), CGPoint(x: 0.05, y: 0.3),
    ]

    /// A locator document: an artifact selector plus one region selector of `ringCount` points.
    static func locatorJSON(ringCount: Int) -> String {
        locatorJSON(points: ring(count: ringCount))
    }

    /// A locator document: an artifact selector, a page selector when `page` is set, and one
    /// region selector of `points`.
    static func locatorJSON(points: [CGPoint], page: Int? = nil) -> String {
        var selectors: [[String: Any]] = [["type": "artifact"]]
        if let page { selectors.append(["type": "page", "artifact_page": page]) }
        selectors.append([
            "type": "region", "unit": "normalized",
            "points": points.map { ["x": Double($0.x), "y": Double($0.y)] },
        ])
        let data = try! JSONSerialization.data(withJSONObject: ["version": 1, "selectors": selectors])
        return String(decoding: data, as: UTF8.self)
    }
}
