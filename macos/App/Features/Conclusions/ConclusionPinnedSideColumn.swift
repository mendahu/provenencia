import SwiftUI

/// A main column with a side column beside it inside a full-width section, the side
/// column held in view while the main scrolls. Both are placed by the kit's
/// `PVPageColumnsLayout` over the same two subviews, so crossing the width where the side
/// column stacks under the main keeps their identity, focus, hover and reads.
///
/// `mainMin` is the narrowest the main column is drawn: with less room the side column
/// stacks under it, unpinned. The pin is an offset on the side column (`pinOffset`): it
/// stays `verticalPadding` below the top of the visible area, leaves with the last rows
/// and, when it is taller than the visible area, scrolls with the page.
struct ConclusionPinnedSideColumn<Main: View, Side: View>: View {
    let mainMin: CGFloat
    private let main: Main
    private let side: Side

    @State private var sideHeight: CGFloat = 0

    init(mainMin: CGFloat, @ViewBuilder main: () -> Main, @ViewBuilder side: () -> Side) {
        self.mainMin = mainMin
        self.main = main()
        self.side = side()
    }

    var body: some View {
        PVPageVisibleRect { visible, size in
            let stacked = PVPageColumnsMetrics.sideColumnWidths(
                proposal: ProposedViewSize(width: size.width, height: nil), mainMin: mainMin
            ).stacked
            PVPageColumnsLayout(widths: { [mainMin] proposal in
                PVPageColumnsMetrics.sideColumnWidths(proposal: proposal, mainMin: mainMin)
            }) {
                main
                side
                    .onGeometryChange(for: CGFloat.self) { $0.size.height } action: { sideHeight = $0 }
                    .offset(y: stacked ? 0 : PVPageColumnsMetrics.pinOffset(
                        visibleMinY: visible.minY,
                        visibleHeight: visible.height,
                        sectionHeight: size.height,
                        sideHeight: sideHeight,
                        inset: PVPageColumnsMetrics.verticalPadding
                    ))
            }
        }
    }
}
