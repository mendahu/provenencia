import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@MainActor
private final class FrameBox {
    var primary: CGRect?
    var secondary: CGRect?
    var header: CGRect?
}

private struct ColumnsProbe: View {
    let box: FrameBox
    let widthClass: PVWidthClass

    var body: some View {
        PVPageColumns(
            header: {
                Color.clear.frame(height: 40)
                    .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("probe")) } action: { box.header = $0 }
            },
            primary: {
                Color.clear.frame(height: 100)
                    .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("probe")) } action: { box.primary = $0 }
            },
            secondary: {
                Color.clear.frame(height: 100)
                    .onGeometryChange(for: CGRect.self) { $0.frame(in: .named("probe")) } action: { box.secondary = $0 }
            }
        )
        .environment(\.pvWidthClass, widthClass)
        .coordinateSpace(name: "probe")
    }
}

@MainActor @Suite(.hostedWindows)
struct PVPageColumnsTests {
    @Test(arguments: [
        (1102, PVWidthClass.regular, true, 1038, 1038, 1038),
        (1102, .regular, false, 1038, nil, 1038),
        (1600, .wide, true, 1164, 340, 1536),
        (1616, .wide, true, 1180, 340, 1552),
        (1656, .wide, true, 1180, 380, 1592),
        (1756, .wide, true, 1180, 480, 1692),
        (2296, .wide, true, 1180, 480, 1692),
        (2296, .wide, false, 1180, nil, 1180),
        (40, .regular, false, 0, nil, 0),
    ] as [(CGFloat, PVWidthClass, Bool, CGFloat, CGFloat?, CGFloat)])
    func columnWidthsFollowTheBoard(
        pane: CGFloat, widthClass: PVWidthClass, hasSecondary: Bool,
        primary: CGFloat, secondary: CGFloat?, group: CGFloat
    ) {
        let widths = PVPageColumnsMetrics.columnWidths(pane: pane, widthClass: widthClass, hasSecondary: hasSecondary)
        #expect(widths.primary == primary)
        #expect(widths.secondary == secondary)
        #expect(widths.group == group)
    }

    @Test(arguments: [
        (1038, 666, 340, 1038, false),
        (1014, 642, 340, 1014, false),
        (1013, 1013, 720, 1013, true),
        (372, 372, 372, 372, true),
        (2232, 1180, 720, 1932, false),
        (1536, 1164, 340, 1536, false),
    ] as [(CGFloat, CGFloat, CGFloat, CGFloat, Bool)])
    func sideColumnWidthsStackBelowTheMainMinimum(
        available: CGFloat, main: CGFloat, side: CGFloat, group: CGFloat, stacked: Bool
    ) {
        let widths = PVPageColumnsMetrics.sideColumnWidths(available: available, mainMin: PersonSourcesColumns.minWidth)
        #expect(widths == PVPageSideColumnWidths(main: main, side: side, group: group, stacked: stacked))
    }

    @Test func sideColumnWidthsFromAProposal() {
        let mainMin = PersonSourcesColumns.minWidth
        let fallback = PVPageColumnsMetrics.fallbackPane - 2 * PVSpacing.gutterPage
        let expected = PVPageColumnsMetrics.sideColumnWidths(available: fallback, mainMin: mainMin)
        for width in [nil, CGFloat.infinity] as [CGFloat?] {
            #expect(PVPageColumnsMetrics.sideColumnWidths(
                proposal: ProposedViewSize(width: width, height: nil), mainMin: mainMin
            ) == expected)
        }
        #expect(PVPageColumnsMetrics.sideColumnWidths(
            proposal: ProposedViewSize(width: 1038, height: nil), mainMin: mainMin
        ) == PVPageColumnsMetrics.sideColumnWidths(available: 1038, mainMin: mainMin))
    }

    @Test func theSideColumnBreakpointIsOneThousandFourteen() {
        #expect(PersonSourcesColumns.minWidth == 642)
    }

    @Test(arguments: [
        (-300, 800, 2000, 400, 0),
        (300, 800, 2000, 400, 324),
        (1800, 800, 2000, 400, 1600),
        (0, 800, 300, 400, 0),
        (300, 400, 2000, 400, 0),
        (300, 448, 2000, 400, 324),
    ] as [(CGFloat, CGFloat, CGFloat, CGFloat, CGFloat)])
    func pinOffsetHoldsTheColumnInView(
        visibleMinY: CGFloat, visibleHeight: CGFloat, sectionHeight: CGFloat, sideHeight: CGFloat, offset: CGFloat
    ) {
        #expect(PVPageColumnsMetrics.pinOffset(
            visibleMinY: visibleMinY, visibleHeight: visibleHeight, sectionHeight: sectionHeight,
            sideHeight: sideHeight, inset: PVPageColumnsMetrics.verticalPadding
        ) == offset)
    }

    @Test func unboundedProposalSizesToTheGroup() {
        for width in [nil, CGFloat.infinity] as [CGFloat?] {
            let widths = PVPageColumnsMetrics.columnWidths(
                proposal: ProposedViewSize(width: width, height: nil), widthClass: .wide, hasSecondary: true
            )
            #expect(widths.pane == widths.group)
            let alone = PVPageColumnsMetrics.columnWidths(
                proposal: ProposedViewSize(width: width, height: nil), widthClass: .wide, hasSecondary: false
            )
            #expect(alone.pane == PVSpacing.widthContentMax && alone.primary == PVSpacing.widthContentMax)
        }
        let bounded = PVPageColumnsMetrics.columnWidths(
            proposal: ProposedViewSize(width: 2296, height: nil), widthClass: .wide, hasSecondary: true
        )
        #expect(bounded.pane == 2296 && bounded.group == 1692)
    }

    @Test func fullWidthModeFillsThePrimaryAtWideOnly() {
        let gutter = PVSpacing.gutterPage
        let fill = PVPageColumnsMetrics.columnWidths(pane: 2296, widthClass: .wide, hasSecondary: false, primaryFill: true)
        #expect(fill.primary == 2296 - 2 * gutter && fill.group == 2296 - 2 * gutter)
        #expect(fill.band == PVSpacing.widthContentMax && fill.secondary == nil)

        let withSecondary = PVPageColumnsMetrics.columnWidths(pane: 2296, widthClass: .wide, hasSecondary: true, primaryFill: true)
        #expect(withSecondary.primary == 1180 && withSecondary.secondary == 480 && withSecondary.group == 1692)
        #expect(withSecondary.band == withSecondary.group)

        let regular = PVPageColumnsMetrics.columnWidths(pane: 1102, widthClass: .regular, hasSecondary: false, primaryFill: true)
        #expect(regular.primary == 1038 && regular.band == regular.group)

        let off = PVPageColumnsMetrics.columnWidths(pane: 2296, widthClass: .wide, hasSecondary: false)
        #expect(off.primary == 1180 && off.band == off.group)

        let unbounded = PVPageColumnsMetrics.columnWidths(
            proposal: ProposedViewSize(width: nil, height: nil), widthClass: .wide, hasSecondary: false, primaryFill: true
        )
        #expect(unbounded.primary == PVSpacing.widthContentMax && unbounded.band == unbounded.group)
    }

    @Test func fillSplitSideMaxTakesTheSplitGroup() {
        let gutter = PVSpacing.gutterPage
        let sideMax = PVSpacing.widthSecondaryMax
        let group = PVPageColumnsMetrics.split(available: 2296 - 2 * gutter, sideMax: sideMax).group
        let split = PVPageColumnsMetrics.columnWidths(
            pane: 2296, widthClass: .wide, hasSecondary: false, primaryFill: true, fillSplitSideMax: sideMax
        )
        #expect(split.primary == group && split.group == group && split.band == group && split.secondary == nil)

        let proposed = PVPageColumnsMetrics.columnWidths(
            proposal: ProposedViewSize(width: 2296, height: nil),
            widthClass: .wide, hasSecondary: false, primaryFill: true, fillSplitSideMax: sideMax
        )
        #expect(proposed.pane == 2296 && proposed.primary == group && proposed.band == group)

        // nil keeps today's numbers.
        let nilMax = PVPageColumnsMetrics.columnWidths(
            pane: 2296, widthClass: .wide, hasSecondary: false, primaryFill: true, fillSplitSideMax: nil
        )
        #expect(nilMax.primary == 2296 - 2 * gutter && nilMax.band == PVSpacing.widthContentMax)

        // Regular, or primaryFill off: unchanged.
        let regular = PVPageColumnsMetrics.columnWidths(
            pane: 1102, widthClass: .regular, hasSecondary: false, primaryFill: true, fillSplitSideMax: sideMax
        )
        #expect(regular.primary == 1038 && regular.band == regular.group)
        let unfilled = PVPageColumnsMetrics.columnWidths(
            pane: 2296, widthClass: .wide, hasSecondary: false, primaryFill: false, fillSplitSideMax: sideMax
        )
        #expect(unfilled.primary == 1180 && unfilled.band == unfilled.group)
    }

    @Test func viewportSizeIsZeroOutsideAPage() {
        #expect(EnvironmentValues().pvPageViewportSize == .zero)
    }

    @Test func tokensMatchTheBoard() {
        #expect(PVSpacing.widthWideBreakpoint == 1600)
        #expect(PVSpacing.widthSecondaryMax == 480)
        #expect(PVSpacing.widthInspector == 340)
        #expect(PVSpacing.widthContentMax == 1180)
    }

    @Test func wideHostPlacesTheSecondaryBeside() async {
        let wide = FrameBox()
        _ = hostInBorderlessWindow(
            ColumnsProbe(box: wide, widthClass: .wide), windowSize: CGSize(width: 2296, height: 600)
        )
        #expect(await waitUntil { wide.primary != nil && wide.secondary != nil && wide.header != nil })
        let p = wide.primary!, s = wide.secondary!
        #expect(s.minX >= p.maxX)
        #expect(abs(s.minY - p.minY) < 0.5)
        #expect(p.width == 1180 && s.width == 480)
        #expect(abs(p.minX - (2296 - 1692) / 2) < 0.5)
        #expect(abs(wide.header!.width - 1692) < 0.5)

        let regular = FrameBox()
        _ = hostInBorderlessWindow(
            ColumnsProbe(box: regular, widthClass: .regular), windowSize: CGSize(width: 1102, height: 600)
        )
        #expect(await waitUntil { regular.primary != nil && regular.secondary != nil })
        let rp = regular.primary!, rs = regular.secondary!
        #expect(rs.minY >= rp.maxY)
        #expect(abs(rs.minX - rp.minX) < 0.5)
        #expect(rp.width == 1038 && rs.width == 1038)
    }
}
