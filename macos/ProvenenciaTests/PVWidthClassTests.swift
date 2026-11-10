import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@MainActor
private final class WidthClassBox {
    var seen: PVWidthClass?
}

private struct WidthClassProbe: View {
    @Environment(\.pvWidthClass) private var widthClass
    let box: WidthClassBox

    var body: some View {
        Text(verbatim: "probe")
            .onAppear { box.seen = widthClass }
            .onChange(of: widthClass) { _, new in box.seen = new }
    }
}

@MainActor @Suite(.hostedWindows)
struct PVWidthClassTests {
    @Test(arguments: [
        (-1, PVWidthClass.regular), (0, .regular), (1102, .regular), (1464, .regular),
        (1599, .regular), (1599.5, .regular),
        (PVSpacing.widthWideBreakpoint, .wide), (1656, .wide), (2296, .wide),
    ] as [(CGFloat, PVWidthClass)])
    func classFollowsPaneWidth(width: CGFloat, expected: PVWidthClass) {
        #expect(PVWidthClass(paneWidth: width) == expected)
    }

    @Test func breakpointIsSixteenHundred() {
        #expect(PVSpacing.widthWideBreakpoint == 1600)
    }

    @Test func modifierFollowsTheWindow() {
        let box = WidthClassBox()
        let view = WidthClassProbe(box: box)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .pvMeasuresWidthClass()
        let (window, host) = hostInBorderlessWindow(view, windowSize: CGSize(width: 2000, height: 400))
        host.layoutSubtreeIfNeeded()
        #expect(box.seen == .wide)

        window.setContentSize(NSSize(width: 1200, height: 400))
        host.layoutSubtreeIfNeeded()
        #expect(box.seen == .regular)
    }

    @Test func aFilledHostMeasuresTheSpaceNotTheChild() {
        let box = WidthClassBox()
        let view = WidthClassProbe(box: box)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .pvMeasuresWidthClass()
        let (_, filled) = hostInBorderlessWindow(view, windowSize: CGSize(width: 2000, height: 400))
        filled.layoutSubtreeIfNeeded()
        #expect(box.seen == .wide)

        let bare = WidthClassBox()
        let (_, probe) = hostInBorderlessWindow(
            WidthClassProbe(box: bare), windowSize: CGSize(width: 2000, height: 400)
        )
        probe.layoutSubtreeIfNeeded()
        #expect(bare.seen == .regular)
    }
}
