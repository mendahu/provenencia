import CoreGraphics
import Foundation

/// The lifeline beside the Timeline's list (S13-13, board S13-D4): where each
/// dated event sits on a vertical line drawn to scale from the first item's top
/// to the last item's bottom, the year ticks, the dashed stubs for a birth or
/// death with no date, and the tethers from a row in view to its point. Pure
/// geometry on the spans Go gave (S13-33); Swift does no date arithmetic.
///
/// Every y is measured from the line's start, the first item's top, except a
/// tether's, which is in the section's own space like the item frames.
struct PersonTimelineLifeline: Equatable {
    /// The width of the line's column.
    static let columnWidth: CGFloat = 80
    /// The line's x in its column.
    static let lineX: CGFloat = 56
    /// What a birth or death with no date takes on the line, outside the scale.
    static let stubLength: CGFloat = 40
    /// The least gap between two year ticks.
    static let minimumTickSpacing: CGFloat = 34
    /// The least height of a year that earns quarter ticks.
    static let quarterTickYearHeight: CGFloat = 136
    /// The most ticks the line draws.
    static let maximumTicks = 200
    /// How far below a row's top its dot sits.
    static let dotInset: CGFloat = 12

    /// A day is a dot: S13-33 divides every year by 365, leap years too.
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

        /// Where a tether leaves it.
        var pointY: CGFloat {
            switch shape {
            case .dot(let y), .cap(let y, _): y
            case .bar(let from, let to, _, _): (from + to) / 2
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
        /// An undated row: a dashed dot at `to` and no curve.
        var isDashed: Bool
    }

    /// An item that gets a dot at its leading edge.
    private struct Target: Equatable {
        var id: String
        var pointY: CGFloat?
    }

    /// The first item's top in the section's space.
    private(set) var origin: CGFloat = 0
    private(set) var length: CGFloat = 0
    private(set) var marks: [Mark] = []
    private(set) var stubs: [ClosedRange<CGFloat>] = []
    private(set) var ticks: [Tick] = []
    private var targets: [Target] = []
    /// Each mark by its item id.
    private var marksByID: [String: Mark] = [:]

    init(content: PersonTimelineContent, itemFrames: [String: CGRect]) {
        let items = content.items
        guard let first = items.first.flatMap({ itemFrames[$0.id] }),
              let last = items.last.flatMap({ itemFrames[$0.id] }),
              last.maxY > first.minY
        else { return }
        origin = first.minY
        length = last.maxY - first.minY

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

        guard let earliest = spans.map(\.span.from).min(), let latest = spans.map(\.span.to).max() else {
            // Nothing to scale (only undated events): the two caps sit at the
            // first and last items, joined by one dashed stretch.
            for item in stubbed {
                let y: CGFloat = item.kind == .birth ? 0 : length
                marks.append(Mark(id: item.item, shape: .cap(y: y, isPlaceholder: !item.recorded)))
            }
            stubs = [0...length]
            targets = Self.targets(items, pointY: { _ in nil })
            return
        }
        let range = (from: earliest, to: max(latest, earliest + 1))

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
        let pointsPerYear = max(0, length - Self.stubLength * CGFloat(insertions.count)) / CGFloat(range.to - range.from)

        func map(_ year: Double) -> CGFloat {
            var y = CGFloat(year - range.from) * pointsPerYear
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
        targets = Self.targets(items, pointY: { marksByID[$0]?.pointY })
    }

    /// The tethers of the rows whose frame meets `visible`, in the section's
    /// space: a curve from the row's point on the line to a dot at the row's
    /// leading edge, or only the dot, dashed, for a row with no date. A
    /// zero-height `visible` (no viewport size yet) draws none.
    func tethers(itemFrames: [String: CGRect], visible: CGRect) -> [Tether] {
        guard visible.height > 0 else { return [] }
        return targets.compactMap { target in
            guard let frame = itemFrames[target.id], frame.intersects(visible) else { return nil }
            let end = CGPoint(x: Self.columnWidth, y: frame.minY + Self.dotInset)
            guard let pointY = target.pointY else { return Tether(id: target.id, from: end, to: end, isDashed: true) }
            return Tether(id: target.id, from: CGPoint(x: Self.lineX, y: origin + pointY), to: end, isDashed: false)
        }
    }

    /// The marks of the rows with a point, and which of them is where.
    func mark(_ id: String) -> Mark? { marksByID[id] }

    // MARK: - Pieces

    /// The items that get a dot at their leading edge: every event row and recorded
    /// marker. A tether leaves a point only when the item has one that is a row's own.
    private static func targets(
        _ items: [PersonTimelineContent.Item],
        pointY: (String) -> CGFloat?
    ) -> [Target] {
        items.compactMap { item in
            guard let row = item.eventRow else { return nil }
            return Target(id: item.id, pointY: row.span == nil ? nil : pointY(item.id))
        }
    }

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
