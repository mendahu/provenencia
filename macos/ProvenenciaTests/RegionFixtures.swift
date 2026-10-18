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

    /// A locator document: an artifact selector plus one region selector of `ringCount` points.
    static func locatorJSON(ringCount: Int) -> String {
        let points = ring(count: ringCount).map { ["x": Double($0.x), "y": Double($0.y)] }
        let document: [String: Any] = [
            "version": 1,
            "selectors": [
                ["type": "artifact"],
                ["type": "region", "unit": "normalized", "points": points],
            ],
        ]
        let data = try! JSONSerialization.data(withJSONObject: document)
        return String(decoding: data, as: UTF8.self)
    }
}
