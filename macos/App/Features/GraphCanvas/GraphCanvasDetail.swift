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

/// Host cutoffs for ``GraphCanvasDetail``.
///
/// Between the two values the current level stays put, so a pinch that hovers
/// on the boundary does not flicker the face.
struct GraphCanvasDetailBand: Equatable, Sendable {
    /// Magnification strictly below this is overview.
    var overviewBelow: CGFloat
    /// Magnification strictly above this is reading.
    var readingAbove: CGFloat

    /// Stays at reading for every positive magnification. A host that never
    /// sets a band does not swap faces.
    static let readingOnly = GraphCanvasDetailBand(overviewBelow: 0, readingAbove: 0)

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
