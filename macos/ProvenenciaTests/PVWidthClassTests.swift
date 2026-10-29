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

@MainActor @Suite
struct PVWidthClassTests {
    @Test(arguments: [
        (-1, PVWidthClass.regular), (0, .regular), (1102, .regular), (1464, .regular),
        (1599, .regular), (1599.5, .regular),
        (PVSpacing.widthWideBreakpoint, .wide), (1600, .wide), (1656, .wide), (2296, .wide),
    ] as [(CGFloat, PVWidthClass)])
    func classFollowsPaneWidth(width: CGFloat, expected: PVWidthClass) {
        #expect(PVWidthClass(paneWidth: width) == expected)
    }

    @Test func modifierFollowsTheWindow() async {
        let box = WidthClassBox()
        let view = WidthClassProbe(box: box)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .pvMeasuresWidthClass()
        let (window, _) = hostInBorderlessWindow(view, windowSize: CGSize(width: 2000, height: 400))
        #expect(await waitUntil { box.seen == .wide })

        window.setContentSize(NSSize(width: 1200, height: 400))
        #expect(await waitUntil { box.seen == .regular })
    }

    @Test func aFilledHostMeasuresTheSpaceNotTheChild() async {
        let box = WidthClassBox()
        let view = WidthClassProbe(box: box)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .pvMeasuresWidthClass()
        let (window, _) = hostInBorderlessWindow(view, windowSize: CGSize(width: 2000, height: 400))
        #expect(await waitUntil { box.seen == .wide })

        let bare = WidthClassBox()
        let (bareWindow, _) = hostInBorderlessWindow(
            WidthClassProbe(box: bare), windowSize: CGSize(width: 2000, height: 400)
        )
        #expect(await waitUntil { bare.seen == .regular })
        withExtendedLifetime((window, bareWindow)) {}
    }
}
