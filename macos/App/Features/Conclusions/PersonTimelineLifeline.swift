import CoreGraphics
import Foundation

/// The lifeline beside the Timeline's list (S13-13, board S13-D4): where each
/// dated event's point sits on a vertical line drawn to scale, the year ticks,
/// the dashed stubs for a birth or death with no date, and the tethers from a
/// point to its row's flag. Pure geometry on the spans Go gave (S13-33); Swift
/// does no date arithmetic. The scale is an input (`PersonTimelinePlacement`
/// chooses it from the rows); every y is in the section's own space.
struct PersonTimelineLifeline: Equatable {
    /// The width of the line's column.
    static let columnWidth: CGFloat = 80
    /// The line's x in its column.
    static let lineX: CGFloat = 56
    /// The gap between the line's column and the list.
    static let listGap: CGFloat = 12
    /// The x of a row's flag: where a tether ends.
    static var flagX: CGFloat { columnWidth + listGap + PersonTimelinePlacement.flagInset }
    /// What a birth or death with no date takes on the line, outside the scale.
    static let stubLength: CGFloat = 40
    /// The least gap between two year ticks.
    static let minimumTickSpacing: CGFloat = 34
    /// The least height of a year that earns quarter ticks.
    static let quarterTickYearHeight: CGFloat = 56
    /// The most ticks the line draws.
    static let maximumTicks = 200
    /// Where the line starts in the section's space, and the room after its end.
    static let pad: CGFloat = 18
    /// The shortest span of years the line scales: a life under it is drawn as this.
    static let minimumYears = 0.25

    /// A day is a dot: at most 1/365 of a year (1/366 in a leap year), so this catches both.
    private static let dayInYears = 1.0 / 365.0 + 1e-9

    enum Shape: Equatable {
        case dot(y: CGFloat)
        case bar(from: CGFloat, to: CGFloat, openStart: Bool, openEnd: Bool)
        case cap(y: CGFloat, isPlaceholder: Bool)
    }

    struct Mark: Equatable {
        /// The item id.
        var id: String
        var shape: Shape

        /// The y range of the mark's click target: 6pt around a dot or cap, and a bar's extent
        /// with 6pt more at each end.
        var hit: ClosedRange<CGFloat> {
            switch shape {
            case .dot(let y), .cap(let y, _): (y - 6)...(y + 6)
            case .bar(let from, let to, _, _): (from - 6)...max(to + 6, from + 6)
            }
        }

        /// Where a tether leaves it: a dot or cap at its y, a closed bar at its
        /// middle, an open bar 8pt in from its known end.
        var pointY: CGFloat {
            switch shape {
            case .dot(let y), .cap(let y, _): y
            case .bar(let from, let to, let openStart, let openEnd):
                if openStart { to - 8 } else if openEnd { from + 8 } else { (from + to) / 2 }
            }
        }
    }

    struct Tick: Equatable {
        var y: CGFloat
        /// Nil on a quarter tick.
        var year: Int?
        var isQuarter: Bool
    }

    struct Tether: Equatable {
        var id: String
        var from: CGPoint
        var to: CGPoint
    }

    /// One click target of the line: a shown item's mark and its name, shown on hover.
    struct Target: Equatable {
        var mark: Mark
        var label: String
    }

    private(set) var length: CGFloat = 0
    private(set) var marks: [Mark] = []
    /// A click target for each item with a mark, in item order. The folded rows' points have
    /// none of their own: they share `foldedHit`.
    private(set) var targets: [Target] = []
    /// The y range that covers the folded rows' points, one shared click target; nil when
    /// nothing is folded.
    private(set) var foldedHit: ClosedRange<CGFloat>?
    /// The folded rows' marks, in item order.
    private var foldedMarks: [Mark] = []
    private(set) var stubs: [ClosedRange<CGFloat>] = []
    private(set) var ticks: [Tick] = []
    /// Each mark by its item id.
    private var marksByID: [String: Mark] = [:]

    /// The span of years the line scales: the earliest `from` to the latest `to`
    /// of the dated rows, at least `minimumYears` long; nil with no dated row.
    static func range(_ content: PersonTimelineContent) -> (from: Double, to: Double)? {
        guard let earliest = content.allSpans.map(\.span.from).min(),
              let latest = content.allSpans.map(\.span.to).max()
        else { return nil }
        return (earliest, max(latest, earliest + minimumYears))
    }

    init(content: PersonTimelineContent, pointsPerYear: CGFloat) {
        let items = content.items
        let spans = content.allSpans
        let spansByID = Dictionary(spans.map { ($0.id, $0.span) }, uniquingKeysWith: { first, _ in first })
        let shown = Set(items.map(\.id))

        // The markers that take a stub of their own: a placeholder, or a
        // recorded one with no span.
        let stubbed: [PersonTimelineContent.Marker] = items.compactMap { item in
            guard case .marker(let marker) = item, marker.row?.span == nil else { return nil }
            return marker
        }

        guard let range = Self.range(content) else {
            // Nothing to scale (only undated events): the two caps and one
            // dashed stretch between them.
            let top = Self.pad, bottom = Self.pad + Self.stubLength
            for marker in stubbed {
                let item = PersonTimelineContent.Item.marker(marker)
                let mark = Mark(id: item.id, shape: .cap(y: marker.kind == .birth ? top : bottom, isPlaceholder: marker.row == nil))
                marks.append(mark)
                targets.append(Target(mark: mark, label: marker.spokenLabel))
            }
            stubs = [top...bottom]
            length = bottom + Self.pad
            marksByID = Dictionary(marks.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
            return
        }

        // Where each stub is inserted on the line of years, from the place the content
        // resolved: before a row, at the least `from` of that row and the shown dated
        // rows after the marker (a row the Timeline's order put after the stub can start
        // earlier than the row it names); after a row, at its `to`; with none, a birth at
        // the range's start and a death at its end.
        struct Insertion { var at: Double; var isBefore: Bool }
        let insertions: [String: Insertion] = Dictionary(stubbed.map { marker in
            let id = PersonTimelineContent.Item.marker(marker).id
            let insertion: Insertion
            switch marker.place {
            case .before(let row)? where spansByID[row] != nil:
                let later = (items.firstIndex { $0.id == id }.map { items[($0 + 1)...] } ?? []).compactMap { item -> Double? in
                    switch item {
                    case .row, .marker: spansByID[item.id]?.from
                    case .undatedHeader, .more: nil
                    }
                }
                insertion = Insertion(at: ([spansByID[row]!.from] + later).min()!, isBefore: true)
            case .after(let row)? where spansByID[row] != nil:
                insertion = Insertion(at: spansByID[row]!.to, isBefore: false)
            default:
                insertion = marker.kind == .birth
                    ? Insertion(at: range.from, isBefore: true) : Insertion(at: range.to, isBefore: false)
            }
            return (id, insertion)
        }, uniquingKeysWith: { first, _ in first })
        length = Self.pad + CGFloat(range.to - range.from) * pointsPerYear + Self.stubLength * CGFloat(stubbed.count) + Self.pad

        func base(_ year: Double) -> CGFloat { Self.pad + CGFloat(year - range.from) * pointsPerYear }

        // Marks, in item order. A point is drawn below the stubs before its item; the folded
        // rows' points are where their fold is.
        let recordedSpanMarkers = Set(items.compactMap { item -> String? in
            guard case .marker(let marker) = item, let row = marker.row, row.span != nil else { return nil }
            return row.id
        })
        func point(_ id: String, _ span: CatalogTimelineSpan, stubsBefore: Int) -> Mark {
            let shift = Self.stubLength * CGFloat(stubsBefore)
            if recordedSpanMarkers.contains(id) {
                return Mark(id: id, shape: .cap(y: base(span.from) + shift, isPlaceholder: false))
            }
            if span.to - span.from <= Self.dayInYears {
                return Mark(id: id, shape: .dot(y: base(span.from) + shift))
            }
            return Mark(id: id, shape: .bar(
                from: base(span.from) + shift, to: base(span.to) + shift, openStart: span.openStart, openEnd: span.openEnd
            ))
        }
        var stubsBefore = 0
        for item in items {
            switch item {
            case .row, .marker:
                let mark: Mark
                if let stub = insertions[item.id] {
                    let start = base(stub.at) + Self.stubLength * CGFloat(stubsBefore)
                    stubs.append(start...(start + Self.stubLength))
                    mark = Mark(
                        id: item.id,
                        shape: .cap(y: stub.isBefore ? start : start + Self.stubLength, isPlaceholder: item.eventRow == nil)
                    )
                    stubsBefore += 1
                } else if let span = spansByID[item.id] {
                    mark = point(item.id, span, stubsBefore: stubsBefore)
                } else {
                    continue
                }
                marks.append(mark)
                targets.append(Target(mark: mark, label: item.spokenLabel ?? ""))
            case .more(.dated, _):
                for span in spans where !shown.contains(span.id) {
                    let mark = point(span.id, span.span, stubsBefore: stubsBefore)
                    marks.append(mark)
                    foldedMarks.append(mark)
                }
            case .undatedHeader, .more: break
            }
        }
        if let low = foldedMarks.map(\.hit.lowerBound).min(), let high = foldedMarks.map(\.hit.upperBound).max() {
            foldedHit = low...high
        }

        // A year on the line, moved down by the stubs inserted before it.
        func map(_ year: Double) -> CGFloat {
            var y = base(year)
            for stub in insertions.values where stub.isBefore ? year >= stub.at : year > stub.at { y += Self.stubLength }
            return y
        }

        // Ticks.
        let years = Double(range.to - range.from)
        let step = Self.tickStep(from: range.from, to: range.to, pointsPerYear: pointsPerYear)
        let firstYear = Int(range.from.rounded(.up))
        let lastYear = Int(range.to.rounded(.down))
        if firstYear <= lastYear {
            for year in firstYear...lastYear where year % step == 0 {
                ticks.append(Tick(y: map(Double(year)), year: year, isQuarter: false))
            }
        }
        if pointsPerYear >= Self.quarterTickYearHeight, years > 0 {
            for year in Int(range.from.rounded(.down))...lastYear {
                for quarter in [0.25, 0.5, 0.75] {
                    let at = Double(year) + quarter
                    if at > range.from, at < range.to { ticks.append(Tick(y: map(at), year: nil, isQuarter: true)) }
                }
            }
            ticks.sort { $0.y < $1.y }
        }

        marksByID = Dictionary(marks.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
    }

    /// The tethers of the rows whose span (the item's top to bottom, from
    /// placement) meets `visible`'s vertical range, in the section's space: a
    /// curve from the row's point on the line to its flag, its end y held within
    /// `visible`. A row with no point of its own (an undated row) has no anchor
    /// and no curve. A zero-height `visible` (no viewport size yet) draws none.
    func tethers(anchors: [String: CGFloat], spans: [String: ClosedRange<CGFloat>], visible: CGRect) -> [Tether] {
        guard visible.height > 0 else { return [] }
        return marks.compactMap { mark in
            guard let y = anchors[mark.id], let span = spans[mark.id],
                  span.upperBound >= visible.minY, span.lowerBound <= visible.maxY
            else { return nil }
            return Tether(
                id: mark.id,
                from: CGPoint(x: Self.lineX, y: mark.pointY),
                to: CGPoint(x: Self.flagX, y: min(max(y, visible.minY), visible.maxY))
            )
        }
    }

    /// The marks of the rows with a point, and which of them is where.
    func mark(_ id: String) -> Mark? { marksByID[id] }

    /// The folded mark whose point is nearest `y`, in the line's space (the earlier in item order on
    /// a tie); nil when nothing is folded.
    func foldedMark(nearest y: CGFloat) -> String? {
        var best: (id: String, distance: CGFloat)?
        for mark in foldedMarks {
            let distance = abs(mark.pointY - y)
            if best == nil || distance < best!.distance { best = (mark.id, distance) }
        }
        return best?.id
    }

    // MARK: - Pieces

    /// The step between year ticks: the smallest of 1, 2, 5, 10, 20, 50 that
    /// leaves them `minimumTickSpacing` apart; if even 50 draws more than
    /// `maximumTicks`, it grows through 100, 200, 500 … until it does not.
    static func tickStep(from: Double, to: Double, pointsPerYear: CGFloat) -> Int {
        let steps = [1, 2, 5, 10, 20, 50]
        if let step = steps.first(where: { CGFloat($0) * pointsPerYear >= minimumTickSpacing }) { return step }
        func count(_ step: Int) -> Int {
            let first = Int(from.rounded(.up)), last = Int(to.rounded(.down))
            guard first <= last else { return 0 }
            return Int((Double(last) / Double(step)).rounded(.down)) - Int((Double(first - 1) / Double(step)).rounded(.down))
        }
        var step = 50
        var next = [100, 200, 500, 1000, 2000, 5000, 10000].makeIterator()
        while count(step) > maximumTicks, let bigger = next.next() { step = bigger }
        return step
    }
}
