import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct PersonTimelinePlacementTests {
    private typealias Data = PersonTimelinePreviewData
    private let en = Locale(identifier: "en_US")

    private func content(_ timeline: CatalogPersonTimeline) -> PersonTimelineContent {
        PersonTimelineContent(timeline: timeline, locale: en)
    }

    /// Every item `height` tall.
    private func heights(_ content: PersonTimelineContent, _ height: CGFloat) -> [String: CGFloat] {
        Dictionary(uniqueKeysWithValues: content.items.map { ($0.id, height) })
    }

    @Test func rowsSitByTheirPoints() throws {
        let johann = content(Data.johann)
        let placement = PersonTimelinePlacement(content: johann, heights: heights(johann, 120), viewportHeight: 900)
        let line = try #require(placement.lifeline)
        let tops = johann.items.map { placement.tops[$0.id]! }

        // No two items overlap, and the order is kept.
        for (top, next) in zip(tops, tops.dropFirst()) { #expect(next >= top + 120 - 0.001) }
        // The birth row's flag centre is on its point: 18 minus 4 + 4 + 10.
        #expect(placement.tops["1"] == 0)
        // The baptism is pushed directly under the birth, not on its own point.
        #expect(placement.tops["2"] == 120)
        // The 1835 census sits on its point: divider 1 + padding 20 + flag inset 4 + half the flag 10.
        let census = try #require(line.mark("3"))
        #expect(abs(placement.tops["3"]! - (census.pointY - 35)) < 0.001)
        #expect(tops[2] > tops[1] + 120)

        // The anchors are the tops plus the flags' centres.
        #expect(placement.anchors["1"] == CGFloat(18))
        #expect(placement.anchors["3"] == placement.tops["3"]! + 35)
        #expect(placement.anchors["undated"] == nil && placement.anchors["12"] == nil)

        // The median distance from a flag to its point is at most 24pt, or the scale is twice its base.
        let range = try #require(PersonTimelineLifeline.range(johann))
        let years = CGFloat(range.to - range.from)
        let base = max(max(300, 900 - 320) / years, min(CGFloat(johann.items.count) * 120 / years, 40))
        let distances = johann.items.compactMap { item -> CGFloat? in
            guard let anchor = placement.anchors[item.id], let mark = line.mark(item.id) else { return nil }
            return anchor - mark.pointY
        }.sorted()
        #expect(distances.allSatisfy { $0 >= -0.001 })
        #expect(distances[distances.count / 2] <= 24 || abs(placement.pointsPerYear - 2 * base) < 0.001)
        #expect(placement.pointsPerYear >= base - 0.001 && placement.pointsPerYear <= 2 * base + 0.001)

        // The space between the baptism (1817) and the census (1835) is a gap.
        #expect(placement.tops["3"]! - (placement.tops["2"]! + 120) > 100)
        #expect(placement.height >= placement.tops["11"]! + 120 && placement.height >= line.length)

        // With any height missing, the items stack and there is no line.
        var partial = heights(johann, 120)
        partial["5"] = nil
        let stacked = PersonTimelinePlacement(content: johann, heights: partial, viewportHeight: 900)
        #expect(stacked.lifeline == nil && stacked.tops["1"] == 0 && stacked.tops["2"] == 120 && stacked.tops["3"] == 240)
        #expect(stacked.anchors.isEmpty)
    }

    @Test func aBirthdayPairSitsTogether() throws {
        // Birth 12 Mar 1817 and baptism 16 Mar 1817: the baptism sits directly under the birth.
        let johann = content(Data.johann)
        let placement = PersonTimelinePlacement(content: johann, heights: heights(johann, 90), viewportHeight: 900)
        #expect(placement.tops["2"] == placement.tops["1"]! + 90)
    }

    @Test func scaleFloorAndRunPast() throws {
        let ernst = content(Data.ernst)
        let range = try #require(PersonTimelineLifeline.range(ernst))
        let years = CGFloat(range.to - range.from)
        let tall = PersonTimelinePlacement(content: ernst, heights: heights(ernst, 120), viewportHeight: 900)
        #expect(tall.pointsPerYear >= 580 / years - 0.001)
        let none = PersonTimelinePlacement(content: ernst, heights: heights(ernst, 120), viewportHeight: 0)
        #expect(none.pointsPerYear >= 300 / years - 0.001)
        #expect(tall.pointsPerYear >= none.pointsPerYear)

        // Only undated: the items stack from the top with a line of two caps.
        let ida = content(Data.ida)
        #expect(ida.items.map(\.id) == ["birth", "undated", "i1", "i2", "death"])
        let stacked = PersonTimelinePlacement(content: ida, heights: heights(ida, 100), viewportHeight: 900)
        #expect(ida.items.map { stacked.tops[$0.id] } == [0, 100, 200, 300, 400])
        #expect(stacked.lifeline?.length == 76 && stacked.height == 500)

        // A short list on a long life runs past the end of the line.
        let johann = content(Data.johann)
        let long = PersonTimelinePlacement(content: johann, heights: heights(johann, 400), viewportHeight: 900)
        let line = try #require(long.lifeline)
        #expect(long.height == long.tops["11"]! + 400 && long.height > line.length)
    }

    @Test func paddingsAreOneSource() {
        let johann = content(Data.johann)
        let items = johann.items
        #expect(PersonTimelinePlacement.topPadding(items[0], after: nil) == 4)
        #expect(!PersonTimelinePlacement.hasDivider(items[0], after: nil))
        #expect(PersonTimelinePlacement.topPadding(items[1], after: items[0]) == 20)
        #expect(PersonTimelinePlacement.hasDivider(items[1], after: items[0]))
        let header = items.first { $0.id == "undated" }!
        let before = items[items.firstIndex { $0.id == "undated" }! - 1]
        let after = items[items.firstIndex { $0.id == "undated" }! + 1]
        #expect(PersonTimelinePlacement.topPadding(header, after: before) == 22 && !PersonTimelinePlacement.hasDivider(header, after: before))
        #expect(PersonTimelinePlacement.topPadding(after, after: header) == 14 && !PersonTimelinePlacement.hasDivider(after, after: header))
        #expect(PersonTimelinePlacement.gutterWidth == 56)
    }
}
