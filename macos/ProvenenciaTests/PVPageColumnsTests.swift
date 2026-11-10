import SwiftUI
import Testing
@testable import Provenencia

@MainActor
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

    @Test func groupXCentersTheGroup() {
        #expect(PVPageColumnsMetrics.groupX(in: 2296, group: 1692) == 302)
        #expect(PVPageColumnsMetrics.groupX(in: 1102, group: 1038) == 32)
        #expect(PVPageColumnsMetrics.groupX(in: 1038, group: 1038) == 0)
    }

    @Test func columnsLayoutPlacesTheSecondaryBesideAtWideAndBelowAtRegular() throws {
        func arrange(width: CGFloat, widthClass: PVWidthClass, hasSecondary: Bool = true)
            -> (widths: PVPageSideColumnWidths, placed: (main: CGRect, side: CGRect?, height: CGFloat), asked: [Int])
        {
            let widths = PVPageColumnsMetrics.layoutWidths(
                proposal: ProposedViewSize(width: width, height: nil), widthClass: widthClass,
                hasSecondary: hasSecondary, primaryFill: false, fillSplitSideMax: nil
            )
            var asked: [Int] = []
            let placed = PVPageColumnsLayout.arrange(width: width, widths: widths) { index, width in
                asked.append(index)
                return CGSize(width: width, height: 100)
            }
            return (widths, placed, asked)
        }

        let wide = arrange(width: 2296, widthClass: .wide)
        let main = wide.placed.main, side = try #require(wide.placed.side)
        #expect(!wide.widths.stacked)
        #expect(main.width == 1180 && main.minX == (2296 - 1692) / 2 && main.minY == 0)
        #expect(side.width == 480 && side.minX == main.maxX + PVSpacing.space9 && side.minY == 0)
        #expect(wide.placed.height == 100)

        let regular = arrange(width: 1102, widthClass: .regular)
        let rMain = regular.placed.main, rSide = try #require(regular.placed.side)
        #expect(regular.widths.stacked)
        #expect(rMain.width == 1038 && rSide.width == 1038 && rMain.minX == 32 && rSide.minX == 32)
        #expect(rSide.minY == rMain.maxY + PVSpacing.space9)
        #expect(regular.placed.height == 100 + PVSpacing.space9 + 100)

        let alone = arrange(width: 2296, widthClass: .wide, hasSecondary: false)
        #expect(alone.widths.side == nil && alone.placed.side == nil)
        #expect(alone.asked == [0])
        #expect(alone.placed.height == 100)
    }
}
