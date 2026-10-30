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
            header: { _ in
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

@MainActor @Suite
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

    @Test func tokensMatchTheBoard() {
        #expect(PVSpacing.widthWideBreakpoint == 1600)
        #expect(PVSpacing.widthSecondaryMax == 480)
        #expect(PVSpacing.widthInspector == 340)
        #expect(PVSpacing.widthContentMax == 1180)
    }

    @Test func wideHostPlacesTheSecondaryBeside() async {
        let wide = FrameBox()
        let (wideWindow, _) = hostInBorderlessWindow(
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
        let (regularWindow, _) = hostInBorderlessWindow(
            ColumnsProbe(box: regular, widthClass: .regular), windowSize: CGSize(width: 1102, height: 600)
        )
        #expect(await waitUntil { regular.primary != nil && regular.secondary != nil })
        let rp = regular.primary!, rs = regular.secondary!
        #expect(rs.minY >= rp.maxY)
        #expect(abs(rs.minX - rp.minX) < 0.5)
        #expect(rp.width == 1038 && rs.width == 1038)
        withExtendedLifetime((wideWindow, regularWindow)) {}
    }

    @Test(arguments: [
        (0, 500, false, false), (8, 500, false, false), (9, 500, false, true),
        (4, 500, true, true), (0, 500, true, false), (-20, 500, true, false),
        (50, 119, false, false), (50, 120, false, true),
        (50, 70, true, true), (50, 0, true, false), (50, 70, false, false), (0, 300, true, false),
    ] as [(CGFloat, CGFloat, Bool, Bool)])
    func scrolledFollowsTheBoard(offset: CGFloat, overflow: CGFloat, was: Bool, expected: Bool) {
        #expect(PVPageColumnsMetrics.isScrolled(offset: offset, overflow: overflow, wasScrolled: was) == expected)
    }

    @Test func compactingDoesNotUndoItself() {
        for overflow in stride(from: CGFloat(120), to: 236, by: 4) {
            let compact = PVPageColumnsMetrics.isScrolled(offset: 50, overflow: overflow, wasScrolled: false)
            let after = PVPageColumnsMetrics.isScrolled(offset: 50, overflow: overflow - 116, wasScrolled: compact)
            #expect(after == compact, "overflow \(overflow)")
        }
    }
}
