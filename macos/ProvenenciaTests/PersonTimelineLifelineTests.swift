import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct PersonTimelineLifelineTests {
    private typealias Data = PersonTimelinePreviewData
    private let en = Locale(identifier: "en_US")

    private func content(_ timeline: CatalogPersonTimeline, expanded: Set<PersonTimelineContent.Fold> = []) -> PersonTimelineContent {
        PersonTimelineContent(timeline: timeline, expanded: expanded, locale: en)
    }

    /// Frames stacked top to bottom summing to `height`, the same height each.
    private func frames(_ content: PersonTimelineContent, height: CGFloat) -> [String: CGRect] {
        let each = height / CGFloat(content.items.count)
        var result: [String: CGRect] = [:]
        for (k, item) in content.items.enumerated() {
            result[item.id] = CGRect(x: 80, y: CGFloat(k) * each, width: 600, height: each)
        }
        return result
    }

    private func lifeline(_ timeline: CatalogPersonTimeline, height: CGFloat, expanded: Set<PersonTimelineContent.Fold> = []) -> (PersonTimelineLifeline, PersonTimelineContent) {
        let content = content(timeline, expanded: expanded)
        return (PersonTimelineLifeline(content: content, itemFrames: frames(content, height: height)), content)
    }

    private func y(_ mark: PersonTimelineLifeline.Mark?) -> CGFloat {
        guard let mark else { return .nan }
        switch mark.shape {
        case .dot(let y), .cap(let y, _): return y
        case .bar(let from, _, _, _): return from
        }
    }

    @Test func pointsToScale() {
        let (line, _) = lifeline(Data.johann, height: 2730)
        #expect(line.length == 2730)
        #expect(abs(y(line.mark("1"))) < 0.001)
        // The burial's dot ends the line, a day short of its end.
        #expect(abs(y(line.mark("11")) - 2730) < 1)
        // 63.5 years over 2730pt: about 43pt a year, so one-year ticks.
        let perYear = 2730 / (1880 + 265.0 / 365 + 1.0 / 365 - (1817 + 70.0 / 365))
        #expect(abs(perYear - 43) < 1)
        let years = line.ticks.compactMap(\.year)
        #expect(years.first == 1818 && years.last == 1880 && years.count == 63)
        // Events 4 and 5 sit 0.67 years apart, about 29pt.
        #expect(abs((y(line.mark("5")) - y(line.mark("4"))) - 0.67 * perYear) < 1)
        // A day is a dot, a year a bar, a range a bar, "bef." an open start.
        if case .dot = line.mark("4")?.shape {} else { Issue.record("a day is a dot") }
        if case .bar(_, _, let openStart, _) = line.mark("3")?.shape { #expect(!openStart) } else { Issue.record("a year is a bar") }

        let (lise, _) = lifeline(Data.lise, height: 4000, expanded: [.dated])
        if case .bar(_, _, let openStart, _) = lise.mark("l119")?.shape { #expect(openStart) } else { Issue.record("bef. is a bar") }
        if case .bar(let from, let to, _, _) = lise.mark("l4")?.shape { #expect(to > from) } else { Issue.record("a range is a bar") }
        let spacing = zip(lise.ticks, lise.ticks.dropFirst()).map { $0.1.y - $0.0.y }
        #expect(spacing.allSatisfy { $0 >= 34 } && lise.ticks.allSatisfy { !$0.isQuarter })
        #expect(lise.ticks.compactMap(\.year).prefix(2) == [1801, 1802])

        let (thin, _) = lifeline(Data.lise, height: 1000, expanded: [.dated])
        #expect(thin.ticks.compactMap(\.year).prefix(2).map { $0 % 5 } == [0, 0])
    }

    @Test func placeholdersAndShortLives() {
        let (unknown, content) = lifeline(Data.unknownEnds, height: 2800)
        #expect(unknown.stubs.count == 2 && unknown.stubs.allSatisfy { abs(($0.upperBound - $0.lowerBound) - 40) < 0.001 })
        if case .cap(let birthY, let placeholder) = unknown.mark("birth")?.shape {
            #expect(birthY == 0 && placeholder)
            // 40pt before the baptism.
            #expect(abs(y(unknown.mark("2")) - 40) < 0.001)
        } else { Issue.record("birth placeholder is a cap") }
        // The burial sits 40pt lower than the scale alone gives.
        let (whole, _) = lifeline(Data.johann, height: 2800)
        _ = whole
        let spans = Dictionary(uniqueKeysWithValues: content.allSpans.map { ($0.id, $0.span) })
        let perYear = (2800 - 80) / (spans["11"]!.to - spans["2"]!.from)
        #expect(abs(y(unknown.mark("11")) - (CGFloat(spans["11"]!.from - spans["2"]!.from) * CGFloat(perYear) + 80)) < 0.5)
        // The death stub follows event 9's end.
        let nineEnd = CGFloat(spans["9"]!.to - spans["2"]!.from) * CGFloat(perYear) + 40
        if case .cap(let deathY, true) = unknown.mark("death")?.shape { #expect(abs(deathY - (nineEnd + 40)) < 0.5) } else { Issue.record("death placeholder") }

        // A funeral and no death: the death stub goes before the funeral.
        var timeline = Data.johann
        timeline.events = [
            Data.event("c", type: "census", when: .year(1875)),
            Data.event("f", type: "funeral", when: .year(1880)),
        ]
        timeline.birth = CatalogLifeMarker(besideEventId: "c")
        timeline.death = CatalogLifeMarker(besideEventId: "c")
        let (funeral, funeralContent) = lifeline(timeline, height: 800)
        #expect(funeralContent.items.map(\.id) == ["birth", "c", "death", "f"])
        #expect(funeral.mark("death").map { y($0) }! < y(funeral.mark("f")))

        // Only undated: two caps joined by one stretch, no tick.
        let (ida, _) = lifeline(Data.ida, height: 300)
        #expect(ida.ticks.isEmpty && ida.stubs == [0...300])
        #expect(ida.marks.map(\.id) == ["birth", "death"])

        // A short life: year labels and quarter ticks.
        let (ernst, _) = lifeline(Data.ernst, height: 1200)
        #expect(ernst.ticks.compactMap(\.year) == [1846, 1847, 1848])
        #expect(ernst.ticks.contains { $0.isQuarter } && ernst.ticks.filter(\.isQuarter).allSatisfy { $0.year == nil })

        // A range under a year is drawn as a year: one dated event, one point.
        var one = Data.ida
        one.events = [Data.event("d", type: "census", when: .day(1850, 6, 1))]
        let (single, _) = lifeline(one, height: 500)
        #expect(single.marks.contains { $0.id == "d" } && single.length == 500)

        // A recorded birth with no span: a filled cap on a stub of its own, no curve.
        var recorded = Data.johann
        recorded.events = recorded.events.map { event in
            var event = event
            if event.id == "1" { event.span = nil }
            return event
        }
        recorded.birth = CatalogLifeMarker(eventId: "1", besideEventId: "2")
        let (cap, capContent) = lifeline(recorded, height: 2800)
        if case .cap(_, let placeholder) = cap.mark("1")?.shape { #expect(!placeholder) } else { Issue.record("filled cap") }
        let capFrames = frames(capContent, height: 2800)
        let tether = cap.tethers(itemFrames: capFrames, visible: CGRect(x: 0, y: 0, width: 800, height: 2800)).first { $0.id == "1" }
        #expect(tether?.isDashed == true && tether?.from == tether?.to)

        // A taller line gives a new scale.
        let (tall, _) = lifeline(Data.johann, height: 5460)
        let (short, _) = lifeline(Data.johann, height: 2730)
        #expect(abs(y(tall.mark("5")) - 2 * y(short.mark("5"))) < 0.5)

        // Tethers only for items meeting `visible`; none without a viewport.
        let (johann, johannContent) = lifeline(Data.johann, height: 1400)
        let johannFrames = frames(johannContent, height: 1400)
        let visible = CGRect(x: 0, y: 0, width: 800, height: 200)
        let tethers = johann.tethers(itemFrames: johannFrames, visible: visible)
        let expected: [String] = johannContent.items
            .filter { item in item.eventRow != nil && johannFrames[item.id]!.intersects(visible) }
            .map { $0.id }
        #expect(tethers.map(\.id) == expected && !tethers.isEmpty && tethers.count < johannContent.items.count)
        #expect(tethers.allSatisfy { $0.to.x == 80 && $0.isDashed == ($0.from == $0.to) })
        #expect(johann.tethers(itemFrames: johannFrames, visible: CGRect(x: 0, y: 0, width: 800, height: 0)).isEmpty)
        // The undated rows' tethers are dashed dots.
        let all = johann.tethers(itemFrames: johannFrames, visible: CGRect(x: 0, y: 0, width: 800, height: 1400))
        #expect(all.first { $0.id == "12" }?.isDashed == true && all.first { $0.id == "4" }?.isDashed == false)
    }

    @Test func oneHundredTwentyEvents() {
        let (line, content) = lifeline(Data.lise, height: 6000, expanded: [.dated])
        let datedIDs = content.allSpans.map(\.id)
        #expect(datedIDs.count == 120)
        #expect(line.marks.map(\.id) == content.items.compactMap { $0.eventRow?.span != nil ? $0.id : nil })
        #expect(line.marks.count == 120)
        // Folded, the line still shows all 120 points.
        let (folded, foldedContent) = lifeline(Data.lise, height: 3000)
        #expect(foldedContent.items.contains(.more(.dated, count: 67)))
        #expect(folded.marks.count == 120)
    }

    @Test func tickCountIsCapped() {
        var timeline = Data.ida
        timeline.events = [
            Data.event("a", type: "census", when: .year(-9999)),
            Data.event("b", type: "census", when: .year(9999)),
        ]
        let (line, _) = lifeline(timeline, height: 4000)
        #expect(line.ticks.count <= 200 && !line.ticks.isEmpty)
        let years = line.ticks.compactMap(\.year)
        let step = years.count > 1 ? years[1] - years[0] : 0
        #expect(step > 50 && years.allSatisfy { $0 % step == 0 })
        #expect(years.first.map { $0 < 0 } == true)
        #expect(String(years.first ?? 0).hasPrefix("-"))

        let (ida, _) = lifeline(Data.ida, height: 300)
        #expect(ida.ticks.isEmpty && ida.marks.count == 2 && ida.stubs.count == 1)
    }
}
