import SwiftUI

/// Every number and colour a circular link diagram draws with. The kit tokens are only the
/// standard style's defaults.
struct CircularLinkDiagramStyle: Equatable {
    var circleColor: Color = PVColor.borderSubtle
    var circleLineWidth: CGFloat = 1
    var arcColor: Color = PVColor.borderStrong
    var highlightedArcColor: Color = PVColor.accentLine
    var dimmedArcOpacity: Double = 0.18
    /// The control point's fraction from the centre toward the chord's midpoint.
    var arcCurvature: CGFloat = 0.3
    /// The gap between neighbouring arcs of one pair, as a fraction of the radius, clamped to `arcSpacingRange`.
    var arcSpacingFraction: CGFloat = 0.06
    var arcSpacingRange: ClosedRange<CGFloat> = 7...14
    var dotSize: CGFloat = 8
    var dotRingWidth: CGFloat = 1.5
    var nodeColor: Color = PVColor.textSecondary
    var highlightedNodeColor: Color = PVColor.accent
    var neighborNodeFill: Color = PVColor.accentSoft
    var neighborNodeRing: Color = PVColor.accentLine
    var recededNodeOpacity: Double = 0.35
    var labelFont: Font = PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.regular)
    var highlightedLabelFont: Font = PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.semibold)
    var labelColor: Color = PVColor.textSecondary
    var highlightedLabelColor: Color = PVColor.accent
    var neighborLabelColor: Color = PVColor.textPrimary
    var labelGap: CGFloat = PVSpacing.space3
    /// Past this many nodes only the highlighted or focused node prints its label.
    var labelsShownUpTo: Int = 30
    var overlayLabelBackground: Color = PVColor.surfacePage
    var nodePadding: CGSize = CGSize(width: 4, height: 2)
    var nodeHoverFill: Color = PVColor.surfaceHover
    /// Room kept outside the circle, left and right of it.
    var sideRoom: CGFloat = 84
    /// Room kept outside the circle, above and below it.
    var verticalRoom: CGFloat = 30
    /// The smallest circle.
    var minimumDiameter: CGFloat = 172
    /// A node is labelled beside its dot when its dot is this far, as a fraction of the radius, from the vertical through the centre.
    var labelSideThreshold: CGFloat = 0.35
    /// Where a single node's dot sits, and the height of the region it is drawn in.
    var singleNodeY: CGFloat = 24
    var singleNodeHeight: CGFloat = 64

    static let standard = CircularLinkDiagramStyle()
}
