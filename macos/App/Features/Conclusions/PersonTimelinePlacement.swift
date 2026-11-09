import CoreGraphics
import Foundation

/// Where the Timeline's rows sit against its line (S13-13, board S13-D4, update
/// 2026-11-03, TL-5b): each row wants its flag centred on its event's point, and
/// goes as near that as the rows above allow, so a stretch with no records opens
/// as a gap between rows. Pure: it takes each item's measured height and the
/// viewport's, chooses the scale, and gives every item's top, the anchors the
/// tethers end at and the line. It also owns every vertical measure of an item
/// (dividers, paddings, the flag), so the view, the anchors and the tethers agree.
struct PersonTimelinePlacement: Equatable {
    typealias Item = PersonTimelineContent.Item

    /// The flag's inset from the gutter's leading and top edges.
    static let flagInset: CGFloat = 4
    static let flagWidth: CGFloat = 52
    static let flagHeight: CGFloat = 20
    /// The flag's column: the flag and its inset.
    static let gutterWidth: CGFloat = flagInset + flagWidth
    static let bottomPadding: CGFloat = 20

    /// The most the median row may sit below its point before the scale grows.
    private static let comfortableDistance: CGFloat = 24
    /// The tallest a year is made by the rows alone, before the viewport's floor.
    private static let densestYear: CGFloat = 40
    private static let growth: [CGFloat] = [1, 1.25, 1.5, 1.75, 2]

    static func topPadding(_ item: Item, after previous: Item?) -> CGFloat {
        guard let previous else { return 4 }
        if case .undatedHeader = item { return 22 }
        if case .undatedHeader = previous { return 14 }
        return 20
    }

    static func hasDivider(_ item: Item, after previous: Item?) -> Bool {
        guard let previous else { return false }
        if case .undatedHeader = item { return false }
        if case .undatedHeader = previous { return false }
        return true
    }

    private static func flagCentre(_ item: Item, after previous: Item?) -> CGFloat {
        (hasDivider(item, after: previous) ? 1 : 0) + topPadding(item, after: previous) + flagInset + flagHeight / 2
    }

    /// How far below an item's top its anchor is: the flag's centre for an event
    /// row or a recorded marker, half the height of a placeholder; nil for the
    /// Undated heading and a "Show N more" row.
    private static func anchorOffset(_ item: Item, after previous: Item?, height: CGFloat) -> CGFloat? {
        switch item {
        case .row:
            return flagCentre(item, after: previous)
        case .marker(let marker):
            return marker.row == nil ? height / 2 : flagCentre(item, after: previous)
        case .undatedHeader, .more:
            return nil
        }
    }

    /// The scale: points a year on the line.
    private(set) var pointsPerYear: CGFloat = 0
    /// The line, nil when the content has no items.
    private(set) var lifeline: PersonTimelineLifeline?
    /// Each item's top in the section's space, by id.
    private(set) var tops: [String: CGFloat] = [:]
    /// Each anchored item's anchor in the section's space, by id.
    private(set) var anchors: [String: CGFloat] = [:]
    /// Each anchored item's top to bottom in the section's space, by id.
    private(set) var spans: [String: ClosedRange<CGFloat>] = [:]
    /// The section's height: the larger of the last item's bottom and the line's length.
    private(set) var height: CGFloat = 0

    /// The height an item with none measured yet takes: the mean of the measured
    /// ones, 120 when none is.
    static func estimatedHeight(heights: [String: CGFloat]) -> CGFloat {
        guard !heights.isEmpty else { return 120 }
        return heights.values.reduce(0, +) / CGFloat(heights.count)
    }

    /// The least the text column may be before the people go under it.
    static let minimumTextWidth: CGFloat = 280

    /// The event row's picture and the gap between its columns.
    static let rowPicture: CGFloat = 72
    static let rowGap: CGFloat = 16
    /// The gap between the people cards of a Wide row.
    static let cardColumnGap: CGFloat = 15

    /// The people column: one compact card at Regular; three and their two gaps at Wide.
    static func peopleWidth(_ widthClass: PVWidthClass) -> CGFloat {
        let card = PVTileCardSize.compact.minimumWidth
        return widthClass == .wide ? 3 * card + 2 * cardColumnGap : card
    }

    /// The text column's width in items `itemsWidth` wide, with the people beside
    /// it: the gutter, picture and people column and four gaps (gutter|picture,
    /// picture|text, text|spacer, spacer|people) taken out.
    static func textWidth(itemsWidth: CGFloat, widthClass: PVWidthClass) -> CGFloat {
        itemsWidth - gutterWidth - rowPicture - peopleWidth(widthClass) - 4 * rowGap
    }

    /// The widest the text column grows: the page's content measure less the
    /// line column, the gutter, the picture and two gaps.
    static let textCap: CGFloat = PVSpacing.widthContentMax - PersonTimelineLifeline.columnWidth - PersonTimelineLifeline.listGap - gutterWidth - rowPicture - 2 * rowGap

    /// Whether the people column goes under the title: the text column beside it
    /// would be under `minimumTextWidth`.
    static func stacksPeople(itemsWidth: CGFloat, widthClass: PVWidthClass) -> Bool {
        textWidth(itemsWidth: itemsWidth, widthClass: widthClass) < minimumTextWidth
    }

    /// `pointsPerYear` keeps a scale already chosen; nil chooses one. An item
    /// with no measured height takes `estimatedHeight`, so every item is placed.
    init(content: PersonTimelineContent, heights measured: [String: CGFloat], viewportHeight: CGFloat, pointsPerYear fixed: CGFloat? = nil) {
        let items = content.items
        guard !items.isEmpty else { return }
        let estimate = Self.estimatedHeight(heights: measured)
        let heights = Dictionary(items.map { ($0.id, measured[$0.id] ?? estimate) }, uniquingKeysWith: { first, _ in first })
        guard let range = PersonTimelineLifeline.range(content) else {
            let line = PersonTimelineLifeline(content: content, pointsPerYear: 0)
            var layout = Self.layout(items, heights: heights, line: line)
            layout.finish(line: line, items: items, heights: heights)
            self = layout.placement
            return
        }
        let years = CGFloat(range.to - range.from)
        if let fixed {
            let line = PersonTimelineLifeline(content: content, pointsPerYear: fixed)
            var layout = Self.layout(items, heights: heights, line: line)
            layout.pointsPerYear = fixed
            layout.finish(line: line, items: items, heights: heights)
            self = layout.placement
            return
        }
        let floor = max(300, viewportHeight - 320) / years
        let total = items.reduce(0) { $0 + heights[$1.id]! }
        let base = max(floor, min(total / years, Self.densestYear))

        var kept: (placement: PersonTimelinePlacement, median: CGFloat)?
        for factor in Self.growth {
            let line = PersonTimelineLifeline(content: content, pointsPerYear: base * factor)
            var layout = Self.layout(items, heights: heights, line: line)
            let median = layout.median
            layout.pointsPerYear = base * factor
            layout.finish(line: line, items: items, heights: heights)
            if let current = kept {
                if median < current.median - 1 { kept = (layout.placement, median) }
            } else {
                kept = (layout.placement, median)
            }
            if kept!.median <= Self.comfortableDistance { break }
        }
        self = kept!.placement
    }

    // MARK: - Pieces

    private init() {}

    /// One try's result before it is finished.
    private struct Layout {
        var placement = PersonTimelinePlacement()
        /// `top - want` of each anchored item.
        var distances: [CGFloat] = []

        var pointsPerYear: CGFloat {
            get { placement.pointsPerYear }
            set { placement.pointsPerYear = newValue }
        }

        var median: CGFloat {
            guard !distances.isEmpty else { return 0 }
            return distances.sorted()[distances.count / 2]
        }

        mutating func finish(line: PersonTimelineLifeline, items: [Item], heights: [String: CGFloat]) {
            placement.lifeline = line
            placement.height = max(items.last.map { (placement.tops[$0.id] ?? 0) + (heights[$0.id] ?? 0) } ?? 0, line.length)
        }
    }

    private static func layout(_ items: [Item], heights: [String: CGFloat], line: PersonTimelineLifeline) -> Layout {
        var result = Layout()
        var bottom: CGFloat = 0
        var previous: Item?
        for item in items {
            let h = heights[item.id] ?? 0
            var top = bottom
            if let offset = anchorOffset(item, after: previous, height: h), let mark = line.mark(item.id) {
                let want = mark.pointY - offset
                top = max(bottom, want)
                result.distances.append(top - want)
                result.placement.anchors[item.id] = top + offset
                result.placement.spans[item.id] = top...(top + h)
            }
            result.placement.tops[item.id] = top
            bottom = top + h
            previous = item
        }
        return result
    }
}

extension PersonTimelinePlacement {
    /// The scale chosen once every item was measured, kept while the items, the line's years
    /// and the viewport stay as they were: a date edit that moves the line's range takes a
    /// new scale.
    struct KeptScale: Equatable {
        private var ids: [String]
        private var years: ClosedRange<Double>?
        private var viewportHeight: CGFloat
        private var pointsPerYear: CGFloat

        init(content: PersonTimelineContent, viewportHeight: CGFloat, pointsPerYear: CGFloat) {
            ids = content.items.map(\.id)
            years = Self.years(content)
            self.viewportHeight = viewportHeight
            self.pointsPerYear = pointsPerYear
        }

        /// The kept scale; nil when `content` or the viewport is not what it was chosen for.
        func pointsPerYear(for content: PersonTimelineContent, viewportHeight: CGFloat) -> CGFloat? {
            guard viewportHeight == self.viewportHeight, years == Self.years(content), ids == content.items.map(\.id)
            else { return nil }
            return pointsPerYear
        }

        private static func years(_ content: PersonTimelineContent) -> ClosedRange<Double>? {
            PersonTimelineLifeline.range(content).map { $0.from...$0.to }
        }
    }
}
