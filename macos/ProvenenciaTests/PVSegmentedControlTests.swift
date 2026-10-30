import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@MainActor
private final class Chosen {
    var values: [Int] = []
    /// True once the hosted control has been laid out and can be clicked.
    var isLaidOut = false
}

@MainActor @Suite
struct PVSegmentedControlTests {
    private func control(_ chosen: Chosen) -> some View {
        PVSegmentedControl(
            "Kind",
            items: [
                .init("One", value: 1, accessibilityIdentifier: "seg.one"),
                .init("Two", value: 2, isDisabled: true, accessibilityIdentifier: "seg.two"),
                .init(text: "Three", value: 3, accessibilityIdentifier: "seg.three"),
            ],
            selection: 1,
            select: { chosen.values.append($0) }
        )
        .frame(width: 400)
        .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { chosen.isLaidOut = $0 > 0 }
    }

    /// Clicks the middle of segment `index` of three equal ones in the 400pt
    /// control, as the pointer would.
    private func click(segment index: Int, in window: NSWindow) {
        let x = 2 + (396.0 / 3) * (Double(index) + 0.5)
        let point = NSPoint(x: x, y: window.contentView!.bounds.midY)
        for type in [NSEvent.EventType.leftMouseDown, .leftMouseUp] {
            guard let event = NSEvent.mouseEvent(
                with: type, location: point, modifierFlags: [], timestamp: ProcessInfo.processInfo.systemUptime,
                windowNumber: window.windowNumber, context: nil, eventNumber: 0, clickCount: 1, pressure: 1
            ) else { continue }
            window.sendEvent(event)
        }
    }

    @Test func selectSendsTheItemValue() async {
        let chosen = Chosen()
        let (window, _) = hostInBorderlessWindow(control(chosen), windowSize: CGSize(width: 400, height: 80))
        #expect(await waitUntil { chosen.isLaidOut })
        click(segment: 2, in: window)
        #expect(await waitUntil { chosen.values == [3] })
        click(segment: 0, in: window)
        #expect(await waitUntil { chosen.values == [3, 1] })
        withExtendedLifetime(window) {}
    }

    @Test func disabledItemDoesNotSelect() async {
        let chosen = Chosen()
        let (window, _) = hostInBorderlessWindow(control(chosen), windowSize: CGSize(width: 400, height: 80))
        #expect(await waitUntil { chosen.isLaidOut })
        click(segment: 1, in: window)
        click(segment: 2, in: window)
        #expect(await waitUntil { chosen.values == [3] })
        withExtendedLifetime(window) {}
    }
}
