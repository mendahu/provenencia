import CoreGraphics

/// How much of a card a host paints.
///
/// The canvas decides when this changes. What the overview face draws stays in
/// the product folder, so a later tree can reuse the cutoffs and paint its own.
enum GraphCanvasDetail: Equatable, Sendable {
    /// The host's full card.
    case reading
    /// The host's zoomed-out face.
    case overview
}

/// Host cut for ``GraphCanvasDetail``.
///
/// The caller names the break point. The band opens ``margin`` (2.5%) on
/// either side, and the level stays put inside that gap so a pinch on the
/// boundary does not flicker the face.
struct GraphCanvasDetailBand: Equatable, Sendable {
    /// Half-width of the gap around the cut, as a fraction of magnification.
    static let margin: CGFloat = 2.5 / 100

    /// Magnification strictly below this is overview.
    var overviewBelow: CGFloat
    /// Magnification strictly above this is reading.
    var readingAbove: CGFloat

    /// Overview below `cut - margin`, reading above `cut + margin`.
    init(cut: CGFloat) {
        self.init(overviewBelow: cut - Self.margin, readingAbove: cut + Self.margin)
    }

    /// Stays at reading for every positive magnification. A host that never
    /// sets a band does not swap faces.
    static let readingOnly = GraphCanvasDetailBand(overviewBelow: 0, readingAbove: 0)

    private init(overviewBelow: CGFloat, readingAbove: CGFloat) {
        self.overviewBelow = overviewBelow
        self.readingAbove = readingAbove
    }

    /// Level for a magnification that has no previous face. Overview only when
    /// already below `overviewBelow`.
    func initial(magnification: CGFloat) -> GraphCanvasDetail {
        magnification < overviewBelow ? .overview : .reading
    }

    /// Next level given the face already showing.
    func next(_ current: GraphCanvasDetail, magnification: CGFloat) -> GraphCanvasDetail {
        if magnification < overviewBelow { return .overview }
        if magnification > readingAbove { return .reading }
        return current
    }
}
