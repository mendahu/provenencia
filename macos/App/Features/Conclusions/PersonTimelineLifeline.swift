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

    private(set) var length: CGFloat = 0
    private(set) var marks: [Mark] = []
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
        let folded = spans.map(\.id).filter { !shown.contains($0) }

        // The markers that take a stub of their own: a placeholder, or a
        // recorded one with no span.
        struct Stubbed { var item: String; var kind: PersonTimelineContent.Marker.Kind; var recorded: Bool }
        let stubbed: [Stubbed] = items.compactMap { item in
            guard case .marker(let marker) = item, marker.row?.span == nil else { return nil }
            return Stubbed(item: item.id, kind: marker.kind, recorded: marker.row != nil)
        }

        guard let range = Self.range(content) else {
            // Nothing to scale (only undated events): the two caps and one
            // dashed stretch between them.
            let top = Self.pad, bottom = Self.pad + Self.stubLength
            for item in stubbed {
                marks.append(Mark(id: item.item, shape: .cap(y: item.kind == .birth ? top : bottom, isPlaceholder: !item.recorded)))
            }
            stubs = [top...bottom]
            length = bottom + Self.pad
            marksByID = Dictionary(marks.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
            return
        }

        // Where each stub is inserted on the line of years: a birth's at the
        // `from` of the row it precedes, a death's at the `to` of the row it
        // follows; the range's start and end when there is none.
        struct Insertion { var item: String; var at: Double; var kind: PersonTimelineContent.Marker.Kind }
        let insertions: [Insertion] = stubbed.map { item in
            switch item.kind {
            case .birth:
                let next = Self.nextRowID(after: item.item, in: items, folded: folded)
                return Insertion(item: item.item, at: next.flatMap { spansByID[$0]?.from } ?? range.from, kind: .birth)
            case .death:
                let previous = Self.previousRowID(before: item.item, in: items, folded: folded)
                return Insertion(item: item.item, at: previous.flatMap { spansByID[$0]?.to } ?? range.to, kind: .death)
            }
        }
        length = Self.pad + CGFloat(range.to - range.from) * pointsPerYear + Self.stubLength * CGFloat(insertions.count) + Self.pad

        func map(_ year: Double) -> CGFloat {
            var y = Self.pad + CGFloat(year - range.from) * pointsPerYear
            for insertion in insertions {
                switch insertion.kind {
                case .birth where year >= insertion.at: y += Self.stubLength
                case .death where year > insertion.at: y += Self.stubLength
                default: break
                }
            }
            return y
        }

        // Marks, in item order; the folded rows' points where their fold is.
        let recordedSpanMarkers = Set(items.compactMap { item -> String? in
            guard case .marker(let marker) = item, let row = marker.row, row.span != nil else { return nil }
            return row.id
        })
        func point(_ id: String, _ span: CatalogTimelineSpan) -> Mark {
            if recordedSpanMarkers.contains(id) {
                return Mark(id: id, shape: .cap(y: map(span.from), isPlaceholder: false))
            }
            if span.to - span.from <= Self.dayInYears {
                return Mark(id: id, shape: .dot(y: map(span.from)))
            }
            return Mark(id: id, shape: .bar(
                from: map(span.from), to: map(span.to), openStart: span.openStart, openEnd: span.openEnd
            ))
        }
        for item in items {
            switch item {
            case .row, .marker:
                if let stub = insertions.first(where: { $0.item == item.id }) {
                    let recorded = item.eventRow != nil
                    let start = stub.kind == .birth ? map(stub.at) - Self.stubLength : map(stub.at)
                    stubs.append(start...(start + Self.stubLength))
                    marks.append(Mark(
                        id: item.id,
                        shape: .cap(y: stub.kind == .birth ? start : start + Self.stubLength, isPlaceholder: !recorded)
                    ))
                } else if let span = spansByID[item.id] {
                    marks.append(point(item.id, span))
                }
            case .more(.dated, _):
                for span in spans where !shown.contains(span.id) { marks.append(point(span.id, span.span)) }
            case .undatedHeader, .more: break
            }
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

    /// The tethers of the rows whose anchor lies in `visible`'s vertical range,
    /// in the section's space: a curve from the row's point on the line to its
    /// flag. A row with no point of its own (an undated row) has no anchor and
    /// no curve. A zero-height `visible` (no viewport size yet) draws none.
    func tethers(anchors: [String: CGFloat], visible: CGRect) -> [Tether] {
        guard visible.height > 0 else { return [] }
        return marks.compactMap { mark in
            guard let y = anchors[mark.id], y >= visible.minY, y <= visible.maxY else { return nil }
            return Tether(
                id: mark.id,
                from: CGPoint(x: Self.lineX, y: mark.pointY),
                to: CGPoint(x: Self.flagX, y: y)
            )
        }
    }

    /// The marks of the rows with a point, and which of them is where.
    func mark(_ id: String) -> Mark? { marksByID[id] }

    // MARK: - Pieces

    /// The dated row after the item `id`, a fold standing for its first hidden row.
    private static func nextRowID(after id: String, in items: [PersonTimelineContent.Item], folded: [String]) -> String? {
        guard let at = items.firstIndex(where: { $0.id == id }) else { return nil }
        for item in items[(at + 1)...] {
            if case .more(.dated, _) = item { return folded.first }
            if let row = item.eventRow, row.span != nil { return row.id }
        }
        return nil
    }

    /// The dated row before the item `id`, a fold standing for its last hidden row.
    private static func previousRowID(before id: String, in items: [PersonTimelineContent.Item], folded: [String]) -> String? {
        guard let at = items.firstIndex(where: { $0.id == id }) else { return nil }
        for item in items[..<at].reversed() {
            if case .more(.dated, _) = item { return folded.last }
            if let row = item.eventRow, row.span != nil { return row.id }
        }
        return nil
    }

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
