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

    @Test func aRowAfterTheDeathKeepsItsDate() throws {
        // A burial "2011" listed after a death on 15 Dec 2011: its bar spans 2011 and its point is above the death's.
        let life = Data.timeline(
            [
                Data.event("d", type: "death", when: .day(2011, 12, 15, from: 2011.9534, to: 2011.9562)),
                Data.event("bur", type: "burial", when: .year(2011)),
            ],
            birth: CatalogLifeMarker(place: .before(eventId: "d")), death: CatalogLifeMarker(eventId: "d")
        )
        let c = content(life)
        let placement = PersonTimelinePlacement(content: c, heights: heights(c, 120), viewportHeight: 900)
        let line = try #require(placement.lifeline)
        #expect(c.items.map(\.id) == ["birth", "d", "bur"])
        guard case .bar? = line.mark("bur")?.shape else {
            Issue.record("the burial is a bar")
            return
        }
        #expect(line.mark("bur")!.pointY < line.mark("d")!.pointY)
        #expect(placement.tops["bur"]! >= placement.tops["d"]! + 120 - 0.001)
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

        // With a height missing, it takes the mean of the measured (120): every item is placed and the line exists.
        var partial = heights(johann, 120)
        partial["5"] = nil
        let estimated = PersonTimelinePlacement(content: johann, heights: partial, viewportHeight: 900)
        #expect(estimated.lifeline != nil)
        #expect(johann.items.allSatisfy { estimated.tops[$0.id] != nil })
        #expect(estimated == placement)
        #expect(PersonTimelinePlacement.estimatedHeight(heights: [:]) == 120)
        #expect(PersonTimelinePlacement.estimatedHeight(heights: ["a": 100, "b": 200]) == 150)
    }

    @Test func aFixedScaleMovesOnlyRowsBelow() throws {
        let johann = content(Data.johann)
        var measured = heights(johann, 120)
        let before = PersonTimelinePlacement(content: johann, heights: measured, viewportHeight: 900)
        let row = johann.items[3].id
        measured[row] = 720
        let after = PersonTimelinePlacement(content: johann, heights: measured, viewportHeight: 900, pointsPerYear: before.pointsPerYear)
        #expect(after.pointsPerYear == before.pointsPerYear)
        for (index, item) in johann.items.enumerated() {
            let was = before.tops[item.id]!, now = after.tops[item.id]!
            if index <= 3 { #expect(now == was) } else { #expect(now >= was) }
        }
    }

    @Test func narrowListStacksPeople() {
        // Regular: gutter 56, picture 72, one 300pt people column, four 16pt gaps and a 280pt text column.
        #expect(PersonTimelinePlacement.stacksPeople(itemsWidth: 771, widthClass: .regular))
        #expect(!PersonTimelinePlacement.stacksPeople(itemsWidth: 772, widthClass: .regular))
        // Wide: the people column is three cards and two gaps, 930pt.
        #expect(PersonTimelinePlacement.stacksPeople(itemsWidth: 1401, widthClass: .wide))
        #expect(!PersonTimelinePlacement.stacksPeople(itemsWidth: 1402, widthClass: .wide))
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

    @Test func aDateEditRescales() {
        let johann = content(Data.johann)
        let kept = PersonTimelinePlacement.KeptScale(content: johann, viewportHeight: 900, pointsPerYear: 37)
        #expect(kept.pointsPerYear(for: johann, viewportHeight: 900) == 37)
        #expect(kept.pointsPerYear(for: johann, viewportHeight: 800) == nil)

        // The last dated event, the burial, moved 50 years later under the same ids: the line's range changes.
        var moved = Data.johann
        moved.events = moved.events.map { event in
            var event = event
            if event.id == "11", var span = event.span {
                span.from += 50
                span.to += 50
                event.span = span
            }
            return event
        }
        let later = content(moved)
        #expect(later.items.map(\.id) == johann.items.map(\.id))
        #expect(kept.pointsPerYear(for: later, viewportHeight: 900) == nil)

        #expect(kept.pointsPerYear(for: content(Data.unknownEnds), viewportHeight: 900) == nil)
    }
}
