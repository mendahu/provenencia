import CoreGraphics

/// Where the Family section puts everything (S13-12, board S13-D3 `layTree3`):
/// the tiles of both columns, their captions, and the one-point lines that join
/// them. A pure function of the content, the width of the primary column and
/// the width class, so the geometry is tested without a view.
///
/// The left column is the origin: each parent set under its caption, then the
/// siblings indented beneath on rails in the margin, the person among them. The
/// right column is the family the person formed: each partner with their
/// children indented beneath. A bus joins the person's tile to every partner.
/// A folded list ends in one "more" row; no line goes to it.
struct PersonFamilyLayout: Equatable {
    struct PlacedTile: Equatable {
        var id: String
        /// Relative to its group.
        var frame: CGRect
    }

    /// The "Show N more" row that ends a folded list. The view words it.
    struct PlacedMore: Equatable {
        var fold: PersonFamilyContent.Fold
        var count: Int
        /// Relative to its group.
        var frame: CGRect
    }

    /// One VoiceOver container, in reading order.
    struct Group: Equatable {
        var label: String
        var frame: CGRect
        var tiles: [PlacedTile]
        var more: [PlacedMore] = []
    }

    struct Caption: Equatable {
        var text: String
        var origin: CGPoint
    }

    static let margin: CGFloat = 20
    static let moreRowHeight: CGFloat = 28
    static let moreGap: CGFloat = 4
    static let tileGap: CGFloat = 12
    static let captionBand: CGFloat = 26
    static let indent: CGFloat = 32
    static let afterSet: CGFloat = 20
    static let betweenGroups: CGFloat = 28
    static let bracketX: CGFloat = 40

    static func gutter(_ widthClass: PVWidthClass) -> CGFloat {
        widthClass == .wide ? 96 : 48
    }

    /// The x of parent set `k`'s rail among `n` sets.
    static func railX(_ k: Int, of n: Int) -> CGFloat {
        switch n {
        case 1: return 10
        case 2: return k == 0 ? 6 : 14
        default: return 4 + (12 * CGFloat(k) / CGFloat(n - 1)).rounded()
        }
    }

    var groups: [Group] = []
    var captions: [Caption] = []
    /// Horizontal and vertical lines, one point thick.
    var segments: [CGRect] = []
    var height: CGFloat = 0
    /// The column width, the right column's x, and the bus's x, as laid out.
    var columnWidth: CGFloat = 0
    var rightX: CGFloat = 0
    var busX: CGFloat = 0

    /// The absolute frame of the tile with `id`.
    func frame(ofTile id: String) -> CGRect? {
        for group in groups {
            if let tile = group.tiles.first(where: { $0.id == id }) {
                return tile.frame.offsetBy(dx: group.frame.minX, dy: group.frame.minY)
            }
        }
        return nil
    }

    init(content: PersonFamilyContent, width: CGFloat, widthClass: PVWidthClass) {
        let m = Self.margin
        let th = PersonTile.height
        let step = th + Self.tileGap
        let gutter = Self.gutter(widthClass)
        let tp = max((width - m - gutter) / 2, 120).rounded(.down)
        let c2 = m + tp + gutter
        columnWidth = tp
        rightX = c2
        busX = c2 - Self.indent

        // Absolute frames collected per container, then made relative.
        func container(_ label: String, _ placed: [(String, CGRect)], more: [(PersonFamilyContent.Fold, Int, CGRect)] = []) -> Group {
            let union = (placed.map(\.1) + more.map(\.2)).dropFirst().reduce(placed[0].1) { $0.union($1) }
            return Group(
                label: label,
                frame: union,
                tiles: placed.map { PlacedTile(id: $0.0, frame: $0.1.offsetBy(dx: -union.minX, dy: -union.minY)) },
                more: more.map { PlacedMore(fold: $0.0, count: $0.1, frame: $0.2.offsetBy(dx: -union.minX, dy: -union.minY)) }
            )
        }
        /// The row under a list that ends at `bottom`.
        func moreRow(x: CGFloat, width: CGFloat, below bottom: CGFloat) -> CGRect {
            CGRect(x: x, y: bottom + Self.moreGap, width: width, height: Self.moreRowHeight)
        }
        let moreBand = Self.moreGap + Self.moreRowHeight
        func horizontal(_ x1: CGFloat, _ x2: CGFloat, y: CGFloat) {
            segments.append(CGRect(x: min(x1, x2), y: y, width: abs(x2 - x1), height: 1))
        }
        func vertical(x: CGFloat, _ y1: CGFloat, _ y2: CGFloat) {
            segments.append(CGRect(x: x, y: min(y1, y2), width: 1, height: abs(y2 - y1) + 1))
        }

        var leftBottom: CGFloat = 0
        var focusCenter: CGFloat?
        var focusRight: CGFloat?

        // Left: the parent sets.
        let n = content.parentSets.count
        var y: CGFloat = 0
        var firstParentCenter: [CGFloat] = []
        var parentCenters: [[CGFloat]] = []
        for (k, set) in content.parentSets.enumerated() {
            captions.append(Caption(text: content.parentCaptions[k], origin: CGPoint(x: m, y: y)))
            var placed: [(String, CGRect)] = []
            var centers: [CGFloat] = []
            for (i, tile) in set.enumerated() {
                let top = y + Self.captionBand + CGFloat(i) * step
                placed.append((tile.id, CGRect(x: m, y: top, width: tp, height: th)))
                centers.append(top + th / 2)
            }
            var bottom = y + Self.captionBand + CGFloat(set.count) * step - Self.tileGap
            var more: [(PersonFamilyContent.Fold, Int, CGRect)] = []
            if k == n - 1, content.parentSetsHidden > 0 {
                more.append((.parentSets, content.parentSetsHidden, moreRow(x: m, width: tp, below: bottom)))
                bottom += moreBand
            }
            if !placed.isEmpty { groups.append(container(content.parentCaptions[k], placed, more: more)) }
            parentCenters.append(centers)
            firstParentCenter.append(centers.first ?? y + Self.captionBand)
            leftBottom = max(leftBottom, bottom)
            y = bottom + Self.afterSet
        }

        // Left: the siblings.
        let siblingX = n > 0 ? m + Self.indent : m
        let siblingWidth = n > 0 ? tp - Self.indent : tp
        let siblingTop = n > 0 ? y : Self.captionBand
        if let caption = content.siblingsCaption {
            captions.append(Caption(text: caption, origin: CGPoint(x: m, y: 0)))
        }
        var siblingCenters: [CGFloat] = []
        var siblingPlaced: [(String, CGRect)] = []
        for (i, sibling) in content.siblings.enumerated() {
            let top = siblingTop + CGFloat(i) * step
            let frame = CGRect(x: siblingX, y: top, width: siblingWidth, height: th)
            siblingPlaced.append((sibling.tile.id, frame))
            siblingCenters.append(top + th / 2)
            if sibling.tile.content.isFocus {
                focusCenter = top + th / 2
                focusRight = frame.maxX
            }
            leftBottom = max(leftBottom, top + th)
        }
        var siblingMore: [(PersonFamilyContent.Fold, Int, CGRect)] = []
        if content.siblingsHidden > 0, let last = siblingPlaced.last {
            let row = moreRow(x: siblingX, width: siblingWidth, below: last.1.maxY)
            siblingMore.append((.siblings, content.siblingsHidden, row))
            leftBottom = max(leftBottom, row.maxY)
        }
        if !siblingPlaced.isEmpty { groups.append(container(content.siblingsLabel, siblingPlaced, more: siblingMore)) }

        // Left lines. One pass finds each parent set's siblings and the first
        // set each sibling is in.
        if n > 0 {
            var members = [[Int]](repeating: [], count: n)
            var firstSet = [Int?](repeating: nil, count: content.siblings.count)
            for (i, sibling) in content.siblings.enumerated() {
                for k in sibling.sets where k >= 0 && k < n {
                    members[k].append(i)
                    firstSet[i] = min(firstSet[i] ?? k, k)
                }
            }
            for k in 0..<n {
                let rail = Self.railX(k, of: n)
                for center in parentCenters[k] { horizontal(rail, m, y: center) }
                if let last = members[k].last {
                    vertical(x: rail, firstParentCenter[k], siblingCenters[last])
                }
                for i in members[k] where firstSet[i] == k {
                    horizontal(rail, siblingX, y: siblingCenters[i])
                }
            }
            let recorded = content.siblings.indices.filter { content.siblings[$0].sets.isEmpty }
            if let focusCenter, let last = recorded.last {
                vertical(x: Self.bracketX, focusCenter, siblingCenters[last])
                horizontal(Self.bracketX, siblingX, y: focusCenter)
                for i in recorded { horizontal(Self.bracketX, siblingX, y: siblingCenters[i]) }
            }
        } else if content.siblings.count > 1 {
            let rail = Self.railX(0, of: 1)
            vertical(x: rail, siblingCenters[0], siblingCenters[siblingCenters.count - 1])
            for center in siblingCenters { horizontal(rail, m, y: center) }
        }

        // Right: the groups.
        var rightBottom: CGFloat = 0
        var joins: [CGFloat] = []
        var gy = Self.captionBand
        if let caption = content.rightCaption {
            captions.append(Caption(text: caption, origin: CGPoint(x: c2, y: 0)))
        }
        let childX = c2 + Self.indent
        let childRail = c2 - 12
        for (index, group) in content.groups.enumerated() {
            var placed: [(String, CGRect)] = []
            var top = gy
            var partnerCenter: CGFloat?
            if let partner = group.partner {
                placed.append((partner.id, CGRect(x: c2, y: top, width: tp, height: th)))
                partnerCenter = top + th / 2
                top += step
            }
            var childCenters: [CGFloat] = []
            for child in group.children {
                placed.append((child.id, CGRect(x: childX, y: top, width: tp - Self.indent, height: th)))
                childCenters.append(top + th / 2)
                top += step
            }
            let join = partnerCenter ?? childCenters.first ?? gy
            joins.append(join)
            if let partnerCenter {
                horizontal(busX, c2, y: partnerCenter)
                if let last = childCenters.last {
                    vertical(x: childRail, partnerCenter, last)
                    for center in childCenters { horizontal(childRail, childX, y: center) }
                }
            } else if childCenters.count == 1 {
                horizontal(busX, childX, y: childCenters[0])
            } else if let first = childCenters.first, let last = childCenters.last {
                horizontal(busX, childRail, y: first)
                vertical(x: childRail, first, last)
                for center in childCenters { horizontal(childRail, childX, y: center) }
            }
            var bottom = top - Self.tileGap
            var more: [(PersonFamilyContent.Fold, Int, CGRect)] = []
            if group.childrenHidden > 0 {
                more.append((.children(index), group.childrenHidden, moreRow(x: childX, width: tp - Self.indent, below: bottom)))
                bottom += moreBand
            }
            if index == content.groups.count - 1, content.groupsHidden > 0 {
                more.append((.partners, content.groupsHidden, moreRow(x: c2, width: tp, below: bottom)))
                bottom += moreBand
            }
            if !placed.isEmpty {
                groups.append(container(group.label, placed, more: more))
                rightBottom = max(rightBottom, bottom)
                gy = bottom + Self.betweenGroups
            }
        }
        if !joins.isEmpty {
            var spans = joins
            if let focusCenter, let focusRight {
                spans.append(focusCenter)
                horizontal(focusRight, busX, y: focusCenter)
            }
            vertical(x: busX, spans.min() ?? 0, spans.max() ?? 0)
        }
        height = max(leftBottom, rightBottom)
    }
}
