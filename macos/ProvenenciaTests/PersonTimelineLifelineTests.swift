import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct PersonTimelineLifelineTests {
    private typealias Data = PersonTimelinePreviewData
    private let en = Locale(identifier: "en_US")

    private func content(_ timeline: CatalogPersonTimeline, unfolded: [PersonTimelineContent.Fold] = []) -> PersonTimelineContent {
        var folds = ConclusionFoldState<PersonTimelineContent.Fold>()
        // A hundred pages shows any list here.
        for fold in unfolded { folds.reveal(fold, pages: 100) }
        return PersonTimelineContent(timeline: timeline, folds: folds, locale: en)
    }

    private func lifeline(_ timeline: CatalogPersonTimeline, ppy: CGFloat, unfolded: [PersonTimelineContent.Fold] = []) -> (PersonTimelineLifeline, PersonTimelineContent) {
        let content = content(timeline, unfolded: unfolded)
        return (PersonTimelineLifeline(content: content, pointsPerYear: ppy), content)
    }

    private func y(_ mark: PersonTimelineLifeline.Mark?) -> CGFloat {
        guard let mark else { return .nan }
        switch mark.shape {
        case .dot(let y), .cap(let y, _): return y
        case .bar(let from, _, _, _): return from
        }
    }

    @Test func pointsToScale() throws {
        let (line, content) = lifeline(Data.johann, ppy: 43)
        let range = try #require(PersonTimelineLifeline.range(content))
        #expect(abs(line.length - (18 + CGFloat(range.to - range.from) * 43 + 18)) < 0.001)
        #expect(abs(y(line.mark("1")) - 18) < 0.001)
        // The burial's dot ends the line, a day short of its end.
        #expect(abs(y(line.mark("11")) - (line.length - 18)) < 1)
        let years = line.ticks.compactMap(\.year)
        #expect(years.first == 1818 && years.last == 1880 && years.count == 63)
        // Events 4 and 5 sit 0.67 years apart, about 29pt.
        #expect(abs((y(line.mark("5")) - y(line.mark("4"))) - 0.67 * 43) < 1)
        // A day is a dot, a year a bar, a range a bar, "bef." an open start.
        if case .dot = line.mark("4")?.shape {} else { Issue.record("a day is a dot") }
        if case .bar(_, _, let openStart, _) = line.mark("3")?.shape { #expect(!openStart) } else { Issue.record("a year is a bar") }

        let (lise, _) = lifeline(Data.lise, ppy: 41, unfolded: [.dated])
        if case .bar(_, _, let openStart, _) = lise.mark("l119")?.shape { #expect(openStart) } else { Issue.record("bef. is a bar") }
        if case .bar(let from, let to, _, _) = lise.mark("l4")?.shape { #expect(to > from) } else { Issue.record("a range is a bar") }
        let spacing = zip(lise.ticks, lise.ticks.dropFirst()).map { $0.1.y - $0.0.y }
        #expect(spacing.allSatisfy { $0 >= 34 } && lise.ticks.allSatisfy { !$0.isQuarter })
        #expect(lise.ticks.compactMap(\.year).prefix(2) == [1801, 1802])

        let (thin, _) = lifeline(Data.lise, ppy: 10, unfolded: [.dated])
        #expect(thin.ticks.compactMap(\.year).prefix(2).map { $0 % 5 } == [0, 0])

        // A tether leaves a closed bar at its middle, an open bar 8pt in from its known end.
        if case .bar(let from, let to, _, _) = line.mark("3")?.shape { #expect(line.mark("3")?.pointY == (from + to) / 2) }
        if case .bar(_, let to, true, _) = lise.mark("l119")?.shape { #expect(lise.mark("l119")?.pointY == to - 8) } else { Issue.record("bef. is an open start") }

        // A bigger scale scales the line's points from its start.
        let (double, _) = lifeline(Data.johann, ppy: 86)
        #expect(abs((y(double.mark("5")) - 18) - 2 * (y(line.mark("5")) - 18)) < 0.001)
    }

    @Test func placeholdersAndShortLives() throws {
        let (unknown, content) = lifeline(Data.unknownEnds, ppy: 43)
        let range = try #require(PersonTimelineLifeline.range(content))
        #expect(abs(unknown.length - (18 + 40 + 40 + CGFloat(range.to - range.from) * 43 + 18)) < 0.001)
        #expect(unknown.stubs.count == 2 && unknown.stubs.allSatisfy { abs(($0.upperBound - $0.lowerBound) - 40) < 0.001 })
        if case .cap(let birthY, let placeholder) = unknown.mark("birth")?.shape {
            #expect(birthY == 18 && placeholder)
            // 40pt before the baptism, the first dated point.
            #expect(abs(y(unknown.mark("2")) - 58) < 0.001)
        } else { Issue.record("birth placeholder is a cap") }
        // The burial sits 40pt lower than the scale alone gives, plus the death stub's.
        let spans = Dictionary(uniqueKeysWithValues: content.allSpans.map { ($0.id, $0.span) })
        let perYear: CGFloat = 43
        #expect(abs(y(unknown.mark("11")) - (18 + CGFloat(spans["11"]!.from - spans["2"]!.from) * perYear + 80)) < 0.5)
        // The death stub sits just before the burial, below the birth stub's shift, with its cap at the stub's top.
        let deathStart = 18 + CGFloat(spans["11"]!.from - spans["2"]!.from) * perYear + 40
        if case .cap(let deathY, true) = unknown.mark("death")?.shape { #expect(abs(deathY - deathStart) < 0.5) } else { Issue.record("death placeholder") }
        #expect(unknown.stubs.contains { abs($0.lowerBound - deathStart) < 0.5 && abs($0.upperBound - (deathStart + 40)) < 0.5 })

        // A funeral and no death: the death stub goes before the funeral.
        var timeline = Data.johann
        timeline.events = [
            Data.event("c", type: "census", when: .year(1875)),
            Data.event("f", type: "funeral", when: .year(1880)),
        ]
        timeline.birth = CatalogLifeMarker(place: .before(eventId: "c"))
        timeline.death = CatalogLifeMarker(place: .before(eventId: "f"))
        let (funeral, funeralContent) = lifeline(timeline, ppy: 43)
        #expect(funeralContent.items.map(\.id) == ["birth", "c", "death", "f"])
        #expect(funeral.mark("death").map { y($0) }! < y(funeral.mark("f")))

        // Only undated: two caps joined by one stretch, no tick.
        let (ida, _) = lifeline(Data.ida, ppy: 43)
        #expect(ida.ticks.isEmpty && ida.stubs == [18...58] && ida.length == 76)
        #expect(ida.marks.map(\.id) == ["birth", "death"])
        #expect(y(ida.mark("birth")) == 18 && y(ida.mark("death")) == 58)

        // A short life: year labels and quarter ticks from 56pt a year.
        let (ernst, _) = lifeline(Data.ernst, ppy: 100)
        #expect(ernst.ticks.compactMap(\.year) == [1846, 1847, 1848])
        #expect(ernst.ticks.contains { $0.isQuarter } && ernst.ticks.filter(\.isQuarter).allSatisfy { $0.year == nil })
        let (sparse, _) = lifeline(Data.ernst, ppy: 55)
        #expect(!sparse.ticks.contains { $0.isQuarter })
        let (dense, _) = lifeline(Data.ernst, ppy: 56)
        #expect(dense.ticks.contains { $0.isQuarter })

        // A range under a quarter year is drawn as a quarter: one dated event, one point.
        var one = Data.ida
        one.events = [Data.event("d", type: "census", when: .day(1850, 6, 1, from: 1850.4136986301369, to: 1850.4164383561642))]
        let (single, _) = lifeline(one, ppy: 100)
        #expect(single.marks.contains { $0.id == "d" } && abs(single.length - (18 + 25 + 40 + 40 + 18)) < 0.001)

        // A recorded birth with no span: a filled cap on a stub of its own, with a curve.
        var recorded = Data.johann
        recorded.events = recorded.events.map { event in
            var event = event
            if event.id == "1" { event.span = nil }
            return event
        }
        recorded.birth = CatalogLifeMarker(eventId: "1", place: .before(eventId: "2"))
        let (cap, _) = lifeline(recorded, ppy: 43)
        if case .cap(_, let placeholder) = cap.mark("1")?.shape { #expect(!placeholder) } else { Issue.record("filled cap") }
        let capTether = cap.tethers(anchors: ["1": 100], spans: ["1": 80...180], visible: CGRect(x: 0, y: 0, width: 800, height: 200)).first
        #expect(capTether?.id == "1" && capTether?.from.y == y(cap.mark("1")) && capTether?.to == CGPoint(x: PersonTimelineLifeline.flagX, y: 100))
    }

    @Test func tethersEndAtTheFlags() {
        let (johann, _) = lifeline(Data.johann, ppy: 43)
        let anchors: [String: CGFloat] = ["1": 40, "2": 90, "4": 300, "12": 500]
        let spans = anchors.mapValues { ($0 - 20)...($0 + 80) }
        let tethers = johann.tethers(anchors: anchors, spans: spans, visible: CGRect(x: 0, y: 0, width: 800, height: 200))
        #expect(tethers.map(\.id) == ["1", "2"])
        #expect(tethers.allSatisfy { $0.to.x == PersonTimelineLifeline.flagX && $0.from.x == PersonTimelineLifeline.lineX })
        #expect(tethers.map(\.to.y) == [40, 90])
        #expect(tethers.first?.from.y == johann.mark("1")?.pointY)
        // None for an undated row (no mark), for an anchor outside `visible`, or without a viewport.
        let all = johann.tethers(anchors: anchors, spans: spans, visible: CGRect(x: 0, y: 0, width: 800, height: 1000))
        #expect(all.map(\.id) == ["1", "2", "4"])
        #expect(johann.tethers(anchors: anchors, spans: spans, visible: CGRect(x: 0, y: 0, width: 800, height: 0)).isEmpty)
    }

    @Test func aTallRowKeepsItsTetherWhileItFillsTheViewport() {
        let (johann, _) = lifeline(Data.johann, ppy: 43)
        // Row 2's anchor is above the viewport but its 2000pt-tall item meets it.
        let visible = CGRect(x: 0, y: 500, width: 800, height: 400)
        let tethers = johann.tethers(anchors: ["2": 100], spans: ["2": 80...2080], visible: visible)
        #expect(tethers.map(\.id) == ["2"] && tethers.first?.to.y == 500)
        #expect(johann.tethers(anchors: ["2": 100], spans: ["2": 80...300], visible: visible).isEmpty)
    }

    @Test func oneHundredTwentyEvents() {
        let (line, content) = lifeline(Data.lise, ppy: 41, unfolded: [.dated])
        let datedIDs = content.allSpans.map(\.id)
        #expect(datedIDs.count == 120)
        #expect(line.marks.map(\.id) == content.items.compactMap { $0.eventRow?.span != nil ? $0.id : nil })
        #expect(line.marks.count == 120)
        // Folded, the line still shows all 120 points.
        let (folded, foldedContent) = lifeline(Data.lise, ppy: 41)
        #expect(foldedContent.items.contains(.more(.dated, count: 67)))
        #expect(folded.marks.count == 120)

        // A click target for each shown item with a point; the folded points share one region.
        #expect(folded.targets.count == foldedContent.items.filter { folded.mark($0.id) != nil }.count)
        #expect(folded.targets.count < 120)
        let foldedIDs = content.allSpans.map(\.id).filter { id in !foldedContent.items.contains { $0.id == id } }
        #expect(foldedIDs.count == 67)
        let foldedMarks = foldedIDs.compactMap { folded.mark($0) }
        let hit = folded.foldedHit
        #expect(hit?.lowerBound == foldedMarks.map(\.hit.lowerBound).min() && hit?.upperBound == foldedMarks.map(\.hit.upperBound).max())
        let middle = foldedMarks[30]
        // Points of one year share a y: the nearest is one of them, the earlier in item order.
        #expect(folded.foldedMark(nearest: middle.pointY).flatMap { folded.mark($0)?.pointY } == middle.pointY)
        #expect(folded.foldedMark(nearest: (hit?.lowerBound ?? 0) - 500) == foldedMarks.first?.id)
        #expect(line.foldedHit == nil && line.foldedMark(nearest: 100) == nil)
        // A target's extent: 6pt around a dot, a bar's extent with 6pt more at each end.
        #expect(PersonTimelineLifeline.Mark(id: "d", shape: .dot(y: 100)).hit == 94...106)
        #expect(PersonTimelineLifeline.Mark(id: "b", shape: .bar(from: 100, to: 200, openStart: false, openEnd: false)).hit == 94...206)
        #expect(PersonTimelineLifeline.Mark(id: "s", shape: .bar(from: 100, to: 100, openStart: false, openEnd: false)).hit == 94...106)
    }

    @Test func aDeathBeforeABurial() throws {
        let ppy: CGFloat = 10
        // A baptism and a burial: the death cap is 40pt above the burial and below the baptism.
        let life = Data.timeline(
            [Data.event("bap", type: "baptism", when: .year(1820)), Data.event("bur", type: "burial", when: .year(1850))],
            birth: CatalogLifeMarker(place: .before(eventId: "bap")), death: CatalogLifeMarker(place: .before(eventId: "bur"))
        )
        let (line, _) = lifeline(life, ppy: ppy)
        if case .cap(let deathY, true) = line.mark("death")?.shape {
            #expect(abs(deathY - (y(line.mark("bur")) - 40)) < 0.001)
            #expect(deathY > y(line.mark("bap")))
        } else { Issue.record("death placeholder") }

        // A baptism and a burial in the same year: the baptism is above the death stub, the burial below it.
        let infant = Data.timeline(
            [
                Data.event("bap", type: "baptism", when: .day(1850, 3, 1, from: 1850.17, to: 1850.1728)),
                Data.event("bur", type: "burial", when: .day(1850, 9, 1, from: 1850.67, to: 1850.6728)),
            ],
            birth: CatalogLifeMarker(place: .before(eventId: "bap")), death: CatalogLifeMarker(place: .before(eventId: "bur"))
        )
        let (short, _) = lifeline(infant, ppy: 100)
        if case .cap(let deathY, true) = short.mark("death")?.shape {
            #expect(y(short.mark("bap")) < deathY && y(short.mark("bur")) >= deathY + 40 - 0.001)
        } else { Issue.record("death placeholder") }

        // Both before the one burial: the birth stub is the 0th, the death stub the 1st, the burial after both.
        let burial = Data.timeline(
            [Data.event("bur", type: "burial", when: .day(1850, 9, 1, from: 1850.67, to: 1850.6728))],
            birth: CatalogLifeMarker(place: .before(eventId: "bur")), death: CatalogLifeMarker(place: .before(eventId: "bur"))
        )
        let (both, _) = lifeline(burial, ppy: ppy)
        #expect(both.stubs == [18...58, 58...98])
        #expect(abs(y(both.mark("bur")) - 98) < 0.001)

        // After the 60th of 60 rows, which a fold hides: the stub is at that row's end.
        var rows = (1...60).map { Data.event("d\($0)", type: "census", when: .year(Int32(1800 + $0))) }
        rows[59] = Data.event("d60", type: "census", when: .year(1860))
        let folded = Data.timeline(rows, birth: CatalogLifeMarker(), death: CatalogLifeMarker(place: .after(eventId: "d60")))
        let (afterFold, foldedContent) = lifeline(folded, ppy: ppy)
        #expect(foldedContent.items.contains(.more(.dated, count: 10)))
        let range = try #require(PersonTimelineLifeline.range(foldedContent))
        let d60 = try #require(foldedContent.allSpans.first { $0.id == "d60" })
        #expect(afterFold.stubs.count == 2)
        #expect(abs(afterFold.stubs[1].lowerBound - (18 + CGFloat(d60.span.to - range.from) * ppy + 40)) < 0.001)

        // A place that is not a dated row: the content drops it and the stub sits at the range's end.
        let unknown = Data.timeline(
            [Data.event("x", type: "census", when: .year(1850))],
            birth: CatalogLifeMarker(), death: CatalogLifeMarker(place: .before(eventId: "nope"))
        )
        let (end, unknownContent) = lifeline(unknown, ppy: ppy)
        guard case .marker(let death) = unknownContent.items.last else {
            Issue.record("death is last")
            return
        }
        #expect(death.place == nil)
        let unknownRange = try #require(PersonTimelineLifeline.range(unknownContent))
        #expect(abs(end.stubs[1].lowerBound - (18 + CGFloat(unknownRange.to - unknownRange.from) * ppy + 40)) < 0.001)
    }

    @Test func tickCountIsCapped() {
        var timeline = Data.ida
        timeline.events = [
            Data.event("a", type: "census", when: .year(-9999)),
            Data.event("b", type: "census", when: .year(9999)),
        ]
        let (line, _) = lifeline(timeline, ppy: 0.2)
        #expect(line.ticks.count <= 200 && !line.ticks.isEmpty)
        let years = line.ticks.compactMap(\.year)
        let step = years.count > 1 ? years[1] - years[0] : 0
        #expect(step > 50 && years.allSatisfy { $0 % step == 0 })
        #expect(years.first.map { $0 < 0 } == true)
        #expect(String(years.first ?? 0).hasPrefix("-"))
    }
}
